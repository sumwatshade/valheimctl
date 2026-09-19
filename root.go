package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	backupcmd "github.com/sumwatshade/valheimctl/cmd/backup"
	initcmd "github.com/sumwatshade/valheimctl/cmd/init"
	registercmd "github.com/sumwatshade/valheimctl/cmd/register"
	startcmd "github.com/sumwatshade/valheimctl/cmd/start"
	statuscmd "github.com/sumwatshade/valheimctl/cmd/status"
	stopcmd "github.com/sumwatshade/valheimctl/cmd/stop"
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

	cmd.AddCommand(initcmd.NewCommand(a))
	cmd.AddCommand(startcmd.NewCommand(a))
	cmd.AddCommand(stopcmd.NewCommand(a))
	cmd.AddCommand(statuscmd.NewCommand(a))
	cmd.AddCommand(registercmd.NewCommand(a))
	cmd.AddCommand(backupcmd.NewCommand(a))

	return &rootCommand{Command: cmd, app: a}
}

func (r *rootCommand) Execute() error {
	return r.Command.Execute()
}

func main() {
	if err := newRootCmd(newApp(".", "", "")).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
