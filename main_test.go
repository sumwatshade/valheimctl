package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatusReportsNotInitialized(t *testing.T) {
	root := t.TempDir()
	app := newApp(root, "", "")

	got, err := app.status()
	if err != nil {
		t.Fatalf("status returned error: %v", err)
	}
	if !strings.Contains(got, "not initialized") {
		t.Fatalf("status output = %q; want not initialized message", got)
	}
}

func TestRegisterCreatesServiceAndEnablesStartup(t *testing.T) {
	root := t.TempDir()
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}

	logPath := filepath.Join(root, "systemctl.log")
	script := "#!/bin/sh\n" +
		"printf '%s\\n' \"$*\" >> \"" + logPath + "\"\n" +
		"exit 0\n"
	systemctlPath := filepath.Join(binDir, "systemctl")
	if err := os.WriteFile(systemctlPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	app := newApp(root, systemctlPath, filepath.Join(root, "etc", "systemd", "system"))
	if err := app.init(); err != nil {
		t.Fatalf("init returned error: %v", err)
	}

	if err := app.register("test-server", filepath.Join(root, "valheim"), "Example Server"); err != nil {
		t.Fatalf("register returned error: %v", err)
	}

	servicePath := filepath.Join(root, "etc", "systemd", "system", "test-server.service")
	content, err := os.ReadFile(servicePath)
	if err != nil {
		t.Fatalf("expected service file: %v", err)
	}
	if !strings.Contains(string(content), "Description=Example Server") {
		t.Fatalf("service file missing description: %q", string(content))
	}

	log, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("systemctl log missing: %v", err)
	}
	if !strings.Contains(string(log), "enable") || !strings.Contains(string(log), "test-server.service") {
		t.Fatalf("expected enable invocation logged, got %q", string(log))
	}
}

func TestBackupIsDeferred(t *testing.T) {
	root := t.TempDir()
	app := newApp(root, "", "")
	out, err := app.backup()
	if err != nil {
		t.Fatalf("backup returned error: %v", err)
	}
	if !strings.Contains(out, "deferred") && !strings.Contains(out, "not implemented") {
		t.Fatalf("backup output = %q; want deferred/not implemented message", out)
	}
}
