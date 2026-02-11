package datasource

import (
	"fmt"
	"time"

	"github.com/Rancheroo/r8s/internal/bundle"
	"github.com/Rancheroo/r8s/internal/rancher"
)

// GenerateSyntheticDemo creates a minimal valid bundle with synthetic data
// Replaces the embedded example-log-bundle to reduce repository size
func GenerateSyntheticDemo() *bundle.Bundle {
	now := time.Now()

	return &bundle.Bundle{
		Path:        "",
		ExtractPath: "",
		Manifest: &bundle.BundleManifest{
			NodeName:    "demo-node",
			CollectedAt: now,
			RKE2Version: "v1.28.0+rke2r1",
			K8sVersion:  "v1.28.0",
			FileCount:   150,
			TotalSize:   1024 * 1024 * 10, // 10MB
			BundleType:  string(bundle.FormatRKE2),
		},
		Pods:        generateSyntheticPodInfo(),
		LogFiles:    []bundle.LogFileInfo{},
		CRDs:        []interface{}{},
		Deployments: []interface{}{},
		Services:    []interface{}{},
		Namespaces:  []interface{}{},
		Events:      generateSyntheticRancherEvents(),
		Loaded:      true,
		Size:        1024 * 1024 * 10,
		IsTemporary: false,
		Health: &bundle.BundleHealth{
			TotalFiles:   10,
			FoundFiles:   10,
			DerivedFiles: 0,
			MissingFiles: []string{},
			Warnings:     []string{},
		},
	}
}

// generateSyntheticPodInfo creates pod info structs for the bundle
func generateSyntheticPodInfo() []bundle.PodInfo {
	return []bundle.PodInfo{
		{Namespace: "default", Name: "nginx-deployment-7c4c7b8f5-x2v9p", Containers: []string{"nginx"}, HasCurrentLogs: true},
		{Namespace: "production", Name: "frontend-app-5d9c8b7f4-k3m8n", Containers: []string{"frontend"}, HasCurrentLogs: true},
		{Namespace: "production", Name: "backend-api-6f8d9c7b5-p9q2r", Containers: []string{"api"}, HasCurrentLogs: true},
		{Namespace: "cache", Name: "redis-cache-0", Containers: []string{"redis"}, HasCurrentLogs: true},
		{Namespace: "database", Name: "postgres-db-0", Containers: []string{"postgres"}, HasCurrentLogs: true},
		{Namespace: "batch", Name: "worker-cronjob-27981234-ab12c", Containers: []string{"worker"}, HasCurrentLogs: true},
		{Namespace: "payments", Name: "payment-service-9a8b7c6d5-e4f3g", Containers: []string{"payment"}, HasCurrentLogs: true},
		{Namespace: "logging", Name: "elasticsearch-0", Containers: []string{"elasticsearch"}, HasCurrentLogs: true},
		{Namespace: "monitoring", Name: "prometheus-server-5c9d4b8f7-x1y2z", Containers: []string{"prometheus"}, HasCurrentLogs: true},
		{Namespace: "monitoring", Name: "grafana-7d8e9f0a1-b2c3d", Containers: []string{"grafana"}, HasCurrentLogs: true},
	}
}

