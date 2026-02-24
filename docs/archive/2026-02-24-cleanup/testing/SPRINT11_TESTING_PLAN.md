# Sprint 11 Testing Plan

## Phase 1: Build & Unit Tests (5 min)

```bash
# Navigate to r8s
cd /home/bradmin/workspace/r8s

# Set up Go
export PATH="/usr/lib/go-1.22/bin:$PATH"

# Build AI package
go build ./internal/ai/...

# Run all unit tests
go test ./internal/ai/... -v

# Check coverage
go test ./internal/ai/... -cover
```

**Expected:**
- Build succeeds without errors
- 31 tests pass
- Coverage >60%

---

## Phase 2: Manual Pattern Test (5 min)

```bash
# Run pattern test harness
go run ./cmd/pattern-test/main.go
```

**Tests:**
1. OOM + CrashLoop patterns
2. etcd (corruption, latency, quorum)
3. Certificates (expired, expiring)
4. Networking (DNS, CNI, timeout)
5. Storage (PVC, disk pressure)
6. Node & Pod states

**Expected:** All 7 test cases pass

---

## Phase 3: Hint Generation Test (5 min)

```bash
# Create test file
cat > /tmp/test_hints.sh << 'EOF'
package main

import (
	"fmt"
	"github.com/Rancheroo/r8s/internal/ai"
)

func main() {
	// Simulate analysis
	analyzer := ai.NewAnalyzer()
	
	// Test content with OOM
	content := `Out of memory: Kill process 1234 (nginx)`
	
	result, err := analyzer.Analyze(content, ai.AnalysisOptions{})
	if err != nil {
		panic(err)
	}
	
	// Print results
	fmt.Println(analyzer.FormatResults(result))
}
EOF

go run /tmp/test_hints.sh
```

**Expected:**
- Pattern detected
- Hint generated with suggestion
- Command provided
- References listed

---

## Phase 4: Bundle Analysis Test (10 min)

```bash
# If you have a real bundle:
mkdir -p /tmp/test-analysis
cd /tmp/test-analysis

# Extract bundle (if you have one)
# tar -xzf ~/support-bundle.tar.gz

# Run hypothetical analysis
cat > /tmp/analyze_bundle.go << 'EOF'
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/Rancheroo/r8s/internal/ai"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: analyze-bundle <bundle-dir>")
		os.Exit(1)
	}
	
	bundleDir := os.Args[1]
	analyzer := ai.NewAnalyzer()
	
	// Quick test: scan journald logs
	logFiles := []string{
		"journald/rke2-server.log",
		"journald/k3s-agent.log",
	}
	
	for _, logFile := range logFiles {
		fullPath := filepath.Join(bundleDir, logFile)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			continue
		}
		
		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}
		
		result, _ := analyzer.Analyze(string(content), ai.AnalysisOptions{})
		if len(result.Hints) > 0 {
			fmt.Printf("\n=== %s ===\n", logFile)
			fmt.Println(analyzer.FormatResults(result))
		}
	}
}
EOF

go run /tmp/analyze_bundle.go /path/to/your/bundle
```

---

## Phase 5: Real Cluster Issues (Optional - 10 min)

### Option A: Existing Bundle
```bash
# Find real bundles in your workspace
find /home/bradmin -name "*.tar.gz" -path "*/bundle*" 2>/dev/null | head -5

# Extract and analyze
# tar -xzf <bundle> -C /tmp/bundle-test
# go run /tmp/analyze_bundle.go /tmp/bundle-test
```

### Option B: Test Against Sample Log Files
```bash
# Create synthetic cluster data
mkdir -p /tmp/test-cluster/journald /tmp/test-cluster/pod-logs

# Create test log with OOM issue
cat > /tmp/test-cluster/journald/rke2-server.log << 'EOF'
Feb 23 10:00:00 server1 rke2[1234]: I0223 10:00:00.123456    1234 kubelet.go:1234] kubelet started
Feb 23 10:05:00 server1 rke2[1234]: E0223 10:05:00.123456    1234 oomkill.go:123] Out of memory: Kill process 5678 (worker-pod) score 912 or sacrifice child
Feb 23 10:05:01 server1 rke2[1234]: E0223 10:05:01.123456    1234 event.go:123] Pod worker-0 in namespace default was killed due to out of memory
EOF

# Create etcd log with latency
cat > /tmp/test-cluster/journald/etcd.log << 'EOF'
Feb 23 10:10:00 server1 etcd[1000]: etcd: slow request took 2.5s
Feb 23 10:10:01 server1 etcd[1000]: etcd: read index took too long (5.123s)
EOF

# Analyze
# go run /tmp/analyze_bundle.go /tmp/test-cluster
```

---

## Phase 6: Validation Checklist

| Test | Expected | Actual |
|------|----------|--------|
| Build succeeds | ✓ | |
| Unit tests pass | 31/31 | |
| Coverage >60% | ✓ | |
| Pattern test passes | 7/7 | |
| OOM pattern detected | ✓ | |
| CrashLoop pattern detected | ✓ | |
| etcd patterns (3) detected | ✓ | |
| Certificate patterns (2) detected | ✓ | |
| Network patterns (3) detected | ✓ | |
| Storage patterns (2) detected | ✓ | |
| Node/Pod patterns (4) detected | ✓ | |
| Correlations detected | ✓ | |
| Hints generated | ✓ | |
| Markdown export works | ✓ | |
| Performance <1s | ✓ | |

---

## Troubleshooting

### Build Fails
```bash
# Check Go version
go version

# Should be 1.22+
```

### Tests Fail
```bash
# Run specific test
go test ./internal/ai/... -run TestAnalyzeV2 -v

# Check for syntax errors
go vet ./internal/ai/...
```

### Patterns Not Matching
```bash
# Debug: print all registered patterns
go run <<'EOF'
package main
import (
	"fmt"
	"github.com/Rancheroo/r8s/internal/ai"
)
func main() {
	registry := ai.NewRegistryV2()
	for _, p := range registry.GetAll() {
		fmt.Printf("%s: %s\n", p.ID, p.Name)
	}
}
EOF
```

---

## Sign-off

**Pattern Engine:** ____(initials) Date: _______

**Hint System:** ____(initials) Date: _______

**Integration:** ____(initials) Date: _______