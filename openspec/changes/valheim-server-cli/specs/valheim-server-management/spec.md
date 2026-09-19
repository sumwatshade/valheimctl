# Spec Delta

## Purpose

This capability gives operators a consistent CLI workflow for preparing, managing, and registering a dedicated Valheim server so it can start reliably and remain observable throughout its lifecycle.

## ADDED Requirements

### Requirement: CLI initializes a dedicated Valheim environment
The system SHALL provide an `init` command that validates required prerequisites, creates the expected server configuration layout, and prepares the server for a supported dedicated Valheim deployment.

#### Scenario: Initialization succeeds with valid prerequisites
- **WHEN** the operator runs `init` in a supported environment with required dependencies available
- **THEN** the system validates prerequisites, creates the managed configuration directories, and reports a successful initialization state

#### Scenario: Initialization fails when prerequisites are missing
- **WHEN** the operator runs `init` and required dependencies or permissions are missing
- **THEN** the system SHALL stop initialization and return a clear error describing the unmet prerequisites

### Requirement: CLI starts and stops the server lifecycle
The system SHALL provide `start` and `stop` commands that manage the dedicated Valheim server lifecycle and report whether the server is currently running.

#### Scenario: Server starts successfully
- **WHEN** the operator runs `start` after a successful initialization and the server is not already running
- **THEN** the system SHALL start the dedicated server process and report the running state

#### Scenario: Server stop succeeds
- **WHEN** the operator runs `stop` while the server is running
- **THEN** the system SHALL stop the dedicated server process and report the stopped state

#### Scenario: Start or stop is rejected when state is invalid
- **WHEN** the operator runs `start` or `stop` while the server state is not valid for the requested action
- **THEN** the system SHALL return a clear state-based error instead of proceeding with an unsafe action

### Requirement: CLI reports server status
The system SHALL provide a `status` command that reports the current lifecycle state of the dedicated Valheim server and whether required configuration and prerequisites are present.

#### Scenario: Status reports healthy running state
- **WHEN** the operator runs `status` for a properly initialized server that is running
- **THEN** the system SHALL return a running status with the relevant configuration and operational health summary

#### Scenario: Status reports unhealthy or missing setup
- **WHEN** the operator runs `status` for a server that is not initialized, not configured, or not running as expected
- **THEN** the system SHALL report the invalid or missing setup state and the known reason

### Requirement: CLI registers server identity and configuration
The system SHALL provide a `register` command that stores the server's identity, configuration metadata, and any required operational details needed for managed hosting, and it SHALL ensure the service is configured to run automatically through the host service manager (`systemctl` or equivalent).

#### Scenario: Registration stores valid server metadata and enables startup
- **WHEN** the operator runs `register` with complete server metadata and a valid service definition
- **THEN** the system SHALL persist the configuration, validate or create the service unit, enable startup on boot, and report that the server is registered and startup-enabled

#### Scenario: Registration rejects incomplete metadata
- **WHEN** the operator runs `register` without required identity or configuration values
- **THEN** the system SHALL reject the request and explain which values are required

#### Scenario: Registration fails when startup integration is unavailable
- **WHEN** the operator runs `register` in an environment without a supported service manager or without permission to configure startup
- **THEN** the system SHALL report the startup integration failure and not claim that the service is registered for automatic startup

### Requirement: CLI exposes a backup scaffold for future implementation
The system SHALL provide a `backup` command as a scaffold for future backup behavior, but it SHALL remain intentionally deferred until the required backup scope, destination, and retention model are specified.

#### Scenario: Backup command is present but intentionally not implemented
- **WHEN** the operator runs `backup` before the backup design is finalized
- **THEN** the system SHALL return a clear "not yet implemented" message and document the missing details required for a full backup workflow

#### Scenario: Backup implementation details are requested
- **WHEN** the operator or project needs to define the backup contract
- **THEN** the system SHALL require explicit specification of the backup source data, destination, retention policy, and restore expectations before implementation proceeds
