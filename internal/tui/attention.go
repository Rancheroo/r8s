package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderAttentionDashboard renders the attention dashboard view
func (a *App) renderAttentionDashboard() string {
	if len(a.attentionItems) == 0 {
		return a.renderAllGood()
	}

	// Group items by severity
	critical := []AttentionItem{}
	warning := []AttentionItem{}
	info := []AttentionItem{}

	for _, item := range a.attentionItems {
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
			itemIdx++
		}
	}

	content := strings.Join(lines, "\n")

	// Pad content to fill screen
	contentHeight := a.height - 6
	contentLines := strings.Split(content, "\n")
	if len(contentLines) < contentHeight {
		padding := make([]string, contentHeight-len(contentLines))
		for i := range padding {
			padding[i] = ""
		}
		content = content + "\n" + strings.Join(padding, "\n")
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorRed).
		Padding(1, 2).
		Width(a.width - 4).
		Render(content)

	status := statusStyle.Render(" [1-9] jump · [Enter] logs · [c] classic view · [r] refresh · [q] quit ")

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
	// Format: "1. 💀 nginx-deploy-xyz    CrashLoopBackOff    kube-system"
	numStr := fmt.Sprintf("%d. ", num)

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

	line := fmt.Sprintf("%s%s %-*s  %-*s  %s",
		numStr,
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
