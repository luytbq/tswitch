package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/tswitch/internal/config"
	"github.com/user/tswitch/internal/keys"
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
			m.syncPreview()
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
			m.syncPreview()
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
		}
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// Filter / Search
// ---------------------------------------------------------------------------

func (m *Model) enterFilterMode() {
	m.filterMode = true
	m.filterQuery = ""
	m.applyFilter()
}

func (m *Model) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.filterMode = false
		m.filterQuery = ""
		m.applyFilter()
	case "enter":
		m.filterMode = false
	case "backspace":
		if len(m.filterQuery) > 0 {
			runes := []rune(m.filterQuery)
			m.filterQuery = string(runes[:len(runes)-1])
			m.applyFilter()
		}
	default:
		if msg.Type == tea.KeyRunes {
			m.filterQuery += msg.String()
			m.applyFilter()
		}
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// Navigation
// ---------------------------------------------------------------------------

// moveFocus moves the active grid's focus and updates the preview panel.
func (m *Model) moveFocus(dx, dy int) {
	grid := m.activeGrid()
	grid.MoveFocus(dx, dy)
	m.syncPreview()
}

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
		m.applyFilter()
		m.currentMode = ModeSessionGrid
		m.syncPreview()
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
		if err := config.SaveConfig(m.config); err != nil {
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
		if err := config.SaveConfig(m.config); err != nil {
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

// syncPreview updates the preview panel to match the currently focused item.
func (m *Model) syncPreview() {
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
}
