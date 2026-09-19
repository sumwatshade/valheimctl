# Design

## Context

See proposal.md - the `init` command needs a clear prerequisite contract before it prepares a Valheim server runtime. The project already treats `init` as a first-class command, but the current implementation does not yet codify external system dependencies or SteamCMD installation requirements.

This project is currently Linux-focused and assumes a systemd-based environment. Package names and installation flows are not portable across operating systems: the underlying libraries and SteamCMD are typically named differently on Debian/Ubuntu, Fedora/RHEL, and Arch-based systems. The design therefore treats OS detection and distro-aware package resolution as a required part of `init` instead of a best-effort convenience layer.

## Goals / Non-Goals

**Goals:**
- Define a reusable prerequisite check for the `init` command on supported Linux distributions
- Handle required system libraries consistently with distro-aware package names and package manager behavior
- Install SteamCMD using an OS-specific flow with clear validation
- Keep the command safe to re-run and easy for operators to diagnose
- Fail clearly on unsupported operating systems, including non-systemd environments and non-Linux hosts

**Non-Goals:**
- Replacing the system package manager with a custom installer
- Supporting every Linux distribution in a single generic flow
- Defining a future server update or configuration migration system
- Claiming support for macOS, Windows, or non-systemd Linux implementations

## Decisions

1. Centralize prerequisite checks inside the `init` command flow.
   - The command should validate system readiness before creating server state.
   - This keeps runtime setup logic close to the command entrypoint and avoids duplicating dependency checks into unrelated commands.
   - Alternative considered: lazy checks spread across `start` and `register`; rejected because failure becomes harder to diagnose and patch consistently.

2. Treat package names as OS-specific, not universal.
   - `libatomic1`, `libpulse-dev`, and `libpulse0` are examples of the Debian/Ubuntu package names, but other Linux distributions use different names for the same underlying libraries.
   - The design therefore requires OS detection and a distro-to-package mapping before installation begins.
   - Alternative considered: using one fixed package list across all systems; rejected because it fails on unsupported distros and makes the init contract inaccurate.

3. Scope the design to Linux with systemd.
   - The current project explicitly depends on Linux behavior such as `systemctl` and service file generation, so support is intentionally limited to Linux-based hosts with systemd.
   - This is not a cross-platform CLI yet; unsupported hosts should exit with a clear message rather than pretending to support them.
   - Alternative considered: claiming generic portability without verification; rejected because it would create broken behavior on non-Linux systems.

4. Install SteamCMD via a distro-aware wrapper that validates the executable.
   - Debian/Ubuntu flows should use the supported package manager path or documented installer method.
   - Other supported Linux distributions should map to the equivalent package or installation flow.
   - Non-supported operating systems should exit with actionable guidance instead of attempting a best-effort install.
   - Alternative considered: silently assuming SteamCMD is installed; rejected because it violates the requirement and creates non-deterministic runtime behavior.

5. Keep initialization idempotent.
   - If the required libraries and SteamCMD are already present, the command should continue without unnecessary reinstall churn.
   - This reduces the risk of repeated installs in automation or operator reruns.
   - Alternative considered: always reinstalling dependencies; rejected as expensive and noisy.

## Risks / Trade-offs

- [Package manager differences] → Mitigation: detect OS family and use the supported Linux path, with a clear unsupported-platform error for anything else. Package names vary by distro and must be mapped explicitly rather than assumed.
- [Linux-only assumptions] → Mitigation: keep the project scoped to Linux with systemd and document that unsupported hosts are intentionally rejected.
- [Installer drift] → Mitigation: keep the install steps aligned with the current canonical SteamCMD documentation and vendor-supported packaging flows.
- [Operator misconfiguration] → Mitigation: validate `steamcmd` availability and fail with a diagnostic error when PATH or installation state is wrong.
- [Repeated init runs] → Mitigation: treat installation checks as idempotent and only install what is missing.

## Migration Plan

1. Add the prerequisite check to the `init` command flow.
2. Detect OS and package manager before attempting package installation and confirm the host is a supported Linux/systemd environment.
3. Resolve the correct package names for the detected distro and install any missing required libraries.
4. Install or validate SteamCMD in the OS-supported manner.
5. Return actionable errors for unsupported systems or failed package installation.
6. Re-run the init flow to confirm the command remains idempotent and diagnostic.

## Open Questions

- Which Linux distributions are explicitly supported, beyond Debian/Ubuntu?
- Should `init` expose a `--skip-prereqs` flag for advanced operator workflows, or is strict enforcement the desired default?
