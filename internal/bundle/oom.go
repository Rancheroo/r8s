package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// OOMAnalysis represents an out-of-memory event analysis
type OOMAnalysis struct {
	PodName       string
	ContainerName string
	MemoryLimit   string // "1Gi"
	MemoryRequest string // "512Mi"
	OOMKillTime   string
	IsNodeOOM     bool // vs container OOM
}

// AnalyzeOOMEvents analyzes kubectl events and pod data to identify OOM kills
func AnalyzeOOMEvents(extractPath string) ([]OOMAnalysis, error) {
	bundleRoot := getBundleRoot(extractPath)

	// Parse events first to find OOM kill messages
	eventsPath := filepath.Join(bundleRoot, "rke2/kubectl/events")
	eventsContent, err := os.ReadFile(eventsPath)
	if err != nil {
		// Events file might not exist
		return nil, nil
	}

	oomEvents := parseOOMEvents(string(eventsContent))
	if len(oomEvents) == 0 {
		return nil, nil
	}

	// Parse pods to get resource specs and correlate with OOM events
	podsPath := filepath.Join(bundleRoot, "rke2/kubectl/pods")
	podsContent, err := os.ReadFile(podsPath)
	if err != nil {
		// Pods file might not exist, return OOM events without resource details
		return oomEvents, nil
	}

	// Enrich OOM events with pod resource information
	return correlateOOMWithResources(oomEvents, string(podsContent)), nil
}

// parseOOMEvents extracts OOM kill events from kubectl events output
func parseOOMEvents(eventsContent string) []OOMAnalysis {
	var analyses []OOMAnalysis

	lines := strings.Split(eventsContent, "\n")

	// Look for OOM kill patterns in events
	oomRegex := regexp.MustCompile(`(?i)(oom|out of memory).*killed?`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		// Events format: NAMESPACE NAME AGE TYPE REASON MESSAGE
		if len(fields) >= 6 {
			eventType := fields[3]
			reason := fields[4]
			message := strings.Join(fields[5:], " ")

			if (eventType == "Warning" && reason == "OOMKilling") ||
				oomRegex.MatchString(message) {

				// Extract pod name from message or event name
				podName := extractPodNameFromOOMMessage(message, fields[1])

				analysis := OOMAnalysis{
					PodName:     podName,
					OOMKillTime: fields[2], // AGE field
				}

				// Determine if node OOM vs container OOM
				analysis.IsNodeOOM = strings.Contains(strings.ToLower(message), "node") ||
					strings.Contains(strings.ToLower(message), "system")

				analyses = append(analyses, analysis)
			}
		}
	}

	return analyses
}

// correlateOOMWithResources enriches OOM events with resource specs from pods
func correlateOOMWithResources(oomEvents []OOMAnalysis, podsContent string) []OOMAnalysis {
	// Parse pods to get resource information
	podResources := parsePodResourceMap(podsContent)

	// Enrich OOM events
	for i := range oomEvents {
		if resources, exists := podResources[oomEvents[i].PodName]; exists {
			// For simplicity, use the first container's resources
			// In a real implementation, we'd need to identify which container OOM'd
			if len(resources) > 0 {
				oomEvents[i].ContainerName = resources[0].ContainerName
				oomEvents[i].MemoryLimit = resources[0].MemoryLimit
				oomEvents[i].MemoryRequest = resources[0].MemoryRequest
			}
		}
	}

	return oomEvents
}

// parsePodResourceMap creates a map of pod name to resource specs
func parsePodResourceMap(podsContent string) map[string][]ResourceSpec {
	resourceMap := make(map[string][]ResourceSpec)

	lines := strings.Split(podsContent, "\n")

	// Skip header line
	if len(lines) > 0 {
		lines = lines[1:]
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		podName := fields[0]
		namespace := fields[1]

		// Create a basic resource spec (in real implementation, parse actual specs)
		// This is simplified - actual implementation would parse pod YAML
		resourceSpec := ResourceSpec{
			PodName:       fmt.Sprintf("%s/%s", namespace, podName),
			ContainerName: podName,     // Simplified
			QoSClass:      "Burstable", // Default assumption
		}

		resourceMap[resourceSpec.PodName] = append(resourceMap[resourceSpec.PodName], resourceSpec)
	}

	return resourceMap
}

// extractPodNameFromOOMMessage tries to extract pod name from OOM message
func extractPodNameFromOOMMessage(message, eventName string) string {
	// Try to extract pod name from message
	podRegex := regexp.MustCompile(`pod\s+([^\s,]+)`)
	if matches := podRegex.FindStringSubmatch(message); len(matches) > 1 {
		return matches[1]
	}

	// Fallback to event name
	return eventName
}
