# Proposal

## Why

The project needs a first-class configuration model for Valheim server runtime settings before it can reliably launch or manage dedicated servers from a reproducible input source. The current CLI architecture is focused on lifecycle operations, but it does not yet define a structured way to load and validate the server arguments that correspond to Valheim’s console flags.

## What Changes

- Add a dedicated internal configuration package for loading a Valheim server config file.
- Define the config keys that map to Valheim command-line arguments such as name, port, world, password, save directory, public visibility, log file, presets, modifiers, and backup settings.
- Validate required values, numeric ranges, enum constraints, and file-path semantics before the server is started.
- Expose a single, reusable configuration model that can be consumed by the CLI and future server startup logic.
- Keep the config contract explicit enough to support safe defaults, validation failures, and documentation for operator configuration.

## Capabilities

### New Capabilities
- `valheim-config`: defines the internal Valheim server configuration contract, file-loading behavior, and validation rules for command arguments that map to the server executable.

### Modified Capabilities
- None.

## Impact

- Affects future Valheim server startup and lifecycle code paths by standardizing the configuration inputs they consume.
- Adds a new internal package and validation layer, with no direct change to the user-facing command surface in this planning phase.
- Enables safer startup behavior by failing early on invalid ports, modifier combinations, log paths, or missing required values.
- Establishes the schema for operator-managed config files that can be reused across development, testing, and deployment scenarios.
