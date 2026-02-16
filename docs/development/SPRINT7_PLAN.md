# Sprint 7 Plan: v0.7.1 K3s Support + Coverage Increase

**Sprint Goal:** Deliver K3s bundle format detection and path abstraction while increasing total test coverage from 33% to 45%.

**Duration:** 2 weeks (Feb 17 - Feb 28, 2026)  
**Target Release:** v0.7.1 (March 1, 2026)

---

## 📋 Scope (Reduced per Strategic Brief)

### P0: K3s Bundle Format Detection (Days 1-3)
- Detect K3s vs RKE2 bundle structure
- Update `BundleFormat` type constants
- Add format detection logic to bundle loader
- Smoke test with sample K3s bundle

### P1: Path Abstraction Layer (Days 4-7)
- Abstract hardcoded `rke2/` paths in 5 core files:
  1. `internal/bundle/bundle.go`
  2. `internal/bundle/manifest.go`
  3. `internal/datasource/file.go`
  4. `internal/bundle/journald.go`
  5. `internal/bundle/dmesg.go`
- Create `PathResolver` interface
- Implement `RKE2PathResolver` and `K3sPathResolver`

### P2: Coverage Increase (Days 8-10)
- Target: 33% → 45% total repo coverage
- Focus on untested packages:
  - `internal/config` (47.1% → 70%)
  - `internal/datasource` (26.0% → 50%)
  - `internal/tui` (14.1% → 40%)

### P3: Documentation + Polish (Days 11-14)
- Update README with K3s support
- Document path abstraction design
- CodeRabbit review items
- Final integration testing

---

## 🤖 Work Distribution (AI Team Roles)

### RancherSRE (Lead Developer)
**Allocation:** 60% of sprint capacity
**Responsibilities:**
- Day 1-3: K3s format detection implementation
- Day 4-7: Path abstraction layer (5 core files)
- Day 8-10: Coverage increase - config and datasource packages
- Day 11-14: Integration, testing, bug fixes
- Code review and quality gate enforcement

**Deliverables:**
- [ ] K3s detection working with sample bundles
- [ ] Path abstraction interface implemented
- [ ] 5 core files refactored to use abstraction
- [ ] Coverage increased to 45%

### CodeRabbit (Code Review)
**Allocation:** Always active, reviews all PRs
**Responsibilities:**
- Review every commit for quality issues
- Catch coverage regressions early
- Flag path abstraction anti-patterns
- Document deferred items for v0.7.2

**Deliverables:**
- [ ] Zero lint warnings in new code
- [ ] All coverage-critical paths tested
- [ ] Review backlog < 24 hours

### Documentation (Technical Writer)
**Allocation:** Days 11-14 (intensify pre-release)
**Responsibilities:**
- Update README.md with K3s support announcement
- Document PathResolver interface
- Add troubleshooting guide for multi-distro bundles
- Update CHANGELOG.md

**Deliverables:**
- [ ] README reflects v0.7.1 features
- [ ] Architecture decision record (ADR) for path abstraction
- [ ] CHANGELOG.md updated with release notes

### Release Manager (Spawn at 80% completion)
**Allocation:** Days 12-14
**Responsibilities:**
- Prepare release notes
- Create version bump PR
- Tag v0.7.1
- Announce in channels

---

## 📊 Coverage Increase Plan

| Package | Current | Target | +Lines | Priority |
|---------|---------|--------|--------|----------|
| internal/config | 47.1% | 70% | ~200 | High |
| internal/datasource | 26.0% | 50% | ~300 | High |
| internal/tui | 14.1% | 40% | ~400 | Medium |
| internal/rancher | 0.0% | 30% | ~150 | Low |
| **Total** | **33%** | **45%** | **~1050** | - |

**Strategy:**
1. Test exported functions first (highest impact)
2. Use table-driven tests for parser variations
3. Mock file system for path abstraction tests
4. Skip TUI rendering tests (complex, lower ROI)

---

## 🗓️ Sprint Schedule

| Week | Focus | Key Milestone |
|------|-------|---------------|
| **Week 1** | K3s + Abstraction | Path abstraction working for RKE2 |
| Mon-Tue | K3s detection | Detect format, load K3s bundle |
| Wed-Fri | Path abstraction | 5 core files refactored |
| **Week 2** | Coverage + Polish | 45% coverage achieved |
| Mon-Wed | Coverage push | config + datasource tested |
| Thu-Fri | Integration + docs | All tests green, docs ready |

---

## ⚠️ Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| K3s bundle format changes | High | Use version detection, not hardcoded paths |
| Path abstraction breaks existing | High | Comprehensive test coverage before refactor |
| Coverage target not met | Medium | Prioritize high-value packages, defer TUI tests |
| CodeRabbit lint job flaky | Low | Document workaround, admin merge if needed |

---

## ✅ Definition of Done

- [ ] K3s bundles load and display correctly
- [ ] RKE2 bundles still work (no regression)
- [ ] Path abstraction used in 5+ files
- [ ] Total coverage ≥ 45%
- [ ] All CI checks passing
- [ ] README updated
- [ ] CHANGELOG.md updated
- [ ] v0.7.1 tagged and released

---

## 🚫 Out of Scope (v0.7.2+)

- RKE1 support (deferred to v0.7.4)
- Full test bundles for all distros
- AI pattern matching
- Performance optimization

---

*Plan ready for review. Once approved, work begins in #sprint channel.*
