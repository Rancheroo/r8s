# R8S v0.9.0 AI Intelligence — Manual Test Plan

**Version:** v0.9.0  
**Branch:** `feature/sprint11-ai-intelligence`  
**Date:** February 24, 2026  
**Tester:** _______________

---

## 🎯 Test Objectives

1. Verify all 19 AI patterns detect correctly
2. Confirm NLQ command responds accurately
3. Validate export formats (SARIF, JUnit, Markdown)
4. Check data quality (no `<no value>` for regex-captured)
5. Ensure no false positives (CNI/DNS fixed)
6. Confirm parallel analysis performance
7. Verify Pattern Registry commands work

---

## 📋 Test Environment Setup

### Prerequisites
```bash
# Ensure you're on the right branch
cd /home/bradmin/.openclaw/workspace/r8s
git checkout feature/sprint11-ai-intelligence
git log --oneline -3  # Should show v0.9.0 tag

# Build fresh
make clean && make build
# Or: go build -o bin/r8s ./main.go

# Verify version
./bin/r8s version  # Should show v0.9.0
```

### Test Bundles
- **Primary:** `~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/`
- **Optional:** Any other RKE2/K3s bundles for variety

---

## ✅ Test Cases

### Section 1: Core AI Analysis (P0 - Critical)

#### TC-1.1: Basic Analysis Command
```bash
./bin/r8s analyze ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/
```

**Expected:**
- [ ] Shows "R8S Bundle Analysis" header
- [ ] Bundle type: RKE2
- [ ] Health: ● (100% complete)
- [ ] "Issues Found:" section present
- [ ] 🔴 CRITICAL: At least 1 issue (CrashLoopBackOff)
- [ ] ⚠️ WARNING: Shows ImagePullBackOff, etcd-latency, etc.
- [ ] Result: "ISSUES FOUND (1 critical, 5 warning)" or similar
- [ ] Exit code: 1 (critical issues found)

**Issues to Check:**
- [ ] No `<no value>` in output (should show actual pod names)
- [ ] Pod names extracted: r8s-test-crash-segfault, r8s-test-crash-exit1, etc.
- [ ] Namespaces shown: r8s-test-app-backend, r8s-test-app-frontend, etc.
- [ ] Restart counts accurate

**Defects:** _______________________________

---

#### TC-1.2: JSON Output Format
```bash
./bin/r8s analyze ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ --format=json | jq '.critical_count'
```

**Expected:**
- [ ] Valid JSON output
- [ ] `.critical_count` shows number ≥ 1
- [ ] `.warning_count` shows number ≥ 1
- [ ] `.issues[]` array contains finding objects
- [ ] Each issue has: severity, type, resource, message, suggestion

**Defects:** _______________________________

---

#### TC-1.3: Severity Filtering
```bash
# Test critical only
./bin/r8s analyze ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ --severity=critical

# Should show only 🔴 issues
```

**Expected:**
- [ ] Only shows CRITICAL issues (no warnings/info)
- [ ] Result shows only critical count
- [ ] Exit code: 1

**Then test:**
```bash
./bin/r8s analyze ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ --severity=warning
```

- [ ] Only shows ⚠️ WARNING issues
- [ ] No 🔴 CRITICAL shown

**Defects:** _______________________________

---

#### TC-1.4: Verbose Mode with Progress
```bash
./bin/r8s analyze ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ -v 2>&1 | tee /tmp/verbose.log
```

**Expected:**
- [ ] Shows "Analyzing bundle..."
- [ ] Shows progress line: "3/10 (30%) - file.log"
- [ ] Progress updates in place (not new lines)
- [ ] Final analysis output after progress

**Defects:** _______________________________

---

### Section 2: Pattern Detection Quality (P0 - Critical)

#### TC-2.1: No False Positives (CNI/DNS Fix)
```bash
./bin/r8s analyze ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ --severity=warning 2>&1 | grep -i "cni"
```

**Expected:**
- [ ] NO output (no CNI issues unless real error)
- [ ] Should NOT match calico, flannel, etc. as errors

