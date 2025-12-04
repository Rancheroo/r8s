package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Rancheroo/r8s/internal/datasource"
)

// AttentionSeverity represents the severity level of an attention item
type AttentionSeverity int

const (
	SeverityCritical AttentionSeverity = 0
	SeverityWarning  AttentionSeverity = 1
	SeverityInfo     AttentionSeverity = 2
)

// AttentionItem represents a single issue requiring attention
type AttentionItem struct {
	Severity     AttentionSeverity
	Emoji        string
	Title        string // e.g., "nginx-deploy-xyz"
	Description  string // e.g., "CrashLoopBackOff"
	Namespace    string
	Count        int       // For aggregated items (e.g., restart count, error count)
	Timestamp    time.Time // When detected
	ResourceType string    // "pod", "node", "etcd", "daemonset", "event", "log", "system"

	// Navigation context for drill-down
	PodName       string
	ContainerName string
	ClusterID     string
}

// ComputeAttentionItems runs all signal detectors and returns prioritized list of issues
func ComputeAttentionItems(ds datasource.DataSource) []AttentionItem {
	var items []AttentionItem

	// Tier 1: Pod Health (Critical)
	items = append(items, detectPodHealth(ds)...)

	// Tier 2: Cluster Health (Critical)
	items = append(items, detectClusterHealth(ds)...)

	// Tier 3: Events (Warning)
	items = append(items, detectEventIssues(ds)...)

	// Tier 4: Logs (Info) - Sample only for performance
	// Commented out for initial implementation - can be slow
	// items = append(items, detectLogIssues(ds)...)

	// Tier 5: System Health (Bundle only)
	items = append(items, detectSystemHealth(ds)...)

	// Sort by severity (Critical → Warning → Info)
	sortAttentionItems(items)

	// Limit to top 15 items
	if len(items) > 15 {
		items = items[:15]
	}

	return items
}

