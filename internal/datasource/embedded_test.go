package datasource

import (
	"testing"
)

func TestNewEmbeddedDataSource(t *testing.T) {
	t.Run("create with verbose false", func(t *testing.T) {
		ds, err := NewEmbeddedDataSource(false)
		if err != nil {
			t.Fatalf("NewEmbeddedDataSource() error = %v", err)
		}
		if ds == nil {
			t.Fatal("Expected non-nil datasource")
		}
	})

	t.Run("create with verbose true", func(t *testing.T) {
		ds, err := NewEmbeddedDataSource(true)
		if err != nil {
			t.Fatalf("NewEmbeddedDataSource() error = %v", err)
		}
		if ds == nil {
			t.Fatal("Expected non-nil datasource")
		}
	})
}

func TestSyntheticDataSource_GetClusters(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	clusters, err := ds.GetClusters()
	if err != nil {
		t.Errorf("GetClusters() error = %v", err)
	}
	if len(clusters) != 1 {
		t.Errorf("Expected 1 cluster, got %d", len(clusters))
	}
	if clusters[0].Name != "demo-cluster" {
		t.Errorf("Expected cluster name 'demo-cluster', got %s", clusters[0].Name)
	}
}

func TestSyntheticDataSource_GetProjects(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	projects, counts, err := ds.GetProjects("demo-cluster")
	if err != nil {
		t.Errorf("GetProjects() error = %v", err)
	}
	if len(projects) == 0 {
		t.Error("Expected projects, got none")
	}
	if len(counts) == 0 {
		t.Error("Expected namespace counts, got none")
	}
}

func TestSyntheticDataSource_GetNamespaces(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	tests := []struct {
		projectID string
		wantEmpty bool
	}{
		{"p-default", false},
		{"p-production", false},
		{"p-monitoring", false},
		{"p-nonexistent", true},
	}

	for _, tt := range tests {
		t.Run(tt.projectID, func(t *testing.T) {
			ns, err := ds.GetNamespaces("demo-cluster", tt.projectID)
			if err != nil {
				t.Errorf("GetNamespaces() error = %v", err)
			}
			if tt.wantEmpty && len(ns) > 0 {
				t.Errorf("Expected empty namespaces for %s, got %d", tt.projectID, len(ns))
			}
			if !tt.wantEmpty && len(ns) == 0 {
				t.Errorf("Expected namespaces for %s, got none", tt.projectID)
			}
		})
	}
}

func TestSyntheticDataSource_GetPods(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	t.Run("with namespace filter", func(t *testing.T) {
		pods, err := ds.GetPods("p-default", "default")
		if err != nil {
			t.Errorf("GetPods() error = %v", err)
		}
		// Demo data may have pods in default namespace
		t.Logf("Got %d pods in default namespace", len(pods))
	})

	t.Run("empty namespace", func(t *testing.T) {
		pods, err := ds.GetPods("p-default", "")
		if err != nil {
			t.Errorf("GetPods() error = %v", err)
		}
		t.Logf("Got %d pods with no filter", len(pods))
	})
}

func TestSyntheticDataSource_GetAllPods(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	pods, err := ds.GetAllPods()
	if err != nil {
		t.Errorf("GetAllPods() error = %v", err)
	}
	if len(pods) == 0 {
		t.Error("Expected pods from demo data, got none")
	}
}

func TestSyntheticDataSource_GetDeployments(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	deployments, err := ds.GetDeployments("p-default", "default")
	if err != nil {
		t.Errorf("GetDeployments() error = %v", err)
	}
	// Demo data has deployments
	t.Logf("Got %d deployments", len(deployments))
}

func TestSyntheticDataSource_GetServices(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	services, err := ds.GetServices("p-default", "default")
	if err != nil {
		t.Errorf("GetServices() error = %v", err)
	}
	t.Logf("Got %d services", len(services))
}

func TestSyntheticDataSource_GetCRDs(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	crds, err := ds.GetCRDs("demo-cluster")
	if err != nil {
		t.Errorf("GetCRDs() error = %v", err)
	}
	t.Logf("Got %d CRDs", len(crds))
}

func TestSyntheticDataSource_GetCRDInstances(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	instances, err := ds.GetCRDInstances("demo-cluster", "group", "v1", "widgets")
	if err != nil {
		t.Errorf("GetCRDInstances() error = %v", err)
	}
	if len(instances) != 0 {
		t.Errorf("Expected empty CRD instances for demo, got %d", len(instances))
	}
}

func TestSyntheticDataSource_GetLogs(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	t.Run("existing pod", func(t *testing.T) {
		logs, err := ds.GetLogs("demo-cluster", "default", "nginx-deployment-7c4c7b8f5-x2v9p", "nginx", false)
		if err != nil {
			t.Errorf("GetLogs() error = %v", err)
		}
		t.Logf("Got %d log lines", len(logs))
	})

	t.Run("non-existent pod", func(t *testing.T) {
		// Demo mode returns synthetic logs even for non-existent pods
		logs, err := ds.GetLogs("demo-cluster", "default", "nonexistent", "container", false)
		if err != nil {
			t.Errorf("GetLogs() error = %v", err)
		}
		// Demo mode generates synthetic logs
		t.Logf("Got %d synthetic log lines", len(logs))
	})
}

