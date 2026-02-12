package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// OOMAnalysis represents an out-of-memory event analysis
type OOMAnalysis struct {
	PodName       string
	ContainerName string
	MemoryLimit   string // "1Gi"
	MemoryRequest string // "512Mi"
	OOMKillTime   string
	IsNodeOOM     bool   // vs container OOM
	QoSClass      string // "Guaranteed", "Burstable", "BestEffort" (S3-MEDIUM-1)
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
// S3-MEDIUM-2: Parse pod manifests to extract QoS class
func enrichWithQoSClass(oomEvents []OOMAnalysis, bundleRoot string) []OOMAnalysis {
	manifestsPath := filepath.Join(bundleRoot, "rke2/pod-manifests")
	if _, err := os.Stat(manifestsPath); os.IsNotExist(err) {
		return oomEvents
	}

	// Build map of pod name to QoS class
	qosMap := buildQoSMapFromManifests(manifestsPath)

	// Enrich OOM events
	for i := range oomEvents {
		if qosClass, exists := qosMap[oomEvents[i].PodName]; exists {
			oomEvents[i].QoSClass = qosClass
		}
	}

	return oomEvents
}

// buildQoSMapFromManifests scans all pod manifest YAMLs and builds a map of pod name to QoS class
func buildQoSMapFromManifests(manifestsPath string) map[string]string {
	qosMap := make(map[string]string)

	files, err := os.ReadDir(manifestsPath)
	if err != nil {
		return qosMap
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(manifestsPath, file.Name()))
		if err != nil {
			continue
		}

		podName, qosClass := parsePodYAMLForQoS(string(content))
		if podName != "" {
			qosMap[podName] = qosClass
		}
	}

	return qosMap
}

// parsePodYAMLForQoS parses a pod YAML and extracts pod name and calculates QoS class
func parsePodYAMLForQoS(yamlContent string) (string, string) {
	// Parse YAML into generic map structure
	var pod map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &pod); err != nil {
		return "", ""
	}

	// Check if this is a Pod kind
	if kind, ok := pod["kind"].(string); !ok || kind != "Pod" {
		return "", ""
	}

	// Extract pod name from metadata
	metadata, ok := pod["metadata"].(map[string]interface{})
	if !ok {
		return "", ""
	}

	podName, ok := metadata["name"].(string)
	if !ok {
		return "", ""
	}

	// Extract spec
	spec, ok := pod["spec"].(map[string]interface{})
	if !ok {
		return podName, "BestEffort"
	}

	// Get containers
	containers, ok := spec["containers"].([]interface{})
	if !ok || len(containers) == 0 {
		return podName, "BestEffort"
	}

	// Calculate QoS class for all containers
	return podName, calculatePodQoSClass(containers)
}

// calculatePodQoSClass determines QoS class based on container resources
// Kubernetes QoS rules:
// - Guaranteed: Every container has memory and CPU limits set, and limits == requests
// - Burstable: At least one container has a request/limit set, but not Guaranteed
// - BestEffort: No requests or limits set for any container
func calculatePodQoSClass(containers []interface{}) string {
	allHaveLimits := true
	allLimitsEqualRequests := true
	hasAnyResources := false

	for _, c := range containers {
		container, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		resources, ok := container["resources"].(map[string]interface{})
		if !ok {
			// No resources for this container
			allHaveLimits = false
			continue
		}

		requests, hasRequests := resources["requests"].(map[string]interface{})
		limits, hasLimits := resources["limits"].(map[string]interface{})

		memRequest := ""
		cpuRequest := ""
		memLimit := ""
		cpuLimit := ""

		if hasRequests {
			hasAnyResources = true
			if m, ok := requests["memory"].(string); ok {
				memRequest = m
			}
			if c, ok := requests["cpu"].(string); ok {
				cpuRequest = c
			}
		}

		if hasLimits {
			hasAnyResources = true
			if m, ok := limits["memory"].(string); ok {
				memLimit = m
			}
			if c, ok := limits["cpu"].(string); ok {
				cpuLimit = c
			}
		}

		// Check if this container has limits
		if memLimit == "" || cpuLimit == "" {
			allHaveLimits = false
		}

		// Check if limits equal requests
		if memRequest != memLimit || cpuRequest != cpuLimit {
			allLimitsEqualRequests = false
		}
	}

	// Determine QoS class
	if allHaveLimits && allLimitsEqualRequests {
		return "Guaranteed"
	}

	if hasAnyResources {
		return "Burstable"
	}

	return "BestEffort"
}

// enrichWithNodeMemory attempts to correlate OOM events with node memory pressure
// Falls back gracefully if node data is not available
func enrichWithNodeMemory(oomEvents []OOMAnalysis, bundleRoot string) []OOMAnalysis {
	// TODO: Parse nodesDescribePath to analyze node memory pressure during OOM events
	nodesDescribePath := filepath.Join(bundleRoot, "rke2/kubectl/nodesdescribe")
	if _, err := os.Stat(nodesDescribePath); os.IsNotExist(err) {
		return oomEvents
	}

	return oomEvents
}