// detectPodHealth detects pod-level issues
func detectPodHealth(ds datasource.DataSource) []AttentionItem {
	var items []AttentionItem

	// Get all pods across all namespaces
	pods, err := ds.GetAllPods()
	if err != nil {
		// Silent failure - we'll detect what we can
		return items
	}

	for _, pod := range pods {
		// Extract namespace from NamespaceID (may be "cluster:namespace" format)
		namespace := pod.NamespaceID
		if strings.Contains(namespace, ":") {
			parts := strings.Split(namespace, ":")
			if len(parts) > 1 {
				namespace = parts[1]
			}
		}

		// Critical: CrashLoopBackOff
		if strings.Contains(pod.State, "CrashLoopBackOff") ||
			strings.Contains(pod.KubectlStatus, "CrashLoopBackOff") {
			items = append(items, AttentionItem{
				Severity:     SeverityCritical,
				Emoji:        "💀",
				Title:        pod.Name,
				Description:  "CrashLoopBackOff",
				Namespace:    namespace,
				ResourceType: "pod",
				PodName:      pod.Name,
				Timestamp:    time.Now(),
			})
			continue
		}

		// Critical: OOMKilled
		if strings.Contains(pod.State, "OOMKilled") ||
			strings.Contains(pod.KubectlStatus, "OOMKilled") {
			items = append(items, AttentionItem{
				Severity:     SeverityCritical,
				Emoji:        "💀",
				Title:        pod.Name,
				Description:  "OOMKilled",
				Namespace:    namespace,
				ResourceType: "pod",
				PodName:      pod.Name,
				Timestamp:    time.Now(),
			})
			continue
		}

		// Critical: Error/Failed state
		if strings.Contains(pod.State, "Error") || strings.Contains(pod.State, "Failed") ||
			strings.Contains(pod.KubectlStatus, "Error") || strings.Contains(pod.KubectlStatus, "Failed") {
			items = append(items, AttentionItem{
				Severity:     SeverityCritical,
				Emoji:        "💀",
				Title:        pod.Name,
				Description:  "Error state",
				Namespace:    namespace,
				ResourceType: "pod",
				PodName:      pod.Name,
				Timestamp:    time.Now(),
			})
			continue
		}

		// Critical: ImagePullBackOff / ErrImagePull
		if strings.Contains(pod.State, "ImagePullBackOff") || strings.Contains(pod.State, "ErrImagePull") ||
			strings.Contains(pod.KubectlStatus, "ImagePullBackOff") || strings.Contains(pod.KubectlStatus, "ErrImagePull") {
			items = append(items, AttentionItem{
				Severity:     SeverityCritical,
				Emoji:        "🚫",
				Title:        pod.Name,
				Description:  "ImagePullBackOff",
				Namespace:    namespace,
				ResourceType: "pod",
				PodName:      pod.Name,
				Timestamp:    time.Now(),
			})
			continue
		}

		// Critical: Evicted
		if strings.Contains(pod.State, "Evicted") || strings.Contains(pod.KubectlStatus, "Evicted") {
			items = append(items, AttentionItem{
				Severity:     SeverityCritical,
				Emoji:        "🚫",
				Title:        pod.Name,
				Description:  "Evicted",
				Namespace:    namespace,
				ResourceType: "pod",
				PodName:      pod.Name,
				Timestamp:    time.Now(),
			})
			continue
		}

		// Warning: High restart count (≥3)
		restartCount := pod.RestartCount
		if pod.KubectlRestarts > 0 {
			restartCount = pod.KubectlRestarts
		}

		if restartCount >= 3 {
			items = append(items, AttentionItem{
				Severity:     SeverityWarning,
				Emoji:        "🔥",
				Title:        pod.Name,
				Description:  fmt.Sprintf("%d restarts", restartCount),
				Namespace:    namespace,
				Count:        restartCount,
				ResourceType: "pod",
				PodName:      pod.Name,
				Timestamp:    time.Now(),
			})
			continue
		}

		// Warning: Not Ready (but not in terminal state)
		// Only flag if containers are actually not ready (e.g., "1/2" not "2/2")
		if pod.KubectlReady != "" && !isHealthyReadyStatus(pod.KubectlReady) &&
			pod.KubectlStatus == "Running" {
			items = append(items, AttentionItem{
				Severity:     SeverityWarning,
				Emoji:        "⚠️",
				Title:        pod.Name,
				Description:  fmt.Sprintf("Not ready (%s)", pod.KubectlReady),
				Namespace:    namespace,
				ResourceType: "pod",
				PodName:      pod.Name,
				Timestamp:    time.Now(),
			})
		}
	}

	return items
}

// detectClusterHealth detects cluster-level issues
func detectClusterHealth(ds datasource.DataSource) []AttentionItem {
	var items []AttentionItem

	// Check node health
	nodes, err := ds.GetNodes()
	if err == nil {
		for _, node := range nodes {
			if strings.Contains(node.Status, "NotReady") ||
				strings.Contains(node.Status, "Unknown") {
				items = append(items, AttentionItem{
					Severity:     SeverityCritical,
					Emoji:        "📍",
					Title:        node.Name,
					Description:  "NotReady",
					Namespace:    "cluster",
					ResourceType: "node",
					Timestamp:    time.Now(),
				})
			}
		}
	}

	// Check etcd health (bundle mode only)
	etcdHealth, err := ds.GetEtcdHealth()
	if err == nil && etcdHealth != nil {
		if etcdHealth.HasAlarms {
			items = append(items, AttentionItem{
				Severity:     SeverityCritical,
				Emoji:        "🚨",
				Title:        "ETCD",
				Description:  fmt.Sprintf("ALARM: %s", etcdHealth.AlarmType),
				Namespace:    "etcd",
				Count:        etcdHealth.AlarmCount,
				ResourceType: "etcd",
				Timestamp:    time.Now(),
			})
		}

		if !etcdHealth.Healthy {
			items = append(items, AttentionItem{
				Severity:     SeverityCritical,
				Emoji:        "⚠️",
				Title:        "ETCD",
				Description:  "Unhealthy endpoints",
				Namespace:    "etcd",
				ResourceType: "etcd",
				Timestamp:    time.Now(),
			})
		}
	}

	// Check DaemonSets
	daemonsets, err := ds.GetDaemonSets()
	if err == nil {
		for _, ds := range daemonsets {
			// Parse "X/Y" ready format
			if strings.Contains(ds.Ready, "/") {
				parts := strings.Split(ds.Ready, "/")
				if len(parts) == 2 && parts[0] != parts[1] {
					items = append(items, AttentionItem{
						Severity:     SeverityWarning,
						Emoji:        "🔧",
						Title:        fmt.Sprintf("%s DS", ds.Name),
						Description:  fmt.Sprintf("%s ready", ds.Ready),
						Namespace:    ds.Namespace,
						ResourceType: "daemonset",
						Timestamp:    time.Now(),
					})
				}
			}
		}
	}

	return items
}

