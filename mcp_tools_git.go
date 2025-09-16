package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetRepositoryInfoParams parameters for get_repository_info tool
type GetRepositoryInfoParams struct {
	Repository string `json:"repository"`
}

// PullRepositoryParams parameters for pull_repository tool
type PullRepositoryParams struct {
	Repository string `json:"repository"`
}

// ListBranchesParams parameters for list_branches tool
type ListBranchesParams struct {
	Repository string `json:"repository"`
	Limit      int    `json:"limit,omitempty"`
}

// SwitchBranchParams parameters for switch_branch tool
type SwitchBranchParams struct {
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
}

// SearchFilesParams parameters for search_files tool
type SearchFilesParams struct {
	Repository      string   `json:"repository"`
	Keywords        []string `json:"keywords"`
	SearchMode      string   `json:"search_mode,omitempty"`      // "and" or "or", defaults to "and"
	IncludeFilename bool     `json:"include_filename,omitempty"` // search in filenames too, defaults to false
	ContextLines    int      `json:"context_lines,omitempty"`    // number of context lines before/after match, 0=no context
	IncludePatterns []string `json:"include_patterns,omitempty"` // file patterns to include (glob)
	ExcludePatterns []string `json:"exclude_patterns,omitempty"` // file patterns to exclude (glob)
	Limit           int      `json:"limit,omitempty"`
}

// ListFilesParams parameters for list_files tool
type ListFilesParams struct {
	Repository      string   `json:"repository"`
	Directory       string   `json:"directory,omitempty"`
	Recursive       bool     `json:"recursive,omitempty"`
	IncludePatterns []string `json:"include_patterns,omitempty"` // file patterns to include (glob)
	ExcludePatterns []string `json:"exclude_patterns,omitempty"` // file patterns to exclude (glob)
	Limit           int      `json:"limit,omitempty"`
}

// GetFileContentParams parameters for get_file_content tool
type GetFileContentParams struct {
	Repository      string   `json:"repository"`
	FilePath        string   `json:"file_path,omitempty"`    // Single file path (for backward compatibility)
	FilePaths       []string `json:"file_paths,omitempty"`   // Multiple file paths
	MaxLines        int      `json:"max_lines,omitempty"`    // Max lines per file
	StartLine       int      `json:"start_line,omitempty"`   // Line to start reading from (1-based)
	ShowLineNumbers bool     `json:"show_line_numbers,omitempty"` // Show line numbers in output
}

// CloneRepositoryParams parameters for clone_repository tool
type CloneRepositoryParams struct {
	URL  string `json:"url"`
	Name string `json:"name,omitempty"` // Optional: will be extracted from URL if not provided
}

// ListWorkspaceRepositoriesParams parameters for list_workspace_repositories tool
type ListWorkspaceRepositoriesParams struct {
	// No parameters needed
}

// RemoveRepositoryParams parameters for remove_repository tool
type RemoveRepositoryParams struct {
	Name string `json:"name"`
}

// GetReadmeFilesParams parameters for get_readme_files tool
type GetReadmeFilesParams struct {
	Repository string `json:"repository"`
	Recursive  bool   `json:"recursive,omitempty"` // Search subdirectories
}

// ListCommitsParams parameters for list_commits tool
type ListCommitsParams struct {
	Repository string `json:"repository"`
	Limit      int    `json:"limit,omitempty"`
}

// GetCommitDiffParams parameters for get_commit_diff tool
type GetCommitDiffParams struct {
	Repository string `json:"repository"`
	CommitHash string `json:"commit_hash"`
}

