# Phase 5B: kubectl Resource Parsing - Test Plan

**Date**: 2025-11-27  
**Phase**: Phase 5B - Offline Cluster Explorer  
**Methodology**: Apply lessons from Phases 2-4  
**Focus**: Edge cases, malformed data, race conditions, partial collections

---

## Executive Summary

**Testing Approach**: Proactive code review + systematic testing + malformed data handling

**Priority Testing**:
1. **P0 (Critical)**: Malformed kubectl files, missing files, permission errors
2. **P1 (High)**: Edge cases (empty data, garbage entries, partial collections)
3. **P2 (Medium)**: Performance, concurrency, integration

**Expected Issues**: Parsing errors, graceful degradation, error propagation

---

## Pre-Test Code Review Analysis

### 🔴 CRITICAL FINDINGS

#### Finding #1: Silent Error Swallowing (SECURITY/RELIABILITY ISSUE)
**Location**: `internal/bundle/bundle.go:54-57`

```go
// Parse kubectl resources (ignore errors - these are optional)
crds, _ := ParseCRDs(extractPath)
deployments, _ := ParseDeployments(extractPath)
services, _ := ParseServices(extractPath)
namespaces, _ := ParseNamespaces(extractPath)
```

**Problem**:
- Errors are silently ignored with `_`
- No logging of what went wrong
- User has no visibility into parsing failures
- Cannot distinguish between "file missing" vs "file corrupt"

**Impact**: 🔴 **HIGH**
- Silent data loss
- False sense of completeness
- Debugging nightmare when data is missing

**Example Scenario**:
```
Bundle has corrupted kubectl/deployments file (permission denied)
Result: Bundle loads "successfully" but shows 0 deployments
User thinks: "Bundle has no deployments"
Reality: "File couldn't be read"
```

**Recommendation**: Log errors but continue loading

---

#### Finding #2: No Field Count Validation (CRASH RISK)
**Location**: Multiple parsers

**ParseDeployments** (line 98):
```go
if len(fields) < 6 {
    continue // Need at least namespace, name, ready, uptodate, available, age
}
```

**Problem**: What if fields[5] exists but fields[6] doesn't?
- `kubectl get deployments` output may vary
- Different kubectl versions might format differently
- Corrupted files could have unexpected columns

**ParseServices** (line 146):
```go
if len(fields) < 6 {
    continue
}
// But then accesses fields[5] without checking if more exist
portsStr := fields[5]
```

**Risk**: Array out of bounds if kubectl output format changes

---

#### Finding #3: No Validation of Parsed Values
**Location**: All parsers

**ParseCRDs** (line 49):
```go
kind := strings.Title(plural)
if strings.HasSuffix(kind, "s") && len(kind) > 1 {
    kind = kind[:len(kind)-1]
}
```

**Problem**:
- `strings.Title()` is deprecated (use `cases.Title()`)
- Simple pluralization fails for: "batches" → "Batche", "statuses" → "Statuse"
- No validation that parsed kind makes sense

**ParseServices** (line 162):
```go
fmt.Sscanf(parts[0], "%d", &port)
```

**Problem**:
- No error checking on Sscanf
- If scan fails, `port` is 0 (silent failure)
- No validation that port is in valid range (1-65535)

---

#### Finding #4: Time Parsing Silently Fails
**Location**: `kubectl.go:56`

```go
created, _ := time.Parse(time.RFC3339, createdAt)
```

**Problem**:
- Error ignored
- If parse fails, `created` is zero time (0001-01-01)
- Displays wrong dates to users

---

### 🟡 MEDIUM FINDINGS

#### Finding #5: No Protection Against Malformed Lines
**All parsers**:
```go
fields := strings.Fields(line)
```

**Problem**:
- Assumes whitespace-delimited fields
- What if line contains tabs, multiple spaces, weird unicode?
- What if line is binary garbage (corrupted file)?

**Example Attack**:
```
# Malformed kubectl/crds file
NAME                 CREATED
addons.cattle.io     \x00\x00\x00garbage binary data
```

