// Package tui - Data fetching and async command generation.
// This file contains all functions that load data from the datasource,
// including fetch operations, describe operations, and data transformation helpers.
package tui

import (
	"encoding/json"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Rancheroo/r8s/internal/rancher"
)

// fetchClusters fetches clusters using the unified data source
func (a *App) fetchClusters() tea.Cmd {
	// Copy needed fields before creating closure to avoid data race
	ds := a.dataSource
	verbose := a.config.Verbose

	return func() tea.Msg {
		if ds == nil {
			return errMsg{err: fmt.Errorf("no data source available")}
		}

		clusters, err := ds.GetClusters()
		if err != nil {
			if verbose {
				return errMsg{err: fmt.Errorf("failed to fetch clusters: %w\n\n"+
					"Context: DataSource fetch\n"+
					"Hint: Check bundle data or API connectivity", err)}
			}
			return errMsg{err: fmt.Errorf("failed to fetch clusters: %w", err)}
		}

		return clustersMsg{clusters: clusters}
	}
}

// fetchProjects fetches projects using the unified data source
func (a *App) fetchProjects(clusterID string) tea.Cmd {
	// Copy needed fields before creating closure to avoid data race
	ds := a.dataSource
	verbose := a.config.Verbose

	return func() tea.Msg {
		if ds == nil {
			return errMsg{err: fmt.Errorf("no data source available")}
		}

		projects, namespaceCounts, err := ds.GetProjects(clusterID)
		if err != nil {
			if verbose {
				return errMsg{err: fmt.Errorf("failed to fetch projects: %w\n\n"+
					"Context: clusterID=%s\n"+
					"Hint: Check bundle data or API connectivity", err, clusterID)}
			}
			return errMsg{err: fmt.Errorf("failed to fetch projects: %w", err)}
		}

		return projectsMsg{projects: projects, namespaceCounts: namespaceCounts}
	}
}

// fetchNamespaces fetches namespaces using the unified data source
func (a *App) fetchNamespaces(clusterID, projectID string) tea.Cmd {
	// Copy needed fields before creating closure to avoid data race
	ds := a.dataSource
	verbose := a.config.Verbose

	return func() tea.Msg {
		if ds == nil {
			return errMsg{err: fmt.Errorf("no data source available")}
		}

		namespaces, err := ds.GetNamespaces(clusterID, projectID)
		if err != nil {
			if verbose {
				return errMsg{err: fmt.Errorf("failed to fetch namespaces: %w\n\n"+
					"Context: clusterID=%s, projectID=%s\n"+
					"Hint: Check bundle data or API connectivity", err, clusterID, projectID)}
			}
			return errMsg{err: fmt.Errorf("failed to fetch namespaces: %w", err)}
		}

		return namespacesMsg{namespaces: namespaces}
	}
}

// fetchPods fetches pods using the unified data source
func (a *App) fetchPods(projectID, namespaceName string) tea.Cmd {
	// Copy needed fields before creating closure to avoid data race
	ds := a.dataSource
	verbose := a.config.Verbose

	return func() tea.Msg {
		if ds == nil {
			return errMsg{err: fmt.Errorf("no data source available")}
		}

		pods, err := ds.GetPods(projectID, namespaceName)
		if err != nil {
			if verbose {
				return errMsg{err: fmt.Errorf("failed to fetch pods for projectID=%s, namespace=%s: %w\n\n"+
					"Context: DataSource fetch\n"+
					"Hint: Check bundle data or API connectivity", projectID, namespaceName, err)}
			}
			return errMsg{err: fmt.Errorf("failed to fetch pods: %w", err)}
		}

		return podsMsg{pods: pods}
	}
}

// fetchDeployments fetches deployments using the unified data source
func (a *App) fetchDeployments(projectID, namespaceName string) tea.Cmd {
	// Copy needed fields before creating closure to avoid data race
	ds := a.dataSource
	verbose := a.config.Verbose

	return func() tea.Msg {
		if ds == nil {
			return errMsg{err: fmt.Errorf("no data source available")}
		}

		deployments, err := ds.GetDeployments(projectID, namespaceName)
		if err != nil {
			if verbose {
				return errMsg{err: fmt.Errorf("failed to fetch deployments for projectID=%s, namespace=%s: %w\n\n"+
					"Context: DataSource fetch\n"+
					"Hint: Check bundle data or API connectivity", projectID, namespaceName, err)}
			}
			return errMsg{err: fmt.Errorf("failed to fetch deployments: %w", err)}
		}

		return deploymentsMsg{deployments: deployments}
	}
}

