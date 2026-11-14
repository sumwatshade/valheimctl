package diagnosticscmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sumwatshade/valheimctl/internal/cli"
)

func NewCommand(s cli.ServerManagement) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diagnostics",
		Short: "Collect diagnostic information about the server setup",
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := s.Diagnostics()
			if err != nil {
				return err
			}
			fmt.Print(msg)
			return nil
		},
	}
	return cmd
}
