package versioncmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sumwatshade/valheimctl/internal/cli"
)

func NewCommand(s cli.ServerManagement) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Report version information about the server installation",
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := s.Version()
			if err != nil {
				return err
			}
			fmt.Print(msg)
			return nil
		},
	}
	return cmd
}
