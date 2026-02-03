package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetRepositoryInfoParams parameters for get_repository_info tool
type GetRepositoryInfoParams struct {
	Repository       string   `json:"repository,omitempty"`
	IncludeMemos     bool     `json:"include_memos,omitempty"`     // Include memos associated with this repository
	MemoLimit        int      `json:"memo_limit,omitempty"`        // Limit for memo list (default: 10)
	ExcludePatterns  []string `json:"exclude_patterns,omitempty"` // File patterns to exclude from statistics
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
	Repository      string   `json:"repository,omitempty"`       // Single repository (uses session default if empty)
	Repositories    []string `json:"repositories,omitempty"`     // Multiple repositories for cross-repo search
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
	Repository string   `json:"repository"`
	FilePath   string   `json:"file_path,omitempty"`  // Single file path (for backward compatibility)
	FilePaths  []string `json:"file_paths,omitempty"` // Multiple file paths
	StartLine  int      `json:"start_line,omitempty"` // Start reading from this line (1-based, default: 1)
	EndLine    int      `json:"end_line,omitempty"`   // End line (inclusive, default: start_line + 100)
	MaxLines   int      `json:"max_lines,omitempty"`  // Deprecated: use end_line instead
}

// CloneRepositoryParams parameters for clone_repository tool
type CloneRepositoryParams struct {
	URL             string `json:"url"`
	Name            string `json:"name,omitempty"`             // Optional: will be extracted from URL if not provided
	IncludeInfo     bool   `json:"include_info,omitempty"`     // Include repository info after clone
	IncludeBranches bool   `json:"include_branches,omitempty"` // Include branch list after clone
}

// ListWorkspaceRepositoriesParams parameters for list_workspace_repositories tool
type ListWorkspaceRepositoriesParams struct {
	IncludeStatus  bool `json:"include_status,omitempty"`  // Include git status for each repo
	IncludeCommits bool `json:"include_commits,omitempty"` // Include recent commits for each repo
	CommitLimit    int  `json:"commit_limit,omitempty"`    // Number of commits to include (default: 5)
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

// SessionParams parameters for session tool (unified set/get/clear)
type SessionParams struct {
	Action                 string   `json:"action"`                            // "set", "get", or "clear"
	DefaultRepository      string   `json:"default_repository,omitempty"`      // for "set"
	DefaultIncludePatterns []string `json:"default_include_patterns,omitempty"` // for "set"
	DefaultExcludePatterns []string `json:"default_exclude_patterns,omitempty"` // for "set"
	DefaultSearchLimit     int      `json:"default_search_limit,omitempty"`     // for "set"
	DefaultListFilesLimit  int      `json:"default_list_files_limit,omitempty"` // for "set"
	DefaultMaxLines        int      `json:"default_max_lines,omitempty"`        // for "set"
	DefaultCommitLimit     int      `json:"default_commit_limit,omitempty"`     // for "set"
}

// BatchParams parameters for batch tool (unified clone/pull/status)
type BatchParams struct {
	Operation    string              `json:"operation"`              // "clone", "pull", or "status"
	URLs         []string            `json:"urls,omitempty"`         // for "clone" - list of URLs to clone
	Repositories []string            `json:"repositories,omitempty"` // for "pull"/"status" - empty = all repos
}

// BatchResult result for batch operations
type BatchResult struct {
	Name      string `json:"name"`
	URL       string `json:"url,omitempty"`
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	Branch    string `json:"branch,omitempty"`
	HasChanges bool   `json:"has_changes,omitempty"`
	Error     string `json:"error,omitempty"`
}

// RepositoryOverview overview of a single repository (used by list_repositories)
type RepositoryOverview struct {
	Name          string   `json:"name"`
	CurrentBranch string   `json:"current_branch"`
	HasChanges    bool     `json:"has_changes"`
	RemoteURL     string   `json:"remote_url,omitempty"`
	RecentCommits []Commit `json:"recent_commits,omitempty"`
	BranchCount   int      `json:"branch_count,omitempty"`
	Error         string   `json:"error,omitempty"`
}

// RepoSearchResult result for cross-repository search (used by search_files)
type RepoSearchResult struct {
	Repository string         `json:"repository"`
	Results    []SearchResult `json:"results"`
	TotalCount int            `json:"total_count"`
	Error      string         `json:"error,omitempty"`
}

// RegisterGitTools registers all Git-related MCP tools
func RegisterGitTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_repository_info",
		Description: "Get repo info. Can include files, READMEs, memos via flags.",
	}, handleGetRepositoryInfo)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "pull_repository",
		Description: "Git pull on repository",
	}, handlePullRepository)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_branches",
		Description: "List branches in repository",
	}, handleListBranches)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "switch_branch",
		Description: "Switch to branch",
	}, handleSwitchBranch)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_files",
		Description: "Search files by keywords. Cross-repo via repositories array.",
	}, handleSearchFiles)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_files",
		Description: "List files in directory with pattern filtering",
	}, handleListFiles)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_file_content",
		Description: "Get file content with line range support",
	}, handleGetFileContent)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "clone_repository",
		Description: "Clone repo. Can include info and branches.",
	}, handleCloneRepository)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_repositories",
		Description: "List workspace repos with optional status/commits",
	}, handleListWorkspaceRepositories)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "remove_repository",
		Description: "Remove repository from workspace",
	}, handleRemoveRepository)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_readme_files",
		Description: "Find README files in repository",
	}, handleGetReadmeFiles)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_commits",
		Description: "List commit history",
	}, handleListCommits)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_commit_diff",
		Description: "Get diff for a commit",
	}, handleGetCommitDiff)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "session",
		Description: "Session config: action=set/get/clear. Set defaults for repo, patterns, limits.",
	}, handleSession)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "batch",
		Description: "Batch ops: operation=clone/pull/status on multiple repos",
	}, handleBatch)
}

