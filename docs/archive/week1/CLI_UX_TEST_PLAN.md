# CLI UX & Bug Fix Verification - Test Plan

**Date**: 2025-11-27  
**Testing Scope**: CLI UX improvements (commit c5a6726) + Bug fixes (commit dec2732)  
**Methodology**: User-centric testing + regression verification  

---

## Executive Summary

**Two Major Changes to Test:**
1. 🔴 **Bug Fixes** (commit dec2732): Empty resources + parse logging
2. 🟢 **CLI UX** (commit c5a6726): Explicit modes, better help, no silent fallbacks

**Testing Strategy**:
- **Category A**: CLI UX & Help System (7 tests)
- **Category B**: Mode Behavior Verification (6 tests)
- **Category C**: Bug Fix Regression Tests (5 tests)
- **Category D**: Integration & Edge Cases (4 tests)

**Total Tests**: 22  
**Priority**: P0 (10), P1 (8), P2 (4)

---

## Pre-Test Analysis

### What Changed (Code Review)

#### 1. Root Command Behavior (BREAKING CHANGE)
**Before**: `r8s` → launches TUI immediately (may show mock data silently)  
**After**: `r8s` → shows help (user must choose explicit mode)

**Files Modified**:
- `cmd/root.go` - Removed RunE, added comprehensive help
- `cmd/tui.go` - NEW file, TUI subcommand with --mockdata and --bundle flags
- `internal/config/config.go` - Added MockMode field
- `internal/tui/app.go` - Three distinct mode initialization paths

#### 2. Mode Control (NEW FLAGS)
**New Flags**:
- `--mockdata`: Explicit demo mode (no silent fallback)
- `--bundle`: Bundle analysis mode (moved from root to tui subcommand)

**Behavior Matrix**:
| Command | Rancher API | Bundle | Mock | Result |
|---------|-------------|--------|------|--------|
| `r8s` | N/A | N/A | N/A | Shows help |
| `r8s tui` | Required | No | No | Connect or error |
| `r8s tui --mockdata` | No | No | Yes | Demo mode |
| `r8s tui --bundle=X` | No | Yes | No | Analyze bundle |

#### 3. Error Handling (IMPROVED)
**Before**: Silent fallback to mock on API failure  
**After**: Clear error with helpful guidance

---

## Category A: CLI UX & Help System

### Test A1: Root Command Shows Help (P0)
**Purpose**: Verify no accidental TUI launch  
**Command**: `./bin/r8s`

**Expected Output**:
```
r8s (Rancheroos) - A TUI for browsing Rancher-managed Kubernetes clusters

FEATURES:
  • Interactive TUI for navigating Rancher clusters
  • View pods, deployments, services, CRDs with live data
  • Analyze RKE2 log bundles offline (no API required)
  ...

Available Commands:
  bundle      Work with support bundles
  config      Manage r8s configuration
  tui         Launch interactive terminal UI
  version     Print version information
```

**Pass Criteria**:
- ✅ Shows comprehensive help
- ✅ Does NOT launch TUI
- ✅ Exit code 0
- ✅ Lists subcommands clearly

---

### Test A2: TUI Subcommand Help (P0)
**Purpose**: Verify TUI-specific help is complete  
**Command**: `./bin/r8s tui --help`

**Expected Sections**:
1. Description of TUI functionality
2. Three mode requirements (API, bundle, mockdata)
3. Examples for each mode
4. Keyboard shortcuts reference
5. Flag documentation

**Pass Criteria**:
- ✅ Contains `--mockdata` flag description
- ✅ Contains `--bundle` flag description
- ✅ Shows usage examples
- ✅ Documents keyboard shortcuts
- ✅ Clear and comprehensive

---

### Test A3: Version Command (P2)
**Purpose**: Ensure version command still works  
**Command**: `./bin/r8s version`

**Expected Output**:
```
r8s version dev (c5a6726)
Built: 2025-11-27T...
```

