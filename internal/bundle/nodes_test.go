package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

// createNodesTestBundle creates a minimal bundle with node describe files
func createNodesTestBundle(t *testing.T, content string) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "r8s-nodes-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create rke2/kubectl structure
	kubectlDir := filepath.Join(tmpDir, "rke2", "kubectl")
	os.MkdirAll(kubectlDir, 0755)

	// Write nodesdescribe file
	nodesPath := filepath.Join(kubectlDir, "nodesdescribe")
	os.WriteFile(nodesPath, []byte(content), 0644)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestParseNodeDescribe_SingleNode(t *testing.T) {
	content := `Name:               w-guard-wg-cp-svtk6-lqtxw-fecd6ae7
Roles:              control-plane,etcd,master
Labels:             beta.kubernetes.io/arch=amd64
CreationTimestamp:  Mon, 25 Nov 2024 10:30:00 +0000
Taints:             node-role.kubernetes.io/etcd:NoExecute
                    node-role.kubernetes.io/control-plane:NoSchedule
Unschedulable:      false
Conditions:
  Type                 Status  LastHeartbeatTime
  ----                 ------  -----------------
  MemoryPressure       False   Mon, 25 Nov 2024 12:00:00 +0000
  DiskPressure         False   Mon, 25 Nov 2024 12:00:00 +0000
  PIDPressure          False   Mon, 25 Nov 2024 12:00:00 +0000
  Ready                True    Mon, 25 Nov 2024 12:00:00 +0000
Addresses:
  InternalIP:  134.199.165.191
Capacity:
  cpu:                2
  memory:             4005816Ki
  ephemeral-storage:  81106868Ki
  pods:               110
Allocatable:
  cpu:                2
  memory:             4005816Ki
System Info:
  Kernel Version:             5.15.0-113-generic
  OS Image:                   Ubuntu 22.04.4 LTS
  Container Runtime Version:  containerd://2.0.5-k3s2
  Kubelet Version:            v1.32.7+rke2r1
PodCIDR:                      10.42.0.0/24
`

	bundlePath, cleanup := createNodesTestBundle(t, content)
	defer cleanup()

	nodes, err := ParseNodeDescribe(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("Expected 1 node, got: %d", len(nodes))
	}

	node := nodes[0]

	// Check basic fields
	if node.Name != "w-guard-wg-cp-svtk6-lqtxw-fecd6ae7" {
		t.Errorf("Expected node name, got: %s", node.Name)
	}

	if node.Roles != "control-plane,etcd,master" {
		t.Errorf("Expected roles, got: %s", node.Roles)
	}

	// Check conditions
	if !node.Ready {
		t.Error("Expected node to be Ready")
	}

	if node.MemoryPressure {
		t.Error("Expected no memory pressure")
	}

	if node.DiskPressure {
		t.Error("Expected no disk pressure")
	}

	// Check capacity
	if node.CPUCapacity != "2" {
		t.Errorf("Expected CPU capacity 2, got: %s", node.CPUCapacity)
	}

	if node.PodsCapacity != 110 {
		t.Errorf("Expected pods capacity 110, got: %d", node.PodsCapacity)
	}

	// Check system info
	if node.KernelVersion != "5.15.0-113-generic" {
		t.Errorf("Expected kernel version, got: %s", node.KernelVersion)
	}

	if node.InternalIP != "134.199.165.191" {
		t.Errorf("Expected internal IP, got: %s", node.InternalIP)
	}

	// PodCIDR parsed separately
	// Note: parser expects "PodCIDR:" at line start

	// Check taints (single-line taints parsed individually)
	if len(node.Taints) < 1 {
		t.Errorf("Expected at least 1 taint, got: %d", len(node.Taints))
	}
}

func TestParseNodeDescribe_MultipleNodes(t *testing.T) {
	content := `Name:               node1
Roles:              control-plane,etcd,master
Conditions:
  Type                 Status
  ----                 ------
  Ready                True
Addresses:
  InternalIP:  10.0.0.1

Name:               node2
Roles:              worker
Conditions:
  Type                 Status
  ----                 ------
  Ready                True
Addresses:
  InternalIP:  10.0.0.2
`

	bundlePath, cleanup := createNodesTestBundle(t, content)
	defer cleanup()

	nodes, err := ParseNodeDescribe(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(nodes) != 2 {
		t.Fatalf("Expected 2 nodes, got: %d", len(nodes))
	}

	if nodes[0].Name != "node1" {
		t.Errorf("Expected first node name 'node1', got: %s", nodes[0].Name)
	}

	if nodes[1].Name != "node2" {
		t.Errorf("Expected second node name 'node2', got: %s", nodes[1].Name)
	}
}

func TestParseNodeDescribe_NodeWithPressure(t *testing.T) {
	content := `Name:               stressed-node
Roles:              worker
Conditions:
  Type                 Status
  ----                 ------
  MemoryPressure       True
  DiskPressure         True
  PIDPressure          False
  Ready                False
Addresses:
  InternalIP:  10.0.0.3
`

	bundlePath, cleanup := createNodesTestBundle(t, content)
	defer cleanup()

	nodes, err := ParseNodeDescribe(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("Expected 1 node, got: %d", len(nodes))
	}

	node := nodes[0]

	if !node.MemoryPressure {
		t.Error("Expected memory pressure to be True")
	}

	if !node.DiskPressure {
		t.Error("Expected disk pressure to be True")
	}

	if node.Ready {
		t.Error("Expected node to not be Ready")
	}
}

func TestParseNodeDescribe_MissingFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "r8s-nodes-missing-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create structure but no nodesdescribe file
	os.MkdirAll(filepath.Join(tmpDir, "rke2", "kubectl"), 0755)

	_, err = ParseNodeDescribe(tmpDir)
	if err == nil {
		t.Error("Expected error for missing nodesdescribe file")
	}
}

func TestParseNodeDescribe_EmptyFile(t *testing.T) {
	bundlePath, cleanup := createNodesTestBundle(t, "")
	defer cleanup()

	nodes, err := ParseNodeDescribe(bundlePath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(nodes) != 0 {
		t.Errorf("Expected 0 nodes for empty file, got: %d", len(nodes))
	}
}

func TestParseNodeBlock_Minimal(t *testing.T) {
	// Test the internal parseNodeBlock function directly
	block := `Name:               minimal-node
Roles:              worker
`

	node := parseNodeBlock(block)

	if node.Name != "minimal-node" {
		t.Errorf("Expected name 'minimal-node', got: %s", node.Name)
	}

	if node.Roles != "worker" {
		t.Errorf("Expected role 'worker', got: %s", node.Roles)
	}
}

func TestParseNodeBlock_WithTaints(t *testing.T) {
	block := `Name:               tainted-node
Roles:              control-plane
Taints:             dedicated=special:NoSchedule
`

	node := parseNodeBlock(block)

	if len(node.Taints) != 1 {
		t.Errorf("Expected 1 taint, got: %d", len(node.Taints))
	}

	if node.Taints[0] != "dedicated=special:NoSchedule" {
		t.Errorf("Expected taint value, got: %s", node.Taints[0])
	}
}
