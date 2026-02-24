package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	currentMode      Mode
	sessions         []tmux.Session
	windows          []tmux.Window
	currentSess      string // session name when in window view
	windowsBySession map[string][]string // session -> window names (for search)
	helpShown     bool
	markingMode   bool   // waiting for a mark-key press
	markingTarget string // "session" or "window"
	filterMode    bool   // search input is active
	filterQuery   string // current fuzzy-search term
	dialog        *Dialog
	pendingAction dialogAction

	// Viewport.
	width  int
	height int

	// Feedback (status bar message with auto-expiry).
	statusMsg     string
	statusMsgTime time.Time
	isStatusError bool
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

	gridW, gridH, previewW, previewH := m.layoutSizes()
	m.sessionGrid = NewGrid(gridW, gridH, styles)
	m.windowGrid = NewGrid(gridW, gridH, styles)
	m.previewPanel = NewPreviewPanel(previewW, previewH, styles)

	if err := m.loadSessions(); err != nil {
		m.setStatusError(err.Error())
	}
	m.syncPreview() //nolint — preview is always metadata on startup, cmd is nil

	return m, nil
}

// captureResultMsg carries the output of an async pane capture.
type captureResultMsg struct{ content string }

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
	case captureResultMsg:
		m.previewPanel.SetCaptureContent(msg.content)
	}
	return m, nil
}

// View implements tea.Model.
func (m *Model) View() string {
	if m.dialog != nil {
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			m.dialog.Render(m.width, m.height))
	}
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

	// Dialog intercepts all keys.
	if m.dialog != nil {
		return m.handleDialogKey(msg)
	}

	// Filter mode intercepts all keys.
	if m.filterMode {
		return m.handleFilterKey(msg)
	}

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
		return m, m.syncPreview()

	case keys.ActionStartMark:
		m.enterMarkingMode()

	case keys.ActionFilter:
		return m, m.enterFilterMode()

	case keys.ActionNew:
		return m.handleNew()

	case keys.ActionRename:
		return m.handleRename()

	case keys.ActionKill:
		return m.handleKill()

	case keys.ActionMoveUp:
		return m, m.moveFocus(0, -1)
	case keys.ActionMoveDown:
		return m, m.moveFocus(0, 1)
	case keys.ActionMoveLeft:
		return m, m.moveFocus(-1, 0)
	case keys.ActionMoveRight:
		return m, m.moveFocus(1, 0)

	case keys.ActionReorderUp:
		return m.handleReorder(0, -1)
	case keys.ActionReorderDown:
		return m.handleReorder(0, 1)
	case keys.ActionReorderLeft:
		return m.handleReorder(-1, 0)
	case keys.ActionReorderRight:
		return m.handleReorder(1, 0)

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
	sessions = m.applySavedSessionOrder(sessions)
	m.sessions = sessions

	// Pre-fetch all window names so session filtering can match against them.
	if wbs, err := m.tmux.ListAllWindowNames(); err == nil {
		m.windowsBySession = wbs
	}

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
	windows = m.applySavedWindowOrder(sessionName, windows)
	m.windows = windows
	m.currentSess = sessionName

	items := make([]GridItem, len(windows))
	for i, w := range windows {
		items[i] = WindowCard{w}
	}
	m.windowGrid.SetItems(items)
	return nil
}

// applyFilter re-filters the current mode's items from the full list and
// updates the grid. Called whenever filterQuery changes.
func (m *Model) applyFilter() {
	switch m.currentMode {
	case ModeSessionGrid:
		filtered := FilterSessions(m.sessions, m.filterQuery, m.windowsBySession)
		items := make([]GridItem, len(filtered))
		for i, s := range filtered {
			items[i] = SessionCard{s}
		}
		m.sessionGrid.SetItems(items)
	case ModeWindowGrid:
		filtered := FilterWindows(m.windows, m.filterQuery)
		items := make([]GridItem, len(filtered))
		for i, w := range filtered {
			items[i] = WindowCard{w}
		}
		m.windowGrid.SetItems(items)
	}
}

