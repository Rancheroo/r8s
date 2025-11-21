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

// ViewType represents different view types
type ViewType int

const (
	ViewClusters ViewType = iota
	ViewProjects
	ViewNamespaces
	ViewPods
)

// ViewContext holds context for the current view
type ViewContext struct {
	viewType    ViewType
	clusterID   string
	clusterName string
	projectID   string
	projectName string
}

// App represents the main TUI application
type App struct {
	config  *config.Config
	client  *rancher.Client
	width   int
	height  int
	
	// Navigation state
	viewStack   []ViewContext
	currentView ViewContext
	
	// Data for different views
	clusters []rancher.Cluster
	projects []rancher.Project
	
	// UI state
	table    table.Model
	error    string
	loading  bool
	showHelp bool
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
		config:      cfg,
		client:      client,
		loading:     true,
		currentView: ViewContext{viewType: ViewClusters},
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
		// Handle help screen
		if a.showHelp {
			if msg.String() == "?" || msg.String() == "esc" || msg.String() == "q" {
				a.showHelp = false
				return a, nil
			}
			return a, nil
		}
		
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case "r", "ctrl+r":
			a.loading = true
			return a, a.refreshCurrentView()
		case "?":
			a.showHelp = true
			return a, nil
		case "enter":
			return a, a.handleEnter()
		case "esc":
			if len(a.viewStack) > 0 {
				// Pop from view stack
				a.currentView = a.viewStack[len(a.viewStack)-1]
				a.viewStack = a.viewStack[:len(a.viewStack)-1]
				a.loading = true
				return a, a.refreshCurrentView()
			}
			return a, nil
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
		
	case projectsMsg:
		a.loading = false
		a.projects = msg.projects
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
	
	if a.showHelp {
		return renderHelp()
	}
	
	// Build view components
	breadcrumb := breadcrumbStyle.Render(a.getBreadcrumb())
	statusText := a.getStatusText()
	status := statusStyle.Render(statusText)
	
	// Render table
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

// updateTable updates the table with current view data
func (a *App) updateTable() {
	switch a.currentView.viewType {
	case ViewClusters:
		a.updateClustersTable()
	case ViewProjects:
		a.updateProjectsTable()
	}
}

// updateClustersTable updates the table with cluster data
func (a *App) updateClustersTable() {
	if len(a.clusters) == 0 {
		return
	}
	
	// Define columns
	columns := []table.Column{
		table.NewColumn("name", "NAME", 30),
		table.NewColumn("state", "STATE", 15),
		table.NewColumn("version", "VERSION", 20),
		table.NewColumn("provider", "PROVIDER", 15),
		table.NewColumn("age", "AGE", 10),
	}
	
	// Build rows with styled state
	rows := []table.Row{}
	for _, cluster := range a.clusters {
		age := formatAge(cluster.Created)
		version := formatVersion(cluster.Version)
		stateStyled := GetStateStyle(cluster.State).Render(cluster.State)
		
		rows = append(rows, table.NewRow(table.RowData{
			"name":     cluster.Name,
			"state":    stateStyled,
			"version":  version,
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

// updateProjectsTable updates the table with project data
func (a *App) updateProjectsTable() {
	if len(a.projects) == 0 {
		return
	}
	
	// Define columns
	columns := []table.Column{
		table.NewColumn("name", "NAME", 40),
		table.NewColumn("state", "STATE", 15),
		table.NewColumn("namespaces", "NAMESPACES", 15),
		table.NewColumn("age", "AGE", 10),
	}
	
	// Build rows with styled state
	rows := []table.Row{}
	for _, project := range a.projects {
		age := formatAge(project.Created)
		stateStyled := GetStateStyle(project.State).Render(project.State)
		
		// Display name or ID if name is empty
		displayName := project.DisplayName
		if displayName == "" {
			displayName = project.Name
		}
		
		rows = append(rows, table.NewRow(table.RowData{
			"name":       displayName,
			"state":      stateStyled,
			"namespaces": "-",  // TODO: fetch namespace count
			"age":        age,
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

type projectsMsg struct {
	projects []rancher.Project
}

type errMsg struct {
	error
}

// formatVersion formats a ClusterVersion object into a string
func formatVersion(v *rancher.ClusterVersion) string {
	if v == nil {
		return "N/A"
	}
	if v.GitVersion != "" {
		return v.GitVersion
	}
	if v.Major != "" && v.Minor != "" {
		return fmt.Sprintf("v%s.%s", v.Major, v.Minor)
	}
	return "Unknown"
}

// renderHelp renders the help screen
func renderHelp() string {
	// ASCII Rancher cow logo
	logo := lipgloss.NewStyle().
		Foreground(colorCyan).
		Bold(true).
		Padding(0, 2).
		Render(`
    /\_/\     r9s - Rancher9s
   ( o.o )    k9s-inspired TUI for Rancher
    > ^ <     
   /|   |\    Press '?' 'Esc' or 'q' to close
  (_|   |_)`)
	
	helpText := lipgloss.NewStyle().
		Foreground(colorWhite).
		Padding(1, 2).
		Render(`NAVIGATION
  ↑/k          Move up
  ↓/j          Move down
  g            Go to top
  G            Go to bottom
  PgUp/PgDn    Page up/down

ACTIONS
  Enter        Navigate into selected resource
  Esc          Go back to previous view
  r / Ctrl+R   Refresh current view
  ?            Show this help
  q / Ctrl+C   Quit application

COMMAND MODE (coming soon)
  :            Enter command mode
  :clusters    List clusters
  :projects    List projects
  :pods        List pods

FILTER MODE (coming soon)
  /            Enter filter mode
  Esc          Exit filter mode

STATUS COLORS
  Green        Active / Running
  Yellow       Pending / Provisioning
  Red          Failed / Error
  Gray         Completed / Terminated`)
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		logo,
		"",
		helpText,
	)
}

// getBreadcrumb returns the breadcrumb string based on current view
func (a *App) getBreadcrumb() string {
	switch a.currentView.viewType {
	case ViewClusters:
		return "Rancher Clusters"
	case ViewProjects:
		return fmt.Sprintf("Cluster: %s > Projects", a.currentView.clusterName)
	case ViewNamespaces:
		return fmt.Sprintf("Cluster: %s > Project: %s > Namespaces",
			a.currentView.clusterName, a.currentView.projectName)
	case ViewPods:
		return fmt.Sprintf("Cluster: %s > Project: %s > Pods",
			a.currentView.clusterName, a.currentView.projectName)
	default:
		return "r9s"
	}
}

// getStatusText returns the status bar text based on current view
func (a *App) getStatusText() string {
	count := 0
	resourceType := "items"
	
	switch a.currentView.viewType {
	case ViewClusters:
		count = len(a.clusters)
		resourceType = "clusters"
	case ViewProjects:
		count = len(a.projects)
		resourceType = "projects"
	}
	
	navHelp := ""
	if len(a.viewStack) > 0 {
		navHelp = " | 'Esc' to go back"
	}
	
	return fmt.Sprintf(" %d %s%s | '?' for help | 'q' to quit | 'r' to refresh ",
		count, resourceType, navHelp)
}

// refreshCurrentView refreshes data for the current view
func (a *App) refreshCurrentView() tea.Cmd {
	switch a.currentView.viewType {
	case ViewClusters:
		return a.fetchClusters()
	case ViewProjects:
		return a.fetchProjects(a.currentView.clusterID)
	default:
		return nil
	}
}

// handleEnter handles Enter key press to navigate into resources
func (a *App) handleEnter() tea.Cmd {
	if a.table.HighlightedRow().Data == nil {
		return nil
	}
	
	switch a.currentView.viewType {
	case ViewClusters:
		// Get selected cluster
		selected := a.table.HighlightedRow()
		if selected.Data == nil {
			return nil
		}
		
		clusterName := selected.Data["name"].(string)
		// Find cluster by name to get ID
		for _, cluster := range a.clusters {
			if cluster.Name == clusterName {
				// Push current view to stack
				a.viewStack = append(a.viewStack, a.currentView)
				
				// Navigate to projects view
				a.currentView = ViewContext{
					viewType:    ViewProjects,
					clusterID:   cluster.ID,
					clusterName: cluster.Name,
				}
				a.loading = true
				return a.fetchProjects(cluster.ID)
			}
		}
	}
	
	return nil
}

// fetchProjects fetches projects for a cluster
func (a *App) fetchProjects(clusterID string) tea.Cmd {
	return func() tea.Msg {
		if a.client == nil {
			return errMsg{fmt.Errorf("client not initialized")}
		}
		
		collection, err := a.client.ListProjects(clusterID)
		if err != nil {
			return errMsg{err}
		}
		
		return projectsMsg{projects: collection.Data}
	}
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
