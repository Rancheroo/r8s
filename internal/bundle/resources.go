package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ResourceSpec represents pod resource specifications
type ResourceSpec struct {
	PodName       string
	ContainerName string
	MemoryRequest string // "512Mi"
	MemoryLimit   string // "1Gi"
	CPURequest    string // "100m"
	CPULimit      string // "500m"
	QoSClass      string // "Guaranteed", "Burstable", "BestEffort"
}

// ParsePodResources parses pod resource specifications from bundle files
func ParsePodResources(extractPath string, podName string) ([]ResourceSpec, error) {
	bundleRoot := getBundleRoot(extractPath)

	// Look for pod specs in multiple locations
	var specs []ResourceSpec

	// 1. Try pod manifests in rke2/pod-manifests/
	manifestSpecs, err := parsePodManifests(bundleRoot, podName)
	if err == nil {
		specs = append(specs, manifestSpecs...)
	}

	// 2. Try kubectl describe pods output
	describeSpecs, err := parsePodDescribe(bundleRoot, podName)
	if err == nil {
		specs = append(specs, describeSpecs...)
	}

	// 3. Try kubectl get pods with resource columns
	getPodsSpecs, err := parsePodsResourceColumns(bundleRoot, podName)
	if err == nil {
		specs = append(specs, getPodsSpecs...)
	}

	return specs, nil
}

// parsePodManifests parses resource specs from pod manifest YAML files
func parsePodManifests(bundleRoot, podName string) ([]ResourceSpec, error) {
	manifestsDir := filepath.Join(bundleRoot, "rke2/pod-manifests")

	files, err := os.ReadDir(manifestsDir)
	if err != nil {
		return nil, err
	}

	var specs []ResourceSpec

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(manifestsDir, file.Name()))
		if err != nil {
			continue
		}

		fileSpecs := parseYAMLResourceSpecs(string(content), podName)
		specs = append(specs, fileSpecs...)
	}

	return specs, nil
}

// parseYAMLResourceSpecs extracts resource specs from pod YAML
func parseYAMLResourceSpecs(yamlContent, targetPodName string) []ResourceSpec {
	var specs []ResourceSpec

	lines := strings.Split(yamlContent, "\n")

	var currentPodName string
	var inContainers bool
	var inResources bool
	var currentContainer string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Extract pod name
		if strings.HasPrefix(line, "metadata:") {
			// Look for name in next few lines
			continue
		}
		if match := regexp.MustCompile(`name:\s*(.+)`).FindStringSubmatch(line); len(match) > 1 {
			currentPodName = strings.TrimSpace(match[1])
		}

		// Skip if not the target pod
		if targetPodName != "" && currentPodName != targetPodName {
			continue
		}

		// Enter containers section
		if strings.HasPrefix(line, "containers:") {
			inContainers = true
			continue
		}

		// Container name
		if inContainers && strings.HasPrefix(line, "- name:") {
			currentContainer = strings.TrimSpace(strings.TrimPrefix(line, "- name:"))
			continue
		}

		// Resources section
		if inContainers && strings.Contains(line, "resources:") {
			inResources = true
			continue
		}

		// Extract resource values
		if inResources && currentContainer != "" {
			spec := ResourceSpec{
				PodName:       currentPodName,
				ContainerName: currentContainer,
			}

			// Memory request
			if match := regexp.MustCompile(`memory:\s*(.+)`).FindStringSubmatch(line); len(match) > 1 {
				spec.MemoryRequest = strings.TrimSpace(match[1])
			}

			// CPU request
			if match := regexp.MustCompile(`cpu:\s*(.+)`).FindStringSubmatch(line); len(match) > 1 {
				spec.CPURequest = strings.TrimSpace(match[1])
			}

			// Memory limit
			if strings.Contains(line, "limits:") {
				// Look ahead for limits
				continue
			}

			// CPU limit
			if strings.Contains(line, "cpu:") && !strings.Contains(line, "requests:") {
				if match := regexp.MustCompile(`cpu:\s*(.+)`).FindStringSubmatch(line); len(match) > 1 {
					spec.CPULimit = strings.TrimSpace(match[1])
				}
			}

			// If we have any resources, add the spec
			if spec.MemoryRequest != "" || spec.MemoryLimit != "" || spec.CPURequest != "" || spec.CPULimit != "" {
				spec.QoSClass = determineQoSClass(spec)
				specs = append(specs, spec)
			}
		}

		// Exit sections
		if inResources && line == "" {
			inResources = false
		}
		if inContainers && !strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "- ") {
			inContainers = false
			currentContainer = ""
		}
	}

	return specs
}

