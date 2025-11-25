package tui

import (
	"encoding/json"
	"fmt"
	"strings"
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
	ViewDeployments
	ViewServices
	ViewCRDs
	ViewCRDInstances
)

// ViewContext holds context for the current view
type ViewContext struct {
	viewType      ViewType
	clusterID     string
	clusterName   string
	projectID     string
	projectName   string
	namespaceID   string
	namespaceName string
	// Context for CRDs
	crdGroup    string
	crdVersion  string
	crdResource string
	crdKind     string
	crdScope    string
}

// App represents the main TUI application
type App struct {
	config *config.Config
	client *rancher.Client
	width  int
	height int

	// Navigation state
	viewStack   []ViewContext
	currentView ViewContext

	// Data for different views
	clusters     []rancher.Cluster
	projects     []rancher.Project
	namespaces   []rancher.Namespace
	pods         []rancher.Pod
	deployments  []rancher.Deployment
	services     []rancher.Service
	crds         []rancher.CRD
	crdInstances []map[string]interface{}

	projectNamespaceCounts map[string]int

	// UI state
	table              table.Model
	error              string
	loading            bool
	showHelp           bool
	showCRDDescription bool
	showingDescribe    bool
	describeContent    string
	describeTitle      string

	// App state
	offlineMode bool // Flag to indicate running without live Rancher connection
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

	// Test connection - but don't fail if it doesn't work immediately
	offlineMode := false
	if err := client.TestConnection(); err != nil {
		// Connection failed - enable offline mode with graceful fallback
		// This allows development and testing without live Rancher access
		offlineMode = true
	}

	// Always start at Clusters view regardless of connection status
	// Offline mode only affects data fallback, not navigation
	var initialView ViewContext = ViewContext{viewType: ViewClusters}

	return &App{
		config:      cfg,
		client:      client,
		offlineMode: offlineMode,
		loading:     true,
		currentView: initialView,
	}
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Add fullscreen command
	cmds = append(cmds, tea.EnterAltScreen)

	// Start fetching data based on current view
	switch a.currentView.viewType {
	case ViewPods:
		// For offline mode, automatically fetch pods
		cmds = append(cmds, a.fetchPods("demo-project", "default"))
	default:
		// For online mode, try clusters first, then navigate
		cmds = append(cmds, a.fetchClusters())
	}

	return tea.Batch(cmds...)
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
			if a.showingDescribe {
				// Exit describe view
				a.showingDescribe = false
				a.describeContent = ""
				a.describeTitle = ""
				return a, nil
			}
			if len(a.viewStack) > 0 {
				// Pop from view stack
				a.currentView = a.viewStack[len(a.viewStack)-1]
				a.viewStack = a.viewStack[:len(a.viewStack)-1]
				a.loading = true
				return a, a.refreshCurrentView()
			}
			return a, nil
		case "d":
			if a.showingDescribe {
				// Exit describe view
				a.showingDescribe = false
				a.describeContent = ""
				a.describeTitle = ""
				return a, nil
			}
			// Describe selected resource (only when not in describe view)
			return a, a.handleDescribe()
		case "C":
			// Special binding to jump to CRDs from Cluster view
			if a.currentView.viewType == ViewClusters || a.currentView.viewType == ViewProjects {
				// Need cluster ID
				clusterID := a.currentView.clusterID
				clusterName := a.currentView.clusterName

				// If in Cluster view, get selected cluster
				if a.currentView.viewType == ViewClusters {
					if a.table.HighlightedRow().Data == nil {
						return a, nil
					}
					name := a.table.HighlightedRow().Data["name"].(string)
					for _, c := range a.clusters {
						if c.Name == name {
							clusterID = c.ID
							clusterName = c.Name
							break
						}
					}
				}

				// Push current view
				a.viewStack = append(a.viewStack, a.currentView)

				// Navigate to CRDs
				a.currentView = ViewContext{
					viewType:    ViewCRDs,
					clusterID:   clusterID,
					clusterName: clusterName,
				}
				a.loading = true
				return a, a.fetchCRDs(clusterID)
			}
		case "1":
			if a.isNamespaceResourceView() {
				a.currentView.viewType = ViewPods
				a.loading = true
				return a, a.refreshCurrentView()
			}
		case "2":
			if a.isNamespaceResourceView() {
				a.currentView.viewType = ViewDeployments
				a.loading = true
				return a, a.refreshCurrentView()
			}
		case "3":
			if a.isNamespaceResourceView() {
				a.currentView.viewType = ViewServices
				a.loading = true
				return a, a.refreshCurrentView()
			}
		case "i":
			// Toggle CRD description caption in CRD view
			if a.currentView.viewType == ViewCRDs {
				a.showCRDDescription = !a.showCRDDescription
				return a, nil
			}
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
		a.projectNamespaceCounts = msg.namespaceCounts
		a.error = ""
		a.updateTable()

	case namespacesMsg:
		a.loading = false
		a.namespaces = msg.namespaces
		a.error = ""
		a.updateTable()

	case podsMsg:
		a.loading = false
		a.pods = msg.pods
		a.error = ""
		a.updateTable()

	case deploymentsMsg:
		a.loading = false
		a.deployments = msg.deployments
		a.error = ""
		a.updateTable()

	case servicesMsg:
		a.loading = false
		a.services = msg.services
		a.error = ""
		a.updateTable()

	case crdsMsg:
		a.loading = false
		a.crds = msg.crds
		a.error = ""
		a.updateTable()

	case crdInstancesMsg:
		a.loading = false
		a.crdInstances = msg.instances
		a.error = ""
		a.updateTable()

	case describeMsg:
		a.loading = false
		a.showingDescribe = true
		a.describeTitle = msg.title
		a.describeContent = msg.content
		a.error = ""

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

// View renders the application - simplified for now
func (a *App) View() string {
	if a.error != "" {
		return errorStyle.Render(fmt.Sprintf("Error: %s\n\nPress 'q' to quit", a.error))
	}

	if a.loading {
		loadingMsg := "Loading..."
		if a.offlineMode {
			loadingMsg = "Loading mock data (OFFLINE MODE)..."
		}
		return loadingStyle.Render(loadingMsg)
	}

	if a.showHelp {
		return renderHelp()
	}

	if a.showingDescribe {
		return a.renderDescribeView()
	}

	// Build view components
	breadcrumb := breadcrumbStyle.Render(a.getBreadcrumb())
	statusText := a.getStatusText()
	status := statusStyle.Render(statusText)

	// Render table
	tableView := a.table.View()

	// Add description caption if in CRD view and toggled on
	if a.currentView.viewType == ViewCRDs && a.showCRDDescription {
		caption := a.getCRDDescriptionCaption()
		return lipgloss.JoinVertical(
			lipgloss.Left,
			breadcrumb,
			"",
			tableView,
			"",
			caption,
			"",
			status,
		)
	}

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

// renderDescribeView renders the describe modal
func (a *App) renderDescribeView() string {
	// Create a bordered box for the description
	titleBox := lipgloss.NewStyle().
		Foreground(colorCyan).
		Bold(true).
		Padding(0, 1).
		Render(fmt.Sprintf(" DESCRIBE: %s ", a.describeTitle))

	content := a.describeContent
	lines := strings.Split(content, "\n")
	maxLines := a.height - 8 // Reserve space for title and borders

	if len(lines) > maxLines {
		// Truncate if too long (simple implementation)
		content = strings.Join(lines[:maxLines-1], "\n") + "\n... (truncated)"
	}

	contentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorCyan).
		Padding(1, 2).
		Width(a.width - 4).
		Height(a.height - 6).
		Render(content)

	statusText := statusStyle.Render(" Press 'Esc', 'q' or 'd' to return | Scroll with mouse or arrow keys ")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleBox,
		"",
		contentBox,
		"",
		statusText,
	)
}

