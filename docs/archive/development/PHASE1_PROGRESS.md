# Phase 1: Log Viewing Foundation - IN PROGRESS

**Started:** November 27, 2025, 4:10 PM  
**Status:** 🟡 In Progress (Step 1/10 Complete)

---

## Completed Steps

### ✅ Step 1: Add ViewLogs to ViewType enum
- **File:** `internal/tui/app.go`
- **Change:** Added `ViewLogs` constant to ViewType enum (line 32)
- **Status:** Complete

---

## Remaining Steps

### Step 2: Add Log Context Fields to ViewContext
**File:** `internal/tui/app.go`  
**Location:** ViewContext struct (around line 36)  
**Changes Needed:**
```go
type ViewContext struct {
    viewType      ViewType
    clusterID     string
    clusterName   string
    projectID     string
    projectName   string
    namespaceID   string
    namespaceName string
    // Context for CRDs
    crdGroup    string
    crdVersion  string
    crdResource string
    crdKind     string
    crdScope    string
    // NEW: Context for logs  
    podName       string
    containerName string
}
```

### Step 3: Add Log Data Storage to App Struct
**File:** `internal/tui/app.go`  
**Location:** App struct (around line 48)  
**Changes Needed:**
```go
type App struct {
    // ... existing fields ...
    
    // Data for different views
    clusters     []rancher.Cluster
    projects     []rancher.Project
    namespaces   []rancher.Namespace
    pods         []rancher.Pod
    deployments  []rancher.Deployment
    services     []rancher.Service
    crds         []rancher.CRD
    crdInstances []map[string]interface{}
    logs         []string  // NEW: Log lines for current pod
    
    // ... rest of fields ...
}
```

### Step 4: Create internal/tui/views Directory
**Command:** `mkdir -p internal/tui/views`

### Step 5: Create logs.go Component
**File:** `internal/tui/views/logs.go`  
**Content:** See implementation in R8S_MIGRATION_PLAN.md Phase 1

### Step 6: Add 'l' Hotkey Handler
**File:** `internal/tui/app.go`  
**Location:** Update() function, key handler section  
**Changes Needed:**
```go
case "l":
    // Open logs for selected pod
    if a.currentView.viewType == ViewPods {
        return a, a.handleViewLogs()
    }
```

### Step 7: Add logsMsg Type
**File:** `internal/tui/app.go`  
**Location:** Messages section (end of file, around line 1590)  
**Changes Needed:**
```go
type logsMsg struct {
    logs []string
}
```

### Step 8: Implement fetchLogs() Function
**File:** `internal/tui/app.go`  
**Location:** Add after fetchServices() function  
**Changes Needed:** Full function implementation for API + mock fallback

### Step 9: Add handleViewLogs() Function
**File:** `internal/tui/app.go`  
**Location:** Add after handleDescribe() function  
**Changes Needed:** Navigate to logs view for selected pod

### Step 10: Update getBreadcrumb() for Logs
**File:** `internal/tui/app.go`  
**Location:** getBreadcrumb() function  
**Changes Needed:**
```go
case ViewLogs:
    return fmt.Sprintf("Cluster: %s > Project: %s > Namespace: %s > Pod: %s > Logs",
        a.currentView.clusterName, a.currentView.projectName, 
        a.currentView.namespaceName, a.currentView.podName)
```

### Step 11: Update getStatusText() for Logs
**File:** `internal/tui/app.go`  
**Location:** getStatusText() function  
**Changes Needed:**
```go
case ViewLogs:
    status = fmt.Sprintf(" %sViewing logs for %s | Press 'Esc' to go back | 'q' to quit ", 
        offlinePrefix, a.currentView.podName)
```

### Step 12: Handle logsMsg in Update()
**File:** `internal/tui/app.go`  
**Location:** Update() function message handling  
**Changes Needed:**
```go
case logsMsg:
    a.loading = false
    a.logs = msg.logs
    a.error = ""
    // Display logs in viewport (to be implemented)
```

---

## Testing Checklist

After implementation:
- [ ] Press `l` on a pod in Pods view
- [ ] Verify logs view opens
- [ ] Verify breadcrumb shows full path including pod name
- [ ] Verify logs display (mock data in offline mode)
- [ ] Verify `Esc` returns to Pods view
- [ ] Run `make test` - all 49 tests should still pass

---

## Next Actions

1. Continue implementing steps 2-12 above
2. Create basic viewport-based log viewer
3. Test with mock data
4. Document completion

---

**Estimated Remaining Time:** 30 minutes  
**Current Progress:** 1/12 steps (8%)
