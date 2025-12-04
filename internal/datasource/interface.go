// Package datasource provides a unified interface for data retrieval across different modes.
// It eliminates mode-specific logic by abstracting Live API, Bundle files, and Demo data
// behind a single interface. The TUI layer only depends on this interface, making the
// codebase mode-agnostic and maintainable.
package datasource

import (
	"github.com/Rancheroo/r8s/internal/rancher"
)

// DataSource abstracts data retrieval for the TUI across all modes.
// Implementations: LiveDataSource (Rancher API), BundleDataSource (log bundles), EmbeddedDataSource (demo)
type DataSource interface {
	// GetClusters returns all available clusters
	GetClusters() ([]rancher.Cluster, error)

	// GetProjects returns projects for the given cluster with namespace counts
	GetProjects(clusterID string) ([]rancher.Project, map[string]int, error)

	// GetNamespaces returns namespaces for the given cluster and project
	GetNamespaces(clusterID, projectID string) ([]rancher.Namespace, error)

	// GetPods returns pods for the given project and namespace
	GetPods(projectID, namespace string) ([]rancher.Pod, error)

	// GetAllPods returns all pods across all namespaces (for attention dashboard)
	GetAllPods() ([]rancher.Pod, error)

	// GetDeployments returns deployments for the given project and namespace
	GetDeployments(projectID, namespace string) ([]rancher.Deployment, error)

	// GetServices returns services for the given project and namespace
	GetServices(projectID, namespace string) ([]rancher.Service, error)

	// GetCRDs returns CRDs for the given cluster
	GetCRDs(clusterID string) ([]rancher.CRD, error)

	// GetCRDInstances returns instances of a CRD
	GetCRDInstances(clusterID, group, version, plural string) ([]map[string]interface{}, error)

	// GetLogs returns log lines for the specified pod and container
	GetLogs(clusterID, namespace, pod, container string, previous bool) ([]string, error)

	// GetContainers returns available containers for a pod
	GetContainers(namespace, pod string) ([]string, error)

	// DescribePod returns detailed pod information as JSON-marshalable data
	DescribePod(clusterID, namespace, name string) (interface{}, error)

	// DescribeDeployment returns detailed deployment information as JSON-marshalable data
	DescribeDeployment(clusterID, namespace, name string) (interface{}, error)

	// DescribeService returns detailed service information as JSON-marshalable data
	DescribeService(clusterID, namespace, name string) (interface{}, error)

	// GetNodes returns cluster nodes (for attention dashboard)
	GetNodes() ([]Node, error)

	// GetAllEvents returns all cluster events (for attention dashboard)
	GetAllEvents() ([]rancher.Event, error)

	// GetDaemonSets returns all DaemonSets (for attention dashboard)
	GetDaemonSets() ([]DaemonSet, error)

	// GetEtcdHealth returns etcd cluster health (bundle mode only, returns nil for live)
	GetEtcdHealth() (*EtcdHealth, error)

	// GetSystemHealth returns system health metrics (bundle mode only, returns nil for live)
	GetSystemHealth() (*SystemHealth, error)

	// Mode returns a display string for the current mode (LIVE, BUNDLE, DEMO)
	Mode() string

	// Close cleans up any resources held by the data source
	Close() error
}

// Node represents a Kubernetes node
type Node struct {
	Name   string
	Status string
}

// DaemonSet represents a DaemonSet with ready status
type DaemonSet struct {
	Name      string
	Namespace string
	Ready     string // Format: "X/Y"
}

// EtcdHealth represents etcd cluster health status
type EtcdHealth struct {
	Healthy    bool
	HasAlarms  bool
	AlarmType  string
	AlarmCount int
}

// SystemHealth represents system resource usage
type SystemHealth struct {
	MemoryUsedPercent float64
	DiskUsedPercent   float64
}