func handleGetRepositoryInfo(ctx context.Context, req *mcp.CallToolRequest, args GetRepositoryInfoParams) (*mcp.CallToolResult, any, error) {
	sc := GetSessionConfig()
	repository := sc.GetRepository(args.Repository)

	if repository == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: repository required (no default set)"}},
			IsError: true,
		}, nil, nil
	}

	var result strings.Builder

	// Get basic repository info
	info, err := GetRepositoryInfo(repository)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to get repository info: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	result.WriteString(fmt.Sprintf("Repository: %s\n", info.Path))
	result.WriteString(strings.Repeat("=", 50) + "\n\n")
	result.WriteString(fmt.Sprintf("Branch: %s\n", info.CurrentBranch))
	if !info.LastUpdate.IsZero() {
		result.WriteString(fmt.Sprintf("Updated: %s\n", info.LastUpdate.Format("2006-01-02 15:04:05")))
	}
	if info.RemoteURL != "" {
		result.WriteString(fmt.Sprintf("Remote: %s\n", info.RemoteURL))
	}
	if info.License != "" {
		result.WriteString(fmt.Sprintf("License: %s\n", info.License))
	}

	// File statistics (always shown)
	excludePatterns := sc.GetExcludePatterns(args.ExcludePatterns)
	stats, err := GetFileStatistics(repository, excludePatterns)
	if err == nil {
		result.WriteString(fmt.Sprintf("\nFiles: %d, Dirs: %d\n", stats.TotalFiles, stats.TotalDirs))

		// Sort extensions by count and show top ones
		if len(stats.ExtensionCounts) > 0 {
			result.WriteString("Extensions: ")
			type extCount struct {
				ext   string
				count int
			}
			var sorted []extCount
			for ext, count := range stats.ExtensionCounts {
				sorted = append(sorted, extCount{ext, count})
			}
			// Sort by count descending
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[j].count > sorted[i].count {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
			// Show top 10
			shown := 0
			others := 0
			for _, ec := range sorted {
				if shown < 10 {
					if shown > 0 {
						result.WriteString(", ")
					}
					result.WriteString(fmt.Sprintf("%s(%d)", ec.ext, ec.count))
					shown++
				} else {
					others += ec.count
				}
			}
			if others > 0 {
				result.WriteString(fmt.Sprintf(", others(%d)", others))
			}
			result.WriteString("\n")
		}
	}

	// Include memos associated with this repository
	if args.IncludeMemos {
		result.WriteString("\n## Memos\n")
		store := GetMemoStore()
		if store == nil {
			result.WriteString("  Error: memo store not initialized\n")
		} else {
			memoLimit := args.MemoLimit
			if memoLimit == 0 {
				memoLimit = 10 // Default limit
			}
			memos := store.GetMemosByRepository(repository, memoLimit)
			if len(memos) == 0 {
				result.WriteString("  No memos found for this repository\n")
			} else {
				for _, memo := range memos {
					result.WriteString(fmt.Sprintf("  📝 %s\n", memo.Title))
					result.WriteString(fmt.Sprintf("     ID: %s\n", memo.ID))
					if len(memo.Tags) > 0 {
						result.WriteString(fmt.Sprintf("     Tags: %s\n", strings.Join(memo.Tags, ", ")))
					}
					// Show content preview
					preview := memo.Content
					if len(preview) > 80 {
						preview = preview[:80] + "..."
					}
					result.WriteString(fmt.Sprintf("     Preview: %s\n", strings.ReplaceAll(preview, "\n", " ")))
				}
				if len(memos) == memoLimit {
					result.WriteString(fmt.Sprintf("  (Limited to %d memos)\n", memoLimit))
				}
			}
		}
	}

	// Include main README content (always shown if available)
	if info.ReadmeContent != "" {
		result.WriteString("\n## README\n")
		result.WriteString(strings.Repeat("-", 30) + "\n")
		result.WriteString(info.ReadmeContent)
		if !strings.HasSuffix(info.ReadmeContent, "\n") {
			result.WriteString("\n")
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
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
	if len(args.Keywords) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: at least one keyword is required"}},
			IsError: true,
		}, nil, nil
	}

	sc := GetSessionConfig()

	// Default limit to prevent token overflow
	limit := sc.GetSearchLimit(args.Limit)

	// Default search mode to "and"
	searchMode := args.SearchMode
	if searchMode == "" {
		searchMode = "and"
	}

	includePatterns := sc.GetIncludePatterns(args.IncludePatterns)
	excludePatterns := sc.GetExcludePatterns(args.ExcludePatterns)

	// Multi-repository search if repositories array is provided
	if len(args.Repositories) > 0 {
		var allResults []RepoSearchResult

		for _, repoName := range args.Repositories {
			repoResult := RepoSearchResult{Repository: repoName}
			results, err := SearchFiles(repoName, args.Keywords, searchMode, args.IncludeFilename, args.ContextLines, includePatterns, excludePatterns, limit)
			if err != nil {
				repoResult.Error = err.Error()
			} else {
				repoResult.Results = results
				repoResult.TotalCount = len(results)
			}
			allResults = append(allResults, repoResult)
		}

		resultText := formatMultiRepoSearchResults(allResults, args.Keywords, searchMode)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
		}, nil, nil
	}

	// Single repository search
	repository := sc.GetRepository(args.Repository)
	if repository == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: repository required (no default set)"}},
			IsError: true,
		}, nil, nil
	}

	results, err := SearchFiles(repository, args.Keywords, searchMode, args.IncludeFilename, args.ContextLines, includePatterns, excludePatterns, limit)
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

	// Default values
	startLine := args.StartLine
	if startLine < 1 {
		startLine = 1
	}

	// Calculate maxLines: end_line > max_lines > session default > 100
	var maxLines int
	if args.EndLine > 0 {
		if args.EndLine < startLine {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error: end_line (%d) must be >= start_line (%d)", args.EndLine, startLine)}},
				IsError: true,
			}, nil, nil
		}
		maxLines = args.EndLine - startLine + 1
	} else {
		// Use session config default (falls back to 100 if not set)
		maxLines = GetSessionConfig().GetMaxLines(args.MaxLines)
	}

	showLineNumbers := true

	if len(filePaths) == 1 {
		// Single file
		content, totalLines, actualStart, actualEnd, err := GetFileContentWithLineNumbers(args.Repository, filePaths[0], startLine, maxLines, showLineNumbers)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("[%s ERR:%v]", filePaths[0], err)}},
				IsError: true,
			}, nil, nil
		}

		resultText := fmt.Sprintf("[%s L%d-%d/%d]\n%s", filePaths[0], actualStart, actualEnd, totalLines, content)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
		}, nil, nil
	} else {
		// Multiple files
		results, err := GetMultipleFileContentsWithLineNumbers(args.Repository, filePaths, startLine, maxLines, showLineNumbers)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("ERR:%v", err)}},
				IsError: true,
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

	var result strings.Builder
	var repoName string
	var cloneSuccess bool

	output, actualName, err := CloneRepository(args.URL, args.Name)
	repoName = actualName

	if err != nil {
		if strings.Contains(err.Error(), "already exists in workspace") {
			pullOutput, pullErr := PullRepository(actualName)
			if pullErr != nil {
				errorMsg := fmt.Sprintf("Repository '%s' already exists but pull failed: %v\nOutput: %s", actualName, pullErr, pullOutput)
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{Text: errorMsg}},
					IsError: true,
				}, nil, nil
			}
			result.WriteString(fmt.Sprintf("Repository '%s' already exists. Pulled latest:\n%s\n", actualName, strings.TrimSpace(pullOutput)))
			cloneSuccess = true
		} else {
			var errorMsg string
			if args.Name == "" {
				errorMsg = fmt.Sprintf("Clone failed for '%s' (from URL): %v\nOutput: %s", actualName, err, output)
			} else {
				errorMsg = fmt.Sprintf("Clone failed for '%s': %v\nOutput: %s", actualName, err, output)
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: errorMsg}},
				IsError: true,
			}, nil, nil
		}
	} else {
		if args.Name == "" {
			result.WriteString(fmt.Sprintf("Cloned as '%s' (from URL):\n%s\n", actualName, strings.TrimSpace(output)))
		} else {
			result.WriteString(fmt.Sprintf("Cloned '%s':\n%s\n", actualName, strings.TrimSpace(output)))
		}
		cloneSuccess = true
	}

	// Include repository info if requested
	if cloneSuccess && args.IncludeInfo {
		result.WriteString("\n## Repository Info\n")
		info, err := GetRepositoryInfo(repoName)
		if err != nil {
			result.WriteString(fmt.Sprintf("Error: %v\n", err))
		} else {
			result.WriteString(fmt.Sprintf("Branch: %s\n", info.CurrentBranch))
			if info.RemoteURL != "" {
				result.WriteString(fmt.Sprintf("Remote: %s\n", info.RemoteURL))
			}
		}
	}

	// Include branches if requested
	if cloneSuccess && args.IncludeBranches {
		result.WriteString("\n## Branches\n")
		branches, err := ListBranches(repoName)
		if err != nil {
			result.WriteString(fmt.Sprintf("Error: %v\n", err))
		} else {
			for _, branch := range branches {
				if branch.IsCurrent {
					result.WriteString(fmt.Sprintf("  * %s (current)\n", branch.Name))
				} else {
					result.WriteString(fmt.Sprintf("    %s\n", branch.Name))
				}
			}
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
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

	// Extended mode: include status and/or commits
	if args.IncludeStatus || args.IncludeCommits {
		sc := GetSessionConfig()
		commitLimit := args.CommitLimit
		if commitLimit == 0 {
			commitLimit = sc.GetCommitLimit(5)
		}

		var overviews []RepositoryOverview
		for _, repoName := range repositories {
			overview := RepositoryOverview{Name: repoName}

			if args.IncludeStatus {
				status, err := GetRepositoryStatus(repoName)
				if err != nil {
					overview.Error = err.Error()
				} else {
					overview.CurrentBranch = status.CurrentBranch
					overview.HasChanges = status.HasChanges
				}

				branches, err := ListBranches(repoName)
				if err == nil {
					overview.BranchCount = len(branches)
				}

				info, err := GetRepositoryInfo(repoName)
				if err == nil {
					overview.RemoteURL = info.RemoteURL
				}
			}

			if args.IncludeCommits {
				commits, err := ListCommits(repoName, commitLimit)
				if err == nil {
					overview.RecentCommits = commits
				}
			}

			overviews = append(overviews, overview)
		}

		resultText := formatWorkspaceOverview(wm.GetWorkspaceDir(), overviews, args.IncludeCommits)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
		}, nil, nil
	}

	// Simple mode: just list repository names
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

	result.WriteString(fmt.Sprintf("Files in '%s' (%s, %d files):\n", directory, modeStr, len(files)))
	result.WriteString(strings.Repeat("-", 50) + "\n")

	for _, file := range files {
		infoStr := ""
		if file.Size > 0 || file.LineCount > 0 {
			var parts []string

			// Add file size
			if file.Size > 0 {
				if file.Size < 1024 {
					parts = append(parts, fmt.Sprintf("%dB", file.Size))
				} else if file.Size < 1024*1024 {
					parts = append(parts, fmt.Sprintf("%.1fKB", float64(file.Size)/1024))
				} else {
					parts = append(parts, fmt.Sprintf("%.1fMB", float64(file.Size)/(1024*1024)))
				}
			}

			// Add line count
			if file.LineCount > 0 {
				parts = append(parts, fmt.Sprintf("%dL", file.LineCount))
			}

			if len(parts) > 0 {
				infoStr = fmt.Sprintf(" (%s)", strings.Join(parts, ", "))
			}
		}
		result.WriteString(fmt.Sprintf("%s%s\n", file.Path, infoStr))
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

func formatMultipleFileContents(results []FileContentResult) string {
	var result strings.Builder

	for i, fileResult := range results {
		if i > 0 {
			result.WriteString("\n")
		}

		if fileResult.Error != "" {
			result.WriteString(fmt.Sprintf("[%s ERR:%s]\n", fileResult.FilePath, fileResult.Error))
		} else {
			result.WriteString(fmt.Sprintf("[%s L%d-%d/%d]\n", fileResult.FilePath, fileResult.StartLine, fileResult.EndLine, fileResult.TotalLines))
			result.WriteString(fileResult.Content)
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

// Unified session handler

func handleSession(ctx context.Context, req *mcp.CallToolRequest, args SessionParams) (*mcp.CallToolResult, any, error) {
	switch args.Action {
	case "set":
		config := &SessionConfig{
			DefaultRepository:      args.DefaultRepository,
			DefaultIncludePatterns: args.DefaultIncludePatterns,
			DefaultExcludePatterns: args.DefaultExcludePatterns,
			DefaultSearchLimit:     args.DefaultSearchLimit,
			DefaultListFilesLimit:  args.DefaultListFilesLimit,
			DefaultMaxLines:        args.DefaultMaxLines,
			DefaultCommitLimit:     args.DefaultCommitLimit,
		}
		SetSessionConfigValues(config)

		var result strings.Builder
		result.WriteString("Session configuration updated:\n")
		result.WriteString(strings.Repeat("-", 30) + "\n")
		if args.DefaultRepository != "" {
			result.WriteString(fmt.Sprintf("default_repository: %s\n", args.DefaultRepository))
		}
		if len(args.DefaultIncludePatterns) > 0 {
			result.WriteString(fmt.Sprintf("default_include_patterns: %v\n", args.DefaultIncludePatterns))
		}
		if len(args.DefaultExcludePatterns) > 0 {
			result.WriteString(fmt.Sprintf("default_exclude_patterns: %v\n", args.DefaultExcludePatterns))
		}
		if args.DefaultSearchLimit > 0 {
			result.WriteString(fmt.Sprintf("default_search_limit: %d\n", args.DefaultSearchLimit))
		}
		if args.DefaultListFilesLimit > 0 {
			result.WriteString(fmt.Sprintf("default_list_files_limit: %d\n", args.DefaultListFilesLimit))
		}
		if args.DefaultMaxLines > 0 {
			result.WriteString(fmt.Sprintf("default_max_lines: %d\n", args.DefaultMaxLines))
		}
		if args.DefaultCommitLimit > 0 {
			result.WriteString(fmt.Sprintf("default_commit_limit: %d\n", args.DefaultCommitLimit))
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
		}, nil, nil

	case "get":
		sc := GetSessionConfig()
		if sc.IsEmpty() {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "No session configuration set.\nUse action='set' with: default_repository, default_include_patterns, default_exclude_patterns, default_search_limit, default_list_files_limit, default_max_lines, default_commit_limit"}},
			}, nil, nil
		}
		var result strings.Builder
		result.WriteString("Current Session Configuration:\n")
		result.WriteString(strings.Repeat("=", 50) + "\n\n")
		configMap := sc.ToMap()
		for key, value := range configMap {
			result.WriteString(fmt.Sprintf("%s: %v\n", key, value))
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
		}, nil, nil

	case "clear":
		ClearSessionConfig()
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Session configuration cleared. All values reset to defaults."}},
		}, nil, nil

	default:
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: action must be 'set', 'get', or 'clear'"}},
			IsError: true,
		}, nil, nil
	}
}