Result: Unpredictable parsing, possible panic

---

#### Finding #6: Resource Exhaustion
**Location**: `bundle.go:68-79`

```go
// Convert to []interface{} to avoid import cycle
var crdsI, deploymentsI, servicesI, namespacesI []interface{}
for i := range crds {
    crdsI = append(crdsI, crds[i])
}
```

**Problem**:
- Allocates twice: once for typed slices, once for interface{} slices
- O(n) memory overhead
- For large bundles (1000+ CRDs), wastes memory

**Better Approach**: Use capacity pre-allocation
```go
crdsI := make([]interface{}, len(crds))
for i := range crds {
    crdsI[i] = crds[i]
}
```

---

## Gap Analysis: Missing Test Cases

### Gap #1: Permission Denied / Unreadable Files
**Missing Test**: What happens if kubectl files have wrong permissions?

**Test Scenarios**:
```bash
# Create bundle with permission denied files
chmod 000 extracted/rke2/kubectl/crds
r8s --bundle=test.tar.gz
```

**Expected**: Bundle loads but CRDs show as empty (with error logged)  
**Risk**: Silent failure, no indication to user

---

### Gap #2: Corrupted/Truncated Files
**Missing Test**: What if file is partially written?

**Test Scenario**:
```bash
# Create truncated kubectl/deployments file
echo "NAMESPACE    NAME" > kubectl/deployments
# Missing actual data rows
```

**Expected**: Empty deployments list  
**Risk**: User confusion - "Why no deployments?"

---

### Gap #3: Malicious/Garbage Data
**Missing Test**: What if file contains binary garbage, SQL injection-like data?

**Test Scenarios**:
```
# kubectl/services with malicious data
NAMESPACE    NAME             TYPE        CLUSTER-IP
$(rm -rf /)  malicious-svc    ClusterIP   10.43.0.1
../../../etc passwd-stealer   NodePort    10.43.0.2
```

**Expected**: Skip invalid lines, log warnings  
**Risk**: Security vulnerability if not sanitized

---

### Gap #4: Unicode/International Characters
**Missing Test**: Namespace names with unicode

**Test Scenario**:
```
NAMESPACE          NAME
中文-namespace      app-deployment
café-system       café-api
```

**Expected**: Parse correctly  
**Risk**: String operations break on non-ASCII

---

### Gap #5: Very Large Files
**Missing Test**: kubectl output with thousands of entries

**Test Scenario**:
```bash
# Create kubectl/crds with 10,000 CRDs
for i in {1..10000}; do
    echo "crd-$i.example.com  2025-11-27T00:00:00Z" >> kubectl/crds
done
```

**Expected**: Parse in <1s, no OOM  
**Risk**: Performance degradation, memory exhaustion

---

### Gap #6: Concurrent Bundle Loading
**Missing Test**: Load multiple bundles simultaneously

**Test Scenario**:
```bash
# Terminal 1
r8s --bundle=bundle1.tar.gz &

# Terminal 2
r8s --bundle=bundle2.tar.gz &
```

**Expected**: Both load independently  
**Risk**: Race conditions in global state

---

### Gap #7: Empty Kubectl Directories
**Missing Test**: Bundle with no kubectl directory at all

**Test Scenario**:
```bash
# Create bundle without rke2/kubectl/* files
tar czf minimal-bundle.tar.gz systeminfo/ systemlogs/
r8s --bundle=minimal-bundle.tar.gz
```

**Expected**: Bundle loads, shows 0 resources  
**Risk**: Crash if path doesn't exist

---

## Test Execution Plan

### Phase 1: P0 Critical Tests (MUST PASS)

#### A1: Missing kubectl Files
**Test**: Bundle with no kubectl directory
```bash
# Prep: Create minimal bundle
mkdir -p test-bundle/systeminfo
echo "test" > test-bundle/systeminfo/hostname
tar czf test-missing-kubectl.tar.gz -C test-bundle .

# Execute
./bin/r8s --bundle=test-missing-kubectl.tar.gz
```

