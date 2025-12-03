package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Rancheroo/r8s/internal/rancher"
)

// ParseCRDs parses kubectl get crds output from bundle
func ParseCRDs(extractPath string) ([]rancher.CRD, error) {
	// FIX BUG-003: Use getBundleRoot() to handle wrapper directories
	bundleRoot := getBundleRoot(extractPath)
	path := filepath.Join(bundleRoot, "rke2/kubectl/crds")
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var crds []rancher.CRD

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header and empty lines
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		name := fields[0]
		createdAt := fields[1]

		// Parse CRD name into group/kind
		// Format: <plural>.<group>
		// Example: addons.k3s.cattle.io -> plural=addons, group=k3s.cattle.io
		parts := strings.Split(name, ".")
		if len(parts) < 2 {
			continue
		}

		plural := parts[0]
		group := strings.Join(parts[1:], ".")

		// Generate kind by capitalizing plural (simple heuristic)
		kind := strings.Title(plural)
		// Remove trailing 's' for kind if present
		if strings.HasSuffix(kind, "s") && len(kind) > 1 {
			kind = kind[:len(kind)-1]
		}

		// Parse timestamp
		created, _ := time.Parse(time.RFC3339, createdAt)

		crds = append(crds, rancher.CRD{
			Metadata: rancher.ObjectMeta{
				Name:              name,
				CreationTimestamp: created,
			},
			Spec: rancher.CRDSpec{
				Group: group,
				Names: rancher.CRDNames{
					Kind:     kind,
					Plural:   plural,
					Singular: plural, // Simple fallback
				},
				Scope: "Cluster", // Default assumption
				Versions: []rancher.CRDVersion{
					{Name: "v1", Served: true, Storage: true},
				},
			},
		})
	}

	return crds, nil
}

// ParseDeployments parses kubectl get deployments output from bundle
func ParseDeployments(extractPath string) ([]rancher.Deployment, error) {
	// FIX BUG-003: Use getBundleRoot() to handle wrapper directories
	bundleRoot := getBundleRoot(extractPath)
	path := filepath.Join(bundleRoot, "rke2/kubectl/deployments")
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var deployments []rancher.Deployment

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue // Need at least namespace, name, ready, uptodate, available, age
		}

		namespace := fields[0]
		name := fields[1]
		ready := fields[2] // Format: "1/1"

		// Parse ready field "1/1"
		readyParts := strings.Split(ready, "/")
		var readyReplicas, totalReplicas int
		if len(readyParts) == 2 {
			fmt.Sscanf(readyParts[0], "%d", &readyReplicas)
			fmt.Sscanf(readyParts[1], "%d", &totalReplicas)
		}

		deployments = append(deployments, rancher.Deployment{
			Name:              name,
			NamespaceID:       namespace,
			State:             "active",
			Replicas:          totalReplicas,
			ReadyReplicas:     readyReplicas,
			AvailableReplicas: readyReplicas,
			UpToDateReplicas:  readyReplicas,
			Created:           time.Now(), // Not in kubectl output
		})
	}

	return deployments, nil
}

// ParseServices parses kubectl get services output from bundle
func ParseServices(extractPath string) ([]rancher.Service, error) {
	// FIX BUG-003: Use getBundleRoot() to handle wrapper directories
	bundleRoot := getBundleRoot(extractPath)
	path := filepath.Join(bundleRoot, "rke2/kubectl/services")
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var services []rancher.Service

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		namespace := fields[0]
		name := fields[1]
		serviceType := fields[2]
		clusterIP := fields[3]
		// externalIP := fields[4]
		portsStr := fields[5]

		// Parse ports: "5473/TCP" or "9093/TCP,9094/TCP,9094/UDP"
		var ports []rancher.ServicePort
		for _, portStr := range strings.Split(portsStr, ",") {
			parts := strings.Split(portStr, "/")
			if len(parts) == 2 {
				var port int
				fmt.Sscanf(parts[0], "%d", &port)
				protocol := parts[1]

				ports = append(ports, rancher.ServicePort{
					Protocol:   protocol,
					Port:       port,
					TargetPort: port,
				})
			}
		}

		services = append(services, rancher.Service{
			Name:        name,
			NamespaceID: namespace,
			State:       "active",
			ClusterIP:   clusterIP,
			Kind:        serviceType,
			Ports:       ports,
			Created:     time.Now(), // Not in kubectl output
		})
	}

	return services, nil
}

