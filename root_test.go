package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandRegistersExpectedCommands(t *testing.T) {
	cmd := newRootCmd(newApp(t.TempDir(), "", ""))
	var names []string
	for _, sub := range cmd.Commands() {
		names = append(names, sub.Name())
	}
	for _, name := range []string{"init", "start", "stop", "status", "register", "backup"} {
		found := false
		for _, got := range names {
			if got == name {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("root command missing subcommand %q; got %v", name, names)
		}
	}
}

func TestStatusCommandOutput(t *testing.T) {
	root := t.TempDir()
	cmd := newRootCmd(newApp(root, "", ""))
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"status"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("status command returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "not initialized") {
		t.Fatalf("status command output = %q; want not initialized message", buf.String())
	}
}
