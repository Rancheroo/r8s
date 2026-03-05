# r8s v1.2.1 Manual Test Plan
**Branch:** `feature/v1.2-simplify`  
**Commit:** `c8459d8`  
**Date:** March 5, 2026  
**Tester:** _______________

---

## Quick Start

```bash
# Build
cd /home/bradmin/.openclaw/workspace/r8s
go build .

# Verify version
./r8s version
```

---

## Part 1: Critical Bug Fixes (Must Pass)

### CRITICAL-1: Execute() Error Display ✅
**Test:** Error messages now show before exit

| # | Command | Expected Result | Actual | Pass |
|---|---------|-----------------|--------|------|
| 1.1 | `./r8s get pods /nonexistent/path` | Shows "Bundle not found" error message + exits with code 2 | | |
| 1.2 | `./r8s ask ./bundle` (no question) | Shows usage error with helpful suggestion | | |
| 1.3 | `./r8s analize ./bundle` | Shows "did you mean 'analyze'?" suggestion | | |
| 1.4 | `./r8s describe pod my-pod ./bad-path` | Shows bundle path error with friendly message | | |

**Edge Cases:**
- [ ] Empty command: `./r8s` 
- [ ] Completely unknown command: `./r8s foobar`
- [ ] Valid command, missing all args: `./r8s get`

---

### CRITICAL-2: Fake Command Cleanup ✅
**Test:** Typos are suggestions, not known commands

| # | Command | Expected Result | Actual | Pass |
|---|---------|-----------------|--------|------|
| 2.1 | `./r8s analize --help` | Shows "unknown command 'analize'" with suggestion | | |
| 2.2 | `./r8s log ./bundle` | Shows "unknown command 'log'" with suggestion | | |
| 2.3 | `./r8s analyse ./bundle` | Shows "unknown command 'analyse'" with suggestion | | |
| 2.4 | `./r8s logs --help` | Shows logs command help (valid command) | | |

**UX Hunt:** Try these typos—are suggestions helpful?
- `analiz`, `analise`, `analyse`
- `validat`, `validte`
- `descibe`, `describ`, `dscribe`

---

### CRITICAL-3: Pod-Stuck-Terminating Regex ✅
**Test:** No false positives on source files

| # | Test Input | Expected Match | Actual | Pass |
|---|------------|----------------|--------|------|
| 3.1 | `kube-controllers/resources.go some stuff` | NO MATCH (source file) | | |
| 3.2 | `default my-pod 1/1 Terminating 5m` | MATCH (real pod) | | |
| 3.3 | `kube-system calico-node-x1abc 1/1 Terminating 12d` | MATCH (real pod with days) | | |
| 3.4 | `cattle-system rancher-xxx 2/2 Terminating 3h30m` | MATCH (complex duration) | | |

**Regression Test:** Create temp file with fake pod output:
```bash
echo "kube-controllers/resources.go 1/1 Terminating 5m" > /tmp/fake.txt
./r8s analyze /tmp/fake.txt 2>&1 | grep -i "terminating" || echo "GOOD: No false positive"
```

---

### MED-1: Get Command Bundle Validation ✅
**Test:** Early validation with clear error

| # | Command | Expected Result | Actual | Pass |
|---|---------|-----------------|--------|------|
| 4.1 | `./r8s get pods /nonexistent` | "Bundle not found: /nonexistent" + friendly error | | |
| 4.2 | `./r8s get pods ./valid-bundle` | Works normally | | |
| 4.3 | `./r8s get pods ./file.txt` | "Bundle not found" (file exists but not bundle) | | |
| 4.4 | `./r8s get pods` | "bundle path required" error | | |

---

### MED-2: Describe Arg Ordering ✅
**Test:** Path vs resource type detection

| # | Command | Expected Result | Actual | Pass |
|---|---------|-----------------|--------|------|
| 5.1 | `./r8s describe pod my-pod ./bundle` | Works (normal order) | | |
| 5.2 | `./r8s describe ./bundle pod my-pod` | Works (swapped order detected) | | |
| 5.3 | `./r8s describe helmcharts.helm.cattle.io ./bundle my-chart` | Works (CRD with dots) | | |
| 5.4 | `./r8s describe ./bundle` | Error: "resource type and name required" | | |

