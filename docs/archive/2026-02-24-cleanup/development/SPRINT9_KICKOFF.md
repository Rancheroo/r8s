# Sprint 9 Kickoff — v0.8.0-alpha (CLI Polish + TUI Sunset)

**Date:** 2026-02-19
**Sprint Duration:** 2 weeks (Feb 19 - Mar 04, 2026)
**Target Release:** v0.8.0-alpha
**Branch:** `feature/sprint9-cli-polish`

---

## 🚀 The Vision: "kubectl for Bundles"

**Sprint 8 was a massive success:**
- ✅ **Pivot:** Shifted from TUI-heavy to CLI-first architecture.
- ✅ **Removal:** Deleted technical debt and bloated views (Musk's Law #2).
- ✅ **Automation:** Upgraded release process for reliability.
- ✅ **Foundation:** Shipped `validate` and `generate` commands.

**Sprint 9 Goal:**
Polish the CLI experience to `kubectl` standards and officially deprecate the heavy TUI views. Make the CLI the primary interface users love.

---

## 📅 Schedule & Scope

### Week 1: CLI Maturity (The "kubectl" Feel)

| Priority | Feature | Description |
|----------|---------|-------------|
| **P0** | `r8s completion` | Shell completion (bash/zsh/fish) — critical for adoption. |
| **P0** | `r8s logs` | Stream logs from bundle (like `kubectl logs`). |
| **P1** | `r8s describe` | Resource details (like `kubectl describe`). |
| **P1** | Output Formats | Standardize `--output=json|yaml|wide` across all commands. |

### Week 2: TUI Sunset & Documentation

| Priority | Feature | Description |
|----------|---------|-------------|
| **P0** | `r8s dashboard` | New command to launch the *lightweight* dashboard-only TUI. |
| **P1** | TUI Cleanup | Final deletion of legacy navigation views (Clusters/Projects). |
| **P1** | Man Pages | Generate man pages for all commands. |
| **P2** | Documentation | Update README and website with CLI-first tutorials. |

---

## 🛠️ Implementation Plan

### 1. Shell Completion (Day 1-2)
- Use Cobra's built-in generation.
- Ensure dynamic completion for pods/namespaces if possible (from bundle index).

### 2. Advanced Commands (Day 3-5)
- **`r8s logs [pod] -n [ns]`**:
  - Locate log file in bundle.
  - Pipe to stdout (support paging with `less` if TTY).
- **`r8s describe [kind] [name]`**:
  - Parse YAML from bundle.
  - Render simplified detail view (events, status, spec).

### 3. TUI Refactor (Day 6-8)
- Move `internal/tui` to `internal/tui/legacy` (optional) or just delete.
- Create `cmd/dashboard.go` that launches *only* the summary dashboard.
- Ensure `r8s` (no args) prints help, not TUI.

### 4. Release Automation (Day 9-10)
- Verify new release workflow handles v0.8.0-alpha builds.
- Ensure binaries are attached to GitHub releases automatically.

---

## ⚠️ Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Users miss TUI | Keep `r8s dashboard` as a first-class citizen. |
| Bundle format variance | Ensure `logs`/`describe` handle different collector versions gracefully. |
| Completion lag | Cache bundle index for fast tab-completion. |

---

## Success Metrics

- [ ] **CLI Parity:** Can I debug a crashloop without opening the TUI?
- [ ] **Speed:** `r8s logs` starts in <500ms.
- [ ] **Deletion:** TUI code reduction >50%.
- [ ] **User Joy:** Tab-completion works.

---

**Musk's Laws applied:**
1. **Delete:** Legacy TUI views gone.
2. **Simplify:** One binary, standard flags.
3. **Accelerate:** Fast shell completion.

*Let's build the tool support engineers deserve.* 🚀
