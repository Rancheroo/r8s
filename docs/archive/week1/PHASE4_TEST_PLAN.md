# Phase 4: Bundle Import - Test Plan

**Test Phase Date:** November 27, 2025  
**Phase:** Phase 4 - Bundle Import Core  
**Tester:** Automated + Manual  
**Priority:** P0 (Must complete before Phase 5)

---

## Test Objectives

1. Verify bundle extraction works correctly
2. Validate size limit enforcement
3. Test error handling and edge cases
4. Confirm security validations
5. Verify metadata parsing accuracy
6. Test bundle format detection
7. Ensure cleanup works properly

---

## Test Environment

**Bundle Available:**
- `example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz` (8.93 MB)

**Test Data Needed:**
- Small test bundle (<1MB) - will create
- Large test bundle (>10MB) - use existing with low limit
- Invalid/corrupted bundle
- Non-RKE2 bundle format

---

## Test Categories

### Category A: Happy Path Tests (P0)

#### A1: Basic Import
**Objective:** Verify basic bundle import works

**Steps:**
1. Run `./bin/r8s bundle import -p example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz -l 100`
2. Observe output

**Expected:**
- ✅ Extraction successful
- ✅ Bundle metadata displayed
- ✅ Node name: w-guard-wg-cp-svtk6-lqtxw
- ✅ RKE2 version shown
- ✅ K8s version shown
- ✅ File count: 319
- ✅ No errors

**Status:** ⏳ Not run

---

#### A2: Default Size Limit
**Objective:** Verify default 10MB limit is applied

**Steps:**
1. Run `./bin/r8s bundle import -p example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz`
2. Note: Bundle is 8.93MB, should pass

**Expected:**
- ✅ Import succeeds with warning about default 10MB limit
- ✅ No size errors

**Status:** ⏳ Not run

---

#### A3: Custom Size Limit (Higher)
**Objective:** Verify custom size limit parameter works

**Steps:**
1. Run `./bin/r8s bundle import -p example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz -l 50`

**Expected:**
- ✅ Import succeeds
- ✅ Shows "Size limit: 50MB"

**Status:** ⏳ Not run

---

### Category B: Error Handling Tests (P0)

#### B1: Missing File
**Objective:** Verify error when bundle doesn't exist

**Steps:**
1. Run `./bin/r8s bundle import -p nonexistent.tar.gz`

**Expected:**
- ❌ Clear error: "bundle file not found: nonexistent.tar.gz"
- ❌ No panic or crash
- ❌ No temp directory created

**Status:** ⏳ Not run

---

#### B2: Size Limit Exceeded
**Objective:** Verify size limit is enforced

**Steps:**
1. Run `./bin/r8s bundle import -p example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz -l 5`
2. Note: Bundle is 8.93MB, limit is 5MB

**Expected:**
- ❌ Clear error: "bundle file size (X bytes) exceeds limit (Y bytes)"
- ❌ No extraction performed
- ❌ No temp directory created
- ❌ Clean exit

**Status:** ⏳ Not run

---

#### B3: Invalid Bundle Format
**Objective:** Verify unknown format detection

**Steps:**
1. Create a tar.gz with random files (not RKE2 structure)
2. Run import command

**Expected:**
- ❌ Clear error: "unknown bundle format"
- ❌ Temp directory cleaned up
- ❌ No crash

**Status:** ⏳ Not run

---

#### B4: Missing Required Flag
**Objective:** Verify --path is required

**Steps:**
1. Run `./bin/r8s bundle import`

**Expected:**
- ❌ Error: "required flag(s) "path" not set"
- ❌ Shows usage help

**Status:** ⏳ Not run

---

### Category C: Edge Cases (P1)

#### C1: Help Command
**Objective:** Verify help works

**Steps:**
1. Run `./bin/r8s bundle --help`
2. Run `./bin/r8s bundle import --help`

**Expected:**
- ✅ Shows bundle subcommands
- ✅ Shows import usage and examples
- ✅ Shows all flags

**Status:** ⏳ Not run

---

#### C2: Zero Size Limit
**Objective:** Verify 0 triggers default

**Steps:**
1. Run `./bin/r8s bundle import -p example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz -l 0`

**Expected:**
- ✅ Uses default 10MB limit
- ✅ Import succeeds

**Status:** ⏳ Not run

---

#### C3: Negative Size Limit
**Objective:** Verify negative values handled

