package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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

func handleListCommits(ctx context.Context, req *mcp.CallToolRequest, args ListCommitsParams) (*mcp.CallToolResult, any, error) {
	if args.Repository == "" {
		return errorResult("Error: repository path is required")
	}

	limit := args.Limit
	if limit == 0 {
		limit = 20
	}

	commits, err := ListCommits(args.Repository, limit)
	if err != nil {
		return errorResult("Failed to list commits: %v", err)
	}

	return textResult(formatCommits(commits, limit))
}

func handleGetCommitDiff(ctx context.Context, req *mcp.CallToolRequest, args GetCommitDiffParams) (*mcp.CallToolResult, any, error) {
	if args.Repository == "" {
		return errorResult("Error: repository path is required")
	}
	if args.CommitHash == "" {
		return errorResult("Error: commit_hash is required")
	}

	diff, err := GetCommitDiff(args.Repository, args.CommitHash)
	if err != nil {
		return errorResult("Failed to get commit diff: %v", err)
	}

	return textResult(formatCommitDiff(args.CommitHash, diff))
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
