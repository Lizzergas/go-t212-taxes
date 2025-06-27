# T212 Taxes Makefile
# Provides common development and build commands

# Variables
BINARY_NAME=t212-taxes
BUILD_DIR=dist
GO_VERSION=1.21
MAIN_PATH=./cmd/t212-taxes

# Build information
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse HEAD)
DATE ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Build flags
LDFLAGS=-ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Color output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

.PHONY: help
help: ## Show this help message
	@echo '$(BLUE)T212 Taxes Development Commands$(NC)'
	@echo ''
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make $(BLUE)<target>$(NC)\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: clean
clean: ## Clean build artifacts
	@echo '$(YELLOW)Cleaning build artifacts...$(NC)'
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@echo '$(GREEN)Clean complete!$(NC)'

.PHONY: deps
deps: ## Download and verify dependencies
	@echo '$(YELLOW)Downloading dependencies...$(NC)'
	@go mod download
	@go mod verify
	@echo '$(GREEN)Dependencies ready!$(NC)'

.PHONY: build
build: clean ## Build the application
	@echo '$(YELLOW)Building $(BINARY_NAME)...$(NC)'
	@go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)
	@echo '$(GREEN)Build complete: $(BINARY_NAME)$(NC)'

.PHONY: build-all
build-all: clean ## Build for all platforms
	@echo '$(YELLOW)Building for all platforms...$(NC)'
	@mkdir -p $(BUILD_DIR)
	
	# Linux
	@echo '$(BLUE)Building for Linux...$(NC)'
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	
	# macOS
	@echo '$(BLUE)Building for macOS...$(NC)'
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	
	# Windows
	@echo '$(BLUE)Building for Windows...$(NC)'
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	
	@echo '$(GREEN)Multi-platform build complete!$(NC)'
	@ls -la $(BUILD_DIR)/

.PHONY: test
test: ## Run tests
	@echo '$(YELLOW)Running tests...$(NC)'
	@go test ./...
	@echo '$(GREEN)Tests passed!$(NC)'

.PHONY: test-verbose
test-verbose: ## Run tests with verbose output
	@echo '$(YELLOW)Running tests with verbose output...$(NC)'
	@go test -v ./...

.PHONY: test-race
test-race: ## Run tests with race detection
	@echo '$(YELLOW)Running tests with race detection...$(NC)'
	@go test -race ./...
	@echo '$(GREEN)Race tests passed!$(NC)'

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo '$(YELLOW)Running tests with coverage...$(NC)'
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo '$(GREEN)Coverage report generated: coverage.html$(NC)'

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo '$(YELLOW)Running benchmarks...$(NC)'
	@go test -bench=. -benchmem ./...

.PHONY: fmt
fmt: ## Format code
	@echo '$(YELLOW)Formatting code...$(NC)'
	@go fmt ./...
	@echo '$(GREEN)Code formatted!$(NC)'

.PHONY: lint
lint: ## Run linter
	@echo '$(YELLOW)Running linter...$(NC)'
	@golangci-lint run
	@echo '$(GREEN)Linting passed!$(NC)'

.PHONY: lint-fix
lint-fix: ## Run linter with auto-fix
	@echo '$(YELLOW)Running linter with auto-fix...$(NC)'
	@golangci-lint run --fix
	@echo '$(GREEN)Linting and fixes applied!$(NC)'

.PHONY: security
security: ## Run security scan
	@echo '$(YELLOW)Running security scan...$(NC)'
	@gosec ./...
	@echo '$(GREEN)Security scan passed!$(NC)'

.PHONY: check
check: fmt lint test-race security ## Run all checks (format, lint, test, security)
	@echo '$(GREEN)All checks passed!$(NC)'

.PHONY: run
run: build ## Build and run the application
	@echo '$(YELLOW)Running $(BINARY_NAME)...$(NC)'
	@./$(BINARY_NAME)

.PHONY: run-dev
run-dev: ## Run in development mode
	@echo '$(YELLOW)Running in development mode...$(NC)'
	@go run $(MAIN_PATH)

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo '$(YELLOW)Building Docker image...$(NC)'
	@docker build -t $(BINARY_NAME):latest .
	@echo '$(GREEN)Docker image built: $(BINARY_NAME):latest$(NC)'

.PHONY: docker-run
docker-run: docker-build ## Build and run Docker container
	@echo '$(YELLOW)Running Docker container...$(NC)'
	@docker run --rm -it $(BINARY_NAME):latest

.PHONY: install
install: ## Install the application
	@echo '$(YELLOW)Installing $(BINARY_NAME)...$(NC)'
	@go install $(LDFLAGS) $(MAIN_PATH)
	@echo '$(GREEN)$(BINARY_NAME) installed!$(NC)'

.PHONY: uninstall
uninstall: ## Uninstall the application
	@echo '$(YELLOW)Uninstalling $(BINARY_NAME)...$(NC)'
	@rm -f $(shell go env GOPATH)/bin/$(BINARY_NAME)
	@echo '$(GREEN)$(BINARY_NAME) uninstalled!$(NC)'

.PHONY: mod-tidy
mod-tidy: ## Tidy go modules
	@echo '$(YELLOW)Tidying go modules...$(NC)'
	@go mod tidy
	@echo '$(GREEN)Modules tidied!$(NC)'

.PHONY: mod-update
mod-update: ## Update go modules
	@echo '$(YELLOW)Updating go modules...$(NC)'
	@go get -u ./...
	@go mod tidy
	@echo '$(GREEN)Modules updated!$(NC)'

.PHONY: release-check
release-check: clean check build-all ## Prepare for release (run all checks and build)
	@echo '$(GREEN)Release preparation complete!$(NC)'
	@echo '$(BLUE)Built artifacts:$(NC)'
	@ls -la $(BUILD_DIR)/

.PHONY: dev-setup
dev-setup: deps ## Set up development environment
	@echo '$(YELLOW)Setting up development environment...$(NC)'
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo '$(GREEN)Development environment ready!$(NC)'

.PHONY: version
version: ## Show version information
	@echo '$(BLUE)Version Information:$(NC)'
	@echo 'Version: $(VERSION)'
	@echo 'Commit:  $(COMMIT)'
	@echo 'Date:    $(DATE)'
	@echo 'Go:      $(shell go version)'

# Default target
.DEFAULT_GOAL := help 