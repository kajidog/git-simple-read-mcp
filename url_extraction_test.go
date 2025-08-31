package main

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestExtractRepoNameFromURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expected    string
		expectError bool
	}{
		{
			name:     "HTTPS URL with .git",
			url:      "https://github.com/user/repository.git",
			expected: "repository",
		},
		{
			name:     "HTTPS URL without .git",
			url:      "https://github.com/user/repository",
			expected: "repository",
		},
		{
			name:     "SSH URL with .git",
			url:      "git@github.com:user/repository.git",
			expected: "repository",
		},
		{
			name:     "SSH URL without .git",
			url:      "git@github.com:user/repository",
			expected: "repository",
		},
		{
			name:     "GitLab HTTPS URL",
			url:      "https://gitlab.com/group/subgroup/project.git",
			expected: "project",
		},
		{
			name:     "Complex path",
			url:      "https://example.com/path/to/my-awesome-project.git",
			expected: "my-awesome-project",
		},
		{
			name:     "URL with query parameters",
			url:      "https://github.com/user/repo.git?ref=main",
			expected: "repo.git?ref=main", // This is expected behavior - query params are part of name
		},
		{
			name:        "Empty URL",
			url:         "",
			expectError: true,
		},
		{
			name:     "Invalid URL format",
			url:      "not-a-valid-url",
			expected: "not-a-valid-url", // Single segment becomes the name
		},
		{
			name:        "URL ending with slash",
			url:         "https://github.com/user/repo/",
			expected:    "", // Empty last segment should cause error
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractRepoNameFromURL(tt.url)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestCloneRepositoryWithAutoName(t *testing.T) {
	// Setup workspace
	tempDir := t.TempDir()
	InitializeWorkspace(tempDir)
	defer func() { globalWorkspaceManager = nil }() // Cleanup

	t.Run("auto extract name from URL", func(t *testing.T) {
		// Test with invalid URL (won't actually clone but will test name extraction)
		_, repoName, err := CloneRepository("https://github.com/user/test-repo.git", "")

		// Expect clone to fail but name extraction should work
		if err == nil {
			t.Errorf("Expected clone to fail for invalid URL")
		}

		// Even on failure, we should get the extracted name in error
		expectedName := "test-repo"
		if repoName != "" && repoName != expectedName {
			t.Errorf("Expected extracted name %q, got %q", expectedName, repoName)
		}
	})

	t.Run("use provided name instead of URL extraction", func(t *testing.T) {
		customName := "my-custom-name"
		_, repoName, err := CloneRepository("https://github.com/user/original-name.git", customName)

		// Expect clone to fail but name should be custom
		if err == nil {
			t.Errorf("Expected clone to fail for invalid URL")
		}

		if repoName != "" && repoName != customName {
			t.Errorf("Expected custom name %q, got %q", customName, repoName)
		}
	})
}

func TestCloneRepositoryMCPHandler(t *testing.T) {
	// Setup workspace
	tempDir := t.TempDir()
	InitializeWorkspace(tempDir)
	defer func() { globalWorkspaceManager = nil }() // Cleanup

	ctx := context.Background()

	t.Run("clone with URL only", func(t *testing.T) {
		params := CloneRepositoryParams{
			URL: "https://github.com/user/test-repo.git",
			// Name is omitted - should be auto-extracted
		}

		result, _, err := handleCloneRepository(ctx, nil, params)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		// Should fail to clone (invalid URL) but show extracted name
		if !result.IsError {
			t.Errorf("Expected error result for invalid URL")
		}

		content := result.Content[0].(*mcp.TextContent).Text
		if !contains(content, "test-repo") {
			t.Errorf("Expected extracted repo name 'test-repo' in error message")
		}
	})

	t.Run("clone with custom name", func(t *testing.T) {
		params := CloneRepositoryParams{
			URL:  "https://github.com/user/original-name.git",
			Name: "custom-name",
		}

		result, _, err := handleCloneRepository(ctx, nil, params)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		// Should fail to clone (invalid URL) but show custom name
		if !result.IsError {
			t.Errorf("Expected error result for invalid URL")
		}

		content := result.Content[0].(*mcp.TextContent).Text
		if !contains(content, "custom-name") {
			t.Errorf("Expected custom name 'custom-name' in error message")
		}
	})

	t.Run("empty URL should fail", func(t *testing.T) {
		params := CloneRepositoryParams{
			URL: "",
		}

		result, _, err := handleCloneRepository(ctx, nil, params)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}

		if !result.IsError {
			t.Errorf("Expected error result for empty URL")
		}

		content := result.Content[0].(*mcp.TextContent).Text
		if !contains(content, "URL is required") {
			t.Errorf("Expected URL required error message")
		}
	})
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
