package stopcmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sumwatshade/valheimctl/internal/cli"
)

type stopFake struct {
	stopped bool
}

func (f *stopFake) Init() error { return nil }
func (f *stopFake) Start() (string, error) { return "", nil }
func (f *stopFake) Stop() (string, error) {
	f.stopped = true
	return "server stopped", nil
}
func (f *stopFake) Status() (string, error) { return "", nil }
func (f *stopFake) Register(serverName, serverDir, description string) error { return nil }
func (f *stopFake) Backup() (string, error) { return "", nil }
func (f *stopFake) BackupList() ([]cli.BackupSummary, error) { return nil, nil }
func (f *stopFake) BackupApply(name string) error { return nil }
func (f *stopFake) BackupCreate(name string) error { return nil }

func TestStopCommandRunsStop(t *testing.T) {
	service := &stopFake{}
	cmd := NewCommand(service)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("stop command returned error: %v", err)
	}
	if !service.stopped {
		t.Fatal("stop command did not call Stop() on the service")
	}
	if !strings.Contains(buf.String(), "server stopped") {
		t.Fatalf("stop output = %q; want server stopped message", buf.String())
	}
}
