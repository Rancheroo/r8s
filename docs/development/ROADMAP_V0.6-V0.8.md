# r8s Development Roadmap: v0.6.x → v0.8.x

**Roadmap Created**: 2026-01-14  
**Based On**: PRINCIPLES_AUDIT.md + FUTURE_WORK.md analysis  
**Time Horizon**: Q1-Q3 2026 (6-9 months)

This document provides a comprehensive roadmap for r8s releases v0.6.9 through v0.8.0, focusing on principle compliance, testing infrastructure, and feature maturity.

---

## 🎯 Vision & Goals

### v0.6.x Series: **"Principle Compliance & Quality Foundation"**
**Theme**: Fix principle violations, establish testing baseline  
**Target**: March 2026  
**Key Metrics**: 98% principle compliance, 30% test coverage

### v0.7.x Series: **"Testing Infrastructure & Robustness"**
**Theme**: Build comprehensive test suite, CI/CD automation  
**Target**: May 2026  
**Key Metrics**: 50% test coverage, automated quality gates

### v0.8.x Series: **"Production Hardening"**
**Theme**: Performance optimization, 80% coverage, scale testing  
**Target**: August 2026  
**Key Metrics**: 80% test coverage, <2s dashboard load with 1000 pods

---

## 📦 v0.6.10 "Architecture Refactor & K3s Support"

**Release Date**: Late February 2026  
**Duration**: 3-4 week sprint  
**Effort**: ~30-40 hours total  
**Theme**: Inherited codebase review, cleanup, and multi-distro support

### Objectives
1. ✅ Complete codebase review (inherited from previous maintainer)
2. ✅ Identify and eliminate redundant code, dead code, unnecessary abstractions
3. ✅ Add K3s bundle compatibility (path abstraction: `rke2/` → `k3s/`)
4. ✅ Consolidate duplicate implementations
5. ✅ Improve code organization and reduce technical debt

### Context
This release addresses the reality that r8s was inherited from a previous developer who left the project. The codebase likely contains:
- Unused functions and types
- Over-engineered abstractions
- Duplicated logic across packages
- Hardcoded assumptions (like `rke2/` paths)
- Inconsistent patterns

This is a **big fix release** — focused on cleanup, not features.

### Implementation Tasks

#### Phase 1: Codebase Audit & Analysis (Week 1, ~8 hours)

**Task 1a: Static Analysis & Dead Code Detection**
```
Tools to use:
- golang.org/x/tools/cmd/deadcode (find unused code)
- go vet ./... (standard analysis)
- staticcheck ./... (deep analysis)
- gocyclo (complexity detection)

Files to create:
- docs/development/CODEBASE_AUDIT.md (findings log)
- scripts/audit-codebase.sh (automated checks)

Areas to investigate:
1. Unused exports (functions, types, constants)
2. Functions called only from tests
3. Interfaces with only one implementation
4. Over-complex functions (cyclomatic >15)
5. Duplicate code blocks
6. Unused imports and dependencies

Estimated Time: 4 hours
```

**Task 1b: Bundle Path Hardcoding Audit**
```
Current issue: 13 hardcoded "rke2" path references

Files to examine:
- internal/bundle/manifest.go (DetectFormat, path construction)
- internal/bundle/validate.go (ValidateBundle)
- internal/bundle/oom.go (event/pod paths)
- internal/datasource/bundle_test.go (test fixtures)

K3s bundle structure (same as RKE2, just different prefix):
  k3s/kubectl/pods
  k3s/podlogs/
  k3s/kubectl/poddescribe/
  k3s/pod-manifests/
  systeminfo/
  systemlogs/

Action: Create distro-agnostic path abstraction

Estimated Time: 2 hours
```

**Task 1c: Dependency & Import Analysis**
```
Check for:
- Unused dependencies in go.mod
- Import cycles
- Heavy dependencies that could be replaced
- Internal imports that bypass interfaces

Commands:
  go mod tidy
  go list -m all | wc -l  (count dependencies)
  go mod why <package>    (understand imports)

Estimated Time: 2 hours
```

#### Phase 2: K3s Compatibility Implementation (Week 2, ~10 hours)

**Task 2a: Bundle Format Extension**
```
File: internal/bundle/types.go

Add:
  FormatK3s BundleFormat = "k3s-support-bundle"

Update Bundle struct:
  DistroDir string // "rke2" or "k3s" - set during detection
  DistroType BundleFormat // FormatRKE2 or FormatK3s

Estimated Time: 1 hour
```