// generateSyntheticRancherEvents creates rancher.Event structs
func generateSyntheticRancherEvents() []interface{} {
	now := time.Now()
	events := []rancher.Event{
		{Namespace: "production", Type: "Warning", Reason: "BackOff", Object: "pod/frontend-app-5d9c8b7f4-k3m8n", Message: "Back-off restarting failed container", Count: 42, FirstSeen: now.Add(-2 * time.Hour).Format(time.RFC3339), LastSeen: now.Format(time.RFC3339), Name: "event-1"},
		{Namespace: "production", Type: "Warning", Reason: "Failed", Object: "pod/frontend-app-5d9c8b7f4-k3m8n", Message: "Error: container has crashed 5 times", Count: 5, FirstSeen: now.Add(-2 * time.Hour).Format(time.RFC3339), LastSeen: now.Add(-95 * time.Minute).Format(time.RFC3339), Name: "event-2"},
		{Namespace: "production", Type: "Warning", Reason: "Failed", Object: "pod/backend-api-6f8d9c7b5-p9q2r", Message: "Failed to pull image \"backend:v2.1.0\": pull access denied", Count: 12, FirstSeen: now.Add(-110 * time.Minute).Format(time.RFC3339), LastSeen: now.Add(-105 * time.Minute).Format(time.RFC3339), Name: "event-3"},
		{Namespace: "production", Type: "Normal", Reason: "Pulling", Object: "pod/backend-api-6f8d9c7b5-p9q2r", Message: "Pulling image \"backend:v2.1.0\"", Count: 15, FirstSeen: now.Add(-110 * time.Minute).Format(time.RFC3339), LastSeen: now.Format(time.RFC3339), Name: "event-4"},
		{Namespace: "database", Type: "Warning", Reason: "OOMKilling", Object: "pod/postgres-db-0", Message: "Memory cgroup out of memory: Killed process 12345", Count: 5, FirstSeen: now.Add(-90 * time.Minute).Format(time.RFC3339), LastSeen: now.Format(time.RFC3339), Name: "event-5"},
		{Namespace: "payments", Type: "Warning", Reason: "FailedScheduling", Object: "pod/payment-service-9a8b7c6d5-e4f3g", Message: "0/2 nodes are available: 2 Insufficient memory", Count: 8, FirstSeen: now.Add(-100 * time.Minute).Format(time.RFC3339), LastSeen: now.Add(-70 * time.Minute).Format(time.RFC3339), Name: "event-6"},
		{Namespace: "default", Type: "Normal", Reason: "Created", Object: "pod/nginx-deployment-7c4c7b8f5-x2v9p", Message: "Created container nginx", Count: 1, FirstSeen: now.Add(-119 * time.Minute).Format(time.RFC3339), Name: "event-7"},
		{Namespace: "default", Type: "Normal", Reason: "Started", Object: "pod/nginx-deployment-7c4c7b8f5-x2v9p", Message: "Started container nginx", Count: 1, FirstSeen: now.Add(-118 * time.Minute).Format(time.RFC3339), Name: "event-8"},
		{Namespace: "production", Type: "Warning", Reason: "Unhealthy", Object: "pod/frontend-app-5d9c8b7f4-k3m8n", Message: "Liveness probe failed: HTTP probe failed with statuscode: 503", Count: 25, FirstSeen: now.Add(-110 * time.Minute).Format(time.RFC3339), LastSeen: now.Format(time.RFC3339), Name: "event-9"},
		{Namespace: "default", Type: "Normal", Reason: "ScalingReplicaSet", Object: "deployment/nginx-deployment", Message: "Scaled up replica set nginx-deployment-7c4c7b8f5 to 3", Count: 1, FirstSeen: now.Add(-115 * time.Minute).Format(time.RFC3339), Name: "event-10"},
	}

	// Convert to []interface{}
	result := make([]interface{}, len(events))
	for i := range events {
		result[i] = events[i]
	}
	return result
}

// generateSyntheticLogs creates realistic-looking logs
func generateSyntheticLogs(podName, namespace, containerName string) string {
	return fmt.Sprintf(`[INFO] 2026-02-10T03:00:00Z Starting container %s
[INFO] 2026-02-10T03:00:01Z Application initializing...
[INFO] 2026-02-10T03:00:02Z Connected to database
[INFO] 2026-02-10T03:00:03Z Server listening on :8080
[INFO] 2026-02-10T03:00:04Z Ready to serve requests
[WARN] 2026-02-10T03:15:23Z High memory usage detected: 85%%
[INFO] 2026-02-10T03:30:45Z Processing request batch #1234
[INFO] 2026-02-10T04:00:12Z Health check passed
[WARN] 2026-02-10T04:15:33Z Response time degraded: 2.3s
[INFO] 2026-02-10T04:30:00Z Garbage collection completed
[INFO] 2026-02-10T05:00:00Z Heartbeat: all systems operational`, containerName)
}
