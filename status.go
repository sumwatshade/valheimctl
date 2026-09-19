package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStatusCommand(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Report the Valheim server status",
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := a.status()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), msg)
			return nil
		},
	}
	return cmd
}
