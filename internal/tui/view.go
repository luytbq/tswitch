package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ---------------------------------------------------------------------------
// Top-level views
// ---------------------------------------------------------------------------

func (m *Model) renderSessionView() string {
	m.sessionGrid.SetMarks(m.buildMarkMap(true))

	header := m.styles.HeaderStyle.Render("Sessions")

	return m.renderLayout(header, m.sessionGrid.Render(), m.previewPanel.Render())
}

func (m *Model) renderWindowView() string {
	m.windowGrid.SetMarks(m.buildMarkMap(false))

	breadcrumb := m.styles.CardSubtle.Render("Sessions > ")
	header := m.styles.HeaderStyle.Render(breadcrumb + m.currentSess)

	return m.renderLayout(header, m.windowGrid.Render(), m.previewPanel.Render())
}

func (m *Model) renderHelp() string {
	s := m.styles
	var b strings.Builder

	b.WriteString(s.HeaderStyle.Render("TSWITCH - TMUX Session Navigator"))
	b.WriteString("\n\n")

	// Navigation
	b.WriteString(s.HelpSection.Render("Navigation"))
	b.WriteString("\n")
	writeHelpLine(&b, s, "h/j/k/l, arrows", "Move between cards")
	writeHelpLine(&b, s, "enter", "Drill into session / Switch to window")
	writeHelpLine(&b, s, "space", "Quick switch to session")
	writeHelpLine(&b, s, "esc", "Go back / Quit")

	b.WriteString("\n")
	b.WriteString(s.HelpSection.Render("Marks"))
	b.WriteString("\n")
	writeHelpLine(&b, s, "m + key", "Mark focused item with a key")
	writeHelpLine(&b, s, "key", "Jump to marked session/window")

	b.WriteString("\n")
	b.WriteString(s.HelpSection.Render("Management"))
	b.WriteString("\n")
	writeHelpLine(&b, s, "n", "New session / window")
	writeHelpLine(&b, s, "r", "Rename session / window")
	writeHelpLine(&b, s, "x", "Kill session / window")

	b.WriteString("\n")
	b.WriteString(s.HelpSection.Render("Search & UI"))
	b.WriteString("\n")
	writeHelpLine(&b, s, "/", "Search (fuzzy filter)")
	writeHelpLine(&b, s, "tab", "Toggle preview mode")
	writeHelpLine(&b, s, "?", "Toggle this help")
	writeHelpLine(&b, s, "q", "Quit")

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func writeHelpLine(b *strings.Builder, s Styles, key, desc string) {
	k := s.HelpKey.Render(fmt.Sprintf("  %-18s", key))
	d := s.HelpDesc.Render(desc)
	b.WriteString(k + d + "\n")
}

// ---------------------------------------------------------------------------
// Layout helpers
// ---------------------------------------------------------------------------

// renderLayout composes the header + grid|preview + status bar.
func (m *Model) renderLayout(header, grid, preview string) string {
	main := lipgloss.JoinHorizontal(lipgloss.Top, grid, " ", preview)
	return lipgloss.JoinVertical(lipgloss.Left, header, main, m.renderStatusBar())
}

func (m *Model) renderStatusBar() string {
	s := m.styles

	// Filter mode: show the search prompt, suppress other content.
	if m.filterMode {
		prompt := s.StatusHints.Render("/") + " " + s.StatusSuccess.Render(m.filterQuery+"█")
		hint := s.StatusHints.Render("  esc:clear  enter:keep")
		return s.StatusBar.Width(m.width).Render(prompt + hint)
	}

	// Left side: keybind hints (always shown).
	var hints string
	if m.currentMode == ModeSessionGrid {
		hints = "j/k:nav  enter:select  space:quick  /:search  n:new  r:rename  x:kill  m:mark  ?:help  q:quit"
	} else {
		hints = "j/k:nav  enter:switch  /:search  n:new  r:rename  x:kill  m:mark  esc:back  ?:help  q:quit"
	}
	left := s.StatusHints.Render(hints)

	// Right side: active filter indicator or feedback message.
	var right string
	switch {
	case m.filterQuery != "":
		right = s.StatusSuccess.Render("  /" + m.filterQuery)
	case m.markingMode:
		right = s.StatusSuccess.Render("  Press a key to assign mark (ESC to cancel)")
	case m.statusMessage() != "":
		msg := m.statusMessage()
		if m.isStatusError {
			right = s.StatusError.Render("  " + msg)
		} else {
			right = s.StatusSuccess.Render("  " + msg)
		}
	}

	bar := left + right

	// StatusBar style has Padding(0,1) which is inside Width(),
	// but no border. So rendered width = Width value exactly.
	return s.StatusBar.Width(m.width).Render(bar)
}

// ---------------------------------------------------------------------------
// Mark map builders
// ---------------------------------------------------------------------------

// buildMarkMap creates a mapping from display-names to mark keys,
// used by Grid to render mark indicators on cards.
// When multiple marks target the same item, keys are concatenated (e.g. "a,b").
func (m *Model) buildMarkMap(forSessions bool) map[string]string {
	mm := make(map[string]string)
	if forSessions {
		for key, mark := range m.config.Marks {
			if existing, ok := mm[mark.SessionName]; ok {
				mm[mark.SessionName] = existing + "," + key
			} else {
				mm[mark.SessionName] = key
			}
		}
	} else {
		for key, mark := range m.config.Marks {
			// Only show marks that belong to the currently viewed session.
			if mark.SessionName != m.currentSess {
				continue
			}
			displayName := fmt.Sprintf("%d: %s", mark.WindowIndex, m.windowName(mark.WindowIndex))
			if existing, ok := mm[displayName]; ok {
				mm[displayName] = existing + "," + key
			} else {
				mm[displayName] = key
			}
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
