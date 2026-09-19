package backupcmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/sumwatshade/valheimctl/internal/cli"
)

func NewCommand(s cli.BackupService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Manage Valheim world backups",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				msg, err := s.Backup()
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), msg)
				return nil
			}
			return cmd.Help()
		},
	}
	cmd.AddCommand(newListCommand(s))
	cmd.AddCommand(newApplyCommand(s))
	cmd.AddCommand(newCreateCommand(s))
	return cmd
}

func newListCommand(s cli.BackupService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List existing Valheim backups",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			list, err := s.BackupList()
			if err != nil {
				return err
			}
			if len(list) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No backups found.")
				return nil
			}
			for _, item := range list {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", item.Name, item.CreatedAt.Format(time.RFC3339))
			}
			return nil
		},
	}
	return cmd
}

func newApplyCommand(s cli.BackupService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply <name>",
		Short: "Apply a named backup to the active world save directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := s.BackupApply(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "valheimctl backup apply: restored %s\n", args[0])
			return nil
		},
	}
	return cmd
}

func newCreateCommand(s cli.BackupService) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <name>",
		Aliases: []string{"snapshot"},
		Short:   "Create a backup of all world saves",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := s.BackupCreate(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "valheimctl backup create: created %s\n", args[0])
			return nil
		},
	}
	return cmd
}
