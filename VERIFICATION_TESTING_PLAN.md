# Verification Testing Plan - Phases A, B, C

## Overview

This document provides a comprehensive testing plan to verify all changes made during Phase A (Documentation), Phase B (Test Infrastructure), and Phase C (Describe Feature).

**Date Created:** 2025-11-26  
**Test Environment:** Development/Local  
**Prerequisites:** Go 1.23, access to Rancher API (optional for offline testing)

---

## Pre-Test Setup

### 1. Verify Git Commits
```bash
cd /home/bradmin/github/r9s

# Check commit history
git log --oneline -4

# Expected output:
# 93a132c feat: add describe support for deployments and services
# fa86c07 test: add comprehensive unit tests with race detection
# 4dfa60b docs: add package-level godoc and fix Go version
# 347b4df Fix data extraction issues in Pods, Deployments, and Projects views
```

### 2. Clean Build
```bash
# Clean any old builds
rm -rf bin/r9s

# Verify go.mod
cat go.mod | grep "^go "
# Expected: go 1.23
```

---

## Phase A: Documentation Verification

### Test A1: Verify Go Version Fix
**Purpose:** Confirm go.mod has correct Go version  
**Expected Result:** Go version is 1.23 (not 1.25)

```bash
# Check go.mod
grep "^go " go.mod

# Expected output:
# go 1.23
```

**Pass Criteria:** ✅ Shows `go 1.23`

---

### Test A2: Verify Package-Level Documentation
**Purpose:** Confirm all packages have godoc comments  
**Expected Result:** Each package has proper documentation

```bash
# Check main.go
head -n 5 main.go | grep "// Package main"

# Check cmd/root.go
head -n 5 cmd/root.go | grep "// Package cmd"

# Check internal/config/config.go
head -n 10 internal/config/config.go | grep "// Package config"

# Check internal/rancher/client.go
head -n 10 internal/rancher/client.go | grep "// Package rancher"

# Check internal/tui/app.go
head -n 10 internal/tui/app.go | grep "// Package tui"
```

**Pass Criteria:** ✅ All 5 packages show godoc comments

---

### Test A3: Verify Typo Fix
**Purpose:** Confirm Pod.HostnameI was renamed to Pod.Hostname  
**Expected Result:** No references to HostnameI

```bash
# Search for old typo (should find nothing)
grep -r "HostnameI" internal/rancher/types.go internal/tui/app.go

# Search for correct field
grep "Hostname" internal/rancher/types.go

# Check app.go uses correct field
grep "pod.Hostname" internal/tui/app.go
```

**Pass Criteria:** ✅ No "HostnameI" found, "Hostname" is used correctly

---

### Test A4: Build Verification
**Purpose:** Confirm code compiles without errors  
**Expected Result:** Clean build with no errors

```bash
# Build the project
go build -o bin/r9s main.go 2>&1

# Check for errors (ignore GOPATH warning)
echo $?
# Expected: 0 (success)

# Verify binary exists
ls -lh bin/r9s
```

**Pass Criteria:** ✅ Build succeeds, binary created

---

## Phase B: Test Infrastructure Verification

### Test B1: Run All Tests
**Purpose:** Verify all unit tests pass  
**Expected Result:** 19 tests pass, 1 skipped, 0 failed

```bash
# Run tests with race detection
make test

# Or manually:
go test -v -race ./...
```

**Expected Output:**
```
=== RUN   TestProfile_GetToken
--- PASS: TestProfile_GetToken (0.00s)
=== RUN   TestConfig_Validate
--- PASS: TestConfig_Validate (0.00s)
...
PASS
ok  	github.com/4realtech/r9s/internal/config	1.022s
PASS
ok  	github.com/4realtech/r9s/internal/rancher	1.038s
```

**Pass Criteria:** 
- ✅ All config tests pass (8 passed, 1 skipped)
- ✅ All rancher tests pass (11 passed)
- ✅ No race conditions detected
- ✅ Total: 19 passed, 1 skipped, 0 failed

---

### Test B2: Verify Race Detection
**Purpose:** Confirm -race flag is enabled in Makefile  
**Expected Result:** Makefile includes -race flag

```bash
# Check Makefile
grep "go test.*-race" Makefile

# Expected output:
# go test -v -race ./...
```

**Pass Criteria:** ✅ -race flag present in test target

---

### Test B3: Test Coverage Report
**Purpose:** Generate coverage report  
**Expected Result:** ~90% coverage for tested packages

