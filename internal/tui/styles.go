package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors (k9s-inspired)
	colorCyan     = lipgloss.Color("14")
	colorBlue     = lipgloss.Color("33")
	colorGreen    = lipgloss.Color("10")
	colorYellow   = lipgloss.Color("11")
	colorRed      = lipgloss.Color("9")
	colorGray     = lipgloss.Color("240")
	colorWhite    = lipgloss.Color("15")
	colorDarkGray = lipgloss.Color("236")

	// Base styles
	baseStyle = lipgloss.NewStyle().
			Foreground(colorWhite)

	// Header style (table headers)
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorCyan).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(colorGray)

	// Breadcrumb style
	breadcrumbStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(colorDarkGray).
			Padding(0, 1)

	// Status bar style
	statusStyle = lipgloss.NewStyle().
			Foreground(colorDarkGray).
			Background(colorCyan).
			Bold(true)

	// Error style
	errorStyle = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true).
			Padding(1, 2)

	// Loading style
	loadingStyle = lipgloss.NewStyle().
			Foreground(colorYellow).
			Bold(true).
			Padding(1, 2)

	// State colors
	stateRunning = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	statePending = lipgloss.NewStyle().
			Foreground(colorYellow).
			Bold(true)

	stateFailed = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)

	stateCompleted = lipgloss.NewStyle().
			Foreground(colorGray)
)

// GetStateStyle returns the appropriate style for a resource state
func GetStateStyle(state string) lipgloss.Style {
	switch state {
	case "active", "running", "Active", "Running":
		return stateRunning
	case "pending", "Pending", "Provisioning", "Updating":
		return statePending
	case "error", "failed", "Error", "Failed":
		return stateFailed
	case "completed", "Completed", "Succeeded":
		return stateCompleted
	default:
		return baseStyle
	}
}
