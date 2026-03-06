# r8s Demo & Speaker Notes

**Audience:** SUSE Rancher Support Engineers
**Goal:** Show how `r8s` accelerates root cause analysis and integrates into existing workflows.
**Duration:** 30 min Demo + 30 min Q&A

---

## 1. Introduction (2 mins)

"r8s (rates) is an AI-powered, offline-first tool that brings `kubectl` superpowers to static support bundles. It doesn't just grep logs; it understands Kubernetes relationships and detects common failure patterns automatically."

**Key Selling Points for Support:**
*   **Offline First:** No cluster access needed. Works on air-gapped machines.
*   **Zero Data Leaks:** All "AI" analysis happens locally on your laptop. No data is sent to OpenAI/Cloud.
*   **kubectl Muscle Memory:** Use commands you already know (`get`, `logs`, `describe`).

---

## 2. Setup

Run the demo generator to create rich, realistic bundles in `/tmp`.

```bash
# Generate data and set variables
eval $(/tmp/generate-demo-bundle.sh env)

# Verify
echo $r8sbundle1
# Output: /tmp/r8s-demo-critical
```

---

## 3. Scenario 1: The "Everything is Broken" Ticket (10 mins)

*Context: A customer uploads a bundle saying "Production is down". You have 5 minutes to find the root cause.*

### Step 1: Instant Analysis & Validation
Run the AI analyzer. It performs an automatic health check before scanning.

```bash
r8s analyze $r8sbundle1
```

*   **Talking Point:** "Notice the first step: `✓ Bundle validated`. r8s instantly checks if the bundle is complete and healthy before wasting your time scanning it. It then scans for 19+ specific patterns..."

### Step 3: Natural Language Query
Instead of grepping, just ask.

```bash
r8s ask $r8sbundle1 "what is the main issue?"
r8s ask $r8sbundle1 "show me all expired certificates"
```

*   **Talking Point:** "This uses a local NLP engine to map your question to the bundle data. It's faster than constructing complex `find | grep` chains."

---

## 4. Scenario 2: Deep Dive Investigation (10 mins)

*Context: You know the issue is in the `cattle-system` namespace, but need details.*

### Step 1: Familiar Exploration
Use standard kubectl commands on the static bundle. No more manual directory navigation.

```bash
# List all namespaces
r8s get namespaces $r8sbundle2

# List nodes with details
r8s get nodes $r8sbundle2

# Get pods in a specific namespace
r8s get pods $r8sbundle2 -n cattle-system

# See IP addresses and Nodes with -o wide
r8s get pods $r8sbundle2 -n cattle-system -o wide

# Describe a node to check capacity and conditions
r8s describe node $r8sbundle2 r8s-wk-jnhwv-4xqzn
```

*   **Talking Point:** "It supports standard flags like `-n`, `-A` (all namespaces), and `-o wide`. You don't need to learn a new syntax."

### Step 2: Logs without Unzipping
Don't navigate the directory structure manually.

```bash
# Find the crashing pod
r8s get pods $r8sbundle2 | grep CrashLoop

# Read its logs
r8s logs $r8sbundle2 rancher-webhook-5d9b7
```

*   **Talking Point:** "r8s automatically finds the right log file for the pod, handling the directory structure differences between RKE2, K3s, and Rancher versions."

### Step 3: Describe
Get the full context (Events, State, Conditions).

```bash
r8s describe pod $r8sbundle2 rancher-webhook-5d9b7
```

---

## 5. Scenario 3: Automation & Tooling (8 mins)

*Context: The engineers in the audience build their own tools or want to automate triage.*

### AI-Ready Diagnostics
Want to use ChatGPT/Claude/OpenCode to analyze the bundle? Don't copy-paste 50 log files.

```bash
# Generate a single, context-rich prompt for an LLM
r8s generate prompt $r8sbundle2 > analysis-context.md

# Pipe directly into OpenCode
r8s generate prompt $r8sbundle2 | opencode run
```

*   **Talking Point:** "This command packages the most critical parts of the bundle (node status, failing pods, recent logs) into a single Markdown file optimized for LLM context windows. You can pipe this straight into `opencode` or `claude` to start an interactive debugging session with full context."

### JSON Output for Scripting
Every command supports `--format=json`.

```bash
# Extract only critical issues for a dashboard
r8s analyze $r8sbundle2 --format=json | jq '.issues[] | select(.severity=="critical")'
```

