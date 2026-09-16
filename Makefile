.PHONY: help build dev test lint clean run release build-snapshot
.DEFAULT_GOAL := help

# Variables
PROJECT_NAME := autox
MAIN_PKG := ./cmd/autox
BINARY_PATH := ./dist/$(PROJECT_NAME)
VERSION ?= dev
COMMIT := $(shell git rev-parse --short=7 HEAD 2>/dev/null || echo none)
BUILD_DIR := ./dist

# Go build variables
CGO_ENABLED := 0
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# Build flags
LDFLAGS := -s -w \
	-X github.com/ivis-kuroda/weko-autox/internal/cli.versionString=$(VERSION) \
	-X github.com/ivis-kuroda/weko-autox/internal/cli.commitString=$(COMMIT)

help: ## Show this help message
	@echo "$(PROJECT_NAME) - Build targets"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: clean build ## Build development binary (same as build)

build: ## Build development binary for current platform
	@echo "Building $(PROJECT_NAME) for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) go build \
		-ldflags "$(LDFLAGS)" \
		-o $(BINARY_PATH) \
		$(MAIN_PKG)
	@echo "✓ Binary built: $(BINARY_PATH)"

test: ## Run unit tests
	go test ./...

lint: ## Run golangci-lint
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Error: golangci-lint not found. Install with: https://golangci-lint.run/welcome/install/"; \
		exit 127; \
	fi
	golangci-lint run ./...

build-snapshot: ## Build snapshot binaries for all platforms using goreleaser
	@echo "Building snapshot binaries..."
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "Error: goreleaser not found. Install with: go install github.com/goreleaser/goreleaser/v2@latest"; \
		exit 127; \
	fi
	goreleaser build --snapshot --clean

release: ## Build and release using goreleaser
	@echo "Building release..."
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "Error: goreleaser not found. Install with: go install github.com/goreleaser/goreleaser/v2@latest"; \
		exit 127; \
	fi
	goreleaser release --snapshot --clean

run: build ## Build and run the binary
	@$(BINARY_PATH)

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@echo "✓ Clean complete"
