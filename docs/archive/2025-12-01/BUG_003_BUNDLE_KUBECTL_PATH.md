# BUG-003: Bundle Mode TUI Fails - kubectl Path Resolution Issue

**Date:** November 28, 2025  
**Severity:** CRITICAL  
**Component:** Bundle Mode / kubectl Parser  
**Status:** Root Cause Identified

---

## Summary

Bundle mode TUI fails to initialize with "client not initialized" error despite kubectl resource files existing in the extracted bundle.

## Environment

- **Test Bundle:** `example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz`
- **Mode:** Bundle mode (`--bundle` flag)
- **OS:** Linux/Ubuntu

## Reproduction Steps

1. Import bundle:
   ```bash
   ./bin/r8s bundle import --path=example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz --limit=100
   ```
   ✅ **Result:** SUCCESS - Bundle extracts to `/tmp/r8s-bundle-{id}/`

2. Launch TUI in bundle mode:
   ```bash
   ./bin/r8s tui --bundle=example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz
   ```
   ❌ **Result:** FAIL - TUI crashes with "client not initialized"

## Error Messages

During bundle import, warnings appear:
```
Warning: Failed to parse CRDs from bundle: open /tmp/r8s-bundle-{id}/rke2/kubectl/crds: no such file or directory
Warning: Failed to parse Deployments from bundle: open /tmp/r8s-bundle-{id}/rke2/kubectl/deployments: no such file or directory
Warning: Failed to parse Services from bundle: open /tmp/r8s-bundle-{id}/rke2/kubectl/services: no such file or directory
Warning: Failed to parse Namespaces from bundle: open /tmp/r8s-bundle-{id}/rke2/kubectl/namespaces: no such file or directory
```

## Root Cause Analysis

### Actual Bundle Structure

When unpacked, the bundle has this structure:
```
/tmp/r8s-bundle-{id}/
└── w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09/    ← Node name directory
    ├── rke2/
    │   ├── kubectl/
    │   │   ├── crds                    ← EXISTS (flat file)
    │   │   ├── deployments             ← EXISTS (flat file)
    │   │   ├── services                ← EXISTS (flat file)
    │   │   ├── namespaces              ← EXISTS (flat file)
    │   │   ├── pods                    ← EXISTS (flat file)
    │   │   └── [30+ other files]
    │   └── podlogs/
    └── systeminfo/
```

**Key Finding:** kubectl outputs are **flat files**, not directories. Each file contains kubectl command output (e.g., `kubectl get crds -A -o wide`).

### Parser Path Mismatch

**Problem Location:** `internal/bundle/kubectl.go` lines 15, 83, 131, 190

The kubectl parsing functions use:
```go
path := filepath.Join(extractPath, "rke2/kubectl/crds")
```

But the actual path is:
```
/tmp/r8s-bundle-{id}/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09/rke2/kubectl/crds
                      ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ Missing node name!
```

### Solution Exists But Not Used

**Location:** `internal/bundle/manifest.go` lines 72-85

The `getBundleRoot()` function **already exists** and correctly handles wrapper directories:

```go
func getBundleRoot(extractPath string) string {
    // Check if there's a single wrapper directory
    entries, err := os.ReadDir(extractPath)
    if err == nil && len(entries) == 1 && entries[0].IsDir() {
        // Check if this wrapper contains the bundle
        wrapperDir := filepath.Join(extractPath, entries[0].Name())
        rke2Dir := filepath.Join(wrapperDir, "rke2")
        if _, err := os.Stat(rke2Dir); err == nil {
            return wrapperDir
        }
    }
    return extractPath
}
```

This function is used by:
- ✅ `ParseManifest()` - Works correctly
- ✅ `InventoryPods()` - Works correctly (86 pods found)
- ✅ `InventoryLogFiles()` - Works correctly (176 log files found)
- ❌ `ParseCRDs()` - **NOT USED** (bug)
- ❌ `ParseDeployments()` - **NOT USED** (bug)
- ❌ `ParseServices()` - **NOT USED** (bug)
- ❌ `ParseNamespaces()` - **NOT USED** (bug)

## Impact Assessment

**Severity:** CRITICAL

### What Works
- ✅ Bundle import command
- ✅ Bundle extraction
- ✅ Pod inventory (86 pods found)
- ✅ Log file inventory (176 files found)
- ✅ Manifest parsing

### What Breaks
- ❌ Bundle TUI mode completely unusable
- ❌ CRDs not parsed (0 instead of actual count)
- ❌ Deployments not parsed (0 instead of actual count)
- ❌ Services not parsed (0 instead of actual count)
- ❌ Namespaces not parsed (0 instead of actual count)
- ❌ TUI initialization fails due to empty resource lists

### User Impact
Bundle mode is **completely broken** for TUI usage. Users cannot browse bundle resources in the TUI.

## Technical Details

### Affected Files
1. `internal/bundle/kubectl.go` - All 4 Parse functions (lines 14, 82, 130, 189)

### Correct Implementation Pattern
The pod/log inventory functions show the correct pattern:

```go
// CORRECT ✅ - from InventoryPods() line 171
bundleRoot := getBundleRoot(extractPath)
podlogsDir := filepath.Join(bundleRoot, "rke2", "podlogs")
```

### Incorrect Implementation Pattern
The kubectl parsers use the wrong pattern:

```go
// INCORRECT ❌ - from ParseCRDs() line 15
path := filepath.Join(extractPath, "rke2/kubectl/crds")
```

## Fix Approach (For Developers)

**Required Changes:**
Each of the 4 kubectl parsing functions needs one additional line at the start:

```go
// ParseCRDs, ParseDeployments, ParseServices, ParseNamespaces
func ParseXXX(extractPath string) ([]Type, error) {
    bundleRoot := getBundleRoot(extractPath)  // ← ADD THIS LINE
    path := filepath.Join(bundleRoot, "rke2/kubectl/xxx")  // ← CHANGE extractPath to bundleRoot
    // ... rest of function unchanged
}
```

**Files to Modify:**
- `internal/bundle/kubectl.go` (4 functions)

**No API Changes Required:**
- Function signatures remain the same
- Return types unchanged
- Backward compatible (works with both wrapper and non-wrapper bundles)

## Verification Steps

After fix is applied:

1. **Import Test:**
   ```bash
   ./bin/r8s bundle import --path=example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz
   ```
   Expected: No warnings about missing kubectl files

2. **TUI Test:**
   ```bash
   ./bin/r8s tui --bundle=example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz
   ```
   Expected: TUI launches successfully

3. **Resource View Tests:**
   - Navigate to CRDs view (C key)
   - Navigate to Deployments view
   - Navigate to Services view
   - Navigate to Namespaces view
   - Verify resource counts match bundle content

## Related Documentation

- `LOG_BUNDLE_ANALYSIS.md` - Expected bundle structure
- `BUNDLE_DISCOVERY_COMPREHENSIVE.md` - Complete resource inventory
- Bundle contains 33 kubectl resource files in `rke2/kubectl/` directory

## Testing Notes

This bug was discovered during systematic bundle mode testing. The bundle import command succeeds and correctly inventories pods/logs, but kubectl resource parsing fails silently (warnings logged), causing TUI initialization to fail.

The fix is straightforward: use the existing `getBundleRoot()` helper function consistently across all bundle parsing code.

---

**Test Report by:** Warp AI Testing Agent  
**Test Date:** November 28, 2025  
**Test Type:** Bundle Mode Integration Testing  
**Test Method:** CLI testing + code analysis