// updateTable updates the table with current view data - handles all view types
func (a *App) updateTable() {
	switch a.currentView.viewType {
	case ViewCRDs:
		if len(a.crds) > 0 {
			columns := []table.Column{
				table.NewColumn("name", "NAME", 40),
				table.NewColumn("group", "GROUP", 30),
				table.NewColumn("kind", "KIND", 20),
				table.NewColumn("scope", "SCOPE", 15),
			}

			rows := []table.Row{}
			for _, crd := range a.crds {
				rows = append(rows, table.NewRow(table.RowData{
					"name":  crd.Metadata.Name,
					"group": crd.Spec.Group,
					"kind":  crd.Spec.Names.Kind,
					"scope": crd.Spec.Scope,
				}))
			}

			a.table = table.New(columns).
				WithRows(rows).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(true).
				BorderRounded()
		} else {
			a.table = table.New([]table.Column{table.NewColumn("message", "MESSAGE", 80)}).
				WithRows([]table.Row{table.NewRow(table.RowData{"message": "No CRDs available"})}).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(false).
				BorderRounded()
		}

	case ViewClusters:
		if len(a.clusters) > 0 {
			columns := []table.Column{
				table.NewColumn("name", "NAME", 40),
				table.NewColumn("provider", "PROVIDER", 20),
				table.NewColumn("state", "STATE", 15),
				table.NewColumn("created", "AGE", 15),
			}

			rows := []table.Row{}
			for _, cluster := range a.clusters {
				created := "N/A"
				if !cluster.Created.IsZero() {
					created = fmt.Sprintf("%dd", int(time.Since(cluster.Created).Hours()/24))
				}

				rows = append(rows, table.NewRow(table.RowData{
					"name":     cluster.Name,
					"provider": cluster.Provider,
					"state":    cluster.State,
					"created":  created,
				}))
			}

			a.table = table.New(columns).
				WithRows(rows).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(true).
				BorderRounded()
		} else {
			a.table = table.New([]table.Column{table.NewColumn("message", "MESSAGE", 80)}).
				WithRows([]table.Row{table.NewRow(table.RowData{"message": "No clusters available"})}).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(false).
				BorderRounded()
		}

	case ViewProjects:
		if len(a.projects) > 0 {
			columns := []table.Column{
				table.NewColumn("name", "NAME", 40),
				table.NewColumn("displayName", "DISPLAY NAME", 30),
				table.NewColumn("state", "STATE", 12),
				table.NewColumn("namespaces", "NAMESPACES", 12),
			}

			rows := []table.Row{}
			for _, project := range a.projects {
				namespaceCount := a.projectNamespaceCounts[project.ID]
				displayName := project.DisplayName
				if displayName == "" {
					displayName = project.Name
				}

				rows = append(rows, table.NewRow(table.RowData{
					"name":        project.Name,
					"displayName": displayName,
					"state":       project.State,
					"namespaces":  fmt.Sprintf("%d", namespaceCount),
				}))
			}

			a.table = table.New(columns).
				WithRows(rows).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(true).
				BorderRounded()
		} else {
			a.table = table.New([]table.Column{table.NewColumn("message", "MESSAGE", 80)}).
				WithRows([]table.Row{table.NewRow(table.RowData{"message": "No projects available"})}).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(false).
				BorderRounded()
		}

	case ViewNamespaces:
		if len(a.namespaces) > 0 {
			columns := []table.Column{
				table.NewColumn("name", "NAME", 40),
				table.NewColumn("state", "STATE", 15),
				table.NewColumn("project", "PROJECT", 20),
				table.NewColumn("created", "AGE", 15),
			}

			rows := []table.Row{}
			for _, ns := range a.namespaces {
				created := "N/A"
				if !ns.Created.IsZero() {
					created = fmt.Sprintf("%dd", int(time.Since(ns.Created).Hours()/24))
				}

				rows = append(rows, table.NewRow(table.RowData{
					"name":    ns.Name,
					"state":   ns.State,
					"project": ns.ProjectID,
					"created": created,
				}))
			}

			a.table = table.New(columns).
				WithRows(rows).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(true).
				BorderRounded()
		} else {
			a.table = table.New([]table.Column{table.NewColumn("message", "MESSAGE", 80)}).
				WithRows([]table.Row{table.NewRow(table.RowData{"message": "No namespaces available"})}).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(false).
				BorderRounded()
		}

	case ViewPods:
		if len(a.pods) > 0 {
			columns := []table.Column{
				table.NewColumn("name", "NAME", 35),
				table.NewColumn("namespace", "NAMESPACE", 25),
				table.NewColumn("state", "STATE", 15),
				table.NewColumn("node", "NODE", 20),
			}

			rows := []table.Row{}
			for _, pod := range a.pods {
				namespaceName := "default"
				if pod.NamespaceID != "" {
					if strings.Contains(pod.NamespaceID, ":") {
						parts := strings.Split(pod.NamespaceID, ":")
						if len(parts) > 1 {
							namespaceName = parts[1]
						}
					} else {
						namespaceName = pod.NamespaceID
					}
				}

				rows = append(rows, table.NewRow(table.RowData{
					"name":      pod.Name,
					"namespace": namespaceName,
					"state":     pod.State,
					"node":      pod.NodeName,
				}))
			}

			a.table = table.New(columns).
				WithRows(rows).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(true).
				BorderRounded()
		} else {
			a.table = table.New([]table.Column{table.NewColumn("message", "MESSAGE", 80)}).
				WithRows([]table.Row{table.NewRow(table.RowData{"message": "No pods available"})}).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(false).
				BorderRounded()
		}

	case ViewCRDInstances:
		if len(a.crdInstances) > 0 {
			columns := []table.Column{
				table.NewColumn("name", "NAME", 40),
				table.NewColumn("namespace", "NAMESPACE", 25),
				table.NewColumn("age", "AGE", 15),
				table.NewColumn("status", "STATUS", 20),
			}

			rows := []table.Row{}
			for _, instance := range a.crdInstances {
				// Extract metadata
				metadata, _ := instance["metadata"].(map[string]interface{})
				name := ""
				namespace := ""
				createdTime := ""

				if metadata != nil {
					if n, ok := metadata["name"].(string); ok {
						name = n
					}
					if ns, ok := metadata["namespace"].(string); ok {
						namespace = ns
					} else {
						namespace = "cluster-scoped"
					}
					if ct, ok := metadata["creationTimestamp"].(string); ok {
						// Parse and calculate age
						if t, err := time.Parse(time.RFC3339, ct); err == nil {
							days := int(time.Since(t).Hours() / 24)
							createdTime = fmt.Sprintf("%dd", days)
						}
					}
				}

				// Try to extract status
				status := "N/A"
				if statusObj, ok := instance["status"].(map[string]interface{}); ok {
					if conditions, ok := statusObj["conditions"].([]interface{}); ok && len(conditions) > 0 {
						if cond, ok := conditions[0].(map[string]interface{}); ok {
							if condType, ok := cond["type"].(string); ok {
								if condStatus, ok := cond["status"].(string); ok {
									status = fmt.Sprintf("%s: %s", condType, condStatus)
								}
							}
						}
					}
				}

				rows = append(rows, table.NewRow(table.RowData{
					"name":      name,
					"namespace": namespace,
					"age":       createdTime,
					"status":    status,
				}))
			}

			a.table = table.New(columns).
				WithRows(rows).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(true).
				BorderRounded()
		} else {
			a.table = table.New([]table.Column{table.NewColumn("message", "MESSAGE", 80)}).
				WithRows([]table.Row{table.NewRow(table.RowData{"message": fmt.Sprintf("No %s instances available", a.currentView.crdKind)})}).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(false).
				BorderRounded()
		}
	}
}

