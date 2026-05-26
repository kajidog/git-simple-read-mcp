package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetRepositoryInfoParams parameters for get_repository_info tool
type GetRepositoryInfoParams struct {
	Repository      string   `json:"repository,omitempty"`
	IncludeMemos    bool     `json:"include_memos,omitempty"`    // Include memos associated with this repository
	MemoLimit       int      `json:"memo_limit,omitempty"`       // Limit for memo list (default: 10)
	ExcludePatterns []string `json:"exclude_patterns,omitempty"` // File patterns to exclude from statistics
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

// CloneRepositoryParams parameters for clone_repository tool
type CloneRepositoryParams struct {
	URL             string `json:"url"`
	Name            string `json:"name,omitempty"`             // Optional: will be extracted from URL if not provided
	IncludeInfo     bool   `json:"include_info,omitempty"`     // Include repository info after clone
	IncludeBranches bool   `json:"include_branches,omitempty"` // Include branch list after clone
}

// ListWorkspaceRepositoriesParams parameters for list_repositories tool
type ListWorkspaceRepositoriesParams struct {
	IncludeStatus  bool `json:"include_status,omitempty"`  // Include git status for each repo
	IncludeCommits bool `json:"include_commits,omitempty"` // Include recent commits for each repo
	CommitLimit    int  `json:"commit_limit,omitempty"`    // Number of commits to include per repo when include_commits is set (default: 5)
}

// RemoveRepositoryParams parameters for remove_repository tool
type RemoveRepositoryParams struct {
	Name string `json:"name"`
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

func handleGetRepositoryInfo(ctx context.Context, req *mcp.CallToolRequest, args GetRepositoryInfoParams) (*mcp.CallToolResult, any, error) {
	sc := GetSessionConfig()
	repository := sc.GetRepository(args.Repository)

	if repository == "" {
		return errorResult("Error: repository required (no default set)")
	}

	var result strings.Builder

	info, err := GetRepositoryInfo(repository)
	if err != nil {
		return errorResult("Failed to get repository info: %v", err)
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

	excludePatterns := sc.GetExcludePatterns(args.ExcludePatterns)
	stats, err := GetFileStatistics(repository, excludePatterns)
	if err == nil {
		result.WriteString(fmt.Sprintf("\nFiles: %d, Dirs: %d\n", stats.TotalFiles, stats.TotalDirs))
		if len(stats.ExtensionCounts) > 0 {
			result.WriteString("Extensions: ")
			result.WriteString(formatExtensionStats(stats.ExtensionCounts, 10))
			result.WriteString("\n")
		}
	}

	if args.IncludeMemos {
		result.WriteString("\n## Memos\n")
		store := GetMemoStore()
		if store == nil {
			result.WriteString("  Error: memo store not initialized\n")
		} else {
			memoLimit := args.MemoLimit
			if memoLimit == 0 {
				memoLimit = 10
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

	if info.ReadmeContent != "" {
		result.WriteString("\n## README\n")
		result.WriteString(strings.Repeat("-", 30) + "\n")
		result.WriteString(info.ReadmeContent)
		if !strings.HasSuffix(info.ReadmeContent, "\n") {
			result.WriteString("\n")
		}
	}

	return textResult(result.String())
}

func handlePullRepository(ctx context.Context, req *mcp.CallToolRequest, args PullRepositoryParams) (*mcp.CallToolResult, any, error) {
	if args.Repository == "" {
		return errorResult("Error: repository path is required")
	}

	output, err := PullRepository(args.Repository)
	if err != nil {
		return errorResult("Pull failed: %v\nOutput: %s", err, output)
	}
	return textResult(fmt.Sprintf("Git pull completed successfully:\n%s", output))
}

func handleListBranches(ctx context.Context, req *mcp.CallToolRequest, args ListBranchesParams) (*mcp.CallToolResult, any, error) {
	if args.Repository == "" {
		return errorResult("Error: repository path is required")
	}

	branches, err := ListBranches(args.Repository)
	if err != nil {
		return errorResult("Failed to list branches: %v", err)
	}

	if args.Limit > 0 && len(branches) > args.Limit {
		branches = branches[:args.Limit]
	}

	return textResult(formatBranches(branches, args.Limit > 0))
}

func handleSwitchBranch(ctx context.Context, req *mcp.CallToolRequest, args SwitchBranchParams) (*mcp.CallToolResult, any, error) {
	if args.Repository == "" {
		return errorResult("Error: repository path is required")
	}
	if args.Branch == "" {
		return errorResult("Error: branch name is required")
	}

	output, err := SwitchBranch(args.Repository, args.Branch)
	if err != nil {
		return errorResult("Branch switch failed: %v\nOutput: %s", err, output)
	}
	return textResult(fmt.Sprintf("Successfully switched to branch '%s':\n%s", args.Branch, output))
}

func handleCloneRepository(ctx context.Context, req *mcp.CallToolRequest, args CloneRepositoryParams) (*mcp.CallToolResult, any, error) {
	if args.URL == "" {
		return errorResult("Error: repository URL is required")
	}

	var result strings.Builder
	var cloneSuccess bool

	output, repoName, err := CloneRepository(args.URL, args.Name)

	if err != nil {
		if strings.Contains(err.Error(), "already exists in workspace") {
			pullOutput, pullErr := PullRepository(repoName)
			if pullErr != nil {
				return errorResult("Repository '%s' already exists but pull failed: %v\nOutput: %s", repoName, pullErr, pullOutput)
			}
			result.WriteString(fmt.Sprintf("Repository '%s' already exists. Pulled latest:\n%s\n", repoName, strings.TrimSpace(pullOutput)))
			cloneSuccess = true
		} else {
			if args.Name == "" {
				return errorResult("Clone failed for '%s' (from URL): %v\nOutput: %s", repoName, err, output)
			}
			return errorResult("Clone failed for '%s': %v\nOutput: %s", repoName, err, output)
		}
	} else {
		if args.Name == "" {
			result.WriteString(fmt.Sprintf("Cloned as '%s' (from URL):\n%s\n", repoName, strings.TrimSpace(output)))
		} else {
			result.WriteString(fmt.Sprintf("Cloned '%s':\n%s\n", repoName, strings.TrimSpace(output)))
		}
		cloneSuccess = true
	}

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

	return textResult(result.String())
}

func handleListWorkspaceRepositories(ctx context.Context, req *mcp.CallToolRequest, args ListWorkspaceRepositoriesParams) (*mcp.CallToolResult, any, error) {
	wm := GetWorkspaceManager()
	if wm == nil {
		return errorResult("Error: workspace not initialized")
	}

	repositories, err := wm.ListRepositories()
	if err != nil {
		return errorResult("Failed to list repositories: %v", err)
	}

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

		return textResult(formatWorkspaceOverview(wm.GetWorkspaceDir(), overviews, args.IncludeCommits))
	}

	return textResult(formatWorkspaceRepositories(repositories, wm.GetWorkspaceDir()))
}

func handleRemoveRepository(ctx context.Context, req *mcp.CallToolRequest, args RemoveRepositoryParams) (*mcp.CallToolResult, any, error) {
	if args.Name == "" {
		return errorResult("Error: repository name is required")
	}

	wm := GetWorkspaceManager()
	if wm == nil {
		return errorResult("Error: workspace not initialized")
	}

	if err := wm.RemoveRepository(args.Name); err != nil {
		return errorResult("Failed to remove repository: %v", err)
	}
	return textResult(fmt.Sprintf("Successfully removed repository '%s'", args.Name))
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
