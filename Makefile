# Firefly - Infrastructure Drift Detection Tool
# Makefile for build automation

# Variables
BINARY_NAME=firefly
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_WINDOWS=$(BINARY_NAME).exe
MAIN_PACKAGE=./cmd/firefly
GO_FILES=$(shell find . -name '*.go' -not -path './vendor/*')

# Default target
.PHONY: all
all: clean test build

# Build the application (Windows default)
.PHONY: build
build:
	@echo "Building $(BINARY_WINDOWS) for Windows..."
	@set GOOS=windows&& set GOARCH=amd64&& go build -o $(BINARY_WINDOWS) $(MAIN_PACKAGE)

# Build for Windows
.PHONY: build-windows
build-windows:
	@echo "Building $(BINARY_WINDOWS) for Windows..."
	@set GOOS=windows&& set GOARCH=amd64&& go build -o $(BINARY_WINDOWS) $(MAIN_PACKAGE)

# Build for Linux
.PHONY: build-linux
build-linux:
	@echo "Building $(BINARY_UNIX) for Linux..."
	@set GOOS=linux&& set GOARCH=amd64&& go build -o $(BINARY_UNIX) $(MAIN_PACKAGE)

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detection
.PHONY: test-race
test-race:
	@echo "Running tests with race detection..."
	go test -v -race ./...

# Lint the code
.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		go vet ./...; \
	fi

# Format the code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Tidy dependencies
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@if exist $(BINARY_WINDOWS) del $(BINARY_WINDOWS)
	@if exist $(BINARY_UNIX) del $(BINARY_UNIX)
	@if exist coverage.out del coverage.out
	@if exist coverage.html del coverage.html

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download

# Run the application
.PHONY: run
run: build
	@echo "Running $(BINARY_WINDOWS)..."
	.\$(BINARY_WINDOWS)

# Development setup
.PHONY: dev-setup
dev-setup:
	@echo "Setting up development environment..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go mod download

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          - Clean, test, and build"
	@echo "  build        - Build the application"
	@echo "  build-windows- Build for Windows"
	@echo "  build-linux  - Build for Linux"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage"
	@echo "  test-race    - Run tests with race detection"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"
	@echo "  tidy         - Tidy dependencies"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Install dependencies"
	@echo "  run          - Build and run the application"
	@echo "  dev-setup    - Set up development environment"
	@echo "  help         - Show this help message"