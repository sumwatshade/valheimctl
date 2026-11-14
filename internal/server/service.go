package server

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/sumwatshade/valheimctl/internal/cli"
	"github.com/sumwatshade/valheimctl/internal/config"
)

const (
	defaultSteamAppID   = 896660
	defaultWorldName    = "val-one-dot-oh"
	defaultServerName   = "My Valheim Server"
	defaultGamePort     = 2456
	defaultServiceUser  = "valheim"
	defaultServiceHome  = "/home/valheim"
	defaultServerDir    = "/home/valheim/server"
	defaultSaveDir      = "/home/valheim/worlds"
	defaultBackupDir    = "/home/valheim-backups"
	defaultStartScript  = "/home/valheim/server/start_valheim.sh"
	defaultServiceName  = "valheim"
	defaultServiceDir   = "/etc/systemd/system"
	defaultJournalctl   = "journalctl"
	defaultTar          = "tar"
	defaultFind         = "find"
	defaultDf           = "df"
	defaultDu           = "du"
	defaultSudo         = "sudo"
)

// Service manages the full Valheim dedicated server lifecycle.
type Service struct {
	RootDir            string
	SystemctlPath      string
	ServiceDir         string
	RuntimeOS          string
	OSReleasePath      string
	PackageManagerPath string
	SteamcmdPath       string

	// Configurable deployment paths (defaults match the spec).
	ServiceUser  string
	ServiceHome  string
	ServerDir    string
	SaveDir      string
	BackupDir    string
	StartScript  string
	ServiceName  string
	SteamAppID   int
	WorldName    string
	ServerName   string
	GamePort     int
	Journalctl   string
	Tar          string
	Find         string
	Df           string
	Du           string
	Sudo         string
}

func NewService(rootDir, systemctlPath, serviceDir string) *Service {
	if rootDir == "" {
		rootDir = "."
	}
	if systemctlPath == "" {
		systemctlPath = "systemctl"
	}
	if serviceDir == "" {
		serviceDir = defaultServiceDir
	}
	if !filepath.IsAbs(serviceDir) {
		serviceDir = filepath.Join("/", serviceDir)
	}
	return &Service{
		RootDir:            rootDir,
		SystemctlPath:      systemctlPath,
		ServiceDir:         serviceDir,
		RuntimeOS:          runtime.GOOS,
		OSReleasePath:      "/etc/os-release",
		PackageManagerPath: "apt-get",
		SteamcmdPath:       "steamcmd",

		ServiceUser:  defaultServiceUser,
		ServiceHome:  defaultServiceHome,
		ServerDir:    defaultServerDir,
		SaveDir:      defaultSaveDir,
		BackupDir:    defaultBackupDir,
		StartScript:  defaultStartScript,
		ServiceName:  defaultServiceName,
		SteamAppID:   defaultSteamAppID,
		WorldName:    defaultWorldName,
		ServerName:   defaultServerName,
		GamePort:     defaultGamePort,
		Journalctl:   defaultJournalctl,
		Tar:          defaultTar,
		Find:         defaultFind,
		Df:           defaultDf,
		Du:           defaultDu,
		Sudo:         defaultSudo,
	}
}

func (s *Service) ConfigDir() string { return config.RootDir(s.RootDir) }
func (s *Service) StatePath() string { return config.StatePath(s.RootDir) }
func (s *Service) MetadataPath() string { return config.MetadataPath(s.RootDir) }

func (s *Service) Init() error {
	if err := config.EnsureInitialized(s.RootDir, ""); err != nil {
		return err
	}
	return s.ensurePrereqs()
}

func (s *Service) ensurePrereqs() error {
	osName := s.RuntimeOS
	if osName == "" {
		osName = runtime.GOOS
	}
	if osName != "linux" {
		return fmt.Errorf("unsupported operating system %q: valheimctl currently supports Linux with systemd", osName)
	}

	distro, err := s.detectLinuxDistro()
	if err != nil {
		return err
	}

	switch distro {
	case "debian", "ubuntu":
		return s.ensureDebianPrereqs()
	default:
		return fmt.Errorf("unsupported Linux distribution: %s", distro)
	}
}

