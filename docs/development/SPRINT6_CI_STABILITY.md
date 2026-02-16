# Sprint 6: CI Stability & Quality Gates
## CORRECTED PLAN - Sequential Approach

**Status:** REALIGNED per strategic brief approval  
**Original Issue:** Day 3 docs were K3s refactor (scope drift)  
**Correction:** Sprint 6 is CI ONLY. K3s deferred to v0.7.1.

---

## 🎯 Sprint 6 Goal (LOCKED)

**Achieve CI pipeline passing with 50% test coverage.**

**NOT in Sprint 6:**
- ❌ K3s support (moved to v0.7.1)
- ❌ Multi-distro refactoring (moved to v0.7.1)  
- ❌ Path abstraction (moved to v0.7.1)
- ❌ Any new parsers or features

**IN Sprint 6:**
- ✅ Fix all golangci-lint warnings (#44)
- ✅ Achieve 50% test coverage (#45)
- ✅ Re-enable all CI jobs
- ✅ Cross-platform build verification
- ✅ Remove dead code
- ✅ Add `make lint` target

---

## Week 1: Foundation (Days 1-3)

### Day 1: CI Audit & Lint Setup
- [ ] Add `make lint` target to Makefile
- [ ] Install golangci-lint locally
- [ ] Run first lint pass - document all warnings
- [ ] Categorize lint issues (critical vs style)

**Deliverable:** Lint issue inventory with priority rankings

### Day 2: Lint Fixes - Critical Issues
- [ ] Fix error handling issues (unchecked returns)
- [ ] Fix security warnings (if any)
- [ ] Fix performance issues
- [ ] Address CodeRabbit items that are actual bugs

**Deliverable:** Critical lint warnings resolved

### Day 3: Lint Fixes - Style & Cleanup  
- [ ] Fix naming conventions
- [ ] Fix documentation comments
- [ ] Remove unused imports/variables
- [ ] Standardize error messages

**Deliverable:** All lint warnings resolved

---

## Week 2: Coverage & CI (Days 4-7)

### Day 4: Coverage Analysis
- [ ] Run coverage report: `make coverage`
- [ ] Identify uncovered packages
- [ ] Prioritize by risk (core > TUI)
- [ ] Create test plan for critical paths

**Current Coverage:** ~10%  
**Target:** 50%  
**Gap:** 40% needs covering

### Day 5: Core Package Tests
- [ ] `internal/bundle/` - Add tests for edge cases
- [ ] `internal/datasource/` - Test error paths
- [ ] `internal/config/` - Expand existing tests

**Target:** +15% coverage

### Day 6: TUI Package Tests
- [ ] `internal/tui/` - Test helper functions (not Bubble Tea UI)
- [ ] Test data transformation functions
- [ ] Test formatting functions
- [ ] Mock testing for async operations

**Target:** +15% coverage  
**Note:** Full TUI testing is hard (requires Bubble Tea framework) - focus on helper functions

### Day 7: Integration & Verification
- [ ] Run full test suite: `go test ./...`
- [ ] Verify coverage ≥ 50%
- [ ] Re-enable CI coverage threshold
- [ ] Re-enable CI lint job
- [ ] Cross-platform build test

**Deliverable:** All CI jobs green

---

## Deliverables (End of Sprint 6)

| Deliverable | Success Criteria |
|-------------|------------------|
| `make lint` | Runs without errors locally |
| CI Pipeline | All jobs passing (not disabled) |
| Test Coverage | ≥ 50% (`go test -cover`) |
| Zero Lint | `golangci-lint` clean |
| Documentation | Update testing guide |

---

## Make Targets to Add

```makefile
# To add to Makefile:

lint: ## Run golangci-lint
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

ci: lint test coverage ## Run full CI pipeline locally
	@echo "CI checks complete"
```

---

## Metrics to Track

**Daily:**
- Lint warning count
- Test coverage percentage  
- CI job status

**End of Sprint:**
- Total lines of test code added
- Bugs caught by new tests
- CI reliability (green builds)

---

## v0.7.1 Preview (After Sprint 6)

Once CI is green, v0.7.1 scope:
- K3s format detection ONLY
- 5 core file path abstraction
- NO RKE1 support
- NO full 18-file refactor

See: `docs/development/V0.7x_STRATEGIC_BRIEF.md`

---

## Related Documents

- Strategic Brief: `V0.7x_STRATEGIC_BRIEF.md` (sequential timeline approved)
- K3s Work: `SPRINT6_DAY3_MULTI_DISTRO_IMPLEMENTATION.md` (DEFERRED to v0.7.1)
- K3s Checklist: `SPRINT6_DAY3_CHECKLIST.md` (DEFERRED to v0.7.1)

---

**Last Updated:** 2026-02-14 (Scope drift corrected)  
**Approved By:** DontStop (sequential approach)  
**Next Review:** Monday Standup
