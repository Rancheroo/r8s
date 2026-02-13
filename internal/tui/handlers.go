// Package tui - Event handlers and navigation logic.
// This file contains all keyboard event handlers, navigation functions,
// and user interaction processing.
package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Rancheroo/r8s/internal/rancher"
)

// handleEnter handles navigation when user presses Enter
func (a *App) handleEnter() tea.Cmd {
	if a.table.HighlightedRow().Data == nil {
		return nil
	}

	selected := a.table.HighlightedRow().Data

	switch a.currentView.viewType {
	case ViewAttention:
		// Jump to pod logs from attention dashboard (for pod issues)
		// CRITICAL FIX: ViewAttention groups by severity (critical→warning→info)
		// Must replicate this grouping to match visual order with cursor position
		if len(a.attentionItems) == 0 {
			return nil
		}

		// Get displayed items (respects sorting and capping)
		// IMPORTANT: This returns items already sorted by the current sort mode
		displayedItems := a.getDisplayedItems()

		// Group items by severity while PRESERVING sort order within each group
		// This MUST match the rendering logic in renderAttentionDashboard() exactly
		var critical, warning, info []AttentionItem
		for _, item := range displayedItems {
			switch item.Severity {
			case SeverityCritical:
				critical = append(critical, item)
			case SeverityWarning:
				warning = append(warning, item)
			case SeverityInfo:
				info = append(info, item)
			}
		}

		// Build visual order: critical → warning → info (matching rendered order)
		visualOrder := make([]AttentionItem, 0, len(displayedItems))
		visualOrder = append(visualOrder, critical...)
		visualOrder = append(visualOrder, warning...)
		visualOrder = append(visualOrder, info...)

		// Validate cursor is in bounds
		if a.attentionCursor < 0 || a.attentionCursor >= len(visualOrder) {
			return nil
		}

		matchedItem := &visualOrder[a.attentionCursor]

		// DEBUG: Log item details for diagnosis (verbose mode only)
		if a.config.Verbose {
			fmt.Printf("DEBUG handleEnter: cursor=%d, ResourceType='%s', Title='%s', PodName='%s'\n",
				a.attentionCursor, matchedItem.ResourceType, matchedItem.Title, matchedItem.PodName)
		}

		// v0.6.1: Handle cluster event drill-down
		// v0.6.8: Extended to support node and etcd drill-down
		if len(matchedItem.AffectedPods) > 0 {
			// Drill-down available for: event, node, etcd
			if matchedItem.ResourceType == "event" ||
				matchedItem.ResourceType == "node" ||
				matchedItem.ResourceType == "etcd" {
				return a.handleClusterEventDrillDown(matchedItem)
			}
		}

		// v0.6.7/v0.6.8: Comprehensive non-pod resource type handling
		// Only pod items with valid PodName can navigate to diagnostics
		// Kubelet, system, daemonset, log items without drill-down return nil
		if matchedItem.ResourceType != "pod" || matchedItem.PodName == "" {
			return nil
		}

		// Navigate to pod logs view with diagnostic-first display
		cmd := a.navigateToLogs(matchedItem.ClusterID, matchedItem.Namespace, matchedItem.PodName, matchedItem.ContainerName)

		// Dashboard-specific: Set filter to show errors and warnings by default
		a.filterLevel = "WARN"
		a.currentContainer = matchedItem.ContainerName
		return cmd

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

		// Validate clusterID was found before proceeding
		if clusterID == "" {
			a.error = fmt.Sprintf("cluster '%s' not found", clusterName)
			a.loading = false
			return nil
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

		// Validate projectID was found before proceeding
		if projectID == "" {
			a.error = fmt.Sprintf("project '%s' not found", projectName)
			a.loading = false
			return nil
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

		// FIX (v0.5.4): In bundle mode with derived namespaces, ID may be empty
		// Use namespace name directly for navigation - name is what matters for pod fetching
		var namespaceID string
		for _, n := range a.namespaces {
			if n.Name == namespaceName {
				namespaceID = n.ID // May be empty for derived namespaces
				break
			}
		}

		// Push current view to stack
		a.viewStack = append(a.viewStack, a.currentView)

		// Navigate to Pods (use name, not ID - works for derived namespaces)
		a.currentView = ViewContext{
			viewType:      ViewPods,
			clusterID:     a.currentView.clusterID,
			clusterName:   a.currentView.clusterName,
			projectID:     a.currentView.projectID,
			projectName:   a.currentView.projectName,
			namespaceID:   namespaceID, // May be empty - that's OK
			namespaceName: namespaceName,
		}
		a.loading = true
		return a.fetchPods(a.currentView.projectID, namespaceName)

	case ViewPods:
		// Navigate to logs for selected pod (diagnostic-first approach)
		podName := safeRowString(selected, "name")
		namespaceName := safeRowString(selected, "namespace")
		if podName == "" || namespaceName == "" {
			return nil // Skip if required fields are missing
		}

		// v0.5.8: Use unified navigation function (eliminates duplicate state clearing)
		return a.navigateToLogs(a.currentView.clusterID, namespaceName, podName, "")

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

		// FIX BUG-001: Use helper function to select best CRD version
		// This correctly handles served versions and avoids 404 errors
		storageVersion, err := selectBestCRDVersion(selectedCRD.Spec.Versions)
		if err != nil {
			a.error = fmt.Sprintf("CRD %s: %v", selectedCRD.Metadata.Name, err)
			return nil
		}

		// Push current view to stack only AFTER successful validation
		a.viewStack = append(a.viewStack, a.currentView)

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

	case ViewContainerSelect:
		// S3-MEDIUM-1: Handle container selection
		containerName := safeRowString(selected, "container")
		if containerName == "" {
			return nil
		}

		a.currentContainer = containerName

		// Clear any previous pod/log state to prevent data leakage between containers
		a.clearPodState()

		// Navigate to logs with selected container
		a.currentView = ViewContext{
			viewType:      ViewLogs,
			clusterID:     a.currentView.clusterID,
			clusterName:   a.currentView.clusterName,
			projectID:     a.currentView.projectID,
			projectName:   a.currentView.projectName,
			namespaceID:   a.currentView.namespaceID,
			namespaceName: a.currentView.namespaceName,
			podName:       a.currentView.podName,
			containerName: containerName,
		}

		a.loading = true
		return a.fetchLogs(a.currentView.clusterID, a.currentView.namespaceName, a.currentView.podName, containerName, a.showPrevious)

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
	case ViewAttention:
		// Describe pod from attention dashboard
		if len(a.attentionItems) == 0 {
			return nil
		}

		// Find matching attention item by pod name
		podName := safeRowString(selected, "title")
		if podName == "" {
			return nil
		}

		var matchedItem *AttentionItem
		for i := range a.attentionItems {
			if a.attentionItems[i].Title == podName {
				matchedItem = &a.attentionItems[i]
				break
			}
		}

		if matchedItem == nil || matchedItem.ResourceType != "pod" || matchedItem.PodName == "" {
			a.error = "Describe is not yet implemented for this resource type"
			return nil
		}

		return a.describePod(matchedItem.ClusterID, matchedItem.Namespace, matchedItem.PodName)

	case ViewLogs:
		// Describe pod from log view (including "No Logs" screen)
		if a.currentView.podName == "" || a.currentView.namespaceName == "" {
			a.error = "No pod information available"
			return nil
		}
		return a.describePod(a.currentView.clusterID, a.currentView.namespaceName, a.currentView.podName)

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

// handleViewLogs navigates to logs view for the selected pod
// S3-MEDIUM-1: Supports multi-container pod selection
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

	// S3-MEDIUM-1: Check for multiple containers
	containers, err := a.dataSource.GetContainers(namespaceName, podName)
	if err != nil {
		// Log error and handle gracefully - don't pass empty string to fetchLogs
		fmt.Printf("ERROR: Failed to get containers for pod %s/%s: %v\n", namespaceName, podName, err)
		// Let downstream handle empty container list appropriately
		containers = nil
	}

	a.containers = containers

	// S3-MEDIUM-1: Fetch diagnostic info for containers
	a.containerDetails = a.fetchContainerDiagnostics(namespaceName, podName, containers)

	// Push current view to stack
	a.viewStack = append(a.viewStack, a.currentView)

	// S3-MEDIUM-1: If multiple containers, show selection UI first
	if len(containers) > 1 {
		a.currentView = ViewContext{
			viewType:      ViewContainerSelect,
			clusterID:     a.currentView.clusterID,
			clusterName:   a.currentView.clusterName,
			projectID:     a.currentView.projectID,
			projectName:   a.currentView.projectName,
			namespaceID:   a.currentView.namespaceID,
			namespaceName: namespaceName,
			podName:       podName,
		}
		a.updateContainerSelectTable()
		return nil
	}

	// Single container - go straight to logs
	if len(containers) == 1 {
		a.currentContainer = containers[0]
	} else {
		a.currentContainer = ""
	}

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
		containerName: a.currentContainer,
	}

	a.loading = true
	return a.fetchLogs(a.currentView.clusterID, namespaceName, podName, a.currentContainer, a.showPrevious)
}

// refreshCurrentView handles refreshing the current view data
func (a *App) refreshCurrentView() tea.Cmd {
	switch a.currentView.viewType {
	case ViewAttention:
		return a.fetchAttention()
	case ViewClusterEvent:
		// v0.6.1: Cluster event view doesn't need refresh - data already in attentionItems
		// Just clear loading state immediately
		a.loading = false
		return nil
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
		return a.fetchCRDInstances(a.currentView.clusterID, a.currentView.crdGroup, a.currentView.crdVersion, a.currentView.crdResource)
	case ViewLogs:
		return a.fetchLogs(a.currentView.clusterID, a.currentView.namespaceName, a.currentView.podName, a.currentContainer, a.showPrevious)
	case ViewContainerSelect:
		// S3-MEDIUM-1: Container selection doesn't need refresh - just ensure table is up to date
		a.updateContainerSelectTable()
		a.loading = false
		return nil
	default:
		return nil
	}
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
// FIX: Track cursor by Title to maintain selection after sort (v0.5.2 "Show, Don't Ask")
func (a *App) cycleSortMode() tea.Cmd {
	// Save currently selected item's Title BEFORE sorting (for Dashboard)
	var selectedTitle string
	if a.currentView.viewType == ViewAttention && a.attentionCursor < len(a.attentionItems) {
		selectedTitle = a.attentionItems[a.attentionCursor].Title
	}

	// Get current sort mode for this view (default to global if not set)
	currentMode, exists := a.sortModes[a.currentView.viewType]
	if !exists {
		currentMode = a.sortMode // Use global default
	}

	// Cycle to next mode (use NumSortModes sentinel for safe wrapping)
	nextMode := SortMode((int(currentMode) + 1) % int(NumSortModes))

	// Store per-view preference
	a.sortModes[a.currentView.viewType] = nextMode

	// For Dashboard: restore cursor position after sort
	// Store title for restoration after data refresh
	a.savedRowName = selectedTitle

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
			a.logViewport.ScrollDown(1)
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
		// No containers available - return immediately
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

	// Fetch logs for the new container
	return a.fetchLogs(a.currentView.clusterID, a.currentView.namespaceName, a.currentView.podName, a.currentContainer, a.showPrevious)
}

// isNamespaceResourceView returns true if the current view is a namespace-scoped resource view
func (a *App) isNamespaceResourceView() bool {
	return a.currentView.viewType == ViewPods ||
		a.currentView.viewType == ViewDeployments ||
		a.currentView.viewType == ViewServices
}

// handleClusterEventDrillDown navigates to cluster event drill-down view (v0.6.1)
// Shows affected pods list for cluster-wide events like "190× BackOff"
func (a *App) handleClusterEventDrillDown(item *AttentionItem) tea.Cmd {
	if item == nil || len(item.AffectedPods) == 0 {
		return nil
	}

	// Prefer EventReason field if available, otherwise parse from Title
	var eventReason string
	if item.EventReason != "" {
		eventReason = item.EventReason
	} else {
		// Fallback: Extract event reason from Title (e.g., "190× BackOff" -> "BackOff")
		titleParts := strings.Split(item.Title, "×")
		if len(titleParts) > 1 {
			eventReason = strings.TrimSpace(titleParts[1])
		} else {
			eventReason = strings.TrimSpace(item.Title)
		}
	}

	// Push current view to stack
	a.viewStack = append(a.viewStack, a.currentView)

	// Use actual event type, fallback to Warning if not set
	eventType := item.EventType
	if eventType == "" {
		eventType = "Warning"
	}

	// v0.6.8: Capture node name for node events
	var nodeName string
	if item.ResourceType == "node" {
		// Extract node name from Title (format: "Node w-guard-wg-wk-pfvjr-4x...")
		nodeName = strings.TrimPrefix(item.Title, "Node ")
		// Truncate if it has description after " - "
		if dashIndex := strings.Index(nodeName, " - "); dashIndex > 0 {
			nodeName = nodeName[:dashIndex]
		}
	}

	// Navigate to cluster event view
	a.currentView = ViewContext{
		viewType:    ViewClusterEvent,
		clusterID:   item.ClusterID,
		clusterName: item.ClusterName,
		eventReason: eventReason,
		eventType:   eventType,
		nodeName:    nodeName, // v0.6.8: Include node name
	}

	// No loading needed - we already have the data
	a.loading = false

	// CRITICAL FIX v0.6.8.1: Return a cmd to trigger re-render
	// Returning nil prevents Bubble Tea from updating the view
	return func() tea.Msg {
		return clusterEventMsg{}
	}
}

// clusterEventMsg triggers view update for cluster event drill-down
type clusterEventMsg struct{}

// handleClusterEventPodSelection handles 1-9 key presses in cluster event view (v0.6.8.1)
func (a *App) handleClusterEventPodSelection(keyNum string) tea.Cmd {
	// Convert key to index (1-based to 0-based)
	podIndex, err := strconv.Atoi(keyNum)
	if err != nil || podIndex < 1 || podIndex > 9 {
		return nil
	}
	podIndex-- // Convert to 0-based

	// Find the event item from attentionItems matching current eventReason
	var eventItem *AttentionItem
	for i := range a.attentionItems {
		item := &a.attentionItems[i]
		// Match by eventReason (works for event, node, and etcd types)
		if item.EventReason == a.currentView.eventReason ||
			(item.ResourceType == "node" && strings.Contains(item.Title, a.currentView.eventReason)) ||
			(item.ResourceType == "etcd" && strings.Contains(item.Title, a.currentView.eventReason)) {
			eventItem = item
			break
		}
	}

	if eventItem == nil || podIndex >= len(eventItem.AffectedPods) {
		return nil
	}

	// Get selected pod name
	podName := eventItem.AffectedPods[podIndex]

	// Find pod details to get namespace
	allPods, err := a.dataSource.GetAllPods()
	if err != nil {
		a.error = fmt.Sprintf("Failed to get pod list: %v", err)
		return nil
	}

	var selectedPod *rancher.Pod
	for i := range allPods {
		// v0.6.8.2 FIX: Match by pod name ONLY (namespace in eventItem is unreliable)
		// In bundle mode, pod names are globally unique so this is safe
		if allPods[i].Name == podName {
			selectedPod = &allPods[i]
			break
		}
	}

	if selectedPod == nil {
		a.error = fmt.Sprintf("Pod '%s' not found", podName)
		return nil
	}

	// Extract namespace from the matched pod's NamespaceID
	// Bundle mode: "cattle-system:default" → "default"
	// Mock mode: "default" → "default"
	namespace := selectedPod.NamespaceID
	if strings.Contains(namespace, ":") {
		parts := strings.Split(namespace, ":")
		if len(parts) > 1 {
			namespace = parts[len(parts)-1] // Take last part after final ":"
		}
	}

	// Navigate to pod logs using unified navigation
	return a.navigateToLogs(a.currentView.clusterID, namespace, podName, "")
}