**Edge Case:** Path with dots in directory name:
```bash
./r8s describe pod my-pod ./v1.2.3/bundle
```

---

## Part 2: UX Hunt (Open-Ended Exploration)

### Challenge 1: Error Message Quality
Find 3 scenarios where error messages could be better:
1. _________________________________
2. _________________________________
3. _________________________________

### Challenge 2: Command Discovery
Can you figure out how to:
- [ ] List all available commands without running `--help` on each?
- [ ] Find the correct syntax for `r8s ask` if you forget it?
- [ ] Get suggestions for a command you've partially typed?

### Challenge 3: Edge Case Inputs
Try these weird inputs and note behavior:
- [ ] Bundle path that's a symlink: `ln -s ./real-bundle ./link-bundle && ./r8s get pods ./link-bundle`
- [ ] Bundle path with spaces: `./r8s get pods "./my bundle"`
- [ ] Very long pod name: create a pod name with 100+ characters
- [ ] Unicode in resource names: `./r8s describe pod "测试-pod" ./bundle`

### Challenge 4: Consistency Hunt
Find inconsistencies between commands:
1. Does `get` behave differently than `describe` for missing bundles?
2. Are error message formats consistent across commands?
3. Do all commands support the same flag styles (-n vs --namespace)?

### Challenge 5: Pattern Detection Stress Test
Create a bundle with these edge cases and run `analyze`:
```bash
# Create test scenarios
echo "default my-pod 0/1 Terminating 999d" > /tmp/old-terminating.txt
echo "kube-system dns-autoscaler 1/1 Terminating 1s" > /tmp/one-second.txt
./r8s analyze /path/to/bundle 2>&1 | head -50
```
- [ ] Does it catch very old terminating pods?
- [ ] Does it handle sub-second durations?
- [ ] Are the matches accurate or noisy?

---

## Part 3: Integration Scenarios

### Scenario A: Full Workflow
1. Start with no bundle: `./r8s get pods ./missing`
2. Use suggestion to find correct command
3. Analyze a real bundle: `./r8s analyze ./real-bundle`
4. Get specific resource: `./r8s describe pod <name> ./real-bundle`
5. Export findings: `./r8s export --format sarif ./real-bundle`

**Track:** Any friction points or confusing moments

### Scenario B: Typo Recovery
1. Type `./r8s analize ./bundle` (typo)
2. Read error message—does it help you recover?
3. Type `./r8s analyze ./bundle` (correct)
4. Did you learn the correct command from the error?

### Scenario C: Bundle Path Confusion
1. `./r8s describe ./bundle pod my-pod` (wrong order)
2. Does it work? Is the error message helpful?
3. `./r8s describe pod my-pod ./bundle` (correct order)
4. Compare UX between attempts

---

## Part 4: Performance & Resource Checks

```bash
# Time basic commands
time ./r8s version
time ./r8s get pods ./bundle
time ./r8s analyze ./bundle

# Check binary size
ls -lh ./r8s

# Memory usage (if possible)
/usr/bin/time -v ./r8s analyze ./bundle 2>&1 | grep -E "Maximum resident|User time"
```

---

## Bug Report Template

Found an issue? Document it:

```
**Command:** 
**Expected:** 
**Actual:** 
**Severity:** (Critical/High/Medium/Low/UX)
**Suggestion:** 
```

---

## Sign-Off

| Check | Result |
|-------|--------|
| All Part 1 tests pass | ⬜ Yes ⬜ No |
| No regressions found | ⬜ Yes ⬜ No |
| UX issues documented | ⬜ Yes ⬜ No |
| Ready for CodeRabbit | ⬜ Yes ⬜ No |

**Notes:**
_______________________________________________
_______________________________________________
_______________________________________________

**Tester Signature:** _______________  **Date:** _______________
