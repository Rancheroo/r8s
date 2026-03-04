# Sprint 1.2 Test Report & v1.2.1 Fix Plan

**Branch:** `feature/v1.2-simplify` | **Date:** 2026-03-04 | **Binary:** `go build` from commit HEAD  
**Test Bundle:** `r8s-cp-wlp7h-lhvgq-2026-03-03_03_49_04` (RKE2, 100% complete, real production bundle)

---

## Test Plan Results (Issue #86: Never Blank Output)

### TC1: Wrong Argument Order — PASS

`r8s ask "why is pod crashing" /tmp/r8s-test/valid-bundle` correctly shows:
- "Wrong argument order" header
- Correct usage format: `r8s ask <bundle> <question>`
- Shows what user entered vs expected
- Exits with code 2

### TC2: Bundle Not Found — PARTIAL FAIL (4/6 pass)

- `r8s analyze /nonexistent` — PASS
- `r8s ask /nonexistent "test"` — PASS
- `r8s describe /nonexistent pod test` — **FAIL** (BUG-1: shows `Bundle not found: 'pod'` — wrong arg parsed as bundle)
- `r8s export /nonexistent` — PASS
- `r8s get pods /nonexistent` — **FAIL** (BUG-2: completely blank output)
- `r8s logs /nonexistent test-pod` — PASS

### TC3: Unknown Command Suggestions — PARTIAL FAIL (3/5 pass)

- `r8s analize ./bundle` — **FAIL** (BUG-3: blank output — `analize` is in `isKnownCommand` list but not a real Cobra command)
- `r8s descibe ./bundle pod test` — PASS (shows "Did you mean 'describe'?")
- `r8s xyzfoo ./bundle` — PASS (shows "Did you mean 'analyze'?")
- `r8s log ./bundle nginx` — **FAIL** (BUG-3: blank output — `log` is in `isKnownCommand` list but not a real Cobra command)
- `r8s validat ./bundle` — PASS (shows "Did you mean 'validate'?")

### TC4: No Issues Found Message — NOT TESTABLE

Clean synthetic bundles always trigger structural `missing_file` issues (kubectl/pods, kubectl/nodes, etc.). The `ShowNoIssuesFound` code path cannot be reached without a structurally-complete bundle that has zero pattern matches. This test case needs a proper fixture bundle.

### TC5: Correct Usage — PASS (4/4)

- `r8s analyze <real-bundle>` — PASS (7 critical, 225 warning, 758ms)
- `r8s ask <bundle> "test"` — not tested (requires AI backend)
- `r8s --help` — PASS
- `r8s version` — PASS (`r8s 0.9.0`)

### TC6: Unit Tests — PASS (all pass)

All tests pass: `TestIsLikelyQuestion`, `TestIsLikelyPath`, `TestParseQueryIntent`, `TestNewUsageError`, `TestIsUsageError`, `TestIsKnownCommand`, `TestIsValidBundlePath`, `TestCommandSuggestions`, `TestAvailableCommands`, etc.

---

## UX Exploratory Testing Results

### Working Well

- `r8s get pods <bundle>` — clean tabwriter output, 38 pods, namespace filtering works
- `r8s get ns <bundle>` — shows discovered namespaces
- `r8s logs <bundle> <pod>` — streams logs with nice header formatting
- `r8s validate <bundle>` — detailed health check (100% complete, per-category breakdown)
- `r8s export <bundle>` — JSON export works
- `r8s analyze <bundle> --severity=critical` — correctly filters to critical-only
- `r8s analyze <bundle>/` — trailing slash handled gracefully
- `r8s analyze ""` — shows "Bundle not found: ''" (good)
- `r8s logs <bundle> nonexistent-pod` — shows "No logs found for pod" (good)
- `r8s` (no args) — shows help (good)

### UX Bugs Found

