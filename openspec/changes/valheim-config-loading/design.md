# Design

## Context

The project is building a dedicated Valheim server management tool, and the CLI will eventually need to assemble and validate the command-line arguments that the Valheim server executable accepts. This change introduces the configuration contract and validation layer before that behavior is tied to startup logic.

## Goals / Non-Goals

**Goals:**
- Define a configuration model for the server settings that map directly to Valheim console flags.
- Keep validation explicit and deterministic so bad configuration is rejected early.
- Allow future CLI startup logic to consume a single internal config object instead of rebuilding argument assembly ad hoc.
- Support explicit default handling without hiding user intent.

**Non-Goals:**
- Replacing the Valheim server binary or changing actual game behavior.
- Hardcoding all startup semantics into the CLI at this stage.
- Implementing a full interactive config editor in the planning phase.

## Decisions

1. Use a typed config model for all supported Valheim arguments.
   - A struct-based representation gives each field a clear type, default, and validation rule.
   - It keeps the public contract stable while still allowing future conversion to the command arguments used by the binary.
   - Alternative considered: passing arbitrary map-based settings everywhere; rejected because type safety and validation become harder to maintain.

2. Validate before startup rather than late in the launch path.
   - Server configuration should fail fast when a value is blank, out of range, or not in the supported enumeration.
   - This reduces operator confusion and prevents partial or unsafe startup states.
   - Alternative considered: letting the runtime fail when the server process starts; rejected because it pushes validation too late and makes diagnosis slower.

3. Keep the loader and validator as an internal package boundary.
   - The package should parse config inputs and return a normalized config object and explicit validation errors.
   - This separates configuration concerns from the CLI command structure and keeps startup logic easier to test.
   - Alternative considered: embedding validation directly in command handlers; rejected because it couples config semantics to command wiring and discourages reuse.

4. Treat defaults as explicit config state, not as hidden behavior.
   - The package should preserve the difference between omitted values and user-specified values while still honoring Valheim defaults.
   - This makes future behavior easier to reason about and document.
   - Alternative considered: ignoring omitted values entirely; rejected because it obscures the actual effective config state.

5. Keep the schema discipline aligned to Valheim’s actual argument names.
   - Each config key will map one-to-one to known Valheim flags such as `name`, `port`, `world`, `public`, `saveinterval`, `backups`, and `preset`.
   - This reduces ambiguity and ensures the config file remains grounded in actual server behavior.
   - Alternative considered: inventing custom names without a direct mapping; rejected because it creates friction between config files and the server binary.

## Risks / Trade-offs

- [Validation drift from upstream Valheim changes] → Mitigation: keep the config model explicitly mapped to Valheim’s documented flags and review it when the server behavior changes.
- [Ambiguous defaults across platforms] → Mitigation: represent platform-specific defaults in the schema as documented values and validate any platform-specific override before use.
- [Overly strict validation] → Mitigation: allow only known-supported values and clearly explain the accepted ranges and enums for each option.
- [Configuration complexity] → Mitigation: separate required settings, optional settings, and defaults into the same typed model so operator intent remains easy to inspect.

## Migration Plan

1. Define the config struct and supported field list for the Valheim server arguments.
2. Add a loader that accepts the configuration file format and returns a typed object.
3. Add validation rules for required values, numeric ranges, enums, path semantics, and compatibility checks.
4. Convert the validated config into the final command-line argument list used by server startup logic.
5. Add focused tests covering valid configs, invalid values, and default behavior.
6. Document the config schema and example values in the project docs once the implementation is ready.

## Open Questions

- Which config file format should be supported first: YAML, JSON, or TOML?
- Should validation treat `public` as an explicit boolean or as a numeric flag mirroring the binary’s `1`/`0` contract?
- Are there any save-path or log-path assumptions that should be platform-specific at the CLI layer rather than in the config package?
