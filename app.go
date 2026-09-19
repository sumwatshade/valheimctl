package main

import (
	"path/filepath"
	"runtime"

	backupsvc "github.com/sumwatshade/valheimctl/internal/backup"
	"github.com/sumwatshade/valheimctl/internal/config"
	serversvc "github.com/sumwatshade/valheimctl/internal/server"
)

type state = config.State

type app struct {
	rootDir            string
	systemctlPath      string
	serviceDir         string
	runtimeOS          string
	osReleasePath      string
	packageManagerPath string
	steamcmdPath       string
	serverSvc          *serversvc.Service
	backupSvc          *backupsvc.Service
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

	app := &app{
		rootDir:            rootDir,
		systemctlPath:      systemctlPath,
		serviceDir:         serviceDir,
		runtimeOS:          runtime.GOOS,
		osReleasePath:      "/etc/os-release",
		packageManagerPath: "apt-get",
		steamcmdPath:       "steamcmd",
		serverSvc:          serversvc.NewService(rootDir, systemctlPath, serviceDir),
		backupSvc:          backupsvc.NewService(rootDir),
	}
	app.syncServices()
	return app
}

func (a *app) syncServices() {
	if a.serverSvc == nil {
		a.serverSvc = serversvc.NewService(a.rootDir, a.systemctlPath, a.serviceDir)
	}
	if a.backupSvc == nil {
		a.backupSvc = backupsvc.NewService(a.rootDir)
	}
	a.serverSvc.RootDir = a.rootDir
	a.serverSvc.SystemctlPath = a.systemctlPath
	a.serverSvc.ServiceDir = a.serviceDir
	a.serverSvc.RuntimeOS = a.runtimeOS
	a.serverSvc.OSReleasePath = a.osReleasePath
	a.serverSvc.PackageManagerPath = a.packageManagerPath
	a.serverSvc.SteamcmdPath = a.steamcmdPath
	a.backupSvc.RootDir = a.rootDir
}

func writeJSON(path string, v any) error {
	return config.WriteJSON(path, v)
}

func readJSON(path string, v any) error {
	return config.ReadJSON(path, v)
}

func copyTree(src, dst string) error {
	return backupsvc.CopyTree(src, dst)
}
