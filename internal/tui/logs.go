// Package tui - Log viewing, filtering, and colorization functionality.
// This file handles all log-related operations including rendering, filtering,
// search, color detection, and level-based highlighting.
package tui

import (
	"fmt"
	"strings"

	"github.com/Rancheroo/r8s/internal/rancher"
	"github.com/charmbracelet/lipgloss"
)

// Package-level pattern slices allocated once for performance
var errorPatterns = []string{
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

var warnKeywords = []string{
	"WARN=",
	"DEPRECATED",
	"DEPRECATION",
	"ALERT:",
	"ALERT=",
}

// renderLogsView renders the logs view for a pod with viewport scrolling
// DIAGNOSTIC-FIRST APPROACH: Always show diagnostic panel first, then 'l' to view logs
func (a *App) renderLogsView() string {
	// If user pressed 'l' to view raw logs
	if a.showRawLogs {
		// If no logs available, show simple message
		if len(a.logs) == 0 {
			return a.renderNoLogsMessage()
		}
		// Show raw logs
		return a.renderRawLogs()
	}

	// Default: Show diagnostic panel (intelligence first!)
	return a.renderEmptyLogsHelp()
}

// renderRawLogs renders the raw log viewer
func (a *App) renderRawLogs() string {

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
		statusText = " [/] search  [l] diagnostics  [Ctrl+E] errors  [Ctrl+W] warnings  [Esc] back  [q] quit "
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

// applyLogFilter applies the current log level filter to the logs with colors
func (a *App) applyLogFilter() {
	// Use colored rendering for all content
	a.logViewport.SetContent(a.renderLogsWithColors())
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
			// Wrap raw text into segments FIRST, preferring to break at whitespace
			remainingLine := line
			segmentIndex := 0
			for len(remainingLine) > 0 {
				// Determine segment length, preferring whitespace breaks
				segmentEnd := wrapWidth
				if segmentEnd > len(remainingLine) {
					segmentEnd = len(remainingLine)
				}

				// If we're not at the end, try to find last whitespace before wrapWidth
				if segmentEnd < len(remainingLine) {
					// Look for the last space/whitespace before wrapWidth
					lastSpace := -1
					for idx := segmentEnd - 1; idx >= 0; idx-- {
						r := rune(remainingLine[idx])
						if r == ' ' || r == '\t' {
							lastSpace = idx
							break
						}
					}
					// Use the whitespace break if found, otherwise use wrapWidth
					if lastSpace > 0 {
						segmentEnd = lastSpace
					}
				}

				segment := remainingLine[:segmentEnd]

				// Trim leading spaces on wrapped segments (not first segment of current line)
				if segmentIndex > 0 {
					segment = strings.TrimLeft(segment, " \t")
				}

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
				segmentIndex++
			}
		}
	}

	return strings.Join(wrappedLines, "\n")
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

// isErrorLog detects ERROR level logs with explicit indicator priority
// Priority: [ERROR] or E#### > keyword patterns
// isErrorLog reports whether the given log line should be classified as an ERROR-level entry.
//
// It performs case-insensitive checks in priority order: first it excludes lines that explicitly
// indicate non-error levels (WARN, INFO, DEBUG and Kubernetes-style W/I/D prefixes at line start),
// then it accepts explicit error indicators ([ERROR], Kubernetes E#### at line start, or "LEVEL=ERROR"),
// and finally it matches configured keyword patterns from the package-level errorPatterns slice.
// Returns true if any error indicator is found, false otherwise.
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
	// Use package-level errorPatterns slice
	for _, pattern := range errorPatterns {
		if strings.Contains(lineUpper, pattern) {
			return true
		}
	}

	return false
}

// isWarnLog reports whether the given log line indicates a warning level.
// It recognizes explicit warning markers (e.g., "[WARN]", "[WARNING]", "WARN:"/"WARNING:", "LEVEL=WARN") and Kubernetes-style "W####" prefixes, and also matches configured warning keyword patterns; lines containing explicit error indicators are treated as errors and do not qualify as warnings.
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

	// Use package-level warnKeywords slice
	for _, pattern := range warnKeywords {
		if strings.Contains(lineUpper, pattern) {
			return true
		}
	}

	return false
}

