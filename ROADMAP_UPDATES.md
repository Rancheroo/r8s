# r8s Roadmap Updates

**Last Updated**: 2026-02-20

---

## 🚨 CRITICAL: CI/CD Review Required (Added 2026-02-20)

**Status:** CI/CD DISABLED for feature branches pending review  
**Issue:** "Cannot open: File exists" errors blocking all CI runs  
**Impact:** Sprint 9 delayed, wasted 4+ hours debugging CI instead of shipping

### Problem Summary
- Tests PASS locally and in CI logs (all green)
- CI job fails with mysterious "Cannot open: File exists" error
- Not a Go error - likely GitHub Actions infrastructure or file contention
- Attempted fixes: `-p 1`, removed `-race`, combined test/coverage steps

### Decision
**Disable test step in CI for feature branches until proper review.**
- Keep: Build verification (cross-platform)
- Keep: Lint checks
- Disable: Test execution in CI (run locally instead)
- Review: Full CI/CD pipeline with fresh eyes

### Action Items
| Task | Owner | Due | Priority |
|------|-------|-----|----------|
| Review CI/CD workflow from scratch | DevOps/Team | Sprint 10 | P0 |
| Document local test requirements | Documentation | Sprint 10 | P1 |
| Consider alternatives (BuildKite, etc) | DevOps | Post-v0.8.0 | P2 |
| Re-enable CI tests after fix | Team | When stable | P0 |

**Reference:** See commit history `feature/sprint9-cli-polish` for attempted fixes.

---

**Source**: Comprehensive Gap Analysis + Product Manager Responsibilities + Musk's Laws Review + README Promise Audit
**Source**: Comprehensive Gap Analysis + Product Manager Responsibilities + Musk's Laws Review + README Promise Audit
**Status**: Sprint-ready tasks with effort estimates + Quick Wins Added
**Audit Reference**: See `docs/development/V0.7.0_README_AUDIT.md` for full promise gap analysis

---

## Product Manager Responsibilities (NEW)

To prevent roadmap gaps like v0.7.0/v0.8.0 promises being missed, the Product Manager role includes:

### Weekly Reviews
| Task | Frequency | Tool | Output |
|------|-----------|------|--------|
| README.md Promise Audit | Weekly | Manual diff | Gap report |
| Log Bundle Repo Watch | Daily | GitHub API / Grok-4.1 | Changelog summary |
| Breaking Change Detection | Real-time | Commit hooks | Alert to #suse-rancher-esc |
| PR Review Coverage | Per-PR | GitHub Actions | Quality gate block |

### Specific Responsibilities
1. **README.md Promise Tracking**
   - Compare `README.md` "What's Coming" section against `ROADMAP_UPDATES.md` active tasks
   - Flag any promises without corresponding sprint tasks within 24 hours
   - Musk's Law #1 (QUESTION): "Is every README promise tracked in a sprint?"

2. **Log Bundle Repo Monitoring**
   - Watch https://github.com/rancher/support-bundle-kit for breaking changes
   - Parse commit messages for new features affecting r8s parsers
   - Grok-4.1-fast research for impact analysis
   - Output: Weekly digest posted to topic #55

3. **Release Readiness Validation**
   - Before any release tag: verify all README promises for that version are complete
   - Block releases with unfulfilled promises
   - Musk's Law #2 (DELETE): Remove promises that won't ship, don't hide them

4. **AI Feature Coordination**
   - Track AI/LLM integration roadmap (see AI Features section below)
   - Coordinate with model providers (Ollama, OpenAI, Anthropic)
   - Monitor AI model capability changes that could enhance r8s

---

## Gap Analysis: v0.7.0 README Promises vs Current Roadmap

### MISSING from current roadmap - to be added:

