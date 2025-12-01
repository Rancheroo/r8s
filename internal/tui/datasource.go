package tui

import (
	"fmt"
	"strings"

	"github.com/Rancheroo/r8s/internal/bundle"
	"github.com/Rancheroo/r8s/internal/rancher"
)

// DataSource abstracts pod and log data retrieval
// This allows the TUI to work with both live Rancher API and offline bundles
type DataSource interface {
	// GetClusters returns all available clusters
	GetClusters() ([]rancher.Cluster, error)

	// GetProjects returns projects for the given cluster with namespace counts
	GetProjects(clusterID string) ([]rancher.Project, map[string]int, error)

	// GetPods returns pods for the given project and namespace
	GetPods(projectID, namespace string) ([]rancher.Pod, error)

	// GetLogs returns log lines for the specified pod and container
	GetLogs(clusterID, namespace, pod, container string, previous bool) ([]string, error)

	// GetContainers returns available containers for a pod
	GetContainers(namespace, pod string) ([]string, error)

	// GetCRDs returns CRDs for the given cluster
	GetCRDs(clusterID string) ([]rancher.CRD, error)

	// GetDeployments returns deployments for the given project and namespace
	GetDeployments(projectID, namespace string) ([]rancher.Deployment, error)

	// GetServices returns services for the given project and namespace
	GetServices(projectID, namespace string) ([]rancher.Service, error)

	// GetNamespaces returns namespaces for the given cluster and project
	GetNamespaces(clusterID, projectID string) ([]rancher.Namespace, error)

	// IsOffline returns true if this is an offline data source
	IsOffline() bool

	// GetMode returns a display string for the current mode
	GetMode() string
}

// LiveDataSource uses the Rancher API for live data
type LiveDataSource struct {
	client      *rancher.Client
	offlineMode bool
}

// NewLiveDataSource creates a new live data source
func NewLiveDataSource(client *rancher.Client, offline bool) *LiveDataSource {
	return &LiveDataSource{
		client:      client,
		offlineMode: offline,
	}
}

// GetClusters fetches clusters from the Rancher API
func (ds *LiveDataSource) GetClusters() ([]rancher.Cluster, error) {
	if ds.offlineMode {
		return nil, fmt.Errorf("offline mode")
	}

	collection, err := ds.client.ListClusters()
	if err != nil {
		return nil, err
	}

	return collection.Data, nil
}

// GetProjects fetches projects from the Rancher API
func (ds *LiveDataSource) GetProjects(clusterID string) ([]rancher.Project, map[string]int, error) {
	if ds.offlineMode {
		return nil, nil, fmt.Errorf("offline mode")
	}

	collection, err := ds.client.ListProjects(clusterID)
	if err != nil {
		return nil, nil, err
	}

	// Count namespaces per project
	namespaceCounts := make(map[string]int)
	nsCollection, err := ds.client.ListNamespaces(clusterID)
	if err == nil {
		for _, ns := range nsCollection.Data {
			if ns.ProjectID != "" {
				namespaceCounts[ns.ProjectID]++
			}
		}
	}

	// Ensure all projects have an entry
	for _, project := range collection.Data {
		if _, exists := namespaceCounts[project.ID]; !exists {
			namespaceCounts[project.ID] = 0
		}
	}

	return collection.Data, namespaceCounts, nil
}

// GetPods fetches pods from the Rancher API
func (ds *LiveDataSource) GetPods(projectID, namespace string) ([]rancher.Pod, error) {
	// If offline, return empty (caller will use mock data)
	if ds.offlineMode {
		return nil, fmt.Errorf("offline mode")
	}

	collection, err := ds.client.ListPods(projectID)
	if err != nil {
		return nil, err
	}

	// Filter by namespace
	var filtered []rancher.Pod
	for _, pod := range collection.Data {
		podNamespace := pod.NamespaceID
		if strings.Contains(podNamespace, ":") {
			parts := strings.Split(podNamespace, ":")
			if len(parts) > 1 {
				podNamespace = parts[1]
			}
		}

		if podNamespace == namespace {
			filtered = append(filtered, pod)
		}
	}

	return filtered, nil
}

// GetLogs fetches logs from the Rancher API
func (ds *LiveDataSource) GetLogs(clusterID, namespace, pod, container string, previous bool) ([]string, error) {
	if ds.offlineMode {
		return nil, fmt.Errorf("offline mode")
	}

	// Fetch logs from Rancher via K8s proxy
	logContent, err := ds.client.GetPodLogs(clusterID, namespace, pod, container, previous, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch logs: %w", err)
	}

	// Split into lines
	lines := strings.Split(logContent, "\n")

	// Remove empty last line if present
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return lines, nil
}

// GetContainers returns containers for a pod
func (ds *LiveDataSource) GetContainers(namespace, pod string) ([]string, error) {
	if ds.offlineMode {
		return []string{"app"}, nil
	}

	// We need clusterID to fetch pod details - extract from current context
	// For now, we'll need to get this from somewhere... let's return a sensible default
	// TODO: Store clusterID in LiveDataSource for this use case

	// Return common container name as fallback
	return []string{"app"}, nil
}

