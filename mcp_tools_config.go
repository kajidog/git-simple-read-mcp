package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SessionParams parameters for session tool (unified set/get/clear)
type SessionParams struct {
	Action                 string   `json:"action"`                              // "set", "get", or "clear"
	DefaultRepository      string   `json:"default_repository,omitempty"`        // for "set"
	DefaultIncludePatterns []string `json:"default_include_patterns,omitempty"`  // for "set"
	DefaultExcludePatterns []string `json:"default_exclude_patterns,omitempty"`  // for "set"
	DefaultSearchLimit     int      `json:"default_search_limit,omitempty"`      // for "set": cap on search_files results
	DefaultListFilesLimit  int      `json:"default_list_files_limit,omitempty"`  // for "set": cap on list_files results
	DefaultLineLimit       int      `json:"default_line_limit,omitempty"`        // for "set": default span for get_file_content
	DefaultCommitLimit     int      `json:"default_commit_limit,omitempty"`      // for "set": cap on list_commits results
}

// BatchParams parameters for batch tool (unified clone/pull/status)
type BatchParams struct {
	Operation    string   `json:"operation"`              // "clone", "pull", or "status"
	URLs         []string `json:"urls,omitempty"`         // for "clone" - list of URLs to clone
	Repositories []string `json:"repositories,omitempty"` // for "pull"/"status" - empty = all repos
}

// BatchResult result for batch operations
type BatchResult struct {
	Name       string `json:"name"`
	URL        string `json:"url,omitempty"`
	Success    bool   `json:"success"`
	Message    string `json:"message,omitempty"`
	Branch     string `json:"branch,omitempty"`
	HasChanges bool   `json:"has_changes,omitempty"`
	Error      string `json:"error,omitempty"`
}

func handleSession(ctx context.Context, req *mcp.CallToolRequest, args SessionParams) (*mcp.CallToolResult, any, error) {
	switch args.Action {
	case "set":
		config := &SessionConfig{
			DefaultRepository:      args.DefaultRepository,
			DefaultIncludePatterns: args.DefaultIncludePatterns,
			DefaultExcludePatterns: args.DefaultExcludePatterns,
			DefaultSearchLimit:     args.DefaultSearchLimit,
			DefaultListFilesLimit:  args.DefaultListFilesLimit,
			DefaultLineLimit:       args.DefaultLineLimit,
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
		if args.DefaultLineLimit > 0 {
			result.WriteString(fmt.Sprintf("default_line_limit: %d\n", args.DefaultLineLimit))
		}
		if args.DefaultCommitLimit > 0 {
			result.WriteString(fmt.Sprintf("default_commit_limit: %d\n", args.DefaultCommitLimit))
		}
		return textResult(result.String())

	case "get":
		sc := GetSessionConfig()
		if sc.IsEmpty() {
			return textResult("No session configuration set.\nUse action='set' with: default_repository, default_include_patterns, default_exclude_patterns, default_search_limit, default_list_files_limit, default_line_limit, default_commit_limit")
		}
		var result strings.Builder
		result.WriteString("Current Session Configuration:\n")
		result.WriteString(strings.Repeat("=", 50) + "\n\n")
		configMap := sc.ToMap()
		for key, value := range configMap {
			result.WriteString(fmt.Sprintf("%s: %v\n", key, value))
		}
		return textResult(result.String())

	case "clear":
		ClearSessionConfig()
		return textResult("Session configuration cleared. All values reset to defaults.")

	default:
		return errorResult("Error: action must be 'set', 'get', or 'clear'")
	}
}

func handleBatch(ctx context.Context, req *mcp.CallToolRequest, args BatchParams) (*mcp.CallToolResult, any, error) {
	wm := GetWorkspaceManager()
	if wm == nil {
		return errorResult("Error: workspace not initialized")
	}

	switch args.Operation {
	case "clone":
		if len(args.URLs) == 0 {
			return errorResult("Error: urls array is required for clone operation")
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
		return textResult(formatBatchResults("clone", results))

	case "pull":
		repos := args.Repositories
		if len(repos) == 0 {
			var err error
			repos, err = wm.ListRepositories()
			if err != nil {
				return errorResult("Failed to list repositories: %v", err)
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
		return textResult(formatBatchResults("pull", results))

	case "status":
		repos := args.Repositories
		if len(repos) == 0 {
			var err error
			repos, err = wm.ListRepositories()
			if err != nil {
				return errorResult("Failed to list repositories: %v", err)
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
		return textResult(formatBatchResults("status", results))

	default:
		return errorResult("Error: operation must be 'clone', 'pull', or 'status'")
	}
}

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
