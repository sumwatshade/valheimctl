package stopcmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sumwatshade/valheimctl/internal/cli"
)

func NewCommand(s cli.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the tracked Valheim server",
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := s.Stop()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), msg)
			return nil
		},
	}
	return cmd
}
