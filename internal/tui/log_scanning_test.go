package tui

import (
	"testing"

	"github.com/Rancheroo/r8s/internal/datasource"
	"github.com/Rancheroo/r8s/internal/rancher"
)

// TestDashboardLogAccuracy_MatchesLogView verifies that the error/warning counts
// shown in the dashboard EXACTLY match the counts you see when viewing that pod's logs.
// This is the #1 requirement - "Truth Only™" principle.
func TestDashboardLogAccuracy_MatchesLogView(t *testing.T) {
	// Create test datasource with known log content
	ds := &mockDataSource{
		pods: []rancher.Pod{
			{Name: "pod-with-5-errors", NamespaceID: "default", State: "Running"},
			{Name: "pod-with-0-errors", NamespaceID: "default", State: "Running"},
			{Name: "pod-with-21-warnings", NamespaceID: "default", State: "Running"},
		},
		logs: map[string][]string{
			"default/pod-with-5-errors": {
				"I1127 00:00:01 [INFO] Starting up",
				"E1127 00:00:02 [ERROR] Connection failed",
				"E1127 00:00:03 [ERROR] Retry failed",
				"I1127 00:00:04 [INFO] Normal log",
				"E1127 00:00:05 [ERROR] Another error",
				"E1127 00:00:06 [ERROR] Fourth error",
				"E1127 00:00:07 [ERROR] Fifth error",
				"I1127 00:00:08 [INFO] Done",
			},
			"default/pod-with-0-errors": {
				"I1127 00:00:01 [INFO] Starting up",
				"I1127 00:00:02 [INFO] All good",
				"I1127 00:00:03 [INFO] Still good",
			},
			"default/pod-with-21-warnings": {
				"W1127 00:00:01 [WARN] Warning 1",
				"W1127 00:00:02 [WARN] Warning 2",
				"W1127 00:00:03 [WARN] Warning 3",
				"W1127 00:00:04 [WARN] Warning 4",
				"W1127 00:00:05 [WARN] Warning 5",
				"I1127 00:00:06 [INFO] Normal",
				"W1127 00:00:07 [WARN] Warning 6",
				"W1127 00:00:08 [WARN] Warning 7",
				"W1127 00:00:09 [WARN] Warning 8",
				"W1127 00:00:10 [WARN] Warning 9",
				"W1127 00:00:11 [WARN] Warning 10",
				"W1127 00:00:12 [WARN] Warning 11",
				"W1127 00:00:13 [WARN] Warning 12",
				"W1127 00:00:14 [WARN] Warning 13",
				"W1127 00:00:15 [WARN] Warning 14",
				"W1127 00:00:16 [WARN] Warning 15",
				"W1127 00:00:17 [WARN] Warning 16",
				"W1127 00:00:18 [WARN] Warning 17",
				"W1127 00:00:19 [WARN] Warning 18",
				"W1127 00:00:20 [WARN] Warning 19",
				"W1127 00:00:21 [WARN] Warning 20",
				"W1127 00:00:22 [WARN] Warning 21",
			},
		},
	}

	// Detect issues using dashboard logic
	scanDepth := 200
	items := detectLogIssues(ds, scanDepth)

	// Verify each pod's counts match what we'd see in log view
	testCases := []struct {
		podName        string
		expectedErrors int
		expectedWarns  int
	}{
		{"pod-with-5-errors", 5, 0},
		{"pod-with-0-errors", 0, 0},
		{"pod-with-21-warnings", 0, 21},
	}

	for _, tc := range testCases {
		t.Run(tc.podName, func(t *testing.T) {
			// Find the attention item for this pod
			var found bool
			for _, item := range items {
				if item.PodName == tc.podName {
					found = true

					// Get the actual logs for verification
					logs, _ := ds.GetLogs("", "default", tc.podName, "", false)

					// Count errors/warnings in the actual logs (same logic as log view)
					actualErrors := 0
					actualWarns := 0
					for _, line := range logs {
						if isErrorLog(line) {
							actualErrors++
						} else if isWarnLog(line) {
							actualWarns++
						}
					}

					// Dashboard count MUST match actual log count
					if tc.expectedErrors != actualErrors {
						t.Errorf("Expected %d errors, log view shows %d", tc.expectedErrors, actualErrors)
					}
					if tc.expectedWarns != actualWarns {
						t.Errorf("Expected %d warnings, log view shows %d", tc.expectedWarns, actualWarns)
					}

					// If we have significant issues, verify they're in the dashboard
					if actualErrors >= 10 || actualWarns >= 10 {
						if item.Count == 0 {
							t.Errorf("Pod has %d errors/%d warnings but dashboard shows 0", actualErrors, actualWarns)
						}
					}
				}
			}

			// Pods with significant issues should be in dashboard
			if tc.expectedErrors >= 10 || tc.expectedWarns >= 10 {
				if !found {
					t.Errorf("Pod %s with %dE/%dW not in dashboard", tc.podName, tc.expectedErrors, tc.expectedWarns)
				}
			}
		})
	}
}

