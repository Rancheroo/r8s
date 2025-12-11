package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// Default dashboard cap - show top N items before requiring expansion
const defaultDashboardCap = 20

// getDisplayedItems returns items to display based on expansion state and cap
// Items are sorted according to the current sort mode before display
// CRITICAL GUARANTEE: ALL critical severity items are ALWAYS included, even if beyond cap
func (a *App) getDisplayedItems() []AttentionItem {
	// Get current sort mode for this view
	sortMode, exists := a.sortModes[ViewAttention]
	if !exists {
		sortMode = a.sortMode // Use global default
	}

	// Apply sorting based on current mode (returns sorted copy)
	items := GetSortedAttentionItems(a.attentionItems, sortMode)

	// If expanded or total items <= cap, show all
	if a.attentionExpanded || len(items) <= defaultDashboardCap {
		return items
	}

	// CRITICAL-SAFE CAPPING: Ensure ALL criticals are included
	// Dynamic cap expansion if needed to show all critical severity items
	cap := defaultDashboardCap

	// Count criticals in the full sorted list
	criticalCount := 0
	lastCriticalPosition := -1
	for i, item := range items {
		if item.Severity == SeverityCritical {
			criticalCount++
			lastCriticalPosition = i
		}
	}

	// If any critical is beyond the cap, expand the cap to include it
	if lastCriticalPosition >= cap {
		cap = lastCriticalPosition + 1
	}

	return items[:cap]
}

// ensureCursorVisible scrolls viewport to keep cursor visible
func (a *App) ensureCursorVisible() {
	if a.currentView.viewType != ViewAttention {
		return
	}

	// Calculate line number of cursor position
	// For now, simple approach: scroll to cursor line
	// Each item is ~1-2 lines depending on expansion
	lineNum := a.attentionCursor * 2 // Approximate

	// Scroll viewport to show this line
	viewportHeight := a.attentionViewport.Height
	if lineNum < a.attentionViewport.YOffset {
		// Cursor above viewport - scroll up
		a.attentionViewport.SetYOffset(lineNum)
	} else if lineNum >= a.attentionViewport.YOffset+viewportHeight {
		// Cursor below viewport - scroll down
		a.attentionViewport.SetYOffset(lineNum - viewportHeight + 1)
	}
}

