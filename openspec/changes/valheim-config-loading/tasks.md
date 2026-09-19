# Tasks

## 1. Config contract design

- [x] 1.1 Define the supported Valheim config fields and map each field to the corresponding console argument and default behavior.
- [x] 1.2 Document required and optional settings, including precedence rules between defaults, explicit values, and overrides.
- [x] 1.3 Verify the field model covers server identity, world settings, network settings, backup timing, modifiers, and crossplay flags.

## 2. Configuration loading and validation

- [x] 2.1 Implement the internal config loader for the supported config file format and verify it produces a typed representation.
- [x] 2.2 Add validation for required fields, invalid ports, missing passwords, and malformed path values.
- [x] 2.3 Add enum validation for preset names, world modifier sets, and supported visibility values.
- [x] 2.4 Add validation for time intervals, backup counts, and other numeric values to reject impossible or negative settings.
- [x] 2.5 Verify the loader returns actionable validation errors with field-level detail for all invalid inputs.

## 3. Startup argument mapping

- [x] 3.1 Map the validated config object to the exact Valheim command-line flags that must be passed to the server binary.
- [x] 3.2 Ensure public visibility, crossplay, world, and instance settings are converted in the expected order and format.
- [x] 3.3 Validate that custom save directories and log paths are passed through without losing operator intent.

## 4. Test coverage and regression safety

- [x] 4.1 Add tests covering a valid configuration file and the expected fields loaded into the internal model.
- [x] 4.2 Add tests covering invalid port numbers, missing values, unsupported modifiers, and invalid timing ranges.
- [x] 4.3 Add tests for default behavior and explicit override precedence.
- [x] 4.4 Run the Go test suite and verify the config package does not regress other lifecycle behaviors.

## 5. Documentation and handoff

- [x] 5.1 Provide an example configuration file showing the supported Valheim options and expected values.
- [x] 5.2 Confirm the config contract is documented clearly enough for future CLI startup and server management work.
- [x] 5.3 Review the change for readiness for implementation and confirm the artifact set is sufficient for the apply phase.
