# TUI Sunset Plan: Sprint 9 Week 2

**Goal:** Strip TUI from 9,360 lines to ~2,000 lines (dashboard only)  
**Target:** r8s dashboard command + clean deletion  
**Timeline:** Days 8-9 (2 days)

---

## Current TUI Structure

```
internal/tui/
├── app.go              # 1,173 lines - Main app logic
├── app_test.go         # 464 lines
├── attention.go        # 489 lines - Attention dashboard
├── attention_ai.go     # 213 lines - AI panel
├── attention_signals.go # 1,368 lines - Signal processing
├── attention_test.go   # 390 lines
├── diagnostics.go      # 116 lines
├── fetch.go            # 432 lines - Data fetching
├── handlers.go         # 737 lines - Event handlers
├── helpers.go          # 634 lines - Helper functions
├── log_detection_test.go # 106 lines
├── log_scanning_test.go # 381 lines
├── logs.go             # 1,191 lines - Log viewer
├── prompt_generator.go # 214 lines
├── prompt_terminal.go  # 165 lines
├── prompt_test.go      # 214 lines
├── prompt_view.go      # 161 lines
├── styles.go           # 128 lines
├── table.go            # 625 lines - Table rendering
├── table_helpers.go    # 161 lines
└── TOTAL:             # ~9,360 lines
```

---

## Deletion Strategy

### Phase 1: Identify What's Essential (Day 8 Morning)

**KEEP (Dashboard Only):**
- `attention.go` - Core dashboard (needs cleanup)
- `attention_ai.go` - AI panel (optional, keep if easy)
- `styles.go` - Color/style definitions
- Minimal helpers from `helpers.go`

**DELETE (Day 8-9):**
- `app.go` - Replace with minimal dashboard launcher
- `fetch.go` - Move logic to CLI commands
- `handlers.go` - Event handlers for navigation
- `logs.go` - Log viewer (CLI `logs` replaces this)
- `table.go` - Table rendering
- `prompt_*.go` - TUI prompts (CLI `generate` replaces)
- `diagnostics.go` - Move to CLI
- All `*_test.go` files for deleted components

### Phase 2: Create `r8s dashboard` (Day 8 Afternoon)

**New file:** `cmd/dashboard.go`
```go
// Minimal dashboard launcher
// Just launches the attention dashboard TUI
// No navigation, no complex state
```

**New minimal TUI:** `internal/tui/dashboard.go`
```go
// ~500 lines max
// Shows attention items only
// Keyboard: Enter (view details), q (quit)
// That's it.
```

### Phase 3: Mass Deletion (Day 9)

**Delete files:**
```bash
rm internal/tui/app.go
rm internal/tui/app_test.go
rm internal/tui/fetch.go
rm internal/tui/handlers.go
rm internal/tui/logs.go
rm internal/tui/table.go
rm internal/tui/table_helpers.go
rm internal/tui/prompt_*.go
rm internal/tui/diagnostics.go
rm internal/tui/log_*_test.go
rm internal/tui/attention_signals.go  # If too complex
```

**Remaining files:**
```
internal/tui/
├── dashboard.go        # NEW - Minimal dashboard (~500 lines)
├── attention.go        # Cleaned up (~300 lines)
├── attention_ai.go     # Optional (~100 lines)
├── styles.go           # Keep (~128 lines)
└── helpers.go          # Minimal (~200 lines)
```

**Total after deletion:** ~1,200 lines (from 9,360)

---

## Dependencies to Remove

### From `go.mod`:
```
github.com/charmbracelet/bubbles
github.com/charmbracelet/bubbletea
github.com/charmbracelet/lipgloss
github.com/charmbracelet/bubble-table
github.com/atotto/clipboard
```

### Impact:
- Binary size reduction: ~5-10MB
- Build time: Faster
- Dependencies: From ~50 to ~30

---

## Migration Path

### For Users:
```bash
# OLD (v0.7.x)
r8s ./bundle/                    # Launches full TUI

# NEW (v0.8.0)
r8s dashboard ./bundle/          # Launches minimal dashboard
r8s validate ./bundle/           # CLI (recommended)
r8s logs ./bundle/               # CLI
r8s describe ./bundle/           # CLI
```

### Breaking Changes:
- `r8s ./bundle/` no longer works (no default TUI)
- Must use explicit `r8s dashboard ./bundle/`
- Or use CLI commands (preferred)

---

## Implementation Checklist

### Day 8: Dashboard Command
- [ ] Create `cmd/dashboard.go`
- [ ] Create `internal/tui/dashboard.go` (minimal)
- [ ] Test `r8s dashboard ./bundle/` works
- [ ] Verify attention items display
- [ ] Verify keyboard shortcuts work

### Day 9: Mass Deletion
- [ ] Backup branch: `git branch backup/pre-tui-deletion`
- [ ] Delete obsolete TUI files
- [ ] Remove Bubble Tea imports
- [ ] Update `go.mod` (go mod tidy)
- [ ] Build passes
- [ ] All CLI commands still work
- [ ] Test dashboard only
- [ ] Update README
- [ ] Push to origin

### Day 10: Documentation
- [ ] Document breaking changes
- [ ] Update quickstart guide
- [ ] Migration guide for users

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Delete too much | Backup branch before deletion |
| Dashboard breaks | Test after each file deletion |
| Users confused | Clear migration guide |
| Dependencies broken | `go mod tidy` after deletion |

---

## Success Metrics

| Metric | Before | After | Target |
|--------|--------|-------|--------|
| TUI Lines | 9,360 | ~1,200 | <2,000 ✅ |
| Dependencies | ~50 | ~30 | <35 ✅ |
| Binary Size | ~15MB | ~10MB | <12MB ✅ |
| Build Time | 30s | 20s | <25s ✅ |

---

## Post-Deletion Opportunities

Once TUI is stripped:

1. **Rename `internal/tui/` to `internal/dashboard/`**
   - Clearer intent
   - Only dashboard code remains

2. **Simplify dashboard further**
   - Static HTML output option?
   - Even simpler TUI?

3. **Focus on CLI**
   - All new features go to CLI
   - Dashboard is display-only

---

## Ready to Execute

**Prerequisites:**
- [ ] Week 1 CLI commands complete and tested
- [ ] `r8s dashboard` command designed
- [ ] User notification of breaking changes

**Execution Order:**
1. Create dashboard command
2. Test dashboard works
3. Backup branch
4. Delete files
5. Verify build
6. Update docs
7. Celebrate 🎉

---

*Plan ready for Days 8-9 execution*  
*Estimated time: 2 days*  
*Risk: Low (backup strategy in place)*
