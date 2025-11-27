# r8s (Rancheroos)

> A terminal UI for navigating and managing Rancher-based Kubernetes clusters

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)

r8s (pronounced "rancheros", a play on k9s and k8s) provides a fast, keyboard-driven terminal interface for browsing Rancher clusters, projects, namespaces, and Kubernetes resources. Inspired by k9s but designed specifically for Rancher's multi-cluster management model.

---

## 📸 Demo

_[Screenshot coming soon - showing TUI with cluster navigator and describe modal]_

---

## ✨ Features

### Currently Available

- ✅ **Browse Rancher Hierarchy**: Clusters → Projects → Namespaces → Resources
- ✅ **View Multiple Resource Types**: Pods, Deployments, Services, CRDs
- ✅ **Describe Resources**: Press `d` to view detailed JSON for Pods/Deployments/Services
- ✅ **CRD Explorer**: Browse Custom Resource Definitions and their instances
- ✅ **Offline Mode**: automatic fallback to mock data for development/demos
- ✅ **Fast Navigation**: Keyboard shortcuts for efficient browsing
- ✅ **Multiple Profiles**: Switch between different Rancher environments

### Current Limitations

- Read-only (no resource modification)
- No log viewing or pod exec yet
- No real-time watching (manual refresh only)
- Describe limited to Pods, Deployments, Services

---

## 📦 Installation

### Prerequisites

