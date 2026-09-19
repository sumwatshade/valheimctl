package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStartCommandSetsRunningState(t *testing.T) {
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
	if err := app.init(); err != nil {
		t.Fatalf("init returned error: %v", err)
	}
	if err := writeJSON(app.metadataPath(), state{Initialized: true, Running: false, ServerName: "example-server"}); err != nil {
		t.Fatalf("writing metadata: %v", err)
	}

	cmd := newRootCmd(app)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"start"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("start command returned error: %v", err)
	}

	var st state
	if err := readJSON(app.statePath(), &st); err != nil {
		t.Fatalf("reading start state: %v", err)
	}
	if !st.Running {
		t.Fatal("start command did not mark server as running")
	}
	if !strings.Contains(buf.String(), "server started") {
		t.Fatalf("start output = %q; want server started message", buf.String())
	}
}
