package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	serversvc "github.com/sumwatshade/valheimctl/internal/server"
)

func TestInitInstallsRequiredPackagesAndSteamCMDForDebian(t *testing.T) {
	root := t.TempDir()
	logPath := filepath.Join(root, "apt.log")
	aptPath := filepath.Join(root, "apt-get")
	script := "#!/bin/sh\n" +
		"printf '%s\\n' \"$*\" >> \"" + logPath + "\"\n" +
		"exit 0\n"
	if err := os.WriteFile(aptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("writing fake apt-get: %v", err)
	}

	osReleasePath := filepath.Join(root, "os-release")
	if err := os.WriteFile(osReleasePath, []byte("NAME=\"Debian GNU/Linux\"\nID=debian\nID_LIKE=debian\n"), 0o644); err != nil {
		t.Fatalf("writing os-release: %v", err)
	}

	steamcmdPath := filepath.Join(root, "steamcmd")
	if err := os.WriteFile(steamcmdPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("writing fake steamcmd: %v", err)
	}

	svc := serversvc.NewService(root, "", "")
	svc.RuntimeOS = "linux"
	svc.OSReleasePath = osReleasePath
	svc.PackageManagerPath = aptPath
	svc.SteamcmdPath = steamcmdPath

	if err := svc.Init(); err != nil {
		t.Fatalf("init returned error: %v", err)
	}

	log, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("expected apt install log: %v", err)
	}
	content := string(log)
	for _, want := range []string{"install", "libatomic1", "libpulse-dev", "libpulse0", "steamcmd"} {
		if !strings.Contains(content, want) {
			t.Fatalf("apt log %q missing %q", content, want)
		}
	}
}

func TestInitRejectsUnsupportedLinuxDistribution(t *testing.T) {
	root := t.TempDir()
	osReleasePath := filepath.Join(root, "os-release")
	if err := os.WriteFile(osReleasePath, []byte("NAME=\"Fedora\"\nID=fedora\nID_LIKE=fedora\n"), 0o644); err != nil {
		t.Fatalf("writing os-release: %v", err)
	}

	svc := serversvc.NewService(root, "", "")
	svc.RuntimeOS = "linux"
	svc.OSReleasePath = osReleasePath

	err := svc.Init()
	if err == nil {
		t.Fatal("expected unsupported distro error")
	}
	if !strings.Contains(err.Error(), "unsupported Linux distribution") {
		t.Fatalf("error = %q; want unsupported distro message", err)
	}
}
