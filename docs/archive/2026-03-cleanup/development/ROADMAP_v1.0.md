# r8s v1.0 Roadmap
**The Path to Production-Ready**

**Status:** ACTIVE — Sprint 12 In Progress  
**Last Updated:** 2026-02-24  
**Approach:** Sequential Sprints (Quality First)

---

## 🎯 v1.0 Vision

**r8s** is the definitive Rancher support bundle analysis tool:
- ✅ Parse any Rancher/RKE2/K3s bundle (90%+ coverage)
- ✅ Detect issues automatically (AI-assisted analysis)
- ✅ Guide users to root cause (wizard interface)
- ✅ Work offline, work fast, work reliably
- ✅ Delightful CLI experience (loading messages, helpful errors)

---

## 📅 Release Timeline

### Current State: v0.9.0 SHIPPED ✅ (Sprint 11 Complete)
- AI Pattern Engine v2 with confidence scoring
- Root Cause Hint system
- Natural Language Queries (`r8s ask`)
- Export formats (SARIF, JUnit, Markdown, JSON)
- kubectl-style commands (`r8s get`, `r8s logs`, `r8s describe`)
- 10 bundles validated successfully

### Road to v1.0 — REVISED with UX-First Approach

| Release | Target Date | Focus | Key Deliverables |
|---------|-------------|-------|------------------|
| **v0.9.0** | Feb 24 | AI Foundation | Shipped ✅ |
| **v1.0-rc1** | Feb 27 | kubectl Plugin + UX | kubectl-r8s wrapper, loading messages, error improvements |
| **v1.0-rc2** | Mar 2 | Final Polish | Documentation, performance validation, code quality |
| **v1.0.0** | Mar 6 | Stable Release | Production-ready, delightful UX |

### Why UX Before Performance?

**The Insight:** Users won't notice a 200ms optimization, but they WILL notice:
- Fun loading messages that make waiting enjoyable
- Helpful error suggestions when they typo a command
- A kubectl plugin that feels native

**The New Order:**
1. **UX First** — Make it delightful (Sprint 12 Days 1-3)
2. **Polish** — Documentation, testing (Sprint 12 Days 4-5)
3. **Ship** — Release v1.0 (Sprint 12 Day 5)
4. **Performance** — Optimize in v1.1 based on real usage

---

## Sprint 12 Revised Schedule: v1.0 Final Push

**Theme:** UX Polish & kubectl Integration  
**Duration:** 5 days  
**Base:** v0.9.0 validated, kubectl scaffold ready

### Day 1: kubectl Plugin Build ✅ VALIDATED
- [x] kubectl-r8s plugin wrapper implementation
- [x] Bundle auto-detection from environment or filesystem
- [x] Command translation: `kubectl r8s get pods` → `r8s get pods <bundle>`
- [x] Installation documentation

**Validation:** Plugin tested successfully on 10 bundles

### Day 2: UX Improvements — Loading Messages & Error Handling
**Focus:** Make the CLI delightful

**Loading Experience (Cowsay-Style Messages):**
```go
// internal/ui/loading.go
var loadingMessages = []string{
    "🐄 Analyzing bundle... Did you know? The 'R' in RKE2 stands for 'Rancher' (and 'Reliable')",
    "🐮 Parsing logs... Fun fact: Rancher was originally called 'Project Longhorn'",
    "🐄 Detecting patterns... Tip: Run 'r8s ask' for natural language queries",
    "🐮 Scanning for issues... Fun fact: K3s is pronounced 'K3s' (like 'Kates')",
    "🐄 Almost there... Tip: Use --output=json for CI/CD pipelines",
    "🐮 Looking for CrashLoopBackOff... The first rule of Rancher: Don't panic (unless etcd is down)",
    "🐄 Analyzing pod logs... Fun fact: r8s stands for 'Rancher8s' (and 'Great!')",
    "🐮 Checking certificates... Tip: Expired certs are the #1 cause of kubelet failures",
    "🐄 Pattern matching... Fun fact: A support bundle can contain 50,000+ log files",
    "🐮 Finalizing results... Did you know? You can pipe r8s output to 'jq' for custom filtering",
}
```

**Error Message Enhancements:**
- Typo detection with suggestions: `r8s analize` → "Did you mean 'r8s analyze'?"
- British spelling support: `r8s analyse` works too
- Command not found guidance with related commands
- Actionable error messages with examples