// Unified batch handler

func handleBatch(ctx context.Context, req *mcp.CallToolRequest, args BatchParams) (*mcp.CallToolResult, any, error) {
	wm := GetWorkspaceManager()
	if wm == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: workspace not initialized"}},
			IsError: true,
		}, nil, nil
	}

	switch args.Operation {
	case "clone":
		if len(args.URLs) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "Error: urls array is required for clone operation"}},
				IsError: true,
			}, nil, nil
		}

		var results []BatchResult
		for _, url := range args.URLs {
			result := BatchResult{URL: url}
			output, actualName, err := CloneRepository(url, "")
			result.Name = actualName

			if err != nil {
				if strings.Contains(err.Error(), "already exists in workspace") {
					pullOutput, pullErr := PullRepository(actualName)
					if pullErr != nil {
						result.Success = false
						result.Error = fmt.Sprintf("Already exists but pull failed: %v", pullErr)
					} else {
						result.Success = true
						result.Message = fmt.Sprintf("Already exists, pulled: %s", strings.TrimSpace(pullOutput))
					}
				} else {
					result.Success = false
					result.Error = err.Error()
				}
			} else {
				result.Success = true
				result.Message = fmt.Sprintf("Cloned: %s", strings.TrimSpace(output))
			}
			results = append(results, result)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: formatBatchResults("clone", results)}},
		}, nil, nil

	case "pull":
		repos := args.Repositories
		if len(repos) == 0 {
			var err error
			repos, err = wm.ListRepositories()
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list repositories: %v", err)}},
					IsError: true,
				}, nil, nil
			}
		}

		var results []BatchResult
		for _, repoName := range repos {
			result := BatchResult{Name: repoName}
			output, err := PullRepository(repoName)
			if err != nil {
				result.Success = false
				result.Error = err.Error()
			} else {
				result.Success = true
				result.Message = strings.TrimSpace(output)
			}
			results = append(results, result)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: formatBatchResults("pull", results)}},
		}, nil, nil

	case "status":
		repos := args.Repositories
		if len(repos) == 0 {
			var err error
			repos, err = wm.ListRepositories()
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to list repositories: %v", err)}},
					IsError: true,
				}, nil, nil
			}
		}

		var results []BatchResult
		for _, repoName := range repos {
			result := BatchResult{Name: repoName}
			status, err := GetRepositoryStatus(repoName)
			if err != nil {
				result.Error = err.Error()
			} else {
				result.Success = true
				result.Branch = status.CurrentBranch
				result.HasChanges = status.HasChanges
			}
			results = append(results, result)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: formatBatchResults("status", results)}},
		}, nil, nil

	default:
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: operation must be 'clone', 'pull', or 'status'"}},
			IsError: true,
		}, nil, nil
	}
}