```bash
# Generate coverage report
go test -race -coverprofile=coverage.out ./internal/config ./internal/rancher

# View coverage
go tool cover -func=coverage.out | tail -5

# Generate HTML report (optional)
go tool cover -html=coverage.out -o coverage.html
```

**Pass Criteria:** ✅ Coverage shows ~90% for core packages

---

### Test B4: Concurrent Safety Test
**Purpose:** Verify TestClient_ConcurrentRequests passes  
**Expected Result:** No race conditions in concurrent requests

```bash
# Run concurrent test specifically
go test -v -race ./internal/rancher -run TestClient_ConcurrentRequests

# Run multiple times to increase confidence
for i in {1..5}; do
  echo "Run $i:"
  go test -race ./internal/rancher -run TestClient_ConcurrentRequests
done
```

**Pass Criteria:** ✅ All runs pass with no race warnings

---

## Phase C: Describe Feature Verification

### Test C1: Build with New Feature
**Purpose:** Confirm new describe methods compile  
**Expected Result:** Clean build

```bash
# Rebuild
go build -o bin/r9s main.go

# Verify methods exist in binary (optional)
nm bin/r9s | grep -i describe
```

**Pass Criteria:** ✅ Build succeeds

---

### Test C2: Interactive Testing - Offline Mode
**Purpose:** Test describe feature with mock data  
**Expected Result:** Describe works for Pods, Deployments, Services

**Setup:**
```bash
# Create test config (if not exists)
mkdir -p ~/.r9s
cat > ~/.r9s/config.yaml << 'EOF'
current_profile: demo
profiles:
  - name: demo
    url: https://invalid-url-for-offline-testing.local
    bearer_token: demo-token
    insecure: false
refresh_interval: 30s
EOF
```

**Test Steps:**

1. **Launch Application:**
   ```bash
   ./bin/r9s
   ```
   Expected: App launches in offline mode with mock data

2. **Navigate to Pods View:**
   - Press Enter on demo-cluster → Enter on demo-project → Enter on default namespace
   - You should see list of mock pods

3. **Test Pod Describe:**
   - Highlight any pod (e.g., nginx-deployment-6bccc6bf79-w6bbq)
   - Press `d` key
   - Expected: Modal appears with JSON pod details
   - Verify title shows "Pod: default/[pod-name]"
   - Press `Esc` to close

4. **Navigate to Deployments View:**
   - Press `2` key
   - You should see list of mock deployments

5. **Test Deployment Describe:**
   - Highlight any deployment (e.g., nginx-deployment)
   - Press `d` key
   - Expected: Modal appears with JSON deployment details
   - Verify title shows "Deployment: default/[deployment-name]"
   - Verify shows replicas, status
   - Press `d` to close

6. **Navigate to Services View:**
   - Press `3` key
   - You should see list of mock services

7. **Test Service Describe:**
   - Highlight any service (e.g., nginx-service)
   - Press `d` key
   - Expected: Modal appears with JSON service details
   - Verify title shows "Service: default/[service-name]"
   - Verify shows clusterIP, ports
   - Press `Esc` to close

8. **Exit:**
   - Press `q` to quit

**Pass Criteria:**
- ✅ Offline mode works correctly
- ✅ Describe works for Pods
- ✅ Describe works for Deployments
- ✅ Describe works for Services
- ✅ JSON is properly formatted
- ✅ Modal opens and closes cleanly
- ✅ Status text shows 'd'=describe hint

---

### Test C3: Test Describe in Other Views (Negative Test)
**Purpose:** Verify describe shows error in unsupported views  
**Expected Result:** Shows "not yet implemented" message

**Test Steps:**
1. Launch app: `./bin/r9s`
2. Press `Esc` to go to Clusters view
3. Highlight a cluster and press `d`
4. Expected: Error message at top: "Describe is not yet implemented for this resource type"

**Pass Criteria:** ✅ Error message appears for unsupported views

---

### Test C4: Interactive Testing - Online Mode (Optional)
**Purpose:** Test describe feature with real Rancher API  
**Expected Result:** Real resource details displayed

**Prerequisites:** Valid Rancher connection in ~/.r9s/config.yaml

**Test Steps:**
```bash
# Verify you have valid config
cat ~/.r9s/config.yaml

# Launch app
./bin/r9s
```

1. Navigate to a real namespace with pods
2. Highlight a pod and press `d`
3. Verify JSON shows real pod details (not mock)
4. Test deployments and services similarly

