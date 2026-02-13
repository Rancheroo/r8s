# Sprint Status - Sprint 5: Quality & Stability

**Current Sprint:** Sprint 5  
**Status:** 🔄 PLANNING  
**Previous:** [Sprint 4 Complete](./SPRINT4_COMPLETE.md)

---

## Sprint 5 Goals

### Primary Objective
Resolve all CI/CD issues and establish quality gates before adding new features.

### Success Criteria
- [ ] All CI checks passing (lint + coverage)
- [ ] 50% test coverage threshold enforced
- [ ] Clean build on all platforms
- [ ] Zero known CI blockers

---

## Sprint 5 Tasks

### P1: CI/CD Stability (Complete First)
| Task | Issue | Effort | Owner |
|------|-------|--------|-------|
| Fix lint warnings | #44 | 2h | TBD |
| Re-enable coverage threshold | #45 | 2h | TBD |
| Resolve Go version compatibility | #45 | 1h | TBD |

### P2: Backlog Parser Work (After CI Clean)
| Task | Issue | Effort | Owner |
|------|-------|--------|-------|
| BACKLOG-4: PV/PVC Parser | #39 | 4h | TBD |
| BACKLOG-5: dmesg OOM Detection | #40 | 3h | TBD |
| BACKLOG-6: RKE2 Journald Parser | #41 | 3h | TBD |

### P3: Bundle Completeness
| Task | Issue | Effort | Owner |
|------|-------|--------|-------|
| Bundle completeness validation | TBD | 4h | TBD |
| Missing file detection | TBD | 2h | TBD |

---

## Sprint 5 Constraints

### CodeRabbit Limit
- **Max 150 files per PR**
- Split large changes into multiple PRs
- Plan file deletions separately

### CI Requirements
- All PRs must pass Build & Test
- Lint must be clean before merge
- Coverage must not decrease

### Workflow
1. Work locally on feature branch
2. Run `make test` locally
3. Push when ready for PR
4. Wait for all green checks
5. Merge via PR

---

## Completed Sprints

| Sprint | Status | Link |
|--------|--------|------|
| Sprint 3 | ✅ Complete | See Sprint 4 doc |
| Sprint 4 | ✅ Complete | [Details](./SPRINT4_COMPLETE.md) |

---

*Last Updated: 2026-02-13*
