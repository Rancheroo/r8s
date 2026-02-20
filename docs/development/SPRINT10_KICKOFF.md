# Sprint 10 Kickoff: CI/CD Recovery + TUI Sunset

**Start Date:** 2026-02-20  
**Goal:** Fix CI/CD infrastructure + Complete TUI sunset (9,360 → 1,200 lines)  
**Target Release:** v0.8.0-alpha  
**Branch:** `feature/sprint10-ci-cleanup`

---

## Sprint Goals (Musk's Law Order)

### 1. QUESTION — CI/CD: Fix or Replace?

**Problem:** "Cannot open: File exists" errors blocking CI tests  
**Options:**

| Option | Effort | Risk | Verdict |
|--------|--------|------|---------|
| A. Debug GitHub Actions | 2-3 days | High (unknown root cause) | Risky |
| B. Simplify CI (build only) | 1 day | Low | ✅ Sprint 9 did this |
| C. Migrate to BuildKite/Drone | 1 week | Medium | Future consideration |
| D. Local-first CI (pre-commit hooks) | 2 days | Low | **Recommended** |

**Decision:** Option D — Local CI with pre-commit hooks + build-only in GitHub Actions

---

## Sprint 10 Tasks

### Week 1: CI/CD Fix (Days 1-3)

#### Day 1: Local CI Setup
- [ ] Install pre-commit hooks for tests
- [ ] Create `scripts/pre-commit.sh` — runs tests before allowing commit
- [ ] Update `CONTRIBUTING.md` with local test requirements
- [ ] Verify: `git commit` triggers test run

#### Day 2: GitHub Actions Cleanup
- [ ] Remove commented test steps from `ci.yml`
- [ ] Simplify workflow to: build + lint + verify
- [ ] Add badge to README: "Build Passing"
- [ ] Document: "Tests run locally, not in CI"

#### Day 3: Test Reliability
- [ ] Add `make test-quick` — fast test subset for pre-commit
- [ ] Add `make test-full` — complete test suite (slower)
- [ ] Ensure all tests pass reliably locally
- [ ] Document test commands in README

### Week 2: TUI Sunset (Days 4-7)

#### Day 4: Mass TUI Deletion — Phase 1
Delete files (backup branch first):
```bash
git branch backup/pre-tui-deletion
rm internal/tui/app.go
rm internal/tui/app_test.go
rm internal/tui/fetch.go
rm internal/tui/handlers.go
rm internal/tui/logs.go
rm internal/tui/table.go
rm internal/tui/table_helpers.go
```

#### Day 5: Mass TUI Deletion — Phase 2
```bash
rm internal/tui/prompt_generator.go
rm internal/tui/prompt_terminal.go
rm internal/tui/prompt_test.go
rm internal/tui/prompt_view.go
rm internal/tui/diagnostics.go
rm internal/tui/attention_signals.go
rm internal/tui/log_detection_test.go
rm internal/tui/log_scanning_test.go
```

#### Day 6: Cleanup & Dependencies
- [ ] Run `go mod tidy` — remove unused Bubble Tea deps
- [ ] Verify build passes
- [ ] Verify `r8s dashboard` still works
- [ ] Check binary size reduction (target: <12MB)

#### Day 7: Documentation Update
- [ ] Update README.md — CLI-first messaging
- [ ] Document breaking changes (TUI removed)
- [ ] Update quickstart guide
- [ ] Migration guide for existing users

### Week 3: Release Prep (Days 8-10)

#### Day 8: v0.8.0-alpha Release
- [ ] Tag release: `git tag v0.8.0-alpha`
- [ ] Push tag: `git push origin v0.8.0-alpha`
- [ ] Create GitHub release notes
- [ ] Attach binaries

#### Day 9: Announcement
- [ ] Post announcement (internal/team)
- [ ] Highlight: CLI-first architecture
- [ ] Highlight: 6 kubectl-style commands
- [ ] Note: TUI sunset (dashboard only)

#### Day 10: Sprint Retrospective
- [ ] What worked?
- [ ] What didn't?
- [ ] Sprint 11 planning

---

## Success Metrics

| Metric | Before | Target | After |
|--------|--------|--------|-------|
| CI Reliability | 0% (broken) | 100% (build) | TBD |
| Local Test Time | ~60s | <30s (quick) | TBD |
| TUI Lines | 9,360 | ~1,200 | TBD |
| Binary Size | ~15MB | <12MB | TBD |
| Dependencies | ~50 | <35 | TBD |
| Release Status | None | v0.8.0-alpha | TBD |

---

## Decisions Made

1. **CI/CD:** Local-first testing, build-only in GitHub Actions
2. **TUI:** Strip to dashboard only (80% deletion)
3. **Release:** v0.8.0-alpha after TUI sunset
4. **Documentation:** Update for CLI-first architecture

---

## Risks

| Risk | Mitigation |
|------|------------|
| TUI deletion breaks dashboard | Test after each deletion batch |
| Users confused by TUI removal | Clear migration guide + README |
| Dependencies not cleaning up | `go mod tidy` + verify build |
| Local CI hooks slow development | `make test-quick` for speed |

---

## Definition of Done

- [ ] CI passes (build + lint only)
- [ ] Local tests run via pre-commit hooks
- [ ] TUI stripped to ~1,200 lines
- [ ] `r8s dashboard` works
- [ ] v0.8.0-alpha released
- [ ] Documentation updated

---

**Ready to start?** Say "GO Day 1" or select a specific day.
