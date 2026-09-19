package server

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sumwatshade/valheimctl/internal/config"
)

type Service struct {
	RootDir            string
	SystemctlPath      string
	ServiceDir         string
	RuntimeOS          string
	OSReleasePath      string
	PackageManagerPath string
	SteamcmdPath       string
}

func NewService(rootDir, systemctlPath, serviceDir string) *Service {
	if rootDir == "" {
		rootDir = "."
	}
	if systemctlPath == "" {
		systemctlPath = "systemctl"
	}
	if serviceDir == "" {
		serviceDir = filepath.Join(rootDir, "etc", "systemd", "system")
	}
	return &Service{
		RootDir:            rootDir,
		SystemctlPath:      systemctlPath,
		ServiceDir:         serviceDir,
		RuntimeOS:          runtime.GOOS,
		OSReleasePath:      "/etc/os-release",
		PackageManagerPath: "apt-get",
		SteamcmdPath:       "steamcmd",
	}
}

func (s *Service) ConfigDir() string { return config.RootDir(s.RootDir) }
func (s *Service) StatePath() string { return config.StatePath(s.RootDir) }
func (s *Service) MetadataPath() string { return config.MetadataPath(s.RootDir) }

func (s *Service) Init() error {
	if err := config.EnsureInitialized(s.RootDir, s.ServiceDir); err != nil {
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
	packages := []string{"libatomic1", "libpulse-dev", "libpulse0", "steamcmd"}
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
