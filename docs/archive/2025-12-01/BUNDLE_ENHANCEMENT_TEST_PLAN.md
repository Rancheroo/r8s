# Bundle Loading Enhancement - Testing Plan

**Date:** November 28, 2025  
**Version:** Post-fixes (commit 9292892)  
**Test Focus:** Dual-mode bundle loading (archives + directories) + Bug fixes

---

## What Changed Since Last Testing

### Bugs Fixed (Commits 3814049, a249562)
1. **BUG-001**: CRD version selection - Fixed to check `served: true`
2. **BUG-002**: Nil pointer crashes in mock mode describe functions
3. **BUG-003**: Bundle kubectl path resolution (getBundleRoot usage)

### New Feature (Commit 9292892)
- **Dual-mode bundle loading**: Support both compressed archives AND extracted directories
- **Auto-detection**: System automatically detects input type
- **Enhanced validation**: Comprehensive error messages
- **Smart cleanup**: Only temp extractions cleaned, user directories preserved

---

## Testing Objectives

### Primary Goals
1. ✅ Verify all 3 critical bugs are fixed
2. 🆕 Test new dual-mode bundle loading (archive + directory)
3. 🆕 Validate auto-detection logic
4. 🆕 Test error handling and validation
5. ✅ Verify previously passing tests still work
6. 🆕 Test bundle TUI mode (was completely broken, should work now)

### Success Criteria
- All previously found bugs resolved
- Archive mode works correctly
- Directory mode works correctly
- Auto-detection accurate
- Error messages clear and actionable
- TUI launches successfully in bundle mode
- All resource views populate with correct data

---

## Test Environment

- **OS:** Linux/Ubuntu
- **Terminal:** bash 5.2.21
- **r8s Version:** Latest (commit 9292892)
- **Test Bundle:** `example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz`
- **Bundle Contents:** 337 files, 86 pods, 176 logs, 33 kubectl resources

---

## Test Suite

### Phase 1: Bug Fix Verification ✅ CRITICAL

#### Test 1.1: BUG-003 Fix - Bundle kubectl Path Resolution
**Status:** Previously broken, should now work

**Test Steps:**
```bash
# Test archive mode
./bin/r8s bundle import --path=example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz --limit=100 --verbose
```

**Expected Results:**
- ✅ No warnings about missing kubectl files
- ✅ CRDs parsed successfully (count > 0)
- ✅ Deployments parsed successfully (count > 0)
- ✅ Services parsed successfully (count > 0)
- ✅ Namespaces parsed successfully (count > 0)

**Previous Behavior:**
- ❌ Warnings: "Failed to parse CRDs/Deployments/Services/Namespaces"
- ❌ All kubectl resource counts = 0

#### Test 1.2: BUG-003 Fix - Bundle TUI Launch
**Status:** Previously crashed, should now work

**Test Steps:**
```bash
# Launch TUI in bundle mode
./bin/r8s tui --bundle=example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz
```

**Expected Results:**
- ✅ TUI launches successfully (no "client not initialized" error)
- ✅ Cluster view displays
- ✅ Can navigate to resource views
- ✅ Resource counts shown correctly

**Previous Behavior:**
- ❌ Crash with "client not initialized" error

#### Test 1.3: BUG-002 Fix - Mock Mode Describe
**Status:** Previously crashed, should now work

**Test Steps:**
```bash
# Launch TUI in mock mode
./bin/r8s tui --mockdata

# Navigate to Pods view
# Select a pod
# Press 'd' to describe
```

**Expected Results:**
- ✅ Describe modal opens (no crash)
- ✅ Shows JSON or message indicating mock mode
- ✅ Can close modal with Esc/d/q

**Previous Behavior:**
- ❌ Nil pointer dereference crash

#### Test 1.4: BUG-001 Fix - CRD Version Selection
**Status:** Code analysis bug, should verify with real data

**Test Method:** Code review + bundle data inspection

**Test Steps:**
```bash
# Launch bundle TUI
./bin/r8s tui --bundle=example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz

# Navigate to CRDs (C key)
# Select a CRD
# Press Enter to view instances
```

**Expected Results:**
- ✅ No 404 errors
- ✅ CRD instances load successfully (or show appropriate empty message)

**Note:** May need live Rancher instance to fully test

---

### Phase 2: New Feature - Dual-Mode Bundle Loading 🆕

#### Test 2.1: Archive Mode (Existing, Should Still Work)
**Status:** Should work as before but with kubectl parsing fixed

