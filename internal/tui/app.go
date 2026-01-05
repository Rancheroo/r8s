// Package tui implements the terminal user interface for r8s using the Bubble Tea framework.
// It provides an interactive, keyboard-driven interface for navigating Rancher clusters, projects,
// namespaces, and Kubernetes resources. The package handles view rendering, state management,
// and user input processing.
package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"

	"github.com/Rancheroo/r8s/internal/config"
	"github.com/Rancheroo/r8s/internal/datasource"
	"github.com/Rancheroo/r8s/internal/rancher"
)

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
	dataSource datasource.DataSource // Abstracted data source (bundle or demo)
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
	showRawLogs      bool     // Show raw logs instead of diagnostic panel (toggle with 'l')

	// App state
	offlineMode bool   // Flag to indicate running without live Rancher connection
	bundleMode  bool   // Flag to indicate bundle mode
	bundlePath  string // Path to loaded bundle

	// Attention Dashboard
	attentionItems    []AttentionItem // Detected issues for attention dashboard
	attentionCursor   int             // Selected item index in dashboard
	attentionViewport viewport.Model  // Scrollable viewport for dashboard
	attentionExpanded bool            // Show all items vs top-20 cap
	expandedItems     map[int]bool    // Which collapsed event items are expanded
	subCursor         int             // Selected pod index within expanded event (-1 = not in sub-nav)

	// Sorting state
	sortMode        SortMode              // Current sort mode (global default)
	sortModes       map[ViewType]SortMode // Per-view sort mode
	cachedPodCounts map[string]PodCounts  // Cached E/W counts (key: "namespace/podname")

	// Selection preservation
	savedRowName string // Saved row name when navigating away
}

// HasError returns true if the app has an initialization error
func (a *App) HasError() bool {
	return a.error != ""
}

// GetError returns the app's error message
func (a *App) GetError() string {
	return a.error
}

