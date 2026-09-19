package config

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	defaultPort         = 2456
	defaultWorld        = "Dedicated"
	defaultSaveDir      = "~/.config/unity3d/IronGate/Valheim"
	defaultPublic       = true
	defaultLogFile      = ""
	defaultSaveInterval = 1800
	defaultBackups      = 4
	defaultBackupShort  = 7200
	defaultBackupLong   = 43200
)

// Config models the subset of Valheim server options that are explicitly
// supported by valheimctl and map to command-line arguments used by the server.
type Config struct {
	Name         string            `yaml:"name"`
	Port         int               `yaml:"port"`
	World        string            `yaml:"world"`
	Password     string            `yaml:"password"`
	SaveDir      string            `yaml:"savedir"`
	Public       bool              `yaml:"public"`
	LogFile      string            `yaml:"logfile"`
	SaveInterval int               `yaml:"saveinterval"`
	Backups      int               `yaml:"backups"`
	BackupShort  int               `yaml:"backupshort"`
	BackupLong   int               `yaml:"backuplong"`
	Crossplay    bool              `yaml:"crossplay"`
	InstanceID   string            `yaml:"instanceid"`
	Preset       string            `yaml:"preset"`
	Modifiers    map[string]string `yaml:"modifiers"`
	SetKey       string            `yaml:"setkey"`
}

var validPresets = map[string]struct{}{
	"normal": {}, "casual": {}, "easy": {}, "hard": {}, "hardcore": {}, "immersive": {}, "hammer": {},
}

var validModifierKinds = map[string]map[string]struct{}{
	"combat": {
		"veryeasy": {}, "easy": {}, "hard": {}, "veryhard": {},
	},
	"deathpenalty": {
		"casual": {}, "veryeasy": {}, "easy": {}, "hard": {}, "hardcore": {},
	},
	"resources": {
		"muchless": {}, "less": {}, "more": {}, "muchmore": {}, "most": {},
	},
	"raids": {
		"none": {}, "muchless": {}, "less": {}, "more": {}, "muchmore": {},
	},
	"portals": {
		"casual": {}, "hard": {}, "veryhard": {},
	},
}

var validSetKeys = map[string]struct{}{
	"nobuildcost": {}, "playerevents": {}, "passivemobs": {}, "nomap": {},
}

func LoadFile(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	return Parse(b)
}

func Parse(data []byte) (Config, error) {
	cfg := Config{
		Port:         defaultPort,
		World:        defaultWorld,
		SaveDir:      defaultSaveDir,
		Public:       defaultPublic,
		SaveInterval: defaultSaveInterval,
		Backups:      defaultBackups,
		BackupShort:  defaultBackupShort,
		BackupLong:   defaultBackupLong,
		Modifiers:    map[string]string{},
	}

	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return Config{}, fmt.Errorf("parse yaml: %w", err)
	}
	if raw == nil {
		raw = map[string]any{}
	}

	if v, ok := raw["name"]; ok {
		cfg.Name = valueString(v)
	}
	if v, ok := raw["port"]; ok {
		cfg.Port = valueInt(v)
	}
	if v, ok := raw["world"]; ok {
		cfg.World = valueString(v)
	}
	if v, ok := raw["password"]; ok {
		cfg.Password = valueString(v)
	}
	if v, ok := raw["savedir"]; ok {
		cfg.SaveDir = valueString(v)
	}
	if v, ok := raw["public"]; ok {
		public, err := valueBool(v)
		if err != nil {
			return Config{}, err
		}
		cfg.Public = public
	}
	if v, ok := raw["logfile"]; ok {
		cfg.LogFile = valueString(v)
	}
	if v, ok := raw["saveinterval"]; ok {
		cfg.SaveInterval = valueInt(v)
	}
	if v, ok := raw["backups"]; ok {
		cfg.Backups = valueInt(v)
	}
	if v, ok := raw["backupshort"]; ok {
		cfg.BackupShort = valueInt(v)
	}
	if v, ok := raw["backuplong"]; ok {
		cfg.BackupLong = valueInt(v)
	}
	if v, ok := raw["crossplay"]; ok {
		crossplay, err := valueBool(v)
		if err != nil {
			return Config{}, err
		}
		cfg.Crossplay = crossplay
	}
	if v, ok := raw["instanceid"]; ok {
		cfg.InstanceID = valueString(v)
	}
	if v, ok := raw["preset"]; ok {
		cfg.Preset = valueString(v)
	}
	if v, ok := raw["modifiers"]; ok {
		cfg.Modifiers = modifierMap(v)
	}
	if v, ok := raw["setkey"]; ok {
		cfg.SetKey = valueString(v)
	}
	if cfg.Modifiers == nil {
		cfg.Modifiers = map[string]string{}
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func valueString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case fmt.Stringer:
		return x.String()
	default:
		return fmt.Sprint(v)
	}
}

func valueInt(v any) int {
	switch x := v.(type) {
	case int:
		return x
	case int32:
		return int(x)
	case int64:
		return int(x)
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return 0
		}
		return int(x)
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(x))
		if err == nil {
			return n
		}
		return 0
	default:
		return 0
	}
}