// 3. Key-value form: contains "LEVEL=INFO".
func isInfoLog(line string) bool {
	lineUpper := strings.ToUpper(line)
	// Bracketed format: [INFO]
	if strings.Contains(lineUpper, "[INFO]") {
		return true
	}
	// K8s format at line start: I#### (check first 5 chars only)
	if len(line) >= 5 && line[0] == 'I' && isDigit(line[1]) && isDigit(line[2]) &&
		isDigit(line[3]) && isDigit(line[4]) {
		return true
	}
	// Also check for level=info format (anywhere in line)
	if strings.Contains(lineUpper, "LEVEL=INFO") {
		return true
	}
	return false
}

// isDebugLog reports whether the log line indicates DEBUG level.
// It recognizes bracketed "[DEBUG]", Kubernetes-style "D####" at the start (checks the first five characters), and a case-insensitive "LEVEL=DEBUG" anywhere in the line.
func isDebugLog(line string) bool {
	lineUpper := strings.ToUpper(line)
	// Bracketed format: [DEBUG]
	if strings.Contains(lineUpper, "[DEBUG]") {
		return true
	}
	// K8s format at line start: D#### (check first 5 chars only)
	if len(line) >= 5 && line[0] == 'D' && isDigit(line[1]) && isDigit(line[2]) &&
		isDigit(line[3]) && isDigit(line[4]) {
		return true
	}
	// Also check for level=debug format (anywhere in line)
	if strings.Contains(lineUpper, "LEVEL=DEBUG") {
		return true
	}
	return false
}

// isDigit reports whether b is an ASCII digit ('0' through '9').
func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

// renderEmptyLogsHelp renders helpful guidance when logs are empty
// Auto-displays without button presses (Show, Don't Ask philosophy)
func (a *App) renderEmptyLogsHelp() string {
	breadcrumb := breadcrumbStyle.Render(a.getBreadcrumb())

	// Get pod information to enrich the diagnostic panel
	// Fetch from dataSource directly (not a.pods which may be empty if coming from Dashboard)
	var foundPod *rancher.Pod

	if a.dataSource != nil {
		pods, err := a.dataSource.GetAllPods()
		if err == nil {
			for _, pod := range pods {
				if pod.Name == a.currentView.podName {
					foundPod = &pod
					break
				}
			}
		}
	}

	if foundPod == nil {
		// Fallback if pod not found
		return a.renderEmptyLogsFallback(breadcrumb)
	}

	// Build enhanced diagnostic panel with maximum intel
	return a.renderMaximumIntelPanel(breadcrumb, foundPod)
}

// renderMaximumIntelPanel renders the enhanced "Maximum Intel" diagnostic panel
func (a *App) renderMaximumIntelPanel(breadcrumb string, pod *rancher.Pod) string {
	// Extract pod diagnostics
	state := pod.KubectlStatus
	if state == "" {
		state = pod.State
	}
	restarts := pod.KubectlRestarts
	if restarts == 0 {
		restarts = pod.RestartCount
	}
	ready := pod.KubectlReady
	node := pod.NodeName
	if node == "" {
		node = pod.NodeID
	}
	age := pod.KubectlAge

	// Build title - context-aware based on log availability
	var titleText string
	if len(a.logs) > 0 {
		titleText = fmt.Sprintf("📊 Pod Diagnostics - %s", pod.Name)
	} else {
		titleText = fmt.Sprintf("📭 No Logs Available - Pod: %s", pod.Name)
	}

	helpTitle := lipgloss.NewStyle().
		Foreground(colorYellow).
		Bold(true).
		Render(titleText)

	// Build diagnostic sections
	var sections []string

	// Section 1: Diagnosis
	diagnosisSection := a.buildDiagnosisSection(state, restarts, age)
	sections = append(sections, diagnosisSection)

	// Section 2: Pod Status
	statusSection := fmt.Sprintf(`━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
💊 POD STATUS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  State:       %s
  Restarts:    %d
  Ready:       %s
  Node:        %s
  Age:         %s`, state, restarts, ready, node, age)
	sections = append(sections, statusSection)

	// Section 3: Container Status (NEW in v0.5.6)
	var podEvents []rancher.Event
	if a.dataSource != nil {
		podEvents, _ = a.dataSource.GetEventsByPod(a.currentView.namespaceName, pod.Name)
	}
	containerSection := a.buildContainerStatusSection(pod, podEvents)
	sections = append(sections, containerSection)

	// Section 4: Recent Events (fetch from global events file for complete history)
	// Fallback to pod.KubectlEvents if no global events (convert strings to display)
	eventsSection := a.buildEventsSection(podEvents)
	sections = append(sections, eventsSection)

	// Section 5: Investigation Guidance
	investigateSection := a.buildInvestigationSection(state, restarts)
	sections = append(sections, investigateSection)

	// Section 6: External Tools
	toolsSection := a.buildExternalToolsSection()
	sections = append(sections, toolsSection)

	// Combine all sections
	helpText := strings.Join(sections, "\n\n")

	helpBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorYellow).
		Padding(1, 2).
		Width(a.width - 6).
		Render(helpText)

	status := statusStyle.Render(" [l]=view logs  [Ctrl+P]=previous logs  [Esc]=back  [q]=quit ")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		breadcrumb,
		"",
		helpTitle,
		"",
		helpBox,
		"",
		status,
	)
}

