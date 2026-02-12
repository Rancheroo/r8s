.PHONY: build install test clean run fmt vet tidy coverage help

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
	go test -v -race ./...

fmt: ## Format Go code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

tidy: ## Tidy go.mod
	go mod tidy

clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)
	@echo "Cleaned build directory"

coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1
	@echo "Coverage report: coverage.out"
	@echo "View HTML: go tool cover -html=coverage.out -o coverage.html"

dev: tidy fmt vet ## Run development checks (tidy, fmt, vet)
	@echo "Development checks complete"

.DEFAULT_GOAL := help
