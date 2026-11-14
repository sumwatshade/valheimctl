package logscmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sumwatshade/valheimctl/internal/cli"
)

func NewCommand(s cli.ServerManagement) *cobra.Command {
	var follow bool
	var lines int

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Show journalctl logs for the Valheim server service",
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := s.Logs(follow, lines)
			if err != nil {
				return err
			}
			fmt.Print(msg)
			return nil
		},
	}

	cmd.Flags().BoolVar(&follow, "follow", false, "follow the log output")
	cmd.Flags().IntVar(&lines, "lines", 200, "number of lines to show")

	return cmd
}
