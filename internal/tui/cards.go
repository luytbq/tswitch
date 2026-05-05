package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/luytbq/tswitch/internal/tmux"
)

// SessionCard wraps a tmux.Session for grid display.
type SessionCard struct {
	session tmux.Session
}

func (c SessionCard) Title() string {
	return c.session.Name
}

func (c SessionCard) Subtitle() string {
	return fmt.Sprintf("%d wins · %d panes · %s", c.session.WindowCount, c.session.PaneCount, formatTimeSince(c.session.LastActive))
}

func (c SessionCard) Indicator() string {
	if c.session.Attached {
		return "●"
	}
	return ""
}

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
		dir := c.window.WorkingDir
		if idx := strings.LastIndexByte(dir, '/'); idx >= 0 && idx < len(dir)-1 {
			dir = dir[idx+1:]
		}
		info += " · " + dir
	}
	return info
}

func (c WindowCard) Indicator() string {
	if c.window.Active {
		return "●"
	}
	return ""
}

// PaneCard wraps a tmux.Pane for grid display.
type PaneCard struct {
	pane tmux.Pane
}

func (c PaneCard) Title() string {
	return fmt.Sprintf("Pane %d", c.pane.Index)
}

func (c PaneCard) Subtitle() string {
	dir := c.pane.WorkingDir
	if idx := strings.LastIndexByte(dir, '/'); idx >= 0 && idx < len(dir)-1 {
		dir = dir[idx+1:]
	}
	return c.pane.Command + " · " + dir
}

func (c PaneCard) Indicator() string {
	if c.pane.Active {
		return "●"
	}
	return ""
}

// formatTimeSince returns a human-readable relative time string.
func formatTimeSince(t time.Time) string {
	if t.IsZero() {
		return "?"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}