// Formatting functions for batch operations

func formatBatchResults(operation string, results []BatchResult) string {
	var sb strings.Builder

	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}

	sb.WriteString(fmt.Sprintf("Batch %s (%d/%d successful):\n", operation, successCount, len(results)))
	sb.WriteString(strings.Repeat("-", 40) + "\n")

	for _, r := range results {
		name := r.Name
		if r.URL != "" && name == "" {
			name = r.URL
		}

		switch operation {
		case "clone":
			if r.Success {
				sb.WriteString(fmt.Sprintf("✓ %s → %s\n", r.URL, r.Name))
				if r.Message != "" {
					sb.WriteString(fmt.Sprintf("  %s\n", r.Message))
				}
			} else {
				sb.WriteString(fmt.Sprintf("✗ %s: %s\n", r.URL, r.Error))
			}
		case "pull":
			if r.Success {
				sb.WriteString(fmt.Sprintf("✓ %s: %s\n", name, r.Message))
			} else {
				sb.WriteString(fmt.Sprintf("✗ %s: %s\n", name, r.Error))
			}
		case "status":
			if r.Error != "" {
				sb.WriteString(fmt.Sprintf("✗ %s: %s\n", name, r.Error))
			} else {
				changeStatus := "clean"
				if r.HasChanges {
					changeStatus = "changes"
				}
				sb.WriteString(fmt.Sprintf("📁 %s [%s] (%s)\n", name, r.Branch, changeStatus))
			}
		}
	}

	return sb.String()
}