**Pass Criteria:**
- ✅ Real data displayed (not mock data)
- ✅ JSON structure is valid
- ✅ All fields populated correctly

---

## Integration Testing

### Test I1: End-to-End Workflow
**Purpose:** Verify complete user workflow  
**Expected Result:** Smooth navigation and feature integration

**Test Steps:**
1. Start app: `./bin/r9s`
2. Navigate: Clusters → Projects → Namespaces → Pods
3. Switch between Pods (1), Deployments (2), Services (3)
4. Describe resources in each view
5. Use Esc to navigate back
6. Verify no crashes or errors

**Pass Criteria:**
- ✅ No crashes
- ✅ All navigation works
- ✅ All describe functions work
- ✅ Memory stable (no leaks)

---

### Test I2: Keyboard Shortcuts
**Purpose:** Verify all keyboard bindings work  
**Expected Result:** All keys respond correctly

**Test Shortcuts:**
- `?` - Show help
- `d` - Describe (in Pods/Deployments/Services)
- `1` - Switch to Pods
- `2` - Switch to Deployments
- `3` - Switch to Services
- `Enter` - Navigate into resource
- `Esc` - Navigate back/close modal
- `r` or `Ctrl+r` - Refresh
- `q` or `Ctrl+c` - Quit

**Pass Criteria:** ✅ All keyboard shortcuts work as expected

---

## Regression Testing

### Test R1: Existing Features Still Work
**Purpose:** Ensure no functionality was broken  
**Expected Result:** All pre-existing features work

**Test Areas:**
1. ✅ Cluster listing
2. ✅ Project listing  
3. ✅ Namespace listing
4. ✅ Pod listing with all columns
5. ✅ Deployment listing with replica counts
6. ✅ Service listing with ports
7. ✅ CRD listing
8. ✅ CRD instance listing
9. ✅ Navigation (Esc, Enter)
10. ✅ View switching (1, 2, 3)

---

### Test R2: Offline Mode Fallback
**Purpose:** Verify graceful degradation  
**Expected Result:** App works without Rancher connection

```bash
# Test with invalid URL
cat > ~/.r9s/config.yaml << 'EOF'
current_profile: offline-test
profiles:
  - name: offline-test
    url: https://invalid-rancher.test
    bearer_token: test
EOF

# Run app
./bin/r9s
```

**Pass Criteria:**
- ✅ App launches successfully
- ✅ Offline mode banner displayed
- ✅ Mock data shown
- ✅ All features work with mock data

---

## Performance Testing

### Test P1: Startup Time
**Purpose:** Measure app startup performance  
**Expected Result:** Quick startup (<2 seconds)

```bash
time ./bin/r9s <<< $'q\n'
```

**Pass Criteria:** ✅ Starts in under 2 seconds

---

### Test P2: Memory Usage
**Purpose:** Check for memory leaks  
**Expected Result:** Stable memory usage

```bash
# Run with memory profiling
go test -memprofile=mem.prof ./internal/rancher -run TestClient_ConcurrentRequests

# Analyze
go tool pprof -top mem.prof
```

**Pass Criteria:** ✅ No unexpected allocations

---

## Documentation Testing

### Test D1: Verify Documentation Files
**Purpose:** Confirm all docs are present and accurate  
**Expected Result:** All documentation exists

```bash
# Check documentation files
ls -la | grep -E "\.md$"

# Should see:
# DOCUMENTATION_AUDIT.md
# TEST_INFRASTRUCTURE_SUMMARY.md
# PHASE_C_DESCRIBE_FEATURE.md
# VERIFICATION_TESTING_PLAN.md (this file)
```

**Pass Criteria:** ✅ All documentation files present

---

### Test D2: Run godoc Server
**Purpose:** Verify godoc generation works  
**Expected Result:** Documentation is accessible

```bash
# Install godoc if needed
go install golang.org/x/tools/cmd/godoc@latest

# Run godoc server
godoc -http=:6060 &

# Open in browser:
xdg-open http://localhost:6060/pkg/github.com/4realtech/r9s/

# Kill server when done
killall godoc
```

**Pass Criteria:** ✅ Package documentation displays correctly

---

## Test Results Summary Template

