package datasource

import (
	"fmt"
	"time"

	"github.com/Rancheroo/r8s/internal/bundle"
	"github.com/Rancheroo/r8s/internal/rancher"
)

// NewEmbeddedDataSource creates a data source from synthetic demo data
// Uses GenerateSyntheticDemo() to create minimal valid bundle in-memory
// This replaces the old file-based example-log-bundle approach (removed to reduce repo size)
func NewEmbeddedDataSource(verbose bool) (DataSource, error) {
	if verbose {
		fmt.Println("🎯 Generating synthetic demo bundle...")
	}

	// Generate synthetic bundle in-memory
	b := GenerateSyntheticDemo()

	if verbose {
		fmt.Printf("✅ Demo bundle generated: %d pods, %d events\n",
			len(b.Pods), len(b.Events))
	}

	return &syntheticDataSource{bundle: b}, nil
}

// syntheticDataSource implements the DataSource interface using synthetic data
type syntheticDataSource struct {
	bundle *bundle.Bundle
}

// Ensure syntheticDataSource implements DataSource interface
var _ DataSource = (*syntheticDataSource)(nil)

// GetClusters returns a single demo cluster
func (ds *syntheticDataSource) GetClusters() ([]rancher.Cluster, error) {
	return []rancher.Cluster{
		{
			ID:       "demo-cluster-id",
			Type:     "cluster",
			Name:     "demo-cluster",
			State:    "active",
			Provider: "demo",
			Created:  time.Now().Add(-30 * 24 * time.Hour),
			Labels:   map[string]string{},
			Links:    map[string]string{},
			Actions:  map[string]string{},
		},
	}, nil
}

// GetProjects returns demo projects with namespace counts
func (ds *syntheticDataSource) GetProjects(clusterID string) ([]rancher.Project, map[string]int, error) {
	projects := []rancher.Project{
		{ID: "p-default", Name: "default", ClusterID: clusterID, State: "active", Created: time.Now()},
		{ID: "p-production", Name: "production", ClusterID: clusterID, State: "active", Created: time.Now()},
		{ID: "p-monitoring", Name: "monitoring", ClusterID: clusterID, State: "active", Created: time.Now()},
		{ID: "p-database", Name: "database", ClusterID: clusterID, State: "active", Created: time.Now()},
		{ID: "p-cache", Name: "cache", ClusterID: clusterID, State: "active", Created: time.Now()},
		{ID: "p-batch", Name: "batch", ClusterID: clusterID, State: "active", Created: time.Now()},
		{ID: "p-payments", Name: "payments", ClusterID: clusterID, State: "active", Created: time.Now()},
		{ID: "p-logging", Name: "logging", ClusterID: clusterID, State: "active", Created: time.Now()},
	}

	// Namespace counts per project
	counts := map[string]int{
		"p-default":    1,
		"p-production": 1,
		"p-monitoring": 1,
		"p-database":   1,
		"p-cache":      1,
		"p-batch":      1,
		"p-payments":   1,
		"p-logging":    1,
	}

	return projects, counts, nil
}

// GetNamespaces returns namespaces for the given cluster and project
func (ds *syntheticDataSource) GetNamespaces(clusterID, projectID string) ([]rancher.Namespace, error) {
	nsMap := map[string][]rancher.Namespace{
		"p-default":    {{ID: "ns-default", Name: "default", ProjectID: projectID, State: "active", Created: time.Now()}},
		"p-production": {{ID: "ns-production", Name: "production", ProjectID: projectID, State: "active", Created: time.Now()}},
		"p-monitoring": {{ID: "ns-monitoring", Name: "monitoring", ProjectID: projectID, State: "active", Created: time.Now()}},
		"p-database":   {{ID: "ns-database", Name: "database", ProjectID: projectID, State: "active", Created: time.Now()}},
		"p-cache":      {{ID: "ns-cache", Name: "cache", ProjectID: projectID, State: "active", Created: time.Now()}},
		"p-batch":      {{ID: "ns-batch", Name: "batch", ProjectID: projectID, State: "active", Created: time.Now()}},
		"p-payments":   {{ID: "ns-payments", Name: "payments", ProjectID: projectID, State: "active", Created: time.Now()}},
		"p-logging":    {{ID: "ns-logging", Name: "logging", ProjectID: projectID, State: "active", Created: time.Now()}},
	}

	if ns, ok := nsMap[projectID]; ok {
		return ns, nil
	}
	return []rancher.Namespace{}, nil
}

