# Lessons Learned - r8s Development

## Overview
This document captures key insights, patterns, and lessons learned during the development of r8s (Rancheroos), a TUI for browsing Rancher-managed Kubernetes clusters and analyzing log bundles.

---

## Development Timeline

### Week 1: Foundation & Core Features
- **Phase 0**: Rebranding from r9s to r8s
- **Phase 1**: Core TUI navigation (Clusters → Projects → Namespaces → Pods)
- **Phase 2**: Enhanced navigation with Deployments and Services
- **Phase 3**: Log viewing with color highlighting and search
- **Phase 4**: Bundle import for offline cluster analysis
- **Phase 5**: Bundle log viewer integration
- **Phase 5B**: kubectl parsing and bug fixes
- **CLI UX**: Explicit mode control and improved help
- **Verbose Errors**: Enhanced error handling for debugging

---

## Key Lessons Learned

### 1. **Silent Fallbacks Are Confusing**

**Problem:**
Early versions would silently fall back to mock data when API connections failed, leaving users confused about whether they were seeing real or fake data.

**Solution:**
- Added explicit `--mockdata` flag for demo mode
- Removed silent fallbacks
- Show clear error messages when API fails
- Added mode indicators in UI

**Lesson:**
> Users deserve to know what data they're seeing. Explicit is better than implicit. Silent fallbacks create mystery and erode trust.

---

### 2. **Help Should Be DefaultAssistant behavior**

**Problem:**
Running `r8s` without arguments tried to launch the TUI, which could fail mysteriously. New users had no way to discover features.

**Solution:**
- Changed root command to show help by default
- Moved TUI to subcommand: `r8s tui`
- Added comprehensive examples in help text
- Documented keyboard shortcuts

**Lesson:**
> Modern CLIs show help when run without args. Make discovery easy. Don't assume users know what to do.

---

### 3. **Verbose Errors Save Time**

**Problem:**
Generic errors like "bundle file not found" required back-and-forth to diagnose. Missing context about file paths, expected formats, etc.

**Solution:**
- Added `--verbose` / `-v` flag
- Enhanced errors with:
  * File paths and current directory
  * Expected vs actual values
  * Actionable hints for fixes

**Lesson:**
> Detailed error messages are cheap to implement but invaluable for debugging. Context + hints = faster problem solving.

---

### 4. **Search Implementation Gotchas**

