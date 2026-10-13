package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/luytbq/tswitch/internal/config"
)

// ---------------------------------------------------------------------------
// Top-level views
// ---------------------------------------------------------------------------

func (m *Model) renderSessionView() string {
	m.sessionGrid.SetMarks(m.buildMarkMap(true))

	header := m.renderHeader("Sessions"+m.countSuffix(len(m.sessionGrid.Items())), "")
	separator := m.styles.CardSubtle.Render(strings.Repeat("─", m.width))

	return m.renderLayout(header, separator, m.sessionGrid.Render(), m.previewPanel.Render())
}

func (m *Model) renderListView() string {
	m.tree.SetMarks(m.buildTreeMarkMap())

	sessionCount := 0
	for _, row := range m.tree.Rows() {
		if row.Kind == NodeSession {
			sessionCount++
		}
	}

	var scroll string
	if first, last, total := m.tree.VisibleRange(); total > last-first+1 {
		scroll = fmt.Sprintf("%d-%d/%d", first, last, total)
	}

	header := m.renderHeader("Sessions"+m.countSuffix(sessionCount), scroll)
	separator := m.styles.CardSubtle.Render(strings.Repeat("─", m.width))

	return m.renderLayout(header, separator, m.tree.Render(), m.previewPanel.Render())
}

// countSuffix renders "(12)", or "(4/29)" while a filter is narrowing the list.
func (m *Model) countSuffix(shown int) string {
	if m.filterQuery != "" && shown != len(m.sessions) {
		return fmt.Sprintf(" (%d/%d)", shown, len(m.sessions))
	}
	return fmt.Sprintf(" (%d)", shown)
}

// renderHeader lays out the title, the sort and view chips, and an optional
// right-aligned note such as the scroll position.
func (m *Model) renderHeader(title, right string) string {
	s := m.styles
	view := config.ViewGrid
	if m.isListView() {
		view = config.ViewList
	}
	chips := fmt.Sprintf("  ·  sort: %s  ·  %s", m.config.Settings.SortBy, view)

	left := s.HeaderStyle.Render(title) + s.CardSubtle.Render(chips)
	if right == "" {
		return left
	}
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 1
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + s.CardSubtle.Render(right)
}

func (m *Model) renderWindowView() string {
	m.windowGrid.SetMarks(m.buildMarkMap(false))

	count := len(m.windowGrid.Items())
	header := m.styles.HeaderStyle.Render(fmt.Sprintf("Sessions › %s (%d)", m.currentSess, count))
	separator := m.styles.CardSubtle.Render(strings.Repeat("─", m.width))

	return m.renderLayout(header, separator, m.windowGrid.Render(), m.previewPanel.Render())
}

func (m *Model) renderPaneView() string {
	count := len(m.paneGrid.Items())
	winName := m.windowName(m.currentWin)
	header := m.styles.HeaderStyle.Render(fmt.Sprintf("Sessions › %s › %s (%d)", m.currentSess, winName, count))
	separator := m.styles.CardSubtle.Render(strings.Repeat("─", m.width))

	return m.renderLayout(header, separator, m.paneGrid.Render(), m.previewPanel.Render())
}

