# Lessons Learned - r8s Development

This document captures key lessons learned during r8s development to improve future decision-making and avoid repeating mistakes.

---

## Architecture & Design Patterns

### DataSource Interface Abstraction (v0.3.2)

**Context:** Originally had mode-specific logic scattered throughout TUI code with fallbacks and branching.

**Lesson:** **Unified interfaces eliminate code duplication and prevent mode-specific bugs**

**What We Did:**
- Created `DataSource` interface in `internal/datasource/`
- Three clean implementations: `LiveDataSource`, `BundleDataSource`, `EmbeddedDataSource`
- TUI layer only depends on interface, not implementations

**Impact:**
- ✅ Eliminated 300+ lines of duplicate/fallback code
- ✅ Single code path for all modes
- ✅ Mode-agnostic TUI - no if/else branching
- ✅ Easy to add new modes (just implement interface)

**Key Insight:** When you find yourself writing `if liveMode { ... } else if bundleMode { ... }` repeatedly, you need an interface.

---

### Bug: Live Mode Describe Broken (v0.3.2)

**Context:** Describe feature ('d' key) worked in Bundle mode but failed in Live mode.

**Lesson:** **Always use your own interfaces - bypassing them creates mode-specific bugs**

**What Went Wrong:**
```go
// WRONG: Bypassed DataSource interface
pods, err := a.dataSource.GetPods("", namespace)  // Empty projectID fails in Live
// Then searched for specific pod in returned list
```

**Root Cause:**
- `describePod()` called `GetPods("")` instead of using `DescribePod()` interface method
- Live mode requires valid `projectID` for GetPods() - empty string causes API failure
- Bundle mode ignores `projectID` parameter, so it worked fine
- Bug only manifested in Live mode

**The Fix:**
```go
// CORRECT: Use proper interface method
data, err := a.dataSource.DescribePod(clusterID, namespace, name)
```

**Impact:**
- ✅ Works in all modes (Live/Bundle/Demo)
- ✅ Removed 138 lines of fallback code
- ✅ Consistent behavior across modes

**Key Insight:** If you design an interface with specific methods (DescribePod), USE THEM. Don't work around your own abstractions.

---

### Selection Preservation Investigation (v0.3.2)

**Context:** Wanted to preserve table selection when navigating back to previous view.

**Lesson:** **Third-party library limitations can block features - document and move on**

**What We Discovered:**
- bubble-table library doesn't expose methods to:
  - Iterate through rows
  - Set selection by row data/content
  - Query row at specific index
- Only provides: GetHighlightedRow(), row count, pagination

**Attempted Workarounds:**
1. **Fork library** - Adds maintenance burden
2. **Parallel data structures** - Duplicates state, risks drift
3. **Switch table libraries** - Major refactor

**Decision:**
- Added infrastructure (`savedRowName` field)
- Documented limitation in code comments
- Selection resets to top (acceptable UX)
- Deferred until library improves or we switch

**Key Insight:** Don't fight library limitations if workarounds are complex. Document, defer, and reassess when priorities change.

---

## Code Quality & Maintenance

### Eliminating Silent Fallbacks (v0.3.0)

**Context:** Originally had silent fallbacks from API to mock data when errors occurred.

**Lesson:** **Silent fallbacks hide bugs and confuse users**

**What We Changed:**
- Removed all silent fallbacks
- Errors now fail explicitly with context
- `--verbose` flag provides detailed error context
- Mode indicators ([LIVE]/[BUNDLE]/[MOCK]) show data source

**Impact:**
- ✅ Users know when something is wrong
- ✅ Easier to debug issues
- ✅ Clear expectations about data source
- ✅ No mysterious "why is this showing demo data?" questions

**Key Insight:** Explicit errors are better than silent fallbacks. If something fails, tell the user why.

---

### Code Organization

**Lesson:** **Clean package boundaries improve maintainability**

**Package Structure:**
```
internal/
├── datasource/     # Data layer - interface + implementations
├── tui/           # UI layer - depends only on datasource interface
├── rancher/       # API client - used by LiveDataSource
├── bundle/        # File parsing - used by BundleDataSource
└── config/        # Configuration management
```

**Benefits:**
- Clear separation of concerns
- Easy to test (mock DataSource interface)
- UI code is mode-agnostic
- Can add new data sources without touching UI

**Key Insight:** Good package boundaries = easier testing + cleaner code + faster development.

