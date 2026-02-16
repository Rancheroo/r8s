package bundle

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// NodeConditions contains full node health from kubectl describe
type NodeConditions struct {
	Name  string
	Roles string // "control-plane,etcd,master" or "worker"

	// Conditions (the critical diagnostic data)
	Ready              bool
	MemoryPressure     bool // True = problem!
	DiskPressure       bool // True = problem!
	PIDPressure        bool // True = problem!
	NetworkUnavailable bool // True = problem!
	EtcdIsVoter        bool // Only for control-plane nodes

	// Capacity/Allocatable
	CPUCapacity       string // "2"
	MemoryCapacity    string // "4005816Ki"
	CPUAllocatable    string // "2"
	MemoryAllocatable string // "4005816Ki"
	StorageCapacity   string // "81106868Ki"
	PodsCapacity      int    // 110

	// Taints
	Taints        []string // ["node-role.kubernetes.io/etcd:NoExecute", ...]
	Unschedulable bool

	// System Info
	KernelVersion    string // "5.15.0-113-generic"
	OSImage          string // "Ubuntu 22.04.4 LTS"
	ContainerRuntime string // "containerd://2.0.5-k3s2"
	KubeletVersion   string // "v1.32.7+rke2r1"

	// Network
	InternalIP string // "134.199.165.191"
	PodCIDR    string // "10.42.0.0/24"

	// Computed flags for quick diagnostics
	HasPressure    bool // True if ANY pressure condition is True
	IsControlPlane bool
	IsEtcd         bool
	IsWorker       bool
}

