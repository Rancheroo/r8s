# R8s Principles

Core principles guiding R8s development and support operations.

*Living document - evolves with every incident and feature.*

---

## The 10 Principles (Revised for CLI-First)

### 1. **Log Bundle is the Source of Truth**

Every feature, fix, and diagnostic **must** account for the log bundle structure.

**Before writing any code:**
- [ ] Check where the data comes from in the bundle
- [ ] Verify the file format (table, JSON, YAML, text)
- [ ] Confirm the path exists in all bundle versions (RKE2, K3s, Rancher)

**Rule:** If the collector script changes, we break. Monitor it weekly.

---

### 2. **Empty is Valid**

Bundles are often partial, corrupted, or from failed collections.

**Every parser must:**
- Handle missing files gracefully
- Work with incomplete data
- Never crash on missing fields
- Provide best-effort results

**Example:** If `poddescribe/` doesn't exist, fall back to `pod-manifests/`, then to basic `kubectl/pods`.

---

### 3. **Fail Gracefully, Explain Clearly**

When something goes wrong, tell the user **what** and **why**.

**Good:**
```
⚠️  QoS class unavailable - pod describe data not found in bundle.
   This bundle was collected before PR #418 (Feb 2026).
   Falling back to manifest parsing...
```

**Bad:**
```
Error: nil pointer dereference
```

---

### 4. **CLI First & Compatible**

We are building `kubectl` for bundles. Muscle memory matters.

**Rules:**
- Use standard kubectl verbs (`get`, `describe`, `logs`)
- Support standard flags (`-n`, `-o`, `-l`)
- Output should mimic kubectl output by default
- Interactive features (like `ask`) should be additive, not replacing standard commands

---

### 5. **Version Awareness**

Bundles change over time. Features appear and disappear.

**Implementation:**
- Detect bundle format version when possible
- Support multiple parsing strategies
- Document version compatibility matrix
- Test with bundles from different time periods

---

### 6. **Parse Defensively**

Kubernetes output formats change. kubectl tables aren't contracts.

**Rules:**
- Don't assume column positions
- Handle extra/missing columns
- Use regex for flexible matching
- Validate before using parsed data

---

### 7. **Multi-Source Fallbacks**

Critical data should be available from multiple sources.

**Example - Container Names:**
1. Primary: `poddescribe/` (PR #418)
2. Fallback: `pod-manifests/`
3. Fallback: Log filename parsing
4. Last resort: Single container assumption

---

### 8. **Test with Real Bundles**

Unit tests are necessary but not sufficient.

**Required:**
- Test with actual customer bundles (anonymized)
- Test bundles from different Rancher versions
- Test partial/failed collections
- Test edge cases (empty namespaces, huge pods)

---

### 9. **Signal When Data is Missing**

Users should know when they're not seeing the full picture.

**UI Patterns:**
- "⚠️ QoS data unavailable - bundle collected before Feb 2026"
- "⚠️ Showing 1 of 3 containers - describe data not found"
- "ℹ️ Partial bundle - some logs may be missing"

---

### 10. **Focus on Value (80/20 Rule)**

Prioritize the **20% of work that delivers 80% of the value**. Look for low-hanging fruit with big impact.

**Guiding Questions:**
- "Is this the smallest valuable slice?"
- "Are we over-engineering a corner case?"
- "Can we solve this simply first?"

Break complexity into quick wins. Never do "all foundation, no features."

---

### 11. **Universal Interoperability & Automation**

We are not just a CLI; we are a data source for the ecosystem.

**The Golden Rule:**
Everything must be consumable by **machines** and **people**.
- **Log Bundles with Everything:** Output must be ready for Salesforce, Slack, GitHub Issues, and other tools.
- **Formats:** Support JSON (for jq/scripts), SARIF (for code scanning), and Markdown (for humans/Slack).
- **Automation:** Enabling automation is a primary goal. If it can be scripted, it should be.

**Don't build dead ends.** Build pipelines that feed into the larger support ecosystem.

---

### 12. **Zero Configuration Start**

Reduce friction to zero.

**Rule:**
The user should never *need* to edit a config file to get value.
- Auto-detect bundle formats.
- Auto-detect log styles.
- Defaults should cover 80% of use cases.

**Violation:** Requiring a user to specify `--type=rke2` when the folder structure screams RKE2.

---

### 13. **Delete Zombie Code**

Maintenance cost > feature value.

**Musk's Law: Delete the part or process.**
If a fallback parser or legacy feature hasn't been triggered or useful in recent versions, **delete it**. You can always add it back later.

**Don't maintain code "just in case."**

---

### 14. **Stream, Don't Load**

Support bundles can be massive (10GB+).

**Performance Rules:**
- Never read an entire file into memory unless it's strictly bounded (e.g., config).
- Use `bufio.Scanner` or streaming decoders.
- Paginate results implicitly.

**Goal:** Start analyzing in < 200ms, regardless of bundle size.

---

### 15. **UX is a First-Class Citizen**

Even in a terminal, User Experience drives adoption.

**Core Tenets:**
- **Clarity:** Error messages should guide the user to a solution (see Principle 3).
- **Intuitiveness:** Flags and commands should guess what the user wants.
- **Delight:** Small touches (colors, spinners, clear summaries) make the tool a partner, not a burden.

---

### 16. **The Engineering Algorithm (The 5 Laws)**

We follow the 5-step process for every feature and process:

1.  **Make Requirements Less Dumb:** Question constraints. Does the user *really* need a full TUI for this?
2.  **Delete the Part/Process:** If a feature isn't used, kill it. (See Principle 13).
3.  **Simplify or Optimize:** Do not optimize what should not exist.
4.  **Accelerate Cycle Time:** Ship faster.
5.  **Automate:** If you're doing it twice, script it. (See Principle 11).

---

## Principle Violations = Technical Debt

Every time we violate these principles, we create future incidents:

| Violation | Debt | Example |
|-----------|------|---------|
| No bundle check | Parser breaks silently | Sprint 4 QoS issue |
| No fallback | Feature fails on old bundles | Container selector |
| No debug logging | Can't diagnose issues | Current debug build needed |
| No version awareness | Can't support multiple formats | N/A yet |

---

*Adopted: 2026-02-12*
*Revised: 2026-03-05 (CLI Pivot)*
