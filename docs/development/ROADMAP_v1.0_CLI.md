# r8s v1.0 Roadmap — CLI-First Edition
**The Path to Production-Ready (Revised Post-TUI Deletion)**

**Status:** ACTIVE — Sprint 8 Complete, TUI Deleted  
**Last Updated:** 2026-02-18  
**Approach:** Sequential Releases, CLI-Only Architecture

---

## 🎯 v1.0 Vision (Revised)

**r8s** is the definitive **kubectl-for-Rancher-bundles**:
- ✅ Parse any Rancher/RKE2/K3s bundle (90%+ coverage)
- ✅ Detect issues automatically (AI-assisted analysis)
- ✅ **Scriptable CLI** — pipe to jq, integrate with CI/CD
- ✅ **Offline-first** — works without cluster access
- ✅ **Fast** — sub-second analysis, minimal memory
- ✅ **kubectl-compatible** — familiar commands, flags, output

**What v1.0 is NOT:**
- ❌ TUI/dashboard (deleted per Sprint 8)
- ❌ Interactive wizard (replaced by `r8s analyze` with smart defaults)
- ❌ Live cluster mode (bundle-only by design)

---

## 📅 Revised Release Timeline

### Current State: v0.8.0-pure-cli IN PROGRESS ✅
- TUI deleted (9,940 lines removed)
- Pure CLI architecture active
- Commands: `analyze`, `validate`, `generate`, `test-cluster`
- Binary: 7.6MB (30% smaller)

### Road to v1.0 (CLI-First)

| Release | Target Date | Focus | Key Deliverables |
|---------|-------------|-------|------------------|
| **v0.8.0** | Mar 2 | CLI MVP | kubectl-compat commands, AI patterns, bundle health |
| **v0.8.1** | Mar 9 | K3s Support | Format detection, 5-file path abstraction |
| **v0.9.0** | Mar 23 | AI Intelligence | Pattern engine v2, root cause hints, export formats |
| **v0.9.5** | Apr 6 | Performance | Parallel parsing, <1s analysis, memory optimization |
| **v1.0.0** | Apr 20 | Stable Release | Production-ready, 80% coverage, complete docs |

**Timeline Compression:** v1.0 moved from May 11 → Apr 20 (3 weeks faster) due to TUI deletion reducing complexity.

---

## Detailed Release Plans

### v0.8.0 — CLI MVP (March 2)
**Status:** Sprint 8 Complete — Polish & Release

**Already Delivered:**
- ✅ `r8s analyze [path]` — Default command, issue detection
- ✅ `r8s validate [path]` — Bundle health check
- ✅ `r8s generate prompt [path]` — AI prompt export
- ✅ `r8s generate export [path]` — JSON/YAML findings
- ✅ `r8s test-cluster [path]` — Automated diagnostics
- ✅ AI pattern engine (3 patterns: OOM, ImagePull, CrashLoop)

**Remaining for v0.8.0:**
- [ ] `r8s get pods [path]` — kubectl-compatible listing
- [ ] `r8s logs [path] [pod]` — Log streaming
- [ ] `r8s describe [path] [resource]` — Resource details
- [ ] Shell completion (bash, zsh, fish)
- [ ] Man pages
- [ ] README rewrite for CLI-first

**Success Criteria:**
- All commands follow kubectl patterns (`r8s <verb> <resource> <path>`)
- JSON output for CI integration
- Proper exit codes (0=ok, 1=issues, 2=error)
- 45%+ test coverage

---

### v0.8.1 — K3s Support (March 9)
**Scope:** Reduced from original 18-file refactor

**In:**
- K3s format detection (`DetectFormat()` enhancement)
- 5 core files path abstraction:
  - `types.go` — Add `FormatK3s`, path helpers
  - `manifest.go` — Use `b.KubectlPath()`
  - `validate.go` — Dynamic distro validation
  - `journald.go` — Service name mapping (k3s-agent vs rke2-server)
  - `completeness.go` — Dynamic path generation
- Smoke tests for K3s bundles

**Out:**
- ❌ RKE1 support (deferred to post-v1.0 — RKE1 is EOL, low demand)
- ❌ Full 18-file refactor (5 files covers 80% of paths)

**Contractor Need:** None (backend only)

---

### v0.9.0 — AI Intelligence (March 23)
**Scope:** Smart analysis, not just detection

**In:**
- Pattern engine v2:
  - 10+ patterns (etcd, certificates, networking, storage)
  - Confidence scoring (Certain/Likely/Possible)
  - Pattern correlation ("This + That = Root cause")
