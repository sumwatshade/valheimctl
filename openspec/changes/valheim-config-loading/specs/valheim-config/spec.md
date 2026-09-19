# Spec Delta

## Purpose

This capability defines the internal Valheim server configuration contract for loading and validating a configuration file that maps to the command-line arguments used by the dedicated server executable.

## ADDED Requirements

### Requirement: Configuration file loads a structured Valheim server definition
The system SHALL accept a configuration file containing the operator-defined Valheim server settings needed to start and run the dedicated server using the equivalent command-line options.

#### Scenario: Configuration loads with valid values
- **WHEN** the operator provides a valid config file that includes supported Valheim settings
- **THEN** the system SHALL parse the configuration into a typed internal representation without losing the mapping to the corresponding executable arguments

#### Scenario: Configuration fails with malformed content
- **WHEN** the operator supplies a config file with invalid types, malformed values, or unreadable structure
- **THEN** the system SHALL reject the file and return a clear validation error that identifies the offending field

### Requirement: Server identity and network settings are validated
The system SHALL validate the server name, port, world name, password, public visibility flag, and save-related settings before the server is launched.

#### Scenario: Port is within the supported range
- **WHEN** the operator sets a server port value that is valid for the chosen deployment model
- **THEN** the system SHALL accept the configuration and preserve the port as the authoritative value for the server

#### Scenario: Invalid port or missing required values are rejected
- **WHEN** the operator provides an invalid port, a blank server name, or a missing required field for a required runtime setting
- **THEN** the system SHALL reject the configuration and explain which values are invalid or missing

### Requirement: Save and log paths are normalized and checked
The system SHALL accept save directory and log file settings, normalize them to the platform-appropriate semantics, and validate that the configured paths are usable.

#### Scenario: Save directory is configured explicitly
- **WHEN** the operator sets a custom `savedir` path
- **THEN** the system SHALL preserve that path in the internal config and ensure it is treated as the override for worlds and permission files

#### Scenario: Log file path is configured
- **WHEN** the operator sets a log file location
- **THEN** the system SHALL preserve the location and validate it as a writable or createable log destination when applicable

### Requirement: World modifiers and preset rules are validated
The system SHALL validate preset names, modifier keys, and modifier values against the supported Valheim options and apply the correct precedence rules.

#### Scenario: Preset and modifier values are valid
- **WHEN** the operator provides a valid preset such as `hard` or a valid modifier assignment such as `-modifier raids none`
- **THEN** the system SHALL accept the configuration and preserve the intended world settings

#### Scenario: Invalid modifier or invalid preset is rejected
- **WHEN** the operator supplies a preset or modifier outside the accepted set of Valheim values
- **THEN** the system SHALL reject the configuration and return the invalid selector or value

### Requirement: Backup and timing settings are validated for coherence
The system SHALL validate backup counts and time intervals so the configuration remains internally consistent and matches Valheim’s expected semantics.

#### Scenario: Backup interval values are accepted
- **WHEN** the operator provides valid backup count and interval settings
- **THEN** the system SHALL accept the values and keep them in the configuration object for later command generation

#### Scenario: Incoherent or negative timing values are rejected
- **WHEN** the operator provides a negative or impossible backup or save interval
- **THEN** the system SHALL reject the configuration and explain the invalid timing settings

### Requirement: Crossplay and instance id settings are recognized
The system SHALL recognize and validate `crossplay`, `instanceid`, and related compatibility flags when presented in a config file.

#### Scenario: Crossplay is enabled deliberately
- **WHEN** the operator sets `crossplay` to the expected enabled state
- **THEN** the system SHALL preserve that state in the config and mark the server as using the PlayFab backend

#### Scenario: Instance ID is optionally declared
- **WHEN** the operator provides an `instanceid` for a multi-server deployment on the same host
- **THEN** the system SHALL validate the value and retain it without rejecting the broader config

### Requirement: The configuration contract separates defaults, overrides, and explicit values
The system SHALL distinguish between default values, explicit user-provided values, and invalid values so that the runtime can reuse the config object without silently dropping operator intent.

#### Scenario: Default values are preserved when not set
- **WHEN** the operator omits a field that has a standard Valheim default
- **THEN** the system SHALL still represent the value consistently in the config model and allow the server runtime to apply the default behavior

#### Scenario: Explicit values override defaults
- **WHEN** the operator sets a value that differs from the standard default
- **THEN** the system SHALL preserve the explicit user setting and ensure it wins over any inherited default behavior

### Requirement: The config package exposes a safe, testable validation surface
The system SHALL provide a single internal validation API that can be used by the CLI or future startup logic to parse a config object and return validation errors in a consistent structure.

#### Scenario: Validation is reusable across call sites
- **WHEN** a caller uses the config loader and validator on a config file or deserialized map
- **THEN** the system SHALL return the same validation semantics regardless of the input source

#### Scenario: Validation supports actionable errors
- **WHEN** the config contains one or more invalid fields
- **THEN** the system SHALL return a deterministic list of validation issues that identify each field and its problem
