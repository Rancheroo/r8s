# r9s Fix Specification for Cline

**Date:** 2025-11-26  
**Priority:** HIGH  
**Scope:** Critical bug fix + CRD enhancements  

---

## Executive Summary

**Critical Issue Found:** Deployments view crashes with 500 error due to JSON unmarshaling failure.  
**Status:** CRD functionality is working perfectly - no fixes needed there, only enhancements requested.  
**Action Required:** Fix Deployment scale field handling + Add CRD descriptions for Longhorn/Rancher resources.

---

## Issue #1: CRITICAL - Deployments View 500 Error

### Problem Description

**Error Message:**
```
Error: failed to fetch deployments: failed to decode response: 
json: cannot unmarshal number into Go struct field Deployment.data.scale 
of type rancher.DeploymentScale
```

**Severity:** HIGH - Completely breaks Deployments view  
**Affects:** All namespaces when trying to view Deployments  
**Root Cause:** The Rancher API returns `scale` as a number in some cases, but our code expects it to always be an object (`DeploymentScale` struct).

### Current Code

**File:** `internal/rancher/types.go` (lines 238, 252-257)

```go
type Deployment struct {
    // ... other fields ...
    Scale *DeploymentScale `json:"scale,omitempty"`
    // ...
}

type DeploymentScale struct {
    Scale int `json:"scale"` // Desired replicas
    Ready int `json:"ready"` // Ready replicas
    Total int `json:"total"` // Total replicas
}
```

### Root Cause Analysis

The Rancher API can return the `scale` field in TWO different formats:

**Format 1 (Object):**
```json
{
  "scale": {
    "scale": 1,
    "ready": 1,
    "total": 1
  }
}
```

**Format 2 (Number):**
```json
{
  "scale": 1
}
```

Our current code assumes Format 1 (object) but the API is returning Format 2 (number), causing the unmarshal error.

### Solution

We need to handle both formats. There are two approaches:

#### Approach A: Custom JSON Unmarshaler (Recommended)

**Pros:** Clean, handles both formats transparently  
**Cons:** Slightly more code

**Implementation:**

1. Change the `Scale` field to use `json.RawMessage`:

```go
// File: internal/rancher/types.go

type Deployment struct {
    ID          string `json:"id"`
    Type        string `json:"type"`
    Name        string `json:"name"`
    NamespaceID string `json:"namespaceId"`
    State       string `json:"state"`
    
    // Use RawMessage to handle both number and object formats
    ScaleRaw json.RawMessage `json:"scale,omitempty"`
    
    // Parsed scale data (populated after unmarshaling)
    Scale *DeploymentScale `json:"-"`
    
    // Alternative: Direct fields if scale is not available
    Replicas          int               `json:"replicas"`
    AvailableReplicas int               `json:"availableReplicas"`
    ReadyReplicas     int               `json:"readyReplicas"`
    UpToDateReplicas  int               `json:"updatedReplicas"`
    UpdatedReplicas   int               `json:"upToDateReplicas"` // Alternative field name
    Created           time.Time         `json:"created"`
    Labels            map[string]string `json:"labels,omitempty"`
    Annotations       map[string]string `json:"annotations,omitempty"`
    Links             map[string]string `json:"links"`
    Actions           map[string]string `json:"actions"`
}

// Add UnmarshalJSON method for Deployment
func (d *Deployment) UnmarshalJSON(data []byte) error {
    // Define a temporary type to avoid recursion
    type Alias Deployment
    aux := &struct {
        *Alias
    }{
        Alias: (*Alias)(d),
    }
    
    if err := json.Unmarshal(data, &aux); err != nil {
        return err
    }
    
    // Parse the scale field
    if len(d.ScaleRaw) > 0 {
        // Try to unmarshal as object first
        var scaleObj DeploymentScale
        if err := json.Unmarshal(d.ScaleRaw, &scaleObj); err == nil {
            d.Scale = &scaleObj
        } else {
            // Try as number
            var scaleNum int
            if err := json.Unmarshal(d.ScaleRaw, &scaleNum); err == nil {
                d.Scale = &DeploymentScale{
                    Scale: scaleNum,
                    Ready: scaleNum, // Assume ready = scale for number format
                    Total: scaleNum,
                }
            }
        }
    }
    
    return nil
}
```

