// Package tui implements the terminal user interface for r8s using the Bubble Tea framework.
// It provides an interactive, keyboard-driven interface for navigating Rancher clusters, projects,
// namespaces, and Kubernetes resources. The package handles view rendering, state management,
// and user input processing.
package tui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"

	"github.com/Rancheroo/r8s/internal/config"
	"github.com/Rancheroo/r8s/internal/datasource"
	"github.com/Rancheroo/r8s/internal/rancher"
)

// safeRowString safely extracts a string value from table row data.
// Returns empty string if key doesn't exist or value is nil/wrong type.
// This prevents panics from nil interface conversions in bundle mode.
func safeRowString(rowData table.RowData, key string) string {
	if rowData == nil {
		return ""
	}
	val, exists := rowData[key]
	if !exists || val == nil {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

// ViewType represents different view types
type ViewType int

const (
	ViewAttention ViewType = iota // Attention Dashboard (default root view)
	ViewClusters
	ViewProjects
	ViewNamespaces
	ViewPods
	ViewDeployments
	ViewServices
	ViewCRDs
	ViewCRDInstances
	ViewLogs
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
	// Context for logs
	podName       string
	containerName string
}

// App represents the main TUI application
type App struct {
	config     *config.Config
	client     *rancher.Client
	dataSource datasource.DataSource // Abstracted data source (live or bundle)
	width      int
	height     int

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
	logs         []string // Log lines for current pod

	projectNamespaceCounts map[string]int

	// UI state
	table              table.Model
	logViewport        viewport.Model
	error              string
	loading            bool
	showHelp           bool
	showCRDDescription bool
	showingDescribe    bool
	describeContent    string
	describeTitle      string

	// Log search state
	searchMode    bool
	searchQuery   string
	searchMatches []int // Line numbers with matches
	currentMatch  int   // Current match index

	// Log viewing state
	currentContainer string   // Current container being viewed
	containers       []string // Available containers for current pod
	tailMode         bool     // Auto-refresh tail mode
	filterLevel      string   // Log level filter: "", "ERROR", "WARN", "INFO"
	showPrevious     bool     // Show previous logs (for crashed containers)
	wordWrap         bool     // Enable word wrapping for long log lines

	// App state
	offlineMode bool   // Flag to indicate running without live Rancher connection
	bundleMode  bool   // Flag to indicate bundle mode
	bundlePath  string // Path to loaded bundle

	// Attention Dashboard
	attentionItems []AttentionItem // Detected issues for attention dashboard

	// Selection preservation
	savedRowName string // Saved row name when navigating away
}

// NewApp creates a new TUI application
func NewApp(cfg *config.Config, bundlePath string) *App {
	// Get current profile
	profile, err := cfg.GetCurrentProfile()
	if err != nil {
		return &App{
			config: cfg,
			error:  fmt.Sprintf("Failed to get profile: %v", err),
		}
	}

	// Determine data source based on mode
	var ds datasource.DataSource
	var client *rancher.Client
	var bundleMode bool
	var offlineMode bool

	if bundlePath != "" {
		// Bundle mode - load bundle as data source
		bds, err := datasource.NewBundleDataSource(bundlePath, cfg.Verbose)
		if err != nil {
			// Provide helpful error message based on common issues
			errorMsg := fmt.Sprintf("Failed to load log bundle from: %s\n\n%v\n\n", bundlePath, err)
			errorMsg += "Common solutions:\n"
			errorMsg += "  • Ensure the path points to an extracted bundle directory\n"
			errorMsg += "  • Check that the bundle contains an rke2/ directory\n"
			errorMsg += "  • Verify the bundle structure: kubectl/, podlogs/, etc.\n"
			errorMsg += "  • See docs/BUNDLE-FORMAT.md for details\n"
			errorMsg += "\nUse --verbose flag for more details"

			return &App{
				config: cfg,
				error:  errorMsg,
			}
		}
		ds = bds
		bundleMode = true
		offlineMode = false // Bundle mode is not "offline", it's bundle analysis
	} else if cfg.MockMode {
		// Demo/Mock mode - explicitly requested via --mockdata flag
		// Uses the example bundle from the repo
		eds, err := datasource.NewEmbeddedDataSource(cfg.Verbose)
		if err != nil {
			return &App{
				config: cfg,
				error: fmt.Sprintf(
					"Failed to load demo bundle: %v\n\n"+
						"The demo bundle may be missing from the repo.\n"+
						"Try using --bundle with the example-log-bundle/ directory instead.",
					err,
				),
			}
		}
		ds = eds
		offlineMode = true // Display as offline in UI
		bundleMode = false // It's demo, not user bundle
	} else {
		// Live mode - use Rancher client
		client = rancher.NewClient(
			profile.URL,
			profile.GetToken(),
			cfg.Insecure || profile.Insecure,
		)

		// Test connection - fail hard if it doesn't work
		if err := client.TestConnection(); err != nil {
			return &App{
				config: cfg,
				error: fmt.Sprintf(
					"Cannot connect to Rancher API at %s\n\n"+
						"Error: %v\n\n"+
						"Options:\n"+
						"  • Check RANCHER_URL and RANCHER_TOKEN\n"+
						"  • Use --mockdata flag for demo mode\n"+
						"  • Use --bundle flag to analyze log bundles\n"+
						"  • Run 'r8s config init' to set up configuration",
					profile.URL, err,
				),
			}
		}

		ds = datasource.NewLiveDataSource(client)
		offlineMode = false
	}

	// Always start at Clusters view regardless of connection status
	var initialView ViewContext = ViewContext{viewType: ViewClusters}

	return &App{
		config:      cfg,
		client:      client,
		dataSource:  ds,
		offlineMode: offlineMode,
		bundleMode:  bundleMode,
		bundlePath:  bundlePath,
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

		// FIX BUG #7: Handle search input BEFORE regular hotkeys
		// This prevents hotkeys from triggering when typing in search mode
		if a.searchMode && a.currentView.viewType == ViewLogs {
			switch msg.String() {
			case "esc":
				// FIX BUG #3: Restore filter state when exiting search
				a.searchMode = false
				a.searchQuery = ""
				a.searchMatches = nil
				a.currentMatch = -1
				// Re-apply any active log filter to restore filtered view
				a.applyLogFilter()
				return a, nil
			case "enter":
				a.searchMode = false
				a.performSearch()
				return a, nil
			case "backspace":
				if len(a.searchQuery) > 0 {
					a.searchQuery = a.searchQuery[:len(a.searchQuery)-1]
				}
				return a, nil
			default:
				// Add character to search query
				if len(msg.String()) == 1 {
					a.searchQuery += msg.String()
				}
				return a, nil
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case "r", "ctrl+r", "ctrl+l":
			// FIX BUG #7: Handle Ctrl+L to refresh (prevent terminal clear conflicts)
			a.loading = true
			return a, a.refreshCurrentView()
		case "j":
			// FIX BUG #14: Vim-style navigation down
			if !a.searchMode && a.currentView.viewType != ViewLogs {
				newTable, cmd := a.table.Update(tea.KeyMsg{Type: tea.KeyDown})
				a.table = newTable
				return a, cmd
			}
		case "k":
			// FIX BUG #14: Vim-style navigation up
			if !a.searchMode && a.currentView.viewType != ViewLogs {
				newTable, cmd := a.table.Update(tea.KeyMsg{Type: tea.KeyUp})
				a.table = newTable
				return a, cmd
			}
		case "?":
			a.showHelp = true
			return a, nil
		case "enter":
			return a, a.handleEnter()
		case "esc", "b":
			// Universal back navigation - 'b' and 'Esc' do the same thing
			if a.showingDescribe {
				// Exit describe view
				a.showingDescribe = false
				a.describeContent = ""
				a.describeTitle = ""
				return a, nil
			}
			// FIX 5: Check search mode BEFORE view stack (priority fix)
			if a.searchMode {
				// Exit search mode without exiting view
				a.searchMode = false
				a.searchQuery = ""
				a.searchMatches = nil
				a.currentMatch = -1
				return a, nil
			}
			if len(a.viewStack) > 0 {
				// Pop from view stack
				// FIX 6: Clean search state when exiting view
				a.searchMode = false
				a.searchQuery = ""
				a.searchMatches = nil
				a.currentMatch = -1

				// Save current selection before navigating back
				// Store the row's primary key (name) so we can restore position after refresh
				if row := a.table.HighlightedRow(); row.Data != nil {
					a.savedRowName = safeRowString(row.Data, "name")
				}

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
					name := safeRowString(a.table.HighlightedRow().Data, "name")
					if name == "" {
						return a, nil
					}
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
		case "l":
			// Open logs view for selected pod
			if a.currentView.viewType == ViewPods {
				return a, a.handleViewLogs()
			}
		case "t":
			// Toggle tail mode in logs view
			if a.currentView.viewType == ViewLogs {
				a.tailMode = !a.tailMode
				if a.tailMode {
					// Start tail mode - position at bottom
					a.logViewport.GotoBottom()
					return a, a.tickTail()
				}
				return a, nil
			}
		case "c":
			// Cycle through containers in logs view
			if a.currentView.viewType == ViewLogs && len(a.containers) > 1 {
				return a, a.cycleContainer()
			}
		case "ctrl+e":
			// Filter to ERROR logs only
			if a.currentView.viewType == ViewLogs {
				if a.filterLevel == "ERROR" {
					a.filterLevel = "" // Toggle off
				} else {
					a.filterLevel = "ERROR"
				}
				// FIX BUG #10: Clear search state when filter changes (prevents stale match indices)
				a.searchMatches = nil
				a.currentMatch = -1
				a.applyLogFilter()
				return a, nil
			}
		case "ctrl+w":
			// Filter to WARN/ERROR logs
			if a.currentView.viewType == ViewLogs {
				if a.filterLevel == "WARN" {
					a.filterLevel = "" // Toggle off
				} else {
					a.filterLevel = "WARN"
				}
				// FIX BUG #10: Clear search state when filter changes
				a.searchMatches = nil
				a.currentMatch = -1
				a.applyLogFilter()
				return a, nil
			}
		case "ctrl+a":
			// Show all logs (clear filter)
			if a.currentView.viewType == ViewLogs {
				a.filterLevel = ""
				// FIX BUG #10: Clear search state when filter changes
				a.searchMatches = nil
				a.currentMatch = -1
				a.applyLogFilter()
				return a, nil
			}
		case "ctrl+p":
			// Toggle previous logs in logs view
			if a.currentView.viewType == ViewLogs {
				a.showPrevious = !a.showPrevious
				a.loading = true
				return a, a.fetchLogs(a.currentView.clusterID, a.currentView.namespaceName, a.currentView.podName)
			}
		case "/":
			// Enter search mode in logs view
			if a.currentView.viewType == ViewLogs && !a.searchMode {
				a.searchMode = true
				a.searchQuery = ""
				a.searchMatches = nil
				a.currentMatch = -1
				return a, nil
			}
		case "n":
			// Next match in search
			if a.currentView.viewType == ViewLogs && len(a.searchMatches) > 0 {
				a.currentMatch = (a.currentMatch + 1) % len(a.searchMatches)
				a.logViewport.SetContent(a.renderLogsWithColors())
				a.logViewport.GotoTop()
				for i := 0; i < a.searchMatches[a.currentMatch]; i++ {
					a.logViewport.LineDown(1)
				}
				return a, nil
			}
		case "N":
			// Previous match in search
			if a.currentView.viewType == ViewLogs && len(a.searchMatches) > 0 {
				a.currentMatch--
				if a.currentMatch < 0 {
					a.currentMatch = len(a.searchMatches) - 1
				}
				a.logViewport.SetContent(a.renderLogsWithColors())
				a.logViewport.GotoTop()
				for i := 0; i < a.searchMatches[a.currentMatch]; i++ {
					a.logViewport.LineDown(1)
				}
				return a, nil
			}
		case "g":
			// Jump to first log line (vim muscle memory)
			if a.currentView.viewType == ViewLogs && !a.searchMode {
				a.logViewport.GotoTop()
				return a, nil
			}
		case "G":
			// Jump to last log line (vim muscle memory)
			if a.currentView.viewType == ViewLogs && !a.searchMode {
				a.logViewport.GotoBottom()
				return a, nil
			}
		case "w":
			// Toggle word wrap in logs view
			if a.currentView.viewType == ViewLogs && !a.searchMode {
				a.wordWrap = !a.wordWrap
				// Re-render with new wrap setting
				a.logViewport.SetContent(a.renderLogsWithColors())
				return a, nil
			}
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// FIX BUG #11: Resize log viewport on window resize
		if a.currentView.viewType == ViewLogs {
			a.logViewport.Width = a.width - 4
			a.logViewport.Height = a.height - 6
		}
		a.updateTable()

	case clustersMsg:
		a.loading = false
		a.clusters = msg.clusters
		a.error = ""
		a.updateTable()
		a.restoreSelection()

	case projectsMsg:
		a.loading = false
		a.projects = msg.projects
		a.projectNamespaceCounts = msg.namespaceCounts
		a.error = ""
		a.updateTable()
		a.restoreSelection()

	case namespacesMsg:
		a.loading = false
		a.namespaces = msg.namespaces
		a.error = ""
		a.updateTable()
		a.restoreSelection()

	case podsMsg:
		a.loading = false
		a.pods = msg.pods
		a.error = ""
		a.updateTable()
		a.restoreSelection()

	case deploymentsMsg:
		a.loading = false
		a.deployments = msg.deployments
		a.error = ""
		a.updateTable()
		a.restoreSelection()

	case servicesMsg:
		a.loading = false
		a.services = msg.services
		a.error = ""
		a.updateTable()
		a.restoreSelection()

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

	case logsMsg:
		a.loading = false
		a.logs = msg.logs
		a.error = ""

		// Initialize viewport for logs view with colored content
		a.logViewport = viewport.New(a.width-4, a.height-6)
		a.logViewport.SetContent(a.renderLogsWithColors())

	case tailTickMsg:
		// Handle tail mode tick - fetch new logs and schedule next tick
		if a.tailMode && a.currentView.viewType == ViewLogs {
			return a, tea.Batch(
				a.fetchLogs(a.currentView.clusterID, a.currentView.namespaceName, a.currentView.podName),
				a.tickTail(), // Schedule next tick
			)
		}

	case attentionMsg:
		a.loading = false
		a.attentionItems = msg.items
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

	// Update viewport if in logs view
	if a.currentView.viewType == ViewLogs {
		newViewport, cmd := a.logViewport.Update(msg)
		a.logViewport = newViewport
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return a, tea.Batch(cmds...)
}

// View renders the application - simplified for now
func (a *App) View() string {
	if a.error != "" {
		return errorStyle.Render(fmt.Sprintf("Error: %s\n\nPress 'q' to quit", a.error))
	}

	if a.loading {
		// FIX BUG #4: Show appropriate loading message for each mode
		loadingMsg := "Loading..."
		if a.bundleMode {
			loadingMsg = "Loading bundle data..."
		} else if a.offlineMode {
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

	// Special rendering for logs view
	if a.currentView.viewType == ViewLogs {
		return a.renderLogsView()
	}

	// Build view components
	breadcrumb := breadcrumbStyle.Render(a.getBreadcrumb())
	statusText := a.getStatusText()
	status := statusStyle.Render(statusText)

	// Render table
	tableView := a.table.View()

	// Build the view with optional offline warning banner
	var components []string
	components = append(components, breadcrumb)

	// Add offline warning banner if in offline mode
	if a.offlineMode {
		warningBanner := offlineWarningStyle.Render("⚠️  OFFLINE MODE - DISPLAYING MOCK DATA  ⚠️")
		components = append(components, "", warningBanner)
	}

	components = append(components, "", tableView)

	// Add description caption if in CRD view and toggled on
	if a.currentView.viewType == ViewCRDs && a.showCRDDescription {
		caption := a.getCRDDescriptionCaption()
		components = append(components, "", caption)
	}

	components = append(components, "", status)

	// Join all components
	return lipgloss.JoinVertical(lipgloss.Left, components...)
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

// renderLogsView renders the logs view for a pod with viewport scrolling
func (a *App) renderLogsView() string {
	// Build breadcrumb
	breadcrumb := breadcrumbStyle.Render(a.getBreadcrumb())

	// Build status text with search info
	var statusText string
	if a.searchMode {
		statusText = fmt.Sprintf(" Search: %s_ | Press 'Enter' to search, 'Esc' to cancel ", a.searchQuery)
	} else if len(a.searchMatches) > 0 {
		// Show visible log count (respecting filters) in search results
		visibleLogs := a.getVisibleLogs()
		statusText = fmt.Sprintf(" %d lines | Match %d/%d | 'n'=next 'N'=prev '/'=new Esc=clear | q=quit ",
			len(visibleLogs), a.currentMatch+1, len(a.searchMatches))
	} else {
		statusText = a.getStatusText()
	}
	status := statusStyle.Render(statusText)

	// Use viewport for scrollable logs - it already has the content set
	viewportContent := a.logViewport.View()

	// Create bordered box around the viewport
	logsBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorCyan).
		Width(a.width - 4).
		Render(viewportContent)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		breadcrumb,
		"",
		logsBox,
		"",
		status,
	)
}

// updateTable updates the table with current view data - handles all view types
func (a *App) updateTable() {
	switch a.currentView.viewType {
	case ViewCRDs:
		if len(a.crds) > 0 {
			columns := []table.Column{
				table.NewColumn("name", "NAME", 35),
				table.NewColumn("group", "GROUP", 25),
				table.NewColumn("kind", "KIND", 18),
				table.NewColumn("scope", "SCOPE", 12),
				table.NewColumn("instances", "INSTANCES", 10),
			}

			rows := []table.Row{}
			for _, crd := range a.crds {
				// Get instance count for this CRD
				instanceCount := a.getCRDInstanceCount(crd.Spec.Group, crd.Spec.Names.Plural)

				rows = append(rows, table.NewRow(table.RowData{
					"name":      crd.Metadata.Name,
					"group":     crd.Spec.Group,
					"kind":      crd.Spec.Names.Kind,
					"scope":     crd.Spec.Scope,
					"instances": fmt.Sprintf("%d", instanceCount),
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

				// Get node name with fallback support
				nodeName := a.getPodNodeName(pod)

				rows = append(rows, table.NewRow(table.RowData{
					"name":      pod.Name,
					"namespace": namespaceName,
					"state":     pod.State,
					"node":      nodeName,
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

	case ViewDeployments:
		if len(a.deployments) > 0 {
			columns := []table.Column{
				table.NewColumn("name", "NAME", 35),
				table.NewColumn("namespace", "NAMESPACE", 20),
				table.NewColumn("ready", "READY", 12),
				table.NewColumn("uptodate", "UP-TO-DATE", 12),
				table.NewColumn("available", "AVAILABLE", 12),
			}

			rows := []table.Row{}
			for _, deployment := range a.deployments {
				namespaceName := "default"
				if deployment.NamespaceID != "" {
					if strings.Contains(deployment.NamespaceID, ":") {
						parts := strings.Split(deployment.NamespaceID, ":")
						if len(parts) > 1 {
							namespaceName = parts[1]
						}
					} else {
						namespaceName = deployment.NamespaceID
					}
				}

				// Get replica counts - prefer Scale field, fallback to direct fields
				var totalReplicas, readyReplicas, updatedReplicas, availableReplicas int

				if deployment.Scale != nil {
					// Use Scale field if available
					totalReplicas = deployment.Scale.Scale
					readyReplicas = deployment.Scale.Ready
					availableReplicas = deployment.Scale.Ready // Scale.Ready represents available
					updatedReplicas = deployment.Scale.Ready   // Assume updated = ready
				} else {
					// Fallback to direct fields
					totalReplicas = deployment.Replicas
					readyReplicas = deployment.ReadyReplicas
					availableReplicas = deployment.AvailableReplicas
					// Try both possible field names for updated replicas
					if deployment.UpToDateReplicas > 0 {
						updatedReplicas = deployment.UpToDateReplicas
					} else {
						updatedReplicas = deployment.UpdatedReplicas
					}
				}

				rows = append(rows, table.NewRow(table.RowData{
					"name":      deployment.Name,
					"namespace": namespaceName,
					"ready":     fmt.Sprintf("%d/%d", readyReplicas, totalReplicas),
					"uptodate":  fmt.Sprintf("%d", updatedReplicas),
					"available": fmt.Sprintf("%d", availableReplicas),
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
				WithRows([]table.Row{table.NewRow(table.RowData{"message": "No deployments available"})}).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(false).
				BorderRounded()
		}

	case ViewServices:
		if len(a.services) > 0 {
			columns := []table.Column{
				table.NewColumn("name", "NAME", 30),
				table.NewColumn("namespace", "NAMESPACE", 20),
				table.NewColumn("type", "TYPE", 15),
				table.NewColumn("cluster_ip", "CLUSTER-IP", 18),
				table.NewColumn("ports", "PORT(S)", 20),
			}

			rows := []table.Row{}
			for _, service := range a.services {
				namespaceName := "default"
				if service.NamespaceID != "" {
					if strings.Contains(service.NamespaceID, ":") {
						parts := strings.Split(service.NamespaceID, ":")
						if len(parts) > 1 {
							namespaceName = parts[1]
						}
					} else {
						namespaceName = service.NamespaceID
					}
				}

				// Format ports
				var portStrings []string
				for _, port := range service.Ports {
					portStr := fmt.Sprintf("%d/%s", port.Port, port.Protocol)
					if port.NodePort > 0 {
						portStr = fmt.Sprintf("%d:%d/%s", port.Port, port.NodePort, port.Protocol)
					}
					portStrings = append(portStrings, portStr)
				}
				portsDisplay := strings.Join(portStrings, ",")

				rows = append(rows, table.NewRow(table.RowData{
					"name":       service.Name,
					"namespace":  namespaceName,
					"type":       service.Kind,
					"cluster_ip": service.ClusterIP,
					"ports":      portsDisplay,
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
				WithRows([]table.Row{table.NewRow(table.RowData{"message": "No services available"})}).
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
	// FIX BUG #8: Add visual mode indicator to breadcrumb
	modeIndicator := "[LIVE] "
	if a.bundleMode {
		modeIndicator = "[BUNDLE] "
	} else if a.offlineMode {
		modeIndicator = "[MOCK] "
	}

	switch a.currentView.viewType {
	case ViewClusters:
		return modeIndicator + "r8s - Clusters"
	case ViewProjects:
		return modeIndicator + fmt.Sprintf("Cluster: %s > Projects", a.currentView.clusterName)
	case ViewNamespaces:
		return modeIndicator + fmt.Sprintf("Cluster: %s > Project: %s > Namespaces",
			a.currentView.clusterName, a.currentView.projectName)
	case ViewPods:
		return modeIndicator + fmt.Sprintf("Cluster: %s > Project: %s > Namespace: %s > Pods",
			a.currentView.clusterName, a.currentView.projectName, a.currentView.namespaceName)
	case ViewDeployments:
		return modeIndicator + fmt.Sprintf("Cluster: %s > Project: %s > Namespace: %s > Deployments",
			a.currentView.clusterName, a.currentView.projectName, a.currentView.namespaceName)
	case ViewServices:
		return modeIndicator + fmt.Sprintf("Cluster: %s > Project: %s > Namespace: %s > Services",
			a.currentView.clusterName, a.currentView.projectName, a.currentView.namespaceName)
	case ViewCRDs:
		return modeIndicator + fmt.Sprintf("Cluster: %s > CRDs", a.currentView.clusterName)
	case ViewCRDInstances:
		return modeIndicator + fmt.Sprintf("Cluster: %s > CRDs > %s", a.currentView.clusterName, a.currentView.crdKind)
	case ViewLogs:
		return modeIndicator + fmt.Sprintf("Cluster: %s > Project: %s > Namespace: %s > Pod: %s > Logs",
			a.currentView.clusterName, a.currentView.projectName, a.currentView.namespaceName, a.currentView.podName)
	default:
		return modeIndicator + "r8s - Rancher Navigator"
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
		status = fmt.Sprintf(" %s%d clusters | Enter=projects 'C'=CRDs 'r'=refresh | '?'=help 'q'=quit ", offlinePrefix, count)

	case ViewProjects:
		count := len(a.projects)
		status = fmt.Sprintf(" %s%d projects | Enter=namespaces 'C'=CRDs 'r'=refresh | '?'=help 'q'=quit ", offlinePrefix, count)

	case ViewNamespaces:
		count := len(a.namespaces)
		status = fmt.Sprintf(" %s%d namespaces | Enter=pods 'r'=refresh | '?'=help 'q'=quit ", offlinePrefix, count)

	case ViewPods:
		count := len(a.pods)
		status = fmt.Sprintf(" %s%d pods | 'l'=logs 'd'=describe '1/2/3'=switch view 'r'=refresh | '?'=help 'q'=quit ", offlinePrefix, count)

	case ViewDeployments:
		count := len(a.deployments)
		status = fmt.Sprintf(" %s%d deployments | 'd'=describe '1/2/3'=switch view 'r'=refresh | '?'=help 'q'=quit ", offlinePrefix, count)

	case ViewServices:
		count := len(a.services)
		status = fmt.Sprintf(" %s%d services | 'd'=describe '1/2/3'=switch view 'r'=refresh | '?'=help 'q'=quit ", offlinePrefix, count)

	case ViewCRDs:
		count := len(a.crds)
		status = fmt.Sprintf(" %s%d CRDs | 'i'=toggle description Enter=instances 'r'=refresh | '?'=help 'q'=quit ", offlinePrefix, count)

	case ViewCRDInstances:
		count := len(a.crdInstances)
		status = fmt.Sprintf(" %s%d %s instances | 'd'=describe(soon) 'r'=refresh | '?'=help 'q'=quit ", offlinePrefix, count, a.currentView.crdKind)

	case ViewLogs:
		// FIX 4: Show visible log count instead of total count
		visibleLogs := a.getVisibleLogs()
		count := len(visibleLogs)
		// Build dynamic status based on active features
		parts := []string{fmt.Sprintf("%d lines", count)}

		if a.tailMode {
			parts = append(parts, "TAIL MODE")
		}
		if a.filterLevel != "" {
			parts = append(parts, fmt.Sprintf("Filter: %s", a.filterLevel))
		}
		if a.showPrevious {
			parts = append(parts, "PREVIOUS LOGS")
		}
		if a.wordWrap {
			parts = append(parts, "Wrap:On")
		}
		if len(a.containers) > 1 {
			parts = append(parts, fmt.Sprintf("Container: %s", a.currentContainer))
		}

		statusInfo := strings.Join(parts, " | ")
		status = fmt.Sprintf(" %s%s | 'w'=wrap 't'=tail Ctrl+E/W/A=filter '/'=search | Esc=back q=quit ", offlinePrefix, statusInfo)

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
	case ViewDeployments:
		return a.fetchDeployments(a.currentView.projectID, a.currentView.namespaceName)
	case ViewServices:
		return a.fetchServices(a.currentView.projectID, a.currentView.namespaceName)
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
		clusterName := safeRowString(selected, "name")
		if clusterName == "" {
			return nil // Skip if name is missing
		}
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
		projectName := safeRowString(selected, "name")
		if projectName == "" {
			return nil // Skip if name is missing
		}
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
		namespaceName := safeRowString(selected, "name")
		if namespaceName == "" {
			return nil // Skip if name is missing
		}
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
		crdName := safeRowString(selected, "name")
		if crdName == "" {
			return nil // Skip if name is missing
		}
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

		// FIX BUG-001: Use helper function to select best CRD version
		// This correctly handles served versions and avoids 404 errors
		storageVersion, err := selectBestCRDVersion(selectedCRD.Spec.Versions)
		if err != nil {
			a.error = fmt.Sprintf("CRD %s: %v", selectedCRD.Metadata.Name, err)
			return nil
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

	selected := a.table.HighlightedRow().Data

	switch a.currentView.viewType {
	case ViewPods:
		podName := safeRowString(selected, "name")
		namespaceName := safeRowString(selected, "namespace")
		if podName == "" || namespaceName == "" {
			return nil // Skip if required fields are missing
		}
		return a.describePod(a.currentView.clusterID, namespaceName, podName)

	case ViewDeployments:
		deploymentName := safeRowString(selected, "name")
		namespaceName := safeRowString(selected, "namespace")
		if deploymentName == "" || namespaceName == "" {
			return nil // Skip if required fields are missing
		}
		return a.describeDeployment(a.currentView.clusterID, namespaceName, deploymentName)

	case ViewServices:
		serviceName := safeRowString(selected, "name")
		namespaceName := safeRowString(selected, "namespace")
		if serviceName == "" || namespaceName == "" {
			return nil // Skip if required fields are missing
		}
		return a.describeService(a.currentView.clusterID, namespaceName, serviceName)

	default:
		// No description available for this resource type
		a.error = "Describe is not yet implemented for this resource type"
		return nil
	}
}

// describePod fetches detailed pod information
func (a *App) describePod(clusterID, namespace, name string) tea.Cmd {
	return func() tea.Msg {
		// Use DataSource interface for describe - works in all modes
		data, err := a.dataSource.DescribePod(clusterID, namespace, name)
		if err != nil {
			return errMsg{fmt.Errorf("failed to describe pod: %w", err)}
		}

		// Format as JSON for display
		jsonBytes, err := json.MarshalIndent(data, "", "  ")
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

// describeDeployment fetches detailed deployment information
func (a *App) describeDeployment(clusterID, namespace, name string) tea.Cmd {
	return func() tea.Msg {
		// Use DataSource interface for describe - works in all modes
		data, err := a.dataSource.DescribeDeployment(clusterID, namespace, name)
		if err != nil {
			return errMsg{fmt.Errorf("failed to describe deployment: %w", err)}
		}

		jsonBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return errMsg{fmt.Errorf("failed to format deployment details: %w", err)}
		}

		content := fmt.Sprintf("Deployment Details (JSON):\n\n%s", string(jsonBytes))

		return describeMsg{
			title:   fmt.Sprintf("Deployment: %s/%s", namespace, name),
			content: content,
		}
	}
}

// describeService fetches detailed service information
func (a *App) describeService(clusterID, namespace, name string) tea.Cmd {
	return func() tea.Msg {
		// Use DataSource interface for describe - works in all modes
		data, err := a.dataSource.DescribeService(clusterID, namespace, name)
		if err != nil {
			return errMsg{fmt.Errorf("failed to describe service: %w", err)}
		}

		jsonBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return errMsg{fmt.Errorf("failed to format service details: %w", err)}
		}

		content := fmt.Sprintf("Service Details (JSON):\n\n%s", string(jsonBytes))

		return describeMsg{
			title:   fmt.Sprintf("Service: %s/%s", namespace, name),
			content: content,
		}
	}
}

// fetchLogs fetches logs for a pod using the data source
func (a *App) fetchLogs(clusterID, namespace, podName string) tea.Cmd {
	return func() tea.Msg {
		// Try to get logs from data source first
		if a.dataSource != nil {
			logs, err := a.dataSource.GetLogs(clusterID, namespace, podName, a.currentContainer, a.showPrevious)
			if err == nil {
				// Return even if empty - empty logs is valid
				return logsMsg{logs: logs}
			}
			// FIX BUG #13: NO SILENT FALLBACK - return error with context
			if a.config.Verbose {
				return errMsg{fmt.Errorf("failed to fetch logs from data source: %w\n\n"+
					"Context: cluster=%s, namespace=%s, pod=%s, container=%s\n"+
					"Hint: Check bundle data or pod status", err, clusterID, namespace, podName, a.currentContainer)}
			}
			return errMsg{fmt.Errorf("failed to fetch logs: %w", err)}
		}

		// Only use mock data if explicitly in mock mode
		if a.offlineMode && a.config.MockMode {
			mockLogs := a.generateMockLogs(podName)
			return logsMsg{logs: mockLogs}
		}

		return errMsg{fmt.Errorf("no data source available")}
	}
}

// generateMockLogs generates realistic mock logs for testing
func (a *App) generateMockLogs(podName string) []string {
	return []string{
		"I1127 00:44:40.476206 [INFO] Kubelet starting up...",
		"I1127 00:44:40.478859 [INFO] None policy: Start",
		"I1127 00:44:40.478889 [INFO] Starting memory manager with policy=None",
		"I1127 00:44:40.479426 [INFO] Initialized iptables rules for IPv6",
		"I1127 00:44:40.479451 [INFO] Starting to sync pod status with apiserver",
		"I1127 00:44:40.479482 [INFO] Starting kubelet main sync loop",
		"E1127 00:44:40.479579 [ERROR] Skipping pod synchronization - PLEG is not healthy",
		"W1127 00:44:40.483015 [WARN] Failed to list RuntimeClass: connection refused",
		"E1127 00:44:40.483087 [ERROR] Unhandled error in reflector: connection refused",
		"I1127 00:44:40.545805 [INFO] Attempting to connect to API server",
		"E1127 00:44:40.545850 [ERROR] Error getting current node from lister: node not found",
		"I1127 00:44:40.586619 [INFO] Failed to read data from checkpoint - checkpoint not found",
		"I1127 00:44:40.586837 [INFO] Eviction manager: starting control loop",
		"I1127 00:44:40.588489 [INFO] Starting Kubelet Plugin Manager",
		"E1127 00:44:40.590954 [ERROR] Eviction manager: failed to check container filesystem",
		"I1127 00:44:40.688184 [INFO] Attempting to register node w-guard-wg-cp-svtk6-lqtxw",
		"E1127 00:44:40.688785 [ERROR] Unable to register node with API server: connection refused",
		"W1127 00:44:40.810298 [WARN] No need to create mirror pod - failed to get node info",
		"I1127 00:44:40.838155 [INFO] VerifyControllerAttachedVolume started for volume file3",
		"I1127 00:44:40.838373 [INFO] VerifyControllerAttachedVolume started for volume file5",
		"I1127 00:44:40.838450 [INFO] VerifyControllerAttachedVolume started for volume file6",
		"I1127 00:44:40.890508 [INFO] Attempting to register node (retry 2)",
		"E1127 00:44:40.890971 [ERROR] Unable to register node with API server: connection refused",
		"E1127 00:44:41.066666 [ERROR] Failed to ensure lease exists - will retry with backoff",
		"W1127 00:44:41.110927 [WARN] Nameserver limits exceeded - some nameservers omitted",
		"I1127 00:44:41.292845 [INFO] Attempting to register node (retry 3)",
		"E1127 00:44:41.293183 [ERROR] Unable to register node with API server: connection refused",
		"W1127 00:44:41.420049 [WARN] Failed to list Services: connection refused",
		"W1127 00:44:41.455586 [WARN] Failed to list RuntimeClass: connection refused",
		"I1127 00:44:42.000000 [INFO] Health check passed for container app-main",
		"I1127 00:44:42.500000 [INFO] Processing HTTP request GET /api/v1/pods",
		"I1127 00:44:42.501234 [INFO] Query execution completed in 15ms",
		"I1127 00:44:42.550000 [INFO] Response sent: 200 OK",
		"W1127 00:44:43.000000 [WARN] Slow query detected: SELECT * FROM pods (duration: 500ms)",
		"I1127 00:44:43.100000 [INFO] Cache invalidated for namespace default",
		"I1127 00:44:43.200000 [INFO] Syncing pod state with etcd",
		"E1127 00:44:43.300000 [ERROR] Connection timeout to metrics server",
		"I1127 00:44:43.400000 [INFO] Retrying connection to metrics server (attempt 1/3)",
		"W1127 00:44:43.500000 [WARN] High memory usage detected: 85% of limit",
		"I1127 00:44:43.600000 [INFO] Garbage collection triggered",
		"I1127 00:44:43.700000 [INFO] Freed 150MB of memory",
		"I1127 00:44:44.000000 [INFO] Pod nginx-deployment-abc123 started successfully",
		"I1127 00:44:44.100000 [INFO] Container nginx pulling image nginx:1.21",
		"I1127 00:44:44.200000 [INFO] Image pull successful",
		"I1127 00:44:44.300000 [INFO] Container nginx started",
		"E1127 00:44:44.400000 [ERROR] Failed to mount volume pvc-data: volume not found",
		"W1127 00:44:44.500000 [WARN] Retrying volume mount with exponential backoff",
		"I1127 00:44:44.800000 [INFO] Volume pvc-data mounted successfully on retry",
		"I1127 00:44:45.000000 [INFO] Readiness probe succeeded for container nginx",
		fmt.Sprintf("I1127 00:44:45.500000 [INFO] Mock logs for pod: %s", podName),
		"I1127 00:44:45.600000 [INFO] All health checks passing",
	}
}

// handleViewLogs navigates to logs view for the selected pod
func (a *App) handleViewLogs() tea.Cmd {
	if a.table.HighlightedRow().Data == nil {
		return nil
	}

	selected := a.table.HighlightedRow().Data
	podName := safeRowString(selected, "name")
	namespaceName := safeRowString(selected, "namespace")
	if podName == "" || namespaceName == "" {
		return nil // Skip if required fields are missing
	}

	// Push current view to stack
	a.viewStack = append(a.viewStack, a.currentView)

	// Navigate to logs view
	a.currentView = ViewContext{
		viewType:      ViewLogs,
		clusterID:     a.currentView.clusterID,
		clusterName:   a.currentView.clusterName,
		projectID:     a.currentView.projectID,
		projectName:   a.currentView.projectName,
		namespaceID:   a.currentView.namespaceID,
		namespaceName: namespaceName,
		podName:       podName,
		containerName: "", // TODO: Support multi-container pods later
	}

	a.loading = true
	return a.fetchLogs(a.currentView.clusterID, namespaceName, podName)
}

// fetchPods fetches pods using the unified data source
func (a *App) fetchPods(projectID, namespaceName string) tea.Cmd {
	return func() tea.Msg {
		if a.dataSource == nil {
			return errMsg{fmt.Errorf("no data source available")}
		}

		pods, err := a.dataSource.GetPods(projectID, namespaceName)
		if err != nil {
			if a.config.Verbose {
				return errMsg{fmt.Errorf("failed to fetch pods: %w\n\n"+
					"Context: projectID=%s, namespace=%s\n"+
					"Hint: Check bundle data or API connectivity", err, projectID, namespaceName)}
			}
			return errMsg{fmt.Errorf("failed to fetch pods: %w", err)}
		}

		return podsMsg{pods: pods}
	}
}

// fetchDeployments fetches deployments using the unified data source
func (a *App) fetchDeployments(projectID, namespaceName string) tea.Cmd {
	return func() tea.Msg {
		if a.dataSource == nil {
			return errMsg{fmt.Errorf("no data source available")}
		}

		deployments, err := a.dataSource.GetDeployments(projectID, namespaceName)
		if err != nil {
			if a.config.Verbose {
				return errMsg{fmt.Errorf("failed to fetch deployments: %w\n\n"+
					"Context: projectID=%s, namespace=%s\n"+
					"Hint: Check bundle data or API connectivity", err, projectID, namespaceName)}
			}
			return errMsg{fmt.Errorf("failed to fetch deployments: %w", err)}
		}

		return deploymentsMsg{deployments: deployments}
	}
}

// fetchServices fetches services using the unified data source
func (a *App) fetchServices(projectID, namespaceName string) tea.Cmd {
	return func() tea.Msg {
		if a.dataSource == nil {
			return errMsg{fmt.Errorf("no data source available")}
		}

		services, err := a.dataSource.GetServices(projectID, namespaceName)
		if err != nil {
			if a.config.Verbose {
				return errMsg{fmt.Errorf("failed to fetch services: %w\n\n"+
					"Context: projectID=%s, namespace=%s\n"+
					"Hint: Check bundle data or API connectivity", err, projectID, namespaceName)}
			}
			return errMsg{fmt.Errorf("failed to fetch services: %w", err)}
		}

		return servicesMsg{services: services}
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

// getMockDeployments generates realistic mock deployment data
func (a *App) getMockDeployments(namespaceName string) []rancher.Deployment {
	return []rancher.Deployment{
		{
			Name:              "nginx-deployment",
			NamespaceID:       namespaceName,
			State:             "active",
			Replicas:          3,
			AvailableReplicas: 3,
			ReadyReplicas:     3,
			UpToDateReplicas:  3,
			Created:           time.Now().Add(-time.Hour * 24),
		},
		{
			Name:              "redis-deployment",
			NamespaceID:       namespaceName,
			State:             "active",
			Replicas:          2,
			AvailableReplicas: 2,
			ReadyReplicas:     2,
			UpToDateReplicas:  2,
			Created:           time.Now().Add(-time.Hour * 48),
		},
		{
			Name:              "api-server",
			NamespaceID:       namespaceName,
			State:             "active",
			Replicas:          5,
			AvailableReplicas: 5,
			ReadyReplicas:     5,
			UpToDateReplicas:  5,
			Created:           time.Now().Add(-time.Hour * 72),
		},
		{
			Name:              "worker-deployment",
			NamespaceID:       namespaceName,
			State:             "updating",
			Replicas:          4,
			AvailableReplicas: 3,
			ReadyReplicas:     3,
			UpToDateReplicas:  1,
			Created:           time.Now().Add(-time.Hour * 12),
		},
	}
}

// getMockServices generates realistic mock service data
func (a *App) getMockServices(namespaceName string) []rancher.Service {
	return []rancher.Service{
		{
			Name:        "nginx-service",
			NamespaceID: namespaceName,
			State:       "active",
			ClusterIP:   "10.43.100.50",
			Kind:        "ClusterIP",
			Ports: []rancher.ServicePort{
				{Name: "http", Protocol: "TCP", Port: 80, TargetPort: 8080},
			},
			Created: time.Now().Add(-time.Hour * 24),
		},
		{
			Name:        "redis-service",
			NamespaceID: namespaceName,
			State:       "active",
			ClusterIP:   "10.43.100.51",
			Kind:        "ClusterIP",
			Ports: []rancher.ServicePort{
				{Name: "redis", Protocol: "TCP", Port: 6379, TargetPort: 6379},
			},
			Created: time.Now().Add(-time.Hour * 48),
		},
		{
			Name:        "api-service",
			NamespaceID: namespaceName,
			State:       "active",
			ClusterIP:   "10.43.100.52",
			Kind:        "NodePort",
			Ports: []rancher.ServicePort{
				{Name: "api", Protocol: "TCP", Port: 8080, TargetPort: 8080, NodePort: 30080},
			},
			Created: time.Now().Add(-time.Hour * 72),
		},
		{
			Name:        "loadbalancer-service",
			NamespaceID: namespaceName,
			State:       "active",
			ClusterIP:   "10.43.100.53",
			Kind:        "LoadBalancer",
			Ports: []rancher.ServicePort{
				{Name: "http", Protocol: "TCP", Port: 80, TargetPort: 8080},
				{Name: "https", Protocol: "TCP", Port: 443, TargetPort: 8443},
			},
			Created: time.Now().Add(-time.Hour * 96),
		},
	}
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

// fetchCRDs fetches CRDs using the unified data source
func (a *App) fetchCRDs(clusterID string) tea.Cmd {
	return func() tea.Msg {
		if a.dataSource == nil {
			return errMsg{fmt.Errorf("no data source available")}
		}

		crds, err := a.dataSource.GetCRDs(clusterID)
		if err != nil {
			if a.config.Verbose {
				return errMsg{fmt.Errorf("failed to fetch CRDs: %w\n\n"+
					"Context: clusterID=%s\n"+
					"Hint: Check bundle data or API connectivity", err, clusterID)}
			}
			return errMsg{fmt.Errorf("failed to fetch CRDs: %w", err)}
		}

		return crdsMsg{crds: crds}
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

// fetchCRDInstances fetches CRD instances using the unified data source
func (a *App) fetchCRDInstances(clusterID, group, version, resource string) tea.Cmd {
	return func() tea.Msg {
		if a.dataSource == nil {
			return errMsg{fmt.Errorf("no data source available")}
		}

		instances, err := a.dataSource.GetCRDInstances(clusterID, group, version, resource)
		if err != nil {
			if a.config.Verbose {
				return errMsg{fmt.Errorf("failed to fetch CRD instances: %w\n\n"+
					"Context: clusterID=%s, group=%s, version=%s, resource=%s\n"+
					"Hint: Check CRD version and API connectivity", err, clusterID, group, version, resource)}
			}
			return errMsg{fmt.Errorf("failed to fetch CRD instances: %w", err)}
		}

		return crdInstancesMsg{instances: instances}
	}
}

// getMockCRDInstances generates mock CRD instance data with varied counts
func (a *App) getMockCRDInstances(group, resource string) []map[string]interface{} {
	now := time.Now()

	// Generate different mock data based on the CRD type
	switch group {
	case "cert-manager.io":
		if resource == "certificates" {
			// 5 certificate instances - common in production
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
					},
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{"type": "Ready", "status": "True"},
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
					},
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{"type": "Ready", "status": "True"},
						},
					},
				},
				{
					"metadata": map[string]interface{}{
						"name":              "web-cert",
						"namespace":         "web",
						"creationTimestamp": now.Add(-time.Hour * 200).Format(time.RFC3339),
					},
					"spec": map[string]interface{}{
						"secretName": "web-tls",
						"dnsNames":   []string{"web.example.com"},
					},
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{"type": "Ready", "status": "True"},
						},
					},
				},
				{
					"metadata": map[string]interface{}{
						"name":              "grafana-cert",
						"namespace":         "monitoring",
						"creationTimestamp": now.Add(-time.Hour * 96).Format(time.RFC3339),
					},
					"spec": map[string]interface{}{
						"secretName": "grafana-tls",
						"dnsNames":   []string{"grafana.example.com"},
					},
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{"type": "Ready", "status": "True"},
						},
					},
				},
				{
					"metadata": map[string]interface{}{
						"name":              "prometheus-cert",
						"namespace":         "monitoring",
						"creationTimestamp": now.Add(-time.Hour * 144).Format(time.RFC3339),
					},
					"spec": map[string]interface{}{
						"secretName": "prometheus-tls",
						"dnsNames":   []string{"prometheus.example.com"},
					},
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{"type": "Ready", "status": "True"},
						},
					},
				},
			}
		}
	case "monitoring.coreos.com":
		if resource == "servicemonitors" {
			// 7 service monitor instances - monitoring setup
			instances := []map[string]interface{}{}
			services := []string{"kube-state-metrics", "prometheus-operator", "node-exporter",
				"grafana", "alertmanager", "prometheus", "blackbox-exporter"}

			for i, svc := range services {
				instances = append(instances, map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":              svc,
						"namespace":         "monitoring",
						"creationTimestamp": now.Add(-time.Hour * time.Duration(24*(i+1))).Format(time.RFC3339),
					},
					"spec": map[string]interface{}{
						"selector": map[string]interface{}{
							"matchLabels": map[string]string{"app": svc},
						},
						"endpoints": []interface{}{
							map[string]interface{}{"port": "metrics", "interval": "30s"},
						},
					},
				})
			}
			return instances
		}
	case "cattle.io":
		if resource == "clusters" {
			// 3 cluster instances - common setup
			return []map[string]interface{}{
				{
					"metadata": map[string]interface{}{
						"name":              "production",
						"creationTimestamp": now.Add(-time.Hour * 720).Format(time.RFC3339),
					},
					"spec": map[string]interface{}{"displayName": "Production Cluster"},
				},
				{
					"metadata": map[string]interface{}{
						"name":              "staging",
						"creationTimestamp": now.Add(-time.Hour * 480).Format(time.RFC3339),
					},
					"spec": map[string]interface{}{"displayName": "Staging Cluster"},
				},
				{
					"metadata": map[string]interface{}{
						"name":              "development",
						"creationTimestamp": now.Add(-time.Hour * 240).Format(time.RFC3339),
					},
					"spec": map[string]interface{}{"displayName": "Development Cluster"},
				},
			}
		}
	case "rio.cattle.io":
		// Few instances - less common CRD
		return []map[string]interface{}{
			{
				"metadata": map[string]interface{}{
					"name":              fmt.Sprintf("%s-1", resource),
					"namespace":         "default",
					"creationTimestamp": now.Add(-time.Hour * 24).Format(time.RFC3339),
				},
				"spec": map[string]interface{}{"field": "value"},
			},
		}
	}

	// Default: 0 instances for unknown CRDs
	return []map[string]interface{}{}
}

// fetchAttention analyzes cluster health and returns attention items
func (a *App) fetchAttention() tea.Cmd {
	return func() tea.Msg {
		if a.dataSource == nil {
			return errMsg{fmt.Errorf("no data source available")}
		}

		// Detect all issues across the cluster
		items := ComputeAttentionItems(a.dataSource)

		return attentionMsg{items: items}
	}
}

// fetchClusters fetches clusters using the unified data source
func (a *App) fetchClusters() tea.Cmd {
	return func() tea.Msg {
		if a.dataSource == nil {
			return errMsg{fmt.Errorf("no data source available")}
		}

		clusters, err := a.dataSource.GetClusters()
		if err != nil {
			if a.config.Verbose {
				return errMsg{fmt.Errorf("failed to fetch clusters: %w\n\n"+
					"Context: DataSource fetch\n"+
					"Hint: Check bundle data or API connectivity", err)}
			}
			return errMsg{fmt.Errorf("failed to fetch clusters: %w", err)}
		}

		return clustersMsg{clusters: clusters}
	}
}

// fetchProjects fetches projects using the unified data source
func (a *App) fetchProjects(clusterID string) tea.Cmd {
	return func() tea.Msg {
		if a.dataSource == nil {
			return errMsg{fmt.Errorf("no data source available")}
		}

		projects, namespaceCounts, err := a.dataSource.GetProjects(clusterID)
		if err != nil {
			if a.config.Verbose {
				return errMsg{fmt.Errorf("failed to fetch projects: %w\n\n"+
					"Context: clusterID=%s\n"+
					"Hint: Check bundle data or API connectivity", err, clusterID)}
			}
			return errMsg{fmt.Errorf("failed to fetch projects: %w", err)}
		}

		return projectsMsg{projects: projects, namespaceCounts: namespaceCounts}
	}
}

// fetchNamespaces fetches namespaces using the unified data source
func (a *App) fetchNamespaces(clusterID, projectID string) tea.Cmd {
	return func() tea.Msg {
		if a.dataSource == nil {
			return errMsg{fmt.Errorf("no data source available")}
		}

		namespaces, err := a.dataSource.GetNamespaces(clusterID, projectID)
		if err != nil {
			if a.config.Verbose {
				return errMsg{fmt.Errorf("failed to fetch namespaces: %w\n\n"+
					"Context: clusterID=%s, projectID=%s\n"+
					"Hint: Check bundle data or API connectivity", err, clusterID, projectID)}
			}
			return errMsg{fmt.Errorf("failed to fetch namespaces: %w", err)}
		}

		a.updateNamespaceCounts(namespaces)
		return namespacesMsg{namespaces: namespaces}
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

// getCRDInstanceCount returns the count of instances for a given CRD using datasource
func (a *App) getCRDInstanceCount(group, resource string) int {
	if a.dataSource == nil {
		return 0
	}

	// Get the storage version for this CRD
	var version string
	for _, crd := range a.crds {
		if crd.Spec.Group == group && crd.Spec.Names.Plural == resource {
			for _, v := range crd.Spec.Versions {
				if v.Storage {
					version = v.Name
					break
				}
			}
			if version == "" && len(crd.Spec.Versions) > 0 {
				version = crd.Spec.Versions[0].Name
			}
			break
		}
	}

	if version == "" {
		return 0
	}

	instances, err := a.dataSource.GetCRDInstances(a.currentView.clusterID, group, version, resource)
	if err != nil {
		return 0 // Silently return 0 for counts (non-critical)
	}

	return len(instances)
}

// restoreSelection restores the previously saved table selection if applicable
// This is called after table updates to maintain user's position when navigating back
func (a *App) restoreSelection() {
	// Note: Full restoration not implemented - bubble-table doesn't provide
	// a way to iterate through rows or set selection by index
	// This would require either:
	// 1. bubble-table library changes to expose rows
	// 2. Maintaining a parallel rows slice ourselves
	// 3. Using a different table library
	// For now, selection resets to top (simple behavior)
	a.savedRowName = "" // Clear any saved state
}

// isNamespaceResourceView returns true if the current view is a namespace-scoped resource view
func (a *App) isNamespaceResourceView() bool {
	return a.currentView.viewType == ViewPods ||
		a.currentView.viewType == ViewDeployments ||
		a.currentView.viewType == ViewServices
}

// getPodNodeName extracts the node name from a Pod with fallback support
func (a *App) getPodNodeName(pod rancher.Pod) string {
	// Try each field in order of preference
	if pod.NodeName != "" {
		return pod.NodeName
	}
	if pod.NodeID != "" {
		return pod.NodeID
	}
	if pod.Node != "" {
		return pod.Node
	}
	if pod.Hostname != "" {
		return pod.Hostname
	}
	// No node information available
	return ""
}

// performSearch searches through logs for the query and populates search matches
// FIX 2: Search through visible (filtered) logs instead of all logs
func (a *App) performSearch() {
	if a.searchQuery == "" {
		return
	}

	// Clear previous matches
	a.searchMatches = nil
	a.currentMatch = -1

	// Get visible logs (respects active filter)
	visibleLogs := a.getVisibleLogs()

	// Search through visible logs (case-insensitive)
	query := strings.ToLower(a.searchQuery)
	for i, line := range visibleLogs {
		if strings.Contains(strings.ToLower(line), query) {
			a.searchMatches = append(a.searchMatches, i)
		}
	}

	// Jump to first match if found
	if len(a.searchMatches) > 0 {
		a.currentMatch = 0
		a.logViewport.SetContent(a.renderLogsWithColors())
		a.logViewport.GotoTop()
		for i := 0; i < a.searchMatches[0]; i++ {
			a.logViewport.LineDown(1)
		}
	}
}

// tickTail returns a command to refresh logs in tail mode
func (a *App) tickTail() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		// FIX BUG #15 REGRESSION: Return tailTickMsg to continue tick chain
		// Cannot invoke cmd() here - breaks event loop
		return tailTickMsg{}
	})
}

// cycleContainer cycles through available containers for the current pod
func (a *App) cycleContainer() tea.Cmd {
	if len(a.containers) == 0 {
		// Initialize mock containers for demonstration
		a.containers = []string{"app", "sidecar", "init"}
		a.currentContainer = a.containers[0]
		return nil
	}

	// Find current container index
	currentIdx := 0
	for i, c := range a.containers {
		if c == a.currentContainer {
			currentIdx = i
			break
		}
	}

	// Move to next container (wrap around)
	nextIdx := (currentIdx + 1) % len(a.containers)
	a.currentContainer = a.containers[nextIdx]

	// In production, would fetch logs for new container
	// For now, just update the display
	return nil
}

// applyLogFilter applies the current log level filter to the logs with colors
func (a *App) applyLogFilter() {
	// Use colored rendering for all content
	a.logViewport.SetContent(a.renderLogsWithColors())
}

// getVisibleLogs returns the currently visible logs based on active filters
// FIX 3: Helper function to get logs respecting current filter state
// Supports both bracketed format ([ERROR], [WARN]) and K8s format (E1120, W1120)
func (a *App) getVisibleLogs() []string {
	if a.filterLevel == "" {
		// No filter - return all logs
		return a.logs
	}

	// Filter logs by level
	var filteredLogs []string
	for _, line := range a.logs {
		include := false

		switch a.filterLevel {
		case "ERROR":
			// Show only ERROR logs - support both formats
			include = isErrorLog(line)
		case "WARN":
			// Show WARN and ERROR logs - support both formats
			include = isWarnLog(line) || isErrorLog(line)
		}

		if include {
			filteredLogs = append(filteredLogs, line)
		}
	}

	return filteredLogs
}

// isErrorLog detects ERROR level logs in both bracketed and K8s formats
func isErrorLog(line string) bool {
	lineUpper := strings.ToUpper(line)
	// Bracketed format: [ERROR]
	if strings.Contains(lineUpper, "[ERROR]") {
		return true
	}
	// K8s format: E1120, E0102, etc. (E followed by 4 digits)
	if len(line) > 5 {
		for i := 0; i < len(line)-5; i++ {
			if line[i] == 'E' && isDigit(line[i+1]) && isDigit(line[i+2]) &&
				isDigit(line[i+3]) && isDigit(line[i+4]) {
				// Check if followed by space or colon
				if i+5 < len(line) && (line[i+5] == ' ' || line[i+5] == ':') {
					return true
				}
			}
		}
	}
	// Also check for level=error format
	if strings.Contains(lineUpper, "LEVEL=ERROR") {
		return true
	}
	return false
}

// isWarnLog detects WARN level logs in both bracketed and K8s formats
func isWarnLog(line string) bool {
	lineUpper := strings.ToUpper(line)
	// Bracketed format: [WARN]
	if strings.Contains(lineUpper, "[WARN]") {
		return true
	}
	// K8s format: W1120, W0102, etc. (W followed by 4 digits)
	if len(line) > 5 {
		for i := 0; i < len(line)-5; i++ {
			if line[i] == 'W' && isDigit(line[i+1]) && isDigit(line[i+2]) &&
				isDigit(line[i+3]) && isDigit(line[i+4]) {
				// Check if followed by space or colon
				if i+5 < len(line) && (line[i+5] == ' ' || line[i+5] == ':') {
					return true
				}
			}
		}
	}
	// Also check for level=warn/warning format
	if strings.Contains(lineUpper, "LEVEL=WARN") {
		return true
	}
	return false
}

// isInfoLog detects INFO level logs in both bracketed and K8s formats
func isInfoLog(line string) bool {
	lineUpper := strings.ToUpper(line)
	// Bracketed format: [INFO]
	if strings.Contains(lineUpper, "[INFO]") {
		return true
	}
	// K8s format: I1120, I0102, etc. (I followed by 4 digits)
	if len(line) > 5 {
		for i := 0; i < len(line)-5; i++ {
			if line[i] == 'I' && isDigit(line[i+1]) && isDigit(line[i+2]) &&
				isDigit(line[i+3]) && isDigit(line[i+4]) {
				// Check if followed by space or colon
				if i+5 < len(line) && (line[i+5] == ' ' || line[i+5] == ':') {
					return true
				}
			}
		}
	}
	// Also check for level=info format
	if strings.Contains(lineUpper, "LEVEL=INFO") {
		return true
	}
	return false
}

// isDebugLog detects DEBUG level logs in both bracketed and K8s formats
func isDebugLog(line string) bool {
	lineUpper := strings.ToUpper(line)
	// Bracketed format: [DEBUG]
	if strings.Contains(lineUpper, "[DEBUG]") {
		return true
	}
	// K8s format: D1120, D0102, etc. (D followed by 4 digits)
	if len(line) > 5 {
		for i := 0; i < len(line)-5; i++ {
			if line[i] == 'D' && isDigit(line[i+1]) && isDigit(line[i+2]) &&
				isDigit(line[i+3]) && isDigit(line[i+4]) {
				// Check if followed by space or colon
				if i+5 < len(line) && (line[i+5] == ' ' || line[i+5] == ':') {
					return true
				}
			}
		}
	}
	// Also check for level=debug format
	if strings.Contains(lineUpper, "LEVEL=DEBUG") {
		return true
	}
	return false
}

// isDigit checks if a byte is an ASCII digit
func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

// Messages
type clustersMsg struct {
	clusters []rancher.Cluster
}

// tailTickMsg is sent periodically when tail mode is active
type tailTickMsg struct{}

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

// logsMsg represents a message containing log data
type logsMsg struct {
	logs []string
}

// attentionMsg represents attention dashboard analysis results
type attentionMsg struct {
	items []AttentionItem
}

// colorizeLogLine applies color styling based on log level
// Supports both bracketed format ([ERROR], [WARN]) and K8s format (E1120, W1120)
func (a *App) colorizeLogLine(line string, lineIndex int) string {
	// Check if this is the current search match
	isCurrentMatch := false
	if len(a.searchMatches) > 0 && a.currentMatch >= 0 && a.currentMatch < len(a.searchMatches) {
		if lineIndex == a.searchMatches[a.currentMatch] {
			isCurrentMatch = true
		}
	}

	// If current search match, highlight the entire line
	if isCurrentMatch {
		return searchMatchStyle.Render(line)
	}

	// Otherwise, colorize by log level using the same detection functions as filtering
	if isErrorLog(line) {
		return logErrorStyle.Render(line)
	}
	if isWarnLog(line) {
		return logWarnStyle.Render(line)
	}
	if isInfoLog(line) {
		return logInfoStyle.Render(line)
	}
	if isDebugLog(line) {
		return logDebugStyle.Render(line)
	}

	// Default: no special styling
	return line
}

// renderLogsWithColors renders logs with color coding and search highlighting
func (a *App) renderLogsWithColors() string {
	visibleLogs := a.getVisibleLogs()

	if !a.wordWrap {
		// No wrapping - colorize and return as-is
		coloredLines := make([]string, len(visibleLogs))
		for i, line := range visibleLogs {
			coloredLines[i] = a.colorizeLogLine(line, i)
		}
		return strings.Join(coloredLines, "\n")
	}

	// Word wrap enabled - wrap long lines to viewport width
	var wrappedLines []string
	wrapWidth := a.logViewport.Width
	if wrapWidth <= 0 {
		wrapWidth = 80 // Fallback width
	}

	for i, line := range visibleLogs {
		colorizedLine := a.colorizeLogLine(line, i)

		// Simple wrapping: break at wrapWidth
		if len(line) <= wrapWidth {
			wrappedLines = append(wrappedLines, colorizedLine)
		} else {
			// Wrap line into multiple lines
			for len(line) > 0 {
				if len(line) <= wrapWidth {
					wrappedLines = append(wrappedLines, a.colorizeLogLine(line, i))
					break
				}
				// Take wrapWidth characters
				wrappedLines = append(wrappedLines, a.colorizeLogLine(line[:wrapWidth], i))
				line = line[wrapWidth:]
			}
		}
	}

	return strings.Join(wrappedLines, "\n")
}

// selectBestCRDVersion selects the best version from a CRD's version list
// Priority: storage+served > storage > first served > error
func selectBestCRDVersion(versions []rancher.CRDVersion) (string, error) {
	var storageVersion string
	var firstServedVersion string

	for _, v := range versions {
		// Track first served version as fallback
		if v.Served && firstServedVersion == "" {
			firstServedVersion = v.Name
		}

		// Prefer storage version if it's also served
		if v.Storage && v.Served {
			return v.Name, nil
		}

		// Track storage version even if not served
		if v.Storage {
			storageVersion = v.Name
		}
	}

	// Fallback 1: Use storage version even if not marked as served
	// (some CRDs have storage=true but don't explicitly mark served)
	if storageVersion != "" {
		return storageVersion, nil
	}

	// Fallback 2: Use first served version
	if firstServedVersion != "" {
		return firstServedVersion, nil
	}

	// No valid version found
	return "", fmt.Errorf("no served versions available")
}

// renderHelp shows comprehensive keybinding reference
func renderHelp() string {
	help := `r8s HELP - KEYBINDINGS

NAVIGATION
  ↑/↓, j/k    Move selection up/down
  Enter       Navigate into selection
  b or Esc    Go back one level
  
ACTIONS
  l           View logs (Pod view)
  d           Describe resource (Pods/Deployments/Services)
  r           Refresh current view
  
VIEW SWITCHING (Namespace Context)
  1           Switch to Pods
  2           Switch to Deployments
  3           Switch to Services
  
CLUSTER VIEWS
  C           Jump to CRDs (from Cluster/Project view)
  i           Toggle CRD description (in CRD view)
  
LOG VIEWING (when viewing logs)
  g           Jump to first line
  G           Jump to last line
  w           Toggle word wrap for long lines
  /           Start search
  n           Next search match
  N           Previous search match
  t           Toggle tail mode (auto-scroll)
  c           Cycle containers (multi-container pods)
  
LOG FILTERS (in log view)
  Ctrl+E      Filter to ERROR only
  Ctrl+W      Filter to WARN and ERROR
  Ctrl+A      Show all logs (clear filter)
  Ctrl+P      Toggle previous container logs
  
GENERAL
  ?           Show/hide this help
  q           Quit application
  Ctrl+C      Force quit
  
Press Esc or ? to close this help`

	return helpStyle.Render(help)
}
