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

// formatCount formats large numbers with K/M/B abbreviation for display
// Examples: 999 → "999", 1500 → "1.5K", 2500000 → "2.5M"
func formatCount(count int) string {
	if count < 1000 {
		return fmt.Sprintf("%d", count)
	}
	if count < 1000000 {
		// Format as K (thousands)
		k := float64(count) / 1000.0
		if k < 10 {
			return fmt.Sprintf("%.1fK", k)
		}
		return fmt.Sprintf("%dK", int(k))
	}
	if count < 1000000000 {
		// Format as M (millions)
		m := float64(count) / 1000000.0
		if m < 10 {
			return fmt.Sprintf("%.1fM", m)
		}
		return fmt.Sprintf("%dM", int(m))
	}
	// Format as B (billions)
	b := float64(count) / 1000000000.0
	if b < 10 {
		return fmt.Sprintf("%.1fB", b)
	}
	return fmt.Sprintf("%dB", int(b))
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
						return a, a.fetchLogs("", item.Namespace, podName)
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
						return a, a.fetchLogs(item.ClusterID, item.Namespace, item.PodName)
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
				return a, nil
			case "g":
				// Jump to first item (vim muscle memory)
				a.attentionCursor = 0
				a.subCursor = -1
				return a, nil
			case "G":
				// Jump to last item (vim muscle memory)
				displayedItems := a.getDisplayedItems()
				if len(displayedItems) > 0 {
					a.attentionCursor = len(displayedItems) - 1
				}
				a.subCursor = -1
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
		return errorStyle.Render(fmt.Sprintf("Error: %s\n\nPress Esc to continue", a.error))
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

	// Build log context header with pod/container details and stats
	visibleLogs := a.getVisibleLogs()
	errorCount := a.countLogLevel(visibleLogs, "ERROR")
	warnCount := a.countLogLevel(visibleLogs, "WARN")

	containerInfo := ""
	if a.currentContainer != "" {
		containerInfo = fmt.Sprintf(" → container: %s", a.currentContainer)
	} else if len(a.containers) > 0 {
		containerInfo = fmt.Sprintf(" → container: %s", a.containers[0])
	}

	contextHeader := fmt.Sprintf("Pod: %s%s (%d lines · %d errors · %d warnings)",
		a.currentView.podName, containerInfo, len(visibleLogs), errorCount, warnCount)
	contextHeaderStyled := lipgloss.NewStyle().
		Foreground(colorCyan).
		Bold(true).
		Render(contextHeader)

	// Build status text with search info
	var statusText string
	if a.searchMode {
		statusText = fmt.Sprintf(" Search: %s_ | Press 'Enter' to search, 'Esc' to cancel ", a.searchQuery)
	} else if len(a.searchMatches) > 0 {
		statusText = fmt.Sprintf(" %d lines | Match %d/%d | 'n'=next 'N'=prev '/'=new Esc=clear | q=quit ",
			len(visibleLogs), a.currentMatch+1, len(a.searchMatches))
	} else {
		statusText = " [/] search  [Ctrl+E] errors only  [Ctrl+W] warnings  [Esc] back  [q] quit "
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
		contextHeaderStyled,
		"",
		logsBox,
		"",
		status,
	)
}

