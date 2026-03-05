# Quick Wins Roadmap: r8s v0.7.x Series

**Session Date:** 2026-02-14  
**Participants:**  
- **Rex (RancherSRE):** "Code talks, plans walk. Let's find the low-hanging fruit."  
- **Luna (Launchpad):** "Agreed, but let's make sure that fruit doesn't rot next week. Strategy first."

---

## 🚀 Executive Summary

**Rex:** We shipped v0.7.0 with 90% coverage. Great. Now the CI is red and linting is off. We need to fix the foundation before we pile on K3s and AI.  
**Luna:** Precisely. This roadmap focuses on "cleaning the kitchen" (Sprint 6) so we can "cook the banquet" (v0.7.1+). We applied Musk's 5 Laws to strip down requirements to their bare essentials.

**Theme:** Stability is Speed.

---

## 1. Immediate Wins (This Week)
*Criteria: <4 hours effort, high visibility, fixes immediate pain.*

### A. Enable CI Coverage Threshold at 10%
- **What:** Re-enable the `coverage` check in GitHub Actions, but set the fail threshold to 10% (current actual).
- **Why:** **Stops the bleeding.** Right now, we can merge code with 0% tests. This puts a floor on quality immediately.
- **Effort:** 0.5 hours
- **Complexity:** 1/5
- **Musk Law:** #1 Question every requirement (Don't aim for 50% *today*, aim for *non-zero* today).
- **Quality Gate:** "Ship Value in Manageable Bites" (Principle #13).

### B. "make lint" Target
- **What:** Add a `lint` target to the `Makefile` that installs and runs `golangci-lint`.
- **Why:** Devs (me) forget the command. If `make test` works, `make lint` should too. Accelerates local feedback loop.
- **Effort:** 0.5 hours
- **Complexity:** 1/5
- **Musk Law:** #5 Automate (Don't ask humans to remember flags).
- **Quality Gate:** "Debug Like You're Not There" (Principle #4 - consistency).

### C. Document `os.Exit(2)` Decision
- **What:** Add a comment/doc explaining why we force exit on bundle load failure (CodeRabbit complaint).
- **Why:** Stops the endless "fix this" noise from the bot. Resolves technical debt by decision, not code.
- **Effort:** 0.25 hours
- **Complexity:** 1/5
- **Musk Law:** #2 Delete unnecessary parts (Delete the ambiguity).
- **Quality Gate:** "Fail Gracefully, Explain Clearly" (Principle #3).

### D. `FormatK3s` Constant Definition
- **What:** Add `FormatK3s` to `internal/bundle/types.go` and a placeholder detector.
- **Why:** It's a tiny change that "reserves the parking space" for v0.7.1. Prevents huge refactors later.
- **Effort:** 0.5 hours
- **Complexity:** 1/5
- **Musk Law:** #3 Simplify (Prepare the structure before the logic).
- **Quality Gate:** "Version Awareness" (Principle #5).

### E. Fix Top 5 Lint Spam
- **What:** Fix the 5 most frequent lint warnings (likely error handling or unused vars) that clutter the logs.
- **Why:** Reduces noise so we can see the real signals.
- **Effort:** 2 hours
- **Complexity:** 2/5
- **Musk Law:** #2 Delete unnecessary parts (Delete the noise).
- **Quality Gate:** "Fail Gracefully" (Principle #3).

---

## 2. Sprint 6 Wins (CI Stability)
*Criteria: Enables v0.7.1, removes blockers, "eat your vegetables" phase.*

### A. The "Golden" Test Bundle
- **What:** Create a minimal, sanitized `test/testdata/golden_bundle` committed to git.
- **Why:** We rely on local bundles too much. CI needs a reliable target.
- **Effort:** 4 hours
- **Complexity:** 2/5
- **Musk Law:** #5 Automate (CI needs data to run).
- **Quality Gate:** "Test with Real Bundles" (Principle #9).

### B. Helper Function Unit Tests (TUI)
- **What:** Test the *logic* of the TUI (formatters, data transformers) without testing the *UI* (Bubble Tea).
- **Why:** TUI testing is hard/flaky. Logic testing is easy/fast. 80% of bugs are in the logic.
- **Effort:** 6 hours
- **Complexity:** 3/5
- **Musk Law:** #3 Simplify (Don't test the rendering, test the data).
- **Quality Gate:** "Ship Value in Manageable Bites" (Principle #13).

### C. Cross-Platform Build Check
- **What:** Add a GitHub Action matrix for Linux/macOS/Windows builds.
- **Why:** We broke Windows support in v0.6.x silently. Catch it early.
- **Effort:** 2 hours
- **Complexity:** 2/5
- **Musk Law:** #5 Automate.
- **Quality Gate:** "Version Awareness" (Principle #5).

---

## 3. v0.7.1 Wins (K3s Support)
*Criteria: Reduced scope, 80/20 rule, high impact.*

### A. The "Magic" Path Interface
- **What:** Refactor `internal/bundle` to use an interface for file paths, defaulting to RKE2 but swappable.
- **Why:** **Rex:** "Stop hardcoding `rke2/` everywhere!" **Luna:** "This allows us to support K3s *and* RKE1 with one pattern."
- **Effort:** 8 hours
- **Complexity:** 4/5
- **Musk Law:** #3 Simplify (One abstraction, multiple implementations).
- **Quality Gate:** "The Collector Script is a Moving Target" (Principle #6).

### B. K3s Detection (Only)
- **What:** Detect K3s bundles and print "K3s Detected" (even if parsing fails).
- **Why:** User value: "At least it knows what I am." Better than "Unknown Format."
- **Effort:** 2 hours
- **Complexity:** 2/5
- **Musk Law:** #1 Question every requirement (Do we need to *parse* it all yet? No, just identify it).
- **Quality Gate:** "Signal When Data is Missing" (Principle #11).

### C. 5 Core File Abstraction
- **What:** Only abstract the top 5 most used files (nodes, pods, events, logs, version). Ignore the obscure ones for now.
- **Why:** Covers 90% of use cases. RKE1/K3s users mostly check pods/nodes.
- **Effort:** 4 hours
- **Complexity:** 3/5
- **Musk Law:** #2 Delete unnecessary parts (Don't refactor the obscure parsers yet).
- **Quality Gate:** "Multi-Source Fallbacks" (Principle #8).

---

## 4. Future Wins (v0.7.3+)
*Criteria: High value, foundation for AI, performance.*

### A. Error Pattern Grouping (AI MVP)
- **What:** Simple regex grouper: "Seen 50 times: 'etcd server timeout'".
- **Why:** **Rex:** "This isn't AI, it's counting." **Luna:** "Users don't care. It looks like magic to them." Delivers 80% of 'AI' value with 1% of the cost.
- **Effort:** 6 hours
- **Complexity:** 3/5
- **Musk Law:** #1 Question every requirement (Do we need an LLM? No, we need `map[string]int`).
- **Quality Gate:** "Parse Defensively" (Principle #7).

### B. Lazy Loading Dashboard
- **What:** Only parse tab data when the user clicks the tab.
- **Why:** Startup time is getting slow.
- **Effort:** 8 hours
- **Complexity:** 4/5
- **Musk Law:** #4 Accelerate cycle time (User perception of speed).
- **Quality Gate:** "Ship Value in Manageable Bites" (Principle #13).

---

## 🏁 Final Thoughts

**Rex:** "I like the 'FormatK3s' placeholder. I can knock that out while my coffee brews. The 'Magic Path' interface scares me a bit, but it's necessary."

**Luna:** "The 'Golden Bundle' is the strategic win here. Without it, our CI is just a syntax checker. We need data-driven tests to support K3s safely."

**Action:**
1. Rex takes **Immediate Wins A & B** today.
2. Luna updates the **Sprint 6 Plan** to include the Golden Bundle task.
3. We review the **Magic Path** design on Monday.

*Signed,*  
*Rex & Luna*
