# Git Remote MCP

Git読み取りリモートMCP - A Model Context Protocol (MCP) server for Git remote operations with workspace management.

## Features

This MCP server provides the following tools for Git repository operations:

### Workspace Management
- **clone_repository**: Clone a Git repository into the managed workspace
- **list_workspace_repositories**: List all repositories in the workspace
- **remove_repository**: Remove a repository from the workspace

**Security**: All operations are restricted to repositories within the specified workspace directory.

### Repository Information
- **get_repository_info**: Get basic repository information including:
  - Commit count
  - Last update date
  - Current branch
  - License file detection
  - README content (first 50 lines)
  - Remote URL

### Repository Operations
- **pull_repository**: Execute `git pull` on the specified repository

### Branch Management
- **list_branches**: List all branches in the repository (supports pagination)
- **switch_branch**: Switch to a specified branch

### File Operations
- **search_files**: Search for files containing specified keywords with enhanced filtering
  - AND/OR logic support
  - File pattern filtering (include/exclude patterns)
  - Context lines around matches
  - Filename and content search
- **list_files**: List files in specified directory with enhanced information
  - Recursive expansion
  - File pattern filtering (include/exclude patterns) 
  - Character count and line count for each file
  - File size information
- **get_file_content**: Get the content of files
  - Single file or multiple files in one request
  - Individual error handling for each file
  - Line limits applied per file

## Installation

1. Clone this repository:
```bash
git clone <repository-url>
cd git-remote-mcp
```

2. Build the project:
```bash
go build .
```

## Usage

### Standalone Server

Start the MCP server using stdio transport (default):
```bash
./git-remote-mcp mcp --workspace ./my-workspace
```

Start the MCP server using HTTP transport:
```bash
./git-remote-mcp mcp --transport http --port 8080 --workspace ./my-workspace
```

The workspace directory will be created automatically if it doesn't exist. All Git operations will be restricted to repositories within this workspace.

## Remote MCP Usage

To use this as a remote MCP server:

### 1. Build and Start HTTP Server
```bash
# Build the server
go build .

# Start HTTP server (default port 8080)
./git-remote-mcp mcp --transport http --port 8080 --workspace ./workspace

# Or use the provided script
./start-server.sh ./workspace 8080
```

### 2. Configure MCP Client

Add to your MCP client configuration (e.g., Claude Code):

```json
{
  "mcpServers": {
    "git-remote": {
      "url": "http://localhost:8080/mcp",
      "description": "Git Remote MCP Server for repository operations"
    }
  }
}
```

### 3. Remote Access

For remote access across network:

```bash
# Start server on all interfaces
./git-remote-mcp mcp --transport http --port 8080 --workspace ./workspace --host 0.0.0.0

# Then connect from client with
# "url": "http://your-server-ip:8080/mcp"
```

**Security Note**: When exposing over network, consider adding authentication, HTTPS, and firewall rules.

### 4. Docker Deployment

Run with Docker:

```bash
# Build the image
docker build -t git-remote-mcp .

# Run the container
docker run -d \
  --name git-remote-mcp \
  -p 8080:8080 \
  -v $(pwd)/workspace:/workspace \
  git-remote-mcp

# Or use docker-compose
docker-compose up -d
```

### 5. Production Deployment

For systemd-based systems:

```bash
# Copy files to production location
sudo cp git-remote-mcp /opt/git-remote-mcp/
sudo cp git-remote-mcp.service /etc/systemd/system/

# Create user and directories
sudo useradd -r -s /bin/false git-mcp
sudo mkdir -p /var/lib/git-remote-mcp/workspace
sudo chown -R git-mcp:git-mcp /var/lib/git-remote-mcp

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable git-remote-mcp
sudo systemctl start git-remote-mcp
```

### Tool Parameters

Each tool accepts JSON parameters:

#### clone_repository
```json
{
  "url": "https://github.com/user/repository.git",
  "name": "my-repo"
}
```

#### list_repositories
```json
{}
```

#### remove_repository
```json
{
  "name": "repository-name"
}
```

#### get_repository_info
```json
{
  "repository": "my-repo"
}
```

#### pull_repository
```json
{
  "repository": "my-repo"
}
```

#### list_branches
```json
{
  "repository": "my-repo",
  "limit": 10
}
```