**Check same for DNS:**
```bash
./bin/r8s analyze ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ 2>&1 | grep -c "coredns"
```

- [ ] Should show 0-1 (real DNS errors only, not pod names)

**Defects:** _______________________________

---

#### TC-2.2: CrashLoop Detection Count
```bash
./bin/r8s analyze ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ -v 2>&1 | grep -c "crashloopbackoff-v2"
```

**Expected:**
- [ ] Should show 6 (one per CrashLoop pod)
- [ ] Each should have unique pod name

**Verify individual pods:**
```bash
./bin/r8s analyze ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ 2>&1 | grep -E "r8s-test-crash-"
```

- [ ] Shows r8s-test-crash-segfault
- [ ] Shows r8s-test-crash-exit1
- [ ] Shows r8s-test-crash-panic
- [ ] Shows worker-processor pods (if crashed)

**Defects:** _______________________________

---

#### TC-2.3: No `<no value>` in Output
```bash
./bin/r8s analyze ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ 2>&1 | grep -c "<no value>"
```

**Expected:**
- [ ] Result: 0 (no `<no value>` placeholders)
- [ ] All fields populated from regex captures

**Spot check specific examples:**
- [ ] "Pod r8s-test-crash-segfault in namespace r8s-test-app-backend" (not `<no value>`)
- [ ] "Restarts: 7" (not `<no value>`)

**Defects:** _______________________________

---

### Section 3: Export Formats (P1 - High Priority)

#### TC-3.1: SARIF Export
```bash
./bin/r8s export ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ --format=sarif --output=/tmp/test.sarif
cat /tmp/test.sarif | head -20
```

**Expected:**
- [ ] Valid JSON structure
- [ ] Contains `version`, `runs`, `results`
- [ ] `ruleId` shows "crashloopbackoff-v2"
- [ ] `level` shows "error" for critical
- [ ] No errors during export

**Validate JSON:**
```bash
cat /tmp/test.sarif | jq '.runs[0].results | length'  # Should be > 0
```

- [ ] Parses correctly with jq
- [ ] Results array has findings

**Defects:** _______________________________

---

#### TC-3.2: JUnit Export
```bash
./bin/r8s export ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ --format=junit --output=/tmp/test.xml
head -30 /tmp/test.xml
```

**Expected:**
- [ ] Valid XML structure
-- [ ] `<?xml version="1.0" encoding="UTF-8"?>` header
- [ ] Contains `<testsuites>` element
- [ ] `<testcase>` elements present
- [ ] `<failure>` or `<error>` for issues

**Quick validate:**
```bash
head -1 /tmp/test.xml | grep -q "xml version" && echo "OK" || echo "FAIL"
```

- [ ] Passes basic XML check

**Defects:** _______________________________

---

#### TC-3.3: Markdown Export
```bash
./bin/r8s export ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ --format=markdown --output=/tmp/test.md
head -40 /tmp/test.md
```

**Expected:**
- [ ] Mark]down heading: "# R8S Bundle Analysis Report"
- [ ] Shows bundle path and type
- [ ] Summary section with issue counts
- [ ] Detailed findings section
- [ ] No raw JSON in output

**Defects:** _______________________________

---

#### TC-3.4: Export to stdout
```bash
./bin/r8s export ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ --format=json 2>&1 | head -5
```

**Expected:**
- [ ] Outputs to terminal (no file)
- [ ] Valid JSON stream
- [ ] No extra messages in stdout

**Defects:** _______________________________

---

### Section 4: Natural Language Queries (P2 - Medium Priority)

#### TC-4.1: Basic NLQ - Why
```bash
./bin/r8s ask ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ "why is r8s-test-crash-panic crashing?"
```

**Expected:**
- [ ] Shows "🤖 R8S Natural Language Query" header
- [ ] Shows the question back
- [ ] Detects CrashLoop context
- [ ] Provides root cause analysis
- [ ] Shows suggested commands

