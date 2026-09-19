package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newBackupCommand(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Report the current backup status",
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := a.backup()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), msg)
			return nil
		},
	}
	return cmd
}
