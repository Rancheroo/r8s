# r8s CLI UX/Code Review Report

**Reviewer:** PIXEL (UX/CLI Lead)  
**Date:** 2026-02-17  
**Scope:** `cmd/` package — root.go, validate.go, generate.go, tui.go, config.go

---

## Executive Summary

The r8s CLI shows **strong UX foundations** with consistent Cobra patterns, helpful help text, and thoughtful output formatting. However, there are **notable inconsistencies** in error handling, flag naming conventions, and command structure that create UX debt. The codebase would benefit from a CLI style guide and refactoring to align all commands with the high standards set by `validate.go`.

**Overall Grade: B+** — Good practices in place, but inconsistencies need addressing before v1.0.

---

## 1. CLI Consistency Issues

### 1.1 Inconsistent Flag Naming Conventions

| Command | Flag | Issue |
|---------|------|-------|
| `validate` | `-f, --format` | Short flag for format |
| `generate prompt` | `-f, --format` | Same flag name, same purpose ✓ |
| `generate prompt` | `-o, --output` | Short flag for output |
| `root` | `-v, --verbose` | Short flag exists ✓ |
| `root` | `--scan` | **No short flag** — inconsistent with `-f`, `-v` |
| `root` | `--context` | **No short flag** — most tools use `-c` |
| `root` | `--config` | **No short flag** — standard is `-c` but conflicts with `--context` |

**Recommendation:**
- Add `-s` for `--scan` (lines to scan)
- Decide on `-c` priority: either `--config` or `--context` gets it
- Consider `-C` for the other, or use long-only for context

### 1.2 Bundle Path Argument Inconsistency

```go
// validate.go — requires positional arg
Args: cobra.ExactArgs(1)  // Must pass: r8s validate ./bundle

// generate prompt — requires positional arg  
Args: cobra.ExactArgs(1)  // Must pass: r8s generate prompt ./bundle

// tui.go — optional positional arg
// But also has: --bundle flag

// root.go — optional positional arg, delegates to TUI
Args: cobra.MaximumNArgs(1)
```

**Issue:** `tui` accepts both positional AND flag for bundle path. This creates ambiguity:
```bash
r8s tui ./bundle        # Works
r8s tui --bundle ./bundle  # Works
r8s ./bundle            # Works (via root.go)
```

**Recommendation:**
- **Pick one pattern** — positional args are more intuitive for this use case
- Remove `--bundle` flag from tuiCmd, rely on positional arg
- Ensure all commands use consistent bundle path handling

### 1.3 Output Format Inconsistency

| Command | JSON Output | Table Output | Summary Output |
|---------|-------------|--------------|----------------|
| `validate` | `--format=json` | `--format=table` (default) | `--summary` flag |
| `generate prompt` | N/A | N/A | N/A (generates text) |
| `config view` | No option | Text format only | No option |
| `version` | No option | Text only | No option |

**Issue:** `validate` uses `--format` enum, but `--summary` is a separate boolean. This is confusing:
- `--format=summary` doesn't exist (would expect it to)
- `--summary` overrides `--format` without warning

**Recommendation:**
```go
// Better pattern — make summary a format option
validateCmd.Flags().StringVarP(&validateFormat, "format", "f", "table", "Output format: table, json, summary")
// Remove separate --summary flag
```

### 1.4 Exit Code Documentation vs Implementation Gap

```go
// validate.go documents:
// EXIT CODES:
//   0 - Bundle is valid
//   1 - Bundle is incomplete but usable
//   2 - Invalid bundle or path

// But the code does:
if !health.IsValid {
    os.Exit(2)
} else if health.Completeness < 100 {
    os.Exit(1)
}
```

**Issue:** No `os.Exit(0)` explicitly shown (relies on return nil). Also, error path calls `os.Exit(2)` twice (once explicitly, once via error return).

**Recommendation:**
```go
// Cleaner pattern
switch {
case err != nil:
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(2)
case !health.IsValid:
    os.Exit(2)
case health.Completeness < 100:
    os.Exit(1)
default:
    os.Exit(0)
}
```

---

## 2. Error Message Improvements

### 2.1 Error Prefix Inconsistency

```go
// validate.go — uses "Error:" prefix with Fprintf to stderr
fmt.Fprintf(os.Stderr, "Error: %v\n", err)
os.Exit(2)

// generate.go — wraps errors with context, prints to stdout via fmt.Println
return fmt.Errorf("cannot access bundle path: %w", err)
// (goes through cobra, which prints to stderr, but no "Error:" prefix)

// tui.go — wraps errors
return fmt.Errorf("failed to load config: %w", err)
```