func formatWorkspaceOverview(workspaceDir string, overviews []RepositoryOverview, includeCommits bool) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Workspace Overview (%s)\n", workspaceDir))
	result.WriteString(strings.Repeat("=", 50) + "\n")
	result.WriteString(fmt.Sprintf("Total repositories: %d\n\n", len(overviews)))

	if len(overviews) == 0 {
		result.WriteString("No repositories found in workspace.\n")
		return result.String()
	}

	for _, o := range overviews {
		if o.Error != "" {
			result.WriteString(fmt.Sprintf("📁 %s: Error - %s\n\n", o.Name, o.Error))
			continue
		}

		changeStatus := "✓"
		if o.HasChanges {
			changeStatus = "●"
		}

		result.WriteString(fmt.Sprintf("📁 %s %s\n", o.Name, changeStatus))
		result.WriteString(fmt.Sprintf("   Branch: %s", o.CurrentBranch))
		if o.BranchCount > 0 {
			result.WriteString(fmt.Sprintf(" (%d total)", o.BranchCount))
		}
		result.WriteString("\n")

		if o.RemoteURL != "" {
			result.WriteString(fmt.Sprintf("   Remote: %s\n", o.RemoteURL))
		}

		if includeCommits && len(o.RecentCommits) > 0 {
			result.WriteString("   Recent commits:\n")
			for _, c := range o.RecentCommits {
				shortHash := c.Hash
				if len(shortHash) > 7 {
					shortHash = shortHash[:7]
				}
				msg := c.Message
				if len(msg) > 50 {
					msg = msg[:47] + "..."
				}
				result.WriteString(fmt.Sprintf("     %s %s\n", shortHash, msg))
			}
		}
		result.WriteString("\n")
	}

	return result.String()
}