#### Approach B: Make Scale a Pointer to Interface (Alternative)

**Pros:** Simpler type definition  
**Cons:** More complex usage in app.go

```go
// Simpler but requires type assertion in app.go
Scale interface{} `json:"scale,omitempty"`
```

### Display Logic Update

**File:** `internal/tui/app.go` (around line 667-670)

**Current code:**
```go
rows = append(rows, table.NewRow(table.RowData{
    "name":      deployment.Name,
    "namespace": namespaceName,
    "ready":     fmt.Sprintf("%d/%d", deployment.ReadyReplicas, deployment.Replicas),
    "uptodate":  fmt.Sprintf("%d", deployment.UpToDateReplicas),
    "available": fmt.Sprintf("%d", deployment.AvailableReplicas),
}))
```

**Updated code with proper fallback:**
```go
// Extract replica counts with fallback logic
var desired, ready, uptodate, available int

if deployment.Scale != nil {
    // Use Scale object if available
    desired = deployment.Scale.Scale
    ready = deployment.Scale.Ready
    uptodate = deployment.Scale.Total
    available = ready // Best approximation
} else {
    // Fall back to direct fields
    desired = deployment.Replicas
    ready = deployment.ReadyReplicas
    uptodate = deployment.UpToDateReplicas
    if uptodate == 0 {
        uptodate = deployment.UpdatedReplicas // Try alternative field
    }
    available = deployment.AvailableReplicas
}

rows = append(rows, table.NewRow(table.RowData{
    "name":      deployment.Name,
    "namespace": namespaceName,
    "ready":     fmt.Sprintf("%d/%d", ready, desired),
    "uptodate":  fmt.Sprintf("%d", uptodate),
    "available": fmt.Sprintf("%d", available),
}))
```

### Testing Requirements

After implementing the fix, test:

1. ✅ Deployments view loads without errors
2. ✅ Replica counts display correctly (not 0/0)
3. ✅ Works across multiple namespaces
4. ✅ Handles both API response formats
5. ✅ No regression in other views

### Files to Modify

1. **`internal/rancher/types.go`**
   - Add `json.RawMessage` field for `ScaleRaw`
   - Add `UnmarshalJSON` method for `Deployment`
   - Update comments

2. **`internal/tui/app.go`**
   - Update deployment table rendering logic (lines ~664-673)
   - Add proper fallback for replica count extraction

---

## Enhancement #1: CRD Descriptions (Low Priority)

### Current State

✅ **CRD functionality is FULLY WORKING:**
- 96 CRDs load successfully
- Instance counts display correctly
- CRD details accessible with 'i' toggle
- Instance browsing works for all tested CRDs:
  - ✅ Longhorn CRDs (volumes, backuptargets, settings, etc.)
  - ✅ Monitoring CRDs (Alertmanager, Prometheus, ServiceMonitor, etc.)
  - ✅ Rancher/Cattle CRDs (AuthConfig, Settings, etc.)

### Enhancement Request

**Goal:** Add human-readable descriptions for common CRD types so users understand what they are.

**Priority:** LOW - This is nice-to-have, not critical  
**Complexity:** MEDIUM - Requires documentation research

### Implementation Approach

**Option 1: Static Description Map (Recommended)**

Add a description lookup function:

