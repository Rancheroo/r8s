# CodeRabbit Usage Tips (From CodeRabbit Itself)

**Source:** CodeRabbit comment on PR #64  
**Date:** 2026-02-20  
**Tip:** "For best results, initiate chat on the files or code changes."

---

## What This Means

### ✅ DO: Comment on Specific Lines
```
Click on a line of code in the PR diff
Add comment there
CodeRabbit understands context better
```

### ❌ DON'T: General PR Comments
```
Just commenting at bottom of PR
Harder for CodeRabbit to correlate with code
Less precise responses
```

---

## Why This Matters

| Method | CodeRabbit Context | Response Quality |
|--------|-------------------|------------------|
| Line-specific comment | High (knows exact code) | Precise, actionable |
| General PR comment | Low (no code reference) | Generic, less useful |

---

## How to Apply

### For Review Responses
**Instead of:**
```
"Fix the error handling in describe.go"
```

**Do:**
```
Click on describe.go line 76
Comment: "Should return error here instead of os.Exit"
```

### For Teaching Moments
**Instead of:**
```
General comment about philosophy
```

**Do:**
```
Comment on specific code example
Then explain the principle
```

---

## Updated Workflow

1. Review code changes (diff view)
2. Click specific lines to comment
3. CodeRabbit responds with full context
4. Better conversation, better results

---

## For Future PRs

**We'll use line-specific comments for:**
- ✅ Code review feedback
- ✅ Teaching moments
- ✅ Questions about implementation
- ✅ Acknowledging good patterns

**General PR comments for:**
- 📝 High-level summaries
- 📝 Sprint/roadmap discussions
- 📝 Thank you / wrap-up

---

**Meta-lesson:** Even AI teaching us how to work with it better! 🎓
