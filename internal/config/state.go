package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type State struct {
	Initialized bool   `json:"initialized"`
	Running     bool   `json:"running"`
	ServerName  string `json:"server_name,omitempty"`
	ServerDir   string `json:"server_dir,omitempty"`
	Description string `json:"description,omitempty"`
}

type Metadata struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Worlds    []string  `json:"worlds,omitempty"`
}

func RootDir(rootDir string) string {
	return filepath.Join(rootDir, ".valheimctl")
}

func StatePath(rootDir string) string {
	return filepath.Join(RootDir(rootDir), "state.json")
}

func MetadataPath(rootDir string) string {
	return filepath.Join(RootDir(rootDir), "server.json")
}

func EnsureInitialized(rootDir, serviceDir string) error {
	if err := os.MkdirAll(RootDir(rootDir), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(serviceDir, 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(StatePath(rootDir)); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	st := State{Initialized: true, Running: false}
	return WriteJSON(StatePath(rootDir), st)
}

func WriteJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

func ReadJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func LoadState(rootDir string) (State, error) {
	var st State
	if err := ReadJSON(StatePath(rootDir), &st); err != nil {
		return State{}, err
	}
	return st, nil
}

func SaveState(rootDir string, st State) error {
	return WriteJSON(StatePath(rootDir), st)
}

func LoadMetadata(rootDir string) (State, error) {
	var st State
	if _, err := os.Stat(MetadataPath(rootDir)); err != nil {
		return State{}, err
	}
	if err := ReadJSON(MetadataPath(rootDir), &st); err != nil {
		return State{}, err
	}
	return st, nil
}

func SaveMetadata(rootDir string, st State) error {
	return WriteJSON(MetadataPath(rootDir), st)
}
