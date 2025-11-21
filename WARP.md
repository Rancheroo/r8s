# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

r9s (Rancher9s) is a k9s-inspired terminal UI (TUI) application for managing Rancher-based Kubernetes clusters. It's built in Go using the Bubble Tea framework and provides keyboard-driven navigation for Rancher's multi-cluster management capabilities.

**Current Status:** Alpha - Basic cluster listing works, project/namespace/pod views are planned but not yet implemented.

## Development Commands

### Building and Running
```bash
# Build binary to ./bin/r9s
make build

# Install to $GOPATH/bin
make install

# Run directly without building
make run

# Run tests
make test
```

### Development Workflow
```bash
# Download/update dependencies
make tidy

# Format code
make fmt

# Run linter
make vet

# Run all dev checks (tidy + fmt + vet)
make dev

# Clean build artifacts
make clean
```

### Testing with Rancher Instance
The test Rancher instance is at: `https://rancher.do.4rl.io`

Configuration is stored in `~/.r9s/config.yaml`. You'll need to set `insecure: true` for this test instance due to certificate issues.

## Git Workflow

### Branching Strategy
We follow a feature-branch workflow to ensure code stability and organized development.

1.  **Master Branch**: The `master` branch should always be stable and buildable.
2.  **Feature Branches**: Create a new branch for each specific feature or fix.
    *   Format: `feature/feature-name` or `fix/issue-description`
    *   Example: `feature/crd-browser`, `fix/pod-navigation`
3.  **Commit Messages**: Use descriptive commit messages explaining *what* and *why*.
4.  **Merging**: When a feature is complete:
    *   Ensure `make dev` passes (fmt, vet, tidy)
    *   Merge into `master` (squash merge preferred for cleaner history)

## Code Architecture

### High-Level Structure
```
r9s/
├── cmd/                          # CLI layer (Cobra)
│   └── root.go                  # Command definitions, flags, version handling
├── internal/
│   ├── config/                  # Configuration management
│   │   └── config.go           # YAML config, multi-profile support
│   ├── rancher/                 # Rancher API client
│   │   ├── client.go           # HTTP client with bearer token auth
│   │   └── types.go            # API response type definitions
│   ├── tui/                     # Terminal UI (Bubble Tea)
│   │   ├── app.go              # Main Bubble Tea model (Elm Architecture)
│   │   └── styles.go           # Lipgloss styles (k9s-inspired colors)
│   └── k8s/                     # Kubernetes operations (future)
└── main.go                      # Entry point
```

### Key Architectural Patterns

#### Bubble Tea (TUI Framework)
The TUI follows The Elm Architecture:
- **Model** (`App` struct): Holds application state (clusters, table, loading, error)
- **Init()**: Returns initial commands (fetch clusters, enter alt screen)
- **Update(Msg)**: Message handler that updates model and returns new commands
- **View()**: Renders the current UI state

Message passing is used for async operations:
- `clustersMsg`: Contains fetched cluster data
- `errMsg`: Contains error information
- `tea.WindowSizeMsg`: Terminal resize events
- `tea.KeyMsg`: Keyboard input

#### Rancher API Client
- Direct HTTP calls to Rancher v3 API (no official SDK dependency)
- Bearer token authentication: `Authorization: Bearer token-key:secret`
- Supports both combined token and separate access/secret key pairs
- TLS insecure skip option for dev environments

**Important API Structure Details:**
- The `sort` field in collections is an object with `order`, `reverse`, and `links` sub-fields, NOT a simple map
- The `version` field in clusters is an object (`ClusterVersion`) with `gitVersion`, `major`, `minor` fields, NOT a string
- These non-obvious structures cause JSON unmarshaling errors if defined incorrectly

#### Configuration
- Multi-profile support via `~/.r9s/config.yaml`
- Profiles contain: URL, bearerToken (or accessKey+secretKey), insecure flag
- CLI flags override config values
- Auto-creates default config on first run

### Navigation State (Not Yet Implemented)
The planned navigation flow:
```
Clusters → Projects → Namespaces → Resources (Pods/Deployments/etc.)
    ↓
    Cluster-level Resources (Nodes, Catalogs)
```

Breadcrumb navigation will show: `Cluster: {name} > Project: {name} > Namespace: {name} > {ResourceType}`