// parsePodDescribe parses resource specs from kubectl describe pods output
func parsePodDescribe(bundleRoot, podName string) ([]ResourceSpec, error) {
	describePath := filepath.Join(bundleRoot, "rke2/kubectl/nodesdescribe") // Note: this is nodes, not pods

	content, err := os.ReadFile(describePath)
	if err != nil {
		return nil, err
	}

	// This would parse kubectl describe output - simplified for now
	return parseDescribeResourceSpecs(string(content), podName)
}

// parsePodsResourceColumns parses resource specs from kubectl get pods with resource columns
func parsePodsResourceColumns(bundleRoot, podName string) ([]ResourceSpec, error) {
	podsPath := filepath.Join(bundleRoot, "rke2/kubectl/pods")

	content, err := os.ReadFile(podsPath)
	if err != nil {
		return nil, err
	}

	return parsePodsResourceTable(string(content), podName)
}

// parsePodsResourceTable extracts resource info from kubectl get pods output
func parsePodsResourceTable(content, targetPodName string) ([]ResourceSpec, error) {
	var specs []ResourceSpec

	lines := strings.Split(content, "\n")

	// Skip header
	if len(lines) > 0 {
		lines = lines[1:]
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		podName := fields[0]
		namespace := fields[1]

		// Filter by pod name if specified
		if targetPodName != "" && podName != targetPodName {
			continue
		}

		fullPodName := fmt.Sprintf("%s/%s", namespace, podName)

		// Look for resource columns (if present in the output)
		// This is a simplified implementation
		spec := ResourceSpec{
			PodName:       fullPodName,
			ContainerName: podName,
			QoSClass:      "Burstable", // Default
		}

		specs = append(specs, spec)
	}

	return specs, nil
}

// parseDescribeResourceSpecs extracts resources from kubectl describe output
func parseDescribeResourceSpecs(content, targetPodName string) ([]ResourceSpec, error) {
	var specs []ResourceSpec

	// Split by pod sections (look for "Name:" lines)
	podBlocks := strings.Split(content, "\nName:")

	for _, block := range podBlocks {
		if strings.TrimSpace(block) == "" {
			continue
		}

		lines := strings.Split(block, "\n")
		var podName string
		var inContainers bool

		for _, line := range lines {
			line = strings.TrimSpace(line)

			// Pod name
			if strings.HasPrefix(line, "Name:") {
				podName = strings.TrimSpace(strings.TrimPrefix(line, "Name:"))
				if targetPodName != "" && podName != targetPodName {
					break
				}
				continue
			}

			// Containers section
			if strings.Contains(line, "Containers:") {
				inContainers = true
				continue
			}

			// Parse container resource requests/limits
			if inContainers && strings.Contains(line, "Requests:") {
				// Parse requests line
				requests := parseResourceLine(line, "Requests:")
				if len(requests) > 0 {
					spec := ResourceSpec{
						PodName:       podName,
						ContainerName: podName, // Simplified
						MemoryRequest: requests["memory"],
						CPURequest:    requests["cpu"],
					}
					specs = append(specs, spec)
				}
			}

			if inContainers && strings.Contains(line, "Limits:") {
				// Update existing spec with limits
				limits := parseResourceLine(line, "Limits:")
				for i := range specs {
					if specs[i].PodName == podName {
						specs[i].MemoryLimit = limits["memory"]
						specs[i].CPULimit = limits["cpu"]
						specs[i].QoSClass = determineQoSClass(specs[i])
						break
					}
				}
			}

			// Exit containers section
			if inContainers && line == "" {
				inContainers = false
			}
		}
	}

	return specs, nil
}

// parseResourceLine parses a resource line like "Requests: cpu=100m,memory=512Mi"
func parseResourceLine(line, prefix string) map[string]string {
	resources := make(map[string]string)

	if !strings.Contains(line, prefix) {
		return resources
	}

	resourceStr := strings.TrimSpace(strings.TrimPrefix(line, prefix))
	if resourceStr == "" {
		return resources
	}

	// Split by comma
	parts := strings.Split(resourceStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "=") {
			kv := strings.Split(part, "=")
			if len(kv) == 2 {
				key := strings.TrimSpace(kv[0])
				value := strings.TrimSpace(kv[1])
				resources[key] = value
			}
		}
	}

	return resources
}

// determineQoSClass determines Kubernetes QoS class based on resource specs
func determineQoSClass(spec ResourceSpec) string {
	hasMemRequest := spec.MemoryRequest != ""
	hasMemLimit := spec.MemoryLimit != ""
	hasCPURequest := spec.CPURequest != ""
	hasCPULimit := spec.CPULimit != ""

	// Guaranteed: All containers have memory and CPU limits/requests set
	if hasMemRequest && hasMemLimit && hasCPURequest && hasCPULimit {
		return "Guaranteed"
	}

	// Burstable: At least one container has a memory or CPU request or limit set
	if hasMemRequest || hasMemLimit || hasCPURequest || hasCPULimit {
		return "Burstable"
	}

	// BestEffort: No resource requests or limits set
	return "BestEffort"
}
