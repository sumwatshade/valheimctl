package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type app struct {
	rootDir            string
	systemctlPath      string
	serviceDir         string
	runtimeOS          string
	osReleasePath      string
	packageManagerPath string
	steamcmdPath       string
}

type state struct {
	Initialized bool   `json:"initialized"`
	Running     bool   `json:"running"`
	ServerName  string `json:"server_name,omitempty"`
	ServerDir   string `json:"server_dir,omitempty"`
	Description string `json:"description,omitempty"`
}

func newApp(rootDir, systemctlPath, serviceDir string) *app {
	if rootDir == "" {
		rootDir = "."
	}
	if systemctlPath == "" {
		systemctlPath = "systemctl"
	}
	if serviceDir == "" {
		serviceDir = filepath.Join(rootDir, "etc", "systemd", "system")
	}
	return &app{
		rootDir:            rootDir,
		systemctlPath:      systemctlPath,
		serviceDir:         serviceDir,
		runtimeOS:          runtime.GOOS,
		osReleasePath:      "/etc/os-release",
		packageManagerPath: "apt-get",
		steamcmdPath:       "steamcmd",
	}
}

func (a *app) configDir() string {
	return filepath.Join(a.rootDir, ".valheimctl")
}

func (a *app) statePath() string {
	return filepath.Join(a.configDir(), "state.json")
}

func (a *app) metadataPath() string {
	return filepath.Join(a.configDir(), "server.json")
}

func (a *app) ensureInitialized() error {
	if err := os.MkdirAll(a.configDir(), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(a.serviceDir, 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(a.statePath()); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	st := state{Initialized: true, Running: false}
	return writeJSON(a.statePath(), st)
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func (a *app) init() error {
	if err := a.ensureInitialized(); err != nil {
		return err
	}
	return a.ensurePrereqs()
}

func (a *app) ensurePrereqs() error {
	osName := a.runtimeOS
	if osName == "" {
		osName = runtime.GOOS
	}
	if osName != "linux" {
		return fmt.Errorf("unsupported operating system %q: valheimctl currently supports Linux with systemd", osName)
	}

	distro, err := a.detectLinuxDistro()
	if err != nil {
		return err
	}

	switch distro {
	case "debian", "ubuntu":
		return a.ensureDebianPrereqs()
	default:
		return fmt.Errorf("unsupported Linux distribution: %s", distro)
	}
}

func (a *app) detectLinuxDistro() (string, error) {
	b, err := os.ReadFile(a.osReleasePath)
	if err != nil {
		return "", fmt.Errorf("unable to read %s: %w", a.osReleasePath, err)
	}

	id := ""
	idLike := ""
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "ID="):
			id = strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
		case strings.HasPrefix(line, "ID_LIKE="):
			idLike = strings.Trim(strings.TrimPrefix(line, "ID_LIKE="), "\"")
		}
	}

	if id == "" && idLike == "" {
		return "", fmt.Errorf("unsupported Linux distribution: unable to determine OS from %s", a.osReleasePath)
	}

	for _, candidate := range []string{id, idLike} {
		for _, v := range strings.FieldsFunc(candidate, func(r rune) bool { return r == ' ' || r == '\t' || r == '|' }) {
			switch strings.TrimSpace(v) {
			case "debian", "ubuntu", "linuxmint":
				return "debian", nil
			}
		}
	}

	if id != "" {
		return strings.TrimSpace(id), nil
	}
	return strings.TrimSpace(idLike), nil
}

func (a *app) ensureDebianPrereqs() error {
	packages := []string{"libatomic1", "libpulse-dev", "libpulse0", "steamcmd"}
	if err := a.installPackages(packages...); err != nil {
		return err
	}
	if err := a.ensureSteamcmd(); err != nil {
		return err
	}
	return nil
}

func (a *app) installPackages(packages ...string) error {
	if len(packages) == 0 {
		return nil
	}
	cmd := exec.Command(a.packageManagerPath, append([]string{"install", "-y"}, packages...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("package installation failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (a *app) ensureSteamcmd() error {
	if _, err := os.Stat(a.steamcmdPath); err == nil {
		return nil
	}
	if _, err := exec.LookPath("steamcmd"); err == nil {
		return nil
	}
	return errors.New("steamcmd is not installed or not available on PATH")
}

func (a *app) status() (string, error) {
	if _, err := os.Stat(a.statePath()); err != nil {
		return "valheimctl status: not initialized", nil
	}
	var st state
	if err := readJSON(a.statePath(), &st); err != nil {
		return "", err
	}
	if !st.Initialized {
		return "valheimctl status: not initialized", nil
	}
	statusText := "stopped"
	if st.Running {
		statusText = "running"
	}
	return fmt.Sprintf("valheimctl status: %s (%s)", statusText, st.ServerName), nil
}

func (a *app) start() (string, error) {
	if err := a.ensureInitialized(); err != nil {
		return "", err
	}
	st := state{Initialized: true, Running: true}
	if _, err := os.Stat(a.metadataPath()); err == nil {
		var meta state
		if err := readJSON(a.metadataPath(), &meta); err == nil {
			st.ServerName = meta.ServerName
			st.ServerDir = meta.ServerDir
			st.Description = meta.Description
		}
	}
	if err := writeJSON(a.statePath(), st); err != nil {
		return "", err
	}
	return "valheimctl start: server started", nil
}

func (a *app) stop() (string, error) {
	if _, err := os.Stat(a.statePath()); err != nil {
		return "", errors.New("server not initialized")
	}
	st := state{Initialized: true, Running: false}
	if err := writeJSON(a.statePath(), st); err != nil {
		return "", err
	}
	return "valheimctl stop: server stopped", nil
}

func (a *app) register(serverName, serverDir, description string) error {
	if err := a.ensureInitialized(); err != nil {
		return err
	}
	if serverName == "" {
		return errors.New("server name is required")
	}
	if serverDir == "" {
		serverDir = a.rootDir
	}
	if description == "" {
		description = serverName
	}

	meta := state{
		Initialized: true,
		Running:     false,
		ServerName:  serverName,
		ServerDir:   serverDir,
		Description: description,
	}
	if err := writeJSON(a.metadataPath(), meta); err != nil {
		return err
	}

	serviceFile := filepath.Join(a.serviceDir, serverName+".service")
	serviceText := fmt.Sprintf("[Unit]\nDescription=%s\nAfter=network.target\n\n[Service]\nType=simple\nWorkingDirectory=%s\nExecStart=/bin/sh -c 'echo valheimctl placeholder'\nRestart=on-failure\n\n[Install]\nWantedBy=multi-user.target\n", description, serverDir)
	if err := os.WriteFile(serviceFile, []byte(serviceText), 0o644); err != nil {
		return err
	}

	if err := a.runSystemctl("daemon-reload"); err != nil {
		return err
	}
	if err := a.runSystemctl("enable", serverName+".service"); err != nil {
		return err
	}

	st := state{Initialized: true, Running: false, ServerName: serverName, ServerDir: serverDir, Description: description}
	return writeJSON(a.statePath(), st)
}

func (a *app) runSystemctl(args ...string) error {
	cmd := exec.Command(a.systemctlPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl %s failed: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}
