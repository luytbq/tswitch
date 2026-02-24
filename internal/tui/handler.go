package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/tswitch/internal/config"
	"github.com/user/tswitch/internal/keys"
)

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
