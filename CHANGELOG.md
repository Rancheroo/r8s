# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.2] - 2026-03-05

### Added
- **Demo Mode**: Added `scripts/setup-demo-bundles.sh` to generate scenarios for team demos.
- **Documentation**: New `docs/DEMO_SCRIPT.md` with speaker notes and `CONTRIBUTING.md` issue templates.
- **Refactor**: Moved debug scripts to `scripts/` directory for cleaner root.

### Changed
- Updated `README.md` to reflect v1.3.2 status.
- Improved error handling in bundle validation.

## [1.2.1] - 2026-03-04

### Added
- **Natural Language Query**: `r8s ask` now supports `why`, `show`, `which` intents.
- **Critical Pattern Detection**: Added 5 new patterns including Etcd Quorum Loss.

### Fixed
- Fixed critical bug where pattern detection wasn't finding issues due to incorrect file path patterns.
- Updated default `bundleType` to `rke2`.

## [1.1.0] - 2026-03-03

### Added
- **Bundle Format v1.1 Support**: Full support for new Rancher log bundle format
  - Virtualization detection via `systemd-detect-virt` (kvm, vmware, docker, etc.)
  - Versions file parser for comprehensive system/component information
  - Pod describe directory support for per-namespace pod descriptions
  - Sysstat data directory support for performance metrics
- **Memory Unit Parsing**: Handles Gi, Mi, Ki, B suffixes in memory files
- **Path Fallback System**: Automatic backward compatibility with old bundle formats
  - Tries new locations first (e.g., `systeminfo/memory`), falls back to old (e.g., `systeminfo/freem`)
  - Supports both `systeminfo/dmesg` (new) and `systemlogs/dmesg` (old)
- **Integration Tests**: Comprehensive tests for both old and new bundle formats

### Fixed
- **test-cluster command not recognized**: Added missing commands to `isKnownCommand()` list
  - Commands affected: `test-cluster`, `patterns`, `completion`
  - Root cause: Hardcoded command list didn't include all subcommands

### Changed
- **SystemHealthInfo struct**: Added `VirtType` field for virtualization platform detection
- **PathResolver interface**: Extended with new methods for v1.1 format support
  - `GetRootVersionsFile()`, `GetSysteminfoPath()`, `GetDmesgPaths()`, `GetMemoryPaths()`
- **Test Coverage**: Increased to 69.6% (exceeds 45% target)

### Technical Details
- New files: `internal/bundle/versions.go`, `internal/bundle/integration_v1.1_test.go`
- Modified: `internal/bundle/systeminfo.go`, `internal/bundle/dmesg.go`, `internal/bundle/paths.go`, `cmd/root.go`
- Bundle format detection: Automatic via file presence (no user action required)

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
## [1.3.3] - 2026-03-05

### Added
- **AI Context Generation**: New `r8s generate prompt` feature to extract and format bundle data for LLM analysis.
- **OpenCode Integration**: Seamless support for piping bundle context to local AI with `| opencode run`.
- **Intelligent Summarization**: Automatically includes failing pods, relevant logs, and warning events in generated prompts.

### Changed
- Updated demo script to include AI integration examples.
