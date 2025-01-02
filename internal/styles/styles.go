package styles

import "github.com/charmbracelet/lipgloss"

var (
	ERROR          = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render
	AGENT_MESSAGE  = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render
	AGENT_QUESTION = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")).Render
)
