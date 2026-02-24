# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.1] - 2026-02-25

### Fixed
- **Critical Bug**: Pattern detection was not finding issues in bundles
  - Root cause: `collectBundleContent()` used incorrect file path patterns
  - Empty `bundleType` created invalid paths like `/kubectl/pods` (leading slash)
  - Fixed paths: `agent-logs/*` (was `agent/logs/*.log`), added `systemlogs/journald-*`
  - Issue #84: https://github.com/Rancheroo/r8s/issues/84

### Changed
- Default `bundleType` to `rke2` when not specified
- Updated path patterns to match actual RKE2 bundle structure

**Note:** All v1.0.0 users should upgrade to v1.0.1 immediately.

## [1.0.0] - 2026-02-24

### Added
- **kubectl-r8s plugin**: Use r8s as a kubectl plugin for familiar UX
- **UX Loading Messages**: Cowsay-style loading messages with Rancher references
- **Error Handling**: Better error messages with typo suggestions
- **British Spelling Support**: `analyse` works alongside `analyze`

### Changed
- **Simplified Help**: Reduced from 60 to 45 lines
- **Removed Bloat**: Deleted 2,500 lines of unused code per Elon's 5 Laws
  - Removed: config command, completion command, get_pods/get_nodes (replaced by plugin)
  - Simplified: Global flags (removed cfgFile, contextName, scanDepth)

### Sprint 12 Complete
- ✅ kubectl-r8s plugin built and tested
- ✅ 10 production bundles validated (100% success)
- ✅ 19 AI patterns working
- ✅ Elon's 5 Laws cleanup applied
- ✅ UX improvements with personality

## [0.9.0] - 2026-02-24

### Added
- **AI Pattern Engine v2**: 19 patterns with confidence scoring
- **Natural Language Queries**: `r8s ask "why is nginx crashing?"`
- **Root Cause Hints**: Explains why and how to fix
- **Export Formats**: SARIF, JUnit, Markdown, JSON
- **Parallel Analyzer**: Performance optimization with goroutines
- **Pattern Registry**: `r8s patterns list/show/search`

## [0.8.0] - 2026-02-13

### Added
- CLI-first architecture
- kubectl-compatible commands (get, logs, describe)
- Bundle validation
- Test cluster command

## [0.7.0] - 2026-02-06

### Added
- Initial release
- Basic kubectl commands
- Bundle parsing