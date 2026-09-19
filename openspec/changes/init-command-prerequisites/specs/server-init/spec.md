# Spec Delta

## Purpose

Defines the prerequisite bootstrap workflow for the Valheim server `init` command. This capability ensures the host machine has the required runtime libraries and a valid SteamCMD installation before server setup proceeds.

## ADDED Requirements

### Requirement: init installs required system libraries
The system SHALL verify that `libatomic1`, `libpulse-dev`, and `libpulse0` are present on supported Linux systems before continuing initialization.

#### Scenario: packages are missing
- **WHEN** the `init` command runs and one or more required packages are absent
- **THEN** the system SHALL install the missing packages using the platform package manager and SHALL stop initialization if installation fails

#### Scenario: packages are already installed
- **WHEN** the `init` command runs and all required packages are already present
- **THEN** the system SHALL continue without unnecessary reinstall steps

### Requirement: init bootstraps SteamCMD
The system SHALL install SteamCMD using the appropriate OS-specific installation flow for the current operating system and SHALL confirm that the `steamcmd` executable is available.

#### Scenario: Debian or Ubuntu environment
- **WHEN** the `init` command runs on a Debian-compatible Linux system
- **THEN** the system SHALL install SteamCMD using the supported package manager or documented installer flow and SHALL verify it is available on `PATH`

#### Scenario: unsupported operating system
- **WHEN** the `init` command runs on an unsupported platform
- **THEN** the system SHALL return a clear error that identifies the unsupported OS and the required SteamCMD setup steps

#### Scenario: SteamCMD installation fails or is unusable
- **WHEN** package installation completes but SteamCMD is missing, not executable, or cannot run
- **THEN** the system SHALL return a clear installation error and SHALL stop initialization

### Requirement: init is safe to rerun
The system SHALL make repeated `init` executions idempotent by checking for prerequisites before installation and by avoiding unnecessary reinstallation work.

#### Scenario: repeated initialization
- **WHEN** the operator runs `init` more than once on a prepared machine
- **THEN** the system SHALL verify existing prerequisites and continue without duplicate installation churn
