# r8s v0.7.x Release Series Roadmap

**Series Theme:** "From Functional to Intelligent"

---

## 📊 Release Overview

| Version | Target | Theme | Key Deliverables |
|---------|--------|-------|------------------|
| **v0.7.0** | Feb 2026 | Maximum Information | ✅ RELEASED - 90% bundle coverage |
| **v0.7.1** | Mar 2026 | Multi-Distro Support | K3s + RKE1 compatibility |
| **v0.7.2** | Mar 2026 | CI Stability | All checks green, 50% coverage |
| **v0.7.3** | Apr 2026 | AI Analysis v1 | Smart issue correlation, root cause hints |
| **v0.7.4** | Apr 2026 | AI Assistant Mode | Natural language queries, guided debugging |
| **v0.7.5** | May 2026 | Performance | <2s dashboard, memory optimization |

---

## 🔮 v0.7.1 "Multi-Distro Support" (March 2026)

**Goal:** Support K3s and RKE1 bundles, not just RKE2

### Features
- [ ] K3s bundle format detection
- [ ] RKE1 bundle format detection  
- [ ] Dynamic path abstraction (no hardcoded `rke2/`)
- [ ] Distro-specific parsers (etcd locations vary)
- [ ] Test bundles for all three distros

### Technical Work
- [ ] Refactor 18+ hardcoded RKE2 paths
- [ ] Add `DistroType` to Bundle struct
- [ ] Create path helper methods
- [ ] Update validation logic

**Files:**
- `internal/bundle/types.go` - Add FormatK3s, FormatRKE1
- `internal/bundle/manifest.go` - Dynamic path resolution
- `internal/bundle/detect.go` - Multi-distro detection

**Estimated:** 2-3 week sprint

---

## 🔧 v0.7.2 "CI Stability & Quality Gates" (March 2026)

**Goal:** All CI checks passing, enforce quality automatically

