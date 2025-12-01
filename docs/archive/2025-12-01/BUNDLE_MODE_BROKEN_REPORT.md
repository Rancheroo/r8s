# Bundle Mode Critical Bug Report - "client not initialized"

**Date:** November 28, 2025  
**Severity:** 🔴 CRITICAL - Bundle mode completely broken  
**Status:** CONFIRMED - Regression introduced

---

## Problem

Bundle mode TUI displays "Error: client not initialized" and cannot be used.

**Command that fails:**
```bash
./bin/r8s tui --bundle=example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz
```

**Error displayed in TUI:**
```
Error: client not initialized

Press 'q' to quit
```

---

## Root Cause

**Location:** `internal/tui/app.go`

### Bug #1: fetchClusters() doesn't use dataSource (lines 2486-2510)

```go
func (a *App) fetchClusters() tea.Cmd {
    return func() tea.Msg {
        // Only use mock data if explicitly in mock mode
        if a.offlineMode && a.config.MockMode {
            mockClusters := a.getMockClusters()
            return clustersMsg{clusters: mockClusters}
        }

        if a.client == nil {  // ❌ BUG: Fails in bundle mode!
            return errMsg{fmt.Errorf("client not initialized")}
        }
        
        collection, err := a.client.ListClusters()
        // ...
    }
}
```

**Problem:** Bundle mode sets `a.client = nil` because there's no live Rancher connection. The function immediately returns "client not initialized" error instead of checking `a.dataSource` first.

### Bug #2: fetchProjects() has same issue (lines 2514-2563)

```go
func (a *App) fetchProjects(clusterID string) tea.Cmd {
    return func() tea.Msg {
        // Only use mock data if explicitly in mock mode
        if a.offlineMode && a.config.MockMode {
            // ...
        }

        if a.client == nil {  // ❌ BUG: Fails in bundle mode!
            return errMsg{fmt.Errorf("client not initialized")}
        }
        
        collection, err := a.client.ListProjects(clusterID)
        // ...
    }
}
```

### Correct Pattern: fetchNamespaces() (lines 2567-2596)

```go
func (a *App) fetchNamespaces(clusterID, projectID string) tea.Cmd {
    return func() tea.Msg {
        // ✅ CORRECT: Try to get namespaces from data source first
        if a.dataSource != nil {
            namespaces, err := a.dataSource.GetNamespaces(clusterID, projectID)
            if err == nil {
                return namespacesMsg{namespaces: namespaces}
            }
            // Handle error...
        }

        // Then fallback to mock or return error
        if a.offlineMode && a.config.MockMode {
            // ...
        }

        return errMsg{fmt.Errorf("no data source available")}
    }
}
```

**This function works correctly** because it checks `a.dataSource` before checking `a.client`.

---

## Impact

**Severity:** CRITICAL

### What's Broken
- ❌ Bundle TUI mode completely unusable
- ❌ Users cannot browse bundle data in TUI
- ❌ Error appears immediately on launch
- ❌ No way to view clusters, projects, or any resources from bundle

### What Works
- ✅ Bundle import CLI works correctly
- ✅ kubectl resource parsing works (96 CRDs, 29 deployments, etc.)
- ✅ Bundle extraction and metadata parsing works
- ✅ Mock mode (`--mockdata`) works

### User Experience
1. User extracts bundle or uses archive
2. User runs: `r8s tui --bundle=<path>`
3. TUI shows red "Error: client not initialized" message
4. User cannot proceed - must press 'q' to quit
5. **Complete failure** - no functionality available

---

## Why This Wasn't Caught Earlier

1. **Bundle import tests pass** because they only test the CLI import command, not TUI launch
2. **Headless timeout tests pass** because they just check "doesn't crash immediately" - they don't verify the TUI actually displays data
3. **My earlier testing** showed "no panic" which was true, but I didn't validate that the TUI was **functional**
4. **The panic fix** (safeRowString) addressed crash issues but this is a **separate logic bug**

---

## Expected Behavior

When launching TUI in bundle mode:
1. TUI should detect bundle data source
2. Use `a.dataSource.GetClusters()` to load cluster data from bundle
3. Display cluster list (should show 1 cluster from the bundle)
4. Allow navigation to projects, namespaces, pods, etc.

---

## Fix Required

### fetchClusters() needs dataSource check

