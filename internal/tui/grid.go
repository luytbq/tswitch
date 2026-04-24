package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Grid layout constants.
const (
	minCardContentW    = 16 // minimum card content width; cards expand beyond this
	cardGap            = 2  // space between cards (right-side breathing room)
	cardBorderPadding  = 4  // border-left(1) + pad-left(1) + pad-right(1) + border-right(1)
	cardContentLines   = 2  // title line + subtitle line
	cardRenderedHeight = cardContentLines + 2 // content lines + border top/bottom
	cardRowGap         = 1  // blank lines between rows
)

// GridItem represents an item that can be displayed in a grid card.
type GridItem interface {
	// Title returns the primary display name (rendered in bold).
	Title() string
	// Subtitle returns a secondary info line (rendered dim).
	Subtitle() string
	// Indicator returns an optional short indicator (e.g. "●" for attached).
	// Empty string means no indicator.
	Indicator() string
}

// Grid manages a grid layout with auto-fit columns and keyboard focus.
type Grid struct {
	items        []GridItem
	width        int
	height       int
	focusIndex   int
	columns      int
	rows         int
	cardContentW int            // computed per-card content width
	usedWidth    int            // actual width used by card columns
	scrollOffset int
	styles       Styles
	markMap      map[string]string // item key -> mark key
}

// NewGrid creates a new grid component.
func NewGrid(width, height int, styles Styles) *Grid {
	return &Grid{
		width:   width,
		height:  height,
		styles:  styles,
		markMap: make(map[string]string),
	}
}

// SetItems replaces the grid items and resets focus/scroll.
func (g *Grid) SetItems(items []GridItem) {
	g.items = items
	g.focusIndex = 0
	g.scrollOffset = 0
	g.recalculate()
}

// SetSize updates the grid viewport dimensions.
func (g *Grid) SetSize(width, height int) {
	g.width = width
	g.height = height
	g.recalculate()
	g.ensureVisible()
}

// SetMarks provides a mapping from item titles to mark key labels.
func (g *Grid) SetMarks(marks map[string]string) {
	g.markMap = marks
}

// MoveFocus moves the focus by (dx, dy) grid cells, clamping to bounds.
func (g *Grid) MoveFocus(dx, dy int) {
	if len(g.items) == 0 {
		return
	}

	row := g.focusIndex / g.columns
	col := g.focusIndex % g.columns

	newRow := clamp(row+dy, 0, g.rows-1)
	newCol := clamp(col+dx, 0, g.columns-1)

	newIndex := newRow*g.columns + newCol
	if newIndex >= len(g.items) {
		// Target cell doesn't exist (incomplete last row).
		// Clamp to the last item on that row.
		newIndex = len(g.items) - 1
	}
	if newIndex >= 0 && newIndex != g.focusIndex {
		g.focusIndex = newIndex
		g.ensureVisible()
	}
}

// MoveItem swaps the focused item with its neighbor at offset (dx, dy)
// and moves focus to the new position. Returns true if a swap occurred.
func (g *Grid) MoveItem(dx, dy int) bool {
	if len(g.items) == 0 {
		return false
	}

	row := g.focusIndex / g.columns
	col := g.focusIndex % g.columns

	newRow := row + dy
	newCol := col + dx
	if newRow < 0 || newRow >= g.rows || newCol < 0 || newCol >= g.columns {
		return false
	}

	newIndex := newRow*g.columns + newCol
	if newIndex < 0 {
		return false
	}
	if newIndex >= len(g.items) {
		// Target cell doesn't exist (incomplete last row); clamp to last item.
		newIndex = len(g.items) - 1
	}
	if newIndex == g.focusIndex {
		return false
	}

	g.items[g.focusIndex], g.items[newIndex] = g.items[newIndex], g.items[g.focusIndex]
	g.focusIndex = newIndex
	g.ensureVisible()
	return true
}

// Items returns the current grid items slice.
func (g *Grid) Items() []GridItem {
	return g.items
}

// FocusIndex returns the current focus position.
func (g *Grid) FocusIndex() int {
	return g.focusIndex
}

// ReplaceItem replaces the item at the given position without changing focus or layout.
func (g *Grid) ReplaceItem(pos int, item GridItem) {
	if pos >= 0 && pos < len(g.items) {
		g.items[pos] = item
	}
}

// UsedWidth returns the actual width consumed by card columns.
func (g *Grid) UsedWidth() int {
	return g.usedWidth
}

// FocusFirstWhere moves focus to the first item satisfying match without
// scrolling the viewport. Does nothing if no item matches.
func (g *Grid) FocusFirstWhere(match func(GridItem) bool) {
	for i, item := range g.items {
		if match(item) {
			g.focusIndex = i
			return
		}
	}
}

// GetFocused returns the currently focused item, or nil.
func (g *Grid) GetFocused() GridItem {
	if g.focusIndex < len(g.items) {
		return g.items[g.focusIndex]
	}
	return nil
}

