package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCommandInitializesState(t *testing.T) {
	root := t.TempDir()
	cmd := newRootCmd(newApp(root, "", ""))
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
