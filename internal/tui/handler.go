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

// handleBack navigates backwards: window→session→quit.
func (m *Model) handleBack() (tea.Model, tea.Cmd) {
	switch {
	case m.markingMode:
		m.markingMode = false
		m.setError("")
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
			m.setError(err.Error())
		} else {
			m.currentMode = ModeWindowGrid
		}

	case ModeWindowGrid:
		card, ok := m.windowGrid.GetFocused().(WindowCard)
		if !ok {
			return m, nil
		}
		if err := m.tmux.SwitchClient(m.currentSess, card.window.Index); err != nil {
			m.setError(err.Error())
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
	if err := m.tmux.SwitchClient(card.session.Name, 0); err != nil {
		m.setError(err.Error())
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
		m.setError("Press a key to mark this session (ESC to cancel)")
	} else {
		m.markingTarget = "window"
		m.setError("Press a key to mark this window (ESC to cancel)")
	}
}

// handleMarkAssignment processes a key press while in marking mode.
func (m *Model) handleMarkAssignment(keyStr string) (tea.Model, tea.Cmd) {
	defer func() { m.markingMode = false }()

	// Cancel on ESC.
	if keyStr == "esc" {
		m.setError("")
		return m, nil
	}

	if keys.IsReserved(keyStr) {
		m.setError(fmt.Sprintf("'%s' is reserved — pick another key", keyStr))
		return m, nil
	}

	switch m.markingTarget {
	case "session":
		card, ok := m.sessionGrid.GetFocused().(SessionCard)
		if !ok {
			return m, nil
		}
		m.config.SetMark(keyStr, card.session.Name, 0, 0)
		if err := config.SaveConfig(m.config); err != nil {
			m.setError(fmt.Sprintf("Failed to save mark: %v", err))
		} else {
			m.setError(fmt.Sprintf("Marked '%s' → %s", card.session.Name, keyStr))
		}

	case "window":
		card, ok := m.windowGrid.GetFocused().(WindowCard)
		if !ok {
			return m, nil
		}
		m.config.SetMark(keyStr, m.currentSess, card.window.Index, 0)
		if err := config.SaveConfig(m.config); err != nil {
			m.setError(fmt.Sprintf("Failed to save mark: %v", err))
		} else {
			m.setError(fmt.Sprintf("Marked '%s:%d' → %s", m.currentSess, card.window.Index, keyStr))
		}
	}

	return m, nil
}

// handleJumpToMark switches to the target of a mark.
func (m *Model) handleJumpToMark(keyStr string) (tea.Model, tea.Cmd) {
	mark := m.config.GetMark(keyStr)
	if mark == nil {
		m.setError(fmt.Sprintf("no mark '%s'", keyStr))
		return m, nil
	}
	if err := m.tmux.SwitchClient(mark.SessionName, mark.WindowIndex); err != nil {
		m.setError(err.Error())
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
