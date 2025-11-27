# Verbose Error Handling - Test Plan

**Date**: 2025-11-27  
**Commit**: bbc8de5  
**Feature**: `--verbose` / `-v` flag for enhanced error messages  
**Status**: 🟢 READY TO TEST

---

## Executive Summary

**What's New**: Global `--verbose` flag that provides detailed, actionable error messages with:
1. **What happened** - The error itself
2. **Context** - File paths, URLs, values attempted
3. **Expected** - What should have been there
4. **Hint** - Actionable guidance to fix

**Why It Matters**: Makes testing and debugging 10x faster by showing exactly what went wrong and where.

---

## Test Categories

### Category A: Flag Integration (3 tests)
- Verify flag availability
- Test flag propagation
- Check flag persistence

### Category B: Bundle Error Enhancement (6 tests)
- Missing bundle path
- File not found
- Extraction failure
- Manifest parsing error
- Pod inventory failure
- Log inventory failure

### Category C: Error Quality (3 tests)
- Message structure validation
- Context completeness
- Actionable hints

### Category D: Performance & Regression (2 tests)
- No performance impact when disabled
- All existing tests still pass

**Total Tests**: 14

---

## Category A: Flag Integration

### ✅ Test A1: Flag Appears in Help (P0)
**Purpose**: Verify `--verbose` flag is documented

**Commands**:
```bash
./bin/r8s --help | grep -i verbose
./bin/r8s tui --help | grep -i verbose
./bin/r8s bundle --help | grep -i verbose
```

**Expected Output**:
```
  -v, --verbose            enable verbose error output for debugging
```

**Pass Criteria**:
- ✅ Flag shows in root help
- ✅ Flag shows in tui help
- ✅ Flag shows in bundle help
- ✅ Short form `-v` documented
- ✅ Long form `--verbose` documented

---

### ✅ Test A2: Flag Propagation (P0)
**Purpose**: Verify flag passes to TUI and bundle commands

**Test**: Code review of flag passing

**Files to Check**:
- `cmd/root.go` - Flag definition
- `cmd/tui.go` - Flag usage
- `internal/config/config.go` - Config field
- `internal/tui/app.go` - TUI propagation
- `internal/bundle/bundle.go` - Bundle usage

**Pass Criteria**:
- ✅ Flag defined as persistent in root.go
- ✅ Config.Verbose set in tui.go
- ✅ Verbose passed to NewBundleDataSource
- ✅ opts.Verbose checked in bundle.go error paths

---

### ✅ Test A3: Default Behavior (P0)
**Purpose**: Verify verbose is OFF by default (no breaking changes)

**Test**: Run commands without `--verbose` flag

**Commands**:
```bash
./bin/r8s tui --bundle=/nonexistent/path.tar.gz 2>&1 | wc -l
./bin/r8s tui --bundle=/nonexistent/path.tar.gz --verbose 2>&1 | wc -l
```

**Expected**:
- Non-verbose: Shorter error (1-2 lines)
- Verbose: Longer error (4-6 lines with context)

**Pass Criteria**:
- ✅ Default errors are concise (backward compatible)
- ✅ Verbose errors are detailed (new functionality)

---

## Category B: Bundle Error Enhancement

### ✅ Test B1: File Not Found Error (P0)
**Purpose**: Test enhanced error for missing bundle file

**Setup**: Use non-existent file path

**Commands**:
```bash
# Without verbose
./bin/r8s tui --bundle=/nonexistent/bundle.tar.gz 2>&1

# With verbose
./bin/r8s tui --bundle=/nonexistent/bundle.tar.gz --verbose 2>&1
```

**Expected WITHOUT verbose**:
```
Error: Failed to load bundle: bundle file not found: /nonexistent/bundle.tar.gz
```

**Expected WITH verbose**:
```
Error: Failed to load bundle: bundle file not found: /nonexistent/bundle.tar.gz
Current directory: /home/bradmin/github/r8s
Hint: Check the file path and ensure the file exists
```

**Pass Criteria**:
- ✅ Shows error in both modes
- ✅ Verbose adds current directory
- ✅ Verbose adds helpful hint
- ✅ Non-verbose is concise

---

### ✅ Test B2: Invalid Bundle Format (P0)
**Purpose**: Test error for corrupted/invalid bundle file

**Setup**: Create invalid .tar.gz file

**Commands**:
```bash
# Create fake bundle
echo "not a valid tarball" > /tmp/fake-bundle.tar.gz

# Test without verbose
./bin/r8s tui --bundle=/tmp/fake-bundle.tar.gz 2>&1

# Test with verbose
./bin/r8s tui --bundle=/tmp/fake-bundle.tar.gz --verbose 2>&1
```

**Expected WITH verbose**:
```
Error: Failed to load bundle: failed to extract bundle: gzip: invalid header
Bundle path: /tmp/fake-bundle.tar.gz
Hint: Ensure the file is a valid .tar.gz archive
```

