# Sprint 11: v0.9.0 — AI Intelligence
**Pattern Engine v2 + Root Cause Analysis + Export Formats + Natural Language Queries**

**Duration:** 2 weeks  
**Target Date:** March 23, 2026  
**Base Branch:** `main` (post-v0.8.1)  
**Theme:** Smart Analysis

---

## 🎯 Sprint Goal

Transform r8s from a "kubectl for bundles" tool into an **AI-powered diagnostic assistant**. Deliver pattern engine v2 with 10+ detection patterns, actionable root cause hints, CI/CD-ready export formats, and natural language query capability.

**User Value Proposition:**
> "Don't just see what's wrong — understand WHY it's wrong and HOW to fix it."

---

## 📋 Scope

### P0: Pattern Engine v2 — Core Expansion (Days 1-6)
**Impact: CRITICAL | Effort: HIGH | User Value: 🔥🔥🔥**

**Deliverables:**
- [ ] Pattern engine architecture v2 with confidence scoring
- [ ] Pattern correlation engine ("This + That = Root cause")
- [ ] 10+ production-ready patterns:
  - [ ] **etcd** — Corruption, latency, space exceeded
  - [ ] **Certificates** — Expired, expiring soon, invalid CA
  - [ ] **Networking** — DNS failures, CNI errors, connectivity issues
  - [ ] **Storage** — PVC binding failures, storage pressure
  - [ ] **OOMKilled** — Out of memory kills with context
  - [ ] **CrashLoopBackOff** — Crash loops with restart analysis
  - [ ] **ImagePullBackOff** — Image pull failures with registry hints
  - [ ] **Node Pressure** — Disk, memory, PID pressure conditions
  - [ ] **Pod Stuck** — Pending, Terminating, Unknown states
  - [ ] **Leader Election** — Control plane leader issues

**Pattern Schema v2:**
```yaml
name: "CertificateExpired"
description: "Detects expired or expiring certificates"
severity: critical
confidence: high  # Certain/Likely/Possible
sources:
  - "kubectl get nodes -o yaml"
  - "kubectl get pods -n cattle-system"
patterns:
  - regex: "certificate has expired|x509: certificate"
    weight: 1.0
  - regex: "NotAfter.*202[0-5]"  # Old dates
    weight: 0.8
correlations:
  - pattern: "NodeNotReady"
    message: "Node may be NotReady due to expired certificate"
root_cause_hint: "Certificate expired {{.DaysAgo}} days ago. Renew with: kubectl certificate approve {{.CSRName}}"
```

**Success Criteria:**
- All 10 patterns detect real issues in test bundles
- Confidence scoring accurate (>80% precision)
- Pattern correlation identifies related issues
- 65%+ test coverage on pattern engine

---

### P0: Root Cause Hints (Days 4-8)
**Impact: HIGH | Effort: MEDIUM | User Value: 🔥🔥🔥**

**Deliverables:**
- [ ] Root cause hint generation system
- [ ] Context-aware hints for each pattern:
  - "Pod is crashlooping because image tag 'latest' doesn't exist"
  - "Certificate expired 3 days ago — renew with: kubectl certificate approve ..."
  - "Node pressure detected — disk usage at 95%, clean up /var/log"
  - "DNS resolution failing — check CoreDNS pods in kube-system"
- [ ] Hint templates with variable substitution
- [ ] Severity-based hint prioritization

**Hint Template System:**
```go
type RootCauseHint struct {
    Pattern      string            // Pattern ID that triggered
    Severity     string            // critical, warning, info
    Confidence   string            // Certain, Likely, Possible
    Summary      string            // Short description
    Explanation  string            // Detailed explanation
    Suggestion   string            // How to fix
    Command      string            // kubectl command to run
    References   []string          // Links to docs/KB
    Metadata     map[string]string // Variables from pattern match
}
```

**Example Output:**
```
🔴 CRITICAL [Certain] Certificate Expired
   Resource: node/worker-1
   
   Root Cause: Certificate expired 3 days ago (Feb 20, 2026)
   
   Explanation: The kubelet client certificate on node/worker-1 has expired.
   This prevents the node from communicating with the API server.
   
   Suggested Fix:
   1. Approve the pending CSR:
      kubectl certificate approve csr-worker-1-abc123
   
   2. Restart kubelet on the node:
      systemctl restart rke2-agent
   
   Reference: https://docs.rke2.io/security/certificates
```

