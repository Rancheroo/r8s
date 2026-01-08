package bundle

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// KubeletIssue represents a node-level issue detected from kubelet logs
type KubeletIssue struct {
	Pattern  string // "HTTP 502", "connection timeout", etc.
	Message  string // Full error message
	Count    int    // Occurrences
	LastSeen string // Timestamp
	Severity string // "error", "warning"
}

// ParseKubeletLogs parses kubelet logs from journald/rke2-server for node-level issues
func ParseKubeletLogs(extractPath string) ([]KubeletIssue, error) {
	bundleRoot := getBundleRoot(extractPath)
	journalPath := filepath.Join(bundleRoot, "journald/rke2-server")

	content, err := os.ReadFile(journalPath)
	if err != nil {
		// Gracefully handle missing journal file
		return nil, nil
	}

	return parseJournaldLogs(string(content))
}

// parseJournaldLogs parses the journald content for error patterns
func parseJournaldLogs(content string) ([]KubeletIssue, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	issues := make(map[string]*KubeletIssue)

	// Define error patterns to detect (based on log analysis)
	patterns := map[string]*regexp.Regexp{
		"HTTP 502":              regexp.MustCompile(`Sending HTTP/1\.1 502 response`),
		"connection timeout":    regexp.MustCompile(`connect: connection timed out`),
		"remotedialer timeout":  regexp.MustCompile(`remotedialer server.*i/o timeout`),
		"DNS nameserver limits": regexp.MustCompile(`Nameserver limits exceeded`),
		"TLS handshake error":   regexp.MustCompile(`TLS handshake error`),
		"EOF during TLS":        regexp.MustCompile(`EOF.*TLS`),
		"unsupported protocol":  regexp.MustCompile(`protocol.*not supported`),
		"version not supported": regexp.MustCompile(`version.*not supported`),
		"no route to host":      regexp.MustCompile(`no route to host`),
		"network unreachable":   regexp.MustCompile(`network is unreachable`),
	}

	// Parse each line
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Extract timestamp and log level
		timestamp, level, message := parseJournaldLine(line)
		if message == "" {
			continue
		}

		// Check for error patterns
		for patternName, regex := range patterns {
			if regex.MatchString(message) {
				key := patternName

				if issues[key] == nil {
					issues[key] = &KubeletIssue{
						Pattern:  patternName,
						Severity: level,
					}
				}

				issues[key].Count++
				issues[key].Message = message // Keep the last matching message
				if timestamp != "" {
					issues[key].LastSeen = timestamp
				}
				break // Only match one pattern per line
			}
		}
	}

	// Convert map to slice
	var result []KubeletIssue
	for _, issue := range issues {
		result = append(result, *issue)
	}

	return result, nil
}

// parseJournaldLine extracts timestamp, level, and message from a journald line
// Format: Dec 13 14:06:17 hostname rke2[PID]: time="..." level=error msg="..."
func parseJournaldLine(line string) (timestamp, level, message string) {
	// Extract timestamp from beginning (e.g., "Dec 13 14:06:17")
	if len(line) < 15 {
		return "", "", ""
	}

	// Look for the colon after the service name
	colonIndex := strings.Index(line, ": ")
	if colonIndex == -1 {
		return "", "", ""
	}

	timestamp = strings.TrimSpace(line[:colonIndex])

	// Extract the log message part after the colon
	logPart := line[colonIndex+2:]

	// Parse the structured log format: time="..." level=error msg="..."
	timeRegex := regexp.MustCompile(`time="([^"]*)"`)
	levelRegex := regexp.MustCompile(`level=(\w+)`)
	msgRegex := regexp.MustCompile(`msg="([^"]*)"`)

	timeMatch := timeRegex.FindStringSubmatch(logPart)
	levelMatch := levelRegex.FindStringSubmatch(logPart)
	msgMatch := msgRegex.FindStringSubmatch(logPart)

	if len(msgMatch) > 1 {
		message = msgMatch[1]
	} else {
		// Fallback: use the whole log part if structured parsing fails
		message = logPart
	}

	if len(levelMatch) > 1 {
		level = levelMatch[1]
	} else {
		level = "info" // default
	}

	// Use structured timestamp if available
	if len(timeMatch) > 1 {
		timestamp = timeMatch[1]
	}

	return timestamp, level, message
}
