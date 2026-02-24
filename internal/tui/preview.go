package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tswitch/internal/tmux"
)

// PreviewMode is a typed enum for the preview panel display mode.
type PreviewMode int

const (
	PreviewMetadata PreviewMode = iota
	PreviewCapture
)

// PreviewPanel displays metadata or captured content for the focused item.
type PreviewPanel struct {
	width   int
	height  int
	styles  Styles
	mode    PreviewMode
	content string
	title   string
}

// NewPreviewPanel creates a new preview panel.
func NewPreviewPanel(width, height int, styles Styles) *PreviewPanel {
	return &PreviewPanel{
		width:  width,
		height: height,
		styles: styles,
		mode:   PreviewMetadata,
		title:  "Preview",
	}
}

// SetSize updates the panel dimensions.
func (pp *PreviewPanel) SetSize(width, height int) {
	pp.width = width
	pp.height = height
}

// IsCapture reports whether the panel is in capture mode.
func (pp *PreviewPanel) IsCapture() bool { return pp.mode == PreviewCapture }

// ToggleMode switches between metadata and capture modes.
func (pp *PreviewPanel) ToggleMode() {
	if pp.mode == PreviewCapture {
		pp.mode = PreviewMetadata
	} else {
		pp.mode = PreviewCapture
	}
}

// SetSessionMetadata populates the panel for a session.
func (pp *PreviewPanel) SetSessionMetadata(session tmux.Session) {
	pp.title = "Session"

	var lines []string
	lines = append(lines, pp.styles.CardTitle.Render(session.Name))
	lines = append(lines, "")

	attached := "no"
	if session.Attached {
		attached = pp.styles.CardAttached.Render("yes")
	}

	lines = append(lines, fmt.Sprintf("Windows:     %d", session.WindowCount))
	lines = append(lines, fmt.Sprintf("Attached:    %s", attached))
	lines = append(lines, fmt.Sprintf("Created:     %s", formatTime(session.Created)))
	lines = append(lines, fmt.Sprintf("Last Active: %s", formatTime(session.LastActive)))

	pp.content = strings.Join(lines, "\n")
}

// SetWindowMetadata populates the panel for a window.
func (pp *PreviewPanel) SetWindowMetadata(window tmux.Window) {
	pp.title = "Window"

	var lines []string
	lines = append(lines, pp.styles.CardTitle.Render(fmt.Sprintf("%d: %s", window.Index, window.Name)))
	lines = append(lines, "")

	active := "no"
	if window.Active {
		active = pp.styles.CardAttached.Render("yes")
	}

	lines = append(lines, fmt.Sprintf("Panes:       %d", window.PaneCount))
	lines = append(lines, fmt.Sprintf("Active:      %s", active))
	lines = append(lines, fmt.Sprintf("Layout:      %s", window.Layout))
	if window.WorkingDir != "" {
		lines = append(lines, fmt.Sprintf("Dir:         %s", window.WorkingDir))
	}

	pp.content = strings.Join(lines, "\n")
}

// SetCaptureContent sets raw capture-pane output.
func (pp *PreviewPanel) SetCaptureContent(content string) {
	pp.content = content
}

// Render returns the rendered panel string.
//
// pp.width  = content width passed to lipgloss Width(). Rendered = pp.width + 2 (border).
// pp.height = total body height. We set lipgloss Height(pp.height - 2) so
//
//	rendered height = (pp.height - 2) + 2 = pp.height.
func (pp *PreviewPanel) Render() string {
	if pp.width < 10 {
		return ""
	}

	titleLine := pp.styles.PreviewTitle.Render(pp.title)

	body := pp.content
	if body == "" {
		body = pp.styles.CardSubtle.Render("(no content)")
	}

	// Usable lines inside the box: total height - border(2) - padding(2) - title(1) - blank after title(1).
	maxLines := pp.height - 6
	if maxLines < 1 {
		maxLines = 1
	}
	lines := strings.Split(body, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}

	// Truncate lines that exceed the panel width to prevent layout overflow.
	// Use lipgloss.Width for accurate visual width (handles wide/multi-byte chars).
	for i, line := range lines {
		if lipgloss.Width(line) > pp.width {
			runes := []rune(line)
			cut := len(runes)
			for cut > 0 && lipgloss.Width(string(runes[:cut])) > pp.width {
				cut--
			}
			lines[i] = string(runes[:cut])
		}
	}

	inner := titleLine + "\n" + strings.Join(lines, "\n")

	// lipgloss: Width/Height set content area; border adds +2 each axis.
	contentH := pp.height - 2
	if contentH < 1 {
		contentH = 1
	}

	return pp.styles.PreviewBorder.
		Copy().
		Padding(1).
		Width(pp.width).
		Height(contentH).
		Render(inner)
}

// formatTime formats a time for display, handling zero times.
func formatTime(t interface{ Format(string) string }) string {
	type zeroChecker interface {
		IsZero() bool
	}
	if z, ok := t.(zeroChecker); ok && z.IsZero() {
		return "?"
	}
	return t.Format("2006-01-02 15:04")
}
