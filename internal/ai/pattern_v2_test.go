package ai

import (
	"strings"
	"testing"
)

// TestPatternV2Matcher tests the v2 pattern matching system
// Sprint 11 Day 1: Pattern Engine v2 Architecture
func TestPatternV2Matcher(t *testing.T) {
	tests := []struct {
		name           string
		pattern        PatternV2
		content        string
		wantMatch      bool
		wantConfidence Confidence
	}{
		{
			name: "OOMKill detection - certain",
			pattern: PatternV2{
				ID:         "test-oom",
				Name:       "Test OOM",
				Category:   "OOM",
				Severity:   SeverityCritical,
				Confidence: ConfidenceCertain,
				Matchers: []Matcher{
					{Type: "keyword", Pattern: "out of memory", Weight: 1.0},
					{Type: "keyword", Pattern: "oomkill", Weight: 1.0},
				},
				Description: "Test OOM pattern",
			},
			content:        "The container was killed due to out of memory error",
			wantMatch:      true,
			wantConfidence: ConfidenceCertain,
		},
		{
			name: "ImagePull detection - certain",
			pattern: PatternV2{
				ID:         "test-image",
				Name:       "Test Image Pull",
				Category:   "Image",
				Severity:   SeverityWarning,
				Confidence: ConfidenceCertain,
				Matchers: []Matcher{
					{Type: "keyword", Pattern: "imagepullbackoff", Weight: 1.0},
					{Type: "keyword", Pattern: "failed to pull", Weight: 0.5},
				},
				Description: "Test image pull pattern",
			},
			content:        "Pod is in ImagePullBackOff state",
			wantMatch:      true,
			wantConfidence: ConfidenceCertain,
		},
		{
			name: "No match - empty content",
			pattern: PatternV2{
				ID:         "test-empty",
				Name:       "Test Empty",
				Category:   "Test",
				Severity:   SeverityInfo,
				Confidence: ConfidenceCertain,
				Matchers: []Matcher{
					{Type: "keyword", Pattern: "notfound", Weight: 1.0},
				},
				Description: "Test empty pattern",
			},
			content:   "This content has no matching keywords",
			wantMatch: false,
		},
		{
			name: "Partial match - possible confidence",
			pattern: PatternV2{
				ID:         "test-partial",
				Name:       "Test Partial",
				Category:   "Test",
				Severity:   SeverityWarning,
				Confidence: ConfidencePossible,
				Matchers: []Matcher{
					{Type: "keyword", Pattern: "keyword1", Weight: 1.0},
					{Type: "keyword", Pattern: "keyword2", Weight: 1.0},
					{Type: "keyword", Pattern: "keyword3", Weight: 1.0},
				},
				Description: "Test partial pattern",
			},
			content:        "Only keyword1 is present here",
			wantMatch:      true,
			wantConfidence: ConfidencePossible, // 1/3 matched = 33%
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewMatcherV2(tt.pattern)
			results := matcher.Match(tt.content)

			// Find first matched result
			var matchedResult *MatchResultV2
			for i := range results {
				if results[i].Matched {
					matchedResult = &results[i]
					break
				}
			}

			gotMatch := matchedResult != nil
			if gotMatch != tt.wantMatch {
				t.Errorf("Match() matched = %v, want %v", gotMatch, tt.wantMatch)
			}

			if matchedResult != nil && matchedResult.Confidence != tt.wantConfidence {
				t.Errorf("Match() confidence = %v, want %v", matchedResult.Confidence, tt.wantConfidence)
			}

			if matchedResult != nil && matchedResult.PatternID != tt.pattern.ID {
				t.Errorf("Match() PatternID = %v, want %v", matchedResult.PatternID, tt.pattern.ID)
			}
		})
	}
}

