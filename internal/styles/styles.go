package styles

import "github.com/charmbracelet/lipgloss"

var (
	RED               = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render
	LIGHT_BLUE        = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render
	LIGHT_YELLOW_BOLD = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")).Render
	WHITE             = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Render
)
