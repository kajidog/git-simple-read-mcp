package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHandleGetFileContent(t *testing.T) {
	repo := CreateTestRepositoryWithContent(t)
	testFile := "test_file_for_reading.txt"
	var fileContent strings.Builder
	for i := 1; i <= 10; i++ {
		fileContent.WriteString(fmt.Sprintf("Line %d\n", i))
	}
	repo.WriteFile(testFile, fileContent.String())
	repo.AddCommit("Add test file for reading")

	tests := []struct {
		name              string
		args              GetFileContentParams
		expectError       bool
		expectedInOutput  []string
		notExpectedInOutput []string
	}{
		{
			name: "single file with start_line",
			args: GetFileContentParams{
				Repository: "test-repo",
				FilePath:   testFile,
				StartLine:  5,
				MaxLines:   2,
			},
			expectError: false,
			expectedInOutput: []string{
				fmt.Sprintf("Content of %s (lines 5-6 of 10)", testFile),
				"Line 5",
				"Line 6",
			},
			notExpectedInOutput: []string{"Line 4", "Line 7"},
		},
		{
			name: "multiple files with start_line",
			args: GetFileContentParams{
				Repository: "test-repo",
				FilePaths:  []string{testFile, "README.md"},
				StartLine:  2,
				MaxLines:   1,
			},
			expectError: false,
			expectedInOutput: []string{
				fmt.Sprintf("Content of 2 files"),
				fmt.Sprintf("📄 %s (lines 2-2 of 10)", testFile),
				"Line 2",
				"📄 README.md (lines 2-2 of 3)", // README has 3 lines
			},
			notExpectedInOutput: []string{
				"Line 1",
				"This is a test repository for Git Simple Read MCP.", // This is on line 3, should not be in output
			},
		},
		{
			name: "start_line out of bounds",
			args: GetFileContentParams{
				Repository: "test-repo",
				FilePath:   testFile,
				StartLine:  11,
			},
			expectError: false,
			expectedInOutput: []string{
				fmt.Sprintf("Content of %s (lines 11-10 of 10)", testFile),
				"No content to display",
				"Note: Start line (11) is beyond the end of the file (10 lines).",
			},
		},
		{
			name: "non-existent file",
			args: GetFileContentParams{
				Repository: "test-repo",
				FilePath:   "nonexistent.txt",
			},
			expectError: true,
			expectedInOutput: []string{
				"Error for nonexistent.txt",
				"failed to open file",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The test repo helper creates the repo inside a workspace, but the handler needs the repo name.
			// The helper sets up a global workspace, so we can use that.
			tt.args.Repository = "test-repo"

			result, _, err := handleGetFileContent(context.Background(), nil, tt.args)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.IsError != tt.expectError {
				t.Errorf("Expected IsError to be %v, but got %v", tt.expectError, result.IsError)
			}

			content := result.Content[0].(*mcp.TextContent).Text
			for _, expected := range tt.expectedInOutput {
				if !strings.Contains(content, expected) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", expected, content)
				}
			}
			for _, notExpected := range tt.notExpectedInOutput {
				if strings.Contains(content, notExpected) {
					t.Errorf("Expected output to not contain %q, but it did.\nOutput:\n%s", notExpected, content)
				}
			}
		})
	}
}

func TestHandleCloneRepository(t *testing.T) {
	// Create a single source repository to be used as the remote for clones
	sourceRepo := CreateTestRepositoryWithContent(t)
	// The helper pollutes the global workspace, so we clear it before starting the actual tests.
	globalWorkspaceManager = nil

	t.Run("clone existing repository should pull", func(t *testing.T) {
		// Setup a clean workspace for the test
		workspaceDir := t.TempDir()
		InitializeWorkspace(workspaceDir)
		defer func() { globalWorkspaceManager = nil }()

		// 1. Initial clone of the source repo into our clean workspace
		cloneArgs := CloneRepositoryParams{URL: sourceRepo.Path, Name: "test-repo"}
		result, _, err := handleCloneRepository(context.Background(), nil, cloneArgs)
		if err != nil {
			t.Fatalf("Initial clone in test setup failed with error: %v", err)
		}
		if result.IsError {
			t.Fatalf("Initial clone in test setup failed with tool error: %s", result.Content[0].(*mcp.TextContent).Text)
		}

		// 2. Call clone again, which should trigger a pull
		result, _, err = handleCloneRepository(context.Background(), nil, cloneArgs)
		if err != nil {
			t.Fatalf("handleCloneRepository on existing repo failed with error: %v", err)
		}
		if result.IsError {
			t.Fatalf("handleCloneRepository on existing repo returned tool error: %s", result.Content[0].(*mcp.TextContent).Text)
		}

		// 3. Check that the output indicates a pull was performed.
		content := result.Content[0].(*mcp.TextContent).Text
		if !strings.Contains(content, "Pulled latest changes") {
			t.Errorf("Expected output to contain 'Pulled latest changes', but got: %s", content)
		}
	})
}
