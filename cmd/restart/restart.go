package restartcmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sumwatshade/valheimctl/internal/cli"
)

func NewCommand(s cli.ServerManagement) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart the Valheim dedicated server service",
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := s.Restart()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), msg)
			return nil
		},
	}
	return cmd
}
