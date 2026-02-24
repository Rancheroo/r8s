# r8s v0.8.0 Manual Testing Plan

**Date:** 2026-02-18  
**Branch:** `feature/v0.8.0-pure-cli`  
**Tester:** DontStop

---

## Prerequisites

```bash
# Pull the branch
git fetch origin
git checkout feature/v0.8.0-pure-cli
git pull origin feature/v0.8.0-pure-cli

# Build
cd /home/node/.openclaw/workspace/r8s
make build
# OR: go build .
```

**Required:** An extracted RKE2 support bundle folder
- Location: `~/bundle/` or similar
- Structure: Contains `metadata.json`, `podlogs/`, `k8s/` directories

---

## Phase 1: Build & Basic Verification

| # | Test | Command | Expected Result | Pass/Fail |
|---|------|---------|-----------------|-----------|
| 1 | Binary builds | `go build .` | No errors, `r8s` binary created | |
| 2 | Version check | `./r8s version` | Shows version, commit, date | |
| 3 | Help works | `./r8s --help` | Shows CLI-first help text | |
| 4 | No TUI refs | `grep -r "bubbletea\|lipgloss" cmd/` | No matches | |
| 5 | Binary size | `ls -lh r8s` | < 8MB (was ~11MB) | |

---

## Phase 2: CLI Entry Point (Default Command)

| # | Test | Command | Expected Result | Pass/Fail |
|---|------|---------|-----------------|-----------|
| 6 | No args shows help | `./r8s` | Shows help (not TUI error) | |
| 7 | Bundle path triggers analyze | `./r8s ./bundle/` | Runs analyze command | |
| 8 | Explicit analyze works | `./r8s analyze ./bundle/` | Same as #7 | |
| 9 | Help mentions CLI-first | `./r8s --help` | "kubectl for Rancher bundles" in text | |

---

## Phase 3: Core kubectl Commands

### 3a: `r8s get` Tests

| # | Test | Command | Expected Result | Pass/Fail |
|---|------|---------|-----------------|-----------|
| 10 | Get pods table | `./r8s get pods ./bundle/` | Table with NAMESPACE, NAME, READY, STATUS | |
| 11 | Get pods wide | `./r8s get pods ./bundle/ -o wide` | Includes NODE column | |
| 12 | Get pods JSON | `./r8s get pods ./bundle/ -o json` | Valid JSON array | |
| 13 | Get pods names only | `./r8s get pods ./bundle/ -o name` | One pod name per line | |
| 14 | Get with namespace filter | `./r8s get pods ./bundle/ -n cattle-system` | Only cattle-system pods | |
| 15 | Get all namespaces | `./r8s get pods ./bundle/ -A` | Shows NAMESPACE column | |
| 16 | Get nodes | `./r8s get nodes ./bundle/` | Shows node name, status, version | |
| 17 | Get namespaces | `./r8s get ns ./bundle/` | Shows namespace list | |
| 18 | Shorthand works | `./r8s get po ./bundle/` | Same as "pods" | |
| 19 | JSON pipes to jq | `./r8s get pods ./bundle/ -o json \| jq '.[0].name'` | Extracts first pod name | |

### 3b: `r8s logs` Tests

| # | Test | Command | Expected Result | Pass/Fail |
|---|------|---------|-----------------|-----------|
| 20 | Logs basic | `./r8s logs ./bundle/ <pod-name>` | Shows log lines | |
| 21 | Logs with container | `./r8s logs ./bundle/ <pod> -c <container>` | Specific container logs | |
| 22 | Logs previous | `./r8s logs ./bundle/ <pod> -p` | Previous container logs (if exist) | |
| 23 | Logs tail | `./r8s logs ./bundle/ <pod> --tail=50` | Last 50 lines only | |
| 24 | Partial name match | `./r8s logs ./bundle/ nginx` | Matches nginx-* pod | |
| 25 | Ambiguous name errors | `./r8s logs ./bundle/ app` | Lists multiple matches, asks for full name | |
| 26 | Missing pod errors | `./r8s logs ./bundle/ nonexistent` | "pod 'nonexistent' not found", exit 1 | |
| 27 | Verbose header | `./r8s logs ./bundle/ <pod> -v` | Shows "# Logs for..." header | |

### 3c: `r8s describe` Tests