```go
// File: internal/tui/crd_descriptions.go (new file)

package tui

// getCRDDescription returns a human-readable description for known CRD types
func getCRDDescription(group, kind string) string {
    key := group + "/" + kind
    
    descriptions := map[string]string{
        // Longhorn Storage CRDs
        "longhorn.io/Volume": "Persistent storage volume managed by Longhorn",
        "longhorn.io/Engine": "Longhorn volume engine instance",
        "longhorn.io/Replica": "Longhorn volume replica instance",
        "longhorn.io/InstanceManager": "Manages Longhorn engine and replica instances",
        "longhorn.io/BackingImage": "Backing image for Longhorn volumes",
        "longhorn.io/BackingImageDataSource": "Data source for Longhorn backing images",
        "longhorn.io/BackingImageManager": "Manages Longhorn backing images",
        "longhorn.io/BackupTarget": "Target location for Longhorn backups",
        "longhorn.io/BackupVolume": "Backup of a Longhorn volume",
        "longhorn.io/Backup": "Individual backup of Longhorn volume data",
        "longhorn.io/EngineImage": "Longhorn engine image version",
        "longhorn.io/Node": "Longhorn node configuration and status",
        "longhorn.io/Orphan": "Orphaned Longhorn resource requiring cleanup",
        "longhorn.io/RecurringJob": "Scheduled job for backups or snapshots",
        "longhorn.io/Setting": "Longhorn configuration setting",
        "longhorn.io/ShareManager": "Manages shared access to Longhorn volumes",
        "longhorn.io/Snapshot": "Point-in-time snapshot of a Longhorn volume",
        "longhorn.io/SupportBundle": "Diagnostic bundle for Longhorn troubleshooting",
        "longhorn.io/SystemBackup": "System-level Longhorn backup",
        "longhorn.io/SystemRestore": "System-level Longhorn restore operation",
        "longhorn.io/VolumeAttachment": "Attachment of a Longhorn volume to a node",
        
        // Rancher Monitoring CRDs (monitoring.coreos.com)
        "monitoring.coreos.com/Alertmanager": "Prometheus Alertmanager deployment and configuration",
        "monitoring.coreos.com/AlertmanagerConfig": "Additional Alertmanager routing and receiver config",
        "monitoring.coreos.com/PodMonitor": "Scrape metrics from pods matching label selectors",
        "monitoring.coreos.com/Probe": "Monitor ingress or static targets with blackbox exporter",
        "monitoring.coreos.com/Prometheus": "Prometheus server deployment and configuration",
        "monitoring.coreos.com/PrometheusAgent": "Prometheus in Agent mode for remote write",
        "monitoring.coreos.com/PrometheusRule": "Prometheus alerting and recording rules",
        "monitoring.coreos.com/ScrapeConfig": "Custom scrape configuration for Prometheus",
        "monitoring.coreos.com/ServiceMonitor": "Scrape metrics from Kubernetes services",
        "monitoring.coreos.com/ThanosRuler": "Thanos Ruler for distributed rule evaluation",
        
        // Rancher Cattle Management CRDs
        "management.cattle.io/AuthConfig": "Authentication provider configuration (LDAP, SAML, etc.)",
        "management.cattle.io/Cluster": "Rancher managed Kubernetes cluster",
        "management.cattle.io/Project": "Rancher project for resource organization",
        "management.cattle.io/Setting": "Global Rancher configuration setting",
        "management.cattle.io/Token": "API token for Rancher authentication",
        "management.cattle.io/User": "Rancher user account",
        "management.cattle.io/RoleTemplate": "Rancher RBAC role template",
        "management.cattle.io/GlobalRole": "Cluster-wide Rancher role",
        "management.cattle.io/ClusterRoleTemplateBinding": "Binding of role template to cluster",
        "management.cattle.io/ProjectRoleTemplateBinding": "Binding of role template to project",
        
        // Rancher Catalog CRDs
        "catalog.cattle.io/App": "Deployed Helm chart application",
        "catalog.cattle.io/ClusterRepo": "Cluster-scoped Helm chart repository",
        "catalog.cattle.io/Operation": "Catalog operation (install, upgrade, uninstall)",
        
        // Fleet (GitOps) CRDs
        "fleet.cattle.io/Bundle": "Fleet bundle deployment",
        "fleet.cattle.io/BundleDeployment": "Deployment of a Fleet bundle to a cluster",
        "fleet.cattle.io/Cluster": "Fleet managed cluster",
        "fleet.cattle.io/ClusterGroup": "Group of clusters for Fleet targeting",
        "fleet.cattle.io/GitRepo": "Git repository tracked by Fleet",
        
        // Harvester CRDs (if present)
        "harvesterhci.io/VirtualMachine": "Harvester virtual machine",
        "harvesterhci.io/VirtualMachineImage": "VM image for Harvester",
        
        // Add more as needed...
    }
    
    if desc, ok := descriptions[key]; ok {
        return desc
    }
    
    // Fallback: try to extract from OpenAPI schema if available
    return ""
}
```

