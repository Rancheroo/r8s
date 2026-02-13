# Branch Protection Ruleset Configuration

## GitHub Repository: Rancheroo/r8s

This document describes the branch protection ruleset for the `main` branch.

---

## Ruleset: Protect Main Branch

### Overview
- **Name**: `Protect Main Branch`
- **Enforcement**: Active
- **Target**: `main` branch (default)

### Purpose
Prevent direct pushes to main and ensure all code passes CI validation before merging.

---

## Rules Configuration

### ✅ Require a pull request before merging
**Status**: Required

**Settings:**
- [x] Require approvals: 1
- [ ] Dismiss stale PR approvals when new commits are pushed
- [ ] Require review from Code Owners
- [ ] Restrict who can dismiss pull request reviews
- [ ] Require approval of the most recent reviewable push

### ✅ Require status checks to pass
**Status**: Required

**Required status checks:**
1. `Lint` - Code quality gates
2. `Build & Test` - Unit tests and coverage
3. `Cross-Platform Build (linux, amd64)`
4. `Cross-Platform Build (linux, arm64)`
5. `Cross-Platform Build (darwin, amd64)`
6. `Cross-Platform Build (darwin, arm64)`
7. `Documentation` - Doc generation
8. `Integration Test` - Full workflow validation

**Additional settings:**
- [x] Require branches to be up to date before merging

### ✅ Block force pushes
**Status**: Enabled

Prevents force pushes that could rewrite history.

### ✅ Require linear history
**Status**: Enabled

Enforces a clean, linear commit history without merge commits.

---

## Bypass List

**Who can bypass**: None

All contributors must follow the PR workflow.

---

## Workflow with Branch Protection

```
┌─────────────────┐     ┌──────────────┐     ┌─────────────────┐
│  Feature Branch │────▶│  Create PR   │────▶│  CI Validates   │
└─────────────────┘     └──────────────┘     └─────────────────┘
                                                        │
                        ┌──────────────┐               │
                        │  Merge to    │◀──────────────┘
                        │  Main        │    All checks pass
                        └──────────────┘
```

### Steps:
1. Create feature branch from main
2. Make changes locally
3. Push branch to origin
4. Create Pull Request
5. Wait for all status checks to pass
6. Get code review approval
7. Merge PR to main

---

## Emergency Procedures

If you need to bypass protection (emergency hotfix):
1. Go to Settings → Rulesets
2. Temporarily disable the ruleset
3. Push the hotfix
4. Re-enable the ruleset immediately after

**Note**: This should be extremely rare. Always prefer the PR workflow.

---

## Related Documentation
- `.github/workflows/ci.yml` - CI pipeline definition
- `docs/CI_CD_GUIDE.md` - CI/CD usage guide
- `BRANCH_WORKFLOW.md` - Branch management guide

---

*Last updated: 2026-02-13*