// buildDiagnosisSection builds intelligent diagnosis based on pod state
func (a *App) buildDiagnosisSection(state string, restarts int, age string) string {
	stateUpper := strings.ToUpper(state)

	var diagnosis string
	var emoji string

	// Diagnose based on state and restart count
	if strings.Contains(stateUpper, "CRASHLOOPBACKOFF") || strings.Contains(stateUpper, "CRASHLOOP") {
		emoji = "🔴"
		if restarts > 0 && age != "" {
			diagnosis = fmt.Sprintf("CRASH LOOP DETECTED\n     %d restarts - Container repeatedly failing to start\n     Container never stayed Running long enough to generate logs", restarts)
		} else {
			diagnosis = "CRASH LOOP DETECTED\n     Container repeatedly failing to start\n     Check exit codes and termination reasons in pod description"
		}
	} else if strings.Contains(stateUpper, "OOMKILLED") {
		emoji = "🔴"
		diagnosis = "OUT OF MEMORY ERROR\n     Container exceeded memory limits and was killed\n     No logs because container terminated before flushing to disk"
	} else if strings.Contains(stateUpper, "IMAGEPULLBACKOFF") || strings.Contains(stateUpper, "ERRIMAGEBACKOFF") {
		emoji = "🟡"
		diagnosis = "IMAGE PULL FAILURE\n     Cannot pull container image from registry\n     Check image name, registry auth, and network connectivity"
	} else if strings.Contains(stateUpper, "ERROR") {
		emoji = "🔴"
		diagnosis = "POD IN ERROR STATE\n     Container failed to start or run\n     Check container configuration and dependencies"
	} else if strings.Contains(stateUpper, "PENDING") {
		emoji = "🟡"
		diagnosis = "POD PENDING\n     Pod not yet scheduled or containers not started\n     Check node resources and scheduling constraints"
	} else if strings.Contains(stateUpper, "EVICTED") {
		emoji = "🟣"
		diagnosis = "POD EVICTED\n     Removed due to resource pressure on node\n     Check node memory/disk pressure events"
	} else if restarts >= 10 {
		emoji = "🟠"
		diagnosis = fmt.Sprintf("HIGH RESTART COUNT\n     %d restarts detected - Instability pattern\n     Container may be crashing intermittently", restarts)
	} else if restarts >= 3 {
		emoji = "🟡"
		diagnosis = fmt.Sprintf("MODERATE RESTARTS\n     %d restarts detected\n     Container experiencing some instability", restarts)
	} else {
		emoji = "ℹ️"
		diagnosis = "NO LOGS GENERATED\n     Container either:\n     • Never started successfully\n     • Started but didn't write any logs\n     • Logs not captured in bundle"
	}

	return fmt.Sprintf(`━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
💡 DIAGNOSIS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  %s %s`, emoji, diagnosis)
}