// GetPods returns pods for the given project and namespace
func (ds *syntheticDataSource) GetPods(projectID, namespace string) ([]rancher.Pod, error) {
	return ds.convertBundlePodsToRancherPods(namespace), nil
}

// GetAllPods returns all pods across all namespaces
func (ds *syntheticDataSource) GetAllPods() ([]rancher.Pod, error) {
	return ds.convertBundlePodsToRancherPods(""), nil
}

// convertBundlePodsToRancherPods converts bundle.PodInfo to rancher.Pod
func (ds *syntheticDataSource) convertBundlePodsToRancherPods(filterNS string) []rancher.Pod {
	var pods []rancher.Pod

	statusMap := map[string]string{
		"nginx-deployment-7c4c7b8f5-x2v9p":  "Running",
		"frontend-app-5d9c8b7f4-k3m8n":      "CrashLoopBackOff",
		"backend-api-6f8d9c7b5-p9q2r":       "ImagePullBackOff",
		"redis-cache-0":                     "Running",
		"postgres-db-0":                     "OOMKilled",
		"worker-cronjob-27981234-ab12c":     "Completed",
		"payment-service-9a8b7c6d5-e4f3g":   "Pending",
		"elasticsearch-0":                   "Running",
		"prometheus-server-5c9d4b8f7-x1y2z": "Running",
		"grafana-7d8e9f0a1-b2c3d":           "ContainerCreating",
	}

	readyMap := map[string]string{
		"nginx-deployment-7c4c7b8f5-x2v9p":  "1/1",
		"frontend-app-5d9c8b7f4-k3m8n":      "0/1",
		"backend-api-6f8d9c7b5-p9q2r":       "0/1",
		"redis-cache-0":                     "1/1",
		"postgres-db-0":                     "0/1",
		"worker-cronjob-27981234-ab12c":     "0/1",
		"payment-service-9a8b7c6d5-e4f3g":   "0/1",
		"elasticsearch-0":                   "1/1",
		"prometheus-server-5c9d4b8f7-x1y2z": "1/1",
		"grafana-7d8e9f0a1-b2c3d":           "0/1",
	}

	for _, podInfo := range ds.bundle.Pods {
		if filterNS != "" && podInfo.Namespace != filterNS {
			continue
		}

		pod := rancher.Pod{
			ID:            fmt.Sprintf("%s/%s", podInfo.Namespace, podInfo.Name),
			Type:          "pod",
			Name:          podInfo.Name,
			NamespaceID:   podInfo.Namespace,
			State:         statusMap[podInfo.Name],
			PodIP:         "10.42.0.10",
			RestartCount:  0,
			Created:       time.Now().Add(-3 * 24 * time.Hour),
			Labels:        map[string]string{},
			Links:         map[string]string{},
			Actions:       map[string]string{},
			KubectlReady:  readyMap[podInfo.Name],
			KubectlStatus: statusMap[podInfo.Name],
			KubectlAge:    "3d",
		}
		pods = append(pods, pod)
	}

	return pods
}

// GetDeployments returns empty deployments (not in demo)
func (ds *syntheticDataSource) GetDeployments(projectID, namespace string) ([]rancher.Deployment, error) {
	return []rancher.Deployment{}, nil
}

// GetServices returns services
func (ds *syntheticDataSource) GetServices(projectID, namespace string) ([]rancher.Service, error) {
	services := []rancher.Service{
		{ID: "svc-1", Name: "kubernetes", NamespaceID: "default", State: "active", ClusterIP: "10.43.0.1", Kind: "ClusterIP", Created: time.Now()},
		{ID: "svc-2", Name: "nginx-service", NamespaceID: "default", State: "active", ClusterIP: "10.43.12.34", Kind: "LoadBalancer", Created: time.Now()},
		{ID: "svc-3", Name: "backend-api", NamespaceID: "production", State: "active", ClusterIP: "10.43.56.78", Kind: "ClusterIP", Created: time.Now()},
	}

	if namespace == "" {
		return services, nil
	}

	var filtered []rancher.Service
	for _, svc := range services {
		if svc.NamespaceID == namespace {
			filtered = append(filtered, svc)
		}
	}
	return filtered, nil
}

// GetCRDs returns empty CRDs
func (ds *syntheticDataSource) GetCRDs(clusterID string) ([]rancher.CRD, error) {
	return []rancher.CRD{}, nil
}

// GetCRDInstances returns empty instances
func (ds *syntheticDataSource) GetCRDInstances(clusterID, group, version, plural string) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

// GetLogs returns synthetic logs
func (ds *syntheticDataSource) GetLogs(clusterID, namespace, pod, container string, previous bool) ([]string, error) {
	logs := generateSyntheticLogs(pod, namespace, container)
	return []string{logs}, nil
}

