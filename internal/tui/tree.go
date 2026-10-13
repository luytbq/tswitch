package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// NodeKind identifies which tmux object a tree row stands for.
type NodeKind int

const (
	NodeSession NodeKind = iota
	NodeWindow
	NodePane
)

// Tree row layout.
const (
	treeIndentW   = 2 // per depth level
	treeChevronW  = 2 // "▸ ", "▾ " or blank
	treeIndicW    = 2 // "● " or blank
	treeMetaGapW  = 2 // minimum gap between label and meta
	treeBadgeW    = 4 // "[x]" plus a trailing space, always reserved
	treeMinLabelW = 6
)

// TreeRow is one visible line of the tree. Rows are rebuilt from scratch on
// every change, so a row carries everything needed to render and act on it.
type TreeRow struct {
	Kind      NodeKind
	Depth     int
	Session   string
	Window    int // -1 for session rows
	Pane      int // -1 for session and window rows
	Label     string
	Meta      string
	Indicator string

	Expandable bool
	Expanded   bool
}

// Key returns a stable identity for the node, used for mark lookup and for
// keeping focus on the same node across rebuilds.
func (r TreeRow) Key() string {
	switch r.Kind {
	case NodeWindow:
		return fmt.Sprintf("w\x00%s\x00%d", r.Session, r.Window)
	case NodePane:
		return fmt.Sprintf("p\x00%s\x00%d\x00%d", r.Session, r.Window, r.Pane)
	default:
		return "s\x00" + r.Session
	}
}

// Tree renders a flat slice of rows as an indented, scrollable list.
type Tree struct {
	rows    []TreeRow
	focus   int
	scroll  int
	width   int
	height  int
	styles  Styles
	markMap map[string]string // TreeRow.Key() -> mark key
}

// NewTree creates a tree view sized to the given viewport.
func NewTree(width, height int, styles Styles) *Tree {
	return &Tree{
		width:   width,
		height:  height,
		styles:  styles,
		markMap: make(map[string]string),
	}
}

// SetRows replaces the visible rows, clamping focus into the new range.
func (t *Tree) SetRows(rows []TreeRow) {
	t.rows = rows
	if t.focus >= len(rows) {
		t.focus = len(rows) - 1
	}
	if t.focus < 0 {
		t.focus = 0
	}
	t.ensureVisible()
}

// SetSize updates the viewport dimensions.
func (t *Tree) SetSize(width, height int) {
	t.width = width
	t.height = height
	t.ensureVisible()
}

// SetMarks provides a mapping from TreeRow.Key() to mark key labels.
func (t *Tree) SetMarks(marks map[string]string) { t.markMap = marks }

// Rows returns the current rows.
func (t *Tree) Rows() []TreeRow { return t.rows }

// FocusIndex returns the focused row position.
func (t *Tree) FocusIndex() int { return t.focus }

// SetFocusIndex moves focus to an absolute row position.
func (t *Tree) SetFocusIndex(i int) {
	if i < 0 || i >= len(t.rows) {
		return
	}
	t.focus = i
	t.ensureVisible()
}

// Focused returns the focused row.
func (t *Tree) Focused() (TreeRow, bool) {
	if t.focus < 0 || t.focus >= len(t.rows) {
		return TreeRow{}, false
	}
	return t.rows[t.focus], true
}

// FocusRowWhere moves focus to the first row satisfying match.
func (t *Tree) FocusRowWhere(match func(TreeRow) bool) bool {
	for i, row := range t.rows {
		if match(row) {
			t.focus = i
			t.ensureVisible()
			return true
		}
	}
	return false
}

// MoveFocus moves focus by delta rows, clamping at both ends.
func (t *Tree) MoveFocus(delta int) {
	if len(t.rows) == 0 {
		return
	}
	t.focus = clamp(t.focus+delta, 0, len(t.rows)-1)
	t.ensureVisible()
}

// FocusParent moves focus to the row that owns the focused one.
func (t *Tree) FocusParent() bool {
	row, ok := t.Focused()
	if !ok || row.Depth == 0 {
		return false
	}
	for i := t.focus - 1; i >= 0; i-- {
		if t.rows[i].Depth < row.Depth {
			t.focus = i
			t.ensureVisible()
			return true
		}
	}
	return false
}

