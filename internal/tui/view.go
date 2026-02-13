package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// ---------------------------------------------------------------------------
// Top-level views
// ---------------------------------------------------------------------------

func (m *Model) renderSessionView() string {
	m.sessionGrid.SetMarks(m.buildMarkMap(true))

	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("33")).
		Bold(true).
		Render("TMUX Sessions")

	return m.renderLayout(header, m.sessionGrid.Render(), m.previewPanel.Render())
}

func (m *Model) renderWindowView() string {
	m.windowGrid.SetMarks(m.buildMarkMap(false))

	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("33")).
		Bold(true).
		Render(fmt.Sprintf("Windows in: %s", m.currentSess))

	return m.renderLayout(header, m.windowGrid.Render(), m.previewPanel.Render())
}

func (m *Model) renderHelp() string {
	help := `
TSWITCH - TMUX Navigator

Navigation:
  h/j/k/l or arrows     Move between cards
  enter                 Zoom in / Switch
  space                 Quick switch (sessions only)
  esc                   Back / Quit
  
Marks:
  m + key               Mark current session/window with key
  key (if marked)       Switch to marked session/window

Management:
  n                     New session/window (coming soon)
  r                     Rename (coming soon)
  x                     Kill (coming soon)
  t                     Tag session (coming soon)
  
Preview:
  tab                   Toggle preview mode
  
Other:
  /                     Filter (coming soon)
  ?                     Toggle help
  q                     Quit
`
	return m.styles.HelpStyle.Render(help)
}

// ---------------------------------------------------------------------------
// Layout helpers
// ---------------------------------------------------------------------------

// renderLayout composes the header + grid|preview + status bar.
func (m *Model) renderLayout(header, grid, preview string) string {
	main := lipgloss.JoinHorizontal(lipgloss.Top, grid, preview)
	return lipgloss.JoinVertical(lipgloss.Left, header, main, m.renderStatusBar())
}

func (m *Model) renderStatusBar() string {
	var text string
	switch {
	case m.markingMode:
		text = "Press a key to mark (ESC to cancel)"
	case m.lastErr != "":
		text = m.lastErr
	case m.currentMode == ModeSessionGrid:
		text = "j/k:nav  enter:select  space:quick  m:mark  tab:toggle  ?:help  q:quit"
	default:
		text = "j/k:nav  enter:switch  m:mark  esc:back  tab:toggle  ?:help  q:quit"
	}
	return m.styles.StatusBar.Render(text)
}

// ---------------------------------------------------------------------------
// Mark map builders
// ---------------------------------------------------------------------------

// buildMarkMap creates a mapping from display-names to mark keys,
// used by Grid to render mark indicators on cards.
func (m *Model) buildMarkMap(forSessions bool) map[string]string {
	mm := make(map[string]string)
	if forSessions {
		for key, mark := range m.config.Marks {
			mm[mark.SessionName] = key
		}
	} else {
		for key, mark := range m.config.Marks {
			displayName := fmt.Sprintf("%d: %s", mark.WindowIndex, m.windowName(mark.WindowIndex))
			mm[displayName] = key
		}
	}
	return mm
}

// windowName finds a window name by index in the current window list.
func (m *Model) windowName(index int) string {
	for _, w := range m.windows {
		if w.Index == index {
			return w.Name
		}
	}
	return ""
}
