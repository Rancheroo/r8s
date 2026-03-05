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

*Assume the following environment variables are set for the demo:*

```bash
export r8sbundle1="./support-bundle-2026-03-01" # A bundle with a critical failure
export r8sbundle2="./support-bundle-2026-03-02" # A bundle for deep-dive investigation
```

---

## 3. Scenario 1: The "Everything is Broken" Ticket (10 mins)

*Context: A customer uploads a bundle saying "Production is down". You have 5 minutes to find the root cause.*

### Step 1: Validate the Artifact
First, check if the bundle is even usable.

```bash
r8s validate $r8sbundle1
```

*   **Talking Point:** "How many times have you debugged a partial bundle for an hour only to realize logs are missing? `r8s` tells you immediately."

### Step 2: Instant Analysis
Run the AI analyzer to find smoking guns.

```bash
r8s analyze $r8sbundle1
```

*   **Talking Point:** "r8s scans for 19+ specific patterns like Etcd Quorum Loss, CNI failures, and OOMKilled pods. It highlights CRITICAL issues in red."

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
r8s describe node $r8sbundle2 worker-2
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

**Q: Where do I get it?**
**A:** It's open source! `github.com/Rancheroo/r8s`.
