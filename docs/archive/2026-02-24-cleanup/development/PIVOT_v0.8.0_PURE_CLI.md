# PIVOT PLAN: Pure CLI v0.8.0

**Decision Date:** 2026-02-17  
**Decision:** Kill TUI, go pure CLI  
**Target:** v0.8.0 in 4 weeks

---

## 🎯 THE VISION

r8s becomes the `kubectl` for Rancher bundles.

```bash
# One-liner that changes everything
r8s analyze ./bundle | jq '.criticalIssues'
```

---

## 📋 PHASES

### Phase 1: Cleanup & Archive (Today)

**Tasks:**
- [ ] Archive old sprint docs (pre-Sprint 8)
- [ ] Update README with CLI-first philosophy
- [ ] Mark TUI code for deletion
- [ ] Create v0.8.0 ROADMAP

**Files to Archive:**
- docs/development/SPRINT[5-7]*
- docs/development/TUI_* (if any)
- All pre-CLI docs

### Phase 2: TUI Deletion (Days 1-3)

**Branch:** `feature/v0.8.0-pure-cli`

**Tasks:**
- [ ] Delete `internal/tui/` (9,360 lines)
- [ ] Remove Bubble Tea dependencies from go.mod
- [ ] Update `cmd/root.go` - no more default TUI
- [ ] Update `cmd/tui.go` - remove or convert to `r8s dashboard` (optional)
- [ ] Fix imports across codebase
- [ ] Ensure tests pass

**Commands:**
```bash
git checkout -b feature/v0.8.0-pure-cli
git rm -rf internal/tui/
go mod tidy
go build .  # Verify clean build
```

### Phase 3: CLI Core (Days 4-10)

**Commands to Build:**

| Priority | Command | Status |
|----------|---------|--------|
| P0 | `r8s analyze [bundle]` | ✅ Exists (needs refactor) |
| P0 | `r8s validate [bundle]` | ✅ Exists |
| P0 | `r8s generate prompt [bundle]` | ✅ Exists |
| P1 | `r8s logs [bundle] [pod]` | ⏳ New |
| P1 | `r8s describe [bundle] [resource]` | ⏳ New |
| P1 | `r8s export [bundle]` | ⏳ New |
| P2 | `r8s doctor [bundle]` | ⏳ New |
| P2 | `r8s completion` | ⏳ Shell completion |

**All commands must support:**
- `--format=json|yaml|table`
- `--output=file`
- Meaningful exit codes
- Pipe-friendly (no prompts unless stdin is TTY)

### Phase 4: Polish & Docs (Days 11-14)

- [ ] Man pages for all commands
- [ ] Shell completion (bash, zsh, fish)
- [ ] README rewrite
- [ ] Quickstart guide
- [ ] Migration guide (TUI → CLI)
- [ ] CI/CD examples

### Phase 5: Rancher Extension PoC (Sprint 9)

- [ ] Extension skeleton
- [ ] Shell out to r8s binary
- [ ] Parse JSON output
- [ ] Render in Vue/React

---

## 🗓️ TIMELINE

```
WEEK OF FEB 17: Phase 1-2 (Cleanup + TUI Deletion)
WEEK OF FEB 24: Phase 3 (CLI Core)
WEEK OF MAR 3:  Phase 4 (Polish)
WEEK OF MAR 10: Phase 5 (Extension PoC)

→ v0.8.0 release: March 17, 2026
```

---

## 🚫 DEPRECATED

The following are **gone** in v0.8.0:

| Feature | Replacement |
|---------|-------------|
| TUI Dashboard | `r8s analyze --format=table` |
| Visual log browser | `r8s logs [pod] \| less -R` |
| Mouse navigation | Shell completion + aliases |
| Bubble Tea | Nothing (deleted) |

---

## ✅ NEW IN v0.8.0

| Feature | Value |
|---------|-------|
| `r8s analyze --format=json \| jq` | Composable |
| `r8s validate \|\| exit 1` | CI/CD ready |
| `r8s doctor > support.md` | Automation |
| Shell completion | Fast UX |
| 1/10th the code | Maintainable |

---

## 🔥 ELON'S LAWS CHECKLIST

- [ ] **Question:** Does TUI add value? → No, delete it.
- [ ] **Delete:** 9,360 lines of TUI code
- [ ] **Simplify:** One interface, many commands
- [ ] **Accelerate:** No Bubble Tea upgrades, no TTY hell
- [ ] **Automate:** CI/CD native from day one

---

## 🎬 NEXT ACTIONS

1. **You:** Review this plan, say "go"
2. **Me:** Create branch, start TUI deletion
3. **Daily:** Push → Verify → "Pull and test" → User Verified ✓

---

*For Elon. For the automation. For the win.*
