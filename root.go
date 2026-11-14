package main

import (
	"github.com/spf13/cobra"
	backupcmd "github.com/sumwatshade/valheimctl/cmd/backup"
	configcmd "github.com/sumwatshade/valheimctl/cmd/config"
	diagnosticscmd "github.com/sumwatshade/valheimctl/cmd/diagnostics"
	installcmd "github.com/sumwatshade/valheimctl/cmd/install"
	logscmd "github.com/sumwatshade/valheimctl/cmd/logs"
	reinstallcmd "github.com/sumwatshade/valheimctl/cmd/reinstall"
	restartcmd "github.com/sumwatshade/valheimctl/cmd/restart"
	updatecmd "github.com/sumwatshade/valheimctl/cmd/update"
	versioncmd "github.com/sumwatshade/valheimctl/cmd/version"
	worldcmd "github.com/sumwatshade/valheimctl/cmd/world"
	initcmd "github.com/sumwatshade/valheimctl/cmd/init"
	registercmd "github.com/sumwatshade/valheimctl/cmd/register"
	startcmd "github.com/sumwatshade/valheimctl/cmd/start"
	statuscmd "github.com/sumwatshade/valheimctl/cmd/status"
	stopcmd "github.com/sumwatshade/valheimctl/cmd/stop"
	backupsvc "github.com/sumwatshade/valheimctl/internal/backup"
	serversvc "github.com/sumwatshade/valheimctl/internal/server"
)

type rootCommand struct {
	*cobra.Command
	app *app
}

func newRootCmd(a *app) *rootCommand {
	cmd := &cobra.Command{
		Use:           "valheimctl",
		Short:         "Manage a dedicated Valheim server",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	serverSvc := serversvc.NewService(a.rootDir, a.systemctlPath, a.serviceDir)
	serverSvc.RuntimeOS = a.runtimeOS
	serverSvc.OSReleasePath = a.osReleasePath
	serverSvc.PackageManagerPath = a.packageManagerPath
	serverSvc.SteamcmdPath = a.steamcmdPath
	backupSvc := backupsvc.NewService(a.rootDir)
	backupSvc.BackupDir = serverSvc.BackupDir

	cmd.AddCommand(initcmd.NewCommand(serverSvc))
	cmd.AddCommand(startcmd.NewCommand(serverSvc))
	cmd.AddCommand(stopcmd.NewCommand(serverSvc))
	cmd.AddCommand(statuscmd.NewCommand(serverSvc))
	cmd.AddCommand(registercmd.NewCommand(serverSvc))
	cmd.AddCommand(backupcmd.NewCommand(backupSvc))

	// Full server management commands.
	cmd.AddCommand(installcmd.NewCommand(serverSvc))
	cmd.AddCommand(updatecmd.NewCommand(serverSvc))
	cmd.AddCommand(restartcmd.NewCommand(serverSvc))
	cmd.AddCommand(logscmd.NewCommand(serverSvc))
	cmd.AddCommand(versioncmd.NewCommand(serverSvc))
	cmd.AddCommand(configcmd.NewCommand(serverSvc))
	cmd.AddCommand(worldcmd.NewCommand(serverSvc))
	cmd.AddCommand(reinstallcmd.NewCommand(serverSvc))
	cmd.AddCommand(diagnosticscmd.NewCommand(serverSvc))

	return &rootCommand{Command: cmd, app: a}
}

func (r *rootCommand) Execute() error {
	return r.Command.Execute()
}
