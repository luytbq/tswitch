package tmux

import "time"

// Service defines the operations the TUI needs from tmux.
// All tmux interactions go through this interface, making the TUI
// testable with a mock implementation.
type Service interface {
	// Queries
	IsInTmux() bool
	ListSessions() ([]Session, error)
	ListWindows(sessionName string) ([]Window, error)
	ListAllWindowNames() (map[string][]string, error) // session -> window names
	ListPanes(sessionName string, windowIndex int) ([]Pane, error)
	CapturePane(sessionName string, windowIndex int, paneIndex int) (string, error)

	// Navigation
	SwitchToSession(sessionName string) error
	SwitchClient(sessionName string, windowIndex int) error
	AttachSession(sessionName string) error

	// Session management
	NewSession(sessionName string) error
	RenameSession(oldName, newName string) error
	KillSession(sessionName string) error

	// Window management
	NewWindow(sessionName string, windowName string) error
	RenameWindow(sessionName string, windowIndex int, newName string) error
	KillWindow(sessionName string, windowIndex int) error
}

// Session represents a TMUX session.
type Session struct {
	Name        string
	WindowCount int
	Attached    bool
	Created     time.Time
	LastActive  time.Time
	Width       int
	Height      int
}

// Window represents a TMUX window.
type Window struct {
	Index      int
	Name       string
	PaneCount  int
	Active     bool
	Layout     string
	WorkingDir string
}

// Pane represents a TMUX pane.
type Pane struct {
	Index      int
	Active     bool
	Width      int
	Height     int
	Command    string
	WorkingDir string
}