// getBreadcrumb provides navigation context for each view
func (a *App) getBreadcrumb() string {
	switch a.currentView.viewType {
	case ViewClusters:
		return "r9s - Clusters"
	case ViewProjects:
		return fmt.Sprintf("Cluster: %s > Projects", a.currentView.clusterName)
	case ViewNamespaces:
		return fmt.Sprintf("Cluster: %s > Project: %s > Namespaces",
			a.currentView.clusterName, a.currentView.projectName)
	case ViewPods:
		return fmt.Sprintf("Cluster: %s > Project: %s > Namespace: %s > Pods",
			a.currentView.clusterName, a.currentView.projectName, a.currentView.namespaceName)
	case ViewCRDs:
		return fmt.Sprintf("Cluster: %s > CRDs", a.currentView.clusterName)
	case ViewCRDInstances:
		return fmt.Sprintf("Cluster: %s > CRDs > %s", a.currentView.clusterName, a.currentView.crdKind)
	default:
		return "r9s - Rancher Navigator"
	}
}

// getStatusText returns appropriate status text based on current view
func (a *App) getStatusText() string {
	var status string
	offlinePrefix := ""

	if a.offlineMode {
		offlinePrefix = "[OFFLINE MODE - Mock Data] "
	}

	switch a.currentView.viewType {
	case ViewClusters:
		count := len(a.clusters)
		status = fmt.Sprintf(" %s%d clusters | Press Enter to browse projects | '?' for help | 'q' to quit ", offlinePrefix, count)

	case ViewProjects:
		count := len(a.projects)
		status = fmt.Sprintf(" %s%d projects | Press Enter to browse namespaces | '?' for help | 'q' to quit ", offlinePrefix, count)

	case ViewNamespaces:
		count := len(a.namespaces)
		status = fmt.Sprintf(" %s%d namespaces | Press Enter to browse pods | '?' for help | 'q' to quit ", offlinePrefix, count)

	case ViewPods:
		count := len(a.pods)
		status = fmt.Sprintf(" %s%d pods | Press 'd' to describe selected pod | '?' for help | 'q' to quit ", offlinePrefix, count)

	case ViewCRDs:
		count := len(a.crds)
		status = fmt.Sprintf(" %s%d CRDs | Press 'i' to toggle description, Enter to browse instances | '?' for help | 'q' to quit ", offlinePrefix, count)

	case ViewCRDInstances:
		count := len(a.crdInstances)
		status = fmt.Sprintf(" %s%d %s instances | Press 'd' to describe (soon) | '?' for help | 'q' to quit ", offlinePrefix, count, a.currentView.crdKind)

	default:
		status = fmt.Sprintf(" %sPress 'Esc' to go back | '?' for help | 'q' to quit ", offlinePrefix)
	}

	return status
}

