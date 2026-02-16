# Multi-Distribution Bundle Support Implementation

**Status**: ⚠️ DEFERRED to v0.7.1 (post-Sprint 6)  
**Priority**: K3s required, RKE1 nice-to-have  
**Estimated Effort**: 8-10 hours  
**Prerequisites**: Sprint 6 CI Stability MUST be complete first

---

## ⚠️ DEFERRAL NOTICE

This work was originally Sprint 6 Day 3, but has been **DEFERRED** per strategic brief approval.

**Why:** Sprint 6 scope is CI Stability ONLY (#44, #45). K3s work requires:
1. CI pipeline passing (lint + coverage)
2. Clean foundation for path refactoring
3. Risk mitigation (no K3s bundles to test with yet)

**New Timeline:** v0.7.1 (after Sprint 6 completes)

**Reduced Scope for v0.7.1:**
- K3s detection ONLY (not full 18-file refactor)
- 5 core files (not all 9+)
- NO RKE1 support (moved to v0.7.4)

See: `SPRINT6_CI_STABILITY.md` for current Sprint 6 work

---

## Overview

This document provides a complete implementation playbook for adding multi-distribution support to r8s, enabling it to parse both RKE2 and K3s support bundles (with RKE1 as a future extension). The work is designed as a standalone Day 3 task that can be executed by any team member in a future sprint.

### Why Now vs v0.6.10

Initially planned for v0.6.10, this refactor is being moved to **Sprint 6 Day 3** because:

1. **18 hardcoded paths exist now**, but new v0.7.0 parsers (journald, completeness, resources) are adding MORE hardcoded rke2 paths
2. **Technical debt is accumulating** - every new parser adds to the migration burden
3. **Context is fresh** - you're already touching `internal/bundle/*.go` files for v0.7.0 work
4. **Saves ~20 hours** - doing it now alongside parser work vs re-learning codebase in v0.6.10
5. **Blocks future distro support** - no RKE1, custom distros until this is done

### Bundle Structure Comparison

```
# RKE2 Bundle
rke2/kubectl/pods
rke2/kubectl/events
rke2/kubectl/nodes
rke2/kubectl/daemonsets
rke2/kubectl/deployments
rke2/podlogs/<namespace>_<pod>_<container>.log
rke2/pod-manifests/<namespace>_<pod>.yaml
rke2/poddescribe/<namespace>_<pod>.txt
systemlogs/rke2-server.log
systemlogs/rke2-agent.log

# K3s Bundle (identical structure, different prefix)
k3s/kubectl/pods
k3s/kubectl/events
k3s/kubectl/nodes
k3s/kubectl/daemonsets
k3s/kubectl/deployments
k3s/podlogs/<namespace>_<pod>_<container>.log
k3s/pod-manifests/<namespace>_<pod>.yaml
k3s/poddescribe/<namespace>_<pod>.txt
systemlogs/k3s.log
systemlogs/k3s-agent.log

# RKE1 Bundle (nice-to-have, check if different)
# Structure TBD - verify with rancher2_logs_collector.sh
```

---

## Phase 1: Foundation Changes (2 hours)

### 1.1 Add Format Constants and Bundle Fields

**File**: `internal/bundle/types.go`

```go
// Add to BundleFormat constants
const (
    FormatRKE2      BundleFormat = "rke2-support-bundle"
    FormatK3s       BundleFormat = "k3s-support-bundle"
    FormatRKE1      BundleFormat = "rke1-support-bundle" // Nice-to-have
    FormatUnknown   BundleFormat = "unknown"
)

// Add to Bundle struct
type Bundle struct {
    // ... existing fields ...
    
    // DistroDir is the top-level directory name ("rke2", "k3s", or "rke1")
    DistroDir string `json:"distro_dir"`
    
    // DistroType identifies the distribution type for service name mapping
    DistroType BundleFormat `json:"distro_type"`
}

// ServiceNameMapping returns the appropriate service names for the distro
func (b *Bundle) ServiceNameMapping() ServiceNames {
    switch b.DistroType {
    case FormatK3s:
        return ServiceNames{
            ServerService: "k3s",
            AgentService:  "k3s-agent",
        }
    case FormatRKE1:
        return ServiceNames{
            ServerService: "rancher-k3s-server", // Verify for RKE1
            AgentService:  "rancher-k3s-agent",
        }
    default: // FormatRKE2
        return ServiceNames{
            ServerService: "rke2-server",
            AgentService:  "rke2-agent",
        }
    }
}

type ServiceNames struct {
    ServerService string
    AgentService  string
}
```

### 1.2 Update DetectFormat()

**File**: `internal/bundle/manifest.go`

```go
func DetectFormat(bundleRoot string) (BundleFormat, string, error) {
    // Check for RKE2
    rke2Dir := filepath.Join(bundleRoot, "rke2")
    if info, err := os.Stat(rke2Dir); err == nil && info.IsDir() {
        // Verify it has expected structure
        if hasBundleStructure(rke2Dir) {
            return FormatRKE2, "rke2", nil
        }
    }
    
    // Check for K3s
    k3sDir := filepath.Join(bundleRoot, "k3s")
    if info, err := os.Stat(k3sDir); err == nil && info.IsDir() {
        if hasBundleStructure(k3sDir) {
            return FormatK3s, "k3s", nil
        }
    }
    
    // Check for RKE1 (nice-to-have)
    rke1Dir := filepath.Join(bundleRoot, "rke1")
    if info, err := os.Stat(rke1Dir); err == nil && info.IsDir() {
        if hasBundleStructure(rke1Dir) {
            return FormatRKE1, "rke1", nil
        }
    }
    
    // Check inside wrapper directories
    entries, err := os.ReadDir(bundleRoot)
    if err != nil {
        return FormatUnknown, "", fmt.Errorf("failed to read bundle root: %w", err)
    }
    
    for _, entry := range entries {
        if entry.IsDir() {
            nestedPath := filepath.Join(bundleRoot, entry.Name())
            format, dir, err := DetectFormat(nestedPath)
            if err == nil && format != FormatUnknown {
                return format, filepath.Join(entry.Name(), dir), nil
            }
        }
    }
    
    return FormatUnknown, "", fmt.Errorf("no valid bundle structure found")
}

func hasBundleStructure(distroDir string) bool {
    // Check for minimum viable structure: kubectl/pods
    kubectlDir := filepath.Join(distroDir, "kubectl")
    podsFile := filepath.Join(kubectlDir, "pods")
    
    info, err := os.Stat(podsFile)
    return err == nil && !info.IsDir()
}
```

### 1.3 Update LoadBundle()

**File**: `internal/bundle/manifest.go`

```go
func LoadBundle(bundleRoot string) (*Bundle, error) {
    // ... existing validation ...
    
    format, distroDir, err := DetectFormat(absoluteRoot)
    if err != nil {
        return nil, fmt.Errorf("bundle detection failed: %w", err)
    }
    
    bundle := &Bundle{
        RootPath:   absoluteRoot,
        Format:     format,
        DistroDir:  distroDir,     // "rke2" or "k3s" (or nested like "host1/rke2")
        DistroType: format,        // FormatRKE2, FormatK3s, etc.
        // ... other fields ...
    }
    
    // ... rest of loading logic ...
    
    return bundle, nil
}
```

---

## Phase 2: Path Abstraction Methods (2 hours)

### 2.1 Add Helper Methods to Bundle

**File**: `internal/bundle/types.go` (or new file `internal/bundle/paths.go`)

```go
// KubectlPath returns the path to a kubectl output file
func (b *Bundle) KubectlPath(resource string) string {
    return filepath.Join(b.RootPath, b.DistroDir, "kubectl", resource)
}

// PodLogsPath returns the directory containing pod logs
func (b *Bundle) PodLogsPath() string {
    return filepath.Join(b.RootPath, b.DistroDir, "podlogs")
}

// PodLogsFile returns the path to a specific pod log file
func (b *Bundle) PodLogsFile(namespace, pod, container string) string {
    filename := fmt.Sprintf("%s_%s_%s.log", namespace, pod, container)
    return filepath.Join(b.PodLogsPath(), filename)
}

// PodManifestsPath returns the directory containing pod manifests
func (b *Bundle) PodManifestsPath() string {
    return filepath.Join(b.RootPath, b.DistroDir, "pod-manifests")
}

// PodManifestFile returns the path to a specific pod manifest
func (b *Bundle) PodManifestFile(namespace, pod string) string {
    filename := fmt.Sprintf("%s_%s.yaml", namespace, pod)
    return filepath.Join(b.PodManifestsPath(), filename)
}

// PodDescribePath returns the directory containing pod describe output
func (b *Bundle) PodDescribePath() string {
    return filepath.Join(b.RootPath, b.DistroDir, "poddescribe")
}

// PodDescribeFile returns the path to a specific pod describe file
func (b *Bundle) PodDescribeFile(namespace, pod string) string {
    filename := fmt.Sprintf("%s_%s.txt", namespace, pod)
    return filepath.Join(b.PodDescribePath(), filename)
}

// SystemLogsPath returns the system logs directory
func (b *Bundle) SystemLogsPath() string {
    return filepath.Join(b.RootPath, "systemlogs")
}

// ServerLogPath returns the server service log path
func (b *Bundle) ServerLogPath() string {
    svc := b.ServiceNameMapping()
    return filepath.Join(b.SystemLogsPath(), svc.ServerService+".log")
}

// AgentLogPath returns the agent service log path (if exists)
func (b *Bundle) AgentLogPath() string {
    svc := b.ServiceNameMapping()
    return filepath.Join(b.SystemLogsPath(), svc.AgentService+".log")
}
```

---

## Phase 3: Refactor Existing Code (4 hours)

### 3.1 Update manifest.go

**Current hardcoded paths in manifest.go** (~7 occurrences):

```go
// BEFORE (example from inventoryPods):
podsPath := filepath.Join(b.RootPath, "rke2", "kubectl", "pods")

// AFTER:
podsPath := b.KubectlPath("pods")
```

**Migration checklist for manifest.go**:
- [ ] `DetectFormat()` - Update to return detected distro (see Phase 1.2)
- [ ] `LoadBundle()` - Set `DistroDir` and `DistroType` fields
- [ ] `inventoryPods()` - Replace `filepath.Join(b.RootPath, "rke2", "kubectl", "pods")`
- [ ] `inventoryNodes()` - Replace kubectl path construction
- [ ] `inventoryEvents()` - Replace kubectl/events path
- [ ] `inventoryDaemonSets()` - Replace kubectl/daemonsets path
- [ ] Any other path construction using `"rke2"` literal

### 3.2 Update validate.go

**Current hardcoded path** (~1 occurrence):

```go
// BEFORE:
rke2Dir := filepath.Join(b.RootPath, "rke2")
if _, err := os.Stat(rke2Dir); os.IsNotExist(err) {
    issues = append(issues, "Missing rke2/ directory")
}

// AFTER:
distroDir := filepath.Join(b.RootPath, b.DistroDir)
if _, err := os.Stat(distroDir); os.IsNotExist(err) {
    issues = append(issues, fmt.Sprintf("Missing %s/ directory", b.DistroDir))
}
```

### 3.3 Update journald.go

**Current hardcoded service names** (~4 occurrences):

```go
// BEFORE:
const (
    RKE2ServerService = "rke2-server"
    RKE2AgentService  = "rke2-agent"
)

// AFTER - use dynamic service names:
func (b *Bundle) GetServiceNames() (server, agent string) {
    svc := b.ServiceNameMapping()
    return svc.ServerService, svc.AgentService
}

// In parser functions:
serverSvc, agentSvc := b.GetServiceNames()
// Parse logs looking for serverSvc and agentSvc
```

**Files to check for service name hardcoding**:
- `internal/bundle/journald.go` - All rke2-server/rke2-agent references
- Any log parsing that looks for specific service names
- System log file opening logic

### 3.4 Update completeness.go

**Current hardcoded paths** (~10+ occurrences in requirements table):

```go
// BEFORE:
var requiredPaths = []string{
    "rke2/kubectl/pods",
    "rke2/kubectl/events",
    "rke2/kubectl/nodes",
    // ...
}

// AFTER:
func (b *Bundle) GetRequiredPaths() []string {
    base := b.DistroDir
    return []string{
        filepath.Join(base, "kubectl", "pods"),
        filepath.Join(base, "kubectl", "events"),
        filepath.Join(base, "kubectl", "nodes"),
        filepath.Join(base, "kubectl", "daemonsets"),
        filepath.Join(base, "kubectl", "deployments"),
        filepath.Join(base, "kubectl", "replicasets"),
        filepath.Join(base, "kubectl", "services"),
        filepath.Join(base, "podlogs"),
        filepath.Join(base, "pod-manifests"),
        "systemlogs",
    }
}
```

### 3.5 Update resources.go

**Current hardcoded paths** (~3 occurrences):

```go
// BEFORE:
manifestPath := filepath.Join(b.RootPath, "rke2", "pod-manifests", filename)
describePath := filepath.Join(b.RootPath, "rke2", "poddescribe", filename)

// AFTER:
manifestPath := b.PodManifestFile(namespace, podName)
describePath := b.PodDescribeFile(namespace, podName)
```

### 3.6 Update oom.go

Check for any `rke2/` path references and replace with helper methods.

### 3.7 Update Test Fixtures

**Files to update**:
- `internal/bundle/kubectl_test.go` - Test data paths
- `internal/datasource/bundle_test.go` - Bundle test fixtures

Create K3s test fixtures:
```
testdata/bundles/minimal-k3s/
  k3s/kubectl/pods
  k3s/kubectl/events
  k3s/kubectl/nodes
  systemlogs/k3s.log

testdata/bundles/full-k3s/
  [complete K3s bundle structure]
```

---

## Phase 4: Testing & Validation (2 hours)

### 4.1 Unit Tests for Detection

**File**: `internal/bundle/manifest_test.go` (add new tests)

```go
func TestDetectFormat_RKE2Bundle(t *testing.T) {
    tmpDir := t.TempDir()
    createRKE2Structure(tmpDir)
    
    format, distroDir, err := DetectFormat(tmpDir)
    
    assert.NoError(t, err)
    assert.Equal(t, FormatRKE2, format)
    assert.Equal(t, "rke2", distroDir)
}

func TestDetectFormat_K3sBundle(t *testing.T) {
    tmpDir := t.TempDir()
    createK3sStructure(tmpDir)
    
    format, distroDir, err := DetectFormat(tmpDir)
    
    assert.NoError(t, err)
    assert.Equal(t, FormatK3s, format)
    assert.Equal(t, "k3s", distroDir)
}

func TestDetectFormat_NestedBundle(t *testing.T) {
    // Test wrapper directory: supportbundle_123/host1/rke2/
    tmpDir := t.TempDir()
    wrapperDir := filepath.Join(tmpDir, "wrapper")
    createRKE2Structure(wrapperDir)
    
    format, distroDir, err := DetectFormat(tmpDir)
    
    assert.NoError(t, err)
    assert.Equal(t, FormatRKE2, format)
    assert.Equal(t, filepath.Join("wrapper", "rke2"), distroDir)
}
```

### 4.2 Integration Tests

**File**: `internal/bundle/bundle_integration_test.go` (new file)

```go
func TestLoadBundle_RKE2andK3sParity(t *testing.T) {
    rke2Bundle := loadTestBundle(t, "testdata/bundles/minimal-rke2")
    k3sBundle := loadTestBundle(t, "testdata/bundles/minimal-k3s")
    
    // Both should load successfully
    assert.NotNil(t, rke2Bundle)
    assert.NotNil(t, k3sBundle)
    
    // Verify path helpers work correctly
    assert.Contains(t, rke2Bundle.KubectlPath("pods"), "rke2")
    assert.Contains(t, k3sBundle.KubectlPath("pods"), "k3s")
    
    // Both should parse same pod data structure
    rke2Pods := rke2Bundle.GetPods()
    k3sPods := k3sBundle.GetPods()
    
    // Data structure should be identical (only paths differ)
    assert.Equal(t, len(rke2Pods), len(k3sPods))
}
```

### 4.3 Regression Tests

Verify existing RKE2 bundles still work:
- [ ] Load existing test bundles
- [ ] Run full test suite: `go test ./internal/bundle/...`
- [ ] Verify 100% pass rate
- [ ] Check no hardcoded "rke2" strings remain (except tests for RKE2 specifically)

### 4.4 Manual Validation

```bash
# Build r8s
cd /home/node/.openclaw/workspace/r8s && make build

# Test with RKE2 bundle (regression)
./r8s --bundle /path/to/rke2/bundle
# → Should work exactly as before

# Test with K3s bundle (new)
./r8s --bundle /path/to/k3s/bundle  
# → Should now work (previously would fail)
```

---

## File Impact Summary

| File | Changes | Lines Modified |
|------|---------|----------------|
| `internal/bundle/types.go` | Add fields, ServiceNameMapping, path helpers | +80-100 |
| `internal/bundle/manifest.go` | Update DetectFormat, LoadBundle, all path construction | ~50 |
| `internal/bundle/validate.go` | Update ValidateBundle to use dynamic paths | ~10 |
| `internal/bundle/journald.go` | Dynamic service names | ~20 |
| `internal/bundle/completeness.go` | Dynamic path requirements | ~15 |
| `internal/bundle/resources.go` | Use helper methods | ~10 |
| `internal/bundle/oom.go` | Check for hardcoded paths | ~5 |
| `internal/bundle/*_test.go` | Add K3s tests, fixtures | +100 |
| `testdata/bundles/*` | Create K3s test fixtures | New files |

**Total estimated changes**: ~300 lines across 9-10 files

---

## RKE1 Extension Notes (Nice-to-Have)

If implementing RKE1 support:

1. **Verify bundle structure** - Check `rancher2_logs_collector.sh` for RKE1 differences
2. **Service names** - RKE1 uses different service names (docker, kubelet, etc.)
3. **Log locations** - May differ from RKE2/K3s structure
4. **Update DetectFormat()** - Add RKE1 directory check
5. **Add FormatRKE1** - Already in types.go from Phase 1.1

**Estimated additional effort**: 2-3 hours

---

## Success Criteria

- [ ] K3s bundles load and parse correctly
- [ ] RKE2 bundles continue to work (regression tested)
- [ ] Zero hardcoded distro paths remain (all use helper methods)
- [ ] All tests pass (`go test ./...`)
- [ ] New K3s test fixtures created and passing
- [ ] Service name abstraction works for journald parsing
- [ ] Documentation updated (BUNDLE_DEPENDENCY_ANALYSIS.md)

---

## Common Pitfalls

1. **Trailing slashes in DistroDir** - Ensure consistent handling (no trailing slash)
2. **Nested wrapper directories** - Handle `wrapper/host1/rke2/` structure correctly
3. **Service name case sensitivity** - Verify exact service names in collector script
4. **Test fixture drift** - Keep K3s fixtures updated if RKE2 structure changes
5. **Partial matches** - Don't match `k3s` inside a directory named `my-k3s-cluster`

---

## References

- Rancher2 logs collector: https://github.com/rancherlabs/support-tools/blob/master/collection/rancher/v2.x/logs-collector/rancher2_logs_collector.sh
- K3s bundle creation: `k3s-setup()` function in collector script
- RKE2 bundle creation: `rke2-setup()` function in collector script
- BUNDLE_DEPENDENCY_ANALYSIS.md - Current bundle structure docs

---

**Last Updated**: 2026-02-13  
**Ready for**: Sprint 6, Day 3 execution
