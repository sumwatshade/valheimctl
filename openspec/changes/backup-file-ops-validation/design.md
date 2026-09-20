# Design

## Context

The project already includes a backup command surface and a backup service that operates on a directory tree. The main gap was not the CLI shape; it was that the service was coupled directly to the host OS, which made it difficult to test and obscured the real filesystem contract behind `os.*` calls. The goal is to keep the backup behavior intact while making the file system boundary explicit, implementation-detail based, and easy to swap for tests.

## Goals / Non-Goals

**Goals:**
- Keep backup creation, listing, and apply behavior tied to the configured save root and world directory semantics.
- Make the service accept an injected filesystem implementation as an internal dependency, without exposing that dependency as a domain-level API.
- Use the Go standard-library `io/fs` interfaces where they match read behavior while keeping the write and remove operations explicit.
- Validate the service with filesystem-backed tests that simulate realistic save trees using `testing/fstest` and/or temporary directories.

**Non-Goals:**
- Replacing the Valheim save format or changing how world files are laid out on disk.
- Introducing a new backup storage format beyond the existing directory layout.
- Making the backup service depend on a live game runtime or external shared state.

## Decisions

1. Treat the backup service as the filesystem boundary, not the CLI layer.
   - The service should own the file operations that are required for listing, creating, and applying backups.
   - This keeps the backup contract stable and makes it straightforward to exercise in tests without changing the user-facing command names.
   - Alternative considered: keeping all filesystem behavior in the command layer; rejected because it bakes OS assumptions into the wrong layer and makes end-to-end behavior harder to validate.

2. Inject the filesystem as an implementation detail of the service constructor.
   - The service can accept an optional filesystem dependency while defaulting to the real OS filesystem in production.
   - This makes tests deterministic and allows a `fstest.MapFS` or fake to be provided at construction time.
   - Alternative considered: making a filesystem dependency a public domain-level abstraction; rejected because it broadens the API surface unnecessarily and exposes an implementation concern outside the service boundary.

3. Use standard library `io/fs` interfaces for the read side, but keep the writable operations explicit.
   - Read behavior maps naturally to `fs.FS`, `fs.ReadDirFS`, `fs.ReadFileFS`, and `fs.StatFS`.
   - Write, mkdir, and remove operations do not have a standard counterpart in the same abstraction, so the service keeps those methods in a small custom interface used only for mutation.
   - This yields the most idiomatic Go contract while remaining explicit about the operations the service actually performs.

4. Keep restore semantics focused and explicit: overwrite the configured world directory only.
   - `backup apply` should verify the backup exists, select the target world directory according to configuration, and copy the backup data into that directory after removing the current target if present.
   - This keeps behavior predictable and avoids clobbering unrelated world folders.
   - Alternative considered: merge into the current directory without deletion; rejected because it leaves stale files behind and makes restore semantics harder to reason about.

5. Validate the service with filesystem-backed tests instead of host-dependent tests.
   - The service should be exercised against temporary directories and synthetic filesystem trees, not only live OS files.
   - These tests can verify list/create/apply semantics and missing-backup failures without requiring a real Valheim install.
   - Alternative considered: only relying on integration tests against the real developer machine; rejected because it creates hidden environment coupling and makes tests flaky.

### Filesystem boundary

The backup logic is built around a small internal contract that abstracts the operations the service actually needs, while still leaning on the standard library for reads when appropriate:

```go
type readFS interface {
    fs.FS
    fs.ReadDirFS
    fs.ReadFileFS
    fs.StatFS
}

type writableFS interface {
    MkdirAll(path string, perm fs.FileMode) error
    WriteFile(name string, data []byte, perm fs.FileMode) error
    RemoveAll(path string) error
}

type FileSystem interface {
    readFS
    writableFS
}
```

The production implementation wraps the real OS filesystem, while the tests can inject a custom fake or a `fstest.MapFS`-based adapter. This makes the backup service portable without requiring the CLI or callers to know about the filesystem implementation at all.

The service constructor is intentionally local to the implementation detail:

```go
type Service struct {
    RootDir string
    fs      FileSystem
}

func NewService(rootDir string, fsys ...FileSystem) *Service {
    service := &Service{RootDir: rootDir, fs: osFS{}}
    if len(fsys) > 0 && fsys[0] != nil {
        service.fs = fsys[0]
    }
    return service
}
```

This allows test code to provide a synthetic filesystem while preserving the default runtime behavior in production.

### Test model using `fstest`

The intended testing pattern is:

1. Build a synthetic filesystem tree that represents a Valheim save root and one or more backups.
2. Instantiate the service with that filesystem at the configured root.
3. Exercise `List`, `Create`, and `Apply` against the synthetic tree.
4. Assert on file presence, metadata semantics, and the overwrite behavior for the configured world directory.

This keeps the tests deterministic and fast while still exercising real file operation semantics in a controlled manner.

## Risks / Trade-offs

- [Host OS coupling] → Mitigation: hidden behind an internal filesystem abstraction and constructor injection.
- [Overexposed abstraction] → Mitigation: keep the interface narrow and local to the service layer, rather than exposing it as a public contract.
- [Read/write contract mismatch] → Mitigation: use stdlib `io/fs` interfaces for read behavior and keep only the mutation methods in the custom portion of the contract.
- [Unsafe restore behavior] → Mitigation: validate the target world directory, remove or replace only the intended world, and test the apply path against synthetic trees.
- [Test brittleness] → Mitigation: prefer simple filesystem adapters and explicit directory layouts to avoid hidden mock behavior.

## Migration Plan

1. Add the backup service constructor overload that accepts an injected filesystem implementation.
2. Wrap the real OS filesystem behind an adapter that satisfies the internal contract.
3. Update backup service methods to use the injected filesystem instead of direct `os.*` calls.
4. Add service-level tests covering create/list/apply behavior and missing-backup failures.
5. Add synthetic-filesystem tests using `fstest.MapFS` to validate the injected-filesystem path end-to-end.
6. Review the command tests and service tests together to ensure the CLI still behaves the same while the service contract becomes more testable.

## Open Questions

- Should the project eventually move the filesystem abstraction into a shared internal utility package if more services require the same pattern?
- Is the current backup service contract narrow enough to stay as a private interface, or would a broader internal filesystem package become worthwhile if more save-system logic is added later?