// NewApp creates a new TUI application
func NewApp(cfg *config.Config, bundlePath string) *App {
	// Determine data source based on mode
	var ds datasource.DataSource
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
		offlineMode = false
	} else {
		// Default: Demo mode - uses the embedded example bundle from the repo
		// This makes r8s work out-of-the-box with zero configuration
		eds, err := datasource.NewEmbeddedDataSource(cfg.Verbose)
		if err != nil {
			return &App{
				config: cfg,
				error: fmt.Sprintf(
					"Failed to load demo bundle: %v\n\n"+
						"The demo bundle may be missing from the repo.\n"+
						"Try specifying a bundle path: r8s /path/to/bundle",
					err,
				),
			}
		}
		ds = eds
		offlineMode = true
		bundleMode = false
	}

	// Always start with Attention Dashboard (the killer feature)
	initialView := ViewContext{viewType: ViewAttention}

	return &App{
		config:          cfg,
		dataSource:      ds,
		offlineMode:     offlineMode,
		bundleMode:      bundleMode,
		bundlePath:      bundlePath,
		loading:         true,
		currentView:     initialView,
		sortMode:        SortByCount,                 // Default to count-based sorting
		sortModes:       make(map[ViewType]SortMode), // Per-view sort state
		cachedPodCounts: make(map[string]PodCounts),  // Pod E/W count cache
	}
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Add fullscreen command
	cmds = append(cmds, tea.EnterAltScreen)

	// Start fetching data based on current view
	switch a.currentView.viewType {
	case ViewAttention:
		// Fetch attention dashboard data (new default)
		cmds = append(cmds, a.fetchAttention())
	case ViewPods:
		// For offline mode, automatically fetch pods
		cmds = append(cmds, a.fetchPods("demo-project", "default"))
	default:
		// For other views, try clusters first
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

		// ATTENTION DASHBOARD NAVIGATION - Handle before general navigation
		if a.currentView.viewType == ViewAttention && len(a.attentionItems) > 0 {
			// Initialize subCursor if not set
			if a.subCursor == 0 && a.expandedItems == nil {
				a.subCursor = -1 // -1 means not in sub-navigation
			}

			switch msg.String() {
			case "j", "down":
				// Check if we're in sub-navigation mode
				if a.subCursor >= 0 {
					// Navigate within pod list
					item := a.attentionItems[a.attentionCursor]
					if a.subCursor < len(item.AffectedPods)-1 {
						a.subCursor++
					}
					return a, nil
				}

				// Check if current item is expanded and has pods - enter sub-nav
				currentItem := a.attentionItems[a.attentionCursor]
				if a.expandedItems != nil && a.expandedItems[a.attentionCursor] && len(currentItem.AffectedPods) > 0 {
					// Enter pod list
					a.subCursor = 0
					return a, nil
				}

				// Normal navigation
				if a.attentionCursor < len(a.attentionItems)-1 {
					a.attentionCursor++
				}
				return a, nil

			case "k", "up":
				// Check if we're in sub-navigation mode
				if a.subCursor >= 0 {
					// Navigate within pod list
					if a.subCursor > 0 {
						a.subCursor--
					}
					return a, nil
				}

				// Normal navigation
				if a.attentionCursor > 0 {
					a.attentionCursor--
				}
				return a, nil
			case "1", "2", "3", "4", "5", "6", "7", "8", "9":
				// Jump to line by number (1-indexed display, 0-indexed storage)
				idx := int(msg.String()[0] - '1')
				if idx < len(a.attentionItems) {
					a.attentionCursor = idx
				}
				return a, nil
			case "enter":
				// Check if we're in sub-navigation - navigate to selected pod's logs
				if a.subCursor >= 0 && a.attentionCursor < len(a.attentionItems) {
					item := a.attentionItems[a.attentionCursor]
					if len(item.AffectedPods) > 0 && a.subCursor < len(item.AffectedPods) {
						podName := item.AffectedPods[a.subCursor]

						// Push current view to stack
						a.viewStack = append(a.viewStack, a.currentView)

						// Navigate to logs for selected pod
						a.currentView = ViewContext{
							viewType:      ViewLogs,
							clusterID:     "", // TODO: get from bundle
							clusterName:   "",
							projectID:     "",
							projectName:   "",
							namespaceID:   "",
							namespaceName: item.Namespace,
							podName:       podName,
							containerName: "",
						}

						a.filterLevel = "" // Show all logs by default
						a.loading = true
						return a, a.fetchLogs("", item.Namespace, podName, a.currentContainer, a.showPrevious)
					}
				}

				// Navigate to logs for the selected item (pod issues only)
				if a.attentionCursor < len(a.attentionItems) {
					item := a.attentionItems[a.attentionCursor]
					if item.ResourceType == "pod" && item.PodName != "" {
						// Push current view to stack
						a.viewStack = append(a.viewStack, a.currentView)

						// Navigate to logs view - show all logs by default
						a.currentView = ViewContext{
							viewType:      ViewLogs,
							clusterID:     item.ClusterID,
							clusterName:   "",
							projectID:     "",
							projectName:   "",
							namespaceID:   "",
							namespaceName: item.Namespace,
							podName:       item.PodName,
							containerName: item.ContainerName,
						}

						// Show all logs by default (user can filter with Ctrl+E/W)
						a.filterLevel = ""
						a.loading = true
						return a, a.fetchLogs(item.ClusterID, item.Namespace, item.PodName, a.currentContainer, a.showPrevious)
					}
				}
				return a, nil
			case "left", "h":
				// Exit sub-navigation if in it, otherwise collapse
				if a.subCursor >= 0 {
					a.subCursor = -1
					return a, nil
				}
				// Collapse current item
				if a.attentionCursor < len(a.attentionItems) && a.expandedItems != nil {
					a.expandedItems[a.attentionCursor] = false
				}
				return a, nil
			case "right", "l":
				// Toggle expansion for event items (future feature)
				if a.attentionCursor < len(a.attentionItems) {
					if a.expandedItems == nil {
						a.expandedItems = make(map[int]bool)
					}
					a.expandedItems[a.attentionCursor] = !a.expandedItems[a.attentionCursor]
				}
				return a, nil
			case "m":
				// Toggle expansion of dashboard (show all vs capped)
				a.attentionExpanded = !a.attentionExpanded
				// Reset cursor to safe position if needed
				displayedItems := a.getDisplayedItems()
				if a.attentionCursor >= len(displayedItems) {
					a.attentionCursor = len(displayedItems) - 1
					if a.attentionCursor < 0 {
						a.attentionCursor = 0
					}
				}
				a.ensureCursorVisible() // FIX: Scroll to keep cursor visible (v0.5.2)
				return a, nil
			case "g":
				// Jump to first item (vim muscle memory)
				a.attentionCursor = 0
				a.subCursor = -1
				a.ensureCursorVisible() // FIX: Scroll to top (v0.5.2)
				return a, nil
			case "G":
				// Jump to last item (vim muscle memory)
				displayedItems := a.getDisplayedItems()
				if len(displayedItems) > 0 {
					a.attentionCursor = len(displayedItems) - 1
				}
				a.subCursor = -1
				a.ensureCursorVisible() // FIX: Scroll to bottom (v0.5.2)
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
			// For other views, keep describe functionality
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
		case "c":
			// Navigate from Attention Dashboard to Clusters
			if a.currentView.viewType == ViewAttention {
				// Push current view to stack (though this is root, allows back navigation)
				a.viewStack = append(a.viewStack, a.currentView)

				// Navigate to Clusters
				a.currentView = ViewContext{viewType: ViewClusters}
				a.loading = true
				return a, a.fetchClusters()
			}
			// Cycle through containers in logs view
			if a.currentView.viewType == ViewLogs && len(a.containers) > 1 {
				return a, a.cycleContainer()
			}
		case "i":
			// Toggle CRD description caption in CRD view
			if a.currentView.viewType == ViewCRDs {
				a.showCRDDescription = !a.showCRDDescription
				return a, nil
			}
		case "l":
			// In logs view: toggle between diagnostic panel and raw logs
			if a.currentView.viewType == ViewLogs {
				a.showRawLogs = !a.showRawLogs
				// If switching to raw logs and logs exist, re-render viewport
				if a.showRawLogs && len(a.logs) > 0 {
					a.logViewport.SetContent(a.renderLogsWithColors())
				}
				return a, nil
			}
			// In pods view: Open logs view
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
				return a, a.fetchLogs(a.currentView.clusterID, a.currentView.namespaceName, a.currentView.podName, a.currentContainer, a.showPrevious)
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
		case "s":
			// Cycle sort mode
			if a.currentView.viewType == ViewAttention {
				// Dashboard: 3-mode cycle (Count → Severity → Name)
				return a, a.cycleSortMode()
			} else if a.currentView.viewType == ViewPods {
				// Classic Pod view: 2-mode toggle (Count ↔ Name)
				return a, a.togglePodSortMode()
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
		a.updateNamespaceCounts(msg.namespaces)
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
				a.fetchLogs(a.currentView.clusterID, a.currentView.namespaceName, a.currentView.podName, a.currentContainer, a.showPrevious),
				a.tickTail(), // Schedule next tick
			)
		}

	case attentionMsg:
		a.loading = false
		a.attentionItems = msg.items
		a.error = ""

		// FIX: Restore cursor position by Title after sort (v0.5.2 "Show, Don't Ask")
		// This prevents cursor jumping when user cycles sort modes
		if a.savedRowName != "" && a.currentView.viewType == ViewAttention {
			// Find the item with matching Title in newly sorted list
			for i, item := range a.attentionItems {
				if item.Title == a.savedRowName {
					a.attentionCursor = i
					a.ensureCursorVisible() // Scroll to keep cursor visible
					break
				}
			}
			a.savedRowName = "" // Clear after restoration
		}

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

	// Update viewport if in attention dashboard view (for scrolling)
	if a.currentView.viewType == ViewAttention && a.attentionViewport.Width > 0 {
		newViewport, cmd := a.attentionViewport.Update(msg)
		a.attentionViewport = newViewport
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return a, tea.Batch(cmds...)
}

// View renders the application - simplified for now
func (a *App) View() string {
	if a.error != "" {
		return errorStyle.Render(fmt.Sprintf("Error: %s\n\n[Esc/b]=back  [q]=quit", a.error))
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
		return renderHelp(a.currentView.viewType)
	}

	if a.showingDescribe {
		return a.renderDescribeView()
	}

	// Special rendering for attention dashboard
	if a.currentView.viewType == ViewAttention {
		return a.renderAttentionDashboard()
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
	err error
}

func (e errMsg) Error() string {
	return e.err.Error()
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