**Problem:**
Search hotkeys conflicted with typing in search mode (Bug #7). Users couldn't type 'd' in their search query.

**Solution:**
- Fixed precedence: Check `searchMode` BEFORE processing hotkeys
- Only non-search keys exit search mode

**Lesson:**
> Input handling order matters. Mode-specific behavior should be checked first. State machines need careful ordering.

---

### 5. **Empty Lists ≠ Mock Data**

**Problem:**
Early code treated empty resource lists as errors and fell back to mock data. Real empty lists (valid state) were never shown.

**Solution:**
- Distinguish between:
  * Error fetching data → show error
  * Successfully fetched empty list → show "No X available"
  * Offline mode → show mock data

**Lesson:**
> Empty is a valid state. Don't conflate "no results" with "error". Bundle data might legitimately have zero resources.

---

### 6. **Bundle Structure Isn't Standard**

**Problem:**
Assumed all RKE2 bundles have consistent structure. Real bundles vary wildly based on collection tool and version.

**Solution:**
- Made kubectl parsing lenient
- Added fallbacks for missing fields
- Log parse failures as warnings, not errors
- Support multiple file naming patterns

**Lesson:**
> Real-world data is messy. Build robustness through fallbacks, not strict schemas. Log issues but continue processing.

---

### 7. **Filter State Must Persist**

**Problem:**
Search results showed all logs when filtered logs were intended. Filters weren't applied to search matches.

**Solution:**
- Created `getVisibleLogs()` helper
- Applied filters before search
- Search only through visible (filtered) logs

**Lesson:**
> When multiple features interact (filtering + search), ensure state consistency. Helper functions clarify intent.

---

### 8. **Auto-Formatting Breaks SEARCH Blocks**

**Problem:**
Using `replace_in_file` would fail because auto-formatter changed the file after writing, breaking exact matches.

**Solution:**
- Always use `final_file_content` as reference for next SEARCH block
- Account for formatting changes (quotes, spacing, imports)

**Lesson:**
> Tools that auto-format break exact-match find/replace. Always reference the final state after edits.

---

### 9. **Testing in Headless Environments**

**Problem:**
TUI can't run without a TTY. Hard to test in CI or SSH sessions.

**Solution:**
- Created mockable data sources
- Unit tests don't require TUI
- Verbose errors help diagnose issues without running UI

**Lesson:**
> Design for testability from the start. Abstract UI from logic. Make components testable in isolation.

---

### 10. **Documentation Organization Matters**

**Problem:**
Root directory cluttered with 30+ markdown files. Hard to find relevant documentation.

**Solution:**
- Archive old phase documents to `docs/archive/`
- Keep only current status and high-level docs in root
- Organize by feature/phase for easy navigation

**Lesson:**
> Documentation accumulates fast. Regular cleanup prevents overwhelm. Archive completed work, keep active docs visible.

---

## Development Patterns

### Good Patterns That Worked

1. **Phased Development**
   - Break large features into phases
   - Complete one phase before starting next
   - Document each phase completion

2. **Test-Driven Bug Fixes**
   - Create test plan before fixing bug
   - Verify fix with specific test cases
   - Document what was broken and how it was fixed

3. **Data Source Abstraction**
   - Single interface for live API and offline bundles
   - Makes switching modes seamless
   - Enables testing without real API

4. **Explicit Mode Control**
   - `--mockdata` for demo
   - `--bundle` for offline analysis
   - Default behavior clearly documented

5. **Comprehensive Completion Docs**
   - Write detailed completion docs for each feature
   - Include before/after examples
   - Document lessons learned immediately

### Patterns to Avoid

1. **Silent Fallbacks**
   - Don't fall back to mock data silently
   - Always tell user what data they're seeing

2. **Unclear Mode Indicators**
   - Make current mode obvious in UI
   - Differentiate live/offline/bundle modes visually

3. **Generic Error Messages**
   - Avoid "failed to load" without context
   - Always include file paths and hints

4. **Cluttered Root Directory**
   - Archive completed phase docs regularly
   - Keep root clean and navigable

---

## Technical Decisions

### Why Go?
- Strong standard library (TUI frameworks, tar handling)
- Cross-platform compilation
- Fast performance for log parsing
- Good Kubernetes ecosystem support

### Why Bubble Tea (TUI framework)?
- Modern, well-maintained
- Component-based architecture
- Good for complex interactive UIs
- Elm-architecture (predictable state management)

### Why Bundle Import?
- Enables offline cluster analysis
- Support teams need this for troubleshooting
- No dependency on live cluster access

### Why Verbose Flag?
- Debugging in production without code changes
- Helps users self-diagnose issues
- Minimal performance cost when disabled

---

## Code Quality Principles

### What Worked Well

1. **Clear function names**: `getVisibleLogs()` vs `getLogs()`
2. **Type safety**: Go's type system caught many bugs early
3. **Error wrapping**: `fmt.Errorf("x: %w", err)` preserves context
4. **Comments**: Explain "why" not "what"
5. **Consistent naming**: `fetchX()` for async, `getX()` for sync

### Areas for Improvement

1. **More unit tests**: Current test coverage is light
2. **Integration tests**: Need end-to-end bundle import tests
3. **Error types**: Custom error types could help categorization
4. **Logging**: Add structured logging for production debugging

---

## User Experience Insights

### What Users Want

1. **Discoverability**: Help should be obvious and comprehensive
2. **Transparency**: Show what data source is active
3. **Helpful errors**: Tell me what's wrong AND how to fix it
4. **Keyboard efficiency**: TUI users love keyboard shortcuts
5. **Demo mode**: Try before configuring API access

### What Surprised Us

1. **Mock data demand**: Users wanted demo mode for screenshots/testing
2. **Bundle analysis**: Offline analysis more popular than anticipated
3. **Search importance**: Log search is make-or-break feature
4. **Error verbosity**: Users appreciated detailed errors more than expected

---

## Performance Learnings

### What's Fast
- Go's bufio for streaming large logs
- Color rendering with lipgloss
- Table rendering with bubble-table
- Tar extraction with Go stdlib

### What's Slow
- JSON parsing of huge kubectl outputs (10MB+)
- Recursive directory walking in large bundles
- Regex search on millions of log lines

### Optimizations Applied
- Stream processing instead of loading all logs in memory
- Lazy loading of resources (fetch on demand)
- Size limits on bundle import (default 10MB)
- Skip parse errors instead of failing

---

## Future Considerations

### Features to Add
1. Live log tailing from API
2. Event timeline view
3. Resource graph visualization
4. Export filtered logs
5. Multi-bundle comparison

### Technical Debt to Address
1. Increase test coverage (target 80%+)
2. Add integration test suite
3. Refactor mock data generation
4. Improve bundle format detection
5. Add metrics/observability

### Documentation Needs
1. User guide with screenshots
2. Bundle format specification
3. Troubleshooting guide
4. Developer contributing guide
5. API client documentation

---

## Team Workflows

### What Worked
- Detailed phase planning before coding
- Testing after each feature
- Documentation as we go
- Regular cleanup/consolidation
- Commit messages with context

### What to Improve
- Earlier user feedback loops
- More pair programming on complex features
- Automated testing in CI
- Regular dependency updates
- Security scanning

---

## Closing Thoughts

Building r8s taught us that:

1. **User empathy matters** - Silent behaviors confuse users
2. **Documentation is development** - Good docs prevent issues
3. **Tests save time** - Bugs caught early are cheap to fix
4. **Cleanup is essential** - Technical debt accumulates fast
5. **Explicit > Implicit** - Clear modes and flags reduce confusion

The project evolved from a simple cluster browser to a comprehensive observability tool. Each phase built on learnings from the previous one, and iterative improvement proved more effective than big-bang rewrites.

---

**Date**: November 27, 2025
**Status**: Week 1 Complete - Solid foundation established
**Next**: Week 2 - Polish, testing, and production readiness
