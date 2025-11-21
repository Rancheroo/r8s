package tui

import (
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
	table             table.Model
	error             string
	loading           bool
	showHelp          bool
	showCRDDescription bool
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

// updateTable updates the table with current view data
func (a *App) updateTable() {
	switch a.currentView.viewType {
	case ViewClusters:
		a.updateClustersTable()
	case ViewProjects:
		a.updateProjectsTable()
	case ViewNamespaces:
		a.updateNamespacesTable()
	case ViewPods:
		a.updatePodsTable()
	case ViewDeployments:
		a.updateDeploymentsTable()
	case ViewServices:
		a.updateServicesTable()
	case ViewCRDs:
		a.updateCRDsTable()
	case ViewCRDInstances:
		a.updateCRDInstancesTable()
	}
}

// updateClustersTable updates the table with cluster data
func (a *App) updateClustersTable() {
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
			"namespaces": fmt.Sprintf("%d", a.projectNamespaceCounts[project.ID]),
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

// updateNamespacesTable updates the table with namespace data
func (a *App) updateNamespacesTable() {
	// Define columns
	columns := []table.Column{
		table.NewColumn("name", "NAME", 40),
		table.NewColumn("project", "PROJECT", 25),
		table.NewColumn("state", "STATE", 15),
		table.NewColumn("age", "AGE", 10),
	}

	// Build rows with styled state
	rows := []table.Row{}
	for _, ns := range a.namespaces {
		age := formatAge(ns.Created)
		stateStyled := GetStateStyle(ns.State).Render(ns.State)

		// Extract project name from ID
		projectName := "-"
		if ns.ProjectID != "" {
			// ProjectID format: c-xxxxx:p-yyyyy
			parts := strings.Split(ns.ProjectID, ":")
			if len(parts) > 1 {
				projectName = parts[1]
			}
		}

		rows = append(rows, table.NewRow(table.RowData{
			"name":    ns.Name,
			"project": projectName,
			"state":   stateStyled,
			"age":     age,
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

// updatePodsTable updates the table with pod data
func (a *App) updatePodsTable() {
	// Define columns
	columns := []table.Column{
		table.NewColumn("name", "NAME", 35),
		table.NewColumn("namespace", "NAMESPACE", 25),
		table.NewColumn("state", "STATE", 15),
		table.NewColumn("node", "NODE", 20),
		table.NewColumn("restarts", "RESTARTS", 10),
		table.NewColumn("age", "AGE", 10),
	}

	// Build rows with styled state
	rows := []table.Row{}
	for _, pod := range a.pods {
		age := formatAge(pod.Created)
		stateStyled := GetStateStyle(pod.State).Render(pod.State)

		// Extract namespace name from ID
		namespaceName := "-"
		if pod.NamespaceID != "" {
			// NamespaceID format: namespace-name or c-xxxxx:namespace-name
			parts := strings.Split(pod.NamespaceID, ":")
			if len(parts) > 1 {
				namespaceName = parts[1]
			} else {
				namespaceName = pod.NamespaceID
			}
		}

		rows = append(rows, table.NewRow(table.RowData{
			"name":      pod.Name,
			"namespace": namespaceName,
			"state":     stateStyled,
			"node":      pod.NodeName,
			"restarts":  fmt.Sprintf("%d", pod.RestartCount),
			"age":       age,
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

// updateDeploymentsTable updates the table with deployment data
func (a *App) updateDeploymentsTable() {
	// Define columns
	columns := []table.Column{
		table.NewColumn("name", "NAME", 35),
		table.NewColumn("ready", "READY", 10),
		table.NewColumn("uptodate", "UP-TO-DATE", 10),
		table.NewColumn("available", "AVAILABLE", 10),
		table.NewColumn("age", "AGE", 10),
	}

	// Build rows
	rows := []table.Row{}
	for _, dep := range a.deployments {
		age := formatAge(dep.Created)
		// Logic for state color could be improved, defaulting to simple state check
		stateStyled := GetStateStyle(dep.State).Render(fmt.Sprintf("%d/%d", dep.ReadyReplicas, dep.Replicas))

		rows = append(rows, table.NewRow(table.RowData{
			"name":      dep.Name,
			"ready":     stateStyled,
			"uptodate":  fmt.Sprintf("%d", dep.UpToDateReplicas),
			"available": fmt.Sprintf("%d", dep.AvailableReplicas),
			"age":       age,
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

// updateServicesTable updates the table with service data
func (a *App) updateServicesTable() {
	// Define columns
	columns := []table.Column{
		table.NewColumn("name", "NAME", 35),
		table.NewColumn("type", "TYPE", 15),
		table.NewColumn("clusterip", "CLUSTER-IP", 15),
		table.NewColumn("ports", "PORTS", 25),
		table.NewColumn("age", "AGE", 10),
	}

	// Build rows
	rows := []table.Row{}
	for _, svc := range a.services {
		age := formatAge(svc.Created)

		ports := []string{}
		for _, p := range svc.Ports {
			ports = append(ports, fmt.Sprintf("%d/%s", p.Port, p.Protocol))
		}
		portsStr := strings.Join(ports, ",")

		// Use Kind as Type (e.g., ClusterIP)
		svcType := svc.Kind
		if svcType == "" {
			svcType = svc.Type
		}

		rows = append(rows, table.NewRow(table.RowData{
			"name":      svc.Name,
			"type":      svcType,
			"clusterip": svc.ClusterIP,
			"ports":     portsStr,
			"age":       age,
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

// updateCRDsTable updates the table with CRD data
func (a *App) updateCRDsTable() {
	if len(a.crds) == 0 {
		return
	}

	// Define columns (removed DESCRIPTION column to fix alignment)
	columns := []table.Column{
		table.NewColumn("name", "NAME", 50),
		table.NewColumn("group", "GROUP", 35),
		table.NewColumn("version", "VERSION", 12),
		table.NewColumn("scope", "SCOPE", 12),
		table.NewColumn("age", "AGE", 10),
	}

	// Build rows
	rows := []table.Row{}
	for _, crd := range a.crds {
		age := formatAge(crd.Metadata.CreationTimestamp)
		version := "-"

		if len(crd.Spec.Versions) > 0 {
			// Find best version
			for _, v := range crd.Spec.Versions {
				if v.Served {
					version = v.Name
					break
				}
			}
			// Fallback if no served version found (shouldn't happen but safe)
			if version == "-" {
				version = crd.Spec.Versions[0].Name
			}
		}

		rows = append(rows, table.NewRow(table.RowData{
			"name":    crd.Metadata.Name,
			"group":   crd.Spec.Group,
			"version": version,
			"scope":   crd.Spec.Scope,
			"age":     age,
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

// updateCRDInstancesTable updates the table with generic CRD instance data
func (a *App) updateCRDInstancesTable() {
	if len(a.crdInstances) == 0 {
		return
	}

	// Define columns - generic since we don't know schema
	columns := []table.Column{
		table.NewColumn("name", "NAME", 40),
		table.NewColumn("namespace", "NAMESPACE", 25),
		table.NewColumn("age", "AGE", 10),
	}

	// Build rows
	rows := []table.Row{}
	for _, instance := range a.crdInstances {
		// Extract metadata
		name := "-"
		namespace := "-"
		age := "-"

		if metadata, ok := instance["metadata"].(map[string]interface{}); ok {
			if n, ok := metadata["name"].(string); ok {
				name = n
			}
			if ns, ok := metadata["namespace"].(string); ok {
				namespace = ns
			}
			if tsStr, ok := metadata["creationTimestamp"].(string); ok {
				if ts, err := time.Parse(time.RFC3339, tsStr); err == nil {
					age = formatAge(ts)
				}
			}
		}

		rows = append(rows, table.NewRow(table.RowData{
			"name":      name,
			"namespace": namespace,
			"age":       age,
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
  1/2/3        Switch views (Pods/Deploy/Svc)

ACTIONS
  Enter        Navigate into selected resource
  Esc          Go back to previous view
  r / Ctrl+R   Refresh current view
  Shift+C      Switch to CRD Explorer (from Cluster/Project)
  i            Toggle CRD description (in CRD view)
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
		return fmt.Sprintf("Cluster: %s > Project: %s > Namespace: %s > Pods",
			a.currentView.clusterName, a.currentView.projectName, a.currentView.namespaceName)
	case ViewDeployments:
		return fmt.Sprintf("Cluster: %s > Project: %s > Namespace: %s > Deployments",
			a.currentView.clusterName, a.currentView.projectName, a.currentView.namespaceName)
	case ViewServices:
		return fmt.Sprintf("Cluster: %s > Project: %s > Namespace: %s > Services",
			a.currentView.clusterName, a.currentView.projectName, a.currentView.namespaceName)
	case ViewCRDs:
		return fmt.Sprintf("Cluster: %s > Custom Resource Definitions", a.currentView.clusterName)
	case ViewCRDInstances:
		return fmt.Sprintf("Cluster: %s > CRD: %s.%s (%s)",
			a.currentView.clusterName, a.currentView.crdResource, a.currentView.crdGroup, a.currentView.crdKind)
	default:
		return "r9s"
	}
}

// getCRDDescriptionCaption returns the description caption for the selected CRD
func (a *App) getCRDDescriptionCaption() string {
	if a.table.HighlightedRow().Data == nil {
		return ""
	}

	crdName := a.table.HighlightedRow().Data["name"].(string)

	// Find the CRD
	for _, crd := range a.crds {
		if crd.Metadata.Name == crdName {
			description := ""
			// Find served version description
			if len(crd.Spec.Versions) > 0 {
				for _, v := range crd.Spec.Versions {
					if v.Served {
						if v.Schema != nil && v.Schema.OpenAPIV3Schema != nil {
							description = v.Schema.OpenAPIV3Schema.Description
						}
						break
					}
				}
			}

			if description == "" {
				description = "No description available for this CRD."
			}

			// Create caption box
			captionStyle := lipgloss.NewStyle().
				Foreground(colorWhite).
				Background(colorDarkGray).
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorCyan).
				Width(a.width - 4)

			title := lipgloss.NewStyle().
				Bold(true).
				Foreground(colorCyan).
				Render("Description: ")

			return captionStyle.Render(title + description)
		}
	}

	return ""
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
	case ViewNamespaces:
		count = len(a.namespaces)
		resourceType = "namespaces"
	case ViewPods:
		count = len(a.pods)
		resourceType = "pods"
	case ViewDeployments:
		count = len(a.deployments)
		resourceType = "deployments"
	case ViewServices:
		count = len(a.services)
		resourceType = "services"
	case ViewCRDs:
		count = len(a.crds)
		resourceType = "CRDs"
	case ViewCRDInstances:
		count = len(a.crdInstances)
		resourceType = strings.ToLower(a.currentView.crdKind) + "s"
	}

	navHelp := ""
	if len(a.viewStack) > 0 {
		navHelp = " | 'Esc' to go back"
	}

	// Add CRD-specific help
	crdHelp := ""
	if a.currentView.viewType == ViewCRDs {
		crdHelp = " | 'i' for description"
	}

	return fmt.Sprintf(" %d %s%s | '?' for help | 'q' to quit | 'r' to refresh | Shift+C for CRDs%s ",
		count, resourceType, navHelp, crdHelp)
}

// refreshCurrentView refreshes data for the current view
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
	case ViewDeployments:
		return a.fetchDeployments(a.currentView.projectID, a.currentView.namespaceName)
	case ViewServices:
		return a.fetchServices(a.currentView.projectID, a.currentView.namespaceName)
	case ViewCRDs:
		return a.fetchCRDs(a.currentView.clusterID)
	case ViewCRDInstances:
		return a.fetchCRDInstances(
			a.currentView.clusterID,
			a.currentView.crdGroup,
			a.currentView.crdVersion,
			a.currentView.crdResource,
			"", // TODO: support namespace filtering
		)
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

	case ViewProjects:
		// Get selected project
		selected := a.table.HighlightedRow()
		if selected.Data == nil {
			return nil
		}

		projectName := selected.Data["name"].(string)
		// Find project by name to get ID
		for _, project := range a.projects {
			displayName := project.DisplayName
			if displayName == "" {
				displayName = project.Name
			}
			if displayName == projectName {
				// Push current view to stack
				a.viewStack = append(a.viewStack, a.currentView)

				// Navigate to namespaces view
				a.currentView = ViewContext{
					viewType:    ViewNamespaces,
					clusterID:   a.currentView.clusterID,
					clusterName: a.currentView.clusterName,
					projectID:   project.ID,
					projectName: displayName,
				}
				a.loading = true
				return a.fetchNamespaces(a.currentView.clusterID, project.ID)
			}
		}

	case ViewNamespaces:
		// Get selected namespace
		selected := a.table.HighlightedRow()
		if selected.Data == nil {
			return nil
		}

		namespaceName := selected.Data["name"].(string)
		// Find namespace by name
		for _, ns := range a.namespaces {
			if ns.Name == namespaceName {
				// Push current view to stack
				a.viewStack = append(a.viewStack, a.currentView)

				// Navigate to pods view
				a.currentView = ViewContext{
					viewType:      ViewPods,
					clusterID:     a.currentView.clusterID,
					clusterName:   a.currentView.clusterName,
					projectID:     a.currentView.projectID,
					projectName:   a.currentView.projectName,
					namespaceID:   ns.ID,
					namespaceName: ns.Name,
				}
				a.loading = true
				return a.fetchPods(a.currentView.projectID, ns.Name)
			}
		}
	case ViewCRDs:
		// Get selected CRD
		selected := a.table.HighlightedRow()
		if selected.Data == nil {
			return nil
		}

		crdName := selected.Data["name"].(string)
		var selectedCRD rancher.CRD
		for _, crd := range a.crds {
			if crd.Metadata.Name == crdName {
				selectedCRD = crd
				break
			}
		}

		// Push current view
		a.viewStack = append(a.viewStack, a.currentView)

		// Navigate to CRD Instances

		// Find best version (served=true)
		bestVersion := ""
		if len(selectedCRD.Spec.Versions) > 0 {
			bestVersion = selectedCRD.Spec.Versions[0].Name // Default
			for _, v := range selectedCRD.Spec.Versions {
				if v.Served {
					bestVersion = v.Name
					break
				}
			}
		}

		a.currentView = ViewContext{
			viewType:    ViewCRDInstances,
			clusterID:   a.currentView.clusterID,
			clusterName: a.currentView.clusterName,
			crdGroup:    selectedCRD.Spec.Group,
			crdVersion:  bestVersion,
			crdResource: selectedCRD.Spec.Names.Plural,
			crdKind:     selectedCRD.Spec.Names.Kind,
			crdScope:    selectedCRD.Spec.Scope,
		}

		// For namespaced CRDs, do we want to filter by namespace?
		// Currently, this lists ALL instances across all namespaces if scope is Namespaced.
		// Future improvement: Allow filtering by namespace.
		a.loading = true
		return a.fetchCRDInstances(
			a.currentView.clusterID,
			selectedCRD.Spec.Group,
			bestVersion,
			selectedCRD.Spec.Names.Plural,
			"", // All namespaces
		)
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

		// Add a pseudo-project for unassigned/system namespaces
		systemProject := rancher.Project{
			ID:          clusterID + ":__UNASSIGNED__",
			Name:        "__UNASSIGNED__",
			DisplayName: "System / Unassigned Namespaces",
			ClusterID:   clusterID,
			State:       "active",
		}

		// Prepend the system project to the list
		projects := append([]rancher.Project{systemProject}, collection.Data...)

		// Fetch namespaces to count them
		nsCollection, err := a.client.ListNamespaces(clusterID)
		counts := make(map[string]int)

		if err == nil {
			// Calculate counts
			unassignedID := clusterID + ":__UNASSIGNED__"
			for _, ns := range nsCollection.Data {
				if ns.ProjectID == "" || ns.ProjectID == "null" {
					counts[unassignedID]++
				} else {
					counts[ns.ProjectID]++
				}
			}
		}

		return projectsMsg{projects: projects, namespaceCounts: counts}
	}
}

// fetchNamespaces fetches namespaces for a cluster, filtered by project
func (a *App) fetchNamespaces(clusterID, projectID string) tea.Cmd {
	return func() tea.Msg {
		if a.client == nil {
			return errMsg{fmt.Errorf("client not initialized")}
		}

		collection, err := a.client.ListNamespaces(clusterID)
		if err != nil {
			return errMsg{err}
		}

		// Filter namespaces by project ID
		filteredNamespaces := []rancher.Namespace{}

		// Check if this is the special unassigned project
		isUnassigned := strings.HasSuffix(projectID, ":__UNASSIGNED__")

		for _, ns := range collection.Data {
			if isUnassigned {
				// Show namespaces with no project or system namespaces
				if ns.ProjectID == "" || ns.ProjectID == "null" {
					filteredNamespaces = append(filteredNamespaces, ns)
				}
			} else {
				// Exact match on ProjectID
				if ns.ProjectID == projectID {
					filteredNamespaces = append(filteredNamespaces, ns)
				}
			}
		}

		return namespacesMsg{namespaces: filteredNamespaces}
	}
}

// fetchPods fetches pods for a project, filtered by namespace
func (a *App) fetchPods(projectID, namespaceName string) tea.Cmd {
	return func() tea.Msg {
		if a.client == nil {
			return errMsg{fmt.Errorf("client not initialized")}
		}

		collection, err := a.client.ListPods(projectID)
		if err != nil {
			return errMsg{err}
		}

		// Filter pods by namespace name
		filteredPods := []rancher.Pod{}
		for _, pod := range collection.Data {
			// Extract namespace from NamespaceID
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

// fetchDeployments fetches deployments for a project, filtered by namespace
func (a *App) fetchDeployments(projectID, namespaceName string) tea.Cmd {
	return func() tea.Msg {
		if a.client == nil {
			return errMsg{fmt.Errorf("client not initialized")}
		}

		collection, err := a.client.ListDeployments(projectID)
		if err != nil {
			return errMsg{err}
		}

		// Filter by namespace
		filtered := []rancher.Deployment{}
		for _, dep := range collection.Data {
			depNamespace := dep.NamespaceID
			if strings.Contains(depNamespace, ":") {
				parts := strings.Split(depNamespace, ":")
				if len(parts) > 1 {
					depNamespace = parts[1]
				}
			}

			if depNamespace == namespaceName {
				filtered = append(filtered, dep)
			}
		}

		return deploymentsMsg{deployments: filtered}
	}
}

// fetchServices fetches services for a project, filtered by namespace
func (a *App) fetchServices(projectID, namespaceName string) tea.Cmd {
	return func() tea.Msg {
		if a.client == nil {
			return errMsg{fmt.Errorf("client not initialized")}
		}

		collection, err := a.client.ListServices(projectID)
		if err != nil {
			return errMsg{err}
		}

		// Filter by namespace
		filtered := []rancher.Service{}
		for _, svc := range collection.Data {
			svcNamespace := svc.NamespaceID
			if strings.Contains(svcNamespace, ":") {
				parts := strings.Split(svcNamespace, ":")
				if len(parts) > 1 {
					svcNamespace = parts[1]
				}
			}

			if svcNamespace == namespaceName {
				filtered = append(filtered, svc)
			}
		}

		return servicesMsg{services: filtered}
	}
}

// fetchCRDs fetches CRDs for a cluster
func (a *App) fetchCRDs(clusterID string) tea.Cmd {
	return func() tea.Msg {
		if a.client == nil {
			return errMsg{fmt.Errorf("client not initialized")}
		}

		collection, err := a.client.ListCRDs(clusterID)
		if err != nil {
			return errMsg{err}
		}

		return crdsMsg{crds: collection.Items}
	}
}

// fetchCRDInstances fetches instances of a CRD
func (a *App) fetchCRDInstances(clusterID, group, version, resource, namespace string) tea.Cmd {
	return func() tea.Msg {
		if a.client == nil {
			return errMsg{fmt.Errorf("client not initialized")}
		}

		collection, err := a.client.ListCustomResources(clusterID, group, version, resource, namespace)
		if err != nil {
			// Handle 404 specifically
			if strings.Contains(err.Error(), "404") {
				return errMsg{fmt.Errorf("Failed to list resources. API endpoint not found (404).\nThe CRD version '%s' might not be served or the resource path is incorrect.\nPlease create an issue if this persists.", version)}
			}
			return errMsg{err}
		}

		return crdInstancesMsg{instances: collection.Items}
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

// isNamespaceResourceView checks if current view is a namespace-level resource view
func (a *App) isNamespaceResourceView() bool {
	return a.currentView.viewType == ViewPods ||
		a.currentView.viewType == ViewDeployments ||
		a.currentView.viewType == ViewServices
}