// GetContainers returns container names for a pod
func (ds *syntheticDataSource) GetContainers(namespace, pod string) ([]string, error) {
	for _, podInfo := range ds.bundle.Pods {
		if podInfo.Namespace == namespace && podInfo.Name == pod {
			return podInfo.Containers, nil
		}
	}
	return []string{}, nil
}

// DescribePod returns basic pod info
func (ds *syntheticDataSource) DescribePod(clusterID, namespace, name string) (interface{}, error) {
	pods, _ := ds.GetPods("", namespace)
	for _, pod := range pods {
		if pod.Name == name {
			return pod, nil
		}
	}
	return nil, fmt.Errorf("pod not found: %s/%s", namespace, name)
}

// DescribeDeployment returns empty (not in demo)
func (ds *syntheticDataSource) DescribeDeployment(clusterID, namespace, name string) (interface{}, error) {
	return nil, fmt.Errorf("deployment not found: %s/%s", namespace, name)
}

// DescribeService returns service info
func (ds *syntheticDataSource) DescribeService(clusterID, namespace, name string) (interface{}, error) {
	services, _ := ds.GetServices("", namespace)
	for _, svc := range services {
		if svc.Name == name {
			return svc, nil
		}
	}
	return nil, fmt.Errorf("service not found: %s/%s", namespace, name)
}

// GetNodes returns synthetic nodes
func (ds *syntheticDataSource) GetNodes() ([]Node, error) {
	return []Node{
		{Name: "demo-node-1", Status: "Ready"},
		{Name: "demo-node-2", Status: "Ready"},
	}, nil
}

// GetAllEvents returns all events
func (ds *syntheticDataSource) GetAllEvents() ([]rancher.Event, error) {
	var events []rancher.Event
	for _, e := range ds.bundle.Events {
		if event, ok := e.(rancher.Event); ok {
			events = append(events, event)
		}
	}
	return events, nil
}

// GetEventsByPod returns events filtered by pod
func (ds *syntheticDataSource) GetEventsByPod(namespace, podName string) ([]rancher.Event, error) {
	allEvents, _ := ds.GetAllEvents()
	var filtered []rancher.Event
	for _, event := range allEvents {
		if event.Namespace == namespace && (event.Object == "pod/"+podName || event.PodName == podName) {
			filtered = append(filtered, event)
		}
	}
	return filtered, nil
}

// GetDaemonSets returns empty
func (ds *syntheticDataSource) GetDaemonSets() ([]DaemonSet, error) {
	return []DaemonSet{}, nil
}

// GetEtcdHealth returns nil (not applicable)
func (ds *syntheticDataSource) GetEtcdHealth() (*EtcdHealth, error) {
	return nil, nil
}

// GetEtcdDetails returns nil (not applicable)
func (ds *syntheticDataSource) GetEtcdDetails() (*EtcdDetails, error) {
	return nil, nil
}

// GetNodeConditions returns nil (not applicable)
func (ds *syntheticDataSource) GetNodeConditions() ([]NodeConditions, error) {
	return nil, nil
}

// GetSystemHealth returns nil (not applicable)
func (ds *syntheticDataSource) GetSystemHealth() (*SystemHealth, error) {
	return nil, nil
}

// GetKubeletIssues returns nil (not applicable)
func (ds *syntheticDataSource) GetKubeletIssues() ([]KubeletIssue, error) {
	return nil, nil
}

// GetOOMAnalysis returns nil (not applicable)
func (ds *syntheticDataSource) GetOOMAnalysis() ([]OOMAnalysis, error) {
	return nil, nil
}

// GetPodResources returns nil (not applicable)
func (ds *syntheticDataSource) GetPodResources(podName string) ([]ResourceSpec, error) {
	return nil, nil
}

// GetDiagnosticContext returns nil (not applicable)
func (ds *syntheticDataSource) GetDiagnosticContext(namespace, podName string) (*DiagnosticContext, error) {
	return nil, nil
}

// GetBundleHealth returns bundle health
func (ds *syntheticDataSource) GetBundleHealth() *BundleHealth {
	return &BundleHealth{
		HasEtcd:       true,
		HasNodes:      true,
		HasSystemInfo: true,
		HasEvents:     true,
		HasPods:       true,
	}
}

// Mode returns the data source mode
func (ds *syntheticDataSource) Mode() string {
	return "DEMO"
}

// Close cleans up resources (no-op for synthetic)
func (ds *syntheticDataSource) Close() error {
	return nil
}
