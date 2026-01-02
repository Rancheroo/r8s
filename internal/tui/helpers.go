// Package tui - Helper utilities and formatting functions for the TUI.
// This file contains reusable utility functions for safe data extraction,
// number formatting, breadcrumb generation, status text, and modal rendering.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"

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