// VisibleRange returns the 1-based range of rows on screen and the total,
// for the scroll counter in the header.
func (t *Tree) VisibleRange() (first, last, total int) {
	total = len(t.rows)
	if total == 0 {
		return 0, 0, 0
	}
	first = t.scroll + 1
	last = min(t.scroll+t.visibleRows(), total)
	return first, last, total
}

// Render returns the rendered tree.
func (t *Tree) Render() string {
	if len(t.rows) == 0 {
		return t.styles.CardSubtle.Render("  No items")
	}
	t.ensureVisible()

	end := min(t.scroll+t.visibleRows(), len(t.rows))
	lines := make([]string, 0, end-t.scroll)
	for i := t.scroll; i < end; i++ {
		lines = append(lines, t.renderRow(t.rows[i], i == t.focus))
	}
	return strings.Join(lines, "\n")
}

// ---------------------------------------------------------------------------
// Private helpers
// ---------------------------------------------------------------------------

func (t *Tree) visibleRows() int {
	if t.height < 1 {
		return 1
	}
	return t.height
}

func (t *Tree) ensureVisible() {
	visible := t.visibleRows()
	if t.focus < t.scroll {
		t.scroll = t.focus
	}
	if t.focus >= t.scroll+visible {
		t.scroll = t.focus - visible + 1
	}
	maxScroll := max(0, len(t.rows)-visible)
	t.scroll = clamp(t.scroll, 0, maxScroll)
}

func (t *Tree) renderRow(row TreeRow, focused bool) string {
	s := t.styles
	bg := lipgloss.Color("236")
	style := func(base lipgloss.Style) lipgloss.Style {
		if focused {
			return base.Copy().Background(bg)
		}
		return base
	}

	labelStyle := s.CardTitle
	switch row.Kind {
	case NodeWindow:
		labelStyle = s.TreeWindow
	case NodePane:
		labelStyle = s.CardSubtle
	}

	chevron := "  "
	if row.Expandable {
		chevron = "▸ "
		if row.Expanded {
			chevron = "▾ "
		}
	}
	indicator := "  "
	if row.Indicator != "" {
		indicator = row.Indicator + " "
	}

	badge := ""
	if key, ok := t.markMap[row.Key()]; ok {
		badge = "[" + key + "]"
	}

	indent := strings.Repeat(" ", row.Depth*treeIndentW)
	prefixW := lipgloss.Width(indent) + treeChevronW + treeIndicW

	// Lay out right to left: the badge column is always reserved so badges
	// line up whether or not a row has a mark, then the meta block, then
	// whatever is left over goes to the label. A mark carrying several keys
	// widens the column rather than overflowing the row.
	badgeW := treeBadgeW
	if w := lipgloss.Width(badge) + 1; w > badgeW {
		badgeW = w
	}
	metaW := lipgloss.Width(row.Meta)
	labelW := t.width - prefixW - treeMetaGapW - metaW - badgeW
	meta := row.Meta
	if labelW < treeMinLabelW {
		// Not enough room for both: drop the meta block, keep the name.
		meta = ""
		metaW = 0
		labelW = t.width - prefixW - treeMetaGapW - badgeW
	}
	if labelW < 1 {
		labelW = 1
	}

	label := truncate(row.Label, labelW)
	gap := t.width - prefixW - lipgloss.Width(label) - metaW - badgeW
	if gap < 1 {
		gap = 1
	}

	var b strings.Builder
	b.WriteString(style(s.CardSubtle).Render(indent + chevron))
	b.WriteString(style(s.CardAttached).Render(indicator))
	b.WriteString(style(labelStyle).Render(label))
	b.WriteString(style(lipgloss.NewStyle()).Render(strings.Repeat(" ", gap)))
	b.WriteString(style(s.CardSubtle).Render(meta))
	b.WriteString(style(s.MarkBadge).Render(padLeft(badge, badgeW)))
	return b.String()
}

// truncate shortens text to at most w display cells, marking the cut with "…".
func truncate(text string, w int) string {
	if w < 1 {
		return ""
	}
	if lipgloss.Width(text) <= w {
		return text
	}
	runes := []rune(text)
	cut := len(runes)
	for cut > 0 && lipgloss.Width(string(runes[:cut]))+1 > w {
		cut--
	}
	return string(runes[:cut]) + "…"
}

// padLeft right-aligns text inside w display cells, which keeps mark badges in
// a column of their own with at least one space between them and the meta text.
func padLeft(text string, w int) string {
	pad := w - lipgloss.Width(text)
	if pad < 0 {
		return text
	}
	return strings.Repeat(" ", pad) + text
}