**Steps:**
1. Run `./bin/r8s bundle import -p example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz -l -10`

**Expected:**
- ❌ Error or uses default
- ❌ No crash

**Status:** ⏳ Not run

---

#### C4: Very Large Size Limit
**Objective:** Verify large limits work

**Steps:**
1. Run `./bin/r8s bundle import -p example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz -l 1000`

**Expected:**
- ✅ Import succeeds
- ✅ Shows "Size limit: 1000MB"

**Status:** ⏳ Not run

---

### Category D: Security Tests (P0)

#### D1: Path Traversal Protection
**Objective:** Verify path traversal is blocked

**Steps:**
1. Create a malicious tar.gz with paths like `../../etc/passwd`
2. Attempt import

**Expected:**
- ❌ Error: "invalid file path in bundle"
- ❌ No files extracted outside temp dir
- ❌ Cleanup performed

**Status:** ⏳ Not run (requires malicious bundle creation)

---

#### D2: Symlink Handling
**Objective:** Verify symlinks handled safely

**Steps:**
1. Bundle contains symlinks (already present in example bundle)
2. Import bundle

**Expected:**
- ✅ Import succeeds
- ✅ Symlinks skipped or handled safely
- ✅ No errors

**Status:** ⏳ Not run

---

### Category E: Cleanup Tests (P1)

#### E1: Successful Import Cleanup
**Objective:** Verify temp directory exists during import

**Steps:**
1. Import bundle
2. Note extraction location from output
3. Check if directory exists

**Expected:**
- ✅ Temp directory exists at shown location
- ✅ Contains extracted files
- ✅ Directory follows pattern `/tmp/r8s-bundle-*`

**Status:** ⏳ Not run

---

#### E2: Failed Import Cleanup
**Objective:** Verify cleanup on error

**Steps:**
1. Trigger an error (size limit exceeded)
2. Check for leftover temp directories

**Expected:**
- ✅ No temp directories left behind
- ✅ `ls /tmp/r8s-bundle-*` shows nothing or old ones only

**Status:** ⏳ Not run

---

### Category F: Metadata Tests (P1)

#### F1: Version Parsing
**Objective:** Verify version extraction accuracy

**Steps:**
1. Import bundle
2. Verify versions match known values

**Expected:**
- ✅ RKE2 Version: v1.32.7+rke2r1
- ✅ K8s Version: v1.32.7+rke2r1
- ✅ Versions shown correctly

**Status:** ⏳ Not run

---

#### F2: File Count Accuracy
**Objective:** Verify file counting is correct

**Steps:**
1. Import bundle
2. Compare file count to actual
3. Manual check: `tar -tzf bundle.tar.gz | wc -l`

**Expected:**
- ✅ File count matches
- ✅ Count includes all non-directory entries

**Status:** ⏳ Not run

---

#### F3: Size Calculation
**Objective:** Verify size reporting

**Steps:**
1. Import bundle
2. Check reported size
3. Compare to: `ls -lh bundle.tar.gz`

**Expected:**
- ✅ Size matches file size
- ✅ Format is MB with 2 decimals

**Status:** ⏳ Not run

---

### Category G: Output Format Tests (P2)

#### G1: Output Formatting
**Objective:** Verify output is well-formatted

**Steps:**
1. Import bundle
2. Review output formatting

**Expected:**
- ✅ Separator lines render correctly
- ✅ Columns aligned
- ✅ No garbled characters
- ✅ Namespace grouping works
- ✅ Log type breakdown shown

**Status:** ⏳ Not run

---

#### G2: Empty Sections
**Objective:** Verify output when sections are empty

**Steps:**
1. Import bundle with no pods (current example)
2. Check pod summary section

**Expected:**
- ✅ Section omitted or shows "0 pods"
- ✅ No errors
- ✅ Other sections still shown

**Status:** ⏳ Not run

---

### Category H: Performance Tests (P2)

#### H1: Import Speed
**Objective:** Verify extraction is fast

**Steps:**
1. Time the import: `time ./bin/r8s bundle import -p bundle.tar.gz -l 100`

**Expected:**
- ✅ Completes in <5 seconds for 8.93MB bundle
- ✅ Most time in extraction, not parsing

**Status:** ⏳ Not run

---

#### H2: Memory Usage
**Objective:** Verify low memory usage

**Steps:**
1. Monitor memory during import
2. Check for leaks after multiple imports

