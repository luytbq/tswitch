package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/luytbq/tswitch/internal/config"
	"github.com/luytbq/tswitch/internal/keys"
	"github.com/luytbq/tswitch/internal/tmux"
)

// ---------------------------------------------------------------------------
// Session / window management
// ---------------------------------------------------------------------------

type dialogAction int

const (
	dialogNone          dialogAction = iota
	dialogNewSession                 // input → tmux new-session
	dialogRenameSession              // input → tmux rename-session
	dialogKillSession                // confirm → tmux kill-session
	dialogNewWindow                  // input → tmux new-window
	dialogRenameWindow               // input → tmux rename-window
	dialogKillWindow                 // confirm → tmux kill-window
)

func (m *Model) handleNew() (tea.Model, tea.Cmd) {
	switch m.currentMode {
	case ModeSessionGrid:
		m.dialog = NewInputDialog("New Session", "Session name:", "", m.styles)
		m.pendingAction = dialogNewSession
	case ModeWindowGrid:
		m.dialog = NewInputDialog("New Window", "Window name:", "", m.styles)
		m.pendingAction = dialogNewWindow
	}
	return m, nil
}

func (m *Model) handleRename() (tea.Model, tea.Cmd) {
	switch m.currentMode {
	case ModeSessionGrid:
		card, ok := m.sessionGrid.GetFocused().(SessionCard)
		if !ok {
			return m, nil
		}
		m.dialog = NewInputDialog("Rename Session", "New name:", card.session.Name, m.styles)
		m.pendingAction = dialogRenameSession
	case ModeWindowGrid:
		card, ok := m.windowGrid.GetFocused().(WindowCard)
		if !ok {
			return m, nil
		}
		m.dialog = NewInputDialog("Rename Window", "New name:", card.window.Name, m.styles)
		m.pendingAction = dialogRenameWindow
	}
	return m, nil
}

