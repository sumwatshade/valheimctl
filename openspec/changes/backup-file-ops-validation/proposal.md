# Proposal

## Why

The backup workflow was tightly coupled to the host operating system, which made the service difficult to exercise in tests and hid the real filesystem contract behind direct `os.*` calls. The result was a backup service that worked for live filesystems but was hard to validate deterministically in unit tests and harder to reason about when the save root and world layout were intentionally synthetic.

The change addresses that by making the filesystem boundary explicit and testable without changing the CLI surface. It keeps the backup contract aligned with the configured save root and world selection, while making the service accept a filesystem implementation as an internal dependency rather than a domain-level API concern.

## What Changes

- Define the backup service around a narrow filesystem dependency that can be swapped for production or test implementations.
- Use the Go standard-library `io/fs` interfaces where they describe read behavior, while keeping the write/mkdir/remove operations explicit because the backup service mutates the filesystem.
- Make the service constructor accept an injected filesystem implementation as an internal detail so tests can use `fstest.MapFS` or a custom fake without changing command behavior.
- Preserve the configured save root and world directory semantics while restoring from a named backup and overwriting only the relevant world folder.
- Add service-level tests that cover backup creation, listing, missing-backup handling, and apply behavior against a synthetic filesystem tree.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `backup-management`: adds a deterministic filesystem abstraction for the internal backup service, explicit restore semantics for the configured world directory, and filesystem-backed validation for backup list/create/apply flows.

## Impact

- Affects the internal backup service and its constructor contract, while keeping the public backup CLI shape unchanged.
- Improves testability by allowing the service to run against temporary directories, `fstest.MapFS`, or a custom fake filesystem without coupling tests to the host OS.
- Clarifies the boundary between CLI behavior and service behavior: command tests validate command wiring; service tests validate real backup operations and filesystem semantics.
