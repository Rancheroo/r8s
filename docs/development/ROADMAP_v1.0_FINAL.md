# r8s v1.0 Roadmap — Finalized
**kubectl for Rancher Support Bundles**

**Status:** LOCKED — Post-Pivot Alignment  
**Version:** v1.0.0  
**Target Date:** April 20, 2026  
**Last Updated:** 2026-02-18

---

## 🎯 Product Vision (Locked)

**r8s** is `kubectl` for Rancher support bundles.

Analyze clusters offline, detect issues with AI, integrate with CI/CD.
Simple. Fast. Scriptable.

**What r8s IS:**
- ✅ kubectl-compatible CLI for bundle navigation
- ✅ AI-powered pattern detection
- ✅ CI/CD integration (JSON output, exit codes)
- ✅ Bundle health validation
- ✅ Support for RKE2, K3s bundles

**What r8s is NOT (v1.0):**
- ❌ Live cluster connection
- ❌ Rancher UI extension
- ❌ Interactive TUI (deleted)
- ❌ Real-time monitoring

---

## 📋 Executive Summary

This document finalizes the v1.0 scope following the strategic pivot assessment.

**Key Decisions:**
1. **v1.0 = Bundle-only CLI** — No live cluster, no extension
2. **kubectl compatibility** — Familiar commands, immediate adoption
3. **AI-assisted analysis** — Pattern detection, not replacement for engineers
4. **April 20, 2026** — Locked release date

**What Changed:**
- Live cluster support → **SHELVED** for v1.1+
- Rancher extension → **SHELVED** for post-v1.0
- Scope reduced to core value: kubectl for bundles

**What Stayed:**
- CLI-first architecture
- Bundle analysis foundation
- AI pattern matching
- K3s support
- CI/CD integration

---

## 📅 Release Timeline (Locked)

| Release | Date | Focus | Key Deliverables |
|---------|------|-------|------------------|
| **v0.8.0** | Mar 2, 2026 | CLI MVP | kubectl commands: `get`, `logs`, `describe` |
| **v0.8.1** | Mar 9, 2026 | K3s Support | Format detection, 5-file path refactor |
| **v0.9.0** | Mar 23, 2026 | AI Intelligence | 10+ patterns, root cause hints |
| **v0.9.5** | Apr 6, 2026 | Performance | <1s analysis, parallel parsing, 70% coverage |
| **v1.0.0** | Apr 20, 2026 | Stable Release | 80% coverage, complete docs, production-ready |

**Post-v1.0 (Future Releases):**
- v1.1: Live cluster support (if user demand validates)
- v1.2: Rancher UI extension (separate initiative)
- v1.3: Advanced AI features (anomaly detection, learning)

---

## 🚀 Detailed Release Plans

### v0.8.0 — CLI MVP (March 2, 2026)
**Theme:** kubectl Compatibility

**Deliverables:**
- [ ] `r8s get pods <bundle>` — List all pods
- [ ] `r8s get pods <bundle> -n <namespace>` — Filter by namespace
- [ ] `r8s get nodes <bundle>` — List nodes
- [ ] `r8s get events <bundle>` — List cluster events
- [ ] `r8s describe pod <bundle> <pod-name>` — Pod details
- [ ] `r8s describe node <bundle> <node-name>` — Node details
- [ ] `r8s logs <bundle> <pod-name>` — Container logs
- [ ] `r8s logs <bundle> <pod-name> --previous` — Previous container
- [ ] Shell completion (bash, zsh, fish)
- [ ] Man pages

**Success Criteria:**
- All commands follow kubectl argument patterns
- JSON output for programmatic use (`--format=json`)
- Proper exit codes (0=ok, 1=issues found, 2=error)
- 50%+ test coverage
- Passes `make ci` (lint + test + build)

**User Value:**
> "I already know kubectl. r8s works the same way, but offline."

---

### v0.8.1 — K3s Support (March 9, 2026)
**Theme:** Multi-Distro Foundation

**Deliverables:**
- [ ] K3s format detection in `DetectFormat()`
- [ ] Path abstraction for 5 core files:
  - `types.go` — Add `FormatK3s`, path helpers
  - `manifest.go` — Use `b.KubectlPath()`
  - `validate.go` — Dynamic distro validation
  - `journald.go` — Service name mapping (k3s-agent vs rke2-server)
  - `completeness.go` — Dynamic path generation
- [ ] K3s bundle smoke tests
- [ ] Documentation update for K3s users

**Out of Scope:**
- ❌ RKE1 support (EOL distro, deferred indefinitely)
- ❌ Full 18-file refactor (5 files covers 80% of value)

**Success Criteria:**
- K3s bundles analyze correctly
- No regression on RKE2 bundles
- 55%+ test coverage

---

### v0.9.0 — AI Intelligence (March 23, 2026)
**Theme:** Smart Analysis