**Success Criteria:**
- Hints generated for all P0 patterns
- >80% hint accuracy on test bundles
- Users can act on hints without additional research

---

### P1: Export Formats — CI/CD Integration (Days 6-10)
**Impact: HIGH | Effort: MEDIUM | User Value: 🔥🔥🔥**

**Deliverables:**
- [ ] **SARIF export** — Security scanner integration (GitHub Advanced Security, etc.)
- [ ] **JUnit XML export** — CI/CD integration (Jenkins, GitHub Actions, etc.)
- [ ] **Markdown export** — Human-readable reports
- [ ] **JSON export v2** — Structured data with patterns, hints, correlations

**Commands:**
```bash
r8s generate export ./bundle/ --format=sarif --output=results.sarif
r8s generate export ./bundle/ --format=junit --output=results.xml
r8s generate export ./bundle/ --format=markdown --output=report.md
r8s generate export ./bundle/ --format=json --output=results.json
```

**SARIF Format:**
- Tool information (r8s version)
- Rule definitions (one per pattern)
- Results with locations (file, line, column)
- Severity mapping (critical/error/warning/note)

**JUnit Format:**
- Test suites per pattern category
- Test cases per detection
- Failure messages with root cause hints
- Duration and timing info

**Markdown Format:**
- Executive summary
- Pattern detections by severity
- Detailed findings with hints
- Recommended actions checklist

**Success Criteria:**
- SARIF imports successfully to GitHub Advanced Security
- JUnit consumed by Jenkins/GitHub Actions with proper failure reporting
- Markdown renders correctly with clear hierarchy
- All formats include pattern detections and root cause hints

---

### P1: Natural Language Queries v1 (Days 8-12)
**Impact: HIGH | Effort: MEDIUM-HIGH | User Value: 🔥🔥🔥**

**Deliverables:**
- [ ] `r8s ask` command for natural language queries
- [ ] Query parser for common troubleshooting questions
- [ ] Intent mapping to patterns and bundle data
- [ ] Response generation with context from bundle

**Supported Queries (v1):**
```bash
r8s ask ./bundle/ "why is nginx-pod crashing?"
r8s ask ./bundle/ "show me certificate issues"
r8s ask ./bundle/ "which pods are using most memory?"
r8s ask ./bundle/ "why is node worker-1 not ready?"
r8s ask ./bundle/ "find all image pull errors"
```

**Query Parsing:**
```go
type QueryIntent struct {
    Type      string   // "why", "show", "find", "which"
    Resource  string   // "pod", "node", "certificate", "image"
    Name      string   // "nginx-pod", "worker-1"
    Condition string   // "crashing", "not ready", "errors"
    Filter    string   // "most memory", "all"
}
```

**Response Generation:**
- Parse query → Identify intent → Query bundle data → Generate response
- Responses include relevant patterns, logs, and root cause hints
- Fall back to pattern summary if specific answer unavailable

**Example Interaction:**
```bash
$ r8s ask ./bundle/ "why is nginx-pod crashing?"

Analyzing bundle for: nginx-pod crash reasons...

🔴 Found: CrashLoopBackOff with 5 restarts in 10 minutes

Root Cause Analysis:
  The container is exiting immediately after start.
  Last exit code: 1
  
Log excerpt (last 10 lines):
  nginx: [emerg] bind() to 0.0.0.0:80 failed (98: Address already in use)
  
Suggested Fix:
  Check for port conflicts. The nginx container cannot bind to port 80.
  Another process may be using this port on the host network.
  
Related Patterns Detected:
  - CrashLoopBackOff (Critical)
  - ExitCode1 (Warning)
```

**Out of Scope (v2):**
- ❌ Complex multi-hop reasoning
- ❌ Historical trend analysis
- ❌ External knowledge base queries
- ❌ Natural language follow-up questions