**Test Steps:**
```bash
# Test with archive file
./bin/r8s bundle import --path=example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz --limit=100 --verbose
```

**Expected Results:**
- ✅ Detects as archive (📦 icon in output)
- ✅ Extracts to /tmp/r8s-bundle-*
- ✅ Shows extraction progress
- ✅ Parses all resources correctly
- ✅ Shows: "86 pods, 176 logs, 29 deployments, 37 services, 96 CRDs, 17 namespaces"
- ✅ Bundle.IsTemporary = true

**Success Criteria:**
- No warnings about missing kubectl files (BUG-003 fix)
- All resource counts > 0

#### Test 2.2: Directory Mode (NEW Feature)
**Status:** Brand new functionality

**Test Steps:**
```bash
# First, extract bundle manually
tar -xzf example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz -C /tmp/

# Test with extracted directory
./bin/r8s bundle import --path=/tmp/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09/ --verbose
```

**Expected Results:**
- ✅ Detects as directory (📁 icon in output)
- ✅ NO extraction step
- ✅ Instant load (no delays)
- ✅ Parses all resources correctly
- ✅ Same resource counts as archive mode
- ✅ Bundle.IsTemporary = false

**Success Criteria:**
- Faster than archive mode (no extraction)
- No size limits applied
- Directory preserved after exit

#### Test 2.3: Auto-Detection - Archive vs Directory
**Status:** Core new feature

**Test Cases:**
```bash
# Test 1: .tar.gz archive
./bin/r8s bundle import --path=bundle.tar.gz
# Expected: Archive mode

# Test 2: .tgz archive
./bin/r8s bundle import --path=bundle.tgz
# Expected: Archive mode

# Test 3: Extracted directory
./bin/r8s bundle import --path=bundle-dir/
# Expected: Directory mode

# Test 4: Nested directory (with node name)
./bin/r8s bundle import --path=/tmp/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09/
# Expected: Directory mode
```

**Expected Results:**
- ✅ Correct mode detected for each input type
- ✅ Appropriate icon shown (📦 vs 📁)
- ✅ Correct processing path taken

#### Test 2.4: Cleanup Behavior
**Status:** Critical for safety

**Test Steps:**
```bash
# Test 1: Archive mode cleanup
./bin/r8s bundle import --path=bundle.tar.gz --verbose
ls /tmp/r8s-bundle-*  # Note the path
# Exit r8s
ls /tmp/r8s-bundle-*  # Should be gone

# Test 2: Directory mode no cleanup
./bin/r8s bundle import --path=/tmp/extracted-bundle/ --verbose
ls /tmp/extracted-bundle/  # Should exist
# Exit r8s
ls /tmp/extracted-bundle/  # Should still exist
```

**Expected Results:**
- ✅ Archive extractions cleaned up
- ✅ User directories preserved
- ✅ No accidental deletions

---

### Phase 3: Error Handling & Validation 🆕

#### Test 3.1: Path Not Found
**Test Steps:**
```bash
./bin/r8s bundle import --path=/nonexistent/bundle.tar.gz --verbose
```

**Expected Error:**
```
❌ path not found: /nonexistent/bundle.tar.gz

Current directory: /home/user/r8s
Absolute path tried: /nonexistent/bundle.tar.gz

TROUBLESHOOTING:
  1. Check the path is correct
  2. Ensure file/folder exists
  3. Check file permissions
  4. Try using an absolute path
```

#### Test 3.2: Invalid Directory Structure
**Test Steps:**
```bash
mkdir /tmp/not-a-bundle
./bin/r8s bundle import --path=/tmp/not-a-bundle/ --verbose
```

**Expected Error:**
```
❌ invalid bundle directory: missing rke2/ directory

Path checked: /tmp/not-a-bundle/rke2

EXPECTED STRUCTURE:
  bundle-folder/
    ├── rke2/
    │   ├── kubectl/
    │   ├── podlogs/
    │   └── ...
    └── metadata.json

HINT: This folder doesn't appear to be an extracted RKE2 support bundle
```

#### Test 3.3: Unsupported Archive Format
**Test Steps:**
```bash
# Create dummy .zip file
touch /tmp/bundle.zip
./bin/r8s bundle import --path=/tmp/bundle.zip --verbose
```

**Expected Error:**
```
❌ unsupported archive format: .zip

Supported formats:
  • .tar.gz  (RKE2 support bundles)
  • .tgz     (compressed tar)

Current file: bundle.zip

SOLUTIONS:
  1. If bundle is already extracted, point to the folder:
     r8s --bundle=/path/to/extracted-folder/
  2. If you have a different archive format, extract it first
  3. Ensure the file extension is preserved
```