**Usage in CRD list view:**

```go
// File: internal/tui/app.go (in updateTable, ViewCRDs section)

// Add description column (optional, or show in detail view)
for _, crd := range a.crds {
    description := getCRDDescription(crd.Spec.Group, crd.Spec.Names.Kind)
    if description == "" && crd.Spec.Versions != nil && len(crd.Spec.Versions) > 0 {
        if crd.Spec.Versions[0].Schema != nil && 
           crd.Spec.Versions[0].Schema.OpenAPIV3Schema != nil {
            description = crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Description
        }
    }
    
    // Add to table row or detail view
}
```

**Option 2: Fetch from OpenAPI Schema**

The CRD's OpenAPIV3Schema already has a description field. We just need to parse it better:

```go
// File: internal/rancher/types.go

type OpenAPIV3Schema struct {
    Description string                      `json:"description,omitempty"`
    Type        string                      `json:"type,omitempty"`
    Properties  map[string]SchemaProperty   `json:"properties,omitempty"` // Add this
}

type SchemaProperty struct {
    Type        string `json:"type,omitempty"`
    Description string `json:"description,omitempty"`
}
```

### CRD Enhancement Files to Create/Modify

1. **`internal/tui/crd_descriptions.go`** (new file)
   - Add description lookup function
   - Add description map for common CRDs

2. **`internal/rancher/types.go`** (enhancement)
   - Expand OpenAPIV3Schema to capture more fields
   - Add helper methods to extract descriptions

3. **`internal/tui/app.go`** (enhancement)
   - Update CRD detail view to show descriptions
   - Optionally add description column to CRD list

---

## Enhancement #2: Better Error Messages (Optional)

### Current State

✅ Error messages are functional but could be more user-friendly.

### Enhancement Suggestions

**For 404 errors on CRD instances:**
```go
// File: internal/tui/app.go (in fetchCRDInstances)

if err != nil {
    // Check for 404 specifically
    if strings.Contains(err.Error(), "404") {
        return errMsg{fmt.Errorf(
            "API endpoint not found for %s instances. "+
            "This CRD version may not be served by the API. "+
            "Try a different version or check if the CRD is properly installed.",
            crdKind,
        )}
    }
    return errMsg{err}
}
```

**For deployment errors (after fix):**
```go
// Add more context to deployment fetch errors
if err != nil {
    return deploymentsMsg{
        deployments: []rancher.Deployment{},
        error: fmt.Errorf("failed to fetch deployments: %w", err),
    }
}
```

---

## Implementation Priority

### Must Fix (HIGH PRIORITY)

1. **✅ Issue #1: Deployments 500 Error**
   - Status: CRITICAL BUG
   - Impact: Blocks entire Deployments feature
   - Effort: 2-3 hours
   - Files: `types.go`, `app.go`

### Should Have (MEDIUM PRIORITY)

