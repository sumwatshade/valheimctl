package registercmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sumwatshade/valheimctl/internal/cli"
)

func NewCommand(s cli.ServerService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register <server-name> <server-dir> <description>",
		Short: "Register a systemd service for the Valheim server",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := s.Register(args[0], args[1], args[2]); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "valheimctl register: server registered")
			return nil
		},
	}
	return cmd
}
