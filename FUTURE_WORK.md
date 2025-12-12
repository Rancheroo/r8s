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

## 🎯 High Priority (Next Release - v0.5.0)

### Dashboard Log Scanning (REMOVED in v0.4.3)
- **Priority**: HIGH - Was displaying false information
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

## 🚀 Long-Term Ideas (v0.5.0+)

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

## 📝 Documentation Enhancements

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

### Test Coverage
- Increase unit test coverage to 80%+
- Add integration tests for bundle loading
- Performance benchmarks for large bundles

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