**Task 2b: Dynamic Path Resolution**
```
File: internal/bundle/manifest.go (and all path constructors)

Pattern: Replace hardcoded paths with helper function

BEFORE:
  podsPath := filepath.Join(bundleRoot, "rke2", "kubectl", "pods")
  eventsPath := filepath.Join(bundleRoot, "rke2", "kubectl", "events")
  podlogsDir := filepath.Join(bundleRoot, "rke2", "podlogs")

AFTER:
  podsPath := b.KubectlPath("pods")
  eventsPath := b.KubectlPath("events")
  podlogsDir := b.PodLogsPath()

New methods on Bundle:
  func (b *Bundle) KubectlPath(resource string) string
  func (b *Bundle) PodLogsPath() string
  func (b *Bundle) PodManifestsPath() string
  func (b *Bundle) PodDescribePath() string

Estimated Time: 3 hours
```

**Task 2c: Multi-Distro Detection**
```
File: internal/bundle/manifest.go

Update DetectFormat():
  1. Check for rke2/ directory → FormatRKE2
  2. Check for k3s/ directory → FormatK3s
  3. Check for wrapper directories containing either
  4. Return FormatUnknown if neither found

Update ValidateBundle():
  Accept either rke2 or k3s structure
  Validate based on detected distro

Estimated Time: 2 hours
```

**Task 2d: Test Fixtures for K3s**
```
Create test bundles:
- testdata/bundles/minimal-k3s/ (minimal valid K3s bundle)
- testdata/bundles/full-k3s/ (complete K3s bundle)

Test cases:
- TestDetectFormat_K3sBundle
- TestLoadBundle_K3sStructure
- TestInventoryPods_K3sPaths
- Validate cross-distro consistency

Estimated Time: 4 hours
```

#### Phase 3: Code Consolidation & Cleanup (Week 3, ~12 hours)

**Task 3a: Remove Dead Code**
```
Based on audit findings:

Common targets:
- Unused exported functions
- Types with no consumers
- Constants only referenced in tests
- "Future use" code that's never used
- Commented-out legacy code

Safety: 
- Only remove if no references in non-test code
- Keep if part of public API (for now)
- Document removals in CHANGELOG

Estimated Time: 3 hours
```

**Task 3b: Consolidate Duplicate Implementations**
```
Known areas to check:

1. Container name parsing:
   - parsePodManifestsForContainers() in manifest.go
   - Similar logic in other files?
   → Consolidate into single parser

2. Pod log filename parsing:
   - parsePodLogFilename() in manifest.go
   - May exist elsewhere
   → Single source of truth

3. Namespace extraction:
   - Various string splitting logic
   → Use shared utility

4. Path construction:
   - filepath.Join() scattered everywhere
   → Use Bundle helper methods

Estimated Time: 4 hours
```

**Task 3c: Simplify Over-Engineered Code**
```
Look for:
- Interfaces with single implementation
- Unnecessary factory patterns
- Complex type hierarchies
- Premature abstractions

Specific areas (from initial review):
- DataSource interface implementations
  - BundleDataSource
  - EmbeddedDataSource  
  - Could some methods be consolidated?

- TUI state management
  - Multiple similar message types
  - Over-complicated update chains

Estimated Time: 3 hours
```

**Task 3d: Standardize Error Handling**
```
Current inconsistencies:
- Some functions return (nil, error)
- Some return (empty, nil)
- Some panic
- Some log and continue

Standardize per Principle #11 (Empty is Valid):
- Not found → return empty, nil
- Parse error → return error with context
- Critical failure → return error

Create helpers:
  func IsNotFound(err error) bool
  func WrapError(context string, err error) error

Estimated Time: 2 hours
```

#### Phase 4: Documentation & Validation (Week 4, ~8 hours)

**Task 4a: Update Bundle Documentation**
```
Files:
- docs/development/BUNDLE_DEPENDENCY_ANALYSIS.md
  - Add K3s section
  - Document path abstractions
  - Update collector script reference

- README.md
  - Mention K3s support
  - Update bundle compatibility table

Estimated Time: 2 hours
```

**Task 4b: Create Architecture Decision Records**
```
New file: docs/development/ADR/

ADR-001: Multi-Distro Bundle Support
- Why we added K3s
- Path abstraction approach
- Trade-offs considered

ADR-002: Code Cleanup Guidelines
- How we identify dead code
- When to keep vs remove
- Safety procedures

Estimated Time: 2 hours
```

**Task 4c: Regression Testing**
```
Validate nothing broke:

1. RKE2 bundles still work
   - Load example bundle
   - Verify all features functional
   - Performance unchanged

2. K3s bundles now work
   - Load K3s test bundle
   - Verify feature parity

3. Edge cases
   - Mixed/wrapper directories
   - Missing subdirectories
   - Corrupted paths

Estimated Time: 3 hours
```

