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
	pp.title = fmt.Sprintf("Session: %s", session.Name)
	pp.content = fmt.Sprintf(
		"Name:        %s\nWindows:     %d\nAttached:    %v\nWidth:       %d\nHeight:      %d\nCreated:     %s\nLast Active: %s",
		session.Name,
		session.WindowCount,
		session.Attached,
		session.Width,
		session.Height,
		session.Created.Format("2006-01-02 15:04:05"),
		session.LastActive.Format("2006-01-02 15:04:05"),
	)
}

// SetWindowMetadata populates the panel for a window.
func (pp *PreviewPanel) SetWindowMetadata(window tmux.Window) {
	pp.title = fmt.Sprintf("Window: %s", window.Name)
	pp.content = fmt.Sprintf(
		"Name:        %s\nIndex:       %d\nPanes:       %d\nLayout:      %s\nWorking Dir: %s\nActive:      %v",
		window.Name,
		window.Index,
		window.PaneCount,
		window.Layout,
		window.WorkingDir,
		window.Active,
	)
}

// SetCaptureContent sets raw capture-pane output.
func (pp *PreviewPanel) SetCaptureContent(content string) {
	pp.content = content
}

// Render returns the rendered panel string.
func (pp *PreviewPanel) Render() string {
	body := pp.content
	if body == "" {
		body = "(no content)"
	}

	// Truncate to fit.
	lines := strings.Split(body, "\n")
	maxLines := pp.height - 4 // borders + title padding
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	truncated := strings.Join(lines, "\n")

	inner := lipgloss.NewStyle().
		Padding(1).
		Width(pp.width - 4).
		Render(truncated)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(pp.width).
		Height(pp.height).
		Render(inner)
}
