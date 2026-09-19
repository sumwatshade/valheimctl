# Spec Delta

## Purpose

This capability organizes the Valheim CLI around a command-first folder structure so each top-level subcommand can be modeled and maintained as an isolated unit while sharing the same application context.

## ADDED Requirements

### Requirement: Top-level commands are packaged by subcommand
The system SHALL organize each root-level CLI action into a dedicated command package or folder under a `cmd/<subcmd>` layout rather than leaving related command logic in flat root-level files.

#### Scenario: A top-level command is added or modified
- **WHEN** a new or existing top-level CLI action is modeled
- **THEN** the system SHALL keep that behavior within a dedicated command package that is scoped to that command’s responsibilities

#### Scenario: The command surface remains stable
- **WHEN** a user invokes a CLI subcommand
- **THEN** the system SHALL present the same external command names and behavior as before, even though the internal package layout changes

### Requirement: Shared runtime wiring remains centralized
The system SHALL keep shared app configuration, state access, and common integration points in a reusable runtime context even as each command is isolated in its own package.

#### Scenario: A command needs common app services
- **WHEN** a subcommand runs
- **THEN** the system SHALL use the shared runtime context instead of duplicating initialization logic inside every command package

#### Scenario: Common validation is reused across commands
- **WHEN** command validation or setup logic is required
- **THEN** the system SHALL source that logic from shared, reused behavior instead of re-implementing it inside each command package

### Requirement: Shared package contracts are defined with Go interfaces
The system SHALL extract general-purpose behaviors behind small Go interfaces so command packages depend on stable contracts rather than on concrete implementation details from the root application.

#### Scenario: A command depends on state and system operations
- **WHEN** a subcommand needs to read persisted state, install dependencies, or interact with the host service manager
- **THEN** the system SHALL provide those capabilities through explicit interfaces that can be implemented by production code and lightweight test doubles

#### Scenario: A new command reuses existing capabilities
- **WHEN** another CLI action requires similar runtime behavior
- **THEN** the system SHALL allow that command to reuse the same shared interfaces without coupling to a specific concrete service implementation

#### Scenario: Behavior boundaries remain clear
- **WHEN** the internal implementation changes
- **THEN** the system SHALL keep command logic stable as long as the underlying interface contract remains valid

### Requirement: Command isolation improves modeling and testing
The system SHALL allow each command package to be reasoned about independently so command-specific logic, flags, and tests remain focused.

#### Scenario: A command is tested in isolation
- **WHEN** a developer verifies a specific command behavior
- **THEN** the system SHALL allow focused validation of that command without coupling the test to unrelated root-level command code

#### Scenario: Command organization scales cleanly
- **WHEN** additional top-level commands are added in the future
- **THEN** the system SHALL allow those commands to fit into the same `cmd/<subcmd>` structure without restructuring the entire CLI