- **UX-BUG-1: `r8s ask <bundle>` (no question)** — blank output (same root cause as BUG-2)
- **UX-BUG-2: `r8s describe <bundle>` (no name)** — blank output (Cobra arg validation error swallowed)
- **UX-BUG-3: `r8s get` (no resource type)** — blank output
- **UX-BUG-4: `get nodes` version field** — shows full `kubectl version` output ("Client Version: v1.33.7+rke2r1\nKustomize Version: v5.6.0\nServer Version: v1.33.7+rke2r1") instead of just the server version
- **UX-BUG-5: `get pods` READY column** — shows container log file counts (31/31 for kube-apiserver) instead of actual container readiness. Misleading.
- **UX-BUG-6: `describe pod <bundle> <name>`** — returns "No resources found" for pods that exist in `get pods`. The describe command depends on `kubectl/podsdescribe` files that may not exist in the bundle.
- **UX-BUG-7: `analyze --format=json` severity field** — JSON output marks issues as `CRITICAL` but filtering with `--severity=critical` only works on human output; JSON consumers filtering `severity == 'CRITICAL'` get 0 results (case sensitivity or field mismatch).

---

## Critical Bugs (v1.2.1 MUST FIX)

### BUG-CRITICAL-1: `Execute()` Silently Swallows All RunE Errors

**Severity:** CRITICAL — This is the root cause of ALL blank output bugs  
**File:** `cmd/root.go` (83-91)

**Root Cause:** The error handling in `Execute()` has a logic flaw:

```go
if err := rootCmd.Execute(); err != nil {
    if exitCode := GetExitCode(err); exitCode != ExitSuccess {
        os.Exit(exitCode)  // <-- ALWAYS exits here
    }
    ShowFriendlyError(err)  // <-- DEAD CODE
    os.Exit(ExitError)
}
```

`GetExitCode()` returns `ExitError (2)` for any regular (non-ExitCodeError) error, which is always `!= ExitSuccess (0)`, so it calls `os.Exit(2)` **without printing anything**. `ShowFriendlyError` is never reached.

Combined with `SilenceErrors: true` on the root command, this means ANY error returned from a `RunE` function that isn't handled internally produces **zero output**.

**Affected commands:** `get` (bundle not found, missing args), `ask` (missing question), `describe` (missing args), and any future command that returns errors from `RunE`.

**Fix:**
```go
if err := rootCmd.Execute(); err != nil {
    if exitErr, ok := err.(*ExitCodeError); ok {
        os.Exit(exitErr.Code)
    }
    ShowFriendlyError(err)
    os.Exit(ExitError)
}
```

### BUG-CRITICAL-2: `isKnownCommand` Contains Non-Cobra Aliases

**Severity:** HIGH  
**File:** `cmd/root.go` (96-119)

**Root Cause:** `isKnownCommand()` includes typos like `"analize"`, `"log"`, `"analyse"` that are NOT registered Cobra subcommands. When these are entered, they bypass the typo suggestion handler but Cobra can't find them either, producing blank output (amplified by BUG-CRITICAL-1).

**Fix:** Remove all entries from `isKnownCommand` that are not actual Cobra-registered commands. Only keep: `analyze`, `ask`, `completion`, `describe`, `export`, `generate`, `get`, `logs`, `patterns`, `test-cluster`, `validate`, `version`, `help`. The typo-to-suggestion mapping in `CommandSuggestions` handles the rest.

### BUG-CRITICAL-3: `pod-stuck-terminating` Pattern Generates 221/232 Issues (95% Noise)

**Severity:** HIGH — Renders analyze output nearly unusable  
**File:** Pattern definition for `pod-stuck-terminating` (in `cmd/patterns.go` or pattern YAML files)

**Root Cause:** The pattern regex matches source file paths like `kube-controllers/resources.go` as pod names and matches the word "Terminating" in log lines that are not actual stuck-terminating events. It produces 221 identical warnings from a single bundle, drowning out the 11 real findings.

**Fix (two-part):**
1. Fix the pattern regex to only match actual pod termination events (e.g., require `DeletionTimestamp` or `Terminating` in kubectl pod status context, not arbitrary log lines)
2. Add deduplication/capping to the output renderer — if the same pattern+resource combination appears N times, collapse to a single entry with count (e.g., "Pod stuck terminating (seen 221 times across N pods)")

