package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStartCommand(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the tracked Valheim server",
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := a.start()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), msg)
			return nil
		},
	}
	return cmd
}