// buildContainerStatusSection shows container-level status details (v0.5.6)
func (a *App) buildContainerStatusSection(pod *rancher.Pod, events []rancher.Event) string {
	header := `━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📋 CONTAINER STATUS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`

	// Parse Ready field (e.g., "1/2" -> 1 ready, 2 total)
	ready := pod.KubectlReady
	if ready == "" {
		return header + "\n\n  Container status not available in bundle"
	}

	// Extract ready/total counts
	var readyCount, totalCount int
	fmt.Sscanf(ready, "%d/%d", &readyCount, &totalCount)

	// Build status summary
	statusLine := fmt.Sprintf("  Containers: %d/%d ready", readyCount, totalCount)

	// Extract container info from events using ContainerName field (v0.5.6)
	containerMap := make(map[string]string) // container name -> status emoji

	// First pass: Look for BackOff/Failed events (mark containers as failed)
	for _, event := range events {
		if event.ContainerName == "" {
			continue // Skip events without container info
		}

		if event.Reason == "BackOff" || (event.Type == "Warning" && event.Reason == "Failed") {
			containerMap[event.ContainerName] = "❌"
		}
	}

	// Second pass: Look for Started/Created events (collect container names)
	for _, event := range events {
		if event.ContainerName == "" {
			continue
		}

		if event.Reason == "Started" || event.Reason == "Created" {
			// Only set if not already marked as failed
			if containerMap[event.ContainerName] == "" {
				// Initially mark as started (will determine actual status below)
				containerMap[event.ContainerName] = "started"
			}
		}
	}

	// Third pass: Apply pod-level status to determine final container status
	// If pod is not fully ready, containers without BackOff events are likely failing
	for containerName, status := range containerMap {
		if status == "started" {
			if readyCount == totalCount {
				// All containers ready - mark as success
				containerMap[containerName] = "✅"
			} else if readyCount == 0 {
				// No containers ready - mark as failing (even without BackOff)
				containerMap[containerName] = "❌"
			} else {
				// Some containers ready - unknown which ones are failing
				containerMap[containerName] = "⚠️"
			}
		}
	}

	// If we identified containers from events, show them
	var containerLines []string
	if len(containerMap) > 0 {
		containerLines = append(containerLines, "")
		for containerName, status := range containerMap {
			containerLines = append(containerLines, fmt.Sprintf("  %s %s", status, containerName))
		}
	}

	// Add diagnostic insights from events (v0.5.6 enhanced)
	var diagnosticLines []string

	// Analyze BackOff patterns for timing intelligence
	var backoffContainers []string
	backoffCount := 0
	for _, event := range events {
		if event.Reason == "BackOff" && event.ContainerName != "" {
			backoffCount += event.Count
			// Track which containers are in backoff
			found := false
			for _, c := range backoffContainers {
				if c == event.ContainerName {
					found = true
					break
				}
			}
			if !found {
				backoffContainers = append(backoffContainers, event.ContainerName)
			}
		}
	}

	if backoffCount > 0 {
		diagnosticLines = append(diagnosticLines, "")
		diagnosticLines = append(diagnosticLines, fmt.Sprintf("  🔄 Restart pattern: %d backoff events detected", backoffCount))
		if len(backoffContainers) > 0 {
			diagnosticLines = append(diagnosticLines, fmt.Sprintf("     Affected containers: %v", backoffContainers))
		}
	}

	// Analyze Failed events for error details
	for _, event := range events {
		if event.Type == "Warning" && event.Reason == "Failed" && event.ContainerName != "" {
			diagnosticLines = append(diagnosticLines, "")
			diagnosticLines = append(diagnosticLines, fmt.Sprintf("  ❌ %s: %s", event.ContainerName, event.Message))
		}
	}

	// Add note about data limitations
	dataGapNote := `

  ⚠️  Exit codes not captured in bundle
  → For detailed container status, run:
    kubectl describe pod ` + pod.Name + ` -n ` + pod.NamespaceID

	// Combine all parts
	if len(containerLines) > 0 {
		return header + "\n\n" + statusLine + strings.Join(containerLines, "\n") + dataGapNote
	}

	return header + "\n\n" + statusLine + dataGapNote
}

// buildEventsSection formats recent pod events
func (a *App) buildEventsSection(events []rancher.Event) string {
	header := `━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📋 RECENT EVENTS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`

	if len(events) == 0 {
		return header + "\n\n  No events recorded for this pod"
	}

	// Show last 5 events max
	maxEvents := 5
	displayEvents := events
	if len(events) > maxEvents {
		displayEvents = events[:maxEvents]
	}

	var eventLines []string
	for _, event := range displayEvents {
		// Choose emoji based on event type
		emoji := "ℹ️"
		if event.Type == "Warning" {
			emoji = "⚠️"
		}

		// Add count indicator if event occurred multiple times
		countStr := ""
		if event.Count > 1 {
			countStr = fmt.Sprintf(" (x%d)", event.Count)
		}

		// Format: emoji + reason + message + count
		eventLines = append(eventLines, fmt.Sprintf("  %s  %s: %s%s",
			emoji, event.Reason, event.Message, countStr))
	}

	return header + "\n\n" + strings.Join(eventLines, "\n")
}

