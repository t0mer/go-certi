package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

// Config holds the runtime configuration for go-certi.
// Only fields that should be persisted to config.json carry json tags;
// runtime-only fields use json:"-".
type Config struct {
	Port    int    `json:"port"`
	ConfDir string `json:"-"` // runtime-only, not persisted
}

// DefaultConfDir returns the platform-appropriate config directory.
//
// Resolution order:
//  1. $XDG_CONFIG_HOME/go-certi  (Linux/Mac when XDG_CONFIG_HOME is set)
//  2. %AppData%/go-certi          (Windows)
//  3. $HOME/.config/go-certi      (fallback for Linux/Mac)
func DefaultConfDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "go-certi")
	}
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("AppData"); appData != "" {
			return filepath.Join(appData, "go-certi")
		}
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "go-certi")
}

// OverrideFromEnv replaces *val with the value of envKey when the env var is
// set and non-empty. This implements the "env wins over CLI flag" rule from §8
// of CLAUDE.md.
func OverrideFromEnv(envKey string, val *string) {
	if v, ok := os.LookupEnv(envKey); ok && v != "" {
		*val = v
	}
}

// LoadOrCreate reads config.json from confDir, creating the directory and a
// default config file if they do not already exist. The returned *Config
// always has ConfDir set to the provided confDir.
func LoadOrCreate(confDir string) (*Config, error) {
	if err := os.MkdirAll(confDir, 0o700); err != nil {
		return nil, err
	}

	cfgPath := filepath.Join(confDir, "config.json")
	cfg := &Config{Port: 8111, ConfDir: confDir}

	data, err := os.ReadFile(cfgPath)
	if os.IsNotExist(err) {
		return cfg, writeJSON(cfgPath, cfg)
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	cfg.ConfDir = confDir
	return cfg, nil
}

// writeJSON serialises v as indented JSON and writes it to path with mode 0600.
func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
