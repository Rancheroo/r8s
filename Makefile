.PHONY: build install test clean run fmt vet tidy coverage help lint ci dev

# Build variables
BINARY_NAME=r8s
BUILD_DIR=bin
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the r8s binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

install: ## Install r8s to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS)
	@echo "Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

run: ## Run r8s directly without building
	go run $(LDFLAGS) main.go

test: ## Run all tests
	go test -v -race -p 1 ./...

fmt: ## Format Go code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

tidy: ## Tidy go.mod
	go mod tidy

clean: ## Remove build artifacts and coverage files
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "Cleaned build directory and coverage files"

coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1
	@echo "Coverage report: coverage.out"
	@echo "View HTML: go tool cover -html=coverage.out -o coverage.html"

lint: ## Run golangci-lint (Sprint 6 #44)
	@echo "Running linter..."
	@which golangci-lint > /dev/null 2>&1 || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run --timeout=5m ./...

ci: lint test coverage ## Run full CI pipeline locally (Sprint 6)
	@echo "✓ CI checks complete"

dev: tidy fmt vet lint ## Run development checks (tidy, fmt, vet, lint)
	@echo "Development checks complete"

check-sync: ## Check if local branch is behind origin
	@git fetch origin --quiet 2>/dev/null || true
	@if [ -n "$(shell git log HEAD..origin/$(shell git branch --show-current) --oneline 2>/dev/null)" ]; then \
		echo "\033[31m✗ ERROR: Local branch is behind origin\033[0m"; \
		echo "Run: git reset --hard origin/$(shell git branch --show-current)"; \
		echo "Or:  make sync"; \
		exit 1; \
	else \
		echo "\033[32m✓ Branch is up to date with origin\033[0m"; \
	fi

sync: ## Reset local branch to match origin (use with caution)
	@echo "Fetching origin..."
	@git fetch origin
	@echo "Resetting to origin/$(shell git branch --show-current)..."
	@git reset --hard origin/$(shell git branch --show-current)
	@echo "\033[32m✓ Synced with origin\033[0m"

.DEFAULT_GOAL := help
