package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration.
type Config struct {
	Tags         map[string][]string `yaml:"tags"`
	Marks        map[string]Mark     `yaml:"marks"`
	Settings     Settings            `yaml:"settings"`
	SessionOrder []string            `yaml:"session_order,omitempty"`
	WindowOrder  map[string][]int    `yaml:"window_order,omitempty"`
}

// Mark represents a bookmarked session/window/pane.
type Mark struct {
	SessionName string `yaml:"session"`
	WindowIndex int    `yaml:"window"`
	PaneIndex   int    `yaml:"pane"`
}

// Settings holds user-level preferences.
type Settings struct {
	DefaultPreview string `yaml:"default_preview"`
	Theme          string `yaml:"theme"`
	SortBy         string `yaml:"sort_by"`
}

// Default returns a Config with sensible defaults.
func Default() *Config {
	return &Config{
		Tags:  make(map[string][]string),
		Marks: make(map[string]Mark),
		Settings: Settings{
			DefaultPreview: "capture",
			Theme:          "default",
			SortBy:         "activity",
		},
	}
}

// LoadConfig reads the config from ~/.tswitch/state.yaml.
// If the file does not exist, it returns defaults.
func LoadConfig() (*Config, error) {
	dir, err := configDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, "state.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return Default(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := Default() // start with defaults so missing fields are populated
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.fillDefaults()
	return cfg, nil
}

// SaveConfig writes the config to ~/.tswitch/state.yaml.
func SaveConfig(cfg *Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, "state.yaml"), data, 0644)
}

// ---------------------------------------------------------------------------
// Marks helpers
// ---------------------------------------------------------------------------

func (c *Config) SetMark(key, sessionName string, windowIndex, paneIndex int) {
	if c.Marks == nil {
		c.Marks = make(map[string]Mark)
	}
	c.Marks[key] = Mark{SessionName: sessionName, WindowIndex: windowIndex, PaneIndex: paneIndex}
}

func (c *Config) GetMark(key string) *Mark {
	if m, ok := c.Marks[key]; ok {
		return &m
	}
	return nil
}

func (c *Config) DeleteMark(key string) { delete(c.Marks, key) }

// RemoveMarksForTarget deletes all existing marks that point to the
// same session+window combination, so that reassigning a new key to
// the same target replaces the old key rather than accumulating duplicates.
func (c *Config) RemoveMarksForTarget(sessionName string, windowIndex int) {
	for key, m := range c.Marks {
		if m.SessionName == sessionName && m.WindowIndex == windowIndex {
			delete(c.Marks, key)
		}
	}
}

func (c *Config) HasMark(key string) bool {
	_, ok := c.Marks[key]
	return ok
}

// GetSessionMarks returns all mark keys whose target session matches.
func (c *Config) GetSessionMarks(sessionName string) []string {
	var out []string
	for key, m := range c.Marks {
		if m.SessionName == sessionName {
			out = append(out, key)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Order helpers
// ---------------------------------------------------------------------------

func (c *Config) SetSessionOrder(order []string) {
	c.SessionOrder = order
}

func (c *Config) SetWindowOrder(session string, indices []int) {
	if c.WindowOrder == nil {
		c.WindowOrder = make(map[string][]int)
	}
	c.WindowOrder[session] = indices
}

// ---------------------------------------------------------------------------
// Tags helpers
// ---------------------------------------------------------------------------

func (c *Config) GetSessionTags(sessionName string) []string {
	var tags []string
	for tag, sessions := range c.Tags {
		for _, s := range sessions {
			if s == sessionName {
				tags = append(tags, tag)
			}
		}
	}
	return tags
}

func (c *Config) AddSessionTag(sessionName, tag string) {
	if c.Tags == nil {
		c.Tags = make(map[string][]string)
	}
	for _, s := range c.Tags[tag] {
		if s == sessionName {
			return
		}
	}
	c.Tags[tag] = append(c.Tags[tag], sessionName)
}

func (c *Config) RemoveSessionTag(sessionName, tag string) {
	sessions := c.Tags[tag]
	for i, s := range sessions {
		if s == sessionName {
			c.Tags[tag] = append(sessions[:i], sessions[i+1:]...)
			return
		}
	}
}

// ---------------------------------------------------------------------------
// Private
// ---------------------------------------------------------------------------

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	return filepath.Join(home, ".tswitch"), nil
}

func (c *Config) fillDefaults() {
	if c.Settings.DefaultPreview == "" {
		c.Settings.DefaultPreview = "capture"
	}
	if c.Settings.Theme == "" {
		c.Settings.Theme = "default"
	}
	if c.Settings.SortBy == "" {
		c.Settings.SortBy = "activity"
	}
}
