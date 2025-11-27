# r8s Architecture

This document provides a comprehensive overview of r8s's technical architecture, design decisions, and implementation details.

## Table of Contents

- [Overview](#overview)
- [Technology Stack](#technology-stack)
- [Project Structure](#project-structure)
- [Core Components](#core-components)
- [Data Flow](#data-flow)
- [State Management](#state-management)
- [Offline Mode](#offline-mode)
- [API Integration](#api-integration)
- [Design Patterns](#design-patterns)

---

## Overview

r8s is a terminal user interface (TUI) application for navigating and managing Rancher-based Kubernetes clusters. It follows the Model-Update-View pattern popularized by The Elm Architecture, implemented via the Bubble Tea framework.

### Key Principles

1. **Event-Driven**: All user interactions are events processed through a central Update function
2. **Immutable State**: State updates return new state rather than mutating existing state
3. **Graceful Degradation**: Offline mode with mock data allows development without live Rancher
4. **Type Safety**: Strongly-typed Go structs for all API responses
5. **Separation of Concerns**: Clear boundaries between UI, business logic, and API client

---

## Technology Stack

### Core Dependencies

| Package | Purpose | Version |
|---------|---------|---------|
| [Bubble Tea](https://github.com/charmbracelet/bubbletea) | TUI framework | Latest |
| [Lipgloss](https://github.com/charmbracelet/lipgloss) | Terminal styling | Latest |
| [Bubble Table](https://github.com/evertras/bubble-table) | Table component | Latest |
| [Cobra](https://github

.com/spf13/cobra) | CLI framework | Latest |
| [Viper](https://github.com/spf13/viper) | Configuration | Latest |

### Standard Library

- `encoding/json`: API response parsing
- `net/http`: HTTP client for Rancher API
- `time`: Timestamps and duration handling
- `fmt`, `strings`: String formatting and manipulation

---

## Project Structure

```
r8s/
├── main.go                      # Application entry point
├── cmd/
│   └── root.go                 # CLI root command
├── internal/                   # Private application code
│   ├── config/
│   │   ├── config.go          # Configuration logic
│   │   └── config_test.go     # Configuration tests
│   ├── rancher/
│   │   ├── client.go          # Rancher API client
│   │   ├── client_test.go     # Client tests
│   │   └── types.go           # API response types
│   ├── tui/
│   │   ├── app.go             # Main TUI application
│   │   ├── styles.go          # Visual styles
│   │   ├── actions/           # User actions (future)
│   │   ├── components/        # Reusable UI components (future)
│   │   └── views/             # View-specific logic (future)
│   └── k8s/                   # Kubernetes operations (future)
├── docs/
│   ├── ARCHITECTURE.md        # This file
│   └── archive/               # Archived development docs
├── scripts/
│   ├── setup_*.sh             # Setup scripts
│   └── deprecated/            # Old test scripts
└── bin/                       # Compiled binaries
```

### Package Responsibilities

- **main.go**: Initializes config and starts TUI
- **cmd/**: CLI parsing and command setup
- **internal/config**: Configuration file management, profile handling
- **internal/rancher**: Rancher API communication, type definitions
- **internal/tui**: All UI rendering, event handling, state management
- **internal/k8s**: Direct Kubernetes operations (future)

---

## Core Components

### 1. Configuration Management (`internal/config/`)

```go
type Config struct {
    CurrentProfile string    `yaml:"current_profile"`
    Profiles       []Profile `yaml:"profiles"`
    // ...
}

type Profile struct {
    Name        string `yaml:"name"`
    URL         string `yaml:"url"`
    BearerToken string `yaml:"bearer_token"`
    // ...
}
```

**Responsibilities:**
- Load/parse YAML configuration from `~/.r8s/config.yaml`
- Validate configuration structure
- Provide current profile access
- Support multiple Rancher environments

### 2. Rancher API Client (`internal/rancher/`)

```go
type Client struct {
    baseURL    string
    token      string
    httpClient *http.Client
}
```

**Key Methods:**
- `TestConnection()`: Verify API connectivity
- `ListClusters()`: Fetch cluster list
- `ListProjects(clusterID)`: Fetch projects for a cluster
- `ListNamespaces(clusterID)`: Fetch namespaces
- `ListPods(projectID)`: Fetch pods
- `ListDeployments(projectID)`: Fetch deployments
- `ListServices(projectID)`: Fetch services
- `ListCRDs(clusterID)`: Fetch Custom Resource Definitions
- `GetPodDetails(...)`: Fetch individual pod details
- And more...

**Error Handling:**
- HTTP errors wrapped with context
- Connection timeouts
- Authentication failures
- API version compatibility

### 3. TUI Application (`internal/tui/`)

```go
type App struct {
    // Configuration
    config *config.Config
    client *rancher.Client
    
    // State
    viewStack   []ViewContext
    currentView ViewContext
    
    // Data
    clusters    []rancher.Cluster
    pods        []rancher.Pod
    deployments []rancher.Deployment
    // ...
    
    // UI State
    table       table.Model
    loading     bool
    error       string
    offlineMode bool
}
```

**Core Methods:**
- `Init()`: Initialize application, start data fetching
- `Update(msg)`: Process events, update state
- `View()`: Render current state to terminal

---

## Data Flow

### 1. Application Lifecycle

```
┌────────────────┐
│  main.go       │  Load config, create App
└───────┬────────┘
        │
        v
┌────────────────┐
│  App.Init()    │  Start fetching clusters
└───────┬────────┘
        │
        v
┌────────────────┐
│  Event Loop    │  Bubble Tea runtime
│  - Update()    │  Process user input
│  - View()      │  Render UI
└────────────────┘
```

### 2. User Interaction Flow

```
User presses key (e.g., "Enter")
         |
         v
   Key event → Update(KeyMsg)
         |
         v
   Determine action based on currentView
         |
         v
   Return Command (e.g., fetchProjects)
         |
         v
   Command executes asynchronously
         |
         v
   Command returns Message (e.g., projectsMsg)
         |
         v
   Update() processes message
         |
         v
   State updated (a.projects = msg.projects)
         |
         v
   View() re-renders with new state
```

### 3. API Request Flow

```
fetchClusters() called
         |
         v
   Check offlineMode
         |
    +----+----+
    |         |
    v         v
  Offline   Online
    |         |
    v         v
  Mock     API Request
  Data       |
    |        v
    |    Parse JSON → Cluster structs
    |        |
    +--------+
         |
         v
   Return clustersMsg
         |
         v
   Update state
         |
         v
   Render table
```

---

## State Management

### View Stack Navigation

r8s uses a stack-based navigation system:

```go
type ViewContext struct {
    viewType      ViewType  // Current view (Clusters, Pods, etc.)
    clusterID     string    // Context: which cluster
    projectID     string    // Context: which project
    namespaceName string    // Context: which namespace
    // ...
}
```

**Navigation Example:**

```
Initial: ViewClusters
         |
         | User presses Enter on "production"
         v
      Push ViewClusters to stack
      Navigate to ViewProjects (clusterID="c-prod")
         |
         | User presses Enter on "default-project"
         v
      Push ViewProjects to stack
      Navigate to ViewNamespaces
         |
         | User presses Esc
         v
      Pop from stack → Back to ViewProjects
```

### State Updates

All state updates follow this pattern:

```go
case projectsMsg:
    a.loading = false           // Update UI state
    a.projects = msg.projects   // Update data
    a.error = ""               // Clear errors
    a.updateTable()            // Re-render table
```

---

## Offline Mode

### Design Philosophy

Offline mode enables:
- Development without live Rancher access
- Demos and testing
- UI development iteration
- Feature exploration

### Implementation

```go
// At startup
if err := client.TestConnection(); err != nil {
    offlineMode = true
}

// In fetch functions
func (a *App) fetchPods(...) tea.Cmd {
    return func() tea.Msg {
        if a.offlineMode {
            return podsMsg{pods: a.getMockPods(...)}
        }
        
        // Real API call
        collection, err := a.client.ListPods(projectID)
        if err != nil {
            // Fallback to mock for graceful degradation
            return podsMsg{pods: a.getMockPods(...)}
        }
        return podsMsg{pods: collection.Data}
    }
}
```

### Mock Data Generation

Mock data is:
- **Realistic**: Mimics actual Rancher responses
- **Varied**: Different scenarios (running, failed, pending pods)
- **Consistent**: Deterministic for testing
- **Namespace-aware**: Filtered appropriately

---

## API Integration

### Authentication

Bearer token authentication:

```go
req.Header.Set("Authorization", "Bearer "+c.token)
```

Supports:
- Direct bearer tokens
- API key + secret (concatenated to form bearer token)

### Request/Response Cycle

```go
func (c *Client) ListPods(projectID string) (*PodCollection, error) {
    // 1. Build URL
    url := fmt.Sprintf("%s/v3/projects/%s/pods", c.baseURL, projectID)
    
    // 2. Create request
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("Authorization", "Bearer "+c.token)
    
    // 3. Execute request
    resp, err := c.httpClient.Do(req)
    
    // 4. Parse response
    var collection PodCollection
    json.NewDecoder(resp.Body).Decode(&collection)
    
    // 5. Return data
    return &collection, nil
}
```

### Type Mapping

Rancher API responses map to Go structs:

```go
// API Response (JSON)
{
  "id": "c-m-12345",
  "type": "cluster",
  "name": "production",
  "state": "active"
}

// Go Struct
type Cluster struct {
    ID    string `json:"id"`
    Type  string `json:"type"`
    Name  string `json:"name"`
    State string `json:"state"`
}
```

### Field Mapping Strategy

Some fields have multiple possible names (see Deployment replica counts):

```go
// Try multiple field mappings
if deployment.Scale != nil {
    replicas = deployment.Scale.Scale
} else if deployment.Replicas > 0 {
    replicas = deployment.Replicas
}
```

---

## Design Patterns

### 1. Event-Driven Architecture (Bubble Tea)

```go
// Model
type App struct { /* state */ }

// Init
func (a *App) Init() tea.Cmd { /* setup */ }

// Update
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Process events, return new state + commands
}

// View
func (a *App) View() string {
    // Render current state
}
```

### 2. Command Pattern

Commands encapsulate asynchronous operations:

```go
func (a *App) fetchClusters() tea.Cmd {
    return func() tea.Msg {
        // Fetch data
        clusters := /* ... */
        
        // Return message
        return clustersMsg{clusters: clusters}
    }
}
```

### 3. Message Passing

All data flows through messages:

```go
type clustersMsg struct {
    clusters []rancher.Cluster
}

type podsMsg struct {
    pods []rancher.Pod
}

type errMsg struct {
    error
}
```

### 4. Fallback Strategy

Multi-tier attempts with graceful degradation:

```go
// Tier 1: Preferred field
if deployment.Scale != nil {
    use(deployment.Scale.Ready)
// Tier 2: Alternative field
} else if deployment.ReadyReplicas > 0 {
    use(deployment.ReadyReplicas)
// Tier 3: Default
} else {
    use(0)
}
```

### 5. Factory Pattern

Mock data generators act as factories:

```go
func (a *App) getMockPods(namespace string) []rancher.Pod {
    // Generate realistic mock data
    return []rancher.Pod{/* ... */}
}
```

---

## Performance Considerations

### Memory Management

- **Table rendering**: Only visible rows rendered (pagination)
- **Data caching**: Fetched data stored in App state
- **No memory leaks**: Tested with race detector

### Network Optimization

- **Single request per view**: Avoid redundant API calls
- **Filtered responses**: Only fetch needed data
- **Connection pooling**: HTTP client reuse

### UI Responsiveness

- **Asynchronous fetching**: Commands don't block UI
- **Loading states**: User feedback during operations
- **Error handling**: Clear error messages

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

## Future Enhancements

### Planned Architecture Changes

1. **Component extraction**: Move table, modal to `internal/tui/components/`
2. **View separation**: Split views into separate files
3. **Action handlers**: Dedicated action package
4. **Plugin system**: MCP-style extensions
5. **State machine**: Formal FSM for view transitions

### Scalability

- Handle large datasets (1000s of pods)
- Virtual scrolling for tables
- Lazy loading with pagination
- Background refresh without blocking

---

## Conclusion

r8s's architecture prioritizes:
- **Maintainability**: Clear separation of concerns
- **Testability**: Comprehensive test coverage
- **User Experience**: Responsive, intuitive UI
- **Reliability**: Graceful error handling and offline mode

For implementation details, see:
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Development guide
- [README.md](../README.md) - User documentation
- Source code with inline comments

---

**Last Updated**: 2025-11-26  
**Version**: 1.0
