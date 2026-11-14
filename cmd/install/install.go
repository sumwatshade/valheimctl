package installcmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sumwatshade/valheimctl/internal/cli"
)

func NewCommand(s cli.ServerManagement) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install the Valheim dedicated server binaries via SteamCMD",
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := s.Install()
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), msg)
			return nil
		},
	}
	return cmd
}