**Pass Criteria**:
- ✅ Shows version, commit, build date
- ✅ Exit code 0

---

### Test A4: Bundle Subcommand Help (P1)
**Purpose**: Verify bundle commands are documented  
**Command**: `./bin/r8s bundle --help`

**Expected**:
- Lists bundle import/list commands
- Shows examples

**Pass Criteria**:
- ✅ Help displays correctly
- ✅ Clear command structure

---

### Test A5: Invalid Subcommand Error (P1)
**Purpose**: Test error handling for typos  
**Command**: `./bin/r8s invalid-command`

**Expected**:
```
Error: unknown command "invalid-command" for "r8s"
Run 'r8s --help' for usage.
```

**Pass Criteria**:
- ✅ Clear error message
- ✅ Helpful suggestion
- ✅ Non-zero exit code

---

### Test A6: Flag Conflicts (P1)
**Purpose**: Test mutually exclusive flags  
**Command**: `./bin/r8s tui --mockdata --bundle=test.tar.gz`

**Expected**: Error or precedence (bundle mode wins)

**Pass Criteria**:
- ✅ Doesn't crash
- ✅ Clear behavior (document which mode wins)

---

### Test A7: Help Output Quality (P2)
**Purpose**: Manual review of help text readability  

**Checklist**:
- [ ] Formatting is clean (alignment, spacing)
- [ ] Examples are realistic and helpful
- [ ] Terminology is consistent (bundle vs log bundle)
- [ ] No typos or awkward phrasing

---

## Category B: Mode Behavior Verification

### Test B1: Live Mode Without API (P0) - CRITICAL
**Purpose**: Verify hard failure instead of silent mock fallback  
**Setup**: No RANCHER_URL or invalid credentials  
**Command**: `./bin/r8s tui`

**Expected Behavior**:
```
Error: Cannot connect to Rancher API at https://rancher.example.com

Options:
  • Check RANCHER_URL and RANCHER_TOKEN
  • Use --mockdata flag for demo mode
  • Use --bundle flag to analyze log bundles
  • Run 'r8s config init' to set up configuration
```

**Pass Criteria**:
- ✅ TUI does NOT launch
- ✅ Clear error message displayed
- ✅ Helpful suggestions provided
- ✅ NO mock data shown
- ✅ Exit code non-zero

**This is THE most important test** - verifies no silent fallback!

---

### Test B2: Mock Mode Explicit (P0)
**Purpose**: Verify --mockdata flag works  
**Command**: `./bin/r8s tui --mockdata`

**Expected Behavior**:
- ✅ TUI launches immediately
- ✅ Shows demo data (4 clusters, mock resources)
- ✅ Status bar indicates "Demo Mode" or similar
- ✅ No API connection attempts
- ✅ Behaves like old offline mode

---

### Test B3: Bundle Mode (P0)
**Purpose**: Verify bundle analysis works with new CLI  
**Command**: `./bin/r8s tui --bundle=example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz`

**Expected Behavior**:
- ✅ TUI launches
- ✅ Shows "Bundle: w-guard-wg-cp-svtk6-lqtxw" in breadcrumb
- ✅ Displays bundle resources
- ✅ No mock data
- ✅ No API connection attempts

---

### Test B4: Live Mode With Valid API (P1)
**Purpose**: Verify live mode still works  
**Setup**: Valid RANCHER_URL and RANCHER_TOKEN  
**Command**: `./bin/r8s tui`

**Expected Behavior**:
- ✅ TUI launches
- ✅ Connects to Rancher API
- ✅ Shows real clusters
- ✅ No mock data
- ✅ Can navigate normally

**Note**: May skip if no test Rancher instance available

---

### Test B5: Config File Mode Override (P1)
**Purpose**: Test runtime flag overrides config  
**Setup**: Config file with valid API URL  
**Command**: `./bin/r8s tui --mockdata`

**Expected**: Uses mock mode (flag overrides config)

**Pass Criteria**:
- ✅ Mock mode activates despite config
- ✅ No API connection

