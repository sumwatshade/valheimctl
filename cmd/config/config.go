package configcmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sumwatshade/valheimctl/internal/cli"
)

func NewCommand(s cli.ServerManagement) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show or validate the server configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := s.ConfigShow()
			if err != nil {
				return err
			}
			fmt.Print(msg)
			return nil
		},
	}

	cmd.AddCommand(newValidateCommand(s))
	return cmd
}

func newValidateCommand(s cli.ServerManagement) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate the server configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := s.ConfigValidate()
			if err != nil {
				return err
			}
			fmt.Println(msg)
			return nil
		},
	}
	return cmd
}
