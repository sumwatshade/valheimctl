package backupcmd

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/sumwatshade/valheimctl/internal/cli"
)

type backupFake struct {
	listed    []cli.BackupSummary
	applied   string
	created   string
	backupMsg string
}

func (f *backupFake) Init() error { return nil }
func (f *backupFake) Start() (string, error) { return "", nil }
func (f *backupFake) Stop() (string, error) { return "", nil }
func (f *backupFake) Status() (string, error) { return "", nil }
func (f *backupFake) Register(serverName, serverDir, description string) error { return nil }
func (f *backupFake) Backup() (string, error) { return f.backupMsg, nil }
func (f *backupFake) BackupList() ([]cli.BackupSummary, error) { return f.listed, nil }
func (f *backupFake) BackupApply(name string) error { f.applied = name; return nil }
func (f *backupFake) BackupCreate(name string) error { f.created = name; return nil }

func TestBackupRootCommandPrintsBackupStatus(t *testing.T) {
	service := &backupFake{backupMsg: "backup status"}
	cmd := NewCommand(service)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("backup command returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "backup status") {
		t.Fatalf("backup output = %q; want backup status", buf.String())
	}
}

func TestBackupListCommandShowsBackups(t *testing.T) {
	service := &backupFake{listed: []cli.BackupSummary{{Name: "alpha", CreatedAt: time.Now()}}}
	cmd := newListCommand(service)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("backup list returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "alpha") {
		t.Fatalf("backup list output = %q; want backup name", buf.String())
	}
}

func TestBackupApplyCommandCallsService(t *testing.T) {
	service := &backupFake{}
	cmd := newApplyCommand(service)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"alpha"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("backup apply returned error: %v", err)
	}
	if service.applied != "alpha" {
		t.Fatalf("backup apply did not pass alpha to the service; got %q", service.applied)
	}
	if !strings.Contains(buf.String(), "restored alpha") {
		t.Fatalf("backup apply output = %q; want restored alpha message", buf.String())
	}
}

func TestBackupCreateCommandCallsService(t *testing.T) {
	service := &backupFake{}
	cmd := newCreateCommand(service)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"alpha"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("backup create returned error: %v", err)
	}
	if service.created != "alpha" {
		t.Fatalf("backup create did not pass alpha to the service; got %q", service.created)
	}
	if !strings.Contains(buf.String(), "created alpha") {
		t.Fatalf("backup create output = %q; want created alpha message", buf.String())
	}
}
