# Contributing to r8s

Thank you for your interest in contributing to r8s! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Environment](#development-environment)
- [Code Style](#code-style)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Commit Guidelines](#commit-guidelines)
- [Architecture Overview](#architecture-overview)

---

## Getting Started

### Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.23 or higher** - [Download Go](https://go.dev/dl/)
- **Git** - For version control
- **Make** - For build automation (optional but recommended)
- **Access to a Rancher instance** - For testing (can use offline mode for development)

### Clone the Repository

```bash
git clone https://github.com/Rancheroo/r8s.git
cd r8s
```

### Install Dependencies

```bash
go mod download
```

### Build the Project

```bash
# Using Make
make build

# Or directly with Go
go build -o bin/r8s main.go
```

### Run the Application

```bash
# Run from source
go run main.go

# Or run the built binary
./bin/r8s
```

---

## Version Management

r8s uses **git tags** for version management. The version is automatically detected during build and embedded into the binary.

### How It Works

The `Makefile` automatically detects the version from git tags:

```makefile
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
```

This produces versions like:
- `v0.1.0` - Clean tagged version
- `v0.1.0-5-gabcdef` - 5 commits after v0.1.0 tag
- `v0.1.0-dirty` - Uncommitted changes present
- `dev` - No git repository or no tags

### Creating a New Release

1. **Commit all changes**:
   ```bash
   git add .
   git commit -m "feat: description of changes"
   ```

2. **Create a version tag**:
   ```bash
   git tag -a v0.2.0 -m "Release v0.2.0 - Description of changes"
   ```

3. **Build with the new version**:
   ```bash
   make build
   ./bin/r8s version
   # Output: r8s v0.2.0 (commit: abc123, built: 2025-12-01T...)
   ```

4. **Push the tag to remote** (optional):
   ```bash
   git push origin v0.2.0        # Push specific tag
   git push origin --tags        # Push all tags
   ```

### Version Override

You can manually override the version during development:

```bash
make build VERSION=0.2.0-dev
```

### Semantic Versioning

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** (v1.0.0): Breaking API changes
- **MINOR** (v0.2.0): New features, backwards compatible
- **PATCH** (v0.1.1): Bug fixes, backwards compatible

### Pre-release Versions

For pre-release versions, use suffixes:

```bash
git tag -a v0.2.0-beta -m "Beta release for v0.2.0"
git tag -a v0.2.0-rc1 -m "Release candidate 1 for v0.2.0"
```

---

## Development Environment

### Recommended IDE Setup

- **Visual Studio Code** with Go extension
- **GoLand** by JetBrains
- **Vim/Neovim** with gopls

### VS Code Extensions

- Go (golang.go)
- Go Test Explorer
- Go Doc

### Configuration

Create or edit `~/.r8s/config.yaml`:

```yaml
current_profile: dev
profiles:
  - name: dev
    url: https://rancher-dev.example.com
    bearer_token: your-token-here
    insecure: true  # For development only
```

---

## Code Style

### Go Conventions

We follow standard Go conventions and idioms:

1. **Formatting**: Use `gofmt` (automatically applied)
2. **Linting**: Code should pass `go vet`
3. **Naming**: 
   - Use camelCase for private members
   - Use PascalCase for exported members
   - Use descriptive names (avoid single-letter variables except in loops)

### Code Organization

```
r8s/
├── cmd/              # CLI commands
│   └── root.go      # Root command setup
├── internal/        # Private application code
│   ├── config/     # Configuration management
│   ├── rancher/    # Rancher API client
│   ├── tui/        # Terminal UI components
│   └── k8s/        # Kubernetes operations
├── docs/           # Documentation
└── main.go         # Application entry point
```

### Documentation

- **All exported functions** must have godoc comments
- **Package-level comments** at the top of each package
- **Inline comments** for complex logic

Example:

```go
// Package tui implements the terminal user interface.
// It provides interactive navigation through Rancher resources.
package tui

// NewApp creates a new TUI application instance.
// It initializes the Rancher client and sets up the initial view.
func NewApp(cfg *config.Config) *App {
    // Implementation
}
```

---

## Testing

### Running Tests

```bash
# Run all tests with race detection
make test

# Or directly
go test -v -race ./...

# Run specific package tests
go test -v -race ./internal/config
go test -v -race ./internal/rancher

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Writing Tests

- **Table-driven tests** are preferred
- **Test file naming**: `*_test.go`
- **Test function naming**: `TestFunctionName` or `TestType_Method`

Example:

```go
func TestConfig_Validate(t *testing.T) {
    tests := []struct {
        name    string
        config  *Config
        wantErr bool
    }{
        {
            name: "valid config",
            config: &Config{
                CurrentProfile: "test",
                Profiles: []Profile{{Name: "test"}},
            },
            wantErr: false,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Test Coverage Goals

- **Target**: 80% coverage for non-UI code
- **Current**: ~65%
- All new code should include tests
- Critical paths must have >90% coverage

---

## Pull Request Process

### Before Submitting

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Write tests** for your changes

3. **Run tests** and ensure they pass:
   ```bash
   make test
   ```

4. **Format code**:
   ```bash
   go fmt ./...
   ```

5. **Lint code**:
   ```bash
   go vet ./...
   ```

6. **Update documentation** if needed

### Submitting a Pull Request

1. **Push your branch** to the repository
2. **Create a Pull Request** with:
   - Clear description of changes
   - Link to related issues
   - Screenshots for UI changes
3. **Ensure CI passes** (tests, linting)
4. **Request review** from maintainers
5. **Address feedback** and update PR

### PR Checklist

- [ ] Tests added/updated and passing
- [ ] Code formatted with `go fmt`
- [ ] Code passes `go vet`
- [ ] Documentation updated
- [ ] CHANGELOG.md updated (if applicable)
- [ ] No race conditions (`go test -race` passes)
- [ ] Commit messages follow guidelines

---

## Commit Guidelines

### Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>: <description>

[optional body]

[optional footer]
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Maintenance tasks
- `style`: Code style changes (formatting, etc.)

### Examples

```bash
# Feature
git commit -m "feat: add describe support for deployments"

# Bug fix
git commit -m "fix: correct deployment replica count display"

# Documentation
git commit -m "docs: update README with offline mode info"

# Test
git commit -m "test: add unit tests for config validation"

# Multiple commits
git commit -m "feat: implement filter mode

- Add filter input handling
- Implement live filtering
- Add regex pattern support"
```

---

## Architecture Overview

### High-Level Architecture

```
┌─────────────────────────────────────────────┐
│           Terminal (User Input)             │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│        Bubble Tea Framework (TUI)           │
│  ┌────────────────────────────────────────┐ │
│  │  App (State Management)                │ │
│  │  - viewStack    - currentView          │ │
│  │  - clusters     - deployments          │ │
│  │  - pods         - services             │ │
│  └─────────────┬──────────────────────────┘ │
└────────────────┼────────────────────────────┘
                 │
┌────────────────▼────────────────────────────┐
│      Rancher API Client                     │
│  - Authentication (Bearer Token)            │
│  - Resource Fetching (Clusters, Pods, etc.) │
│  - Error Handling & Retry Logic             │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│         Rancher API / Kubernetes            │
└─────────────────────────────────────────────┘
```

### Key Components

1. **TUI (internal/tui/)**
   - `app.go`: Main application state and logic
   - `styles.go`: Visual styling
   - Uses Bubble Tea for event-driven architecture

2. **Rancher Client (internal/rancher/)**
   - `client.go`: API client implementation
   - `types.go`: Data structures matching Rancher API

3. **Config (internal/config/)**
   - `config.go`: Configuration management
   - Handles multiple profiles

### State Flow

1. User presses key → Bubble Tea Update()
2. Update() processes key → triggers command (e.g., fetchPods)
3. Command executes → returns message
4. Message processed → updates app state
5. App re-renders with new state

---

## Common Tasks

### Adding a New Resource Type

1. Add struct to `internal/rancher/types.go`
2. Add fetch method to `internal/rancher/client.go`
3. Add view case to `internal/tui/app.go`
4. Update `updateTable()` to render the resource
5. Add tests

### Adding a New Key Binding

1. Add case to `Update()` in `app.go`
2. Implement handler function
3. Update help screen
4. Document in README

### Debugging

```bash
# Enable verbose logging
export LOG_LEVEL=debug
./bin/r8s

# Run with race detector
go run -race main.go

# Profile memory
go test -memprofile=mem.prof ./internal/tui
go tool pprof mem.prof
```

---

## Communication

- **Issues**: Report bugs or request features via GitHub Issues
- **Discussions**: Use GitHub Discussions for questions
- **Code Review**: All PRs require at least one review

---

## License

By contributing to r8s, you agree that your contributions will be licensed under the Apache License 2.0.

---

## Questions?

If you have questions about contributing, please:
1. Check existing documentation
2. Search closed issues
3. Open a new discussion
4. Ask in pull request comments

Thank you for contributing to r8s! 🎉
