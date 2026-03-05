# r8s v1.0 Vision: kubectl for Bundles

**Date:** 2026-02-17  
**Status:** Strategic Vision Document  
**Target:** v1.0 Release

---

## 🎯 The Killer Feature

**r8s becomes `kubectl` for offline bundles.**

```bash
# Same commands, same muscle memory
# Just point at a bundle instead of a cluster

r8s get pods ./support-bundle.tar.gz
r8s logs ./bundle nginx-pod
r8s describe ./bundle deployment app
r8s top ./bundle nodes
```

---

## Why This Is Game-Changing

| Problem | Solution |
|---------|----------|
| Customer can't share cluster access | Send bundle, explore offline |
| Support needs historical data | Bundle = snapshot in time |
| Training on production issues | Safe bundle exploration |
| CI/CD testing | Test scripts against bundle snapshots |

---

## Roadmap to v1.0

### v0.8.0 — Pure CLI Foundation (4 weeks)
**Goal:** Strip TUI, establish CLI patterns

| Feature | Status | Priority |
|---------|--------|----------|
| Delete TUI (9,360 lines) | ⏳ Ready | P0 |
| `r8s analyze` | ✅ Exists | P0 |
| `r8s validate` | ✅ Exists | P0 |
| `r8s generate prompt` | ✅ Exists | P1 |
| `r8s logs` | ⏳ New | P1 |
| `r8s export` | ⏳ New | P1 |
| JSON output everywhere | ⏳ Refactor | P0 |
| Shell completion | ⏳ New | P2 |

**Exit Criteria:** 
- Zero TUI code
- All commands output JSON
- CI/CD friendly (exit codes, pipes)

---

### v0.9.0 — kubectl Compatibility Layer (4 weeks)
**Goal:** Make bundles quack like clusters

| Feature | Command | Priority |
|---------|---------|----------|
| Get resources | `r8s get pods ./bundle` | P0 |
| Get nodes | `r8s get nodes ./bundle` | P0 |
| Stream logs | `r8s logs ./bundle <pod>` | P0 |
| Describe resources | `r8s describe ./bundle <type> <name>` | P0 |
| Top/resources | `r8s top ./bundle nodes` | P1 |
| Events | `r8s get events ./bundle` | P1 |
| kubectl plugin | `kubectl bundle get pods ./bundle` | P2 |

**Exit Criteria:**
- 80% of `kubectl get/describe/logs` work on bundles
- Tab completion for resource names
- Familiar output formats (`-o yaml`, `-o json`)

---

### v1.0 — AI-Powered Insights (4 weeks)
**Goal:** Intelligence layer on top of kubectl compatibility

| Feature | Description | Priority |
|---------|-------------|----------|
| Pattern detection | Auto-detect OOM, CrashLoop, etc. | P0 |
| Context-aware hints | "This pod failed because..." | P0 |
| LLM integration | `r8s analyze \| claude` | P1 |
| Smart suggestions | "Try: r8s logs ./bundle <pod>" | P1 |
| Root cause hints | "Likely cause: memory pressure" | P1 |
| Report generation | `r8s doctor > support-case.md` | P2 |

**Exit Criteria:**
- r8s suggests fixes, not just finds issues
- LLM can consume r8s output directly
- Automated support case generation

---

## The Vision: Three Layers

```
┌─────────────────────────────────────────┐
│  Layer 3: AI & Intelligence (v1.0)     │
│  • Pattern detection                    │
│  • Root cause hints                     │
│  • LLM integration                      │
│  • Automated reports                    │
├─────────────────────────────────────────┤
│  Layer 2: kubectl Compatibility (v0.9) │
│  • r8s get pods/logs/describe           │
│  • kubectl plugin                       │
│  • Familiar UX                          │
├─────────────────────────────────────────┤
│  Layer 1: Pure CLI Foundation (v0.8)   │
│  • analyze, validate, export            │
│  • JSON output                          │
│  • Scriptable                           │
└─────────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────┐
│  Bundle (tar.gz with cluster data)     │
│  • Pod logs                            │
│  • Node info                           │
│  • Events                              │
│  • Metrics                             │
└─────────────────────────────────────────┘
```

---

## User Journey

### Support Engineer (Primary User)

```bash
# 1. Customer sends bundle
$ ls
support-bundle-2026-02-17.tar.gz

# 2. Explore like live cluster
$ r8s get pods ./support-bundle-2026-02-17.tar.gz
NAME                     READY   STATUS
nginx-7d8c9f4b2-x1z9q   1/1     Running
app-backend-9f3k2m1n-x   0/1     ImagePullBackOff

# 3. Investigate
$ r8s logs ./bundle app-backend-9f3k2m1n-x
Error: manifest not found

# 4. AI helps diagnose
$ r8s analyze ./bundle --format=json | llm "Why is this failing?"
The image tag v2.3.1 doesn't exist. Check registry or tag name.

# 5. Generate support case
$ r8s doctor ./bundle > case-12345.md
```

---

## LLM Integration Strategy

### Phase 1: Structured Output (v0.8.0)
All commands output JSON that LLMs can parse:
```json
{
  "criticalIssues": [
    {"type": "OOMKill", "pod": "nginx", "count": 5},
    {"type": "ImagePullBackOff", "pod": "app", "error": "manifest not found"}
  ],
  "suggestedActions": ["Increase memory limits", "Check image tag"]
}
```

### Phase 2: Prompt Generation (v0.8.0)
```bash
r8s analyze ./bundle --generate-prompt > claude-prompt.txt
# Optimized prompt with context, findings, and questions
```

### Phase 3: Direct Integration (v1.0)
```bash
# Built-in LLM support
r8s analyze ./bundle --llm=claude --interactive

# Or plugin architecture
r8s analyze ./bundle | llm-claude-plugin
```

---

## Success Metrics

| Metric | v0.8.0 | v0.9.0 | v1.0 |
|--------|--------|--------|------|
| CLI commands | 6 | 10+ | 15+ |
| TUI code | 0 | 0 | 0 |
| kubectl compatibility | 0% | 80% | 95% |
| Lines of code | ~15K | ~18K | ~22K |
| AI features | 0 | Basic | Full |

---

## Key Decisions

1. **Pure CLI, no TUI** — Scriptable, CI/CD friendly
2. **kubectl compatibility** — Familiar UX, no learning curve  
3. **JSON output first** — LLM friendly, pipeable
4. **Bundle as first-class** — Everything operates on bundles
5. **AI as layer on top** — Intelligence, not replacement

---

## For Elon

- **Question:** Does it need a TUI? No.
- **Delete:** 9,360 lines of Bubble Tea.
- **Simplify:** One tool, one job.
- **Accelerate:** CLI is faster than TUI.
- **Automate:** Pipe to anything.

---

**This is r8s v1.0: kubectl for bundles, powered by AI.**
