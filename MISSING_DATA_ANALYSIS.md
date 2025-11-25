# r9s Missing Data Analysis

**Analysis Date:** 2025-11-25  
**Version:** dev (commit a4846b6)  
**Purpose:** Identify columns with missing data in live views vs. mock data

---

## Executive Summary

Several columns in live data views are displaying empty values or incorrect counts, while the same columns display correctly in mock/offline mode. This indicates issues with:
1. **Data extraction from Rancher API responses**
2. **Namespace counting logic for Projects**
3. **Field mapping between API responses and displayed data**

---

## Issues Identified

### 🔴 CRITICAL: Issue #1 - Pod NODE Column Empty

**View:** Pods  
**Column:** NODE  
**Severity:** HIGH  

**Current Behavior:**
- Live data: NODE column is completely empty (blank)
- Mock data: NODE column shows values like "worker-node-1", "worker-node-2", "worker-node-3"

**Expected Behavior:**
- Should display the node name where the pod is running (e.g., "worker-node-1")

**Code Location:**
- File: `internal/tui/app.go`
- Lines: 592-638 (Pods table rendering)
- Line 619: `"node": pod.NodeName,`

**Root Cause Analysis:**
- The code correctly reads `pod.NodeName` from the API
- Issue likely: Rancher API field name mismatch
- Rancher API may use different field name than `nodeName`

**Suggested Fix:**
```go
// Check actual Rancher API response - it might be:
// - node
// - nodeId
// - hostId
// - spec.nodeName (nested)
// Add debug logging to see actual API response
```

**Mock Data (Works Correctly):**
```go
NodeName: "worker-node-1",  // Line 1309, 1316, 1323, etc.
```

---

### 🔴 CRITICAL: Issue #2 - Deployment Replica Counts All Zero

**View:** Deployments  
**Columns:** READY, UP-TO-DATE, AVAILABLE  
**Severity:** HIGH  

**Current Behavior:**
- Live data: All three columns show "0" or "0/0"
  - READY: "0/0"
  - UP-TO-DATE: "0"
  - AVAILABLE: "0"
- Mock data: Shows actual counts
  - READY: "3/3", "2/2", "5/5", "3/4"
  - UP-TO-DATE: "3", "2", "5", "1"
  - AVAILABLE: "3", "2", "5", "3"

**Expected Behavior:**
- Should display actual replica counts from live deployments

**Code Location:**
- File: `internal/tui/app.go`
- Lines: 640-688 (Deployments table rendering)
- Line 667: `"ready": fmt.Sprintf("%d/%d", deployment.ReadyReplicas, deployment.Replicas),`
- Line 668: `"uptodate": fmt.Sprintf("%d", deployment.UpToDateReplicas),`
- Line 669: `"available": fmt.Sprintf("%d", deployment.AvailableReplicas),`

**Type Definition:**
- File: `internal/rancher/types.go`
- Lines: 207-223
```go
type Deployment struct {
    Replicas          int    `json:"replicas"`
    AvailableReplicas int    `json:"availableReplicas"`
    ReadyReplicas     int    `json:"readyReplicas"`
    UpToDateReplicas  int    `json:"updatedReplicas"`  // ⚠️ NOTE: JSON tag is "updatedReplicas"
}
```

**Root Cause Analysis:**
1. **Field name mismatch for UpToDateReplicas:**
   - Struct field: `UpToDateReplicas`
   - JSON tag: `updatedReplicas`
   - Rancher API likely returns: `updatedReplicas` ✓ (correct)
   
2. **Possible issues with other fields:**
   - Rancher API might use different field names:
     - `replicas` → might be `scale` or `spec.replicas`
     - `availableReplicas` → might be `status.availableReplicas`
     - `readyReplicas` → might be `status.readyReplicas`

**Suggested Fix:**
1. Verify actual Rancher API response structure for deployments
2. Check if fields are nested under `status` or `spec`
3. Possible fix:
```go
type Deployment struct {
    // ... other fields
    Status DeploymentStatus `json:"status"`
}

type DeploymentStatus struct {
    Replicas          int `json:"replicas"`
    AvailableReplicas int `json:"availableReplicas"`
    ReadyReplicas     int `json:"readyReplicas"`
    UpdatedReplicas   int `json:"updatedReplicas"`
}
```

**Mock Data (Works Correctly):**
```go
Replicas:          3,
AvailableReplicas: 3,
ReadyReplicas:     3,
UpToDateReplicas:  3,
```

---

### 🟡 IMPORTANT: Issue #3 - Project Namespace Counts Show Zero