**Task 4d: Cleanup Documentation**
```
Create CODEBASE_AUDIT_SUMMARY.md:
- What was found
- What was removed/changed
- Lessons learned
- Recommendations for future

Update CHANGELOG.md with v0.6.10 entry

Estimated Time: 1 hour
```

### Success Criteria
- [ ] Codebase audit complete (findings documented)
- [ ] K3s bundles load and parse correctly
- [ ] RKE2 bundles continue to work (regression tested)
- [ ] Dead code removed (target: 5-10% reduction)
- [ ] Duplicate logic consolidated
- [ ] All paths use distro-agnostic abstractions
- [ ] Test coverage maintained or improved
- [ ] Documentation updated

### Release Prompt for v0.6.10

```markdown
# v0.6.10 Release Prompt: Architecture Refactor & K3s Support

## Objectives
Clean up inherited codebase, add K3s compatibility, and establish 
maintainable architecture for multi-distro support.

## Implementation Checklist

### Codebase Audit (8 hours)
1. Run static analysis tools (deadcode, staticcheck, gocyclo)
2. Document findings in CODEBASE_AUDIT.md
3. Identify all hardcoded "rke2" paths (13 occurrences)
4. Catalog redundant/duplicate implementations
5. List over-engineered abstractions

### K3s Support (10 hours)
6. Add FormatK3s to bundle/types.go
7. Extend DetectFormat() for k3s/ directory
8. Create Bundle.DistroDir and path helper methods
9. Replace all hardcoded "rke2" with dynamic paths
10. Create K3s test bundles and validation tests
11. Update ValidateBundle() for multi-distro

### Code Cleanup (12 hours)
12. Remove dead code identified in audit
13. Consolidate duplicate implementations:
    - Container name parsers
    - Log filename parsers
    - Namespace extraction utilities
14. Simplify over-engineered code
15. Standardize error handling patterns

### Documentation & Testing (8 hours)
16. Update BUNDLE_DEPENDENCY_ANALYSIS.md with K3s
17. Create ADR-001 (Multi-Distro Support)
18. Create ADR-002 (Code Cleanup Guidelines)
19. Regression test: RKE2 bundles still work
20. Validate: K3s bundles work correctly
21. Update CHANGELOG.md
22. Create CODEBASE_AUDIT_SUMMARY.md

## Acceptance Criteria
- K3s and RKE2 bundles both load correctly
- Zero hardcoded distro paths (all use abstractions)
- Code size reduced (fewer lines, simpler structure)
- All tests passing
- No performance regressions
- Documentation complete

## Definition of Done
- [ ] Can load and analyze K3s support bundles
- [ ] Can load and analyze RKE2 support bundles (regression)
- [ ] Code audit findings addressed
- [ ] Technical debt reduced
- [ ] Architecture ready for future distros (RKE1, etc.)

## Estimated Effort: 30-40 hours
```

---

## 📦 v0.6.9 "Principle Compliance Sprint"

**Release Date**: Mid-February 2026  
**Duration**: 1 week sprint  
**Effort**: ~10-15 hours total

### Objectives
1. ✅ Achieve 98% principle compliance (from 76%)
2. ✅ Add regression tests for all fixes
3. ✅ Increase test coverage to 30% (from ~13%)
4. ✅ Zero critical principle violations

### Implementation Tasks

#### Phase 1: Critical Fixes (Day 1, ~2 hours)

**Task 1a: Replace Bubble Sort with sort.Slice**
```
File: internal/datasource/bundle.go
Function: GetEventsByPod()
Lines: ~377-395

BEFORE (O(n²) bubble sort):
for i := 0; i < len(podEvents)-1; i++ {
    for j := 0; j < len(podEvents)-i-1; j++ {
        // sorting logic
    }
}

AFTER (O(n log n) stdlib):
sort.Slice(podEvents, func(i, j int) bool {
    // Warnings before Normal
    if podEvents[i].Type != podEvents[j].Type {
        return podEvents[i].Type == "Warning"
    }
    // Same type: most recent first
    return podEvents[i].LastSeen > podEvents[j].LastSeen
})

Estimated Time: 15 minutes
```