**Expected:**
- ✅ <100MB memory for import
- ✅ No memory leaks
- ✅ Memory released after completion

**Status:** ⏳ Not run (requires monitoring tool)

---

### Category I: Integration Tests (P1)

#### I1: Build Test
**Objective:** Verify clean build

**Steps:**
1. Run `go build -o bin/r8s`

**Expected:**
- ✅ Build succeeds
- ✅ No errors
- ✅ Only GOPATH warning (unrelated)
- ✅ Binary created

**Status:** ⏳ Not run

---

#### I2: Existing Features Unchanged
**Objective:** Verify no breaking changes

**Steps:**
1. Run `./bin/r8s` (TUI mode)
2. Test navigation
3. Test log viewing

**Expected:**
- ✅ TUI starts normally
- ✅ All navigation works
- ✅ Log viewing works
- ✅ No regressions

**Status:** ⏳ Not run

---

## Test Execution Plan

### Phase 1: P0 Tests (Must Pass)
1. Run all Category A tests (Happy Path)
2. Run all Category B tests (Error Handling)
3. Run all Category D tests (Security)

**Stop Criteria:** Any P0 test failure

### Phase 2: P1 Tests (Should Pass)
1. Run all Category C tests (Edge Cases)
2. Run all Category E tests (Cleanup)
3. Run all Category F tests (Metadata)
4. Run all Category I tests (Integration)

**Stop Criteria:** >2 P1 test failures

### Phase 3: P2 Tests (Nice to Have)
1. Run all Category G tests (Output Format)
2. Run all Category H tests (Performance)

**Stop Criteria:** None (informational)

---

## Test Data Preparation

### Create Small Test Bundle
```bash
# Create a minimal RKE2-like structure
mkdir -p test-bundle-small/rke2
echo "v1.28.0" > test-bundle-small/rke2/version
tar -czf test-bundle-small.tar.gz test-bundle-small/
```

### Create Invalid Bundle
```bash
# Create a bundle without RKE2 structure
mkdir -p test-bundle-invalid
echo "random" > test-bundle-invalid/file.txt
tar -czf test-bundle-invalid.tar.gz test-bundle-invalid/
```

### Create Path-Traversal Bundle (Optional)
```bash
# Only for security testing - DO NOT USE IN PROD
mkdir -p test-bundle-malicious
(cd test-bundle-malicious && ln -s ../../../etc/passwd bad-link)
tar -czf test-bundle-malicious.tar.gz test-bundle-malicious/
```

---

## Success Criteria

**Phase 4 Testing Complete When:**
- ✅ All P0 tests pass (Category A, B, D)
- ✅ ≥90% P1 tests pass (Category C, E, F, I)
- ✅ ≥70% P2 tests pass (Category G, H)
- ✅ No critical bugs found
- ✅ No regressions in existing features

**If Criteria Not Met:**
- Fix bugs
- Re-run failed tests
- Update documentation with known issues

---

## Bug Reporting Template

```markdown
**Bug ID:** PHASE4-BXX
**Priority:** P0/P1/P2
**Category:** A/B/C/D/E/F/G/H/I
**Test Case:** Test ID

**Description:**
[What went wrong]

**Steps to Reproduce:**
1. [Step 1]
2. [Step 2]

**Expected:**
[What should happen]

**Actual:**
[What actually happened]

**Impact:**
[How this affects users/system]

**Fix Priority:**
[Must fix before Phase 5 / Can defer]
```

---

## Test Report Template

Will be generated as: `PHASE4_TEST_REPORT.md`

Sections:
1. Executive Summary
2. Test Execution Results
3. Bugs Found
4. Performance Metrics
5. Recommendations
6. Sign-off

---

## Estimated Effort

- **Test Data Preparation:** 5 minutes
- **P0 Tests:** 15 minutes
- **P1 Tests:** 20 minutes
- **P2 Tests:** 10 minutes
- **Bug Fixes:** Variable (if needed)
- **Report Writing:** 10 minutes

**Total:** ~60 minutes (assuming minimal bugs)

---

## Next Steps

1. Review this test plan
2. Create test data bundles
3. Execute P0 tests
4. Execute P1 tests
5. Execute P2 tests
6. Generate test report
7. Fix any critical bugs
8. Sign off on Phase 4
9. Proceed to Phase 5

---

**Test Plan Status:** Ready for Execution  
**Dependencies:** None  
**Blocking:** Phase 5 start