- **Go 1.23 or higher** - [Install Go](https://go.dev/dl/)
- **Access to Rancher** (v2.7+ recommended)
- **Rancher API credentials** (Bearer token or API key/secret)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/Rancheroo/r8s.git
cd r8s

# Build
make build

# Binary will be in: ./bin/r8s
```

### Quick Install

```bash
# Install dependencies
go mod download

# Build and run
go build -o bin/r8s main.go
./bin/r8s
```

---

## 🚀 Quick Start

### 1. First Run - Create Configuration

On first run, r8s creates a default config template:

```bash
./bin/r8s
# Creates ~/.r8s/config.yaml
```

### 2. Get Rancher API Credentials

#### Option A: Create API Key in Rancher UI

1. Log in to your Rancher UI
2. Click your avatar (top right) → **Account & API Keys**
3. Click **Create API Key**
4. Set a description (e.g., "r8s CLI access")
5. Set expiration (optional) and scope (optional)
6. Click **Create**
7. **Copy both keys immediately** (won't be shown again):
   - Access Key (e.g., `token-xxxxx`)
   - Secret Key (e.g., `long-secret-string`)
   - Or Bearer Token (pre-concatenated: `token-xxxxx:secret`)

#### Option B: Use Existing Bearer Token

If you already have a bearer token, use that directly.

### 3. Configure r8s

Edit `~/.r8s/config.yaml`:

```yaml
current_profile: production
profiles:
  # Using Bearer Token (recommended)
  - name: production
    url: https://rancher.example.com
    bearer_token: token-xxxxx:your-secret-here
    insecure: false  # Set true to skip TLS verification

  # OR using separate Access Key + Secret Key
  - name: development
    url: https://rancher-dev.example.com
    access_key: token-xxxxx
    secret_key: your-secret-here
    insecure: true  # For dev with self-signed certs

  # Additional profiles as needed
  - name: staging
    url: https://rancher-staging.example.com
    bearer_token: token-yyyyy:staging-secret
```

### 4. Launch r8s

```bash
# Start with default profile
./bin/r8s

# Or specify a profile
./bin/r8s --profile development
```

---

## ⌨️ Key Bindings

### Navigation

| Key | Action |
|-----|--------|
| `↑` / `k` | Move selection up |
| `↓` / `j` | Move selection down |
| `Enter` | Navigate into selected resource |
| `Esc` | Navigate back / Exit modal |
| `q` | Quit application |
| `Ctrl+C` | Quit application |

### Views & Actions

| Key | Action | Available In |
|-----|--------|--------------|
| `d` | Describe selected resource (show JSON details) | Pods, Deployments, Services |
| `1` | Switch to Pods view | Namespace resources |
| `2` | Switch to Deployments view | Namespace resources |
| `3` | Switch to Services view | Namespace resources |
| `C` | Jump to CRDs view | Clusters, Projects |
| `i` | Toggle CRD description | CRDs view |

### Utility

| Key | Action |
|-----|--------|
| `r` / `Ctrl+R` | Refresh current view |
| `?` | Show help screen |

---

## 📚 Usage Guide

### Navigation Hierarchy

r8s follows Rancher's natural hierarchy:

```
Clusters
  ↓ (Enter)
Projects
  ↓ (Enter)
Namespaces
  ↓ (Enter)
Resources (Pods / Deployments / Services)
```

**Example Navigation Flow:**

```bash
1. Start r8s
2. See list of clusters
3. Press ↓ to select "production-cluster"
4. Press Enter → Now viewing projects in that cluster
5. Press ↓ to select "default-project"
6. Press Enter → Now viewing namespaces in that project
7. Press ↓ to select "kube-system"
8. Press Enter → Now viewing pods in that namespace
9. Press 2 → Switch to deployments in same namespace
10. Press d → View deployment details in JSON modal
11. Press Esc → Close modal
12. Press Esc → Back to namespaces
13. Press Esc → Back to projects
14. Press Esc → Back to clusters
```

### Resource Views

When viewing a namespace, you can switch between resource types:

- **Press `1`**: View Pods
- **Press `2`**: View Deployments
- **Press `3`**: View Services

Each view shows relevant information:

**Pods:**
```
NAME                              NAMESPACE     STATE      NODE
nginx-deployment-abc123-xyz       default       Running    worker-1
api-server-def456-uvw             default       Running    worker-2
```

**Deployments:**
```
NAME               NAMESPACE   READY   UP-TO-DATE   AVAILABLE
nginx-deployment   default     3/3     3            3
api-server         default     2/2     2            2
```

**Services:**
```
NAME            NAMESPACE   TYPE        CLUSTER-IP      PORT(S)
nginx-service   default     ClusterIP   10.43.100.50    80/TCP
api-service     default     NodePort    10.43.100.51    8080:30080/TCP
```

### Describe Feature

Press `d` on any Pod, Deployment, or Service to view full details:

```json
{
  "apiVersion": "v1",
  "kind": "Pod",
  "metadata": {
    "name": "nginx-deployment-abc123-xyz",
    "namespace": "default"
  },
  "spec": { ... },
  "status": { ... }
}
```

Press `Esc`, `d`, or `q` to close the describe modal.

### CRD Explorer

Custom Resource Definitions can be browsed:

1. From Clusters or Projects view, press `C`
2. Browse available CRDs
3. Press `i` to toggle description
4. Press `Enter` to view CRD instances

### Offline Mode

If r8s can't connect to Rancher, it automatically enters offline mode:

```
⚠️  OFFLINE MODE - DISPLAYING MOCK DATA  ⚠️
```

This allows:
- Testing the UI without Rancher access
- Development and debugging
- Feature demonstrations
- Learning the interface

**Note:** Offline mode uses realistic mock data, clearly labeled.

---

## 🔧 Configuration

### Config File Location

- **Default**: `~/.r8s/config.yaml`
- **Override**: `r8s --config /path/to/config.yaml`

### Multiple Profiles

Manage multiple Rancher environments:

```yaml
current_profile: production
profiles:
  - name: production
    url: https://rancher-prod.example.com
    bearer_token: prod-token-here
    
  - name: staging
    url: https://rancher-staging.example.com
    bearer_token: staging-token-here
    
  - name: development
    url: https://rancher-dev.example.com
    bearer_token: dev-token-here
    insecure: true
```

Switch profiles:

```bash
r8s --profile staging
r8s --profile development
```

### TLS/SSL Options

```yaml
profiles:
  - name: dev
    url: https://rancher-dev.local
    bearer_token: token
    insecure: true  # Skip TLS certificate verification
```

**⚠️ Security Warning:** Only use `insecure: true` for development environments with self-signed certificates. Never use in production.

---

## 👥 Team Onboarding

### For New Team Members

Welcome! Here's how to get started with r8s:

#### Step 1: Install Go

```bash
# Check if Go is installed
go version

# Should show: go version go1.23 or higher
# If not, install from: https://go.dev/dl/
```

#### Step 2: Clone and Build

```bash
# Clone the repo
git clone https://github.com/Rancheroo/r8s.git
cd r8s

# Install dependencies
go mod download

# Build
make build

# Verify build
./bin/r8s --version
```

#### Step 3: Get Credentials

Ask your team lead for:
1. Rancher URL (e.g., `https://rancher.company.com`)
2. How to generate your API token (see "Get Rancher API Credentials" above)

#### Step 4: Configure

```bash
# r8s will create a config template on first run
./bin/r8s

# Edit the config
vim ~/.r8s/config.yaml   # or use your preferred editor

# Add your credentials
```

#### Step 5: Learn the Basics

1. **Browse around**: Use arrow keys and Enter to navigate
2. **Try describe**: Press `d` on a pod to see details
3. **Practice switching views**: Use 1, 2, 3 keys
4. **Get help**: Press `?` to see all shortcuts

#### Helpful Resources

- [ARCHITECTURE.md](docs/ARCHITECTURE.md) - How r8s works internally
- [CONTRIBUTING.md](CONTRIBUTING.md) - Development guide
- [CHANGELOG.md](CHANGELOG.md) - Version history

---

## 🛠️ Development

### Running Tests

```bash
# Run all tests with race detection
make test

# Run tests for specific package
go test -v -race ./internal/config
go test -v -race ./internal/rancher

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Development Workflow

```bash
# Run from source (no build needed)
go run main.go

# Build for testing
make build

# Format code
go fmt ./...

# Vet code
go vet ./...

# Clean build artifacts
make clean
```

### Project Structure

```
r8s/
├── main.go              # Entry point
├── cmd/root.go          # CLI setup
├── internal/
│   ├── config/         # Configuration management
│   ├── rancher/        # Rancher API client
│   └── tui/            # Terminal UI
├── docs/                # Documentation
└── scripts/             # Helper scripts
```

### Making Changes

1. Create a feature branch: `git checkout -b feature/my-feature`
2. Make your changes
3. Add tests
4. Run tests: `make test`
5. Commit with conventional format: `git commit -m "feat: add new feature"`
6. Push and create PR

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

---

## 📖 Documentation

- [Architecture Guide](docs/ARCHITECTURE.md) - Technical design and patterns
- [Contributing Guide](CONTRIBUTING.md) - How to contribute
- [Changelog](CHANGELOG.md) - Version history and changes

---

## 🐛 Troubleshooting

### "Connection refused" error

**Cause:** Can't reach Rancher API  
**Solutions:**
- Check Rancher URL is correct
- Verify network connectivity
- Check firewall rules
- Try with `insecure: true` if using self-signed cert

### "Authentication failed" error

**Cause:** Invalid or expired token  
**Solutions:**
- Generate new API token in Rancher UI
- Update `~/.r8s/config.yaml` with new token
- Check token hasn't expired

### Offline mode when not expected

**Cause:** r8s couldn't connect on startup  
**Solutions:**
- Check Rancher URL and credentials
- Restart r8s after fixing config
- Look for error details in terminal

### Deployment replica counts show 0/0

**Cause:** Fixed in v0.2.1  
**Solution:** Update to latest version

---

## 🗺️ Roadmap

### v0.3.0 (Planned)
- Command mode (`:` key) for advanced operations
- Filter/search mode (`/` key)
- Scrollable describe modal
- Namespace describe support

### v0.4.0 (Planned)
- CI/CD pipeline
- Enhanced test coverage (80%+)
- Performance optimizations

### v1.0.0 (Future)
- Resource actions (delete, scale, edit)
- Pod logs viewing
- Pod exec/shell
- Real-time watch mode
- Multi-cluster switching

See [PHASE_D_PREPARATION.md](docs/archive/development/PHASE_D_PREPARATION.md) for detailed roadmap.

---

## 📄 License

Apache License 2.0 - See [LICENSE](LICENSE) for details

---

## 🙏 Acknowledgments

- [k9s](https://k9scli.io/) - Inspiration for this project
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Excellent TUI framework
- [Rancher](https://rancher.com/) - Multi-cluster Kubernetes management platform

---

## 📞 Support & Contact

- **Issues**: [GitHub Issues](https://github.com/Rancheroo/r8s/issues)
- **Documentation**: [GitHub Wiki](https://github.com/Rancheroo/r8s/wiki)
- **Contributing**: See [CONTRIBUTING.md](CONTRIBUTING.md)

---

**Made with ❤️ for the Rancher community**

**Current Version**: 0.2.1 (Bug fixes and improvements)  
**Status**: Active Development  
**Test Coverage**: 65% (target: 80%)
