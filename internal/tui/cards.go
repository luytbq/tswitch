package tui

import (
	"fmt"

	"github.com/user/tswitch/internal/tmux"
)

// SessionCard wraps a tmux.Session for grid display.
type SessionCard struct {
	session tmux.Session
}

func (c SessionCard) GetName() string {
	return c.session.Name
}

func (c SessionCard) GetMetadata() string {
	return fmt.Sprintf("%d wins\n%s ago", c.session.WindowCount, formatTimeSince(c.session.LastActive))
}

// WindowCard wraps a tmux.Window for grid display.
type WindowCard struct {
	window tmux.Window
}

func (c WindowCard) GetName() string {
	return fmt.Sprintf("%d: %s", c.window.Index, c.window.Name)
}

func (c WindowCard) GetMetadata() string {
	return fmt.Sprintf("%d pane(s)", c.window.PaneCount)
}
