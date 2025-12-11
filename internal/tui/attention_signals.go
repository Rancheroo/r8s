package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Rancheroo/r8s/internal/datasource"
	"github.com/Rancheroo/r8s/internal/rancher"
)

// AttentionSeverity represents the severity level of an attention item
type AttentionSeverity int

const (
	SeverityCritical AttentionSeverity = 0
	SeverityWarning  AttentionSeverity = 1
	SeverityInfo     AttentionSeverity = 2
)

// SortMode defines how attention items should be sorted
type SortMode int

const (
	SortByCount    SortMode = iota // Default: highest ERR+WARN total descending
	SortBySeverity                 // Current behavior (Critical→Warning→Info)
	SortByName                     // Alphabetical by title
)

// String returns human-readable sort mode name for status bar
func (s SortMode) String() string {
	switch s {
	case SortByCount:
		return "Count ▼"
	case SortBySeverity:
		return "Severity"
	case SortByName:
		return "Name"
	default:
		return "Unknown"
	}
}

// PodCounts holds cached error/warning counts for a pod
type PodCounts struct {
	Errors   int
	Warnings int
	Total    int
}

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

	// Expandable content for aggregate items (events)
	AffectedPods      []string       // Top 10 pod names involved in this event
	AffectedPodCounts map[string]int // Event count per pod
}

