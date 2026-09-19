# Tasks

## 1. Boundary audit

- [x] 1.1 Inventory all state, lifecycle, registration, and backup behaviors currently owned by the root app and map them to their owning command or shared service boundary
- [x] 1.2 Validate which behaviors are command-specific versus shared runtime concerns and record the intended ownership before moving code

## 2. Contract extraction

- [x] 2.1 Define the narrow service contracts needed for app state, lifecycle actions, backup operations, and environment checks without leaking root-specific implementation details
- [x] 2.2 Adjust command constructors to depend on the defined interface rather than the full concrete app implementation where that boundary is useful

## 3. Command package cleanup

- [x] 3.1 Move each command’s domain logic into its package-owned implementation while preserving the existing command names and output messages
- [x] 3.2 Keep the root command and `main.go` focused on runtime bootstrap and command registration only
- [x] 3.3 Confirm the help output and root command registration still list the same top-level commands after the refactor

## 4. Test and validation

- [x] 4.1 Reposition command-specific tests to the command package or command-owned boundary and keep root tests focused on registration behavior
- [x] 4.2 Run the Go test suite and validate that the refactor preserves CLI behavior and regression coverage
- [x] 4.3 Review the final package layout and verify the root app no longer acts as the owner of command-domain logic
