package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// BrowseDir defines a directory to scan for subdirectories when browsing.
type BrowseDir struct {
	Path  string `json:"path"`
	Depth int    `json:"depth"`
}

// UIConfig holds visual/layout preferences.
type UIConfig struct {
	CardMinWidth int `json:"card_min_width"` // minimum card content width; 0 = use built-in default

	// PreviewRefreshMs is how often the capture preview refetches. A pointer
	// so that an explicit 0 (turn the refresh off) stays distinct from the
	// field being absent.
	PreviewRefreshMs *int `json:"preview_refresh_ms"`
}

const (
	defaultPreviewRefreshMs = 1000
	// Every refresh forks a tmux process, so a config asking for a 5ms period
	// would spend the whole terminal on capture-pane.
	minPreviewRefreshMs = 100
)

// PreviewRefresh returns the capture refresh period, or 0 when the live
// refresh is off.
func (c *AppConfig) PreviewRefresh() time.Duration {
	ms := defaultPreviewRefreshMs
	if c.UI.PreviewRefreshMs != nil {
		ms = *c.UI.PreviewRefreshMs
	}
	if ms <= 0 {
		return 0
	}
	if ms < minPreviewRefreshMs {
		ms = minPreviewRefreshMs
	}
	return time.Duration(ms) * time.Millisecond
}

// AppConfig holds read-only application settings loaded from tswitch-config.json.
// This is separate from Config (config.yaml) which stores runtime state.
type AppConfig struct {
	Keys          map[string]string `json:"keys"`
	BrowseDirs    []BrowseDir       `json:"browse_dirs"`
	BrowseExclude []string          `json:"browse_exclude"`
	UI            UIConfig          `json:"ui"`
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
