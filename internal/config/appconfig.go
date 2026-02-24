package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AppConfig holds read-only application settings loaded from tswitch-config.json.
// This is separate from Config (config.yaml) which stores runtime state.
type AppConfig struct {
	Keys map[string]string `json:"keys"`
}

// DefaultAppConfig returns an AppConfig with no overrides (all defaults).
func DefaultAppConfig() *AppConfig {
	return &AppConfig{}
}

// LoadAppConfig searches for tswitch-config.json in order:
//  1. Same directory as the binary
//  2. ~/.tswitch/
//
// Returns DefaultAppConfig if no file is found.
func LoadAppConfig() (*AppConfig, error) {
	// 1. Same directory as binary.
	if exePath, err := os.Executable(); err == nil {
		p := filepath.Join(filepath.Dir(exePath), "tswitch-config.json")
		if cfg, err := loadAppConfigFrom(p); err == nil {
			return cfg, nil
		}
	}

	// 2. ~/.tswitch/
	if home, err := os.UserHomeDir(); err == nil {
		p := filepath.Join(home, ".tswitch", "tswitch-config.json")
		if cfg, err := loadAppConfigFrom(p); err == nil {
			return cfg, nil
		}
	}

	return DefaultAppConfig(), nil
}

func loadAppConfigFrom(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := DefaultAppConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
