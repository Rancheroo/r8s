# Branch Workflow Guidelines

**Professional Git Workflow for Sprints**

---

## Sprint Branch Policy

At the **start of each sprint**, create a sprint branch:

```bash
# At sprint start
git checkout main
git pull origin main
git checkout -b sprint[N]  # e.g., sprint4, sprint5
git push origin sprint[N]
```

The sprint branch serves as:
- The integration point for all sprint work
- A visible indicator of current sprint progress
- The branch for sprint demos/reviews

---

## Current Branches

| Branch | Purpose | Status |
|--------|---------|--------|
| `main` | Production-ready code | Default |
| `sprint4` | Sprint 4 integration | ✅ Active |

---

## Task Branch Naming

For individual tasks within a sprint:

```
sprint[N]-[priority]-[brief-description]

Examples:
- sprint4-critical-pv-pvc-support
- sprint4-high-journald-parser
- sprint4-medium-col-widths
```

---

## Workflow Steps

### 1. Start Sprint
```bash
git checkout main
git pull origin main
git checkout -b sprint4
git push origin sprint4
```

### 2. Work on Task
```bash
git checkout sprint4
git pull origin sprint4
git checkout -b sprint4-critical-pv-pvc
git push origin sprint4-critical-pv-pvc
# ... work ...
git add -A
git commit -m "feat: Add PV/PVC support"
git push origin sprint4-critical-pv-pvc
```

### 3. Create PR
Create PR from `sprint4-critical-pv-pvc` → `sprint4`

### 4. Merge to Sprint
After CodeRabbit approval:
```bash
# Merge via GitHub UI or:
gh pr merge [PR_NUMBER] --squash --delete-branch
```

### 5. End Sprint
```bash
git checkout main
git pull origin main
git merge sprint4  # or create PR: sprint4 → main
git push origin main
git push origin --delete sprint4  # Optional: cleanup
```

---

## Why This Matters

1. **Visibility**: Stakeholders can see sprint progress at a glance
2. **Isolation**: Sprint work doesn't pollute main until ready
3. **Rollback**: Easy to revert entire sprint if needed
4. **Demo**: Always have a stable branch for demos

---

## What Went Wrong (Sprint 4)

Sprint 4 was merged directly to main via feature branches (`sprint4-critical-fixes`, `sprint4-medium-polish`) without maintaining a persistent `sprint4` branch.

**Corrected:** Created `sprint4` branch post-merge for continuity.

---

*Document created: 2026-02-12*