// TestPatternRegistryV2 tests the v2 pattern registry
func TestPatternRegistryV2(t *testing.T) {
	registry := NewRegistryV2()

	// Test that built-in patterns are loaded
	allPatterns := registry.GetAll()
	if len(allPatterns) == 0 {
		t.Error("Expected built-in patterns to be loaded")
	}

	// Test GetByID
	pattern, found := registry.GetByID("oomkill-v2")
	if !found {
		t.Error("Expected to find oomkill-v2 pattern")
	}
	if pattern.ID != "oomkill-v2" {
		t.Errorf("Expected ID oomkill-v2, got %s", pattern.ID)
	}

	// Test GetByCategory
	oomPatterns := registry.GetByCategory("OOM")
	if len(oomPatterns) == 0 {
		t.Error("Expected to find OOM category patterns")
	}
}

// TestAnalyzeV2 tests the v2 analysis with correlations
func TestAnalyzeV2(t *testing.T) {
	registry := NewRegistryV2()

	// Content that should trigger both crashloop and potentially correlate with OOM
	content := `
Pod is in CrashLoopBackOff state
Container was killed due to out of memory
Back-off restarting failed container
`

	results := registry.AnalyzeV2(content)

	// Should find at least crashloopbackoff-v2
	var foundCrashLoop bool
	var foundOOM bool
	for _, r := range results {
		if r.PatternID == "crashloopbackoff-v2" {
			foundCrashLoop = true
			// Check for correlation to OOM
			hasOOMCorrelation := false
			for _, corr := range r.Correlated {
				if corr == "oomkill-v2" {
					hasOOMCorrelation = true
					break
				}
			}
			if !hasOOMCorrelation {
				t.Error("Expected CrashLoopBackOff to correlate with OOMKill")
			}
		}
		if r.PatternID == "oomkill-v2" {
			foundOOM = true
		}
	}

	if !foundCrashLoop {
		t.Error("Expected to find CrashLoopBackOff pattern")
	}
	if !foundOOM {
		t.Error("Expected to find OOMKill pattern")
	}
}

// TestRegisterV2 tests pattern registration
func TestRegisterV2(t *testing.T) {
	registry := NewRegistryV2()

	newPattern := PatternV2{
		ID:          "test-pattern",
		Name:        "Test Pattern",
		Category:    "Test",
		Severity:    SeverityInfo,
		Confidence:  ConfidenceCertain,
		Matchers:    []Matcher{{Type: "keyword", Pattern: "test", Weight: 1.0}},
		Description: "A test pattern",
	}

	err := registry.Register(newPattern)
	if err != nil {
		t.Errorf("Register() error = %v", err)
	}

	// Verify it was added
	retrieved, found := registry.GetByID("test-pattern")
	if !found {
		t.Error("Expected to find registered pattern")
	}
	if retrieved.Name != "Test Pattern" {
		t.Errorf("Expected name 'Test Pattern', got %s", retrieved.Name)
	}
}

