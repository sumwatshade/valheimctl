# Spec Delta

## Purpose

This capability updates the CLI command boundary so the entrypoint remains a thin registration layer and the domain logic moves into explicit command or service-owned responsibilities.

## MODIFIED Requirements

### Requirement: The root command is a thin registration layer
The system SHALL keep the root CLI entrypoint responsible only for creating the app runtime and registering subcommands, without embedding business logic or command-specific orchestration in the root package.

#### Scenario: Subcommands own their behavior
- **WHEN** a user invokes a top-level command such as `init`, `start`, `stop`, `register`, or `backup`
- **THEN** the system SHALL route execution to the relevant command package or shared service contract instead of performing the logic within the root package

#### Scenario: Help output remains stable
- **WHEN** an operator runs the CLI help or lists the available commands
- **THEN** the system SHALL continue to expose the same command set and usage patterns even though the implementation is now organized by command boundary

### Requirement: Domain logic is moved behind explicit boundaries
The system SHALL keep server lifecycle, prerequisites, backup, and registration functionality in command-owned or shared internal packages, using narrow interfaces or contracts so behavior is not secretly coupled to the root CLI wrapper.

#### Scenario: Command packages are isolated
- **WHEN** a command needs to perform a lifecycle action or inspect runtime state
- **THEN** the command SHALL call a well-defined contract or internal implementation owned by that command or its shared service layer rather than reach directly into the root entrypoint

#### Scenario: Shared behavior remains reusable
- **WHEN** multiple commands need the same app capabilities such as state persistence or environment validation
- **THEN** the system SHALL expose that capability through shared abstraction rather than duplicating the logic in each command package

### Requirement: Command behavior remains testable in isolation
The system SHALL allow command-specific and service-level tests to validate behavior without requiring the root CLI entrypoint to carry the implementation itself.

#### Scenario: A command can be tested without a full root assembly
- **WHEN** a command or shared service contract is exercised in a unit test
- **THEN** the system SHALL support direct invocation of that contract without coupling the test to root-level wrappers

#### Scenario: root wiring is validated separately
- **WHEN** the root command is checked for registration and discoverability
- **THEN** the system SHALL verify only that command registration is correct, not the full business logic of each command