| Promise | Priority | Sprint Assignment | Musk's Law |
|---------|----------|-------------------|------------|
| **Storage: PersistentVolumes, PVCs, StatefulSets** | CRITICAL | S4-HIGH-1 | 🧠 QUESTION (Why wasn't this in v0.7.0?) |
| **System Health: dmesg analysis (OOM kills)** | HIGH | S4-HIGH-2 | 🧠 QUESTION (Core troubleshooting need) |
| **Control Plane: RKE2 server logs from journald** | CRITICAL | S4-HIGH-3 | 🧠 QUESTION (Essential for RKE2) |
| **Configuration: ConfigMaps, HelmCharts** | MEDIUM | S4-MEDIUM-1 | ⚙️ SIMPLIFY (Reduce config debugging time) |
| **Bundle Completeness: % complete indicator** | HIGH | S4-HIGH-4 | 🚀 ACCELERATE (Faster triage) |
| **Cache Optimization: 800MB+ bundle speed** | CRITICAL | S5-CRITICAL-1 | 🚀 ACCELERATE (Performance) |
| **Namespace Health: Pre-computed rankings** | MEDIUM | S5-HIGH-1 | 🚀 ACCELERATE (UX) |
| **CI/CD Pipeline: Automated testing** | CRITICAL | S4-CRITICAL-1 | 🤖 AUTOMATE (Quality gates) |
| **50% Test Coverage: Enforced automatically** | HIGH | S4-HIGH-5 | 🤖 AUTOMATE (Prevent regression) |

---

## SPRINT TIMELINE

### ✅ SPRINT 1 COMPLETE: Performance & Build Optimization
**Week**: Feb 10, 2026
**Status**: 🎉 ALL TASKS COMPLETE
**Total Effort**: ~8 hours

| Task | Status | Musk's Law |
|------|--------|------------|
| S1-CRITICAL-1: Delete Demo Bundle | ✅ Complete | 🗑️ DELETE |
| S1-HIGH-1: Fix UI Blocking | ✅ Complete | 🚀 ACCELERATE |
| S1-HIGH-2: Pre-Commit Hooks | ✅ Complete | 🤖 AUTOMATE |

---

### ✅ SPRINT 2: Code Quality & Testing (COMPLETE)
**Week**: Feb 10-14, 2026
**Status**: ✅ ALL TASKS COMPLETE

| Task | Status | Musk's Law |
|------|--------|------------|
| S2-HIGH-1: Consolidate Attention Detectors | ✅ Complete | ⚙️ SIMPLIFY |
| S2-MEDIUM-1: Test Coverage Reporting | ✅ Complete | 🤖 AUTOMATE |

---

### 🔄 SPRINT 3: Feature Completeness (ACTIVE)
**Week**: Feb 11-14, 2026
**Status**: 2/4 COMPLETE

| Task | Status | Musk's Law |
|------|--------|------------|
| S3-MEDIUM-1: Multi-Container Pod Support | ✅ Complete | 🧠 QUESTION |
| S3-MEDIUM-2: OOM QoS Analysis | ✅ Complete | 🧠 QUESTION |
| S3-MEDIUM-3: OOM Node Pressure Analysis | 🔄 In Progress | 🧠 QUESTION |
| S3-MEDIUM-4: test-cluster Subcommand | 🔄 In Progress | 🤖 AUTOMATE |

---

### 🆕 SPRINT 4: Foundation Sprint (March 2026) - REPRIORITIZED + QUICK WINS
**Theme**: Low Effort, High Gain + Future-Proofing + Ship Value Faster
**Duration**: 1 week (expanded with quick wins)
**Total Effort**: ~19 hours (was 12h — added 7h of quick wins)

#### Sprint 4A: Infrastructure (Required) — 12h
| ID | Task | Effort | Musk's Law | Rationale |
|----|------|--------|------------|-----------|
| S4-CRITICAL-1 | CI/CD Pipeline (GitHub Actions) | 4h | 🤖 AUTOMATE | Block releases without tests |
| S4-HIGH-4 | Bundle completeness indicator | 4h | 🚀 ACCELERATE | Transparency for users |
| S4-HIGH-5 | 50% coverage enforcement | 4h | 🤖 AUTOMATE | Quality gate automation |

#### Sprint 4B: Quick Wins (Ship Value Faster) — 7h
| ID | Task | Effort | README Link | Musk's Law |
|----|------|--------|-------------|------------|
| S4-QUICK-1 | CodeRabbit polish fixes (CR#10,12,13,7) | 3h | Polish | ⚙️ SIMPLIFY |
| S4-QUICK-2 | Minimal dmesg parser (OOM kills only) | 2h | Promise #2 partial | 🚀 ACCELERATE |
| S4-QUICK-3 | File-count bundle completeness v1 | 2h | Promise #5 fast version | 🚀 ACCELERATE |

**Why Quick Wins Matter:**
- Deliver user value while infrastructure work is in progress
- Break complexity into manageable bites
- Maintain momentum with visible progress
- Prevent "all infrastructure, no features" trap

**Principle Applied:** See `PRINCIPLES.md` #13 — "Ship Value in Manageable Bites"

**Explicitly Deferred to Sprint 6:**
| ID | Promise | README Ref | Effort | Deferral Reason |
|----|---------|------------|--------|-----------------|
| ~~S4-HIGH-1~~ | Storage: PV/PVC/StatefulSet parsers | Promise #1 | 8h | Heavy parsing—needs dedicated sprint |
| ~~S4-HIGH-2~~ | System Health: Full dmesg analysis | Promise #2 | 6h | Partial coverage via S4-QUICK-2 |
| ~~S4-HIGH-3~~ | Control Plane: RKE2 journald parser | Promise #3 | 6h | Complex format—Sprint 5 candidate |
| ~~S4-MEDIUM-1~~ | ConfigMaps + HelmCharts | Promise #4 | 5h | Lower priority, user demand unclear |

**README Promise Gap Tracking:**
| Promise | Status | Sprint | Notes |
|---------|--------|--------|-------|
| #1 Storage (PV/PVC) | 🔴 DEFERRED | Sprint 6 | 90% coverage promise at risk |
| #2 System Health (dmesg) | 🟡 PARTIAL | Sprint 4B | OOM-only via S4-QUICK-2 |
| #3 Control Plane (journald) | 🔴 DEFERRED | Sprint 6 | RKE2 support incomplete |
| #4 ConfigMaps/Helm | 🔴 DEFERRED | Sprint 6 | Lower priority |
| #5 Bundle Completeness | 🟢 ON TRACK | Sprint 4A | Full implementation |
| #6 Cache Optimization | 🟡 MOVED | Sprint 5 | v0.7.x maintenance release |
| #7 Async Operations | 🟢 ON TRACK | Sprint 5 | Partially done Sprint 3 |
| #8 Namespace Health | 🟡 MOVED | Sprint 5 | v0.7.x maintenance release |
| #9 CI/CD Pipeline | 🟢 ON TRACK | Sprint 4A | Core infrastructure |
| #10 50% Test Coverage | 🟢 ON TRACK | Sprint 4A | Quality gate automation |

**Success Criteria**:
- CI/CD runs on every PR
- Bundle completeness shown on startup (file-count v1 acceptable)
- Coverage enforcement blocks <50% PRs
- Sprint 3 branches merged cleanly
- CodeRabbit high-priority items resolved
- dmesg OOM detection working (minimal version)

**Audit Reference:** See `docs/development/V0.7.0_README_AUDIT.md` for full gap analysis

---

### 🆕 SPRINT 5: Performance Optimization (April 2026)
**Theme**: Achieve v0.7.0 performance targets
**Duration**: 2 weeks
**Total Effort**: ~25 hours

| ID | Task | Effort | Musk's Law | Rationale |
|----|------|--------|------------|-----------|
| S5-CRITICAL-1 | Cache optimization (800MB+ bundles) | 10h | 🚀 ACCELERATE | <2s dashboard load target |
| S5-HIGH-1 | Namespace health pre-computation | 6h | 🚀 ACCELERATE | Rank worst-first instantly |
| S5-HIGH-2 | Memory optimization (<500MB target) | 5h | 🚀 ACCELERATE | Large bundle handling |
| S5-MEDIUM-1 | Async resource loading expansion | 4h | 🚀 ACCELERATE | General UI responsiveness |

**Success Criteria**:
- Dashboard loads <2s with 1000 pods
- Memory <500MB for 800MB bundles
- All async operations non-blocking

---

### 🆕 SPRINT 6: v0.8.0 Production Hardening (May-June 2026)
**Theme**: Enterprise-ready reliability + complete bundle analysis
**Duration**: 4 weeks
**Total Effort**: ~40 hours

| ID | Task | Effort | Musk's Law | Rationale |
|----|------|--------|------------|-----------|
| S6-CRITICAL-1 | 80% test coverage achievement | 15h | 🤖 AUTOMATE | Production-grade QA |
| S6-HIGH-1 | Networking: Ingress, Endpoints, Service→Pod | 8h | ⚙️ SIMPLIFY | Network debugging |
| S6-HIGH-2 | Workloads: Jobs, CronJobs, ReplicaSets, HPA | 8h | ⚙️ SIMPLIFY | Complete workload view |
| S6-HIGH-3 | etcd Health: alarms, member status | 6h | 🧠 QUESTION | Critical cluster health |
| S6-MEDIUM-1 | Bundle health scoring (auto quality assessment) | 4h | 🤖 AUTOMATE | "72% complete - missing journald" |
| S6-MEDIUM-2 | Stress testing (1000+ pods, 10M log lines) | 4h | 🧠 QUESTION | Validate at 10× scale |

**Success Criteria**:
- 80%+ test coverage enforced
- <2s dashboard load with 1000 pods
- Zero known critical bugs
- Complete documentation (troubleshooting playbooks)

---

## AI FEATURES ROADMAP (Detailed)

AI features are documented in FUTURE_WORK.md. Musk's Law mapping:

### v0.8.x: AI Phase 1 - AI-Ready Export
| Feature | Task | Musk's Law | Rationale |
|---------|------|------------|-----------|
| `r8s analyze` command | S7-HIGH-1 | 🧠 QUESTION | What data helps LLMs diagnose? |
| Markdown/JSON export | S7-HIGH-2 | ⚙️ SIMPLIFY | Structured signal, not noise |
| Bundle summary format | S7-HIGH-3 | ⚙️ SIMPLIFY | Fits LLM context windows |

**Principle**: Truth Only™ - structured data for LLM analysis, not fabricated

### v0.9.0: AI Phase 2 - Local AI via Ollama
| Feature | Task | Musk's Law | Rationale |
|---------|------|------------|-----------|
| Ollama auto-detection | S8-CRITICAL-1 | 🤖 AUTOMATE | Zero-config AI integration |
| "🤖 AI Analysis" TUI panel | S8-HIGH-1 | 🧠 QUESTION | How do users want AI insights? |
| Natural language diagnosis | S8-HIGH-2 | 🧠 QUESTION | What explanations help? |
| Privacy-first (local only) | S8-HIGH-3 | 🗑️ DELETE | No cloud data transmission |

**Principle**: Explicit > Implicit - AI status clearly shown, user controls

### v0.9.2: AI Phase 3 - Multi-Provider Architecture
| Feature | Task | Musk's Law | Rationale |
|---------|------|------------|-----------|
| Analyzer interface | S9-HIGH-1 | ⚙️ SIMPLIFY | One interface, many providers |
| OpenAI/Claude plugins | S9-MEDIUM-1 | 🤖 AUTOMATE | User choice of AI provider |
| Analysis caching | S9-MEDIUM-2 | 🚀 ACCELERATE | Avoid re-analyzing bundles |
| Custom prompt templates | S9-LOW-1 | 🤖 AUTOMATE | User-defined analysis patterns |

**Principle**: Use Your Interfaces - trust the abstraction

---

## MUSK'S 5 LAWS ROADMAP REVIEW

All roadmap items reviewed against Elon's 5 Laws:

| Law | Applied To | Count |
|-----|-----------|-------|
| **🧠 QUESTION** | Requirements, AI features, bundle completeness | 8 items |
| **🗑️ DELETE** | Live mode removal, mock data removal, dead code | 4 items |
| **⚙️ SIMPLIFY** | Strategy pattern refactor, AI interfaces, data views | 6 items |
| **🚀 ACCELERATE** | Performance, async loading, caching, pre-computation | 7 items |
| **🤖 AUTOMATE** | CI/CD, coverage gates, principle compliance, AI | 8 items |

### Anti-Patterns Flagged for Review
- S3-MEDIUM-4 (test-cluster): May violate "Best Feature = No Feature" - is this necessary?
- AI Phase 3: Risk of over-engineering - ensure Phase 1 & 2 prove value first

---

## DOCUMENTATION TO REVIEW WITH MUSK'S LAWS

All docs need review against the 5 Laws:

| Document | Primary Law | Action |
|----------|-------------|--------|
| PRINCIPLES.md | All | Ensure alignment with Laws |
| MAXIMUM_INFO_EXTRACTION_PLAN.md | 🧠 QUESTION | Is 90% extraction necessary? |
| FUTURE_WORK.md | 🗑️ DELETE | Remove deferred features or commit |
| CONTRIBUTING.md | 🤖 AUTOMATE | Add automated checks |
| TROUBLESHOOTING.md | ⚙️ SIMPLIFY | Simplify common paths |

**Action**: Schedule 2-hour Musk's Laws review session for all docs

---

## COMPLETE SPRINT TIMELINE (Updated)

```text
2026 Timeline:

Jan ████ v0.6.8 (current)
Feb ████████ Sprint 1, 2, 3 (COMPLETE + IN PROGRESS)
     └── S1: Demo bundle delete, UI blocking, pre-commit
     └── S2: Attention detector consolidation, coverage reporting
     └── S3: Multi-container, OOM QoS, test-cluster
Mar ████████████ Sprint 4 (Data Coverage)
     └── PV/PVC, journald, dmesg, CI/CD, 50% coverage
Apr ████████ Sprint 5 (Performance)
     └── Cache optimization, namespace rankings, <2s load
May ████████████ Sprint 6 (v0.8.0 Production)
     └── 80% coverage, networking, etcd, stress testing
Jun ████████ Sprint 7 (AI Phase 1)
     └── r8s analyze command, structured export
Jul ████████████ Sprint 8 (AI Phase 2)
     └── Ollama integration, local AI analysis
Aug ████████ Sprint 9 (AI Phase 3)
     └── Multi-provider, analysis history
Sep ████ v1.0 planning
```

---

## BACKLOG: Technical Debt + Deferred Work

### Technical Debt
| ID | Task | Priority | Musk's Law |
|----|------|----------|------------|
| BACKLOG-1 | Remove Unused Rancher API Structs | MEDIUM | 🗑️ DELETE |
| BACKLOG-2 | Audit Redundant Error Wrapping | LOW | ⚙️ SIMPLIFY |
| BACKLOG-3 | Automate Changelog Generation | LOW | 🤖 AUTOMATE |

### Deferred from Sprint 4 (Add When Requested)
| ID | Task | Effort | Trigger |
|----|------|--------|---------|
| BACKLOG-4 | Storage: PV/PVC/StatefulSet parsers | 8h | Customer asks for storage debugging |
| BACKLOG-5 | System Health: dmesg OOM detection | 6h | Need kernel-level diagnostics |
| BACKLOG-6 | Control Plane: RKE2 journald parser | 6h | Control plane issues in support cases |
| BACKLOG-7 | ConfigMaps + HelmCharts parsing | 5h | Config debugging becomes priority |

---

**Roadmap Maintained By**: Development Team + Product Manager
**Last Updated**: 2026-02-12
**Next Review**: 2026-02-19 09:00 AM AEST (Weekly sprint review)
