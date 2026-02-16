# Test Patterns Guide

> For CodeRabbit and contributors: how we write tests in r8s.

## Philosophy

We test **support bundle analyzers** — functions that parse messy real-world log/config files and extract structured findings. Our tests must handle missing files, malformed input, and edge cases gracefully because support bundles from the field are never perfect.

**Principles:**
1. **Fail gracefully** — missing files return empty results, not errors
2. **Test the edges** — empty input, short lines, nil maps, invalid YAML
3. **Table-driven by default** — every function with >2 test cases uses table-driven tests
4. **Fixtures via temp dirs** — integration tests create realistic bundle directory structures

---

## Pattern 1: Fixture Creation (Temp Dir → Write → Defer Cleanup)

Integration tests that need filesystem fixtures follow this exact pattern:

```go
func createXxxTestBundle(t *testing.T) (string, func()) {
    t.Helper()
    tmpDir, err := os.MkdirTemp("", "r8s-xxx-test-")
    if err != nil {
        t.Fatalf("Failed to create temp dir: %v", err)
    }

    // Create directory structure matching real bundle layout
    for _, dir := range []string{"rke2/kubectl", "rke2/pod-manifests"} {
        os.MkdirAll(filepath.Join(tmpDir, dir), 0755)
    }

    cleanup := func() { os.RemoveAll(tmpDir) }
    return tmpDir, cleanup
}
```

**Usage at call site:**
```go
bundleRoot, cleanup := createXxxTestBundle(t)
defer cleanup()
```

