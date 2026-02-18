# CLI-First Pivot: Team Review & Roadmap Validation

**Date:** 2026-02-17  
**Status:** Sprint 8 Mid-Pivot Review  
**Revelation:** CLI-first architecture (TUI becomes thin wrapper)

---

## The Revelation

**From today's debugging session:**
- User (DontStop) discovered `r8s validate` is powerful
- Support engineers need kubectl-style tools, not TUI navigation
- CLI is 80% of the value, TUI is 20%
- Pipe-to-AI workflows are the killer feature

**New Direction:** CLI-first, TUI-second

---

## Team Input Required

### 🎪 Luna (Ringmaster/Release Manager)

**Questions for Luna:**
1. Does CLI-first fit v0.7.x roadmap or push to v0.8.0?
2. Resource allocation: Can we spare UX Engineer for CLI review?
3. Risk assessment: Stripping TUI mid-sprint vs finishing Sprint 8 first?

**Luna's Decision Framework:**
- User value: HIGH (support engineers want CLI)
- Technical risk: MEDIUM (refactoring, not rewriting)
- Timeline impact: +3-5 days to Sprint 8
- Musk's Law #2: TUI is 9,360 lines → CLI target is <3,000

### 🎨 UX Engineer

**Review Request:** CLI Command Structure

Current commands:
```
r8s validate [path]              # ✅ Done
r8s generate prompt [path]       # ✅ Done
r8s analyze [path]               # ⬜ Sprint 8 Day 3
r8s export [path]                # ⬜ Sprint 8 Day 4-5
r8s logs [pod] [path]            # ⬜ Future
r8s describe [resource] [path]   # ⬜ Future
```

**UX Questions:**
1. Consistency: Should all commands accept `[path]` as first or last arg?
2. Flags: `--format`, `--output`, `--filter` — consistent naming?
3. Output: Auto-detect TTY vs pipe? (color for humans, JSON for scripts)
4. Exit codes: 0=ok, 1=issues, 2=error — intuitive?

**Deliverable:** CLI style guide + 80/20 quick wins

### ⚡ RancherSRE (Technical Lead)

**Technical Assessment:**

**Pros:**
- Faster development (no TUI state management)
- Easier testing (CLI is scriptable)
- Better for CI/CD integration
- Pipe-to-AI is natural with CLI

**Cons:**
- Dashboard still valuable for first impression
- Some users prefer TUI navigation
- Breaking change for existing users

**Recommendation:** Hybrid approach
- CLI: Full feature set (6+ commands)
- TUI: Dashboard only (strip 80% of views)
- Migration: Gradual, not abrupt

**80/20 Wins:**
1. Strip dead TUI views (ViewClusters, ViewProjects, ViewNamespaces) → 400 lines
2. Simplify sort modes (remove Count) → 30 lines  
3. Add 4 CLI commands → high value, medium effort
4. Dashboard-only TUI → keep the 20% that matters

---

## Roadmap Fit Analysis

### Current Roadmap (ROADMAP_v0.7.x.md)

| Version | Original Scope | CLI-First Pivot |
|---------|---------------|-----------------|
| v0.7.2 | Bundle Health + AI + CLI (Sprint 8) | ✅ Keep |
| v0.7.3 | AI Analysis Engine v1 | ✅ CLI integration instead |
| v0.7.4 | RKE1 + AI Assistant | ⬜ Evaluate need |
| v0.8.0 | Production Hardening | ⬜ CLI-first architecture |

### Recommendation: Adjust v0.7.x

**v0.7.2 (Current Sprint 8):**
- ✅ Finish: `validate`, `generate prompt`
- ⬜ Add: `analyze`, `export` (CLI only)
- ⬜ Strip: TUI to dashboard-only
- Target: "CLI MVP + Dashboard"

**v0.7.3 (Next):**
- Scope: CLI polish + shell completion
- Add: `logs`, `describe` commands
- TUI: Freeze (no new features)

