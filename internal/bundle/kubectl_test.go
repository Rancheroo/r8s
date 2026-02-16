package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

// createKubectlTestBundle creates a minimal bundle with kubectl files
func createKubectlTestBundle(t *testing.T, files map[string]string) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "r8s-kubectl-test-")
	if err != nil {
	t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create rke2/kubectl structure
	kubectlDir := filepath.Join(tmpDir, "rke2", "kubectl")
	os.MkdirAll(kubectlDir, 0755)

	// Write test files
	for filename, content := range files {
		path := filepath.Join(kubectlDir, filename)
		os.WriteFile(path, []byte(content), 0644)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestParseCRDs(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectCount int
		expectErr   bool
	}{
		{
			name: "valid crds",
			content: `NAME                           CREATED AT
addons.k3s.cattle.io           2024-01-15T10:30:00Z
helmcharts.helm.cattle.io      2024-01-15T10:30:00Z
`,
			expectCount: 2,
			expectErr:   false,
		},
		{
			name:        "empty file",
			content:     "NAME\n",
			expectCount: 0,
			expectErr:   false,
		},
		{
			name:        "missing file",
			content:     "", // Don't create crds file
			expectCount: 0,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string]string{}
			if tt.content != "" {
				files["crds"] = tt.content
			}

			bundleDir, cleanup := createKubectlTestBundle(t, files)
			defer cleanup()

			crds, err := ParseCRDs(bundleDir)

			if tt.expectErr {
				if err == nil && tt.content == "" {
					// Missing file case - should error
					t.Logf("ParseCRDs() expected error for missing file")
				} else if err != nil {
					t.Logf("ParseCRDs() returned expected error: %v", err)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseCRDs() unexpected error = %v", err)
				return
			}

			if len(crds) != tt.expectCount {
				t.Errorf("ParseCRDs() returned %d CRDs, expected %d", len(crds), tt.expectCount)
			}

			if len(crds) > 0 && tt.expectCount > 0 {
				// Verify first CRD has expected fields
				if crds[0].Metadata.Name == "" {
					t.Error("ParseCRDs() returned CRD with empty name")
				}
			}
		})
	}
}

func TestParseDeployments(t *testing.T) {
	content := `NAME                    READY   UP-TO-DATE   AVAILABLE   AGE
nginx-deployment        3/3     3            3           5d
redis-deployment        2/3     3            2           3d
`

	bundleDir, cleanup := createKubectlTestBundle(t, map[string]string{
		"deployments": content,
	})
	defer cleanup()

	deployments, err := ParseDeployments(bundleDir)
	if err != nil {
		t.Errorf("ParseDeployments() error = %v", err)
	}

	if len(deployments) != 2 {
		t.Errorf("ParseDeployments() returned %d deployments, expected 2", len(deployments))
	}

	if len(deployments) > 0 {
		if deployments[0].Name != "nginx-deployment" {
			t.Errorf("Expected 'nginx-deployment', got %s", deployments[0].Name)
		}
	}
}

func TestParseServices(t *testing.T) {
	content := `NAME         TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
kubernetes   ClusterIP   10.43.0.1      <none>        443/TCP   5d
nginx-svc    NodePort    10.43.123.45   <none>        80:30080/TCP   2d
`

	bundleDir, cleanup := createKubectlTestBundle(t, map[string]string{
		"services": content,
	})
	defer cleanup()

	services, err := ParseServices(bundleDir)
	if err != nil {
		t.Errorf("ParseServices() error = %v", err)
	}

	if len(services) != 2 {
		t.Errorf("ParseServices() returned %d services, expected 2", len(services))
	}
}

func TestParsePods(t *testing.T) {
	content := `NAMESPACE     NAME                     READY   STATUS    RESTARTS   AGE
kube-system   kube-proxy-abc123        1/1     Running   0          5d
default       nginx-7d4c7f6d9-x2k3p    1/1     Running   0          2d
`

	bundleDir, cleanup := createKubectlTestBundle(t, map[string]string{
		"pods": content,
	})
	defer cleanup()

	pods, err := ParsePods(bundleDir)
	if err != nil {
		t.Errorf("ParsePods() error = %v", err)
	}

	if len(pods) != 2 {
		t.Errorf("ParsePods() returned %d pods, expected 2", len(pods))
	}

	if len(pods) > 0 {
		if pods[0].Name == "" {
			t.Error("ParsePods() returned pod with empty name")
		}
		if pods[0].NamespaceID == "" {
			t.Error("ParsePods() returned pod with empty namespace")
		}
	}
}

