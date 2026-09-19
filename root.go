package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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

	cmd.AddCommand(newInitCommand(a))
	cmd.AddCommand(newStartCommand(a))
	cmd.AddCommand(newStopCommand(a))
	cmd.AddCommand(newStatusCommand(a))
	cmd.AddCommand(newRegisterCommand(a))
	cmd.AddCommand(newBackupCommand(a))

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