// fetchServices fetches services using the unified data source
func (a *App) fetchServices(projectID, namespaceName string) tea.Cmd {
	// Copy needed fields before creating closure to avoid data race
	ds := a.dataSource
	verbose := a.config.Verbose

	return func() tea.Msg {
		if ds == nil {
			return errMsg{err: fmt.Errorf("no data source available")}
		}

		services, err := ds.GetServices(projectID, namespaceName)
		if err != nil {
			if verbose {
				return errMsg{err: fmt.Errorf("failed to fetch services for projectID=%s, namespace=%s: %w\n\n"+
					"Context: DataSource fetch\n"+
					"Hint: Check bundle data or API connectivity", projectID, namespaceName, err)}
			}
			return errMsg{err: fmt.Errorf("failed to fetch services: %w", err)}
		}

		return servicesMsg{services: services}
	}
}

// fetchCRDs fetches CRDs using the unified data source
func (a *App) fetchCRDs(clusterID string) tea.Cmd {
	// Copy needed fields before creating closure to avoid data race
	ds := a.dataSource
	verbose := a.config.Verbose

	return func() tea.Msg {
		if ds == nil {
			return errMsg{err: fmt.Errorf("no data source available")}
		}

		crds, err := ds.GetCRDs(clusterID)
		if err != nil {
			if verbose {
				return errMsg{err: fmt.Errorf("failed to fetch CRDs: %w\n\n"+
					"Context: clusterID=%s\n"+
					"Hint: Check bundle data or API connectivity", err, clusterID)}
			}
			return errMsg{err: fmt.Errorf("failed to fetch CRDs: %w", err)}
		}

		return crdsMsg{crds: crds}
	}
}

// fetchCRDInstances fetches CRD instances using the unified data source
func (a *App) fetchCRDInstances(clusterID, group, version, resource string) tea.Cmd {
	// Copy needed fields before creating closure to avoid data race
	ds := a.dataSource
	verbose := a.config.Verbose

	return func() tea.Msg {
		if ds == nil {
			return errMsg{err: fmt.Errorf("no data source available")}
		}

		instances, err := ds.GetCRDInstances(clusterID, group, version, resource)
		if err != nil {
			if verbose {
				return errMsg{err: fmt.Errorf("failed to fetch CRD instances: %w\n\n"+
					"Context: clusterID=%s, group=%s, version=%s, resource=%s\n"+
					"Hint: Check CRD version and API connectivity", err, clusterID, group, version, resource)}
			}
			return errMsg{err: fmt.Errorf("failed to fetch CRD instances: %w", err)}
		}

		return crdInstancesMsg{instances: instances}
	}
}

// fetchAttention analyzes cluster health and returns attention items
func (a *App) fetchAttention() tea.Cmd {
	// Copy needed fields before creating closure to avoid data race
	ds := a.dataSource
	scanDepth := a.config.ScanDepth
	if scanDepth <= 0 {
		scanDepth = 200
	}

	return func() tea.Msg {
		if ds == nil {
			return errMsg{err: fmt.Errorf("no data source available")}
		}

		// Detect all issues across the cluster
		items := ComputeAttentionItems(ds, scanDepth)

		return attentionMsg{items: items}
	}
}

// fetchLogs fetches logs for a pod using the data source.
// Accepts container and showPrevious as parameters to avoid race conditions
// by capturing snapshot values instead of closing over mutable App fields.
func (a *App) fetchLogs(clusterID, namespace, podName, container string, showPrevious bool) tea.Cmd {
	// Copy needed fields before creating closure to avoid data race
	ds := a.dataSource
	verbose := a.config.Verbose

	return func() tea.Msg {
		if ds == nil {
			return errMsg{err: fmt.Errorf("no data source available")}
		}

		// Use the captured parameter values instead of reading from App state
		logs, err := ds.GetLogs(clusterID, namespace, podName, container, showPrevious)
		if err == nil {
			// v0.6.2: Get full container list from pod spec
			containers, containerErr := ds.GetContainers(namespace, podName)
			if containerErr != nil || len(containers) == 0 {
				// Fallback: if container detection fails, use current container or clear indicator
				if container != "" {
					containers = []string{container}
				} else {
					containers = []string{} // Empty list indicates unknown containers
				}
			}

			// Return even if empty - empty logs is valid
			// FIX (v0.5.4): Include pod name and namespace to prevent race conditions
			// FIX (v0.5.9): Include containers list
			// v0.6.2: Full container detection from pod spec
			return logsMsg{
				logs:       logs,
				containers: containers,
				podName:    podName,
				namespace:  namespace,
			}
		}

		// FIX BUG #13: NO SILENT FALLBACK - return error with context
		if verbose {
			return errMsg{err: fmt.Errorf("failed to fetch logs from data source for cluster=%s, namespace=%s, pod=%s, container=%s: %w\n\n"+
				"Context: DataSource fetch\n"+
				"Hint: Check bundle data or pod status", clusterID, namespace, podName, container, err)}
		}
		return errMsg{err: fmt.Errorf("failed to fetch logs: %w", err)}
	}
}

