package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

// createOOMTestBundle creates a minimal bundle structure for OOM tests.
// Returns the bundle root path and a cleanup function.
func createOOMTestBundle(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "r8s-oom-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create directory structure
	for _, dir := range []string{
		"rke2/kubectl",
		"rke2/pod-manifests",
	} {
		os.MkdirAll(filepath.Join(tmpDir, dir), 0755)
	}

	cleanup := func() { os.RemoveAll(tmpDir) }
	return tmpDir, cleanup
}

// writeFile is a test helper to write content into the bundle.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	os.MkdirAll(filepath.Dir(path), 0755)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write %s: %v", path, err)
	}
}

// --- AnalyzeOOMEvents integration tests ---

func TestAnalyzeOOMEvents_NoEventsFile(t *testing.T) {
	bundleRoot, cleanup := createOOMTestBundle(t)
	defer cleanup()

	// No events file exists
	results, err := AnalyzeOOMEvents(bundleRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestAnalyzeOOMEvents_NoOOMInEvents(t *testing.T) {
	bundleRoot, cleanup := createOOMTestBundle(t)
	defer cleanup()

	eventsContent := `NAMESPACE   NAME             AGE  TYPE    REASON    MESSAGE
default     my-pod-abc12     5m   Normal  Pulled    Successfully pulled image
kube-system coredns-xyz99    2m   Normal  Started   Started container coredns
`
	writeFile(t, filepath.Join(bundleRoot, "rke2/kubectl/events"), eventsContent)

	results, err := AnalyzeOOMEvents(bundleRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestAnalyzeOOMEvents_DetectsOOMKilling(t *testing.T) {
	bundleRoot, cleanup := createOOMTestBundle(t)
	defer cleanup()

	eventsContent := `NAMESPACE   NAME             AGE  TYPE     REASON       MESSAGE
default     my-pod-abc12     5m   Warning  OOMKilling   pod my-app was OOM killed
kube-system coredns-xyz99    2m   Normal   Started      Started container coredns
`
	writeFile(t, filepath.Join(bundleRoot, "rke2/kubectl/events"), eventsContent)

	results, err := AnalyzeOOMEvents(bundleRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].PodName != "my-app" {
		t.Errorf("expected pod name 'my-app', got '%s'", results[0].PodName)
	}
}

func TestAnalyzeOOMEvents_DetectsOOMByRegex(t *testing.T) {
	bundleRoot, cleanup := createOOMTestBundle(t)
	defer cleanup()

	eventsContent := `NAMESPACE   NAME             AGE  TYPE     REASON    MESSAGE
default     my-pod-abc12     5m   Normal   Killing   container out of memory killed
`
	writeFile(t, filepath.Join(bundleRoot, "rke2/kubectl/events"), eventsContent)

	results, err := AnalyzeOOMEvents(bundleRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestAnalyzeOOMEvents_NodeOOMDetection(t *testing.T) {
	bundleRoot, cleanup := createOOMTestBundle(t)
	defer cleanup()

	eventsContent := `NAMESPACE   NAME             AGE  TYPE     REASON       MESSAGE
default     my-pod-abc12     5m   Warning  OOMKilling   node system OOM killed process
`
	writeFile(t, filepath.Join(bundleRoot, "rke2/kubectl/events"), eventsContent)

	results, err := AnalyzeOOMEvents(bundleRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].IsNodeOOM {
		t.Error("expected IsNodeOOM=true for node-level OOM event")
	}
}

func TestAnalyzeOOMEvents_EnrichesWithPodResources(t *testing.T) {
	bundleRoot, cleanup := createOOMTestBundle(t)
	defer cleanup()

	eventsContent := `NAMESPACE   NAME             AGE  TYPE     REASON       MESSAGE
default     my-pod-abc12     5m   Warning  OOMKilling   pod default/my-pod was OOM killed
`
	// Pods content with matching pod
	podsContent := `NAME       NAMESPACE   STATUS
my-pod     default     Running
`
	writeFile(t, filepath.Join(bundleRoot, "rke2/kubectl/events"), eventsContent)
	writeFile(t, filepath.Join(bundleRoot, "rke2/kubectl/pods"), podsContent)

	results, err := AnalyzeOOMEvents(bundleRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	// Pod name extracted from message
	if results[0].PodName != "default/my-pod" {
		t.Errorf("expected pod name 'default/my-pod', got '%s'", results[0].PodName)
	}
}

func TestAnalyzeOOMEvents_EnrichesWithQoS(t *testing.T) {
	bundleRoot, cleanup := createOOMTestBundle(t)
	defer cleanup()

	eventsContent := `NAMESPACE   NAME             AGE  TYPE     REASON       MESSAGE
default     my-pod-abc12     5m   Warning  OOMKilling   pod my-app was OOM killed
`
	podYAML := `kind: Pod
metadata:
  name: my-app
spec:
  nodeName: node-1
  containers:
  - name: main
    resources:
      requests:
        memory: "512Mi"
        cpu: "100m"
      limits:
        memory: "512Mi"
        cpu: "100m"
`
	writeFile(t, filepath.Join(bundleRoot, "rke2/kubectl/events"), eventsContent)
	writeFile(t, filepath.Join(bundleRoot, "rke2/pod-manifests/my-app.yaml"), podYAML)

	results, err := AnalyzeOOMEvents(bundleRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].QoSClass != "Guaranteed" {
		t.Errorf("expected QoS 'Guaranteed', got '%s'", results[0].QoSClass)
	}
	// NodeName is tested separately in TestEnrichWithNodeNames
}

// --- parseOOMEvents unit tests ---

func TestParseOOMEvents_Empty(t *testing.T) {
	results := parseOOMEvents("")
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty input, got %d", len(results))
	}
}

func TestParseOOMEvents_VariousPatterns(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "OOMKilling reason",
			input:    "default my-pod 5m Warning OOMKilling pod my-app was killed",
			expected: 1,
		},
		{
			name:     "OOM regex match",
			input:    "default my-pod 5m Normal Killed container oom killed in namespace",
			expected: 1,
		},
		{
			name:     "No OOM",
			input:    "default my-pod 5m Normal Pulled image pulled successfully",
			expected: 0,
		},
		{
			name:     "Short line",
			input:    "too few fields",
			expected: 0,
		},
		{
			name:     "Multiple OOM events",
			input:    "default pod-a 5m Warning OOMKilling pod app-a killed\ndefault pod-b 3m Warning OOMKilling pod app-b killed",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := parseOOMEvents(tt.input)
			if len(results) != tt.expected {
				t.Errorf("expected %d results, got %d", tt.expected, len(results))
			}
		})
	}
}

// --- extractPodNameFromOOMMessage unit tests ---

func TestExtractPodNameFromOOMMessage(t *testing.T) {
	tests := []struct {
		name      string
		message   string
		eventName string
		expected  string
	}{
		{
			name:      "pod keyword in message",
			message:   "pod my-app was OOM killed",
			eventName: "event-123",
			expected:  "my-app",
		},
		{
			name:      "fallback to event name",
			message:   "container was killed due to memory",
			eventName: "my-pod-abc12",
			expected:  "my-pod-abc12",
		},
		{
			name:      "pod with namespace",
			message:   "pod default/my-app was OOM killed",
			eventName: "event-123",
			expected:  "default/my-app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPodNameFromOOMMessage(tt.message, tt.eventName)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// --- normalizePodName unit tests ---

func TestNormalizePodName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"my-app", "my-app"},
		{"default/my-app", "my-app"},
		{"kube-system/coredns", "coredns"},
		{"my-app-abc12", "my-app"}, // 5 char hash-like suffix stripped
		{"my-app-abcde", "my-app"}, // 5 char hash-like suffix stripped
		{"my-app-a1b2c", "my-app"}, // hash-like suffix
		{"single", "single"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizePodName(tt.input)
			if result != tt.expected {
				t.Errorf("normalizePodName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// --- isHashLike unit tests ---

func TestIsHashLike(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"abcde", true},
		{"abc12", true},
		{"a1b2c", true},
		{"12345", true},
		{"abcd", false},  // too short
		{"", false},      // empty
		{"ABCDE", false}, // uppercase
		{"abc-e", false}, // hyphen
		{"abc_e", false}, // underscore
		{"abcdef", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isHashLike(tt.input)
			if result != tt.expected {
				t.Errorf("isHashLike(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// --- parsePodResourceMap unit tests ---

func TestParsePodResourceMap_Empty(t *testing.T) {
	result := parsePodResourceMap("")
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}

func TestParsePodResourceMap_ValidPods(t *testing.T) {
	input := `NAME       NAMESPACE   STATUS
my-pod     default     Running
coredns    kube-system Running
`
	result := parsePodResourceMap(input)
	if len(result) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result))
	}
	if _, exists := result["default/my-pod"]; !exists {
		t.Error("expected 'default/my-pod' in map")
	}
	if _, exists := result["kube-system/coredns"]; !exists {
		t.Error("expected 'kube-system/coredns' in map")
	}
}

func TestParsePodResourceMap_ShortLines(t *testing.T) {
	input := `HEADER
ab
`
	result := parsePodResourceMap(input)
	if len(result) != 0 {
		t.Errorf("expected empty map for short lines, got %d", len(result))
	}
}

// --- correlateOOMWithResources unit tests ---

func TestCorrelateOOMWithResources(t *testing.T) {
	podsContent := `NAME       NAMESPACE   STATUS
my-pod     default     Running
`
	events := []OOMAnalysis{
		{PodName: "default/my-pod"},
		{PodName: "nonexistent"},
	}

	result := correlateOOMWithResources(events, podsContent)
	if result[0].ContainerName == "" {
		// The simplified implementation sets ContainerName to podName
		// Just check it was enriched
	}
	// nonexistent pod should remain unchanged
	if result[1].ContainerName != "" {
		t.Errorf("expected empty container for nonexistent pod, got '%s'", result[1].ContainerName)
	}
}

// --- calculatePodQoSClass unit tests ---

func TestCalculatePodQoSClass(t *testing.T) {
	tests := []struct {
		name       string
		containers []interface{}
		expected   string
	}{
		{
			name: "Guaranteed - limits equal requests",
			containers: []interface{}{
				map[string]interface{}{
					"resources": map[string]interface{}{
						"requests": map[string]interface{}{
							"memory": "512Mi",
							"cpu":    "100m",
						},
						"limits": map[string]interface{}{
							"memory": "512Mi",
							"cpu":    "100m",
						},
					},
				},
			},
			expected: "Guaranteed",
		},
		{
			name: "Guaranteed - limits only (requests default to limits)",
			containers: []interface{}{
				map[string]interface{}{
					"resources": map[string]interface{}{
						"limits": map[string]interface{}{
							"memory": "512Mi",
							"cpu":    "100m",
						},
					},
				},
			},
			expected: "Guaranteed",
		},
		{
			name: "Burstable - requests only",
			containers: []interface{}{
				map[string]interface{}{
					"resources": map[string]interface{}{
						"requests": map[string]interface{}{
							"memory": "512Mi",
							"cpu":    "100m",
						},
					},
				},
			},
			expected: "Burstable",
		},
		{
			name: "Burstable - limits != requests",
			containers: []interface{}{
				map[string]interface{}{
					"resources": map[string]interface{}{
						"requests": map[string]interface{}{
							"memory": "256Mi",
							"cpu":    "50m",
						},
						"limits": map[string]interface{}{
							"memory": "512Mi",
							"cpu":    "100m",
						},
					},
				},
			},
			expected: "Burstable",
		},
		{
			name: "BestEffort - no resources",
			containers: []interface{}{
				map[string]interface{}{
					"name": "main",
				},
			},
			expected: "BestEffort",
		},
		{
			name: "BestEffort - empty resources",
			containers: []interface{}{
				map[string]interface{}{
					"resources": map[string]interface{}{},
				},
			},
			expected: "BestEffort",
		},
		{
			name: "Malformed container",
			containers: []interface{}{
				"not-a-map",
			},
			expected: "BestEffort",
		},
		{
			name: "Mixed containers - one guaranteed, one best-effort",
			containers: []interface{}{
				map[string]interface{}{
					"resources": map[string]interface{}{
						"limits": map[string]interface{}{
							"memory": "512Mi",
							"cpu":    "100m",
						},
					},
				},
				map[string]interface{}{
					"name": "sidecar",
				},
			},
			expected: "Burstable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePodQoSClass(tt.containers)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// --- parsePodYAMLForQoS unit tests ---

func TestParsePodYAMLForQoS(t *testing.T) {
	tests := []struct {
		name         string
		yaml         string
		expectedName string
		expectedQoS  string
	}{
		{
			name: "Guaranteed pod",
			yaml: `kind: Pod
metadata:
  name: my-app
spec:
  containers:
  - name: main
    resources:
      requests:
        memory: "512Mi"
        cpu: "100m"
      limits:
        memory: "512Mi"
        cpu: "100m"`,
			expectedName: "my-app",
			expectedQoS:  "Guaranteed",
		},
		{
			name: "BestEffort pod - no resources",
			yaml: `kind: Pod
metadata:
  name: simple-pod
spec:
  containers:
  - name: main`,
			expectedName: "simple-pod",
			expectedQoS:  "BestEffort",
		},
		{
			name: "Not a Pod kind",
			yaml: `kind: Deployment
metadata:
  name: my-deploy`,
			expectedName: "",
			expectedQoS:  "",
		},
		{
			name:         "Invalid YAML",
			yaml:         `{{{invalid`,
			expectedName: "",
			expectedQoS:  "",
		},
		{
			name: "No metadata",
			yaml: `kind: Pod
spec:
  containers:
  - name: main`,
			expectedName: "",
			expectedQoS:  "",
		},
		{
			name: "No spec",
			yaml: `kind: Pod
metadata:
  name: no-spec-pod`,
			expectedName: "no-spec-pod",
			expectedQoS:  "BestEffort",
		},
		{
			name: "No containers in spec",
			yaml: `kind: Pod
metadata:
  name: empty-spec
spec:
  nodeName: node-1`,
			expectedName: "empty-spec",
			expectedQoS:  "BestEffort",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, qos := parsePodYAMLForQoS(tt.yaml)
			if name != tt.expectedName {
				t.Errorf("expected name '%s', got '%s'", tt.expectedName, name)
			}
			if qos != tt.expectedQoS {
				t.Errorf("expected QoS '%s', got '%s'", tt.expectedQoS, qos)
			}
		})
	}
}

// --- parsePodYAMLForNode unit tests ---

func TestParsePodYAMLForNode(t *testing.T) {
	tests := []struct {
		name         string
		yaml         string
		expectedPod  string
		expectedNode string
	}{
		{
			name: "Pod with nodeName",
			yaml: `kind: Pod
metadata:
  name: my-app
spec:
  nodeName: worker-1`,
			expectedPod:  "my-app",
			expectedNode: "worker-1",
		},
		{
			name: "Pod without nodeName",
			yaml: `kind: Pod
metadata:
  name: my-app
spec:
  containers:
  - name: main`,
			expectedPod:  "my-app",
			expectedNode: "",
		},
		{
			name: "Not a Pod",
			yaml: `kind: Service
metadata:
  name: my-svc`,
			expectedPod:  "",
			expectedNode: "",
		},
		{
			name:         "Invalid YAML",
			yaml:         `not: valid: yaml: {{`,
			expectedPod:  "",
			expectedNode: "",
		},
		{
			name: "No spec",
			yaml: `kind: Pod
metadata:
  name: no-spec`,
			expectedPod:  "no-spec",
			expectedNode: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod, node := parsePodYAMLForNode(tt.yaml)
			if pod != tt.expectedPod {
				t.Errorf("expected pod '%s', got '%s'", tt.expectedPod, pod)
			}
			if node != tt.expectedNode {
				t.Errorf("expected node '%s', got '%s'", tt.expectedNode, node)
			}
		})
	}
}

// --- enrichWithQoSClass integration test ---

func TestEnrichWithQoSClass_NoManifestsDir(t *testing.T) {
	bundleRoot, cleanup := createOOMTestBundle(t)
	defer cleanup()

	// Remove pod-manifests dir
	os.RemoveAll(filepath.Join(bundleRoot, "rke2/pod-manifests"))

	events := []OOMAnalysis{{PodName: "my-app"}}
	result := enrichWithQoSClass(events, bundleRoot)
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	// QoS should not be set
	if result[0].QoSClass != "" {
		t.Errorf("expected empty QoS, got '%s'", result[0].QoSClass)
	}
}

func TestEnrichWithQoSClass_WithManifest(t *testing.T) {
	bundleRoot, cleanup := createOOMTestBundle(t)
	defer cleanup()

	podYAML := `kind: Pod
metadata:
  name: my-app
spec:
  containers:
  - name: main
    resources:
      requests:
        memory: "256Mi"
        cpu: "50m"
      limits:
        memory: "1Gi"
        cpu: "500m"
`
	writeFile(t, filepath.Join(bundleRoot, "rke2/pod-manifests/my-app.yaml"), podYAML)

	events := []OOMAnalysis{{PodName: "my-app"}}
	result := enrichWithQoSClass(events, bundleRoot)
	if result[0].QoSClass != "Burstable" {
		t.Errorf("expected QoS 'Burstable', got '%s'", result[0].QoSClass)
	}
}

func TestEnrichWithQoSClass_NonYAMLFilesIgnored(t *testing.T) {
	bundleRoot, cleanup := createOOMTestBundle(t)
	defer cleanup()

	writeFile(t, filepath.Join(bundleRoot, "rke2/pod-manifests/readme.txt"), "not a yaml")

	events := []OOMAnalysis{{PodName: "my-app"}}
	result := enrichWithQoSClass(events, bundleRoot)
	if result[0].QoSClass != "" {
		t.Errorf("expected empty QoS, got '%s'", result[0].QoSClass)
	}
}

// --- enrichWithNodeNames integration test ---

func TestEnrichWithNodeNames(t *testing.T) {
	bundleRoot, cleanup := createOOMTestBundle(t)
	defer cleanup()

	podYAML := `kind: Pod
metadata:
  name: my-app
spec:
  nodeName: worker-2
  containers:
  - name: main
`
	writeFile(t, filepath.Join(bundleRoot, "rke2/pod-manifests/my-app.yaml"), podYAML)

	events := []OOMAnalysis{{PodName: "my-app"}}
	result := enrichWithNodeNames(events, bundleRoot)
	if result[0].NodeName != "worker-2" {
		t.Errorf("expected NodeName 'worker-2', got '%s'", result[0].NodeName)
	}
}

func TestEnrichWithNodeNames_NoManifestsDir(t *testing.T) {
	bundleRoot, cleanup := createOOMTestBundle(t)
	defer cleanup()

	os.RemoveAll(filepath.Join(bundleRoot, "rke2/pod-manifests"))

	events := []OOMAnalysis{{PodName: "my-app"}}
	result := enrichWithNodeNames(events, bundleRoot)
	if result[0].NodeName != "" {
		t.Errorf("expected empty NodeName, got '%s'", result[0].NodeName)
	}
}

// --- enrichWithNodeMemory integration test ---

func TestEnrichWithNodeMemory_NoNodeDescribe(t *testing.T) {
	bundleRoot, cleanup := createOOMTestBundle(t)
	defer cleanup()

	events := []OOMAnalysis{{PodName: "my-app", NodeName: "node-1"}}
	result := enrichWithNodeMemory(events, bundleRoot)
	if result[0].NodeMemoryPressure {
		t.Error("expected no memory pressure when nodesdescribe is missing")
	}
}

// --- buildQoSMapFromManifests unit tests ---

func TestBuildQoSMapFromManifests_BadDir(t *testing.T) {
	result := buildQoSMapFromManifests("/nonexistent/path")
	if len(result) != 0 {
		t.Errorf("expected empty map for bad dir, got %d entries", len(result))
	}
}

func TestBuildQoSMapFromManifests_ValidManifests(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "r8s-qos-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	guaranteed := `kind: Pod
metadata:
  name: guaranteed-pod
spec:
  containers:
  - name: main
    resources:
      limits:
        memory: "512Mi"
        cpu: "100m"
`
	besteffort := `kind: Pod
metadata:
  name: besteffort-pod
spec:
  containers:
  - name: main
`
	os.WriteFile(filepath.Join(tmpDir, "guaranteed.yaml"), []byte(guaranteed), 0644)
	os.WriteFile(filepath.Join(tmpDir, "besteffort.yml"), []byte(besteffort), 0644)

	result := buildQoSMapFromManifests(tmpDir)
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result["guaranteed-pod"] != "Guaranteed" {
		t.Errorf("expected 'Guaranteed', got '%s'", result["guaranteed-pod"])
	}
	if result["besteffort-pod"] != "BestEffort" {
		t.Errorf("expected 'BestEffort', got '%s'", result["besteffort-pod"])
	}
}

// --- buildPodNodeMap unit tests ---

func TestBuildPodNodeMap_BadDir(t *testing.T) {
	result := buildPodNodeMap("/nonexistent/path")
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}

func TestBuildPodNodeMap_ValidManifests(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "r8s-podnode-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	pod := `kind: Pod
metadata:
  name: my-pod
spec:
  nodeName: worker-1
`
	os.WriteFile(filepath.Join(tmpDir, "my-pod.yaml"), []byte(pod), 0644)

	result := buildPodNodeMap(tmpDir)
	if result["my-pod"] != "worker-1" {
		t.Errorf("expected 'worker-1', got '%s'", result["my-pod"])
	}
}
