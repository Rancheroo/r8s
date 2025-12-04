package datasource

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Rancheroo/r8s/internal/bundle"
	"github.com/Rancheroo/r8s/internal/rancher"
)

// BundleDataSource uses bundle files for offline data
type BundleDataSource struct {
	bundle *bundle.Bundle
}

// NewBundleDataSource creates a new bundle data source
func NewBundleDataSource(bundlePath string, verbose bool) (*BundleDataSource, error) {
	opts := bundle.ImportOptions{
		Path:    bundlePath,
		MaxSize: 100 * 1024 * 1024, // 100MB for TUI mode
		Verbose: verbose,
	}

	b, err := bundle.Load(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to load bundle: %w", err)
	}

	return &BundleDataSource{bundle: b}, nil
}

// GetClusters returns a single cluster from bundle metadata
func (ds *BundleDataSource) GetClusters() ([]rancher.Cluster, error) {
	// Bundle represents a single cluster snapshot
	clusterName := "bundle-cluster"
	if ds.bundle.Manifest != nil && ds.bundle.Manifest.NodeName != "" {
		clusterName = ds.bundle.Manifest.NodeName
	}

	cluster := rancher.Cluster{
		ID:       "bundle-cluster",
		Name:     clusterName,
		State:    "active",
		Provider: "bundle",
	}

	return []rancher.Cluster{cluster}, nil
}

// GetProjects returns projects from the bundle with namespace counts
func (ds *BundleDataSource) GetProjects(clusterID string) ([]rancher.Project, map[string]int, error) {
	// Get unique projects from namespaces
	projectMap := make(map[string]*rancher.Project)
	namespaceCounts := make(map[string]int)

	for _, item := range ds.bundle.Namespaces {
		if ns, ok := item.(rancher.Namespace); ok {
			projectID := ns.ProjectID
			if projectID == "" {
				projectID = "default"
			}

			// Count namespace
			namespaceCounts[projectID]++

			// Create project if not exists
			if _, exists := projectMap[projectID]; !exists {
				projectMap[projectID] = &rancher.Project{
					ID:        projectID,
					Name:      projectID,
					ClusterID: clusterID,
					State:     "active",
				}
			}
		}
	}

	// Convert map to slice
	var projects []rancher.Project
	for _, project := range projectMap {
		projects = append(projects, *project)
	}

	// If no projects found, create a default one
	if len(projects) == 0 {
		projects = []rancher.Project{
			{
				ID:        "default",
				Name:      "default",
				ClusterID: clusterID,
				State:     "active",
			},
		}
		namespaceCounts["default"] = len(ds.bundle.Namespaces)
	}

	return projects, namespaceCounts, nil
}

// GetNamespaces returns namespaces from the bundle
func (ds *BundleDataSource) GetNamespaces(clusterID, projectID string) ([]rancher.Namespace, error) {
	var namespaces []rancher.Namespace
	for _, item := range ds.bundle.Namespaces {
		if namespace, ok := item.(rancher.Namespace); ok {
			// Filter by project if specified
			if projectID != "" && namespace.ProjectID != projectID && namespace.ProjectID != "" {
				continue
			}
			namespaces = append(namespaces, namespace)
		}
	}
	return namespaces, nil
}

// GetPods returns pods from the bundle with enriched kubectl data
func (ds *BundleDataSource) GetPods(projectID, namespace string) ([]rancher.Pod, error) {
	var pods []rancher.Pod

	// Build event map for quick lookup: namespace/podname -> []event messages
	eventMap := make(map[string][]string)
	for _, item := range ds.bundle.Events {
		if event, ok := item.(rancher.Event); ok {
			if event.ObjectKind == "pod" && event.PodName != "" {
				key := event.Namespace + "/" + event.PodName
				msg := fmt.Sprintf("[%s] %s: %s (count: %d)", event.Type, event.Reason, event.Message, event.Count)
				eventMap[key] = append(eventMap[key], msg)
			}
		}
	}

	// Parse kubectl pods directly for enriched data
	kubectlPods, err := bundle.ParsePods(ds.bundle.ExtractPath)
	kubectlPodsFound := false
	if err == nil && len(kubectlPods) > 0 {
		kubectlPodsFound = true
		for _, pod := range kubectlPods {
			// Filter by namespace if specified
			if namespace != "" && pod.NamespaceID != namespace {
				continue
			}

			// Attach events to this pod
			key := pod.NamespaceID + "/" + pod.Name
			if events, ok := eventMap[key]; ok {
				pod.KubectlEvents = events
			}

			pods = append(pods, pod)
		}
	}

	// Fallback to basic PodInfo if kubectl parsing failed
	if !kubectlPodsFound {
		for _, podInfo := range ds.bundle.Pods {
			// Filter by namespace if specified
			if namespace != "" && podInfo.Namespace != namespace {
				continue
			}

			// Convert bundle.PodInfo to rancher.Pod
			pod := rancher.Pod{
				Name:        podInfo.Name,
				NamespaceID: podInfo.Namespace,
				State:       "Bundle",
				NodeName:    "bundle",
			}

			// Attach events
			key := pod.NamespaceID + "/" + pod.Name
			if events, ok := eventMap[key]; ok {
				pod.KubectlEvents = events
			}

			pods = append(pods, pod)
		}
	}

	return pods, nil
}

