# Project Structure

This document describes the organization and structure of the git-remote-mcp project.

## Directory Structure

```
git-remote-mcp/
├── .claude/                    # Claude AI configuration
├── docs/                      # Documentation files
│   ├── PROJECT_STRUCTURE.md   # This file
│   └── TEST_SUMMARY.md       # Test results and coverage summary
├── scripts/                   # Utility scripts
│   ├── run_tests.sh          # Test runner script
│   └── start-server.sh       # Server startup script
├── .gitignore                # Git ignore patterns
├── CLAUDE.md                 # Claude Code project instructions
├── Dockerfile               # Container build instructions
├── Makefile                 # Build and test targets
├── README.md                # Project overview and usage
├── docker-compose.yml       # Container orchestration
├── example-config.json      # Example configuration
├── git-remote-mcp.service   # Systemd service file
├── go.mod                   # Go module definition
├── go.sum                   # Go dependencies
└── *.go                     # Go source files (see below)
```

## Core Go Source Files

### Application Layer
- `main.go` - CLI entry point with Cobra commands
- `mcp_server.go` - MCP server initialization and transport handling

### MCP Layer
- `mcp_tools_git.go` - MCP tool definitions and handlers for Git operations

### Business Logic Layer
- `git_operations.go` - Core Git operations (info, pull, branches, search, etc.)
- `search_enhanced.go` - Enhanced search functionality with filename search and context lines
- `workspace.go` - Workspace management and security validation

### Test Files
- `*_test.go` - Test files for each corresponding source module
- `test_helpers.go` - Shared test utilities and fixtures

## Key Features by File

### `main.go`
- CLI interface using Cobra framework
- Command definitions and flag handling
- Application bootstrapping

### `mcp_server.go`
- MCP server creation and configuration
- Transport protocol handling (stdio/HTTP)
- Server lifecycle management

### `mcp_tools_git.go`
- 10 registered MCP tools for Git operations
- Parameter validation and error handling
- Response formatting and content generation
- Tools:
  - `get_repository_info` - Repository metadata
  - `clone_repository` - Clone Git repositories
  - `list_workspace_repositories` - List available repositories
  - `remove_repository` - Remove repositories from workspace
  - `pull_repository` - Update repositories
  - `list_branches` - Branch enumeration
  - `switch_branch` - Branch switching
  - `search_files` - Enhanced file search
  - `list_files` - File system navigation
  - `get_file_content` - File content retrieval

### `git_operations.go`
- Direct Git command execution
- Repository validation and operations
- URL parsing and name extraction
- Core data structures (RepositoryInfo, Branch, SearchResult, etc.)

### `search_enhanced.go`
- Advanced search capabilities
- Filename and content search
- AND/OR search modes
- Context lines for search results
- Git grep command optimization

### `workspace.go`
- Security-focused workspace management
- Path validation and traversal prevention
- Repository isolation
- Directory management operations

## Architecture Principles

### Security Model
- All operations restricted to configured workspace directory
- Path validation prevents directory traversal attacks
- Repository isolation and validation

### MCP Protocol Compliance
- Full MCP v2024-11-05 protocol implementation
- Both stdio and HTTP transport support
- Proper error handling and response formatting

### Testability
- Comprehensive test coverage across all layers
- Mock repository creation utilities
- Performance and edge case testing
- Integration test support

### Extensibility
- Modular design with clear separation of concerns
- Easy addition of new Git operations
- Configurable workspace management
- Plugin-ready architecture

## Build and Deployment

### Build Targets (Makefile)
- `make build` - Build binary
- `make test` - Run all tests
- `make test-coverage` - Generate coverage reports
- `make lint` - Code quality checks
- `make clean` - Cleanup build artifacts

### Docker Support
- `Dockerfile` for containerized deployment
- `docker-compose.yml` for service orchestration
- Volume mounting for workspace persistence

### System Service
- `git-remote-mcp.service` for systemd integration
- Automated startup and lifecycle management
- Log management and monitoring support

## Development Guidelines

### Code Organization
- One primary responsibility per file
- Test files co-located with source files
- Shared utilities in dedicated files
- Clear import structure and dependencies

### Testing Strategy
- Unit tests for individual functions
- Integration tests for MCP handlers
- Performance tests for resource usage
- Edge case tests for error conditions

### Documentation Standards
- Code comments for public APIs
- README for user-facing documentation
- CLAUDE.md for development instructions
- Inline documentation for complex logic

This structure provides a maintainable, secure, and extensible foundation for Git remote operations via the Model Context Protocol.