// GetCRDs fetches CRDs from the Rancher API
func (ds *LiveDataSource) GetCRDs(clusterID string) ([]rancher.CRD, error) {
	if ds.offlineMode {
		return nil, fmt.Errorf("offline mode")
	}

	crdList, err := ds.client.ListCRDs(clusterID)
	if err != nil {
		return nil, err
	}
	return crdList.Items, nil
}

// GetDeployments fetches deployments from the Rancher API
func (ds *LiveDataSource) GetDeployments(projectID, namespace string) ([]rancher.Deployment, error) {
	if ds.offlineMode {
		return nil, fmt.Errorf("offline mode")
	}

	collection, err := ds.client.ListDeployments(projectID)
	if err != nil {
		return nil, err
	}

	// Filter by namespace if specified
	if namespace == "" {
		return collection.Data, nil
	}

	var filtered []rancher.Deployment
	for _, deployment := range collection.Data {
		deploymentNamespace := deployment.NamespaceID
		if strings.Contains(deploymentNamespace, ":") {
			parts := strings.Split(deploymentNamespace, ":")
			if len(parts) > 1 {
				deploymentNamespace = parts[1]
			}
		}

		if deploymentNamespace == namespace {
			filtered = append(filtered, deployment)
		}
	}

	return filtered, nil
}

// GetServices fetches services from the Rancher API
func (ds *LiveDataSource) GetServices(projectID, namespace string) ([]rancher.Service, error) {
	if ds.offlineMode {
		return nil, fmt.Errorf("offline mode")
	}

	collection, err := ds.client.ListServices(projectID)
	if err != nil {
		return nil, err
	}

	// Filter by namespace if specified
	if namespace == "" {
		return collection.Data, nil
	}

	var filtered []rancher.Service
	for _, service := range collection.Data {
		serviceNamespace := service.NamespaceID
		if strings.Contains(serviceNamespace, ":") {
			parts := strings.Split(serviceNamespace, ":")
			if len(parts) > 1 {
				serviceNamespace = parts[1]
			}
		}

		if serviceNamespace == namespace {
			filtered = append(filtered, service)
		}
	}

	return filtered, nil
}

// GetNamespaces fetches namespaces from the Rancher API
func (ds *LiveDataSource) GetNamespaces(clusterID, projectID string) ([]rancher.Namespace, error) {
	if ds.offlineMode {
		return nil, fmt.Errorf("offline mode")
	}

	collection, err := ds.client.ListNamespaces(clusterID)
	if err != nil {
		return nil, err
	}

	// Filter by project if specified
	if projectID == "" {
		return collection.Data, nil
	}

	var filtered []rancher.Namespace
	for _, namespace := range collection.Data {
		if namespace.ProjectID == projectID {
			filtered = append(filtered, namespace)
		}
	}

	return filtered, nil
}

// IsOffline returns true if in offline mode
func (ds *LiveDataSource) IsOffline() bool {
	return ds.offlineMode
}

// GetMode returns the display string for this mode
func (ds *LiveDataSource) GetMode() string {
	if ds.offlineMode {
		return "OFFLINE"
	}
	return "LIVE"
}

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
	// Create a cluster from bundle metadata
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

// GetPods returns pods from the bundle
func (ds *BundleDataSource) GetPods(projectID, namespace string) ([]rancher.Pod, error) {
	var pods []rancher.Pod

	for _, podInfo := range ds.bundle.Pods {
		// Filter by namespace if specified
		if namespace != "" && podInfo.Namespace != namespace {
			continue
		}

		// Convert bundle.PodInfo to rancher.Pod
		pod := rancher.Pod{
			Name:        podInfo.Name,
			NamespaceID: podInfo.Namespace,
			State:       "Bundle", // Special state for bundle pods
			NodeName:    "bundle", // Placeholder
		}

		pods = append(pods, pod)
	}

	return pods, nil
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
			// Fallback to single container if not found
			return []string{"unknown"}, nil
		}
	}
	return []string{"unknown"}, nil
}

// IsOffline returns true (bundle mode is always offline)
func (ds *BundleDataSource) IsOffline() bool {
	return true
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

// GetNamespaces returns namespaces from the bundle
func (ds *BundleDataSource) GetNamespaces(clusterID, projectID string) ([]rancher.Namespace, error) {
	var namespaces []rancher.Namespace
	for _, item := range ds.bundle.Namespaces {
		if namespace, ok := item.(rancher.Namespace); ok {
			namespaces = append(namespaces, namespace)
		}
	}
	return namespaces, nil
}

// GetMode returns the display string for bundle mode
func (ds *BundleDataSource) GetMode() string {
	return "BUNDLE"
}

// Close cleans up bundle resources
func (ds *BundleDataSource) Close() error {
	if ds.bundle != nil {
		return ds.bundle.Close()
	}
	return nil
}
