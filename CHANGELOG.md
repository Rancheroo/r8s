# Changelog

All notable changes to the r9s project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Released for Testing]

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