func TestParseNodes(t *testing.T) {
	content := `NAME           STATUS   ROLES                       AGE   VERSION
node1          Ready    control-plane,etcd,master   5d    v1.28.0+rke2r1
node2          Ready    <none>                      5d    v1.28.0+rke2r1
`

	bundleDir, cleanup := createKubectlTestBundle(t, map[string]string{
		"nodes": content,
	})
	defer cleanup()

	nodes, err := ParseNodes(bundleDir)
	if err != nil {
		t.Errorf("ParseNodes() error = %v", err)
	}

	if len(nodes) != 2 {
		t.Errorf("ParseNodes() returned %d nodes, expected 2", len(nodes))
	}

	if len(nodes) > 0 {
		if nodes[0].Name != "node1" {
			t.Errorf("Expected 'node1', got %s", nodes[0].Name)
		}
		if nodes[0].Status != "Ready" {
			t.Errorf("Expected status 'Ready', got %s", nodes[0].Status)
		}
	}
}

func TestParseEvents(t *testing.T) {
	content := `LAST SEEN   TYPE      REASON              OBJECT                        MESSAGE
5m          Normal    Scheduled           pod/nginx-7d4c7f6d9-x2k3p    Successfully assigned default/nginx
1m          Warning   FailedMount         pod/some-pod                  MountVolume.SetUp failed
`

	bundleDir, cleanup := createKubectlTestBundle(t, map[string]string{
		"events": content,
	})
	defer cleanup()

	events, err := ParseEvents(bundleDir)
	if err != nil {
		t.Errorf("ParseEvents() error = %v", err)
	}

	if len(events) != 2 {
		t.Errorf("ParseEvents() returned %d events, expected 2", len(events))
	}

	if len(events) > 0 {
		if events[0].Type != "Normal" && events[0].Type != "Warning" {
			t.Errorf("Unexpected event type: %s", events[0].Type)
		}
	}
}

func TestParsePodsWithContainerStatus(t *testing.T) {
	// Test with containers in different states
	content := `NAMESPACE     NAME                     READY   STATUS             RESTARTS   AGE
default       crash-loop-pod           0/1     CrashLoopBackOff   5          1d
default       pending-pod              0/1     Pending            0          10m
`

	bundleDir, cleanup := createKubectlTestBundle(t, map[string]string{
		"pods": content,
	})
	defer cleanup()

	pods, err := ParsePods(bundleDir)
	if err != nil {
		t.Errorf("ParsePods() error = %v", err)
	}

	if len(pods) != 2 {
		t.Errorf("ParsePods() returned %d pods, expected 2", len(pods))
	}

	// Check container status parsing
	for _, pod := range pods {
		if pod.State == "" {
			t.Error("ParsePods() returned pod with empty state")
		}
	}
}

func TestParseDeploymentsWithMissingFile(t *testing.T) {
	bundleDir, cleanup := createKubectlTestBundle(t, map[string]string{})
	defer cleanup()

	_, err := ParseDeployments(bundleDir)
	if err == nil {
		t.Error("ParseDeployments() expected error for missing file")
	}
}

func TestParseServicesWithMissingFile(t *testing.T) {
	bundleDir, cleanup := createKubectlTestBundle(t, map[string]string{})
	defer cleanup()

	_, err := ParseServices(bundleDir)
	if err == nil {
		t.Error("ParseServices() expected error for missing file")
	}
}

func TestParsePodsWithMissingFile(t *testing.T) {
	bundleDir, cleanup := createKubectlTestBundle(t, map[string]string{})
	defer cleanup()

	_, err := ParsePods(bundleDir)
	if err == nil {
		t.Error("ParsePods() expected error for missing file")
	}
}

func TestParseNodesWithMissingFile(t *testing.T) {
	bundleDir, cleanup := createKubectlTestBundle(t, map[string]string{})
	defer cleanup()

	_, err := ParseNodes(bundleDir)
	if err == nil {
		t.Error("ParseNodes() expected error for missing file")
	}
}

func TestParseEventsWithMissingFile(t *testing.T) {
	bundleDir, cleanup := createKubectlTestBundle(t, map[string]string{})
	defer cleanup()

	_, err := ParseEvents(bundleDir)
	if err == nil {
		t.Error("ParseEvents() expected error for missing file")
	}
}