**Pass Criteria**:
- ✅ Error detected in both modes
- ✅ Verbose shows bundle path
- ✅ Verbose suggests format check
- ✅ Hint is actionable

---

### ✅ Test B3: Missing Manifest (P1)
**Purpose**: Test error when bundle has no metadata.json

**Setup**: Create bundle without proper structure

**Commands**:
```bash
# Create minimal broken bundle
mkdir -p /tmp/broken-bundle/rke2
touch /tmp/broken-bundle/rke2/empty
cd /tmp && tar -czf broken-bundle.tar.gz broken-bundle/
cd /home/bradmin/github/r8s

# Test with verbose
./bin/r8s tui --bundle=/tmp/broken-bundle.tar.gz --verbose 2>&1
```

**Expected WITH verbose**:
```
Error: failed to parse bundle manifest: ...
Extract path: /tmp/r8s-bundle-xxxxx/
Expected: metadata.json with bundle info
Hint: This may not be a valid RKE2 support bundle
```

**Pass Criteria**:
- ✅ Shows extract path
- ✅ Explains expected format
- ✅ Provides useful hint

---

### ✅ Test B4: Missing Pod Logs (P1)
**Purpose**: Test error when pod log directory missing

**Setup**: Use bundle without pod logs directory

**Expected WITH verbose**:
```
Warning: Failed to parse ... (should be graceful)
Extract path: /tmp/r8s-bundle-xxxxx/
Searched: rke2/podlogs/ directory
Hint: Bundle may not contain pod logs
```

**Pass Criteria**:
- ✅ Shows what was searched
- ✅ Explains expected location
- ✅ Doesn't crash (graceful degradation)

---

### ✅ Test B5: Empty Bundle (P1)
**Purpose**: Test bundle with no resources

**Setup**: Use our test bundle from earlier

**Command**:
```bash
./bin/r8s tui --bundle=/tmp/r8s-bundle-1676042228 --verbose 2>&1
```

**Expected**:
- Bundle loads successfully (warnings logged)
- Shows extract path and searched locations
- TUI launches (no crash)

**Pass Criteria**:
- ✅ Verbose warnings show search paths
- ✅ Bundle still usable
- ✅ Clear about what's missing

---

### ✅ Test B6: Real Bundle with Verbose (P0)
**Purpose**: Test verbose mode with working bundle (regression)

**Command**:
```bash
./bin/r8s tui --bundle=/tmp/r8s-bundle-1208343738 --verbose 2>&1 | head -20
```

**Expected**:
- Bundle loads successfully
- Verbose may show parse warnings (expected)
- TUI launches normally

**Pass Criteria**:
- ✅ No false positives
- ✅ Real bundles work with --verbose
- ✅ Only shows actual issues

---

## Category C: Error Quality

### ✅ Test C1: Error Structure Validation (P0)
**Purpose**: Verify all verbose errors follow the pattern

**Error Pattern**:
```
1. What happened (the error message)
2. Context (file paths, directories, values)
3. Expected (what should be present)
4. Hint (actionable fix guidance)
```

**Test Method**: Code review of bundle.go error paths

**Files to Review**:
- `internal/bundle/bundle.go` - All verbose error blocks

**Pass Criteria**:
- ✅ All 6 enhanced errors follow pattern
- ✅ Context includes file paths
- ✅ Expected values are clear
- ✅ Hints are actionable

---

### ✅ Test C2: Context Completeness (P1)
**Purpose**: Verify context provides enough information

**Test**: Trigger each error and check if context is sufficient

**Questions to Answer**:
- Can you find the file being accessed?
- Can you understand what was expected?
- Can you reproduce the error?
- Can you fix it based on the hint?

**Pass Criteria**:
- ✅ File paths are absolute (not relative)
- ✅ Directory locations shown
- ✅ Search paths documented
- ✅ Expected format explained

---

### ✅ Test C3: Hint Actionability (P1)
**Purpose**: Verify hints are actually helpful

**Test**: Review all hints for actionability

**Good Hints**:
- "Check the file path and ensure the file exists"
- "Ensure the file is a valid .tar.gz archive"
- "This may not be a valid RKE2 support bundle"

**Bad Hints** (to avoid):
- "An error occurred"
- "Check your input"
- "Try again"

**Pass Criteria**:
- ✅ Hints suggest specific actions
- ✅ Hints are not generic
- ✅ Hints relate to the actual error

---

## Category D: Performance & Regression

### ✅ Test D1: No Performance Impact (P0)
**Purpose**: Verify --verbose doesn't slow down normal operations

**Test**: Compare execution time with/without verbose

**Commands**:
```bash
# Without verbose (baseline)
time ./bin/r8s bundle import --path=/tmp/r8s-bundle-1208343738 >/dev/null 2>&1

# With verbose
time ./bin/r8s bundle import --path=/tmp/r8s-bundle-1208343738 --verbose >/dev/null 2>&1
```

