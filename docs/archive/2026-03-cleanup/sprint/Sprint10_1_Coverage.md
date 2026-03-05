# Sprint 10.1: Coverage Boost

**Branch:** `feature/sprint10-1-coverage`
**Goal:** cmd/ coverage 5.3% → 80%
**Duration:** 3-5 days
**Reviewer:** @CodeRabbitAI

---

## Scope

### Must Have (80/20)
- [ ] Fix existing test failures (assertion mismatches)
- [ ] cmd/validate.go tests → 60% coverage
- [ ] cmd/export.go tests → 60% coverage  
- [ ] cmd/completion.go tests → 60% coverage

### Nice to Have
- [ ] cmd/logs.go tests → 40% coverage
- [ ] cmd/describe.go tests → 40% coverage
- [ ] cmd/get.go tests → 30% coverage

### Won't Do (defer)
- Integration tests with real bundles (Sprint 11)
- Mock bundle infrastructure (separate PR)

---

## Success Criteria

1. CI passes (no test failures)
2. Coverage ≥ 60% for cmd/ package
3. CodeRabbit approval with 80/20 framework
4. All tests exercise real code paths (not just structs)

---

## 80/20 Framework for CodeRabbit

**Always Fix:**
- Tests that pass but don't test anything
- Coverage gaps > 10% on critical functions
- Race conditions or flaky tests

**Always Defer:**
- Table formatting nits
- "Could be cleaner" comments
- Perfect over good

**Blockers:**
- Security issues
- Panics in tests
- Broken CI

---

## Progress Tracking

| File | Current | Target | Status |
|------|---------|--------|--------|
| cmd/validate.go | 0% | 60% | ⬜ |
| cmd/export.go | ~5% | 60% | ⬜ |
| cmd/completion.go | ~10% | 60% | ⬜ |
| cmd/logs.go | 0% | 40% | ⬜ |
| cmd/describe.go | 0% | 40% | ⬜ |
| cmd/get.go | 0% | 30% | ⬜ |

---

## Notes

- Focus on **executable code paths**, not just structs
- Use table-driven tests where possible
- Mock external dependencies (bundle, filesystem)
- Test error cases, not just happy paths

---

**Ready for CodeRabbit review once tests pass locally.**