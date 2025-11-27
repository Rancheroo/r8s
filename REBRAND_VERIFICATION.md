# r9s → r8s Rebrand Verification Plan

**Date:** 2025-11-27  
**Rebrand Version:** Phase 1 Complete  
**Module:** github.com/Rancheroo/r8s

---

## Summary of Changes

This document outlines all changes made during the r9s → r8s rebrand and provides a comprehensive test plan to verify no regressions were introduced.

### Changed Files (8 total)

| File | Changes | Risk Level |
|------|---------|------------|
| `go.mod` | Module path: `github.com/Rancheroo/r8s` | Low |
| `main.go` | Import path, package comment | Low |
| `cmd/root.go` | Imports, CLI text, help messages, config path | Medium |
| `internal/tui/app.go` | Imports, breadcrumb text (2 locations) | Medium |
| `internal/config/config.go` | Default config paths: `~/.r8s/config.yaml` | **HIGH** |
| `Makefile` | Binary name, help text | Low |

### Key Behavioral Changes

1. **Config File Location**: `~/.r9s/config.yaml` → `~/.r8s/config.yaml`
2. **Binary Name**: `bin/r9s` → `bin/r8s`
3. **Module Import Path**: All internal imports updated

---

## Pre-Flight Checks ✅

All completed successfully:

- [x] `go test -race ./...` - All tests pass
- [x] `make build` - Binary builds without errors
- [x] No compilation errors
- [x] No import path conflicts

---

## Regression Test Plan

### Test 1: Build & Installation

**Objective**: Verify binary builds and installs correctly

```bash
# Clean build
make clean
make build

# Verify binary exists and is named correctly
ls -lh bin/r8s
file bin/r8s

# Test binary runs
./bin/r8s version

# Test install (optional - installs to GOPATH)
# make install
# which r8s
```

**Expected Results:**
- Binary exists at `bin/r8s` (not `bin/r9s`)
- Binary is executable
- Version command displays: "r8s dev (commit: ..., built: ...)"

**Potential Issues:**
- ❌ Binary named `r9s` instead of `r8s`
- ❌ Version output shows "r9s" instead of "r8s"

---

### Test 2: Config File Handling (CRITICAL)

**Objective**: Verify config file location change doesn't break existing workflows

#### Test 2A: New Config Creation

```bash
# Remove any existing config
rm -rf ~/.r8s ~/.r9s

# Run binary (should create default config)
./bin/r8s
# Expected: Config created at ~/.r8s/config.yaml
```

**Expected Results:**
- Directory created: `~/.r8s/`
- File created: `~/.r8s/config.yaml`
- Program exits with message: "Created default config file at /home/[user]/.r8s/config.yaml"
- **NOT** created at `~/.r9s/` (old location)

**Potential Issues:**
- ❌ Config created at `~/.r9s/` (old path)
- ❌ Multiple configs created
- ❌ Permission errors

#### Test 2B: Existing Config Migration

```bash
# Simulate old config exists
mkdir -p ~/.r9s
cat > ~/.r9s/config.yaml <<EOF
currentProfile: default
profiles:
  - name: default
    url: https://rancher.example.com
    bearerToken: test-token-12345
refreshInterval: 5s
logLevel: info
EOF

# Run r8s - should NOT find old config
./bin/r8s
# Expected: Creates new config at ~/.r8s/, ignores ~/.r9s/
```

**Expected Results:**
- r8s creates new config at `~/.r8s/config.yaml`
- r8s does **NOT** read from `~/.r9s/config.yaml`
- User must manually migrate old config

**Migration Steps (Document for Users):**
```bash
# If you have an existing r9s config, migrate it:
cp ~/.r9s/config.yaml ~/.r8s/config.yaml
```

**Potential Issues:**
- ❌ r8s reads from old `~/.r9s/` path
- ❌ r8s fails if old config exists

#### Test 2C: Custom Config Path

```bash
# Test --config flag still works
./bin/r8s --config /tmp/custom-config.yaml
```

**Expected Results:**
- r8s respects custom config path
- No interference with default paths

---

### Test 3: TUI Branding

**Objective**: Verify UI shows correct branding

```bash
# Run in offline mode (will show mock data)
./bin/r8s
```

**Expected Results:**
- Top breadcrumb shows: "r8s - Clusters" (not "r9s - Clusters")
- Default fallback breadcrumb shows: "r8s - Rancher Navigator"
- No "r9s" text visible in UI

**How to Verify:**
1. Launch application (will auto-create config and exit)
2. Edit config to add valid credentials OR
3. Just check that error messages don't say "r9s"

**Potential Issues:**
- ❌ Breadcrumb displays "r9s"
- ❌ Help text references "r9s"

---

### Test 4: Help & Version Commands

**Objective**: Verify CLI help text is updated

```bash
# Test version command
./bin/r8s version

# Test help
./bin/r8s --help
./bin/r8s config --help

# Test Makefile help
make help
```

