# r8s v1.0 Roadmap
**The Path to Production-Ready**

**Status:** DRAFT for Monday Standup Review  
**Last Updated:** 2026-02-14  
**Approach:** Sequential Sprints (Quality First)

---

## 🎯 v1.0 Vision

**r8s** is the definitive Rancher support bundle analysis tool:
- ✅ Parse any Rancher/RKE2/K3s bundle (90%+ coverage)
- ✅ Detect issues automatically (AI-assisted analysis)
- ✅ Guide users to root cause (wizard interface)
- ✅ Work offline, work fast, work reliably

---

## 📅 Release Timeline

### Current State: v0.7.0 SHIPPED ✅
- 8 parsers delivered
- 90% bundle coverage
- CodeRabbit integrated
- CI partially disabled (tech debt)

### Road to v1.0

| Release | Target Date | Focus | Key Deliverables |
|---------|-------------|-------|------------------|
| **v0.7.1** | Mar 2 | K3s Support | Format detection, 5-file refactor |
| **v0.7.2** | Mar 9 | AI Foundations | Pattern grouping, 10 patterns |
| **v0.7.3** | Mar 23 | RKE1 + Assistant | RKE1 support, debugging wizard |
| **v0.8.0** | Apr 6 | Performance | <2s dashboard, parallel parsing |
| **v0.9.0** | Apr 20 | Enterprise | Config mgmt, auth, reporting |
| **v1.0.0** | May 11 | Stable Release | Production-ready, full docs |

---

## Detailed Release Plans

### v0.7.1 — K3s Support (March 2)
**Scope:** REDUCED from original 18-file refactor

**In:**
- K3s format detection (`DetectFormat()` enhancement)
- 5 core files path abstraction:
  - `types.go` - Add `FormatK3s`, path helpers
  - `manifest.go` - Use `b.KubectlPath()`
  - `validate.go` - Dynamic distro validation
  - `journald.go` - Service name mapping
  - `completeness.go` - Dynamic path generation

**Out:**
- ❌ RKE1 support (moved to v0.7.3)
- ❌ Full 18-file refactor (risk reduction)
- ❌ RKE1 service names (not needed yet)

**Contractor Need:** None (backend only)

---

### v0.7.2 — AI Foundations (March 9)
**Scope:** MVP only (pattern grouping)

**In:**
- Pattern matching engine
- Knowledge base (10 common patterns):
  - etcd leader election issues
  - OOM kills detected
  - Certificate expiration
  - Image pull failures
  - CrashLoopBackOff
  - Node pressure
  - Network plugin errors
  - API server timeouts
  - Pod scheduling failures
  - Storage mount issues

**Out:**
- ❌ Root cause hints (defer to v0.7.3)
- ❌ Anomaly detection (defer to v0.8.0)
- ❌ Natural language queries (defer to v0.7.3)

**Contractor Need:** None (backend patterns)

---

### v0.7.3 — RKE1 + AI Assistant (March 23)
**Scope:** Deferred work consolidation

**In:**
- RKE1 format support (full)
- Guided debugging wizard (TUI)
- Basic report generation
- Natural language queries (simple)

**Out:**
- ❌ Complex NLQ (defer to v0.9.0)
- ❌ Advanced AI analysis (defer to v0.8.0)

**Contractor Need:** 
- 🎨 **TUI/UX Expert** (2-3 weeks) — Debugging wizard UX

---

### v0.8.0 — Performance (April 6)
**Scope:** Speed optimization

**In:**
- Parallel bundle parsing
- <2s dashboard target
- Memory optimization
- Benchmarking suite

**Contractor Need:** None

---

### v0.9.0 — Enterprise Features (April 20)
**Scope:** Production deployment ready

**In:**
- Configuration management
- Authentication (basic)
- Report exports (PDF, HTML)
- Team collaboration features

**Contractor Need:**
- 👔 **Product Manager** (part-time) — Feature prioritization, user research

---

### v1.0.0 — Stable Release (May 11)
**Scope:** Production-ready

