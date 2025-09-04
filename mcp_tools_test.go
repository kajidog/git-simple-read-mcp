package main

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHandleGetRepositoryInfo(t *testing.T) {
	tests := []struct {
		name        string
		args        GetRepositoryInfoParams
		expectError bool
		checkResult func(t *testing.T, result *mcp.CallToolResult)
	}{
		{
			name: "valid repository",
			args: GetRepositoryInfoParams{
				Repository: "", // Will be set in test
			},
			expectError: false,
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				if result.IsError {
					t.Errorf("Expected success but got error")
				}
				if len(result.Content) == 0 {
					t.Errorf("Expected content in result")
				}
			},
		},
		{
			name: "empty repository path",
			args: GetRepositoryInfoParams{
				Repository: "",
			},
			expectError: true,
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				if !result.IsError {
					t.Errorf("Expected error for empty repository path")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.expectError {
				repo := CreateTestRepositoryWithContent(t)
				tt.args.Repository = repo.Path
			}

			ctx := context.Background()
			result, _, err := handleGetRepositoryInfo(ctx, nil, tt.args)
			if err != nil {
				t.Fatalf("Handler returned unexpected error: %v", err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func TestHandleSearchFiles(t *testing.T) {
	repo := CreateTestRepositoryWithContent(t)

	tests := []struct {
		name        string
		args        SearchFilesParams
		expectError bool
	}{
		{
			name: "search with AND mode",
			args: SearchFilesParams{
				Repository: repo.Path,
				Keywords:   []string{"database"},
				SearchMode: "and",
				Limit:      10,
			},
			expectError: false,
		},
		{
			name: "search with OR mode",
			args: SearchFilesParams{
				Repository: repo.Path,
				Keywords:   []string{"database", "nonexistent"},
				SearchMode: "or",
				Limit:      10,
			},
			expectError: false,
		},
		{
			name: "search with default mode (and)",
			args: SearchFilesParams{
				Repository: repo.Path,
				Keywords:   []string{"database"},
				Limit:      10,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, _, err := handleSearchFiles(ctx, nil, tt.args)
			if err != nil {
				t.Fatalf("Handler returned unexpected error: %v", err)
			}

			if tt.expectError && !result.IsError {
				t.Errorf("Expected error but got success")
			}
			if !tt.expectError && result.IsError {
				content := result.Content[0].(*mcp.TextContent).Text
				t.Errorf("Expected success but got error: %s", content)
			}
		})
	}
}

// Test the formatSearchResults function with different modes
func TestFormatSearchResults(t *testing.T) {
	results := []SearchResult{
		{Path: "file1.go"},
		{Path: "file2.go"},
	}
	keywords := []string{"keyword1", "keyword2"}

	// Test AND mode
	andResult := formatSearchResults(results, keywords, "and")
	if !strings.Contains(andResult, "keyword1 AND keyword2") {
		t.Errorf("AND mode result should contain 'AND': %s", andResult)
	}

	// Test OR mode
	orResult := formatSearchResults(results, keywords, "or")
	if !strings.Contains(orResult, "keyword1 OR keyword2") {
		t.Errorf("OR mode result should contain 'OR': %s", orResult)
	}
}

func TestHandleSearchFilesEnhanced(t *testing.T) {
	tests := []struct {
		name        string
		args        SearchFilesParams
		expectError bool
		checkResult func(t *testing.T, result *mcp.CallToolResult)
	}{
		{
			name: "filename search",
			args: SearchFilesParams{
				Repository:      "test-repo",
				Keywords:        []string{"main"},
				SearchMode:      "and",
				IncludeFilename: true,
				ContextLines:    0,
				Limit:           10,
			},
			expectError: false,
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				if result.IsError {
					t.Errorf("Expected success but got error: %v", result.Content)
				}
				// Should find main.go filename match
				content := result.Content[0].(*mcp.TextContent).Text
				if !strings.Contains(content, "main.go") || strings.Contains(content, "No files found") {
					t.Errorf("Expected main.go filename match results, got: %s", content)
				}
			},
		},
		{
			name: "content search with context lines",
			args: SearchFilesParams{
				Repository:      "test-repo",
				Keywords:        []string{"func"},
				SearchMode:      "and",
				IncludeFilename: false,
				ContextLines:    2,
				Limit:           10,
			},
			expectError: false,
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				if result.IsError {
					t.Errorf("Expected success but got error: %v", result.Content)
				}
				content := result.Content[0].(*mcp.TextContent).Text
				// Should contain line numbers and context
				if strings.Contains(content, "No files found") {
					t.Errorf("Search didn't find 'func' in test repository, got: %s", content)
				} else if !strings.Contains(content, "Line") {
					t.Errorf("Expected line number information, got: %s", content)
				}
			},
		},
		{
			name: "OR search mode",
			args: SearchFilesParams{
				Repository:      "test-repo",
				Keywords:        []string{"func", "import", "nonexistent"},
				SearchMode:      "or",
				IncludeFilename: false,
				ContextLines:    0,
				Limit:           10,
			},
			expectError: false,
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				if result.IsError {
					t.Errorf("Expected success but got error: %v", result.Content)
				}
				// Should find matches even though one keyword doesn't exist
				content := result.Content[0].(*mcp.TextContent).Text
				if strings.Contains(content, "No files found") {
					t.Errorf("Expected to find results with OR search, got: %s", content)
				}
			},
		},
		{
			name: "combined filename and content search",
			args: SearchFilesParams{
				Repository:      "test-repo",
				Keywords:        []string{"main"},
				SearchMode:      "and",
				IncludeFilename: true,
				ContextLines:    1,
				Limit:           10,
			},
			expectError: false,
			checkResult: func(t *testing.T, result *mcp.CallToolResult) {
				if result.IsError {
					t.Errorf("Expected success but got error: %v", result.Content)
				}
				content := result.Content[0].(*mcp.TextContent).Text
				// Should contain both filename and content matches
				if strings.Contains(content, "No files found") {
					t.Errorf("Expected search results for 'main', got: %s", content)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test has a flawed setup. The CreateTestRepositoryWithContent helper
			// pollutes the global workspace. We work around this by creating our own
			// clean workspace directory and re-initializing the workspace manager
			// after the helper has run.
			workspaceDir := t.TempDir()
			repo := CreateTestRepositoryWithContent(t) // This pollutes the global workspace.

			// Re-initialize with our clean directory.
			InitializeWorkspace(workspaceDir)
			defer func() { globalWorkspaceManager = nil }() // Cleanup

			// Now we can clone from the 'remote' repo path into our clean workspace.
			_, repoName, err := CloneRepository(repo.Path, "test-repo")
			if err != nil {
				t.Fatalf("Failed to clone repo into workspace: %v", err)
			}

			tt.args.Repository = repoName

			ctx := context.Background()
			result, _, err := handleSearchFiles(ctx, nil, tt.args)

			if (err != nil) != tt.expectError {
				t.Errorf("Expected error status %v but got err = %v", tt.expectError, err)
			}

			if result != nil && tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}
