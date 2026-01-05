// Package tui - Event handlers and navigation logic.
// This file contains all keyboard event handlers, navigation functions,
// and user interaction processing.
package tui

import (
	"fmt"
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
				clusterName:   matchedItem.ClusterName,
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
			// Set internal state to match view context before fetching
			a.currentContainer = matchedItem.ContainerName
			return a.fetchLogs(matchedItem.ClusterID, matchedItem.Namespace, matchedItem.PodName, matchedItem.ContainerName, a.showPrevious)
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
		return a.fetchLogs(a.currentView.clusterID, namespaceName, podName, a.currentContainer, a.showPrevious)

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
	return a.fetchLogs(a.currentView.clusterID, namespaceName, podName, a.currentContainer, a.showPrevious)
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
	case ViewCRDInstances:
		return a.fetchCRDInstances(a.currentView.clusterID, a.currentView.crdGroup, a.currentView.crdVersion, a.currentView.crdResource)
	case ViewLogs:
		return a.fetchLogs(a.currentView.clusterID, a.currentView.namespaceName, a.currentView.podName, a.currentContainer, a.showPrevious)
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
