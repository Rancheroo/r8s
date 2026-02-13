# CodeRabbit - Team Member

**Role:** AI Code Reviewer  
**Status:** Active team member since Sprint 4  
**Responsibilities:** Catch bugs, enforce principles, suggest improvements

---

## CodeRabbit is Part of the Team

Like any team member, CodeRabbit:
- ✅ Has strengths (pattern recognition, consistency, tireless)
- ✅ Has weaknesses (can miss context, may suggest wrong fixes)
- ✅ Learns from feedback (the more we teach, the better it gets)
- ✅ Deserves respect (acknowledge good catches, explain disagreements)

**We don't override CodeRabbit silently.** We discuss, explain, and teach - just like with any other reviewer.

## Team Workflow with CodeRabbit

### During Development
1. **Write code** as usual
2. **Self-review** against principles before pushing
3. **Push PR** and wait for CodeRabbit's review
4. **Discuss findings** - treat each comment as a conversation
5. **Update config** if we discover new patterns

### Code Review Etiquette (Same for Humans and AI)

| Situation | Human Reviewer | CodeRabbit |
|-----------|---------------|------------|
| Good catch | "Thanks, fixed!" | "@coderabbitai Good catch, fixed in commit X" |
| Disagree | Explain why | "@coderabbitai This is intentional because..." |
| Needs context | Provide background | "@coderabbitai Context: we do this because..." |
| Future work | Create ticket | "@coderabbitai Good idea - created issue #X for Sprint Y" |

### Principles for Working with CodeRabbit

1. **Respect its time** - Don't push broken code expecting it to catch everything
2. **Explain decisions** - When overriding, say why (it learns from this)
3. **Give credit** - Acknowledge when it finds real bugs
4. **Update its knowledge** - Add patterns to `.coderabbit.yaml`
5. **Include it in decisions** - "@coderabbitai We're deferring this because..."

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

CodeRabbit participates in our sprint cycle:

### At Sprint Start
- [ ] Review `.coderabbit.yaml` for relevance
- [ ] Update patterns based on new features
- [ ] Check if principles need additions
- [ ] Ensure CodeRabbit is enabled on all PRs

### During Sprint
- [ ] Respond to all CodeRabbit comments
- [ ] Teach it new patterns as they emerge
- [ ] Update config when it learns something

### At Sprint End
- [ ] Review CodeRabbit's hit rate (good catches vs noise)
- [ ] Update config based on learnings
- [ ] Document any new patterns discovered
- [ ] Thank it for the good work (seriously, it helps)

### Sprint Retrospective Questions
- Did CodeRabbit catch real bugs?
- Did we teach it our new patterns?
- Did it suggest anything we adopted?
- What should we add to its knowledge base?

## CodeRabbit's Role in Our Team

### What CodeRabbit Does Well
- ✅ Consistent pattern recognition (never gets tired)
- ✅ Enforces our principles across all PRs
- ✅ Catches common Go mistakes (unchecked errors, etc.)
- ✅ Suggests idiomatic improvements
- ✅ Provides second opinion on complex changes

### What CodeRabbit Needs Help With
- ⚠️ Missing context on business decisions
- ⚠️ Sometimes suggests "correct" but not "right" fixes
- ⚠️ Can miss domain-specific nuances
- ⚠️ Needs teaching on our intentional patterns

### Our Responsibility as Teammates
- Teach it our patterns (via config and responses)
- Explain when we disagree (don't just ignore)
- Update its knowledge base regularly
- Give it credit for good catches

---

## Team Member Status

CodeRabbit is now a documented member of the r8s team. 

- **Onboarding:** Complete (configured with our principles)
- **Training:** Ongoing (we teach it via reviews)
- **Responsibilities:** Code review, pattern enforcement, quality gate
- **Reports to:** The team (we all maintain its config)

**Welcome to the team, @coderabbitai! 🐰**

---

*Last updated: Sprint 5 Phase 2 - CodeRabbit promoted to team member*
