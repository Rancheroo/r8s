// Package tui - Log viewing, filtering, and colorization functionality.
// This file handles all log-related operations including rendering, filtering,
// search, color detection, and level-based highlighting.
package tui

import (
	"fmt"
	"strings"

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
func (a *App) renderLogsView() string {
	// Auto-show helpful message when logs are empty (Show, Don't Ask philosophy)
	if len(a.logs) == 0 {
		return a.renderEmptyLogsHelp()
	}

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
	// Use package-level errorPatterns slice
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

	// Use package-level warnKeywords slice
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

// isDebugLog detects DEBUG level logs in both bracketed and K8s formats
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

// isDigit checks if a byte is an ASCII digit
func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

// renderEmptyLogsHelp renders helpful guidance when logs are empty
// Auto-displays without button presses (Show, Don't Ask philosophy)
func (a *App) renderEmptyLogsHelp() string {
	breadcrumb := breadcrumbStyle.Render(a.getBreadcrumb())

	// Build helpful message
	helpTitle := lipgloss.NewStyle().
		Foreground(colorYellow).
		Bold(true).
		Render("📭 No Logs Available")

	helpText := `This pod has no logs to display.

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
	contentHeight := 12 // Approximate lines of content
	availableHeight := a.height - 10
	topPadding := (availableHeight - contentHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	// Use strings.Repeat for vertical padding
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
