# Changelog

All notable changes to the r8s project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.3-final] - 2025-12-04

### Fixed
- **CRITICAL: Dashboard keyboard navigation completely restored**
  - ↑/↓ or j/k moves focus with visible cyan highlight (inverted row)
  - 1-9 instant jump to issue lines
  - Enter drills down to pod logs (showing ALL logs by default)
  - → or l expands collapsed event lines showing affected pods
  - ← or h collapses expanded items or exits sub-navigation
  - c = classic cluster view, r = refresh, q = quit
  - All keys work instantly with no input delay
- **CRITICAL: Launch errors now display cleanly in terminal**
  - Invalid bundle paths print helpful error messages and exit immediately
  - No more "Press Esc" messages when TUI can't start
  - Clean CLI error handling with proper exit codes
  - Eliminates confusing "could not open TTY" errors for initialization failures

### Changed
- **Dashboard is now BUNDLE-ONLY** (architectural decision locked in permanently)
  - Live mode skips dashboard entirely, goes directly to Clusters view
  - Clear message in live mode: "Live cluster browser — use --bundle for Attention Dashboard"
  - Removes all live-mode attention-related bugs permanently
  - Doubles development velocity by eliminating dual-mode complexity
- **Log filter defaults to ALL** when navigating from dashboard
  - Users can still apply Ctrl+E (ERROR) or Ctrl+W (WARN+ERROR) filters as needed
  - Better default UX - see all context first, then filter down

### Added
- Visible selection highlighting in dashboard (cyan background with inverted colors)
- [BUNDLE] prefix in status bar when using bundle mode
- Session state tracking for dashboard cursor position
- Expandable event line infrastructure with pod sub-navigation:
  - Arrow down (↓) or 'j' enters pod list when item is expanded
  - Arrow up/down navigates within pod list with visible highlighting
  - Enter on selected pod jumps to that pod's logs
  - Left arrow (←) or 'h' exits pod list back to main items
- Pod event counts displayed next to each affected pod in expanded view

### Technical
- Added `attentionCursor`, `expandedItems`, and `subCursor` state fields to App struct
- Added `HasError()` and `GetError()` methods to App for pre-TUI error checking
- Keyboard navigation handled before general table navigation in Update()
- Initial view selection based on mode: bundle → Attention, live/mock → Clusters
- Selection rendering uses `isSelected` parameter in `renderAttentionItem()`
- Pod highlighting in expanded views uses `inSubNav` parameter
- Cyan/dark-gray color scheme for maximum visibility
- Error check in `cmd/tui.go` before launching Bubble Tea program

## [0.3.3] - 2025-12-04 (IN PROGRESS)

### Added
- **🔥 Attention Dashboard**: New default root view that immediately shows cluster health status
  - Detects critical issues: CrashLoopBackOff, OOMKilled, ImagePullBackOff, Evicted pods
  - Detects pod restarts (≥3 in recent period)
  - Identifies high error/warning counts in logs
  - Shows etcd health issues (bundle mode)
  - Detects NotReady nodes
  - Displays DaemonSet incomplete deployments
  - Parses cluster events for warnings and failures
  - Clean "All good ✨" state when no issues detected
  - Severity-based grouping: Critical, Warning, Info
  - One-key drill-down (1-9 for quick jump, Enter for details)
  - Toggle classic navigation with 'c' key
  - Configurable default view preference

### Added (Technical)
- `internal/tui/attention.go` - Attention Dashboard view and orchestration
- `internal/tui/attention_signals.go` - Signal detection engine with 5 detector tiers
- `internal/bundle/etcd.go` - etcd health file parsers (alarmlist, endpointhealth)
- `internal/bundle/systeminfo.go` - System health parsers (memory, disk)
- Extended kubectl parsers: ParseNodes(), ParseDaemonSets()
- New DataSource interface methods: GetNodes(), GetEtcdHealth(), GetSystemHealth()
- `ViewAttention` type added to navigation flow
- Attention-specific styles with emoji indicators and severity colors

### Changed
- Default root view is now Attention Dashboard (classic view accessible via 'c' key)
- Config supports `defaultView` setting for user preference persistence

## [0.3.2] - 2025-12-03