**Task 1b: Remove Mock Data Fallbacks**
```
File: internal/datasource/bundle.go
Functions: DescribeDeployment(), DescribeService()
Lines: ~262-290

BEFORE (returns fabricated data):
return map[string]interface{}{
    "apiVersion": "apps/v1",
    "kind": "Deployment",
    "metadata": map[string]interface{}{
        "name": name,
        "namespace": namespace,
    },
    "note": "Bundle data - limited details available",
}, nil

AFTER (returns error):
return nil, fmt.Errorf("deployment %s/%s not found in bundle", namespace, name)

Apply to both DescribeDeployment() and DescribeService()

Estimated Time: 30 minutes
```

**Task 1c: Fix Container Name Fallbacks**
```
File: internal/datasource/bundle.go
Function: GetContainers()
Line: ~225

BEFORE (returns generic "default"):
containers[i] = "default"

AFTER (returns empty or explicit unknown):
// Option 1: Return empty slice
return []string{}, nil

// Option 2: Return explicit unknown
return []string{"(unknown)"}, nil

Choose Option 1 (empty) - aligns with "Empty is Valid" principle

Estimated Time: 20 minutes
```

#### Phase 2: Regression Testing (Day 2-3, ~6 hours)

**Test Suite 1: Sorting Compliance Tests**
```
File: internal/datasource/bundle_test.go (expand existing)

Add tests:

func TestGetEventsByPod_UsesStdlibSort(t *testing.T) {
    // Purpose: Ensure O(n log n) sorting, prevent bubble sort regression
    // Create 100+ test events
    events := generateTestEvents(100)
    
    // Call GetEventsByPod
    result, err := ds.GetEventsByPod("test-ns", "test-pod")
    
    // Verify: Warnings before Normal
    // Verify: Same type sorted by LastSeen desc
    // Verify: Completes in <10ms (performance check)
}

func TestGetEventsByPod_PerformanceAtScale(t *testing.T) {
    // Purpose: Prevent O(n²) regression
    events := generateTestEvents(500)
    
    start := time.Now()
    result, err := ds.GetEventsByPod("test-ns", "test-pod")
    duration := time.Since(start)
    
    // Assert: duration < 50ms (O(n log n) performance)
    // Assert: correct sort order maintained
}

Estimated Time: 2 hours
```

**Test Suite 2: Truth Only™ Compliance Tests**
```
File: internal/datasource/bundle_test.go

Add tests:

func TestDescribeDeployment_NotFound_ReturnsError(t *testing.T) {
    // Purpose: Ensure no mock data returned
    ds := setupTestDataSource()
    
    result, err := ds.DescribeDeployment("cluster", "ns", "nonexistent")
    
    // Assert: err != nil
    // Assert: result == nil (not mock data)
    // Assert: error message contains "not found"
}

func TestDescribeService_NotFound_ReturnsError(t *testing.T) {
    // Similar to deployment test
}

func TestGetContainers_NoPodInfo_ReturnsEmpty(t *testing.T) {
    // Purpose: Verify empty return (not "default" fallback)
    ds := setupTestDataSource()
    
    containers, err := ds.GetContainers("ns", "unknown-pod")
    
    // Assert: err == nil
    // Assert: len(containers) == 0 (not ["default"])
}

Estimated Time: 2 hours
```

**Test Suite 3: Bundle Parsing Core Tests**
```
File: internal/bundle/kubectl_test.go (NEW FILE)

Purpose: Bring bundle coverage from 0% to basic safety net

func TestParsePods_ValidKubectlOutput(t *testing.T) {
    // Test valid kubectl get pods output
    // Verify pod names, status, restarts extracted correctly
}

func TestParsePods_EmptyOutput_ReturnsEmpty(t *testing.T) {
    // Test with empty kubectl output
    // Verify returns empty slice (not error)
}

func TestParseNodes_ValidOutput(t *testing.T) {
    // Test node parsing
}

func TestParseEvents_MultipleTypes(t *testing.T) {
    // Test event parsing with Warning/Normal mix
}

func TestParseDaemonSets_ReadyStatus(t *testing.T) {
    // Test daemonset ready parsing
}

Estimated Time: 4 hours
```

#### Phase 3: Documentation & Release (Day 4, ~1 hour)

**Task 3a: Update Documentation**
```
Files to update:
- CHANGELOG.md - Add v0.6.9 entry
- PRINCIPLES_AUDIT.md - Mark fixes as completed
- FUTURE_WORK.md - Move completed items to CHANGELOG

Estimated Time: 30 minutes
```

**Task 3b: Validation Testing**
```
Commands to run:
1. go test -cover ./... 
   Target: Overall coverage ≥30%
   
2. go test -bench=. ./internal/datasource
   Target: Sorting benchmarks complete in <50ms
   
3. scripts/test-bundle-stress.sh
   Target: No crashes with large bundles
   
4. Manual testing with example bundle
   Target: Describe errors handled gracefully

Estimated Time: 30 minutes
```

