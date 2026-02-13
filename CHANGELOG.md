# Changelog

All notable changes to r8s will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.7.0] - 2026-02-13

### Added
- **Bundle Completeness Indicator** (#34) - Shows % of expected files found in support bundles
- **dmesg OOM Detection Parser** (#40) - Detects Out-of-Memory kills from kernel logs
- **PV/PVC/StatefulSet Parsers** (#39) - Full persistent volume analysis
- **ConfigMaps Parser** (#42) - Extract and analyze ConfigMap data
- **HelmCharts Parser** (#42) - Identify Helm releases and chart versions
- **RKE2 Journald Parser** (#41) - Control plane log analysis with critical error detection
- **test-cluster Command** - Automated cluster diagnostics subcommand
- **CodeRabbit Integration** - Automatic team welcome on PR reviews

### Fixed
- Exit codes, JSON output, and unbound PV handling (#37, #38, #49)
- Corrected ParsePVs column mapping (#48)

### Stats
- +2,216/-55 lines changed across 19 files
- All new parsers: 100% test coverage
- No breaking changes

## [0.6.9] - 2026-01-15

### Fixed
- **CRITICAL**: O(n²) bubble sort replaced with O(n log n) stdlib sort in event sorting (Principle #8: O(n log n) Always)
  - `GetEventsByPod()` now uses `sort.Slice` for optimal performance
  - Prevents performance degradation with large event lists
  - Maintains correct sort order: Warnings first, then by LastSeen descending

### Changed
- **Truth Only™**: Removed mock data fallbacks (Principle #1: Truth Only™)
  - `DescribeDeployment()` now returns error instead of fabricated data when deployment not found
  - `DescribeService()` now returns error instead of fabricated data when service not found  
  - `GetContainers()` now returns empty slice instead of "default" fallback when containers unknown
  - Better to show nothing than wrong data

### Added
- Regression tests for principle compliance (`internal/datasource/bundle_regression_test.go`)
  - Prevents future violations of core development principles
  - Documents expected behavior for Truth Only™, Empty is Valid, and O(n log n) Always

### Developer Notes
- This release achieves 98% principle compliance (up from 76%)
- All critical principle violations resolved
- Foundation for v0.7.0 CI/CD and testing infrastructure

## [0.6.8.1] - 2026-01-14

### Fixed
- **CRITICAL**: Node/etcd drill-down feature now fully functional
  - Added missing message handler for `clusterEventMsg` in Update() method
  - Implemented pod selection function `handleClusterEventPodSelection()`
  - Number keys (1-9) now work correctly in drill-down views
  - Fixed ResourceType check to support event, node, and etcd types
  - GetAllPods error handling improved
  
### Added
- Node name display in drill-down breadcrumb for node events
  - Shows "Node Event: MemoryPressure on w-guard-wg-wk-pfvjr-4x..." instead of generic "Cluster Event"
  - Provides better context when investigating node-specific issues

### Changed
- Increased description column width from 25 to 50 characters in attention dashboard
  - Better visibility for long node pressure descriptions
  - Follows "Truth Only™" principle - show full information

### Known Issues
- Emoji alignment in attention dashboard may vary across terminal emulators (see FUTURE_WORK.md)
  - Attempted fix using runewidth library + manual padding
  - Terminal/font dependent - cosmetic only, functionality unaffected

## [0.6.8] - 2026-01-XX

### Added
- Event message truncation with smart truncation logic
- Cluster-level drill-down for node and etcd events

## [0.6.7] - 2026-01-XX

### Added
- Inline diagnostics in attention dashboard
- Diagnostic context for pod issues (root cause + recommendations)

### Fixed
- Navigation bug where kubelet/etcd/node items navigated to wrong pods
- Pod navigation from attention dashboard now works correctly

---

## Version History

- **0.6.8.1** - Hotfix: Node drill-down complete + UX improvements
- **0.6.8** - Event truncation + cluster drill-down
- **0.6.7** - Inline diagnostics + navigation fixes
- **0.6.x** - Attention dashboard enhancements
- **0.5.x** - Core TUI functionality
- **0.4.x** - Bundle support
- **0.3.x** - Initial release

[Unreleased]: https://github.com/Rancheroo/r8s/compare/v0.7.0...HEAD
[0.7.0]: https://github.com/Rancheroo/r8s/compare/v0.6.9...v0.7.0
[0.6.8.1]: https://github.com/Rancheroo/r8s/compare/v0.6.8...v0.6.8.1
[0.6.8]: https://github.com/Rancheroo/r8s/compare/v0.6.7...v0.6.8
[0.6.7]: https://github.com/Rancheroo/r8s/releases/tag/v0.6.7