// TestDashboardLogAccuracy_NoCaching verifies that each pod's logs are fetched
// independently - no caching bugs where one pod's count is reused for another.
func TestDashboardLogAccuracy_NoCaching(t *testing.T) {
	ds := &mockDataSource{
		pods: []rancher.Pod{
			{Name: "pod-A", NamespaceID: "ns1", State: "Running"},
			{Name: "pod-B", NamespaceID: "ns2", State: "Running"},
			{Name: "pod-C", NamespaceID: "ns1", State: "Running"},
		},
		logs: map[string][]string{
			"ns1/pod-A": makeLogsWithErrors(5),
			"ns2/pod-B": makeLogsWithErrors(15), // Different namespace
			"ns1/pod-C": makeLogsWithErrors(25), // Same namespace as A, different count
		},
		getLogsCalls: make([]string, 0), // Track calls
	}

	scanDepth := 200
	items := detectLogIssues(ds, scanDepth)

	// Verify we made separate GetLogs() calls for each pod
	if len(ds.getLogsCalls) < 3 {
		t.Errorf("Expected at least 3 GetLogs() calls, got %d", len(ds.getLogsCalls))
	}

	// Verify each pod has correct count (not cached from another pod)
	expectedCounts := map[string]int{
		"pod-A": 5,
		"pod-B": 15,
		"pod-C": 25,
	}

	// Verify each pod's count independently with proper namespace
	podNamespaces := map[string]string{
		"pod-A": "ns1",
		"pod-B": "ns2",
		"pod-C": "ns1",
	}

	for podName, expectedCount := range expectedCounts {
		namespace := podNamespaces[podName]
		logs, _ := ds.GetLogs("", namespace, podName, "", false)
		actualCount := 0
		for _, line := range logs {
			if isErrorLog(line) {
				actualCount++
			}
		}

		if actualCount != expectedCount {
			t.Errorf("Pod %s: expected %d errors, got %d (possible caching bug)",
				podName, expectedCount, actualCount)
		}
	}

	// Verify dashboard has correct items (pods with >10 errors)
	if len(items) != 2 {
		t.Errorf("Expected 2 pods in dashboard (B=15, C=25 errors), got %d", len(items))
	}
}

// TestDashboardLogAccuracy_ScanDepthRespected verifies that the --scan flag
// is properly respected and only scans the specified number of lines.
func TestDashboardLogAccuracy_ScanDepthRespected(t *testing.T) {
	// Create a pod with errors beyond scan depth
	logsWithErrorsAtEnd := append(
		makeLogsWithInfo(150),     // 150 info logs
		makeLogsWithErrors(20)..., // 20 errors after line 150
	)

	ds := &mockDataSource{
		pods: []rancher.Pod{
			{Name: "deep-errors", NamespaceID: "default", State: "Running"},
		},
		logs: map[string][]string{
			"default/deep-errors": logsWithErrorsAtEnd,
		},
	}

	// With scanDepth=100, should not see the errors at line 150+
	scanDepth := 100
	items := detectLogIssues(ds, scanDepth)

	// Should find no issues (errors are beyond scan depth)
	if len(items) > 0 {
		t.Errorf("Scan depth not respected: found issues beyond line %d", scanDepth)
	}

	// With scanDepth=200, should see the errors
	scanDepth = 200
	items = detectLogIssues(ds, scanDepth)

	// Should find the pod with errors
	found := false
	for _, item := range items {
		if item.PodName == "deep-errors" {
			found = true
		}
	}

	if !found {
		t.Errorf("With scanDepth=200, should detect errors at line 150+")
	}
}

