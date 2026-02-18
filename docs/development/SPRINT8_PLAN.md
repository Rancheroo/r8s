# Sprint 8 Plan: Bundle Health v2 + AI Groundwork (v0.7.2)

**Sprint Goal:** Deliver **management showcase-ready** features on v0.7.x branch: enhanced bundle health + AI pattern detection + coverage increase. Apply 80/20 ruthlessly.

**Duration:** 2 weeks  
**Target Release:** v0.7.2  
**Base Branch:** `release/v0.7.x` (branched from v0.7.1 tag)  
**Strategic Rationale:** 
- **RKE1 = EOL** — zero value, skip entirely
- **Management showcase** — needs visible AI, not infrastructure
- **Bundle health v2** — immediate user value, quick win
- **AI groundwork** — pattern engine + 3-5 demo patterns, not 100

---

## 🎯 Musk's 5 Laws Applied

| Law | Application |
|-----|-------------|
| **1. Question** | Do we need 10 AI patterns or 3 good ones? → **3** |
| **2. Delete** | Remove RKE1, remove complex pattern schemas → **YAML simplicity** |
| **3. Simplify** | Bundle health = missing files list + impact score. No complex scoring. |
| **2. Delete** | TUI dead code audit — [TUI_DELETION_AUDIT.md](TUI_DELETION_AUDIT.md) — **9% code reduction** |
| **4. Accelerate** | Demo-ready Day 8, polish Day 12-14 |
| **5. Automate** | CI validates patterns work; UX engineer validates polish |

---

## 📋 80/20 Scope (High Impact, Low Effort)

### P0: Bundle Health v2 — Partial Bundle Support (Days 1-4)
**Impact: HIGH | Effort: MEDIUM | Demo Value: 🔥🔥🔥**

What it does:
- Detects missing bundle files automatically
- Shows "Missing: podlogs/" with 🔴 impact score
- Gracefully degrades — partial bundles still work

Implementation:
```
internal/bundle/health.go            // Health checker
internal/bundle/health_test.go       // Unit tests
cmd/flags.go                         // --health flag?
```

Success metric: Load partial bundle → see clear health indicator → still works.

---

### P1: AI Pattern Engine — Foundation (Days 5-8)
**Impact: MEDIUM | Effort: MEDIUM | Demo Value: 🔥🔥🔥**

80/20 scope (management demo-ready):
```
internal/ai/
├── pattern.go        // Interface: Match(log string) -> MatchResult
├── registry.go       // Pattern registry
├── registry_test.go  
└── patterns/         // YAML pattern definitions
    ├── oomkill.yaml      // Simple: "Out of memory" + kill + oom_kill_process
    ├── imagepull.yaml    // "ImagePullBackOff" + errImagePull
    └── crashloop.yaml    // "Back-off restarting" + CrashLoopBackOff
```

Demo scenario: Load demo bundle → AI panel shows:
```
🔴 Critical: OOMKill detected in rancher/rancher pod
🟡 Warning: ImagePullBackOff for nginx:latest
🟢 Info: CrashLoopBackOff resolved after 3 restarts
```

**Out of scope (v0.8.1+):**
- ❌ 10 patterns → **3 patterns**
- ❌ Complex regex → **Keyword matching**
- ❌ Root cause hints → **Just detection**
- ❌ Anomaly detection → **Pattern matching only**
- ❌ Semantic analysis → **String contains**

---

### P2: CLI Commands — Headless Mode (Days 5-7)
**Impact: HIGH | Effort: MEDIUM | Demo Value: 🔥🔥🔥**

Headless commands for CI integration and AI workflows.

| Command | Purpose | Demo Moment |
|---------|---------|-------------|
| `r8s validate bundle [path]` | Health check without TUI | "Works in CI pipelines" |
| `r8s generate prompt [path]` | Export AI-ready prompts | "Pipe to Claude Code" |
| `r8s export findings [path]` | JSON/YAML for monitoring | "Integrate with your tools" |
| `r8s create demo-bundle --out` | Export synthetic demo | "Test without real bundles" |

