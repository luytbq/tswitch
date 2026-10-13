package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	PreviewModeCapture  = "capture"
	PreviewModeMetadata = "metadata"

	// Session sort modes.
	SortRecent = "recent" // most recently attached first
	SortName   = "name"
	SortManual = "manual" // follow SessionOrder

	// Main view modes.
	ViewGrid = "grid"
	ViewList = "list"
)

// Config holds the application configuration.
type Config struct {
	Tags         map[string][]string `yaml:"tags"`
	Marks        map[string]Mark     `yaml:"marks"`
	Settings     Settings            `yaml:"settings"`
	SessionOrder []string            `yaml:"session_order,omitempty"`
	WindowOrder  map[string][]int    `yaml:"window_order,omitempty"`
	// Expanded holds the tree-view expansion state: a session name is present
	// when that session is expanded, and its value lists the expanded window
	// indices inside it. Absent means collapsed.
	Expanded map[string][]int `yaml:"expanded,omitempty"`
}

// Mark represents a bookmarked session/window/pane.
type Mark struct {
	SessionName string `yaml:"session"`
	WindowIndex int    `yaml:"window"`
	PaneIndex   int    `yaml:"pane"`
}

// Settings holds user-level preferences.
type Settings struct {
	PreviewMode string `yaml:"preview_mode"` // last-used preview toggle state; empty = capture
	ViewMode    string `yaml:"view_mode"`    // grid | list
	Theme       string `yaml:"theme"`
	SortBy      string `yaml:"sort_by"` // recent | name | manual
	// PinCurrentFirst keeps the session the client is attached to at the top
	// of the list regardless of sort mode. Pointer so an absent key can be
	// told apart from an explicit false.
	PinCurrentFirst *bool `yaml:"pin_current_first,omitempty"`
}

// Default returns a Config with sensible defaults.
func Default() *Config {
	// PinCurrentFirst is a pointer so an absent key can be told apart from an
	// explicit false; every Config needs its own, not a shared one.
	pinCurrent := true
	return &Config{
		Tags:  make(map[string][]string),
		Marks: make(map[string]Mark),
		Settings: Settings{
			Theme:           "default",
			SortBy:          SortRecent,
			ViewMode:        ViewGrid,
			PinCurrentFirst: &pinCurrent,
		},
	}
}

// LoadState reads the state from ~/.tswitch/state.yaml.
// If the file does not exist, it returns defaults.
func LoadState() (*Config, error) {
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

// SaveState writes the state to ~/.tswitch/state.yaml.
func SaveState(cfg *Config) error {
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

func (c *Config) ClearWindowOrder(session string) {
	delete(c.WindowOrder, session)
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
	if c.Settings.Theme == "" {
		c.Settings.Theme = "default"
	}
	switch c.Settings.SortBy {
	case SortRecent, SortName, SortManual:
	default:
		// "activity" was the name this mode shipped under before it was wired up.
		c.Settings.SortBy = SortRecent
	}
	if c.Settings.ViewMode != ViewList {
		c.Settings.ViewMode = ViewGrid
	}
	if c.Settings.PinCurrentFirst == nil {
		on := true
		c.Settings.PinCurrentFirst = &on
	}
}

// PinCurrentFirst reports whether the attached session is pinned to the top.
func (c *Config) PinCurrentFirst() bool {
	return c.Settings.PinCurrentFirst == nil || *c.Settings.PinCurrentFirst
}

// ---------------------------------------------------------------------------
// Tree expansion helpers
// ---------------------------------------------------------------------------

// IsSessionExpanded reports whether the session's node is open.
func (c *Config) IsSessionExpanded(session string) bool {
	_, ok := c.Expanded[session]
	return ok
}

// IsWindowExpanded reports whether the window's node is open. A window can
// only be open while its session is open.
func (c *Config) IsWindowExpanded(session string, windowIndex int) bool {
	for _, idx := range c.Expanded[session] {
		if idx == windowIndex {
			return true
		}
	}
	return false
}

// SetSessionExpanded opens or closes a session. Closing it also drops the
// expanded windows inside, since they cannot be seen while it is closed.
func (c *Config) SetSessionExpanded(session string, expanded bool) {
	if !expanded {
		delete(c.Expanded, session)
		return
	}
	if c.Expanded == nil {
		c.Expanded = make(map[string][]int)
	}
	if _, ok := c.Expanded[session]; !ok {
		c.Expanded[session] = []int{}
	}
}

// SetWindowExpanded opens or closes a window inside a session.
func (c *Config) SetWindowExpanded(session string, windowIndex int, expanded bool) {
	if !expanded {
		rest := c.Expanded[session][:0]
		for _, idx := range c.Expanded[session] {
			if idx != windowIndex {
				rest = append(rest, idx)
			}
		}
		if _, ok := c.Expanded[session]; ok {
			c.Expanded[session] = rest
		}
		return
	}
	c.SetSessionExpanded(session, true)
	if !c.IsWindowExpanded(session, windowIndex) {
		c.Expanded[session] = append(c.Expanded[session], windowIndex)
	}
}

// CollapseAll closes every node.
func (c *Config) CollapseAll() { c.Expanded = nil }

// RenameSessionState moves the state keyed by session name onto a new name.
func (c *Config) RenameSessionState(oldName, newName string) {
	if windows, ok := c.Expanded[oldName]; ok {
		delete(c.Expanded, oldName)
		if c.Expanded == nil {
			c.Expanded = make(map[string][]int)
		}
		c.Expanded[newName] = windows
	}
	for i, name := range c.SessionOrder {
		if name == oldName {
			c.SessionOrder[i] = newName
		}
	}
	if order, ok := c.WindowOrder[oldName]; ok {
		delete(c.WindowOrder, oldName)
		c.WindowOrder[newName] = order
	}
}

// PruneExpanded drops expansion state for sessions and windows that no longer
// exist, so the file does not accumulate dead entries.
func (c *Config) PruneExpanded(live map[string][]int) {
	for session, windows := range c.Expanded {
		liveWindows, ok := live[session]
		if !ok {
			delete(c.Expanded, session)
			continue
		}
		kept := make([]int, 0, len(windows))
		for _, idx := range windows {
			for _, liveIdx := range liveWindows {
				if idx == liveIdx {
					kept = append(kept, idx)
					break
				}
			}
		}
		c.Expanded[session] = kept
	}
}
