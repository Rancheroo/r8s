package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"

	"github.com/4realtech/r9s/internal/config"
	"github.com/4realtech/r9s/internal/rancher"
)

// App represents the main TUI application
type App struct {
	config  *config.Config
	client  *rancher.Client
	width   int
	height  int
	
	// Current view state
	clusters []rancher.Cluster
	table    table.Model
	error    string
	loading  bool
}

// NewApp creates a new TUI application
func NewApp(cfg *config.Config) *App {
	// Get current profile
	profile, err := cfg.GetCurrentProfile()
	if err != nil {
		return &App{
			config: cfg,
			error:  fmt.Sprintf("Failed to get profile: %v", err),
		}
	}
	
	// Create Rancher client
	client := rancher.NewClient(
		profile.URL,
		profile.GetToken(),
		cfg.Insecure || profile.Insecure,
	)
	
	// Test connection
	if err := client.TestConnection(); err != nil {
		return &App{
			config: cfg,
			client: client,
			error:  fmt.Sprintf("Failed to connect to Rancher: %v", err),
		}
	}
	
	return &App{
		config:  cfg,
		client:  client,
		loading: true,
	}
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.fetchClusters(),
		tea.EnterAltScreen,
	)
}

// Update handles messages and updates the model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case "r", "ctrl+r":
			a.loading = true
			return a, a.fetchClusters()
		}
		
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.updateTable()
		
	case clustersMsg:
		a.loading = false
		a.clusters = msg.clusters
		a.error = ""
		a.updateTable()
		
	case errMsg:
		a.loading = false
		a.error = msg.Error()
	}
	
	// Update table
	newTable, cmd := a.table.Update(msg)
	a.table = newTable
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	return a, tea.Batch(cmds...)
}

// View renders the application
func (a *App) View() string {
	if a.error != "" {
		return errorStyle.Render(fmt.Sprintf("Error: %s\n\nPress 'q' to quit", a.error))
	}
	
	if a.loading {
		return loadingStyle.Render("Loading clusters...")
	}
	
	// Build view components
	breadcrumb := breadcrumbStyle.Render("Rancher Clusters")
	status := statusStyle.Render(fmt.Sprintf(" %d clusters | Press 'q' to quit | 'r' to refresh ", len(a.clusters)))
	
	// Calculate content height
	contentHeight := a.height - lipgloss.Height(breadcrumb) - lipgloss.Height(status) - 2
	
	// Render table with proper height
	tableView := a.table.View()
	
	// Join all components
	return lipgloss.JoinVertical(
		lipgloss.Left,
		breadcrumb,
		"",
		tableView,
		"",
		status,
	)
}

// updateTable updates the table with current cluster data
func (a *App) updateTable() {
	if len(a.clusters) == 0 {
		return
	}
	
	// Define columns
	columns := []table.Column{
		table.NewColumn("name", "NAME", 30),
		table.NewColumn("state", "STATE", 15),
		table.NewColumn("version", "VERSION", 15),
		table.NewColumn("provider", "PROVIDER", 15),
		table.NewColumn("age", "AGE", 10),
	}
	
	// Build rows
	rows := []table.Row{}
	for _, cluster := range a.clusters {
		age := formatAge(cluster.Created)
		
		rows = append(rows, table.NewRow(table.RowData{
			"name":     cluster.Name,
			"state":    cluster.State,
			"version":  cluster.Version,
			"provider": cluster.Provider,
			"age":      age,
		}))
	}
	
	// Create or update table
	a.table = table.New(columns).
		WithRows(rows).
		HeaderStyle(headerStyle).
		WithBaseStyle(baseStyle).
		WithPageSize(a.height - 8).
		Focused(true).
		BorderRounded()
}

// fetchClusters fetches clusters from Rancher
func (a *App) fetchClusters() tea.Cmd {
	return func() tea.Msg {
		if a.client == nil {
			return errMsg{fmt.Errorf("client not initialized")}
		}
		
		collection, err := a.client.ListClusters()
		if err != nil {
			return errMsg{err}
		}
		
		return clustersMsg{clusters: collection.Data}
	}
}

// Messages
type clustersMsg struct {
	clusters []rancher.Cluster
}

type errMsg struct {
	error
}

// formatAge formats a time.Time into a human-readable age string
func formatAge(t time.Time) string {
	age := time.Since(t)
	
	if age < time.Minute {
		return fmt.Sprintf("%ds", int(age.Seconds()))
	}
	if age < time.Hour {
		return fmt.Sprintf("%dm", int(age.Minutes()))
	}
	if age < 24*time.Hour {
		return fmt.Sprintf("%dh", int(age.Hours()))
	}
	days := int(age.Hours() / 24)
	if days < 365 {
		return fmt.Sprintf("%dd", days)
	}
	return fmt.Sprintf("%dy", days/365)
}