### Fixed
- **Describe function in Live mode**: Fixed pod/deployment/service describe breaking in Live mode
  - Root cause: describe functions were calling `GetPods("")` which fails without projectID in Live mode
  - Solution: Use proper `DataSource.DescribePod/Deployment/Service()` interface methods
  - All three describe functions now work correctly in Live, Bundle, and Demo modes
  - Removed mock data fallbacks in favor of proper datasource abstraction

### Refactored
- **DataSource architecture**: Unified data layer with clean interface abstraction
  - New `internal/datasource/` package with `DataSource` interface
  - Three implementations: `LiveDataSource`, `BundleDataSource`, `EmbeddedDataSource`
  - Zero code duplication between modes
  - No fallback code paths - each mode is self-contained
- Demo mode (`--mockdata`) now uses embedded example bundle instead of synthetic data
- Selection preservation infrastructure added (resets to top for now, full implementation deferred)

### Investigated
- **Selection preservation**: Table position restoration when navigating back
  - Added `savedRowName` field for storing selected row identifier
  - Implementation blocked by bubble-table library API limitations
  - Library doesn't expose row iteration or selection-by-data methods
  - Documented in code comments for future implementation options
  - Workarounds would require: library fork, parallel data structures, or different table library

### Removed
- Dead code in fetch functions (300+ lines of fallback calls eliminated)
- Silent fallback behaviors between modes

### Technical
- `getMock*()` functions retained for test suite only
- All fetch functions simplified to single datasource call
- Clean separation: Live → API, Bundle → Files, Demo → Example
- Selection preservation requires architectural changes to implement fully

### Known Limitations
- Table selection currently resets to top when navigating back (library API constraint)
- Full implementation deferred until bubble-table API enhancement or library replacement

## [0.3.1] - 2025-12-03

### Added
- **Vim-style log navigation**: `g` key jumps to first log line, `G` jumps to last line (instant even on 5M-line logs)
- **Universal back navigation**: `b` key works everywhere alongside `Esc` for intuitive navigation
- **Word wrap toggle**: `w` key toggles soft word-wrap in log view with "Wrap:On" indicator
- **Enriched bundle pod details**: Bundle mode now shows full pod metadata from kubectl output
  - Pod status (Ready, Status, Age, IP, Restarts, ReadinessGates)
  - Kubernetes events attached to pods (17 events loaded from example bundle)
  - 93 pods with full kubectl data vs basic 86 pod inventory
- **Events parsing**: ParseEvents() function extracts pod events from kubectl output

### Changed
- Log view horizontal scrolling improved with word wrap support
- Bundle pod describe now shows data comparable to live mode
- Help screen updated with new keyboard shortcuts

### User Experience
- Log navigation feels instant and responsive with vim muscle memory
- Long log lines are now readable with toggleable word wrap
- Back navigation is more intuitive with `b` key option
- Bundle mode provides richer pod context for troubleshooting

## [0.3.0] - 2025-12-01

### Fixed
- **BUG-001**: CRD version selection 404 errors - now properly handles CRDs with multiple versions
- **BUG-002/003**: Nil pointer crashes and bundle path validation issues
- **Bugbash 2025-11** - Fixed 14 bugs across 4 systematic rounds:
  - Symlink panic during tar.gz extraction (now skips with warning)
  - Silent mock data fallbacks eliminated - all 10 fetch functions now fail explicitly with verbose errors
  - Filter state persistence after search exit
  - Search results becoming stale after filter changes
  - Tail mode tick pattern breaking continuous log updates
  - Log viewport not resizing on terminal window changes
  - vim j/k navigation advertised but not implemented
  - Ctrl+L eating next keystroke (remapped to refresh)
  - Incomplete error context in verbose mode
  - Bundle loading confusing messages
- Pod parsing now handles dash-separated filenames correctly

### Added
- Auto-version detection from git tags in Makefile
- Mode indicators in breadcrumb: [LIVE]/[BUNDLE]/[MOCK] for clear data source visibility
- vim j/k navigation in all table views (finally implemented)
- Verbose error context with `--verbose` flag across all operations
- Early validation for tarball paths in TUI command
- Config validate and set commands
- Comprehensive help system improvements
- Bundle size limit flag: `--limit` for bundle command

### Changed
- **Breaking**: Removed tarball support - bundles must be pre-extracted before import
- Bundle command terminology: 'import' → 'info' for clarity
- Loading messages now mode-aware (differentiate live/bundle/mock loading states)
- Filter clearing now also clears search state for consistency

