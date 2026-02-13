package bundle

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// JournaldEntry represents a single journald log entry
// NOTE: Currently unused but kept for future structured parsing
// TODO: Implement structured journald parsing or remove in cleanup
//lint:ignore U1000 Reserved for future use
type JournaldEntry struct {
	Timestamp   time.Time
	Unit        string           // Service unit name (e.g., rke2-server)
	Message     string           // Log message
	Priority    string           // debug, info, warn, error, fatal
	PID         string           // Process ID
	RawLine     string           // Original log line for reference
}

// RKE2ControlPlaneEvents contains extracted control plane events
type RKE2ControlPlaneEvents struct {
	ServerRestarts   []ControlPlaneEvent
	AgentIssues      []ControlPlaneEvent
	CertificateIssues []ControlPlaneEvent
	EtcdIssues       []ControlPlaneEvent
	APIServerIssues  []ControlPlaneEvent
	UnknownErrors    []ControlPlaneEvent
}

// ControlPlaneEvent represents a single control plane event
type ControlPlaneEvent struct {
	Timestamp time.Time
	Unit      string
	Message   string
	Severity  string // info, warning, error, critical
}

// ParseJournald parses journald logs for RKE2 control plane components
// Bundle paths: systemlogs/journald-rke2-server, systemlogs/journald-rke2-agent
func ParseJournald(extractPath string) (*RKE2ControlPlaneEvents, error) {
	bundleRoot := getBundleRoot(extractPath)
	events := &RKE2ControlPlaneEvents{
		ServerRestarts:    make([]ControlPlaneEvent, 0),
		AgentIssues:       make([]ControlPlaneEvent, 0),
		CertificateIssues: make([]ControlPlaneEvent, 0),
		EtcdIssues:        make([]ControlPlaneEvent, 0),
		APIServerIssues:   make([]ControlPlaneEvent, 0),
		UnknownErrors:     make([]ControlPlaneEvent, 0),
	}

	// Try multiple possible paths for journald logs
	possiblePaths := []string{
		"systemlogs/journald-rke2-server",
		"systemlogs/journald-rke2-agent",
		"rke2/agent-logs/rke2-server",
		"rke2/agent-logs/rke2-agent",
		"systemlogs/syslog", // Fallback to syslog
	}

	for _, path := range possiblePaths {
		fullPath := filepath.Join(bundleRoot, path)
		if _, err := os.Stat(fullPath); err == nil {
			parseJournaldFile(fullPath, events)
		}
	}

	return events, nil
}

// parseJournaldFile parses a single journald log file
func parseJournaldFile(filePath string, events *RKE2ControlPlaneEvents) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	// Detect file type from path
	unit := detectUnitFromPath(filePath)

	scanner := bufio.NewScanner(file)

	// Common RKE2 error patterns
	patterns := map[string]*regexp.Regexp{
		"server_restart":   regexp.MustCompile(`(?i)rke2.*starting|Starting.*rke2|started.*rke2`),
		"cert_error":       regexp.MustCompile(`(?i)certificate.*error|cert.*failed|x509|tls.*error`),
		"etcd_error":       regexp.MustCompile(`(?i)etcd.*error|etcd.*unhealthy|etcd.*timeout|raft.*error`),
		"apiserver_error":  regexp.MustCompile(`(?i)apiserver.*error|api.*server.*failed|k8s.*io.*error`),
		"agent_connection": regexp.MustCompile(`(?i)failed.*to.*connect|connection.*refused|dial.*tcp.*i/o.*timeout`),
		"fatal_error":      regexp.MustCompile(`(?i)fatal|panic|critical.*error`),
	}

	for scanner.Scan() {
		line := scanner.Text()

		// Try to parse timestamp (various formats)
		timestamp := parseJournaldTimestamp(line)

		// Check against patterns
		for errorType, pattern := range patterns {
			if pattern.MatchString(line) {
				event := ControlPlaneEvent{
					Timestamp: timestamp,
					Unit:      unit,
					Message:   extractMessage(line),
					Severity:  determineSeverity(errorType, line),
				}

				switch errorType {
				case "server_restart":
					events.ServerRestarts = append(events.ServerRestarts, event)
				case "cert_error":
					events.CertificateIssues = append(events.CertificateIssues, event)
				case "etcd_error":
					events.EtcdIssues = append(events.EtcdIssues, event)
				case "apiserver_error":
					events.APIServerIssues = append(events.APIServerIssues, event)
				case "agent_connection", "fatal_error":
					events.AgentIssues = append(events.AgentIssues, event)
				default:
					events.UnknownErrors = append(events.UnknownErrors, event)
				}
				break // One classification per line (first match wins)
			}
		}
	}
}

