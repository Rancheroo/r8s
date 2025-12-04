# Attention Dashboard Test Plan

**Version:** 0.3.3  
**Date:** 2025-12-04  
**Purpose:** Comprehensive testing of the new Attention Dashboard feature  
**Tester:** QA / Developer  

---

## 🎯 Test Objectives

1. ✅ Verify zero false positives for healthy multi-container pods
2. ✅ Confirm real issues are detected and displayed correctly
3. ✅ Test keyboard navigation and user experience
4. ✅ Validate performance (<800ms load time)
5. ✅ Ensure proper behavior in bundle mode vs live mode

---

## 📋 Pre-Test Setup

### Environment Preparation

```bash
# 1. Build r8s
cd /home/bradmin/github/r8s
go build -o r8s

# 2. Verify test bundle exists
ls -la example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09/

# 3. Note: You'll need a live cluster for live mode tests (optional)
```

### Expected Test Bundle Path
```
example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09
```

---

## 🧪 Test Cases

### TEST 1: Dashboard Loads as Default View

**Priority:** P0 (Critical)  
**Objective:** Verify Attention Dashboard appears immediately on launch

**Steps:**
```bash
./r8s example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09
```

**Expected Results:**
- ✅ Dashboard appears within 2 seconds
- ✅ Screen shows "🔍 ATTENTION DASHBOARD" title
- ✅ Breadcrumb shows "[BUNDLE] r8s - Attention Dashboard"
- ✅ Issues are listed OR "All good ✨" message appears

**Pass Criteria:** Dashboard is the first view shown, not clusters list

---

### TEST 2: Zero False Positives for Multi-Container Pods

**Priority:** P0 (Critical - Bug Fix Verification)  
**Objective:** Confirm healthy multi-container pods are NOT flagged

**Steps:**
1. Launch r8s with test bundle
2. Check if any pods with "Not ready (2/2)" or "Not ready (3/3)" appear

**Expected Results:**
- ❌ NO "Not ready (2/2)" alerts (2 out of 2 is healthy!)
- ❌ NO "Not ready (3/3)" alerts (3 out of 3 is healthy!)
- ❌ NO "Not ready (X/X)" where X equals X (healthy status)

**Known Previous Bug:**
Before fix, these would incorrectly appear:
- alertmanager-rancher-monitoring... Not ready (2/2) ❌
- prometheus-rancher-monitoring... Not ready (3/3) ❌
- longhorn-csi-plugin-* Not ready (3/3) ❌

**Pass Criteria:** All multi-container pods with matching ready counts (e.g., 2/2, 3/3) are NOT flagged

---

### TEST 3: Real Issues Are Detected

**Priority:** P0 (Critical)  
**Objective:** Verify genuine problems are still caught

**Expected Detections:**

#### Cluster-Wide Events
- ✅ "221480 Warning events" with description "DNSConfigForming"
  - Emoji: 🟨 (yellow square)
  - Namespace: cluster

**Manual Verification:**
1. Read each alert line
2. Confirm emoji matches severity:
   - 💀 = Critical pod issues (CrashLoopBackOff, OOMKilled, Error)
   - 🚫 = Resource issues (Evicted, ImagePullBackOff)
   - 🔥 = High restarts (≥3)
   - ⚠️  = Not ready (containers not matching, e.g., "1/2")
   - 🟨 = Warning events
   - 🟥 = Critical events

**Pass Criteria:** Dashboard shows only legitimate issues, not false alarms

---

### TEST 4: Keyboard Navigation

**Priority:** P1 (High)  
**Objective:** Test all keyboard shortcuts work correctly

**Steps & Expected Behavior:**

| Key | Expected Action |
|-----|-----------------|
| `c` | Navigate to Clusters view |
| `Esc` / `b` | Return to Attention Dashboard (if navigated away) |
| `r` / `Ctrl+R` | Refresh dashboard (reload signals) |
| `?` | Show help screen |
| `q` | Quit application |
| `↑` / `↓` | Navigate through issues (if multiple) |
| `Enter` | Drill down to selected issue (future feature) |

