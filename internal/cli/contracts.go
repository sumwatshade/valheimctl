package cli

import "time"

type BackupSummary struct {
	Name      string
	CreatedAt time.Time
}

type Service interface {
	Init() error
	Start() (string, error)
	Stop() (string, error)
	Status() (string, error)
	Register(serverName, serverDir, description string) error
	Backup() (string, error)
	BackupList() ([]BackupSummary, error)
	BackupApply(name string) error
	BackupCreate(name string) error
}