**Deliverables:**
- [ ] Loading spinner with rotating fun facts
- [ ] Typo detection and command suggestions
- [ ] British spelling aliases (`analyse` → `analyze`)
- [ ] Context-aware help in error messages

### Day 3: Documentation Polish
**Focus:** Make it discoverable

**README Updates:**
- [ ] Quickstart guide (30-second demo)
- [ ] kubectl plugin installation & usage
- [ ] Loading messages showcase (fun factor!)
- [ ] `r8s ask` natural language examples
- [ ] Export format examples (CI/CD integration)
- [ ] Comparison table: r8s vs kubectl vs dashboard

**CLI Help:**
- [ ] All commands have `--help` with examples
- [ ] Man page stubs
- [ ] Shell completion docs

**Deliverables:**
- [ ] README.md rewrite
- [ ] Installation guide
- [ ] Feature highlights documentation

### Day 4: Final Testing
**Focus:** Validate everything works together

**Test Matrix:**
| Test | Expected | Status |
|------|----------|--------|
| kubectl r8s get pods | Works with auto-detected bundle | ⬜ |
| Loading messages display | Fun facts rotate every 2-3s | ⬜ |
| Typo correction | "analize" suggests "analyze" | ⬜ |
| British spelling | "analyse" works | ⬜ |
| Error messages | Helpful suggestions shown | ⬜ |
| 10 bundles re-tested | All pass | ⬜ |

**Performance Baseline:**
- Document current timing (not optimizing yet, just measuring)
- `r8s analyze` on 100MB bundle: target <5s acceptable for v1.0
- `r8s get pods`: <1s

**Deliverables:**
- [ ] Full integration test suite passed
- [ ] 10 bundles re-validated
- [ ] Performance baseline documented

### Day 5: Release v1.0
**Focus:** Ship it! 🚀

**Release Checklist:**
- [ ] CHANGELOG.md updated with v1.0 highlights
- [ ] Git tag `v1.0.0`
- [ ] GitHub Release with binaries
- [ ] Release notes highlighting:
  - kubectl plugin integration
  - Delightful loading messages
  - AI-powered analysis
  - 10+ pattern library
- [ ] Announcement tweet/toot

**Deliverables:**
- [ ] v1.0.0 tagged and released
- [ ] Documentation live
- [ ] Release announcement published

---

## Detailed Release Plans

### v1.0-rc1 — kubectl Plugin + UX (Feb 27)
**Scope:** Integration and delight

**In:**
- kubectl-r8s plugin wrapper
- Loading messages with personality (cowsay-style)
- Error message improvements (typo detection, suggestions)
- British spelling support

**Out:**
- ❌ Performance optimization (v1.1)
- ❌ Additional patterns beyond current 3 (v1.1)
- ❌ krew distribution (v1.1)

**Contractor Need:** None

---

### v1.0-rc2 — Final Polish (Mar 2)
**Scope:** Documentation and quality

**In:**
- Complete README rewrite
- Installation guides
- CLI help improvements
- Final bug fixes from rc1

**Out:**
- ❌ New features (feature freeze)

**Contractor Need:** None

---

### v1.0.0 — Stable Release (Mar 6)
**Scope:** Production-ready delight

**In:**
- All v1.0-rc2 fixes
- Final validation
- Release packaging

**Success Criteria:**
- kubectl plugin works seamlessly
- Loading messages bring joy
- Error messages are helpful
- Documentation is complete
- 10 bundles validated

**Contractor Need:** None

---

## UX Enhancements Detail

### Loading Experience Design

**Why Cowsay-Style?**
- Rancher has cow/steer branding heritage
- Fun facts reduce perceived wait time
- Personality differentiates from generic tools
- Memorable = shareable

**Implementation:**
```go
// Show random fun fact every 2-3 seconds during analysis
spinner := NewSpinner("🐄 Analyzing your bundle...")
ticker := time.NewTicker(2 * time.Second)
go func() {
    for i, msg := range loadingMessages {
        <-ticker.C
        spinner.UpdateMessage(msg)
    }
}()
```

**Message Categories:**
1. **Fun Facts** — Rancher/K8s history and trivia
2. **Tips** — How to use r8s better
3. **Humor** — Light Rancher/K8s jokes
4. **Progress** — What's currently happening

