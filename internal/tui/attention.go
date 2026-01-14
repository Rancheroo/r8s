package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// Default dashboard cap - show top N items before requiring expansion
// Updated from 20 to 100 to match production usage and test expectations
const defaultDashboardCap = 100

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
	infoCount := len(info)

	headerText := fmt.Sprintf("🚨 ATTENTION DASHBOARD       %s%s", mode, clusterName)

	// Auto-display health summary (Show, Don't Ask philosophy)
	// Show health breakdown without requiring any button press
	summaryText := fmt.Sprintf("🔥 %d critical · ⚠️  %d warnings · ℹ️  %d info  (%d total)",
		criticalCount, warningCount, infoCount, totalIssues)

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

	// Calculate dynamic number width for alignment (Issue #18 fix)
	numWidth := len(fmt.Sprintf("%d", displayedCount)) + 2 // +2 for ". "

	// Build issue list with cursor tracking
	var lines []string
	itemIdx := 0 // Track actual item index (for cursor)

	// FIX (v0.5.9): Simplified rendering - removed expansion complexity
	if len(critical) > 0 {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Foreground(colorRed).Bold(true).Render("CRITICAL:"))
		for _, item := range critical {
			isSelected := (itemIdx == a.attentionCursor)
			line := a.renderAttentionItem(itemIdx+1, item, isSelected, numWidth)
			lines = append(lines, line)
			itemIdx++
		}
	}

	if len(warning) > 0 {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Foreground(colorYellow).Bold(true).Render("WARNING:"))
		for _, item := range warning {
			isSelected := (itemIdx == a.attentionCursor)
			line := a.renderAttentionItem(itemIdx+1, item, isSelected, numWidth)
			lines = append(lines, line)
			itemIdx++
		}
	}

	if len(info) > 0 {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Foreground(colorCyan).Bold(true).Render("INFO:"))
		for _, item := range info {
			isSelected := (itemIdx == a.attentionCursor)
			line := a.renderAttentionItem(itemIdx+1, item, isSelected, numWidth)
			lines = append(lines, line)
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

	// Build status with bundle health, critical count visibility, position indicator, and sort mode
	var statusParts []string

	// BUNDLE HEALTH FIRST (if in bundle mode) - use interface method
	if a.bundleMode && a.dataSource != nil {
		health := a.dataSource.GetBundleHealth()
		if health != nil {
			pct := health.Percentage()
			healthIndicator := fmt.Sprintf("📦 BUNDLE %d%%", pct)
			statusParts = append(statusParts, healthIndicator)
		}
	}

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

	// CRITICAL COUNT (high visibility)
	if totalCriticals > 0 {
		if displayedCriticals < totalCriticals {
			statusParts = append(statusParts, fmt.Sprintf("🔥 Criticals: %d/%d shown", displayedCriticals, totalCriticals))
		} else {
			statusParts = append(statusParts, fmt.Sprintf("🔥 Criticals: %d", totalCriticals))
		}
	}

	// Item count indicator with expansion state clarity (v0.5.7)
	if displayedCount < totalIssues {
		statusParts = append(statusParts, fmt.Sprintf("Showing %d/%d (capped)", displayedCount, totalIssues))
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
	// v0.5.7: Show clear expansion state with exact item count
	if displayedCount < totalIssues {
		statusParts = append(statusParts, fmt.Sprintf("[m]=show all %d", totalIssues))
	} else {
		statusParts = append(statusParts, "[m]=cap")
	}
	statusParts = append(statusParts, "[g/G]=top/bottom")
	statusParts = append(statusParts, "[Enter]=logs")
	statusParts = append(statusParts, "[c]=classic")

	// v0.5.7: Show help hint for first 3 launches
	if a.launchCount < 3 {
		statusParts = append(statusParts, "💡 Press ? for help")
	}

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
// FIX (v0.5.9): Simplified rendering - removed expansion indicators
// FIX (v0.6.7): Added numWidth parameter for dynamic alignment (Issue #18)
func (a *App) renderAttentionItem(num int, item AttentionItem, isSelected bool, numWidth int) string {
	// Format line 1: "1. 💀 nginx-deploy-xyz    CrashLoopBackOff    kube-system"
	numStr := fmt.Sprintf("%*d. ", numWidth-2, num) // Right-align number with dynamic width

	titleWidth := 30
	descWidth := 50 // v0.6.8.1: Increased from 25 to show full node pressure details (Truth Only™)
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

	line1 := fmt.Sprintf("%s%s %-*s  %-*s  %s",
		numStr,
		item.Emoji,
		titleWidth, title,
		descWidth, desc,
		ns,
	)

	// v0.6.7: Add line 2 with diagnostic context if available
	var line2 string
	if item.DiagnosticContext != nil {
		// Format: "   → Root cause | Fix: Recommendation"
		line2Prefix := strings.Repeat(" ", numWidth) // Align with content
		rootCause := item.DiagnosticContext.RootCause
		fix := item.DiagnosticContext.Recommendation

		// Truncate if too long (leave room for formatting)
		maxLen := 80
		if len(rootCause) > maxLen {
			rootCause = rootCause[:maxLen-3] + "..."
		}
		if len(fix) > maxLen {
			fix = fix[:maxLen-3] + "..."
		}

		line2 = fmt.Sprintf("%s→ %s | Fix: %s", line2Prefix, rootCause, fix)
	}

	// Apply selection highlight (inverts colors for visibility)
	if isSelected {
		styledLine1 := lipgloss.NewStyle().
			Background(colorCyan).
			Foreground(colorDarkGray).
			Bold(true).
			Render(line1)

		if line2 != "" {
			styledLine2 := lipgloss.NewStyle().
				Background(colorCyan).
				Foreground(colorDarkGray).
				Render(line2)
			return styledLine1 + "\n" + styledLine2
		}
		return styledLine1
	}

	// Color the entire line based on severity when not selected
	var style lipgloss.Style
	var diagnosticStyle lipgloss.Style
	switch item.Severity {
	case SeverityCritical:
		style = lipgloss.NewStyle().Foreground(colorRed)
		diagnosticStyle = lipgloss.NewStyle().Foreground(colorGray).Italic(true)
	case SeverityWarning:
		style = lipgloss.NewStyle().Foreground(colorYellow)
		diagnosticStyle = lipgloss.NewStyle().Foreground(colorGray).Italic(true)
	case SeverityInfo:
		style = lipgloss.NewStyle().Foreground(colorWhite)
		diagnosticStyle = lipgloss.NewStyle().Foreground(colorGray).Italic(true)
	}

	// Combine line1 and optional line2
	if line2 != "" {
		return style.Render(line1) + "\n" + diagnosticStyle.Render(line2)
	}
	return style.Render(line1)
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
