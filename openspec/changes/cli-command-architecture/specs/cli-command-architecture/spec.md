# Spec Delta

## Purpose

This capability establishes a durable command architecture for the Valheim CLI so commands can be authored and tested independently while sharing a consistent runtime model, UX pattern, and operational behavior.

## ADDED Requirements

### Requirement: CLI command entries are modeled as first-class subcommands
The system SHALL organize the command surface around a root CLI entrypoint with clearly defined subcommands instead of a monolithic `main.go` execution path.

#### Scenario: Command registration is explicit and structured
- **WHEN** a new CLI action is added to the tool
- **THEN** the system SHALL expose it as an individual command with consistent naming, flags, and lifecycle wiring

#### Scenario: Command help is discoverable
- **WHEN** an operator runs the CLI help output
- **THEN** the system SHALL list the available commands and their purpose in a discoverable, predictable format

### Requirement: Shared command patterns are reusable across actions
The system SHALL provide a reusable command authoring pattern for shared app context, execution setup, and error handling so individual commands do not duplicate boilerplate.

#### Scenario: Common context is reused by commands
- **WHEN** multiple commands require configuration, environment validation, or runtime state access
- **THEN** the system SHALL provide one shared context path rather than duplicating setup logic per command

#### Scenario: Command-specific errors remain actionable
- **WHEN** a command fails due to missing prerequisites or invalid input
- **THEN** the system SHALL surface a focused error message that identifies the relevant action and what needs to change

### Requirement: Interactive terminal patterns are supported where needed
The system SHALL allow interactive terminal experiences, including prompt or rendering patterns, for commands that benefit from richer user interaction without forcing all commands into a TUI model.

#### Scenario: Interactive prompt flows use reusable terminal patterns
- **WHEN** a command requires confirmation, selection, or richer terminal output
- **THEN** the system SHALL use a shared, reusable pattern compatible with the CLI framework rather than ad hoc terminal code

#### Scenario: Non-interactive commands stay straightforward
- **WHEN** a command is designed for automation or scripting
- **THEN** the system SHALL support clear non-interactive behavior without requiring terminal UI complexity

### Requirement: Command implementation is testable in isolation
The system SHALL support unit and command-level testing of handlers and command behavior independent of the entire process entrypoint.

#### Scenario: Command behavior is validated without invoking the whole CLI process
- **WHEN** a command’s logic is tested
- **THEN** the system SHALL allow verification through focused tests against the command or handler layer

#### Scenario: Common validation rules are reused in tests
- **WHEN** the same validation or state logic is used across commands
- **THEN** the system SHALL make that logic available for testing without duplicating business rules in command code
