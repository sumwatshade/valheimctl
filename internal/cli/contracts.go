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