---

## Development Process

### Incremental Refactoring

**Lesson:** **Large refactors work better when done incrementally with testing at each step**

**Our Approach for DataSource Refactor:**
1. Create interface + implementations
2. Keep old code paths working
3. Switch one fetch function at a time
4. Test after each switch
5. Remove old code once all switched
6. Final build + test

**Benefits:**
- Always have working code
- Easy to identify which change broke something
- Can pause/resume refactor
- Less stressful than big-bang rewrites

**Key Insight:** "Make it work, make it right, make it fast" - in that order.

---

### Documentation During Development

**Lesson:** **Document decisions and limitations while context is fresh**

**What We Did:**
- CHANGELOG.md updated with each version
- Code comments explain "why" for complex logic
- Selection preservation limitation documented in code
- This LESSONS_LEARNED.md captures insights

**Key Insight:** Document the "why" immediately. Future you (and others) will thank you.

---

## Interface Design Best Practices

### DataSource Interface Success Factors

**What Made It Work:**

1. **Comprehensive Methods:**
   ```go
   GetPods(projectID, namespace) ([]Pod, error)
   DescribePod(clusterID, namespace, name) (interface{}, error)
   GetLogs(...) ([]string, error)
   ```

2. **Mode() Method:**
   - Returns "LIVE", "BUNDLE", or "MOCK"
   - Used for UI indicators
   - Helps with debugging

3. **Single Responsibility:**
   - Each method does one thing
   - No overloaded "get data somehow" methods
   - Clear contracts

**Anti-Pattern to Avoid:**
```go
// DON'T: One method that does different things based on mode
GetData(mode string, params ...interface{}) (interface{}, error)

// DO: Separate methods with clear signatures
GetPods(projectID, namespace string) ([]Pod, error)
GetDeployments(projectID, namespace string) ([]Deployment, error)
```

**Key Insight:** Specific methods > generic methods. Type safety > flexibility.

---

## Testing Insights

### Build-Test-Commit Cycle

**Lesson:** **Always build before committing**

**Our Process:**
1. Make code changes
2. `make build` - verify it compiles
3. Manual smoke test if significant change
4. Update CHANGELOG if user-facing
5. Git commit with detailed message
6. Tag version if release

**Benefits:**
- Catch compilation errors before commit
- Commits are always buildable
- Easy to bisect if bugs appear
- Clean git history

**Key Insight:** "If it doesn't build, it doesn't ship" - even for commits.

---

## Future Considerations

### When to Refactor

**Green Flags (Do Refactor):**
- Same code pattern repeated 3+ times
- Mode-specific if/else appearing everywhere
- Adding new feature requires changing many files
- Silent failures or mysterious bugs

**Red Flags (Don't Refactor Yet):**
- Code works and not changing frequently
- Only one instance of the pattern
- Would require library fork or major dependency change
- Time better spent on new features

**Key Insight:** Refactor when pain > effort to refactor. Not before.

---

### Library Selection Criteria

**Lessons from bubble-table Experience:**

**Must Haves:**
- ✅ Core functionality works well
- ✅ Active maintenance
- ✅ Good documentation

**Nice to Haves:**
- Row iteration support
- Programmatic selection control
- Extensibility points
- Type safety

**Dealbreakers:**
- Abandoned (<6 months since last update)
- No documentation
- Frequent breaking changes

**Key Insight:** Can work around missing "nice to haves" if core is solid. Can't fix abandoned libraries.

---

## Summary

### Top 5 Lessons

1. **Unified interfaces eliminate mode-specific bugs**
2. **Use your own abstractions - don't bypass them**
3. **Silent fallbacks hide bugs - fail explicitly**
4. **Document limitations when you hit them**
5. **Incremental refactoring > big-bang rewrites**

### Evolution of r8s Architecture

**v0.2.0 and earlier:**
- API calls directly from TUI
- Silent fallbacks to mock data
- Mode-specific logic everywhere

**v0.3.0:**
- Eliminated silent fallbacks
- Added mode indicators
- Better error messages

**v0.3.2:**
- DataSource interface abstraction
- Complete mode separation
- Single code path for all modes
- Bug-free describe across all modes

**Next Steps:**
- Consider table library alternatives if selection preservation becomes priority
- Add more comprehensive tests
- Continue refining interface as needs emerge

---

*Last Updated: 2025-12-03*
*Version: 0.3.2*