// TestRegisterV2Validation tests pattern validation
func TestRegisterV2Validation(t *testing.T) {
	registry := NewRegistryV2()

	tests := []struct {
		name    string
		pattern PatternV2
		wantErr bool
	}{
		{
			name:    "Empty ID",
			pattern: PatternV2{Name: "Test", Matchers: []Matcher{{Type: "keyword", Pattern: "test", Weight: 1.0}}},
			wantErr: true,
		},
		{
			name:    "Empty Name",
			pattern: PatternV2{ID: "test", Matchers: []Matcher{{Type: "keyword", Pattern: "test", Weight: 1.0}}},
			wantErr: true,
		},
		{
			name:    "No Matchers",
			pattern: PatternV2{ID: "test", Name: "Test"},
			wantErr: true,
		},
		{
			name:    "Valid Pattern",
			pattern: PatternV2{ID: "test", Name: "Test", Matchers: []Matcher{{Type: "keyword", Pattern: "test", Weight: 1.0}}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.Register(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestExtractEvidence tests evidence extraction
func TestExtractEvidence(t *testing.T) {
	matcher := NewMatcherV2(PatternV2{
		ID:   "test",
		Name: "Test",
	})

	content := `Line one
Line two with keyword here
Line three`

	evidence := matcher.extractEvidence(Matcher{Type: "keyword", Pattern: "keyword"}, content)

	if !strings.Contains(evidence, "keyword") {
		t.Errorf("Expected evidence to contain 'keyword', got: %s", evidence)
	}

	// Test truncation for long lines
	longLine := strings.Repeat("a", 250)
	contentLong := "Short line\n" + longLine + "\nAnother line"
	evidenceLong := matcher.extractEvidence(Matcher{Type: "keyword", Pattern: longLine[:10]}, contentLong)

	if len(evidenceLong) > 210 {
		t.Errorf("Expected evidence to be truncated, got length %d", len(evidenceLong))
	}
	if !strings.HasSuffix(evidenceLong, "...") {
		t.Errorf("Expected evidence to end with '...', got: %s", evidenceLong)
	}
}

// BenchmarkAnalyzeV2 benchmarks the v2 analysis performance
func BenchmarkAnalyzeV2(b *testing.B) {
	registry := NewRegistryV2()
	content := strings.Repeat("Test content with crashloopbackoff and out of memory errors. ", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.AnalyzeV2(content)
	}
}

// Sprint 11 Day 2: etcd Pattern Tests

func TestEtcdCorruptionPattern(t *testing.T) {
	registry := NewRegistryV2()

	tests := []struct {
		name      string
		content   string
		wantMatch bool
	}{
		{
			name:      "Database space exceeded",
			content:   "etcdserver: mvcc: database space exceeded",
			wantMatch: true,
		},
		{
			name:      "Data corruption",
			content:   "etcd data corruption detected in member data",
			wantMatch: true,
		},
		{
			name:      "Raft log corruption",
			content:   "raft: log corruption detected",
			wantMatch: true,
		},
		{
			name:      "No match",
			content:   "etcd is running normally",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := registry.AnalyzeV2(tt.content)
			found := false
			for _, r := range results {
				if r.PatternID == "etcd-corruption" {
					found = true
					break
				}
			}
			if found != tt.wantMatch {
				t.Errorf("etcd-corruption match = %v, want %v", found, tt.wantMatch)
			}
		})
	}
}

func TestEtcdLatencyPattern(t *testing.T) {
	registry := NewRegistryV2()

	content := "etcd: slow request took 2.5s, read index took too long (5s)"
	results := registry.AnalyzeV2(content)

	var found bool
	for _, r := range results {
		if r.PatternID == "etcd-latency" {
			found = true
			if r.Confidence != ConfidenceLikely {
				t.Errorf("Expected Likely confidence, got %v", r.Confidence)
			}
			break
		}
	}
	if !found {
		t.Error("Expected to find etcd-latency pattern")
	}
}

func TestEtcdQuorumPattern(t *testing.T) {
	registry := NewRegistryV2()

	content := "etcdserver: no leader, raft: no elected leader cluster is unavailable"
	results := registry.AnalyzeV2(content)

	var found bool
	for _, r := range results {
		if r.PatternID == "etcd-quorum" {
			found = true
			if r.Confidence != ConfidenceCertain {
				t.Errorf("Expected Certain confidence, got %v", r.Confidence)
			}
			break
		}
	}
	if !found {
		t.Error("Expected to find etcd-quorum pattern")
	}
}

// Sprint 11 Day 2: Certificate Pattern Tests

func TestCertificateExpiredPattern(t *testing.T) {
	registry := NewRegistryV2()

	content := "x509: certificate has expired or is not yet valid: current time 2026-02-23 is after 2026-02-20"
	results := registry.AnalyzeV2(content)

	var found bool
	for _, r := range results {
		if r.PatternID == "certificate-expired" {
			found = true
			if r.Confidence != ConfidenceCertain {
				t.Errorf("Expected Certain confidence, got %v", r.Confidence)
			}
			// TODO: Re-enable correlation check when node-not-ready pattern is added
			break
		}
	}
	if !found {
		t.Error("Expected to find certificate-expired pattern")
	}
}

func TestCertificateExpiringPattern(t *testing.T) {
	registry := NewRegistryV2()

	content := "Warning: certificate will expire in 10 days"
	results := registry.AnalyzeV2(content)

	var found bool
	for _, r := range results {
		if r.PatternID == "certificate-expiring" {
			found = true
			if r.Confidence != ConfidenceLikely {
				t.Errorf("Expected Likely confidence, got %v", r.Confidence)
			}
			break
		}
	}
	if !found {
		t.Error("Expected to find certificate-expiring pattern")
	}
}

func TestCertificateInvalidCAPattern(t *testing.T) {
	registry := NewRegistryV2()

	content := "x509: certificate signed by unknown authority"
	results := registry.AnalyzeV2(content)

	var found bool
	for _, r := range results {
		if r.PatternID == "certificate-invalid-ca" {
			found = true
			if r.Confidence != ConfidenceCertain {
				t.Errorf("Expected Certain confidence, got %v", r.Confidence)
			}
			break
		}
	}
	if !found {
		t.Error("Expected to find certificate-invalid-ca pattern")
	}
}

// Sprint 11 Day 3: Networking Pattern Tests

func TestDNSFailurePattern(t *testing.T) {
	registry := NewRegistryV2()

	tests := []struct {
		name      string
		content   string
		wantMatch bool
	}{
		{
			name:      "DNS resolution failed",
			content:   "Error: dns resolution failed for service.namespace.svc.cluster.local",
			wantMatch: true,
		},
		{
			name:      "Lookup failed",
			content:   "lookup failed: could not resolve hostname example.com",
			wantMatch: true,
		},
		{
			name:      "CoreDNS mentioned",
			content:   "coredns is having issues resolving names",
			wantMatch: false, // coredns keyword removed - matches pod names not errors
		},
		{
			name:      "No match",
			content:   "DNS is working correctly",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := registry.AnalyzeV2(tt.content)
			found := false
			for _, r := range results {
				if r.PatternID == "dns-failure" {
					found = true
					break
				}
			}
			if found != tt.wantMatch {
				t.Errorf("dns-failure match = %v, want %v", found, tt.wantMatch)
			}
		})
	}
}

func TestCNIErrorPattern(t *testing.T) {
	registry := NewRegistryV2()

	tests := []struct {
		name      string
		content   string
		wantMatch bool
	}{
		{
			name:      "CNI plugin failed",
			content:   "Failed to create pod sandbox: cni plugin failed",
			wantMatch: true,
		},
		{
			name:      "Calico error",
			content:   "calico-node pod is in CrashLoopBackOff",
			wantMatch: false, // calico keyword removed - matches pod names not errors
		},
		{
			name:      "Cilium error",
			content:   "cilium agent failed to start",
			wantMatch: false, // cilium keyword removed - matches pod names not errors
		},
		{
			name:      "Network plugin",
			content:   "networkplugin cni failed to setup",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := registry.AnalyzeV2(tt.content)
			found := false
			for _, r := range results {
				if r.PatternID == "cni-error" {
					found = true
					if r.Confidence != ConfidenceCertain {
						t.Errorf("Expected Certain confidence, got %v", r.Confidence)
					}
					break
				}
			}
			if found != tt.wantMatch {
				t.Errorf("cni-error match = %v, want %v", found, tt.wantMatch)
			}
		})
	}
}

func TestConnectivityTimeoutPattern(t *testing.T) {
	registry := NewRegistryV2()

	content := "Connection timed out while connecting to backend service. i/o timeout after 30s"
	results := registry.AnalyzeV2(content)

	var found bool
	for _, r := range results {
		if r.PatternID == "connectivity-timeout" {
			found = true
			if r.Confidence != ConfidenceLikely {
				t.Errorf("Expected Likely confidence, got %v", r.Confidence)
			}
			break
		}
	}
	if !found {
		t.Error("Expected to find connectivity-timeout pattern")
	}
}

// Sprint 11 Day 3: Storage Pattern Tests

func TestPVCBindingFailurePattern(t *testing.T) {
	registry := NewRegistryV2()

	tests := []struct {
		name      string
		content   string
		wantMatch bool
	}{
		{
			name:      "PVC not bound",
			content:   "persistentvolumeclaim my-pvc is not bound",
			wantMatch: true,
		},
		{
			name:      "Failed to provision",
			content:   "failed to provision volume with StorageClass standard",
			wantMatch: true,
		},
		{
			name:      "No PVs available",
			content:   "no persistent volumes available for this claim",
			wantMatch: true,
		},
		{
			name:      "Waiting for volume",
			content:   "waiting for a volume to be created",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := registry.AnalyzeV2(tt.content)
			found := false
			for _, r := range results {
				if r.PatternID == "pvc-binding-failure" {
					found = true
					break
				}
			}
			if found != tt.wantMatch {
				t.Errorf("pvc-binding-failure match = %v, want %v", found, tt.wantMatch)
			}
		})
	}
}

func TestStoragePressurePattern(t *testing.T) {
	registry := NewRegistryV2()

	tests := []struct {
		name      string
		content   string
		wantMatch bool
	}{
		{
			name:      "Disk pressure",
			content:   "Node has disk pressure condition",
			wantMatch: true,
		},
		{
			name:      "No space left",
			content:   "filesystem has no space left on device",
			wantMatch: true,
		},
		{
			name:      "Inode pressure",
			content:   "InodePressure condition detected on node",
			wantMatch: true,
		},
		{
			name:      "Runtime GC failed",
			content:   "runtime garbage collection failed due to disk pressure",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := registry.AnalyzeV2(tt.content)
			found := false
			for _, r := range results {
				if r.PatternID == "storage-pressure" {
					found = true
					if r.Confidence != ConfidenceCertain {
						t.Errorf("Expected Certain confidence, got %v", r.Confidence)
					}
					break
				}
			}
			if found != tt.wantMatch {
				t.Errorf("storage-pressure match = %v, want %v", found, tt.wantMatch)
			}
		})
	}
}