**View:** Projects  
**Column:** NAMESPACES  
**Severity:** MEDIUM  

**Current Behavior:**
- Live data: NAMESPACES column shows "0" for all projects
- Mock data: Shows actual counts (e.g., "3", "5")

**Expected Behavior:**
- Should display the actual count of namespaces in each project

**Code Location:**
- File: `internal/tui/app.go`
- Lines: 509-549 (Projects table rendering)
- Line 520: `namespaceCount := a.projectNamespaceCounts[project.ID]`
- Line 530: `"namespaces": fmt.Sprintf("%d", namespaceCount),`

**Namespace Counting Logic:**
- Lines: 1923-1927 (fetchProjects function)
```go
// Count namespaces per project
namespaceCounts := make(map[string]int)
for _, project := range collection.Data {
    namespaceCounts[project.ID] = 0 // ⚠️ Real implementation would count namespaces
}
```

**Root Cause:**
- **KNOWN ISSUE:** Line 1926 comment explicitly states this is a placeholder
- The count is hardcoded to 0 for all projects in live mode
- Mock data provides real counts (lines 1901-1904, 1916-1919)

**Suggested Fix:**
The Projects view should fetch namespaces and count them:
```go
// Count namespaces per project
namespaceCounts := make(map[string]int)

// Fetch all namespaces for this cluster
nsCollection, err := a.client.ListNamespaces(clusterID)
if err == nil {
    // Count namespaces by project ID
    for _, ns := range nsCollection.Data {
        if ns.ProjectID != "" {
            namespaceCounts[ns.ProjectID]++
        }
    }
}

// If fetch failed, set to 0 or "-" to indicate unknown
for _, project := range collection.Data {
    if _, exists := namespaceCounts[project.ID]; !exists {
        namespaceCounts[project.ID] = 0
    }
}

return projectsMsg{projects: collection.Data, namespaceCounts: namespaceCounts}
```

**Alternative Fix (Performance-Focused):**
Don't fetch namespaces synchronously. Instead:
1. Show "-" initially for namespace counts
2. Fetch namespace counts asynchronously when Projects view loads
3. Update display when counts are available

**Mock Data (Works Correctly):**
```go
mockNamespaceCounts := map[string]int{
    "demo-project": 3,
    "system":       5,
}
```

---

## Working Correctly

### ✅ Issue-Free Views/Columns

1. **Clusters View** - All columns working:
   - NAME ✓
   - PROVIDER ✓
   - STATE ✓
   - AGE ✓

