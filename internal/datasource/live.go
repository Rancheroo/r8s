package datasource

import (
	"fmt"
	"strings"

	"github.com/Rancheroo/r8s/internal/rancher"
)

// LiveDataSource uses the Rancher API for live cluster data
type LiveDataSource struct {
	client *rancher.Client
}

// NewLiveDataSource creates a new live data source backed by Rancher API
func NewLiveDataSource(client *rancher.Client) *LiveDataSource {
	return &LiveDataSource{
		client: client,
	}
}

// GetClusters fetches clusters from the Rancher API
func (ds *LiveDataSource) GetClusters() ([]rancher.Cluster, error) {
	collection, err := ds.client.ListClusters()
	if err != nil {
		return nil, err
	}
	return collection.Data, nil
}

// GetProjects fetches projects from the Rancher API
func (ds *LiveDataSource) GetProjects(clusterID string) ([]rancher.Project, map[string]int, error) {
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

// GetNamespaces fetches namespaces from the Rancher API
func (ds *LiveDataSource) GetNamespaces(clusterID, projectID string) ([]rancher.Namespace, error) {
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

// GetPods fetches pods from the Rancher API
func (ds *LiveDataSource) GetPods(projectID, namespace string) ([]rancher.Pod, error) {
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

// GetDeployments fetches deployments from the Rancher API
func (ds *LiveDataSource) GetDeployments(projectID, namespace string) ([]rancher.Deployment, error) {
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

// GetCRDs fetches CRDs from the Rancher API
func (ds *LiveDataSource) GetCRDs(clusterID string) ([]rancher.CRD, error) {
	crdList, err := ds.client.ListCRDs(clusterID)
	if err != nil {
		return nil, err
	}
	return crdList.Items, nil
}

// GetCRDInstances fetches instances of a CRD from the Rancher API
func (ds *LiveDataSource) GetCRDInstances(clusterID, group, version, plural string) ([]map[string]interface{}, error) {
	instanceList, err := ds.client.ListCustomResources(clusterID, group, version, plural, "")
	if err != nil {
		return nil, err
	}
	return instanceList.Items, nil
}

// GetLogs fetches logs from the Rancher API
func (ds *LiveDataSource) GetLogs(clusterID, namespace, pod, container string, previous bool) ([]string, error) {
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
	// For live mode, we'd need to fetch pod details to get containers
	// For now, return a sensible default - the TUI can be enhanced later
	// to fetch this from pod details if needed
	return []string{"app"}, nil
}

// DescribePod fetches detailed pod information
func (ds *LiveDataSource) DescribePod(clusterID, namespace, name string) (interface{}, error) {
	return ds.client.GetPodDetails(clusterID, namespace, name)
}

// DescribeDeployment fetches detailed deployment information
func (ds *LiveDataSource) DescribeDeployment(clusterID, namespace, name string) (interface{}, error) {
	return ds.client.GetDeploymentDetails(clusterID, namespace, name)
}

// DescribeService fetches detailed service information
func (ds *LiveDataSource) DescribeService(clusterID, namespace, name string) (interface{}, error) {
	return ds.client.GetServiceDetails(clusterID, namespace, name)
}

// Mode returns the display string for live mode
func (ds *LiveDataSource) Mode() string {
	return "LIVE"
}

// Close cleans up resources (no-op for live data source)
func (ds *LiveDataSource) Close() error {
	return nil
}