// countLogLevel counts logs matching a specific level
func (a *App) countLogLevel(logs []string, level string) int {
	count := 0
	for _, line := range logs {
		switch level {
		case "ERROR":
			if isErrorLog(line) {
				count++
			}
		case "WARN":
			if isWarnLog(line) {
				count++
			}
		}
	}
	return count
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
			// Compute namespace health (scan all pods for E/W counts)
			scanDepth := a.config.ScanDepth
			if scanDepth <= 0 {
				scanDepth = 200
			}
			nsHealth := ComputeNamespaceHealth(a.dataSource, scanDepth)

			// Sort namespaces by total issue count (worst first)
			sortedNS := make([]rancher.Namespace, len(a.namespaces))
			copy(sortedNS, a.namespaces)

			// Bubble sort by total issues descending
			for i := 0; i < len(sortedNS); i++ {
				for j := i + 1; j < len(sortedNS); j++ {
					health1 := nsHealth[sortedNS[i].Name]
					health2 := nsHealth[sortedNS[j].Name]

					// Sort descending (highest count first)
					if health1.Total < health2.Total {
						sortedNS[i], sortedNS[j] = sortedNS[j], sortedNS[i]
					}
				}
			}

			columns := []table.Column{
				table.NewColumn("name", "NAME", 32),
				table.NewColumn("issues", "ISSUES", 18),
				table.NewColumn("state", "STATE", 10),
				table.NewColumn("project", "PROJECT", 16),
				table.NewColumn("created", "AGE", 8),
			}

			rows := []table.Row{}
			for _, ns := range sortedNS {
				created := "N/A"
				if !ns.Created.IsZero() {
					days := int(time.Since(ns.Created).Hours() / 24)
					if days > 0 {
						created = fmt.Sprintf("%dd", days)
					} else {
						hours := int(time.Since(ns.Created).Hours())
						if hours > 0 {
							created = fmt.Sprintf("%dh", hours)
						} else {
							created = fmt.Sprintf("%dm", int(time.Since(ns.Created).Minutes()))
						}
					}
				}

				// Format ISSUES column with color coding
				health := nsHealth[ns.Name]
				issuesDisplay := "✅ Clean"

				if health.Total > 0 {
					// Color coding logic:
					// Red (🔥): >50 errors
					// Yellow (⚠️): >20 warnings OR 1-50 errors
					// Green: Minor issues
					emoji := "✅"
					if health.Errors > 50 {
						emoji = "🔥"
					} else if health.Warnings > 20 || health.Errors > 0 {
						emoji = "⚠️"
					}

					// Format with K/M abbreviation for large numbers
					errStr := formatCount(health.Errors)
					warnStr := formatCount(health.Warnings)
					issuesDisplay = fmt.Sprintf("%s %sE/%sW", emoji, errStr, warnStr)
				}

				rows = append(rows, table.NewRow(table.RowData{
					"name":    ns.Name,
					"issues":  issuesDisplay,
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
			// Pre-populate cache with E/W counts for sorting
			a.populatePodCounts()

			// Apply sorting based on current sort mode
			sortMode, exists := a.sortModes[ViewPods]
			if !exists {
				sortMode = a.sortMode // Use global default
			}

			// Sort pods according to mode
			sortedPods := a.pods
			switch sortMode {
			case SortByCount:
				sortedPods = SortPodsByCount(a.pods, a.cachedPodCounts)
			case SortBySeverity:
				sortedPods = SortPodsBySeverity(a.pods)
			case SortByName:
				sortedPods = SortPodsByName(a.pods)
			}

			columns := []table.Column{
				table.NewColumn("name", "NAME", 28),
				table.NewColumn("namespace", "NAMESPACE", 18),
				table.NewColumn("state", "STATE", 12),
				table.NewColumn("we", "W/E", 8),
				table.NewColumn("node", "NODE", 20),
			}

			rows := []table.Row{}
			for _, pod := range sortedPods {
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

				// Get warning/error counts by scanning pod logs (same as dashboard)
				weCount := "-"
				if a.dataSource != nil {
					// Try to fetch logs for this pod
					logs, err := a.dataSource.GetLogs("", namespaceName, pod.Name, "", false)
					if err == nil && len(logs) > 0 {
						// Get scan depth from config (tunable via --scan flag, default 200)
						scanDepth := a.config.ScanDepth
						if scanDepth <= 0 {
							scanDepth = 200
						}

						// Limit scan to first N lines for table performance
						scanLines := logs
						if len(scanLines) > scanDepth {
							scanLines = scanLines[:scanDepth]
						}

						warnCount := 0
						errorCount := 0
						for _, line := range scanLines {
							if isErrorLog(line) {
								errorCount++
							} else if isWarnLog(line) {
								warnCount++
							}
						}

						// Only show if there are actual errors/warnings - format: "XE/YW"
						if warnCount > 0 || errorCount > 0 {
							weCount = fmt.Sprintf("%dE/%dW", errorCount, warnCount)
						}
					}
				}

				rows = append(rows, table.NewRow(table.RowData{
					"name":      pod.Name,
					"namespace": namespaceName,
					"state":     pod.State,
					"we":        weCount,
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
		// Add sort mode indicator
		sortMode, exists := a.sortModes[ViewPods]
		if !exists {
			sortMode = a.sortMode
		}
		status = fmt.Sprintf(" %s%d pods | Sort: %s | 's'=sort 'l'=logs 'd'=describe '1/2/3'=switch | '?'=help 'q'=quit ", offlinePrefix, count, sortMode.String())

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
	case ViewAttention:
		return a.fetchAttention()
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
	case ViewAttention:
		// Jump to pod logs from attention dashboard (for pod issues)
		if len(a.attentionItems) == 0 {
			return nil
		}

		// Find matching attention item by pod name from selected row
		// The table rows are built in the same order as attentionItems
		podName := safeRowString(selected, "title")
		if podName == "" {
			return nil
		}

		// Find the attention item with matching title (pod name)
		var matchedItem *AttentionItem
		for i := range a.attentionItems {
			if a.attentionItems[i].Title == podName {
				matchedItem = &a.attentionItems[i]
				break
			}
		}

		if matchedItem == nil {
			return nil
		}

		// Only navigate for pod-related issues
		if matchedItem.ResourceType == "pod" && matchedItem.PodName != "" {
			// Push current view to stack
			a.viewStack = append(a.viewStack, a.currentView)

			// Navigate to logs view with error/warning filter
			a.currentView = ViewContext{
				viewType:      ViewLogs,
				clusterID:     matchedItem.ClusterID,
				clusterName:   "",
				projectID:     "",
				projectName:   "",
				namespaceID:   "",
				namespaceName: matchedItem.Namespace,
				podName:       matchedItem.PodName,
				containerName: matchedItem.ContainerName,
			}

			// Set filter to show errors and warnings by default
			a.filterLevel = "WARN"
			a.loading = true
			return a.fetchLogs(matchedItem.ClusterID, matchedItem.Namespace, matchedItem.PodName)
		}

		return nil

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

	case ViewPods:
		// Navigate to logs for selected pod (UX consistency: Enter = logs)
		podName := safeRowString(selected, "name")
		namespaceName := safeRowString(selected, "namespace")
		if podName == "" || namespaceName == "" {
			return nil // Skip if required fields are missing
		}

		// Push current view to stack
		a.viewStack = append(a.viewStack, a.currentView)

		// Navigate to logs view with WARN filter (shows errors + warnings by default)
		a.currentView = ViewContext{
			viewType:      ViewLogs,
			clusterID:     a.currentView.clusterID,
			clusterName:   a.currentView.clusterName,
			projectID:     a.currentView.projectID,
			projectName:   a.currentView.projectName,
			namespaceID:   a.currentView.namespaceID,
			namespaceName: namespaceName,
			podName:       podName,
			containerName: "",
		}

		// Auto-apply WARN filter to show errors and warnings
		a.filterLevel = "WARN"
		a.loading = true
		return a.fetchLogs(a.currentView.clusterID, namespaceName, podName)

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
}

// fetchAttention analyzes cluster health and returns attention items
func (a *App) fetchAttention() tea.Cmd {
	return func() tea.Msg {
		if a.dataSource == nil {
			return errMsg{fmt.Errorf("no data source available")}
		}

		// Get scan depth from config (default 200 if not set)
		scanDepth := a.config.ScanDepth
		if scanDepth <= 0 {
			scanDepth = 200
		}

		// Detect all issues across the cluster
		items := ComputeAttentionItems(a.dataSource, scanDepth)

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

// cycleSortMode cycles through sort modes: Count → Severity → Name → Count (Dashboard)
func (a *App) cycleSortMode() tea.Cmd {
	// Get current sort mode for this view (default to global if not set)
	currentMode, exists := a.sortModes[a.currentView.viewType]
	if !exists {
		currentMode = a.sortMode // Use global default
	}

	// Cycle to next mode
	nextMode := (currentMode + 1) % 3

	// Store per-view preference
	a.sortModes[a.currentView.viewType] = nextMode

	// Trigger refresh to re-sort
	a.loading = true
	return a.refreshCurrentView()
}

// togglePodSortMode toggles between Count ↔ Name (Pod view only)
func (a *App) togglePodSortMode() tea.Cmd {
	// Get current sort mode for pod view
	currentMode, exists := a.sortModes[ViewPods]
	if !exists {
		currentMode = a.sortMode // Use global default
	}

	// Toggle: Count ↔ Name (skip Severity for pod view)
	var nextMode SortMode
	if currentMode == SortByCount {
		nextMode = SortByName
	} else {
		nextMode = SortByCount
	}

	// Store per-view preference
	a.sortModes[ViewPods] = nextMode

	// Trigger refresh to re-sort
	a.loading = true
	return a.refreshCurrentView()
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

// isErrorLog detects ERROR level logs with explicit indicator priority
// Priority: [ERROR] or E#### > keyword patterns
// This prevents "W1204 [WARN] failed..." or "I1127 [INFO] Failed..." from being detected as ERROR
func isErrorLog(line string) bool {
	lineUpper := strings.ToUpper(line)

	// PRIORITY 1: Check for explicit non-ERROR indicators first (to exclude these lines)
	// This prevents false positives from keyword matching

	// Exclude WARN logs
	if strings.Contains(lineUpper, "[WARN]") || strings.Contains(lineUpper, "[WARNING]") {
		return false
	}

	// K8s WARN format at line start: W####
	if len(line) >= 5 && line[0] == 'W' && isDigit(line[1]) && isDigit(line[2]) &&
		isDigit(line[3]) && isDigit(line[4]) {
		return false
	}

	// Exclude INFO logs
	if strings.Contains(lineUpper, "[INFO]") {
		return false
	}

	// K8s INFO format at line start: I####
	if len(line) >= 5 && line[0] == 'I' && isDigit(line[1]) && isDigit(line[2]) &&
		isDigit(line[3]) && isDigit(line[4]) {
		return false
	}

	// Exclude DEBUG logs
	if strings.Contains(lineUpper, "[DEBUG]") {
		return false
	}

	// K8s DEBUG format at line start: D####
	if len(line) >= 5 && line[0] == 'D' && isDigit(line[1]) && isDigit(line[2]) &&
		isDigit(line[3]) && isDigit(line[4]) {
		return false
	}

	// PRIORITY 2: Explicit ERROR indicators
	// Bracketed format: [ERROR]
	if strings.Contains(lineUpper, "[ERROR]") {
		return true
	}

	// K8s format at line start: E####
	if len(line) >= 5 && line[0] == 'E' && isDigit(line[1]) && isDigit(line[2]) &&
		isDigit(line[3]) && isDigit(line[4]) {
		return true
	}

	// level=error format
	if strings.Contains(lineUpper, "LEVEL=ERROR") {
		return true
	}

	// PRIORITY 3: Keyword patterns (only if no explicit level indicator present)
	errorPatterns := []string{
		"ERROR:",
		"ERR=",
		"FAILED",
		"FATAL",
		"PANIC",
		"OOMKILLED",
		"CRASHLOOP",
		"BACK-OFF",
		"BACKOFF",
		"UNAUTHORIZED",
		"DENIED",
		"EXCEPTION",
	}

	for _, pattern := range errorPatterns {
		if strings.Contains(lineUpper, pattern) {
			return true
		}
	}

	return false
}

// isWarnLog detects WARN level logs with explicit indicator priority
func isWarnLog(line string) bool {
	lineUpper := strings.ToUpper(line)

	// PRIORITY 1: Explicit WARN indicators
	// Bracketed formats
	if strings.Contains(lineUpper, "[WARN]") || strings.Contains(lineUpper, "[WARNING]") {
		return true
	}

	// K8s format at line start: W#### (check first 5 chars only)
	if len(line) >= 5 && line[0] == 'W' && isDigit(line[1]) && isDigit(line[2]) &&
		isDigit(line[3]) && isDigit(line[4]) {
		return true
	}

	// Colon-based formats
	if strings.Contains(lineUpper, "WARNING:") || strings.Contains(lineUpper, "WARN:") {
		return true
	}

	// level=warn format
	if strings.Contains(lineUpper, "LEVEL=WARN") || strings.Contains(lineUpper, "LEVEL=WARNING") {
		return true
	}

	// PRIORITY 2: Keyword patterns (only if no explicit ERROR indicator)
	if strings.Contains(lineUpper, "[ERROR]") {
		return false
	}
	if len(line) >= 5 && line[0] == 'E' && isDigit(line[1]) && isDigit(line[2]) &&
		isDigit(line[3]) && isDigit(line[4]) {
		return false
	}

	warnKeywords := []string{
		"WARN=",
		"DEPRECATED",
		"DEPRECATION",
		"ALERT:",
		"ALERT=",
	}

	for _, pattern := range warnKeywords {
		if strings.Contains(lineUpper, pattern) {
			return true
		}
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
// FIXED: Colors are now applied AFTER wrapping to prevent ANSI escape code splits
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

	// Word wrap enabled - wrap FIRST, then colorize EACH segment
	// This prevents ANSI escape codes from being split across wrapped lines
	var wrappedLines []string
	wrapWidth := a.logViewport.Width
	if wrapWidth <= 0 {
		wrapWidth = 80 // Fallback width
	}

	for i, line := range visibleLogs {
		// Check if this is the current search match
		isCurrentMatch := false
		if len(a.searchMatches) > 0 && a.currentMatch >= 0 && a.currentMatch < len(a.searchMatches) {
			if i == a.searchMatches[a.currentMatch] {
				isCurrentMatch = true
			}
		}

		if len(line) <= wrapWidth {
			// No wrap needed - colorize entire line
			wrappedLines = append(wrappedLines, a.colorizeLogLine(line, i))
		} else {
			// Wrap raw text into segments FIRST
			remainingLine := line
			for len(remainingLine) > 0 {
				// Determine segment length
				segmentEnd := wrapWidth
				if segmentEnd > len(remainingLine) {
					segmentEnd = len(remainingLine)
				}

				segment := remainingLine[:segmentEnd]

				// Apply color styling to EACH wrapped segment
				// This preserves colors across all wrapped lines
				if isCurrentMatch {
					wrappedLines = append(wrappedLines, searchMatchStyle.Render(segment))
				} else if isErrorLog(line) {
					wrappedLines = append(wrappedLines, logErrorStyle.Render(segment))
				} else if isWarnLog(line) {
					wrappedLines = append(wrappedLines, logWarnStyle.Render(segment))
				} else if isInfoLog(line) {
					wrappedLines = append(wrappedLines, logInfoStyle.Render(segment))
				} else if isDebugLog(line) {
					wrappedLines = append(wrappedLines, logDebugStyle.Render(segment))
				} else {
					wrappedLines = append(wrappedLines, segment)
				}

				remainingLine = remainingLine[segmentEnd:]
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
