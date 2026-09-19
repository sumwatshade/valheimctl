package main

import (
	"github.com/spf13/cobra"
	backupcmd "github.com/sumwatshade/valheimctl/cmd/backup"
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

	cmd.AddCommand(initcmd.NewCommand(serverSvc))
	cmd.AddCommand(startcmd.NewCommand(serverSvc))
	cmd.AddCommand(stopcmd.NewCommand(serverSvc))
	cmd.AddCommand(statuscmd.NewCommand(serverSvc))
	cmd.AddCommand(registercmd.NewCommand(serverSvc))
	cmd.AddCommand(backupcmd.NewCommand(backupSvc))

	return &rootCommand{Command: cmd, app: a}
}

func (r *rootCommand) Execute() error {
	return r.Command.Execute()
}
