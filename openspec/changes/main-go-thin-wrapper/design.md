# Design

## Context

The project already uses a root command plus command packages, but the underlying application object still owns much of the domain logic. In the current structure, `main.go` and the shared app model are responsible for both bootstrapping the CLI and carrying the runtime behavior for init, start, stop, register, status, and backup. This makes the root layer more than a thin entrypoint and creates a boundary problem when the command surface grows.

## Goals / Non-Goals

**Goals:**
- Keep the root package focused on constructing the CLI and registering subcommands.
- Move command-specific business logic into command packages or narrow shared service interfaces.
- Preserve the command names, behavior, and user-facing output of the current CLI.
- Keep the app runtime as the shared dependency for state/config access without allowing it to become a monolithic behavioral owner.

**Non-Goals:**
- Changing the Valheim server lifecycle or user workflow semantics.
- Introducing a new framework or a different CLI model solely for this refactor.
- Rewriting unrelated project code outside the command boundary and shared service layer.

## Decisions

1. Root command bootstrap stays minimal.
   - The root command remains the repository’s registration point, but its responsibilities are limited to creating the shared runtime and attaching command constructors.
   - Alternative considered: leaving all domain logic in the root and command-specific packages as wrappers; rejected because it preserves the same coupling and keeps root-level behavior hard to reason about.

2. Command packages own the command behavior, while the app remains the shared runtime.
   - Commands such as `init`, `start`, `stop`, `status`, `register`, and `backup` should call narrow interfaces exposed by the runtime or service layer instead of implementing logic inline in the root.
   - This preserves a stable app model while making ownership and testing clearer.
   - Alternative considered: duplicating app functions in each command package; rejected because it fragments logic and makes the architecture harder to maintain.

3. Shared logic is extracted behind interfaces or internal packages before it becomes a command dependency.
   - Areas such as configuration persistence, OS detection, package installation, systemd registration, and backup storage should be expressed through consistent contracts that command packages depend on.
   - This supports future command additions and keeps logic testable without root coupling.
   - Alternative considered: continuing to pass the full concrete app through all command constructors; rejected because that keeps behaviors coupled to implementation detail and reduces testability.

4. Command tests live at the command boundary.
   - Package-level tests validate the command layer itself, while root tests confirm the registry and command names remain stable.
   - This keeps behavior checks close to the owners of that behavior and avoids monolithic root tests.
   - Alternative considered: keeping all tests at the root package; rejected because it undermines the package ownership model introduced by the refactor.

## Risks / Trade-offs

- [Coupling remains hidden in the app layer] → Mitigation: keep the shared runtime narrow and expose only the behaviors required by commands through interfaces.
- [Refactor churn] → Mitigation: preserve command names and outputs exactly so the CLI surface remains stable during the extraction.
- [Over-abstracting too early] → Mitigation: introduce interfaces only around behaviors that are already shared or clearly cross-command in scope.

## Migration Plan

1. Inventory the current behaviors owned by the root app and map each one to a command-specific owner or a shared service contract.
2. Define the narrow interfaces for runtime actions such as state initialization, lifecycle control, service registration, and backup management.
3. Move command execution logic into the appropriate `cmd/<subcmd>` package while keeping the app object as the shared runtime dependency.
4. Reduce `main.go` and the root command to creating the runtime and registering the commands.
5. Update the command-level tests to validate the ownership boundaries and keep root-level tests focused on command registration.
6. Re-run the Go test suite to confirm the CLI surface and behavior remain unchanged.

## Open Questions

- Should the remaining shared runtime behaviors be extracted into a dedicated internal service package immediately, or should the refactor stop at command-level contracts for now?
- Are there any future commands expected to require a stronger system-level abstraction beyond the current service interface pattern?
