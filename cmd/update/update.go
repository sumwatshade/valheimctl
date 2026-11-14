package updatecmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sumwatshade/valheimctl/internal/cli"
)

func NewCommand(s cli.ServerManagement) *cobra.Command {
	var validate bool
	var start bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the Valheim dedicated server binaries via SteamCMD",
		Long: `Update the Valheim dedicated server binaries via SteamCMD.

The service must be stopped before updating. A backup of the save root is
created automatically before the update.

Use --start to restart the service after a successful update.
Use --validate to validate the app files during update.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := s.Update(validate, start)
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), msg)
			return nil
		},
	}

	cmd.Flags().BoolVar(&validate, "validate", false, "validate app files during update")
	cmd.Flags().BoolVar(&start, "start", false, "start the service after a successful update")

	return cmd
}
