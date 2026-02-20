# CodeRabbit Communication Protocol (Updated 2026-02-20)

## Summary of Changes

### 1. Use @coderabbit (Not "Cotton")
**Reason:** Avoid alias confusion — one clear name everywhere

### 2. Threaded Replies (Not Main Thread)
**Reason:** CodeRabbit gets better context, cleaner PR review

---

## How to Reply to CodeRabbit

**Correct Handle:** @CodeRabbitAI (one word, capitalized)

### Threaded Reply (Preferred)
```bash
COMMIT=$(git rev-parse HEAD)

echo '{
  "body": "@CodeRabbitAI Your response here",
  "commit_id": "'$COMMIT'",
  "path": "cmd/file.go",
  "line": 45,
  "in_reply_to": COMMENT_ID
}' | gh api repos/Rancheroo/r8s/pulls/64/comments --input -
```

**Result:** Appears under CodeRabbit's comment as threaded conversation

### General PR Comment (For Summaries Only)
```bash
gh pr comment 64 --body "@coderabbit Summary..."
```

**Result:** Appears in main PR thread

---

## When to Use Each

| Type | Method | Example |
|------|--------|---------|
| **Acknowledge issue** | Threaded | "Good catch - this is blocking because..." |
| **Explain 80/20 decision** | Threaded | "Deferring to #65 - here's why..." |
| **Mark as fixed** | Threaded | "Fixed in commit X - marking resolved" |
| **High-level summary** | General | "All critical issues addressed..." |
| **Thank you / wrap-up** | General | "Thanks for the thorough review!" |

---

## Resolve Workflow

1. **Reply to thread:** "Fixed in commit X"
2. **User clicks Resolve** in web UI (no API for this)
3. **Or** mark PR review as approved

---

## Example Responses

### Critical Issue (Fixed)
```
@CodeRabbitAI Fixed in commit 7626f2d.

What we changed: Added OnlyValidArgs validation
Why: Broke exit code promise to users
Blocking: Yes - user trust issue
```

### Major Issue (Deferred)
```
@coderabbit Acknowledged - deferring to #65.

80/20 reasoning: Risk mitigated, fix >30 min
Decision: Ship with workaround, fix in Sprint 10
Type: Healthy debt (tracked, bounded)
```

### Teaching Moment
```
@coderabbit Teaching moment:

We prioritize shipping > perfection because:
1. Real usage teaches more than review
2. 95% solution that ships > 100% that's late
3. Technical debt documented in #65

See COTTON_FEEDBACK_FRAMEWORK.md for full philosophy.
```

---

## Verification

✅ Tested threaded replies on PR #64 - works perfectly
✅ CodeRabbit receives proper context
✅ Clean threaded conversation in PR
⚠️ Resolve button requires web UI (no API)

---

## Quick Reference

| Task | Command |
|------|---------|
| Get comment IDs | `gh api repos/.../pulls/64/comments \| jq '.[].id'` |
| Threaded reply | `gh api repos/.../pulls/64/comments --input -` |
| General comment | `gh pr comment 64 --body "..."` |

---

**Effective immediately:** All CodeRabbit responses use threaded replies with @coderabbit.
