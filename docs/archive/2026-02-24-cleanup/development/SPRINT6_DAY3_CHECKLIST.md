# Multi-Distro Refactor - Detailed Checklist

**Status**: ⚠️ DEFERRED to v0.7.1 (post-Sprint 6)  
**Total Hardcoded Paths**: 18+ occurrences across 9 files  
**Estimated Effort**: 8-10 hours

---

## ⚠️ DEFERRAL NOTICE

This checklist was originally Sprint 6 Day 3, but has been **DEFERRED** per strategic brief approval (sequential timeline).

**Reason:** Sprint 6 is CI Stability ONLY. Path refactoring without CI linting = integration hell.

**New Plan:**
- Sprint 6: Fix CI (#44, #45) - see `SPRINT6_CI_STABILITY.md`
- v0.7.1: K3s support with REDUCED scope (5 files, not 18)

This document retained for v0.7.1 planning.

---

## Pre-Migration Commands

Run these before starting to get current state:

```bash
# Count all rke2 references
grep -rn "rke2" /home/node/.openclaw/workspace/r8s --include="*.go" | wc -l

# Find all hardcoded path constructions
grep -rn '"rke2"' /home/node/.openclaw/workspace/r8s --include="*.go"

# Find service name references
grep -rn "rke2-server\|rke2-agent" /home/node/.openclaw/workspace/r8s --include="*.go"
```

---

## Phase 1: Foundation Changes

### Task 1.1: Update types.go

**File**: `internal/bundle/types.go`

- [ ] Add `FormatK3s` constant
- [ ] Add `FormatRKE1` constant (nice-to-have)
- [ ] Add `DistroDir string` field to Bundle struct
- [ ] Add `DistroType BundleFormat` field to Bundle struct
- [ ] Add `ServiceNameMapping()` method
- [ ] Add `ServiceNames` struct type

**Verification**:
```bash
grep -n "FormatK3s\|DistroDir\|ServiceNameMapping" internal/bundle/types.go
```

---

## Phase 2: Path Helper Methods

### Task 2.1: Add Helper Methods

**File**: `internal/bundle/types.go` or create `internal/bundle/paths.go`

Add these methods to Bundle struct:

- [ ] `func (b *Bundle) KubectlPath(resource string) string`
- [ ] `func (b *Bundle) PodLogsPath() string`
- [ ] `func (b *Bundle) PodLogsFile(namespace, pod, container string) string`
- [ ] `func (b *Bundle) PodManifestsPath() string`
- [ ] `func (b *Bundle) PodManifestFile(namespace, pod string) string`
- [ ] `func (b *Bundle) PodDescribePath() string`
- [ ] `func (b *Bundle) PodDescribeFile(namespace, pod string) string`
- [ ] `func (b *Bundle) SystemLogsPath() string`
- [ ] `func (b *Bundle) ServerLogPath() string`
- [ ] `func (b *Bundle) AgentLogPath() string`

**Verification**:
```bash
grep -n "func (b \*Bundle)" internal/bundle/types.go | head -20
```

---

## Phase 3: Refactor Existing Files

### File: internal/bundle/manifest.go

**Lines to find and update** (use `grep -n` to get exact line numbers):

- [ ] Line __: `DetectFormat()` - Update to return distro directory
  - **Current**: Returns only format
  - **Change**: Return `(BundleFormat, string, error)` with distro dir

- [ ] Line __: `LoadBundle()` - Set DistroDir and DistroType fields
  - **Current**: Only sets Format field
  - **Change**: Add `DistroDir: distroDir` and `DistroType: format`

- [ ] Line __: Path construction with `"rke2"` literal (inventoryPods)
  - **Current**: `filepath.Join(b.RootPath, "rke2", "kubectl", "pods")`
  - **Change**: `b.KubectlPath("pods")`

- [ ] Line __: Path construction with `"rke2"` literal (inventoryNodes)
  - **Current**: `filepath.Join(b.RootPath, "rke2", "kubectl", "nodes")`
  - **Change**: `b.KubectlPath("nodes")`

- [ ] Line __: Path construction with `"rke2"` literal (inventoryEvents)
  - **Current**: `filepath.Join(b.RootPath, "rke2", "kubectl", "events")`
  - **Change**: `b.KubectlPath("events")`

- [ ] Line __: Path construction with `"rke2"` literal (inventoryDaemonSets)
  - **Current**: `filepath.Join(b.RootPath, "rke2", "kubectl", "daemonsets")`
  - **Change**: `b.KubectlPath("daemonsets")`

**Verification**:
```bash
grep -n "rke2" internal/bundle/manifest.go
grep -n 'KubectlPath\|PodLogsPath\|PodManifest' internal/bundle/manifest.go
```

---

### File: internal/bundle/validate.go

- [ ] Line __: ValidateBundle() rke2 directory check
  - **Current**: `rke2Dir := filepath.Join(b.RootPath, "rke2")`
  - **Change**: `distroDir := filepath.Join(b.RootPath, b.DistroDir)`

**Verification**:
```bash
grep -n "rke2" internal/bundle/validate.go
```

---

### File: internal/bundle/journald.go

- [ ] Line __: Service name constants
  - **Current**: `const RKE2ServerService = "rke2-server"`
  - **Change**: Remove constants, use `b.ServiceNameMapping()`

- [ ] Line __: Service name usage in parsing
  - **Current**: Hardcoded `"rke2-server"` in log parsing
  - **Change**: `svc := b.ServiceNameMapping(); svc.ServerService`

- [ ] Line __: Service name usage in log opening
  - **Current**: Hardcoded path to `systemlogs/rke2-server.log`
  - **Change**: `b.ServerLogPath()`

**Verification**:
```bash
grep -n "rke2-server\|rke2-agent" internal/bundle/journald.go
```

---

### File: internal/bundle/completeness.go

- [ ] Line __: Required paths table
  - **Current**: Static slice with `"rke2/kubectl/pods"` etc
  - **Change**: Dynamic generation using `b.DistroDir`

- [ ] Line __: Path validation logic
  - **Current**: Hardcoded expectations for rke2 structure
  - **Change**: Use `b.GetRequiredPaths()` helper

**Verification**:
```bash
grep -n "rke2" internal/bundle/completeness.go
```

---

### File: internal/bundle/resources.go

- [ ] Line __: Pod manifest path
  - **Current**: `filepath.Join(b.RootPath, "rke2", "pod-manifests", filename)`
  - **Change**: `b.PodManifestFile(namespace, podName)`

- [ ] Line __: Pod describe path
  - **Current**: `filepath.Join(b.RootPath, "rke2", "poddescribe", filename)`
  - **Change**: `b.PodDescribeFile(namespace, podName)`

- [ ] Line __: Kubectl path for resources
  - **Current**: `filepath.Join(b.RootPath, "rke2", "kubectl", resource)`
  - **Change**: `b.KubectlPath(resource)`

**Verification**:
```bash
grep -n "rke2" internal/bundle/resources.go
```

---

### File: internal/bundle/oom.go

- [ ] Line __: Check for rke2 path references
  - Search for any hardcoded paths
  - Replace with helper methods

**Verification**:
```bash
grep -n "rke2" internal/bundle/oom.go || echo "No hardcoded paths found"
```

---

## Phase 4: Test Updates

### File: internal/bundle/kubectl_test.go

- [ ] Add `TestDetectFormat_K3sBundle()`
- [ ] Add `TestDetectFormat_RKE2Bundle()`
- [ ] Update existing tests to work with new Bundle struct fields

### File: internal/bundle/manifest_test.go

- [ ] Add K3s bundle fixtures
- [ ] Add RKE2 nested bundle test

### Create Test Fixtures

- [ ] Create `testdata/bundles/minimal-k3s/k3s/kubectl/pods`
- [ ] Create `testdata/bundles/minimal-k3s/k3s/kubectl/events`
- [ ] Create `testdata/bundles/minimal-k3s/k3s/kubectl/nodes`
- [ ] Create `testdata/bundles/minimal-k3s/systemlogs/k3s.log`

---

## Verification Commands

Run these after all changes:

```bash
# 1. Verify build passes
cd /home/node/.openclaw/workspace/r8s && go build .

# 2. Run tests
go test ./internal/bundle/... -v

# 3. Check no hardcoded rke2 paths remain (except in test data)
grep -rn "rke2" /home/node/.openclaw/workspace/r8s --include="*.go" | grep -v "_test.go" | grep -v "testdata"

# 4. Verify helper methods are used
grep -rn "KubectlPath\|PodLogsPath\|PodManifest" /home/node/.openclaw/workspace/r8s --include="*.go"

# 5. Count method usages vs old pattern
echo "Helper method usages:"
grep -rn "b\.KubectlPath\|b\.PodLogsPath\|b\.PodManifest" /home/node/.openclaw/workspace/r8s --include="*.go" | wc -l
echo "Old filepath.Join patterns:"
grep -rn 'filepath.Join.*rke2' /home/node/.openclaw/workspace/r8s --include="*.go" | wc -l
```

---

## Success Criteria Checklist

- [ ] `types.go` has FormatK3s, DistroDir, DistroType fields
- [ ] `types.go` has ServiceNameMapping() method
- [ ] All path helper methods implemented (KubectlPath, PodLogsPath, etc.)
- [ ] `manifest.go` uses helper methods, no hardcoded "rke2"
- [ ] `validate.go` uses b.DistroDir
- [ ] `journald.go` uses ServiceNameMapping() for service names
- [ ] `completeness.go` generates paths dynamically
- [ ] `resources.go` uses helper methods
- [ ] `oom.go` checked and updated if needed
- [ ] K3s test fixtures created
- [ ] All tests pass (`go test ./...`)
- [ ] Build succeeds (`go build .`)
- [ ] Manual test with RKE2 bundle (regression)
- [ ] Manual test with K3s bundle (new functionality)

---

## Post-Implementation Notes

**To be filled in during implementation**:

- Actual line numbers found: ___________
- Total files modified: ___________
- Test coverage change: ___________
- Performance impact: ___________
- Issues encountered: ___________

---

**Last Updated**: 2026-02-13  
**Assigned To**: ___________  
**Completion Date**: ___________
