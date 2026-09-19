package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newInitCommand(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize valheimctl state",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.init(); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "valheimctl init: initialized")
			return nil
		},
	}
	return cmd
}
