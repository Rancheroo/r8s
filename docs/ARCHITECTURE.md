# r8s Architecture

This document provides a comprehensive overview of r8s's technical architecture, design decisions, and implementation details.

## Table of Contents

- [Overview](#overview)
- [Technology Stack](#technology-stack)
- [Project Structure](#project-structure)
- [Core Components](#core-components)
- [Data Flow](#data-flow)
- [Analysis Engine](#analysis-engine)
- [API Integration](#api-integration)
- [Design Patterns](#design-patterns)

---

## Overview

r8s is a **CLI Automation Engine** for Kubernetes triage. Its core architecture is designed around a headless analysis engine that can ingest, parse, and diagnose support bundles programmatically.

### Key Principles

1.  **Automation First**: Every feature must be accessible via CLI flags and machine-readable output (JSON/SARIF).
2.  **Pipeline Ready**: Designed to run in CI/CD, creating exit codes and reports.
3.  **Stateless Analysis**: Each run is independent and idempotent.
4.  **Graceful Degradation**: Robust handling of incomplete or corrupted bundles.
5.  **Separation of Concerns**: Clear boundaries between the Analysis Engine and CLI presentation layer.

---

## Technology Stack

### Core Dependencies

| Package | Purpose | Version |
|---------|---------|---------|
| [Cobra](https://github.com/spf13/cobra) | CLI framework & command routing | Latest |
| [Viper](https://github.com/spf13/viper) | Configuration | Latest |
| `regexp` | AI Pattern Matching Engine | Stdlib |

---

## Project Structure

```
r8s/
├── main.go                      # Application entry point
├── cmd/                         # CLI Commands
│   ├── root.go                  # Root & Routing
│   ├── analyze.go               # Analysis Command
│   ├── ask.go                   # NLP/Query Command
│   ├── get.go                   # Resource Listing
│   └── tui.go                   # (Removed)
├── internal/                    # Private application code
│   ├── ai/                      # ⭐ Analysis & Pattern Engine
│   │   ├── analyzer.go          # Core analysis logic
│   │   └── patterns/            # Detection patterns
│   ├── bundle/                  # Bundle Parsing & Ingestion
│   ├── config/                  # Configuration
│   ├── rancher/                 # Rancher/K8s API Client
│   └── ui/                      # CLI Presentation Layer
└── docs/                        # Documentation
```

---

## Core Components

### 1. Analysis Engine (`internal/ai/`)

The heart of r8s. It accepts a `Bundle` object and runs a battery of heuristic checks and pattern matching algorithms.

```go
type Analyzer struct {
    patterns []Pattern
}

func (a *Analyzer) Run(bundle *bundle.Bundle) Report {
    // 1. Scan for known log patterns
    // 2. Check resource states (CrashLoop, OOM)
    // 3. Correlate events
    return Report{Issues: [...]}
}
```

### 2. Bundle Processor (`internal/bundle/`)

Responsible for ingesting data from various sources (Tarballs, Directories, Live API) and normalizing it into a standard internal model.

### 3. Presentation Layer (`internal/ui/`)

Handles all user output formatting, including spinners, tables, colors, and structured data (JSON/YAML/SARIF) generation. It ensures consistent experience across all CLI commands.

---

## Data Flow

### 1. CLI / Automation Flow

```
User Command (r8s analyze)
         |
         v
   cmd/analyze.go
         |
         v
   Bundle Processor (Ingest data)
         |
         v
   Analysis Engine (Run patterns)
         |
         v
   JSON/SARIF Formatter
         |
         v
   Stdout (Pipe to jq/Slack)
```
---

## Design Patterns

### 1. Command Pattern (Cobra)

Each CLI command (`analyze`, `get`, `logs`) is encapsulated in a Cobra command struct, allowing for consistent flag parsing, help generation, and execution flow.

### 2. Strategy Pattern (Data Sources)

The `Bundle` interface abstracts the underlying data source. The Analysis Engine doesn't care if the data comes from:
- A live Kubernetes API
- A local directory (extracted bundle)
- A tarball (future support)
- A mock data generator

### 3. Factory Pattern (Mock Data)

Mock data generators act as factories for testing patterns without needing real bundles.

---

## Performance Considerations

### Memory Management

- **Streaming Processing**: Large files (logs) are processed line-by-line where possible to avoid loading 10GB+ files into RAM.
- **Bounded Buffers**: Log tailing uses fixed-size circular buffers.

### Execution Speed

- **Parallel Analysis**: Independent analysis rules run concurrently.
- **Lazy Loading**: Resources are only parsed when requested by a specific rule or command.

---

## Future Enhancements

### Planned Architecture Changes

1.  **Plugin System**: MCP-style extensions for custom analysis rules.
2.  **Remote Analysis**: Support fetching bundles directly from S3/GCS.
3.  **Live Watch Mode**: Continuous analysis of a live cluster.

---

## Testing Strategy

### Unit Tests

- **Config validation**: `internal/config/config_test.go`
- **API client**: `internal/rancher/client_test.go`
- **Mock HTTP responses**: Table-driven tests

### Race Detection

All tests run with `-race` flag:

```bash
go test -race ./...
```

### Coverage

- Current: ~65%
- Target: 80%
- Critical paths: >90%


---

## Conclusion

r8s's architecture prioritizes:
- **Maintainability**: Clear separation of concerns
- **Testability**: Comprehensive test coverage
- **User Experience**: Consistent, clear CLI output
- **Reliability**: Graceful error handling and offline mode

For implementation details, see:
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Development guide
- [README.md](../README.md) - User documentation
- Source code with inline comments

---

**Last Updated**: 2025-11-26  
**Version**: 1.0
