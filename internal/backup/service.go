package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sumwatshade/valheimctl/internal/cli"
	"github.com/sumwatshade/valheimctl/internal/config"
)

type Metadata struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Worlds    []string  `json:"worlds,omitempty"`
}

type Service struct {
	RootDir string
}

func NewService(rootDir string) *Service {
	return &Service{RootDir: rootDir}
}

func (s *Service) StoreDir() string {
	return filepath.Join(s.RootDir, ".valheimctl", "backups")
}

func (s *Service) List() ([]Metadata, error) {
	backupsDir := s.StoreDir()
	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	items := make([]Metadata, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		metaPath := filepath.Join(backupsDir, entry.Name(), "meta.json")
		meta := Metadata{Name: entry.Name()}
		if _, err := os.ReadFile(metaPath); err == nil {
			if err := config.ReadJSON(metaPath, &meta); err != nil {
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

func (s *Service) Create(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("backup name is required")
	}
	worldsDir := filepath.Join(s.RootDir, "worlds_local")
	if _, err := os.Stat(worldsDir); err != nil {
		return fmt.Errorf("save directory %q does not exist: %w", worldsDir, err)
	}

	backupDir := filepath.Join(s.StoreDir(), name)
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

	meta := Metadata{
		Name:      name,
		CreatedAt: time.Now().UTC(),
		Worlds:    worlds,
	}
	if err := config.WriteJSON(filepath.Join(backupDir, "meta.json"), meta); err != nil {
		return err
	}
	return nil
}

func (s *Service) Apply(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("backup name is required")
	}
	backupDir := filepath.Join(s.StoreDir(), name)
	if _, err := os.Stat(backupDir); err != nil {
		return fmt.Errorf("backup %q not found", name)
	}

	worldsDir := filepath.Join(s.RootDir, "worlds_local")
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

	primary := s.primaryWorldName()
	targetWorld := primary
	if _, err := os.Stat(filepath.Join(worldsDir, targetWorld)); err != nil {
		targetWorld = worlds[0]
	}

	src := filepath.Join(backupDir, targetWorld)
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("backup %q does not contain world %q", name, targetWorld)
	}
	if err := copyTree(src, filepath.Join(worldsDir, targetWorld)); err != nil {
		return err
	}
	return nil
}

func (s *Service) primaryWorldName() string {
	if st, err := config.LoadMetadata(s.RootDir); err == nil && st.ServerName != "" {
		return st.ServerName
	}
	return "Dedicated"
}

func (s *Service) Describe() (string, error) {
	items, err := s.List()
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

func (s *Service) Backup() (string, error) {
	return s.Describe()
}

func (s *Service) BackupList() ([]cli.BackupSummary, error) {
	items, err := s.List()
	if err != nil {
		return nil, err
	}
	out := make([]cli.BackupSummary, 0, len(items))
	for _, item := range items {
		out = append(out, cli.BackupSummary{Name: item.Name, CreatedAt: item.CreatedAt})
	}
	return out, nil
}

func (s *Service) BackupApply(name string) error {
	return s.Apply(name)
}

func (s *Service) BackupCreate(name string) error {
	return s.Create(name)
}

func CopyTree(src, dst string) error {
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

func copyTree(src, dst string) error {
	return CopyTree(src, dst)
}
