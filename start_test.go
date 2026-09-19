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
	svc := app.serverSvc
	svc.RuntimeOS = "linux"
	svc.OSReleasePath = filepath.Join(root, "os-release")
	if err := os.WriteFile(svc.OSReleasePath, []byte("NAME=\"Debian GNU/Linux\"\nID=debian\nID_LIKE=debian\n"), 0o644); err != nil {
		t.Fatalf("writing os-release: %v", err)
	}
	svc.PackageManagerPath = filepath.Join(root, "apt-get")
	svc.SteamcmdPath = filepath.Join(root, "steamcmd")
	if err := os.WriteFile(svc.PackageManagerPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("writing fake apt-get: %v", err)
	}
	if err := os.WriteFile(svc.SteamcmdPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("writing fake steamcmd: %v", err)
	}
	if err := svc.Init(); err != nil {
		t.Fatalf("init returned error: %v", err)
	}
	if err := writeJSON(svc.MetadataPath(), state{Initialized: true, Running: false, ServerName: "example-server"}); err != nil {
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
	if err := readJSON(svc.StatePath(), &st); err != nil {
		t.Fatalf("reading start state: %v", err)
	}
	if !st.Running {
		t.Fatal("start command did not mark server as running")
	}
	if !strings.Contains(buf.String(), "server started") {
		t.Fatalf("start output = %q; want server started message", buf.String())
	}
}
