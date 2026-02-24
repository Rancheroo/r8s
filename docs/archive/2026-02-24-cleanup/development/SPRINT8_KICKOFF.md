# Sprint 8 Kickoff — v0.7.2 (Bundle Health + AI + CLI)

**Date:** 2026-02-17  
**Sprint Duration:** 2 weeks (Feb 17 - Feb 28, 2026)  
**Target Release:** v0.7.2  
**Branch:** `release/v0.7.x`

---

## Sprint Goal

Deliver **management showcase-ready** features: enhanced bundle health + AI pattern detection + CLI headless mode. Apply 80/20 ruthlessly.

---

## Quick Reference

| Resource | Location |
|----------|----------|
| Full Sprint Plan | [SPRINT8_PLAN.md](SPRINT8_PLAN.md) |
| TUI Deletion Audit | [TUI_DELETION_AUDIT.md](TUI_DELETION_AUDIT.md) |
| Strategic Brief | [V0.7x_STRATEGIC_BRIEF.md](V0.7x_STRATEGIC_BRIEF.md) |

---

## Day-by-Day Tracker

### Week 1: Core Features + CLI Foundation

| Day | Focus | Tasks | Status |
|-----|-------|-------|--------|
| **1** | Bundle Health Core | Create `internal/bundle/health.go`, tests, `r8s validate bundle` skeleton | ⬜ |
| **2** | `r8s validate` | Complete command, add impact scoring, reach 80% coverage | ⬜ |
| **3** | AI Engine | Pattern interface, YAML loader, registry structure | ⬜ |
| **4** | `r8s generate prompt` | Integrate PromptGenerator, 3 formats (chatbot\|terminal\|script) | ⬜ |
| **5** | Pattern Registry | 3 patterns: OOMKill, ImagePull, CrashLoopBackOff | ⬜ |
| **6** | `r8s export findings` | JSON/YAML output, findings serialization | ⬜ |
| **7** | `r8s create demo-bundle` | Export synthetic demo to disk — **DEMO DAY** | ⬜ |

### Week 2: Polish + Documentation + Cleanup

| Day | Focus | Tasks | Status |
|-----|-------|-------|--------|
| **8-10** | TUI Polish | AI Analysis tab, UX Engineer accessibility review | ⬜ |
| **11-13** | Documentation | Demo script, CLI reference, README update | ⬜ |
| **14** | TUI Cleanup | Delete dead views (per TUI_DELETION_AUDIT.md) — **RELEASE DAY** | ⬜ |

---

## Today's Tasks (Day 1 — Feb 17)

### Morning: Setup + Foundation
- [ ] Review [SPRINT8_PLAN.md](SPRINT8_PLAN.md) one more time
- [ ] Create feature branch: `feature/sprint8-cli-commands`
- [ ] Create `internal/bundle/health.go` with basic health checker
- [ ] Define `BundleHealth` struct (expand existing)

### Afternoon: `r8s validate` Skeleton
- [ ] Create `cmd/validate.go` with Cobra command structure
- [ ] Implement basic bundle loading + validation
- [ ] Add output formatting (table + JSON)

### End of Day
- [ ] Build passes: `make build`
- [ ] Tests pass: `make test`
- [ ] Commit progress

---

## Implementation Order

1. **Bundle Health Core** (Days 1-2)
   - `internal/bundle/health.go` — Health checker
   - `internal/bundle/health_test.go` — Unit tests (target: 80% coverage)
   - `cmd/validate.go` — CLI command

2. **AI Pattern Engine** (Days 3-4)
   - `internal/ai/pattern.go` — Pattern interface
   - `internal/ai/registry.go` — Pattern registry
   - `internal/ai/patterns/*.yaml` — OOMKill, ImagePull, CrashLoop
   - `cmd/generate.go` — `r8s generate prompt`

3. **CLI Export + Create** (Days 5-7)
   - `cmd/export.go` — `r8s export findings`
   - `cmd/create.go` — `r8s create demo-bundle`
   - Integration tests for all 4 commands

4. **TUI Integration** (Days 8-10)
   - `internal/tui/attention_ai.go` — AI Analysis panel
   - Bundle health indicator in TUI
   - UX Engineer accessibility review

5. **Documentation** (Days 11-13)
   - Demo script (3 minutes)
   - CLI reference in README
   - Pattern authoring guide

6. **TUI Cleanup** (Day 14 — Buffer)
   - [TUI_DELETION_AUDIT.md](TUI_DELETION_AUDIT.md) execution

---

## Demo Script (3 Minutes)

```bash
# 1. Validate bundle health (CI-friendly)
$ r8s validate ./production-bundle/
Bundle Health: 68% ⚠️
Missing: podlogs/ (medium impact), sysstat/ (low impact)

# 2. Generate AI prompt for troubleshooting
$ r8s generate prompt ./production-bundle/ --format=terminal
# Copy-paste to Claude Code → get kubectl commands

# 3. Export for monitoring integration
$ r8s export findings ./production-bundle/ --format=json | jq '.critical[]'

# 4. Create demo bundle for testing
$ r8s create demo-bundle --output ./my-demo/
✓ Demo bundle created: ./my-demo/ (10 pods, 3 patterns)
```

---

## Success Metrics

| Metric | Target | Current |
|--------|--------|---------|
| Bundle Health v2 | Working + tests | ⬜ |
| CLI Commands (4) | All functional | ⬜ |
| AI Patterns (3) | Detect correctly | ⬜ |
| TUI Polish | Demo-ready | ⬜ |
| **Coverage** | **≥ 45%** | 36.8% |
| Showcase Ready | 3-min demo script | ⬜ |

---

## Blockers / Dependencies

| Item | Status | Owner |
|------|--------|-------|
| UX Engineer availability | Days 8-14 | External |
| Sample K3s bundles for testing | Need to acquire | RancherSRE |
| Management demo date | TBD | Management |

---

## Quick Commands

```bash
# Build
make build

# Test
make test

# Coverage
make coverage

# Lint
make lint

# All checks
make ci
```

---

## Notes

- **Law #2 (Delete):** TUI cleanup deferred to Day 14 — focus on features first
- **80/20 Rule:** 3 patterns only, not 10. Quality over quantity.
- **Demo-First:** Day 7 milestone is non-negotiable.

---

**Musk's Laws: Question → Delete → Simplify → Accelerate → Automate**

*Let's ship v0.7.2* 🚀
