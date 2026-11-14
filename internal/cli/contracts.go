package cli

import "time"

type BackupSummary struct {
	Name      string
	CreatedAt time.Time
}

type ServerService interface {
	Init() error
	Start() (string, error)
	Stop() (string, error)
	Status() (string, error)
	Register(serverName, serverDir, description string) error
}

// ServerManagement extends ServerService with full Valheim dedicated server
// lifecycle management: install, update, restart, logs, version, config,
// world operations, backup, reinstall, and diagnostics.
type ServerManagement interface {
	ServerService

	// Install installs the Valheim dedicated server binaries via SteamCMD.
	Install() (string, error)

	// Update updates the Valheim dedicated server binaries via SteamCMD.
	// It backs up the save root before updating.
	Update(validate bool, start bool) (string, error)

	// Restart restarts the systemd service.
	Restart() (string, error)

	// Logs returns recent journalctl logs for the service.
	// follow tails the log; lines limits the number of lines returned.
	Logs(follow bool, lines int) (string, error)

	// Version reports version information about the server installation.
	Version() (string, error)

	// ConfigShow displays the current server configuration.
	ConfigShow() (string, error)

	// ConfigValidate validates the current server configuration.
	ConfigValidate() (string, error)

	// WorldList lists save directories and backup candidates.
	WorldList() (string, error)

	// WorldInspect inspects a specific world directory.
	WorldInspect(name string) (string, error)

	// WorldBackup creates a backup of the save root.
	WorldBackup(label string) (string, error)

	// WorldRestore restores a world from a backup archive.
	// dryRun performs a validation-only check when true.
	WorldRestore(archive string, dryRun bool) (string, error)

	// WorldVerify verifies a backup archive's integrity.
	WorldVerify(archive string) (string, error)

	// BackupCreate creates a backup of the save root.
	BackupCreate(label string) (string, error)

	// BackupList lists existing backups.
	BackupList() (string, error)

	// BackupInspect inspects a backup archive.
	BackupInspect(archive string) (string, error)

	// ServerReinstall performs a clean reinstall preserving worlds.
	// dryRun performs a validation-only check when true.
	ServerReinstall(dryRun bool) (string, error)

	// Diagnostics collects diagnostic information about the server setup.
	Diagnostics() (string, error)
}

// BackupService remains for the backup subcommand.
type BackupService interface {
	Backup() (string, error)
	BackupList() ([]BackupSummary, error)
	BackupApply(name string) error
	BackupCreate(name string) error
}

// Service remains as a compatibility interface for callers that still need the
// combined server+backup contract, but command packages should consume the
// narrower interfaces they actually require.
type Service interface {
	ServerService
	BackupService
}