**Expected Results:**
- `r8s version` outputs: "r8s dev ..."
- Help text shows "r8s" in usage examples
- No references to "r9s" in help output
- Makefile shows "Build the r8s binary" (not "r9s")

**Potential Issues:**
- ❌ Help shows "r9s" in examples
- ❌ Usage shows "r9s" instead of "r8s"

---

### Test 5: Unit Tests

**Objective**: Verify all tests still pass

```bash
# Run with race detector
go test -race -v ./...

# Check specific packages
go test -v ./internal/config
go test -v ./internal/rancher
```

**Expected Results:**
- All tests pass (especially config tests)
- No race conditions detected
- Config tests correctly use new paths

**Potential Issues:**
- ❌ Config tests fail due to hardcoded paths
- ❌ Race conditions introduced

---

### Test 6: Import Path Verification

**Objective**: Ensure no broken imports

```bash
# Verify imports
go list -m all | grep -i r9s
# Should return NOTHING

go list -m all | grep -i r8s
# Should show: github.com/Rancheroo/r8s

# Check for any remaining r9s references in code
grep -r "r9s" --include="*.go" .
# Should only show OLD comments/docs, not active code
```

**Expected Results:**
- No dependency on `github.com/4realtech/r9s`
- No dependency on `github.com/Rancheroos/r8s` (wrong handle)
- Dependency shows `github.com/Rancheroo/r8s` (correct)

**Potential Issues:**
- ❌ Imports reference old module path
- ❌ Old module appears in go.mod

---

### Test 7: Full Integration Test

**Objective**: End-to-end workflow verification

```bash
# 1. Clean slate
rm -rf ~/.r8s ~/.r9s bin/

# 2. Build
make clean
make build

# 3. First run (creates config)
./bin/r8s

# 4. Verify config created
ls -la ~/.r8s/config.yaml

# 5. Edit config (add test credentials)
nano ~/.r8s/config.yaml

# 6. Run TUI
./bin/r8s

# 7. Verify branding in TUI
# - Top breadcrumb says "r8s - Clusters"
# - No "r9s" anywhere in interface
```

---

## Rollback Procedure

If critical issues are found:

```bash
# Revert all changes
git revert HEAD

# Or restore from backup
git checkout HEAD~1 -- go.mod main.go cmd/ internal/

# Rebuild
make clean
make build
```

---

## Known Non-Issues (Safe to Ignore)

These are expected and do NOT indicate problems:

1. **GOPATH Warning**: `warning: both GOPATH and GOROOT are the same directory`
   - This is a system configuration issue, not related to rebrand
   
2. **Old Config Presence**: `~/.r9s/config.yaml` exists
   - r8s will NOT automatically migrate this
   - Users must manually copy to `~/.r8s/` if needed

3. **Documentation References**: Old docs may still say "r9s"
   - README, ARCHITECTURE, etc. will be updated in Phase 2
   - Not a code regression

---

## Reporting Issues

If you find regressions, report with:

1. **Test that failed** (Test 1, Test 2A, etc.)
2. **Expected vs Actual behavior**
3. **Error messages** (full text)
4. **Steps to reproduce**

Example:
```
REGRESSION FOUND: Test 2A failed
Expected: Config created at ~/.r8s/config.yaml
Actual: Config created at ~/.r9s/config.yaml
Error: None, but wrong path used
Reproduced by: Running ./bin/r8s with no existing config
```

---

## Completion Checklist

After running all tests, verify:

- [ ] Test 1: Build & Installation - PASS
- [ ] Test 2A: New Config Creation - PASS
- [ ] Test 2B: Existing Config Migration - PASS
- [ ] Test 2C: Custom Config Path - PASS
- [ ] Test 3: TUI Branding - PASS
- [ ] Test 4: Help & Version Commands - PASS
- [ ] Test 5: Unit Tests - PASS
- [ ] Test 6: Import Path Verification - PASS
- [ ] Test 7: Full Integration Test - PASS
- [ ] No regressions found
- [ ] Ready for Phase 2 (Documentation Review)

---

## Next Steps After Verification

Once all tests pass:

1. **Commit the rebrand**:
   ```bash
   git add -A
   git commit -m "rebrand: r9s → r8s (Phase 1)
   
   - Update module path to github.com/Rancheroo/r8s
   - Rename binary from r9s to r8s
   - Update config path from ~/.r9s to ~/.r8s
   - Update all UI branding and breadcrumbs
   - All tests passing, no regressions
   
   BREAKING CHANGE: Config location moved from ~/.r9s to ~/.r8s"
   ```

2. **Fork to new repository** (if not already done):
   ```bash
   gh repo fork 4realtech/r9s --org=Rancheroo --rename=r8s
   ```

3. **Proceed to Phase 2**: Documentation review and godoc audit

---

**End of Verification Plan**