**Deliverables:**
- [ ] Pattern engine v2:
  - 10+ patterns (etcd, certificates, networking, storage, OOM, CrashLoop, ImagePull)
  - Confidence scoring (Certain/Likely/Possible)
  - Pattern correlation ("This + That = Root cause")
- [ ] Root cause hints:
  - "Pod is crashlooping because image tag 'latest' doesn't exist"
  - "Certificate expired 3 days ago"
  - "Node pressure detected — check disk/memory"
- [ ] Export formats:
  - Markdown reports (human-readable)
  - SARIF (security scanner integration)
  - JUnit XML (CI/CD integration)
- [ ] Natural language queries v1:
  - `r8s ask <bundle> "why is nginx-pod crashing?"`

**Out of Scope:**
- ❌ Self-learning patterns
- ❌ Complex NLQ with context awareness
- ❌ Predictive analysis

**Success Criteria:**
- Patterns detect real issues in test bundles
- Root cause hints are accurate >80% of time
- 60%+ test coverage
- Export formats work with popular tools (GitHub Actions, Jenkins, etc.)

---

### v0.9.5 — Performance (April 6, 2026)
**Theme:** Speed at Scale

**Deliverables:**
- [ ] Parallel bundle parsing (goroutines per file type)
- [ ] Lazy loading (parse on demand)
- [ ] Memory-mapped files for large logs
- [ ] Progress indicators for slow operations
- [ ] Benchmarking suite:
  - <1s analysis for 100MB bundles
  - <5s analysis for 1GB bundles
  - <500MB memory for any bundle size

**Success Criteria:**
- Performance targets met on test hardware
- No memory leaks (monitored over 100+ runs)
- 70%+ test coverage
- CI performance regression detection

---

### v1.0.0 — Stable Release (April 20, 2026)
**Theme:** Production-Ready

**Deliverables:**
- [ ] **Complete Documentation:**
  - User guide (CLI reference, examples)
  - Troubleshooting playbook
  - CI/CD integration guide
  - Pattern authoring guide
  - kubectl compatibility matrix
- [ ] **Quality Gates:**
  - 80%+ test coverage
  - Fuzz testing (random bundle inputs)
  - Cross-platform builds (Linux, macOS, Windows)
  - No disabled CI jobs
  - No known critical bugs
- [ ] **Enterprise Features:**
  - Config file support (`~/.r8s/config.yaml`)
  - Plugin architecture (custom parsers)
  - Stable JSON output schema (versioned)
- [ ] **Release Artifacts:**
  - Signed binaries for all platforms
  - Homebrew formula
  - Docker image
  - GitHub release notes

**Success Criteria:**
- All quality gates pass
- Documentation complete and reviewed
- v1.0 tag created and signed
- Release announced

---

## 🎨 kubectl Compatibility Reference

### Command Structure
```bash
r8s <verb> <resource> <bundle-path> [flags]
```

### Core Commands (v1.0)
```bash
# Resource listing
r8s get pods ./bundle/
r8s get pods ./bundle/ -n kube-system
r8s get pods ./bundle/ --all-namespaces
r8s get nodes ./bundle/
r8s get events ./bundle/

# Resource details
r8s describe pod ./bundle/ nginx-pod
r8s describe node ./bundle/ worker-1

# Logs
r8s logs ./bundle/ nginx-pod
r8s logs ./bundle/ nginx-pod --previous
r8s logs ./bundle/ nginx-pod -c container-name

# Analysis
r8s analyze ./bundle/
r8s analyze ./bundle/ --format=json
r8s analyze ./bundle/ --severity=critical

# Validation
r8s validate ./bundle/
r8s validate ./bundle/ --strict

# AI integration
r8s generate prompt ./bundle/
r8s generate export ./bundle/ --format=markdown
r8s ask ./bundle/ "why is nginx crashing?"
```

### Flags (kubectl-Compatible)
| Flag | Short | Description |
|------|-------|-------------|
| `--namespace` | `-n` | Filter by namespace |
| `--all-namespaces` | `-A` | All namespaces |
| `--output` | `-o` | Output format (json, yaml, table) |
| `--selector` | `-l` | Label selector |
| `--container` | `-c` | Container name |
| `--previous` | | Previous container logs |
| `--tail` | | Number of lines to show |
| `--since` | | Show logs newer than duration |

---

## 🔒 Scope Boundaries (What's NOT in v1.0)

### Live Cluster Support (SHELVED → v1.1+)
**Rationale:**
- High complexity (auth, RBAC, certificate management)
- Security surface area increases dramatically
- Different problem space from bundle analysis
- User demand unvalidated

**Decision:** Implement only if post-v1.0 user feedback demands it.