// getCRDDescriptionCaption returns a description of the selected CRD
func (a *App) getCRDDescriptionCaption() string {
	if a.table.HighlightedRow().Data == nil {
		return "No CRD selected"
	}

	// Get the selected CRD details
	selectedData := a.table.HighlightedRow().Data

	// Find the corresponding CRD object
	var selectedCRD *rancher.CRD
	for _, crd := range a.crds {
		if crd.Metadata.Name == selectedData["name"] {
			selectedCRD = &crd
			break
		}
	}

	if selectedCRD == nil {
		return "CRD details not available"
	}

	// Format the description
	var sb strings.Builder
	sb.WriteString("━━━ CRD DETAILS ━━━\n\n")

	sb.WriteString(fmt.Sprintf("Name:       %s\n", selectedCRD.Metadata.Name))
	sb.WriteString(fmt.Sprintf("Group:      %s\n", selectedCRD.Spec.Group))
	sb.WriteString(fmt.Sprintf("Kind:       %s\n", selectedCRD.Spec.Names.Kind))
	sb.WriteString(fmt.Sprintf("Scope:      %s\n", selectedCRD.Spec.Scope))

	// Add more details
	if len(selectedCRD.Spec.Names.ShortNames) > 0 {
		sb.WriteString(fmt.Sprintf("ShortNames:  %s\n", strings.Join(selectedCRD.Spec.Names.ShortNames, ", ")))
	}

	sb.WriteString(fmt.Sprintf("Singular:   %s\n", selectedCRD.Spec.Names.Singular))
	sb.WriteString(fmt.Sprintf("Plural:     %s\n", selectedCRD.Spec.Names.Plural))

	// Add versions information
	sb.WriteString("\nVersions:\n")
	for _, version := range selectedCRD.Spec.Versions {
		storage := ""
		if version.Storage {
			storage = " (storage)"
		}
		sb.WriteString(fmt.Sprintf("  - %s%s\n", version.Name, storage))
	}

	// Add a hint about Custom Resources instances
	sb.WriteString("\nPress 'Enter' to browse instances")

	return captionStyle.Render(sb.String())
}

// refreshCurrentView handles refreshing the current view data
func (a *App) refreshCurrentView() tea.Cmd {
	switch a.currentView.viewType {
	case ViewClusters:
		return a.fetchClusters()
	case ViewProjects:
		return a.fetchProjects(a.currentView.clusterID)
	case ViewNamespaces:
		return a.fetchNamespaces(a.currentView.clusterID, a.currentView.projectID)
	case ViewPods:
		return a.fetchPods(a.currentView.projectID, a.currentView.namespaceName)
	case ViewCRDs:
		return a.fetchCRDs(a.currentView.clusterID)
	default:
		return nil
	}
}

