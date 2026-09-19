package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type backupMetadata struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Worlds    []string  `json:"worlds,omitempty"`
}

func newBackupCommand(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Manage Valheim world backups",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				msg, err := a.backup()
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), msg)
				return nil
			}
			return cmd.Help()
		},
	}
	cmd.AddCommand(newBackupListCommand(a))
	cmd.AddCommand(newBackupApplyCommand(a))
	cmd.AddCommand(newBackupCreateCommand(a))
	return cmd
}

func newBackupListCommand(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List existing Valheim backups",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			list, err := a.listBackups()
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

func newBackupApplyCommand(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply <name>",
		Short: "Apply a named backup to the active world save directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.applyBackup(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "valheimctl backup apply: restored %s\n", args[0])
			return nil
		},
	}
	return cmd
}

func newBackupCreateCommand(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <name>",
		Aliases: []string{"snapshot"},
		Short:   "Create a backup of all world saves",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.createBackup(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "valheimctl backup create: created %s\n", args[0])
			return nil
		},
	}
	return cmd
}

func (a *app) backupStoreDir() string {
	return filepath.Join(a.rootDir, ".valheimctl", "backups")
}

func (a *app) primaryWorldName() string {
	if st, err := a.readMetadata(); err == nil && st.ServerName != "" {
		return st.ServerName
	}
	return "Dedicated"
}

func (a *app) readMetadata() (state, error) {
	var st state
	if _, err := os.Stat(a.metadataPath()); err != nil {
		return state{}, err
	}
	if err := readJSON(a.metadataPath(), &st); err != nil {
		return state{}, err
	}
	return st, nil
}

func (a *app) listBackups() ([]backupMetadata, error) {
	backupsDir := a.backupStoreDir()
	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	items := make([]backupMetadata, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		metaPath := filepath.Join(backupsDir, entry.Name(), "meta.json")
		meta := backupMetadata{Name: entry.Name()}
		if _, err := os.ReadFile(metaPath); err == nil {
			if err := readJSON(metaPath, &meta); err != nil {
				return nil, err
			}
		} else if !os.IsNotExist(err) {
			return nil, err
		}
		if meta.CreatedAt.IsZero() {
			if info, err := os.Stat(filepath.Join(backupsDir, entry.Name())); err == nil {
				meta.CreatedAt = info.ModTime().UTC()
			}
		}
		items = append(items, meta)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
	return items, nil
}

func (a *app) createBackup(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("backup name is required")
	}
	worldsDir := filepath.Join(a.rootDir, "worlds_local")
	if _, err := os.Stat(worldsDir); err != nil {
		return fmt.Errorf("save directory %q does not exist: %w", worldsDir, err)
	}

	backupDir := filepath.Join(a.backupStoreDir(), name)
	if _, err := os.Stat(backupDir); err == nil {
		return fmt.Errorf("backup %q already exists", name)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return err
	}

	worlds := make([]string, 0)
	entries, err := os.ReadDir(worldsDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		worldName := entry.Name()
		worlds = append(worlds, worldName)
		src := filepath.Join(worldsDir, worldName)
		dst := filepath.Join(backupDir, worldName)
		if err := copyTree(src, dst); err != nil {
			return fmt.Errorf("copy world %q to backup: %w", worldName, err)
		}
	}
	if len(worlds) == 0 {
		return fmt.Errorf("no world folders found in %q", worldsDir)
	}

	meta := backupMetadata{
		Name:      name,
		CreatedAt: time.Now().UTC(),
		Worlds:    worlds,
	}
	if err := writeJSON(filepath.Join(backupDir, "meta.json"), meta); err != nil {
		return err
	}
	return nil
}

func (a *app) applyBackup(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("backup name is required")
	}
	backupDir := filepath.Join(a.backupStoreDir(), name)
	if _, err := os.Stat(backupDir); err != nil {
		return fmt.Errorf("backup %q not found", name)
	}

	worldsDir := filepath.Join(a.rootDir, "worlds_local")
	if err := os.MkdirAll(worldsDir, 0o755); err != nil {
		return err
	}

	worlds := make([]string, 0)
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "." {
			worlds = append(worlds, entry.Name())
		}
	}
	if len(worlds) == 0 {
		return fmt.Errorf("backup %q does not contain any world data", name)
	}

	primary := a.primaryWorldName()
	targetWorld := primary
	if _, err := os.Stat(filepath.Join(worldsDir, targetWorld)); err != nil {
		targetWorld = worlds[0]
	}

	for _, worldName := range worlds {
		src := filepath.Join(backupDir, worldName)
		dst := filepath.Join(worldsDir, worldName)
		if err := os.RemoveAll(dst); err != nil {
			return err
		}
		if err := copyTree(src, dst); err != nil {
			return fmt.Errorf("restore world %q: %w", worldName, err)
		}
	}
	if targetWorld != "" {
		if _, err := os.Stat(filepath.Join(worldsDir, targetWorld)); err == nil {
			return nil
		}
	}
	return nil
}

func (a *app) backup() (string, error) {
	items, err := a.listBackups()
	if err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "No backups found.", nil
	}
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, fmt.Sprintf("%s\t%s", item.Name, item.CreatedAt.Format(time.RFC3339)))
	}
	return strings.Join(lines, "\n"), nil
}

func copyTree(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}
