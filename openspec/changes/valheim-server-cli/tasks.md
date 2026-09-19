# Tasks

## 1. CLI shell and command routing
- [x] Define the root CLI entry point in Go.
- [x] Add subcommands for `init`, `start`, `stop`, `status`, `register`, and `backup`.
- [x] Wire each command to a shared error/output model for consistent UX.

## 2. Configuration and state model
- [x] Define the managed server configuration directory and metadata schema.
- [x] Add persisted state for lifecycle status and registration metadata.
- [x] Ensure each state file is created and validated through a single helper path.

## 3. Initialization and prerequisite validation
- [x] Implement dependency and permission checks for the supported environment.
- [x] Create the required directories and config templates.
- [x] Fail fast with actionable errors when prerequisites are missing.

## 4. Lifecycle operations
- [x] Implement `start` with state validation and service launch logic.
- [x] Implement `stop` with graceful shutdown and state updates.
- [x] Implement `status` to report running, stopped, misconfigured, or unhealthy states.

## 5. Registration and startup enablement
- [x] Implement `register` to persist server identity and operational configuration.
- [x] Validate or create the system service definition for the dedicated Valheim instance.
- [x] Use `systemctl` or the equivalent OS service manager to enable startup on boot and confirm service availability.
- [x] Fail clearly when startup integration is unavailable or permission is denied.

## 6. Backup scaffold
- [x] Keep `backup` in the CLI surface as a placeholder command.
- [x] Document the missing details required before real backup support is implemented: source data, destination, retention, and restore expectations.
- [x] Delay implementation until those details are specified.

## 7. Verification and polish
- [x] Add focused tests for command parsing, error handling, and state transitions.
- [x] Validate the CLI behavior on a Debian-compatible environment using a real smoke test path.
- [x] Update user-facing documentation for setup, operational lifecycle, and recovery steps.
