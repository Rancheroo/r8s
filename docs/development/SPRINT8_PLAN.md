# Sprint 8 Plan: AI Analysis Engine v1 (Pattern Matching)

**Sprint Goal:** Deliver foundational AI features — pattern matching engine and knowledge base for intelligent issue grouping.

**Duration:** 2 weeks (March 3 - March 14, 2026)  
**Target Release:** v0.8.0 (March 15, 2026)  
**Strategic Rationale:** RKE1 is EOL; AI features deliver more user value than legacy distro support.

---

## 📋 Scope (MVP per Strategic Brief)

### P0: Pattern Matching Engine (Days 1-5)
- Design pattern matcher interface
- Implement regex + keyword matching
- Create issue grouper (groups related errors)
- Smoke tests with sample bundles

### P1: Knowledge Base v1 (Days 6-8)
- Define pattern schema (JSON/YAML)
- Create 10 core patterns:
  - etcd leader election issues
  - OOMKill patterns
  - ImagePullBackOff
  - CrashLoopBackOff
  - Node pressure (memory/disk/PID)
  - Certificate expiry
  - API server connectivity
  - DNS resolution failures
  - PersistentVolume mount issues
  - Network policy blocks
- Knowledge base loader
- Unit tests for all patterns

### P2: Integration with TUI (Days 9-10)
- Show pattern matches in diagnostic panel
- Group related issues in Attention Dashboard
- Basic severity scoring (info/warning/critical)

### P3: Documentation + Polish (Days 11-14)
- Pattern authoring guide
- Update README with AI features
- CodeRabbit review items
- Final integration testing

---

## 🤖 Smart Work Distribution

| Role | Allocation | Focus |
|------|------------|-------|
| RancherSRE | 70% | Pattern engine, knowledge base, integration |
| CodeRabbit | Continuous | Review all PRs, catch regressions |
| Documentation | Days 12-14 | Pattern authoring guide, README updates |

---

## 📊 Success Criteria

| Metric | Target |
|--------|--------|
| Pattern Matcher | Interface + 2 implementations (regex, keyword) |
| Knowledge Base | 10 patterns with 80%+ accuracy on test bundles |
| Coverage | Maintain 36.8%+ (focus on new code) |
| CI | All checks green (no flaky jobs) |

---

## 🗓 Timeline

| Week | Focus | Key Milestone |
|------|-------|---------------|
| **Week 1** | Engine + KB | Pattern matcher working, 5 patterns defined |
| Mon-Tue | Pattern interface | Matcher design, regex implementation |
| Wed-Fri | Knowledge base | 10 patterns, loader, tests |
| **Week 2** | Integration + Docs | AI features visible in TUI |
| Mon-Tue | TUI integration | Diagnostic panel shows pattern matches |
| Wed-Thu | Documentation | Pattern authoring guide complete |
| Fri | Polish + Release | v0.8.0 tagged |

---

## ⚠️ Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Pattern accuracy low | High | Start with high-confidence patterns (OOM, ImagePull) |
| Performance impact | Medium | Compile patterns once, cache results |
| False positives | Medium | Severity scoring, user feedback loop |

---

## 🚫 Out of Scope (v0.8.1+)

- Root cause hints (needs more pattern validation)
- Anomaly detection (needs baseline data)
- Natural language queries
- ML/semantic matching

---

## ✅ Definition of Done

- [ ] Pattern matcher interface implemented
- [ ] 10 patterns in knowledge base
- [ ] Pattern matches display in TUI
- [ ] Coverage ≥ 36.8%
- [ ] All CI checks passing
- [ ] Pattern authoring guide published
- [ ] README updated with AI features
- [ ] v0.8.0 tagged and released

---

## Why AI Over RKE1?

**RKE1 Status:** End-of-Life, declining user base  
**AI Value:** Immediate productivity gains for ALL users (RKE2 + K3s)  
**Strategic:** Positions r8s as intelligent diagnostic tool, not just log viewer

**Musk's Law #3 (Simplify):** 10 patterns that work > 100 patterns that don't.

---

*Plan approved? Pick it up after Sprint 7 release.*