### Success Criteria
- [x] All 3 principle violations fixed
- [x] 5+ new regression tests added
- [x] Test coverage ≥30%
- [x] No performance regressions
- [x] PRINCIPLES_AUDIT score: 98%

### Release Prompt for v0.6.9

```markdown
# v0.6.9 Release Prompt

## Objectives
Fix all critical principle violations and establish regression testing baseline.

## Implementation Checklist

### Code Fixes (2 hours)
1. Replace bubble sort in GetEventsByPod() with sort.Slice
2. Remove mock data fallbacks in DescribeDeployment/DescribeService
3. Fix GetContainers() to return empty (not "default")

### Testing (6 hours)
4. Add TestGetEventsByPod_UsesStdlibSort()
5. Add TestGetEventsByPod_PerformanceAtScale()
6. Add TestDescribe*_NotFound_ReturnsError() tests
7. Add TestGetContainers_NoPodInfo_ReturnsEmpty()
8. Create internal/bundle/kubectl_test.go with 5 core tests

### Validation
9. Run: go test -cover ./...
   Expected: Coverage ≥30%
10. Run: go test -bench=. ./internal/datasource
    Expected: All benchmarks <50ms
11. Manual test: Describe nonexistent resource
    Expected: Clear error (not mock data)

### Documentation
12. Update CHANGELOG.md with v0.6.9 entry
13. Mark PRINCIPLES_AUDIT.md fixes as complete

## Acceptance Criteria
- Zero principle violations (98% compliance)
- Test coverage ≥30%
- All tests passing
- No performance regressions

## Estimated Effort: 8-10 hours
```

---

## 📦 v0.7.0 "Testing Infrastructure & Performance"

**Release Date**: Late April 2026  
**Duration**: 4-5 week sprint  
**Effort**: ~40-50 hours total

### Objectives
1. ✅ Establish CI/CD pipeline with automated testing
2. ✅ Achieve 50% test coverage
3. ✅ Create principle compliance automation
4. ✅ Complete core feature testing
5. ✅ **Fix large bundle performance issues** (NEW - CRITICAL)

### Implementation Tasks

#### Phase 1: CI/CD Foundation (Week 1, ~8 hours)

**Task 1a: GitHub Actions Workflow**
```
File: .github/workflows/test.yml (NEW)

name: Test & Quality Gates

on:
  pull_request:
    branches: [ main ]
  push:
    branches: [ main ]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests with coverage
        run: go test -v -cover -coverprofile=coverage.out ./...
      
      - name: Coverage gate
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$coverage < 50" | bc -l) )); then
            echo "Coverage $coverage% is below 50% threshold"
            exit 1
          fi
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: golangci/golangci-lint-action@v3
        with:
          version: latest

  principle-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Check for bubble sort violations
        run: |
          if grep -r "for.*len.*-1.*for.*len.*-1" internal/; then
            echo "ERROR: Bubble sort pattern detected (violates Principle #8)"
            exit 1
          fi

Estimated Time: 4 hours
```

**Task 1b: Principle Compliance Checker**
```
File: scripts/check-principles.sh (NEW)

#!/bin/bash
# Automated PRINCIPLES.md compliance verification

set -e

echo "🔍 Checking r8s Principle Compliance..."

# Principle #8: O(n log n) Always
echo "Checking for manual sorting (O(n²))..."
if grep -r "for.*len.*-1.*for.*len.*-1" internal/ --exclude="*_test.go"; then
    echo "❌ FAIL: Bubble sort pattern detected"
    echo "   Violates Principle #8: O(n log n) Always"
    exit 1
fi

# Principle #1: Truth Only™  
echo "Checking for mock data fallbacks..."
if grep -r '"note":.*"Bundle data"' internal/datasource/; then
    echo "⚠️  WARN: Mock data fallback detected"
    echo "   May violate Principle #1: Truth Only™"
fi

# Principle #7: Use Your Interfaces
echo "Checking for interface bypassing..."
if grep -r '\.bundle\.' internal/tui/ --exclude="*_test.go"; then
    echo "❌ FAIL: Direct bundle access from TUI"
    echo "   Violates Principle #7: Use Your Interfaces"
    exit 1
fi

echo "✅ All principle checks passed!"

Estimated Time: 2 hours
```

