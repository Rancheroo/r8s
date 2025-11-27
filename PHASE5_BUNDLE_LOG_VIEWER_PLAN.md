# Phase 5: Bundle Log Viewer - Implementation Plan

**Phase:** 5 of 6  
**Status:** Planning  
**Dependencies:** Phase 4 Complete ✅  
**Estimated Effort:** 60-90 minutes

---

## Objectives

Integrate the bundle import system with the TUI to enable:
1. Loading bundles in the TUI
2. Browsing pods from bundle data
3. Viewing logs from bundle files
4. Reusing all existing log viewer features (search, filters, colors)

---

## Current State Analysis

### What We Have ✅

**From Phase 4:**
- `bundle.Load()` - Loads a bundle from tar.gz
- `bundle.GetPod()` - Retrieves pod info
- `bundle.ReadLogFile()` - Reads log contents
- Bundle metadata (node, versions, pods, logs)
- CLI import command working

**From Phases 1-3:**
- Log viewer with viewport scrolling
- Search (/, n, N)
- Filters (Ctrl+E/W/A)
- Color coding (ERROR/WARN/INFO/DEBUG)
- Search highlighting
- Tail mode, container cycling

### What We Need 🎯

1. **Bundle Mode Toggle**
   - Flag to run TUI in bundle mode
   - Load bundle on startup
   - Show bundle metadata

2. **Pod List from Bundle**
   - Display pods from bundle inventory
   - Group by namespace
   - Show container counts

3. **Log Viewer Integration**
   - Read logs from bundle files (not API)
   - Support current/previous logs
   - Reuse existing viewer features

4. **Seamless UX**
   - No breaking changes to live mode
   - Clear mode indicator
   - Smooth transitions

---

## Architecture Design

### Option 1: Mode Flag Approach (RECOMMENDED)

**Concept:** Add `--bundle=<path>` flag to TUI startup

```bash
# Live mode (existing)
./bin/r8s

# Bundle mode (new)
./bin/r8s --bundle=example-log-bundle/bundle.tar.gz
```

**Pros:**
- Clean separation of concerns
- No changes to existing code paths
- Easy to switch between modes
- Clear user intent

**Cons:**
- Can't switch modes without restart
- Two separate code paths

### Option 2: Runtime Mode Switch

**Concept:** Add menu option to load bundle

**Pros:**
- More flexible
- Can switch between live and bundle

**Cons:**
- More complex state management
- Risk of breaking existing features
- Harder to test

**Decision:** Use Option 1 (Mode Flag) for Phase 5  
*Can add Option 2 in future if needed*

---

## Implementation Plan

### Step 1: Add Bundle Flag to TUI (10 min)

**File:** `cmd/root.go`

```go
var bundlePath string

func init() {
    rootCmd.Flags().StringVar(&bundlePath, "bundle", "", "Path to bundle for offline mode")
}
```

### Step 2: Create Bundle Data Source Interface (15 min)

**File:** `internal/tui/datasource.go` (new)

```go
// DataSource abstracts pod and log data retrieval
type DataSource interface {
    GetPods(namespace string) ([]Pod, error)
    GetLogs(namespace, pod, container string) ([]string, error)
    GetContainers(namespace, pod string) ([]string, error)
    IsOffline() bool
}

// LiveDataSource uses Rancher API
type LiveDataSource struct {
    client *rancher.Client
}

// BundleDataSource uses bundle files
type BundleDataSource struct {
    bundle *bundle.Bundle
}
```

### Step 3: Implement BundleDataSource (20 min)

**File:** `internal/tui/datasource.go`

