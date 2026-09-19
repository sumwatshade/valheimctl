# Proposal

## Why

The current Valheim CLI is a single-file Go entrypoint, which makes commands harder to evolve, test, and maintain as new operations are added. A more durable command architecture is needed so each subcommand can be authored, validated, and extended independently without duplicating setup logic or making the root command brittle.

## What Changes

- Introduce a structured command architecture built around Cobra for command registration and execution.
- Split command behavior into individual command packages or files so each CLI action can be authored in isolation.
- Add durable patterns for shared app state, dependency injection, and common error handling across commands.
- Introduce Bubbles-based patterns where interactive prompts or richer terminal UX are warranted.
- Add command-specific tests so behaviors such as lifecycle management, registration, and validation remain stable as the CLI grows.

## Capabilities

### New Capabilities
- `cli-command-architecture`: defines the reusable command framework, command decomposition model, and testability patterns for the Valheim CLI.

### Modified Capabilities
- None.

## Impact

- Affects the project’s Go CLI authoring model and command layout.
- Introduces external dependencies on Cobra and Bubbles-style terminal patterns.
- Changes how the main entrypoint is structured and how future commands are added, validated, and tested.
- Improves maintainability for individual server-management commands without changing the overall operating goal of the tool.
