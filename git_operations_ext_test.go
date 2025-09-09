package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var testMutex = &sync.Mutex{}

// setupTestRepo creates a temporary git repository for testing in a safe way
func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()
	testMutex.Lock()

	// Store original workspace manager
	originalWorkspaceManager := GetWorkspaceManager()

	// Create a temporary directory for the workspace
	workspacePath, err := os.MkdirTemp("", "test-workspace-ext-")
	if err != nil {
		t.Fatalf("Failed to create temp dir for workspace: %v", err)
	}

	// Initialize the temporary workspace
	if err := InitializeWorkspace(workspacePath); err != nil {
		t.Fatalf("Failed to initialize workspace: %v", err)
	}

	// Create a repository inside the workspace
	repoName := "test-repo"
	repoPath := filepath.Join(workspacePath, repoName)
	if err := os.Mkdir(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create repo dir: %v", err)
	}

	// Git init
	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Git init failed: %v", err)
	}

	// Configure git user
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Git config user.name failed: %v", err)
	}
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Git config user.email failed: %v", err)
	}

	// Create first commit
	filePath := filepath.Join(repoPath, "test.txt")
	if err := os.WriteFile(filePath, []byte("initial content"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Git add failed: %v", err)
	}
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Git commit failed: %v", err)
	}

	// Create second commit
	if err := os.WriteFile(filePath, []byte("updated content"), 0644); err != nil {
		t.Fatalf("Failed to write file again: %v", err)
	}
	cmd = exec.Command("git", "commit", "-am", "Second commit")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Second git commit failed: %v", err)
	}

	// Return the repo name and a cleanup function
	return repoName, func() {
		globalWorkspaceManager = originalWorkspaceManager
		os.RemoveAll(workspacePath)
		testMutex.Unlock()
	}
}

func TestListCommits(t *testing.T) {
	repoName, cleanup := setupTestRepo(t)
	defer cleanup()

	commits, err := ListCommits(repoName, 10)
	if err != nil {
		t.Fatalf("ListCommits failed: %v", err)
	}

	if len(commits) != 2 {
		t.Errorf("Expected 2 commits, but got %d", len(commits))
	}

	if commits[0].Message != "Second commit" {
		t.Errorf("Expected latest commit message to be 'Second commit', but got '%s'", commits[0].Message)
	}

	// Test limit
	commits, err = ListCommits(repoName, 1)
	if err != nil {
		t.Fatalf("ListCommits with limit failed: %v", err)
	}

	if len(commits) != 1 {
		t.Errorf("Expected 1 commit with limit, but got %d", len(commits))
	}
}

func TestGetCommitDiff(t *testing.T) {
	repoName, cleanup := setupTestRepo(t)
	defer cleanup()

	commits, err := ListCommits(repoName, 1)
	if err != nil || len(commits) == 0 {
		t.Fatalf("Could not get latest commit to test diff")
	}
	latestCommitHash := commits[0].Hash

	diff, err := GetCommitDiff(repoName, latestCommitHash)
	if err != nil {
		t.Fatalf("GetCommitDiff failed: %v", err)
	}

	if !strings.Contains(diff, "commit "+latestCommitHash) {
		t.Errorf("Diff output should contain the commit hash")
	}

	if !strings.Contains(diff, "Second commit") {
		t.Errorf("Diff output should contain the commit message")
	}

	if !strings.Contains(diff, "diff --git a/test.txt b/test.txt") {
		t.Errorf("Diff output should contain the diff header for test.txt")
	}

	if !strings.Contains(diff, "+updated content") {
		t.Errorf("Diff output should show the added line")
	}
}
