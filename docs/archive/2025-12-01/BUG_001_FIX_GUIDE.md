# BUG-001: CRD Version Selection Fix Guide

**Bug ID**: BUG-001  
**Severity**: CRITICAL  
**File**: `internal/tui/app.go`  
**Lines**: 1395-1406

---

## Problem

When navigating to CRD instances (pressing Enter on a CRD), the code selects an API version to query. The current logic:

1. ✅ First tries to find a version with `storage: true`
2. ❌ Falls back to first version (`Versions[0]`) without checking `served: true`

If the first version has `served: false`, the API endpoint doesn't exist → **404 error**.

---

## Current Code (BROKEN)

```go
// Get the storage version
storageVersion := ""
for _, v := range selectedCRD.Spec.Versions {
    if v.Storage {
        storageVersion = v.Name
        break
    }
}
// Fallback to first version if no storage version
if storageVersion == "" && len(selectedCRD.Spec.Versions) > 0 {
    storageVersion = selectedCRD.Spec.Versions[0].Name  // ❌ BUG: Doesn't check served!
}
```

---

## Fixed Code

```go
// Get the storage version
storageVersion := ""
for _, v := range selectedCRD.Spec.Versions {
    if v.Storage {
        storageVersion = v.Name
        break
    }
}

// Fallback: find first served version
if storageVersion == "" {
    for _, v := range selectedCRD.Spec.Versions {
        if v.Served {
            storageVersion = v.Name
            break
        }
    }
}

// Final fallback: if no served version found (shouldn't happen), log error
if storageVersion == "" {
    // This shouldn't happen in a valid CRD, but handle gracefully
    a.error = fmt.Sprintf("CRD %s has no served versions", selectedCRD.Metadata.Name)
    return nil
}
```

---

## Alternative: Prefer Storage, then Served

More robust approach that prioritizes storage version but ensures it's served:

```go
// Helper function to select best CRD version
func selectBestCRDVersion(versions []rancher.CRDVersion) (string, error) {
    var storageVersion string
    var firstServedVersion string
    
    for _, v := range versions {
        // Track first served version as fallback
        if v.Served && firstServedVersion == "" {
            firstServedVersion = v.Name
        }
        
        // Prefer storage version if it's also served
        if v.Storage && v.Served {
            return v.Name, nil
        }
        
        // Track storage version even if not served
        if v.Storage {
            storageVersion = v.Name
        }
    }
    
    // Fallback 1: Use storage version even if not marked as served
    // (some CRDs have storage=true but don't explicitly mark served)
    if storageVersion != "" {
        return storageVersion, nil
    }
    
    // Fallback 2: Use first served version
    if firstServedVersion != "" {
        return firstServedVersion, nil
    }
    
    // No valid version found
    return "", fmt.Errorf("no served versions available")
}

// In handleEnter for ViewCRDs case:
storageVersion, err := selectBestCRDVersion(selectedCRD.Spec.Versions)
if err != nil {
    a.error = fmt.Sprintf("CRD %s: %v", selectedCRD.Metadata.Name, err)
    return nil
}
```

---

## Testing the Fix

### Unit Test

Add to `internal/tui/app_test.go`:

```go
func TestSelectBestCRDVersion(t *testing.T) {
    tests := []struct {
        name     string
        versions []rancher.CRDVersion
        want     string
        wantErr  bool
    }{
        {
            name: "storage version that is served",
            versions: []rancher.CRDVersion{
                {Name: "v1beta1", Served: true, Storage: false},
                {Name: "v1", Served: true, Storage: true},
            },
            want:    "v1",
            wantErr: false,
        },
        {
            name: "first version not served, second is",
            versions: []rancher.CRDVersion{
                {Name: "v1alpha1", Served: false, Storage: false},
                {Name: "v1beta1", Served: true, Storage: false},
                {Name: "v1", Served: true, Storage: true},
            },
            want:    "v1",
            wantErr: false,
        },
        {
            name: "no storage version, use first served",
            versions: []rancher.CRDVersion{
                {Name: "v1alpha1", Served: false, Storage: false},
                {Name: "v1beta1", Served: true, Storage: false},
            },
            want:    "v1beta1",
            wantErr: false,
        },
        {
            name: "no served versions",
            versions: []rancher.CRDVersion{
                {Name: "v1alpha1", Served: false, Storage: false},
                {Name: "v1beta1", Served: false, Storage: false},
            },
            want:    "",
            wantErr: true,
        },
        {
            name:     "empty versions list",
            versions: []rancher.CRDVersion{},
            want:     "",
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := selectBestCRDVersion(tt.versions)
            if (err != nil) != tt.wantErr {
                t.Errorf("selectBestCRDVersion() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("selectBestCRDVersion() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Test

Test with real Rancher instance that has CRDs with multiple versions:

1. Find a CRD with multiple versions where first version has `served: false`
2. Navigate: Clusters → 'C' → Select that CRD → Enter
3. Verify: Should show instances without 404 error

Example CRDs to test:
- `monitoring.coreos.com/servicemonitors` (often has v1alpha1 deprecated)
- `cert-manager.io/certificates` (often has v1alpha2, v1alpha3, v1 versions)

---

## Deployment Checklist

- [ ] Apply fix to `internal/tui/app.go`
- [ ] Add unit tests for version selection
- [ ] Run existing tests: `make test`
- [ ] Build binary: `make build`
- [ ] Test manually with `--mockdata`
- [ ] Test with real Rancher instance
- [ ] Test with CRDs having deprecated versions
- [ ] Update CHANGELOG.md
- [ ] Create PR with test results

---

## Related Files

- `internal/tui/app.go` - Main fix location (handleEnter function)
- `internal/rancher/types.go` - CRDVersion type definition (lines 138-143)
- `internal/tui/app_test.go` - Add unit tests here

---

## Expected Impact

**Before Fix**:
- Some CRDs show 404 errors when viewing instances
- Confusing error messages
- Blocking workflow for affected CRDs

**After Fix**:
- All CRDs with served versions work correctly
- Clear error message if no served versions exist
- Smooth navigation to CRD instances

---

## Rollout Strategy

1. **Stage 1**: Fix in development branch
2. **Stage 2**: Test with internal Rancher instances
3. **Stage 3**: Beta release with fix notes
4. **Stage 4**: Production release

---

## Questions?

Contact the r8s development team or file an issue at:
https://github.com/Rancheroo/r8s/issues