2. **Enhancement #2: Better Error Messages**
   - Status: IMPROVEMENT
   - Impact: Better UX
   - Effort: 1 hour
   - Files: `app.go`

### Nice to Have (LOW PRIORITY)

3. **Enhancement #1: CRD Descriptions**
   - Status: ENHANCEMENT
   - Impact: Better documentation
   - Effort: 4-6 hours (research + implementation)
   - Files: `crd_descriptions.go` (new), `types.go`, `app.go`

---

## Testing Checklist

### After Fixing Issue #1 (Deployments)

- [ ] Deployments view loads without errors in default namespace
- [ ] Replica counts show actual values (not 0/0)
- [ ] Test in System project namespaces
- [ ] Test in calico-system namespace
- [ ] Verify no regression in Pods view
- [ ] Verify no regression in Services view
- [ ] Check mock data still works (offline mode)

### After Enhancement #1 (CRD Descriptions)

- [ ] Longhorn CRDs show descriptions
- [ ] Monitoring CRDs show descriptions
- [ ] Cattle CRDs show descriptions
- [ ] Unknown CRDs gracefully show no description
- [ ] Description doesn't break layout
- [ ] Works in both online and offline mode

### After Enhancement #2 (Error Messages)

- [ ] 404 errors show helpful message
- [ ] 500 errors show helpful message
- [ ] Network errors show helpful message
- [ ] Error messages don't expose sensitive info

---

## Implementation Notes for Cline

### Development Environment

```bash
# Working directory
cd /home/bradmin/github/r9s

# Build command
make build
# or
go build -o bin/r9s main.go

# Test command
make test
# or
go test -v -race ./...

# Run application
./bin/r9s
```

### Key Files Reference

```
r9s/
├── internal/
│   ├── rancher/
│   │   ├── types.go        # Type definitions (FIX Issue #1 here)
│   │   └── client.go       # API client methods
│   └── tui/
│       ├── app.go          # Main TUI logic (FIX Issue #1 display here)
│       └── crd_descriptions.go  # NEW file for Enhancement #1
├── cmd/
│   └── root.go
├── main.go
└── Makefile
```

### Coding Standards

1. **Follow existing patterns** in the codebase
2. **Add comments** for complex logic
3. **Handle errors** gracefully
4. **Test with both** online and offline modes
5. **Preserve backward compatibility**
6. **Update mock data** if changing structures

### Git Workflow

```bash
# Create feature branch
git checkout -b fix/deployment-scale-error

# Make changes...

# Test
make test
./bin/r9s  # Manual testing

# Commit
git add .
git commit -m "fix: handle both number and object formats for deployment scale

- Add custom UnmarshalJSON for Deployment type
- Handle scale as number or object
- Update display logic with proper fallbacks
- Fixes #2 (deployment replica counts showing 0/0)

Tested:
- Default namespace deployments
- System project deployments
- Both API response formats
- No regression in other views"

# Push
git push origin fix/deployment-scale-error
```

---

## Expected Outcomes

### After Fix Implementation

1. ✅ Deployments view loads successfully
2. ✅ Replica counts display correctly (e.g., "1/1" instead of "0/0")
3. ✅ No errors in any namespace
4. ✅ Both API response formats handled
5. ✅ All existing functionality preserved

### After Enhancements

1. ✅ CRD descriptions help users understand resources
2. ✅ Error messages guide users to solutions
3. ✅ Better overall user experience

---

## Support Resources

- **Rancher API Docs:** https://rancher.com/docs/rancher/v2.x/en/api/
- **Longhorn Docs:** https://longhorn.io/docs/
- **Prometheus Operator CRDs:** https://prometheus-operator.dev/docs/operator/api/
- **Go JSON Package:** https://pkg.go.dev/encoding/json

---

**Document Version:** 1.0  
**Last Updated:** 2025-11-26  
**Status:** Ready for Implementation  
**Estimated Total Effort:** 6-10 hours (depending on enhancements included)
