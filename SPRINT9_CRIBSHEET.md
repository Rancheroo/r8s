# 📋 Sprint 9 Crib Sheet — Code Sprint Channel

**Sprint:** 9  
**Dates:** Feb 19 - Mar 04, 2026  
**Target:** v0.8.0-alpha  
**Branch:** `feature/sprint9-cli-polish` ⬅️ USE THIS  

---

## 🎯 THE GOAL

Make r8s CLI feel like `kubectl`. Delete the TUI. Ship v0.8.0-alpha.

**Musk's Laws:**
1. Delete - Legacy TUI views gone
2. Simplify - One binary, standard flags
3. Accelerate - Fast shell completion

---

## 📅 TWO WEEKS AT A GLANCE

### WEEK 1: CLI MATURITY (Feb 19-26)

| Day | Task | Priority | Output |
|-----|------|----------|--------|
| **Day 1** | `r8s completion` | P0 | bash/zsh/fish completion |
| **Day 1** | Create sprint branch | P0 | `feature/sprint9-cli-polish` |
| **Day 2** | `r8s logs [pod] -n [ns]` | P0 | Stream logs from bundle |
| **Day 3** | `r8s describe [kind] [name]` | P1 | Resource details (kubectl style) |
| **Day 4** | Output formats | P1 | `--output=json|yaml|wide` on ALL commands |
| **Day 5** | Testing & polish | P1 | All commands have tests |
| **Day 6-7** | Buffer / Issues | - | Fix bugs, prep for Week 2 |

### WEEK 2: TUI SUNSET (Feb 26 - Mar 4)

| Day | Task | Priority | Output |
|-----|------|----------|--------|
| **Day 8** | `r8s dashboard` command | P0 | Lightweight dashboard-only TUI |
| **Day 9** | Delete legacy TUI | P1 | Remove Clusters/Projects views |
| **Day 10** | Release automation setup | P1 | Auto-binaries on GitHub releases |
| **Day 11** | Man pages | P1 | All commands documented |
| **Day 12** | Documentation updates | P2 | README with CLI-first tutorials |
| **Day 13** | Final testing | P0 | Demo script, user verification |
| **Day 14** | Release v0.8.0-alpha | P0 | Tag, release, celebrate |

---

## 📋 DELIVERABLES

### Commands to Build

```bash
# Week 1 targets
r8s completion bash > /etc/bash_completion.d/r8s
r8s logs ./bundle nginx-pod -n cattle-system
r8s describe ./bundle pod nginx-pod
r8s analyze ./bundle --output=json|yaml|wide

# Week 2 targets
r8s dashboard ./bundle           # Optional lightweight TUI
r8s --help                       # Comprehensive help
man r8s                          # Man page
```

### Code to Delete

```bash
# Day 9: Delete these (confirmed safe to remove)
internal/tui/clusters.go
internal/tui/projects.go
internal/tui/namespaces.go
internal/tui/cluster.go
internal/tui/project.go
internal/tui/namespace.go
```

### Documentation to Update

- README.md — CLI-first quickstart
- docs/development/CHANGELOG.md — v0.8.0-alpha notes
- Man pages for all commands

---

## 🛠️ TECHNICAL GUIDANCE

### Branch Setup
```bash
cd /workspace/r8s
git checkout -b feature/sprint9-cli-polish
git push -u origin feature/sprint9-cli-polish
```

### Completion Implementation
Use Cobra's built-in:
```go
// cmd/completion.go
var completionCmd = &cobra.Command{
    Use:   "completion [bash|zsh|fish]",
    Short: "Generate shell completion script",
    RunE: func(cmd *cobra.Command, args []string) error {
        switch args[0] {
        case "bash":
            return rootCmd.GenBashCompletion(os.Stdout)
        case "zsh":
            return rootCmd.GenZshCompletion(os.Stdout)
        case "fish":
            return rootCmd.GenFishCompletion(os.Stdout, true)
        }
        return nil
    },
}
```

### Logs Command Pattern
```bash
r8s logs [bundle-path] [pod-name] [flags]

Flags:
  -n, --namespace string   Filter by namespace
      --tail int          Show last N lines (default 100)
      --previous          Show previous container logs
      --follow            Stream logs (if bundle supports it)
```

### Output Format Standard
All commands must support:
```bash
--output=json     # Machine readable
--output=yaml     # Human readable structured
--output=wide     # Extra columns (like kubectl)
--output=table    # Default, human friendly
```

---

## ✅ DEFINITION OF DONE

For each feature:
- [ ] Code committed
- [ ] Pushed to `origin/feature/sprint9-cli-polish`
- [ ] Tests pass (`make test`)
- [ ] User Verified ✓ (human tested)
- [ ] Documentation updated

For Sprint 9:
- [ ] All P0 features complete
- [ ] TUI code reduced >50%
- [ ] `r8s logs` starts in <500ms
- [ ] Shell completion works on tab
- [ ] v0.8.0-alpha released on GitHub

---

## 📊 SUCCESS METRICS

| Metric | Target | Measurement |
|--------|--------|-------------|
| CLI commands | 8+ | Count in `r8s --help` |
| TUI code | <4,500 lines | `wc -l internal/tui/*.go` |
| Logs latency | <500ms | `time r8s logs ./bundle pod` |
| Completion | Works | Tab completion in bash/zsh |
| CI/CD ready | Yes | Exit codes meaningful, JSON output |

---

## 🚨 RELEASE AUTOMATION SETUP

**Week 2, Day 10 task:**

```yaml
# .github/workflows/release.yml (to create)
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - name: Build binaries
        run: |
          GOOS=linux GOARCH=amd64 go build -o bin/r8s-linux-amd64
          GOOS=darwin GOARCH=amd64 go build -o bin/r8s-darwin-amd64
          GOOS=darwin GOARCH=arm64 go build -o bin/r8s-darwin-arm64
      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: bin/*
          body: "$(cat CHANGELOG.md | head -50)"
```

**Assign to:** DevOps/Release Manager persona
**Due:** Day 10 (Feb 28)

---

## 🔗 RESOURCES

- **Sprint 9 Kickoff Doc:** `docs/development/SPRINT9_KICKOFF.md`
- **Vision v1.0:** `docs/development/R8S_VISION_v1.0.md`
- **Killer Feature:** `docs/development/V1.0_KILLER_FEATURE.md`
- **Pivot Plan:** `docs/development/PIVOT_v0.8.0_PURE_CLI.md`

---

## 🎯 DAILY PROCESS

**Start of Day:**
1. Pull latest: `git pull origin feature/sprint9-cli-polish`
2. Implement feature
3. Commit: `git commit -m "feat: r8s logs command"`
4. Push: `git push origin feature/sprint9-cli-polish`
5. Verify: `git log origin/feature/sprint9-cli-polish --oneline -3`
6. Message: "[Feature] ready, pull and test"

**End of Day:**
1. Human tests feature
2. Feedback in thread
3. User Verified ✓ or fixes issued

---

## 🚀 READY?

Let's build the `kubectl` for bundles.  
**For Elon. For the CLI. For v0.8.0-alpha.**