### CI/CD Integration
*   Show the GitHub Actions example from the README.
*   Explain how this can be used in `suse/support-tools` to auto-triage incoming tickets.

---

## 6. Q&A Preparation (Anticipated Questions)

**Q: Does `r8s ask` send customer data to an external AI API?**
**A:** **NO.** The NLP engine is 100% local. It uses keyword matching and heuristics (Go code) to query the bundle. Your customer's data never leaves your machine.

**Q: Does it work with RKE1?**
**A:** Primarily designed for RKE2 and K3s bundles (which follow the standard layout). RKE1 support is limited to generic Kubernetes resources found in `k8s/` folders.

**Q: Can I add my own patterns?**
**A:** Yes! The patterns are defined in Go. We welcome PRs to `internal/ai/patterns.go` to add new detection logic for issues you see frequently.

**Q: How does it handle huge bundles (10GB+)?**
**A:** `r8s` streams logs and uses efficient parsing. It doesn't load the whole bundle into RAM. However, `grep` might still be faster for *massive* raw log searches, but `r8s` is faster for structured data.

**Q: Can I integrate this into Slack or Jira?**
**A:** Yes. Since `r8s` outputs structured JSON (`--format=json`), you can easily wrap it in a script (like a GitHub Action or a simple webhook bot) to post analysis results directly to a Slack channel or Jira ticket when a bundle is uploaded.

**Q: Does it support custom patterns?**
**A:** Currently, patterns are compiled into the binary for performance. However, we are planning a feature to load custom YAML patterns from a `~/.r8s/patterns/` directory in a future release.

**Q: Is it safe to run on my laptop with customer data?**
**A:** Yes. `r8s` runs entirely locally. It does not upload any data to the cloud. Even the AI features (`r8s ask` and `r8s generate prompt`) are local algorithms or local formatting. If you pipe output to `opencode` or `claude`, *you* control that data flow explicitly.

**Q: Does it work with k3s bundles?**
**A:** Yes, it fully supports both RKE2 and K3s bundles produced by the standard Rancher support tool. It auto-detects the format.

**Q: Can I filter logs by container in a multi-container pod?**
**A:** Yes. Use `r8s logs ./bundle/ pod-name -c container-name` just like kubectl.

**Q: Does it show previous (crashed) logs?**
**A:** Yes. If a pod has restarted, `r8s` automatically detects the `-previous.log` files and can display them. It often prioritizes them for crash analysis.

**Q: How do I know if the bundle is complete?**
**A:** Run `r8s validate ./bundle/`. It checks for core files (nodes, pods, events) and gives you a completeness score.

**Q: Can I use label selectors?**
**A:** Basic filtering is supported (e.g. by namespace). Full label selector support (`-l app=nginx`) is on the roadmap for v1.4.

**Q: Does it check for certificates?**
**A:** Yes. It has specific patterns to detect "x509: certificate has expired" errors in logs and events.

**Q: Can I export the report to PDF?**
**A:** Not directly, but you can export to Markdown (`--format=markdown`) and then use any Markdown-to-PDF converter (like Pandoc or VS Code).

**Q: Does it analyze system logs (journald)?**
**A:** Yes. It scans `systemlogs/` and `journald/` directories for system-level issues like OOM kills, disk pressure, and CNI failures.

**Q: What if the bundle is just a `kubectl cluster-info dump`?**
**A:** It has partial support for generic kubectl dumps, but it works best with the structured Rancher support bundle format.

**Q: Can I contribute a new pattern?**
**A:** Absolutely. Open a PR adding a new YAML file to `internal/ai/patterns/`. It's very easy to add regex-based detection rules.

**Q: How often is it updated?**
**A:** We aim for a release every sprint (2 weeks). Since it's a CLI tool, you can update it instantly with `curl`.

**Q: Does it require Docker?**
**A:** No. It is a single static Go binary. No dependencies required.

**Q: Can I use it on Windows?**
**A:** Yes, we build binaries for Linux, macOS, and Windows. It works in PowerShell or WSL.

**Q: Does it handle Windows containers/nodes?**
**A:** It parses Windows node logs if they are captured in the standard bundle format, but some Linux-specific checks (like dmesg) won't apply.

**Q: Can I diff two bundles?**
**A:** You can run `r8s analyze --format=json` on both and use `diff` or `jd` (JSON diff) to compare the outputs programmatically.

**Q: Where do I get it?**
**A:** It's open source! `github.com/Rancheroo/r8s`.