func (s *Service) detectLinuxDistro() (string, error) {
	b, err := os.ReadFile(s.OSReleasePath)
	if err != nil {
		return "", fmt.Errorf("unable to read %s: %w", s.OSReleasePath, err)
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
		return "", fmt.Errorf("unsupported Linux distribution: unable to determine OS from %s", s.OSReleasePath)
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

func (s *Service) ensureDebianPrereqs() error {
	packages := []string{"libatomic1", "libpulse-dev", "libpulse0"}
	if err := s.installPackages(packages...); err != nil {
		return err
	}
	if err := s.ensureSteamcmd(); err != nil {
		return err
	}
	return nil
}

func (s *Service) installPackages(packages ...string) error {
	if len(packages) == 0 {
		return nil
	}
	cmd := exec.Command(s.PackageManagerPath, append([]string{"install", "-y"}, packages...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("package installation failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (s *Service) ensureSteamcmd() error {
	if _, err := os.Stat(s.SteamcmdPath); err == nil {
		return nil
	}
	if _, err := exec.LookPath("steamcmd"); err == nil {
		return nil
	}
	return errors.New("steamcmd is not installed or not available on PATH")
}

func (s *Service) Status() (string, error) {
	if _, err := os.Stat(s.StatePath()); err != nil {
		return "valheimctl status: not initialized", nil
	}
	var st config.State
	if err := config.ReadJSON(s.StatePath(), &st); err != nil {
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

func (s *Service) Start() (string, error) {
	if err := config.EnsureInitialized(s.RootDir, s.ServiceDir); err != nil {
		return "", err
	}
	st := config.State{Initialized: true, Running: true}
	if _, err := os.Stat(s.MetadataPath()); err == nil {
		var meta config.State
		if err := config.ReadJSON(s.MetadataPath(), &meta); err == nil {
			st.ServerName = meta.ServerName
			st.ServerDir = meta.ServerDir
			st.Description = meta.Description
		}
	}
	if err := config.WriteJSON(s.StatePath(), st); err != nil {
		return "", err
	}
	return "valheimctl start: server started", nil
}

func (s *Service) Stop() (string, error) {
	if _, err := os.Stat(s.StatePath()); err != nil {
		return "", errors.New("server not initialized")
	}
	st := config.State{Initialized: true, Running: false}
	if err := config.WriteJSON(s.StatePath(), st); err != nil {
		return "", err
	}
	return "valheimctl stop: server stopped", nil
}

func (s *Service) Register(serverName, serverDir, description string) error {
	if err := config.EnsureInitialized(s.RootDir, s.ServiceDir); err != nil {
		return err
	}
	if serverName == "" {
		return errors.New("server name is required")
	}
	if serverDir == "" {
		serverDir = s.RootDir
	}
	if description == "" {
		description = serverName
	}

	meta := config.State{
		Initialized: true,
		Running:     false,
		ServerName:  serverName,
		ServerDir:   serverDir,
		Description: description,
	}
	if err := config.WriteJSON(s.MetadataPath(), meta); err != nil {
		return err
	}

	serviceFile := filepath.Join(s.ServiceDir, serverName+".service")
	serviceText := fmt.Sprintf("[Unit]\nDescription=%s\nAfter=network.target\n\n[Service]\nType=simple\nWorkingDirectory=%s\nExecStart=/bin/sh -c 'echo valheimctl placeholder'\nRestart=on-failure\n\n[Install]\nWantedBy=multi-user.target\n", description, serverDir)
	if err := os.WriteFile(serviceFile, []byte(serviceText), 0o644); err != nil {
		return err
	}

	if err := s.runSystemctl("daemon-reload"); err != nil {
		return err
	}
	if err := s.runSystemctl("enable", serverName+".service"); err != nil {
		return err
	}

	st := config.State{Initialized: true, Running: false, ServerName: serverName, ServerDir: serverDir, Description: description}
	return config.WriteJSON(s.StatePath(), st)
}

func (s *Service) runSystemctl(args ...string) error {
	cmd := exec.Command(s.SystemctlPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl %s failed: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

// AsServiceUser runs a command as the configured service user via sudo -iu.
func (s *Service) AsServiceUser(args ...string) *exec.Cmd {
	allArgs := append([]string{"-iu", s.ServiceUser}, args...)
	return exec.Command(s.Sudo, allArgs...)
}

// runAsServiceUser runs a command as the service user and returns combined output.
func (s *Service) runAsServiceUser(args ...string) (string, error) {
	cmd := s.AsServiceUser(args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// Install installs the Valheim dedicated server binaries via SteamCMD.
func (s *Service) Install() (string, error) {
	if err := s.ensureSteamcmd(); err != nil {
		return "", err
	}

	// Initialize state directory.
	if err := config.EnsureInitialized(s.RootDir, s.ServiceDir); err != nil {
		return "", fmt.Errorf("initialize state: %w", err)
	}

	// Create directories.
	if err := os.MkdirAll(s.ServerDir, 0o755); err != nil {
		return "", fmt.Errorf("create server dir: %w", err)
	}
	if err := os.MkdirAll(s.SaveDir, 0o755); err != nil {
		return "", fmt.Errorf("create save dir: %w", err)
	}
	if err := os.MkdirAll(s.BackupDir, 0o755); err != nil {
		return "", fmt.Errorf("create backup dir: %w", err)
	}

	// Fix ownership.
	if err := s.chownRecursive(s.ServiceHome, s.ServiceUser); err != nil {
		return "", fmt.Errorf("fix ownership: %w", err)
	}

	// Install server binaries via SteamCMD as the service user.
	out, err := s.runAsServiceUser(
		"steamcmd",
		"+force_install_dir", s.ServerDir,
		"+login", "anonymous",
		"+app_update", fmt.Sprintf("%d", s.SteamAppID), "validate",
		"+quit",
	)
	if err != nil {
		return "", fmt.Errorf("steamcmd install failed: %w\n%s", err, out)
	}

	// Create the startup script.
	startScript := s.generateStartScript()
	if err := os.WriteFile(s.StartScript, []byte(startScript), 0o750); err != nil {
		return "", fmt.Errorf("write start script: %w", err)
	}
	if err := os.Chown(s.StartScript, 0, 0); err != nil {
		// Best effort; may fail in non-root environments.
		_ = os.Chown(s.StartScript, 0, 0)
	}

	// Create the systemd unit file.
	unitFile := filepath.Join(s.ServiceDir, s.ServiceName+".service")
	unitText := s.generateUnitFile()
	if err := os.WriteFile(unitFile, []byte(unitText), 0o644); err != nil {
		return "", fmt.Errorf("write systemd unit: %w", err)
	}

	// Try to reload systemd and enable the service, but don't fail if
	// systemctl is unavailable (e.g. inside a container without systemd).
	_ = s.runSystemctl("daemon-reload")
	_ = s.runSystemctl("enable", s.ServiceName+".service")

	// Update state.
	st := config.State{
		Initialized: true,
		Running:     false,
		ServerName:  s.ServerName,
		ServerDir:   s.ServerDir,
		Description: s.ServerName,
	}
	if err := config.WriteJSON(s.StatePath(), st); err != nil {
		return "", err
	}

	var buf bytes.Buffer
	buf.WriteString("valheimctl install: server installed\n")
	buf.WriteString(fmt.Sprintf("  server dir: %s\n", s.ServerDir))
	buf.WriteString(fmt.Sprintf("  save dir: %s\n", s.SaveDir))
	buf.WriteString(fmt.Sprintf("  backup dir: %s\n", s.BackupDir))
	buf.WriteString(fmt.Sprintf("  start script: %s\n", s.StartScript))
	buf.WriteString(fmt.Sprintf("  systemd unit: %s\n", unitFile))
	buf.WriteString(fmt.Sprintf("  steam app ID: %d\n", s.SteamAppID))
	buf.WriteString("Run 'valheimctl start' to start the server.\n")
	return buf.String(), nil
}

// generateStartScript creates the server startup script content.
func (s *Service) generateStartScript() string {
	args := []string{
		"-name", s.ServerName,
		"-port", fmt.Sprintf("%d", s.GamePort),
		"-world", s.WorldName,
		"-savedir", s.SaveDir,
		"-public", "1",
		"-saveinterval", "900",
		"-backups", "4",
		"-backupshort", "7200",
		"-backuplong", "43200",
	}
	var b bytes.Buffer
	b.WriteString("#!/usr/bin/env bash\n")
	b.WriteString("set -e\n")
	b.WriteString(fmt.Sprintf("cd %s\n", s.ServerDir))
	b.WriteString("exec ./valheim_server.x86_64 \\\n")
	for i, arg := range args {
		if i == len(args)-1 {
			b.WriteString(fmt.Sprintf("  %s\n", arg))
		} else {
			b.WriteString(fmt.Sprintf("  %s \\\n", arg))
		}
	}
	return b.String()
}

// generateUnitFile creates the systemd unit file content.
func (s *Service) generateUnitFile() string {
	return fmt.Sprintf(`[Unit]
Description=Valheim Dedicated Server
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
User=%s
Group=%s
WorkingDirectory=%s
ExecStart=%s
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
`, s.ServiceUser, s.ServiceUser, s.ServerDir, s.StartScript)
}

// Update updates the Valheim dedicated server binaries via SteamCMD.
// It backs up the save root before updating. If start is true, the service
// is started after a successful update.
func (s *Service) Update(validate bool, start bool) (string, error) {
	// Check if service is active.
	active, err := s.isServiceActive()
	if err != nil {
		return "", err
	}
	if active {
		return "", errors.New("service is running; stop it before updating, or use --start to restart after update")
	}

	// Back up the save root before updating.
	backupLabel := fmt.Sprintf("pre-update-%s", time.Now().UTC().Format("20060102-150405"))
	if _, err := s.WorldBackup(backupLabel); err != nil {
		return "", fmt.Errorf("pre-update backup failed: %w", err)
	}

	// Run SteamCMD update as the service user.
	appUpdate := fmt.Sprintf("%d", s.SteamAppID)
	if validate {
		appUpdate += " validate"
	}
	out, err := s.runAsServiceUser(
		"steamcmd",
		"+force_install_dir", s.ServerDir,
		"+login", "anonymous",
		"+app_update", appUpdate,
		"+quit",
	)
	if err != nil {
		return "", fmt.Errorf("steamcmd update failed: %w\n%s", err, out)
	}

	// Fix ownership on server dir.
	if err := s.chownRecursive(s.ServerDir, s.ServiceUser); err != nil {
		return "", fmt.Errorf("fix ownership: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("valheimctl update: server updated\n")
	buf.WriteString(fmt.Sprintf("  backup: %s\n", backupLabel))
	buf.WriteString(fmt.Sprintf("  steam app ID: %d\n", s.SteamAppID))
	buf.WriteString(out)
	buf.WriteString("\n")
	if start {
		msg, err := s.Start()
		if err != nil {
			return buf.String(), err
		}
		buf.WriteString(msg)
		buf.WriteString("\n")
	} else {
		buf.WriteString("Run 'valheimctl start' to start the server.\n")
	}
	return buf.String(), nil
}

// Restart restarts the systemd service.
func (s *Service) Restart() (string, error) {
	if err := s.runSystemctl("restart", s.ServiceName+".service"); err != nil {
		return "", err
	}
	st := config.State{Initialized: true, Running: true}
	if err := config.WriteJSON(s.StatePath(), st); err != nil {
		return "", err
	}
	return "valheimctl restart: server restarted", nil
}

// Logs returns recent journalctl logs for the service.
func (s *Service) Logs(follow bool, lines int) (string, error) {
	args := []string{"-u", s.ServiceName, "--no-pager"}
	if lines > 0 {
		args = append(args, "-n", fmt.Sprintf("%d", lines))
	}
	if follow {
		args = append(args, "-f")
	}
	cmd := exec.Command(s.Journalctl, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("journalctl failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

// Version reports version information about the server installation.
func (s *Service) Version() (string, error) {
	var buf bytes.Buffer

	// Check server binary.
	binaryPath := filepath.Join(s.ServerDir, "valheim_server.x86_64")
	if info, err := os.Stat(binaryPath); err == nil {
		buf.WriteString(fmt.Sprintf("server binary: %s\n", binaryPath))
		buf.WriteString(fmt.Sprintf("  size: %d bytes\n", info.Size()))
		buf.WriteString(fmt.Sprintf("  modified: %s\n", info.ModTime().UTC().Format(time.RFC3339)))
	} else {
		buf.WriteString(fmt.Sprintf("server binary not found: %s\n", binaryPath))
	}

	// Check SteamCMD.
	if _, err := exec.LookPath("steamcmd"); err == nil {
		out, err := exec.Command("steamcmd", "+quit").CombinedOutput()
		if err != nil {
			buf.WriteString(fmt.Sprintf("steamcmd: error (%v)\n", err))
		} else {
			buf.WriteString(fmt.Sprintf("steamcmd: available\n%s\n", strings.TrimSpace(string(out))))
		}
	} else {
		buf.WriteString("steamcmd: not found\n")
	}

	// Check systemd service.
	active, _ := s.isServiceActive()
	buf.WriteString(fmt.Sprintf("service active: %v\n", active))

	return buf.String(), nil
}

// ConfigShow displays the current server configuration.
func (s *Service) ConfigShow() (string, error) {
	var buf bytes.Buffer
	buf.WriteString("valheimctl config:\n")
	buf.WriteString(fmt.Sprintf("  service user: %s\n", s.ServiceUser))
	buf.WriteString(fmt.Sprintf("  service home: %s\n", s.ServiceHome))
	buf.WriteString(fmt.Sprintf("  server dir: %s\n", s.ServerDir))
	buf.WriteString(fmt.Sprintf("  save dir: %s\n", s.SaveDir))
	buf.WriteString(fmt.Sprintf("  backup dir: %s\n", s.BackupDir))
	buf.WriteString(fmt.Sprintf("  start script: %s\n", s.StartScript))
	buf.WriteString(fmt.Sprintf("  systemd unit: %s/%s.service\n", s.ServiceDir, s.ServiceName))
	buf.WriteString(fmt.Sprintf("  steam app ID: %d\n", s.SteamAppID))
	buf.WriteString(fmt.Sprintf("  world name: %s\n", s.WorldName))
	buf.WriteString(fmt.Sprintf("  server name: %s\n", s.ServerName))
	buf.WriteString(fmt.Sprintf("  game port: %d\n", s.GamePort))
	return buf.String(), nil
}

// ConfigValidate validates the current server configuration.
func (s *Service) ConfigValidate() (string, error) {
	var issues []string

	if s.ServiceUser == "" {
		issues = append(issues, "service user is empty")
	}
	if s.ServerDir == "" {
		issues = append(issues, "server dir is empty")
	}
	if s.SaveDir == "" {
		issues = append(issues, "save dir is empty")
	}
	if s.BackupDir == "" {
		issues = append(issues, "backup dir is empty")
	}
	if s.StartScript == "" {
		issues = append(issues, "start script is empty")
	}
	if s.SteamAppID <= 0 {
		issues = append(issues, "steam app ID must be greater than zero")
	}
	if s.WorldName == "" {
		issues = append(issues, "world name is empty")
	}
	if s.ServerName == "" {
		issues = append(issues, "server name is empty")
	}
	if s.GamePort <= 0 || s.GamePort > 65535 {
		issues = append(issues, "game port must be between 1 and 65535")
	}

	// Check paths exist.
	for _, path := range []string{s.ServerDir, s.SaveDir, s.BackupDir} {
		if path != "" {
			if _, err := os.Stat(path); err != nil {
				issues = append(issues, fmt.Sprintf("path does not exist: %s (%v)", path, err))
			}
		}
	}

	if len(issues) > 0 {
		var buf bytes.Buffer
		buf.WriteString("valheimctl config validate: invalid\n")
		for _, issue := range issues {
			buf.WriteString(fmt.Sprintf("  - %s\n", issue))
		}
		return buf.String(), errors.New("configuration validation failed")
	}
	return "valheimctl config validate: valid", nil
}

// WorldList lists save directories and backup candidates.
func (s *Service) WorldList() (string, error) {
	entries, err := os.ReadDir(s.SaveDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "No save directory found.", nil
		}
		return "", err
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("Save directory: %s\n", s.SaveDir))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		buf.WriteString(fmt.Sprintf("  %s  %s  %d bytes\n", info.ModTime().UTC().Format(time.RFC3339), entry.Name(), info.Size()))
	}
	return buf.String(), nil
}

// WorldInspect inspects a specific world directory.
func (s *Service) WorldInspect(name string) (string, error) {
	worldPath := filepath.Join(s.SaveDir, name)
	info, err := os.Stat(worldPath)
	if err != nil {
		return "", fmt.Errorf("world not found: %s", name)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("not a directory: %s", name)
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("World: %s\n", name))
	buf.WriteString(fmt.Sprintf("  path: %s\n", worldPath))
	buf.WriteString(fmt.Sprintf("  size: %d bytes\n", info.Size()))
	buf.WriteString(fmt.Sprintf("  modified: %s\n", info.ModTime().UTC().Format(time.RFC3339)))

	// List contents up to max depth 5.
	err = filepath.WalkDir(worldPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(worldPath, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		depth := strings.Count(rel, string(filepath.Separator))
		if depth > 5 {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		indent := strings.Repeat("  ", depth+1)
		if d.IsDir() {
			buf.WriteString(fmt.Sprintf("%s%s/\n", indent, d.Name()))
		} else {
			info, err := d.Info()
			if err != nil {
				return err
			}
			buf.WriteString(fmt.Sprintf("%s%s  %d bytes\n", indent, d.Name(), info.Size()))
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// WorldBackup creates a backup of the save root.
func (s *Service) WorldBackup(label string) (string, error) {	if label == "" {
		label = time.Now().UTC().Format("2006-01-02-150405")
	}

	if _, err := os.Stat(s.SaveDir); err != nil {
		return "", fmt.Errorf("save directory does not exist: %w", err)
	}

	archiveName := fmt.Sprintf("worlds-%s.tar.gz", label)
	archivePath := filepath.Join(s.BackupDir, archiveName)

	if _, err := os.Stat(archivePath); err == nil {
		return "", fmt.Errorf("backup already exists: %s", archivePath)
	}

	cmd := exec.Command(s.Tar, "-czvf", archivePath, "-C", filepath.Dir(s.SaveDir), filepath.Base(s.SaveDir))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("backup failed: %w: %s", err, strings.TrimSpace(string(out)))
	}

	// Fix ownership.
	if err := os.Chown(archivePath, 0, 0); err != nil {
		_ = os.Chown(archivePath, 0, 0)
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("valheimctl backup: created %s\n", archivePath))
	buf.Write(out)
	return buf.String(), nil
}

// WorldRestore restores a world from a backup archive.
// dryRun performs a validation-only check when true.
func (s *Service) WorldRestore(archive string, dryRun bool) (string, error) {
	if _, err := os.Stat(archive); err != nil {
		return "", fmt.Errorf("archive not found: %s", archive)
	}

	// Inspect archive paths first.
	listCmd := exec.Command(s.Tar, "-tzf", archive)
	listOut, err := listCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("cannot read archive: %w: %s", err, strings.TrimSpace(string(listOut)))
	}

	// Look for the save directory path in the archive.
	saveBase := filepath.Base(s.SaveDir)
	expectedPrefix := saveBase + "/"
	var foundPaths []string
	for _, line := range strings.Split(string(listOut), "\n") {
		if strings.HasPrefix(line, expectedPrefix) {
			foundPaths = append(foundPaths, line)
		}
	}

	if len(foundPaths) == 0 {
		return "", fmt.Errorf("archive does not contain expected path prefix %q", expectedPrefix)
	}

	if dryRun {
		var buf bytes.Buffer
		buf.WriteString("valheimctl restore: dry-run (no changes made)\n")
		buf.WriteString(fmt.Sprintf("  archive: %s\n", archive))
		buf.WriteString(fmt.Sprintf("  found %d path(s) under %q\n", len(foundPaths), expectedPrefix))
		for _, p := range foundPaths[:min(len(foundPaths), 10)] {
			buf.WriteString(fmt.Sprintf("    %s\n", p))
		}
		if len(foundPaths) > 10 {
			buf.WriteString(fmt.Sprintf("    ... and %d more\n", len(foundPaths)-10))
		}
		buf.WriteString("Run without --dry-run to perform the restore.\n")
		return buf.String(), nil
	}

	// Stop the service if active.
	active, err := s.isServiceActive()
	if err != nil {
		return "", err
	}
	if active {
		if err := s.runSystemctl("stop", s.ServiceName+".service"); err != nil {
			return "", err
		}
	}

	// Preserve current save directory.
	if _, err := os.Stat(s.SaveDir); err == nil {
		preserveName := fmt.Sprintf("worlds_local-before-restore-%s", time.Now().UTC().Format("20060102-150405"))
		preservePath := filepath.Join(filepath.Dir(s.SaveDir), preserveName)
		if err := os.Rename(s.SaveDir, preservePath); err != nil {
			return "", fmt.Errorf("preserve current save: %w", err)
		}
	}

	// Extract the save subtree.
	extractArgs := []string{"-xzf", archive, "-C", filepath.Dir(s.SaveDir)}
	for _, p := range foundPaths {
		// Only extract paths under the expected prefix.
		if strings.HasPrefix(p, expectedPrefix) {
			extractArgs = append(extractArgs, p)
		}
	}
	cmd := exec.Command(s.Tar, extractArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("restore extraction failed: %w: %s", err, strings.TrimSpace(string(out)))
	}

	// Fix ownership.
	if err := s.chownRecursive(s.SaveDir, s.ServiceUser); err != nil {
		return "", fmt.Errorf("fix ownership: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("valheimctl restore: restored %s\n", archive))
	buf.WriteString(fmt.Sprintf("  extracted %d path(s)\n", len(foundPaths)))
	buf.WriteString("Run 'valheimctl start' to start the server.\n")
	return buf.String(), nil
}

// WorldVerify verifies a backup archive's integrity.
func (s *Service) WorldVerify(archive string) (string, error) {
	if _, err := os.Stat(archive); err != nil {
		return "", fmt.Errorf("archive not found: %s", archive)
	}

	cmd := exec.Command(s.Tar, "-tzf", archive)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("archive integrity check failed: %w: %s", err, strings.TrimSpace(string(out)))
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("valheimctl verify: archive %s is valid\n", archive))
	return buf.String(), nil
}

// BackupCreate creates a backup of the save root (ServerManagement interface).
func (s *Service) BackupCreate(label string) (string, error) {
	return s.WorldBackup(label)
}

// BackupList lists existing backups (ServerManagement interface).
func (s *Service) BackupList() (string, error) {
	entries, err := os.ReadDir(s.BackupDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "No backups found.", nil
		}
		return "", err
	}

	var backups []fs.FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tar.gz") {
			backups = append(backups, info)
		}
	}
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].ModTime().After(backups[j].ModTime())
	})

	if len(backups) == 0 {
		return "No backups found.", nil
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("Backup directory: %s\n", s.BackupDir))
	for _, info := range backups {
		buf.WriteString(fmt.Sprintf("  %s  %s  %d bytes\n", info.ModTime().UTC().Format(time.RFC3339), info.Name(), info.Size()))
	}
	return buf.String(), nil
}

// BackupInspect inspects a backup archive.
func (s *Service) BackupInspect(archive string) (string, error) {
	if _, err := os.Stat(archive); err != nil {
		return "", fmt.Errorf("archive not found: %s", archive)
	}

	cmd := exec.Command(s.Tar, "-tzf", archive)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("cannot read archive: %w: %s", err, strings.TrimSpace(string(out)))
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("Archive: %s\n", archive))
	buf.Write(out)
	return buf.String(), nil
}

// ServerReinstall performs a clean reinstall preserving worlds.
// dryRun performs a validation-only check when true.
func (s *Service) ServerReinstall(dryRun bool) (string, error) {
	if dryRun {
		var buf bytes.Buffer
		buf.WriteString("valheimctl reinstall: dry-run (no changes made)\n")
		buf.WriteString(fmt.Sprintf("  would remove: %s\n", s.ServerDir))
		buf.WriteString(fmt.Sprintf("  would preserve: %s\n", s.SaveDir))
		buf.WriteString(fmt.Sprintf("  would preserve: %s\n", s.BackupDir))
		buf.WriteString(fmt.Sprintf("  would reinstall app ID: %d\n", s.SteamAppID))
		return buf.String(), nil
	}

	// Stop the service if active.
	active, err := s.isServiceActive()
	if err != nil {
		return "", err
	}
	if active {
		if err := s.runSystemctl("stop", s.ServiceName+".service"); err != nil {
			return "", err
		}
	}

	// Back up the save root first.
	if _, err := s.WorldBackup("pre-reinstall"); err != nil {
		return "", fmt.Errorf("pre-reinstall backup failed: %w", err)
	}

	// Back up startup script and unit file.
	if _, err := os.Stat(s.StartScript); err == nil {
		backupPath := filepath.Join(s.BackupDir, fmt.Sprintf("start_valheim.sh-%s", time.Now().UTC().Format("20060102-150405")))
		if err := copyFile(s.StartScript, backupPath); err != nil {
			return "", fmt.Errorf("backup start script: %w", err)
		}
	}
	unitFile := filepath.Join(s.ServiceDir, s.ServiceName+".service")
	if _, err := os.Stat(unitFile); err == nil {
		backupPath := filepath.Join(s.BackupDir, fmt.Sprintf("%s.service-%s", s.ServiceName, time.Now().UTC().Format("20060102-150405")))
		if err := copyFile(unitFile, backupPath); err != nil {
			return "", fmt.Errorf("backup unit file: %w", err)
		}
	}

	// Remove only the server installation directory.
	if err := os.RemoveAll(s.ServerDir); err != nil {
		return "", fmt.Errorf("remove server dir: %w", err)
	}
	if err := os.MkdirAll(s.ServerDir, 0o755); err != nil {
		return "", fmt.Errorf("create server dir: %w", err)
	}

	// Install via SteamCMD.
	out, err := s.runAsServiceUser(
		"steamcmd",
		"+force_install_dir", s.ServerDir,
		"+login", "anonymous",
		"+app_update", fmt.Sprintf("%d", s.SteamAppID), "validate",
		"+quit",
	)
	if err != nil {
		return "", fmt.Errorf("steamcmd install failed: %w\n%s", err, out)
	}

	// Recreate startup script and unit file.
	if err := os.WriteFile(s.StartScript, []byte(s.generateStartScript()), 0o750); err != nil {
		return "", fmt.Errorf("write start script: %w", err)
	}
	if err := os.WriteFile(unitFile, []byte(s.generateUnitFile()), 0o644); err != nil {
		return "", fmt.Errorf("write systemd unit: %w", err)
	}

	if err := s.runSystemctl("daemon-reload"); err != nil {
		return "", err
	}

	// Fix ownership.
	if err := s.chownRecursive(s.ServerDir, s.ServiceUser); err != nil {
		return "", fmt.Errorf("fix ownership: %w", err)
	}
	if err := s.chownRecursive(s.SaveDir, s.ServiceUser); err != nil {
		return "", fmt.Errorf("fix ownership: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("valheimctl reinstall: server reinstalled\n")
	buf.WriteString(fmt.Sprintf("  server dir: %s\n", s.ServerDir))
	buf.WriteString(fmt.Sprintf("  save dir preserved: %s\n", s.SaveDir))
	buf.WriteString(out)
	buf.WriteString("Run 'valheimctl start' to start the server.\n")
	return buf.String(), nil
}

// Diagnostics collects diagnostic information about the server setup.
func (s *Service) Diagnostics() (string, error) {
	var buf bytes.Buffer
	buf.WriteString("valheimctl diagnostics\n")
	buf.WriteString("====================\n\n")

	// Service status.
	buf.WriteString("Service status:\n")
	active, err := s.isServiceActive()
	if err != nil {
		buf.WriteString(fmt.Sprintf("  error checking status: %v\n", err))
	} else {
		buf.WriteString(fmt.Sprintf("  active: %v\n", active))
	}
	enabled, err := s.isServiceEnabled()
	if err != nil {
		buf.WriteString(fmt.Sprintf("  error checking enabled: %v\n", err))
	} else {
		buf.WriteString(fmt.Sprintf("  enabled: %v\n", enabled))
	}

	// Recent logs.
	buf.WriteString("\nRecent logs:\n")
	logs, err := s.Logs(false, 20)
	if err != nil {
		buf.WriteString(fmt.Sprintf("  error reading logs: %v\n", err))
	} else {
		for _, line := range strings.Split(logs, "\n") {
			if line != "" {
				buf.WriteString(fmt.Sprintf("  %s\n", line))
			}
		}
	}

	// Server process.
	buf.WriteString("\nServer process:\n")
	cmd := exec.Command("pgrep", "-a", "-f", "valheim_server")
	out, err := cmd.CombinedOutput()
	if err != nil {
		buf.WriteString("  no valheim_server process found\n")
	} else {
		buf.WriteString(string(out))
	}

	// Listening UDP sockets.
	buf.WriteString("\nListening UDP sockets:\n")
	cmd = exec.Command("ss", "-lunp")
	out, err = cmd.CombinedOutput()
	if err != nil {
		buf.WriteString(fmt.Sprintf("  error: %v\n", err))
	} else {
		buf.WriteString(string(out))
	}

	// Disk space.
	buf.WriteString("\nDisk space:\n")
	cmd = exec.Command("df", "-h", s.ServiceHome)
	out, err = cmd.CombinedOutput()
	if err != nil {
		buf.WriteString(fmt.Sprintf("  error: %v\n", err))
	} else {
		buf.WriteString(string(out))
	}

	// Directory sizes.
	buf.WriteString("\nDirectory sizes:\n")
	for _, dir := range []string{s.ServerDir, s.SaveDir, s.BackupDir} {
		cmd = exec.Command("du", "-sh", dir)
		out, err = cmd.CombinedOutput()
		if err != nil {
			buf.WriteString(fmt.Sprintf("  %s: error (%v)\n", dir, err))
		} else {
			buf.WriteString(fmt.Sprintf("  %s\n", strings.TrimSpace(string(out))))
		}
	}

	// File ownership.
	buf.WriteString("\nFile ownership:\n")
	cmd = exec.Command("find", s.ServerDir, s.SaveDir, "-maxdepth", "2", "!", "-user", s.ServiceUser, "-ls")
	out, err = cmd.CombinedOutput()
	if err != nil {
		buf.WriteString(fmt.Sprintf("  error: %v\n", err))
	} else if strings.TrimSpace(string(out)) == "" {
		buf.WriteString(fmt.Sprintf("  all files owned by %s\n", s.ServiceUser))
	} else {
		buf.WriteString(string(out))
	}

	// SteamCMD availability.
	buf.WriteString("\nSteamCMD:\n")
	if _, err := exec.LookPath("steamcmd"); err == nil {
		buf.WriteString("  available\n")
	} else {
		buf.WriteString("  not found\n")
	}

	return buf.String(), nil
}

func (s *Service) isServiceActive() (bool, error) {
	cmd := exec.Command(s.SystemctlPath, "is-active", s.ServiceName+".service")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(string(out)) == "active", nil
}

func (s *Service) isServiceEnabled() (bool, error) {
	cmd := exec.Command(s.SystemctlPath, "is-enabled", s.ServiceName+".service")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(string(out)) == "enabled", nil
}

func (s *Service) chownRecursive(path, user string) error {
	cmd := exec.Command("chown", "-R", user+":"+user, path)
	return cmd.Run()
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Compile-time interface check.
var _ cli.ServerManagement = (*Service)(nil)
