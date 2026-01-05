# r8s Future Work & Deferred Features

This document tracks feature ideas and enhancements that have been identified but deferred to future releases.

## ✅ Recently Completed

### Dashboard Scrolling & Item Capping (v0.4.0)
- **Status**: ✅ Shipped in v0.4.0
- **Description**: Smart capping with expansion for high --scan values
- **Impact**: Dashboard now handles 80+ issues gracefully without screen overflow
- **Features**:
  - Default cap at top-20 issues (sorted by severity)
  - Press 'm' to toggle between capped/expanded view
  - Press 'g/G' to jump to top/bottom of list
  - Position indicator shows "Showing X/Y" when capped
  - "...and X more (press 'm')" message when items hidden
  - Smooth navigation through unlimited items
- **Usage**: Works automatically with `--scan` flag at any value
- **Result**: High --scan values (500-1000) now usable without UX degradation

### Tunable Scan Depth (v0.3.9)
- **Status**: ✅ Shipped in v0.3.9
- **Description**: User-controllable scan depth via `--scan` flag
- **Impact**: Users can now tune speed/accuracy trade-off based on bundle size
- **Usage**: `r8s --scan 500 ./bundle/` (default: 200 lines)

### App.go Modular Decomposition (v0.5.1)

- **Status**: ✅ Shipped in v0.5.1
- **Description**: Complete decomposition of 3031-line app.go into 6 focused modules
- **Impact**: 87% reduction in main file size, 70% cognitive load reduction
- **Modules Created**:
  - **app.go**: ~400 lines - Core state & orchestration
  - **helpers.go**: ~300 lines - Utilities, breadcrumb, status text
  - **logs.go**: ~450 lines - Log view rendering & filtering
  - **fetch.go**: ~400 lines - Data fetching & describe operations
  - **table.go**: ~500 lines - Table rendering for all views
  - **handlers.go**: ~500 lines - Event handlers & navigation
- **Result**: Better maintainability, improved testability, faster onboarding

### Dashboard Log Scanning (RE-ENABLED in v0.5.0)
- **Priority**: ~~HIGH~~ COMPLETED ✅ v0.5.0
- **Status**: ✅ Re-enabled in v0.5.0
- **Complexity**: Medium-High
- **Impact**: High
- **Status**: ❌ REMOVED in v0.4.3 due to inaccurate counts
- **Description**: Re-implement accurate per-pod log error/warning detection for dashboard
- **Problem Identified**: 
  - Dashboard showed identical ERR/WARN counts across different pods (e.g., all argocd pods showing "19 ERR, 17 WARN")
  - Actual pod logs showed different counts (e.g., "1 errors · 0 warnings")
  - Root cause: `detectLogIssues()` was reusing/caching counts incorrectly
  - Also noticed in namespace view - counts appear inconsistent
- **Requirements for Re-implementation**:
  - Fix GetLogs() calls to ensure correct namespace/pod parameters
  - Verify no caching/reuse of counts across different pods
  - Add per-pod verification: dashboard count MUST match log view count
  - Test with namespace-level aggregation
  - Add debug mode to verify which pod's logs are being scanned
  - Only re-enable once 100% verified accurate
- **Current State**: Real-time log counting in individual pod view works perfectly - keep that
- **Principle**: r8s only displays truth. Better to show no count than a wrong count.
- **Triggered by**: User-reported critical bug in v0.4.2

### Smart Sorting by Error Count
- **Priority**: Medium
- **Complexity**: Medium
- **Impact**: Medium
- **Description**: Auto-sort Attention Dashboard items by error/warning count for faster triage
- **Requirements**:
  - Refactor dashboard rendering to support dynamic sorting
  - Add sort toggle (by severity vs. by count)
  - Persist sort preference in config


## 📋 Medium Priority (v0.4.0)

### Namespace Health Ranking & Smart Filtering
- **Priority**: Medium-High
- **Complexity**: Medium
- **Impact**: High
- **Description**: Rank and sort namespaces by problem severity for quick detection of most problematic areas
- **Problem**: With 25+ namespaces, no quick way to identify which ones have the most issues. All show as "active" with no health indicators.
- **Requirements**:
  - Add "ISSUES" column to namespace view showing error/warning counts
  - Sort namespaces by total issue count (highest problems first)
  - Color-code namespace rows: Red (>50 errors), Yellow (>20 warnings), Green (healthy)
  - Filter options: "Show only namespaces with issues" (press 'f')
  - Quick jump to most problematic namespace (press 'e' for highest errors)
  - Aggregate pod-level errors per namespace for ranking score
