package startcmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sumwatshade/valheimctl/internal/cli"
)

func NewCommand(s cli.ServerService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the tracked Valheim server",
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := s.Start()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), msg)
			return nil
		},
	}
	return cmd
}
