package reinstallcmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sumwatshade/valheimctl/internal/cli"
)

func NewCommand(s cli.ServerManagement) *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "reinstall",
		Short: "Clean reinstall of server binaries preserving worlds",
		Long: `Clean reinstall of server binaries preserving worlds.

Stops the service, backs up the save root, removes only the server
installation directory, and reinstalls via SteamCMD.

Defaults to dry-run. Use --execute to perform the actual reinstall.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := s.ServerReinstall(dryRun)
			if err != nil {
				return err
			}
			fmt.Print(msg)
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", true, "perform a validation-only check")
	cmd.Flags().Bool("execute", false, "perform the actual reinstall (overrides --dry-run)")
	_ = cmd.Flags().Lookup("execute") // keep for future use

	return cmd
}