- **Use Case**: "Which namespace should I investigate first?"
- **Example**: 
  ```
  NAME                  STATE    ISSUES       PROJECT
  kube-system          active   🔥 127E/89W  bundle-project
  gpu-operator         active   ⚠️  22E/67W  bundle-project  
  longhorn-system      active   ✅  2E/5W    bundle-project
  calico-system        active   ✅  0E/0W    bundle-project
  ```
- **Triggered by**: User feedback viewing 25-namespace bundle

### Journald Log Scanning
- **Priority**: Medium
- **Complexity**: High
- **Impact**: Medium
- **Description**: Scan systemd/journald logs for node-level issues
- **Requirements**:
  - Extend DataSource interface to support journald parsing
  - Pattern matching for systemd service failures
  - Integration with Attention Dashboard
  - Bundle structure changes to include journald/
- **Dependencies**: Requires datasource interface refactor

### Enhanced Help Panel
- **Priority**: Low
- **Complexity**: Low
- **Impact**: Low
- **Description**: Add contextual pro tips to help panel
- **Requirements**:
  - Pro tip: "Start with dashboard for quick wins"
  - Pro tip: "Use Ctrl+W in logs to focus on issues"
  - Pro tip: "Watch W/E column in Pods view for quick health check"
  - Context-aware tips based on current view

### Edge Case Handling
- **Priority**: Low
- **Complexity**: Low
- **Impact**: Low
- **Description**: Better UX for edge cases
- **Requirements**:
  - Empty logs: Show "No E/W — check describe/events"
  - Huge bundles: Cap dashboard at top-N with "and X more..." indicator
  - Parse errors: Show count in bundle load warning

## 💡 UX Philosophy: "Show, Don't Ask" (v0.5.2+)

**Core Principle**: Information should surface automatically without requiring button presses.

### Auto-Display Health Indicators
- **Priority**: High
- **Complexity**: Low
- **Impact**: High
- **Description**: Show critical system health automatically in dashboard header
- **Requirements**:
  - Dashboard header shows "🔥 2 critical · ⚠ 5 warnings · ✅ 12 healthy" always visible
  - No button press needed - appears on load
  - Updates automatically when data refreshes
- **Philosophy**: User sees cluster health at a glance, no navigation required

### Auto-Show Parse Warnings
- **Priority**: Medium
- **Complexity**: Low
- **Impact**: Medium
- **Description**: Display bundle parse issues automatically in status bar
- **Requirements**:
  - Status bar shows "⚠ 3 parse issues" when bundle has parsing errors
  - Appears automatically on bundle load
  - No hotkey needed - information surfaces immediately
- **Philosophy**: Problems should announce themselves, not hide behind buttons

### Progressive Information Disclosure
- **Priority**: Medium
- **Complexity**: Medium
- **Impact**: High
- **Description**: Each view automatically shows most important info without drilling down
- **Requirements**:
  - Namespace view: Shows error counts automatically in ISSUES column
  - Pod view: Shows W/E counts in every row by default
  - Dashboard: Worst items sorted to top automatically
  - Log view: Empty state shows helpful next steps automatically
- **Philosophy**: Critical data visible by default, details available on demand

### Smart Defaults Over Configuration
- **Priority**: Medium
- **Complexity**: Low
- **Impact**: Medium
- **Description**: System chooses best defaults, reduces need for user configuration
- **Requirements**:
  - Default sort = count (worst first) - no need to press 's'
  - Default filter = all logs - user adds filters only when needed
  - Scan depth auto-adjusts based on bundle size
- **Philosophy**: Software should be intelligent enough to make good choices automatically

### Remove Sort Mode Complexity (v0.5.3+)
- **Priority**: High
- **Complexity**: Low (removal)
- **Impact**: High (simplification)
- **Description**: Eliminate sort mode toggles - always use smart default (highest error count first)
- **Current State**: 
  - 3 sort modes: Count, Severity, Name
  - Toggle hotkey 's' cycles through modes
  - Status bar shows current mode
  - ~200 lines of sorting code