**Variant:**
```bash
./bin/r8s ask ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ "why is etcd slow?"
```

- [ ] Detects etcd-latency issue
- [ ] Explains latency causes

**Defects:** _______________________________

---

#### TC-4.2: NLQ - Show
```bash
./bin/r8s ask ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ "show me all certificate issues"
```

**Expected:**
- [ ] Lists certificate-related patterns
- [ ] May show "No patterns found" (if no cert issues)

**Variant:**
```bash
./bin/r8s ask ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ "show me crashloop issues"
```

- [ ] Lists all 6 CrashLoop issues

**Defects:** _______________________________

---

#### TC-4.3: NLQ - Unknown Query
```bash
./bin/r8s ask ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ "what is the meaning of life?"
```

**Expected:**
- [ ] Gracefully handles unknown query
- [ ] Shows supported patterns with examples
- [ ] No panic or crash

**Defects:** _______________________________

---

### Section 5: Pattern Registry Commands (P2 - Medium Priority)

#### TC-5.1: List Patterns
```bash
./bin/r8s patterns list
```

**Expected:**
- [ ] Shows "Available Patterns" header
- [ ] Table with ID, CATEGORY, SEVERITY, DESCRIPTION
- [ ] At least 19 patterns shown
- [ ] Critical patterns shown in 🔴 red
- [ ] Footer: "Total: X patterns"

**Filters:**
```bash
./bin/r8s patterns list --severity=critical
./bin/r8s patterns list --category=network
```

- [ ] Filters work correctly
- [ ] Only shows matching patterns

**Defects:** _______________________________

---

#### TC-5.2: Show Pattern Details
```bash
./bin/r8s patterns show crashloopbackoff-v2
```

**Expected:**
- [ ] Shows "Pattern Details" header
- [ ] ID, Name, Category, Severity, Confidence
- [ ] Description section
- [ ] Matchers list (with regex/keyword types)
- [ ] Correlations (if any)
- [ ] Hint template shown
- [ ] Suggestion shown
- [ ] Command example shown
- [ ] References (URLs) shown

**Defects:** _______________________________

---

#### TC-5.3: Search Patterns
```bash
./bin/r8s patterns search crash
./bin/r8s patterns search cert
./bin/r8s patterns search network
```

**Expected:**
- [ ] Shows "Search Results for '...'" header
- [ ] Lists matching patterns
- [ ] Shows pattern ID + severity + description
- [ ] "Found X patterns" footer

**Edge case:**
```bash
./bin/r8s patterns search xyznotfound
```

- [ ] Shows "No patterns found" message

**Defects:** _______________________________

---

### Section 6: Data Quality & Edge Cases (P1 - High Priority)

#### TC-6.1: Empty/Invalid Bundle
```bash
./bin/r8s analyze /tmp/empty-dir-xyz/ 2>&1
```

**Expected:**
- [ ] Error message: "cannot access bundle path" or "failed to analyze bundle"
- [ ] Exit code: 2

**Defects:** _______________________________

---

#### TC-6.2: Bundle with No Issues
```bash
# If you have a healthy bundle test case
# ./bin/r8s analyze ~/healthy-bundle
```

**Expected:**
- [ ] Shows "✓ No issues detected"
- [ ] Result: HEALTHY
- [ ] Exit code: 0

**Defects:** _______________________________

---

#### TC-6.3: Parallel Analysis Performance
```bash
time ./bin/r8s analyze ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ --format=json > /dev/null
```

**Expected:**
- [ ] Completes in < 500ms for this bundle
- [ ] CPU usage spikes (parallel goroutines)

**Compare to sequential (if possible):**
```bash
# Check if there's a --workers=1 flag or similar
# time ./bin/r8s analyze ... --workers=1
```

**Defects:** _______________________________

---

#### TC-6.4: Exit Codes (CI/CD)
```bash
./bin/r8s analyze ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ --severity=info
echo "Exit code: $?"  # Should be 1 (critical found)

./bin/r8s get pods ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ > /dev/null
echo "Exit code: $?"  # Should be 0 (success)
```

