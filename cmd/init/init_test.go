package initcmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sumwatshade/valheimctl/internal/cli"
)

type initFake struct {
	initCalled bool
}

func (f *initFake) Init() error {
	f.initCalled = true
	return nil
}
func (f *initFake) Start() (string, error)                                   { return "", nil }
func (f *initFake) Stop() (string, error)                                    { return "", nil }
func (f *initFake) Status() (string, error)                                  { return "", nil }
func (f *initFake) Register(serverName, serverDir, description string) error { return nil }
func (f *initFake) Backup() (string, error)                                  { return "", nil }
func (f *initFake) BackupList() ([]cli.BackupSummary, error)                 { return nil, nil }
func (f *initFake) BackupApply(name string) error                            { return nil }
func (f *initFake) BackupCreate(name string) error                           { return nil }

func TestInitCommandInitializesState(t *testing.T) {
	service := &initFake{}
	cmd := NewCommand(service)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(nil)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}
	if !service.initCalled {
		t.Fatal("init command did not call Init() on the service")
	}
	if !strings.Contains(buf.String(), "initialized") {
		t.Fatalf("init output = %q; want initialized message", buf.String())
	}
}
