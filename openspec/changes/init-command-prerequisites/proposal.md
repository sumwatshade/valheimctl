# Proposal

## Why

The `init` command needs a formal prerequisite contract before it creates or prepares a Valheim server environment. Without explicitly defined package and SteamCMD requirements, setup can fail unpredictably across Linux distributions and leave the system in a partially configured state.

## What Changes

- Add a formal capability for bootstrapping a server environment during `init`
- Require installation of the OS libraries needed by Valheim and SteamCMD
- Require SteamCMD installation through the supported OS-specific workflow
- Fail fast with clear operator guidance when prerequisites cannot be satisfied
- Keep initialization safe to rerun and idempotent when packages are already present

## Capabilities

### New Capabilities
- `server-init`: Defines the prerequisite installation and bootstrap workflow for initializing a dedicated Valheim server environment.

### Modified Capabilities
- None

## Impact

- `init` command behavior and validation flow
- Linux package installation and dependency checks
- SteamCMD installation path and OS-specific requirements
- operator documentation and troubleshooting guidance