```go
func (a *App) fetchClusters() tea.Cmd {
    return func() tea.Msg {
        // ✅ ADD: Try data source first (for bundle mode)
        if a.dataSource != nil {
            clusters, err := a.dataSource.GetClusters()
            if err == nil {
                return clustersMsg{clusters: clusters}
            }
            // Log error but continue to try other sources
        }

        // Only use mock data if explicitly in mock mode
        if a.offlineMode && a.config.MockMode {
            mockClusters := a.getMockClusters()
            return clustersMsg{clusters: mockClusters}
        }

        if a.client == nil {
            return errMsg{fmt.Errorf("client not initialized")}
        }

        collection, err := a.client.ListClusters()
        // ...
    }
}
```

### fetchProjects() needs same fix

```go
func (a *App) fetchProjects(clusterID string) tea.Cmd {
    return func() tea.Msg {
        // ✅ ADD: Try data source first (for bundle mode)
        if a.dataSource != nil {
            projects, namespaceCounts, err := a.dataSource.GetProjects(clusterID)
            if err == nil {
                return projectsMsg{projects: projects, namespaceCounts: namespaceCounts}
            }
            // Log error but continue
        }

        // Only use mock data if explicitly in mock mode
        if a.offlineMode && a.config.MockMode {
            // ...
        }

        if a.client == nil {
            return errMsg{fmt.Errorf("client not initialized")}
        }

        collection, err := a.client.ListProjects(clusterID)
        // ...
    }
}
```

---

## Verification Steps

After fix is applied:

### Test 1: Bundle TUI Launch
```bash
./bin/r8s tui --bundle=example-log-bundle/*.tar.gz
```
**Expected:** TUI displays cluster list, no "client not initialized" error

### Test 2: Bundle Navigation
1. Launch bundle TUI
2. Navigate to cluster (Enter)
3. Navigate to project (Enter)
4. Navigate to namespace (Enter)
5. View pods, deployments, services

**Expected:** All navigation works, data from bundle displayed

### Test 3: Directory Mode
```bash
./bin/r8s tui --bundle=/tmp/extracted-bundle/
```
**Expected:** Same functionality as archive mode

### Test 4: No Regression
```bash
./bin/r8s tui --mockdata
```
**Expected:** Mock mode still works

---

## Related Issues

- **BUG-003:** kubectl path resolution (FIXED) - this allowed bundle import to parse resources
- **Panic fix:** safeRowString() (FIXED) - this prevented crashes on nil data
- **This bug:** dataSource not checked in fetchClusters/fetchProjects (NEW)

---

## DataSource Interface

Bundle mode should use this interface (defined in `internal/tui/datasource.go`):

```go
type DataSource interface {
    GetClusters() ([]rancher.Cluster, error)
    GetProjects(clusterID string) ([]rancher.Project, map[string]int, error)
    GetNamespaces(clusterID, projectID string) ([]rancher.Namespace, error)
    GetPods(clusterID, namespaceID string) ([]rancher.Pod, error)
    GetDeployments(clusterID, namespaceID string) ([]rancher.Deployment, error)
    GetServices(clusterID, namespaceID string) ([]rancher.Service, error)
    GetCRDs(clusterID string) ([]rancher.CRD, error)
}
```

**BundleDataSource** should implement this interface and return data from the loaded bundle.

---

## Test Coverage Gap

Current testing validates:
- ✅ Bundle extraction
- ✅ Resource parsing
- ✅ No crashes/panics

Missing tests:
- ❌ TUI actually displays data in bundle mode
- ❌ Navigation works in bundle mode  
- ❌ Resource views populate in bundle mode

**Recommendation:** Add functional tests that verify TUI behavior, not just "doesn't crash"

---

## Conclusion

**Bundle mode is completely broken** due to missing `dataSource` checks in `fetchClusters()` and `fetchProjects()`. These functions work for live API mode and mock mode, but fail immediately in bundle mode.

The fix is straightforward: follow the same pattern used by `fetchNamespaces()` and check `a.dataSource` before checking `a.client`.

**Priority:** P0 - Must fix before any bundle mode can be used

---

**Report Status:** COMPLETE  
**Next Action:** Developer must fix fetchClusters() and fetchProjects()  
**Testing Required:** Full TUI navigation testing after fix