**In:**
- Complete documentation
- Stable API
- Full test coverage (80%+)
- Security audit
- Performance benchmarks
- User guides

**Contractor Need:**
- 📝 **Technical Writer** (2-3 weeks) — Documentation polish

---

## Contractor Requirements

### 1. TUI/UX Expert (v0.7.3)
**Duration:** 2-3 weeks  
**Budget:** $$  
**Skills:**
- Bubble Tea (Go TUI framework) expertise
- Terminal UI design patterns
- User flow optimization
- Accessibility in terminal apps

**Deliverables:**
- Debugging wizard interface
- Navigation improvements
- Help system design
- Keyboard shortcut optimization

**When:** March 2026

---

### 2. Product Manager (v0.9.0)
**Duration:** Part-time, 4-6 weeks  
**Budget:** $  
**Skills:**
- Developer tool product management
- Open source community engagement
- User research and interviews
- Feature prioritization frameworks

**Deliverables:**
- User personas and journeys
- Feature prioritization matrix
- Community engagement plan
- v1.0+ roadmap planning

**When:** April 2026

---

### 3. Technical Writer (v1.0)
**Duration:** 2-3 weeks  
**Budget:** $  
**Skills:**
- Developer documentation
- CLI tool documentation
- Tutorial and guide writing
- README and landing page copy

**Deliverables:**
- Complete user guide
- API reference
- Quick start tutorial
- Troubleshooting guide

**When:** May 2026

---

## Success Metrics by Release

| Release | Coverage | CI Status | Performance | Distros | AI Features |
|---------|----------|-----------|-------------|---------|-------------|
| v0.7.0 | 90% | Partial | Baseline | 1 | 0 |
| v0.7.1 | 90% | ✅ 100% | Baseline | 2 | 0 |
| v0.7.2 | 90% | ✅ 100% | Baseline | 2 | 1 (grouping) |
| v0.7.3 | 95% | ✅ 100% | Baseline | 3 | 2 (wizard) |
| v0.8.0 | 95% | ✅ 100% | <2s | 3 | 2 |
| v0.9.0 | 95% | ✅ 100% | <2s | 3 | 3 |
| v1.0.0 | 95%+ | ✅ 100% | <2s | 3 | 4 |

---

## Risk Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| K3s format differs significantly | Medium | High | Research phase before coding; collect sample bundles |
| AI features underwhelm users | Medium | Medium | Ship pattern grouping first; validate before expanding |
| Contractor availability | Medium | Medium | Start recruitment 2 weeks early; have backup plans |
| Scope creep | High | High | Strict sequential approach; no parallel tracks |
| Performance targets unachievable | Low | Medium | Benchmark before optimizing; set realistic targets |

---

## Current Sprint (Sprint 6) Alignment

**Sprint 6 (Now):** CI Stability ONLY  
**NOT in Sprint 6:** K3s, AI, RKE1, performance work  

**Why this matters:**
- CI passing = foundation for everything else
- No K3s work until CI is green (approved sequential timeline)
- Quality debt must be paid before compounding

See: `SPRINT6_CI_STABILITY.md`

---

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-02-13 | Sequential timeline approved | Quality first; reduce integration risk |
| 2026-02-13 | K3s scope reduced | 5 files vs 18; RKE1 deferred to v0.7.3 |
| 2026-02-13 | AI scope reduced | Pattern grouping only; root cause hints deferred |
| 2026-02-14 | Contractor needs identified | UX expert (v0.7.3), PM (v0.9.0), Writer (v1.0) |

---

## Action Items for Monday Standup

1. **Review v1.0 roadmap** — Approve timeline and scope
2. **Confirm contractor budget** — UX expert, PM, Writer
3. **Sprint 6 kickoff** — CI stability focus (no K3s)
4. **K3s bundle acquisition** — Need sample bundles for v0.7.1
5. **User research planning** — PM contractor to lead (v0.9.0)

---

*This roadmap assumes sequential sprints with no parallel work.*  
*Parallel tracks risk integration failures and should be avoided.*

**Prepared for:** Monday Standup  
**Decision needed:** Approve roadmap, confirm contractor budget