// GetDeployments returns deployments from the bundle
func (ds *BundleDataSource) GetDeployments(projectID, namespace string) ([]rancher.Deployment, error) {
	var deployments []rancher.Deployment
	for _, item := range ds.bundle.Deployments {
		if deployment, ok := item.(rancher.Deployment); ok {
			// Filter by namespace if specified
			if namespace == "" || deployment.NamespaceID == namespace {
				deployments = append(deployments, deployment)
			}
		}
	}
	return deployments, nil
}

// GetServices returns services from the bundle
func (ds *BundleDataSource) GetServices(projectID, namespace string) ([]rancher.Service, error) {
	var services []rancher.Service
	for _, item := range ds.bundle.Services {
		if service, ok := item.(rancher.Service); ok {
			// Filter by namespace if specified
			if namespace == "" || service.NamespaceID == namespace {
				services = append(services, service)
			}
		}
	}
	return services, nil
}

// GetCRDs returns CRDs from the bundle
func (ds *BundleDataSource) GetCRDs(clusterID string) ([]rancher.CRD, error) {
	var crds []rancher.CRD
	for _, item := range ds.bundle.CRDs {
		if crd, ok := item.(rancher.CRD); ok {
			crds = append(crds, crd)
		}
	}
	return crds, nil
}

// GetCRDInstances returns CRD instances from the bundle
func (ds *BundleDataSource) GetCRDInstances(clusterID, group, version, plural string) ([]map[string]interface{}, error) {
	// Bundle mode doesn't have CRD instances in the current implementation
	// Return empty list rather than error
	return []map[string]interface{}{}, nil
}

// GetLogs returns logs from bundle files
func (ds *BundleDataSource) GetLogs(clusterID, namespace, pod, container string, previous bool) ([]string, error) {
	// Find log file for this pod/container
	for _, logFile := range ds.bundle.LogFiles {
		if logFile.Namespace == namespace &&
			logFile.PodName == pod &&
			logFile.ContainerName == container &&
			logFile.IsPrevious == previous {

			content, err := ds.bundle.ReadLogFile(&logFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read log file: %w", err)
			}

			// Split into lines
			lines := strings.Split(string(content), "\n")

			// Remove empty last line if present
			if len(lines) > 0 && lines[len(lines)-1] == "" {
				lines = lines[:len(lines)-1]
			}

			return lines, nil
		}
	}

	return nil, fmt.Errorf("log file not found for pod %s/%s container %s", namespace, pod, container)
}

// GetContainers returns containers from bundle pod info
func (ds *BundleDataSource) GetContainers(namespace, pod string) ([]string, error) {
	for _, podInfo := range ds.bundle.Pods {
		if podInfo.Namespace == namespace && podInfo.Name == pod {
			if len(podInfo.Containers) > 0 {
				return podInfo.Containers, nil
			}
			return []string{"unknown"}, nil
		}
	}
	return []string{"unknown"}, nil
}