// Render returns the rendered grid string.
func (g *Grid) Render() string {
	if len(g.items) == 0 {
		return g.styles.CardSubtle.Render("  No items")
	}

	g.recalculate()

	effectiveRowH := cardRenderedHeight + cardRowGap
	visibleRows := g.height / effectiveRowH
	if visibleRows < 1 {
		visibleRows = 1
	}
	endRow := min(g.scrollOffset+visibleRows, g.rows)

	var lines []string
	for row := g.scrollOffset; row < endRow; row++ {
		var rowCards []string
		for col := 0; col < g.columns; col++ {
			idx := row*g.columns + col
			if idx >= len(g.items) {
				break
			}
			rowCards = append(rowCards, g.renderCard(g.items[idx], idx == g.focusIndex))
		}
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, rowCards...))
	}

	rowSep := "\n" + strings.Repeat("\n", cardRowGap)
	return strings.Join(lines, rowSep)
}

// ---------------------------------------------------------------------------
// Private helpers
// ---------------------------------------------------------------------------

func (g *Grid) recalculate() {
	minSlot := minCardContentW + cardBorderPadding + cardGap
	g.columns = max(1, g.width/minSlot)
	slotW := g.width / g.columns
	g.cardContentW = slotW - cardBorderPadding - cardGap
	if g.cardContentW < minCardContentW {
		g.cardContentW = minCardContentW
	}
	g.usedWidth = g.columns * (g.cardContentW + cardBorderPadding + cardGap)
	if len(g.items) > 0 {
		g.rows = (len(g.items) + g.columns - 1) / g.columns
	} else {
		g.rows = 0
	}
}

func (g *Grid) ensureVisible() {
	focusRow := g.focusIndex / g.columns
	effectiveRowH := cardRenderedHeight + cardRowGap
	visibleRows := g.height / effectiveRowH
	if visibleRows < 1 {
		visibleRows = 1
	}

	if focusRow < g.scrollOffset {
		g.scrollOffset = focusRow
	}
	if focusRow >= g.scrollOffset+visibleRows {
		g.scrollOffset = focusRow - visibleRows + 1
	}
}

func (g *Grid) renderCard(item GridItem, focused bool) string {
	title := item.Title()
	subtitle := item.Subtitle()
	indicator := item.Indicator()

	contentW := g.cardContentW
	if contentW < minCardContentW {
		contentW = minCardContentW
	}

	// Look up mark using the full title before truncation.
	markKey, hasMark := g.markMap[title]

	// Truncate long titles so cards stay uniform.
	maxTitleLen := contentW - 2 // leave room for indicator
	if hasMark {
		maxTitleLen = contentW - 4 - len(markKey) // room for " [x]"
	}
	if indicator != "" {
		maxTitleLen -= 2 // room for "● "
	}
	if maxTitleLen < 6 {
		maxTitleLen = 6
	}

	displayTitle := title
	if len(displayTitle) > maxTitleLen {
		displayTitle = displayTitle[:maxTitleLen-3] + "..."
	}

	// Select styles; focused cards carry the card background on all elements
	// so the highlight is continuous.
	titleStyle := g.styles.CardTitle
	subtitleStyle := g.styles.CardSubtle
	attachedStyle := g.styles.CardAttached
	markStyle := g.styles.MarkBadge
	if focused {
		bg := lipgloss.Color("236")
		titleStyle = g.styles.CardTitle.Copy().Background(bg)
		subtitleStyle = g.styles.CardSubtle.Copy().Background(bg)
		attachedStyle = g.styles.CardAttached.Copy().Background(bg)
		markStyle = g.styles.MarkBadge.Copy().Background(bg)
	}

	// Build the title line with indicator and mark badge.
	titleRendered := titleStyle.Render(displayTitle)
	if indicator != "" {
		titleRendered = attachedStyle.Render(indicator+" ") + titleRendered
	}

	// Build the full first line: title on the left, mark badge on the right.
	if hasMark {
		badge := markStyle.Render("[" + markKey + "]")
		titleWidth := lipgloss.Width(titleRendered)
		badgeWidth := lipgloss.Width(badge)
		gap := contentW - titleWidth - badgeWidth
		if gap < 1 {
			gap = 1
		}
		spacer := strings.Repeat(" ", gap)
		if focused {
			spacer = lipgloss.NewStyle().Background(lipgloss.Color("236")).Render(spacer)
		}
		titleRendered = titleRendered + spacer + badge
	}

	subtitleRendered := subtitleStyle.Render(subtitle)
	content := titleRendered + "\n" + subtitleRendered

	// Apply dynamic width: border(2) + padding(2) + contentW.
	// For focused cards, wrap content in a background-filled block (including
	// padding) so the highlight covers the entire card interior seamlessly.
	if focused {
		inner := lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Width(contentW + 2). // content + padding
			Padding(0, 1).
			Render(content)
		style := g.styles.CardFocusedStyle.Copy().Width(contentW + 2)
		return style.Render(inner)
	}
	style := g.styles.CardStyle.Copy().Width(contentW + 2)
	return style.Render(content)
}

// clamp restricts v to [lo, hi].
func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