#### Test 3.4: Size Limit Exceeded
**Test Steps:**
```bash
# Test with very low limit
./bin/r8s bundle import --path=example-log-bundle/*.tar.gz --limit=1 --verbose
```

**Expected Error:**
```
❌ bundle uncompressed size (XX.X MB) exceeds limit (1.0 MB)

The bundle is too large for the current size limit.

SOLUTION:
  Increase the limit with --limit flag:
  r8s bundle import --path=bundle.tar.gz --limit=50

ALTERNATIVE:
  Extract manually and use folder mode:
  $ tar -xzf bundle.tar.gz
  $ r8s bundle=./extracted-folder/
```

---

### Phase 4: Bundle TUI Integration Testing 🆕

#### Test 4.1: Launch TUI with Archive
**Test Steps:**
```bash
./bin/r8s tui --bundle=example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz
```

**Expected Results:**
- ✅ TUI launches successfully
- ✅ Shows cluster view
- ✅ Bundle mode indicator visible
- ✅ Can navigate views

#### Test 4.2: Launch TUI with Directory
**Test Steps:**
```bash
# Extract first
tar -xzf example-log-bundle/*.tar.gz -C /tmp/

# Launch with directory
./bin/r8s tui --bundle=/tmp/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09/
```

**Expected Results:**
- ✅ TUI launches successfully (faster than archive)
- ✅ Same functionality as archive mode
- ✅ Directory not deleted on exit

#### Test 4.3: Resource View Navigation
**Test Steps:**
```bash
# Launch bundle TUI
./bin/r8s tui --bundle=<path>

# Navigate through views:
# 1. Clusters → Enter
# 2. Projects → Enter  
# 3. Namespaces → Enter
# 4. Pods view (default)
# 5. Deployments (2 key)
# 6. Services (3 key)
# 7. CRDs (C key)
```

**Expected Results:**
- ✅ All views load correctly
- ✅ Resource counts accurate
- ✅ Tables show data
- ✅ Navigation smooth
- ✅ No crashes

#### Test 4.4: Resource Actions in Bundle Mode
**Test Steps:**
```bash
# Launch bundle TUI
# Navigate to Pods view
# Select a pod
# Press 'd' for describe
```

**Expected Results:**
- ✅ Describe modal opens
- ✅ Shows pod details (from bundle data)
- ✅ Can close modal
- ✅ No crashes

**Test Deployments, Services similarly**

#### Test 4.5: CRD Explorer in Bundle Mode
**Test Steps:**
```bash
# Launch bundle TUI
# Press 'C' for CRDs
# View CRD list
# Select a CRD
# Press Enter to see instances
```

**Expected Results:**
- ✅ CRD list shows (should see 96 CRDs from bundle)
- ✅ CRD instances can be viewed (or shows empty)
- ✅ No crashes
- ✅ Can navigate back

---

### Phase 5: Regression Testing ✅

#### Test 5.1: Mock Mode Still Works
**Test Steps:**
```bash
./bin/r8s tui --mockdata

# Test all previously passing features:
# - Navigation (7/7 tests)
# - Resource Views (6/6 tests)
# - CRD Explorer (6/6 tests)
```

**Expected Results:**
- ✅ All previously passing tests still pass
- ✅ BUG-002 fix doesn't break anything else

#### Test 5.2: CLI Tests Still Pass
**Test Steps:**
```bash
./test_interactive_tui.sh
```

**Expected Results:**
- ✅ All 8 CLI tests pass

---

### Phase 6: Performance & Edge Cases 🆕

#### Test 6.1: Performance Comparison
**Test Steps:**
```bash
# Time archive mode
time ./bin/r8s bundle import --path=bundle.tar.gz --limit=100

# Time directory mode
tar -xzf bundle.tar.gz -C /tmp/
time ./bin/r8s bundle import --path=/tmp/extracted-bundle/
```

**Expected Results:**
- ✅ Directory mode significantly faster
- ✅ Archive mode has extraction overhead
- ✅ Both load same data correctly

#### Test 6.2: Large Bundle Handling
**Test Scenario:** Bundle > 50MB