### Error Handling Design

**Typo Detection:**
```go
// Levenshtein distance for command suggestions
commands := []string{"analyze", "validate", "export", "generate", "logs", "describe", "ask", "get"}
suggest := findClosest(input, commands)
// Output: "Unknown command 'analize'. Did you mean 'r8s analyze'?"
```

**Contextual Help:**
```go
// When user types wrong command, show related commands
Did you mean 'r8s analyze'?

Usage:
  r8s analyze [bundle-path] [flags]

Examples:
  r8s analyze ./support-bundle.tar.gz
  r8s analyze ./bundle --output=json

Or try:
  • r8s validate ./bundle - Quick health check
  • r8s ask ./bundle "why is nginx crashing?" - Natural language query

See all commands: r8s --help
```

**British Spelling:**
- `r8s analyse` → aliases to `r8s analyze`
- Automatic, no user thought required

---

## Success Metrics by Release

| Release | Coverage | CI Status | UX Delight | kubectl Plugin | Validation | Status |
|---------|----------|-----------|------------|----------------|------------|--------|
| v0.9.0 | 90% | ✅ 100% | ⬜ None | ⬜ Missing | ✅ 10 bundles | **SHIPPED** |
| v1.0-rc1 | 90% | ✅ 100% | 🎯 Loading msgs | ✅ Working | ✅ 10 bundles | **IN PROGRESS** |
| v1.0-rc2 | 90% | ✅ 100% | ✅ Complete | ✅ Complete | ✅ 10 bundles | **PLANNED** |
| v1.0.0 | 90%+ | ✅ 100% | ✅ Delightful | ✅ Native | ✅ Validated | **TARGET** |

**New Metric: UX Delight Score**
- v1.0 Target: Users smile during analysis
- Measure: User feedback on loading messages
- Success: >80% positive mentions of UX in feedback

---

## Post-v1.0 Roadmap (v1.1+)

### v1.1 — Performance & More Patterns
- Performance optimization (<2s target)
- Additional pattern library expansion (7+ new patterns)
- kubectl krew distribution
- VS Code extension (stretch)

### v1.2 — Enterprise Features
- Configuration management
- Report exports (PDF, HTML)
- Team collaboration features
- Rancher UI integration research

**But First:** Ship v1.0 with delight.

---

## Risk Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Loading messages annoy power users | Low | Low | Add `--quiet` flag for silent mode |
| kubectl plugin edge cases | Medium | Medium | Extensive testing Day 4 |
| Documentation incomplete | Low | Medium | Day 3 dedicated to docs |
| Scope creep into performance | Medium | High | Strict v1.0 freeze; perf is v1.1 |
| British spelling bugs | Low | Low | Simple alias implementation |

---

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-02-13 | Sequential timeline approved | Quality first; reduce integration risk |
| 2026-02-13 | K3s scope reduced | 5 files vs 18; RKE1 deferred to v0.7.3 |
| 2026-02-13 | AI scope reduced | Pattern grouping only; root cause hints deferred |
| 2026-02-24 | UX before Performance | Users notice delight more than 200ms optimization |
| 2026-02-24 | Loading messages added | Cowsay-style personality matches Rancher brand |
| 2026-02-24 | Typo detection added | Reduces friction for new users |

---

## Action Items for Sprint 12 Completion

1. **Day 2 (UX):** Implement loading messages and error improvements
2. **Day 3 (Docs):** Rewrite README with personality and clarity
3. **Day 4 (Test):** Full validation of kubectl plugin + UX features
4. **Day 5 (Ship):** Tag v1.0.0 and announce

---

## Deliverables Checklist

**PR Contents:**
- [ ] Updated ROADMAP_v1.0.md (this file)
- [ ] Loading messages implementation
- [ ] Error handling improvements
- [ ] kubectl plugin complete
- [ ] Documentation updates
- [ ] Release checklist

**Release Checklist:**
- [ ] CHANGELOG.md updated
- [ ] Git tag `v1.0.0`
- [ ] GitHub Release created
- [ ] Binaries attached
- [ ] Release notes published
- [ ] Announcement posted

---

*This roadmap prioritizes user delight over premature optimization.*  
*Perf is v1.1. Delight is v1.0.*

**Prepared for:** Sprint 12 Execution  
**Next Review:** Post-v1.0 Retrospective