### Removed
- Tarball extraction support (security and complexity reasons)
- Silent fallback behaviors (replaced with explicit errors)

## [0.2.1] - 2025-11-26

### Fixed
- Deployment replica counts now display correctly instead of always showing 0/0
- Services no longer show mock data in online mode when API errors occur
- Added DeploymentScale struct for nested replica data support
- Implemented multi-tier fallback strategy for replica count extraction

### Changed
- Updated fetchServices() to match fetchDeployments() error handling pattern

## [0.2.0] - 2025-11-26

### Added
- **Describe feature** for Pods, Deployments, and Services (press 'd' key)
- Comprehensive unit test suite with race detection (53 tests)
- Package-level godoc documentation for all packages
- Offline mode with automatic fallback to mock data
- Test coverage reporting (~90% for core packages)

### Fixed
- Go version in go.mod (corrected from 1.25 to 1.23)
- Pod.HostnameI typo renamed to Pod.Hostname
- Pod NODE column now displays correctly using multiple field fallbacks
- Namespace counts in Projects view now show accurate data
- Data extraction issues in Pods, Deployments, and Projects views

### Changed
- Enabled race detection in Makefile test target
- Renamed test_describe.go to avoid conflicts

## [0.1.0] - 2025-11-20

### Added
- **CRD (Custom Resource Definition) Explorer** with instance browsing
- CRD description toggle with 'i' key
- Instance counter column in CRD table
- Realistic and varied CRD instance counts
- Prominent offline mode warning banner
- Deployments and Services views with keyboard navigation
- View switching: 1=Pods, 2=Deployments, 3=Services

### Fixed
- CRD instance counter now uses live API data
- CRD navigation performance issue documented
- Navigation flow and offline mode functionality
- Table alignment (removed border styling from header)

## [0.0.3] - 2025-11-15

### Added
- Navigation stack for breadcrumb-style navigation

back/forward
- Projects view with namespace counts
- Improved help screen with ASCII logo
- Better styling and formatting throughout

### Changed
- Improved CRD browser filtering and navigation

## [0.0.2] - 2025-11-10

### Added
- Initial TUI framework using Bubble Tea
- Cluster listing view
- Basic navigation between views

### Fixed
- Rancher API type definitions corrected
- Live instance testing functionality

## [0.0.1] - 2025-11-01

### Added
- Initial project scaffolding
- Basic CLI structure with Cobra
- Configuration management with Viper
- Rancher API client implementation
- README with project vision
- LICENSE (Apache 2.0)
- Git repository initialization

---

## Version History Summary

- **v0.3.0** (2025-12-01): Bugbash 2025-11 - 14 bugs fixed, no silent fallbacks, mode indicators
- **v0.2.1** (2025-11-26): Bug fixes for replica counts and error handling
- **v0.2.0** (2025-11-26): Describe feature, comprehensive testing, documentation
- **v0.1.0** (2025-11-20): CRD Explorer, offline mode, view switching
- **v0.0.3** (2025-11-15): Navigation improvements, projects view
- **v0.0.2** (2025-11-10): TUI framework, cluster listing
- **v0.0.1** (2025-11-01): Initial release

---

## Upgrade Notes

### Upgrading to 0.3.0
- **Breaking Change**: Tarball extraction removed - bundles must be pre-extracted before using `r8s bundle` command
- Bundle command changed from `r8s bundle import` to `r8s bundle info`
- All silent mock data fallbacks eliminated - errors now shown explicitly (use `--verbose` for details)
- New mode indicators in UI: [LIVE], [BUNDLE], or [MOCK] show active data source
- vim j/k navigation now works in all table views
- Use `--limit` flag to adjust bundle size limits if needed

### Upgrading to 0.2.1
- No breaking changes
- Deployment replica counts will now display correctly
- Error messages in online mode are now more informative

### Upgrading to 0.2.0
- New 'd' key binding for describe functionality
- Offline mode automatically detects connection failures
- Run `go test -race ./...` to verify tests pass

### Upgrading to 0.1.0
- New view navigation keys: 1, 2, 3 for Pods/Deployments/Services
- CRD explorer accessible from cluster view with 'C' key

---

## Contributors

- Development Team

For detailed commit history, see: `git log --oneline`
