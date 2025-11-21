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
- Phase 4: Project, Namespace, Pod, Workload views
- Phase 5: Actions (describe, edit, delete, logs, exec, port-forward)
- Phase 6: Command mode (`:pods`, `:deployments`, etc.) and filter mode (`/`)
- Phase 7: Real-time updates via WebSocket or polling
- Phase 8: Rancher-specific features (Catalog Apps, Multi-Cluster Apps, Fleet)