**Implementation:**
```
cmd/validate.go      # Bundle health validation
cmd/generate.go      # Prompt generation (uses existing PromptGenerator)
cmd/export.go        # Findings export
cmd/create.go        # Demo bundle creation
```

**Demo Script (2 minutes):**
```bash
# 1. Validate bundle health (CI-friendly)
$ r8s validate ./production-bundle/
Bundle Health: 68% ⚠️
Missing: podlogs/ (medium impact)

# 2. Generate AI prompt for troubleshooting
$ r8s generate prompt ./production-bundle/ --format=terminal
# Copy-paste to Claude Code → get kubectl commands

# 3. Export for monitoring integration
$ r8s export findings ./production-bundle/ --format=json | jq '.critical[]'
```

---

### P3: TUI Integration — Demo Polish (Days 8-10)
**Impact: MEDIUM | Effort: MEDIUM | Demo Value: 🔥🔥🔥**

What makes it showcase-ready:
- **Tab: AI Analysis** — Shows pattern matches with icons
- **Tab: Bundle Health** — Missing files + completeness %
- **Severity colors:** 🔴 🟡 🟢 (visual, immediate understanding)
- **Hotkey: `a`** — Jump to AI panel

UX Engineer focus:
- Icon alignment
- Color accessibility
- Keyboard flow (Enter/Esc/Tab)
- Empty states ("No patterns matched" → friendly message)

---

### P4: Documentation + Showcase Prep (Days 11-13)
**Impact: MEDIUM | Effort: LOW | Demo Value: 🔥🔥**

Deliverables:
1. **Demo script** (3 min walkthrough)
2. **Pattern authoring guide** (YAML format)
3. **README update** (AI features section)
4. **CLI reference** (new commands documentation)
5. **Recording-ready build** (clean terminal, no debug output)

---

### P5: TUI Cleanup — Musk's Law #2 (Day 14, Buffer)
**Impact: LOW-MEDIUM | Effort: LOW | Code Health: 🔥🔥🔥**

Background: [TUI_DELETION_AUDIT.md](TUI_DELETION_AUDIT.md)

**What:**
- Delete dead view types (`ViewClusters`, `ViewProjects`, `ViewNamespaces`)
- Remove classic view toggle (`c` key) if product confirms
- Simplify sort modes (remove Count sort)

**Why:**
- 9% code reduction (~860 lines)
- Removes pre-v0.3.5 live API mode vestiges
- Faster builds, less maintenance

**Risk:** Low — Dead code, unreachable in bundle-only mode

**Deliverable:** PR with deletions + passing tests

---

## 🤝 Team Composition

| Role | Allocation | Focus |
|------|------------|-------|
| **RancherSRE** | 70% | Health checker, CLI commands, pattern engine |
| **CodeRabbit** | Continuous | Review all PRs |
| **🎨 UX Engineer** | Days 8-14 **NEW** | TUI polish, accessibility, demo flow |
| **Management** | Demo Day | See AI features in action |

**RancherSRE Scope:**
- Days 1-2: Bundle health core + `r8s validate`
- Days 3-4: AI pattern engine + `r8s generate prompt`
- Days 5-7: Pattern registry + remaining CLI commands
- Days 11-13: Documentation + demo script
- Day 14: TUI cleanup (buffer)

**UX Engineer Scope:** Days 8-14 (TUI polish + showcase readiness)
- Review TUI panels for consistency
- Ensure keyboard navigation (Enter/Esc/Tab)
- Accessibility: color contrast, clear labels
- Demo polish: clean output, no debug noise
- Recording-preview: terminal size, font, colors

---

## 📊 Success Criteria (Binary)

| Criterion | Target |
|-----------|--------|
| Bundle Health v2 | Partial bundles show health indicator + still work |
| AI Patterns | 3 patterns (OOM, ImagePull, CrashLoop) detect correctly |
| **CLI Commands** | **4 commands working: validate, generate, export, create** |
| TUI Polish | Keyboard nav works, colors accessible, clean demo |
| **Coverage** | **Increase to 45%+ (from 36.8%)** |
| Showcase Ready | 3-min demo script, clean build, no debug noise |
| TUI Cleanup (bonus) | Dead view types removed, ~9% code reduction |

### 🎯 Coverage Strategy (80/20)

