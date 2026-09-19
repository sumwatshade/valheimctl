package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStopCommand(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the tracked Valheim server",
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := a.stop()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), msg)
			return nil
		},
	}
	return cmd
}
