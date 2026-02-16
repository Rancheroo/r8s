package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetermineQoSClass_Guaranteed(t *testing.T) {
	spec := ResourceSpec{
		MemoryRequest: "512Mi",
		MemoryLimit:   "512Mi",
		CPURequest:    "100m",
		CPULimit:      "100m",
	}

	qos := determineQoSClass(spec)
	if qos != "Guaranteed" {
		t.Errorf("Expected QoS 'Guaranteed', got: %s", qos)
	}
}

func TestDetermineQoSClass_Burstable(t *testing.T) {
	spec := ResourceSpec{
		MemoryRequest: "512Mi",
		CPURequest:    "100m",
	}

	qos := determineQoSClass(spec)
	if qos != "Burstable" {
		t.Errorf("Expected QoS 'Burstable', got: %s", qos)
	}
}

func TestDetermineQoSClass_BestEffort(t *testing.T) {
	spec := ResourceSpec{}

	qos := determineQoSClass(spec)
	if qos != "BestEffort" {
		t.Errorf("Expected QoS 'BestEffort', got: %s", qos)
	}
}

func TestParseResourceLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		prefix   string
		expected map[string]string
	}{
		{
			name:     "Requests line",
			line:     "Requests: cpu=100m,memory=512Mi",
			prefix:   "Requests:",
			expected: map[string]string{"cpu": "100m", "memory": "512Mi"},
		},
		{
			name:     "Limits line",
			line:     "Limits: cpu=500m,memory=1Gi",
			prefix:   "Limits:",
			expected: map[string]string{"cpu": "500m", "memory": "1Gi"},
		},
		{
			name:     "Empty line",
			line:     "Requests:",
			prefix:   "Requests:",
			expected: map[string]string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseResourceLine(tc.line, tc.prefix)
			for k, v := range tc.expected {
				if result[k] != v {
					t.Errorf("Expected %s=%s, got %s", k, v, result[k])
				}
			}
		})
	}
}

func TestParsePodsResourceTable(t *testing.T) {
	content := `NAME                    NAMESPACE       CPU(cores)   MEMORY(bytes)   
nginx-pod               default         100m         512Mi             
mysql-pod               database        500m         1Gi               
`

	specs, err := parsePodsResourceTable(content, "")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(specs) != 2 {
		t.Fatalf("Expected 2 specs, got: %d", len(specs))
	}

	if specs[0].PodName != "default/nginx-pod" {
		t.Errorf("Expected pod name 'default/nginx-pod', got: %s", specs[0].PodName)
	}

	if specs[1].PodName != "database/mysql-pod" {
		t.Errorf("Expected pod name 'database/mysql-pod', got: %s", specs[1].PodName)
	}
}

func TestParsePodsResourceTable_WithFilter(t *testing.T) {
	content := `NAME                    NAMESPACE       CPU(cores)   MEMORY(bytes)   
nginx-pod               default         100m         512Mi             
mysql-pod               database        500m         1Gi               
`

	specs, err := parsePodsResourceTable(content, "mysql-pod")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(specs) != 1 {
		t.Fatalf("Expected 1 spec with filter, got: %d", len(specs))
	}

	if specs[0].PodName != "database/mysql-pod" {
		t.Errorf("Expected mysql-pod, got: %s", specs[0].PodName)
	}
}

func TestParsePodResources_MissingFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "r8s-resources-missing-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create minimal structure
	os.MkdirAll(filepath.Join(tmpDir, "rke2"), 0755)

	specs, _ := ParsePodResources(tmpDir, "nonexistent")
	// Should handle missing files gracefully (returns empty specs)
	// Note: function doesn't return error for missing files, just empty specs
	_ = specs
}
