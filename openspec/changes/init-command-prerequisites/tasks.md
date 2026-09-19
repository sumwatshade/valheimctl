## 1. Define init prerequisite contract
- [x] 1.1 Capture required libraries and SteamCMD installation requirements in the delta spec
- [x] 1.2 Confirm OS-specific expectations for package manager and SteamCMD installation
- [x] 1.3 Validate the requirement is represented as a formal capability and not a hidden implementation detail

## 2. Implement init bootstrapping behavior
- [x] 2.1 Add prerequisite detection for `libatomic1`, `libpulse-dev`, and `libpulse0`
- [x] 2.2 Install missing packages via the platform package manager on supported Linux systems
- [x] 2.3 Add SteamCMD installation logic for the supported distro flow
- [x] 2.4 Validate the `steamcmd` binary is on PATH and executable after installation

## 3. Error handling and safety
- [x] 3.1 Fail clearly when the OS is unsupported or package installation does not succeed
- [x] 3.2 Make the flow idempotent so repeated `init` calls do not reinstall unnecessary packages
- [x] 3.3 Produce actionable operator guidance when prerequisite installation fails

## 4. Validation
- [x] 4.1 Add tests covering missing dependency installation, SteamCMD availability, and unsupported-OS behavior
- [x] 4.2 Run the Go test suite to confirm the init workflow remains valid
- [x] 4.3 Review the command output and documentation to ensure operator guidance is clear
