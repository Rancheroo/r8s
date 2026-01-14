# r8s Development Principles

This document captures the core principles that guide r8s development. When making design decisions, refer to these principles.

---

## Core Principles

### 1. Truth Only™
> **r8s only displays truth. Better to show nothing than wrong data. Remove features that lie.**

**Example**: v0.4.3 removed dashboard log scanning entirely when counts were inaccurate, rather than ship misleading data.

**Application**: Before displaying any metric, verify it's accurate. If accuracy can't be guaranteed, don't show it.

---

### 2. Best Feature = No Feature
> **Smart defaults beat options. Delete complexity, don't improve it.**

**Example**: v0.3.5 deleted 1,200 lines (live mode) in 22 minutes because bundle analysis was the primary use case.

**Application**: When facing complexity, ask "Can we delete this instead of fixing it?" Removal is often the right answer.

---

### 3. Show, Don't Ask
> **Information should surface automatically without button presses.**

**Example**: Dashboard auto-displays health summary ("🔥 5 critical · ⚠️ 4 warnings") without requiring any hotkey.

**Application**: Users shouldn't need to press keys to see important information. Display it proactively.

---

### 4. Explicit > Implicit
> **Clear modes, flags, and indicators reduce confusion.**

**Example**: Mode indicators in breadcrumb ([BUNDLE] / [DEMO] / [LIVE]) show exactly what data source is active.

**Application**: Make the current state obvious. Users should never wonder "what am I looking at?"

---

### 5. Complete Removal
> **Partial deletions create tech debt. Delete entirely or keep fully.**

**Example**: Mock mode removal in v0.5.0 - deleted all 9 getMock* functions completely, not just commented out.

**Application**: When removing a feature, delete ALL related code. No commented-out blocks, no "maybe we'll need this later."

---

### 6. Test at 10× Scale
> **Test features at max values. UI scaling breaks exponentially.**

**Example**: v0.4.0 dashboard overflow when --scan=1000 detected 80+ issues. Now always test at maximum parameter values.

**Application**: Test with --scan=1000, bundles with 500+ pods, logs with millions of lines. Find breaking points early.

---

### 7. Use Your Interfaces
> **Trust abstractions you designed. Bypassing creates bugs.**

**Example**: v0.3.2 bug where describePod() bypassed DataSource interface, causing mode-specific failures.

**Application**: If you designed an interface, use it everywhere. Type assertions and direct access = code smell.

---

### 8. O(n log n) Always
> **Algorithm complexity matters. Use standard library sorting.**

**Example**: Replaced bubble sort O(n²) with sort.Slice O(n log n) across 4 functions in v0.5.2.

**Application**: Never write manual sorting. Use sort.Slice, sort.SliceStable from stdlib.

---

### 9. 30-Day Branch Rule
> **Delete merged branches within 30 days. Clean repos accelerate development.**

**Example**: v0.5.1 cleanup deleted 7 stale merged branches (80% reduction), eliminated mental overhead.

**Application**: After merging, immediately delete the branch. If it's been merged >30 days, delete it without asking.

---

### 10. Minimal Keys
> **Fewer hotkeys = better UX. Enter/Esc navigation should suffice for 90% of workflows.**

**Example**: Dashboard → logs in 2 keys (Enter, l). Classic mode exists for exploration, not as primary flow.

**Application**: Before adding a hotkey, ask "Can we make this the default behavior instead?"

---

### 11. Empty is Valid
> **Don't conflate "no results" with "error". Empty is a legitimate state.**

**Example**: Bundle with zero pods in namespace shows "📭 Empty", not an error or fake data.

**Application**: Design for zero-state UX. Empty results deserve clear messaging, not errors.

---

### 12. Pause for Review
> **Stop after feature branch commit. Only merge to main after explicit approval.**

**Example**: v0.6.8 - Should have stopped after committing to feature branch, waited for review before merge/tag/push.

**Application**: Release workflow must pause after branch commit:
1. Implement feature on feature branch
2. Run automated tests (`make test`)
3. Commit to feature branch
4. **STOP** - Request review before proceeding
5. Only after approval: merge to main, tag, push

**Why**: Allows review of implementation before it becomes production. Easier to iterate on feature branch than to revert from main.

---

## Development Philosophy

### Maximum Information Extraction (v0.5.5-v0.5.6)
> **Extract and display all available bundle data, even when incomplete. Be transparent about data gaps.**

**Example**: Container status section shows "1/2 ready" even without exit codes, provides kubectl command for details.

**Application**: Parse everything in the bundle. When data is missing, explain what's missing and how to get it.

---

### "r8s interprets, user acts" (v0.5.4)
> **Show intelligence, not just information. Users should know WHAT to investigate and WHY.**

**Example**: Diagnostic panel translates "CrashLoopBackOff" into actionable "Container repeatedly failing to start" with investigation steps.

**Application**: Don't just display raw Kubernetes output. Interpret it and provide actionable guidance.

---

## Practical Decision Framework

### When to Remove vs Fix

**Remove immediately if**:
- Data accuracy cannot be verified
- Bug affects user trust in the tool
- Fix is complex and feature is rarely used

**Fix instead if**:
- Data is mostly correct (>95% accuracy)
- Feature is critical to core workflow
- Root cause is clear and fix is simple

---

### Code Quality Patterns

**Prefer**:
- Specific methods > generic methods
- Type safety > flexibility
- Wrap text first, THEN apply styling (ANTML escape codes)
- Critical items deserve special handling (dynamic caps)

**Always**:
- Use interfaces you designed
- Extract reusable functions (DRY principle)
- Name things clearly (no abbreviations unless obvious)

---

### Development Process

1. **Make it work** - Get the feature functional
2. **Make it right** - Apply principles, clean up code
3. **Make it fast** - Optimize only if needed

**In that order.** Don't optimize prematurely.

---

## Examples in Action

### v0.3.5: Complete Feature Removal
**Situation**: Live mode barely used, bundle mode is 95% of use cases  
**Decision**: Delete all 1,200 lines of live mode code  
**Time**: 22 minutes from decision to tagged release  
**Principle**: Best Feature = No Feature

### v0.4.3: Truth Over Features
**Situation**: Dashboard log counts inaccurate (showed same counts for all pods)  
**Decision**: Remove log scanning entirely until it can be verified accurate  
**Result**: Dashboard shows less data, but 100% truthful  
**Principle**: Truth Only™

### v0.5.0: Complete Cleanup
**Situation**: Mock mode deleted, but test helpers used mock functions  
**Decision**: Delete ALL 9 getMock* functions, update tests to use real data sources  
**Result**: Zero dead code, clean codebase  
**Principle**: Complete Removal

---

## Quick Reference

When designing a feature, ask:

1. ✅ **Is this data 100% accurate?** (Truth Only™)
2. ✅ **Can we delete this instead?** (Best Feature = No Feature)
3. ✅ **Does it show automatically?** (Show, Don't Ask)
4. ✅ **Is the state explicit?** (Explicit > Implicit)
5. ✅ **Did we test at 10× scale?** (Test at 10× Scale)

**Before merging to main:**

6. ✅ **Did we pause for review?** (Pause for Review)
   - Feature branch committed?
   - Tests passing?
   - Review requested before merge?

---

## Updates

This document should be updated whenever:
- A new principle emerges from development experience
- An existing principle is refined or clarified
- A significant decision demonstrates principle application

**Last Updated**: v0.6.8 (2026-01-14)