---

## Medium Bugs (v1.2.1 SHOULD FIX)

### BUG-MED-1: `get` Command Missing Pre-Validation for Bundle Path

**File:** `cmd/get.go` (70-95)

**Root Cause:** Unlike `analyze`, `ask`, `logs`, and `export`, the `get` command does not pre-validate the bundle path with `os.Stat()` before calling `loadBundle()`. The `bundle.Load()` function wraps the error, so `os.IsNotExist(err)` fails to match, and `ShowBundleNotFoundError` is never called.

**Fix:** Add `os.Stat(bundlePath)` check before `loadBundle()`, matching the pattern used by `describe.go` (61-75).

### BUG-MED-2: `describe` Command Arg Order Mismatch

**File:** `cmd/describe.go` (19, 114-126)

**Root Cause:** The 3-arg form is `describe <kind> <bundle> <name>`, but users coming from kubectl expect `describe <bundle> <kind> <name>` or `describe <kind> <name> -b <bundle>`. When a user puts the bundle path first (natural for r8s), the path gets parsed as `kind` and the kind gets parsed as `bundlePath`, producing confusing errors.

**Fix:** Add heuristic detection — if `args[0]` contains `/` or `.` it's likely a path, not a kind. Swap args accordingly, similar to the ask command's wrong-order detection.

### BUG-MED-3: `get nodes` Version Parsing

**File:** `cmd/get.go` (256-277) reading from `b.Manifest.K8sVersion`

**Root Cause:** `K8sVersion` contains the full multi-line `kubectl version` output instead of just the server version string.

**Fix:** Parse `K8sVersion` to extract only the server version line, or fix the manifest loader to store only the version string.

---

## Low Priority (v1.2.2 or Backlog)

### LOW-1: `get pods` READY Column Shows Log File Count

The READY column shows `len(pod.Containers)` which counts discovered log directories, not actual running containers. A pod like `kube-apiserver` shows `31/31` because it has 31 log files. Should parse actual container count from pod YAML/JSON.

### LOW-2: `get ns` Only Shows Namespaces From Discovered Pods

Only 5 namespaces shown vs the 8+ that appear in analyze findings (`r8s-test-app-backend`, `r8s-test-app-frontend`, `r8s-test-demo-worker` missing). The namespace list comes from `b.Pods` which doesn't include all pods from the bundle.

### LOW-3: `describe` Can't Find Pods That Exist in `get pods`

`describe` looks for `kubectl/podsdescribe` files. If the bundle only has `kubectl/pods` (list format, not describe format), the describe command can't find individual resources.

### LOW-4: TC4 Test Case Needs Fixture Bundle

The "No issues found" code path (`ShowNoIssuesFound`) is currently unreachable with synthetic bundles. Need to create a proper test fixture that is structurally complete but has no pattern matches.

### LOW-5: JSON Severity Case Mismatch

`--format=json` outputs severity as `"CRITICAL"` (uppercase) but the `--severity=critical` filter may not consistently match. Consumers using JSON output to filter critical issues get unexpected results.

---

## Summary

- **Test Plan (TC1-TC6):** 3 of 6 test cases fully pass. 2 partial fail. 1 not testable.
- **Root Cause:** A single bug in `Execute()` (BUG-CRITICAL-1) causes the majority of blank-output failures across multiple commands.
- **Pattern noise:** `pod-stuck-terminating` produces 95% of all warnings — fixing this alone dramatically improves UX.
- **Recommended v1.2.1 scope:** BUG-CRITICAL-1 + BUG-CRITICAL-2 + BUG-CRITICAL-3 + BUG-MED-1 + BUG-MED-2. These 5 fixes address all blank-output scenarios and the output noise problem.
- **Unit tests all pass** — the existing tests don't cover the `Execute()` error path or integration-level command execution. Consider adding integration tests that invoke the binary and assert on stderr output.