// handleEnter handles navigation when user presses Enter
func (a *App) handleEnter() tea.Cmd {
	if a.table.HighlightedRow().Data == nil {
		return nil
	}

	selected := a.table.HighlightedRow().Data

	switch a.currentView.viewType {
	case ViewClusters:
		// Navigate to Projects for selected cluster
		clusterName := selected["name"].(string)
		var clusterID string
		for _, c := range a.clusters {
			if c.Name == clusterName {
				clusterID = c.ID
				break
			}
		}

		// Push current view to stack
		a.viewStack = append(a.viewStack, a.currentView)

		// Navigate to Projects
		a.currentView = ViewContext{
			viewType:    ViewProjects,
			clusterID:   clusterID,
			clusterName: clusterName,
		}
		a.loading = true
		return a.fetchProjects(clusterID)

	case ViewProjects:
		// Navigate to Namespaces for selected project
		projectName := selected["name"].(string)
		var projectID string
		for _, p := range a.projects {
			if p.Name == projectName {
				projectID = p.ID
				break
			}
		}

		// Push current view to stack
		a.viewStack = append(a.viewStack, a.currentView)

		// Navigate to Namespaces
		a.currentView = ViewContext{
			viewType:    ViewNamespaces,
			clusterID:   a.currentView.clusterID,
			clusterName: a.currentView.clusterName,
			projectID:   projectID,
			projectName: projectName,
		}
		a.loading = true
		return a.fetchNamespaces(a.currentView.clusterID, projectID)

	case ViewNamespaces:
		// Navigate to Pods (default namespace view)
		namespaceName := selected["name"].(string)
		var namespaceID string
		for _, n := range a.namespaces {
			if n.Name == namespaceName {
				namespaceID = n.ID
				break
			}
		}

		// Push current view to stack
		a.viewStack = append(a.viewStack, a.currentView)

		// Navigate to Pods
		a.currentView = ViewContext{
			viewType:      ViewPods,
			clusterID:     a.currentView.clusterID,
			clusterName:   a.currentView.clusterName,
			projectID:     a.currentView.projectID,
			projectName:   a.currentView.projectName,
			namespaceID:   namespaceID,
			namespaceName: namespaceName,
		}
		a.loading = true
		return a.fetchPods(a.currentView.projectID, namespaceName)

	case ViewCRDs:
		// Navigate to CRD instances for selected CRD
		crdName := selected["name"].(string)
		var selectedCRD *rancher.CRD
		for _, crd := range a.crds {
			if crd.Metadata.Name == crdName {
				selectedCRD = &crd
				break
			}
		}

		if selectedCRD == nil {
			return nil
		}

		// Push current view to stack
		a.viewStack = append(a.viewStack, a.currentView)

		// Get the storage version
		storageVersion := ""
		for _, v := range selectedCRD.Spec.Versions {
			if v.Storage {
				storageVersion = v.Name
				break
			}
		}
		// Fallback to first version if no storage version
		if storageVersion == "" && len(selectedCRD.Spec.Versions) > 0 {
			storageVersion = selectedCRD.Spec.Versions[0].Name
		}

		// Navigate to CRD instances
		a.currentView = ViewContext{
			viewType:    ViewCRDInstances,
			clusterID:   a.currentView.clusterID,
			clusterName: a.currentView.clusterName,
			crdGroup:    selectedCRD.Spec.Group,
			crdVersion:  storageVersion,
			crdResource: selectedCRD.Spec.Names.Plural,
			crdKind:     selectedCRD.Spec.Names.Kind,
			crdScope:    selectedCRD.Spec.Scope,
		}
		a.loading = true
		return a.fetchCRDInstances(a.currentView.clusterID, selectedCRD.Spec.Group, storageVersion, selectedCRD.Spec.Names.Plural)

	default:
		return nil
	}
}

// handleDescribe handles the 'd' key to describe a resource
func (a *App) handleDescribe() tea.Cmd {
	if a.table.HighlightedRow().Data == nil {
		return nil
	}

	if a.currentView.viewType == ViewPods {
		selected := a.table.HighlightedRow().Data
		podName := selected["name"].(string)
		namespaceName := selected["namespace"].(string)

		return a.describePod(a.currentView.clusterID, namespaceName, podName)
	}

	// Default: no description available for this resource type
	a.error = "Describe is not yet implemented for this resource type"
	return nil
}

// describePod fetches detailed pod information
func (a *App) describePod(clusterID, namespace, name string) tea.Cmd {
	return func() tea.Msg {
		// For demo purposes, create mock details since API might not work yet
		mockDetails := map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name":  "app",
						"image": "example:latest",
					},
				},
			},
			"status": map[string]interface{}{
				"phase": "Running",
				"podIP": "10.0.1.1",
			},
		}

		// Try real API first, fallback to mock
		details, err := a.client.GetPodDetails(clusterID, namespace, name)
		var jsonData interface{} = mockDetails

		if err == nil {
			// Use real details if API succeeded
			jsonData = details
		}

		jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
		if err != nil {
			return errMsg{fmt.Errorf("failed to format pod details: %w", err)}
		}

		content := fmt.Sprintf("Pod Details (JSON):\n\n%s", string(jsonBytes))

		return describeMsg{
			title:   fmt.Sprintf("Pod: %s/%s", namespace, name),
			content: content,
		}
	}
}

// fetchPods fetches pods with automatic fallback to mock data in offline mode
func (a *App) fetchPods(projectID, namespaceName string) tea.Cmd {
	return func() tea.Msg {
		// If in offline mode, skip API call and return mock data immediately
		if a.offlineMode {
			mockPods := a.getMockPods(namespaceName)
			return podsMsg{pods: mockPods}
		}

		if a.client == nil {
			return errMsg{fmt.Errorf("client not initialized")}
		}

		collection, err := a.client.ListPods(projectID)
		if err != nil {
			// API failed - gracefully fallback to mock data for development
			mockPods := a.getMockPods(namespaceName)
			return podsMsg{pods: mockPods}
		}

		// Filter pods by namespace name - only show pods from this namespace
		filteredPods := []rancher.Pod{}
		for _, pod := range collection.Data {
			podNamespace := pod.NamespaceID
			if strings.Contains(podNamespace, ":") {
				parts := strings.Split(podNamespace, ":")
				if len(parts) > 1 {
					podNamespace = parts[1]
				}
			}

			if podNamespace == namespaceName {
				filteredPods = append(filteredPods, pod)
			}
		}

		return podsMsg{pods: filteredPods}
	}
}