**Task 1c: Pre-commit Hooks**
```
File: .git/hooks/pre-commit (or use pre-commit framework)

#!/bin/bash
# Pre-commit quality checks

# Run tests on changed packages
changed_packages=$(git diff --cached --name-only | grep '\.go$' | xargs -I {} dirname {} | sort -u | xargs)
for pkg in $changed_packages; do
    if [ -f "$pkg"/*.go ]; then
        echo "Testing $pkg..."
        go test "./$pkg" || exit 1
    fi
done

# Run formatter
go fmt ./...

# Check for TODOs in committed code (warn only)
if git diff --cached | grep -i "// TODO"; then
    echo "⚠️  Warning: Committing code with TODO comments"
fi

Estimated Time: 2 hours
```

#### Phase 2: Core Test Coverage (Week 2-3, ~20 hours)

**Task 2a: DataSource Interface Tests**
```
File: internal/datasource/interface_test.go (NEW)

Purpose: Ensure all DataSource implementations comply with interface contract

func TestBundleDataSource_InterfaceCompliance(t *testing.T) {
    // Test all interface methods
    // Verify consistent behavior
}

func TestEmbeddedDataSource_InterfaceCompliance(t *testing.T) {
    // Same tests as Bundle
}

func TestDataSource_EmptyResults_ConsistentHandling(t *testing.T) {
    // Verify "Empty is Valid" principle across implementations
    // GetPods("", "nonexistent") returns []Pod, not error
}

Estimated Time: 4 hours
```

**Task 2b: Bundle Loading Integration Tests**
```
File: internal/bundle/loader_test.go (NEW)

func TestLoadBundle_ExampleBundle_Success(t *testing.T) {
    // Load the example bundle end-to-end
    // Verify all parsers execute
    // Check bundle health calculation
}

func TestLoadBundle_IncompleteBundle_HandlesGracefully(t *testing.T) {
    // Test with missing directories
    // Verify doesn't crash, shows appropriate warnings
}

func TestLoadBundle_InvalidData_ReturnsError(t *testing.T) {
    // Test with corrupted kubectl output
    // Verify returns error (not crash)
}

Estimated Time: 6 hours
```

**Task 2c: Attention Dashboard Tests**
```
File: internal/tui/attention_test.go (expand existing)

func TestComputeAttentionItems_AccurateCountsAtScale(t *testing.T) {
    // Test with 1000+ pods
    // Verify E/W counts accurate
    // Performance: completes in <2 seconds
}

func TestAttentionSorting_PreservesVisualOrder(t *testing.T) {
    // Regression test for Issue #17
    // Verify cursor position matches visual order after sort
}

func TestCriticalSafeCapping_AlwaysShowsCriticals(t *testing.T) {
    // Test Principle #6: Test at 10× Scale
    // Create 200 items, 50 criticals
    // Verify all 50 criticals shown even if beyond cap
}

Estimated Time: 5 hours
```

**Task 2d: Sort Algorithm Consistency**
```
File: internal/tui/attention_signals_test.go (NEW)

func TestAllSortFunctions_UseStdlib(t *testing.T) {
    // Test sortItemsByCount, sortItemsByName, sortItemsBySeverity
    // Verify use sort.Slice (not manual loops)
}

func BenchmarkSortAttentionItems_10000Items(b *testing.B) {
    // Benchmark with 10,000 items
    // Target: <100ms
}

Estimated Time: 3 hours
```

#### Phase 3: Performance Optimization (Week 4, ~10 hours)

**Task 3a: Cache Namespace Health Computation**
```
File: internal/tui/table.go

Problem: Namespace health recomputed on every render - causes 2-5s pauses with 800MB bundles

Solution: Add caching layer with invalidation

type App struct {
    // ... existing fields
    namespaceHealthCache map[string]*NamespaceHealth
    cacheTimestamp      time.Time
}

func (a *App) getNamespaceHealth(namespace string) *NamespaceHealth {
    // Check cache validity
    if time.Since(a.cacheTimestamp) < 5*time.Minute {
        if cached, ok := a.namespaceHealthCache[namespace]; ok {
            return cached
        }
    }
    
    // Compute and cache
    health := a.computeNamespaceHealth(namespace)
    a.namespaceHealthCache[namespace] = health
    a.cacheTimestamp = time.Now()
    return health
}

// Invalidate on refresh
func (a *App) handleRefresh() {
    a.namespaceHealthCache = make(map[string]*NamespaceHealth)
    // ... rest of refresh logic
}

Estimated Time: 3 hours
```