```go
func NewBundleDataSource(bundlePath string) (*BundleDataSource, error) {
    opts := bundle.ImportOptions{
        Path: bundlePath,
        MaxSize: 100 * 1024 * 1024, // 100MB for TUI mode
    }
    
    b, err := bundle.Load(opts)
    if err != nil {
        return nil, err
    }
    
    return &BundleDataSource{bundle: b}, nil
}

func (ds *BundleDataSource) GetPods(namespace string) ([]Pod, error) {
    var pods []Pod
    for _, podInfo := range ds.bundle.Pods {
        if namespace == "" || podInfo.Namespace == namespace {
            pods = append(pods, Pod{
                Name:       podInfo.Name,
                Namespace:  podInfo.Namespace,
                Containers: podInfo.Containers,
                Status:     "Bundle", // Special status for bundle pods
            })
        }
    }
    return pods, nil
}

func (ds *BundleDataSource) GetLogs(namespace, pod, container string) ([]string, error) {
    // Find log file for this pod/container
    for _, logFile := range ds.bundle.LogFiles {
        if logFile.Namespace == namespace &&
           logFile.PodName == pod &&
           logFile.ContainerName == container &&
           !logFile.IsPrevious { // Get current logs by default
            
            content, err := ds.bundle.ReadLogFile(&logFile)
            if err != nil {
                return nil, err
            }
            
            // Split into lines
            lines := strings.Split(string(content), "\n")
            return lines, nil
        }
    }
    return nil, fmt.Errorf("log file not found")
}
```

### Step 4: Update App to Use DataSource (15 min)

**File:** `internal/tui/app.go`

```go
type App struct {
    // ... existing fields
    dataSource DataSource  // New field
    bundleMode bool       // New field
}

func NewApp(cfg *config.Config, bundlePath string) (*App, error) {
    app := &App{
        config: cfg,
        // ... initialize other fields
    }
    
    // Determine data source
    if bundlePath != "" {
        ds, err := NewBundleDataSource(bundlePath)
        if err != nil {
            return nil, err
        }
        app.dataSource = ds
        app.bundleMode = true
    } else {
        // Use live Rancher client
        client, err := rancher.NewClient(cfg)
        if err != nil {
            return nil, err
        }
        app.dataSource = NewLiveDataSource(client)
        app.bundleMode = false
    }
    
    return app, nil
}
```

### Step 5: Update Views to Use DataSource (20 min)

**Affected Views:**
- Pod list view
- Log view
- Namespace view

**Changes:**
```go
// Before (direct API call)
pods, err := app.client.GetPods(namespace)

// After (data source abstraction)
pods, err := app.dataSource.GetPods(namespace)
```

### Step 6: Add Bundle Mode Indicator (10 min)

**File:** `internal/tui/app.go`

```go
func (a *App) renderHeader() string {
    mode := "LIVE"
    if a.bundleMode {
        mode = "BUNDLE"
    }
    
    return fmt.Sprintf("[%s] r8s - Rancher Log Viewer", mode)
}
```

### Step 7: Handle Previous Logs (10 min)

Add hotkey to toggle between current and previous logs:

```go
case "ctrl+p": // Toggle previous logs
    a.showPrevious = !a.showPrevious
    a.refreshLogs()
```

---

## File Changes Summary

### New Files (2)
1. `internal/tui/datasource.go` (~200 lines)
   - DataSource interface
   - BundleDataSource implementation
   - LiveDataSource implementation

2. `internal/tui/bundle_helpers.go` (~100 lines)
   - Bundle-specific utility functions
   - Log file parsing helpers

### Modified Files (4)
1. `cmd/root.go` (~10 lines)
   - Add --bundle flag
   - Pass to NewApp()

2. `internal/tui/app.go` (~30 lines)
   - Add dataSource field
   - Add bundleMode field
   - Update NewApp() constructor
   - Add mode indicator to header

3. `internal/tui/views/pods.go` (~15 lines)
   - Use dataSource.GetPods()
   - Handle bundle mode display

4. `internal/tui/views/logs.go` (~20 lines)
   - Use dataSource.GetLogs()
   - Add previous log toggle

**Total New Code:** ~375 lines

---

## Testing Strategy

