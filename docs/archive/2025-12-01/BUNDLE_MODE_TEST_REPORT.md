# Bundle Mode Testing Report

**Date:** November 28, 2025  
**Tester:** Warp AI Testing Agent  
**Test Type:** Bundle Mode Functional Testing  
**Bundle Tested:** `example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz`

---

## Executive Summary

Bundle mode testing revealed **1 CRITICAL bug** that completely breaks TUI functionality. The bundle import command works correctly, but TUI initialization fails due to a path resolution bug in kubectl resource parsing.

**Status:** 🔴 **BUNDLE TUI MODE BROKEN**

### Test Results
- ✅ **Bundle Import CLI:** PASS
- ❌ **Bundle TUI Mode:** FAIL (critical bug)
- 📋 **Root Cause:** Identified and documented

---

## Test Environment

- **OS:** Linux/Ubuntu
- **Terminal:** bash 5.2.21
- **r8s Version:** Latest (commit 4e19414)
- **Test Bundle:** RKE2 support bundle from node `w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09`
- **Bundle Size:** 8.93 MB compressed, ~30 MB extracted
- **Bundle Contents:** 337 files, 86 pods, 176 log files, 33 kubectl resource files

---

## Test Execution

### Test 1: Bundle Import (CLI)

**Command:**
```bash
./bin/r8s bundle import --path=example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz --limit=100
```

**Result:** ✅ **PASS**

**Output:**
```
Importing bundle from: example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz

Bundle imported successfully!

Extract Path: /tmp/r8s-bundle-3049774990
Bundle Type: rke2
Node: w-guard-wg-cp-svtk6-lqtxw
RKE2 Version: v1.31.2+rke2r1
K8s Version: unknown
Files: 337
Total Size: 8.93 MB

Pods Inventory:
  Total Pods: 86
  Current Logs: 86
  Previous Logs: 0

Log Files:
  Total: 176
  Pod Logs: 176
  System Logs: 0
```

**Observations:**
- Bundle extraction successful
- Pod inventory working (86 pods found)
- Log file inventory working (176 files found)
- Manifest parsing working
- **WARNINGS PRESENT** (see below)

**Warnings Logged:**
```
Warning: Failed to parse CRDs from bundle: open /tmp/r8s-bundle-3049774990/rke2/kubectl/crds: no such file or directory
Warning: Failed to parse Deployments from bundle: open /tmp/r8s-bundle-3049774990/rke2/kubectl/deployments: no such file or directory
Warning: Failed to parse Services from bundle: open /tmp/r8s-bundle-3049774990/rke2/kubectl/services: no such file or directory
Warning: Failed to parse Namespaces from bundle: open /tmp/r8s-bundle-3049774990/rke2/kubectl/namespaces: no such file or directory
```

**Analysis:**
These warnings indicate the kubectl resource parsers are looking in the wrong location. The paths they're checking don't exist because they're missing the node name subdirectory.

---

### Test 2: Bundle TUI Launch

**Command:**
```bash
./bin/r8s tui --bundle=example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz
```

**Result:** ❌ **FAIL**

**Error:**
```
client not initialized
```

**Behavior:**
- TUI fails to initialize
- Application exits with error
- No TUI screen displayed
- Same warnings as Test 1 appear

**Root Cause:**
TUI initialization requires valid resource data (CRDs, Deployments, Services, Namespaces). Since the kubectl parsers fail to find these files, the bundle object has empty arrays for all resources, causing TUI initialization to fail.

---

## Root Cause Investigation

### Bundle Structure Analysis

**Expected Path (by parser):**
```
/tmp/r8s-bundle-{id}/rke2/kubectl/crds
```

**Actual Path (in bundle):**
```
/tmp/r8s-bundle-{id}/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09/rke2/kubectl/crds
                      ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
                      Node name wrapper directory (missing in parser path)
```

### Verification of File Existence

The user confirmed kubectl files exist when the bundle is unpacked:
```
./example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09/rke2/kubectl/
├── api-resources
├── apiservices
├── clusterrolebindings
├── clusterroles
├── configmaps
├── crds                    ← EXISTS!
├── cronjobs
├── daemonsets
├── deployments             ← EXISTS!
├── endpoints
├── events
├── helmcharts
├── ingress
├── jobs
├── leases
├── mutatingwebhookconfigurations
├── namespaces              ← EXISTS!
├── networkpolicies
├── nodes
├── nodesdescribe
├── pods
├── pv
├── pvc
├── replicasets
├── rolebindings
├── roles
├── services                ← EXISTS!
├── statefulsets
├── validatingwebhookconfigurations
├── version
├── volumeattachments
```

**Total:** 33 kubectl resource files (all flat files containing kubectl command output)

### Code Analysis

**Problem Files:** `internal/bundle/kubectl.go`

**Affected Functions:**
1. `ParseCRDs()` - Line 15
2. `ParseDeployments()` - Line 83
3. `ParseServices()` - Line 131
4. `ParseNamespaces()` - Line 190

**Bug Pattern:**
All 4 functions build paths using `extractPath` directly:
```go
path := filepath.Join(extractPath, "rke2/kubectl/crds")
```

**Correct Pattern (used by other functions):**
```go
bundleRoot := getBundleRoot(extractPath)
path := filepath.Join(bundleRoot, "rke2/kubectl/crds")
```

