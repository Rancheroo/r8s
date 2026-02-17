package tui

import (
	"fmt"

	"github.com/Rancheroo/r8s/internal/ai"
	"github.com/Rancheroo/r8s/internal/datasource"
)

// detectAIPatterns runs the AI pattern engine on available data sources
// and converts findings to AttentionItems for the dashboard
func detectAIPatterns(ds datasource.DataSource) []AttentionItem {
	var items []AttentionItem

	// Initialize AI engine with built-in patterns
	engine := ai.NewEngine()

	// Scan kubelet logs for patterns
	items = append(items, scanKubeletWithAI(ds, engine)...)

	// Scan dmesg for patterns
	items = append(items, scanDmesgWithAI(ds, engine)...)

	// Scan pod logs for patterns
	items = append(items, scanPodLogsWithAI(ds, engine)...)

	return items
}

// scanKubeletWithAI scans kubelet logs using AI pattern matching
func scanKubeletWithAI(ds datasource.DataSource, engine *ai.Engine) []AttentionItem {
	var items []AttentionItem

	kubeletIssues, err := ds.GetKubeletIssues()
	if err != nil || kubeletIssues == nil || len(kubeletIssues) == 0 {
		return items
	}

	// Convert kubelet issues to text for AI analysis
	for _, issue := range kubeletIssues {
		content := fmt.Sprintf("%s: %s", issue.Pattern, issue.Message)
		metadata := ai.MatchMetadata{
			SourceType: "logs",
		}

		findings := engine.Analyze(content, metadata)
		items = append(items, convertAIFindingsToItems(findings, "kubelet")...)
	}

	return items
}

// scanDmesgWithAI scans dmesg logs using AI pattern matching
func scanDmesgWithAI(ds datasource.DataSource, engine *ai.Engine) []AttentionItem {
	var items []AttentionItem

	oomAnalysis, err := ds.GetOOMAnalysis()
	if err != nil || oomAnalysis == nil {
		return items
	}

	// Analyze OOM events for additional patterns
	for _, oom := range oomAnalysis {
		// Build content from OOM data
		content := fmt.Sprintf("OOM Kill: %s/%s - Memory Limit: %s, QoS: %s",
			oom.PodName, oom.ContainerName, oom.MemoryLimit, oom.QoSClass)
		if oom.IsNodeOOM {
			content = fmt.Sprintf("Node OOM: System memory exhausted on %s", oom.NodeName)
		}

		metadata := ai.MatchMetadata{
			SourceType: "dmesg",
			PodName:    oom.PodName,
			NodeName:   oom.NodeName,
		}

		findings := engine.Analyze(content, metadata)
		items = append(items, convertAIFindingsToItems(findings, "dmesg")...)
	}

	return items
}

// scanPodLogsWithAI scans pod logs for crash patterns
func scanPodLogsWithAI(ds datasource.DataSource, engine *ai.Engine) []AttentionItem {
	var items []AttentionItem

	// Get all pods to scan their logs
	pods, err := ds.GetAllPods()
	if err != nil || pods == nil {
		return items
	}

	// Limit scanning to pods with issues (not running/completed)
	for _, pod := range pods {
		if isHealthyPodState(pod.State) {
			continue // Skip healthy pods
		}

		// Get pod logs (current container)
		containers, err := ds.GetContainers(pod.NamespaceID, pod.Name)
		if err != nil || len(containers) == 0 {
			continue
		}

		// Sample first container's logs
		logs, err := ds.GetLogs("", pod.NamespaceID, pod.Name, containers[0], false)
		if err != nil || len(logs) == 0 {
			continue
		}

		// Join last 50 log lines for analysis
		logContent := joinLogLines(logs, 50)

		metadata := ai.MatchMetadata{
			SourceType:    "logs",
			PodName:       pod.Name,
			Namespace:     pod.NamespaceID,
			ContainerName: containers[0],
		}

		findings := engine.Analyze(logContent, metadata)

		// Only add high-confidence findings from logs to avoid noise
		for _, f := range findings {
			if f.Confidence >= 0.7 {
				items = append(items, aiFindingToItem(f))
			}
		}
	}

	return items
}

// convertAIFindingsToItems converts AI findings to AttentionItems
func convertAIFindingsToItems(findings []ai.Finding, source string) []AttentionItem {
	var items []AttentionItem

	for _, f := range findings {
		items = append(items, aiFindingToItem(f))
	}

	return items
}

// aiFindingToItem converts a single AI Finding to an AttentionItem
func aiFindingToItem(f ai.Finding) AttentionItem {
	// Map AI severity to TUI severity
	var severity AttentionSeverity
	switch f.Severity {
	case "critical":
		severity = SeverityCritical
	case "high":
		severity = SeverityWarning
	case "medium", "low":
		severity = SeverityInfo
	default:
		severity = SeverityInfo
	}

	// Select emoji based on category
	emoji := "🤖" // Default AI indicator
	switch f.Category {
	case "resource":
		emoji = "💾"
	case "image":
		emoji = "📦"
	case "crash":
		emoji = "💥"
	}

	// Build description with confidence indicator
	description := f.Message
	if f.Confidence < 1.0 {
		description = fmt.Sprintf("%s (%.0f%% confidence)", f.Message, f.Confidence*100)
	}

	return AttentionItem{
		Severity:      severity,
		Emoji:         emoji,
		Title:         f.PatternName,
		Description:   description,
		Namespace:     f.Context.Namespace,
		ResourceType:  "ai-detected",
		PodName:       f.Context.PodName,
		ContainerName: f.Context.ContainerName,
		Count:         1,
		Timestamp:     f.Timestamp,
	}
}

// isHealthyPodState returns true if pod state indicates no issues
func isHealthyPodState(state string) bool {
	switch state {
	case "running", "Running", "completed", "Completed", "succeeded", "Succeeded":
		return true
	default:
		return false
	}
}

// joinLogLines combines log lines into a single string for analysis
func joinLogLines(lines []string, maxLines int) string {
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}

	result := ""
	for _, line := range lines {
		result += line + "\n"
	}
	return result
}