---

### Test B6: Bundle Import Backward Compat (P1)
**Purpose**: Verify old `bundle import` command still works  
**Command**: `./bin/r8s bundle import --path=example-log-bundle/*.tar.gz`

**Expected**:
- ✅ Imports bundle successfully
- ✅ Can launch TUI from imported bundle
- ✅ No breaking changes

---

## Category C: Bug Fix Regression Tests

### Test C1: Empty Resources Show Correctly (P0) - BUG #1 FIX
**Purpose**: Verify Bug #1 fix (empty resources no longer show mock data)  
**Setup**: Create bundle with missing kubectl files

**Steps**:
1. Create test bundle without kubectl/deployments, kubectl/services
2. Load in TUI: `./bin/r8s tui --bundle=test-empty.tar.gz`
3. Navigate to Deployments view (press `2`)
4. Navigate to Services view (press `3`)

**Expected Behavior** (AFTER FIX):
- ✅ Deployments view shows "No deployments available" or empty table
- ✅ Services view shows "No services available" or empty table
- ✅ NO mock data shown (no nginx-deployment, redis-deployment, etc.)

**Failure Scenario** (BEFORE FIX):
- ❌ Shows 4 fake deployments
- ❌ Shows 4 fake services
- ❌ User thinks bundle has these resources

**Validation**: This is the CRITICAL regression test for Bug #1

---

### Test C2: Parse Errors Are Logged (P0) - BUG #2 FIX
**Purpose**: Verify Bug #2 fix (parse errors logged)  
**Setup**: Bundle with corrupted kubectl files

**Steps**:
1. Create bundle with malformed kubectl/crds file (add garbage data)
2. Run: `./bin/r8s tui --bundle=test-malformed.tar.gz 2>&1 | grep -i warning`

**Expected Output**:
```
Warning: Failed to parse CRDs from bundle: <error details>
```

**Pass Criteria**:
- ✅ Parse errors are logged to stderr
- ✅ Bundle still loads (graceful degradation)
- ✅ Error message is helpful
- ✅ No silent failures

---

### Test C3: Empty kubectl Files (P1)
**Purpose**: Verify empty files (not missing) work correctly  
**Setup**: Create bundle with empty kubectl/deployments file

**Steps**:
1. Create kubectl/deployments with only header line
2. Load bundle
3. Navigate to Deployments

**Expected**:
- ✅ No crash
- ✅ Shows empty list (not mock data)
- ✅ No errors logged (valid empty file)

---

### Test C4: Real Bundle Resources (P1)
**Purpose**: Regression test - real data still works  
**Command**: `./bin/r8s tui --bundle=example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09.tar.gz`

**Verification**:
1. Navigate to Deployments (press `2`)
2. Check if real deployments shown (should have 19)
3. Navigate to Services (press `3`)
4. Check if real services shown

**Pass Criteria**:
- ✅ Shows actual bundle data
- ✅ Correct count of resources
- ✅ No mock data mixed in

---

### Test C5: CRD and Namespace Views (P1)
**Purpose**: Verify fetchCRDs/fetchNamespaces also fixed  
**Setup**: Bundle with missing kubectl/crds and kubectl/namespaces

**Expected**:
- ✅ CRD view shows empty (not 50+ mock CRDs)
- ✅ Namespace view shows empty (not mock namespaces)

---

## Category D: Integration & Edge Cases

### Test D1: Mode Switching (P1)
**Purpose**: Test launching different modes sequentially  

**Sequence**:
```bash
./bin/r8s tui --mockdata    # Demo mode
# Quit (q)
./bin/r8s tui --bundle=X    # Bundle mode
# Quit
./bin/r8s tui               # Live mode (should error if no API)
```

**Pass Criteria**:
- ✅ Each mode launches correctly
- ✅ No state pollution between runs
- ✅ Clean exit each time

---

### Test D2: Large Bundle Performance (P2)
**Purpose**: Test with large bundle to verify performance  
**Command**: `./bin/r8s tui --bundle=<large bundle>`

