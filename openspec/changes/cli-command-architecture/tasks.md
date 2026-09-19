## 1. CLI framework setup
- [ ] 1.1 Add Cobra as the root command framework and wire the application entrypoint to it.
- [ ] 1.2 Define a shared app context for state access, service helpers, and configuration directories.
- [ ] 1.3 Establish consistent command execution, flag parsing, and help output patterns.

## 2. Command decomposition
- [ ] 2.1 Extract the current `init`, `start`, `stop`, `status`, `register`, and `backup` behaviors into individual command handlers.
- [ ] 2.2 Replace the monolithic `main.go` dispatch logic with dedicated command definitions and command registration.
- [ ] 2.3 Keep shared validation logic centralized so command code stays small and testable.

## 3. Interactive and terminal UX patterns
- [ ] 3.1 Decide which commands require a richer interactive terminal experience and which remain command-line only.
- [ ] 3.2 Add reusable Bubbles-style prompt or UI patterns only where interactive flows are justified.
- [ ] 3.3 Ensure non-interactive commands remain script-friendly and deterministic.

## 4. Testing and regression safety
- [ ] 4.1 Add unit tests for command registration and root command behavior.
- [ ] 4.2 Add command-level tests for lifecycle, validation, and service registration flows.
- [ ] 4.3 Verify command help output and failure modes remain clear and stable.

## 5. Final validation
- [ ] 5.1 Run the Go test suite to validate the refactor.
- [ ] 5.2 Review the command layout and confirm it supports durable future extension.
- [ ] 5.3 Document the architectural pattern for future command authors.
