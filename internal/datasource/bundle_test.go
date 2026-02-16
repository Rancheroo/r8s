package datasource

import (
	"os"
	"path/filepath"
	"testing"
)

// createTestBundleDir creates a minimal valid bundle structure for testing
func createTestBundleDir(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "r8s-datasource-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create minimal bundle structure
	rke2Dir := filepath.Join(tmpDir, "rke2")
	kubectlDir := filepath.Join(rke2Dir, "kubectl")
	serverDir := filepath.Join(rke2Dir, "server")

	os.MkdirAll(kubectlDir, 0755)
	os.MkdirAll(serverDir, 0755)

	// Create a minimal nodes file
	nodesContent := `NAME           STATUS   ROLES                       AGE   VERSION
node1          Ready    control-plane,etcd,master   5d    v1.28.0
`
	os.WriteFile(filepath.Join(kubectlDir, "nodes"), []byte(nodesContent), 0644)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestNewBundleDataSource(t *testing.T) {
	t.Run("invalid path", func(t *testing.T) {
		ds, err := NewBundleDataSource("/nonexistent/path", false)
		if err == nil {
			t.Error("Expected error for nonexistent path, got nil")
		}
		if ds != nil {
			t.Error("Expected nil datasource for error case")
		}
	})

	t.Run("valid bundle directory", func(t *testing.T) {
		bundleDir, cleanup := createTestBundleDir(t)
		defer cleanup()

		ds, err := NewBundleDataSource(bundleDir, false)
		if err != nil {
			t.Fatalf("Failed to create datasource: %v", err)
		}
		if ds == nil {
			t.Fatal("Expected non-nil datasource")
		}
		if ds.bundle == nil {
			t.Fatal("Expected bundle to be loaded")
		}
	})
}

func TestBundleDataSource_GetClusters(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	clusters, err := ds.GetClusters()
	if err != nil {
		t.Errorf("GetClusters() error = %v", err)
	}

	if len(clusters) != 1 {
		t.Errorf("Expected 1 cluster, got %d", len(clusters))
	}

	if clusters[0].ID != "bundle-cluster" {
		t.Errorf("Expected cluster ID 'bundle-cluster', got %s", clusters[0].ID)
	}
}

func TestBundleDataSource_GetProjects(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	projects, counts, err := ds.GetProjects("bundle-cluster")
	if err != nil {
		t.Errorf("GetProjects() error = %v", err)
	}

	// Should return at least default project
	if len(projects) == 0 {
		t.Error("Expected at least one project")
	}

	if counts == nil {
		t.Error("Expected namespace counts map")
	}
}

func TestBundleDataSource_GetNamespaces(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	namespaces, err := ds.GetNamespaces("bundle-cluster", "")
	if err != nil {
		t.Errorf("GetNamespaces() error = %v", err)
	}

	// May be empty but shouldn't error
	t.Logf("Got %d namespaces", len(namespaces))
}

func TestBundleDataSource_GetPods(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	pods, err := ds.GetPods("default", "")
	if err != nil {
		t.Errorf("GetPods() error = %v", err)
	}

	t.Logf("Got %d pods", len(pods))
}

func TestBundleDataSource_GetAllPods(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	pods, err := ds.GetAllPods()
	if err != nil {
		// Error expected for minimal bundle without pods file
		t.Logf("GetAllPods() error (expected): %v", err)
	}

	t.Logf("Got %d pods", len(pods))
}

func TestBundleDataSource_GetNodes(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	nodes, err := ds.GetNodes()
	if err != nil {
		t.Errorf("GetNodes() error = %v", err)
	}

	// May be empty if parsing fails, but shouldn't error
	t.Logf("Got %d nodes", len(nodes))
}

func TestBundleDataSource_GetAllEvents(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	events, err := ds.GetAllEvents()
	if err != nil {
		t.Errorf("GetAllEvents() error = %v", err)
	}

	t.Logf("Got %d events", len(events))
}

func TestBundleDataSource_GetEventsByPod(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	events, err := ds.GetEventsByPod("default", "nonexistent-pod")
	if err != nil {
		t.Errorf("GetEventsByPod() error = %v", err)
	}

	t.Logf("Got %d events for nonexistent pod", len(events))
}

func TestBundleDataSource_GetDeployments(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	deployments, err := ds.GetDeployments("default", "")
	if err != nil {
		t.Errorf("GetDeployments() error = %v", err)
	}

	t.Logf("Got %d deployments", len(deployments))
}

func TestBundleDataSource_GetServices(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	services, err := ds.GetServices("default", "")
	if err != nil {
		t.Errorf("GetServices() error = %v", err)
	}

	t.Logf("Got %d services", len(services))
}

func TestBundleDataSource_GetCRDs(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	crds, err := ds.GetCRDs("bundle-cluster")
	if err != nil {
		t.Errorf("GetCRDs() error = %v", err)
	}

	t.Logf("Got %d CRDs", len(crds))
}

func TestBundleDataSource_GetDaemonSets(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	daemonsets, err := ds.GetDaemonSets()
	if err != nil {
		t.Errorf("GetDaemonSets() error = %v", err)
	}

	t.Logf("Got %d daemonsets", len(daemonsets))
}

func TestBundleDataSource_GetContainers(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	containers, err := ds.GetContainers("default", "nonexistent-pod")
	if err != nil {
		t.Errorf("GetContainers() error = %v", err)
	}

	t.Logf("Got %d containers", len(containers))
}

func TestBundleDataSource_GetLogs(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	logs, err := ds.GetLogs("bundle-cluster", "default", "nonexistent-pod", "container", false)
	if err != nil {
		t.Logf("GetLogs() returned error (expected for missing pod): %v", err)
	}

	t.Logf("Got %d log lines", len(logs))
}

func TestBundleDataSource_DescribePod(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	desc, err := ds.DescribePod("bundle-cluster", "default", "nonexistent-pod")
	if err != nil {
		t.Logf("DescribePod() returned error (expected): %v", err)
	}

	t.Logf("Got description: %v", desc)
}

func TestBundleDataSource_GetEtcdHealth(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	health, err := ds.GetEtcdHealth()
	if err != nil {
		t.Logf("GetEtcdHealth() returned error: %v", err)
	}

	t.Logf("Got etcd health: %v", health)
}

func TestBundleDataSource_GetNodeConditions(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	conditions, err := ds.GetNodeConditions()
	if err != nil {
		t.Logf("GetNodeConditions() returned error: %v", err)
	}

	t.Logf("Got %d node conditions", len(conditions))
}

func TestBundleDataSource_GetSystemHealth(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	health, err := ds.GetSystemHealth()
	if err != nil {
		t.Logf("GetSystemHealth() returned error: %v", err)
	}

	t.Logf("Got system health: %v", health)
}

func TestBundleDataSource_GetKubeletIssues(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	issues, err := ds.GetKubeletIssues()
	if err != nil {
		t.Logf("GetKubeletIssues() returned error: %v", err)
	}

	t.Logf("Got %d kubelet issues", len(issues))
}

func TestBundleDataSource_Close(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	err = ds.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestBundleDataSource_MultipleCalls(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	// Multiple calls should work consistently
	for i := 0; i < 3; i++ {
		clusters, err := ds.GetClusters()
		if err != nil {
			t.Errorf("GetClusters() call %d error = %v", i, err)
		}
		if len(clusters) != 1 {
			t.Errorf("GetClusters() call %d: expected 1 cluster, got %d", i, len(clusters))
		}
	}
}

func TestBundleDataSource_Caching(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	// First call should cache
	pods1, err := ds.getKubectlPods()
	if err != nil {
		t.Logf("getKubectlPods() error (may be expected): %v", err)
	}

	// Second call should use cache
	pods2, err := ds.getKubectlPods()
	if err != nil {
		t.Logf("getKubectlPods() second call error: %v", err)
	}

	// Should return same results
	if len(pods1) != len(pods2) {
		t.Errorf("Caching inconsistent: first call got %d, second got %d", len(pods1), len(pods2))
	}
}

func TestBundleDataSource_GetCRDInstances(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	// GetCRDInstances returns empty list for bundle mode
	instances, err := ds.GetCRDInstances("cluster1", "group", "v1", "widgets")
	if err != nil {
		t.Errorf("GetCRDInstances() error = %v", err)
	}
	if len(instances) != 0 {
		t.Errorf("GetCRDInstances() expected empty list, got %d items", len(instances))
	}
}

func TestBundleDataSource_DescribeDeployment(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	desc, err := ds.DescribeDeployment("bundle-cluster", "default", "nonexistent-deployment")
	if err != nil {
		t.Logf("DescribeDeployment() error (expected): %v", err)
	}

	t.Logf("Got description: %v", desc)
}

func TestBundleDataSource_DescribeService(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	desc, err := ds.DescribeService("bundle-cluster", "default", "nonexistent-service")
	if err != nil {
		t.Logf("DescribeService() error (expected): %v", err)
	}

	t.Logf("Got description: %v", desc)
}

func TestBundleDataSource_GetEtcdDetails(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	details, err := ds.GetEtcdDetails()
	if err != nil {
		t.Logf("GetEtcdDetails() error: %v", err)
	}

	t.Logf("Got etcd details: %v", details)
}

func TestBundleDataSource_GetPodResources(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	resources, err := ds.GetPodResources("nonexistent-pod")
	if err != nil {
		t.Logf("GetPodResources() error: %v", err)
	}

	t.Logf("Got %d resource specs", len(resources))
}

func TestBundleDataSource_GetDiagnosticContext(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	ctx, err := ds.GetDiagnosticContext("default", "nonexistent-pod")
	if err != nil {
		t.Logf("GetDiagnosticContext() error: %v", err)
	}

	t.Logf("Got diagnostic context: %v", ctx)
}

func TestBundleDataSource_GetBundleHealth(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	health := ds.GetBundleHealth()
	t.Logf("Got bundle health: %+v", health)
}

func TestBundleDataSource_Mode(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	mode := ds.Mode()
	if mode != "BUNDLE" {
		t.Errorf("Mode() = %v, want BUNDLE", mode)
	}
}

func TestBundleDataSource_GetOOMAnalysis(t *testing.T) {
	bundleDir, cleanup := createTestBundleDir(t)
	defer cleanup()

	ds, err := NewBundleDataSource(bundleDir, false)
	if err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	analysis, err := ds.GetOOMAnalysis()
	if err != nil {
		t.Logf("GetOOMAnalysis() error: %v", err)
	}

	t.Logf("Got %d OOM analyses", len(analysis))
}
