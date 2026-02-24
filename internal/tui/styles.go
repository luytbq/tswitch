package tui

import "github.com/charmbracelet/lipgloss"

// Styles holds all styling for the TUI.
type Styles struct {
	// Cards
	CardStyle        lipgloss.Style
	CardFocusedStyle lipgloss.Style
	CardTitle        lipgloss.Style
	CardFocusedTitle lipgloss.Style
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
	StatusMode    lipgloss.Style

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
			BorderForeground(lipgloss.Color("238")).
			Padding(0, 1),

		CardFocusedStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("75")).
			Padding(0, 1),

		CardTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("75")),

		CardFocusedTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("75")).
			Background(lipgloss.Color("236")),

		CardSubtle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),

		CardAttached: lipgloss.NewStyle().
			Foreground(lipgloss.Color("114")).
			Bold(true),

		MarkBadge: lipgloss.NewStyle().
			Foreground(lipgloss.Color("222")).
			Bold(true),

		PreviewBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")),

		PreviewTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("75")),

		StatusBar: lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Padding(0, 1),

		StatusHints: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),

		StatusSuccess: lipgloss.NewStyle().
			Foreground(lipgloss.Color("114")),

		StatusError: lipgloss.NewStyle().
			Foreground(lipgloss.Color("168")),

		StatusMode: lipgloss.NewStyle().
			Foreground(lipgloss.Color("114")).
			Bold(true),

		HeaderStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("75")).
			Bold(true).
			PaddingLeft(1),

		HelpStyle: lipgloss.NewStyle().
			Padding(1, 2),

		HelpKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color("75")).
			Bold(true),

		HelpSection: lipgloss.NewStyle().
			Foreground(lipgloss.Color("114")).
			Bold(true).
			MarginTop(1),

		HelpDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")),

		BorderStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")),
	}
}
