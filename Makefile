.PHONY: help build test test-coverage test-verbose lint fmt vet security clean install-tools all check ci

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME=gomgr
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X github.com/DragonSecurity/github-org-manager-go/internal/version.Version=$(VERSION)"
BUILD_DIR=build
COVERAGE_DIR=coverage

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOFMT=gofmt
GOMOD=$(GOCMD) mod

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

all: clean fmt vet lint test build ## Run all checks and build

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -trimpath $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

install: ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(LDFLAGS) .

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -covermode=atomic -coverprofile=$(COVERAGE_DIR)/coverage.out ./internal/config ./internal/sync
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report: $(COVERAGE_DIR)/coverage.html"
	@$(GOCMD) tool cover -func=$(COVERAGE_DIR)/coverage.out | tail -1

test-verbose: ## Run tests with verbose output
	@echo "Running verbose tests..."
	$(GOTEST) -v -race ./...

test-short: ## Run short tests only
	@echo "Running short tests..."
	$(GOTEST) -short ./...

fmt: ## Format Go code
	@echo "Formatting code..."
	@$(GOFMT) -w -s .
	@echo "Code formatted"

fmt-check: ## Check if code is formatted
	@echo "Checking code formatting..."
	@test -z "$$($(GOFMT) -l .)" || (echo "Code is not formatted. Run 'make fmt'" && exit 1)
	@echo "Code is properly formatted"

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOVET) ./...
	@echo "go vet passed"

lint: ## Run golangci-lint (requires golangci-lint to be installed)
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout 5m ./...; \
		echo "Linting passed"; \
	elif [ -x "$(shell go env GOPATH)/bin/golangci-lint" ]; then \
		$(shell go env GOPATH)/bin/golangci-lint run --timeout 5m ./...; \
		echo "Linting passed"; \
	else \
		echo "golangci-lint not found. Install it with: make install-tools"; \
		exit 1; \
	fi

security: ## Run security checks with gosec (requires gosec to be installed)
	@echo "Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec -quiet ./...; \
		echo "Security check passed"; \
	elif [ -x "$(shell go env GOPATH)/bin/gosec" ]; then \
		$(shell go env GOPATH)/bin/gosec -quiet ./...; \
		echo "Security check passed"; \
	else \
		echo "gosec not found. Install it with: make install-tools"; \
		exit 1; \
	fi

mod-tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	$(GOMOD) tidy
	@echo "Modules tidied"

mod-verify: ## Verify go modules
	@echo "Verifying go modules..."
	$(GOMOD) verify
	@echo "Modules verified"

mod-download: ## Download go modules
	@echo "Downloading go modules..."
	$(GOMOD) download
	@echo "Modules downloaded"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR) $(COVERAGE_DIR)
	@rm -f $(BINARY_NAME)
	@echo "Cleaned"

install-tools: ## Install development tools (golangci-lint, gosec)
	@echo "Installing development tools..."
	@echo "Installing golangci-lint..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	else \
		echo "golangci-lint already installed"; \
	fi
	@echo "Installing gosec..."
	@if ! command -v gosec >/dev/null 2>&1; then \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	else \
		echo "gosec already installed"; \
	fi
	@echo "All tools installed"

check: fmt-check vet test ## Run basic checks (format, vet, test)

check-all: fmt-check vet lint security test ## Run all checks including lint and security

ci: clean check build ## Run CI pipeline (clean, run basic checks, build)

version: ## Display version information
	@echo "Version: $(VERSION)"

.PHONY: mod-tidy mod-verify mod-download version
