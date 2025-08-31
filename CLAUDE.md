# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test Commands

```bash
# Build the application
make build
# or
go build -o git-remote-mcp .

# Run all tests
make test

# Run tests with race detection
make test-verbose

# Run specific test categories
make test-unit          # Core functionality tests
make test-integration   # MCP handler tests
make test-performance   # Performance and resource tests
make test-edge         # Edge cases and error handling

# Run single test
go test -v -run TestSpecificTest

# Generate test coverage
make test-coverage     # Creates coverage.html

# Linting and formatting
make lint             # Run go vet and golangci-lint
make fmt              # Format all Go code

# Start server for testing
make run-stdio        # stdio transport (default)
make run-http         # HTTP transport on port 8080
```

## Architecture Overview

This is a Go-based Model Context Protocol (MCP) server that provides Git remote operations with workspace security and enhanced file operations.

### Core Architecture Layers

1. **MCP Layer** (`mcp_server.go`, `mcp_tools_git.go`)
   - MCP server initialization with both stdio and HTTP transport support
   - Tool parameter definitions and handler registration
   - 10 registered tools: repository info, cloning, branch operations, enhanced file operations, pattern-based search

2. **Workspace Security Layer** (`workspace.go`)
   - `WorkspaceManager` enforces all operations within a specified workspace directory
   - Path validation prevents directory traversal attacks
   - Repository isolation and management

3. **Git Operations Layer** (`git_operations.go`, `search_enhanced.go`)
   - Direct Git command execution via `os/exec`
   - Core Git operations: info, pull, branches, enhanced file listing, content reading
   - Pattern-based file filtering with glob support
   - Character and line counting for text files
   - Multiple file content retrieval with individual error handling
   - Automatic repository name extraction from Git URLs

4. **Application Layer** (`main.go`)
   - Cobra CLI framework with `mcp` subcommand
   - Transport selection (stdio/HTTP) with configurable host/port
   - Workspace directory initialization

### Key Security Model

All Git operations are restricted to repositories within the configured workspace:
- `ValidateRepositoryPath()` ensures paths stay within workspace bounds
- Repository names are validated and paths are resolved to prevent attacks
- Clone operations automatically extract repository names from URLs if not provided

### MCP Tool Parameter Design

Tools use consistent parameter naming:
- `repository` (not `path`) - more intuitive for Git operations
- Optional parameters use `omitempty` JSON tags
- Automatic defaults: `limit=20` (search), `limit=50` (list_files), `max_lines=100`

### Enhanced File Operations

**Pattern Filtering:**
- `include_patterns` and `exclude_patterns` support glob patterns (*.go, src/*, etc.)
- Applied to both `list_files` and `search_files` operations
- Uses Go's `filepath.Match` for pattern matching

**File Information Enhancement:**
- Character count and line count for text files
- File size with human-readable formatting (bytes/KB/MB)
- Modification timestamps

**Multiple File Content:**
- `get_file_content` supports single file (backward compatible) and multiple files
- Individual error handling per file in multi-file requests
- Per-file line limits applied consistently

### Test Structure

Tests are organized into categories matching the Makefile targets:
- `*_test.go` - Core functionality tests
- `enhanced_features_test.go` - Tests for new pattern filtering, character count, and multi-file features
- `test_helpers.go` - Shared test utilities with `TestRepository` struct
- `performance_test.go` - Resource usage and concurrency tests
- `edge_cases_test.go` - Error conditions and boundary cases
- `workspace_test.go` - Security and workspace validation tests

### Transport Modes

- **stdio**: Default mode for direct MCP client integration
- **http**: Remote server mode with configurable host/port for network access

### Repository URL Handling

The `extractRepoNameFromURL()` function handles various Git URL formats:
- HTTPS: `https://github.com/user/repo.git` → `repo`
- SSH: `git@github.com:user/repo.git` → `repo`
- Handles `.git` suffix removal and query parameters

## Development Notes

- All workspace operations require `InitializeWorkspace()` before use
- Git operations return `(output, repoName, error)` pattern for consistent error handling
- MCP handlers return `*mcp.CallToolResultFor[any]` with `IsError` flag
- Parameter validation happens at the MCP handler level before calling Git operations
- Test setup creates isolated temporary workspaces to avoid conflicts