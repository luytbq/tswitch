package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/tswitch/internal/config"
	"github.com/user/tswitch/internal/keys"
	"github.com/user/tswitch/internal/tmux"
)

// Mode represents the current navigation level.
type Mode int

const (
	ModeSessionGrid Mode = iota
	ModeWindowGrid
)

// Model is the top-level Bubbletea model.
type Model struct {
	// Dependencies (injected via constructor).
	tmux   tmux.Service
	config *config.Config
	styles Styles

	// UI components.
	sessionGrid  *Grid
	windowGrid   *Grid
	previewPanel *PreviewPanel

	// State.
	currentMode   Mode
	sessions      []tmux.Session
	windows       []tmux.Window
	currentSess   string // session name when in window view
	helpShown     bool
	markingMode   bool   // waiting for a mark-key press
	markingTarget string // "session" or "window"

	// Viewport.
	width  int
	height int

	// Feedback.
	lastErr     string
	lastErrTime time.Time
}

// NewModel creates a Model wired to a real tmux client.
func NewModel() (*Model, error) {
	return NewModelWith(tmux.NewClient())
}

// NewModelWith creates a Model using the given tmux.Service (useful for tests).
func NewModelWith(svc tmux.Service) (*Model, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		cfg = config.Default()
	}

	styles := NewStyles()
	m := &Model{
		tmux:        svc,
		config:      cfg,
		styles:      styles,
		width:       80,
		height:      24,
		currentMode: ModeSessionGrid,
	}

	m.sessionGrid = NewGrid(50, m.height-3, styles)
	m.windowGrid = NewGrid(50, m.height-3, styles)
	m.previewPanel = NewPreviewPanel(25, m.height-3, styles)

	if err := m.loadSessions(); err != nil {
		m.setError(err.Error())
	}
	if len(m.sessions) > 0 {
		m.previewPanel.SetSessionMetadata(m.sessions[0])
	}

	return m, nil
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model. It dispatches to focused handlers.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
	}
	return m, nil
}

// View implements tea.Model.
func (m *Model) View() string {
	if m.helpShown {
		return m.renderHelp()
	}
	switch m.currentMode {
	case ModeSessionGrid:
		return m.renderSessionView()
	case ModeWindowGrid:
		return m.renderWindowView()
	}
	return ""
}

// ---------------------------------------------------------------------------
// Key dispatch
// ---------------------------------------------------------------------------

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()

	// Marking mode intercepts all keys.
	if m.markingMode {
		return m.handleMarkAssignment(keyStr)
	}

	action := keys.Resolve(keyStr)

	switch action {
	case keys.ActionQuit:
		return m, tea.Quit

	case keys.ActionBack:
		return m.handleBack()

	case keys.ActionToggleHelp:
		m.helpShown = !m.helpShown

	case keys.ActionTogglePreview:
		m.previewPanel.ToggleMode()

	case keys.ActionStartMark:
		m.enterMarkingMode()

	case keys.ActionMoveUp:
		m.moveFocus(0, -1)
	case keys.ActionMoveDown:
		m.moveFocus(0, 1)
	case keys.ActionMoveLeft:
		m.moveFocus(-1, 0)
	case keys.ActionMoveRight:
		m.moveFocus(1, 0)

	case keys.ActionConfirm:
		return m.handleConfirm()

	case keys.ActionQuickSwap:
		return m.handleQuickSwap()

	case keys.ActionNone:
		// Might be a mark-jump key.
		if m.config.HasMark(keyStr) {
			return m.handleJumpToMark(keyStr)
		}
	}

	return m, nil
}

// ---------------------------------------------------------------------------
// Data loading
// ---------------------------------------------------------------------------

func (m *Model) loadSessions() error {
	sessions, err := m.tmux.ListSessions()
	if err != nil {
		return err
	}
	m.sessions = sessions

	items := make([]GridItem, len(sessions))
	for i, s := range sessions {
		items[i] = SessionCard{s}
	}
	m.sessionGrid.SetItems(items)
	return nil
}

func (m *Model) loadWindows(sessionName string) error {
	windows, err := m.tmux.ListWindows(sessionName)
	if err != nil {
		return err
	}
	m.windows = windows
	m.currentSess = sessionName

	items := make([]GridItem, len(windows))
	for i, w := range windows {
		items[i] = WindowCard{w}
	}
	m.windowGrid.SetItems(items)

	if len(m.windows) > 0 {
		m.previewPanel.SetWindowMetadata(m.windows[0])
	}
	return nil
}

// ---------------------------------------------------------------------------
// Resize
// ---------------------------------------------------------------------------

func (m *Model) resize(w, h int) {
	m.width = w
	m.height = h
	m.sessionGrid.SetSize(w-28, h-3)
	m.windowGrid.SetSize(w-28, h-3)
	m.previewPanel.SetSize(25, h-3)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (m *Model) setError(msg string) {
	m.lastErr = msg
	m.lastErrTime = time.Now()
}

func (m *Model) activeGrid() *Grid {
	if m.currentMode == ModeWindowGrid {
		return m.windowGrid
	}
	return m.sessionGrid
}

// formatTimeSince returns a human-readable relative time string.
func formatTimeSince(t time.Time) string {
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