**Test Each Key:**
```bash
# 1. Press 'c' → Should show Clusters view
# 2. Press 'Esc' → Should return to Dashboard
# 3. Press 'r' → Should refresh (brief "Loading..." then dashboard again)
# 4. Press '?' → Should show help
# 5. Press 'Esc' → Should close help
# 6. Press 'q' → Should quit
```

**Pass Criteria:** All shortcuts work as documented, no crashes

---

### TEST 5: Visual Appearance & Readability

**Priority:** P1 (High)  
**Objective:** Verify dashboard is easy to read and understand

**Visual Checklist:**
- ✅ Colors are distinct (red for critical, yellow for warnings)
- ✅ Emojis render correctly in terminal
- ✅ Text is aligned properly (no overlapping columns)
- ✅ Summary line is clear (e.g., "7 issues need attention (3 critical)")
- ✅ Namespace column shows correct context
- ✅ Status bar shows clear instructions

**Screenshot Test:**
Take a screenshot and verify:
1. Title row is prominent
2. Issue count summary is visible
3. Each issue has: Icon | Name | Description | Namespace
4. Bottom status shows available actions

**Pass Criteria:** Dashboard is immediately understandable, professional appearance

---

### TEST 6: Performance Test

**Priority:** P1 (High)  
**Objective:** Verify dashboard loads quickly (<800ms)

**Test Method:**
```bash
# Time the load
time ./r8s example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09
# (Quit immediately after dashboard appears)
```

**Expected Results:**
- ✅ Dashboard appears in <2 seconds (including all computations)
- ✅ No freezing or lag during load
- ✅ Refresh (press 'r') completes quickly

**Acceptable Performance:**
- Initial load: <2 seconds
- Refresh: <1 second
- On 200MB bundle: <5 seconds

**Pass Criteria:** User doesn't perceive any significant delay

---

### TEST 7: "All Good" State

**Priority:** P2 (Medium)  
**Objective:** Test dashboard when no issues exist

**Note:** Test bundle likely has issues. To test this:

**Option A:** Use a healthy live cluster (if available)
```bash
# If you have a healthy cluster configured:
./r8s  # (no --bundle flag)
```

**Option B:** Code inspection
Review `internal/tui/attention.go` and verify:
- "All good ✨" message is implemented
- Instructions show how to continue with 'c'

**Expected Results:**
- ✅ Shows positive "All good ✨" message
- ✅ Instructions: "Press 'c' to continue to clusters, 'r' to refresh"
- ✅ No scary red/yellow text

**Pass Criteria:** Healthy clusters show reassuring message, not empty table

---

### TEST 8: Edge Cases & Error Handling

**Priority:** P2 (Medium)  
**Objective:** Test unusual scenarios

#### Test 8a: Invalid Bundle Path
```bash
./r8s /nonexistent/path
```
**Expected:** Error message (not crash)

#### Test 8b: Partial Bundle Data
```bash
# If bundle has missing directories (e.g., no kubectl/ folder)
# Should gracefully handle missing data
```
**Expected:** Dashboard shows what it can detect, no crashes

#### Test 8c: Very Large Bundle
```bash
# If you have a large production bundle (>500MB)
# Test performance
```
**Expected:** Still loads, may take longer but completes

**Pass Criteria:** No crashes, graceful degradation

---

### TEST 9: Context Awareness

**Priority:** P2 (Medium)  
**Objective:** Verify dashboard shows correct context for each issue

**Verification Steps:**
1. Review each alert
2. Check namespace column matches the actual pod/resource location
3. Verify issue counts are accurate (e.g., restart counts)

**Sample Verification:**
- "alertmanager-xyz" should show namespace "cattle-monitoring-system"
- "longhorn-csi-plugin-xyz" should show namespace "longhorn-system"
- Events should show namespace "cluster"

**Pass Criteria:** All namespace assignments are correct and helpful

---

### TEST 10: Mode Indicator

**Priority:** P3 (Low)  
**Objective:** Verify mode is clearly displayed

**Test:**
```bash
# Bundle mode
./r8s example-log-bundle/...
# Check breadcrumb shows "[BUNDLE]"

# Live mode (if available)
./r8s
# Check breadcrumb shows "[LIVE]"
```

**Expected Results:**
- ✅ Bundle mode clearly labeled
- ✅ Live mode clearly labeled
- ✅ User knows what data source they're viewing

