# Proposal

## Why

The Valheim CLI already separates command behaviors into individual Go files, but the top-level command structure is still organized around a flat repository layout rather than a command-first package model. This makes it harder to isolate each command’s implementation, test surface, and modeling concerns as the CLI grows.

## What Changes

- Introduce a `cmd/<subcmd>` package structure for each top-level Cobra command so command logic can be isolated by function and responsibility.
- Move command-specific entrypoints, flags, and behavior into dedicated subcommand folders instead of keeping them alongside unrelated root-level files.
- Preserve the existing CLI behavior and command surface while improving maintainability and modeling boundaries.
- Add a command-package pattern that keeps shared app wiring in one place while each subcommand remains independently focused.
- Maintain a single root command that registers the subcommand packages without forcing a monolithic command implementation.

## Capabilities

### New Capabilities
- `command-organization`: defines the package-level command structure for isolating each top-level CLI subcommand into its own `cmd/<subcmd>` folder.

### Modified Capabilities
- None.

## Impact

- Affects the Go command layout for the Valheim CLI and the organization of command-related files in the repository.
- Improves how each top-level action is modeled, tested, and extended without coupling unrelated command logic.
- Keeps the tool’s external UX stable while clarifying the internal project structure for future command additions.
- Reduces the risk of root-level command drift as new operations are introduced.