**Focus on untested, high-impact code:**

| Package | Current | Target | Focus |
|---------|---------|--------|-------|
| `internal/bundle` | ~60% | **75%** | Health checker tests |
| `internal/ai` | 0% | **70%** | Pattern registry + matchers |
| `internal/tui` | ~14% | **20%** | AI panel integration tests |
| **Total Repo** | **36.8%** | **45%+** | New code + untested bundles |

**Skip (low ROI):**
- ❌ Complex TUI render tests (hard to test, low value)
- ❌ `cmd/` package (mostly boilerplate)
- ❌ `pkg/rancher/` (deprecated)

**NOT Required (v0.7.2):**
- ❌ 10 patterns (doing 3)
- ❌ Root cause hints
- ❌ Pattern confidence scores
- ❌ Anomaly detection
- ❌ Natural language queries

**Buffer/Filler (if time permits):**
- 🧹 TUI cleanup — [TUI_DELETION_AUDIT.md](TUI_DELETION_AUDIT.md) — 9% code reduction

---

## 🗓 Timeline (Demo-First)

| Day | Focus | Milestone |
|-----|-------|-----------|
| 1-2 | Bundle health core + `validate` command | `r8s validate` working, 80% coverage |
| 3-4 | AI engine + `generate prompt` | Pattern interface + prompt export ready |
| 5-7 | Pattern registry + CLI commands | 3 patterns + `export` + `create demo-bundle` |
| **7** | **🎉 DEMO MILESTONE** | "r8s as a power tool" demo ready |
| 8-10 | TUI polish + UX Engineer | Accessibility, keyboard flow, visual polish |
| 11-13 | Docs + showcase prep | Demo script, CLI reference, README updates |
| 14 | TUI cleanup + release | Delete dead code, tag v0.7.2 |

---

## ⚠️ Risks & 80/20 Cuts

| Risk | 80/20 Mitigation |
|------|------------------|
| Pattern accuracy low | **Only 3 patterns, high-confidence keywords** |
| UX polish takes too long | **UX Engineer starts Day 9, dedicated** |
| Demo fails | **Record backup Day 12, don't do live demo** |
| Health calculation complex | **Simple "files present / total files" % only** |

---

## 🚫 Out of Scope (v0.9+)

**Cut to hit demo date:**
- ❌ Pattern #4-10 (keep it at 3)
- ❌ Semantic analysis (use simple strings)
- ❌ Root cause explanations (just "detected")
- ❌ Confidence scores (binary match/no match)
- ❌ Anomaly detection (needs baselines)
- ❌ Natural language queries
- ❌ Pattern self-learning
- ❌ Complex health scoring (just % complete)

---

## ✅ Definition of Done

- [ ] Bundle health v2: Partial bundles show health indicator
- [ ] AI engine: Pattern interface + registry
- [ ] 3 patterns: OOMKill, ImagePullBackOff, CrashLoopBackOff
- [ ] **CLI: `r8s validate bundle` working**
- [ ] **CLI: `r8s generate prompt` with 3 formats**
- [ ] **CLI: `r8s export findings` JSON/YAML output**
- [ ] **CLI: `r8s create demo-bundle` export to disk**
- [ ] TUI: AI Analysis tab with 🔴🟡🟢 severity
- [ ] TUI: Bundle Health visible in status
- [ ] UX Engineer polish: Accessibility + keyboard flow
- [ ] Demo script: 3-minute walkthrough written (includes CLI demo)
- [ ] Showcase recording: Backup ready
- [ ] Coverage ≥ 45%
- [ ] CI: All green, no flaky jobs

---

## Why This Scope?

**Management Showcase = Needs to LOOK smart, not BE smart.**

80/20 truth:
- 3 patterns that work = demo success
- 10 patterns that fail = demo disaster
- Simple health check = user value
- Complex scoring = engineering time sink
- UX polish = credibility
- More features = confusion

**Musk's Law #2 (Delete):** If it doesn't demo well, delete it.

---

*Team: RancherSRE + CodeRabbit + 🎨 UX Engineer  
Demo target: Management showcase  
Success: "That's impressive — when can we show customers?"*
