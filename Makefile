# Git Remote MCP Makefile

.PHONY: build test test-verbose test-short benchmark clean lint fmt help

# Default target
all: build test

# Build the application
build:
	@echo "Building git-remote-mcp..."
	@go build -o git-remote-mcp .

# Run all tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with verbose output and race detection
test-verbose:
	@echo "Running verbose tests with race detection..."
	@go test -v -race ./...

# Run short tests (skip performance tests)
test-short:
	@echo "Running short tests..."
	@go test -v -short ./...

# Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	@go test -v -bench=. -benchmem ./...

# Run specific test categories
test-unit:
	@echo "Running unit tests..."
	@go test -v -run "TestGet|TestList|TestSwitch|TestSearch|TestPull|TestHelper" ./...

test-integration:
	@echo "Running integration tests..."
	@go test -v -run "TestHandle|TestMCP|TestFormat" ./...

test-performance:
	@echo "Running performance tests..."
	@go test -v -run "TestPerformance|TestMemory|TestConcurrent|TestResource" ./...

test-edge:
	@echo "Running edge case tests..."
	@go test -v -run "TestEdge|TestError|TestBoundary" ./...

# Test coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Linting
lint:
	@echo "Running linter..."
	@go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f git-remote-mcp coverage.out coverage.html
	@go clean

# Initialize go modules
init:
	@echo "Initializing Go modules..."
	@go mod tidy

# Run the MCP server in stdio mode
run-stdio:
	@echo "Starting MCP server (stdio mode)..."
	@./git-remote-mcp mcp

# Run the MCP server in HTTP mode
run-http:
	@echo "Starting MCP server (HTTP mode on port 8080)..."
	@./git-remote-mcp mcp --transport http --port 8080

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod verify

# Security check
security:
	@echo "Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Run: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Static analysis
analyze:
	@echo "Running static analysis..."
	@go vet ./...
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo "staticcheck not installed. Run: go install honnef.co/go/tools/cmd/staticcheck@latest"; \
	fi

# Complete CI pipeline
ci: fmt lint test-verbose benchmark

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	@go mod tidy
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Help target
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  test          - Run all tests"
	@echo "  test-verbose  - Run tests with verbose output and race detection"
	@echo "  test-short    - Run short tests (skip performance tests)"
	@echo "  test-unit     - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-performance - Run performance tests only"
	@echo "  test-edge     - Run edge case tests only"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  benchmark     - Run benchmarks"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  clean         - Clean build artifacts"
	@echo "  init          - Initialize go modules"
	@echo "  run-stdio     - Run MCP server in stdio mode"
	@echo "  run-http      - Run MCP server in HTTP mode"
	@echo "  deps          - Install dependencies"
	@echo "  security      - Run security checks"
	@echo "  analyze       - Run static analysis"
	@echo "  ci            - Run complete CI pipeline"
	@echo "  dev-setup     - Set up development environment"
	@echo "  help          - Show this help message"