**Success Criteria:**
- 5+ query patterns working reliably
- Response time <2 seconds
- >70% query understanding accuracy
- Helpful responses for supported queries

---

### P2: Pattern Registry & Management (Days 10-13)
**Impact: MEDIUM | Effort: LOW | User Value: 🔥🔥**

**Deliverables:**
- [ ] Pattern registry with versioning
- [ ] Hot-reload pattern definitions (YAML files)
- [ ] Pattern listing command: `r8s patterns list`
- [ ] Pattern validation command: `r8s patterns validate`
- [ ] Custom pattern support (user-defined patterns)

**Commands:**
```bash
r8s patterns list                    # Show all patterns
r8s patterns list --severity=critical # Filter by severity
r8s patterns show CertificateExpired # Show pattern details
r8s patterns validate ./my-pattern.yaml # Validate custom pattern
```

**Success Criteria:**
- Patterns load from YAML files at runtime
- Custom patterns work alongside built-ins
- Pattern listing shows all metadata

---

### P2: Performance & Polish (Days 12-14)
**Impact: MEDIUM | Effort: LOW | User Value: 🔥🔥**

**Deliverables:**
- [ ] Pattern analysis <2s for 100MB bundles
- [ ] Parallel pattern matching (goroutines per pattern)
- [ ] Progress indicator for long analyses
- [ ] Memory optimization for large log files
- [ ] CLI help and documentation for all new commands

**Success Criteria:**
- Analysis completes in <2s for standard bundles
- No memory leaks during batch processing
- Progress shown for bundles >50MB
- All commands have comprehensive help

---

## 📅 Day-by-Day Plan

### Week 1: Pattern Engine & Root Cause Foundation

#### Day 1: Pattern Engine v2 Architecture
- [ ] Design confidence scoring system
- [ ] Implement pattern correlation engine
- [ ] Update pattern schema to v2
- [ ] Refactor existing patterns (OOM, CrashLoop, ImagePull) to v2

#### Day 2: Core Patterns — etcd & Certificates
- [ ] Implement etcd pattern (corruption, latency, space)
- [ ] Implement certificate pattern (expired, expiring, invalid CA)
- [ ] Add correlation rules between certificates and node status
- [ ] Write unit tests

#### Day 3: Core Patterns — Networking & Storage
- [ ] Implement networking pattern (DNS, CNI, connectivity)
- [ ] Implement storage pattern (PVC binding, storage pressure)
- [ ] Add cross-pattern correlations
- [ ] Write unit tests

#### Day 4: Core Patterns — Node & Pod States
- [ ] Implement node pressure pattern (disk, memory, PID)
- [ ] Implement pod stuck pattern (Pending, Terminating, Unknown)
- [ ] Implement leader election pattern
- [ ] Write unit tests

#### Day 5: Root Cause Hint System
- [ ] Design hint template system
- [ ] Implement variable substitution
- [ ] Create hint generators for all 10 patterns
- [ ] Write hint tests

#### Day 6: Pattern Integration & Testing
- [ ] Integrate all patterns into analysis flow
- [ ] Run integration tests with sample bundles
- [ ] Tune confidence thresholds
- [ ] Fix any detection issues

#### Day 7: Buffer / Mid-Sprint Review
- [ ] Complete any unfinished Day 1-6 work
- [ ] Review pattern accuracy with real bundles
- [ ] Adjust priorities if needed
- [ ] **Milestone:** All 10 patterns detecting correctly

---

### Week 2: Export Formats, NLQ, & Release

#### Day 8: SARIF Export
- [ ] Implement SARIF schema generation
- [ ] Map patterns to SARIF rules
- [ ] Add tool information and version
- [ ] Test with GitHub Advanced Security import

#### Day 9: JUnit & Markdown Export
- [ ] Implement JUnit XML generation
- [ ] Implement Markdown report generation
- [ ] Add formatting and styling
- [ ] Test with Jenkins and manual review

#### Day 10: Natural Language Queries v1
- [ ] Implement query parser
- [ ] Create intent mapping for 5 query types
- [ ] Implement `r8s ask` command
- [ ] Add response generation