// applySavedSessionOrder reorders sessions according to the saved order.
// Sessions not in the saved order are appended at the end.
func (m *Model) applySavedSessionOrder(sessions []tmux.Session) []tmux.Session {
	if len(m.config.SessionOrder) == 0 {
		return sessions
	}

	byName := make(map[string]tmux.Session, len(sessions))
	for _, s := range sessions {
		byName[s.Name] = s
	}

	result := make([]tmux.Session, 0, len(sessions))
	seen := make(map[string]bool)

	for _, name := range m.config.SessionOrder {
		if s, ok := byName[name]; ok {
			result = append(result, s)
			seen[name] = true
		}
	}
	for _, s := range sessions {
		if !seen[s.Name] {
			result = append(result, s)
		}
	}
	return result
}

// applySavedWindowOrder reorders windows according to the saved order.
// Windows not in the saved order are appended at the end.
func (m *Model) applySavedWindowOrder(sessionName string, windows []tmux.Window) []tmux.Window {
	order, ok := m.config.WindowOrder[sessionName]
	if !ok || len(order) == 0 {
		return windows
	}

	byIndex := make(map[int]tmux.Window, len(windows))
	for _, w := range windows {
		byIndex[w.Index] = w
	}

	result := make([]tmux.Window, 0, len(windows))
	seen := make(map[int]bool)

	for _, idx := range order {
		if w, ok := byIndex[idx]; ok {
			result = append(result, w)
			seen[idx] = true
		}
	}
	for _, w := range windows {
		if !seen[w.Index] {
			result = append(result, w)
		}
	}
	return result
}

// resetFilter clears filter state. The grid already contains all items
// (no filter was applied when the mode was entered), so no SetItems call needed.
func (m *Model) resetFilter() {
	m.filterMode = false
	m.filterQuery = ""
}

// ---------------------------------------------------------------------------
// Resize
// ---------------------------------------------------------------------------

func (m *Model) resize(w, h int) {
	m.width = w
	m.height = h
	m.applyLayout()
}

// layoutSizes computes the width/height for the grid and preview panel.
//
// Session view: preview ~40%, grid ~60%.
// Window view:  50-50 split so there's room to show window content.
func (m *Model) layoutSizes() (gridW, gridH, previewContentW, previewH int) {
	const gap = 1           // space between grid and preview
	const previewBorder = 4 // border(1 each side=2) + Padding(1) horizontal(1 each side=2)

	previewPct := 40
	if m.currentMode == ModeWindowGrid {
		previewPct = 50
	}

	previewRendered := m.width * previewPct / 100
	if previewRendered < previewBorder+minCardContentW {
		previewRendered = previewBorder + minCardContentW
	}
	previewContentW = previewRendered - previewBorder

	gridW = m.width - previewRendered - gap
	if gridW < 1 {
		gridW = 1
	}

	bodyH := m.height - 2 // header(1) + statusbar(1)
	if bodyH < 1 {
		bodyH = 1
	}
	gridH = bodyH
	previewH = bodyH
	return
}

// applyLayout recalculates and pushes updated sizes to all components.
// Call after any mode switch in addition to terminal resize.
func (m *Model) applyLayout() {
	gridW, gridH, previewW, previewH := m.layoutSizes()
	m.sessionGrid.SetSize(gridW, gridH)
	m.windowGrid.SetSize(gridW, gridH)
	m.previewPanel.SetSize(previewW, previewH)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (m *Model) setStatus(msg string) {
	m.statusMsg = msg
	m.statusMsgTime = time.Now()
	m.isStatusError = false
}

func (m *Model) setStatusError(msg string) {
	m.statusMsg = msg
	m.statusMsgTime = time.Now()
	m.isStatusError = true
}

// statusMessage returns the current status message, or empty if it has expired.
const statusMsgTTL = 5 * time.Second

func (m *Model) statusMessage() string {
	if m.statusMsg == "" {
		return ""
	}
	if time.Since(m.statusMsgTime) > statusMsgTTL {
		m.statusMsg = ""
		return ""
	}
	return m.statusMsg
}

func (m *Model) activeGrid() *Grid {
	if m.currentMode == ModeWindowGrid {
		return m.windowGrid
	}
	return m.sessionGrid
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
