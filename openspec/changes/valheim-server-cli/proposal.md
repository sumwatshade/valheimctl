# Proposal

## Why

Managing a dedicated Valheim server today requires a mix of ad hoc shell scripts, manual system setup, and inconsistent service state tracking. This makes initialization fragile, lifecycle operations harder to reason about, and recovery slower when the server needs to be restarted or restored. A single Go-based CLI gives operators one place to prepare prerequisites, manage the server lifecycle, register server metadata, and create backups.

## What Changes

- Add a Go CLI with commands for `init`, `start`, `stop`, `status`, `register`, and `backup`.
- Add a consistent command surface for preparing a dedicated Valheim environment, including prerequisite validation and configuration generation.
- Add lifecycle state tracking so users can determine whether the server is running, stopped, or unhealthy.
- Add support for registering server configuration metadata and creating backups of server data and config.
- Add operational safeguards so failed prerequisite checks stop the workflow before the server is started.

## Capabilities

### New Capabilities
- `valheim-server-management`: provides the CLI workflow for initializing, starting, stopping, checking, registering, and backing up a dedicated Valheim server.

### Modified Capabilities
- None.

## Impact

- Adds a new Go CLI surface in the project root for server administration.
- Introduces local runtime state, persisted configuration, and backup artifact paths for machine setup and operations.
- Relies on OS-level dependencies such as service management, filesystem layout, and any required Valheim install tooling.
- Affects how operators manage a dedicated Valheim instance and may require documentation updates for installation and recovery procedures.