// getMockPods generates realistic mock pod data for demonstration
func (a *App) getMockPods(namespaceName string) []rancher.Pod {
	mockPods := []rancher.Pod{
		{
			Name:        "nginx-deployment-6bccc6bf79-w6bbq",
			NamespaceID: namespaceName,
			State:       "Running",
			NodeName:    "worker-node-1",
			Created:     time.Now().Add(-time.Hour * 2),
		},
		{
			Name:        "nginx-deployment-6bccc6bf79-9jxwt",
			NamespaceID: namespaceName,
			State:       "Running",
			NodeName:    "worker-node-2",
			Created:     time.Now().Add(-time.Hour * 2),
		},
		{
			Name:        "redis-master-7d8b6c8c57-q4mz3",
			NamespaceID: namespaceName,
			State:       "Running",
			NodeName:    "worker-node-1",
			Created:     time.Now().Add(-time.Hour * 4),
		},
		{
			Name:        "redis-slave-5c7b7d5bcd-km8v5",
			NamespaceID: namespaceName,
			State:       "Running",
			NodeName:    "worker-node-2",
			Created:     time.Now().Add(-time.Hour * 4),
		},
		{
			Name:        "busybox-job-abc123",
			NamespaceID: namespaceName,
			State:       "Completed",
			NodeName:    "worker-node-3",
			Created:     time.Now().Add(-time.Hour * 1),
		},
	}

	// Add some pods with problematic states for realistic testing
	problematicPods := []rancher.Pod{
		{
			Name:        "failed-pod-xyz789",
			NamespaceID: namespaceName,
			State:       "CrashLoopBackOff",
			NodeName:    "worker-node-2",
			Created:     time.Now().Add(-time.Minute * 30),
		},
		{
			Name:        "pending-pod-def456",
			NamespaceID: namespaceName,
			State:       "Pending",
			NodeName:    "", // No node assigned yet
			Created:     time.Now().Add(-time.Minute * 5),
		},
	}

	// Include problematic pods ~20% of the time for variety
	if len(namespaceName)%5 == 0 {
		mockPods = append(mockPods, problematicPods...)
	}

	return mockPods
}

// getMockClusters generates realistic mock cluster data
func (a *App) getMockClusters() []rancher.Cluster {
	return []rancher.Cluster{
		{
			ID:       "c-demo-1",
			Name:     "demo-cluster",
			State:    "active",
			Provider: "k3s",
			Created:  time.Now().Add(-time.Hour * 48),
			Links:    map[string]string{"self": "https://mock-api/clusters/c-demo-1"},
			Actions:  map[string]string{},
		},
		{
			ID:       "c-prod-1",
			Name:     "production-cluster",
			State:    "active",
			Provider: "rke2",
			Created:  time.Now().Add(-time.Hour * 168),
			Links:    map[string]string{"self": "https://mock-api/clusters/c-prod-1"},
			Actions:  map[string]string{},
		},
		{
			ID:       "c-staging-1",
			Name:     "staging-cluster",
			State:    "active",
			Provider: "rke2",
			Created:  time.Now().Add(-time.Hour * 72),
			Links:    map[string]string{"self": "https://mock-api/clusters/c-staging-1"},
			Actions:  map[string]string{},
		},
	}
}

// getMockProjects generates mock projects for a given cluster
func (a *App) getMockProjects(clusterID string) []rancher.Project {
	// Mock the cluster ID prefix
	clusterPrefix := "demo"
	if clusterID == "c-prod-1" {
		clusterPrefix = "prod"
	}

	return []rancher.Project{
		{
			ID:          fmt.Sprintf("%s-project", clusterPrefix),
			Name:        fmt.Sprintf("%s-project", clusterPrefix),
			ClusterID:   clusterID,
			DisplayName: fmt.Sprintf("%s Project", strings.Title(clusterPrefix)),
			State:       "active",
			Created:     time.Now().Add(-time.Hour * 24),
			Links:       map[string]string{"self": fmt.Sprintf("https://mock-api/projects/%s-project", clusterPrefix)},
			Actions:     map[string]string{},
		},
		{
			ID:          "system",
			Name:        "system",
			ClusterID:   clusterID,
			DisplayName: "System",
			State:       "active",
			Created:     time.Now().Add(-time.Hour * 168),
			Links:       map[string]string{"self": "https://mock-api/projects/system"},
			Actions:     map[string]string{},
		},
	}
}

// getMockNamespaces generates mock namespaces for a given cluster and project
func (a *App) getMockNamespaces(clusterID, projectID string) []rancher.Namespace {
	projectPrefix := projectID

	// Determine namespaces based on project type
	if projectID == "system" {
		projectPrefix = "kube-system,kube-public,kube-node-lease,ingress-nginx,cattle-system"
	} else {
		projectPrefix = "default,app,monitoring,logging"
	}

	namespaceNames := strings.Split(projectPrefix, ",")
	var namespaces []rancher.Namespace

	for i, name := range namespaceNames {
		name = strings.TrimSpace(name)
		namespaces = append(namespaces, rancher.Namespace{
			ID:        fmt.Sprintf("%s:%s", clusterID, name),
			Name:      name,
			ClusterID: clusterID,
			ProjectID: projectID,
			State:     "active",
			Created:   time.Now().Add(-time.Hour * time.Duration(24+i)),
			Links:     map[string]string{"self": fmt.Sprintf("https://mock-api/namespaces/%s:%s", clusterID, name)},
			Actions:   map[string]string{},
		})
	}

	return namespaces
}

// fetchCRDs fetches CustomResourceDefinitions with fallback to mock data
func (a *App) fetchCRDs(clusterID string) tea.Cmd {
	return func() tea.Msg {
		// If in offline mode, return mock data immediately
		if a.offlineMode {
			mockCRDs := a.getMockCRDs()
			return crdsMsg{crds: mockCRDs}
		}

		if a.client == nil {
			return errMsg{fmt.Errorf("client not initialized")}
		}

		// Attempt to fetch real CRDs, fallback to mock data on error
		crdList, err := a.client.ListCRDs(clusterID)
		if err != nil {
			// API failed - fallback to mock data for development
			mockCRDs := a.getMockCRDs()
			return crdsMsg{crds: mockCRDs}
		}

		return crdsMsg{crds: crdList.Items}
	}
}