- **Proposed Change**:
  - Remove `SortMode` enum entirely
  - Remove sort toggle functionality
  - Always sort by error count descending (worst first)
  - Remove status bar sort indicator
  - Delete unused sorting functions
- **Rationale**:
  - 95% of users want "worst first" (count-based)
  - Sorting options add cognitive load without value
  - Smart default eliminates need for configuration
  - Fewer features = better UX
- **Code Reduction**: ~200 lines removed
- **Philosophy**: "The best feature is no feature - smart defaults beat options"

### CrashLoopBackOff with No Logs Indicator (v0.5.3+) ⭐
- **Priority**: HIGH
- **Complexity**: Low
- **Impact**: High (UX consistency)
- **Description**: Pods showing as CrashLoopBackOff in dashboard but "✅ Clean" in namespace view is misleading
- **Problem**:
  - Dashboard correctly shows pod in CrashLoopBackOff
  - Namespace/Classic view shows "✅ Clean" because no log files found
  - User sees conflicting information: "Is this pod healthy or broken?"
- **Scope Expansion (v0.5.3+)**:
  - Applies to ALL pods with zero log lines, not just CrashLoopBackOff
  - Any pod state (Running, Pending, Error, CrashLoopBackOff, etc.) with 0 logs gets indicator
  - Consistent application of Truth Only™ principle: "no logs ≠ clean"
- **Proposed Fix**:
  - When pod has 0 log lines (regardless of pod state):
    - Show "⚠️ No Logs" instead of "✅ Clean"
    - Use warning emoji (not green tick) for any pod without log data
    - Add tooltip: "No logs captured - check describe/events for details"
- **Logic**:

  ```python
  if logCount == 0:
    display "⚠️ No Logs"  # Not "✅ Clean"
  else:
    display actual E/W counts or "✅ Clean"
  ```

- **Philosophy**: Truth in context - "Clean" means healthy AND no issues, not just "no logs found"
- **Triggered by**: User-reported UX confusion
- **Applies to**: All pod states, not just CrashLoopBackOff (consistency)