```
============================================
VERIFICATION TEST RESULTS
Date: [DATE]
Tester: [NAME]
Environment: [DETAILS]
============================================

Phase A - Documentation:
[ ] Test A1: Go Version Fix - PASS/FAIL
[ ] Test A2: Package Documentation - PASS/FAIL
[ ] Test A3: Typo Fix - PASS/FAIL
[ ] Test A4: Build Verification - PASS/FAIL

Phase B - Test Infrastructure:
[ ] Test B1: Run All Tests - PASS/FAIL (X passed, Y skipped)
[ ] Test B2: Race Detection - PASS/FAIL
[ ] Test B3: Coverage Report - PASS/FAIL (X% coverage)
[ ] Test B4: Concurrent Safety - PASS/FAIL

Phase C - Describe Feature:
[ ] Test C1: Build with New Feature - PASS/FAIL
[ ] Test C2: Offline Mode Testing - PASS/FAIL
    [ ] Pod Describe - PASS/FAIL
    [ ] Deployment Describe - PASS/FAIL
    [ ] Service Describe - PASS/FAIL
[ ] Test C3: Negative Testing - PASS/FAIL
[ ] Test C4: Online Mode (Optional) - PASS/FAIL/SKIPPED

Integration Testing:
[ ] Test I1: End-to-End Workflow - PASS/FAIL
[ ] Test I2: Keyboard Shortcuts - PASS/FAIL

Regression Testing:
[ ] Test R1: Existing Features - PASS/FAIL
[ ] Test R2: Offline Mode Fallback - PASS/FAIL

Performance Testing:
[ ] Test P1: Startup Time - PASS/FAIL (X seconds)
[ ] Test P2: Memory Usage - PASS/FAIL

Documentation Testing:
[ ] Test D1: Documentation Files - PASS/FAIL
[ ] Test D2: Godoc Server - PASS/FAIL

============================================
OVERALL RESULT: PASS/FAIL
NOTES:
[Add any additional notes here]
============================================
```

---

## Quick Test Script

For automated basic verification:

```bash
#!/bin/bash
# quick-test.sh - Basic verification script

set -e

echo "=== Quick Verification Test ==="
echo

echo "1. Checking Go version in go.mod..."
grep "^go 1.23" go.mod && echo "✅ PASS" || echo "❌ FAIL"

echo
echo "2. Building project..."
go build -o bin/r9s main.go && echo "✅ PASS" || echo "❌ FAIL"

echo
echo "3. Running unit tests..."
go test -race ./internal/config ./internal/rancher && echo "✅ PASS" || echo "❌ FAIL"

echo
echo "4. Checking for HostnameI typo..."
! grep -r "HostnameI" internal/ && echo "✅ PASS (typo fixed)" || echo "❌ FAIL (typo found)"

echo
echo "5. Checking describe methods exist..."
grep -q "describeDeployment" internal/tui/app.go && \
grep -q "describeService" internal/tui/app.go && \
echo "✅ PASS" || echo "❌ FAIL"

echo
echo "=== Quick Test Complete ==="
```

**Save and run:**
```bash
chmod +x quick-test.sh
./quick-test.sh
```

---

## Critical Issues Checklist

Before approving for production, verify:

- [ ] All 19 unit tests pass with race detection
- [ ] Build completes without errors
- [ ] No references to typo "HostnameI"
- [ ] Describe works for Pods, Deployments, Services
- [ ] Offline mode works correctly
- [ ] No memory leaks detected
- [ ] Documentation is complete
- [ ] No regression in existing features
- [ ] Keyboard shortcuts all work
- [ ] App starts in under 2 seconds

---

## Support & Troubleshooting

### Common Issues

**Issue 1: Tests fail with "GOPATH and GOROOT" warning**
- This is a warning, not an error
- Tests still pass
- Can be ignored or fixed by separating GOPATH/GOROOT

**Issue 2: Offline mode doesn't activate**
- Check ~/.r9s/config.yaml has invalid URL
- Restart application

**Issue 3: Describe modal doesn't appear**
- Verify you pressed 'd' (not 'D')
- Check you're in Pods/Deployments/Services view
- Try refreshing with 'r'

**Issue 4: Build fails**
- Run `go mod tidy`
- Check Go version: `go version` (should be 1.23)
- Clear build cache: `go clean -cache`

---

## Next Steps After Verification

✅ **If all tests pass:**
1. Tag release: `git tag v0.2.0`
2. Push changes: `git push && git push --tags`
3. Build production binary
4. Deploy to target environment
5. Update README with new features

❌ **If tests fail:**
1. Document failures in issue tracker
2. Investigate root cause
3. Fix issues
4. Re-run verification
5. Do not proceed to production

---

**Testing Plan Version:** 1.0  
**Last Updated:** 2025-11-26  
**Maintainer:** Development Team
