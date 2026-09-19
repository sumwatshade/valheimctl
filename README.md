# valheimctl

Tools to help run and manage Valheim servers, primarily focused on Debian Linux systems.

## Overview

`valheimctl` is a small Go CLI for managing a dedicated Valheim server lifecycle. The current structure favors a simple command-oriented design with a shared application context, a Cobra root command, and thin handlers for each subcommand.

## CLI architecture

The project uses Cobra as the command framework and keeps command setup separate from server behavior.

- Root command and command registration live in [root.go](root.go)
- Shared runtime configuration and state helpers live in [main.go](main.go)
- Each subcommand is implemented as its own file:
  - [init.go](init.go)
  - [start.go](start.go)
  - [stop.go](stop.go)
  - [status.go](status.go)
  - [register.go](register.go)
  - [backup.go](backup.go)

This layout keeps the command surface discoverable and makes it easier to add new commands without growing a single monolithic dispatcher.

## Shared app context

The `app` type centralizes the runtime state that commands need:

- root directory for configuration and state files
- custom `systemctl` binary path for testing and environment overrides
- service directory for generated systemd unit files
- helpers for state initialization, JSON read/write, and service registration

Commands are intentionally thin wrappers around `app` methods, which keeps the behavior consistent across CLI and tests.

## State and service patterns

The tool stores operational state under a local `.valheimctl` directory in the configured root:

- `.valheimctl/state.json` stores lifecycle state like initialization and running status
- `.valheimctl/server.json` stores server metadata such as name, path, and description
- generated service files are written under the systemd service directory for the target host

`register` is the place where the CLI creates a service unit and invokes `systemctl enable`, matching the project requirement to support startup registration on Linux systems.

## Testing patterns

Tests are organized by command rather than keeping all cases in one file. This makes it easier to understand what each command does and keeps failures focused.

Examples:

- [init_test.go](init_test.go)
- [start_test.go](start_test.go)
- [stop_test.go](stop_test.go)
- [register_test.go](register_test.go)
- [backup_test.go](backup_test.go)
- [root_test.go](root_test.go)

The test suite favors real filesystem behavior over mock-heavy assertions. It creates temporary directories, writes realistic state/config files, and validates the command outputs and side effects that matter in practice.

## Local development

Run the current regression suite with:

```bash
go test ./...
```

This project is intentionally small and straightforward: keep commands thin, business logic centralized, and tests focused on observable behavior.
