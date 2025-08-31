package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestWorkspaceManager(t *testing.T) {
	t.Run("create workspace manager", func(t *testing.T) {
		tempDir := t.TempDir()
		
		wm, err := NewWorkspaceManager(tempDir)
		if err != nil {
			t.Fatalf("Failed to create workspace manager: %v", err)
		}
		
		if wm.GetWorkspaceDir() != tempDir {
			t.Errorf("Expected workspace dir %s, got %s", tempDir, wm.GetWorkspaceDir())
		}
	})

	t.Run("validate repository path within workspace", func(t *testing.T) {
		tempDir := t.TempDir()
		wm, _ := NewWorkspaceManager(tempDir)
		
		// Test relative path
		validPath, err := wm.ValidateRepositoryPath("myrepo")
		if err != nil {
			t.Fatalf("Failed to validate relative path: %v", err)
		}
		
		expectedPath := filepath.Join(tempDir, "myrepo")
		if validPath != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, validPath)
		}
		
		// Test absolute path within workspace
		absolutePath := filepath.Join(tempDir, "another-repo")
		validPath, err = wm.ValidateRepositoryPath(absolutePath)
		if err != nil {
			t.Fatalf("Failed to validate absolute path: %v", err)
		}
		
		if validPath != absolutePath {
			t.Errorf("Expected path %s, got %s", absolutePath, validPath)
		}
	})

	t.Run("reject path outside workspace", func(t *testing.T) {
		tempDir := t.TempDir()
		wm, _ := NewWorkspaceManager(tempDir)
		
		outsidePath := "/tmp/outside-repo"
		_, err := wm.ValidateRepositoryPath(outsidePath)
		if err == nil {
			t.Errorf("Expected error for path outside workspace")
		}
	})

	t.Run("list repositories in workspace", func(t *testing.T) {
		tempDir := t.TempDir()
		wm, _ := NewWorkspaceManager(tempDir)
		
		// Create some mock repositories
		repo1Path := filepath.Join(tempDir, "repo1")
		repo2Path := filepath.Join(tempDir, "repo2")
		nonRepoPath := filepath.Join(tempDir, "not-a-repo")
		
		// Create directories
		os.MkdirAll(repo1Path, 0755)
		os.MkdirAll(repo2Path, 0755)
		os.MkdirAll(nonRepoPath, 0755)
		
		// Initialize git repositories
		os.MkdirAll(filepath.Join(repo1Path, ".git"), 0755)
		os.MkdirAll(filepath.Join(repo2Path, ".git"), 0755)
		
		repositories, err := wm.ListRepositories()
		if err != nil {
			t.Fatalf("Failed to list repositories: %v", err)
		}
		
		if len(repositories) != 2 {
			t.Errorf("Expected 2 repositories, got %d", len(repositories))
		}
		
		found := make(map[string]bool)
		for _, repo := range repositories {
			found[repo] = true
		}
		
		if !found["repo1"] || !found["repo2"] {
			t.Errorf("Expected to find repo1 and repo2, got %v", repositories)
		}
		
		if found["not-a-repo"] {
			t.Errorf("Should not include non-git directory")
		}
	})

	t.Run("repository exists check", func(t *testing.T) {
		tempDir := t.TempDir()
		wm, _ := NewWorkspaceManager(tempDir)
		
		// Create a mock repository
		repoPath := filepath.Join(tempDir, "testrepo")
		os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)
		
		if !wm.RepositoryExists("testrepo") {
			t.Errorf("Repository should exist")
		}
		
		if wm.RepositoryExists("nonexistent") {
			t.Errorf("Repository should not exist")
		}
	})

	t.Run("remove repository", func(t *testing.T) {
		tempDir := t.TempDir()
		wm, _ := NewWorkspaceManager(tempDir)
		
		// Create a mock repository
		repoPath := filepath.Join(tempDir, "to-remove")
		os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)
		
		// Create a test file
		testFile := filepath.Join(repoPath, "test.txt")
		os.WriteFile(testFile, []byte("test"), 0644)
		
		// Verify it exists
		if !wm.RepositoryExists("to-remove") {
			t.Fatalf("Repository should exist before removal")
		}
		
		// Remove it
		err := wm.RemoveRepository("to-remove")
		if err != nil {
			t.Fatalf("Failed to remove repository: %v", err)
		}
		
		// Verify it's gone
		if wm.RepositoryExists("to-remove") {
			t.Errorf("Repository should not exist after removal")
		}
		
		if _, err := os.Stat(repoPath); !os.IsNotExist(err) {
			t.Errorf("Repository directory should be completely removed")
		}
	})
}

