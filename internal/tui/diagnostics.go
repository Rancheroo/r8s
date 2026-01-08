package tui

import (
	"fmt"
	"github.com/Rancheroo/r8s/internal/datasource"
)

// generateCrashLoopContext creates diagnostic context for CrashLoopBackOff pods
func generateCrashLoopContext(ds datasource.DataSource, namespace, podName string) *datasource.DiagnosticContext {
	events, _ := ds.GetEventsByPod(namespace, podName)

	ctx := &datasource.DiagnosticContext{
		Severity:       "high",
		FixPriority:    "investigate",
		RootCause:      "Container repeatedly failing to start",
		Recommendation: "Check pod events for startup errors, review container logs, verify image availability",
		RelatedData:    []string{"State: CrashLoopBackOff", fmt.Sprintf("Events: %d recent", len(events))},
	}

	return ctx
}

// generateOOMContext creates diagnostic context for OOM kills
func generateOOMContext(ds datasource.DataSource, namespace, podName string) *datasource.DiagnosticContext {
	oom, _ := ds.GetOOMAnalysis()
	resources, _ := ds.GetPodResources(podName)

	ctx := &datasource.DiagnosticContext{
		Severity:       "critical",
		FixPriority:    "immediate",
		RootCause:      "Container exceeded memory limit",
		Recommendation: "Increase MemoryLimit or optimize application memory usage. Check for memory leaks.",
		RelatedData:    []string{fmt.Sprintf("OOM Events: %d", len(oom)), fmt.Sprintf("Resources: %d specs", len(resources))},
	}

	return ctx
}

// generateImagePullContext creates diagnostic context for image pull failures
func generateImagePullContext(ds datasource.DataSource, namespace, podName string) *datasource.DiagnosticContext {
	events, _ := ds.GetEventsByPod(namespace, podName)

	ctx := &datasource.DiagnosticContext{
		Severity:       "high",
		FixPriority:    "immediate",
		RootCause:      "Cannot pull container image from registry",
		Recommendation: "Verify registry access, check image name/tag, test with docker pull, check network policies",
		RelatedData:    []string{"Typical Events: ImagePullBackOff", fmt.Sprintf("Recent Events: %d", len(events))},
	}

	return ctx
}

// generateNodeContext creates diagnostic context for node issues
func generateNodeContext(ds datasource.DataSource) *datasource.DiagnosticContext {
	nodes, _ := ds.GetNodeConditions()

	pressure := false
	for _, node := range nodes {
		if node.HasPressure {
			pressure = true
			break
		}
	}

	ctx := &datasource.DiagnosticContext{
		Severity:       "critical",
		FixPriority:    "immediate",
		RootCause:      "Node resource pressure detected",
		Recommendation: "Check node conditions (MemoryPressure, DiskPressure, PIDPressure). Evacuate pods or add capacity.",
		RelatedData:    []string{fmt.Sprintf("Nodes with Pressure: %d/%d", countPressureNodes(nodes), len(nodes))},
	}

	if !pressure {
		ctx.Severity = "low"
		ctx.RootCause = "Nodes healthy"
		ctx.Recommendation = "All nodes reporting ready"
	}

	return ctx
}

// generateEtcdContext creates diagnostic context for etcd issues
func generateEtcdContext(ds datasource.DataSource) *datasource.DiagnosticContext {
	etcd, _ := ds.GetEtcdDetails()

	ctx := &datasource.DiagnosticContext{
		Severity:       "critical",
		FixPriority:    "immediate",
		RootCause:      "ETCD cluster unhealthy",
		Recommendation: "Check member health, alarms, and compaction status. Verify disk space and network connectivity.",
		RelatedData:    []string{fmt.Sprintf("Members: %d", etcd.MemberCount), fmt.Sprintf("DB Size: %s", etcd.DBSize)},
	}

	if etcd.Healthy && !etcd.NeedsCompaction {
		ctx.Severity = "low"
		ctx.RootCause = "ETCD healthy"
		ctx.Recommendation = "Cluster stable"
	} else if etcd.NeedsCompaction {
		ctx.Severity = "medium"
		ctx.FixPriority = "investigate"
		ctx.RootCause = "ETCD database needs compaction"
		ctx.Recommendation = etcd.CompactionReason
	}

	return ctx
}

// generateKubeletContext creates diagnostic context for kubelet issues
func generateKubeletContext(ds datasource.DataSource) *datasource.DiagnosticContext {
	issues, _ := ds.GetKubeletIssues()

	if len(issues) == 0 {
		return &datasource.DiagnosticContext{
			Severity:       "low",
			RootCause:      "No kubelet issues detected",
			Recommendation: "Node-level kubelet logs clean",
		}
	}

	ctx := &datasource.DiagnosticContext{
		Severity:       "high",
		FixPriority:    "investigate",
		RootCause:      "Kubelet reporting node-level errors",
		Recommendation: "Review kubelet issues for connectivity, timeout, or remotedialer problems",
		RelatedData:    []string{fmt.Sprintf("Top Issue: %s (%d occurrences)", issues[0].Pattern, issues[0].Count)},
	}

	return ctx
}

func countPressureNodes(nodes []datasource.NodeConditions) int {
	count := 0
	for _, node := range nodes {
		if node.HasPressure {
			count++
		}
	}
	return count
}
