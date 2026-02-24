package tui

import (
	"fmt"

	"github.com/user/tswitch/internal/tmux"
)

// SessionCard wraps a tmux.Session for grid display.
type SessionCard struct {
	session tmux.Session
}

func (c SessionCard) Title() string {
	return c.session.Name
}

func (c SessionCard) Subtitle() string {
	return fmt.Sprintf("%d wins  %s", c.session.WindowCount, formatTimeSince(c.session.LastActive))
}

func (c SessionCard) Indicator() string {
	if c.session.Attached {
		return "*"
	}
	return ""
}

// Deprecated: kept for backward compat with mark map lookup that uses Title().
func (c SessionCard) GetName() string { return c.Title() }

// WindowCard wraps a tmux.Window for grid display.
type WindowCard struct {
	window tmux.Window
}

func (c WindowCard) Title() string {
	return fmt.Sprintf("%d: %s", c.window.Index, c.window.Name)
}

func (c WindowCard) Subtitle() string {
	info := fmt.Sprintf("%d pane(s)", c.window.PaneCount)
	if c.window.WorkingDir != "" {
		// Show just the last path component to save space.
		dir := c.window.WorkingDir
		if idx := lastIndexByte(dir, '/'); idx >= 0 && idx < len(dir)-1 {
			dir = dir[idx+1:]
		}
		info += "  " + dir
	}
	return info
}

func (c WindowCard) Indicator() string {
	if c.window.Active {
		return "*"
	}
	return ""
}

// Deprecated: kept for backward compat with mark map lookup that uses Title().
func (c WindowCard) GetName() string { return c.Title() }

// lastIndexByte returns the index of the last instance of c in s, or -1.
func lastIndexByte(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}
