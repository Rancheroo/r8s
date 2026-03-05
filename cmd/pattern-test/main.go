// Package main provides a simple test harness for Sprint 11 patterns
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Rancheroo/r8s/internal/ai"
)

func main() {
	// Test cases for each pattern category
	testCases := []struct {
		name    string
		content string
		want    []string // Expected pattern IDs to match
	}{
		{
			name: "OOM + CrashLoop",
			content: `
				Out of memory: Kill process 1234 (nginx)
				Container in CrashLoopBackOff with 5 restarts
			`,
			want: []string{"oomkill-v2", "crashloopbackoff-v2"},
		},
		{
			name: "etcd Issues",
			content: `
				etcdserver: mvcc: database space exceeded
				etcd: slow request took 2.5s
				etcdserver: no leader
			`,
			want: []string{"etcd-corruption", "etcd-latency", "etcd-quorum"},
		},
		{
			name: "Certificate Problems",
			content: `
				x509: certificate has expired or is not yet valid
				certificate will expire in 10 days
			`,
			want: []string{"certificate-expired", "certificate-expiring"},
		},
		{
			name: "Network Issues",
			content: `
				dns resolution failed for service.default.svc.cluster.local
				cni plugin failed to set up sandbox
				connection timed out while connecting to backend
			`,
			want: []string{"dns-failure", "cni-error", "connectivity-timeout"},
		},
		{
			name: "Storage Issues",
			content: `
				persistentvolumeclaim my-pvc is not bound
				disk pressure detected on node
			`,
			want: []string{"pvc-binding-failure", "storage-pressure"},
		},
		{
			name: "Node & Pod States",
			content: `
				node worker-1 status NotReady
				Node has MemoryPressure condition
				pod status pending for 10 minutes
				failed to acquire lease kube-system/kube-controller-manager
			`,
			want: []string{"node-not-ready", "node-pressure", "pod-stuck-pending", "leader-election-failure"},
		},
		{
			name: "Image Pull Failures",
			content: `
				ImagePullBackOff: failed to pull image "nginx:nonexistent"
			`,
			want: []string{"imagepullbackoff-v2"},
		},
	}

	fmt.Println("Sprint 11 Pattern Engine Test")
	fmt.Println("=" + strings.Repeat("=", 50))

	allPassed := true

	for _, tc := range testCases {
		fmt.Printf("\nTest: %s\n", tc.name)
		fmt.Println(strings.Repeat("-", 40))

		// Run analysis
		registry := ai.NewRegistryV2()
		matches := registry.AnalyzeV2(tc.content)

		// Check expected patterns
		foundIDs := make(map[string]bool)
		for _, m := range matches {
			foundIDs[m.PatternID] = true
		}

		passed := true
		for _, wantID := range tc.want {
			if foundIDs[wantID] {
				fmt.Printf("  ✓ Found: %s\n", wantID)
			} else {
				fmt.Printf("  ✗ Missing: %s\n", wantID)
				passed = false
				allPassed = false
			}
		}

		// Check for unexpected correlations
		if len(matches) > 0 {
			fmt.Printf("  Total matches: %d\n", len(matches))
			for _, m := range matches {
				if len(m.Correlated) > 0 {
					fmt.Printf("  Correlation: %s ↔ %v\n", m.PatternID, m.Correlated)
				}
			}
		}

		if passed {
			fmt.Println("  [PASS]")
		} else {
			fmt.Println("  [FAIL]")
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	if allPassed {
		fmt.Println("All tests PASSED ✓")
		os.Exit(0)
	} else {
		fmt.Println("Some tests FAILED ✗")
		os.Exit(1)
	}
}
