// Package tui - Table rendering and view updates.
// This file handles table configuration and rendering for all 9 view types,
// including columns, rows, sorting, and styling for each resource view.
package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/evertras/bubble-table/table"

	"github.com/Rancheroo/r8s/internal/rancher"
)

// updateTable updates the table content based on current view
func (a *App) updateTable() {
	switch a.currentView.viewType {
	case ViewCRDs:
		if len(a.crds) > 0 {
			// Dynamic column widths - auto-adapt to terminal size ("Show, Don't Ask")
			columns := a.calculateColumnWidths(getCRDColumnSpecs())

			rows := []table.Row{}
			for _, crd := range a.crds {
				// Get instance count for this CRD
				instanceCount := a.getCRDInstanceCount(crd.Spec.Group, crd.Spec.Names.Plural)

				rows = append(rows, table.NewRow(table.RowData{
					"name":      crd.Metadata.Name,
					"group":     crd.Spec.Group,
					"kind":      crd.Spec.Names.Kind,
					"scope":     crd.Spec.Scope,
					"instances": fmt.Sprintf("%d", instanceCount),
				}))
			}

			a.table = table.New(columns).
				WithRows(rows).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(true).
				BorderRounded()
		} else {
			a.table = table.New([]table.Column{table.NewColumn("message", "MESSAGE", 80)}).
				WithRows([]table.Row{table.NewRow(table.RowData{"message": "No CRDs available"})}).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(false).
				BorderRounded()
		}

	case ViewClusters:
		if len(a.clusters) > 0 {
			// Dynamic column widths - auto-adapt to terminal size
			columns := a.calculateColumnWidths(getClusterColumnSpecs())

			rows := []table.Row{}
			for _, cluster := range a.clusters {
				created := "N/A"
				if !cluster.Created.IsZero() {
					created = fmt.Sprintf("%dd", int(time.Since(cluster.Created).Hours()/24))
				}

				rows = append(rows, table.NewRow(table.RowData{
					"name":     cluster.Name,
					"provider": cluster.Provider,
					"state":    cluster.State,
					"created":  created,
				}))
			}

			a.table = table.New(columns).
				WithRows(rows).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(true).
				BorderRounded()
		} else {
			a.table = table.New([]table.Column{table.NewColumn("message", "MESSAGE", 80)}).
				WithRows([]table.Row{table.NewRow(table.RowData{"message": "No clusters available"})}).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(false).
				BorderRounded()
		}

	case ViewProjects:
		if len(a.projects) > 0 {
			// Dynamic column widths - auto-adapt to terminal size
			columns := a.calculateColumnWidths(getProjectColumnSpecs())

			rows := []table.Row{}
			for _, project := range a.projects {
				namespaceCount := a.projectNamespaceCounts[project.ID]
				displayName := project.DisplayName
				if displayName == "" {
					displayName = project.Name
				}

				rows = append(rows, table.NewRow(table.RowData{
					"name":        project.Name,
					"displayName": displayName,
					"state":       project.State,
					"namespaces":  fmt.Sprintf("%d", namespaceCount),
				}))
			}

			a.table = table.New(columns).
				WithRows(rows).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(true).
				BorderRounded()
		} else {
			a.table = table.New([]table.Column{table.NewColumn("message", "MESSAGE", 80)}).
				WithRows([]table.Row{table.NewRow(table.RowData{"message": "No projects available"})}).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(false).
				BorderRounded()
		}

	case ViewNamespaces:
		if len(a.namespaces) > 0 {
			// Compute namespace health (scan all pods for E/W counts)
			scanDepth := a.config.ScanDepth
			if scanDepth <= 0 {
				scanDepth = 200
			}
			nsHealth := ComputeNamespaceHealth(a.dataSource, scanDepth)

			// Sort namespaces by total issue count (worst first)
			sortedNS := make([]rancher.Namespace, len(a.namespaces))
			copy(sortedNS, a.namespaces)

			// Sort by total issues descending using sort.Slice (sorts in-place)
			SortNamespacesByHealth(sortedNS, nsHealth)

			// Dynamic column widths - auto-adapt to terminal size
			columns := a.calculateColumnWidths(getNamespaceColumnSpecs())

			rows := []table.Row{}
			for _, ns := range sortedNS {
				created := "N/A"
				if !ns.Created.IsZero() {
					days := int(time.Since(ns.Created).Hours() / 24)
					if days > 0 {
						created = fmt.Sprintf("%dd", days)
					} else {
						hours := int(time.Since(ns.Created).Hours())
						if hours > 0 {
							created = fmt.Sprintf("%dh", hours)
						} else {
							created = fmt.Sprintf("%dm", int(time.Since(ns.Created).Minutes()))
						}
					}
				}

				// Format ISSUES column with color coding
				health := nsHealth[ns.Name]
				issuesDisplay := "✅ Clean"

				if health.Total > 0 {
					// Color coding logic:
					// Red (🔥): >50 errors
					// Yellow (⚠️): >20 warnings OR 1-50 errors
					// Green: Minor issues
					emoji := "✅"
					if health.Errors > 50 {
						emoji = "🔥"
					} else if health.Warnings > 20 || health.Errors > 0 {
						emoji = "⚠️"
					}

					// Format with K/M abbreviation for large numbers
					errStr := formatCount(health.Errors)
					warnStr := formatCount(health.Warnings)
					issuesDisplay = fmt.Sprintf("%s %sE/%sW", emoji, errStr, warnStr)
				}

				rows = append(rows, table.NewRow(table.RowData{
					"name":    ns.Name,
					"issues":  issuesDisplay,
					"state":   ns.State,
					"project": ns.ProjectID,
					"created": created,
				}))
			}

			a.table = table.New(columns).
				WithRows(rows).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(true).
				BorderRounded()
		} else {
			a.table = table.New([]table.Column{table.NewColumn("message", "MESSAGE", 80)}).
				WithRows([]table.Row{table.NewRow(table.RowData{"message": "No namespaces available"})}).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(false).
				BorderRounded()
		}

	case ViewPods:
		if len(a.pods) > 0 {
			// Pre-populate cache with E/W counts for sorting
			a.populatePodCounts()

			// Apply sorting based on current sort mode
			sortMode, exists := a.sortModes[ViewPods]
			if !exists {
				sortMode = a.sortMode // Use global default
			}

			// Sort pods according to mode (create a copy to avoid in-place modification)
			var sortedPods []rancher.Pod
			switch sortMode {
			case SortByCount:
				sortedPods = SortPodsByCount(a.pods, a.cachedPodCounts)
			case SortBySeverity:
				sortedPods = SortPodsBySeverity(a.pods)
			case SortByName:
				sortedPods = SortPodsByName(a.pods)
			default:
				// Default: copy the slice to avoid aliasing
				sortedPods = make([]rancher.Pod, len(a.pods))
				copy(sortedPods, a.pods)
			}

			// Dynamic column widths - auto-adapt to terminal size
			columns := a.calculateColumnWidths(getPodColumnSpecs())

			rows := []table.Row{}
			for _, pod := range sortedPods {
				namespaceName := "default"
				if pod.NamespaceID != "" {
					if strings.Contains(pod.NamespaceID, ":") {
						parts := strings.Split(pod.NamespaceID, ":")
						if len(parts) > 1 {
							namespaceName = parts[1]
						}
					} else {
						namespaceName = pod.NamespaceID
					}
				}

				// Get node name with fallback support
				nodeName := a.getPodNodeName(pod)

				// Get error/warning counts from cache (matches namespace view approach)
				weCount := "✅" // Default clean state
				cacheKey := fmt.Sprintf("%s/%s", namespaceName, pod.Name)
				if counts, exists := a.cachedPodCounts[cacheKey]; exists {
					if counts.Errors > 0 || counts.Warnings > 0 {
						// Format like namespace view: use formatCount() for large numbers
						errStr := formatCount(counts.Errors)
						warnStr := formatCount(counts.Warnings)
						weCount = fmt.Sprintf("%sE/%sW", errStr, warnStr)
					}
				}

				rows = append(rows, table.NewRow(table.RowData{
					"name":      pod.Name,
					"namespace": namespaceName,
					"state":     pod.State,
					"we":        weCount,
					"node":      nodeName,
				}))
			}

			a.table = table.New(columns).
				WithRows(rows).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(true).
				BorderRounded()
		} else {
			a.table = table.New([]table.Column{table.NewColumn("message", "MESSAGE", 80)}).
				WithRows([]table.Row{table.NewRow(table.RowData{"message": "No pods available"})}).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(false).
				BorderRounded()
		}

	case ViewDeployments:
		if len(a.deployments) > 0 {
			// Dynamic column widths - auto-adapt to terminal size
			columns := a.calculateColumnWidths(getDeploymentColumnSpecs())

			rows := []table.Row{}
			for _, deployment := range a.deployments {
				namespaceName := "default"
				if deployment.NamespaceID != "" {
					if strings.Contains(deployment.NamespaceID, ":") {
						parts := strings.Split(deployment.NamespaceID, ":")
						if len(parts) > 1 {
							namespaceName = parts[1]
						}
					} else {
						namespaceName = deployment.NamespaceID
					}
				}

				// Get replica counts - prefer Scale field, fallback to direct fields
				var totalReplicas, readyReplicas, updatedReplicas, availableReplicas int

				if deployment.Scale != nil {
					// Use Scale field if available
					totalReplicas = deployment.Scale.Scale
					readyReplicas = deployment.Scale.Ready
					availableReplicas = deployment.Scale.Ready // Scale.Ready represents available
					updatedReplicas = deployment.Scale.Ready   // Assume updated = ready
				} else {
					// Fallback to direct fields
					totalReplicas = deployment.Replicas
					readyReplicas = deployment.ReadyReplicas
					availableReplicas = deployment.AvailableReplicas
					// Try both possible field names for updated replicas
					if deployment.UpToDateReplicas > 0 {
						updatedReplicas = deployment.UpToDateReplicas
					} else {
						updatedReplicas = deployment.UpdatedReplicas
					}
				}

				rows = append(rows, table.NewRow(table.RowData{
					"name":      deployment.Name,
					"namespace": namespaceName,
					"ready":     fmt.Sprintf("%d/%d", readyReplicas, totalReplicas),
					"uptodate":  fmt.Sprintf("%d", updatedReplicas),
					"available": fmt.Sprintf("%d", availableReplicas),
				}))
			}

			a.table = table.New(columns).
				WithRows(rows).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(true).
				BorderRounded()
		} else {
			a.table = table.New([]table.Column{table.NewColumn("message", "MESSAGE", 80)}).
				WithRows([]table.Row{table.NewRow(table.RowData{"message": "No deployments available"})}).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(false).
				BorderRounded()
		}

	case ViewServices:
		if len(a.services) > 0 {
			// Dynamic column widths - auto-adapt to terminal size
			columns := a.calculateColumnWidths(getServiceColumnSpecs())

			rows := []table.Row{}
			for _, service := range a.services {
				namespaceName := "default"
				if service.NamespaceID != "" {
					if strings.Contains(service.NamespaceID, ":") {
						parts := strings.Split(service.NamespaceID, ":")
						if len(parts) > 1 {
							namespaceName = parts[1]
						}
					} else {
						namespaceName = service.NamespaceID
					}
				}

				// Format ports
				var portStrings []string
				for _, port := range service.Ports {
					portStr := fmt.Sprintf("%d/%s", port.Port, port.Protocol)
					if port.NodePort > 0 {
						portStr = fmt.Sprintf("%d:%d/%s", port.Port, port.NodePort, port.Protocol)
					}
					portStrings = append(portStrings, portStr)
				}
				portsDisplay := strings.Join(portStrings, ",")

				rows = append(rows, table.NewRow(table.RowData{
					"name":       service.Name,
					"namespace":  namespaceName,
					"type":       service.Kind,
					"cluster_ip": service.ClusterIP,
					"ports":      portsDisplay,
				}))
			}

			a.table = table.New(columns).
				WithRows(rows).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(true).
				BorderRounded()
		} else {
			a.table = table.New([]table.Column{table.NewColumn("message", "MESSAGE", 80)}).
				WithRows([]table.Row{table.NewRow(table.RowData{"message": "No services available"})}).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(false).
				BorderRounded()
		}

	case ViewCRDInstances:
		if len(a.crdInstances) > 0 {
			// Dynamic column widths - auto-adapt to terminal size
			columns := a.calculateColumnWidths(getCRDInstanceColumnSpecs())

			rows := []table.Row{}
			for _, instance := range a.crdInstances {
				// Extract metadata
				metadata, _ := instance["metadata"].(map[string]interface{})
				name := ""
				namespace := ""
				createdTime := ""

				if metadata != nil {
					if n, ok := metadata["name"].(string); ok {
						name = n
					}
					if ns, ok := metadata["namespace"].(string); ok {
						namespace = ns
					} else {
						namespace = "cluster-scoped"
					}
					if ct, ok := metadata["creationTimestamp"].(string); ok {
						// Parse and calculate age with human-readable format (same as namespace view)
						if t, err := time.Parse(time.RFC3339, ct); err == nil {
							days := int(time.Since(t).Hours() / 24)
							if days > 0 {
								createdTime = fmt.Sprintf("%dd", days)
							} else {
								hours := int(time.Since(t).Hours())
								if hours > 0 {
									createdTime = fmt.Sprintf("%dh", hours)
								} else {
									createdTime = fmt.Sprintf("%dm", int(time.Since(t).Minutes()))
								}
							}
						}
					}
				}

				// Try to extract status
				status := "N/A"
				if statusObj, ok := instance["status"].(map[string]interface{}); ok {
					if conditions, ok := statusObj["conditions"].([]interface{}); ok && len(conditions) > 0 {
						if cond, ok := conditions[0].(map[string]interface{}); ok {
							if condType, ok := cond["type"].(string); ok {
								if condStatus, ok := cond["status"].(string); ok {
									status = fmt.Sprintf("%s: %s", condType, condStatus)
								}
							}
						}
					}
				}

				rows = append(rows, table.NewRow(table.RowData{
					"name":      name,
					"namespace": namespace,
					"age":       createdTime,
					"status":    status,
				}))
			}

			a.table = table.New(columns).
				WithRows(rows).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(true).
				BorderRounded()
		} else {
			a.table = table.New([]table.Column{table.NewColumn("message", "MESSAGE", 80)}).
				WithRows([]table.Row{table.NewRow(table.RowData{"message": fmt.Sprintf("No %s instances available", a.currentView.crdKind)})}).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(false).
				BorderRounded()
		}

	case ViewAttention:
		// Attention Dashboard table rendering
		if len(a.attentionItems) == 0 {
			// No issues - show clean state
			a.table = table.New([]table.Column{table.NewColumn("message", "MESSAGE", 80)}).
				WithRows([]table.Row{table.NewRow(table.RowData{"message": "✨ All systems healthy - no attention items detected"})}).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(false).
				BorderRounded()
		} else {
			// Build attention dashboard with issues - dynamic column widths
			columns := a.calculateColumnWidths(getAttentionColumnSpecs())

			// Get items to display (respect expansion state)
			displayedItems := a.attentionItems
			if !a.attentionExpanded && len(a.attentionItems) > defaultDashboardCap {
				displayedItems = a.attentionItems[:defaultDashboardCap]
			}

			rows := []table.Row{}
			for _, item := range displayedItems {
				// Format severity emoji
				severityEmoji := "ℹ️"
				if item.Severity == SeverityCritical {
					severityEmoji = "🔥"
				} else if item.Severity == SeverityWarning {
					severityEmoji = "⚠️"
				}

				// Format context (namespace/resource type)
				context := fmt.Sprintf("%s/%s", item.Namespace, item.ResourceType)
				if item.Namespace == "" {
					context = item.ResourceType
				}

				// Format count
				countStr := ""
				if item.Count > 1 {
					countStr = fmt.Sprintf("×%d", item.Count)
				}

				rows = append(rows, table.NewRow(table.RowData{
					"severity": severityEmoji,
					"title":    item.Title,
					"context":  context,
					"count":    countStr,
				}))
			}

			a.table = table.New(columns).
				WithRows(rows).
				HeaderStyle(headerStyle).
				WithBaseStyle(baseStyle).
				WithPageSize(a.height - 8).
				Focused(true).
				BorderRounded()
		}
	}
}
