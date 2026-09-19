package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCommandInitializesState(t *testing.T) {
	root := t.TempDir()
	app := newApp(root, "", "")
	app.runtimeOS = "linux"
	app.osReleasePath = filepath.Join(root, "os-release")
	if err := os.WriteFile(app.osReleasePath, []byte("NAME=\"Debian GNU/Linux\"\nID=debian\nID_LIKE=debian\n"), 0o644); err != nil {
		t.Fatalf("writing os-release: %v", err)
	}
	app.packageManagerPath = filepath.Join(root, "apt-get")
	app.steamcmdPath = filepath.Join(root, "steamcmd")
	if err := os.WriteFile(app.packageManagerPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("writing fake apt-get: %v", err)
	}
	if err := os.WriteFile(app.steamcmdPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("writing fake steamcmd: %v", err)
	}
	cmd := newRootCmd(app)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"init"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command returned error: %v", err)
	}

	var st state
	if err := readJSON(filepath.Join(root, ".valheimctl", "state.json"), &st); err != nil {
		t.Fatalf("reading init state: %v", err)
	}
	if !st.Initialized {
		t.Fatal("init command did not create initialized state")
	}
	if !strings.Contains(buf.String(), "initialized") {
		t.Fatalf("init output = %q; want initialized message", buf.String())
	}
}
