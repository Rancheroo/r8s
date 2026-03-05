# Multi-Branch Merge Strategy: Sprint 9 → Sprint 10

**Goal:** Merge both branches through proper Cotton review  
**Branches:**
1. `feature/sprint9-cli-polish` (base: Sprint 9 completion)
2. `feature/sprint10-ci-cleanup` (depends on Sprint 9)

**Reviewers:** Cotton (primary), + optional Security/Documentation roles

---

## The Problem

Sprint 10 branch (`feature/sprint10-ci-cleanup`) was branched from Sprint 9, but Sprint 9 never merged to main. This creates a dependency chain.

**Branch Structure:**
```
main
  └── feature/sprint9-cli-polish (Day 6 complete, not merged)
        └── feature/sprint10-ci-cleanup (v0.8.0-alpha released)
```

**Risk:** If Sprint 9 has issues, Sprint 10 inherits them.

---

## Recommended Strategy: Sequential PRs

### Step 1: PR #1 — Sprint 9 (feature/sprint9-cli-polish → main)

**Scope:** Everything up to Day 6:
- Exit code fixes
- `r8s dashboard` command
- CI triggers for feature branches
- Test plans

**Cotton Review Focus:**
```
@cotton review focus:
- Exit code standardization in cmd/export.go, cmd/describe.go, cmd/logs.go
- Dashboard command structure (cmd/dashboard.go)
- CI workflow changes (.github/workflows/ci.yml)
- No breaking changes to existing commands
```

**Merge Condition:** Cotton approves + CI passes

---

### Step 2: Rebase Sprint 10

After Sprint 9 merges to main:

```bash
# Update local main
git checkout main
git pull origin main

# Rebase Sprint 10 onto main
git checkout feature/sprint10-ci-cleanup
git rebase main

# Force push (careful!)
git push origin feature/sprint10-ci-cleanup --force-with-lease
```

**Now Sprint 10 is clean:**
```
main (includes Sprint 9)
  └── feature/sprint10-ci-cleanup (clean, only Sprint 10 changes)
```

---

### Step 3: PR #2 — Sprint 10 (feature/sprint10-ci-cleanup → main)

**Scope:** Sprint 10 changes only:
- TUI Phase 1 deletion (4,898 lines)
- Simplified CI workflow
- README updates
- Release automation script

**Cotton Review Focus:**
```
@cotton review focus:
- TUI deletions (verify no orphaned references)
- CI simplification (removed test steps)
- README accuracy for CLI-first messaging
- Release script structure (scripts/release.sh)
- No accidental deletions of needed code
```

**Additional Roles:**
- @security-auditor: Check release script for security issues
- @documentation: Verify README changes

**Merge Condition:** Cotton + Security approve + CI passes

---

## Alternative: Single Combined PR

**If you want to skip sequential merges:**

```bash
# Create a combined branch
git checkout -b feature/v0.8.0-alpha-release feature/sprint10-ci-cleanup
git push origin feature/v0.8.0-alpha-release
```

**One PR with everything:**
- Cotton reviews entire delta from main
- Can use CodeRabbit for automated review
- Single merge, cleaner history

**Risk:** Larger PR = longer review time

---

## Cotton Interaction Template

**For PR #1 (Sprint 9):**
```
@cotton Please review this PR for Sprint 9 completion.

**Focus areas:**
1. Exit code standardization (cmd/export.go, describe.go, logs.go)
2. New dashboard command (cmd/dashboard.go, internal/tui/dashboard.go)
3. CI workflow changes (feature branch triggers)

**Context:**
- All verification tests pass (11/11)
- Builds successfully on all platforms
- Breaking change: None, all existing commands work

Please review for:
- Code quality
- Standards compliance
- Potential issues
```

**For PR #2 (Sprint 10):**
```
@cotton Please review Sprint 10 changes for v0.8.0-alpha.

**Focus areas:**
1. TUI deletions (4,898 lines removed) - verify clean deletion
2. CI simplification - removed test steps
3. README updates for CLI-first messaging

**Context:**
- Sprint 9 already merged (dependency satisfied)
- All verification tests pass
- v0.8.0-alpha already tagged and released

Please review for:
- Accidental deletions
- Documentation accuracy
- Release script security

@security-auditor Please review scripts/release.sh
@documentation Please verify README changes
```

---

## Decision Matrix

| Approach | Effort | Risk | Best For |
|----------|--------|------|----------|
| **Sequential PRs** | Medium | Low | Clear separation, easier review |
| **Combined PR** | Low | Medium | Faster merge, larger review |
| **Skip Sprint 9** | Low | High | Not recommended |

---

## Recommendation

**Sequential PRs** — Clean, manageable, proper dependency tracking.

**Timeline:**
1. **Today:** Create PR #1 (Sprint 9), assign Cotton
2. **Tomorrow:** Cotton review, address feedback, merge
3. **Day 3:** Rebase Sprint 10, create PR #2
4. **Day 4-5:** Cotton + Security review, merge
5. **Day 6:** Main branch has v0.8.0-alpha, clean history

---

## What Do You Want?

- **"Sequential"** — I'll prepare PR #1 for Sprint 9
- **"Combined"** — I'll create a single release branch
- **"Show me the commands"** — Step-by-step git commands
