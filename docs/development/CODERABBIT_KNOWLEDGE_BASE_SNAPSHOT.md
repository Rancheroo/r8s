# CodeRabbit Knowledge Base Snapshot — 2026-02-20

**Source:** https://github.com/Rancheroo/r8s/pull/64#issuecomment-3931893467  
**Status:** ✅ AI Reviewer Successfully Calibrated  
**Method:** Interactive teaching via PR comments

---

## CodeRabbit's Internalized Learnings

### Framework-Level (Decision Making)
```
80/20 decision framework:
- Always Fix: security, crashes, broken promises
- Always Defer: typos, cosmetic issues, whitespace
- 80/20 Rule: fix if <30 min, else document and ship
- Golden Rule: Shipping working software > Perfect code
```

### Domain-Specific (r8s Project)
```
Exit codes = "promises to users":
- Must be fixed before merge
- Documented behavior is non-negotiable
- Example: completion.go exit code fix was blocking
```

```
Configuration philosophy:
- Start with validated defaults
- Add CLI flags only after user feedback
- "Best Feature = No Feature"
```

```
Error handling philosophy:
- Fail fast in development (panic on programming errors)
- Graceful degradation in production (log and continue)
- Never return mock data on error
- Empty result is valid (not an error)
```

```
Exit code contracts:
- test-cluster: 0=success, 1=error, 2=tests failed
- This is a CI contract for pipeline integration
- os.Exit(2) is intentional when no cleanup needed
```

---

## How CodeRabbit Will Apply This

**Going forward, CodeRabbit will:**

1. **Filter all findings through our lens:**
   ```
   Will this break user trust? → 🔴 Block merge
   Is this architectural purity? → 🟡 Suggest, don't block
   Could this be simpler? → 🔵 Defer to v0.8.1
   ```

2. **Reference our framework:**
   - Check if issue is <30 min fix
   - Suggest documentation if >30 min
   - Distinguish healthy vs toxic debt

3. **Respect our philosophy:**
   - "Best Feature = No Feature" → bias toward deletion
   - User promises are sacred → block on broken contracts
   - Context over rules → apply judgment, not rigid metrics

---

## Teaching Method That Worked

### What We Did
1. Explained our philosophy (3 comments)
2. Gave specific examples (completion.go, CI tests)
3. Acknowledged good catches
4. Showed reasoning (80/20, shipping > perfection)

### What CodeRabbit Did
1. Internalized framework
2. Applied to specific cases
3. Offered improvement (debt review cadence)
4. Confirmed understanding (handshake comment)

---

## Success Metrics

| Indicator | Status |
|-----------|--------|
| Framework understood | ✅ Yes |
| Domain rules learned | ✅ Yes |
| Applied to examples | ✅ Yes |
| Offered improvements | ✅ Yes |
| Persistent memory | ✅ Yes (cross-PR) |

---

## For Future Projects

**This method scales:**
1. Create COTTON_FEEDBACK_FRAMEWORK.md
2. Have 3-5 teaching interactions
3. Confirm understanding
4. Reference in future PRs

**Result:** Calibrated AI reviewer aligned with team values.

---

## Reference

- PR #64: Full teaching conversation
- COTTON_FEEDBACK_FRAMEWORK.md: Our philosophy
- CODERABBIT_LEARNING_PR64.md: Detailed log
- MEMORY.md: Quick reference

---

**We successfully trained an AI code reviewer.** 🎓
