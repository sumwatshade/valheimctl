# Design

## Context

The project is a Go-based CLI for managing a dedicated Valheim server. The current codebase is a minimal starter and does not yet define a command model, configuration layout, or server management workflow. The design needs to establish a clear structure for operational state, dependency checks, and lifecycle actions without assuming a specific hosting platform beyond the project’s Debian-focused objective.

## Goals / Non-Goals

**Goals:**
- Provide a consistent CLI entry point with commands for initialization, start, stop, status, registration, and backup scaffolding.
- Validate operating prerequisites before performing destructive or state-changing actions.
- Keep configuration and runtime state in a predictable directory layout for a dedicated server setup.
- Ensure server registration configures the service for automatic startup through the host service manager.

**Non-Goals:**
- Building a multiplayer matchmaking service or remote fleet orchestration.
- Managing multiple server instances in a single host abstraction beyond a single dedicated server profile.
- Implementing a full backup-and-restore workflow before the backup requirements are fully specified.

## Decisions

1. Command-first architecture with a single main entrypoint.
   - The CLI will use a root command and subcommands rather than a single monolithic action handler.
   - This keeps each operation explicit and matches the project goal of a server-management tool.
   - Alternative considered: a single command with flags for all actions; rejected because lifecycle operations and validation become harder to reason about and document.

2. Structured configuration and runtime state under a dedicated local directory.
   - The tool will define a managed configuration directory, server metadata file, and state file for lifecycle status.
   - This keeps initialization, registration, status, and startup behavior consistent and auditable.
   - Alternative considered: storing state only in a transient process memory; rejected because the CLI must support restart, inspection, and service lifecycle checks across sessions.

3. Prerequisite validation before state-changing operations.
   - `init`, `start`, and registration flows will verify directories, permissions, and external tooling before continuing.
   - This reduces the risk of starting a broken instance or creating a service entry that cannot run correctly on boot.
   - Alternative considered: eager execution with best-effort logging; rejected because silent misconfiguration is a serious operational risk.

4. Service registration through the host service manager.
   - The `register` command will validate the service definition and invoke `systemctl` or the equivalent service manager to enable the service for startup.
   - This ensures that a registered server is both configured and operationally ready to survive a reboot.
   - Alternative considered: only storing metadata without managing service startup; rejected because the CLI is meant to manage a dedicated server as a real service, not just a config record.

5. Backup remains intentionally deferred as a scaffold.
   - The `backup` command will exist as a placeholder so the CLI surface is stable, but real backup behavior will not be implemented until data sources, destination strategy, and retention rules are defined.
   - This avoids premature contract design for a feature that still requires operational details.
   - Alternative considered: implementing a backup command with guessed semantics; rejected because it would produce a brittle and potentially misleading behavior contract.

## Risks / Trade-offs

- [State drift] → Mitigation: persist explicit metadata and validate configuration before start/stop actions.
- [Incomplete prerequisite detection] → Mitigation: centralize validation in one init and lifecycle helper layer so all commands use the same checks.
- [Startup registration failures] → Mitigation: validate the operating system service manager and fail with actionable guidance when startup integration is unavailable.
- [Host-specific differences] → Mitigation: treat OS assumptions as configuration constraints and keep the CLI logic explicit about supported Debian-like setups.

## Migration Plan

1. Introduce the Go CLI and basic command scaffolding.
2. Add initialization and prerequisite validation as the first operational path.
3. Add server registration and state tracking, including service-manager integration for startup enablement.
4. Implement start/stop/status with explicit state checks.
5. Revisit backup implementation after the required data sources, destinations, and retention policy are defined.
6. Validate each phase through focused CLI tests and a manual operational smoke check in a Debian-compatible environment.

## Open Questions

- What exact external runtime dependencies are required for the first supported Valheim setup (for example, SteamCMD, service manager, or installation directory conventions)?
- Which local directory structure should be the canonical default for managed server state in the target Linux environment?
- What exact backup source, destination, and retention model should be defined before the backup command becomes a real feature?