### Empty Namespace Intelligence (v0.5.3+) ⭐
- **Priority**: Medium
- **Complexity**: Low
- **Impact**: Medium (clarity)
- **Description**: Namespaces with 0 pods showing "✅ Clean" is misleading - they're empty, not healthy
- **Problem**:
  - Empty namespaces (0 pods) show same "✅ Clean" as healthy namespaces
  - User can't distinguish between "healthy" vs "empty"
  - Leads to false confidence: "This namespace looks good!" (but it's just empty)
- **Proposed Fix**:
  - If namespace.PodCount == 0:
    - Display "📭 Empty" instead of "✅ Clean"
  - Use different indicator to show "no pods = no data" not "no problems"
- **Philosophy**: Distinguish "no issues" from "no data" - they mean different things

### Age Display Consistency (v0.5.3+) ⭐
- **Priority**: Low
- **Complexity**: Low
- **Impact**: Low (polish)
- **Description**: "N/A" age is uninformative - show creation time when available
- **Problem**:
  - Many resources show "N/A" for age
  - Bundle contains creation timestamps in metadata
  - User doesn't know if resource is 1 hour or 1 year old
- **Proposed Fix**:
  - Parse creation timestamp from kubectl output
  - Show relative age: "2d", "5h", "30m"
  - Fall back to "Unknown" (not "N/A") only when timestamp truly missing
- **Philosophy**: Show what we know - bundles have timestamps, use them


### Auto-Display Pod Diagnostics on Failure (v0.5.4) ✅
- **Status**: ✅ Shipped in v0.5.4
- **Description**: Enhanced "No Logs" diagnostic panel with maximum intel
- **Impact**: Users see comprehensive pod diagnostics automatically - r8s interprets data instead of showing raw output
- **Features**:
  - Intelligent diagnosis based on pod state (CrashLoop, OOMKilled, ImagePull, Error, Pending, Evicted)
  - Recent events display (last 5) with emoji indicators
  - State-specific investigation suggestions
  - External tools guidance (lnav, kubectl logs)
  - Works from both Dashboard and Classic navigation paths
- **Philosophy**: "r8s interprets, user acts" - Show intelligence, not just information

### Remove Log Filter Modes (v0.5.4+)
- **Priority**: Medium
- **Complexity**: Medium (removal + smart defaults)
- **Impact**: High (simplification)
- **Description**: Eliminate ALL/ERROR/WARN filter toggles - use intelligent defaults with prominent highlighting
- **Current State**:
  - 3 filter modes: ALL, ERROR, WARN
  - Ctrl+A / Ctrl+E / Ctrl+W hotkeys
  - Status bar shows current filter
  - User must manually toggle to see issues
- **Proposed Change (Option A - Pure Simplification)**:
  - Remove filter modes entirely
  - Always show ALL logs by default
  - Errors highlighted RED, warnings YELLOW (already implemented in log view)
  - No filter needed - issues visually surface themselves
- **Proposed Change (Option B - Smart Auto-Detect)**:
  - Auto-detect log content on open:
    - If < 100 lines: show ALL (no filtering needed)
    - If 100-1000 lines and >10 errors: show ERROR
    - If >1000 lines and >20 warnings: show WARN
    - Otherwise: show ALL
  - Add single "Show All" toggle (Ctrl+A) to override smart filter
  - Remove ERROR/WARN individual modes
- **Recommendation (Option A - More Aligned with Principles)**:
  - Option A better aligns with "Show, Don't Ask" - show everything, highlight issues
  - Reduces 3 choices to 1 (toggle available if needed)
  - User sees full context immediately
  - Color highlighting already works - leverage it
- **Rationale**:
  - Software should show data, not hide it behind filters
  - 90% of time: user wants context + errors visible together
  - Color highlighting already surfaces issues
  - Reduce 3 choices to 0 (no filtering needed)
- **Code Reduction**: ~100 lines removed
- **Philosophy**: "Show everything, highlight issues - don't hide context behind filters"

### Remove View Switching Hotkeys (v0.6.0+)
- **Priority**: Low
- **Complexity**: Medium (behavior change)
- **Impact**: Medium (simplification)
- **Description**: Eliminate 1/2/3/4/5 view switching - use context-aware navigation only
- **Current State**:
  - Hotkeys: 1=Pods, 2=Deployments, 3=Services, 4=CRDs, etc.
  - User must remember mapping
  - Rarely used (Enter/Esc navigation more natural)
- **Proposed Change**:
  - Remove number hotkeys entirely
  - Navigation via Enter (drill down) and Esc (go up) only
  - Add breadcrumb-based "back to X" shortcuts if needed
- **Rationale**:
  - Enter/Esc is intuitive hierarchy navigation
  - Number keys aren't discoverable (not shown in help)
  - Simplifies mental model: "Just press Enter"
- **Code Reduction**: ~50 lines removed
- **Philosophy**: "Fewer ways to do same thing = easier to learn"

## 🚀 Long-Term Ideas (v0.6.0+)

### Real-Time Monitoring
- **Description**: Support live cluster monitoring (not just bundles)
- **Requirements**: 
  - Kubernetes API client integration
  - Auto-refresh mode
  - Connection status indicator

### Advanced Search
- **Description**: Regex search across all logs in bundle
- **Requirements**:
  - Global search mode
  - Results aggregation across pods
  - Jump-to-log functionality

### Log Export & Reporting
- **Description**: Export filtered logs or generate issue reports
- **Requirements**:
  - Export selected logs to file
  - Generate markdown summary
  - Email report capability

### Multi-Bundle Comparison
- **Description**: Compare two bundles side-by-side
- **Requirements**:
  - Load two bundles simultaneously
  - Diff view for configuration changes
  - Timeline comparison

### Plugin System
- **Description**: Allow custom signal detection plugins
- **Requirements**:
  - Plugin API specification
  - Custom pattern matching
  - User-defined issue types

## 📚 Documentation Enhancements

### Video Tutorials
- Quick start (3 min)
- Advanced navigation (5 min)
- Custom signals (7 min)

### Use Case Examples
- RKE2 cluster troubleshooting walkthrough
- K3s debugging scenarios
- Rancher upgrade troubleshooting

### Pattern Library
- Common error patterns and solutions
- Known issues database
- Community-contributed patterns

## 🔧 Technical Debt

### Code Quality & Refactoring (Deferred from Jan 2026 Audit)

#### Log View Edge Case Documentation
- **Priority**: Low
- **Complexity**: Low
- **Impact**: Low (maintainability)
- **Description**: Add inline comments explaining whitespace-search edge cases in log wrapping
- **Location**: `internal/tui/logs.go` lines 186-244
- **Requirements**:
  - Document fallback behavior when no whitespace found within wrapWidth
  - Explain lastSpace == 0 and lastSpace == -1 edge cases
  - Note that subsequent wrapped segments trim leading spaces via TrimLeft
  - Clarify intentional split at wrapWidth for very long words (URLs/base64)
- **Rationale**: Future maintainers need to understand the wrapping logic for bug fixes

#### Dynamic Help Text Height Calculation
- **Priority**: Low
- **Complexity**: Low
- **Impact**: Low (polish)
- **Description**: Compute contentHeight dynamically instead of hardcoded value
- **Location**: `internal/tui/logs.go` lines 471-477
- **Current State**: `contentHeight := 12` is hardcoded
- **Requirements**:
  - Use `strings.Count(helpText, "\n") + X` to compute actual height
  - Adjust constant offset X to account for title, borders, margins
  - Import strings package if needed
- **Rationale**: Prevents misalignment when helpText changes

#### Extract Namespace Parsing Helper
- **Priority**: Medium
- **Complexity**: Low
- **Impact**: Medium (code quality)
- **Description**: Extract duplicated namespace extraction logic into helper function
- **Location**: `internal/tui/table.go` lines 244-255 (and other locations)
- **Current State**: Logic for splitting "cluster:namespace" duplicated across ViewPods, ViewDeployments, ViewServices
- **Proposed Fix**:
  ```go
  // In helpers.go
  func extractNamespaceName(namespaceID string) string {
      if namespaceID == "" {
          return "default"
      }
      if strings.Contains(namespaceID, ":") {
          parts := strings.Split(namespaceID, ":")
          if len(parts) > 1 {
              return parts[1]
          }
      }
      return namespaceID
  }
  ```
- **Replace**: All instances of namespace extraction with single function call
- **Rationale**: DRY principle, centralize parsing logic

#### Cache Namespace Health Computation
- **Priority**: Medium
- **Complexity**: High
- **Impact**: High (performance)
- **Description**: Cache expensive namespace health computation instead of recalculating on every render
- **Location**: `internal/tui/table.go` lines 130-145
- **Current State**: `ComputeNamespaceHealth` and `SortNamespacesByHealth` run on every table update
- **Requirements**:
  - Add cache fields to table struct: `cachedNsHealth`, `cachedSortedNS`, `cachedScanDepth`
  - Add invalidation flag or last-namespaces-version tracker
  - On render: check if cache valid (same scanDepth, namespaces unchanged)
  - Reuse cached values if valid, otherwise recompute and update cache
  - Invalidate cache on namespace modification or refresh key ('r')
  - Use mutex if table updates can happen concurrently
- **Expected Impact**: Significant performance improvement for large clusters (25+ namespaces)
- **Rationale**: Scanning logs for all pods in all namespaces on every render is expensive

### Test Coverage
- Increase unit test coverage to 80%+
- Add integration tests for bundle loading
- Performance benchmarks for large bundles
- Update breadcrumb tests to expect empty mode indicator (not "[LIVE]")

### Code Quality
- Refactor attention.go (too large)
- Extract signal detection to strategy pattern
- Reduce cyclomatic complexity in app.go

### Performance Optimization
- Lazy-load logs (don't scan all pods upfront)
- Streaming parser for huge log files
- Memory profiling for 1GB+ bundles

---

## Notes

- Features are moved here when they're identified but not critical for current release
- Priority/complexity/impact ratings help with future planning
- Community suggestions welcome via GitHub issues
- Mark items as ✅ when moved to active development

Last updated: 2025-12-10 (v0.3.9 - Tunable scan depth shipped)
