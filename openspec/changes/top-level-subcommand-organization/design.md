# Design

## Context

The project already has individual command implementations such as `init.go`, `start.go`, `stop.go`, and `backup.go`, but they are laid out as flat top-level files around the root CLI entrypoint. The next step is to make the command structure more model-friendly by grouping each top-level action into its own `cmd/<subcmd>` package while keeping the command surface and runtime behavior consistent.

## Goals / Non-Goals

**Goals:**
- Group each top-level subcommand into a dedicated package folder.
- Preserve the current external CLI names and semantics.
- Keep shared app state and initialization centralized.
- Make each command easier to model and test independently.
- Extract reusable service contracts so command packages depend on interfaces instead of concrete implementations.

**Non-Goals:**
- Redesign the Valheim server lifecycle or command behaviors themselves.
- Introduce additional external frameworks or runtime dependencies.
- Reorganize unrelated project code beyond the command package structure.

## Decisions

1. Use a `cmd/<subcmd>` layout for each root command.
   - Each subcommand gets a logically isolated folder with its own command constructor, flags, and execution path.
   - This aligns the repo with a command-first structure and makes the CLI easier to reason about as it expands.
   - Alternative considered: continuing to keep all command logic in flat root files; rejected because it keeps command responsibilities mixed together and makes future additions harder to model.

2. Keep shared app wiring in the existing application runtime model.
   - The underlying `app` object and config/state helpers remain the shared source of truth for environment and server operations.
   - Command packages consume that shared runtime rather than each re-implementing setup logic.
   - Alternative considered: duplicating app initialization inside every command package; rejected because it increases drift and weakens testability.

3. Preserve a single root Cobra registration layer.
   - The root command remains responsible for registering subcommands, but the actual command implementation is delegated to the package-specific command constructors.
   - This keeps the CLI API stable while moving implementation organization closer to command boundaries.
   - Alternative considered: creating a fully custom dispatcher or moving logic to unrelated package names; rejected because it introduces more indirection than the project needs.

4. Extract reusable service contracts behind Go interfaces.
   - Shared behavior should be defined in small interfaces such as `StateStore`, `ServiceManager`, `PackageInstaller`, `OSDetector`, and command execution contracts rather than by depending on the concrete `app` type.
   - This keeps command packages testable with lightweight fakes and allows future implementations to swap in alternate backends without changing command logic.
   - Example boundaries:
     - `internal/config` or `internal/store`: persistence and state reading/writing
     - `internal/platform`: OS detection, package installation, and systemd/service integration
     - `internal/cli`: command registration and command execution contracts used by root wiring
     - `internal/ports` or `internal/contracts`: shared interfaces that command packages depend on
   - Alternative considered: passing the full concrete `app` object from every command package; rejected because it couples command logic to implementation details and makes testing more brittle.

## Risks / Trade-offs

- [Command-package churn] → Mitigation: keep a small, consistent contract for each `cmd/<subcmd>` package so refactors remain predictable.
- [Hidden coupling across commands] → Mitigation: centralize shared app configuration and validation instead of scattering them across each command package.
- [Migration complexity] → Mitigation: move one command group at a time and validate the root command registration after each step.

## Migration Plan

1. Define the target `cmd/<subcmd>` structure and command package contract.
2. Identify the concrete behaviors that belong in shared packages: state persistence, Linux distro detection, package installation, and service management.
3. Extract those behaviors behind narrow Go interfaces and move command-specific logic to subcommand packages that depend on those contracts.
4. Move each top-level command to its package-specific folder while preserving its Cobra registration and flags.
5. Keep shared app state and initialization in place so commands still access the same runtime configuration through the interface layer.
6. Re-run the CLI and focused tests to confirm the command surface is unchanged.
7. Add or refine tests to cover command registration and individual command behavior after the layout move.

## Open Questions

- Should the refactor also move any shared helper functions into a small internal “command support” package, or keep them in the app layer for now?
- Is the repository prepared for gradually relocating commands one package at a time, or should the entire command tree be migrated in one pass?
- Which shared behaviors should be promoted into explicit Go interfaces first to maximize future reuse without over-engineering the initial refactor?