**Recommendation:** Standardize error presentation:
```go
// cmd/helpers.go — shared error handling
func PrintError(msg string) {
    fmt.Fprintf(os.Stderr, "r8s: error: %s\n", msg)
}

func PrintErrorf(format string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, "r8s: error: "+format+"\n", args...)
}
```

### 2.2 User-Actionable Error Messages

**Current (generate.go):**
```go
return fmt.Errorf("cannot access bundle path: %w", err)
// Output: "cannot access bundle path: stat /bad/path: no such file or directory"
```

**Better:**
```go
return fmt.Errorf("bundle not found: %q\n\nEnsure the path exists and you have read permissions.\nExample: r8s generate prompt ./my-bundle/", bundlePath)
```

### 2.3 Silent Failures

```go
// tui.go
if err := cfg.Save(""); err != nil {
    // Non-fatal: Help hint will still work, just won't persist count
    if cfg.Verbose {
        fmt.Printf("Warning: Failed to save launch count: %v\n", err)
    }
}
```

**Issue:** Silent failure unless verbose mode. User never knows config isn't persisting.

**Recommendation:** Always inform user of non-fatal issues:
```go
if err := cfg.Save(""); err != nil {
    fmt.Fprintf(os.Stderr, "Warning: could not save preferences: %v\n", err)
}
```

---

## 3. Flag Naming Recommendations

### 3.1 Global Flags Review

| Flag | Current | Recommended | Rationale |
|------|---------|-------------|-----------|
| `--scan` | Lines to scan | `--scan-depth` or `--lines` | More descriptive |
| `--verbose` | `-v` | Keep as-is | Standard convention ✓ |
| `--context` | Context name | Keep as-is | Matches kubectl ✓ |
| `--namespace` | `-n` | Keep as-is | Matches kubectl ✓ |

### 3.2 `generate prompt` Flags

| Flag | Issue | Recommendation |
|------|-------|----------------|
| `--severity` | Only works for `all`, `critical`, `warning` | Document better or add `error`, `info` levels |
| `--format=script` | Name implies executable script | Rename to `--format=bash-prompt` or clarify in help |

### 3.3 Missing Standard Flags

**Consider adding:**
- `--no-color` / `--color=false` — for CI/scripting environments
- `--quiet` / `-q` — suppress non-error output
- `--output` / `-o` — consistent across all commands (validate lacks this)

---

## 4. Help Text Assessment

### 4.1 Strong Examples ✓

**validate.go Long description:**
```go
Long: `Check if a support bundle is complete and usable.

