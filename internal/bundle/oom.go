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
// Robust against partial bundles - returns what data is available
func AnalyzeOOMEvents(extractPath string) ([]OOMAnalysis, error) {
	bundleRoot := getBundleRoot(extractPath)

	// Parse events first to find OOM kill messages
	eventsPath := filepath.Join(bundleRoot, "rke2/kubectl/events")
	eventsContent, err := os.ReadFile(eventsPath)
	if err != nil {
		// Events file might not exist - gracefully return empty
		return []OOMAnalysis{}, nil
	}

	oomEvents := parseOOMEvents(string(eventsContent))
	if len(oomEvents) == 0 {
		// No OOM events found - not an error, just no data
		return []OOMAnalysis{}, nil
	}

	// Try to enrich with pod resource specs from multiple sources
	// Source 1: kubectl pods output
	podsPath := filepath.Join(bundleRoot, "rke2/kubectl/pods")
	podsContent, err := os.ReadFile(podsPath)
	if err == nil && len(podsContent) > 0 {
		// Enrich with resource information
		oomEvents = correlateOOMWithResources(oomEvents, string(podsContent))
	}

	// Source 2: Try to get QoS class from pod manifests
	oomEvents = enrichWithQoSClass(oomEvents, bundleRoot)

	// Source 3: Try to correlate with node memory pressure
	oomEvents = enrichWithNodeMemory(oomEvents, bundleRoot)

	return oomEvents, nil
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

// enrichWithQoSClass attempts to add QoS class information from pod manifests
// Falls back gracefully if manifests are not available
func enrichWithQoSClass(oomEvents []OOMAnalysis, bundleRoot string) []OOMAnalysis {
	// Try to get QoS class from pod manifests directory
	manifestsPath := filepath.Join(bundleRoot, "rke2/pod-manifests")

	// Check if directory exists
	if _, err := os.Stat(manifestsPath); os.IsNotExist(err) {
		// Manifests not available - return events as-is
		return oomEvents
	}

	// For now, return events as-is
	// Full implementation would parse YAML manifests to extract QoS class
	// This is deferred to future enhancement
	return oomEvents
}

// enrichWithNodeMemory attempts to correlate OOM events with node memory pressure
// Falls back gracefully if node data is not available
func enrichWithNodeMemory(oomEvents []OOMAnalysis, bundleRoot string) []OOMAnalysis {
	// Try to read node describe data
	nodesDescribePath := filepath.Join(bundleRoot, "rke2/kubectl/nodesdescribe")

	// Check if file exists
	if _, err := os.Stat(nodesDescribePath); os.IsNotExist(err) {
		// Node data not available - return events as-is
		return oomEvents
	}

	// For now, return events as-is
	// Full implementation would parse node conditions and memory allocatable
	// to determine if node was under memory pressure during OOM
	// This is deferred to future enhancement
	return oomEvents
}
