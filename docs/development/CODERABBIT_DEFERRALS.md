# CodeRabbit Deferred Items

**Status:** Intentional design decisions documented for v0.7.0  
**Date:** 2026-02-13  
**Revisit:** Post-Sprint 6 (v0.7.2)

---

## Item #1: os.Exit() in Cobra RunE Functions

**File:** `cmd/testcluster.go`  
**Lines:** 95-97, 117  
**Issue:** Using `os.Exit()` inside `RunE` instead of returning errors

### Current Code
```go
if err != nil {
    fmt.Fprintf(os.Stderr, "Error: failed to load bundle: %v\n", err)
    os.Exit(2)  // ← CodeRabbit: "Should return error, not exit"
}
// ...
os.Exit(exitCode)  // ← CodeRabbit: "Should return nil/error"
return nil
```

### Why Deferred (Intentional)

**CI Contract Requirement:**
- Exit code 0 = All tests passed (healthy cluster)
- Exit code 1 = Tests failed, issues detected
- Exit code 2 = Bundle parsing error (can't run tests)

This is a **command-line interface contract** enforced by CI/CD pipelines:
```bash
r8s test-cluster ./bundle
if [ $? -eq 2 ]; then
    echo "Bundle corrupt - skip this artifact"
    continue
fi
```

Returning an error would give exit code 1, breaking the contract between "can't test" vs "tested and found issues."

### Design Principle
**Explicit Mode > Implicit:** Exit codes are part of the CLI API surface, not an implementation detail.

### Future Resolution
Consider adding a `SilenceErrors: true` and `PostRun` hook for cleaner separation, but only post-Sprint 6 when CI is stable enough to validate the change.

---

## Item #5: Unchecked Error Returns

**Files:** Various parsers (`internal/bundle/*.go`)  
**Pattern:** `file, _ := os.Open(path)` instead of `if err != nil`

### Current Code
```go
func parseSomething(path string) []Item {
    file, _ := os.Open(path)  // ← CodeRabbit: "Error not checked"
    defer file.Close()
    // ... parse ...
}
```

### Why Deferred (Intentional)

**Tolerant Parsing Philosophy:**

From `PRINCIPLES.md` - "Truth Only™":
> "Partial truth > No truth. Missing files should return empty results, not errors."

Bundles are **incomplete by design** (users collect what they can). A missing file means "no data" not "error condition":

- Missing `kubectl get nodes` → Return empty node list (valid state)
- Missing `dmesg` → No kernel events to analyze (valid state)
- Corrupt file → Log warning, return what we can (partial truth)

### Design Principle
**Graceful Degradation:** Users need visibility into what exists, not error cascades from what's missing.

### Future Resolution
Add structured logging for "file not found" vs "parse error" distinction. Track in v0.7.4 telemetry work.

---

## Item #7: Multiple Completeness/Health Systems

**Files:** 
- `internal/bundle/completeness.go` (validation layer)
- `internal/tui/fetch.go` (UI layer)  
**Issue:** Two different "completeness" calculations

### Current State

**Bundle Completeness** (precise, for validation):
```go
func AnalyzeCompleteness(bundlePath string) CompletenessReport {
    // Checks 15+ required files
    // Calculates percentage from expected bundle
    // Returns detailed breakdown
}
```

**UI Completeness** (simple, for display):
```go
func GetStatus(bundle Bundle) string {
    // Simple "Complete/Partial/Missing" buckets
    // Based on parser availability
}
```

### Why Deferred (Intentional)

**Separation of Concerns:**

| Layer | Responsibility | Audience |
|-------|---------------|----------|
| Validation | "Is this bundle sufficient for analysis?" | CI/CD, automated tooling |
| UI | "What's the user-visible status?" | Human operators |

Merging them would:
1. Complicate the UI with validation details users don't need
2. Slow validation with UI rendering logic
3. Create coupling between parsing and display

### Design Principle
**Single Responsibility:** Each system serves one audience well rather than both poorly.

### Current Sync
Both systems use the same thresholds (70% = Complete, 40% = Partial) defined in `internal/bundle/completeness.go`.

### Future Resolution
Extract shared threshold constants to prevent drift. Consider unifying only if maintenance burden proves high (not currently an issue).

---

## How to Handle These in Reviews

When CodeRabbit flags these patterns in future PRs:

```
@coderabbitai Acknowledged - deferred item

This follows intentional pattern from CODERABBIT_DEFERRALS.md:
- Item #X: [Brief rationale]

Deferred until [milestone].
```

This teaches CodeRabbit while documenting the decision.

---

## Sprint 6 Note

None of these deferrals block Sprint 6 (CI Stability). They're architectural decisions that:
- ✅ Work correctly in production
- ✅ Are well-documented
- ✅ Serve specific design goals
- ⚠️ Trigger lint warnings (acceptable trade-off)

Revisit in v0.7.4 if technical debt burden increases.
