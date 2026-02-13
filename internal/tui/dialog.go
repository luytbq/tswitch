package tui

import "github.com/charmbracelet/lipgloss"

// DialogKind distinguishes confirmation from input dialogs.
type DialogKind int

const (
	DialogConfirm DialogKind = iota
	DialogInput
)

// Dialog represents a modal overlay (confirm or text-input).
type Dialog struct {
	Kind        DialogKind
	Title       string
	Message     string
	Input       string
	Options     []string
	SelectedIdx int
	styles      Styles
}

// NewConfirmDialog creates a yes/no confirmation dialog.
func NewConfirmDialog(title, message string, styles Styles) *Dialog {
	return &Dialog{
		Kind:        DialogConfirm,
		Title:       title,
		Message:     message,
		Options:     []string{"Yes", "No"},
		SelectedIdx: 1, // default to "No"
		styles:      styles,
	}
}

// NewInputDialog creates a text-input dialog.
func NewInputDialog(title, message, defaultValue string, styles Styles) *Dialog {
	return &Dialog{
		Kind:    DialogInput,
		Title:   title,
		Message: message,
		Input:   defaultValue,
		styles:  styles,
	}
}

// Render returns the dialog overlay string.
func (d *Dialog) Render(width, height int) string {
	const dialogWidth = 40

	var body string
	switch d.Kind {
	case DialogConfirm:
		opts := ""
		for i, opt := range d.Options {
			s := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			if i == d.SelectedIdx {
				s = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true).Underline(true)
			}
			opts += " " + s.Render(opt)
		}
		body = d.Message + "\n\n" + opts

	case DialogInput:
		body = d.Message + "\n\n> " + d.Input
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1).
		Width(dialogWidth).
		Render(body)
}