**Expected**:
- ✅ Bundle loads successfully
- ✅ Shows 0 CRDs, 0 Deployments, 0 Services, 0 Namespaces
- ✅ No crashes
- ⚠️ Warning logged (if verbose logging added)

---

#### A2: Permission Denied on kubectl File
**Test**: Bundle with unreadable kubectl/crds file
```bash
# Prep: Extract bundle, chmod 000
tar xzf example-log-bundle/*.tar.gz
chmod 000 */rke2/kubectl/crds
tar czf test-permission-denied.tar.gz */

# Execute
./bin/r8s --bundle=test-permission-denied.tar.gz
```

**Expected**:
- ✅ Bundle loads
- ✅ CRDs show as empty
- ⚠️ Error logged: "permission denied"

---

#### A3: Corrupted Binary kubectl File
**Test**: kubectl file with binary garbage
```bash
# Prep: Create garbage file
mkdir -p test-bundle/rke2/kubectl
dd if=/dev/urandom of=test-bundle/rke2/kubectl/crds bs=1024 count=10
tar czf test-corrupted.tar.gz -C test-bundle .

# Execute
./bin/r8s --bundle=test-corrupted.tar.gz
```

**Expected**:
- ✅ Bundle loads
- ✅ Parsers skip garbage lines
- ✅ No crash
- ⚠️ Warning about unparseable lines

---

#### A4: Malformed Column Counts
**Test**: kubectl output with inconsistent columns
```bash
# Prep: Create malformed deployments file
cat > test-bundle/rke2/kubectl/deployments <<EOF
NAMESPACE    NAME            READY
kube-system  coredns         # Missing columns
default      web             1/1     EXTRA   COLUMNS   HERE
cattle       agent
EOF
tar czf test-malformed-columns.tar.gz -C test-bundle .

# Execute
./bin/r8s --bundle=test-malformed-columns.tar.gz
```

**Expected**:
- ✅ Bundle loads
- ✅ Lines with wrong column count are skipped
- ✅ Valid lines are parsed
- ⚠️ Warning about skipped lines

---

### Phase 2: P1 High Priority Tests

#### B1: Empty kubectl Files
**Test**: kubectl files exist but are empty
```bash
# Prep
mkdir -p test-bundle/rke2/kubectl
touch test-bundle/rke2/kubectl/{crds,deployments,services,namespaces}
tar czf test-empty-files.tar.gz -C test-bundle .

# Execute
./bin/r8s --bundle=test-empty-files.tar.gz
```

**Expected**:
- ✅ Bundle loads
- ✅ Shows 0 resources
- ✅ No crash

---

#### B2: Header-Only kubectl Files
**Test**: kubectl files with only header line
```bash
# Prep
mkdir -p test-bundle/rke2/kubectl
echo "NAMESPACE    NAME    READY    UP-TO-DATE    AVAILABLE    AGE" > test-bundle/rke2/kubectl/deployments
tar czf test-header-only.tar.gz -C test-bundle .

# Execute
./bin/r8s --bundle=test-header-only.tar.gz
```

**Expected**:
- ✅ Bundle loads
- ✅ Shows 0 deployments
- ✅ No crash

---

#### B3: Unicode/International Characters
**Test**: Resources with unicode names
```bash
# Prep
mkdir -p test-bundle/rke2/kubectl
cat > test-bundle/rke2/kubectl/namespaces <<EOF
NAME                STATUS    AGE
中文-namespace       Active    30d
café-system         Active    20d
über-system         Active    10d
EOF
tar czf test-unicode.tar.gz -C test-bundle .

# Execute
./bin/r8s --bundle=test-unicode.tar.gz
```

**Expected**:
- ✅ Bundle loads
- ✅ Displays 3 namespaces with unicode names correctly
- ✅ No mojibake or encoding errors

---