// RegisterGitTools registers all Git-related MCP tools
func RegisterGitTools(server *mcp.Server) {
	// Repository information tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_repository_info",
		Description: "Get basic repository information including commit count, last update, current branch, license, and README content",
	}, handleGetRepositoryInfo)

	// Repository pull tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pull_repository",
		Description: "Execute git pull on the specified repository",
	}, handlePullRepository)

	// List branches tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_branches",
		Description: "List all branches in the repository with pagination support",
	}, handleListBranches)

	// Switch branch tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "switch_branch",
		Description: "Switch to the specified branch in the repository",
	}, handleSwitchBranch)

	// Search files tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_files",
		Description: "Search for files containing specified keywords in content and/or filenames with AND/OR logic, context lines, and pagination",
	}, handleSearchFiles)

	// List files tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_files",
		Description: "List files in specified directory with optional recursive expansion and pagination",
	}, handleListFiles)

	// Get file content tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_file_content",
		Description: "Get the content of a specific file with optional line limit",
	}, handleGetFileContent)

	// Clone repository tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "clone_repository",
		Description: "Clone a Git repository into the workspace",
	}, handleCloneRepository)

	// List workspace repositories tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_repositories",
		Description: "List all repositories in the workspace",
	}, handleListWorkspaceRepositories)

	// Remove repository tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "remove_repository",
		Description: "Remove a repository from the workspace",
	}, handleRemoveRepository)

	// Get README files tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_readme_files",
		Description: "Find and list all README files in the repository with optional recursive search",
	}, handleGetReadmeFiles)

	// Add new tools for commit history and diff
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_commits",
		Description: "List the commit history for a repository",
	}, handleListCommits)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_commit_diff",
		Description: "Get the diff for a specific commit, including the commit message",
	}, handleGetCommitDiff)
}

func handleGetRepositoryInfo(ctx context.Context, req *mcp.CallToolRequest, args GetRepositoryInfoParams) (*mcp.CallToolResult, any, error) {
	if args.Repository == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: repository path is required"}},
			IsError: true,
		}, nil, nil
	}

	info, err := GetRepositoryInfo(args.Repository)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to get repository info: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	resultText := formatRepositoryInfo(info)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
	}, nil, nil
}

func handlePullRepository(ctx context.Context, req *mcp.CallToolRequest, args PullRepositoryParams) (*mcp.CallToolResult, any, error) {
	if args.Repository == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: repository path is required"}},
			IsError: true,
		}, nil, nil
	}

	output, err := PullRepository(args.Repository)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Pull failed: %v\nOutput: %s", err, output)}},
			IsError: true,
		}, nil, nil
	}

	resultText := fmt.Sprintf("Git pull completed successfully:\n%s", output)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
	}, nil, nil
}

func handleListBranches(ctx context.Context, req *mcp.CallToolRequest, args ListBranchesParams) (*mcp.CallToolResult, any, error) {
	if args.Repository == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: repository path is required"}},
			IsError: true,
		}, nil, nil
	}

	branches, err := ListBranches(args.Repository)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list branches: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	// Apply limit if specified
	if args.Limit > 0 && len(branches) > args.Limit {
		branches = branches[:args.Limit]
	}

	resultText := formatBranches(branches, args.Limit > 0)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
	}, nil, nil
}

func handleSwitchBranch(ctx context.Context, req *mcp.CallToolRequest, args SwitchBranchParams) (*mcp.CallToolResult, any, error) {
	if args.Repository == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: repository path is required"}},
			IsError: true,
		}, nil, nil
	}

	if args.Branch == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: branch name is required"}},
			IsError: true,
		}, nil, nil
	}

	output, err := SwitchBranch(args.Repository, args.Branch)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Branch switch failed: %v\nOutput: %s", err, output)}},
			IsError: true,
		}, nil, nil
	}

	resultText := fmt.Sprintf("Successfully switched to branch '%s':\n%s", args.Branch, output)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
	}, nil, nil
}

func handleSearchFiles(ctx context.Context, req *mcp.CallToolRequest, args SearchFilesParams) (*mcp.CallToolResult, any, error) {
	if args.Repository == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: repository path is required"}},
			IsError: true,
		}, nil, nil
	}

	if len(args.Keywords) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: at least one keyword is required"}},
			IsError: true,
		}, nil, nil
	}

	// Default limit to prevent token overflow
	limit := args.Limit
	if limit == 0 {
		limit = 20
	}

	// Default search mode to "and"
	searchMode := args.SearchMode
	if searchMode == "" {
		searchMode = "and"
	}

	results, err := SearchFiles(args.Repository, args.Keywords, searchMode, args.IncludeFilename, args.ContextLines, args.IncludePatterns, args.ExcludePatterns, limit)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Search failed: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	resultText := formatSearchResults(results, args.Keywords, searchMode)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
	}, nil, nil
}