**Task 3b: Async CRD Instance Counting**
```
File: internal/tui/fetch.go

Problem: getCRDInstanceCount() blocks UI during table rendering

Solution: Background goroutine with channel communication

type App struct {
    crdCountCache   map[string]int
    crdCountChan    chan CRDCount
}

type CRDCount struct {
    Name  string
    Count int
}

// Start background worker
func (a *App) startCRDCountWorker() {
    go func() {
        for crdName := range a.crdWorkQueue {
            count := a.fetchCRDInstanceCount(crdName)
            a.crdCountChan <- CRDCount{Name: crdName, Count: count}
        }
    }()
}

// Non-blocking accessor
func (a *App) getCRDInstanceCount(crdName string) string {
    if count, ok := a.crdCountCache[crdName]; ok {
        return fmt.Sprintf("%d", count)
    }
    // Queue for background fetch
    a.crdWorkQueue <- crdName
    return "..."  // Show loading indicator
}

Estimated Time: 4 hours
```

**Task 3c: Profile and Optimize Hot Paths**
```
Use Go profiling tools to identify bottlenecks:

1. CPU Profiling:
   go test -cpuprofile=cpu.prof -bench=. ./internal/tui
   go tool pprof cpu.prof

2. Memory Profiling:
   go test -memprofile=mem.prof -bench=. ./internal/tui
   go tool pprof mem.prof

3. Identify hot paths:
   - AttentionItems computation
   - Log scanning
   - Table rendering

4. Add lazy loading for attention items:
   - Compute once on dashboard load
   - Cache results until refresh
   - Don't recompute on view switches

Estimated Time: 3 hours
```

#### Phase 4: Documentation & Tooling (Week 5, ~5 hours)

**Task 4a: Testing Best Practices Guide**
```
File: docs/development/TESTING_GUIDE.md (NEW)

Content:
- How to write tests for r8s
- Test fixtures and helpers
- Coverage expectations by module
- Running tests locally
- CI/CD integration

Estimated Time: 2 hours
```

**Task 3b: Update Contributing Guide**
```
File: CONTRIBUTING.md (update)

Add sections:
- Testing requirements for PRs
- Coverage expectations
- Principle compliance checks
- How to run local tests

Estimated Time: 1 hour
```

**Task 3c: Coverage Reporting**
```
Set up Codecov or similar:
- Badge in README.md
- PR comments with coverage delta
- Fail PR if coverage drops

Estimated Time: 2 hours
```

### Success Criteria
- [x] CI/CD pipeline operational
- [x] Test coverage ≥50%
- [x] Principle compliance automated
- [x] All core modules tested
- [x] Pre-commit hooks working

### Release Prompt for v0.7.0

```markdown
# v0.7.0 Release Prompt

## Objectives
Establish comprehensive testing infrastructure and CI/CD automation.

## Implementation Checklist

### CI/CD Setup (8 hours)
1. Create .github/workflows/test.yml with coverage gates
2. Create scripts/check-principles.sh for automated compliance
3. Set up pre-commit hooks
4. Configure Codecov integration

### Test Coverage Expansion (20 hours)
5. Create internal/datasource/interface_test.go (contract tests)
6. Create internal/bundle/loader_test.go (E2E tests)
7. Expand internal/tui/attention_test.go (scale tests)
8. Create internal/tui/attention_signals_test.go (sort tests)
9. Add performance benchmarks

### Documentation (5 hours)
10. Create docs/development/TESTING_GUIDE.md
11. Update CONTRIBUTING.md with testing requirements
12. Add coverage badge to README.md

### Validation
13. Run full test suite: go test -v -cover ./...
    Expected: ≥50% coverage
14. Verify CI pipeline passes on sample PR
15. Run scripts/check-principles.sh
    Expected: All checks pass

## Acceptance Criteria
- CI/CD pipeline prevents principle violations
- Coverage ≥50% enforced automatically
- Pre-commit hooks catch issues early
- Comprehensive test suite in place

## Estimated Effort: 30-40 hours
```

---

## 📦 v0.7.5 "Performance & Edge Cases"

**Release Date**: Mid-June 2026  
**Duration**: 2-3 week sprint  
**Effort**: ~20-25 hours

### Objectives
1. ✅ Achieve 70% test coverage
2. ✅ Performance benchmarks for all core functions
3. ✅ Edge case handling comprehensive
4. ✅ Stress testing at 10× scale

### Key Tasks
- Performance benchmarks (internal/bundle/benchmark_test.go)
- Stress test scripts (scripts/stress-test-*.sh)
- Table rendering edge cases
- CLI command tests
- OOM analysis completion

### Release Prompt for v0.7.5

