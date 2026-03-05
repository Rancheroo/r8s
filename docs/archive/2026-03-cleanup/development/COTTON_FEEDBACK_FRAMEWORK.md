# How We Respond to Cotton (Training Our AI Reviewer)

**Goal:** Help Cotton learn our 80/20 decision-making philosophy

---

## Response Templates

### For Critical Issues We Fix
```
@cotton Good catch! Fixed in commit X.

**Why this mattered:** [Explain the impact]

**What we changed:** [Brief description]
```

### For Issues We Defer (80/20)
```
@cotton Valid point, deferring to v0.8.1.

**80/20 reasoning:** This is [correctness/documentation/optimization] debt, not [security/crash/data loss] debt. The fix adds [X% effort] for [Y% value] — below our threshold for blocking merge.

**Tracked in:** #65
```

### For "It Depends" Decisions
```
@cotton Interesting — we made a conscious tradeoff here.

**Context:** [Why we did it this way]

**Tradeoff:** [A vs B]

**Decision:** We chose X because [reasoning]. Open to revisiting if [condition].
```

---

## Our Decision Framework (Teach Cotton)

### 1. Shipping > Perfection
"We prioritize getting working software to users over perfect code."

### 2. Critical Path First
"Block merge only for: crashes, data loss, security, or broken core functionality."

### 3. Documentation Over Code
"If it can't be fixed in 30 minutes, document it and ship."

### 4. Test the Fix, Not the Principle
"Don't delay for architectural purity when the pragmatic fix works."

### 5. Learn in Production
"Real usage teaches us more than hypothetical review scenarios."

---

## Example Responses to CodeRabbit (Practice for Cotton)

### Issue: CI Tests Disabled
**Our response:**
```
@cotton Valid concern — acknowledged.

**80/20:** Re-enabling CI tests requires infrastructure investigation (2-3 days). The risk of silent regression is real but mitigated by:
1. Local test enforcement (pre-commit hooks planned)
2. Manual verification before merge
3. Issue #65 tracking full fix

**Decision:** Merge with documented workaround, fix infrastructure in Sprint 10.
```

### Issue: Inconsistent Error Handling
**Our response:**
```
@cotton Good architectural observation.

**Tradeoff:** Standardizing error handling across all commands is a 2-hour refactor. The current code works correctly — it just uses two patterns (return error vs os.Exit).

**80/20:** This is refactoring debt, not correctness debt. Tracking in #65 for v0.8.1 cleanup.
```

### Issue: Hardcoded Confidence Threshold
**Our response:**
```
@cotton Agreed — should be configurable.

**Context:** The 0.7 threshold was chosen based on 50+ test cases during Sprint 8 development. It's empirically stable.

**Decision:** Add `--confidence` flag in v0.8.1 when we add user-facing AI tuning. Not blocking for current release.
```

---

## When We Override Cotton

### Always Fix (No Discussion)
- Security vulnerabilities
- Data loss risks
- Crash bugs
- Broken exit codes (promised behavior)

### Always Defer (Explain Why)
- Typos in comments
- Whitespace issues
- Function complexity (unless unreadable)
- Missing docstrings (unless public API)
- "Could be cleaner" suggestions

### Discuss (Edge Cases)
- Performance concerns (need benchmarks)
- Architectural changes (need tradeoff analysis)
- API design (need usage data)

---

## Teaching Cotton Our Voice

### Instead of: "We don't have time"
**Say:** "We prioritized shipping v0.8.0-alpha over this optimization. Tracked in #65."

### Instead of: "It's good enough"
**Say:** "This meets our correctness threshold. Refinements welcome in follow-up PRs."

### Instead of: "We don't care"
**Say:** "This is valid technical debt — we've documented it and will address in v0.8.1."

### Instead of: "You're wrong"
**Say:** "We made a different tradeoff based on [context]. Here's our reasoning..."

---

## Building Trust with Cotton

1. **Acknowledge valid concerns** — Don't dismiss, explain
2. **Show work** — Reference commits, issues, tests
3. **Be consistent** — Same framework every time
4. **Invite learning** — "Watch how we handle this in v0.8.1"

---

## Response Checklist

Before posting reply to Cotton:
- [ ] Did we fix it? Reference commit.
- [ ] Did we defer it? Reference issue #65.
- [ ] Did we explain our reasoning? (80/20, shipping > perfection, etc.)
- [ ] Did we invite continued learning? ("See how we handle X in Sprint 10")

---

**Next:** When Cotton comments on PR #64, use these templates to respond and teach.