func (m *Model) handleKill() (tea.Model, tea.Cmd) {
	switch m.currentMode {
	case ModeSessionGrid:
		card, ok := m.sessionGrid.GetFocused().(SessionCard)
		if !ok {
			return m, nil
		}
		m.dialog = NewConfirmDialog("Kill Session",
			fmt.Sprintf("Kill session %q?", card.session.Name), m.styles)
		m.pendingAction = dialogKillSession
	case ModeWindowGrid:
		card, ok := m.windowGrid.GetFocused().(WindowCard)
		if !ok {
			return m, nil
		}
		m.dialog = NewConfirmDialog("Kill Window",
			fmt.Sprintf("Kill window %q?", card.window.Name), m.styles)
		m.pendingAction = dialogKillWindow
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// Dialog key handling
// ---------------------------------------------------------------------------

func (m *Model) handleDialogKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	d := m.dialog
	switch d.Kind {
	case DialogInput:
		switch msg.String() {
		case "esc":
			m.dialog = nil
		case "enter":
			return m.submitDialog()
		case "backspace":
			if len(d.Input) > 0 {
				runes := []rune(d.Input)
				d.Input = string(runes[:len(runes)-1])
			}
		default:
			if msg.Type == tea.KeyRunes {
				d.Input += msg.String()
			}
		}

	case DialogConfirm:
		switch msg.String() {
		case "esc":
			m.dialog = nil
		case "y":
			d.SelectedIdx = 0
			return m.submitDialog()
		case "n":
			m.dialog = nil
		case "enter":
			return m.submitDialog()
		case "left", "h", "right", "l":
			if len(d.Options) == 2 {
				d.SelectedIdx = 1 - d.SelectedIdx
			}
		}
	}
	return m, nil
}

func (m *Model) submitDialog() (tea.Model, tea.Cmd) {
	d := m.dialog
	action := m.pendingAction
	m.dialog = nil
	m.pendingAction = dialogNone

	switch action {
	case dialogNewSession:
		name := strings.TrimSpace(d.Input)
		if name == "" {
			m.setStatusError("session name cannot be empty")
			return m, nil
		}
		if err := m.tmux.NewSession(name); err != nil {
			m.setStatusError(err.Error())
		} else {
			m.setStatus("Created: " + name)
			_ = m.loadSessions()
			m.applyFilter()
			m.sessionGrid.FocusFirstWhere(func(item GridItem) bool { return item.Title() == name })
			return m, m.syncPreview()
		}

	case dialogRenameSession:
		card, ok := m.sessionGrid.GetFocused().(SessionCard)
		if !ok {
			return m, nil
		}
		name := strings.TrimSpace(d.Input)
		if name == "" {
			m.setStatusError("session name cannot be empty")
			return m, nil
		}
		if err := m.tmux.RenameSession(card.session.Name, name); err != nil {
			m.setStatusError(err.Error())
		} else {
			m.setStatus("Renamed to: " + name)
			_ = m.loadSessions()
			m.applyFilter()
			return m, m.syncPreview()
		}

	case dialogKillSession:
		if d.SelectedIdx != 0 {
			return m, nil // "No" selected
		}
		card, ok := m.sessionGrid.GetFocused().(SessionCard)
		if !ok {
			return m, nil
		}
		name := card.session.Name
		if err := m.tmux.KillSession(name); err != nil {
			m.setStatusError(err.Error())
		} else {
			m.setStatus("Killed: " + name)
			_ = m.loadSessions()
			m.applyFilter()
			return m, m.syncPreview()
		}

	case dialogNewWindow:
		name := strings.TrimSpace(d.Input)
		if name == "" {
			m.setStatusError("window name cannot be empty")
			return m, nil
		}
		if err := m.tmux.NewWindow(m.currentSess, name); err != nil {
			m.setStatusError(err.Error())
		} else {
			m.setStatus("Created: " + name)
			_ = m.loadWindows(m.currentSess)
			m.applyFilter()
			m.windowGrid.FocusFirstWhere(func(item GridItem) bool { return item.Title() == name })
			return m, m.syncPreview()
		}

	case dialogRenameWindow:
		card, ok := m.windowGrid.GetFocused().(WindowCard)
		if !ok {
			return m, nil
		}
		name := strings.TrimSpace(d.Input)
		if name == "" {
			m.setStatusError("window name cannot be empty")
			return m, nil
		}
		if err := m.tmux.RenameWindow(m.currentSess, card.window.Index, name); err != nil {
			m.setStatusError(err.Error())
		} else {
			m.setStatus("Renamed to: " + name)
			_ = m.loadWindows(m.currentSess)
			m.applyFilter()
			return m, m.syncPreview()
		}

	case dialogKillWindow:
		if d.SelectedIdx != 0 {
			return m, nil // "No" selected
		}
		card, ok := m.windowGrid.GetFocused().(WindowCard)
		if !ok {
			return m, nil
		}
		winName := card.window.Name
		if err := m.tmux.KillWindow(m.currentSess, card.window.Index); err != nil {
			m.setStatusError(err.Error())
		} else {
			m.setStatus("Killed: " + winName)
			_ = m.loadWindows(m.currentSess)
			m.applyFilter()
			return m, m.syncPreview()
		}
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// Filter / Search
// ---------------------------------------------------------------------------

func (m *Model) enterFilterMode() tea.Cmd {
	m.filterMode = true
	m.filterQuery = ""
	m.applyFilter()
	return m.syncPreview()
}

func (m *Model) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.filterMode = false
		m.filterQuery = ""
		m.applyFilter()
		return m, m.syncPreview()
	case "enter":
		m.filterMode = false
	case "backspace":
		if len(m.filterQuery) > 0 {
			runes := []rune(m.filterQuery)
			m.filterQuery = string(runes[:len(runes)-1])
			m.applyFilter()
			return m, m.syncPreview()
		}
	default:
		if msg.Type == tea.KeyRunes {
			m.filterQuery += msg.String()
			m.applyFilter()
			return m, m.syncPreview()
		}
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// Navigation
// ---------------------------------------------------------------------------

// handleBack navigates backwards: window->session->quit.
func (m *Model) handleBack() (tea.Model, tea.Cmd) {
	switch {
	case m.markingMode:
		m.markingMode = false
		m.setStatus("")
	case m.helpShown:
		m.helpShown = false
	case m.currentMode == ModeWindowGrid:
		m.resetFilter()
		m.currentMode = ModeSessionGrid
		m.applyLayout()
		return m, m.syncPreview()
	default:
		return m, tea.Quit
	}
	return m, nil
}

// handleConfirm drills into a session or switches to a window.
func (m *Model) handleConfirm() (tea.Model, tea.Cmd) {
	switch m.currentMode {
	case ModeSessionGrid:
		card, ok := m.sessionGrid.GetFocused().(SessionCard)
		if !ok {
			return m, nil
		}
		if err := m.loadWindows(card.session.Name); err != nil {
			m.setStatusError(err.Error())
		} else {
			m.resetFilter()
			m.currentMode = ModeWindowGrid
			m.applyLayout()
			return m, m.syncPreview()
		}

	case ModeWindowGrid:
		card, ok := m.windowGrid.GetFocused().(WindowCard)
		if !ok {
			return m, nil
		}
		if err := m.tmux.SwitchClient(m.currentSess, card.window.Index); err != nil {
			m.setStatusError(err.Error())
		} else {
			return m, tea.Quit
		}
	}
	return m, nil
}

// handleQuickSwap quick-switches to a session's active window.
func (m *Model) handleQuickSwap() (tea.Model, tea.Cmd) {
	if m.currentMode != ModeSessionGrid {
		return m, nil
	}
	card, ok := m.sessionGrid.GetFocused().(SessionCard)
	if !ok || card.session.WindowCount == 0 {
		return m, nil
	}

	if err := m.tmux.SwitchToSession(card.session.Name); err != nil {
		m.setStatusError(err.Error())
		return m, nil
	}
	return m, tea.Quit
}

// ---------------------------------------------------------------------------
// Marks
// ---------------------------------------------------------------------------

// enterMarkingMode puts the model into mark-assignment mode.
func (m *Model) enterMarkingMode() {
	m.markingMode = true
	if m.currentMode == ModeSessionGrid {
		m.markingTarget = "session"
	} else {
		m.markingTarget = "window"
	}
}

// handleMarkAssignment processes a key press while in marking mode.
func (m *Model) handleMarkAssignment(keyStr string) (tea.Model, tea.Cmd) {
	defer func() { m.markingMode = false }()

	// Cancel on ESC.
	if keyStr == "esc" {
		m.setStatus("")
		return m, nil
	}

	if keys.IsReserved(keyStr) {
		m.setStatusError(fmt.Sprintf("'%s' is reserved", keyStr))
		return m, nil
	}

	switch m.markingTarget {
	case "session":
		card, ok := m.sessionGrid.GetFocused().(SessionCard)
		if !ok {
			return m, nil
		}
		// Remove any existing mark pointing to the same target before setting new one.
		m.config.RemoveMarksForTarget(card.session.Name, -1)
		m.config.SetMark(keyStr, card.session.Name, -1, 0)
		if err := config.SaveState(m.config); err != nil {
			m.setStatusError(fmt.Sprintf("Failed to save: %v", err))
		} else {
			m.setStatus(fmt.Sprintf("Marked %s -> [%s]", card.session.Name, keyStr))
		}

	case "window":
		card, ok := m.windowGrid.GetFocused().(WindowCard)
		if !ok {
			return m, nil
		}
		m.config.RemoveMarksForTarget(m.currentSess, card.window.Index)
		m.config.SetMark(keyStr, m.currentSess, card.window.Index, 0)
		if err := config.SaveState(m.config); err != nil {
			m.setStatusError(fmt.Sprintf("Failed to save: %v", err))
		} else {
			m.setStatus(fmt.Sprintf("Marked %s:%d -> [%s]", m.currentSess, card.window.Index, keyStr))
		}
	}

	return m, nil
}

// handleJumpToMark switches to the target of a mark.
func (m *Model) handleJumpToMark(keyStr string) (tea.Model, tea.Cmd) {
	mark := m.config.GetMark(keyStr)
	if mark == nil {
		m.setStatusError(fmt.Sprintf("no mark '%s'", keyStr))
		return m, nil
	}

	var err error
	if mark.WindowIndex < 0 {
		// Session-level mark: let tmux pick the active window.
		err = m.tmux.SwitchToSession(mark.SessionName)
	} else {
		err = m.tmux.SwitchClient(mark.SessionName, mark.WindowIndex)
	}
	if err != nil {
		m.setStatusError(err.Error())
		return m, nil
	}
	return m, tea.Quit
}

// ---------------------------------------------------------------------------
// Preview sync
// ---------------------------------------------------------------------------

// syncPreview updates the preview panel for the currently focused item.
// In capture mode it clears stale content and returns a tea.Cmd that fetches
// the pane capture asynchronously; in metadata mode it updates synchronously
// and returns nil.
func (m *Model) syncPreview() tea.Cmd {
	if m.previewPanel.IsCapture() {
		m.previewPanel.SetCaptureContent("") // clear stale content
		return m.fetchCapture()
	}
	switch m.currentMode {
	case ModeSessionGrid:
		if card, ok := m.sessionGrid.GetFocused().(SessionCard); ok {
			m.previewPanel.SetSessionMetadata(card.session)
		}
	case ModeWindowGrid:
		if card, ok := m.windowGrid.GetFocused().(WindowCard); ok {
			m.previewPanel.SetWindowMetadata(card.window)
		}
	}
	return nil
}

// fetchCapture returns a Cmd that runs tmux capture-pane for the focused item
// in a goroutine and delivers the result as a captureResultMsg.
func (m *Model) fetchCapture() tea.Cmd {
	var sessName string
	winIdx := -1 // -1 = active window of the session

	switch m.currentMode {
	case ModeSessionGrid:
		card, ok := m.sessionGrid.GetFocused().(SessionCard)
		if !ok {
			return nil
		}
		sessName = card.session.Name
	case ModeWindowGrid:
		card, ok := m.windowGrid.GetFocused().(WindowCard)
		if !ok {
			return nil
		}
		sessName = m.currentSess
		winIdx = card.window.Index
	default:
		return nil
	}

	return func() tea.Msg {
		content, err := m.tmux.CapturePane(sessName, winIdx, 0)
		if err != nil {
			return captureResultMsg{"(capture error: " + err.Error() + ")"}
		}
		// tmux pads every line with spaces to the full pane width and may
		// include \r; strip both so the preview box doesn't overflow.
		content = strings.ReplaceAll(content, "\r", "")
		rawLines := strings.Split(content, "\n")
		for i, l := range rawLines {
			rawLines[i] = strings.TrimRight(l, " ")
		}
		return captureResultMsg{strings.Join(rawLines, "\n")}
	}
}

// moveFocus moves the active grid's focus and syncs the preview.
func (m *Model) moveFocus(dx, dy int) tea.Cmd {
	m.activeGrid().MoveFocus(dx, dy)
	return m.syncPreview()
}

// handleReorder swaps the focused item with its neighbor and persists the new order.
func (m *Model) handleReorder(dx, dy int) (tea.Model, tea.Cmd) {
	grid := m.activeGrid()
	if !grid.MoveItem(dx, dy) {
		return m, nil
	}

	// Extract and persist the new order.
	switch m.currentMode {
	case ModeSessionGrid:
		items := grid.Items()
		order := make([]string, len(items))
		for i, item := range items {
			order[i] = item.(SessionCard).session.Name
		}
		m.config.SetSessionOrder(order)
		// Update m.sessions to match new order.
		m.sessions = make([]tmux.Session, len(items))
		for i, item := range items {
			m.sessions[i] = item.(SessionCard).session
		}

	case ModeWindowGrid:
		items := grid.Items()
		indices := make([]int, len(items))
		for i, item := range items {
			indices[i] = item.(WindowCard).window.Index
		}
		m.config.SetWindowOrder(m.currentSess, indices)
		// Update m.windows to match new order.
		m.windows = make([]tmux.Window, len(items))
		for i, item := range items {
			m.windows[i] = item.(WindowCard).window
		}
	}

	if err := config.SaveState(m.config); err != nil {
		m.setStatusError(err.Error())
	}
	return m, m.syncPreview()
}
