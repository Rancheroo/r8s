// Package tui provides a minimal dashboard for r8s.
// Sprint 9 Week 2: Stripped down to essentials only.
package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Dashboard represents the minimal TUI dashboard
type Dashboard struct {
	bundlePath string
	width      int
	height     int
	items      []DashboardItem
	cursor     int
	quitting   bool
}

// DashboardItem represents a single item in the dashboard
type DashboardItem struct {
	Severity    string
	Title       string
	Description string
	Resource    string
}

// NewDashboard creates a new minimal dashboard
func NewDashboard(bundlePath string) (*Dashboard, error) {
	d := &Dashboard{
		bundlePath: bundlePath,
		items:      []DashboardItem{},
	}

	// Demo items for now
	d.items = []DashboardItem{
		{Severity: "critical", Title: "Demo: OOMKill Detected", Description: "Container killed due to memory limits", Resource: "pod/app-123"},
		{Severity: "warning", Title: "Demo: High Memory Usage", Description: "Memory usage above 80%", Resource: "pod/app-456"},
		{Severity: "info", Title: "Demo: Pod Restarted", Description: "Pod restarted 3 times", Resource: "pod/app-789"},
	}

	return d, nil
}

// Run starts the dashboard TUI
func (d *Dashboard) Run() error {
	p := tea.NewProgram(d, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// Init initializes the dashboard
func (d *Dashboard) Init() tea.Cmd {
	return nil
}

// Update handles keyboard input
func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			d.quitting = true
			return d, tea.Quit

		case "up", "k":
			if d.cursor > 0 {
				d.cursor--
			}

		case "down", "j":
			if d.cursor < len(d.items)-1 {
				d.cursor++
			}

		case "enter":
			// View details (simplified - just print to stdout)
			if d.cursor < len(d.items) {
				item := d.items[d.cursor]
				fmt.Printf("\n%s: %s\n%s\nResource: %s\n\n",
					item.Severity, item.Title, item.Description, item.Resource)
			}

		case "?":
			// Show help
			fmt.Println("\nHelp:")
			fmt.Println("  ↑/k - Move up")
			fmt.Println("  ↓/j - Move down")
			fmt.Println("  Enter - View details")
			fmt.Println("  q - Quit")
			fmt.Println()
		}

	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
	}

	return d, nil
}

// View renders the dashboard
func (d *Dashboard) View() string {
	if d.quitting {
		return "Goodbye!\n"
	}

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D4AA"))
	header := headerStyle.Render("═══ R8S Dashboard ═══")

	// Bundle info
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	info := ""
	if d.bundlePath != "" {
		info = infoStyle.Render(fmt.Sprintf("Bundle: %s", d.bundlePath))
	} else {
		info = infoStyle.Render("Demo Mode - No bundle loaded")
	}

	// Items list
	itemsView := "\n"
	for i, item := range d.items {
		cursor := " "
		if d.cursor == i {
			cursor = ">"
		}

		// Color by severity
		severityColor := "#888888"
		switch item.Severity {
		case "critical":
			severityColor = "#FF0000"
		case "warning":
			severityColor = "#FFAA00"
		case "info":
			severityColor = "#00AAFF"
		}

		itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(severityColor))
		itemsView += fmt.Sprintf("%s %s %s\n", cursor, itemStyle.Render("["+item.Severity+"]"), item.Title)
	}

	// Footer
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	footer := footerStyle.Render("\n↑/↓: Navigate • Enter: Details • q: Quit • ?: Help")

	return fmt.Sprintf("%s\n%s\n%s\n%s", header, info, itemsView, footer)
}

// RunDashboard is a convenience function
func RunDashboard(bundlePath string) error {
	d, err := NewDashboard(bundlePath)
	if err != nil {
		return err
	}
	return d.Run()
}
