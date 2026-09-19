package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	backupsvc "github.com/sumwatshade/valheimctl/internal/backup"
)

func TestBackupCreateListAndApply(t *testing.T) {
	root := t.TempDir()
	app := newApp(root, "", "")
	svc := app.backupSvc
	worldDir := filepath.Join(root, "worlds_local", "Dedicated")
	if err := os.MkdirAll(filepath.Join(worldDir, "chunks"), 0o755); err != nil {
		t.Fatalf("mkdir world dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(worldDir, "level.dat"), []byte("live"), 0o600); err != nil {
		t.Fatalf("write level.dat: %v", err)
	}
	if err := os.WriteFile(filepath.Join(worldDir, "chunks", "a.chunk"), []byte("chunk-data"), 0o600); err != nil {
		t.Fatalf("write chunk: %v", err)
	}

	if err := svc.Create("alpha"); err != nil {
		t.Fatalf("backup create returned error: %v", err)
	}

	listCmd := newRootCmd(app)
	buf := &bytes.Buffer{}
	listCmd.SetOut(buf)
	listCmd.SetErr(buf)
	listCmd.SetArgs([]string{"backup", "list"})
	if err := listCmd.Execute(); err != nil {
		t.Fatalf("backup list returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "alpha") {
		t.Fatalf("backup list output = %q; want backup name", buf.String())
	}

	if err := os.WriteFile(filepath.Join(worldDir, "level.dat"), []byte("changed"), 0o600); err != nil {
		t.Fatalf("overwrite level.dat: %v", err)
	}
	if err := svc.Apply("alpha"); err != nil {
		t.Fatalf("backup apply returned error: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(worldDir, "level.dat"))
	if err != nil {
		t.Fatalf("read restored level.dat: %v", err)
	}
	if string(b) != "live" {
		t.Fatalf("restored world file = %q; want %q", string(b), "live")
	}
}

func TestBackupApplyRejectsMissingBackup(t *testing.T) {
	root := t.TempDir()
	svc := backupsvc.NewService(root)
	if err := svc.Apply("missing"); err == nil {
		t.Fatal("Apply() succeeded for missing backup; want error")
	}
}
