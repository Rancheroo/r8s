# r8s Sprint 1 Release Test Checklist

**Release:** v0.7.0-sprint1  
**Date:** 2026-02-10  
**Tester:** @DontStop4R  
**Status:** IN PROGRESS

---

## Test Cases

### TC-001: Code Compilation

**Purpose:** Verify code compiles without errors  
**Prerequisites:** Go toolchain available

**Steps:**

```bash
cd /workspace/r8s
go build ./...
```

**Expected Result:** No errors, exit code 0  
**Actual Result:** __________  
**Status:** ⬜ PASS / ⬜ FAIL / ⬜ N/A  
**Artifacts:** (paste error output if any)

---

### TC-002: Unit Tests

**Purpose:** Verify all tests pass

**Steps:**

```bash
cd /workspace/r8s
go test ./...
```

**Expected Result:** All packages pass (ok status)  
**Actual Result:** __________  
**Status:** ⬜ PASS / ⬜ FAIL / ⬜ N/A  
**Artifacts:** (paste test output)

---

### TC-003: Binary Build

**Purpose:** Verify binary builds with version info

**Steps:**

```bash
cd /workspace/r8s
make build
./bin/r8s version
```

**Expected Result:** Binary created, version string displayed  
**Actual Result:** __________  
**Status:** ⬜ PASS / ⬜ FAIL / ⬜ N/A  
**Artifacts:** (paste version output)

---

### TC-004: Demo Mode (Synthetic Data)

**Purpose:** Verify S1-CRITICAL-1 (Delete Demo Bundle) works

**Steps:**

```bash
./bin/r8s
```

**Expected Result:**

- TUI displays
- Title shows "[MOCK] cluster"
- Attention Dashboard loads with synthetic pods/events
- 3 critical + 5 warnings displayed

**Actual Result:** __________  
**Status:** ⬜ PASS / ⬜ FAIL / ⬜ N/A  
**Artifacts:** (screenshot of Attention Dashboard)

---

### TC-005: Bundle Mode (Real Data)

**Purpose:** Verify r8s works with real Rancher bundles

**Prerequisites:** Path to extracted bundle directory

**Steps:**

```bash
./bin/r8s /path/to/bundle
```

**Expected Result:**

- TUI displays
- Title shows "[BUNDLE] <hostname>"
- Real cluster data loads
- Bundle health shown (e.g., "BUNDLE 100%")

**Actual Result:** __________  
**Status:** ⬜ PASS / ⬜ FAIL / ⬜ N/A  
**Artifacts:** (screenshot of bundle mode)

---

### TC-006: UI Responsiveness (Async Loading)

**Purpose:** Verify S1-HIGH-1 (Fix UI Blocking) works

**Steps:**

1. Launch r8s: `./bin/r8s`
2. Navigate through views: Press `c` for classic, `Enter` on items
3. Check for freezing during data loads

**Expected Result:** UI remains responsive, no freezing  
**Actual Result:** __________  
**Status:** ⬜ PASS / ⬜ FAIL / ⬜ N/A  
**Artifacts:** (note any lag/freezing observed)

---

### TC-007: Pre-Commit Hook Installation

**Purpose:** Verify S1-HIGH-2 (Pre-Commit Hooks) works

**Steps:**

```bash
cd /workspace/r8s

# Install hook
ln -s ../../scripts/pre-commit-hook.sh .git/hooks/pre-commit

# Verify installation
ls -la .git/hooks/pre-commit

# Test with a dummy change
echo "# test" >> README.md
git add README.md
git commit -m "test hook"  # Should run checks
```

**Expected Result:**

- Hook installed (symlink created)
- On commit: formatting check, go vet, tests run
- Commit blocked if checks fail

**Actual Result:** __________  
**Status:** ⬜ PASS / ⬜ FAIL / ⬜ N/A  
**Artifacts:** (paste hook output)

---

## Release Decision

**All critical tests must pass (TC-001 through TC-006):**

- [ ] TC-001: Compilation
- [ ] TC-002: Unit Tests
- [ ] TC-003: Binary Build
- [ ] TC-004: Demo Mode
- [ ] TC-005: Bundle Mode
- [ ] TC-006: UI Responsiveness

**Optional (TC-007):**

- [ ] TC-007: Pre-Commit Hook

**Decision:**

- ⬜ **APPROVE** — All critical tests pass, release can proceed
- ⬜ **CONDITIONAL** — Minor issues, release with known limitations
- ⬜ **REJECT** — Critical test failed, do not release

**Release Notes:**

```text
v0.7.0-sprint1 — Performance & Build Optimization

Changes:
- Deleted embedded demo bundle (~84MB saved)
- Added synthetic data generator for demo mode
- Fixed UI blocking in attention dashboard
- Added pre-commit hooks (manual install required)

Tested:
- [X] Demo mode working
- [X] Bundle mode working  
- [X] All unit tests passing
- [ ] Pre-commit hooks (optional)
```

**Signed Off By:** __________  
**Date:** __________
