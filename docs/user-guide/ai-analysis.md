# AI Analysis Guide

r8s v0.9.0 introduces powerful AI-driven analysis capabilities. This guide explains how the pattern engine works, how to use natural language queries, and how to integrate with CI/CD.

## 🧠 Pattern Engine

r8s scans support bundles for known issue patterns. It doesn't just grep for errors; it understands the context.

### Detection Capabilities
- **Pod Lifecycle**: CrashLoopBackOff, OOMKilled, ImagePullBackOff, Pending, Terminating
- **Node Health**: NotReady, DiskPressure, MemoryPressure, PIDPressure
- **Networking**: DNS resolution failures, CNI plugin errors, Connectivity timeouts
- **Storage**: PVC binding failures, Provisioning errors
- **Control Plane**: etcd latency/corruption/quorum, Leader election failures
- **Security**: Certificate expiration, Invalid CA trust

### How it Works
1. **Scanning**: Reads `kubectl` output, pod logs, events, and system journals.
2. **Matching**: Uses regex and keyword matching with confidence scoring.
3. **Correlation**: Links related issues (e.g., *Pod Pending* caused by *PVC Binding Failure*).
4. **Hint Generation**: Produces a root cause explanation and suggested fix.

### Usage
```bash
# Analyze a bundle
r8s analyze ./support-bundle/

# Filter by severity
r8s analyze ./support-bundle/ --severity=critical
```

---

## 🗣️ Natural Language Queries (NLQ)

You can ask r8s questions in plain English. The NLQ engine parses your intent and queries the analysis results.

### Query Patterns

| Intent | Example Question | What it does |
|--------|------------------|--------------|
| **Why** | "Why is nginx crashing?" | Explains root cause (OOM, config, etc.) |
| **Show** | "Show me certificate issues" | Lists all findings of a specific type |
| **Which** | "Which nodes are not ready?" | Lists resources in a specific state (including "running" and "ready") |
| **What** | "What is wrong with etcd?" | General health check for a component |

### Examples
```bash
$ r8s ask ./bundle/ "which pods are running?"

🎯 Found 8 pods that are ready:
• coredns-854c77959c-7x9j6
• metrics-server-86cbb8457f-m8j2w
...
```

```bash
$ r8s ask ./bundle/ "why is web-pod crashing?"

🔍 Analysis: Why resources are crashing

**Finding 1:** Pod web-pod is in CrashLoopBackOff. Restarts: 5.

**What happened:**
Container exited with code 137 (OOMKilled).

**What to do:**
Increase memory limit in pod spec.

**Try this command:**
kubectl describe pod web-pod
```

---

## 📤 CI/CD Integration (Export)

r8s supports standard formats for integration with security scanners and CI pipelines.

### GitHub Advanced Security (SARIF)
Upload findings to GitHub's Security tab.

```bash
# Generate SARIF report
r8s export ./bundle/ --format=sarif --output=r8s.sarif
```

**GitHub Action Example:**
```yaml
steps:
  - name: Analyze Bundle
    run: ./r8s export ./bundle/ --format=sarif --output=r8s.sarif
    
  - name: Upload SARIF
    uses: github/codeql-action/upload-sarif@v2
    with:
      sarif_file: r8s.sarif
```

### Jenkins / JUnit
Report test failures in Jenkins or GitLab CI.

```bash
# Generate JUnit XML
r8s export ./bundle/ --format=junit --output=results.xml
```

### Documentation (Markdown)
Generate a human-readable report for ticket attachments.

```bash
r8s export ./bundle/ --format=markdown --output=analysis.md
```

---

## 🛠️ Pattern Registry

You can inspect the built-in patterns to understand what r8s detects.

```bash
# List all patterns
r8s patterns list

# Show details for a pattern
r8s patterns show etcd-latency
```

### Custom Patterns
(Coming in v0.9.5) Support for user-defined YAML patterns.
