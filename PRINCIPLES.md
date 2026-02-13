# R8s Principles

Core principles guiding R8s development and support operations.

*Living document - evolves with every incident and feature.*

---

## The 12 Principles

### 1. **Log Bundle is the Source of Truth**

Every feature, fix, and diagnostic **must** account for the log bundle structure.

**Before writing any code:**
- [ ] Check where the data comes from in the bundle
- [ ] Verify the file format (table, JSON, YAML, text)
- [ ] Confirm the path exists in all bundle versions
- [ ] Document the dependency in `docs/development/BUNDLE_DEPENDENCY_ANALYSIS.md`

**Rule:** If the collector script changes, we break. Monitor it weekly.

---

### 2. **Empty is Valid**

Bundles are often partial, corrupted, or from failed collections.

**Every parser must:**
- Handle missing files gracefully
- Work with incomplete data
- Never crash on missing fields
- Provide best-effort results

**Example:** If `poddescribe/` doesn't exist, fall back to `pod-manifests/`, then to basic `kubectl/pods`.

---

### 3. **Fail Gracefully, Explain Clearly**

When something goes wrong, tell the user **what** and **why**.

**Good:**
```
⚠️  QoS class unavailable - pod describe data not found in bundle.
   This bundle was collected before PR #418 (Feb 2026).
   Falling back to manifest parsing...
```

**Bad:**
```
Error: nil pointer dereference
```

---

### 4. **Debug Like You're Not There**

Write code assuming you'll be debugging it at 3 AM without context.

**Required:**
- Verbose debug logging (--debug flag)
- Show what files were found/missing
- Display parsing decisions
- Log data transformations

---

### 5. **Version Awareness**

Bundles change over time. Features appear and disappear.

**Implementation:**
- Detect bundle format version when possible
- Support multiple parsing strategies
- Document version compatibility matrix
- Test with bundles from different time periods

---

### 6. **The Collector Script is a Moving Target**

`rancherlabs/support-tools` changes without warning.

**Monitoring:**
- Weekly check of collector script changes
- Automated diff alerts
- Version pinning for critical features
- Fallback parsers for format changes

**Reference:** See `skills/log-bundle-analysis/SKILL.md`

---

### 7. **Parse Defensively**

Kubernetes output formats change. kubectl tables aren't contracts.

**Rules:**
- Don't assume column positions
- Handle extra/missing columns
- Use regex for flexible matching
- Validate before using parsed data

---

### 8. **Multi-Source Fallbacks**

Critical data should be available from multiple sources.

**Example - Container Names:**
1. Primary: `poddescribe/` (PR #418)
2. Fallback: `pod-manifests/`
3. Fallback: Log filename parsing
4. Last resort: Single container assumption

**Example - QoS Class:**
1. Primary: `poddescribe/` QoS Class field
2. Fallback: Calculate from resource requests/limits
3. Fallback: "Unknown"

---

### 9. **Test with Real Bundles**

Unit tests are necessary but not sufficient.

**Required:**
- Test with actual customer bundles (anonymized)
- Test bundles from different Rancher versions
- Test partial/failed collections
- Test edge cases (empty namespaces, huge pods)

---

### 10. **Document the Dependency**

Every feature has a bundle dependency. Document it.

**Template:**
```go
// GetContainerLogs retrieves logs for a specific container.
// 
// Bundle Dependencies:
//   - rke2/podlogs/<namespace>-<pod> (REQUIRED)
//   - rke2/kubectl/poddescribe/ (for container list, OPTIONAL)
//
// Compatibility:
//   - Pre-2026 bundles: May have limited container detection
//   - PR #418+ bundles: Full container and QoS support
//
// Fallbacks:
//   - Missing poddescribe: Uses pod-manifests/
//   - Missing manifests: Assumes single container
```

---

### 11. **Signal When Data is Missing**

Users should know when they're not seeing the full picture.

**UI Patterns:**
- "⚠️ QoS data unavailable - bundle collected before Feb 2026"
- "⚠️ Showing 1 of 3 containers - describe data not found"
- "ℹ️ Partial bundle - some logs may be missing"

---

### 12. **Bundle Changes are Breaking Changes**

A change in the collector script is a P1 issue for R8s.

**When the script changes:**
1. Analyze impact immediately
2. Create compatibility patch
3. Test with old and new bundles
4. Document version requirements
5. Update monitoring alerts

---

## Principle Violations = Technical Debt

Every time we violate these principles, we create future incidents:

| Violation | Debt | Example |
|-----------|------|---------|
| No bundle check | Parser breaks silently | Sprint 4 QoS issue |
| No fallback | Feature fails on old bundles | Container selector |
| No debug logging | Can't diagnose issues | Current debug build needed |
| No version awareness | Can't support multiple formats | N/A yet |

---

## Applying Principles

### Before Starting Any Feature

```markdown
## Feature: Multi-Container Pod Support

### Bundle Impact Assessment
- [ ] New data source? → Document in BUNDLE_DEPENDENCY_ANALYSIS.md
- [ ] Path changes? → Add fallback for old bundles
- [ ] Format changes? → Support both formats
- [ ] PR #418 dependency? → Test with pre-418 bundles

### Implementation Checklist
- [ ] Parser handles missing files
- [ ] Debug logging added
- [ ] Fallback chain documented
- [ ] Version compatibility noted
- [ ] Real bundle tested
```

---

*Adopted: 2026-02-12*
*Last updated: 2026-02-12*