**Pass Criteria**:
- ✅ Time difference < 5% (within measurement error)
- ✅ No new allocations in hot paths
- ✅ Simple if-checks have zero cost

---

### ✅ Test D2: Regression Check (P0)
**Purpose**: Ensure all previous tests still pass

**Test**: Re-run critical tests from previous test plan

**Tests to Re-run**:
- Test C1: Empty resources (Bug #1)
- Test C2: Parse errors logged (Bug #2)
- Test C4: Real bundle resources
- Test B3: Bundle mode works

**Pass Criteria**:
- ✅ All previous tests still pass
- ✅ No new failures introduced
- ✅ Error messages improved (not broken)

---

## Testing Execution Plan

### Phase 1: Quick Validation (5 minutes)
1. Test A1 - Flag in help
2. Test A3 - Default behavior
3. Test B1 - File not found
4. Test B6 - Real bundle with verbose

**Goal**: Verify basic functionality works

### Phase 2: Error Coverage (10 minutes)
1. Test B2 - Invalid format
2. Test B3 - Missing manifest
3. Test B4 - Missing pod logs
4. Test B5 - Empty bundle

**Goal**: Verify all error paths enhanced

### Phase 3: Quality Check (5 minutes)
1. Test C1 - Error structure
2. Test C2 - Context completeness
3. Test C3 - Hint quality

**Goal**: Verify error quality

### Phase 4: Integration (5 minutes)
1. Test D1 - Performance
2. Test D2 - Regression

**Goal**: Ensure no breaking changes

**Total Time**: 25 minutes

---

## Success Criteria

### Critical (Must Pass)
- ✅ Flag available globally (A1)
- ✅ File not found shows context (B1)
- ✅ Real bundles work with verbose (B6)
- ✅ No regressions (D2)

### High Priority
- ✅ All error paths enhanced (B1-B6)
- ✅ Error structure consistent (C1)
- ✅ No performance impact (D1)

### Nice to Have
- ✅ Hints are excellent (C3)
- ✅ Context is complete (C2)

---

## Integration with Existing Tests

### Enhanced Test Cases

**From CLI_UX_TEST_PLAN.md**:

**Test C2** (Parse Errors) can now be enhanced:
```bash
# OLD TEST
./bin/r8s bundle import --path=malformed.tar.gz 2>&1 | grep -i warning

# NEW TEST WITH VERBOSE
./bin/r8s bundle import --path=malformed.tar.gz --verbose 2>&1
# Should show extract path, searched locations, and hints
```

**Test D3** (Bundle Path Validation) can now be enhanced:
```bash
# OLD TEST
./bin/r8s tui --bundle=/nonexistent/path.tar.gz

# NEW TEST WITH VERBOSE
./bin/r8s tui --bundle=/nonexistent/path.tar.gz --verbose
# Should show current directory and helpful hint
```

---

## Test Artifacts

### Files to Create
1. `/tmp/fake-bundle.tar.gz` - Invalid bundle format
2. `/tmp/broken-bundle.tar.gz` - Bundle without metadata
3. Test scripts for automated error checking

### Documentation to Update
1. CLI_UX_TEST_RESULTS.md - Add verbose error tests
2. TESTING_SUMMARY.md - Update with verbose testing
3. README.md - Document --verbose flag usage

---

## Expected Benefits

### For This Test Session
- Faster error diagnosis during testing
- Better understanding of bundle structure issues
- Immediate feedback on malformed bundles

### For Future Testing
- New test cases can use --verbose for debugging
- Error reproduction becomes easier
- Test failure analysis is faster

### For Production
- Users can troubleshoot their own issues
- Support tickets have better diagnostic info
- Bug reports are more actionable

---

## Risk Assessment

### Low Risk
- Flag is optional (default behavior unchanged)
- Simple if-checks (no performance impact)
- Enhanced errors are additive (not replacing)

### Medium Risk
- Error message changes may affect error parsing scripts
- Verbose output may be too long for some terminals

### Mitigation
- Keep non-verbose errors backward compatible
- Document verbose flag in troubleshooting guide
- Ensure verbose errors fit typical terminal width

---

**Status**: 🟢 READY TO EXECUTE  
**Estimated Duration**: 25 minutes  
**Priority Tests**: A1, A3, B1, B6, D2 (10 minutes for critical path)

---

## Quick Reference

### Test Commands

```bash
# Check flag availability
./bin/r8s tui --help | grep verbose

# Test file not found (verbose)
./bin/r8s tui --bundle=/nonexistent.tar.gz --verbose

# Test invalid format (verbose)
echo "fake" > /tmp/fake.tar.gz
./bin/r8s tui --bundle=/tmp/fake.tar.gz --verbose

# Test real bundle (verbose)
./bin/r8s tui --bundle=/tmp/r8s-bundle-1208343738 --verbose

# Regression test
./bin/r8s bundle import --path=/tmp/r8s-bundle-1208343738 --verbose
```

---

**Testing Ready!** Let's execute and verify this powerful debugging feature! 🚀