### Features
- [ ] Fix all golangci-lint warnings (#44)
- [ ] Re-enable lint job in CI
- [ ] Achieve 50% test coverage (#45)
- [ ] Re-enable coverage threshold check
- [ ] Cross-platform builds (Linux, macOS, Windows)
- [ ] Remove dead code
- [ ] Standardize error handling

### Quality Gates (Enforced)
- [ ] All PRs must pass lint
- [ ] All PRs must have >50% coverage on new code
- [ ] No disabled CI jobs
- [ ] Clean build on all platforms

**Estimated:** 1-2 week sprint

---

## 🤖 v0.7.3 "AI Analysis Engine v1" (April 2026)

**Goal:** AI-powered issue correlation and root cause analysis

### Features

#### 1. Smart Issue Correlation
```
Before: Show 50 separate warnings
After:  "These 12 warnings are all caused by etcd leader election issues"
```

- [ ] Pattern matching engine for related issues
- [ ] Event clustering (group by root cause)
- [ ] Cross-resource correlation (pod → node → etcd)
- [ ] Severity scoring (not just Warning/Error)

#### 2. Root Cause Hints
```
Issue: Pod stuck in Pending
AI Hint: "Node node-1 has DiskPressure. Clean up images with: crictl rmi --prune"
```

- [ ] Knowledge base of common root causes
- [ ] Context-aware suggestions
- [ ] Confidence scoring (Certain/Likely/Possible)

#### 3. Anomaly Detection
- [ ] Detect unusual patterns ("This pod has restarted 50x more than others")
- [ ] Compare against healthy baselines
- [ ] Highlight outliers in resource usage

#### 4. Log Insight Extraction
- [ ] Automatically extract key errors from large logs
- [ ] Summarize repeated patterns
- [ ] Link errors to specific pods/timelines

### Technical Implementation
```go
// New package: internal/ai
package ai

// IssueCorrelator finds related problems
type IssueCorrelator struct {
    patterns []CorrelationPattern
}

func (ic *IssueCorrelator) FindClusters(issues []Issue) []IssueCluster

// RootCauseAnalyzer suggests fixes
type RootCauseAnalyzer struct {
    knowledgeBase KnowledgeBase
}

func (rca *RootCauseAnalyzer) Analyze(issue Issue) RootCauseHint

// AnomalyDetector finds outliers
type AnomalyDetector struct {
    baselines BaselineStore
}

func (ad *AnomalyDetector) Detect(bundle *Bundle) []Anomaly
```

### AI Knowledge Base
- [ ] Common Kubernetes failure patterns
- [ ] Rancher/RKE2 specific issues
- [ ] Historical issue database (from memories)
- [ ] CVE correlations

**Estimated:** 3-4 week sprint

---

## 🗣️ v0.7.4 "AI Assistant Mode" (April 2026)

**Goal:** Natural language interaction with r8s

### Features

#### 1. Natural Language Queries
```bash
$ r8s --ask "why is pod nginx crashing?"

AI: Looking at pod nginx in default namespace...
    Found: 3 OOMKilled events in last hour
    Memory limit: 128Mi, Peak usage: 156Mi
    Suggestion: Increase memory limit to 256Mi
    Command: kubectl set resources deployment/nginx --limits=memory=256Mi
```

- [ ] Parse natural language questions
- [ ] Query bundle data intelligently
- [ ] Generate human-readable answers

#### 2. Guided Debugging
```bash
$ r8s --guided-debug

AI: I see your cluster has issues. Let's diagnose:
    1. Is it a node problem? [Check nodes - 2 NotReady]
    2. Is it a networking issue? [Check CNI - calico pods crashing]
    3. Is it resource pressure? [Check top - node-1 at 95% memory]
    
    Root cause: Node memory pressure causing CNI failures
    Fix: Add memory or reduce workload
```

- [ ] Interactive troubleshooting wizard
- [ ] Step-by-step diagnosis
- [ ] Context-aware questions

#### 3. Report Generation
```bash
$ r8s --generate-report --format=markdown

Creates: comprehensive analysis report with:
- Executive summary
- Critical issues (top 5)
- Recommendations (prioritized)
- Supporting evidence (logs, events)
```

- [ ] One-click report generation
- [ ] Multiple formats (Markdown, PDF, HTML)
- [ ] Executive summary vs detailed view

### Technical Implementation
```go
// New package: internal/assistant
package assistant

// NLQueryParser converts questions to queries
type NLQueryParser struct {
    model LLMClient  // Local or API-based
}

func (nlp *NLQueryParser) Parse(query string) (*QueryIntent, error)

// GuidedDebugger interactive troubleshooting
type GuidedDebugger struct {
    bundle *Bundle
    state  DebugState
}

func (gd *GuidedDebugger) NextStep() DebugStep

// ReportGenerator creates analysis reports
type ReportGenerator struct {
    template ReportTemplate
}

func (rg *ReportGenerator) Generate(bundle *Bundle, format Format) (Report, error)
```

### Integration Options
**Option A: Local LLM (Privacy-first)**
- Use local model (llama.cpp, ollama)
- No data leaves machine
- Slower but fully private

**Option B: API with opt-in**
- Use OpenAI/Anthropic APIs
- User must explicitly enable
- Clear data usage warnings
- Fallback to local if no connectivity

**Option C: Hybrid**
- Simple queries: local model
- Complex analysis: API (with permission)
- User controls per-session

**Estimated:** 3-4 week sprint

---

## ⚡ v0.7.5 "Performance Optimization" (May 2026)

**Goal:** <2s dashboard load even with 1000+ pods

### Features
- [ ] Parallel bundle parsing
- [ ] Incremental dashboard updates
- [ ] Memory-mapped file access for large bundles
- [ ] Lazy loading for off-screen data
- [ ] Compression for parsed data
- [ ] Benchmark suite

### Targets
- Dashboard load: <2s (from current 5-10s on large bundles)
- Memory usage: <500MB for 1GB bundles (from ~1GB)
- 1000+ pods handled smoothly

**Estimated:** 2-3 week sprint

---

## 📅 Release Cadence

```
Feb: v0.7.0 ✅ RELEASED
Mar: v0.7.1 (K3s) + v0.7.2 (CI)
Apr: v0.7.3 (AI Analysis) + v0.7.4 (Assistant)
May: v0.7.5 (Performance)
Jun: v0.8.0 Planning Begins
```

**Hotfix Policy:**
- Critical bug: v0.7.X+1 within 24h
- Security issue: v0.7.X+1 immediately
- Minor fix: Roll into next planned release

---

## 🎯 Success Metrics

| Metric | v0.7.0 | v0.7.5 Target |
|--------|--------|---------------|
| Bundle Coverage | 90% | 95% |
| Test Coverage | 10% | 70% |
| Dashboard Load | 5-10s | <2s |
| Distros Supported | 1 (RKE2) | 3 (RKE2, K3s, RKE1) |
| AI Features | 0 | 2 major |
| CI Passing | Partial | 100% |

---

## 🔄 v0.8.0 Preview (June-August 2026)

**Theme:** "Production Hardening"

- 80% test coverage
- Complete documentation
- Security audit
- Stress testing (10M+ log lines)
- Enterprise features

---

*This roadmap is living - update as priorities shift.*
