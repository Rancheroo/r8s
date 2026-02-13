# Working with CodeRabbit

This guide helps us teach CodeRabbit our patterns and get better reviews over time.

## How CodeRabbit Learns

### 1. Configuration (Immediate)
The `.coderabbit.yaml` file teaches CodeRabbit:
- Our principles (Truth Only™, etc.)
- Common patterns in our codebase
- Files that define our architecture
- How we typically respond to certain issues

### 2. Review Responses (Over Time)
When we reply to CodeRabbit comments, it learns:
- What patterns are intentional vs bugs
- Our tolerance for different types of issues
- How we prioritize fixes

**Good response format:**
```
@coderabbitai [Acknowledged/Fixed/Deferred]

Reason: [Why this is or isn't an issue]

Action: [What we're doing about it]

Future: [If applicable, when we'll address it]
```

### 3. Code Patterns (Continuous)
CodeRabbit learns from:
- How we structure new code
- What we accept vs reject in PRs
- Our test patterns
- Our error handling approaches

## Teaching CodeRabbit New Patterns

### When CodeRabbit Catches a Real Bug
✅ **Good:** "Fixed in commit X. Good catch!"
- This reinforces that pattern detection is valuable

### When CodeRabbit is Wrong
✅ **Good:** "@coderabbitai This is intentional because [reason]. Pattern is [explanation]."
- This teaches it to recognize our intentional patterns

### When CodeRabbit Suggests Something We Don't Want
✅ **Good:** "@coderabbitai We don't do this because [principle]. See PRINCIPLES.md #[number]."
- This connects suggestions to our principles

## Common Patterns to Reinforce

### Pattern: Tolerant Parsing
```go
// We do this:
_, _ = fmt.Sscanf(s, "%d", &n) // Tolerant parsing

// Not this:
if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
    return err // Too strict for bundle parsing
}
```

### Pattern: Intentional Exit Codes
```go
// We do this in test-cluster command:
os.Exit(2) // Signal test failure to CI

// Because:
// - Exit code 0 = success
// - Exit code 1 = general error  
// - Exit code 2 = tests failed (CI contract)
```

### Pattern: Multiple Systems
```go
// We have multiple "completeness" systems:
// - datasource.BundleHealth (UI layer, simple)
// - bundle.AnalyzeCompleteness (validation layer, detailed)
// 
// They're intentionally separate - don't suggest consolidation
```

### Pattern: Test Names
```go
// Bad: TestAnalyzeCompleteness_Complete
// (implies 100% complete)

// Good: TestAnalyzeCompleteness_RequiredFilesPresent
// (clarifies what we're testing)
```

## Updating the Config

When we find patterns CodeRabbit should know about:

1. **Add to `.coderabbit.yaml`**:
   - New principles
   - New patterns
   - New common responses

2. **Update this file**:
   - Document the pattern
   - Show examples
   - Explain rationale

3. **Reference in PRs**:
   - "See .coderabbit.yaml principle #[id]"
   - Helps connect config to reality

## Reviewing CodeRabbit's Reviews

### Quality Checklist
- [ ] Did it catch real issues?
- [ ] Did it suggest principle violations?
- [ ] Did it understand our patterns?
- [ ] Were suggestions actionable?
- [ ] Did it avoid false positives?

### Giving Feedback
If CodeRabbit consistently misses something:
1. Add it to `knowledge_base.patterns` in config
2. Reference it in next review response
3. Update this guide

## Sprint Integration

### At Sprint Start
- Review `.coderabbit.yaml` for relevance
- Update patterns based on new features
- Check if principles need additions

### At Sprint End
- Review CodeRabbit's hit rate (good catches vs noise)
- Update config based on learnings
- Document any new patterns discovered

## Principles for CodeRabbit Itself

When working with CodeRabbit, remember:

1. **It's a tool, not a replacement** - Final decisions are ours
2. **Teach, don't just correct** - Explain why something is wrong
3. **Be consistent** - Same response to same issue type
4. **Update the config** - Don't rely on memory
5. **Give credit** - When it catches real bugs, acknowledge it

---

*This file helps us get better reviews by teaching CodeRabbit our patterns.*

*Last updated: Sprint 5 Phase 2*