2. **Projects View** - Partial:
   - NAME ✓
   - DISPLAY NAME ✓
   - STATE ✓
   - NAMESPACES ❌ (see Issue #3)

3. **Namespaces View** - All columns working:
   - NAME ✓
   - STATE ✓
   - PROJECT ✓
   - AGE ✓

4. **Pods View** - Partial:
   - NAME ✓
   - NAMESPACE ✓
   - STATE ✓
   - NODE ❌ (see Issue #1)

5. **Deployments View** - Partial:
   - NAME ✓
   - NAMESPACE ✓
   - READY ❌ (see Issue #2)
   - UP-TO-DATE ❌ (see Issue #2)
   - AVAILABLE ❌ (see Issue #2)

6. **Services View** - All columns working:
   - NAME ✓
   - NAMESPACE ✓
   - TYPE ✓
   - CLUSTER-IP ✓
   - PORT(S) ✓

7. **CRDs View** - All columns working:
   - NAME ✓
   - GROUP ✓
   - KIND ✓
   - SCOPE ✓
   - INSTANCES ✓

---

## Comparison: Mock Data vs. Live Data

### Mock Data Structure (Working Reference)

**Pods:**
```go
{
    Name:        "nginx-deployment-6bccc6bf79-w6bbq",
    NamespaceID: namespaceName,
    State:       "Running",
    NodeName:    "worker-node-1",        // ✓ Populated
    Created:     time.Now().Add(-time.Hour * 2),
}
```

**Deployments:**
```go
{
    Name:              "nginx-deployment",
    NamespaceID:       namespaceName,
    State:             "active",
    Replicas:          3,                // ✓ Populated
    AvailableReplicas: 3,                // ✓ Populated
    ReadyReplicas:     3,                // ✓ Populated
    UpToDateReplicas:  3,                // ✓ Populated
    Created:           time.Now().Add(-time.Hour * 24),
}
```

**Services:**
```go
{
    Name:        "nginx-service",
    NamespaceID: namespaceName,
    State:       "active",
    ClusterIP:   "10.43.100.50",        // ✓ Populated
    Kind:        "ClusterIP",           // ✓ Populated
    Ports: []rancher.ServicePort{       // ✓ Populated
        {Name: "http", Protocol: "TCP", Port: 80, TargetPort: 8080},
    },
    Created: time.Now().Add(-time.Hour * 24),
}
```

### Live Data Issues

**Pods (Live API):**
```go
{
    Name:        "basic-web-b95b8bcb8-pbc9f",  // ✓ Has value
    NamespaceID: "c-xxx:default",              // ✓ Has value
    State:       "running",                     // ✓ Has value
    NodeName:    "",                            // ❌ EMPTY!
    Created:     time.Time{...},                // ✓ Has value
}
```

**Deployments (Live API):**
```go
{
    Name:              "basic-web",            // ✓ Has value
    NamespaceID:       "c-xxx:default",        // ✓ Has value
    State:             "active",               // ✓ Has value
    Replicas:          0,                      // ❌ Should be > 0
    AvailableReplicas: 0,                      // ❌ Should be > 0
    ReadyReplicas:     0,                      // ❌ Should be > 0
    UpToDateReplicas:  0,                      // ❌ Should be > 0
    Created:           time.Time{...},         // ✓ Has value
}
```

---

## Debugging Steps

### Step 1: Capture Raw API Responses

Add debug logging to capture actual Rancher API responses:

```go
// In internal/rancher/client.go, add logging before unmarshaling

func (c *Client) ListPods(clusterID, projectID, namespace string) (*PodCollection, error) {
    // ... existing code ...
    
    // Add before json.Unmarshal:
    log.Printf("DEBUG: Raw Pod API Response:\n%s\n", string(body))
    
    var collection PodCollection
    if err := json.Unmarshal(body, &collection); err != nil {
        return nil, fmt.Errorf("failed to unmarshal pods: %w", err)
    }
    
    // Add after unmarshal:
    log.Printf("DEBUG: Parsed Pods: %+v\n", collection.Data)
    
    return &collection, nil
}
```

### Step 2: Compare Field Names

1. Run r9s with debug logging enabled
2. Capture actual API responses
3. Compare with type definitions in `types.go`
4. Look for field name mismatches

### Step 3: Check Rancher API Documentation

Consult Rancher API documentation for:
- Correct field names for pod node information
- Deployment status structure
- Whether fields are nested (e.g., `status.replicas` vs. `replicas`)

---

## Recommended Fixes

### Fix Priority

1. **HIGH PRIORITY:** Issue #2 - Deployment replica counts (blocks deployment monitoring)
2. **HIGH PRIORITY:** Issue #1 - Pod NODE column (important for troubleshooting)
3. **MEDIUM PRIORITY:** Issue #3 - Project namespace counts (UX improvement)

### Implementation Order

1. **Add debug logging** to capture raw API responses
2. **Fix Deployment type definition** based on actual API structure
3. **Fix Pod node name field** based on actual API structure
4. **Implement namespace counting** for Projects view
5. **Add unit tests** to prevent regressions

### Testing Plan

After fixes are implemented:
1. Test with live Rancher instance
2. Verify all columns show data
3. Compare with mock data to ensure consistency
4. Test across multiple clusters/projects/namespaces
5. Update TEST_RESULTS.md with findings

---

## Code Changes Required

### File: `internal/rancher/types.go`

**Change 1: Deployment Structure**

Current:
```go
type Deployment struct {
    ID                string            `json:"id"`
    Type              string            `json:"type"`
    Name              string            `json:"name"`
    NamespaceID       string            `json:"namespaceId"`
    State             string            `json:"state"`
    Replicas          int               `json:"replicas"`
    AvailableReplicas int               `json:"availableReplicas"`
    ReadyReplicas     int               `json:"readyReplicas"`
    UpToDateReplicas  int               `json:"updatedReplicas"`
    Created           time.Time         `json:"created"`
    Labels            map[string]string `json:"labels,omitempty"`
    Annotations       map[string]string `json:"annotations,omitempty"`
    Links             map[string]string `json:"links"`
    Actions           map[string]string `json:"actions"`
}
```

Likely needed (verify with API):
```go
type Deployment struct {
    ID          string            `json:"id"`
    Type        string            `json:"type"`
    Name        string            `json:"name"`
    NamespaceID string            `json:"namespaceId"`
    State       string            `json:"state"`
    
    // Spec contains desired state
    Spec DeploymentSpec `json:"spec,omitempty"`
    
    // Status contains current state
    Status DeploymentStatus `json:"status,omitempty"`
    
    Created     time.Time         `json:"created"`
    Labels      map[string]string `json:"labels,omitempty"`
    Annotations map[string]string `json:"annotations,omitempty"`
    Links       map[string]string `json:"links"`
    Actions     map[string]string `json:"actions"`
}

type DeploymentSpec struct {
    Replicas int `json:"replicas"`
}

type DeploymentStatus struct {
    Replicas          int `json:"replicas"`
    AvailableReplicas int `json:"availableReplicas"`
    ReadyReplicas     int `json:"readyReplicas"`
    UpdatedReplicas   int `json:"updatedReplicas"`
}
```

**Change 2: Pod NodeName Field**

Current:
```go
type Pod struct {
    // ...
    NodeName     string            `json:"nodeName"`
    // ...
}
```

Check if it should be (verify with API):
```go
type Pod struct {
    // ...
    NodeName     string            `json:"nodeName"`  // or "node" or "nodeId"
    NodeID       string            `json:"nodeId,omitempty"`  // backup field
    // ...
}
```

### File: `internal/tui/app.go`

**Change 1: Update Deployment Table Rendering**

Current (lines 664-670):
```go
rows = append(rows, table.NewRow(table.RowData{
    "name":      deployment.Name,
    "namespace": namespaceName,
    "ready":     fmt.Sprintf("%d/%d", deployment.ReadyReplicas, deployment.Replicas),
    "uptodate":  fmt.Sprintf("%d", deployment.UpToDateReplicas),
    "available": fmt.Sprintf("%d", deployment.AvailableReplicas),
}))
```

Update to (if Status structure is used):
```go
// Use Status for current state, Spec for desired state
desiredReplicas := deployment.Spec.Replicas
if desiredReplicas == 0 {
    desiredReplicas = deployment.Status.Replicas // fallback
}

rows = append(rows, table.NewRow(table.RowData{
    "name":      deployment.Name,
    "namespace": namespaceName,
    "ready":     fmt.Sprintf("%d/%d", deployment.Status.ReadyReplicas, desiredReplicas),
    "uptodate":  fmt.Sprintf("%d", deployment.Status.UpdatedReplicas),
    "available": fmt.Sprintf("%d", deployment.Status.AvailableReplicas),
}))
```

**Change 2: Fix Namespace Counting**

Current (lines 1923-1927):
```go
// Count namespaces per project
namespaceCounts := make(map[string]int)
for _, project := range collection.Data {
    namespaceCounts[project.ID] = 0 // Real implementation would count namespaces
}
```

Replace with:
```go
// Count namespaces per project
namespaceCounts := make(map[string]int)

// Fetch namespaces to get accurate counts
nsCollection, err := a.client.ListNamespaces(clusterID)
if err == nil {
    // Count namespaces by project
    for _, ns := range nsCollection.Data {
        if ns.ProjectID != "" {
            namespaceCounts[ns.ProjectID]++
        }
    }
}

// Ensure all projects have an entry (even if 0)
for _, project := range collection.Data {
    if _, exists := namespaceCounts[project.ID]; !exists {
        namespaceCounts[project.ID] = 0
    }
}
```

---

## Additional Columns to Verify

While testing, also verify these columns that weren't explicitly checked:

### Pods View - Additional Columns (Mock Data Has)
- Ready (e.g., "1/1", "0/1") - Currently not displayed
- Restarts - Currently not displayed
- Age - Currently not displayed
- IP - Currently not displayed

### Deployments View - Additional Columns (Mock Data Has)
- Age - Currently not displayed

### Services View - Additional Columns (Mock Data Has)
- Age - Currently not displayed
- External IP (for LoadBalancer services) - Currently not displayed

---

## Summary Checklist

- [ ] Issue #1: Pod NODE column - Field name verification needed
- [ ] Issue #2: Deployment replica counts - Structure verification needed
- [ ] Issue #3: Project namespace counts - Implementation needed
- [ ] Add debug logging to capture raw API responses
- [ ] Update type definitions based on actual API structure
- [ ] Update table rendering logic to use corrected fields
- [ ] Test all fixes with live Rancher instance
- [ ] Add unit tests for data extraction
- [ ] Update documentation with correct field mappings

---

**Next Steps:**
1. Enable debug logging to capture actual API responses
2. Compare API responses with current type definitions
3. Update type definitions with correct field names/structure
4. Implement namespace counting for Projects view
5. Test thoroughly with live data
6. Update TEST_RESULTS.md

---

*Analysis completed by: Warp AI Assistant*  
*Date: 2025-11-25*  
*Documentation version: 1.0*
