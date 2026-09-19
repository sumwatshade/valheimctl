package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/sumwatshade/valheimctl/internal/cli"
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

func (a *app) Init() error {
	return a.init()
}

func (a *app) Start() (string, error) {
	return a.start()
}

func (a *app) Stop() (string, error) {
	return a.stop()
}

func (a *app) Status() (string, error) {
	return a.status()
}

func (a *app) Register(serverName, serverDir, description string) error {
	return a.register(serverName, serverDir, description)
}

func (a *app) Backup() (string, error) {
	return a.backup()
}

func (a *app) BackupList() ([]cli.BackupSummary, error) {
	list, err := a.listBackups()
	if err != nil {
		return nil, err
	}
	out := make([]cli.BackupSummary, 0, len(list))
	for _, item := range list {
		out = append(out, cli.BackupSummary{Name: item.Name, CreatedAt: item.CreatedAt})
	}
	return out, nil
}

func (a *app) BackupApply(name string) error {
	return a.applyBackup(name)
}

func (a *app) BackupCreate(name string) error {
	return a.createBackup(name)
}

type backupMetadata struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Worlds    []string  `json:"worlds,omitempty"`
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

func (a *app) backupStoreDir() string {
	return filepath.Join(a.rootDir, ".valheimctl", "backups")
}

func (a *app) primaryWorldName() string {
	if st, err := a.readMetadata(); err == nil && st.ServerName != "" {
		return st.ServerName
	}
	return "Dedicated"
}

func (a *app) readMetadata() (state, error) {
	var st state
	if _, err := os.Stat(a.metadataPath()); err != nil {
		return state{}, err
	}
	if err := readJSON(a.metadataPath(), &st); err != nil {
		return state{}, err
	}
	return st, nil
}

func (a *app) listBackups() ([]backupMetadata, error) {
	backupsDir := a.backupStoreDir()
	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	items := make([]backupMetadata, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		metaPath := filepath.Join(backupsDir, entry.Name(), "meta.json")
		meta := backupMetadata{Name: entry.Name()}
		if _, err := os.ReadFile(metaPath); err == nil {
			if err := readJSON(metaPath, &meta); err != nil {
				return nil, err
			}
		} else if !os.IsNotExist(err) {
			return nil, err
		}
		if meta.CreatedAt.IsZero() {
			if info, err := os.Stat(filepath.Join(backupsDir, entry.Name())); err == nil {
				meta.CreatedAt = info.ModTime().UTC()
			}
		}
		items = append(items, meta)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
	return items, nil
}

func (a *app) createBackup(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("backup name is required")
	}
	worldsDir := filepath.Join(a.rootDir, "worlds_local")
	if _, err := os.Stat(worldsDir); err != nil {
		return fmt.Errorf("save directory %q does not exist: %w", worldsDir, err)
	}

	backupDir := filepath.Join(a.backupStoreDir(), name)
	if _, err := os.Stat(backupDir); err == nil {
		return fmt.Errorf("backup %q already exists", name)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return err
	}

	worlds := make([]string, 0)
	entries, err := os.ReadDir(worldsDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		worldName := entry.Name()
		worlds = append(worlds, worldName)
		src := filepath.Join(worldsDir, worldName)
		dst := filepath.Join(backupDir, worldName)
		if err := copyTree(src, dst); err != nil {
			return fmt.Errorf("copy world %q to backup: %w", worldName, err)
		}
	}
	if len(worlds) == 0 {
		return fmt.Errorf("no world folders found in %q", worldsDir)
	}

	meta := backupMetadata{
		Name:      name,
		CreatedAt: time.Now().UTC(),
		Worlds:    worlds,
	}
	if err := writeJSON(filepath.Join(backupDir, "meta.json"), meta); err != nil {
		return err
	}
	return nil
}

func (a *app) applyBackup(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("backup name is required")
	}
	backupDir := filepath.Join(a.backupStoreDir(), name)
	if _, err := os.Stat(backupDir); err != nil {
		return fmt.Errorf("backup %q not found", name)
	}

	worldsDir := filepath.Join(a.rootDir, "worlds_local")
	if err := os.MkdirAll(worldsDir, 0o755); err != nil {
		return err
	}

	worlds := make([]string, 0)
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "." {
			worlds = append(worlds, entry.Name())
		}
	}
	if len(worlds) == 0 {
		return fmt.Errorf("backup %q does not contain any world data", name)
	}

	primary := a.primaryWorldName()
	targetWorld := primary
	if _, err := os.Stat(filepath.Join(worldsDir, targetWorld)); err != nil {
		targetWorld = worlds[0]
	}

	src := filepath.Join(backupDir, targetWorld)
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("backup %q does not contain world %q", name, targetWorld)
	}
	if err := copyTree(src, filepath.Join(worldsDir, targetWorld)); err != nil {
		return err
	}
	return nil
}

func (a *app) backup() (string, error) {
	items, err := a.listBackups()
	if err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "No backups found.", nil
	}
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, fmt.Sprintf("%s\t%s", item.Name, item.CreatedAt.Format(time.RFC3339)))
	}
	return strings.Join(lines, "\n"), nil
}

func copyTree(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}
