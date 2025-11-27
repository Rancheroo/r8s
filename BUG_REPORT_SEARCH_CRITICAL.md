# CRITICAL BUG REPORT: Search Functionality

**Date**: 2025-11-27  
**Severity**: CRITICAL - Feature Unusable  
**Status**: CONFIRMED  
**Build**: dev (8f17801)

---

## Executive Summary

The search functionality in the logs view has multiple critical bugs that make it completely unusable. The feature doesn't crash the application but becomes unresponsive and exhibits incorrect behavior.

---

## Critical Bugs Identified

### Bug 1: Search Input Handler Lacks View Context Check
**Location**: `internal/tui/app.go` lines 362-386

**Issue**: The search mode input handler processes keystrokes without checking if we're in `ViewLogs`. This means search mode can capture input even when it shouldn't.

**Expected**: Input should only be processed when `a.currentView.viewType == ViewLogs`

**Actual**: Input is processed regardless of view type

**Code**:
```go
// Handle search input when in search mode
if a.searchMode {
    // Missing check: && a.currentView.viewType == ViewLogs
    switch msg.String() {
```

---

### Bug 2: Search Mode Not Properly Updating Filtered Logs

**Location**: `internal/tui/app.go` line 2495-2497

**Issue**: `performSearch()` searches through `a.logs` (the original log array) but doesn't account for filtered logs when a filter is active.

**Expected**: Search should operate on currently visible logs (filtered if filter is active)

**Actual**: Searches all logs, including filtered-out entries

**Impact**: Search matches reference line numbers that don't exist in the filtered view, causing viewport navigation to fail

---

### Bug 3: Line Count in Status Bar Shows Original Count Not Filtered

**Location**: `internal/tui/app.go` line 1118

**Issue**: Status bar always shows `len(a.logs)` even when filters are active

**Expected**: Show count of visible/filtered logs

**Actual**: Shows total count, misleading when filters reduce visible lines

**Code**:
```go
case ViewLogs:
    count := len(a.logs)  // Wrong - should be len(filtered logs)
```

---

### Bug 4: Search Doesn't Update Viewport Content

**Location**: `internal/tui/app.go` line 2486-2511

**Issue**: After `performSearch()` executes, the viewport content is not updated with highlighting or any visual indication of matches

**Expected**: Matched lines should be highlighted or marked in some way

**Actual**: No visual feedback, just viewport navigation

---

### Bug 5: Missing Search State Cleanup on View Exit

**Location**: No cleanup in escape/back navigation

**Issue**: When exiting logs view with Esc, search state (`searchMode`, `searchQuery`, `searchMatches`) persists

**Expected**: Clean search state when leaving logs view

**Actual**: Search query and state persist, causing confusion on re-entry

**Impact**: Old search terms appear when re-entering logs view

---

### Bug 6: Viewport Navigation Uses Filtered Log Indices on Unfiltered Data

**Location**: `internal/tui/app.go` lines 341-343, 354-356

**Issue**: Navigation to search matches uses line indices from original logs array, but viewport may be showing filtered content

**Expected**: Match indices should correspond to visible viewport lines

**Actual**: Indices are from original logs, causing incorrect navigation when filters are active

**Code**:
```go
for i := 0; i < a.searchMatches[a.currentMatch]; i++ {
    a.logViewport.LineDown(1)
}
// ^ This assumes viewport has all logs, but it may have filtered subset
```

---

## Reproduction Steps

1. Build and run `./bin/r8s`
2. Navigate to: Clusters → Projects → Namespaces → Pods → Select pod → Press 'l' for logs
3. Press '/' to enter search mode
4. Type "ERROR"
5. **OBSERVE**: Only first character appears ("E_")
6. Press Enter to execute search
7. **OBSERVE**: Search doesn't execute visibly, viewport doesn't move
8. Press Esc to cancel
9. **OBSERVE**: Wrong behavior - exits logs view instead of just canceling search
10. Re-enter logs
11. **OBSERVE**: Previous search term still there ("E_")

---

## Root Causes

1. **Insufficient State Management**: Search state not properly scoped to ViewLogs
2. **Missing Filter Integration**: Search doesn't integrate with log filtering system
3. **No Visual Feedback**: No highlighting or indication of matches
4. **State Persistence**: Search state not cleaned up on view exit
5. **Index Mismatch**: Search indices don't account for filtered content

---

## Impact Assessment

**User Experience**: ⚠️ **SEVERE**
- Feature appears broken
- Confusing behavior (Esc exits view)
- No visual feedback
- Persistent stale state

