package tui

import (
	"fmt"
	"strings"

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
// Enhanced in v0.6.6 with actual OOM analysis data
func generateOOMContext(ds datasource.DataSource, namespace, podName string) *datasource.DiagnosticContext {
	oomEvents, _ := ds.GetOOMAnalysis()

	// Find OOM event for this pod
	var matchedOOM *datasource.OOMAnalysis
	for i := range oomEvents {
		// Handle both "namespace/podname" and just "podname" formats
		oomPodName := oomEvents[i].PodName
		if strings.Contains(oomPodName, "/") {
			parts := strings.Split(oomPodName, "/")
			if len(parts) == 2 {
				oomPodName = parts[1]
			}
		}

		if oomPodName == podName {
			matchedOOM = &oomEvents[i]
			break
		}
	}

	// Default context if no specific OOM data
	if matchedOOM == nil {
		return &datasource.DiagnosticContext{
			Severity:       "critical",
			FixPriority:    "immediate",
			RootCause:      "Container OOMKilled",
			Recommendation: "Review pod resource limits. Check rke2/kubectl/events for memory details.",
			RelatedData:    []string{"OOM analysis data not available in bundle"},
		}
	}

	// Build context from actual OOM data
	ctx := &datasource.DiagnosticContext{
		Severity:    "critical",
		FixPriority: "immediate",
	}

	// Determine root cause
	if matchedOOM.IsNodeOOM {
		ctx.RootCause = "Node memory exhausted - system OOM"
		ctx.Recommendation = "Investigate node memory pressure. Check node describe for memory allocatable."
		ctx.RelatedData = []string{
			"Node-level OOM (not container limit)",
			"Check: kubectl describe nodes",
			"Review: rke2/kubectl/nodesdescribe in bundle",
		}
	} else if matchedOOM.MemoryLimit != "" {
		ctx.RootCause = fmt.Sprintf("Container exceeded memory limit: %s", matchedOOM.MemoryLimit)

		// Build recommendation based on limit
		if matchedOOM.MemoryRequest != "" {
			ctx.Recommendation = fmt.Sprintf("Consider increasing limit from %s (request: %s). Monitor actual usage first.",
				matchedOOM.MemoryLimit, matchedOOM.MemoryRequest)
		} else {
			ctx.Recommendation = fmt.Sprintf("Increase memory limit above %s or optimize application", matchedOOM.MemoryLimit)
		}

		ctx.RelatedData = []string{
			fmt.Sprintf("Limit: %s", matchedOOM.MemoryLimit),
		}
		if matchedOOM.MemoryRequest != "" {
			ctx.RelatedData = append(ctx.RelatedData, fmt.Sprintf("Request: %s", matchedOOM.MemoryRequest))
		}
		if matchedOOM.ContainerName != "" {
			ctx.RelatedData = append(ctx.RelatedData, fmt.Sprintf("Container: %s", matchedOOM.ContainerName))
		}
	} else {
		ctx.RootCause = "Container OOMKilled (limit not found in bundle)"
		ctx.Recommendation = "Check pod spec for memory limits. Review rke2/pod-manifests/ or kubectl describe pod."
		ctx.RelatedData = []string{"Memory limit data not parsed from bundle"}
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

