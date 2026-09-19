package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegisterCommandCreatesServiceAndEnablesStartup(t *testing.T) {
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
	cmd := newRootCmd(app)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"register", "test-server", filepath.Join(root, "valheim"), "Example Server"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("register command returned error: %v", err)
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
	if !strings.Contains(buf.String(), "server registered") {
		t.Fatalf("register output = %q; want server registered message", buf.String())
	}
}
