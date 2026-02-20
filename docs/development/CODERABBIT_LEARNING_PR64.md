# CodeRabbit Learning Log: Successful Teaching Moment

**Date:** 2026-02-20  
**PR:** #64  
**Comment:** https://github.com/Rancheroo/r8s/pull/64#issuecomment-3931889935  
**Status:** ✅ LEARNING CONFIRMED

---

## What CodeRabbit Learned

### 1. Our Philosophy (Internalized ✅)
- MVP Patterns: Start simple, iterate based on usage
- Refactor When Needed: Not for metrics, for real needs
- Defer Configuration: Wait for user requests

### 2. It Distinguished Debt Types
**Healthy Debt (Ship It):**
- Known, documented, bounded
- Low blast radius
- Clear payoff threshold

**Toxic Debt (Block It):**
- Silent failures (broken exit codes)
- Compounding complexity
- Security/data loss

### 3. It Applied Our 80/20 Rule
```
Fix if: impact × probability > 30 min investment
Defer if: documented + mitigated + monitored
Never ship: broken promises, security, crashes
```

### 4. It Offered an Improvement! 🎓
**Suggestion:** Debt review cadence every 2-3 sprints
- Close: "No user reports → won't fix"
- Promote: "Reported 3x → must fix"
- Archive: "Superseded by new design"

**Why This is Brilliant:** Prevents #65 from becoming a junk drawer

---

## Teaching Success Metrics

| Indicator | Status |
|-----------|--------|
| Acknowledged framework | ✅ Yes |
| Applied to specific examples | ✅ Yes |
| Distinguishes debt types | ✅ Yes |
| Offers constructive additions | ✅ Yes |
| Suggests process improvement | ✅ Yes |

**Result:** CodeRabbit is now a calibrated reviewer for our project!

---

## How to Preserve This Learning

### Option A: Document in MEMORY.md
Add to curated knowledge base:
```
CodeRabbit (Cotton) has been trained on our 80/20 framework.
See: COTTON_FEEDBACK_FRAMEWORK.md
Applied successfully in PR #64.
```

### Option B: Create .coderabbit.yaml
Configure CodeRabbit with our preferences:
```yaml
# Teach CodeRabbit our priorities
review_profile:
  critical: [security, crashes, broken_promises]
  major: [performance, api_design]
  minor: [typos, formatting]
  defer_to: technical_debt_label

decision_framework:
  max_fix_time: 30min
  shipping_over_perfection: true
  context_over_rules: true
```

### Option C: Reference in Future PRs
When CodeRabbit reviews future PRs:
```
@cotton See PR #64 for our established framework.
This is [type of issue] → [action per framework].
```

---

## The Meta-Lesson

**AI reviewers can learn!** By:
1. Explaining our reasoning
2. Giving specific examples
3. Being consistent
4. Acknowledging their good catches

**Result:** CodeRabbit went from generic bot to calibrated reviewer aligned with our values.

---

## Action Items

- [ ] Reply to CodeRabbit's comment acknowledging its learning
- [ ] Document this success in MEMORY.md
- [ ] Consider adding debt review cadence to Sprint planning
- [ ] Reference PR #64 in future teaching moments

---

**This is how we train AI teammates.** 🎓
