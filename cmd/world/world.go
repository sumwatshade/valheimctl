package worldcmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sumwatshade/valheimctl/internal/cli"
)

func NewCommand(s cli.ServerManagement) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "world",
		Short: "Manage Valheim world saves",
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := s.WorldList()
			if err != nil {
				return err
			}
			fmt.Print(msg)
			return nil
		},
	}

	cmd.AddCommand(newListCommand(s))
	cmd.AddCommand(newInspectCommand(s))
	cmd.AddCommand(newBackupCommand(s))
	cmd.AddCommand(newRestoreCommand(s))
	cmd.AddCommand(newVerifyCommand(s))
	return cmd
}

func newListCommand(s cli.ServerManagement) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List save directories and backup candidates",
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := s.WorldList()
			if err != nil {
				return err
			}
			fmt.Print(msg)
			return nil
		},
	}
	return cmd
}

func newInspectCommand(s cli.ServerManagement) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect <name>",
		Short: "Inspect a specific world directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := s.WorldInspect(args[0])
			if err != nil {
				return err
			}
			fmt.Print(msg)
			return nil
		},
	}
	return cmd
}

func newBackupCommand(s cli.ServerManagement) *cobra.Command {
	var label string
	cmd := &cobra.Command{
		Use:   "backup [label]",
		Short: "Create a backup of the save root",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				label = args[0]
			}
			msg, err := s.WorldBackup(label)
			if err != nil {
				return err
			}
			fmt.Print(msg)
			return nil
		},
	}
	cmd.Flags().StringVar(&label, "label", "", "backup label (defaults to timestamp)")
	return cmd
}

func newRestoreCommand(s cli.ServerManagement) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "restore <archive>",
		Short: "Restore a world from a backup archive",
		Long: `Restore a world from a backup archive.

Defaults to dry-run. Use --execute to perform the actual restore.

The current save directory is preserved before restoration.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := s.WorldRestore(args[0], !dryRun)
			if err != nil {
				return err
			}
			fmt.Print(msg)
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", true, "perform a validation-only check")
	cmd.Flags().Bool("execute", false, "perform the actual restore (overrides --dry-run)")
	_ = cmd.Flags().Lookup("execute") // keep for future use
	return cmd
}

func newVerifyCommand(s cli.ServerManagement) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify <archive>",
		Short: "Verify a backup archive's integrity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := s.WorldVerify(args[0])
			if err != nil {
				return err
			}
			fmt.Print(msg)
			return nil
		},
	}
	return cmd
}
