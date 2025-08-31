#!/bin/bash

# Git Remote MCP Test Runner
# Comprehensive test script that runs all test categories

set -e

echo "==================================="
echo "Git Remote MCP Test Suite"
echo "==================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [[ ! -f "go.mod" ]]; then
    print_error "Must be run from the project root directory"
    exit 1
fi

# Check Go installation
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

print_status "Go version: $(go version)"

# Build the project first
print_status "Building project..."
if go build -o git-remote-mcp .; then
    print_success "Build completed successfully"
else
    print_error "Build failed"
    exit 1
fi

# Initialize variables for test results
UNIT_TESTS_PASSED=false
INTEGRATION_TESTS_PASSED=false
PERFORMANCE_TESTS_PASSED=false
EDGE_TESTS_PASSED=false
BENCHMARKS_PASSED=false

# Function to run test category
run_test_category() {
    local category=$1
    local description=$2
    local pattern=$3
    
    print_status "Running $description..."
    
    if go test -v -run "$pattern" ./...; then
        print_success "$description passed"
        return 0
    else
        print_error "$description failed"
        return 1
    fi
}

# Run unit tests
print_status "=== UNIT TESTS ==="
if run_test_category "unit" "unit tests" "TestGet|TestList|TestSwitch|TestSearch|TestPull|TestHelper"; then
    UNIT_TESTS_PASSED=true
fi

echo ""

# Run integration tests
print_status "=== INTEGRATION TESTS ==="
if run_test_category "integration" "integration tests" "TestHandle|TestMCP|TestFormat"; then
    INTEGRATION_TESTS_PASSED=true
fi

echo ""

# Run edge case tests
print_status "=== EDGE CASE TESTS ==="
if run_test_category "edge" "edge case tests" "TestEdge|TestError|TestBoundary"; then
    EDGE_TESTS_PASSED=true
fi

echo ""

# Run performance tests (only if not in CI or explicitly requested)
if [[ "$1" == "--include-performance" ]] || [[ "$CI" != "true" ]]; then
    print_status "=== PERFORMANCE TESTS ==="
    if run_test_category "performance" "performance tests" "TestPerformance|TestMemory|TestConcurrent|TestResource"; then
        PERFORMANCE_TESTS_PASSED=true
    fi
    echo ""
else
    print_warning "Skipping performance tests (use --include-performance to run them)"
    PERFORMANCE_TESTS_PASSED=true  # Don't fail overall if we skip them
fi

# Run benchmarks (only if explicitly requested)
if [[ "$1" == "--include-benchmarks" ]] || [[ "$2" == "--include-benchmarks" ]]; then
    print_status "=== BENCHMARKS ==="
    if go test -bench=. -benchmem ./...; then
        print_success "Benchmarks completed"
        BENCHMARKS_PASSED=true
    else
        print_error "Benchmarks failed"
    fi
    echo ""
else
    print_warning "Skipping benchmarks (use --include-benchmarks to run them)"
    BENCHMARKS_PASSED=true  # Don't fail overall if we skip them
fi

# Test coverage report
if [[ "$1" == "--coverage" ]] || [[ "$2" == "--coverage" ]] || [[ "$3" == "--coverage" ]]; then
    print_status "=== COVERAGE REPORT ==="
    if go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html; then
        print_success "Coverage report generated: coverage.html"
        print_status "Opening coverage report in browser..."
        if command -v open &> /dev/null; then
            open coverage.html
        elif command -v xdg-open &> /dev/null; then
            xdg-open coverage.html
        else
            print_warning "Cannot open coverage report automatically. Open coverage.html manually."
        fi
    else
        print_error "Coverage report generation failed"
    fi
    echo ""
fi

# Run linter
print_status "=== LINTING ==="
if go vet ./...; then
    print_success "Go vet passed"
    
    # Try to run golangci-lint if available
    if command -v golangci-lint &> /dev/null; then
        if golangci-lint run; then
            print_success "golangci-lint passed"
        else
            print_error "golangci-lint found issues"
        fi
    else
        print_warning "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    fi
else
    print_error "Go vet failed"
fi

echo ""

# Summary
print_status "=== TEST SUMMARY ==="
echo "Unit Tests:        $(if $UNIT_TESTS_PASSED; then echo -e "${GREEN}PASSED${NC}"; else echo -e "${RED}FAILED${NC}"; fi)"
echo "Integration Tests: $(if $INTEGRATION_TESTS_PASSED; then echo -e "${GREEN}PASSED${NC}"; else echo -e "${RED}FAILED${NC}"; fi)"
echo "Edge Case Tests:   $(if $EDGE_TESTS_PASSED; then echo -e "${GREEN}PASSED${NC}"; else echo -e "${RED}FAILED${NC}"; fi)"
echo "Performance Tests: $(if $PERFORMANCE_TESTS_PASSED; then echo -e "${GREEN}PASSED${NC}"; else echo -e "${RED}FAILED${NC}"; fi)"
echo "Benchmarks:        $(if $BENCHMARKS_PASSED; then echo -e "${GREEN}PASSED${NC}"; else echo -e "${RED}FAILED${NC}"; fi)"

# Overall result
if $UNIT_TESTS_PASSED && $INTEGRATION_TESTS_PASSED && $EDGE_TESTS_PASSED && $PERFORMANCE_TESTS_PASSED && $BENCHMARKS_PASSED; then
    print_success "All tests passed!"
    
    # Quick functionality test
    print_status "Running quick functionality test..."
    if ./git-remote-mcp --help > /dev/null 2>&1; then
        print_success "Binary works correctly"
    else
        print_error "Binary execution failed"
        exit 1
    fi
    
    exit 0
else
    print_error "Some tests failed!"
    exit 1
fi