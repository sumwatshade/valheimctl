package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadFileParsesAndValidatesValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "valheim.yaml")
	content := `
name: "My server"
port: 2456
world: "Dedicated"
password: "Secret"
public: 1
savedir: "/tmp/valheim-saves"
logfile: "/tmp/valheim.log"
saveinterval: 1800
backups: 4
backupshort: 7200
backuplong: 43200
crossplay: false
instanceid: "1"
preset: hard
modifiers:
  raids: none
setkey: nomap
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	if cfg.Name != "My server" {
		t.Fatalf("Name = %q; want %q", cfg.Name, "My server")
	}
	if cfg.Port != 2456 {
		t.Fatalf("Port = %d; want 2456", cfg.Port)
	}
	if !cfg.Public {
		t.Fatal("Public should be true for explicit public: 1")
	}
	if cfg.Preset != "hard" {
		t.Fatalf("Preset = %q; want %q", cfg.Preset, "hard")
	}
	if got := cfg.Modifiers["raids"]; got != "none" {
		t.Fatalf("Modifier raids = %q; want %q", got, "none")
	}
	if got := cfg.Args(); !reflect.DeepEqual(got[:2], []string{"-name", "My server"}) {
		t.Fatalf("Args prefix = %v; want [-name My server]", got)
	}
}

func TestParseRejectsInvalidPortAndModifier(t *testing.T) {
	input := []byte(`
name: "Bad server"
port: 70000
password: "Secret"
modifiers:
  raids: impossible
`)

	_, err := Parse(input)
	if err == nil {
		t.Fatal("Parse() succeeded for invalid config; want error")
	}
	if !strings.Contains(err.Error(), "port") && !strings.Contains(err.Error(), "raids") {
		t.Fatalf("error = %q; want validation details for invalid port or modifier", err)
	}
}

func TestDefaultConfigFillsExpectedValheimDefaults(t *testing.T) {
	cfg, err := Parse([]byte("name: \"Test\"\npassword: \"pw\"\n"))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cfg.Port != 2456 {
		t.Fatalf("Port = %d; want default 2456", cfg.Port)
	}
	if cfg.World != "Dedicated" {
		t.Fatalf("World = %q; want %q", cfg.World, "Dedicated")
	}
	if !cfg.Public {
		t.Fatal("Public default should be true")
	}
	if cfg.SaveInterval != 1800 {
		t.Fatalf("SaveInterval = %d; want 1800", cfg.SaveInterval)
	}
	if cfg.BackupShort != 7200 {
		t.Fatalf("BackupShort = %d; want 7200", cfg.BackupShort)
	}
	if cfg.BackupLong != 43200 {
		t.Fatalf("BackupLong = %d; want 43200", cfg.BackupLong)
	}
}