**Helper Function:** `getBundleRoot()` exists in `internal/bundle/manifest.go` (lines 72-85)

**Functions Using It Correctly:**
- ✅ `ParseManifest()` → Works
- ✅ `InventoryPods()` → Works (86 pods found)
- ✅ `InventoryLogFiles()` → Works (176 files found)

**Functions NOT Using It:**
- ❌ `ParseCRDs()` → Fails
- ❌ `ParseDeployments()` → Fails
- ❌ `ParseServices()` → Fails
- ❌ `ParseNamespaces()` → Fails

---

## Bug Report Summary

### BUG-003: Bundle kubectl Path Resolution Issue

**Severity:** 🔴 CRITICAL

**Impact:**
- Bundle TUI mode completely broken
- All kubectl resource parsing fails
- CRDs, Deployments, Services, Namespaces not loaded
- Users cannot use TUI to browse bundle data

**Affected Components:**
- `internal/bundle/kubectl.go` (4 functions)

**Fix Complexity:** LOW (1-line change per function)

**Detailed Report:** See `BUG_003_BUNDLE_KUBECTL_PATH.md`

---

## Testing Methodology

### Approach
1. **CLI Testing:** Execute bundle import command to verify extraction and basic parsing
2. **TUI Testing:** Attempt to launch TUI in bundle mode
3. **Code Analysis:** Review source code to identify root cause
4. **Path Verification:** Confirm actual bundle structure vs. expected paths
5. **Comparison Analysis:** Compare working vs. broken parsing functions

### Tools Used
- `r8s bundle import` command
- `r8s tui --bundle` command
- Code reading and analysis
- User-provided bundle structure information

### Test Coverage
- ✅ Bundle import functionality
- ✅ Bundle extraction
- ✅ Manifest parsing
- ✅ Pod inventory
- ✅ Log file inventory
- ❌ kubectl resource parsing (blocked by bug)
- ❌ TUI initialization in bundle mode (blocked by bug)
- ❌ TUI navigation in bundle mode (blocked by bug)
- ❌ TUI resource views in bundle mode (blocked by bug)

---

## Recommendations

### Immediate Actions (For Developers)

1. **Fix BUG-003** by updating 4 functions in `internal/bundle/kubectl.go`:
   - Add `bundleRoot := getBundleRoot(extractPath)` at start of each function
   - Change `extractPath` to `bundleRoot` in path construction
   - Estimated time: 5 minutes

2. **Rebuild and Test:**
   ```bash
   make build
   ./bin/r8s bundle import --path=<bundle>    # Should have no warnings
   ./bin/r8s tui --bundle=<bundle>            # Should launch TUI
   ```

3. **Verify Resource Counts:**
   - CRDs should show actual count (not 0)
   - Deployments should show actual count (not 0)
   - Services should show actual count (not 0)
   - Namespaces should show actual count (not 0)

### Future Testing (After Fix)

Once BUG-003 is fixed, conduct:

1. **Bundle TUI Navigation Testing:**
   - Launch TUI in bundle mode
   - Navigate through all views (Pods, Deployments, Services, CRDs)
   - Test resource filtering and search
   - Test describe modal
   - Test log viewer

2. **Bundle Mode Feature Testing:**
   - Verify all mock mode features work in bundle mode
   - Test edge cases (empty resources, missing data)
   - Test bundle cleanup on exit

3. **Multiple Bundle Testing:**
   - Test with different bundle formats
   - Test with bundles from different RKE2 versions
   - Test with bundles with/without wrapper directories

### Documentation Updates

After fix:
1. Update `STATUS.md` - Mark bundle mode as working
2. Update `README.md` - Add bundle mode usage examples
3. Update `CHANGELOG.md` - Document bug fix
4. Create `BUNDLE_MODE_GUIDE.md` - User guide for bundle mode

---

## Related Documentation

- `BUG_003_BUNDLE_KUBECTL_PATH.md` - Detailed bug report with fix guidance
- `LOG_BUNDLE_ANALYSIS.md` - Bundle structure analysis
- `BUNDLE_DISCOVERY_COMPREHENSIVE.md` - Complete resource inventory
- `INTERACTIVE_TUI_TEST_REPORT.md` - Mock mode TUI testing results
- `TUI_UX_BUG_REPORT.md` - Mock mode bug findings

---

## Test Artifacts

### Files Created
- `BUG_003_BUNDLE_KUBECTL_PATH.md` - Bug report
- `BUNDLE_MODE_TEST_REPORT.md` - This report

### Test Data Used
- `example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz`

### Extracted Bundle Location
- `/tmp/r8s-bundle-3049774990/`

---

## Conclusion

Bundle mode CLI functionality (import command) works correctly, successfully extracting and inventorying pods and logs. However, **bundle mode TUI is completely broken** due to a path resolution bug in kubectl resource parsing.

The bug is well-understood and has a straightforward fix. The same helper function (`getBundleRoot()`) that works correctly for pod/log inventory needs to be used consistently in kubectl resource parsers.

**Testing Status:** Bundle mode testing **blocked** until BUG-003 is fixed.

**Next Steps:**
1. Developer fixes BUG-003
2. Resume bundle mode TUI testing
3. Conduct comprehensive bundle mode feature testing

---

**Report Status:** COMPLETE  
**Bug Severity:** CRITICAL  
**Developer Action Required:** YES