func formatMultiRepoSearchResults(results []RepoSearchResult, keywords []string, searchMode string) string {
	var sb strings.Builder

	totalMatches := 0
	reposWithMatches := 0
	for _, r := range results {
		if r.TotalCount > 0 {
			totalMatches += r.TotalCount
			reposWithMatches++
		}
	}

	var keywordStr string
	if searchMode == "or" {
		keywordStr = strings.Join(keywords, " OR ")
	} else {
		keywordStr = strings.Join(keywords, " AND ")
	}

	sb.WriteString(fmt.Sprintf("Multi-Repository Search: %s\n", keywordStr))
	sb.WriteString(strings.Repeat("=", 50) + "\n")
	sb.WriteString(fmt.Sprintf("Found %d matches in %d/%d repositories\n\n", totalMatches, reposWithMatches, len(results)))

	for _, r := range results {
		if r.Error != "" {
			sb.WriteString(fmt.Sprintf("📁 %s: Error - %s\n\n", r.Repository, r.Error))
			continue
		}

		if r.TotalCount == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("📁 %s (%d matches)\n", r.Repository, r.TotalCount))

		for _, searchResult := range r.Results {
			sb.WriteString(fmt.Sprintf("  📄 %s\n", searchResult.Path))
			if len(searchResult.Matches) > 0 {
				for _, match := range searchResult.Matches {
					if match.LineNumber > 0 {
						sb.WriteString(fmt.Sprintf("     L%d: %s\n", match.LineNumber, strings.TrimSpace(match.Content)))
					}
				}
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