**Functionality**: ⚠️ **CRITICAL**
- Search completely unusable
- Combines with filter system incorrectly
- Navigation broken when filters active

**Data Integrity**: ✅ **NO IMPACT**
- No data corruption or loss
- Application remains stable

---

## Recommended Fixes

### Fix 1: Add View Context Check to Search Input Handler

```go
// Handle search input when in search mode
if a.searchMode && a.currentView.viewType == ViewLogs {
    switch msg.String() {
    case "esc":
        a.searchMode = false
        a.searchQuery = ""
        a.searchMatches = nil
        a.currentMatch = -1
        return a, nil
    // ... rest of handler
}
```

### Fix 2: Search Should Use Filtered Logs

```go
func (a *App) performSearch() {
    if a.searchQuery == "" {
        return
    }

    // Clear previous matches
    a.searchMatches = nil
    a.currentMatch = -1

    // Get the current visible logs (respecting filters)
    visibleLogs := a.getVisibleLogs() // New helper function

    // Search through visible logs only
    query := strings.ToLower(a.searchQuery)
    for i, line := range visibleLogs {
        if strings.Contains(strings.ToLower(line), query) {
            a.searchMatches = append(a.searchMatches, i)
        }
    }

    // Jump to first match
    if len(a.searchMatches) > 0 {
        a.currentMatch = 0
        a.logViewport.GotoTop()
        for i := 0; i < a.searchMatches[0]; i++ {
            a.logViewport.LineDown(1)
        }
    }
}
```

### Fix 3: Add Helper to Get Visible Logs

```go
func (a *App) getVisibleLogs() []string {
    if a.filterLevel == "" {
        return a.logs
    }

    var filteredLogs []string
    for _, line := range a.logs {
        lineUpper := strings.ToUpper(line)
        switch a.filterLevel {
        case "ERROR":
            if strings.Contains(lineUpper, "[ERROR]") {
                filteredLogs = append(filteredLogs, line)
            }
        case "WARN":
            if strings.Contains(lineUpper, "[WARN]") || strings.Contains(lineUpper, "[ERROR]") {
                filteredLogs = append(filteredLogs, line)
            }
        }
    }
    return filteredLogs
}
```

### Fix 4: Update Status Bar to Show Visible Log Count

```go
case ViewLogs:
    visibleLogs := a.getVisibleLogs()
    count := len(visibleLogs)  // Use visible count, not total
    // Build dynamic status based on active features
    parts := []string{fmt.Sprintf("%d lines", count)}
    // ... rest of status building
```

### Fix 5: Clean Search State on View Exit

```go
case "esc":
    if a.showingDescribe {
        // ... existing describe handling
    } else if a.searchMode {
        // Exit search mode (new check)
        a.searchMode = false
        a.searchQuery = ""
        a.searchMatches = nil
        a.currentMatch = -1
        return a, nil
    } else if len(a.viewStack) > 0 {
        // Pop view from stack
        // ALSO clean search state here:
        a.searchMode = false
        a.searchQuery = ""
        a.searchMatches = nil
        a.currentMatch = -1
        
        // ... rest of existing code
    }
```

---

## Testing After Fixes

1. **Basic Search**: Enter search, type full query, verify execution
2. **Search + Filter**: Enable ERROR filter, then search - verify indices correct
3. **Status Bar**: Verify line count shows filtered count
4. **View Exit**: Exit logs view, re-enter, verify clean state
5. **Escape Behavior**: 
   - In search mode: Esc cancels search
   - Not in search mode: Esc exits view
6. **Navigation**: 'n' and 'N' work correctly with filtered logs

---

## Priority

**P0 - CRITICAL**: This completely breaks the search feature. Must fix before any release.

**Recommended Action**: 
1. Implement all 5 fixes immediately
2. Test thoroughly with the test plan
3. Update TEST_REPORT to mark search tests as FAILED
4. Re-test and update status to PASS once fixed

---

## Related Issues

- Filters work correctly in isolation (PASS)
- Tail mode works correctly (PASS)
- Container cycling works correctly (PASS)
- Search is the only broken feature

---

**Status Update Needed**: TEST_REPORT_PHASE2_STEPS345.md must be updated to reflect:
- Test 7: FAIL (search execution broken)
- Test 12: FAIL (search+filter integration broken)
- Overall status: PARTIAL PASS → FAIL (critical feature broken)

---

**Discovered By**: Interactive testing session  
**Reported By**: AI Agent (Warp)  
**Assigned To**: Development team  
**Target Fix**: ASAP
