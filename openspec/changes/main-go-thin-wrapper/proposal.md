# Proposal

## Why

The CLI’s business logic and runtime orchestration are still concentrated in the root application layer, so the entrypoint and shared app object end up carrying command-specific responsibilities. That makes the command surface harder to reason about, harder to isolate, and harder to extend without coupling unrelated behavior into the same package boundary.

## What Changes

- Move the server lifecycle, state management, prerequisite checks, and backup operations behind command-specific or shared service boundaries instead of keeping them rooted in the CLI entrypoint.
- Keep `main.go` and the root command bootstrap as a thin wrapper whose only job is bootstrapping the app and registering subcommands.
- Preserve the existing external command names and UX while aligning the internal layout to a command-first model.
- Continue to centralize shared runtime concerns, but do so behind small interfaces or internal packages rather than a monolithic root app implementation.
- Refine the command test layout so behavior is verified in the package or command boundary that owns it.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `cli-command-architecture`: expand the existing command architecture requirement so the root CLI entrypoint is a thin registration layer and all domain behavior lives in subcommand or shared service contracts.

## Impact

- Affects the Go structure of the Valheim CLI, especially the root bootstrap and the ownership boundaries for lifecycle, register, and backup logic.
- Improves maintainability, testability, and future extension of command-specific behavior without changing the user-facing command names.
- Encourages reuse of narrow interfaces and internal packages for configuration, OS detection, service setup, and backup operations.
- Keeps the command UX stable while reducing drift between command behavior and root-level wiring.