// Sprint 11 Day 4: Node & Pod State Pattern Tests

func TestNodePressurePattern(t *testing.T) {
	registry := NewRegistryV2()

	tests := []struct {
		name      string
		content   string
		wantMatch bool
	}{
		{
			name:      "DiskPressure",
			content:   "Node has DiskPressure condition",
			wantMatch: true,
		},
		{
			name:      "MemoryPressure",
			content:   "MemoryPressure detected on node",
			wantMatch: true,
		},
		{
			name:      "PIDPressure",
			content:   "PIDPressure condition is true",
			wantMatch: true,
		},
		{
			name:      "OutOfDisk",
			content:   "Node has OutOfDisk condition",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := registry.AnalyzeV2(tt.content)
			found := false
			for _, r := range results {
				if r.PatternID == "node-pressure" {
					found = true
					if r.Confidence != ConfidenceCertain {
						t.Errorf("Expected Certain confidence, got %v", r.Confidence)
					}
					break
				}
			}
			if found != tt.wantMatch {
				t.Errorf("node-pressure match = %v, want %v", found, tt.wantMatch)
			}
		})
	}
}

func TestPodStuckPendingPattern(t *testing.T) {
	registry := NewRegistryV2()

	tests := []struct {
		name      string
		content   string
		wantMatch bool
	}{
		{
			name:      "Pod status pending",
			content:   "pod status pending - waiting for scheduling",
			wantMatch: true,
		},
		{
			name:      "0/1 nodes available",
			content:   "0/1 nodes are available: 1 Insufficient cpu",
			wantMatch: true,
		},
		{
			name:      "Insufficient memory",
			content:   "insufficient memory on available nodes",
			wantMatch: true,
		},
		{
			name:      "Failed to pull image",
			content:   "failed to pull image causing pod to be pending",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := registry.AnalyzeV2(tt.content)
			found := false
			for _, r := range results {
				if r.PatternID == "pod-stuck-pending" {
					found = true
					break
				}
			}
			if found != tt.wantMatch {
				t.Errorf("pod-stuck-pending match = %v, want %v", found, tt.wantMatch)
			}
		})
	}
}