**Pass Criteria:** Mode is unambiguous from breadcrumb

---

## 📊 Test Results Template

Copy this template to record your results:

```
# Attention Dashboard Test Results
Date: ___________
Tester: ___________
Version: 0.3.3

## Summary
- Total Tests: 10
- Passed: ___
- Failed: ___
- Blocked: ___

## Detailed Results

### TEST 1: Dashboard Loads as Default View
Status: [ ] Pass [ ] Fail [ ] Blocked
Notes: 

### TEST 2: Zero False Positives for Multi-Container Pods
Status: [ ] Pass [ ] Fail [ ] Blocked
Notes: 
Verified pods with X/X status not flagged: [ ] Yes [ ] No

### TEST 3: Real Issues Are Detected
Status: [ ] Pass [ ] Fail [ ] Blocked
Notes: 
DNSConfigForming events shown: [ ] Yes [ ] No
Count: _____

### TEST 4: Keyboard Navigation
Status: [ ] Pass [ ] Fail [ ] Blocked
Notes: 
- 'c' works: [ ]
- 'Esc' works: [ ]
- 'r' works: [ ]
- '?' works: [ ]
- 'q' works: [ ]

### TEST 5: Visual Appearance
Status: [ ] Pass [ ] Fail [ ] Blocked
Notes: 
Emojis render: [ ] Yes [ ] No
Colors clear: [ ] Yes [ ] No

### TEST 6: Performance
Status: [ ] Pass [ ] Fail [ ] Blocked
Notes: 
Load time: _____ seconds

### TEST 7: "All Good" State
Status: [ ] Pass [ ] Fail [ ] Blocked [ ] N/A
Notes: 

### TEST 8: Edge Cases
Status: [ ] Pass [ ] Fail [ ] Blocked
Notes: 

### TEST 9: Context Awareness
Status: [ ] Pass [ ] Fail [ ] Blocked
Notes: 

### TEST 10: Mode Indicator
Status: [ ] Pass [ ] Fail [ ] Blocked
Notes: 

## Issues Found
1. 
2. 
3. 

## Overall Assessment
[ ] Ready for Release
[ ] Needs Fixes
[ ] Major Issues Found

Tester Signature: ___________
```

---

## 🐛 Known Limitations

Document any expected behaviors that aren't bugs:

1. **Bundle timestamp awareness:** Dashboard uses current time for "last 1h/24h" calculations - in bundle mode this may not reflect actual bundle capture time
2. **Log analysis disabled:** Log scanning is commented out for performance (see TEST 3 - Tier 4)
3. **Navigation not yet implemented:** Pressing Enter on an issue doesn't drill down (future enhancement)

---

## ✅ Success Criteria Summary

**For Release Approval, ALL of these must pass:**

1. ✅ TEST 2 must PASS (zero false positives - critical bug fix)
2. ✅ TEST 3 must PASS (real issues detected)
3. ✅ TEST 4 must PASS (core navigation works)
4. ✅ No crashes or panics in any test
5. ✅ Performance is acceptable (<5s on test bundle)

---

## 📝 Test Execution Notes

### Quick Smoke Test (5 minutes)
Run these minimum viable tests:
- TEST 1: Dashboard loads
- TEST 2: No false positives
- TEST 4: 'c', 'r', 'q' keys work

### Full Test Suite (30 minutes)
Run all tests 1-10 sequentially

### Regression Test (after code changes)
Re-run TEST 2, TEST 3, TEST 6 to verify fixes don't break existing functionality

---

## 🔄 After Testing

1. **Record results** using template above
2. **Report issues** via `/reportbug` or GitHub issues
3. **Update this document** with any new edge cases discovered
4. **Take screenshots** of interesting visual bugs
5. **Measure performance** on different bundle sizes

---

## 📞 Questions?

If you encounter unexpected behavior:
1. Check if it's listed in "Known Limitations" above
2. Try with `--verbose` flag for more details
3. Report via GitHub issues with:
   - r8s version (0.3.3)
   - Test being run
   - Expected vs actual behavior
   - Bundle size/type (if relevant)

---

**Good luck with testing! 🚀**