**Expected:**
- [ ] Exit code 1 for critical issues
- [ ] Exit code 0 for success
- [ ] Exit code 2 for errors

**Defects:** _______________________________

---

### Section 7: Regression Testing (P1 - High Priority)

#### TC-7.1: Legacy kubectl Commands
```bash
./bin/r8s get pods ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/
./bin/r8s get nodes ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/
./bin/r8s logs ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ calico-node-xyz
./bin/r8s describe pod ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ kube-apiserver-r8s-cp
```

**Expected:**
- [ ] All legacy commands still work
- [ ] Output format unchanged
- [ ] No AI analysis mixed in

**Defects:** _______________________________

---

#### TC-7.2: Generate Prompts (Sprint 7 feature)
```bash
./bin/r8s generate prompt ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ --format=terminal
```

**Expected:**
- [ ] Works as before (no regression)
- [ ] Generates AI prompt for troubleshooting

**Defects:** _______________________________

---

#### TC-7.3: Validate Command
```bash
./bin/r8s validate ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/
./bin/r8s validate ~/Downloads/r8s-cp-wlp7h-lhvgq/r8s-cp-wlp7h-lhvgq-2026-02-03_05_38_57/ --format=json
```

**Expected:**
- [ ] Bundle validation still works
- [ ] Shows completeness percentage
- [ ] Lists missing files (if any)
- [ ] JSON output valid

**Defects:** _______________________________

---

### Section 8: Documentation (P3 - Low Priority / Optional)

#### TC-8.1: README Accuracy
```bash
cat README.md | grep -A5 "v0.9.0"
```

**Expected:**
- [ ] Mentions v0.9.0
- [ ] Lists AI features
- [ ] Examples match actual CLI

**Manual check:**
- [ ] Quick start examples work
- [ ] No broken commands

**Defects:** _______________________________

---

#### TC-8.2: Help Text
```bash
./bin/r8s --help
./bin/r8s analyze --help
./bin/r8s ask --help
./bin/r8s export --help
./bin/r8s patterns --help
```

**Expected:**
- [ ] Help text displays correctly
- [ ] No formatting issues
- [ ] All new commands have help

**Defects:** _______________________________

---

## 📊 Summary Matrix

| Test Category | Total | Passed | Failed | Blocked |
|--------------|-------|--------|--------|---------|
| Core AI Analysis (P0) | 4 | | | |
| Pattern Quality (P0) | 3 | | | |
| Export Formats (P1) | 4 | | | |
| NLQ (P2) | 3 | | | |
| Pattern Registry (P2) | 3 | | | |
| Data Quality (P1) | 4 | | | |
| Regression (P1) | 3 | | | |
| Documentation (P3) | 2 | | | |
| **TOTAL** | **26** | | | |

---

## 🔴 Critical Issues Found

| ID | Description | Severity | Status |
|----|-------------|----------|--------|
| | | | |

---

## 📝 Notes for Developers

1. **Template Extraction:** Some patterns still show `<no value>` — check regex capture group names match template variables (e.g., `{{.PodName}}` needs `?P<PodName>`).

2. **Performance:** Parallel analysis should be ~4-8x faster than sequential. If not, check `runtime.NumCPU()`.

3. **False Positives:** If CNI/DNS patterns trigger on pod names, the fix in `37349cf` may need adjustment.

4. **Exit Codes:** CI/CD relies on exit codes. Verify code 1 for issues, 2 for errors, 0 for success.

---

## 🎯 Go/No-Go Decision

| Criteria | Status | Notes |
|----------|--------|-------|
| All P0 tests pass | ⬜ | |
| No critical defects | ⬜ | |
| Performance acceptable | ⬜ | |
| Documentation complete | ⬜ | |
| **RELEASE?** | ⬜ | |

---

**Tester Signature:** _________________  
**Date:** _________________  
**Branch:** `feature/sprint11-ai-intelligence` at commit: `6c71ebc`