package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Grid layout constants.
const (
	cardContentWidth = 14 // inner content width (chars)
	cardGap          = 1  // gap between cards (chars)
	cardHeight       = 5  // approximate rendered card height (lines)
)

// cardWidthTotal is the total space one card occupies including gap.
var cardWidthTotal = cardContentWidth + 4 + cardGap // content + padding/border + gap

// GridItem represents an item that can be displayed in a grid card.
type GridItem interface {
	GetName() string
	GetMetadata() string
}

// Grid manages a grid layout with auto-fit columns and keyboard focus.
type Grid struct {
	items        []GridItem
	width        int
	height       int
	focusIndex   int
	columns      int
	rows         int
	scrollOffset int
	styles       Styles
	markMap      map[string]string // item name -> mark key
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
}

// SetMarks provides a mapping from item names to mark key labels.
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
	if newIndex < len(g.items) {
		g.focusIndex = newIndex
		g.ensureVisible()
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
		return "No items"
	}

	g.recalculate()
	g.ensureVisible()

	visibleRows := g.height / cardHeight
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

	return strings.Join(lines, "\n")
}

// ---------------------------------------------------------------------------
// Private helpers
// ---------------------------------------------------------------------------

func (g *Grid) recalculate() {
	g.columns = max(1, g.width/cardWidthTotal)
	if len(g.items) > 0 {
		g.rows = (len(g.items) + g.columns - 1) / g.columns
	} else {
		g.rows = 0
	}
}

func (g *Grid) ensureVisible() {
	focusRow := g.focusIndex / g.columns
	visibleRows := g.height / cardHeight

	if focusRow < g.scrollOffset {
		g.scrollOffset = focusRow
	}
	if focusRow >= g.scrollOffset+visibleRows {
		g.scrollOffset = focusRow - visibleRows + 1
	}
}

func (g *Grid) renderCard(item GridItem, focused bool) string {
	name := item.GetName()
	metadata := item.GetMetadata()

	// Truncate long names so cards stay uniform.
	if len(name) > 12 {
		name = name[:9] + "..."
	}

	// Build card content, optionally showing a mark indicator.
	var content string
	if mark, ok := g.markMap[name]; ok {
		content = fmt.Sprintf("%s [%s]\n%s", name, mark, metadata)
	} else {
		content = fmt.Sprintf("%s\n%s", name, metadata)
	}

	style := g.styles.CardStyle
	if focused {
		style = g.styles.CardFocusedStyle
	}
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
