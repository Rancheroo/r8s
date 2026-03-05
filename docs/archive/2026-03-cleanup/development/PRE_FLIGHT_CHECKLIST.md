# Pre-Flight Checklist

**Required for every feature, fix, and issue before writing code.**

*Based on R8s Principles - violations create technical debt.*

---

## Quick Check (30 seconds)

Before starting any work, answer:

1. **Does this read from the bundle?** → Continue to Full Check
2. **Does this change how we parse bundle data?** → Continue to Full Check  
3. **Is this purely UI/internal logic?** → Skip to Implementation Checklist

---

## Full Bundle Impact Assessment (5 minutes)

### Step 1: Identify Data Sources

What files in the bundle does this feature need?

| Data Needed | Bundle Path | Status |
|-------------|-------------|--------|
| Pod list | `rke2/kubectl/pods` | ⬜ Verified |
| Pod describe | `rke2/kubectl/poddescribe/` | ⬜ Verified |
| Pod logs | `rke2/podlogs/*` | ⬜ Verified |
| Pod manifests | `rke2/pod-manifests/` | ⬜ Verified |
| Events | `rke2/kubectl/events` | ⬜ Verified |
| Node info | `systeminfo/*` | ⬜ Verified |
| **Your feature:** | `___________________` | ⬜ Verified |

**Verification:** Check `BUNDLE_DEPENDENCY_ANALYSIS.md` for path and format.

---

### Step 2: Check Collector Script Compatibility

Is the required data available in all bundle versions?

```bash
# Check PR #418 availability (poddescribe)
# Available: Feb 2026+
# Not available: Pre-Feb 2026 bundles

# Check your feature's minimum bundle date
```

- [ ] Data available in latest bundles (PR #418+)
- [ ] Data available in older bundles (pre-PR #418)
- [ ] Fallback path identified for missing data
- [ ] Version compatibility documented

---

### Step 3: Format Analysis

What's the file format?

- [ ] **Table** (kubectl output) → Use flexible column parsing
- [ ] **JSON** → Use structured unmarshaling
- [ ] **YAML** → Use yaml.v3 parser
- [ ] **Plain text** → Use regex with fallback patterns

**Risk Assessment:**
- ⬜ Format is stable (unlikely to change)
- ⬜ Format may change (kubectl tables)
- ⬜ Format will change (documented upstream plans)

---

### Step 4: Fallback Chain

If primary source is missing, what's the fallback?

```
Primary:   ___________________
Fallback 1: ___________________
Fallback 2: ___________________
Final:      ___________________ (graceful degradation)
```

Example (Container Names):
```
Primary:   poddescribe/ (PR #418)
Fallback 1: pod-manifests/
Fallback 2: Log filename parsing
Final:      Single container assumption
```

---

## Implementation Checklist

### Parser Requirements

- [ ] **Handles missing files** (returns empty, not error)
- [ ] **Handles partial data** (parses what exists)
- [ ] **Handles format variations** (extra/missing columns)
- [ ] **Debug logging** added (--debug flag support)
- [ ] **Error messages** explain what was expected vs found

### Code Documentation

```go
// FunctionName does X.
//
// Bundle Dependencies:
//   - path/to/file (REQUIRED/OPTIONAL)
//
// Compatibility:
//   - Pre-2026 bundles: behavior
//   - PR #418+ bundles: behavior
//
// Fallbacks:
//   - Primary: path
//   - Fallback: path
```

- [ ] Bundle dependencies documented
- [ ] Compatibility noted
- [ ] Fallback chain explained

### Testing Requirements

- [ ] Test with **latest bundle** (PR #418+)
- [ ] Test with **older bundle** (pre-PR #418)
- [ ] Test with **partial bundle** (missing files)
- [ ] Test with **empty bundle** (no data)
- [ ] Test debug output is useful

---

## Sign-Off

**Before marking task as complete:**

- [ ] Bundle impact assessment reviewed
- [ ] Fallback chain implemented and tested
- [ ] Debug logging verified
- [ ] Documentation updated
- [ ] Real bundle tested (not just unit tests)

**Reviewer Check (if applicable):**

- [ ] Bundle dependencies verified
- [ ] Fallback logic reviewed
- [ ] Edge cases considered

---

## Example: Multi-Container Pod Support

**Feature:** Show container selector for multi-container pods

### Assessment

**Data Sources:**
- Container names: `rke2/kubectl/poddescribe/` (PR #418)
- Fallback: `rke2/pod-manifests/`

**Compatibility:**
- PR #418+ bundles: Full support
- Pre-PR #418: Limited support (manifests only)

**Fallback Chain:**
```
Primary:   poddescribe/ → parsePodDescribeForContainers()
Fallback 1: pod-manifests/ → parsePodManifestsForContainers()
Fallback 2: Log filename → parsePodLogFilename()
Final:      Single container assumption
```

### Implementation

```go
// parsePodDescribeForContainers extracts container names from PR #418 format
//
// Bundle Dependencies:
//   - rke2/kubectl/poddescribe/* (OPTIONAL, PR #418+)
//
// Compatibility:
//   - PR #418+ bundles (Feb 2026+): Full multi-container support
//   - Pre-PR #418 bundles: Falls back to pod-manifests/
//
// Fallbacks:
//   - Primary: poddescribe/ directory
//   - Fallback: pod-manifests/ YAML files
//   - Last resort: Single container from log filename
func parsePodDescribeForContainers(poddescribeDir string, podMap map[string]*PodInfo) {
    // Implementation with debug logging...
}
```

### Verification

- [x] Tested with PR #418 bundle ✓
- [x] Tested with pre-PR #418 bundle ✓
- [x] Tested with missing poddescribe/ ✓
- [x] Debug logging verified ✓

---

## Template for New Features

Copy this for your feature:

```markdown
## Feature: ___________________

### Bundle Impact Assessment

**Data Sources:**
| Data | Path | Required? |
|------|------|-----------|
|      |      |           |

**Compatibility:**
- Minimum bundle date: ___________
- PR/issue introducing data: #_____

**Fallback Chain:**
```
Primary:    ___________________
Fallback 1: ___________________
Final:      ___________________
```

### Implementation

```go
// Function documentation with bundle deps...
```

### Verification

- [ ] Latest bundle tested
- [ ] Old bundle tested
- [ ] Partial bundle tested
- [ ] Debug logging verified
```

---

*Document: R8s Pre-Flight Checklist*
*Version: 2026-02-12*
*Principles: PRINCIPLES.md*