- Root cause hints:
  - "Pod is crashlooping because image tag 'latest' doesn't exist"
  - "Certificate expired 3 days ago — run `rke2 certificate rotate`"
- Export formats:
  - Markdown reports (human-readable)
  - SARIF (security scanner integration)
  - JUnit XML (CI/CD integration)
- Natural language queries (simple):
  - `r8s ask "why is nginx-pod crashing?"`

**Out:**
- ❌ Complex NLQ (deferred)
- ❌ Self-learning patterns (deferred)

**Contractor Need:** None

---

### v0.9.5 — Performance (April 6)
**Scope:** Speed for large bundles

**In:**
- Parallel bundle parsing (goroutines per file type)
- Lazy loading (parse on demand, not upfront)
- Memory-mapped files for large logs
- Benchmarking:
  - <1s analysis for 100MB bundles
  - <5s analysis for 1GB bundles
  - <500MB memory for any bundle size
- Progress indicators for slow operations

**Out:**
- ❌ Caching (unnecessary with lazy loading)
- ❌ Incremental updates (bundle files don't change)

**Contractor Need:** None

---

### v1.0.0 — Stable Release (April 20)
**Scope:** Production-ready, enterprise-grade

**In:**
- **Complete documentation:**
  - User guide (CLI reference, examples)
  - Troubleshooting playbook
  - CI/CD integration guide
  - Pattern authoring guide
- **Quality gates:**
  - 80%+ test coverage
  - Fuzz testing (random bundle inputs)
  - Cross-platform builds (Linux, macOS, Windows)
  - No disabled CI jobs
- **Enterprise features:**
  - Config file support (`~/.r8s/config.yaml`)
  - Plugin architecture (custom parsers)
  - Team collaboration (shared pattern libraries)
- **Stable API:**
  - JSON output schema versioned
  - Exit codes documented
  - Backward compatibility guarantee

**Contractor Need:**
- 📝 **Technical Writer** (2 weeks) — Documentation polish

---

## Revised Contractor Requirements

### Original Plan (With TUI):
| Role | When | Duration |
|------|------|----------|
| TUI/UX Expert | v0.7.3 | 2-3 weeks |
| Product Manager | v0.9.0 | 4-6 weeks |
| Technical Writer | v1.0 | 2-3 weeks |

### Revised Plan (CLI-Only):
| Role | When | Duration | Savings |
|------|------|----------|---------|
| ~~TUI/UX Expert~~ | ~~v0.7.3~~ | ~~2-3 weeks~~ | **Eliminated** |
| ~~Product Manager~~ | ~~v0.9.0~~ | ~~4-6 weeks~~ | **Eliminated** |
| Technical Writer | v1.0 | 2 weeks | Reduced |

**Total Savings:** 6-9 weeks of contractor budget

**Rationale:**
- TUI deletion removes need for Bubble Tea expertise
- CLI is self-documenting (kubectl patterns are familiar)
- Simpler product = less PM coordination needed

---

## CLI Command Reference (v1.0 Target)

### Core Commands (kubectl-compatible)
```bash
# Analysis
r8s analyze ./bundle/                    # Default — detect all issues
r8s analyze ./bundle/ --format=json      # CI-friendly output
r8s analyze ./bundle/ --severity=critical # Filter by severity

# Validation
r8s validate ./bundle/                   # Bundle health check
r8s validate ./bundle/ --strict          # Fail on warnings

# kubectl-compatible resource inspection
r8s get pods ./bundle/                   # List pods
r8s get pods ./bundle/ -n kube-system    # Filter by namespace
r8s get nodes ./bundle/                  # List nodes
r8s get events ./bundle/                 # List events

r8s describe pod ./bundle/ nginx-pod     # Pod details
r8s describe node ./bundle/ worker-1     # Node details

r8s logs ./bundle/ nginx-pod             # Stream logs
r8s logs ./bundle/ nginx-pod --previous  # Previous container logs
r8s logs ./bundle/ nginx-pod -f          # Follow (simulated)
```

### AI & Export Commands
```bash
# AI integration
r8s generate prompt ./bundle/            # Generate AI troubleshooting prompt
r8s generate prompt ./bundle/ --format=claude   # Claude Code format
r8s generate prompt ./bundle/ --format=terminal # Terminal-friendly

r8s ask ./bundle/ "why is nginx crashing?" # Natural language query

# Export findings
r8s export ./bundle/ --format=json       # Machine-readable
r8s export ./bundle/ --format=markdown   # Human-readable report
r8s export ./bundle/ --format=sarif      # Security scanner
r8s export ./bundle/ --format=junit      # CI/CD integration
```

### Utility Commands
```bash
r8s test-cluster ./bundle/               # Automated diagnostics
r8s create demo-bundle --output ./demo/  # Generate test bundle
r8s completion bash > /etc/bash_completion.d/r8s  # Shell completion
r8s version                              # Version info
r8s config init                          # Create config file
```

---

## Success Metrics by Release (Revised)

| Release | Coverage | CI Status | Performance | Distros | AI Features | Binary Size |
|---------|----------|-----------|-------------|---------|-------------|-------------|
| v0.8.0 | 45% | ✅ 100% | Baseline | 1 (RKE2) | 1 (3 patterns) | 7.6MB |
| v0.8.1 | 50% | ✅ 100% | Baseline | 2 (+K3s) | 1 | 7.8MB |
| v0.9.0 | 60% | ✅ 100% | Baseline | 2 | 3 (+root cause) | 8.0MB |
| v0.9.5 | 70% | ✅ 100% | <1s/100MB | 2 | 3 | 8.0MB |
| **v1.0.0** | **80%+** | **✅ 100%** | **<1s/100MB** | **2** | **4** | **<10MB** |

---

## Musk's 5 Laws Applied to v1.0

| Law | Application | Impact |
|-----|-------------|--------|
| **1. Question** | Deleted TUI (was it necessary?) → Users want CLI | -9,940 lines |
| **2. Delete** | Removed Bubble Tea, dead views, mock data | -30% binary size |
| **3. Simplify** | kubectl-compatible commands (familiar patterns) | Zero learning curve |
| **4. Accelerate** | Parallel parsing, lazy loading | <1s analysis |
| **5. Automate** | CI/CD integration, JSON output, exit codes | Zero-config pipelines |

---

## Risk Mitigation (Revised)

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| K3s format differs significantly | Medium | High | Research before coding; collect sample bundles |
| AI features underwhelm | Low | Medium | Pattern grouping already works; root cause is bonus |
| kubectl-compat confusing | Low | Medium | Clear error messages; `r8s help` examples |
| Scope creep | Medium | High | **TUI deletion prevents this** — no UI to add features to |

**Key Risk Eliminated:** TUI scope creep is impossible — the TUI is gone.

---

## Decision Log (Updated)

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-02-13 | Sequential timeline approved | Quality first; reduce integration risk |
| 2026-02-17 | CLI-first pivot | User testing showed CLI is 80% of value |
| 2026-02-17 | **TUI deletion executed** | Law #2 (Delete); -9,940 lines, -30% binary |
| 2026-02-18 | RKE1 deferred post-v1.0 | EOL distro, low user demand |
| 2026-02-18 | v1.0 moved to Apr 20 | TUI deletion reduced complexity by 3 weeks |

---

## Immediate Next Steps

### v0.8.0 Release (This Week)
1. [ ] Implement `r8s get pods` (kubectl-compatible)
2. [ ] Implement `r8s logs` (log streaming)
3. [ ] Implement `r8s describe` (resource details)
4. [ ] Add shell completion
5. [ ] Rewrite README for CLI-first
6. [ ] Tag v0.8.0

### v0.8.1 Planning (Next Week)
1. [ ] Acquire K3s sample bundles
2. [ ] Document K3s path differences
3. [ ] Implement 5-file path abstraction

---

## Communication for Team

### What Changed
1. **TUI deleted** — Sprint 8 pivot executed (9,940 lines removed)
2. **CLI is the product** — kubectl-compatible commands
3. **Timeline accelerated** — v1.0 moved from May 11 → Apr 20
4. **Contractors eliminated** — No TUI/UX expert, no PM needed
5. **Simpler architecture** — Pure Go, no Bubble Tea dependencies

### What Stayed the Same
1. **Quality first** — CI passing required for all releases
2. **Sequential releases** — No parallel tracks
3. **K3s support** — Still planned (v0.8.1)
4. **AI features** — Still planned (v0.9.0)
5. **Production goal** — v1.0 is still the target

### Why This is Better
- **Faster to v1.0** — 3 weeks sooner
- **Lower cost** — No contractors for TUI/PM
- **Simpler codebase** — 30% smaller binary, fewer dependencies
- **User-aligned** — CLI is what support engineers want
- **Maintainable** — No TUI state management complexity

---

**Prepared for:** Team Alignment  
**Decision needed:** Approve revised timeline, greenlight v0.8.0 release

---

*This roadmap reflects the post-TUI-deletion reality.*  
*The CLI is the product. Everything else is details.*
