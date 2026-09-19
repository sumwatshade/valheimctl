package startcmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sumwatshade/valheimctl/internal/cli"
)

type startFake struct {
	started bool
}

func (f *startFake) Init() error { return nil }
func (f *startFake) Start() (string, error) {
	f.started = true
	return "server started", nil
}
func (f *startFake) Stop() (string, error)                                    { return "", nil }
func (f *startFake) Status() (string, error)                                  { return "", nil }
func (f *startFake) Register(serverName, serverDir, description string) error { return nil }
func (f *startFake) Backup() (string, error)                                  { return "", nil }
func (f *startFake) BackupList() ([]cli.BackupSummary, error)                 { return nil, nil }
func (f *startFake) BackupApply(name string) error                            { return nil }
func (f *startFake) BackupCreate(name string) error                           { return nil }

func TestStartCommandRunsStart(t *testing.T) {
	service := &startFake{}
	cmd := NewCommand(service)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("start command returned error: %v", err)
	}
	if !service.started {
		t.Fatal("start command did not call Start() on the service")
	}
	if !strings.Contains(buf.String(), "server started") {
		t.Fatalf("start output = %q; want server started message", buf.String())
	}
}