func TestSyntheticDataSource_GetContainers(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	containers, err := ds.GetContainers("default", "nginx-deployment-7c4c7b8f5-x2v9p")
	if err != nil {
		t.Errorf("GetContainers() error = %v", err)
	}
	t.Logf("Got %d containers", len(containers))
}

func TestSyntheticDataSource_DescribePod(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	desc, err := ds.DescribePod("demo-cluster", "default", "nginx-deployment-7c4c7b8f5-x2v9p")
	if err != nil {
		t.Errorf("DescribePod() error = %v", err)
	}
	if desc == nil {
		t.Error("Expected pod description, got nil")
	}
}

func TestSyntheticDataSource_DescribeDeployment(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	// Try to describe a deployment (may or may not exist in demo data)
	desc, err := ds.DescribeDeployment("demo-cluster", "default", "nginx-deployment")
	if err != nil {
		t.Logf("DescribeDeployment() returned error (may be expected): %v", err)
	}
	if desc != nil {
		t.Logf("Got deployment description")
	}
}

func TestSyntheticDataSource_DescribeService(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	desc, err := ds.DescribeService("demo-cluster", "default", "nginx-service")
	if err != nil {
		t.Errorf("DescribeService() error = %v", err)
	}
	if desc == nil {
		t.Error("Expected service description, got nil")
	}
}

func TestSyntheticDataSource_GetNodes(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	nodes, err := ds.GetNodes()
	if err != nil {
		t.Errorf("GetNodes() error = %v", err)
	}
	if len(nodes) == 0 {
		t.Error("Expected nodes from demo data, got none")
	}
}

func TestSyntheticDataSource_GetAllEvents(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	events, err := ds.GetAllEvents()
	if err != nil {
		t.Errorf("GetAllEvents() error = %v", err)
	}
	t.Logf("Got %d events", len(events))
}

func TestSyntheticDataSource_GetEventsByPod(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	events, err := ds.GetEventsByPod("default", "nginx-deployment-7c4c7b8f5-x2v9p")
	if err != nil {
		t.Errorf("GetEventsByPod() error = %v", err)
	}
	t.Logf("Got %d events for pod", len(events))
}

func TestSyntheticDataSource_GetDaemonSets(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	daemonsets, err := ds.GetDaemonSets()
	if err != nil {
		t.Errorf("GetDaemonSets() error = %v", err)
	}
	t.Logf("Got %d daemonsets", len(daemonsets))
}

func TestSyntheticDataSource_GetEtcdHealth(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	health, err := ds.GetEtcdHealth()
	if err != nil {
		t.Errorf("GetEtcdHealth() error = %v", err)
	}
	if health != nil {
		t.Logf("Got etcd health: healthy=%v", health.Healthy)
	}
}

func TestSyntheticDataSource_GetEtcdDetails(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	details, err := ds.GetEtcdDetails()
	if err != nil {
		t.Errorf("GetEtcdDetails() error = %v", err)
	}
	if details != nil {
		t.Logf("Got etcd details: healthy=%v", details.Healthy)
	}
}

func TestSyntheticDataSource_GetNodeConditions(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	conditions, err := ds.GetNodeConditions()
	if err != nil {
		t.Errorf("GetNodeConditions() error = %v", err)
	}
	if conditions != nil {
		t.Logf("Got %d node conditions", len(conditions))
	}
}

func TestSyntheticDataSource_GetSystemHealth(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	health, err := ds.GetSystemHealth()
	if err != nil {
		t.Errorf("GetSystemHealth() error = %v", err)
	}
	if health != nil {
		t.Logf("Got system health: memory=%.1f%%", health.MemoryUsedPercent)
	}
}

func TestSyntheticDataSource_GetKubeletIssues(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	issues, err := ds.GetKubeletIssues()
	if err != nil {
		t.Errorf("GetKubeletIssues() error = %v", err)
	}
	if issues != nil {
		t.Logf("Got %d kubelet issues", len(issues))
	}
}

func TestSyntheticDataSource_GetOOMAnalysis(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	analysis, err := ds.GetOOMAnalysis()
	if err != nil {
		t.Errorf("GetOOMAnalysis() error = %v", err)
	}
	if analysis != nil {
		t.Logf("Got %d OOM analyses", len(analysis))
	}
}

func TestSyntheticDataSource_GetPodResources(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	resources, err := ds.GetPodResources("nginx-deployment-7c4c7b8f5-x2v9p")
	if err != nil {
		t.Errorf("GetPodResources() error = %v", err)
	}
	t.Logf("Got %d resource specs", len(resources))
}

func TestSyntheticDataSource_GetDiagnosticContext(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	ctx, err := ds.GetDiagnosticContext("default", "nginx-deployment-7c4c7b8f5-x2v9p")
	if err != nil {
		t.Errorf("GetDiagnosticContext() error = %v", err)
	}
	if ctx != nil {
		t.Logf("Got diagnostic context: %s", ctx.RootCause)
	}
}

func TestSyntheticDataSource_GetBundleHealth(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	health := ds.GetBundleHealth()
	if health == nil {
		t.Error("Expected bundle health, got nil")
	} else {
		t.Logf("Got bundle health: %d%%", health.Percentage())
	}
}

func TestSyntheticDataSource_Mode(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	mode := ds.Mode()
	if mode != "DEMO" {
		t.Errorf("Expected mode 'DEMO', got %s", mode)
	}
}

func TestSyntheticDataSource_Close(t *testing.T) {
	ds, _ := NewEmbeddedDataSource(false)

	err := ds.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
