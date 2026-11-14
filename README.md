# valheimctl

Tools to help run and manage Valheim servers, primarily focused on Debian Linux systems.

## Overview

`valheimctl` is a small Go CLI for managing a dedicated Valheim server lifecycle. The current structure favors a simple command-oriented design with a shared application context, a Cobra root command, and thin handlers for each subcommand.

## CLI architecture

The project uses Cobra as the command framework and keeps command setup separate from server behavior.

- Root command and command registration live in [root.go](root.go)
- Shared runtime configuration and state helpers live in [main.go](main.go)
- Each subcommand is implemented as its own file:
  - [init.go](cmd/init/init.go)
  - [start.go](cmd/start/start.go)
  - [stop.go](cmd/stop/stop.go)
  - [status.go](cmd/status/status.go)
  - [register.go](cmd/register/register.go)
  - [backup.go](cmd/backup/backup.go)
  - [install.go](cmd/install/install.go)
  - [update.go](cmd/update/update.go)
  - [restart.go](cmd/restart/restart.go)
  - [logs.go](cmd/logs/logs.go)
  - [version.go](cmd/version/version.go)
  - [config.go](cmd/config/config.go)
  - [world.go](cmd/world/world.go)
  - [reinstall.go](cmd/reinstall/reinstall.go)
  - [diagnostics.go](cmd/diagnostics/diagnostics.go)

This layout keeps the command surface discoverable and makes it easier to add new commands without growing a single monolithic dispatcher.

## CLI commands

### Lifecycle commands

| Command | Purpose |
|---------|---------|
| `init` | Initialize valheimctl state and check prerequisites |
| `start` | Start the tracked Valheim server |
| `stop` | Stop the tracked Valheim server |
| `status` | Report the Valheim server status |
| `register <name> <dir> <desc>` | Register a systemd service for the Valheim server |
| `restart` | Restart the Valheim server systemd service |

### Server installation and update

| Command | Purpose |
|---------|---------|
| `install` | Install Valheim dedicated server binaries via SteamCMD (AppID 896660) |
| `update [--validate] [--start]` | Update server binaries via SteamCMD; backs up saves first |
| `reinstall [--dry-run\|--execute]` | Clean reinstall preserving worlds and backups |

### Logs and version

| Command | Purpose |
|---------|---------|
| `logs [--follow] [--lines N]` | Show journalctl logs for the server service |
| `version` | Report version information about the server installation |

### Configuration

| Command | Purpose |
|---------|---------|
| `config` | Show the current server configuration |
| `config validate` | Validate the server configuration |

### World management

| Command | Purpose |
|---------|---------|
| `world [list]` | List save directories and backup candidates |
| `world inspect <name>` | Inspect a specific world directory |
| `world backup [label]` | Create a backup of the save root |
| `world restore <archive> [--dry-run]` | Restore a world from a backup archive (dry-run by default) |
| `world verify <archive>` | Verify a backup archive's integrity |

### Backup management

| Command | Purpose |
|---------|---------|
| `backup` | List existing backups (describe) |
| `backup list` | List existing backups with timestamps |
| `backup apply <name>` | Apply a named backup to the active world save directory |
| `backup create <name>` | Create a backup of all world saves (alias: `snapshot`) |

### Diagnostics

| Command | Purpose |
|---------|---------|
| `diagnostics` | Collect diagnostic information about the server setup |

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

- [init_test.go](cmd/init/init_test.go)
- [start_test.go](cmd/start/start_test.go)
- [stop_test.go](cmd/stop/stop_test.go)
- [register_test.go](cmd/register/register_test.go)
- [backup_test.go](cmd/backup/backup_test.go)
- [root_test.go](root_test.go)

The test suite favors real filesystem behavior over mock-heavy assertions. It creates temporary directories, writes realistic state/config files, and validates the command outputs and side effects that matter in practice.

## Local development

Run the current regression suite with:

```bash
go test ./...
```

## Valheim configuration contract

The project now includes an internal configuration package that accepts a Valheim server config file and converts the supported server settings into the equivalent command-line arguments for the server binary.

Example config:

```yaml
name: "My server"
port: 2456
world: "Dedicated"
password: "Secret"
public: 1
savedir: "/tmp/valheim-saves"
logfile: "/tmp/valheim.log"
saveinterval: 1800
backups: 4
backupshort: 7200
backuplong: 43200
crossplay: false
instanceid: "1"
preset: hard
modifiers:
  raids: none
setkey: nomap
```

This package validates required values, allowed enums, numeric ranges, and backup/timing conditions before the server configuration is used.

This project is intentionally small and straightforward: keep commands thin, business logic centralized, and tests focused on observable behavior.
