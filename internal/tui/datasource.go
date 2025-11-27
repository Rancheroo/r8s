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
	// GetPods returns pods for the given project and namespace
	GetPods(projectID, namespace string) ([]rancher.Pod, error)

	// GetLogs returns log lines for the specified pod and container
	GetLogs(clusterID, namespace, pod, container string, previous bool) ([]string, error)

	// GetContainers returns available containers for a pod
	GetContainers(namespace, pod string) ([]string, error)

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
	// For now, return error to trigger mock data
	// In production, would call: ds.client.GetPodLogs(...)
	return nil, fmt.Errorf("live logs not yet implemented")
}

// GetContainers returns containers for a pod
func (ds *LiveDataSource) GetContainers(namespace, pod string) ([]string, error) {
	// For now, return default container
	return []string{"app"}, nil
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
func NewBundleDataSource(bundlePath string) (*BundleDataSource, error) {
	opts := bundle.ImportOptions{
		Path:    bundlePath,
		MaxSize: 100 * 1024 * 1024, // 100MB for TUI mode
	}

	b, err := bundle.Load(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to load bundle: %w", err)
	}

	return &BundleDataSource{bundle: b}, nil
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