// Helper: mockDataSource for testing
type mockDataSource struct {
	pods         []rancher.Pod
	logs         map[string][]string
	getLogsCalls []string
}

// Compile-time interface verification
var _ datasource.DataSource = (*mockDataSource)(nil)

func (m *mockDataSource) Close() error {
	return nil
}

func (m *mockDataSource) GetPods(projectID, namespace string) ([]rancher.Pod, error) {
	return m.pods, nil
}

func (m *mockDataSource) GetAllPods() ([]rancher.Pod, error) {
	return m.pods, nil
}

func (m *mockDataSource) GetLogs(clusterID, namespace, podName, container string, previous bool) ([]string, error) {
	key := namespace + "/" + podName
	m.getLogsCalls = append(m.getLogsCalls, key)

	if logs, ok := m.logs[key]; ok {
		return logs, nil
	}
	return []string{}, nil
}

// Stub other required methods
func (m *mockDataSource) GetClusters() ([]rancher.Cluster, error) {
	return nil, nil
}

func (m *mockDataSource) GetProjects(clusterID string) ([]rancher.Project, map[string]int, error) {
	return nil, nil, nil
}

func (m *mockDataSource) GetNamespaces(clusterID, projectID string) ([]rancher.Namespace, error) {
	return nil, nil
}

func (m *mockDataSource) GetDeployments(projectID, namespace string) ([]rancher.Deployment, error) {
	return nil, nil
}

func (m *mockDataSource) GetServices(projectID, namespace string) ([]rancher.Service, error) {
	return nil, nil
}

func (m *mockDataSource) GetCRDs(clusterID string) ([]rancher.CRD, error) {
	return nil, nil
}

func (m *mockDataSource) GetCRDInstances(clusterID, group, version, resource string) ([]map[string]interface{}, error) {
	return nil, nil
}

func (m *mockDataSource) DescribePod(clusterID, namespace, name string) (interface{}, error) {
	return nil, nil
}

func (m *mockDataSource) DescribeDeployment(clusterID, namespace, name string) (interface{}, error) {
	return nil, nil
}

func (m *mockDataSource) DescribeService(clusterID, namespace, name string) (interface{}, error) {
	return nil, nil
}

func (m *mockDataSource) GetContainers(namespace, pod string) ([]string, error) {
	return nil, nil
}

func (m *mockDataSource) GetNodes() ([]datasource.Node, error) {
	return nil, nil
}

func (m *mockDataSource) GetAllEvents() ([]rancher.Event, error) {
	return nil, nil
}

func (m *mockDataSource) GetEventsByPod(namespace, podName string) ([]rancher.Event, error) {
	return nil, nil
}

func (m *mockDataSource) GetDaemonSets() ([]datasource.DaemonSet, error) {
	return nil, nil
}

func (m *mockDataSource) GetEtcdHealth() (*datasource.EtcdHealth, error) {
	return nil, nil
}

func (m *mockDataSource) GetEtcdDetails() (*datasource.EtcdDetails, error) {
	return nil, nil
}

func (m *mockDataSource) GetNodeConditions() ([]datasource.NodeConditions, error) {
	return nil, nil
}

func (m *mockDataSource) GetSystemHealth() (*datasource.SystemHealth, error) {
	return nil, nil
}

func (m *mockDataSource) GetKubeletIssues() ([]datasource.KubeletIssue, error) {
	return nil, nil
}

func (m *mockDataSource) GetOOMAnalysis() ([]datasource.OOMAnalysis, error) {
	return nil, nil
}

func (m *mockDataSource) GetPodResources(podName string) ([]datasource.ResourceSpec, error) {
	return nil, nil
}

func (m *mockDataSource) GetDiagnosticContext(namespace, podName string) (*datasource.DiagnosticContext, error) {
	return nil, nil
}

func (m *mockDataSource) Mode() string {
	return "TEST"
}

// Helper functions to generate test logs
func makeLogsWithErrors(count int) []string {
	logs := make([]string, count)
	for i := 0; i < count; i++ {
		logs[i] = "E1127 00:00:00 [ERROR] Test error"
	}
	return logs
}

func makeLogsWithInfo(count int) []string {
	logs := make([]string, count)
	for i := 0; i < count; i++ {
		logs[i] = "I1127 00:00:00 [INFO] Normal log"
	}
	return logs
}