#### B4: Very Long Lines
**Test**: kubectl lines with extremely long names
```bash
# Prep
LONG_NAME=$(python3 -c "print('a' * 10000)")
mkdir -p test-bundle/rke2/kubectl
echo "NAME                STATUS" > test-bundle/rke2/kubectl/namespaces
echo "$LONG_NAME    Active" >> test-bundle/rke2/kubectl/namespaces
tar czf test-long-lines.tar.gz -C test-bundle .

# Execute
./bin/r8s --bundle=test-long-lines.tar.gz
```

**Expected**:
- ✅ Bundle loads
- ✅ Long name is parsed (or truncated gracefully)
- ✅ No buffer overflow
- ✅ No performance degradation

---

#### B5: Special Characters in Names
**Test**: Resource names with special characters
```bash
# Prep
mkdir -p test-bundle/rke2/kubectl
cat > test-bundle/rke2/kubectl/services <<EOF
NAMESPACE    NAME                TYPE        CLUSTER-IP      PORT(S)
default      api-v1.0            ClusterIP   10.43.0.1       80/TCP
system       svc-with-dash_      NodePort    10.43.0.2       8080/TCP
test         svc.with.dots       ClusterIP   10.43.0.3       443/TCP
EOF
tar czf test-special-chars.tar.gz -C test-bundle .

# Execute
./bin/r8s --bundle=test-special-chars.tar.gz
```

**Expected**:
- ✅ Bundle loads
- ✅ All 3 services parsed correctly
- ✅ Special characters preserved

---

#### B6: Ports Parsing Edge Cases
**Test**: Services with complex port specifications
```bash
# Prep
mkdir -p test-bundle/rke2/kubectl
cat > test-bundle/rke2/kubectl/services <<EOF
NAMESPACE    NAME      TYPE        CLUSTER-IP    PORT(S)
default      multi     ClusterIP   10.43.0.1     80/TCP,443/TCP,8080/UDP
default      range     NodePort    10.43.0.2     80:30080/TCP
default      none      ClusterIP   None          <none>
EOF
tar czf test-ports.tar.gz -C test-bundle .

# Execute
./bin/r8s --bundle=test-ports.tar.gz
```

**Expected**:
- ✅ Multi-port service parsed correctly
- ✅ Port range handled (or skipped gracefully)
- ✅ `<none>` handled without crash

---

### Phase 3: P2 Medium Priority Tests

#### C1: Large kubectl Files
**Test**: kubectl files with 1000+ entries
```bash
# Prep
mkdir -p test-bundle/rke2/kubectl
echo "NAME                CREATED" > test-bundle/rke2/kubectl/crds
for i in {1..1000}; do
    echo "crd-$i.example.com  2025-11-27T00:00:00Z" >> test-bundle/rke2/kubectl/crds
done
tar czf test-large-files.tar.gz -C test-bundle .

# Execute
time ./bin/r8s --bundle=test-large-files.tar.gz
```

**Expected**:
- ✅ Bundle loads in <2 seconds
- ✅ All 1000 CRDs parsed
- ✅ Memory usage reasonable (<100MB extra)
- ✅ TUI scrolling responsive

---

#### C2: Concurrent Bundle Loading
**Test**: Load 2 bundles simultaneously
```bash
# Prep (use existing bundle)

# Execute
./bin/r8s --bundle=example-log-bundle/*.tar.gz &
./bin/r8s --bundle=example-log-bundle/*.tar.gz &
wait
```

**Expected**:
- ✅ Both instances start
- ✅ No interference between processes
- ✅ No temp directory collisions (already tested in Phase 4)

---

#### C3: Integration with Phase 5 Features
**Test**: Navigate through all resource types
```bash
# Execute
./bin/r8s --bundle=example-log-bundle/*.tar.gz

# Navigate:
# 1. Clusters → Enter
# 2. Projects → Enter
# 3. Namespaces → Enter (REAL data)
# 4. Select namespace → '2' for Deployments (REAL data)
# 5. Back → '3' for Services (REAL data)
# 6. Back to cluster → 'C' for CRDs (REAL data)
```

**Expected**:
- ✅ All views show real data
- ✅ No crashes during navigation
- ✅ Filtering works
- ✅ Search works