**v0.8.0:**
- Full CLI-first architecture
- TUI deprecated (or removed)
- Enterprise features (CI plugins, etc.)

---

## 80/20 Analysis: CLI Opportunities

### High Impact, Low Effort (Do First)

| Win | Effort | Impact | Owner |
|-----|--------|--------|-------|
| Consistent `--format` flag | 2 hrs | HIGH | RancherSRE |
| Auto-detect TTY vs pipe | 3 hrs | HIGH | RancherSRE |
| Command aliases | 1 hr | MEDIUM | RancherSRE |
| Strip dead TUI views | 1 day | HIGH | RancherSRE |
| Help text template | 2 hrs | MEDIUM | UX Engineer |

### High Impact, Medium Effort (Do Next)

| Feature | Effort | Impact | Owner |
|---------|--------|--------|-------|
| `r8s analyze` command | 2 days | HIGH | RancherSRE |
| `r8s export` command | 2 days | HIGH | RancherSRE |
| Dashboard-only TUI | 2 days | MEDIUM | RancherSRE |
| Shell completion | 1 day | HIGH | RancherSRE |

### Low Impact or High Effort (Defer)

| Feature | Effort | Impact | Decision |
|---------|--------|--------|----------|
| `r8s logs` streaming | 3 days | MEDIUM | v0.7.3 |
| `r8s describe` full | 3 days | MEDIUM | v0.7.3 |
| Man pages | 2 days | LOW | v0.8.0 |
| Remove TUI entirely | 5 days | HIGH (risk) | v0.8.0 |

---

## Team Consensus Needed

### Decision 1: Sprint 8 Scope Adjustment

**Option A:** Finish Sprint 8 as planned (TUI-heavy + 2 CLI commands) → v0.7.2
**Option B:** Pivot mid-Sprint 8 to CLI-first (4 CLI commands + dashboard-only TUI) → v0.7.2-extended
**Option C:** Finish Sprint 8 minimal, full CLI pivot in Sprint 9 → v0.7.2 + v0.7.3

**Recommendation:** Option B (adjust now, user wants CLI)

### Decision 2: TUI Fate

**Option A:** Strip TUI to dashboard-only now (Sprint 8)
**Option B:** Freeze TUI, focus on CLI (Sprint 8-9)
**Option C:** Deprecate TUI gradually (v0.8.0)

**Recommendation:** Option A (strip dead code now, per Musk's Law #2)

### Decision 3: CLI Target

**Minimal (v0.7.2):** `validate`, `generate`, `analyze`, `export`
**Standard (v0.7.3):** + `logs`, `describe`, `dashboard`
**Full (v0.8.0):** + scripting, completion, plugins

**Recommendation:** Minimal for v0.7.2, Standard for v0.7.3

---

## Action Items

### Immediate (Today)
- [ ] Luna: Approve Sprint 8 scope adjustment
- [ ] UX Engineer: Review CLI structure, provide 80/20 wins
- [ ] RancherSRE: Implement `analyze` and `export` commands

### This Week
- [ ] RancherSRE: Strip dead TUI views
- [ ] UX Engineer: CLI style guide document
- [ ] Luna: Update roadmap (ROADMAP_v0.7.x.md)

### Next Sprint
- [ ] Team: Sprint 9 planning (CLI polish)
- [ ] RancherSRE: Shell completion
- [ ] UX Engineer: Final CLI review

---

## User Quote (DontStop)

> "BUNDLE Health command line is amazing. Support engineers and customers that know enough to be dangerous would appreciate some features as they do with kubectl and others."

**Translation:** CLI is the product. TUI is a dashboard viewer.

---

## Decision Required From Luna

**Approve CLI-first pivot for Sprint 8?**

**Yes** → Adjust scope, add 2 CLI commands, strip TUI dead code
**No** → Finish Sprint 8 as planned, CLI pivot in Sprint 9

**Default:** Yes (user has spoken)

---

*Prepared by: RancherSRE*  
*For: Luna (Ringmaster), UX Engineer, DontStop (Product)*