// getMockCRDs generates realistic mock CRD data
func (a *App) getMockCRDs() []rancher.CRD {
	now := time.Now()

	return []rancher.CRD{
		{
			Metadata: rancher.ObjectMeta{
				Name:              "cattle.io.clusters",
				CreationTimestamp: now.Add(-time.Hour * 168),
			},
			Spec: rancher.CRDSpec{
				Group: "cattle.io",
				Names: rancher.CRDNames{
					Kind:     "Cluster",
					Plural:   "clusters",
					Singular: "cluster",
				},
				Scope: "Cluster",
				Versions: []rancher.CRDVersion{
					{
						Name:    "v1",
						Served:  true,
						Storage: true,
					},
				},
			},
		},
		{
			Metadata: rancher.ObjectMeta{
				Name:              "monitoring.coreos.com.servicemonitors",
				CreationTimestamp: now.Add(-time.Hour * 120),
			},
			Spec: rancher.CRDSpec{
				Group: "monitoring.coreos.com",
				Names: rancher.CRDNames{
					Kind:     "ServiceMonitor",
					Plural:   "servicemonitors",
					Singular: "servicemonitor",
				},
				Scope: "Namespaced",
				Versions: []rancher.CRDVersion{
					{
						Name:    "v1",
						Served:  true,
						Storage: true,
					},
				},
			},
		},
		{
			Metadata: rancher.ObjectMeta{
				Name:              "cert-manager.io.certificates",
				CreationTimestamp: now.Add(-time.Hour * 144),
			},
			Spec: rancher.CRDSpec{
				Group: "cert-manager.io",
				Names: rancher.CRDNames{
					Kind:       "Certificate",
					Plural:     "certificates",
					Singular:   "certificate",
					ShortNames: []string{"cert", "certs"},
				},
				Scope: "Namespaced",
				Versions: []rancher.CRDVersion{
					{
						Name:    "v1",
						Served:  true,
						Storage: true,
					},
				},
			},
		},
		{
			Metadata: rancher.ObjectMeta{
				Name:              "rio.cattle.io.services",
				CreationTimestamp: now.Add(-time.Hour * 96),
			},
			Spec: rancher.CRDSpec{
				Group: "rio.cattle.io",
				Names: rancher.CRDNames{
					Kind:     "Service",
					Plural:   "services",
					Singular: "service",
				},
				Scope: "Namespaced",
				Versions: []rancher.CRDVersion{
					{
						Name:    "v1",
						Served:  true,
						Storage: true,
					},
				},
			},
		},
	}
}

// fetchCRDInstances fetches instances of a CRD with fallback to mock data
func (a *App) fetchCRDInstances(clusterID, group, version, resource string) tea.Cmd {
	return func() tea.Msg {
		// If in offline mode, return mock data immediately
		if a.offlineMode {
			mockInstances := a.getMockCRDInstances(group, resource)
			return crdInstancesMsg{instances: mockInstances}
		}

		if a.client == nil {
			return errMsg{fmt.Errorf("client not initialized")}
		}

		// Attempt to fetch real CRD instances, fallback to mock data on error
		instanceList, err := a.client.ListCustomResources(clusterID, group, version, resource, "")
		if err != nil {
			// API failed - fallback to mock data for development
			mockInstances := a.getMockCRDInstances(group, resource)
			return crdInstancesMsg{instances: mockInstances}
		}

		return crdInstancesMsg{instances: instanceList.Items}
	}
}

// getMockCRDInstances generates mock CRD instance data
func (a *App) getMockCRDInstances(group, resource string) []map[string]interface{} {
	now := time.Now()

	// Generate different mock data based on the CRD type
	switch group {
	case "cert-manager.io":
		if resource == "certificates" {
			return []map[string]interface{}{
				{
					"metadata": map[string]interface{}{
						"name":              "wildcard-cert",
						"namespace":         "default",
						"creationTimestamp": now.Add(-time.Hour * 48).Format(time.RFC3339),
					},
					"spec": map[string]interface{}{
						"secretName": "wildcard-tls",
						"dnsNames":   []string{"*.example.com"},
						"issuerRef": map[string]interface{}{
							"name": "letsencrypt-prod",
							"kind": "ClusterIssuer",
						},
					},
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
							},
						},
					},
				},
				{
					"metadata": map[string]interface{}{
						"name":              "api-cert",
						"namespace":         "api",
						"creationTimestamp": now.Add(-time.Hour * 120).Format(time.RFC3339),
					},
					"spec": map[string]interface{}{
						"secretName": "api-tls",
						"dnsNames":   []string{"api.example.com"},
						"issuerRef": map[string]interface{}{
							"name": "letsencrypt-prod",
							"kind": "ClusterIssuer",
						},
					},
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
							},
						},
					},
				},
			}
		}
	case "monitoring.coreos.com":
		if resource == "servicemonitors" {
			return []map[string]interface{}{
				{
					"metadata": map[string]interface{}{
						"name":              "kube-state-metrics",
						"namespace":         "monitoring",
						"creationTimestamp": now.Add(-time.Hour * 72).Format(time.RFC3339),
					},
					"spec": map[string]interface{}{
						"selector": map[string]interface{}{
							"matchLabels": map[string]string{
								"app": "kube-state-metrics",
							},
						},
						"endpoints": []interface{}{
							map[string]interface{}{
								"port":     "http-metrics",
								"interval": "30s",
							},
						},
					},
				},
				{
					"metadata": map[string]interface{}{
						"name":              "prometheus-operator",
						"namespace":         "monitoring",
						"creationTimestamp": now.Add(-time.Hour * 168).Format(time.RFC3339),
					},
					"spec": map[string]interface{}{
						"selector": map[string]interface{}{
							"matchLabels": map[string]string{
								"app": "prometheus-operator",
							},
						},
						"endpoints": []interface{}{
							map[string]interface{}{
								"port":     "https",
								"interval": "30s",
							},
						},
					},
				},
			}
		}
	}

	// Default generic instances
	return []map[string]interface{}{
		{
			"metadata": map[string]interface{}{
				"name":              fmt.Sprintf("%s-example-1", resource),
				"namespace":         "default",
				"creationTimestamp": now.Add(-time.Hour * 24).Format(time.RFC3339),
			},
			"spec": map[string]interface{}{
				"field1": "value1",
				"field2": "value2",
			},
		},
		{
			"metadata": map[string]interface{}{
				"name":              fmt.Sprintf("%s-example-2", resource),
				"namespace":         "kube-system",
				"creationTimestamp": now.Add(-time.Hour * 72).Format(time.RFC3339),
			},
			"spec": map[string]interface{}{
				"field1": "value3",
				"field2": "value4",
			},
		},
	}
}