func TestPodStuckTerminatingPattern(t *testing.T) {
	registry := NewRegistryV2()

	tests := []struct {
		name      string
		content   string
		wantMatch bool
	}{
		{
			name:      "Pod status terminating",
			content:   "pod status terminating for 5 minutes",
			wantMatch: true,
		},
		{
			name:      "Deletion timestamp",
			content:   "deletiontimestamp set but pod not deleted",
			wantMatch: true,
		},
		{
			name:      "Failed to delete",
			content:   "failed to delete pod due to finalizer",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := registry.AnalyzeV2(tt.content)
			found := false
			for _, r := range results {
				if r.PatternID == "pod-stuck-terminating" {
					found = true
					break
				}
			}
			if found != tt.wantMatch {
				t.Errorf("pod-stuck-terminating match = %v, want %v", found, tt.wantMatch)
			}
		})
	}
}

func TestLeaderElectionFailurePattern(t *testing.T) {
	registry := NewRegistryV2()

	tests := []struct {
		name      string
		content   string
		wantMatch bool
	}{
		{
			name:      "Failed to acquire lease",
			content:   "failed to acquire lease kube-system/kube-controller-manager",
			wantMatch: true,
		},
		{
			name:      "Leader election lost",
			content:   "leader election lost for kube-scheduler",
			wantMatch: true,
		},
		{
			name:      "Failed to renew lease",
			content:   "failed to renew lease kube-system/kube-controller-manager: timed out",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := registry.AnalyzeV2(tt.content)
			found := false
			for _, r := range results {
				if r.PatternID == "leader-election-failure" {
					found = true
					if r.Confidence != ConfidenceCertain {
						t.Errorf("Expected Certain confidence, got %v", r.Confidence)
					}
					break
				}
			}
			if found != tt.wantMatch {
				t.Errorf("leader-election-failure match = %v, want %v", found, tt.wantMatch)
			}
		})
	}
}