#### switch_branch
```json
{
  "repository": "my-repo",
  "branch": "branch-name"
}
```

#### search_files
```json
{
  "repository": "my-repo",
  "keywords": ["keyword1", "keyword2"],
  "search_mode": "and",
  "include_filename": false,
  "context_lines": 0,
  "include_patterns": ["*.go", "*.js"],
  "exclude_patterns": ["*_test.go", "vendor/*"],
  "limit": 20
}
```

**Parameters:**
- `keywords`: Array of search terms
- `search_mode`: "and" (all keywords) or "or" (any keyword), default: "and"
- `include_filename`: Search in filenames too, default: false
- `context_lines`: Lines of context around matches, default: 0
- `include_patterns`: File patterns to include (glob format)
- `exclude_patterns`: File patterns to exclude (glob format)
- `limit`: Maximum results, default: 20

#### list_files
```json
{
  "repository": "my-repo",
  "directory": "src",
  "recursive": true,
  "include_patterns": ["*.go", "*.js"],
  "exclude_patterns": ["*_test.go", "node_modules/*"],
  "limit": 50
}
```

**Parameters:**
- `directory`: Target directory, default: "."
- `recursive`: Include subdirectories, default: false
- `include_patterns`: File patterns to include (glob format)
- `exclude_patterns`: File patterns to exclude (glob format)
- `limit`: Maximum files to return, default: 50

**Output includes:**
- File path and name
- Directory flag
- File size (bytes/KB/MB)
- Character count (for text files)
- Line count (for text files)
- Modification time

#### get_file_content

**Single file:**
```json
{
  "repository": "my-repo",
  "file_path": "src/main.go",
  "max_lines": 100
}
```

**Multiple files:**
```json
{
  "repository": "my-repo",
  "file_paths": ["src/main.go", "src/utils.go", "config.json"],
  "max_lines": 100
}
```

**Parameters:**
- `file_path`: Single file path (for backward compatibility)
- `file_paths`: Array of file paths (for multiple files)
- `max_lines`: Maximum lines per file, default: 100

**Multiple file output:**
- Individual success/error status per file
- File path identification
- Content or error message for each file

## Enhanced Features Examples

### File Pattern Filtering

**List only Go files:**
```json
{
  "repository": "my-repo",
  "include_patterns": ["*.go"]
}
```

**List all files except tests and vendor:**
```json
{
  "repository": "my-repo",
  "exclude_patterns": ["*_test.go", "vendor/*", "node_modules/*"]
}
```

**Search for functions only in source files:**
```json
{
  "repository": "my-repo", 
  "keywords": ["func"],
  "include_patterns": ["src/*.go", "lib/*.go"],
  "exclude_patterns": ["*_test.go"]
}
```

### Character Count and File Information

The `list_files` tool now returns detailed file information:
```
📄 main.go (2.1 KB, 156 chars, 8 lines)
📄 utils.go (1.5 KB, 98 chars, 5 lines)
📁 src/
📄 src/app.js (856 bytes, 45 chars, 3 lines)
```

### Multiple File Content Retrieval

```json
{
  "repository": "my-repo",
  "file_paths": ["main.go", "config.json", "README.md"],
  "max_lines": 50
}
```

Returns content for all files with individual error handling:
```
Content of 3 files:
==================================================

📄 main.go
```
package main
func main() {}
```

------------------------------------------

📄 config.json  
❌ Error: file not found

------------------------------------------

📄 README.md
```
# My Project
This is a sample project
```
```

## Error Handling

All tools return appropriate error messages when:
- Repository path is invalid or not a Git repository
- Git commands fail
- File operations encounter errors
- Parameters are missing or invalid

## Pagination

Several tools support pagination to prevent large outputs:
- `list_branches`: Limit number of branches returned
- `search_files`: Limit search results (default: 20)
- `list_files`: Limit file listing (default: 50)
- `get_file_content`: Limit lines read (default: 100)

## Security Considerations

- This server performs read-only operations on Git repositories
- The `switch_branch` operation modifies the working directory but doesn't commit changes
- The `pull_repository` operation updates the repository from its remote origin
- Always ensure the server has appropriate permissions for the target repositories

## Dependencies

- Go 1.23.0 or later
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- [Cobra CLI](https://github.com/spf13/cobra)
- Git command-line tools

## License

This project is provided as-is for educational and development purposes.