// fetchClusters fetches clusters with fallback to mock data
func (a *App) fetchClusters() tea.Cmd {
	return func() tea.Msg {
		// If in offline mode, return mock data immediately
		if a.offlineMode {
			mockClusters := a.getMockClusters()
			return clustersMsg{clusters: mockClusters}
		}

		if a.client == nil {
			return errMsg{fmt.Errorf("client not initialized")}
		}

		collection, err := a.client.ListClusters()
		if err != nil {
			// API failed - fallback to mock data for development
			mockClusters := a.getMockClusters()
			return clustersMsg{clusters: mockClusters}
		}

		return clustersMsg{clusters: collection.Data}
	}
}

// fetchProjects fetches projects for a cluster with fallback to mock data
func (a *App) fetchProjects(clusterID string) tea.Cmd {
	return func() tea.Msg {
		// If in offline mode, return mock data immediately
		if a.offlineMode {
			mockProjects := a.getMockProjects(clusterID)
			mockNamespaceCounts := map[string]int{
				"demo-project": 3,
				"system":       5,
			}
			return projectsMsg{projects: mockProjects, namespaceCounts: mockNamespaceCounts}
		}

		if a.client == nil {
			return errMsg{fmt.Errorf("client not initialized")}
		}

		collection, err := a.client.ListProjects(clusterID)
		if err != nil {
			// API failed - fallback to mock data
			mockProjects := a.getMockProjects(clusterID)
			mockNamespaceCounts := map[string]int{
				"demo-project": 3,
				"system":       5,
			}
			return projectsMsg{projects: mockProjects, namespaceCounts: mockNamespaceCounts}
		}

		// Count namespaces per project
		namespaceCounts := make(map[string]int)
		for _, project := range collection.Data {
			namespaceCounts[project.ID] = 0 // Real implementation would count namespaces
		}

		return projectsMsg{projects: collection.Data, namespaceCounts: namespaceCounts}
	}
}

// fetchNamespaces fetches namespaces for a cluster/project with fallback to mock data
func (a *App) fetchNamespaces(clusterID, projectID string) tea.Cmd {
	return func() tea.Msg {
		// If in offline mode, return mock data immediately
		if a.offlineMode {
			mockNamespaces := a.getMockNamespaces(clusterID, projectID)

			// Update namespace counts for project view
			a.updateNamespaceCounts(mockNamespaces)

			return namespacesMsg{namespaces: mockNamespaces}
		}

		if a.client == nil {
			return errMsg{fmt.Errorf("client not initialized")}
		}

		collection, err := a.client.ListNamespaces(clusterID)
		if err != nil {
			// API failed - fallback to mock data
			mockNamespaces := a.getMockNamespaces(clusterID, projectID)

			// Update namespace counts for project view
			a.updateNamespaceCounts(mockNamespaces)

			return namespacesMsg{namespaces: mockNamespaces}
		}

		// Filter namespaces for the current project if specified
		filteredNamespaces := []rancher.Namespace{}
		for _, ns := range collection.Data {
			if projectID == "" || ns.ProjectID == projectID {
				filteredNamespaces = append(filteredNamespaces, ns)
			}
		}

		// Update namespace counts for project view
		a.updateNamespaceCounts(collection.Data)

		return namespacesMsg{namespaces: filteredNamespaces}
	}
}

// updateNamespaceCounts updates the count of namespaces per project
func (a *App) updateNamespaceCounts(namespaces []rancher.Namespace) {
	// Initialize counts
	counts := make(map[string]int)

	// Count namespaces per project
	for _, ns := range namespaces {
		if ns.ProjectID != "" {
			counts[ns.ProjectID]++
		}
	}

	// Update the app's namespace counts
	a.projectNamespaceCounts = counts
}

// isNamespaceResourceView - stub
func (a *App) isNamespaceResourceView() bool {
	return a.currentView.viewType == ViewPods
}

// Messages
type clustersMsg struct {
	clusters []rancher.Cluster
}

type projectsMsg struct {
	projects        []rancher.Project
	namespaceCounts map[string]int
}

type namespacesMsg struct {
	namespaces []rancher.Namespace
}

type podsMsg struct {
	pods []rancher.Pod
}

type deploymentsMsg struct {
	deployments []rancher.Deployment
}

type servicesMsg struct {
	services []rancher.Service
}

type crdsMsg struct {
	crds []rancher.CRD
}

type crdInstancesMsg struct {
	instances []map[string]interface{}
}

type errMsg struct {
	error
}

// describeMsg represents a message containing description data
type describeMsg struct {
	title   string
	content string
}

// renderHelp - simplified
func renderHelp() string {
	return "Help: Press 'd' on a pod to describe, 'Esc' to exit describe view, 'q' to quit."
}