### Rancher UI Extension (SHELVED → post-v1.0)
**Rationale:**
- Separate product lifecycle from CLI
- Requires Rancher UI framework expertise (Vue.js, Steve API)
- Not required for core value proposition
- Can be developed independently

**Decision:** Separate initiative, not tied to v1.0.

### Advanced AI Features (SHELVED → v1.2+)
**Rationale:**
- Pattern detection is sufficient for v1.0
- Self-learning requires data collection infrastructure
- Complexity not justified by user value

**Decision:** Defer until pattern detection proves value.

---

## 📊 Success Metrics (v1.0)

| Metric | Target | Measurement |
|--------|--------|-------------|
| Test Coverage | 80%+ | `make coverage` |
| CI Status | 100% green | No disabled jobs |
| Performance | <1s/100MB | Benchmark suite |
| Binary Size | <10MB | `ls -lh r8s` |
| Distros Supported | 2 | RKE2, K3s |
| AI Patterns | 10+ | Pattern registry |
| kubectl Commands | 8+ | Command reference |
| Documentation | Complete | All sections reviewed |

---

## 🛡️ Musk's 5 Laws Applied

| Law | Application | Impact |
|-----|-------------|--------|
| **1. Question** | Why live cluster? Why extension? → Not essential for v1.0 | Scope reduced |
| **2. Delete** | TUI deleted (9,940 lines), live cluster shelved | -30% binary size |
| **3. Simplify** | kubectl patterns (familiar), bundle-only (focused) | Zero learning curve |
| **4. Accelerate** | April 20 target (was May 11 with live cluster) | 3 weeks faster |
| **5. Automate** | CI/CD integration, JSON output, exit codes | Zero-config pipelines |

---

## ⚠️ Risk Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| kubectl-compat confusion | Low | Medium | Clear error messages, comprehensive help |
| K3s format differs significantly | Medium | High | Research before coding; sample bundles |
| Pattern accuracy low | Low | Medium | Confidence scores, manual verification |
| Scope creep | Low | High | **Scope locked** — live cluster/extension are post-v1.0 |
| Performance targets unachievable | Low | Medium | Benchmark before optimizing |

**Key Risk Eliminated:** TUI/live cluster scope creep prevented by explicit shelving.

---

## 👥 Team Responsibilities

### RancherSRE (Technical Lead)
- kubectl command implementation
- AI pattern engine development
- Performance optimization
- Code review and quality gates

### Luna (Release Manager)
- Timeline management
- Release coordination
- Risk tracking
- Communication

### CodeRabbit (Automated Review)
- All PR reviews
- Quality gate enforcement
- Architecture compliance

### Technical Writer (v1.0 phase only)
- User guide
- CI/CD integration guide
- Pattern authoring guide

---

## 📝 Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-02-13 | Sequential timeline approved | Quality first; reduce integration risk |
| 2026-02-17 | CLI-first pivot | User testing showed CLI is 80% of value |
| 2026-02-17 | TUI deletion executed | Law #2 (Delete); -9,940 lines, -30% binary |
| 2026-02-18 | **Live cluster shelved** | High complexity, unvalidated demand → v1.1+ |
| 2026-02-18 | **Rancher extension shelved** | Separate product, high effort → post-v1.0 |
| 2026-02-18 | **Scope locked** | v1.0 = kubectl + AI + bundles only |

---

## 🚦 Immediate Next Steps

### This Week (Sprint 8 Completion)
1. [ ] Implement `r8s get pods`
2. [ ] Implement `r8s get nodes`
3. [ ] Implement `r8s describe`
4. [ ] Implement `r8s logs`
5. [ ] Add shell completion
6. [ ] Tag v0.8.0

### Next Week (v0.8.1 Planning)
1. [ ] Acquire K3s sample bundles
2. [ ] Document K3s path differences
3. [ ] Implement 5-file path abstraction

### v0.9.0 Preparation
1. [ ] Expand pattern registry to 10+ patterns
2. [ ] Implement confidence scoring
3. [ ] Add root cause hint generation

---

## 📚 Related Documents

- `CLI_FIRST_TEAM_REVIEW.md` — Pivot rationale
- `TUI_DELETION_AUDIT.md` — Cleanup details
- `ROADMAP_v1.0_CLI.md` — Previous version (pre-pivot-finalization)
- `V0.7x_STRATEGIC_BRIEF.md` — Original strategic planning

---

## ✅ Approval

This roadmap is **LOCKED** for v1.0 development.

**Scope Changes:** Require explicit approval from DontStop
**Timeline Changes:** Require explicit approval from DontStop
**Post-v1.0 Features:** Live cluster, Rancher extension shelved for future releases

---

**Prepared by:** RancherSRE, Luna, Product Manager  
**Approved by:** DontStop (SUSE Escalation Support Engineer)  
**Date:** 2026-02-18

---

*"kubectl for Rancher bundles. Simple. Fast. Scriptable."*
