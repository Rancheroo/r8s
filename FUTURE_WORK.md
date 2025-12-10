# r8s Future Work & Deferred Features

This document tracks feature ideas and enhancements that have been identified but deferred to future releases.

## ✅ Recently Completed

### Tunable Scan Depth (v0.3.9)
- **Status**: ✅ Shipped in v0.3.9
- **Description**: User-controllable scan depth via `--scan` flag
- **Impact**: Users can now tune speed/accuracy trade-off based on bundle size
- **Usage**: `r8s --scan 500 ./bundle/` (default: 200 lines)

## 🎯 High Priority (Next Release - v0.4.0)

### Smart Sorting by Error Count
- **Priority**: Medium
- **Complexity**: Medium
- **Impact**: Medium
- **Description**: Auto-sort Attention Dashboard items by error/warning count for faster triage
- **Requirements**:
  - Refactor dashboard rendering to support dynamic sorting
  - Add sort toggle (by severity vs. by count)
  - Persist sort preference in config

### Hotkeys for Quick Navigation
- **Priority**: Medium  
- **Complexity**: Low
- **Impact**: High
- **Description**: Global hotkeys to jump directly to highest error/warn pod
- **Requirements**:
  - `e` hotkey: Jump to pod with most errors
  - `w` hotkey: Jump to pod with most warnings
  - Works from any view (global binding)
  - Visual indicator showing which pod would be selected

## 📋 Medium Priority (v0.4.0)

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