## Important Implementation Notes

### Working with Rancher API Types
When adding new Rancher API endpoints:

1. **Always inspect the actual JSON response first** using curl:
   ```bash
   curl -ks -H "Authorization: Bearer token" https://rancher.url/v3/endpoint | jq '.'
   ```

2. **Common gotchas:**
   - `sort` field is NOT `map[string]string`, it's a nested struct
   - `version` field in clusters is NOT a string, it's a version object
   - Many fields are optional and should use `omitempty` tags
   - Project IDs have format `c-xxxxx:p-yyyyy` (cluster:project)

3. **Type definition pattern:**
   ```go
   type Collection struct {
       Type         string                   `json:"type"`
       ResourceType string                   `json:"resourceType"`
       Sort         *Sort                    `json:"sort,omitempty"`
       // ... use pointers for optional complex objects
   }
   ```

### Bubble Tea Patterns Used
- Use `tea.Batch()` to return multiple commands
- Keep async operations (API calls) as commands that return messages
- Table updates trigger on both data changes and window resize
- `lipgloss.JoinVertical()` for layout composition
- Alt screen mode for clean TUI experience

### Bubble-Table (evertras/bubble-table)
Current version: v0.15.2 (not v0.16.3 which doesn't exist)

Usage pattern:
```go
columns := []table.Column{
    table.NewColumn("key", "HEADER", width),
}
rows := []table.Row{
    table.NewRow(table.RowData{"key": value}),
}
t := table.New(columns).
    WithRows(rows).
    HeaderStyle(headerStyle).
    Focused(true)
```

### Color Scheme (k9s-inspired)
- **Cyan**: Headers, breadcrumbs, highlights
- **Green**: Running/active/healthy resources
- **Yellow**: Pending/provisioning states
- **Red**: Failed/error states
- **Gray**: Completed/terminated resources

## Common Development Tasks

### Adding a New Resource View
1. Add API types to `internal/rancher/types.go` (inspect JSON first!)
2. Add client methods to `internal/rancher/client.go`
3. Create view in `internal/tui/views/` (future structure)
4. Add message types for async data fetching
5. Add command mode handler (e.g., `:pods`)
6. Add navigation from parent view (e.g., from namespace list)

### Adding a New Action (d, e, l, s, etc.)
1. Add key binding to `Update()` message handler
2. Create action handler that returns a `tea.Cmd`
3. For actions requiring external tools (kubectl):
   - Generate kubeconfig from Rancher API
   - Execute kubectl command
   - Clean up temp files

### Testing Against Rancher Instance
1. Ensure config has valid credentials and `insecure: true` for test instance
2. Build with `make build`
3. Run `./bin/r9s` in terminal
4. Use `q` to quit, `r` to refresh, `j/k` to navigate

### Debugging JSON Unmarshaling Issues
If you see "json: cannot unmarshal..." errors:
1. Capture raw API response: `curl -ks -H "Authorization: Bearer token" URL | jq '.'`
2. Compare with Go struct definitions in `types.go`
3. Check field types (string vs object vs array)
4. Verify all optional fields use `omitempty` tags
5. Use pointers for optional complex types

### Implementing CRD Browser
**Overview**: The CRD Browser lets users discover, understand, and interact with Custom Resource Definitions.

**API Endpoints:**
```
GET /v3/clusters/{clusterID}/customresourcedefinitions  # List CRDs
GET /v3/clusters/{clusterID}/customresourcedefinitions/{crdName}  # Get CRD details
GET /v3/clusters/{clusterID}/customresources/{group}/{version}/{kind}  # List CR instances
```

**Data Structures:**
```go
type CRD struct {
    ID          string              `json:"id"`
    Name        string              `json:"name"`
    Group       string              `json:"spec.group"`
    Kind        string              `json:"spec.names.kind"`
    Plural      string              `json:"spec.names.plural"`
    Scope       string              `json:"spec.scope"` // Namespaced or Cluster
    Versions    []CRDVersion        `json:"spec.versions"`
    Created     time.Time           `json:"created"`
}

type CRDVersion struct {
    Name    string        `json:"name"`
    Served  bool          `json:"served"`
    Storage bool          `json:"storage"`
    Schema  *OpenAPIV3Schema `json:"schema.openAPIV3Schema"`
}

type OpenAPIV3Schema struct {
    Type        string                      `json:"type"`
    Properties  map[string]SchemaProperty   `json:"properties"`
    Required    []string                    `json:"required"`
    Description string                      `json:"description"`
}

type SchemaProperty struct {
    Type        string                    `json:"type"`
    Description string                    `json:"description"`
    Format      string                    `json:"format,omitempty"`
    Properties  map[string]SchemaProperty `json:"properties,omitempty"`
    Items       *SchemaProperty           `json:"items,omitempty"`
    Enum        []interface{}             `json:"enum,omitempty"`
    Minimum     *float64                  `json:"minimum,omitempty"`
    Maximum     *float64                  `json:"maximum,omitempty"`
}
```

**UI Flow:**
1. **CRD List View**
   - Columns: Name, Group, Kind, Scope, Versions, Age
   - Navigate with j/k, Enter to view details
   
2. **CRD Details View**
   - Show Group/Version/Kind header
   - Display schema in tree format
   - Highlight required fields
   - Show field types and descriptions
   - Actions: `i` to list instances, `e` for example YAML
   
3. **CR Instances View**
   - List all instances of the CRD
   - Columns: Name, Namespace (if namespaced), Status, Age
   - Actions: `d` describe, `e` edit, `Del` delete

**Schema Display Format:**
```
apiVersion: apps.cattle.io/v1
kind: App
metadata:
  name: string (required)
  namespace: string (required)
spec:
  chart: string
    Chart name from catalog
  version: string
    Chart version
  targetNamespace: string
    Namespace to install chart into
  values: object
    Helm values override
```

**Implementation Steps:**
1. Add CRD types to `internal/rancher/types.go`
2. Add CRD client methods to `internal/rancher/client.go`
3. Create `ViewCRDs` and `ViewCRInstances` view types
4. Add CRD table rendering functions
5. Add CRD schema parser and tree renderer
6. Add `:crds` command mode handler
7. Add CRD actions (describe, list instances, create/edit/delete CRs)

## Dependencies

- Go 1.25+ (1.23+ will work)
- `github.com/charmbracelet/bubbletea` v1.2.4 - TUI framework
- `github.com/charmbracelet/lipgloss` v1.0.0 - Terminal styling
- `github.com/evertras/bubble-table` v0.15.2 - Table component
- `github.com/spf13/cobra` v1.8.1 - CLI framework
- `github.com/spf13/viper` v1.19.0 - Config management (not yet used)

## Version Information
Version info is injected at build time via ldflags in the Makefile:
- `main.version` - Version string (default: "dev")
- `main.commit` - Git commit hash
- `main.date` - Build timestamp

## Future Implementation Phases
See STATUS.md for detailed implementation roadmap. Key upcoming features:

### Phase 4: Additional Resource Views
- Deployments, Services, ConfigMaps, Secrets, Nodes, PVCs, Ingresses

### Phase 5: CRD Browser (NEW FEATURE)
**Purpose**: Browse and interact with Custom Resource Definitions and their instances

**Features:**
- **CRD Discovery**: List all CRDs in the cluster with group, version, kind
- **Schema Viewer**: Display OpenAPI v3 schema with field types and descriptions
- **CRD Explanation**: AI-powered or annotation-based descriptions of what each CRD does
- **Instance Browser**: List all instances of a selected CRD type
- **Interaction Guide**: Show how to create/modify/delete CR instances
- **Rancher CRDs**: Special handling for cattle.io, fleet.cattle.io resources

**Implementation Notes:**
- Access via `:crds` command mode or dedicated view
- Parse OpenAPI schema from CRD spec.versions[].schema
- Extract field descriptions from schema.properties
- Show required vs optional fields
- Display validation rules (min/max, enum, pattern)
- Provide example manifests
- Link to related documentation

### Phase 6: Actions (describe, edit, delete, logs, exec, port-forward)
### Phase 7: Command mode (`:pods`, `:deployments`, `:crds`, etc.) and filter mode (`/`)
### Phase 8: Real-time updates via WebSocket or polling
### Phase 9: Rancher-specific features (Catalog Apps, Multi-Cluster Apps, Fleet)