func (m *Model) renderHelp() string {
	s := m.styles
	var b strings.Builder

	b.WriteString(s.HeaderStyle.Render("TSWITCH - TMUX Session Navigator"))
	b.WriteString("\n\n")

	// Navigation
	b.WriteString(s.HelpSection.Render("Navigation"))
	b.WriteString("\n")
	writeHelpLine(&b, s, "h/j/k/l, arrows", "Move between items")
	writeHelpLine(&b, s, "o", "Grid: drill in · List: fold/unfold")
	writeHelpLine(&b, s, "enter, space", "Switch to focused item")
	writeHelpLine(&b, s, "esc", "Go back / Fold / Quit")

	b.WriteString("\n")
	b.WriteString(s.HelpSection.Render("View"))
	b.WriteString("\n")
	writeHelpLine(&b, s, "v", "Toggle card grid / tree list")
	writeHelpLine(&b, s, "s", "Cycle sort: recent, name, manual")
	writeHelpLine(&b, s, "H/L", "List: fold / unfold all sessions")

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
	writeHelpLine(&b, s, "d", "Kill session / window")
	writeHelpLine(&b, s, "x", "Cut window/pane (toggle to clear)")
	writeHelpLine(&b, s, "p", "Paste cut window/pane onto focus")
	writeHelpLine(&b, s, "H/J/K/L", "Reorder items (sort: manual)")
	writeHelpLine(&b, s, "f", "Browse dirs (fzf)")

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

// renderLayout composes the header + separator + grid|preview + status bar.
// The grid is padded to the actual width used by cards (not the full
// allocated width) so the preview sits snugly next to the grid.
// The final output is clamped to the terminal height to prevent overflow.
func (m *Model) renderLayout(header, separator, grid, preview string) string {
	usedW := m.mainUsedWidth()
	paddedGrid := lipgloss.NewStyle().Width(usedW).Render(grid)
	main := lipgloss.JoinHorizontal(lipgloss.Top, paddedGrid, " ", preview)

	var full string
	if banner := m.renderClipboardBanner(); banner != "" {
		full = lipgloss.JoinVertical(lipgloss.Left, header, banner, separator, main, m.renderStatusBar())
	} else {
		full = lipgloss.JoinVertical(lipgloss.Left, header, separator, main, m.renderStatusBar())
	}

	// Clamp output to terminal height so long preview content doesn't push
	// the grid off-screen.
	lines := strings.Split(full, "\n")
	if len(lines) > m.height {
		lines = lines[:m.height]
	}
	return strings.Join(lines, "\n")
}

// renderClipboardBanner returns a one-line banner shown above the grid when
// a cut is pending. Returns "" when the clipboard is empty.
func (m *Model) renderClipboardBanner() string {
	if m.clipboard == nil {
		return ""
	}
	s := m.styles
	text := fmt.Sprintf("[CUT] %s  ·  p=paste  x=clear", m.clipboard.label)
	return s.StatusSuccess.Render(text)
}

func (m *Model) renderStatusBar() string {
	s := m.styles

	// Filter mode: show the search prompt, suppress other content.
	if m.filterMode {
		prompt := s.StatusHints.Render("/") + " " + s.StatusSuccess.Render(m.filterQuery+"█")
		hint := s.StatusHints.Render("  esc:clear  enter:keep")
		return s.StatusBar.Width(m.width).Render(prompt + hint)
	}

	// Left side: mode label + keybind hints.
	var modeLabel, hints string
	if m.isListView() {
		modeLabel = s.StatusMode.Render("TREE")
		hints = " jk:nav  hl:fold  o:fold/switch  enter:switch  JK:move  HL:fold all  s:sort  v:grid  /:search  n:new  r:rename  d:kill  x:cut  p:paste  m:mark  ?:help  q:quit"
		return m.statusLine(modeLabel, hints)
	}
	switch m.currentMode {
	case ModeSessionGrid:
		modeLabel = s.StatusMode.Render("SESSIONS")
		hints = " hjkl/HJKL:nav/reorder  o:open  enter/space:switch  tab:preview  /:search  n:new  r:rename  d:kill  p:paste  m:mark  f:browse  ?:help  q:quit"
	case ModeWindowGrid:
		modeLabel = s.StatusMode.Render("WINDOWS")
		hints = " hjkl/HJKL:nav/reorder  o:open  enter/space:switch  tab:preview  /:search  n:new  r:rename  d:kill  x:cut  p:paste  m:mark  esc:back  ?:help  q:quit"
	case ModePaneGrid:
		modeLabel = s.StatusMode.Render("PANES")
		hints = " hjkl/HJKL:nav/reorder  enter/space:switch  tab:preview  /:search  x:cut  p:paste  m:mark  esc:back  ?:help  q:quit"
	}
	return m.statusLine(modeLabel, hints)
}

// statusLine fits the key hints and the feedback message onto one row. The
// hints are cut first: the message is the part the user is waiting for, and
// the hint list is long enough to push it off screen on a wide terminal.
func (m *Model) statusLine(modeLabel, hints string) string {
	s := m.styles
	right := m.statusRight()

	// StatusBar style has Padding(0,1) which is inside Width(), but no border,
	// so the usable room is the terminal width less that padding.
	avail := m.width - 2 - lipgloss.Width(modeLabel) - lipgloss.Width(right)
	if lipgloss.Width(hints) > avail {
		hints = truncate(hints, avail)
	}
	return s.StatusBar.Width(m.width).Render(modeLabel + s.StatusHints.Render(hints) + right)
}

// statusRight renders the trailing part of the status bar: the active filter,
// the marking prompt, or the most recent feedback message.
func (m *Model) statusRight() string {
	s := m.styles
	switch {
	case m.filterQuery != "":
		return s.StatusSuccess.Render("  /" + m.filterQuery)
	case m.markingMode:
		return s.StatusSuccess.Render("  Press a key to assign mark (ESC to cancel)")
	case m.statusMessage() != "":
		if m.isStatusError {
			return s.StatusError.Render("  " + m.statusMessage())
		}
		return s.StatusSuccess.Render("  " + m.statusMessage())
	}
	return ""
}

// ---------------------------------------------------------------------------
// Mark map builders
// ---------------------------------------------------------------------------

// buildTreeMarkMap maps TreeRow keys to mark key labels.
func (m *Model) buildTreeMarkMap() map[string]string {
	mm := make(map[string]string)
	for key, mark := range m.config.Marks {
		row := TreeRow{Kind: NodeSession, Session: mark.SessionName}
		switch {
		case mark.WindowIndex < 0:
		case mark.PaneIndex > 0:
			row.Kind = NodePane
			row.Window = mark.WindowIndex
			row.Pane = mark.PaneIndex
		default:
			row.Kind = NodeWindow
			row.Window = mark.WindowIndex
		}
		if existing, ok := mm[row.Key()]; ok {
			mm[row.Key()] = existing + "," + key
		} else {
			mm[row.Key()] = key
		}
	}
	return mm
}

// buildMarkMap creates a mapping from display-names to mark keys,
// used by Grid to render mark indicators on cards.
// When multiple marks target the same item, keys are concatenated (e.g. "a,b").
func (m *Model) buildMarkMap(forSessions bool) map[string]string {
	mm := make(map[string]string)
	for key, mark := range m.config.Marks {
		var displayKey string
		if forSessions {
			displayKey = mark.SessionName
		} else {
			if mark.SessionName != m.currentSess {
				continue
			}
			displayKey = fmt.Sprintf("%d: %s", mark.WindowIndex, m.windowName(mark.WindowIndex))
		}
		if existing, ok := mm[displayKey]; ok {
			mm[displayKey] = existing + "," + key
		} else {
			mm[displayKey] = key
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
