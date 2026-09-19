package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestStartCommandSetsRunningState(t *testing.T) {
	root := t.TempDir()
	app := newApp(root, "", "")
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
