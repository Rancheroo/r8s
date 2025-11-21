# r9s (Rancher9s)

> A k9s-inspired terminal UI for managing Rancher-based Kubernetes clusters

r9s provides a powerful, keyboard-driven terminal interface for navigating and managing Rancher clusters, projects, namespaces, and Kubernetes resources. Built with the same philosophy as k9s, it's designed specifically for Rancher's multi-cluster management capabilities.

## ✨ Features

- 🎯 **k9s-style navigation** - Vim keybindings, command mode, filter mode
- 🌐 **Multi-cluster support** - Seamlessly manage multiple Rancher clusters
- 📁 **Project-aware** - Navigate Rancher's project hierarchy naturally
- 🔄 **Real-time updates** - Watch resources change in real-time
- 🎨 **Beautiful UI** - Polished terminal interface with color-coded states
- ⚡ **Fast & efficient** - Built in Go with minimal resource usage
- 🔐 **Secure** - Bearer token or API key/secret authentication

## 📦 Installation

### Prerequisites

- **Go 1.25+** (for building from source)
- Access to a Rancher instance (v2.7+)
- Rancher API credentials (Bearer token or API key + secret)

### Install from source

```bash
# Clone the repository
git clone https://github.com/4realtech/r9s.git
cd r9s

# Build and install
make install

# Or just build to ./bin/r9s
make build
```

### Install with go install

```bash
go install github.com/4realtech/r9s@latest
```

## 🚀 Quick Start

### 1. Initial Setup

On first run, r9s will create a default configuration file:

```bash
r9s
```

This creates `~/.r9s/config.yaml` with a template you need to edit.

### 2. Configure Rancher Access

Edit `~/.r9s/config.yaml`:

```yaml
currentProfile: default
profiles:
  - name: default
    url: https://rancher.example.com
    bearerToken: token-xxxxx:yyyyy  # Option 1: Direct bearer token
    # OR
    accessKey: token-xxxxx           # Option 2: API key
    secretKey: yyyyy                 # Option 2: API secret
    insecure: false                  # Skip TLS verification if needed
    
  - name: dev
    url: https://rancher-dev.example.com
    bearerToken: token-aaaaa:bbbbb
    insecure: true

refreshInterval: 5s
logLevel: info
```

#### Creating API Keys in Rancher

1. Log in to your Rancher UI
2. Click your avatar in the top right → **Account & API Keys**
3. Click **Create API Key**
4. Set description and optional expiration/scope
5. Copy the **Access Key** (token-xxxxx) and **Secret Key**
6. Alternatively, copy the **Bearer Token** (pre-concatenated)

### 3. Launch r9s

```bash
r9s                          # Start with default profile
r9s --profile dev           # Use a specific profile
r9s --insecure              # Skip TLS verification
r9s --context my-cluster    # Start in specific cluster
```

## ⌨️ Key Bindings

### Navigation

| Key | Action |
|-----|--------|
| `↑/k` | Move up |
| `↓/j` | Move down |
| `←/h` | Move left (if applicable) |
| `→/l` | Move right (if applicable) |
| `g` | Go to top |
| `G` | Go to bottom |
| `PgUp` | Page up |
| `PgDn` | Page down |

### Actions

| Key | Action |
|-----|--------|
| `Enter` | Navigate into selected resource |
| `d` | Describe resource |
| `e` | Edit resource YAML |
| `Ctrl+d` | Delete resource |
| `l` | View logs (pods) |
| `s` | Shell into container (pods) |
| `p` | Port forward (pods) |

### Modes

| Key | Action |
|-----|--------|
| `:` | Enter command mode |
| `/` | Enter filter mode |
| `Esc` | Exit mode |
| `?` | Show help |
| `q` | Quit |
| `Ctrl+r` | Refresh current view |
| `Ctrl+c` | Show cluster list |

### Commands (Command Mode)

Enter command mode with `:` and type:

- `:clusters` - List clusters
- `:projects` - List projects
- `:namespaces` or `:ns` - List namespaces
- `:pods` - List pods
- `:deployments` - List deployments
- `:services` - List services
- `:configmaps` - List config maps
- `:secrets` - List secrets
- `:ingresses` - List ingresses
- `:apps` - List catalog apps (Rancher-specific)
- `:mca` - List multi-cluster apps
- `:fleet` - List Fleet workspaces
- `:help` - Show help
- `:quit` or `:q` - Exit r9s

## 🎯 Usage Examples

### Navigate cluster hierarchy

```
r9s
↓/j to select cluster → Enter
↓/j to select project → Enter
↓/j to select namespace → Enter
:pods to view pods
```

### Quick resource access

```
:pods             # Jump directly to pods view
/nginx            # Filter for "nginx"
Enter             # Select pod
l                 # View logs
```

### Edit and apply resources

```
:deployments      # List deployments
e                 # Edit selected deployment
                  # Opens $EDITOR with YAML
                  # Save and exit to apply
```

## 🔧 Configuration

### Config File Location

- Default: `~/.r9s/config.yaml`
- Override: `r9s --config /path/to/config.yaml`

### Multiple Profiles

```yaml
profiles:
  - name: production
    url: https://rancher-prod.example.com
    accessKey: token-prod-key
    secretKey: prod-secret
    
  - name: staging
    url: https://rancher-staging.example.com
    accessKey: token-staging-key
    secretKey: staging-secret
```

Switch profiles:
```bash
r9s --profile production
r9s --profile staging
```

### Environment Variables

- `EDITOR` - Editor for YAML editing (default: vim)
- `TERM` - Should be set to support 256 colors

## 🏗️ Architecture

r9s is built with:

- **Go 1.25** - Modern, fast, compiled language
- **Bubble Tea** - Terminal UI framework based on The Elm Architecture
- **Lipgloss** - Terminal styling and layout
- **Bubble-table** - Interactive table component
- **Cobra** - CLI framework
- **Viper** - Configuration management

## 🚧 Project Status

**Current Version:** Alpha / Early Development

This project is actively under development. Core features are being implemented phase by phase:

- [x] Phase 1: Project scaffolding & basic structure
- [x] Phase 2: Configuration & authentication
- [ ] Phase 3: Core TUI framework with cluster navigation
- [ ] Phase 4: Resource views (projects, namespaces, pods, workloads)
- [ ] Phase 5-7: Actions (logs, exec, edit) & polish
- [ ] Phase 8: Real-time updates & watch mechanism
- [ ] Phase 9-12: Advanced features, error handling, testing

## 🤝 Contributing

Contributions are welcome! Please feel free to submit issues, feature requests, or pull requests.

### Development Setup

```bash
# Clone repository
git clone https://github.com/4realtech/r9s.git
cd r9s

# Install dependencies
go mod tidy

# Run development checks
make dev

# Run locally
make run

# Build
make build
```

### Code Structure

```
r9s/
├── cmd/              # CLI commands (Cobra)
├── internal/
│   ├── config/      # Configuration management
│   ├── rancher/     # Rancher API client
│   ├── tui/         # Terminal UI (Bubble Tea)
│   └── k8s/         # Kubernetes operations
├── docs/            # Documentation
└── main.go          # Entry point
```

## 📄 License

Apache License 2.0

## 🙏 Acknowledgments

- [k9s](https://k9scli.io/) - Inspiration for this project
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Excellent TUI framework
- [Rancher](https://rancher.com/) - Multi-cluster Kubernetes management

## 📞 Support

- GitHub Issues: https://github.com/4realtech/r9s/issues
- Documentation: https://github.com/4realtech/r9s/docs

---

**Made with ❤️ for the Rancher community**