// detectEventIssues detects issues from cluster events
func detectEventIssues(ds datasource.DataSource) []AttentionItem {
	var items []AttentionItem

	events, err := ds.GetAllEvents()
	if err != nil {
		return items
	}

	// Aggregate Warning events by reason
	warningCounts := make(map[string]int)
	for _, event := range events {
		if event.Type == "Warning" && event.Count > 0 {
			warningCounts[event.Reason] += event.Count
		}
	}

	// Report high-impact warning types
	for reason, count := range warningCounts {
		if count >= 5 { // At least 5 occurrences
			emoji := "🟨"
			severity := SeverityWarning

			// Elevate critical event types
			if strings.Contains(reason, "Failed") ||
				strings.Contains(reason, "Error") ||
				reason == "BackOff" {
				emoji = "🟥"
				severity = SeverityWarning // Keep as warning, but with red emoji
			}

			items = append(items, AttentionItem{
				Severity:     severity,
				Emoji:        emoji,
				Title:        fmt.Sprintf("%d Warning events", count),
				Description:  reason,
				Namespace:    "cluster",
				Count:        count,
				ResourceType: "event",
				Timestamp:    time.Now(),
			})
		}
	}

	return items
}

// detectSystemHealth detects system-level issues (bundle mode only)
func detectSystemHealth(ds datasource.DataSource) []AttentionItem {
	var items []AttentionItem

	sysHealth, err := ds.GetSystemHealth()
	if err != nil || sysHealth == nil {
		return items
	}

	// Memory pressure (>90% used)
	if sysHealth.MemoryUsedPercent > 90 {
		items = append(items, AttentionItem{
			Severity:     SeverityInfo,
			Emoji:        "💾",
			Title:        "Memory",
			Description:  fmt.Sprintf("%.0f%% used", sysHealth.MemoryUsedPercent),
			Namespace:    "system",
			ResourceType: "system",
			Timestamp:    time.Now(),
		})
	}

	// Disk pressure (>85% used)
	if sysHealth.DiskUsedPercent > 85 {
		items = append(items, AttentionItem{
			Severity:     SeverityInfo,
			Emoji:        "💿",
			Title:        "Disk",
			Description:  fmt.Sprintf("%.0f%% used", sysHealth.DiskUsedPercent),
			Namespace:    "system",
			ResourceType: "system",
			Timestamp:    time.Now(),
		})
	}

	return items
}

// isHealthyReadyStatus checks if a pod's ready status indicates all containers are ready
// Examples: "2/2" → true, "3/3" → true, "1/2" → false, "0/3" → false
func isHealthyReadyStatus(ready string) bool {
	if ready == "" {
		return false
	}

	parts := strings.Split(ready, "/")
	if len(parts) != 2 {
		return false
	}

	// Check if ready count equals total count (e.g., "2/2" or "3/3")
	return parts[0] == parts[1]
}

// sortAttentionItems sorts items by severity (Critical first)
func sortAttentionItems(items []AttentionItem) {
	// Simple bubble sort by severity
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].Severity > items[j].Severity {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}
