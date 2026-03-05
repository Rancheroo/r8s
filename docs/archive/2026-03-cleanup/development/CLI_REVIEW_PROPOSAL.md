# CLI Structure Review & Quick Wins Analysis

**Status:** Proposed for Sprint 9 or v0.8.0  
**Owner:** UX Team + CLI Lead  
**Effort:** Medium (analysis) + Low (implementation)

---

## Current State (Sprint 8)

### Existing Commands

```
r8s [bundle-path]          # Root - launches TUI with bundle
r8s                        # Root - launches TUI with demo

r8s completion             # Shell completion
r8s config                 # Config management (init, view, edit, validate, set)
r8s test-cluster [path]    # Automated diagnostics
r8s tui [path]             # Launch TUI explicitly
r8s validate [path]        # NEW: Bundle health check
r8s generate prompt [path] # NEW: AI prompt generation
r8s version                # Version info
```

### Problems Identified

| Issue | Example | Severity |
|-------|---------|----------|
| **Inconsistent arg patterns** | `test-cluster [path]` vs `validate [path]` vs `tui [path]` | Medium |
| **No unified output format** | Some use tables, some JSON, inconsistent | Medium |
| **Missing commands** | No `export`, no `create`, no `analyze` | Low |
| **Flag inconsistency** | Global flags vs command-specific flags unclear | Medium |
| **Help text quality** | Varies significantly between commands | Low |
| **No config file support** | Most commands ignore `--config` | Medium |

---

## UX Review Tasks

### 1. Command Naming Audit

**Question:** Are command names intuitive?

| Current | Alternative | Pros/Cons |
|---------|-------------|-----------|
| `test-cluster` | `diagnose`, `check` | `test` implies testing tool, not analysis |
| `generate prompt` | `prompt`, `ai` | Shorter, clearer intent |
| `validate` | `health`, `check` | `validate` is good, but `health` more specific |

**Recommendation:** Keep current names (too late to change), but add aliases.

### 2. Argument Patterns

**Current inconsistency:**
```bash
r8s ./bundle                    # Implicit TUI
r8s tui ./bundle                # Explicit TUI  
r8s validate ./bundle           # Explicit command
r8s test-cluster ./bundle       # Different pattern
r8s generate prompt ./bundle    # Subcommand with arg
```

**UX Question:** Should all commands accept `[bundle-path]` as first arg?

**Proposed Standard:**
```bash
r8s [command] [bundle-path] [flags]
```

Where `[bundle-path]` is:
- Optional for TUI (defaults to demo)
- Required for analysis commands
- Not used for config/version

### 3. Output Format Consistency

**Current:**
```bash
r8s test-cluster --format=table|json|summary
r8s validate --format=table|json|summary
r8s generate prompt --format=chatbot|terminal|script
```

**Proposed Standard:**
```bash
--format=auto       # Detect based on stdout (TTY vs pipe)
--format=table      # Human-readable (default for TTY)
--format=json       # Machine-readable (default for pipe)
--format=yaml       # Alternative machine-readable
--format=summary    # One-line summary
```

### 4. Global vs Command Flags

**Current mess:**
```bash
# Global flags (work on all commands)
--config, --verbose, --scan

# But they don't always make sense:
r8s version --scan=500    # Nonsense, but accepted
```

**Proposed:**
```bash
# Truly global
--config     # All commands
--verbose    # All commands  
--help       # All commands

# Command-specific only
--scan       # Only for log analysis commands
--format     # Only for output commands
```

### 5. Exit Codes

**Current:**
- `test-cluster`: 0=ok, 1=issues, 2=bundle error
- `validate`: 0=valid, 1=incomplete, 2=invalid

**Proposed Standard:**
```bash
0   # Success / No issues
1   # Issues found / Incomplete
2   # Invalid arguments / Bundle error
3   # Runtime error / Exception
```

---

## Quick Wins (Low Effort, High Impact)

### Win 1: Standardize `--format` Flag (2 hours)

**Files:** `cmd/validate.go`, `cmd/testcluster.go`, `cmd/generate.go`

**Change:**
```go
// Consistent format flag
testClusterCmd.Flags().StringVar(&testClusterFormat, "format", "auto", "Output format: auto, table, json, summary")
validateCmd.Flags().StringVar(&validateFormat, "format", "auto", "Output format: auto, table, json, summary")
```

