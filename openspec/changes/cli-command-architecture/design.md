# Design

## Context

The project currently uses a single-file Go entrypoint for all CLI behavior, which makes command growth and validation harder to manage. The command surface has expanded to include lifecycle and registration workflows, and the next step is to adopt a more durable architectural pattern that preserves behavior while reducing duplication and making each command independently testable.

## Goals / Non-Goals

**Goals:**
- Introduce a reusable CLI framework for authoring commands individually.
- Separate root command setup from command-specific business logic.
- Provide a shared context and dependency pattern for state, configuration, and service interactions.
- Add testing-friendly boundaries around handlers and shared command logic.

**Non-Goals:**
- Replacing the functional requirements of the Valheim management commands.
- Converting every command into a full interactive terminal application.
- Reworking the underlying server lifecycle design beyond the command architecture itself.

## Decisions

1. Use Cobra as the root command framework.
   - Cobra provides a structured command model, flag parsing, and consistent help output.
   - This is a better match than building custom CLI dispatch logic in one file.
   - Alternative considered: continuing with a manual switch statement in `main`; rejected because it scales poorly and is harder to test.

2. Split command definitions into individual files or packages.
   - Each command should own its flags, execution logic, and validation behavior without being coupled to the rest of the CLI.
   - This makes command growth manageable and reduces accidental cross-command logic leakage.
   - Alternative considered: a large shared command registry file; rejected because it still centralizes logic and makes tests noisy.

3. Use shared app context for runtime configuration.
   - Each command will receive a shared runtime context containing configuration paths, service-manager hooks, and state access helpers.
   - This keeps command code focused on behavior while preserving reuse across commands.
   - Alternative considered: global variables; rejected because they are harder to test and couple command execution to process state.

4. Use Bubbles-oriented patterns selectively for interactivity.
   - Bubbles and Bubble Tea are appropriate for prompts, status screens, or interactive confirmation flows when they genuinely improve UX.
   - Commands that are meant for scripting or automation should remain non-interactive and deterministic.
   - Alternative considered: forcing a TUI for every command; rejected because that creates unnecessary complexity for simple operational commands.

## Risks / Trade-offs

- [Command sprawl] → Mitigation: maintain a common application context and shared validation hooks so commands do not diverge in structure.
- [Over-engineering the CLI] → Mitigation: keep the TUI layer optional and only use interactive components where the user workflow benefits from them.
- [Testing complexity] → Mitigation: isolate handlers and shared logic behind reusable interfaces so unit tests remain focused and quick.
- [Library churn] → Mitigation: prefer a stable core architecture with clear boundaries, leaving dependency updates isolated to command lifecycle and terminal integration points.

## Migration Plan

1. Add the Cobra-based root command and shared app context.
2. Move existing server management behavior into individual command handlers or command-specific files.
3. Standardize validation and error handling around the shared app context.
4. Add command-specific tests for configuration, state, and CLI routing behaviors.
5. Introduce Bubbles-style interactive patterns only for flows that need user prompts or richer terminal output.
6. Validate the refactor by checking help output, command execution, and isolated tests for the new structure.

## Open Questions

- Which individual commands need a richer interactive prompt model versus a simple flag-based interface?
- Should the command package structure be flat per-command files or grouped by functional modules?
- Which command behaviors should be extracted into reusable service interfaces before additional commands are added?
