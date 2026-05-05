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
	ListAllWindowNames() (map[string][]string, error)  // session -> window names
	ListAllPaneCounts() (map[string]int, error)        // session -> total pane count
	ListPanes(sessionName string, windowIndex int) ([]Pane, error)
	CapturePane(sessionName string, windowIndex int, paneIndex int) (string, error)

	// Navigation
	SwitchToSession(sessionName string) error
	SwitchToLast() error
	SwitchClient(sessionName string, windowIndex int) error
	SelectPane(sessionName string, windowIndex int, paneIndex int) error
	AttachSession(sessionName string) error

	// Session management
	NewSession(sessionName string) error
	NewSessionInDir(sessionName string, dir string) error
	HasSession(sessionName string) bool
	RenameSession(oldName, newName string) error
	KillSession(sessionName string) error

	// Window management
	NewWindow(sessionName string, windowName string) error
	RenameWindow(sessionName string, windowIndex int, newName string) error
	KillWindow(sessionName string, windowIndex int) error
	MoveWindow(srcSession string, srcIndex int, dstSession string) error
	SwapWindow(sessionName string, srcIndex, dstIndex int) error

	// Pane management
	JoinPane(srcSession string, srcWindow, srcPane int, dstSession string, dstWindow int) error
}

// Session represents a TMUX session.
type Session struct {
	Name        string
	WindowCount int
	PaneCount   int
	Attached    bool
	Created     time.Time
	LastActive  time.Time
	Width       int
	Height      int
	// Active pane state in the active window (populated from list-sessions).
	ActivePaneDir   string
	ActivePaneCmd   string
	ActivePaneTitle string
	ActivePanePID   int
}

// Window represents a TMUX window.
type Window struct {
	Index      int
	Name       string
	PaneCount  int
	Active     bool
	Layout     string
	WorkingDir string // CWD of the active pane (pane_current_path)
	// Active pane state (populated from list-windows).
	ActivePaneCmd   string
	ActivePaneTitle string
	ActivePanePID   int
}

// Pane represents a TMUX pane.
type Pane struct {
	Index      int
	Active     bool
	Width      int
	Height     int
	Command    string
	WorkingDir string
	Title      string // pane_title — used for SSH/FTP connection detection
	PID        int    // pane_pid — used to read SSH process args via ps
}
