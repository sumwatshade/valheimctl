package statuscmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sumwatshade/valheimctl/internal/cli"
)

func NewCommand(s cli.ServerService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Report the Valheim server status",
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := s.Status()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), msg)
			return nil
		},
	}
	return cmd
}