#### Day 11: NLQ Enhancement & Pattern Registry
- [ ] Expand query understanding
- [ ] Implement pattern registry commands
- [ ] Add pattern listing and validation
- [ ] Support custom patterns

#### Day 12: Performance & Polish
- [ ] Parallel pattern matching
- [ ] Progress indicators
- [ ] Memory optimization
- [ ] CLI help completion

#### Day 13: Testing & Documentation
- [ ] Full integration testing
- [ ] Export format validation
- [ ] Documentation updates
- [ ] README updates for v0.9.0

#### Day 14: Release
- [ ] Final regression testing
- [ ] Coverage check (60%+ target)
- [ ] Tag v0.9.0
- [ ] Release notes

---

## ✅ Success Criteria

| Metric | Target | Measurement |
|--------|--------|-------------|
| Patterns | 10+ working | `r8s patterns list` shows 10+ patterns |
| Pattern Accuracy | >80% | Manual validation on test bundles |
| Root Cause Hints | All P0 patterns | Each critical pattern has hint |
| Export Formats | 4 formats | SARIF, JUnit, Markdown, JSON v2 |
| NLQ Queries | 5+ types | Documented query patterns work |
| Performance | <2s analysis | Benchmark on 100MB bundle |
| Test Coverage | 60%+ | `make coverage` |
| CI Status | 100% green | No disabled jobs |

---

## 🛠️ Implementation Notes

### Pattern Engine v2 Design
```go
// Pattern with confidence and correlation
type PatternV2 struct {
    ID            string
    Name          string
    Description   string
    Severity      Severity
    Confidence    Confidence // Certain, Likely, Possible
    Matchers      []Matcher
    Correlations  []Correlation
    HintGenerator HintGenerator
}

type MatchResult struct {
    PatternID   string
    Matched     bool
    Confidence  Confidence
    Resources   []Resource
    Evidence    []string
    Correlated  []string // IDs of correlated patterns
}
```

### Confidence Scoring
- **Certain:** Direct evidence, unambiguous (e.g., "certificate has expired")
- **Likely:** Strong evidence, minor ambiguity possible (e.g., "NotAfter date in past")
- **Possible:** Some evidence, needs verification (e.g., "connection refused" could be network or service)

### Correlation Engine
```go
type Correlation struct {
    PatternID   string // Pattern to correlate with
    Condition   string // Condition for correlation
    Message     string // Message to show when correlated
}

// Example: Certificate + NodeNotReady = Certificate likely cause
```

### Export Architecture
```go
type Exporter interface {
    Export(analysis *AnalysisResult, w io.Writer) error
}

type SARIFExporter struct{}
type JUnitExporter struct{}
type MarkdownExporter struct{}
type JSONExporter struct{}
```

---

## 📚 Dependencies from Previous Sprints

### From Sprint 10 (Quality Gates)
- [ ] v0.8.1 K3s support complete
- [ ] CI pipeline stable (all green)
- [ ] Test coverage baseline ≥55%
- [ ] Bundle parsing robust for all formats

### From Sprint 10.1 (CI Cleanup)
- [ ] No disabled CI jobs
- [ ] Coverage reporting reliable
- [ ] Build process optimized

### From v0.8.0/v0.8.1
- [ ] kubectl commands stable (get, logs, describe)
- [ ] Bundle format detection working (RKE2, K3s)
- [ ] Pattern engine v1 foundation (if applicable)

---

## ⚠️ Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Pattern accuracy below target | Medium | High | Start with keyword matching, add regex later; confidence scoring allows partial credit |
| NLQ complexity underestimated | Medium | High | Scope v1 to 5 query patterns only; use simple intent matching |
| Export format compliance | Low | Medium | Test early with target tools (GitHub, Jenkins); iterate |
| Performance with 10 patterns | Low | Medium | Parallel pattern matching; benchmark early |
| Root cause hints too generic | Low | High | Template-based with variable substitution; manual review |
| Scope creep to v2 features | Medium | High | Strict adherence to v1 scope; document v2 ideas for future |