**Impact:** Users learn one pattern, works everywhere.

### Win 2: Add Command Aliases (1 hour)

**Files:** `cmd/*.go`

**Change:**
```go
testClusterCmd = &cobra.Command{
    Use:     "test-cluster [bundle-path]",
    Aliases: []string{"check", "diagnose"},  // NEW
    ...
}
```

**Impact:** Users can use intuitive names.

### Win 3: Consistent Help Text (2 hours)

**Files:** All `cmd/*.go`

**Template:**
```go
Long: `One-line description of what this does.

USE CASES:
  • When to use this command
  • What it tells you

EXAMPLES:
  # Basic usage
  r8s {{.UseLine}}

  # With options  
  r8s {{.UseLine}} --format=json

EXIT CODES:
  0 - Success
  1 - Issues found
  2 - Error
`,
```

**Impact:** Professional, consistent CLI experience.

### Win 4: Auto-Format Detection (3 hours)

**Files:** `cmd/root.go` or new helper

**Change:**
```go
func detectFormat(cmd *cobra.Command, flagValue string) string {
    if flagValue != "auto" {
        return flagValue
    }
    // If stdout is TTY, use table; else JSON
    if isatty.IsTerminal(os.Stdout.Fd()) {
        return "table"
    }
    return "json"
}
```

**Impact:** Pipe-friendly by default.

### Win 5: Validation for Unused Flags (2 hours)

**Files:** Command pre-run hooks

**Change:**
```go
PersistentPreRun: func(cmd *cobra.Command, args []string) {
    // Warn if --scan used with version
    if cmd.Name() == "version" && scanDepth != 200 {
        fmt.Fprintln(os.Stderr, "Warning: --scan flag has no effect with 'version' command")
    }
},
```

**Impact:** Prevents user confusion.

---

## Medium Effort Improvements

### Improvement 1: `r8s analyze` Umbrella Command (1 day)

**New command structure:**
```bash
r8s analyze health [path]      # Same as validate
r8s analyze patterns [path]    # Run pattern detection
r8s analyze summary [path]     # Quick summary
r8s analyze full [path]        # Everything (health + patterns + summary)
```

**Benefit:** Single entry point for all analysis.

### Improvement 2: Output Templating (2 days)

**New flag:**
```bash
r8s validate ~/bundle --template="{{.Completeness}}% complete"
```

**Benefit:** Users can customize output for scripts.

### Improvement 3: Config File Integration (1 day)

**Support in all commands:**
```bash
r8s validate ~/bundle --config=~/custom-config.yaml
# Uses config for defaults (scan depth, format preferences, etc.)
```

**Benefit:** Corporate environments with shared configs.

---

## Sprint 9 Proposal: CLI Polish Sprint

### Week 1: Audit + Quick Wins
- Day 1-2: UX review, document inconsistencies
- Day 3-4: Implement quick wins (1-5 above)
- Day 5: Testing, documentation

### Week 2: Medium Improvements  
- Day 6-7: Implement `r8s analyze` umbrella
- Day 8-9: Output templating
- Day 10: Final polish, release

### Deliverables
- [ ] CLI consistency audit document
- [ ] All commands use standard `--format`
- [ ] All commands have aliases
- [ ] Consistent help text
- [ ] Auto-format detection
- [ ] `r8s analyze` umbrella command
- [ ] Updated README with CLI guide

---

## Success Metrics

| Metric | Before | After |
|--------|--------|-------|
| Command consistency | 60% | 95% |
| Help text quality | Mixed | Excellent |
| User confusion reports | 5/mo | <1/mo |
| Script-friendliness | 70% | 95% |

---

## Why This Matters

**80/20 Principle:**
- 20% of CLI issues cause 80% of user confusion
- Quick wins (10 hours total) fix most problems
- Medium improvements (5 days) elevate to professional grade

**User Impact:**
- New users learn faster
- Script writers have consistent interface
- Support tickets decrease
- Demo quality improves

---

*Proposed for Sprint 9 or v0.8.0*
*Estimated effort: 2 weeks (1 dev + UX review)*