func handleListFiles(ctx context.Context, req *mcp.CallToolRequest, args ListFilesParams) (*mcp.CallToolResult, any, error) {
	if args.Repository == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: repository path is required"}},
			IsError: true,
		}, nil, nil
	}

	directory := args.Directory
	if directory == "" {
		directory = "."
	}

	// Default limit to prevent token overflow
	limit := args.Limit
	if limit == 0 {
		limit = 50
	}

	files, err := ListFiles(args.Repository, directory, args.Recursive, args.IncludePatterns, args.ExcludePatterns, limit)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list files: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	resultText := formatFileList(files, directory, args.Recursive, limit)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
	}, nil, nil
}

func handleGetFileContent(ctx context.Context, req *mcp.CallToolRequest, args GetFileContentParams) (*mcp.CallToolResult, any, error) {
	if args.Repository == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: repository path is required"}},
			IsError: true,
		}, nil, nil
	}

	// Determine which file paths to use (backward compatibility)
	var filePaths []string
	if len(args.FilePaths) > 0 {
		filePaths = args.FilePaths
	} else if args.FilePath != "" {
		filePaths = []string{args.FilePath}
	} else {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: file path(s) required"}},
			IsError: true,
		}, nil, nil
	}

	// Default limit to prevent token overflow
	maxLines := args.MaxLines
	if maxLines == 0 {
		maxLines = 100
	}

	// Handle start_line, ensuring it's at least 1
	startLine := args.StartLine
	if startLine < 1 {
		startLine = 1
	}

	if len(filePaths) == 1 {
		// Single file
		result, err := GetFileContentWithLineNumbers(args.Repository, filePaths[0], maxLines, startLine, args.ShowLineNumbers)
		if err != nil {
			// Ensure we always have a result object for consistent error formatting
			if result == nil {
				result = &FileContentResult{
					FilePath: filePaths[0],
					Error:    err.Error(),
				}
			} else if result.Error == "" {
				result.Error = err.Error()
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: formatSingleFileContent(*result)}},
				IsError: true,
			}, nil, nil
		}

		resultText := formatSingleFileContent(*result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
		}, nil, nil
	} else {
		// Multiple files
		results, err := GetMultipleFileContentsWithLineNumbers(args.Repository, filePaths, maxLines, startLine, args.ShowLineNumbers)
		if err != nil {
			// The call returns partial results, so we format them anyway
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: formatMultipleFileContents(results)}},
				IsError: true, // Mark as error despite sending content
			}, nil, nil
		}

		resultText := formatMultipleFileContents(results)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
		}, nil, nil
	}
}

func handleCloneRepository(ctx context.Context, req *mcp.CallToolRequest, args CloneRepositoryParams) (*mcp.CallToolResult, any, error) {
	if args.URL == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: repository URL is required"}},
			IsError: true,
		}, nil, nil
	}

	output, actualName, err := CloneRepository(args.URL, args.Name)
	if err != nil {
		// Check if the error is because the repository already exists
		if strings.Contains(err.Error(), "already exists in workspace") {
			// If it exists, pull the latest changes
			pullOutput, pullErr := PullRepository(actualName)
			if pullErr != nil {
				errorMsg := fmt.Sprintf("Repository '%s' already exists but pull failed: %v\nOutput: %s", actualName, pullErr, pullOutput)
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{Text: errorMsg}},
					IsError: true,
				}, nil, nil
			}
			resultText := fmt.Sprintf("Repository '%s' already exists. Pulled latest changes:\n%s", actualName, pullOutput)
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
			}, nil, nil
		}

		// Handle other clone errors
		var errorMsg string
		if args.Name == "" {
			errorMsg = fmt.Sprintf("Clone failed for repository '%s' (extracted from URL): %v\nOutput: %s", actualName, err, output)
		} else {
			errorMsg = fmt.Sprintf("Clone failed for repository '%s': %v\nOutput: %s", actualName, err, output)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: errorMsg}},
			IsError: true,
		}, nil, nil
	}

	var resultText string
	if args.Name == "" {
		resultText = fmt.Sprintf("Successfully cloned repository as '%s' (extracted from URL):\n%s", actualName, output)
	} else {
		resultText = fmt.Sprintf("Successfully cloned repository '%s':\n%s", actualName, output)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
	}, nil, nil
}

func handleListWorkspaceRepositories(ctx context.Context, req *mcp.CallToolRequest, args ListWorkspaceRepositoriesParams) (*mcp.CallToolResult, any, error) {
	wm := GetWorkspaceManager()
	if wm == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: workspace not initialized"}},
			IsError: true,
		}, nil, nil
	}

	repositories, err := wm.ListRepositories()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list repositories: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	resultText := formatWorkspaceRepositories(repositories, wm.GetWorkspaceDir())
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
	}, nil, nil
}