// buildInvestigationSection provides actionable next steps
func (a *App) buildInvestigationSection(state string, restarts int) string {
	stateUpper := strings.ToUpper(state)

	var suggestions []string

	// State-specific suggestions
	if strings.Contains(stateUpper, "OOMKILLED") {
		suggestions = []string{
			"Check memory limits - OOMKilled signals resource exhaustion",
			"Review resource requests/limits in pod spec",
			"Check if recent config change increased memory usage",
		}
	} else if strings.Contains(stateUpper, "CRASHLOOP") || restarts >= 10 {
		suggestions = []string{
			"Check container exit codes in pod description",
			"Review application startup configuration",
			"Check dependency services (DB, API endpoints, etc.)",
		}
	} else if strings.Contains(stateUpper, "IMAGEPULL") {
		suggestions = []string{
			"Verify image name and tag are correct",
			"Check registry authentication (imagePullSecrets)",
			"Verify network connectivity to registry",
		}
	} else if strings.Contains(stateUpper, "PENDING") {
		suggestions = []string{
			"Check node resources (CPU, memory, disk)",
			"Review pod scheduling constraints (nodeSelector, taints)",
			"Verify required volumes are available",
		}
	} else {
		suggestions = []string{
			"Check recent pod events above for failure patterns",
			"Try [Ctrl+P] to view previous container logs if available",
			"Review pod logs (if available) for application errors",
		}
	}

	header := `━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🔍 INVESTIGATE NEXT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`

	var steps []string
	for i, suggestion := range suggestions {
		steps = append(steps, fmt.Sprintf("  %d. %s", i+1, suggestion))
	}

	return header + "\n\n" + strings.Join(steps, "\n")
}

// buildExternalToolsSection provides guidance on external analysis tools
func (a *App) buildExternalToolsSection() string {
	return `━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🛠️  EXTERNAL TOOLS FOR LOG ANALYSIS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  When logs ARE available elsewhere:

  • lnav - Advanced terminal log viewer with filtering & search
    Install: brew install lnav  OR  apt install lnav
    Usage:   lnav /path/to/bundle/rke2/podlogs/namespace_*/

  • kubectl logs - View logs from live cluster
    Usage:   kubectl logs <pod> -n <namespace>
    Previous: kubectl logs <pod> -n <namespace> --previous

  • Integrated in r8s - Press [Ctrl+P] for previous container logs`
}

// renderNoLogsMessage renders a simple "no logs" message when user pressed 'l' but pod has no logs
func (a *App) renderNoLogsMessage() string {
	breadcrumb := breadcrumbStyle.Render(a.getBreadcrumb())

	helpTitle := lipgloss.NewStyle().
		Foreground(colorYellow).
		Bold(true).
		Render("📭 No Logs Available")

	helpText := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorYellow).
		Padding(2, 4).
		Width(a.width - 8).
		Render("No container logs available for this pod")

	status := statusStyle.Render(" [Esc]=back to diagnostics  [q]=quit ")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		breadcrumb,
		"",
		helpTitle,
		"",
		helpText,
		"",
		status,
	)
}

// renderEmptyLogsFallback renders a simple fallback when pod not found
func (a *App) renderEmptyLogsFallback(breadcrumb string) string {
	helpTitle := lipgloss.NewStyle().
		Foreground(colorYellow).
		Bold(true).
		Render("📭 No Logs Available")

	helpText := `No container logs available

Possible reasons:
  • Pod hasn't generated any logs yet
  • Pod recently restarted (try Ctrl+P for previous logs)
  • Container hasn't started successfully

Next steps:
  • Press 'd' to describe pod (check status/events)
  • Press Esc to go back and check pod state
  • Look for errors in pod description`

	helpBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorYellow).
		Padding(2, 4).
		Width(a.width - 8).
		Render(helpText)

	// Calculate padding for vertical centering
	contentHeight := 12
	availableHeight := a.height - 10
	topPadding := (availableHeight - contentHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	paddingStr := strings.Repeat("\n", topPadding)
	centeredContent := paddingStr + strings.Join([]string{helpTitle, "", helpBox}, "\n")

	status := statusStyle.Render(" [d]=describe pod  [Ctrl+P]=previous logs  [Esc]=back  [q]=quit ")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		breadcrumb,
		"",
		centeredContent,
		"",
		status,
	)
}