// ParseNodeDescribe parses kubectl describe node output
func ParseNodeDescribe(extractPath string) ([]NodeConditions, error) {
	bundleRoot := getBundleRoot(extractPath)
	format := DetectFormat(extractPath)
	resolver := NewPathResolver(bundleRoot, format)
	path := filepath.Join(resolver.GetKubectlDir(), "nodesdescribe")

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Split nodes by double newline (each node description is separated)
	nodeBlocks := strings.Split(string(content), "\n\n")
	var nodes []NodeConditions

	for _, block := range nodeBlocks {
		if strings.TrimSpace(block) == "" {
			continue
		}

		node := parseNodeBlock(block)
		if node.Name != "" {
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

// parseNodeBlock parses a single node's describe output
func parseNodeBlock(block string) NodeConditions {
	node := NodeConditions{}
	lines := strings.Split(block, "\n")

	inConditions := false
	inCapacity := false
	inAllocatable := false
	inSystemInfo := false
	inAddresses := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Parse Name
		if strings.HasPrefix(line, "Name:") {
			node.Name = strings.TrimSpace(strings.TrimPrefix(line, "Name:"))
			continue
		}

		// Parse Roles
		if strings.HasPrefix(line, "Roles:") {
			node.Roles = strings.TrimSpace(strings.TrimPrefix(line, "Roles:"))
			// Set role flags
			node.IsControlPlane = strings.Contains(node.Roles, "control-plane")
			node.IsEtcd = strings.Contains(node.Roles, "etcd")
			node.IsWorker = strings.Contains(node.Roles, "worker")
			continue
		}

		// Parse Unschedulable
		if strings.HasPrefix(line, "Unschedulable:") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "Unschedulable:"))
			node.Unschedulable = value == "true"
			continue
		}

		// Section markers
		if strings.HasPrefix(line, "Conditions:") {
			inConditions = true
			inCapacity = false
			inAllocatable = false
			inSystemInfo = false
			inAddresses = false
			continue
		}
		if strings.HasPrefix(line, "Capacity:") {
			inConditions = false
			inCapacity = true
			inAllocatable = false
			inSystemInfo = false
			inAddresses = false
			continue
		}
		if strings.HasPrefix(line, "Allocatable:") {
			inConditions = false
			inCapacity = false
			inAllocatable = true
			inSystemInfo = false
			inAddresses = false
			continue
		}
		if strings.HasPrefix(line, "System Info:") {
			inConditions = false
			inCapacity = false
			inAllocatable = false
			inSystemInfo = true
			inAddresses = false
			continue
		}
		if strings.HasPrefix(line, "Addresses:") {
			inConditions = false
			inCapacity = false
			inAllocatable = false
			inSystemInfo = false
			inAddresses = true
			continue
		}

		// Parse Conditions section
		if inConditions && strings.Contains(trimmed, "True") || strings.Contains(trimmed, "False") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				condType := fields[0]
				status := fields[1]

				switch condType {
				case "Ready":
					node.Ready = status == "True"
				case "MemoryPressure":
					node.MemoryPressure = status == "True"
					if node.MemoryPressure {
						node.HasPressure = true
					}
				case "DiskPressure":
					node.DiskPressure = status == "True"
					if node.DiskPressure {
						node.HasPressure = true
					}
				case "PIDPressure":
					node.PIDPressure = status == "True"
					if node.PIDPressure {
						node.HasPressure = true
					}
				case "NetworkUnavailable":
					node.NetworkUnavailable = status == "True"
					if node.NetworkUnavailable {
						node.HasPressure = true
					}
				case "EtcdIsVoter":
					node.EtcdIsVoter = status == "True"
				}
			}
			continue
		}

		// Parse Capacity section
		if inCapacity {
			if strings.Contains(line, "cpu:") {
				node.CPUCapacity = extractValue(line, "cpu:")
			} else if strings.Contains(line, "memory:") {
				node.MemoryCapacity = extractValue(line, "memory:")
			} else if strings.Contains(line, "ephemeral-storage:") {
				node.StorageCapacity = extractValue(line, "ephemeral-storage:")
			} else if strings.Contains(line, "pods:") {
				podsStr := extractValue(line, "pods:")
				node.PodsCapacity, _ = strconv.Atoi(podsStr)
			}
			continue
		}

		// Parse Allocatable section
		if inAllocatable {
			if strings.Contains(line, "cpu:") {
				node.CPUAllocatable = extractValue(line, "cpu:")
			} else if strings.Contains(line, "memory:") {
				node.MemoryAllocatable = extractValue(line, "memory:")
			}
			continue
		}

		// Parse System Info section
		if inSystemInfo {
			if strings.Contains(line, "Kernel Version:") {
				node.KernelVersion = extractValue(line, "Kernel Version:")
			} else if strings.Contains(line, "OS Image:") {
				node.OSImage = extractValue(line, "OS Image:")
			} else if strings.Contains(line, "Container Runtime Version:") {
				node.ContainerRuntime = extractValue(line, "Container Runtime Version:")
			} else if strings.Contains(line, "Kubelet Version:") {
				node.KubeletVersion = extractValue(line, "Kubelet Version:")
			}
			continue
		}

		// Parse Addresses section
		if inAddresses {
			if strings.Contains(line, "InternalIP:") {
				node.InternalIP = extractValue(line, "InternalIP:")
			}
			continue
		}

		// Parse Taints
		if strings.HasPrefix(line, "Taints:") {
			taintValue := strings.TrimSpace(strings.TrimPrefix(line, "Taints:"))
			if taintValue != "<none>" && taintValue != "" {
				node.Taints = append(node.Taints, taintValue)
			}
			continue
		}

		// Parse PodCIDR
		if strings.HasPrefix(line, "PodCIDR:") {
			node.PodCIDR = strings.TrimSpace(strings.TrimPrefix(line, "PodCIDR:"))
			continue
		}

		// Multi-line taint continuation (indented lines after Taints:)
		if len(node.Taints) > 0 && strings.HasPrefix(line, "                    ") && !strings.Contains(line, ":") {
			taint := strings.TrimSpace(line)
			if taint != "" && taint != "<none>" {
				node.Taints = append(node.Taints, taint)
			}
			continue
		}
	}

	return node
}

// extractValue extracts value after a key in format "  key: value"
func extractValue(line, key string) string {
	if idx := strings.Index(line, key); idx >= 0 {
		value := strings.TrimSpace(line[idx+len(key):])
		return value
	}
	return ""
}