This command analyzes the bundle structure and reports:
  • Completeness percentage (what's present vs expected)
  • Missing files with impact scoring
  ...`,
```
- Clear bulleted list
- Concrete examples
- Exit codes documented

**root.go Long description:**
- Feature list with bullets
- Quickstart section
- Multiple examples

### 4.2 Weak Examples ✗

**config.go root command:**
```go
var configCmd = &cobra.Command{
    Use:   "config",
    Short: "Manage r8s configuration",
    Long:  "Initialize, view, or edit r8s configuration file",
    // ... just prints subcommand list
}
```
- No examples
- No mention of config file location
- Subcommands not shown in help (prints custom text instead of letting cobra handle it)

**version command:**
```go
var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print version information",
    // No Long description
}
```
- Missing examples (e.g., `r8s version --short` if that existed)

### 4.3 Cobra Help Template Customization

**Missing:** The CLI doesn't customize cobra's help template. Consider:
- Adding "See also" sections
- Grouping commands (Analysis: validate, generate; Interface: tui; Config: config)
- Adding footer with docs URL

---

## 5. Command Structure Issues

### 5.1 `generate` Command Hierarchy

```
r8s generate prompt    # Currently only subcommand
```

**Issue:** `generate` has only one subcommand. Is this premature abstraction?

**Options:**
1. **Keep hierarchy** — if planning `generate script`, `generate report` in future
2. **Flatten** — `r8s prompt` instead of `r8s generate prompt`

**Recommendation:** Keep hierarchy IF there are concrete plans for more generators. Otherwise, flatten to `r8s prompt` for v1.0 and re-introduce nesting later.

### 5.2 `config` Command Completeness

**config set** supports: `url`, `token`, `insecure`, `currentProfile`

**Missing common operations:**
- No `config get` — can't read a single value
- No `config unset` — can't remove values
- No `config delete-profile` — requires manual editing

**Example gap:**
```bash
r8s config get url          # Not possible — must use view and grep
r8s config unset token      # Not possible
r8s config delete-profile staging  # Not possible
```

### 5.3 Command Naming

| Command | Verb | Noun | Pattern |
|---------|------|------|---------|
| `validate` | ✓ | bundle (implied) | verb-only |
| `generate` | ✓ | prompt (subcmd) | verb + noun |
| `config` | ✓ | — | noun (management) |
| `tui` | ✗ | — | acronym |
| `version` | ✓ | — | noun |

**Issue:** `tui` is jargon. Non-technical users won't know what it means.

**Recommendation:** Consider alias:
```go
var tuiCmd = &cobra.Command{
    Use:     "tui",
    Aliases: []string{"browse", "explore", "interactive"},
    // ...
}
```

---

## 6. Datasource Interface Usage

### 6.1 Commands Don't Use DataSource

**Finding:** Neither `validate` nor `generate` commands use the `datasource.DataSource` interface!

```go
// validate.go — uses bundle package directly
health, err := bundle.CheckHealth(bundlePath)

// generate.go — uses bundle package directly
health, err := bundle.CheckHealth(bundlePath)

// Only tui.go uses datasource
cfg, err := config.Load(cfgFile, "")
app := tui.NewApp(cfg, tuiBundlePath)  // Creates internal DataSource
```

**Issue:** The datasource abstraction exists but is only used by TUI. This creates:
1. **Code duplication** — bundle health checks implemented twice
2. **Inconsistency** — CLI and TUI may behave differently
3. **Missed abstraction** — DataSource.BundleHealth has Percentage() method not used

### 6.2 Recommendation: Unify Through DataSource

```go
// cmd/helpers.go
func OpenBundle(path string) (datasource.DataSource, error) {
    // Return BundleDataSource or error
}

// Then in validate:
func runValidate(cmd *cobra.Command, args []string) error {
    ds, err := OpenBundle(bundlePath)
    if err != nil {
        return err
    }
    defer ds.Close()
    
    health := ds.GetBundleHealth()
    // ... use health.Percentage(), etc.
}
```

---

## 7. UX Debt Summary (Priority Order)

### 🔴 High Priority (Fix Before v1.0)

1. **Standardize error handling** — Create shared error formatting helpers
2. **Fix flag inconsistency** — Add short flags (`-s` for scan) or remove from others
3. **Bundle path ambiguity** — Remove `--bundle` flag, use positional only
4. **DataSource unification** — Refactor validate/generate to use datasource interface

### 🟡 Medium Priority (Fix in v1.x)

5. **Add `--no-color` flag** — Required for CI/scripting
6. **Expand config commands** — Add `get`, `unset`, `delete-profile`
7. **Improve error messages** — Make all errors actionable
8. **Help template customization** — Group commands, add "See also"

### 🟢 Low Priority (Nice to Have)

9. **Flatten `generate prompt`** — Or commit to multi-subcommand structure
10. **Add `tui` aliases** — `browse`, `explore`, `interactive`
11. **JSON output for all commands** — `config view --format=json`, `version --format=json`
12. **Progress indicators** — For long operations (bundle analysis)

---

## 8. Positive Patterns to Preserve

### ✅ Excellent Help Text
- `validate.go` and `root.go` set the standard
- Bulleted features, clear examples, exit codes

### ✅ Consistent Cobra Patterns
- All commands use `RunE` for error handling
- Proper `Args` validation (`ExactArgs`, `MaximumNArgs`)
- Persistent flags on root for global options

### ✅ Thoughtful Output Formatting
- `validate` table output uses color and icons effectively
- JSON output properly indented with `encoder.SetIndent`
- Summary mode for quick scanning

### ✅ Version Command
- Shows version, commit, and build date — all essential info

---

## Appendix: Quick Reference

### Current Command Structure
```
r8s [bundle-path]              # Launch TUI with bundle (root command)
r8s validate <bundle-path>     # Validate bundle health
r8s generate prompt <bundle-path>  # Generate AI prompt
r8s tui [bundle-path]          # Launch TUI (explicit)
r8s config [init|view|edit|validate|set]
r8s version                    # Show version
```

### Flag Inventory
```
Global (root):
  --config, -v/--verbose, --context, -n/--namespace, --scan

validate:
  -f/--format=[table|json|summary], --summary

generate prompt:
  -f/--format=[chatbot|terminal|script], -o/--output, -s/--severity

tui:
  --bundle (REDUNDANT with positional arg)

config:
  (uses global --config, --profile on some subcommands)
```

---

*Report compiled by PIXEL — r8s UX/CLI Lead*