**Key rules:**
- Helper returns `(string, func())` — root path and cleanup function
- Always call `t.Helper()` so failures point to the test, not the helper
- `t.Fatalf` on setup failure (can't continue without fixtures)
- Caller always `defer cleanup()`
- Prefix temp dirs with `r8s-xxx-test-` for easy identification

**File writing helper:**
```go
func writeFile(t *testing.T, path, content string) {
    t.Helper()
    os.MkdirAll(filepath.Dir(path), 0755)
    if err := os.WriteFile(path, []byte(content), 0644); err != nil {
        t.Fatalf("Failed to write %s: %v", path, err)
    }
}
```

### What CodeRabbit should flag:
- ❌ Missing `defer cleanup()` after fixture creation
- ❌ Missing `t.Helper()` in helper functions
- ❌ Using `t.Errorf` instead of `t.Fatalf` for setup failures

---

## Pattern 2: Table-Driven Tests

Every function with multiple input/output combinations uses this structure:

```go
func TestFunctionName_Category(t *testing.T) {
    tests := []struct {
        name     string
        input    <type>
        expected <type>
    }{
        {"descriptive name", inputVal, expectedVal},
        // ...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := FunctionUnderTest(tt.input)
            if got != tt.expected {
                t.Errorf("FunctionUnderTest(%v) = %v, want %v", tt.input, got, tt.expected)
            }
        })
    }
}
```

**Naming conventions:**
- Test names use `tt.input` as the subtest name when the input is self-describing (e.g., `TestNormalizePodName` uses `tt.input`)
- Otherwise use a descriptive `name` field (e.g., `"nil uses default"`, `"short victim name skipped"`)
- Group related tests: `TestParseJournaldLogs_AllPatterns`, `TestParseOOMEvents_VariousPatterns`

**Real example from `safeget_test.go`:**
```go
func TestSafeInt(t *testing.T) {
    tests := []struct {
        name     string
        input    interface{}
        expected int
    }{
        {"nil", nil, 0},
        {"int", 42, 42},
        {"int64", int64(42), 42},
        {"float64", 3.14, 3},
        {"string number", "42", 42},
        {"string invalid", "abc", 0},
        {"bool", true, 0},
    }
    // ...
}
```

### What CodeRabbit should flag:
- ❌ Non-table-driven tests when there are >2 cases
- ❌ Missing `t.Run` wrapper (prevents subtest isolation)
- ❌ Test case without a descriptive name

---

## Pattern 3: Assertion Strategy — `t.Fatalf` vs `t.Errorf`

**`t.Fatalf`** — Stop the test immediately. Used when:
- Setup/fixture creation fails
- The result count is wrong and subsequent assertions would panic
- An error is returned when none expected

```go
// Stop: can't check results[0] if len is wrong
if len(results) != 1 {
    t.Fatalf("expected 1 result, got %d", len(results))
}
// Safe to access results[0] now
if results[0].PodName != "my-app" {
    t.Errorf("expected pod name 'my-app', got '%s'", results[0].PodName)
}
```

**`t.Errorf`** — Continue the test. Used when:
- Checking field values (more failures = more signal)
- The test can safely continue after this assertion

**Pattern: Guard then inspect**
```go
results, err := AnalyzeOOMEvents(bundleRoot)
if err != nil {
    t.Fatalf("unexpected error: %v", err)           // Fatal: can't continue
}
if len(results) != 1 {
    t.Fatalf("expected 1 result, got %d", len(results))  // Fatal: prevents panic
}
if results[0].PodName != "my-app" {
    t.Errorf("expected pod 'my-app', got '%s'", results[0].PodName)  // Errorf: informational
}
if results[0].QoSClass != "Guaranteed" {
    t.Errorf("expected QoS 'Guaranteed', got '%s'", results[0].QoSClass)  // Errorf: collect all failures
}
```

### What CodeRabbit should flag:
- ❌ `t.Errorf` on length checks followed by index access (will panic)
- ❌ `t.Fatalf` inside `t.Run` for non-critical field checks (loses parallel info)
- ❌ Missing error check after functions that return `(result, error)`

---

## Pattern 4: Edge Case Coverage

Every analyzer test suite includes these edge cases:

| Edge Case | Example | Expected Behavior |
|-----------|---------|-------------------|
| Missing file | No events file exists | Return empty results, no error |
| Empty input | `""` passed to parser | Return empty slice |
| Whitespace only | `"   \n\n   \n"` | Return empty slice |
| Short/malformed lines | `"too few fields"` | Skip line silently |
| Nil inputs | `nil` map, `nil` slice | Return zero value |
| Invalid YAML | `"{{{invalid"` | Return empty, no panic |
| Wrong type | Non-Pod kind in YAML | Skip gracefully |
| Nonexistent directory | `/nonexistent/path` | Return empty map |

**Real example from `kubelet_test.go`:**
```go
func TestParseKubeletLogs_MissingFile(t *testing.T) {
    bundleRoot, cleanup := createKubeletTestBundle(t)
    defer cleanup()
    // No journal file exists — just the empty dir
    issues, err := ParseKubeletLogs(bundleRoot)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if issues != nil {
        t.Errorf("expected nil for missing file, got %v", issues)
    }
}
```

### What CodeRabbit should flag:
- ❌ New analyzer without a "missing file" test
- ❌ New parser without an "empty input" test
- ❌ Functions accepting `interface{}` without nil test case

---

## Pattern 5: Substring Matching in Tests

⚠️ **This is a potential confusion point for CodeRabbit.**

Our `dmesg_more_test.go` tests **substring/contains matching** for pod correlation — a victim process named `"nginx"` matches a pod named `"nginx-deployment-abc123"`. This is **intentional bidirectional substring matching**, not sloppy assertions.

```go
{
    name: "exact substring match victim in pod name",
    kills: []DMesgOOMKill{{VictimName: "nginx", VictimPID: 100}},
    pods:  map[string]int{"nginx-deployment-abc123": 1},
    expected: map[string]bool{"nginx-deployment-abc123": true},
},
{
    name: "short victim name skipped (less than 3 chars)",
    kills: []DMesgOOMKill{{VictimName: "sh", VictimPID: 1}},
    pods:  map[string]int{"shell-runner-abc": 1},
    expected: map[string]bool{},  // "sh" is too short, skip it
},
```

**The correlation logic:**
- Victim names < 3 chars are skipped (too many false positives)
- Matching is case-insensitive
- Either direction: victim-in-pod OR pod-in-victim

### What CodeRabbit should NOT flag:
- ✅ Substring matching in `CorrelateWithPods` tests — this is the actual business logic
- ✅ Short name filtering (< 3 chars) — intentional false-positive prevention

---

## Pattern 6: Test Function Naming

```
Test<Function>_<Scenario>
```

**Integration tests** (use filesystem fixtures):
- `TestAnalyzeOOMEvents_NoEventsFile`
- `TestAnalyzeOOMEvents_DetectsOOMKilling`
- `TestParseKubeletLogs_HTTP502`
- `TestEnrichWithQoSClass_WithManifest`

**Unit tests** (pure function, no I/O):
- `TestParseOOMEvents_VariousPatterns` (table-driven)
- `TestIsHashLike` (table-driven)
- `TestSafeString` (table-driven)
- `TestParseJournaldLine` (table-driven)

**Separator `_` signals scenario grouping** — functions with many tests use it to group by category.

---

## Pattern 7: RecoverToError for Panic Safety

Functions that process untrusted input (support bundle data) use panic recovery:

```go
func TestRecoverToError(t *testing.T) {
    t.Run("with panic", func(t *testing.T) {
        fn := func() error { panic("something went wrong") }
        err := RecoverToError(fn, "test context")
        if err == nil {
            t.Error("expected error from panic recovery, got nil")
        }
    })
}
```

### What CodeRabbit should flag:
- ❌ New analyzer parsing untrusted YAML/JSON without panic recovery
- ❌ `RecoverToError` tests missing the non-string panic case

---

## Summary: CodeRabbit Review Checklist

When reviewing new test files, check for:

1. **[ ] Fixture cleanup** — Every `createXxxTestBundle` call has `defer cleanup()`
2. **[ ] t.Helper()** — All test helper functions call it
3. **[ ] Table-driven** — Multiple cases use `[]struct` + `t.Run`
4. **[ ] Fatal vs Error** — Length/error guards use `t.Fatalf`; field checks use `t.Errorf`
5. **[ ] Edge cases** — Missing file, empty input, nil, invalid format
6. **[ ] No panics** — Index access only after length guard with `t.Fatalf`
7. **[ ] Naming** — `Test<Function>_<Scenario>` format
8. **[ ] Graceful degradation** — Analyzers return empty results for missing data, not errors
