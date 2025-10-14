# Makefile for bom-dagger

BINARY_NAME := bom-dagger
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Build flags
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

# Directories
DIST_DIR := dist
CMD_DIR := cmd/bom-dagger

.PHONY: all build clean test coverage run help install fmt vet lint

## help: Display this help message
help:
	@echo "Available targets:"
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/^## //'

## all: Build for all platforms
all: clean build-all

## build: Build the binary for current platform
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(CMD_DIR)/main.go

## build-all: Build for all platforms
build-all: clean
	@mkdir -p $(DIST_DIR)
	@echo "Building for Linux AMD64..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)/main.go
	@echo "Building for Linux ARM64..."
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)/main.go
	@echo "Building for Darwin ARM64 (Apple Silicon)..."
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)/main.go
	@echo "Build complete! Binaries in $(DIST_DIR)/"

## install: Install the binary to /usr/local/bin
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo mv $(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete!"

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -f $(BINARY_NAME)
	@rm -rf $(DIST_DIR)
	@rm -f coverage.txt coverage.out
	@echo "Clean complete!"

## test: Run all tests
test:
	$(GOTEST) -v ./...

## test-race: Run tests with race detector
test-race:
	$(GOTEST) -race -v ./...

## coverage: Run tests with coverage
coverage:
	$(GOTEST) -race -coverprofile=coverage.txt -covermode=atomic ./...
	@echo "Coverage report generated: coverage.txt"
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "HTML coverage report: coverage.html"

## fmt: Format all Go files
fmt:
	@echo "Formatting code..."
	@go fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

## lint: Run golangci-lint (must be installed)
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install from https://golangci-lint.run/usage/install/" && exit 1)
	@golangci-lint run

## deps: Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

## run: Run with example SBOM
run: build
	./$(BINARY_NAME) -i example-sbom.json

## example: Run various examples
example: build
	@echo "Simple example:"
	@./$(BINARY_NAME) -i testdata/sboms/simple-1.6.json
	@echo "\nServices example:"
	@./$(BINARY_NAME) -i testdata/sboms/services-1.6.json
	@echo "\nMicroservices example (groups):"
	@./$(BINARY_NAME) -i testdata/sboms/microservices-1.6.json -g | head -20

## release-dry: Test release process locally
release-dry: clean
	@echo "Simulating release for version $(VERSION)..."
	@mkdir -p $(DIST_DIR)
	@echo "Building Linux AMD64..."
	@GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)/main.go
	@cd $(DIST_DIR) && tar czf $(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && rm $(BINARY_NAME)-linux-amd64
	@echo "Building Linux ARM64..."
	@GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)/main.go
	@cd $(DIST_DIR) && tar czf $(BINARY_NAME)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64 && rm $(BINARY_NAME)-linux-arm64
	@echo "Building Darwin ARM64 (Apple Silicon)..."
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)/main.go
	@cd $(DIST_DIR) && tar czf $(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 && rm $(BINARY_NAME)-darwin-arm64
	@echo "Release artifacts in $(DIST_DIR)/"
	@ls -lh $(DIST_DIR)/

## release: Create a new release tag and push it (requires TAG env var)
release:
	@if [ -z "$$TAG" ]; then \
		echo "Error: TAG environment variable is not set"; \
		echo "Usage: make release TAG=v1.0.0"; \
		exit 1; \
	fi
	@echo "Preparing release for tag: $$TAG"
	@echo "Checking if tag already exists..."
	@if git rev-parse $$TAG >/dev/null 2>&1; then \
		echo "Error: Tag $$TAG already exists"; \
		exit 1; \
	fi
	@echo "Checking for uncommitted changes..."
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Error: There are uncommitted changes. Please commit or stash them first."; \
		git status --short; \
		exit 1; \
	fi
	@echo "Ensuring we're on main branch..."
	@current_branch=$$(git rev-parse --abbrev-ref HEAD); \
	if [ "$$current_branch" != "main" ]; then \
		echo "Error: Not on main branch (currently on $$current_branch)"; \
		echo "Please checkout main branch first: git checkout main"; \
		exit 1; \
	fi
	@echo "Pulling latest changes from origin..."
	@git pull origin main
	@echo "Running tests before release..."
	@$(GOTEST) ./... > /dev/null 2>&1 || (echo "Tests failed! Aborting release." && exit 1)
	@echo "✓ All tests passed"
	@echo "Creating tag $$TAG..."
	@git tag -a $$TAG -m "Release $$TAG"
	@echo "✓ Tag created successfully"
	@echo "Pushing tag to origin..."
	@git push origin $$TAG
	@echo "✓ Tag pushed successfully"
	@echo ""
	@echo "🎉 Release $$TAG created successfully!"
	@echo ""
	@echo "GitHub Actions will now:"
	@echo "  1. Build binaries for all platforms"
	@echo "  2. Create a GitHub release"
	@echo "  3. Upload the binaries as release assets"
	@echo ""
	@echo "Monitor the progress at:"
	@echo "  https://github.com/nprimmer/bom-dagger/actions"
	@echo ""
	@echo "Once complete, the release will be available at:"
	@echo "  https://github.com/nprimmer/bom-dagger/releases/tag/$$TAG"

# Default target
.DEFAULT_GOAL := help