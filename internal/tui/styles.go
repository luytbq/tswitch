package tui

import "github.com/charmbracelet/lipgloss"

// Styles holds all styling for the TUI.
type Styles struct {
	// Cards
	CardStyle        lipgloss.Style
	CardFocusedStyle lipgloss.Style
	CardTitle        lipgloss.Style
	CardSubtle       lipgloss.Style
	CardAttached     lipgloss.Style
	MarkBadge        lipgloss.Style

	// Preview
	PreviewBorder lipgloss.Style
	PreviewTitle  lipgloss.Style

	// Status bar
	StatusBar     lipgloss.Style
	StatusHints   lipgloss.Style
	StatusSuccess lipgloss.Style
	StatusError   lipgloss.Style

	// Header
	HeaderStyle lipgloss.Style

	// Help
	HelpStyle   lipgloss.Style
	HelpKey     lipgloss.Style
	HelpSection lipgloss.Style
	HelpDesc    lipgloss.Style

	// Borders
	BorderStyle lipgloss.Style
}

// NewStyles creates the default Styles.
func NewStyles() Styles {
	return Styles{
		CardStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),

		CardFocusedStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")).
			Padding(0, 1),

		CardTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("33")),

		CardSubtle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")),

		CardAttached: lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true),

		MarkBadge: lipgloss.NewStyle().
			Foreground(lipgloss.Color("228")).
			Bold(true),

		PreviewBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")),

		PreviewTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("33")),

		StatusBar: lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Padding(0, 1),

		StatusHints: lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")),

		StatusSuccess: lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")),

		StatusError: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")),

		HeaderStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")).
			Bold(true).
			PaddingLeft(1),

		HelpStyle: lipgloss.NewStyle().
			Padding(1, 2),

		HelpKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")).
			Bold(true),

		HelpSection: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			MarginTop(1),

		HelpDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("248")),

		BorderStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")),
	}
}