| # | Test | Command | Expected Result | Pass/Fail |
|---|------|---------|-----------------|-----------|
| 28 | Describe pod text | `./r8s describe pod ./bundle/ <pod-name>` | Name, Namespace, Status, Containers sections | |
| 29 | Describe pod JSON | `./r8s describe pod ./bundle/ <pod> -o json` | JSON structure | |
| 30 | Describe node | `./r8s describe node ./bundle/` | Node name, RKE2 version, K8s version | |
| 31 | Describe namespace | `./r8s describe ns ./bundle/ cattle-system` | Namespace name, pod count | |
| 32 | Shorthand works | `./r8s describe po ./bundle/ <pod>` | Same as "pod" | |
| 33 | Partial name match | `./r8s describe pod ./bundle/ nginx` | Matches partial name | |

---

## Phase 4: Bundle Analysis Commands

| # | Test | Command | Expected Result | Pass/Fail |
|---|------|---------|-----------------|-----------|
| 34 | Analyze table | `./r8s analyze ./bundle/` | Health percentage, issues list | |
| 35 | Analyze JSON | `./r8s analyze ./bundle/ -f json` | Valid JSON with bundle_path, issues array | |
| 36 | Analyze critical only | `./r8s analyze ./bundle/ -s critical` | Only critical issues | |
| 37 | Validate table | `./r8s validate ./bundle/` | Bundle Health Check table | |
| 38 | Validate JSON | `./r8s validate ./bundle/ -f json` | Health check JSON | |
| 39 | Validate summary | `./r8s validate ./bundle/ --summary` | One-line summary | |
| 40 | Generate prompt | `./r8s generate prompt ./bundle/` | AI prompt output | |
| 41 | Generate terminal format | `./r8s generate prompt ./bundle/ -f terminal` | Command-focused prompt | |

---

## Phase 5: Error Handling & Edge Cases

| # | Test | Command | Expected Result | Pass/Fail |
|---|------|---------|-----------------|-----------|
| 42 | Invalid bundle path | `./r8s analyze /nonexistent/` | Error message, exit 2 | |
| 43 | Unknown resource type | `./r8s get secrets ./bundle/` | "unknown resource type: secrets" | |
| 44 | Missing pod in describe | `./r8s describe pod ./bundle/ fake-pod` | "pod 'fake-pod' not found", exit 1 | |
| 45 | Missing bundle path | `./r8s get pods` | "bundle path required" | |
| 46 | Wrong output format | `./r8s get pods ./bundle/ -o xml` | Unknown format error | |
| 47 | Exit code 0 on success | `./r8s validate ./bundle/; echo $?` | 0 if valid | |
| 48 | Exit code 1 on issues | `./r8s analyze ./bundle/; echo $?` | 1 if critical issues found | |

---

## Phase 6: kubectl UX Parity

| # | Test | kubectl | r8s | Pass/Fail |
|---|------|---------|-----|-----------|
| 49 | Same get syntax | `kubectl get pods -n ns` | `r8s get pods ./bundle/ -n ns` | |
| 50 | Same logs syntax | `kubectl logs pod -c c` | `r8s logs ./bundle/ pod -c c` | |
| 51 | Same describe syntax | `kubectl describe pod name` | `r8s describe pod ./bundle/ name` | |
| 52 | Same -o json flag | `kubectl ... -o json` | `r8s ... -o json` | |
| 53 | Same -n namespace flag | `kubectl ... -n kube-system` | `r8s ... -n kube-system` | |

---

## Phase 7: Performance Check

| # | Test | Command | Expected | Pass/Fail |
|---|------|---------|----------|-----------|
| 54 | Fast startup | `time ./r8s get pods ./bundle/` | < 2 seconds | |
| 55 | Fast logs | `time ./r8s logs ./bundle/ <pod>` | < 1 second | |
| 56 | JSON parse speed | `time ./r8s analyze ./bundle/ -f json` | < 3 seconds | |

---

## Summary

**Total Tests:** 56  
**Passed:** ___  
**Failed:** ___  
**Blockers:** ___

### Blockers (If Any)

List any failing tests that prevent release:
1. 
2. 
3. 

### Notes

Additional observations:
- 
- 

---

**Test Complete:** ✅ Ship v0.8.0-alpha / ❌ Fix issues first

**Tester Signature:** _________________  
**Date:** _________________