```markdown
# v0.7.5 Release Prompt

## Focus: Performance optimization and edge case coverage

### Performance Benchmarks
- BenchmarkParsePods_1000Pods <500ms
- BenchmarkComputeAttentionItems <2s with 1000 pods
- BenchmarkLogScanning_1MLines <5s

### Stress Tests
- Dashboard with 1000 pods
- 500+ events per pod
- 10M log lines
- All must remain responsive

### Edge Cases
- Zero-state handling
- Incomplete bundles
- Corrupted data
- Terminal resize during operation

Target: 70% coverage, all stress tests passing
```

---

## 📦 v0.8.0 "Production Hardening"

**Release Date**: Late August 2026  
**Duration**: 4-5 week sprint  
**Effort**: ~40-50 hours

### Objectives
1. ✅ Achieve 80%+ test coverage
2. ✅ Production-ready performance
3. ✅ Comprehensive documentation
4. ✅ Long-term maintenance plan

### Key Features
- 80% test coverage across all modules
- <2s dashboard load with 1000 pods
- Complete error handling
- Performance profiling
- Security audit
- Production deployment guide

### Success Criteria
- [x] 80%+ test coverage
- [x] All benchmarks meet targets
- [x] Zero known critical bugs
- [x] Complete documentation
- [x] Ready for 1.0 planning

### Release Prompt for v0.8.0

```markdown
# v0.8.0 Release Prompt

## Objective: Production-hardening milestone

### Quality Targets
- 80%+ test coverage (comprehensive)
- Zero critical/high severity bugs
- All performance benchmarks green
- Security audit complete

### Performance Targets
- Dashboard load: <2s (1000 pods)
- Log scanning: <5s (10M lines)
- Memory: <500MB (large bundles)
- No memory leaks

### Documentation
- Complete API documentation
- Deployment guide
- Troubleshooting playbook
- Upgrade paths documented

### Preparation for v1.0
- Feature freeze for stability
- Long-term support planning
- Backward compatibility guarantees

This release establishes r8s as production-ready.
```

---

## 📊 Release Timeline

```
2026 Timeline:

Jan ████ v0.6.8 (current)
Feb ████████████ v0.6.9 (compliance) + v0.6.10 (refactor + K3s)
Mar ████ 
Apr ████████ v0.7.0 (CI/CD + 50% coverage)
May ████
Jun ████████ v0.7.5 (performance + 70% coverage)
Jul ████
Aug ████████████ v0.8.0 (production hardening + 80% coverage)
Sep ████ v1.0 planning

Legend:
████ = Active development
```

---

## 🎯 Metrics Dashboard

### Test Coverage Progression
| Release | Coverage | Delta | Target |
|---------|----------|-------|--------|
| v0.6.8 (current) | ~13% | - | Baseline |
| v0.6.9 | 30% | +17% | ✅ 30% |
| v0.6.10 | 35% | +5% | 🟢 Maintain + K3s tests |
| v0.7.0 | 50% | +15% | ✅ 50% |
| v0.7.5 | 70% | +20% | ✅ 70% |
| v0.8.0 | 80% | +10% | ✅ 80% |

### Principle Compliance
| Release | Score | Status |
|---------|-------|--------|
| v0.6.8 | 76% (91/120) | 🟡 Partial |
| v0.6.9 | 98% (117/120) | 🟢 Excellent |
| v0.7.0+ | 98%+ | 🟢 Maintained |

### Performance Targets
| Metric | v0.6.9 | v0.7.5 | v0.8.0 |
|--------|--------|--------|--------|
| Dashboard (1000 pods) | N/A | <5s | <2s |
| Event sort (500 events) | <50ms | <30ms | <20ms |
| Log scan (1M lines) | N/A | <10s | <5s |
| Memory (large bundle) | N/A | <1GB | <500MB |

---

## 🔗 Dependencies & Ordering

### Critical Path
```
v0.6.9 (Principle Fixes)
    ↓
v0.6.10 (Architecture Refactor + K3s)
    ↓
v0.7.0 (CI/CD Infrastructure)
    ↓
v0.7.5 (Performance Testing)
    ↓
v0.8.0 (Production Hardening)
    ↓
v1.0 (Stable Release)
```

### Parallel Workstreams
- **Testing**: Can progress independently per module
- **Documentation**: Can be written alongside features
- **Performance**: Requires baseline from v0.6.9 first

---

## 📚 References

- **PRINCIPLES_AUDIT.md** - Detailed compliance analysis
- **FUTURE_WORK.md** - Complete feature backlog
- **PRINCIPLES.md** - Development principles
- **CHANGELOG.md** - Release history

---

**Roadmap Maintained By**: Development Team  
**Last Updated**: 2026-02-13  
**Next Review**: After v0.6.10 release