func handleRemoveRepository(ctx context.Context, req *mcp.CallToolRequest, args RemoveRepositoryParams) (*mcp.CallToolResult, any, error) {
	if args.Name == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: repository name is required"}},
			IsError: true,
		}, nil, nil
	}

	wm := GetWorkspaceManager()
	if wm == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: workspace not initialized"}},
			IsError: true,
		}, nil, nil
	}

	err := wm.RemoveRepository(args.Name)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to remove repository: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	resultText := fmt.Sprintf("Successfully removed repository '%s'", args.Name)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
	}, nil, nil
}

func handleGetReadmeFiles(ctx context.Context, req *mcp.CallToolRequest, args GetReadmeFilesParams) (*mcp.CallToolResult, any, error) {
	if args.Repository == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: repository path is required"}},
			IsError: true,
		}, nil, nil
	}

	readmeFiles, err := GetReadmeFiles(args.Repository, args.Recursive)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to find README files: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	resultText := formatReadmeFiles(readmeFiles, args.Recursive)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
	}, nil, nil
}

// Formatting functions

func formatRepositoryInfo(info *RepositoryInfo) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Repository Information for: %s\n", info.Path))
	result.WriteString(strings.Repeat("=", 50) + "\n\n")

	result.WriteString(fmt.Sprintf("Current Branch: %s\n", info.CurrentBranch))
	result.WriteString(fmt.Sprintf("Total Commits: %d\n", info.CommitCount))

	if !info.LastUpdate.IsZero() {
		result.WriteString(fmt.Sprintf("Last Update: %s\n", info.LastUpdate.Format("2006-01-02 15:04:05")))
	}

	if info.RemoteURL != "" {
		result.WriteString(fmt.Sprintf("Remote URL: %s\n", info.RemoteURL))
	}

	if info.License != "" {
		result.WriteString(fmt.Sprintf("License File: %s\n", info.License))
	}

	if info.ReadmeContent != "" {
		result.WriteString("\nREADME Content:\n")
		result.WriteString(strings.Repeat("-", 30) + "\n")
		result.WriteString(info.ReadmeContent)
		if !strings.HasSuffix(info.ReadmeContent, "\n") {
			result.WriteString("\n")
		}
	}

	return result.String()
}

func handleListCommits(ctx context.Context, req *mcp.CallToolRequest, args ListCommitsParams) (*mcp.CallToolResult, any, error) {
	if args.Repository == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: repository path is required"}},
			IsError: true,
		}, nil, nil
	}

	limit := args.Limit
	if limit == 0 {
		limit = 20 // Default limit
	}

	commits, err := ListCommits(args.Repository, limit)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list commits: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	resultText := formatCommits(commits, limit)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
	}, nil, nil
}

func handleGetCommitDiff(ctx context.Context, req *mcp.CallToolRequest, args GetCommitDiffParams) (*mcp.CallToolResult, any, error) {
	if args.Repository == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: repository path is required"}},
			IsError: true,
		}, nil, nil
	}
	if args.CommitHash == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: commit_hash is required"}},
			IsError: true,
		}, nil, nil
	}

	diff, err := GetCommitDiff(args.Repository, args.CommitHash)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to get commit diff: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	resultText := formatCommitDiff(args.CommitHash, diff)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
	}, nil, nil
}

func formatCommits(commits []Commit, limit int) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Commit History (%d commits):\n", len(commits)))
	result.WriteString(strings.Repeat("=", 50) + "\n\n")

	if len(commits) == 0 {
		result.WriteString("No commits found.\n")
		return result.String()
	}

	for _, commit := range commits {
		result.WriteString(fmt.Sprintf("commit %s\n", commit.Hash))
		result.WriteString(fmt.Sprintf("Author: %s\n", commit.Author))
		result.WriteString(fmt.Sprintf("Date:   %s\n", commit.Date))
		result.WriteString(fmt.Sprintf("\n    %s\n\n", commit.Message))
	}

	if len(commits) == limit {
		result.WriteString(fmt.Sprintf("(Limited to %d commits)\n", limit))
	}

	return result.String()
}

func formatCommitDiff(commitHash, diff string) string {
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Diff for commit %s:\n", commitHash))
	result.WriteString(strings.Repeat("=", 50) + "\n\n")
	result.WriteString(diff)
	return result.String()
}

