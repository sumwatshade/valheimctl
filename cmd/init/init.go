package initcmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sumwatshade/valheimctl/internal/cli"
)

func NewCommand(s cli.ServerService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize valheimctl state",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := s.Init(); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "valheimctl init: initialized")
			return nil
		},
	}
	return cmd
}