// detectUnitFromPath extracts the service unit from the file path
func detectUnitFromPath(path string) string {
	path = strings.ToLower(path)
	if strings.Contains(path, "rke2-server") {
		return "rke2-server"
	}
	if strings.Contains(path, "rke2-agent") {
		return "rke2-agent"
	}
	if strings.Contains(path, "k3s") {
		return "k3s"
	}
	return "unknown"
}

// parseJournaldTimestamp attempts to parse various timestamp formats
func parseJournaldTimestamp(line string) time.Time {
	// Try common syslog/journald formats
	formats := []string{
		"Jan 02 15:04:05",
		"Jan  2 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if len(line) >= len(format) {
			if t, err := time.Parse(format, line[:len(format)]); err == nil {
				// Assume current year if not specified
				if t.Year() == 0 {
					t = t.AddDate(time.Now().Year(), 0, 0)
				}
				return t
			}
		}
	}

	return time.Now()
}

// extractMessage extracts the message part from a log line
func extractMessage(line string) string {
	// Look for "]: " which typically follows syslog PID brackets
	if idx := strings.Index(line, "]: "); idx != -1 {
		return strings.TrimSpace(line[idx+3:])
	}
	// Fallback: look for first ": " after position 15 (skip timestamp region)
	if len(line) > 15 {
		if idx := strings.Index(line[15:], ": "); idx != -1 {
			return strings.TrimSpace(line[15+idx+2:])
		}
	}
	return strings.TrimSpace(line)
}

// determineSeverity determines the severity of an event
func determineSeverity(errorType, message string) string {
	messageLower := strings.ToLower(message)

	if strings.Contains(messageLower, "fatal") ||
	   strings.Contains(messageLower, "panic") ||
	   strings.Contains(messageLower, "critical") {
		return "critical"
	}

	if strings.Contains(messageLower, "error") ||
	   errorType == "cert_error" ||
	   errorType == "etcd_error" {
		return "error"
	}

	if strings.Contains(messageLower, "warn") {
		return "warning"
	}

	return "info"
}

// HasIssues returns true if any control plane issues were detected
func (e *RKE2ControlPlaneEvents) HasIssues() bool {
	return len(e.ServerRestarts) > 0 ||
		len(e.AgentIssues) > 0 ||
		len(e.CertificateIssues) > 0 ||
		len(e.EtcdIssues) > 0 ||
		len(e.APIServerIssues) > 0 ||
		len(e.UnknownErrors) > 0
}

// GetSummary returns a human-readable summary
func (e *RKE2ControlPlaneEvents) GetSummary() string {
	if !e.HasIssues() {
		return "No RKE2 control plane issues detected"
	}

	var parts []string
	parts = append(parts, "RKE2 Control Plane Issues Detected:")

	if len(e.ServerRestarts) > 0 {
		parts = append(parts, fmt.Sprintf("  • %d server restart(s)", len(e.ServerRestarts)))
	}
	if len(e.EtcdIssues) > 0 {
		parts = append(parts, fmt.Sprintf("  • %d etcd issue(s)", len(e.EtcdIssues)))
	}
	if len(e.CertificateIssues) > 0 {
		parts = append(parts, fmt.Sprintf("  • %d certificate issue(s)", len(e.CertificateIssues)))
	}
	if len(e.APIServerIssues) > 0 {
		parts = append(parts, fmt.Sprintf("  • %d API server issue(s)", len(e.APIServerIssues)))
	}
	if len(e.AgentIssues) > 0 {
		parts = append(parts, fmt.Sprintf("  • %d agent connection issue(s)", len(e.AgentIssues)))
	}
	if len(e.UnknownErrors) > 0 {
		parts = append(parts, fmt.Sprintf("  • %d unknown error(s)", len(e.UnknownErrors)))
	}

	return strings.Join(parts, "\n")
}

// GetCriticalIssues returns only critical and error severity issues
func (e *RKE2ControlPlaneEvents) GetCriticalIssues() []ControlPlaneEvent {
	var critical []ControlPlaneEvent

	for _, event := range e.ServerRestarts {
		if event.Severity == "critical" || event.Severity == "error" {
			critical = append(critical, event)
		}
	}
	for _, event := range e.EtcdIssues {
		if event.Severity == "critical" || event.Severity == "error" {
			critical = append(critical, event)
		}
	}
	for _, event := range e.CertificateIssues {
		if event.Severity == "critical" || event.Severity == "error" {
			critical = append(critical, event)
		}
	}
	for _, event := range e.APIServerIssues {
		if event.Severity == "critical" || event.Severity == "error" {
			critical = append(critical, event)
		}
	}
	for _, event := range e.AgentIssues {
		if event.Severity == "critical" || event.Severity == "error" {
			critical = append(critical, event)
		}
	}
	for _, event := range e.UnknownErrors {
		if event.Severity == "critical" || event.Severity == "error" {
			critical = append(critical, event)
		}
	}

	return critical
}
