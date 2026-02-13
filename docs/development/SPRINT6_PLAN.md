# Sprint 6: CI Stability & Quality Gates

**Status:** 🔄 PLANNING  
**Previous:** [Sprint 5 Phase 2 Complete](./SPRINT5_PHASE2_COMPLETE.md)  
**Goal:** Resolve all CI/CD blockers, enforce quality gates

---

## Sprint 6 Goals

### Primary Objective
Clean up all CI/CD issues and establish automated quality gates.

### Success Criteria
- [ ] All GitHub Actions checks passing
- [ ] golangci-lint re-enabled and clean
- [ ] 50% test coverage threshold enforced
- [ ] Clean build on Linux, macOS, Windows
- [ ] No CI warnings or disabled jobs

---

## Sprint 6 Tasks

### P1: Critical CI Fixes (Week 1)

| Task | Issue | Effort | Description |
|------|-------|--------|-------------|
| Fix lint warnings | #44 | 4h | Resolve all golangci-lint errors |
| Re-enable lint job | #44 | 30m | Uncomment lint job in CI config |
| Fix coverage gaps | #45 | 4h | Add tests to reach 50% threshold |
| Re-enable coverage | #45 | 30m | Uncomment coverage threshold check |
| Go version compatibility | #45 | 2h | Resolve Go 1.24 dependency issues |

### P2: Code Quality (Week 1-2)

| Task | Effort | Description |
|------|--------|-------------|
| Remove unused functions | 2h | Delete dead code identified by linter |
| Fix error handling | 2h | Check all unchecked error returns |
| Add missing documentation | 2h | Document exported functions |
| Standardize imports | 1h | Fix import ordering and grouping |

### P3: Testing Infrastructure (Week 2)

| Task | Effort | Description |
|------|--------|-------------|
| Integration tests | 4h | Add bundle parsing integration tests |
| TUI testing | 4h | Add basic TUI smoke tests |
| Benchmark tests | 2h | Add performance benchmarks |

---

## Branch Strategy

```
feature/sprint6-ci-stability
├── fix/lint-warnings          # Issue #44
├── fix/coverage-threshold     # Issue #45
└── fix/go-compatibility       # Issue #45 (Go version)
```

**Rules:**
- Each fix in its own branch for reviewability
- All branches merge to `feature/sprint6-ci-stability`
- Final PR to main only when all checks green

---

## Definition of Done

- [ ] `make lint` passes locally
- [ ] `make test` passes with >50% coverage
- [ ] All GitHub Actions jobs green
- [ ] No disabled or commented-out CI steps
- [ ] Cross-platform builds successful
- [ ] Documentation updated (CI status badge, etc.)

---

## Risks & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Lint fixes break functionality | High | Run full test suite after each fix |
| Coverage gaps require major refactors | Medium | Focus on critical paths first |
| Go 1.24 compatibility issues | Medium | Test on multiple Go versions |
| Time overrun | Low | Prioritize P1, defer P3 if needed |

---

## Post-Sprint 6

Once CI is clean:
- Resume feature development
- All future PRs must pass full CI
- No more "temporary" disables

---

*Sprint 6: The quality foundation sprint.*
