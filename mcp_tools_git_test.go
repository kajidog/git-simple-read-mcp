package main

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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