**Pass Criteria**:
- ✅ Loads within reasonable time (<5 seconds)
- ✅ No memory issues
- ✅ Navigation remains responsive

---

### Test D3: Bundle Path Validation (P1)
**Purpose**: Test invalid bundle paths  
**Command**: `./bin/r8s tui --bundle=/nonexistent/path.tar.gz`

**Expected**:
```
Error: Cannot load bundle: file not found
```

**Pass Criteria**:
- ✅ Clear error message
- ✅ TUI doesn't launch
- ✅ No crash

---

### Test D4: Environment Variable Override (P1)
**Purpose**: Test RANCHER_URL/TOKEN env vars with new CLI  
**Setup**: Set RANCHER_URL and RANCHER_TOKEN  
**Command**: `./bin/r8s tui`

**Expected**:
- ✅ Uses environment variables
- ✅ Connects to specified URL
- ✅ No config file required

---

## Testing Execution Plan

### Phase 1: Quick Smoke Tests (5 minutes)
- Test A1 (root help)
- Test A2 (tui help)
- Test B1 (live mode fails correctly)
- Test B2 (mock mode works)

**Goal**: Verify basic CLI structure works

### Phase 2: Bug Fix Verification (10 minutes)
- Test C1 (empty resources - CRITICAL)
- Test C2 (parse errors logged)
- Test C4 (real data works)

**Goal**: Confirm bugs are fixed

### Phase 3: Mode Testing (10 minutes)
- Test B3 (bundle mode)
- Test B6 (bundle import backward compat)
- Test D1 (mode switching)

**Goal**: Verify all modes work

### Phase 4: Edge Cases (10 minutes)
- Test C3, C5 (empty files, CRDs/namespaces)
- Test D3 (invalid paths)
- Test A5, A6 (error handling)

**Goal**: Break it if possible

### Phase 5: Documentation Review (5 minutes)
- Test A7 (help quality)
- Manual review of error messages

**Goal**: User experience polish

**Total Time Estimate**: 40 minutes

---

## Success Criteria

### Critical (Must Pass)
- ✅ **B1**: Live mode fails gracefully (no silent mock)
- ✅ **C1**: Empty resources show empty lists (not mock)
- ✅ **C2**: Parse errors are logged
- ✅ **A1, A2**: Help system is complete

### High Priority
- ✅ All P0 and P1 tests pass
- ✅ No regressions in bundle mode
- ✅ Error messages are helpful

### Nice to Have
- ✅ P2 tests pass
- ✅ Help text is polished
- ✅ Performance is good

---

## Test Artifacts

### Files to Create
1. `test-empty-resources.tar.gz` - Missing kubectl files
2. `test-malformed-crds.tar.gz` - Corrupted kubectl/crds
3. `test-empty-files.tar.gz` - Empty kubectl files (header only)

### Test Results Log
Will document in: `CLI_UX_TEST_RESULTS.md`

---

## Risk Assessment

### Low Risk Changes
- Help text (A1, A2, A7)
- Version command (A3)
- Error messages (A5)

### Medium Risk Changes
- Mode initialization (B1, B2, B3)
- Flag handling (A6, B5)

### High Risk Changes
- Empty resource display (C1) - **CRITICAL BUG FIX**
- Parse error handling (C2) - **CRITICAL BUG FIX**
- Live mode connection failure (B1) - **BREAKING CHANGE**

---

## Rollback Plan

If critical tests fail:
1. Identify failing test category
2. Git revert problematic commit
3. Document issue in GitHub
4. Re-test with revert

**Revert Order**:
- If C1/C2 fail: Revert `dec2732` (bug fixes)
- If B1 fails: Revert `c5a6726` (CLI UX)
- If both fail: Revert both commits

---

**Status**: 🟡 READY TO EXECUTE  
**Estimated Duration**: 40-60 minutes  
**Priority Tests**: B1, C1, C2, A1 (10 minutes for critical path)