func valueBool(v any) (bool, error) {
	switch x := v.(type) {
	case bool:
		return x, nil
	case int:
		return x != 0, nil
	case int32:
		return x != 0, nil
	case int64:
		return x != 0, nil
	case float64:
		return x != 0, nil
	case string:
		normalized := strings.TrimSpace(strings.ToLower(x))
		switch normalized {
		case "true", "1", "yes", "y", "on":
			return true, nil
		case "false", "0", "no", "n", "off":
			return false, nil
		default:
			return false, fmt.Errorf("boolean value must be true/false or 1/0")
		}
	default:
		return false, fmt.Errorf("boolean value must be true/false or 1/0")
	}
}

func modifierMap(v any) map[string]string {
	out := map[string]string{}
	m, ok := v.(map[string]any)
	if !ok {
		return out
	}
	for key, rawValue := range m {
		out[key] = valueString(rawValue)
	}
	return out
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(c.Password) == "" {
		return fmt.Errorf("password is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if strings.TrimSpace(c.World) == "" {
		return fmt.Errorf("world is required")
	}
	if strings.TrimSpace(c.SaveDir) == "" {
		return fmt.Errorf("savedir must not be empty")
	}
	if c.SaveInterval <= 0 {
		return fmt.Errorf("saveinterval must be greater than zero")
	}
	if c.Backups < 0 {
		return fmt.Errorf("backups must be zero or greater")
	}
	if c.BackupShort <= 0 {
		return fmt.Errorf("backupshort must be greater than zero")
	}
	if c.BackupLong <= 0 {
		return fmt.Errorf("backuplong must be greater than zero")
	}
	if strings.TrimSpace(c.Preset) != "" {
		if _, ok := validPresets[strings.ToLower(c.Preset)]; !ok {
			return fmt.Errorf("preset %q is invalid; expected one of: normal, casual, easy, hard, hardcore, immersive, hammer", c.Preset)
		}
	}
	if strings.TrimSpace(c.SetKey) != "" {
		if _, ok := validSetKeys[strings.ToLower(c.SetKey)]; !ok {
			return fmt.Errorf("setkey %q is invalid; expected one of: nobuildcost, playerevents, passivemobs, nomap", c.SetKey)
		}
	}
	for key, value := range c.Modifiers {
		kind := strings.ToLower(strings.TrimSpace(key))
		if kind == "" {
			return fmt.Errorf("modifier key cannot be empty")
		}
		v := strings.ToLower(strings.TrimSpace(value))
		if v == "" {
			return fmt.Errorf("modifier %q value cannot be empty", key)
		}
		allowed, ok := validModifierKinds[kind]
		if !ok {
			return fmt.Errorf("modifier %q is unsupported; supported kinds: combat, deathpenalty, resources, raids, portals", key)
		}
		if _, ok := allowed[v]; !ok {
			return fmt.Errorf("modifier %q=%q is invalid for kind %q", key, value, kind)
		}
	}
	if strings.TrimSpace(c.LogFile) != "" && strings.TrimSpace(c.LogFile) == c.SaveDir {
		return fmt.Errorf("logfile and savedir must not point to the same path")
	}
	if c.Backups > 0 && c.BackupShort <= 0 {
		return fmt.Errorf("backupshort must be greater than zero when backups are enabled")
	}
	if c.BackupLong <= 0 {
		return fmt.Errorf("backuplong must be greater than zero")
	}
	return nil
}

func (c Config) Args() []string {
	args := []string{}
	if c.Name != "" {
		args = append(args, "-name", c.Name)
	}
	if c.Port > 0 {
		args = append(args, "-port", strconv.Itoa(c.Port))
	}
	if c.World != "" {
		args = append(args, "-world", c.World)
	}
	if c.Password != "" {
		args = append(args, "-password", c.Password)
	}
	if c.SaveDir != "" {
		args = append(args, "-savedir", c.SaveDir)
	}
	if c.Public {
		args = append(args, "-public", "1")
	}
	if c.LogFile != "" {
		args = append(args, "-logFile", c.LogFile)
	}
	if c.SaveInterval > 0 {
		args = append(args, "-saveinterval", strconv.Itoa(c.SaveInterval))
	}
	if c.Backups >= 0 {
		args = append(args, "-backups", strconv.Itoa(c.Backups))
	}
	if c.BackupShort > 0 {
		args = append(args, "-backupshort", strconv.Itoa(c.BackupShort))
	}
	if c.BackupLong > 0 {
		args = append(args, "-backuplong", strconv.Itoa(c.BackupLong))
	}
	if c.Crossplay {
		args = append(args, "-crossplay")
	}
	if c.InstanceID != "" {
		args = append(args, "-instanceid", c.InstanceID)
	}
	if c.Preset != "" {
		args = append(args, "-preset", c.Preset)
	}
	if len(c.Modifiers) > 0 {
		keys := make([]string, 0, len(c.Modifiers))
		for key := range c.Modifiers {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			args = append(args, "-modifier", key, c.Modifiers[key])
		}
	}
	if c.SetKey != "" {
		args = append(args, "-setkey", c.SetKey)
	}
	return args
}
