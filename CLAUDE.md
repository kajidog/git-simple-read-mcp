# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test Commands

```bash
# Build the application
make build
# or
go build -o git-simple-read-mcp .

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

This is a Go-based Model Context Protocol (MCP) server that provides Git read operations with workspace security and enhanced file operations.

### Core Architecture Layers

1. **MCP Layer** (`mcp_server.go`, `mcp_tools_git.go`, `mcp_tools_memo.go`)
   - MCP server initialization with both stdio and HTTP transport support
   - Tool parameter definitions and handler registration
   - 28 registered tools: repository info, cloning, branch listing, enhanced file operations, pattern-based search, README discovery, session configuration, batch operations, composite tools, cross-repository search, memo management (add/get/update/delete/list/delete-all)

2. **Workspace Security Layer** (`workspace.go`)
   - `WorkspaceManager` enforces all operations within a specified workspace directory
   - Path validation prevents directory traversal attacks
   - Repository isolation and management

3. **Git Operations Layer** (`git_operations.go`, `search_enhanced.go`)
   - Direct Git command execution via `os/exec`
   - Core Git read operations: info, pull, branch listing, enhanced file listing, content reading
   - Pattern-based file filtering with glob support
   - Character and line counting for text files
   - Multiple file content retrieval with individual error handling
   - Automatic repository name extraction from Git URLs
   - README file discovery with recursive search support
   - Line numbers for file content display (always enabled)
   - Start line offset for reading from specific line

4. **Memo Management Layer** (`memo.go`)
   - Persistent document memo storage and retrieval
   - JSON-based file storage in workspace directory (`memos.json`)
   - Thread-safe operations with mutex synchronization
   - CRUD operations: Create, Read, Update, Delete
   - Search and filtering by title, content, and tags
   - Cross-session persistence

5. **Application Layer** (`main.go`)
   - Cobra CLI framework with `mcp` subcommand
   - Transport selection (stdio/HTTP) with configurable host/port
   - Workspace directory initialization
   - Memo store initialization

### Key Security Model

All Git operations are restricted to repositories within the configured workspace:
- `ValidateRepositoryPath()` ensures paths stay within workspace bounds
- Repository names are validated and paths are resolved to prevent attacks
- Clone operations automatically extract repository names from URLs if not provided

### MCP Tool Parameter Design

Tools use consistent parameter naming:
- `repository` (not `path`) - more intuitive for Git operations
- Optional parameters use `omitempty` JSON tags
- Automatic defaults:
  - `limit=20` (search), `limit=50` (list_files)
  - `max_lines=100`, `start_line=1` (get_file_content)
  - `show_line_numbers=true` (always enabled for AI-friendly output)

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
- `start_line` parameter allows reading from specified line (1-based, default: 1)
- Line numbers always displayed (AI-friendly default)
- Returns metadata: total lines, actual start/end lines
- AI-optimized minimal output format: `[path L{start}-{end}/{total}]`
- Individual error handling per file in multi-file requests
- Per-file line limits applied consistently

**README File Discovery:**
- `get_readme_files` tool finds all README files in repository
- Supports both non-recursive (root only) and recursive search modes
- Matches multiple README patterns: README, README.*, readme, readme.*, Readme, Readme.*
- Provides file metadata: size, modification time, line count

### Session Configuration (`session_config.go`)

Server-side session state management to reduce tool call overhead:
- `set_session_config`: Set default values (repository, patterns, limits)
- `get_session_config`: View current session configuration
- `clear_session_config`: Reset to defaults

**Session-aware Parameters:**
- `default_repository`: Used when no repository is specified
- `default_include_patterns` / `default_exclude_patterns`: Default file patterns
- `default_search_limit`, `default_list_files_limit`, `default_max_lines`, `default_commit_limit`

### Batch Operations

Execute operations on multiple repositories in a single call:
- `batch_clone`: Clone multiple repositories at once
- `batch_pull`: Pull all (or specified) repositories
- `batch_status`: Get status of all (or specified) repositories

### Composite Tools

Combine frequently-used operations into single calls:
- `explore_repository`: get_repository_info + list_files + get_readme_files
- `setup_repository`: clone + get_repository_info + list_branches
- `get_workspace_overview`: Summary of all repositories (branches, status, commits)

### Cross-Repository Search

- `cross_repo_search`: Search across multiple (or all) repositories at once
- Supports all search parameters (keywords, patterns, context lines)
- Returns results grouped by repository

### Memo Management

Document memo system for persistent note-taking across sessions with repository association:

**Data Structure:**
- ID: Unique identifier (UUID) - full UUID displayed in all outputs for AI usability
- Repository: Associated repository name (optional, for organizing memos by project)
- Title: Memo title (required)
- Content: Memo body text
- Tags: Optional tags for categorization
- CreatedAt / UpdatedAt: Timestamps

**Available Tools:**
- `add_memo`: Create a new memo. Parameters: repository (optional), title (required), content, tags
- `get_memo`: Retrieve a specific memo by ID (displays full UUID and repository)
- `update_memo`: Update memo. Parameters: id (required), repository, title, content, tags
- `delete_memo`: Delete a memo by ID
- `list_memos`: Search/list memos. Parameters: repository (filter by repo), query (search title/content), tags, limit
- `delete_all_memos`: Delete all memos (use with caution)

**Repository Integration:**
- `get_repository_info` with `include_memos=true` shows associated memos
- `memo_limit` parameter controls how many memos to display (default: 10)
- Memos can be filtered by repository name in list_memos

**Storage:**
- Memos are stored in `<workspace>/memos.json`
- Persistent across sessions
- Thread-safe with mutex synchronization
- Automatic save on every modification

**Search Capabilities:**
- Search by keywords in title/content (case-insensitive)
- Filter by tags
- Result limiting with configurable max results

### Test Structure

Tests are organized into categories matching the Makefile targets:
- `*_test.go` - Core functionality tests
- `enhanced_features_test.go` - Tests for new pattern filtering, character count, and multi-file features
- `readme_line_numbers_test.go` - Tests for README discovery and line numbers functionality
- `memo_test.go` - Tests for memo management functionality (CRUD, persistence, concurrency)
- `test_helpers.go` - Shared test utilities with `TestRepository` struct
- `performance_test.go` - Resource usage and concurrency tests
- `edge_cases_test.go` - Error conditions and boundary cases
- `workspace_test.go` - Security and workspace validation tests

### Transport Modes

- **stdio**: Default mode for direct MCP client integration
- **http**: Remote server mode with configurable host/port for network access

### HTTP Endpoints

When running in HTTP mode:
- `/mcp` - MCP protocol endpoint
- `/health` - Health check endpoint (returns "ok")

### Repository URL Handling

The `extractRepoNameFromURL()` function handles various Git URL formats:
- HTTPS: `https://github.com/user/repo.git` → `repo`
- SSH: `git@github.com:user/repo.git` → `repo`
- Handles `.git` suffix removal and query parameters

## Development Notes

- All workspace operations require `InitializeWorkspace()` before use
- Memo management requires `InitializeMemoStore()` after workspace initialization
- Git operations return `(output, repoName, error)` pattern for consistent error handling
- MCP handlers return `*mcp.CallToolResultFor[any]` with `IsError` flag
- Parameter validation happens at the MCP handler level before calling Git operations
- Test setup creates isolated temporary workspaces to avoid conflicts