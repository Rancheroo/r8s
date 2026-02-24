# Sprint 4 Re-Prioritized: Quick Wins + Foundation

**Date:** 2026-02-12
**Theme:** Low Effort, High Gain + Future-Proofing
**Duration:** 1 week (reduced from 3 weeks)
**Total Effort:** ~12 hours

---

## Musk's Law Applied: DELETE Before Building

**Removed from Sprint 4 (Pushed to Future/Backlog):**

| Task | Reason | Future Home |
|------|--------|-------------|
| S4-HIGH-1: PV/PVC/StatefulSet parsers | 8h effort, complex parsing | Sprint 6 (if customer demand) |
| S4-HIGH-2: dmesg OOM detection | 6h effort, new parser type | Backlog (nice-to-have) |
| S4-HIGH-3: RKE2 journald parser | 6h effort, complex format | Sprint 5 (performance focus) |
| S4-MEDIUM-1: ConfigMaps + HelmCharts | 5h effort, lower priority | Backlog |

**Why Delete:** These are "would be nice" features that don't block Sprint 5/6. We can add them when:
- A customer explicitly asks
- We have bundle samples to test against
- Other critical work is done

---

## Sprint 4: NEW FOCUS (Quick Wins Only)

### 🎯 Three Tasks, Maximum Impact

| ID | Task | Effort | Gain | Sets Up Future? |
|----|------|--------|------|-----------------|
| **S4-CRITICAL-1** | CI/CD Pipeline (GitHub Actions) | 4h | 🔥 HIGH | ✅ Sprint 5/6 quality |
| **S4-HIGH-4** | Bundle completeness indicator | 4h | 🔥 HIGH | ✅ User trust |
| **S4-HIGH-5** | 50% coverage enforcement | 4h | 🔥 HIGH | ✅ Code quality |

**Total:** 12 hours, 1 week

---

## Task Details

### S4-CRITICAL-1: CI/CD Pipeline (Foundation for Everything)
**Why First:** Blocks bad code from reaching main, enables confident sprints

**Scope (Minimal Viable):**
```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - run: make test
      - run: make build
      - run: make coverage  # Generate coverage report
```

**Success Criteria:**
- [ ] PRs blocked if tests fail
- [ ] Coverage report uploaded as artifact
- [ ] Build verification on every push

**Future Extension:** Add linting, cross-platform builds later

---

### S4-HIGH-4: Bundle Completeness Indicator (User Trust)
**Why:** Users immediately know if bundle is useful

**Scope:**
```
┌─────────────────────────────────────┐
│ r8s v0.7.0-sprint4                  │
│                                     │
│ Bundle: ./support-bundle/           │
│ Health: ████████░░ 80% (4/5)        │
│                                      │
│ Missing: journald logs              │
│                                      │
│ [Continue] [Exit]                   │
└─────────────────────────────────────┘
```

**Implementation:**
- Reuse existing `BundleHealth` struct
- Check 5 key files on startup
- Show progress bar + missing items list
- Non-blocking (user can continue with incomplete bundle)

**Files:**
- `internal/tui/app.go` - Add startup check
- `internal/bundle/health.go` - Enhance percentage calculation
- `internal/tui/views.go` - Add bundle health view

---

### S4-HIGH-5: 50% Coverage Enforcement (Quality Gate)
**Why:** Prevents regression, forces test discipline

**Scope:**
- Add coverage check to CI
- Fail build if <50% coverage
- Report coverage in PR comments (optional)

**Makefile addition:**
```makefile
coverage-check:
	@echo "Checking coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//' | \
		awk '{if ($$1 < 50.0) { print "FAIL: Coverage is $$1%, minimum is 50%"; exit 1 } \
		else { print "PASS: Coverage is $$1%" } }'
```

**CI Integration:**
```yaml
- name: Coverage Check
  run: make coverage-check
```

---

## Sprint 4 Success Criteria (Revised)

1. ✅ CI/CD runs on every PR
2. ✅ Bundle completeness shown on startup
3. ✅ Coverage enforcement blocks <50% PRs
4. ✅ Sprint 3 branches merged cleanly

**NOT in Sprint 4 (and that's OK):**
- ❌ PV/PVC parsers (backlog)
- ❌ dmesg analysis (backlog)
- ❌ RKE2 journald parser (Sprint 5 candidate)
- ❌ ConfigMaps/HelmCharts (backlog)

---

## Future Sprint Impact

| Sprint | Benefit from Sprint 4 |
|--------|----------------------|
| Sprint 5 (Performance) | CI catches regressions automatically |
| Sprint 6 (Production) | Coverage enforcement already in place |
| All Future | Bundle completeness = user confidence |

---

## Anti-Patterns Avoided

1. **No Parser Overload:** We resisted adding 4 complex parsers in one sprint
2. **Foundation First:** CI/CD enables confident speed later
3. **User Trust:** Bundle completeness indicator = immediate value
4. **Quality Gates:** Coverage enforcement prevents tech debt

---

**Decision:** Sprint 4 is now a 1-week "foundation sprint" - setting up CI, coverage, and user trust. Heavy parsing work deferred until we have clearer demand or bundle samples.

**Next:** After Sprint 4, we have clean CI + coverage gates + user trust. Sprint 5 can focus on performance with confidence.