// renderAttentionDashboard renders the attention dashboard view with scrolling
func (a *App) renderAttentionDashboard() string {
	if len(a.attentionItems) == 0 {
		return a.renderAllGood()
	}

	// Get displayed items (respects capping/expansion)
	displayedItems := a.getDisplayedItems()

	// Group items by severity
	critical := []AttentionItem{}
	warning := []AttentionItem{}
	info := []AttentionItem{}

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

	// Build header
	mode := ""
	if a.bundleMode {
		mode = "[BUNDLE] "
	} else if a.offlineMode {
		mode = "[MOCK] "
	} else {
		mode = "[LIVE] "
	}

	clusterName := "cluster"
	if a.bundleMode && a.dataSource != nil {
		clusters, err := a.dataSource.GetClusters()
		if err == nil && len(clusters) > 0 {
			clusterName = clusters[0].Name
		}
	}

	totalIssues := len(a.attentionItems)
	displayedCount := len(displayedItems)
	criticalCount := len(critical)
	warningCount := len(warning)

	headerText := fmt.Sprintf("🚨 ATTENTION DASHBOARD       %s%s", mode, clusterName)
	summaryText := fmt.Sprintf("%d issues (%d critical, %d warning)", totalIssues, criticalCount, warningCount)

	header := lipgloss.NewStyle().
		Foreground(colorWhite).
		Background(colorRed).
		Bold(true).
		Padding(0, 1).
		Width(a.width - 4).
		Render(headerText)

	summary := lipgloss.NewStyle().
		Foreground(colorYellow).
		Bold(true).
		Padding(0, 1).
		Render(summaryText)

	// Build issue list with cursor tracking
	var lines []string
	itemIdx := 0 // Track actual item index (for cursor)

	if len(critical) > 0 {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Foreground(colorRed).Bold(true).Render("CRITICAL:"))
		for _, item := range critical {
			isSelected := (itemIdx == a.attentionCursor)
			line := a.renderAttentionItem(itemIdx+1, item, isSelected)
			lines = append(lines, line)

			// Show expanded content if this item is expanded
			if a.expandedItems != nil && a.expandedItems[itemIdx] && len(item.AffectedPods) > 0 {
				// Pass current position to know if we should highlight pods
				inSubNav := (itemIdx == a.attentionCursor && a.subCursor >= 0)
				lines = append(lines, a.renderExpandedContent(item, inSubNav)...)
			}
			itemIdx++
		}
	}

	if len(warning) > 0 {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Foreground(colorYellow).Bold(true).Render("WARNING:"))
		for _, item := range warning {
			isSelected := (itemIdx == a.attentionCursor)
			line := a.renderAttentionItem(itemIdx+1, item, isSelected)
			lines = append(lines, line)

			// Show expanded content if this item is expanded
			if a.expandedItems != nil && a.expandedItems[itemIdx] && len(item.AffectedPods) > 0 {
				inSubNav := (itemIdx == a.attentionCursor && a.subCursor >= 0)
				lines = append(lines, a.renderExpandedContent(item, inSubNav)...)
			}
			itemIdx++
		}
	}

	if len(info) > 0 {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Foreground(colorCyan).Bold(true).Render("INFO:"))
		for _, item := range info {
			isSelected := (itemIdx == a.attentionCursor)
			line := a.renderAttentionItem(itemIdx+1, item, isSelected)
			lines = append(lines, line)

			// Show expanded content if this item is expanded
			if a.expandedItems != nil && a.expandedItems[itemIdx] && len(item.AffectedPods) > 0 {
				inSubNav := (itemIdx == a.attentionCursor && a.subCursor >= 0)
				lines = append(lines, a.renderExpandedContent(item, inSubNav)...)
			}
			itemIdx++
		}
	}

	// Add capping indicator if items are hidden
	if displayedCount < totalIssues {
		hiddenCount := totalIssues - displayedCount
		cappingMsg := fmt.Sprintf("\n...and %d more issues (press 'm' to show all)", hiddenCount)
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Foreground(colorGray).Italic(true).Render(cappingMsg))
	}

	content := strings.Join(lines, "\n")

	// Use viewport for scrollable content when expanded
	contentHeight := a.height - 10 // Reserve space for header, summary, status

	// Initialize or update viewport
	if a.attentionViewport.Width == 0 {
		a.attentionViewport = viewport.New(a.width-8, contentHeight)
		a.attentionViewport.SetContent(content)
	} else {
		a.attentionViewport.Width = a.width - 8
		a.attentionViewport.Height = contentHeight
		a.attentionViewport.SetContent(content)
	}

	// Get scrollable viewport view
	viewportContent := a.attentionViewport.View()

	// Create bordered box around viewport
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorRed).
		Padding(1, 2).
		Width(a.width - 4).
		Render(viewportContent)

	// Build status with critical count visibility, position indicator, and sort mode
	var statusParts []string

	// Count total criticals in full list (not just displayed)
	totalCriticals := 0
	displayedCriticals := 0
	for _, item := range a.attentionItems {
		if item.Severity == SeverityCritical {
			totalCriticals++
		}
	}
	for _, item := range displayedItems {
		if item.Severity == SeverityCritical {
			displayedCriticals++
		}
	}

	// CRITICAL COUNT FIRST (highest visibility)
	if totalCriticals > 0 {
		if displayedCriticals < totalCriticals {
			statusParts = append(statusParts, fmt.Sprintf("🔥 Criticals: %d/%d shown", displayedCriticals, totalCriticals))
		} else {
			statusParts = append(statusParts, fmt.Sprintf("🔥 Criticals: %d", totalCriticals))
		}
	}

	// Item count indicator
	if displayedCount < totalIssues {
		statusParts = append(statusParts, fmt.Sprintf("Showing %d/%d", displayedCount, totalIssues))
	} else {
		statusParts = append(statusParts, fmt.Sprintf("%d items", displayedCount))
	}

	// Add sort mode indicator
	sortMode, exists := a.sortModes[ViewAttention]
	if !exists {
		sortMode = a.sortMode
	}
	statusParts = append(statusParts, fmt.Sprintf("Sort: %s", sortMode.String()))

	statusParts = append(statusParts, "[s]=sort")
	statusParts = append(statusParts, "[m]=expand")
	statusParts = append(statusParts, "[g/G]=top/bottom")
	statusParts = append(statusParts, "[Enter]=logs")
	statusParts = append(statusParts, "[c]=classic")

	statusText := " " + strings.Join(statusParts, " · ") + " "
	status := statusStyle.Render(statusText)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		summary,
		"",
		box,
		"",
		status,
	)
}