### Unit Tests
- DataSource interface implementation
- Bundle pod listing
- Bundle log retrieval
- Previous log handling

### Integration Tests
1. **Live Mode (Regression)**
   - Start without --bundle flag
   - Verify all existing features work
   - No changes detected

2. **Bundle Mode (New)**
   - Start with --bundle flag
   - Load test bundle
   - View pod list
   - View logs
   - Test search
   - Test filters
   - Test colors

3. **Mode Indicator**
   - Live mode shows "LIVE"
   - Bundle mode shows "BUNDLE"

4. **Previous Logs**
   - Ctrl+P toggles to previous
   - Error if no previous logs
   - Back to current logs

---

## Success Criteria

- [ ] TUI starts with `--bundle` flag
- [ ] Bundle loads successfully
- [ ] Pod list shows bundle pods
- [ ] Logs display from bundle files
- [ ] All Phase 1-3 features work (search, filters, colors)
- [ ] Previous logs accessible with Ctrl+P
- [ ] Mode indicator shows "BUNDLE"
- [ ] Zero regressions in live mode
- [ ] Clean error messages
- [ ] Performance acceptable (<1s to display logs)

---

## Risks & Mitigations

### Risk 1: Breaking Live Mode
**Mitigation:** 
- Keep existing code paths intact
- Use abstraction layer
- Comprehensive regression testing

### Risk 2: Bundle Log Format Differences
**Mitigation:**
- Log files are plain text (same format)
- Reuse existing parser
- Test with real bundle

### Risk 3: Missing Previous Logs
**Mitigation:**
- Check if file exists before reading
- Show clear error message
- Graceful fallback

### Risk 4: Large Log Files
**Mitigation:**
- Stream reading (existing pattern)
- Truncate if needed
- Show file size warnings

---

## User Flow Examples

### Example 1: View Bundle Logs

```bash
# Import and view in one step
./bin/r8s --bundle=bundle.tar.gz

# TUI starts in bundle mode
# Navigate: Namespaces → Pods → Logs
# Use existing hotkeys: /, n, Ctrl+E, etc.
```

### Example 2: Compare Current vs Previous

```bash
./bin/r8s --bundle=bundle.tar.gz

# Navigate to crashed pod logs
# Press Ctrl+P to see previous logs (before crash)
# Press Ctrl+P again to go back to current
```

### Example 3: Search Across Bundle

```bash
./bin/r8s --bundle=bundle.tar.gz

# Navigate to pod
# Press / to search
# Pattern highlights work same as live mode
```

---

## Performance Targets

- Bundle load: <5 seconds for 100MB bundle
- Pod list display: <100ms
- Log display: <500ms for 10K lines
- Search: <200ms for 10K lines
- Filter toggle: <100ms
- No memory leaks

---

## Documentation Updates

### User Documentation
1. Update README.md with bundle mode usage
2. Add --bundle flag to help text
3. Create bundle mode guide

### Developer Documentation
1. Document DataSource interface
2. Add architecture diagram
3. Update testing guide

---

## Phase Breakdown

### Part A: Data Source Abstraction (30 min)
- Create interface
- Implement BundleDataSource
- Unit tests

### Part B: TUI Integration (35 min)
- Add bundle flag
- Update App initialization
- Update views to use DataSource
- Integration testing

### Part C: Previous Logs Feature (15 min)
- Add Ctrl+P hotkey
- Implement toggle logic
- Test with bundle

### Part D: Polish & Testing (10 min)
- Mode indicator
- Error messages
- Final testing
- Documentation

**Total:** 90 minutes

---

## Next Steps

1. Review this plan
2. Confirm approach
3. Begin implementation with Part A
4. Iterate through parts
5. Test comprehensively
6. Document and commit

---

**Plan Status:** Ready for Approval  
**Estimated Effort:** 60-90 minutes  
**Risk Level:** 🟡 Medium (new integration, but well-scoped)  
**Value:** 🔴 High (unlocks offline diagnostics)
