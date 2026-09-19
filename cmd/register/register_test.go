package registercmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sumwatshade/valheimctl/internal/cli"
)

type registerFake struct {
	serverName string
	serverDir  string
	description string
}

func (f *registerFake) Init() error { return nil }
func (f *registerFake) Start() (string, error) { return "", nil }
func (f *registerFake) Stop() (string, error) { return "", nil }
func (f *registerFake) Status() (string, error) { return "", nil }
func (f *registerFake) Register(serverName, serverDir, description string) error {
	f.serverName = serverName
	f.serverDir = serverDir
	f.description = description
	return nil
}
func (f *registerFake) Backup() (string, error) { return "", nil }
func (f *registerFake) BackupList() ([]cli.BackupSummary, error) { return nil, nil }
func (f *registerFake) BackupApply(name string) error { return nil }
func (f *registerFake) BackupCreate(name string) error { return nil }

func TestRegisterCommandRunsRegister(t *testing.T) {
	service := &registerFake{}
	cmd := NewCommand(service)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"test-server", "/tmp/valheim", "Example Server"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("register command returned error: %v", err)
	}
	if service.serverName != "test-server" || service.serverDir != "/tmp/valheim" || service.description != "Example Server" {
		t.Fatalf("register command passed unexpected values: %+v", service)
	}
	if !strings.Contains(buf.String(), "server registered") {
		t.Fatalf("register output = %q; want server registered message", buf.String())
	}
}