**Test Steps:**
```bash
# Test 1: Archive with default limit
./bin/r8s bundle import --path=large-bundle.tar.gz
# Expected: Size limit error with helpful message

# Test 2: Archive with increased limit
./bin/r8s bundle import --path=large-bundle.tar.gz --limit=200

# Test 3: Directory mode (no limits)
tar -xzf large-bundle.tar.gz -C /tmp/
./bin/r8s bundle import --path=/tmp/large-bundle/
# Expected: Works without limit restrictions
```

#### Test 6.3: Bundle with Missing Resources
**Test Scenario:** Bundle missing some kubectl files

**Test Steps:**
```bash
# Create modified bundle
tar -xzf bundle.tar.gz -C /tmp/test-bundle/
rm /tmp/test-bundle/*/rke2/kubectl/deployments
./bin/r8s bundle import --path=/tmp/test-bundle/*/ --verbose
```

**Expected Results:**
- ✅ Graceful handling of missing files
- ✅ Loads available resources
- ⚠️ Warning for missing resources (optional)
- ✅ TUI still launches

---

## Test Execution Order

### Priority 1: Critical Bug Fixes
1. Test 1.1 - BUG-003 kubectl parsing ✅ MUST PASS
2. Test 1.2 - BUG-003 TUI launch ✅ MUST PASS
3. Test 1.3 - BUG-002 describe crash ✅ MUST PASS

### Priority 2: Core New Feature
4. Test 2.1 - Archive mode ✅ MUST PASS
5. Test 2.2 - Directory mode 🆕 MUST PASS
6. Test 2.3 - Auto-detection 🆕 MUST PASS
7. Test 2.4 - Cleanup behavior 🆕 MUST PASS

### Priority 3: Safety & UX
8. Test 3.1-3.4 - Error handling 🆕 SHOULD PASS
9. Test 4.1-4.5 - Bundle TUI 🆕 SHOULD PASS

### Priority 4: Validation
10. Test 5.1-5.2 - Regression tests ✅ MUST PASS
11. Test 6.1-6.3 - Performance & edge cases 🆕 NICE TO HAVE

---

## Test Results Template

### Test Execution Log

| Test ID | Description | Status | Notes |
|---------|-------------|--------|-------|
| 1.1 | BUG-003 kubectl parsing | ⏳ | |
| 1.2 | BUG-003 TUI launch | ⏳ | |
| 1.3 | BUG-002 describe crash | ⏳ | |
| 2.1 | Archive mode | ⏳ | |
| 2.2 | Directory mode | ⏳ | |
| 2.3 | Auto-detection | ⏳ | |
| 2.4 | Cleanup behavior | ⏳ | |
| 3.1 | Path not found error | ⏳ | |
| 3.2 | Invalid directory error | ⏳ | |
| 3.3 | Unsupported format error | ⏳ | |
| 3.4 | Size limit error | ⏳ | |
| 4.1 | TUI with archive | ⏳ | |
| 4.2 | TUI with directory | ⏳ | |
| 4.3 | Resource navigation | ⏳ | |
| 4.4 | Resource actions | ⏳ | |
| 4.5 | CRD explorer | ⏳ | |
| 5.1 | Mock mode regression | ⏳ | |
| 5.2 | CLI tests | ⏳ | |

Legend: ✅ PASS | ❌ FAIL | ⚠️ WARNING | ⏳ PENDING

---

## Success Criteria Summary

### Must Pass (Blockers)
- ✅ All 3 critical bugs fixed
- ✅ Archive mode works (existing)
- ✅ Directory mode works (new)
- ✅ Auto-detection accurate
- ✅ Bundle TUI launches successfully
- ✅ No regressions in mock mode

### Should Pass (Important)
- ✅ Error messages clear and helpful
- ✅ Cleanup behavior correct (no data loss)
- ✅ Performance acceptable
- ✅ Bundle resource views populate

### Nice to Have
- ✅ Edge cases handled gracefully
- ✅ Performance optimization evident
- ✅ Large bundles work

---

## Documentation to Create

After testing:
1. `BUNDLE_ENHANCEMENT_TEST_REPORT.md` - Full test execution results
2. Update `TESTING_MASTER_SUMMARY` with new results
3. `BUG_FIX_VERIFICATION.md` - Confirmation all bugs resolved
4. Update `STATUS.md` - Mark features as tested

---

## Next Steps

1. Execute test suite systematically
2. Document all results
3. Report any new issues found
4. Validate documentation matches behavior
5. Provide recommendations for release

---

**Test Plan Status:** READY FOR EXECUTION  
**Estimated Time:** 2-3 hours for full suite  
**Critical Path:** Tests 1.1-2.4 (bug fixes + core feature)