// DescribePod returns detailed pod information from bundle
func (ds *BundleDataSource) DescribePod(clusterID, namespace, name string) (interface{}, error) {
	// Get the pod from bundle (has enriched fields)
	pods, err := ds.GetPods("", namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	// Find the specific pod
	for i := range pods {
		if pods[i].Name == name {
			// Return the pod as JSON-marshalable data
			return pods[i], nil
		}
	}

	return nil, fmt.Errorf("pod not found: %s/%s", namespace, name)
}

// DescribeDeployment returns detailed deployment information from bundle
func (ds *BundleDataSource) DescribeDeployment(clusterID, namespace, name string) (interface{}, error) {
	deployments, err := ds.GetDeployments("", namespace)
	if err != nil {
		return nil, err
	}

	for i := range deployments {
		if deployments[i].Name == name && deployments[i].NamespaceID == namespace {
			return deployments[i], nil
		}
	}

	// Return a mock structure if not found in bundle
	return map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"note": "Bundle data - limited details available",
	}, nil
}

// DescribeService returns detailed service information from bundle
func (ds *BundleDataSource) DescribeService(clusterID, namespace, name string) (interface{}, error) {
	services, err := ds.GetServices("", namespace)
	if err != nil {
		return nil, err
	}

	for i := range services {
		if services[i].Name == name && services[i].NamespaceID == namespace {
			return services[i], nil
		}
	}

	// Return a mock structure if not found in bundle
	return map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"note": "Bundle data - limited details available",
	}, nil
}

// Mode returns the display string for bundle mode
func (ds *BundleDataSource) Mode() string {
	return "BUNDLE"
}

// GetAllPods returns all pods across all namespaces
func (ds *BundleDataSource) GetAllPods() ([]rancher.Pod, error) {
	// Use kubectl parser which has all pods
	pods, err := bundle.ParsePods(ds.bundle.ExtractPath)
	if err != nil {
		return nil, err
	}
	return pods, nil
}

// GetNodes returns cluster nodes
func (ds *BundleDataSource) GetNodes() ([]Node, error) {
	nodeInfos, err := bundle.ParseNodes(ds.bundle.ExtractPath)
	if err != nil {
		// Nodes file might not exist in all bundles
		return []Node{}, nil
	}

	var nodes []Node
	for _, ni := range nodeInfos {
		nodes = append(nodes, Node{
			Name:   ni.Name,
			Status: ni.Status,
		})
	}
	return nodes, nil
}

// GetAllEvents returns all cluster events
func (ds *BundleDataSource) GetAllEvents() ([]rancher.Event, error) {
	// Events are already parsed and stored in bundle
	var events []rancher.Event
	for _, item := range ds.bundle.Events {
		if event, ok := item.(rancher.Event); ok {
			events = append(events, event)
		}
	}
	return events, nil
}

// GetDaemonSets returns all DaemonSets
func (ds *BundleDataSource) GetDaemonSets() ([]DaemonSet, error) {
	dsInfos, err := bundle.ParseDaemonSets(ds.bundle.ExtractPath)
	if err != nil {
		// DaemonSets file might not exist
		return []DaemonSet{}, nil
	}

	var daemonsets []DaemonSet
	for _, dsi := range dsInfos {
		daemonsets = append(daemonsets, DaemonSet{
			Name:      dsi.Name,
			Namespace: dsi.Namespace,
			Ready:     dsi.Ready,
		})
	}
	return daemonsets, nil
}

// GetEtcdHealth returns etcd health info (bundle only)
func (ds *BundleDataSource) GetEtcdHealth() (*EtcdHealth, error) {
	healthInfo, err := bundle.ParseEtcdHealth(ds.bundle.ExtractPath)
	if err != nil {
		// etcd dir might not exist
		return nil, nil
	}

	return &EtcdHealth{
		Healthy:    healthInfo.Healthy,
		HasAlarms:  healthInfo.HasAlarms,
		AlarmType:  healthInfo.AlarmType,
		AlarmCount: healthInfo.AlarmCount,
	}, nil
}

// GetSystemHealth returns system health info (bundle only)
func (ds *BundleDataSource) GetSystemHealth() (*SystemHealth, error) {
	healthInfo, err := bundle.ParseSystemHealth(ds.bundle.ExtractPath)
	if err != nil {
		// systeminfo dir might not exist
		return nil, nil
	}

	return &SystemHealth{
		MemoryUsedPercent: healthInfo.MemoryUsedPercent,
		DiskUsedPercent:   healthInfo.DiskUsedPercent,
	}, nil
}

// Close cleans up bundle resources
func (ds *BundleDataSource) Close() error {
	if ds.bundle != nil {
		return ds.bundle.Close()
	}
	return nil
}

// Helper function to pretty-print JSON for describe views
func prettifyJSON(v interface{}) string {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%+v", v)
	}
	return string(jsonBytes)
}
