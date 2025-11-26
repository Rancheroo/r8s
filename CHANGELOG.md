# Changelog

All notable changes to the r9s project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive API field documentation for Deployment struct
- Consistent mock data fallback behavior across all resource types

### Changed
- Services fetch now returns errors in online mode instead of silently falling back to mock data

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

- **v0.2.1** (2025-11-26): Bug fixes for replica counts and error handling
- **v0.2.0** (2025-11-26): Describe feature, comprehensive testing, documentation
- **v0.1.0** (2025-11-20): CRD Explorer, offline mode, view switching
- **v0.0.3** (2025-11-15): Navigation improvements, projects view
- **v0.0.2** (2025-11-10): TUI framework, cluster listing
- **v0.0.1** (2025-11-01): Initial release

---

## Upgrade Notes

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