func formatBranches(branches []Branch, limited bool) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Branches (%d):\n", len(branches)))
	result.WriteString(strings.Repeat("-", 30) + "\n")

	for _, branch := range branches {
		if branch.IsCurrent {
			result.WriteString(fmt.Sprintf("* %s (current)\n", branch.Name))
		} else {
			result.WriteString(fmt.Sprintf("  %s\n", branch.Name))
		}
	}

	if limited {
		result.WriteString("\n(Results may be limited)")
	}

	return result.String()
}

func formatSearchResults(results []SearchResult, keywords []string, searchMode string) string {
	var result strings.Builder

	var keywordStr string
	if searchMode == "or" {
		keywordStr = strings.Join(keywords, " OR ")
	} else {
		keywordStr = strings.Join(keywords, " AND ")
	}
	result.WriteString(fmt.Sprintf("Search Results for: %s (%d files found)\n", keywordStr, len(results)))
	result.WriteString(strings.Repeat("-", 50) + "\n")

	if len(results) == 0 {
		result.WriteString("No files found matching the specified keywords.\n")
		return result.String()
	}

	for i, searchResult := range results {
		if i > 0 {
			result.WriteString("\n")
		}

		// Show file path with match type
		matchTypeStr := ""
		if searchResult.MatchType == "filename" {
			matchTypeStr = " [filename match]"
		} else if searchResult.MatchType == "content" {
			matchTypeStr = " [content match]"
		} else if searchResult.MatchType == "both" {
			matchTypeStr = " [filename + content match]"
		}

		result.WriteString(fmt.Sprintf("📄 %s%s\n", searchResult.Path, matchTypeStr))

		// Show detailed matches
		if len(searchResult.Matches) > 0 {
			for _, match := range searchResult.Matches {
				if match.LineNumber == 0 {
					// Filename match
					result.WriteString(fmt.Sprintf("   └─ Filename: %s\n", match.Content))
				} else {
					// Content match with line number
					result.WriteString(fmt.Sprintf("   └─ Line %d: %s\n", match.LineNumber, strings.TrimSpace(match.Content)))
				}
			}
		}
	}

	return result.String()
}

func formatFileList(files []FileInfo, directory string, recursive bool, limit int) string {
	var result strings.Builder

	modeStr := "non-recursive"
	if recursive {
		modeStr = "recursive"
	}

	result.WriteString(fmt.Sprintf("Files in '%s' (%s, %d items):\n", directory, modeStr, len(files)))
	result.WriteString(strings.Repeat("-", 50) + "\n")

	for _, file := range files {
		if file.IsDir {
			result.WriteString(fmt.Sprintf("📁 %s/\n", file.Path))
		} else {
			infoStr := ""
			if file.Size > 0 || file.CharCount > 0 {
				var parts []string

				// Add file size
				if file.Size > 0 {
					if file.Size < 1024 {
						parts = append(parts, fmt.Sprintf("%d bytes", file.Size))
					} else if file.Size < 1024*1024 {
						parts = append(parts, fmt.Sprintf("%.1f KB", float64(file.Size)/1024))
					} else {
						parts = append(parts, fmt.Sprintf("%.1f MB", float64(file.Size)/(1024*1024)))
					}
				}

				// Add character and line count
				if file.CharCount > 0 {
					parts = append(parts, fmt.Sprintf("%d chars", file.CharCount))
				}
				if file.LineCount > 0 {
					parts = append(parts, fmt.Sprintf("%d lines", file.LineCount))
				}

				if len(parts) > 0 {
					infoStr = fmt.Sprintf(" (%s)", strings.Join(parts, ", "))
				}
			}
			result.WriteString(fmt.Sprintf("📄 %s%s\n", file.Path, infoStr))
		}
	}

	if len(files) == limit {
		result.WriteString(fmt.Sprintf("\n(Limited to %d results)", limit))
	}

	return result.String()
}

func formatWorkspaceRepositories(repositories []string, workspaceDir string) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Workspace Repositories (%s):\n", workspaceDir))
	result.WriteString(strings.Repeat("=", 50) + "\n\n")

	if len(repositories) == 0 {
		result.WriteString("No repositories found in workspace.\n")
		result.WriteString("Use 'clone_repository' tool to add repositories.\n")
		return result.String()
	}

	for _, repo := range repositories {
		result.WriteString(fmt.Sprintf("📁 %s\n", repo))
	}

	result.WriteString(fmt.Sprintf("\nTotal: %d repositories\n", len(repositories)))

	return result.String()
}

