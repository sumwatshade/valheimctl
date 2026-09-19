# Tasks

## 1. Command package structure

- [x] 1.1 Define the target `cmd/<subcmd>` layout and the shared command package contract to preserve root CLI behavior.
- [x] 1.2 Create the initial directory structure for each top-level command and verify file placement matches the planned package boundaries.

## 2. Command extraction and wiring

- [x] 2.1 Move each top-level command constructor and handler into its dedicated package while preserving existing flags and command names.
- [x] 2.2 Update the root Cobra registration to add subcommands from the package-based command modules and verify help output still lists the same command set.
- [x] 2.3 Keep shared app state and initialization logic centralized so command packages use one runtime contract instead of duplicating setup code.

## 3. Validation and regression safety

- [x] 3.1 Run the focused Go test suite for CLI behavior and command registration to verify the package layout change preserved functionality.
- [x] 3.2 Check the command help output and a representative subcommand invocation to confirm the external interface remains stable.
- [x] 3.3 Review the final command layout and confirm it supports isolated modeling for future top-level commands.