func TestCloneRepository(t *testing.T) {
	// Setup workspace
	tempDir := t.TempDir()
	InitializeWorkspace(tempDir)
	defer func() { globalWorkspaceManager = nil }() // Cleanup
	
	t.Run("clone repository validation", func(t *testing.T) {
		// Test empty URL
		_, _, err := CloneRepository("", "testrepo")
		if err == nil {
			t.Errorf("Expected error for empty URL")
		}
		
		// Test auto-extraction with invalid URL
		_, repoName, err := CloneRepository("invalid-url", "")
		if err == nil {
			t.Errorf("Expected error for invalid URL")
		}
		// Even on error, name extraction should work for single segments
		if repoName != "" && repoName != "invalid-url" {
			t.Errorf("Expected extracted name 'invalid-url', got %q", repoName)
		}
	})
	
	t.Run("clone existing repository", func(t *testing.T) {
		// Create a mock existing repository
		repoPath := filepath.Join(tempDir, "existing")
		os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)
		
		// Attempt to clone to existing name
		_, _, err := CloneRepository("https://github.com/user/repo.git", "existing")
		if err == nil {
			t.Errorf("Expected error when cloning to existing repository name")
		}
		
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("Expected error message about existing repository, got: %v", err)
		}
	})
}

func TestWorkspacePathValidation(t *testing.T) {
	// Setup workspace
	tempDir := t.TempDir()
	InitializeWorkspace(tempDir)
	defer func() { globalWorkspaceManager = nil }() // Cleanup
	
	t.Run("validate workspace paths", func(t *testing.T) {
		// Test relative path
		validPath, err := ValidateWorkspacePath("myrepo")
		if err != nil {
			t.Fatalf("Failed to validate relative path: %v", err)
		}
		
		expectedPath := filepath.Join(tempDir, "myrepo")
		if validPath != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, validPath)
		}
		
		// Test path outside workspace
		_, err = ValidateWorkspacePath("/tmp/outside")
		if err == nil {
			t.Errorf("Expected error for path outside workspace")
		}
	})

	t.Run("uninitialized workspace", func(t *testing.T) {
		globalWorkspaceManager = nil // Simulate uninitialized workspace
		
		_, err := ValidateWorkspacePath("test")
		if err == nil {
			t.Errorf("Expected error for uninitialized workspace")
		}
		
		if !strings.Contains(err.Error(), "not initialized") {
			t.Errorf("Expected error message about uninitialized workspace")
		}
		
		// Restore workspace
		InitializeWorkspace(tempDir)
	})
}

func TestWorkspaceMCPHandlers(t *testing.T) {
	// Setup workspace
	tempDir := t.TempDir()
	InitializeWorkspace(tempDir)
	defer func() { globalWorkspaceManager = nil }() // Cleanup
	
	ctx := context.Background()
	
	t.Run("handle clone repository", func(t *testing.T) {
		// Test empty URL
		params := CloneRepositoryParams{URL: "", Name: "test"}
		
		result, _, err := handleCloneRepository(ctx, nil, params)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}
		
		if !result.IsError {
			t.Errorf("Expected error result for empty URL")
		}
		
		// Test empty name
		params = CloneRepositoryParams{URL: "https://github.com/user/repo.git", Name: ""}
		result, _, err = handleCloneRepository(ctx, nil, params)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}
		
		if !result.IsError {
			t.Errorf("Expected error result for empty name")
		}
	})

	t.Run("handle list workspace repositories", func(t *testing.T) {
		// Create some mock repositories
		repo1Path := filepath.Join(tempDir, "repo1")
		repo2Path := filepath.Join(tempDir, "repo2")
		
		os.MkdirAll(filepath.Join(repo1Path, ".git"), 0755)
		os.MkdirAll(filepath.Join(repo2Path, ".git"), 0755)
		
		params := ListWorkspaceRepositoriesParams{}
		
		result, _, err := handleListWorkspaceRepositories(ctx, nil, params)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}
		
		if result.IsError {
			t.Errorf("Expected success result")
		}
		
		content := result.Content[0].(*mcp.TextContent).Text
		if !strings.Contains(content, "Workspace Repositories") {
			t.Errorf("Expected workspace repositories header in content")
		}
		
		if !strings.Contains(content, "repo1") || !strings.Contains(content, "repo2") {
			t.Errorf("Expected repository names in content")
		}
	})

	t.Run("handle list empty workspace", func(t *testing.T) {
		// Clean workspace
		emptyDir := t.TempDir()
		InitializeWorkspace(emptyDir)
		
		params := ListWorkspaceRepositoriesParams{}
		
		result, _, err := handleListWorkspaceRepositories(ctx, nil, params)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}
		
		if result.IsError {
			t.Errorf("Expected success result")
		}
		
		content := result.Content[0].(*mcp.TextContent).Text
		if !strings.Contains(content, "No repositories found") {
			t.Errorf("Expected empty workspace message")
		}
		
		if !strings.Contains(content, "clone_repository") {
			t.Errorf("Expected clone instruction in empty workspace")
		}
		
		// Restore original workspace
		InitializeWorkspace(tempDir)
	})

	t.Run("handle remove repository", func(t *testing.T) {
		// Test empty name
		params := RemoveRepositoryParams{Name: ""}
		
		result, _, err := handleRemoveRepository(ctx, nil, params)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}
		
		if !result.IsError {
			t.Errorf("Expected error result for empty name")
		}
		
		// Test non-existent repository
		params = RemoveRepositoryParams{Name: "nonexistent"}
		result, _, err = handleRemoveRepository(ctx, nil, params)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}
		
		if !result.IsError {
			t.Errorf("Expected error result for non-existent repository")
		}
		
		// Test successful removal
		// Create a mock repository
		repoPath := filepath.Join(tempDir, "to-remove")
		os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)
		
		params = RemoveRepositoryParams{Name: "to-remove"}
		result, _, err = handleRemoveRepository(ctx, nil, params)
		if err != nil {
			t.Fatalf("Handler returned error: %v", err)
		}
		
		if result.IsError {
			t.Errorf("Expected success result for valid removal")
		}
		
		content := result.Content[0].(*mcp.TextContent).Text
		if !strings.Contains(content, "Successfully removed") {
			t.Errorf("Expected success message in content")
		}
	})
}

