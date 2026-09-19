package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestStopCommandSetsStoppedState(t *testing.T) {
	root := t.TempDir()
	app := newApp(root, "", "")
	if err := os.MkdirAll(app.serverSvc.ConfigDir(), 0o755); err != nil {
		t.Fatalf("creating config dir: %v", err)
	}
	if err := writeJSON(app.serverSvc.StatePath(), state{Initialized: true, Running: true, ServerName: "example-server"}); err != nil {
		t.Fatalf("writing initial state: %v", err)
	}

	cmd := newRootCmd(app)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"stop"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("stop command returned error: %v", err)
	}

	var st state
	if err := readJSON(app.serverSvc.StatePath(), &st); err != nil {
		t.Fatalf("reading stop state: %v", err)
	}
	if st.Running {
		t.Fatal("stop command did not mark server as stopped")
	}
	if !strings.Contains(buf.String(), "server stopped") {
		t.Fatalf("stop output = %q; want server stopped message", buf.String())
	}
}
