package statuscmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sumwatshade/valheimctl/internal/cli"
)

type statusFake struct{}

func (f *statusFake) Init() error { return nil }
func (f *statusFake) Start() (string, error) { return "", nil }
func (f *statusFake) Stop() (string, error) { return "", nil }
func (f *statusFake) Status() (string, error) { return "not initialized", nil }
func (f *statusFake) Register(serverName, serverDir, description string) error { return nil }
func (f *statusFake) Backup() (string, error) { return "", nil }
func (f *statusFake) BackupList() ([]cli.BackupSummary, error) { return nil, nil }
func (f *statusFake) BackupApply(name string) error { return nil }
func (f *statusFake) BackupCreate(name string) error { return nil }

func TestStatusCommandReportsStatus(t *testing.T) {
	service := &statusFake{}
	cmd := NewCommand(service)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("status command returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "not initialized") {
		t.Fatalf("status output = %q; want not initialized message", buf.String())
	}
}