func TestNodeNotReadyPattern(t *testing.T) {
	registry := NewRegistryV2()

	tests := []struct {
		name      string
		content   string
		wantMatch bool
	}{
		{
			name:      "Node NotReady",
			content:   "node worker-1 status NotReady",
			wantMatch: true,
		},
		{
			name:      "Kubelet stopped posting",
			content:   "kubelet stopped posting node status",
			wantMatch: true,
		},
		{
			name:      "Node controller marked",
			content:   "node controller marked node as NotReady",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := registry.AnalyzeV2(tt.content)
			found := false
			for _, r := range results {
				if r.PatternID == "node-not-ready" {
					found = true
					if r.Confidence != ConfidenceCertain {
						t.Errorf("Expected Certain confidence, got %v", r.Confidence)
					}
					break
				}
			}
			if found != tt.wantMatch {
				t.Errorf("node-not-ready match = %v, want %v", found, tt.wantMatch)
			}
		})
	}
}

// TestPatternCount verifies all 19 patterns are present
func TestPatternCount(t *testing.T) {
	registry := NewRegistryV2()
	patterns := registry.GetAll()

	// Should have at least 19 patterns (3 base + 3 etcd + 3 cert + 3 net + 2 storage + 5 node/pod/control)
	if len(patterns) < 19 {
		t.Errorf("Expected at least 19 patterns, got %d", len(patterns))
	}

	// Verify specific patterns exist
	expectedPatterns := []string{
		"oomkill-v2",
		"imagepullbackoff-v2",
		"crashloopbackoff-v2",
		"etcd-corruption",
		"etcd-latency",
		"etcd-quorum",
		"certificate-expired",
		"certificate-expiring",
		"certificate-invalid-ca",
		"dns-failure",
		"cni-error",
		"connectivity-timeout",
		"pvc-binding-failure",
		"storage-pressure",
		"node-pressure",
		"pod-stuck-pending",
		"pod-stuck-terminating",
		"leader-election-failure",
		"node-not-ready",
	}

	for _, id := range expectedPatterns {
		_, found := registry.GetByID(id)
		if !found {
			t.Errorf("Expected to find pattern: %s", id)
		}
	}
}