// renderAttentionItem renders a single attention item with number prefix and selection highlight
func (a *App) renderAttentionItem(num int, item AttentionItem, isSelected bool) string {
	// Format: "1. ► 💀 nginx-deploy-xyz    CrashLoopBackOff    kube-system"
	numStr := fmt.Sprintf("%d. ", num)

	// Add ►/▼ indicator for collapsible event items
	expandIndicator := ""
	if item.ResourceType == "event" || item.ResourceType == "cluster" {
		// Check if this item is expanded
		itemIdx := num - 1 // Convert to 0-based index
		if a.expandedItems != nil && a.expandedItems[itemIdx] {
			expandIndicator = "▼ "
		} else {
			expandIndicator = "► "
		}
	}

	titleWidth := 30
	descWidth := 25
	nsWidth := 20

	title := item.Title
	if len(title) > titleWidth {
		title = title[:titleWidth-3] + "..."
	}

	desc := item.Description
	if len(desc) > descWidth {
		desc = desc[:descWidth-3] + "..."
	}

	ns := item.Namespace
	if len(ns) > nsWidth {
		ns = ns[:nsWidth-3] + "..."
	}

	line := fmt.Sprintf("%s%s%s %-*s  %-*s  %s",
		numStr,
		expandIndicator,
		item.Emoji,
		titleWidth, title,
		descWidth, desc,
		ns,
	)

	// Apply selection highlight (inverts colors for visibility)
	if isSelected {
		return lipgloss.NewStyle().
			Background(colorCyan).
			Foreground(colorDarkGray).
			Bold(true).
			Render(line)
	}

	// Color the entire line based on severity when not selected
	var style lipgloss.Style
	switch item.Severity {
	case SeverityCritical:
		style = lipgloss.NewStyle().Foreground(colorRed)
	case SeverityWarning:
		style = lipgloss.NewStyle().Foreground(colorYellow)
	case SeverityInfo:
		style = lipgloss.NewStyle().Foreground(colorWhite)
	}

	return style.Render(line)
}

// renderExpandedContent renders the expanded pod list for event items
func (a *App) renderExpandedContent(item AttentionItem, inSubNav bool) []string {
	var lines []string

	// Show top pods with event counts
	for i, podName := range item.AffectedPods {
		if i >= 5 { // Show max 5 pods to avoid clutter
			remaining := len(item.AffectedPods) - 5
			if remaining > 0 {
				hint := lipgloss.NewStyle().Foreground(colorGray).Render(
					fmt.Sprintf("       ... and %d more pods (press Enter for logs)", remaining))
				lines = append(lines, hint)
			}
			break
		}

		// Get event count for this pod
		eventCount := 0
		if item.AffectedPodCounts != nil {
			eventCount = item.AffectedPodCounts[podName]
		}

		// Format: "       ├─ pod-name-abc123 (123 events)"
		prefix := "       ├─ "
		if i == len(item.AffectedPods)-1 || i == 4 {
			prefix = "       └─ "
		}

		podText := fmt.Sprintf("%s%s (%d events)", prefix, podName, eventCount)

		// Highlight if this pod is selected in sub-navigation
		isSelectedPod := inSubNav && i == a.subCursor
		if isSelectedPod {
			podLine := lipgloss.NewStyle().
				Background(colorCyan).
				Foreground(colorDarkGray).
				Bold(true).
				Render(podText)
			lines = append(lines, podLine)
		} else {
			podLine := lipgloss.NewStyle().Foreground(colorGray).Render(podText)
			lines = append(lines, podLine)
		}
	}

	return lines
}

// renderAllGood renders the "all systems operational" screen
func (a *App) renderAllGood() string {
	mode := ""
	if a.bundleMode {
		mode = "[BUNDLE] "
	} else if a.offlineMode {
		mode = "[MOCK] "
	} else {
		mode = "[LIVE] "
	}

	header := lipgloss.NewStyle().
		Foreground(colorWhite).
		Background(colorGreen).
		Bold(true).
		Padding(0, 1).
		Width(a.width - 4).
		Render(fmt.Sprintf("✨ ATTENTION DASHBOARD       %s", mode))

	message := lipgloss.NewStyle().
		Foreground(colorGreen).
		Bold(true).
		Align(lipgloss.Center).
		Width(a.width - 8).
		Render("All systems operational ✨")

	subtext := lipgloss.NewStyle().
		Foreground(colorWhite).
		Align(lipgloss.Center).
		Width(a.width - 8).
		Render("No issues detected in this cluster")

	hint := lipgloss.NewStyle().
		Foreground(colorGray).
		Align(lipgloss.Center).
		Width(a.width - 8).
		Render("Press [c] or [Enter] to continue to cluster navigation")

	// Calculate padding for vertical centering
	contentHeight := 6 // Lines of actual content
	availableHeight := a.height - 6
	topPadding := (availableHeight - contentHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	var paddingLines []string
	for i := 0; i < topPadding; i++ {
		paddingLines = append(paddingLines, "")
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorGreen).
		Padding(1, 2).
		Width(a.width - 4).
		Height(availableHeight).
		Render(strings.Join(append(paddingLines, "",
			message,
			"",
			subtext,
			"",
			"",
			hint,
		), "\n"))

	status := statusStyle.Render(" [c] classic view · [r] refresh · [q] quit ")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		box,
		"",
		status,
	)
}