// describePod fetches detailed pod information
func (a *App) describePod(clusterID, namespace, name string) tea.Cmd {
	// Copy needed fields before creating closure to avoid data race
	ds := a.dataSource

	return func() tea.Msg {
		if ds == nil {
			return errMsg{err: fmt.Errorf("data source not initialized")}
		}

		// Use DataSource interface for describe - works in all modes
		data, err := ds.DescribePod(clusterID, namespace, name)
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to describe pod: %w", err)}
		}

		// Format as JSON for display
		jsonBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to format pod details: %w", err)}
		}

		content := fmt.Sprintf("Pod Details (JSON):\n\n%s", string(jsonBytes))

		return describeMsg{
			title:   fmt.Sprintf("Pod: %s/%s", namespace, name),
			content: content,
		}
	}
}

// describeDeployment fetches detailed deployment information
func (a *App) describeDeployment(clusterID, namespace, name string) tea.Cmd {
	// Copy needed fields before creating closure to avoid data race
	ds := a.dataSource

	return func() tea.Msg {
		if ds == nil {
			return errMsg{err: fmt.Errorf("data source not initialized")}
		}

		// Use DataSource interface for describe - works in all modes
		data, err := ds.DescribeDeployment(clusterID, namespace, name)
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to describe deployment: %w", err)}
		}

		jsonBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to format deployment details: %w", err)}
		}

		content := fmt.Sprintf("Deployment Details (JSON):\n\n%s", string(jsonBytes))

		return describeMsg{
			title:   fmt.Sprintf("Deployment: %s/%s", namespace, name),
			content: content,
		}
	}
}

// describeService fetches detailed service information
func (a *App) describeService(clusterID, namespace, name string) tea.Cmd {
	// Copy needed fields before creating closure to avoid data race
	ds := a.dataSource

	return func() tea.Msg {
		if ds == nil {
			return errMsg{err: fmt.Errorf("data source not initialized")}
		}

		// Use DataSource interface for describe - works in all modes
		data, err := ds.DescribeService(clusterID, namespace, name)
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to describe service: %w", err)}
		}

		jsonBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to format service details: %w", err)}
		}

		content := fmt.Sprintf("Service Details (JSON):\n\n%s", string(jsonBytes))

		return describeMsg{
			title:   fmt.Sprintf("Service: %s/%s", namespace, name),
			content: content,
		}
	}
}

// updateNamespaceCounts updates the count of namespaces per project
func (a *App) updateNamespaceCounts(namespaces []rancher.Namespace) {
	// Initialize counts
	counts := make(map[string]int)

	// Count namespaces per project
	for _, ns := range namespaces {
		if ns.ProjectID != "" {
			counts[ns.ProjectID]++
		}
	}

	// Update the app's namespace counts
	a.projectNamespaceCounts = counts
}

// getCRDInstanceCount returns the count of instances for a given CRD using datasource
// TODO(FUTURE_WORK): This function blocks UI rendering during table updates.
// Should be refactored to:
// 1. Return cached counts during render (non-blocking)
// 2. Fetch counts asynchronously in background goroutines
// 3. Store results in App.crdInstanceCounts map[string]int (key: clusterID+group+resource)
// 4. Signal UI refresh after cache update
// See FUTURE_WORK.md for detailed async implementation plan
func (a *App) getCRDInstanceCount(group, resource string) int {
	if a.dataSource == nil {
		return 0
	}

	// Get the storage version for this CRD
	var version string
	for _, crd := range a.crds {
		if crd.Spec.Group == group && crd.Spec.Names.Plural == resource {
			for _, v := range crd.Spec.Versions {
				if v.Storage {
					version = v.Name
					break
				}
			}
			if version == "" && len(crd.Spec.Versions) > 0 {
				version = crd.Spec.Versions[0].Name
			}
			break
		}
	}

	if version == "" {
		return 0
	}

	instances, err := a.dataSource.GetCRDInstances(a.currentView.clusterID, group, version, resource)
	if err != nil {
		return 0 // Silently return 0 for counts (non-critical)
	}

	return len(instances)
}
