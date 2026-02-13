package tui

import "github.com/charmbracelet/lipgloss"

// Styles holds all styling for the TUI.
type Styles struct {
	// Cards
	CardStyle        lipgloss.Style
	CardFocusedStyle lipgloss.Style
	CardTitle        lipgloss.Style
	CardSubtle       lipgloss.Style

	// Preview
	PreviewStyle lipgloss.Style
	PreviewTitle lipgloss.Style

	// Filter
	FilterStyle       lipgloss.Style
	FilterPlaceholder lipgloss.Style

	// Status bar
	StatusBar    lipgloss.Style
	StatusBarKey lipgloss.Style

	// Help
	HelpStyle lipgloss.Style
	HelpKey   lipgloss.Style

	// Borders
	BorderStyle lipgloss.Style
}

// NewStyles creates the default Styles.
func NewStyles() Styles {
	return Styles{
		CardStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1).
			Width(cardContentWidth),

		CardFocusedStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")).
			Padding(0, 1).
			Width(cardContentWidth),

		CardTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("33")),

		CardSubtle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")),

		PreviewStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1),

		PreviewTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("33")),

		FilterStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),

		FilterPlaceholder: lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")),

		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Background(lipgloss.Color("235")).
			Padding(0, 1),

		StatusBarKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")).
			Bold(true),

		HelpStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("248")).
			Padding(1),

		HelpKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")).
			Bold(true),

		BorderStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")),
	}
}