---

## Code Review Recommendations (Pre-Test)

### Fix #1: Add Error Logging
**Location**: `bundle.go:54-57`

```go
// BEFORE:
crds, _ := ParseCRDs(extractPath)

// AFTER:
crds, err := ParseCRDs(extractPath)
if err != nil {
    // Log but don't fail - kubectl files are optional
    log.Printf("Warning: failed to parse CRDs: %v", err)
}
```

### Fix #2: Validate Field Access
**Location**: All parsers

```go
// BEFORE:
fields := strings.Fields(line)
if len(fields) < 6 {
    continue
}
namespace := fields[0]
portsStr := fields[5]  // Unsafe!

// AFTER:
fields := strings.Fields(line)
if len(fields) < 6 {
    continue
}
namespace := fields[0]
portsStr := ""
if len(fields) > 5 {
    portsStr = fields[5]
}
```

### Fix #3: Validate Parsed Values
**Location**: `kubectl.go:162`

```go
// BEFORE:
var port int
fmt.Sscanf(parts[0], "%d", &port)

// AFTER:
var port int
if n, err := fmt.Sscanf(parts[0], "%d", &port); err != nil || n != 1 {
    continue // Skip invalid port
}
if port < 1 || port > 65535 {
    continue // Skip out-of-range port
}
```

### Fix #4: Pre-allocate Slices
**Location**: `bundle.go:68-79`

```go
// BEFORE:
var crdsI []interface{}
for i := range crds {
    crdsI = append(crdsI, crds[i])
}

// AFTER:
crdsI := make([]interface{}, len(crds))
for i := range crds {
    crdsI[i] = crds[i]
}
```

---

## Test Execution Summary Template

```markdown
## Test Results

### P0 Tests (Critical)
- [ ] A1: Missing kubectl files
- [ ] A2: Permission denied
- [ ] A3: Corrupted binary file
- [ ] A4: Malformed columns

### P1 Tests (High)
- [ ] B1: Empty files
- [ ] B2: Header-only files
- [ ] B3: Unicode characters
- [ ] B4: Very long lines
- [ ] B5: Special characters
- [ ] B6: Ports parsing

### P2 Tests (Medium)
- [ ] C1: Large files (1000+ entries)
- [ ] C2: Concurrent loading
- [ ] C3: Integration testing

### Bugs Found
(Document any issues discovered during testing)

### Code Review Issues
- [ ] Fix #1: Add error logging
- [ ] Fix #2: Validate field access
- [ ] Fix #3: Validate parsed values
- [ ] Fix #4: Pre-allocate slices
```

---

## Success Criteria

### Must Pass (P0)
- ✅ All P0 tests pass without crashes
- ✅ Graceful handling of missing/corrupted files
- ✅ No silent data loss
- ✅ Error logging implemented

### Should Pass (P1)
- ✅ ≥90% P1 tests pass
- ✅ Edge cases handled gracefully
- ✅ Unicode support works

### Nice to Have (P2)
- ✅ Performance acceptable (<2s for large files)
- ✅ Memory usage reasonable
- ✅ Concurrent loading works

---

## Comparison with Previous Phases

| Phase | P0 Tests | Bugs Found (Predicted) | Focus Area |
|-------|----------|------------------------|------------|
| Phase 2 | 9 | 9 (found during test) | Integration bugs |
| Phase 3 | 9 | 1 (found via code review) | Index mismatches |
| Phase 4 | 9 | 1 minor (display bug) | Concurrency, security |
| **Phase 5B** | **7** | **2-3 (predicted)** | **Malformed data, edge cases** |

**Prediction**: Based on code review, expect to find 2-3 issues:
1. Silent error swallowing (already identified)
2. Panic on malformed data (likely)
3. Performance issue with large files (possible)

---

**Test Plan Complete**: 2025-11-27  
**Tests Planned**: 13 (7 P0, 6 P1, 3 P2)  
**Code Review Issues**: 6 critical findings  
**Ready for Execution**: ✅ YES
