# Post-Merge Cleanup Instructions - v0.5.2

## After PR is merged to main

Once the `feature/v0.5.0-refactor-app` branch is merged into `main`, follow these steps to clean up:

### Step 1: Switch to main and pull

```bash
git checkout main
git pull origin main
```

### Step 2: Delete local branch

```bash
git branch -d feature/v0.5.0-refactor-app
```

### Step 3: Delete remote branch

```bash
git push origin --delete feature/v0.5.0-refactor-app
```

### Step 4: Verify cleanup

```bash
git branch -vv
# Should only show main (and any other active work branches)

git branch -r
# Should not show origin/feature/v0.5.0-refactor-app
```

### Step 5: Clean up old merged branches (per 30-Day Rule)

```bash
# Delete the 8 stale merged branches identified in audit
git branch -d audit docs-and-release docs-release-0.4.3 feat/namespace-health-ranking feat/smart-sorting-both-views fix-core-bugs remove-live-mode

# Delete the merged branch with remote
git branch -d feature/v0.4.4-post-audit-improvements
git push origin --delete feature/v0.4.4-post-audit-improvements
```

### Final State
After cleanup, you should have:
- **Local:** Only `main` branch (plus any new active work)
- **Remote:** Only `origin/main`
- **Result:** Clean 80% branch reduction, fresh repository feel

---

**Note:** This file can be deleted after completing the cleanup steps.