func TestWorkspacePathValidationInOperations(t *testing.T) {
	// Setup workspace
	tempDir := t.TempDir()
	InitializeWorkspace(tempDir)
	defer func() { globalWorkspaceManager = nil }() // Cleanup
	
	// Create a test repository in workspace
	repoName := "test-validation"
	repoPath := filepath.Join(tempDir, repoName)
	_ = CreateTestRepositoryAt(t, repoPath)
	
	t.Run("repository operations require workspace validation", func(t *testing.T) {
		// Test GetRepositoryInfo with relative path
		info, err := GetRepositoryInfo(repoName)
		if err != nil {
			t.Fatalf("Failed to get repository info with relative path: %v", err)
		}
		
		if info == nil {
			t.Errorf("Expected repository info")
		}
		
		// Test with path outside workspace
		outsideRepo := "/tmp/outside-repo"
		_, err = GetRepositoryInfo(outsideRepo)
		if err == nil {
			t.Errorf("Expected error for path outside workspace")
		}
		
		if !strings.Contains(err.Error(), "workspace") {
			t.Errorf("Expected workspace validation error, got: %v", err)
		}
	})
	
	t.Run("all operations validate workspace paths", func(t *testing.T) {
		// Test various operations with invalid paths
		outsidePath := "/tmp/outside"
		
		operations := []func(string) error{
			func(path string) error {
				_, err := GetRepositoryInfo(path)
				return err
			},
			func(path string) error {
				_, err := ListBranches(path)
				return err
			},
			func(path string) error {
				_, err := SwitchBranch(path, "main")
				return err
			},
			func(path string) error {
				_, err := SearchFiles(path, []string{"test"}, "and", false, 0, 10)
				return err
			},
			func(path string) error {
				_, err := ListFiles(path, ".", false, 10)
				return err
			},
			func(path string) error {
				_, err := GetFileContent(path, "test.txt", 10)
				return err
			},
			func(path string) error {
				_, err := PullRepository(path)
				return err
			},
		}
		
		for i, operation := range operations {
			err := operation(outsidePath)
			if err == nil {
				t.Errorf("Operation %d should have failed for path outside workspace", i)
			}
			
			if !strings.Contains(err.Error(), "workspace") {
				t.Errorf("Operation %d should have returned workspace validation error, got: %v", i, err)
			}
		}
	})
}

// CreateTestRepositoryAt creates a test repository at a specific path
func CreateTestRepositoryAt(t *testing.T, path string) *TestRepository {
	os.MkdirAll(path, 0755)
	
	repo := &TestRepository{
		Path: path,
		T:    t,
	}
	
	// Initialize git repository
	repo.runGitCommand("init")
	repo.runGitCommand("config", "user.name", "Test User")
	repo.runGitCommand("config", "user.email", "test@example.com")
	repo.runGitCommand("config", "init.defaultBranch", "main")
	
	// Add initial content
	repo.WriteFile("README.md", "# Test Repository")
	repo.runGitCommand("add", ".")
	repo.runGitCommand("commit", "-m", "Initial commit")
	
	return repo
}