// ParseNamespaces parses kubectl get namespaces output from bundle
func ParseNamespaces(extractPath string) ([]rancher.Namespace, error) {
	// FIX BUG-003: Use getBundleRoot() to handle wrapper directories
	bundleRoot := getBundleRoot(extractPath)
	path := filepath.Join(bundleRoot, "rke2/kubectl/namespaces")
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var namespaces []rancher.Namespace

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		name := fields[0]
		status := fields[1]
		// age := fields[2] if needed

		namespaces = append(namespaces, rancher.Namespace{
			Name:      name,
			State:     strings.ToLower(status),
			ClusterID: "bundle",
			ProjectID: "bundle-project",
			Created:   time.Now(), // Not in kubectl output
		})
	}

	return namespaces, nil
}

// ParsePods parses kubectl get pods output from bundle
// Format: NAMESPACE NAME READY STATUS RESTARTS AGE IP NODE NOMINATED_NODE READINESS_GATES
func ParsePods(extractPath string) ([]rancher.Pod, error) {
	bundleRoot := getBundleRoot(extractPath)
	path := filepath.Join(bundleRoot, "rke2/kubectl/pods")
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var pods []rancher.Pod

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header and empty lines
		}

		fields := strings.Fields(line)
		if len(fields) < 8 {
			continue // Need at least namespace, name, ready, status, restarts, age, ip, node
		}

		namespace := fields[0]
		name := fields[1]
		ready := fields[2]  // e.g., "1/1", "2/2"
		status := fields[3] // Running, Completed, etc.
		restartsStr := fields[4]
		age := fields[5]
		ip := fields[6]
		node := fields[7]

		// Parse readiness gates if present (fields 8 and 9)
		readinessGates := "<none>"
		if len(fields) >= 10 {
			readinessGates = fields[9]
		}

		// Parse restart count
		var restarts int
		fmt.Sscanf(restartsStr, "%d", &restarts)

		pods = append(pods, rancher.Pod{
			Name:                  name,
			NamespaceID:           namespace,
			NodeName:              node,
			State:                 status,
			PodIP:                 ip,
			RestartCount:          restarts,
			Created:               time.Now(), // Not available in kubectl output
			KubectlReady:          ready,
			KubectlStatus:         status,
			KubectlAge:            age,
			KubectlIP:             ip,
			KubectlReadinessGates: readinessGates,
			KubectlRestarts:       restarts,
		})
	}

	return pods, nil
}

// ParseEvents parses kubectl get events output from bundle
// Format: NAMESPACE LAST_SEEN TYPE REASON OBJECT SUBOBJECT SOURCE MESSAGE FIRST_SEEN COUNT NAME
func ParseEvents(extractPath string) ([]rancher.Event, error) {
	bundleRoot := getBundleRoot(extractPath)
	path := filepath.Join(bundleRoot, "rke2/kubectl/events")
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var events []rancher.Event

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header and empty lines
		}

		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue // Need minimum fields
		}

		namespace := fields[0]
		lastSeen := fields[1]
		eventType := fields[2]
		reason := fields[3]
		object := fields[4]
		// subobject := fields[5] (often empty)
		source := fields[6]

		// Message spans multiple fields, find FIRST_SEEN marker
		messageStart := 7
		messageEnd := len(fields) - 3 // Last 3 are FIRST_SEEN, COUNT, NAME

		message := ""
		if messageEnd > messageStart {
			message = strings.Join(fields[messageStart:messageEnd], " ")
		}

		firstSeen := fields[len(fields)-3]
		countStr := fields[len(fields)-2]
		name := fields[len(fields)-1]

		var count int
		fmt.Sscanf(countStr, "%d", &count)

		// Extract pod name from object field (format: "pod/pod-name")
		podName := ""
		objectKind := ""
		if strings.Contains(object, "/") {
			parts := strings.SplitN(object, "/", 2)
			if len(parts) == 2 {
				objectKind = parts[0]
				podName = parts[1]
			}
		}

		events = append(events, rancher.Event{
			Namespace:  namespace,
			Type:       eventType,
			Reason:     reason,
			Object:     object,
			Message:    message,
			Source:     source,
			FirstSeen:  firstSeen,
			LastSeen:   lastSeen,
			Count:      count,
			Name:       name,
			PodName:    podName,
			ObjectKind: objectKind,
		})
	}

	return events, nil
}