func formatSingleFileContent(result FileContentResult) string {
	var builder strings.Builder

	// Header
	header := fmt.Sprintf("Content of %s (lines %d-%d of %d)", result.FilePath, result.StartLine, result.EndLine, result.TotalLines)
	if result.Error != "" {
		header = fmt.Sprintf("Error for %s", result.FilePath)
	}
	builder.WriteString(header + "\n")
	builder.WriteString(strings.Repeat("=", len(header)) + "\n\n")

	if result.Error != "" {
		builder.WriteString(fmt.Sprintf("❌ Error: %s\n", result.Error))
	}

	if result.Content != "" {
		builder.WriteString("```\n")
		builder.WriteString(result.Content)
		if !strings.HasSuffix(result.Content, "\n") {
			builder.WriteString("\n")
		}
		builder.WriteString("```\n")
	} else if result.Error == "" {
		builder.WriteString("No content to display.\n")
		if result.StartLine > result.TotalLines && result.TotalLines > 0 {
			builder.WriteString(fmt.Sprintf("Note: Start line (%d) is beyond the end of the file (%d lines).\n", result.StartLine, result.TotalLines))
		}
	}

	return builder.String()
}

func formatMultipleFileContents(results []FileContentResult) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Content of %d files:\n", len(results)))
	result.WriteString(strings.Repeat("=", 50) + "\n\n")

	for i, fileResult := range results {
		if i > 0 {
			result.WriteString("\n" + strings.Repeat("-", 40) + "\n\n")
		}

		// Header
		header := fmt.Sprintf("📄 %s (lines %d-%d of %d)", fileResult.FilePath, fileResult.StartLine, fileResult.EndLine, fileResult.TotalLines)
		if fileResult.Error != "" {
			header = fmt.Sprintf("📄 %s", fileResult.FilePath)
		}
		result.WriteString(header + "\n")

		if fileResult.Error != "" {
			result.WriteString(fmt.Sprintf("❌ Error: %s\n", fileResult.Error))
		} else {
			result.WriteString("```\n")
			result.WriteString(fileResult.Content)
			if !strings.HasSuffix(fileResult.Content, "\n") {
				result.WriteString("\n")
			}
			result.WriteString("```\n")
			if fileResult.StartLine > fileResult.TotalLines && fileResult.TotalLines > 0 {
				result.WriteString(fmt.Sprintf("Note: Start line (%d) is beyond the end of the file (%d lines).\n", fileResult.StartLine, fileResult.TotalLines))
			}
		}
	}

	return result.String()
}

func formatReadmeFiles(readmeFiles []ReadmeFileInfo, recursive bool) string {
	var result strings.Builder

	searchMode := "root directory only"
	if recursive {
		searchMode = "recursive search"
	}

	result.WriteString(fmt.Sprintf("README Files (%s, %d found):\n", searchMode, len(readmeFiles)))
	result.WriteString(strings.Repeat("=", 50) + "\n\n")

	if len(readmeFiles) == 0 {
		result.WriteString("No README files found in the repository.\n")
		return result.String()
	}

	for _, readme := range readmeFiles {
		result.WriteString(fmt.Sprintf("📄 %s\n", readme.Path))
		if readme.Size > 0 {
			var sizeStr string
			if readme.Size < 1024 {
				sizeStr = fmt.Sprintf("%d bytes", readme.Size)
			} else if readme.Size < 1024*1024 {
				sizeStr = fmt.Sprintf("%.1f KB", float64(readme.Size)/1024)
			} else {
				sizeStr = fmt.Sprintf("%.1f MB", float64(readme.Size)/(1024*1024))
			}
			result.WriteString(fmt.Sprintf("   Size: %s", sizeStr))
			
			if readme.LineCount > 0 {
				result.WriteString(fmt.Sprintf(" | Lines: %d", readme.LineCount))
			}
			result.WriteString("\n")
		}
		if !readme.ModTime.IsZero() {
			result.WriteString(fmt.Sprintf("   Modified: %s\n", readme.ModTime.Format("2006-01-02 15:04:05")))
		}
		result.WriteString("\n")
	}

	return result.String()
}
