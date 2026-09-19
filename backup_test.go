package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestBackupCommandReportsDeferred(t *testing.T) {
	root := t.TempDir()
	cmd := newRootCmd(newApp(root, "", ""))
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"backup"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("backup command returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "deferred") {
		t.Fatalf("backup output = %q; want deferred message", buf.String())
	}
}