// ComputeAttentionItems runs all signal detectors and returns prioritized list of issues
func ComputeAttentionItems(ds datasource.DataSource, scanDepth int) []AttentionItem {
	var items []AttentionItem

	// Default scan depth if not set
	if scanDepth <= 0 {
		scanDepth = 200
	}

	// Tier 1: Pod Health (Critical)
	items = append(items, detectPodHealth(ds)...)

	// Tier 2: Cluster Health (Critical)
	items = append(items, detectClusterHealth(ds)...)

	// Tier 3: Events (Warning)
	items = append(items, detectEventIssues(ds)...)

	// Tier 4: Logs (Critical/Warning) - Sample logs for error/warn counts
	items = append(items, detectLogIssues(ds, scanDepth)...)

	// Tier 5: System Health (Bundle only)
	items = append(items, detectSystemHealth(ds)...)

	// Sort by severity (Critical → Warning → Info)
	sortAttentionItems(items)

	// Limit to top 100 items (can scroll to see all)
	if len(items) > 100 {
		items = items[:100]
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

		// Critical: CrashLoopBackOff (case-insensitive for Rancher API compatibility)
		stateLower := strings.ToLower(pod.State)
		kubectlStatusLower := strings.ToLower(pod.KubectlStatus)

		if strings.Contains(stateLower, "crashloopbackoff") ||
			strings.Contains(kubectlStatusLower, "crashloopbackoff") {
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

		// Critical: OOMKilled (distinct emoji for visibility)
		if strings.Contains(stateLower, "oomkilled") ||
			strings.Contains(kubectlStatusLower, "oomkilled") {
			items = append(items, AttentionItem{
				Severity:     SeverityCritical,
				Emoji:        "🧨", // Distinct from CrashLoop
				Title:        pod.Name,
				Description:  "OOMKilled",
				Namespace:    namespace,
				ResourceType: "pod",
				PodName:      pod.Name,
				Timestamp:    time.Now(),
			})
			continue
		}

		// Critical: Error/Failed state (case-insensitive)
		if strings.Contains(stateLower, "error") || strings.Contains(stateLower, "failed") ||
			strings.Contains(kubectlStatusLower, "error") || strings.Contains(kubectlStatusLower, "failed") {
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

		// Critical: ImagePullBackOff / ErrImagePull (case-insensitive)
		if strings.Contains(stateLower, "imagepullbackoff") || strings.Contains(stateLower, "errimagepull") ||
			strings.Contains(kubectlStatusLower, "imagepullbackoff") || strings.Contains(kubectlStatusLower, "errimagepull") {
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

		// Critical: Evicted (case-insensitive)
		if strings.Contains(stateLower, "evicted") || strings.Contains(kubectlStatusLower, "evicted") {
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

	// Aggregate Warning events by reason and track affected pods
	type eventStats struct {
		count int
		pods  map[string]int // pod name -> event count
	}
	warningStats := make(map[string]*eventStats)

	for _, event := range events {
		if event.Type == "Warning" && event.Count > 0 {
			if warningStats[event.Reason] == nil {
				warningStats[event.Reason] = &eventStats{
					pods: make(map[string]int),
				}
			}
			warningStats[event.Reason].count += event.Count

			// Track pod name if available
			if event.PodName != "" {
				warningStats[event.Reason].pods[event.PodName] += event.Count
			}
		}
	}

	// Report high-impact warning types (collapsed format with pod names)
	for reason, stats := range warningStats {
		if stats.count >= 5 { // At least 5 occurrences
			emoji := "🟨"
			severity := SeverityWarning

			// Elevate critical event types
			if strings.Contains(reason, "Failed") ||
				strings.Contains(reason, "Error") ||
				reason == "BackOff" {
				emoji = "🟥"
				severity = SeverityWarning // Keep as warning, but with red emoji
			}

			// Get top 10 affected pods
			affectedPods := getTopPods(stats.pods, 10)

			// Collapsed format: "467339× DNSConfigForming" with expandable pod list
			items = append(items, AttentionItem{
				Severity:          severity,
				Emoji:             emoji,
				Title:             fmt.Sprintf("%d× %s", stats.count, reason),
				Description:       "Warning events",
				Namespace:         "cluster",
				Count:             stats.count,
				ResourceType:      "event",
				AffectedPods:      affectedPods,
				AffectedPodCounts: stats.pods, // Store full count map for display
				Timestamp:         time.Now(),
			})
		}
	}

	return items
}

// getTopPods returns the top N pods by event count
func getTopPods(pods map[string]int, n int) []string {
	type podCount struct {
		name  string
		count int
	}

	var sorted []podCount
	for name, count := range pods {
		sorted = append(sorted, podCount{name, count})
	}

	// Simple bubble sort by count (descending)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].count < sorted[j].count {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Return top N names
	result := []string{}
	for i := 0; i < len(sorted) && i < n; i++ {
		result = append(result, sorted[i].name)
	}

	return result
}

// detectLogIssues scans pod logs for error and warning patterns
// Samples first N lines per pod for performance (N = scanDepth, tunable via --scan flag)
func detectLogIssues(ds datasource.DataSource, scanDepth int) []AttentionItem {
	var items []AttentionItem

	// Get all pods to scan their logs
	pods, err := ds.GetAllPods()
	if err != nil {
		return items
	}

	// Sample max 10 pods to avoid performance issues
	maxPodsToScan := 10
	if len(pods) > maxPodsToScan {
		pods = pods[:maxPodsToScan]
	}

	for _, pod := range pods {
		// Extract namespace
		namespace := pod.NamespaceID
		if strings.Contains(namespace, ":") {
			parts := strings.Split(namespace, ":")
			if len(parts) > 1 {
				namespace = parts[1]
			}
		}

		// Try to get logs for this pod
		logs, err := ds.GetLogs("", namespace, pod.Name, "", false)
		if err != nil {
			// Skip pods without logs (common for init containers, etc.)
			continue
		}

		// Sample first N lines for performance (tunable via --scan flag)
		if len(logs) > scanDepth {
			logs = logs[:scanDepth]
		}

		// Count errors and warnings using the same detection functions as log view
		errorCount := 0
		warnCount := 0

		for _, line := range logs {
			if isErrorLog(line) {
				errorCount++
			} else if isWarnLog(line) {
				warnCount++
			}
		}

		// Report pods with significant error counts (>10 errors)
		if errorCount > 10 {
			items = append(items, AttentionItem{
				Severity:     SeverityCritical,
				Emoji:        "🔥",
				Title:        pod.Name,
				Description:  fmt.Sprintf("%d ERR, %d WARN", errorCount, warnCount),
				Namespace:    namespace,
				Count:        errorCount,
				ResourceType: "pod",
				PodName:      pod.Name,
				Timestamp:    time.Now(),
			})
		} else if warnCount > 20 {
			// Report pods with many warnings but few errors
			items = append(items, AttentionItem{
				Severity:     SeverityWarning,
				Emoji:        "⚠️",
				Title:        pod.Name,
				Description:  fmt.Sprintf("%d WARN, %d ERR", warnCount, errorCount),
				Namespace:    namespace,
				Count:        warnCount,
				ResourceType: "pod",
				PodName:      pod.Name,
				Timestamp:    time.Now(),
			})
		}
	}

	return items
}

// NOTE: isErrorLog and isWarnLog are defined in app.go and reused here (same package)

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

// sortItemsByCount sorts items by total error+warning count (descending)
func sortItemsByCount(items []AttentionItem) {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].Count < items[j].Count {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

// sortItemsByName sorts items alphabetically by title
func sortItemsByName(items []AttentionItem) {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].Title > items[j].Title {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

// GetSortedAttentionItems returns items sorted by the specified mode
func GetSortedAttentionItems(items []AttentionItem, mode SortMode) []AttentionItem {
	// Make a copy to avoid modifying original
	sorted := make([]AttentionItem, len(items))
	copy(sorted, items)

	switch mode {
	case SortByCount:
		sortItemsByCount(sorted)
	case SortBySeverity:
		sortAttentionItems(sorted)
	case SortByName:
		sortItemsByName(sorted)
	}

	return sorted
}

// SortPodsByCount sorts pods by their E/W count (descending)
// Note: This is a simplified version that will be enhanced with actual counts from cache
func SortPodsByCount(pods []rancher.Pod, counts map[string]PodCounts) []rancher.Pod {
	// Make a copy to avoid modifying original
	sorted := make([]rancher.Pod, len(pods))
	copy(sorted, pods)

	// Bubble sort by count
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			// Extract namespace from NamespaceID
			ns1 := extractNamespace(sorted[i].NamespaceID)
			ns2 := extractNamespace(sorted[j].NamespaceID)

			key1 := ns1 + "/" + sorted[i].Name
			key2 := ns2 + "/" + sorted[j].Name

			count1 := counts[key1].Total
			count2 := counts[key2].Total

			// Sort descending (highest count first)
			if count1 < count2 {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// SortPodsByName sorts pods alphabetically by name
func SortPodsByName(pods []rancher.Pod) []rancher.Pod {
	sorted := make([]rancher.Pod, len(pods))
	copy(sorted, pods)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	return sorted
}

// populatePodCounts scans all pods and populates the error/warning count cache
// This is called before sorting to ensure the cache is fresh
func (a *App) populatePodCounts() {
	if a.dataSource == nil {
		return
	}

	// Get scan depth from config
	scanDepth := a.config.ScanDepth
	if scanDepth <= 0 {
		scanDepth = 200
	}

	// Scan each pod and cache counts
	for _, pod := range a.pods {
		// Extract namespace name
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

		// Create cache key
		key := namespaceName + "/" + pod.Name

		// Skip if already cached (performance optimization)
		if _, exists := a.cachedPodCounts[key]; exists {
			continue
		}

		// Fetch and scan logs
		logs, err := a.dataSource.GetLogs("", namespaceName, pod.Name, "", false)
		if err != nil || len(logs) == 0 {
			continue
		}

		// Limit scan to first N lines
		scanLines := logs
		if len(scanLines) > scanDepth {
			scanLines = scanLines[:scanDepth]
		}

		// Count errors and warnings
		errorCount := 0
		warnCount := 0
		for _, line := range scanLines {
			if isErrorLog(line) {
				errorCount++
			} else if isWarnLog(line) {
				warnCount++
			}
		}

		// Store in cache
		a.cachedPodCounts[key] = PodCounts{
			Errors:   errorCount,
			Warnings: warnCount,
			Total:    errorCount + warnCount,
		}
	}
}

// SortPodsBySeverity sorts pods by state severity (errors first, then warnings, then ok)
func SortPodsBySeverity(pods []rancher.Pod) []rancher.Pod {
	// Make a copy to avoid modifying original
	sorted := make([]rancher.Pod, len(pods))
	copy(sorted, pods)

	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			severity1 := getPodStateSeverity(sorted[i].State)
			severity2 := getPodStateSeverity(sorted[j].State)

			// Sort by severity (critical first)
			if severity1 > severity2 {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// NamespaceHealth tracks error/warning counts for a namespace
type NamespaceHealth struct {
	Errors   int
	Warnings int
	Total    int
}

// ComputeNamespaceHealth scans all pods and returns health stats per namespace
func ComputeNamespaceHealth(ds datasource.DataSource, scanDepth int) map[string]NamespaceHealth {
	health := make(map[string]NamespaceHealth)

	if scanDepth <= 0 {
		scanDepth = 200
	}

	// Get all pods across all namespaces
	pods, err := ds.GetAllPods()
	if err != nil {
		return health
	}

	// Scan each pod and aggregate by namespace
	for _, pod := range pods {
		// Extract namespace
		namespace := extractNamespace(pod.NamespaceID)
		if namespace == "" {
			namespace = "default"
		}

		// Try to get logs for this pod
		logs, err := ds.GetLogs("", namespace, pod.Name, "", false)
		if err != nil || len(logs) == 0 {
			continue
		}

		// Limit scan to first N lines
		scanLines := logs
		if len(scanLines) > scanDepth {
			scanLines = scanLines[:scanDepth]
		}

		// Count errors and warnings
		errorCount := 0
		warnCount := 0
		for _, line := range scanLines {
			if isErrorLog(line) {
				errorCount++
			} else if isWarnLog(line) {
				warnCount++
			}
		}

		// Accumulate to namespace total
		stats := health[namespace]
		stats.Errors += errorCount
		stats.Warnings += warnCount
		stats.Total += errorCount + warnCount
		health[namespace] = stats
	}

	return health
}

// extractNamespace extracts namespace from NamespaceID (handles "cluster:namespace" format)
func extractNamespace(namespaceID string) string {
	if strings.Contains(namespaceID, ":") {
		parts := strings.Split(namespaceID, ":")
		if len(parts) > 1 {
			return parts[1]
		}
	}
	return namespaceID
}

// getPodStateSeverity assigns severity score to pod states (lower = more critical)
func getPodStateSeverity(state string) int {
	stateLower := strings.ToLower(state)

	// Critical states (0-9)
	if strings.Contains(stateLower, "crash") || strings.Contains(stateLower, "error") ||
		strings.Contains(stateLower, "failed") || strings.Contains(stateLower, "oom") {
		return 0
	}
	if strings.Contains(stateLower, "imagepull") || strings.Contains(stateLower, "evicted") {
		return 5
	}

	// Warning states (10-19)
	if strings.Contains(stateLower, "pending") || strings.Contains(stateLower, "unknown") {
		return 10
	}

	// Normal states (20+)
	if strings.Contains(stateLower, "running") {
		return 20
	}
	if strings.Contains(stateLower, "completed") || strings.Contains(stateLower, "succeeded") {
		return 25
	}

	// Unknown states
	return 30
}
