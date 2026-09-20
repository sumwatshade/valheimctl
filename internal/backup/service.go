package backup

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
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

type readFS interface {
	fs.FS
	fs.ReadDirFS
	fs.ReadFileFS
	fs.StatFS
}

type writableFS interface {
	MkdirAll(path string, perm fs.FileMode) error
	WriteFile(name string, data []byte, perm fs.FileMode) error
	RemoveAll(path string) error
}

type FileSystem interface {
	readFS
	writableFS
}

type osFS struct{}

func (osFS) ReadDir(name string) ([]fs.DirEntry, error) {
	entries, err := os.ReadDir(name)
	if err != nil {
		return nil, err
	}
	out := make([]fs.DirEntry, len(entries))
	for i, entry := range entries {
		out[i] = entry
	}
	return out, nil
}

func (osFS) Open(name string) (fs.File, error)    { return os.Open(name) }
func (osFS) ReadFile(name string) ([]byte, error) { return os.ReadFile(name) }
func (osFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}
func (osFS) MkdirAll(path string, perm fs.FileMode) error { return os.MkdirAll(path, perm) }
func (osFS) RemoveAll(path string) error                  { return os.RemoveAll(path) }
func (osFS) Stat(name string) (fs.FileInfo, error)        { return os.Stat(name) }

type Service struct {
	RootDir string
	fs      FileSystem
}

func NewService(rootDir string, fsys ...FileSystem) *Service {
	service := &Service{RootDir: rootDir, fs: osFS{}}
	if len(fsys) > 0 && fsys[0] != nil {
		service.fs = fsys[0]
	}
	return service
}

func (s *Service) StoreDir() string {
	return filepath.Join(s.RootDir, ".valheimctl", "backups")
}

func readJSONFile(fsys FileSystem, path string, v any) error {
	b, err := fsys.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func writeJSONFile(fsys FileSystem, path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return fsys.WriteFile(path, b, 0o600)
}

func (s *Service) List() ([]Metadata, error) {
	backupsDir := s.StoreDir()
	entries, err := s.fs.ReadDir(backupsDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
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
		if _, err := s.fs.Stat(metaPath); err == nil {
			if err := readJSONFile(s.fs, metaPath, &meta); err != nil {
				return nil, err
			}
		} else if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		if meta.CreatedAt.IsZero() {
			if info, err := s.fs.Stat(filepath.Join(backupsDir, entry.Name())); err == nil {
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
	if _, err := s.fs.Stat(worldsDir); err != nil {
		return fmt.Errorf("save directory %q does not exist: %w", worldsDir, err)
	}

	backupDir := filepath.Join(s.StoreDir(), name)
	if _, err := s.fs.Stat(backupDir); err == nil {
		return fmt.Errorf("backup %q already exists", name)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := s.fs.MkdirAll(backupDir, 0o755); err != nil {
		return err
	}

	worlds := make([]string, 0)
	entries, err := s.fs.ReadDir(worldsDir)
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
		if err := s.copyTree(src, dst); err != nil {
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
	if err := writeJSONFile(s.fs, filepath.Join(backupDir, "meta.json"), meta); err != nil {
		return err
	}
	return nil
}

func (s *Service) Apply(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("backup name is required")
	}
	backupDir := filepath.Join(s.StoreDir(), name)
	if _, err := s.fs.Stat(backupDir); err != nil {
		return fmt.Errorf("backup %q not found", name)
	}

	worldsDir := filepath.Join(s.RootDir, "worlds_local")
	if err := s.fs.MkdirAll(worldsDir, 0o755); err != nil {
		return err
	}

	worlds := make([]string, 0)
	entries, err := s.fs.ReadDir(backupDir)
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
	if _, err := s.fs.Stat(filepath.Join(worldsDir, targetWorld)); err != nil {
		targetWorld = worlds[0]
	}

	src := filepath.Join(backupDir, targetWorld)
	if _, err := s.fs.Stat(src); err != nil {
		return fmt.Errorf("backup %q does not contain world %q", name, targetWorld)
	}
	if err := s.fs.RemoveAll(filepath.Join(worldsDir, targetWorld)); err != nil {
		return err
	}
	if err := s.copyTree(src, filepath.Join(worldsDir, targetWorld)); err != nil {
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

func (s *Service) copyTree(src, dst string) error {
	entries, err := s.fs.ReadDir(src)
	if err != nil {
		return err
	}
	if err := s.fs.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	for _, entry := range entries {
		fullPath := filepath.Join(src, entry.Name())
		targetPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := s.copyTree(fullPath, targetPath); err != nil {
				return err
			}
			continue
		}
		data, err := s.fs.ReadFile(fullPath)
		if err != nil {
			return err
		}
		if err := s.fs.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		if err := s.fs.WriteFile(targetPath, data, 0o600); err != nil {
			return err
		}
	}
	return nil
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