### Key Risk Mitigations
1. **Pattern Accuracy:** Use existing log data to validate patterns before implementation
2. **NLQ Scope:** Hard limit of 5 query types; complex queries out of scope
3. **Performance:** Parallel matching from day 1; benchmark on Day 3
4. **Hint Quality:** Review all hints with support engineers for accuracy

---

## 🚫 Out of Scope (Deferred to v0.9.5/v1.0)

**Intentionally NOT in Sprint 11:**
- ❌ Self-learning patterns (anomaly detection, baselines)
- ❌ Historical trend analysis
- ❌ Complex NLQ with context/follow-ups
- ❌ Predictive analysis ("this will fail soon")
- ❌ External knowledge base integration
- ❌ Live cluster pattern detection
- ❌ Pattern marketplace/sharing
- ❌ Advanced correlations (>2 patterns)
- ❌ Automatic remediation suggestions

**Why:** Focus on core value — accurate detection and clear explanations. Advanced features build on this foundation.

---

## 🎯 Definition of Done

- [ ] 10+ patterns implemented and tested
- [ ] Pattern confidence scoring working
- [ ] Pattern correlation identifying related issues
- [ ] Root cause hints for all P0 patterns
- [ ] SARIF export tested with GitHub
- [ ] JUnit export tested with Jenkins/Actions
- [ ] Markdown export generating readable reports
- [ ] `r8s ask` command with 5+ query types
- [ ] Pattern registry with list/validate commands
- [ ] Custom pattern support
- [ ] Analysis performance <2s for 100MB bundles
- [ ] Test coverage ≥60%
- [ ] All CI jobs passing
- [ ] Documentation updated
- [ ] v0.9.0 tagged and released

---

## 📝 Success Metrics Validation

### How to Verify

**Pattern Accuracy:**
```bash
# Run analysis on test bundles with known issues
r8s analyze ./test-bundle/ --format=json

# Check that patterns match expected issues
# Manual review: Expected 8 issues, detected 7 = 87.5% accuracy
```

**Export Formats:**
```bash
# SARIF validation
r8s generate export ./bundle/ --format=sarif --output=test.sarif
# Upload to GitHub repository → Security tab → Code scanning

# JUnit validation
r8s generate export ./bundle/ --format=junit --output=test.xml
# Import to Jenkins or view in GitHub Actions

# Markdown review
r8s generate export ./bundle/ --format=markdown --output=report.md
cat report.md  # Verify formatting and content
```

**Natural Language Queries:**
```bash
# Test each query type
r8s ask ./bundle/ "why is nginx-pod crashing?"
r8s ask ./bundle/ "show me certificate issues"
r8s ask ./bundle/ "which pods are using most memory?"
r8s ask ./bundle/ "why is node worker-1 not ready?"
r8s ask ./bundle/ "find all image pull errors"

# Verify helpful responses for each
```

---

## 👥 Team Responsibilities

| Role | Focus |
|------|-------|
| **RancherSRE** | Pattern engine v2, root cause hints, NLQ |
| **CodeRabbit** | All PR reviews, quality gates |
| **Community/Contributors** | Pattern definitions, test bundles |

---

## 📚 Related Documents

- `ROADMAP_v1.0_FINAL.md` — v1.0 roadmap with v0.9.0 scope
- `SPRINT9_v0.8.1_PLAN.md` — Previous sprint (K3s support)
- `QUICK_WINS_RESEARCH.md` — Feature research and kubectl gaps
- Pattern definitions: `internal/ai/patterns/*.yaml`

---

## 🎉 Post-Sprint Vision

After Sprint 11, r8s will:

1. **Detect** 10+ common cluster issues automatically
2. **Explain** root causes in plain English
3. **Suggest** specific remediation steps
4. **Integrate** with CI/CD via standard formats
5. **Answer** natural language troubleshooting questions

**User Workflow:**
```bash
# Analyze bundle and get actionable insights
r8s analyze ./production-bundle/
# Shows: 🔴 3 critical issues with root causes and fixes

# Export for CI/CD
r8s generate export ./bundle/ --format=sarif

# Ask questions in plain English
r8s ask ./bundle/ "why is the API server slow?"
```

---

*"From kubectl compatibility to AI-powered diagnostics. This is where r8s becomes indispensable."*
