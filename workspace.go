package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WorkspaceManager manages the workspace directory for Git operations
type WorkspaceManager struct {
	workspaceDir string
}

// NewWorkspaceManager creates a new workspace manager
func NewWorkspaceManager(workspaceDir string) (*WorkspaceManager, error) {
	if workspaceDir == "" {
		return nil, fmt.Errorf("workspace directory cannot be empty")
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("invalid workspace path: %v", err)
	}

	// Create workspace directory if it doesn't exist
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workspace directory: %v", err)
	}

	return &WorkspaceManager{
		workspaceDir: absPath,
	}, nil
}

// GetWorkspaceDir returns the absolute path to the workspace directory
func (wm *WorkspaceManager) GetWorkspaceDir() string {
	return wm.workspaceDir
}

// ValidateRepositoryPath validates that the given path is within the workspace
// and converts relative paths to absolute paths within the workspace
func (wm *WorkspaceManager) ValidateRepositoryPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("repository path cannot be empty")
	}

	var fullPath string

	// If path is relative, resolve it within workspace
	if !filepath.IsAbs(path) {
		fullPath = filepath.Join(wm.workspaceDir, path)
	} else {
		fullPath = path
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("invalid repository path: %v", err)
	}

	// Check if the path is within workspace
	if !wm.isWithinWorkspace(absPath) {
		return "", fmt.Errorf("repository path must be within workspace directory: %s", wm.workspaceDir)
	}

	return absPath, nil
}

// GetRepositoryName extracts the repository name from a repository path within workspace
func (wm *WorkspaceManager) GetRepositoryName(repoPath string) (string, error) {
	validPath, err := wm.ValidateRepositoryPath(repoPath)
	if err != nil {
		return "", err
	}

	// Get relative path from workspace
	relPath, err := filepath.Rel(wm.workspaceDir, validPath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %v", err)
	}

	// Return the first directory component as repository name
	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) == 0 || parts[0] == "." {
		return "", fmt.Errorf("invalid repository path")
	}

	return parts[0], nil
}

// GetRepositoryPath returns the full path to a repository by name
func (wm *WorkspaceManager) GetRepositoryPath(repoName string) string {
	return filepath.Join(wm.workspaceDir, repoName)
}

// ListRepositories lists all repositories in the workspace
func (wm *WorkspaceManager) ListRepositories() ([]string, error) {
	entries, err := os.ReadDir(wm.workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read workspace directory: %v", err)
	}

	var repositories []string
	for _, entry := range entries {
		if entry.IsDir() {
			repoPath := filepath.Join(wm.workspaceDir, entry.Name())
			if isGitRepository(repoPath) {
				repositories = append(repositories, entry.Name())
			}
		}
	}

	return repositories, nil
}

// RepositoryExists checks if a repository exists in the workspace
func (wm *WorkspaceManager) RepositoryExists(repoName string) bool {
	repoPath := wm.GetRepositoryPath(repoName)
	return isGitRepository(repoPath)
}

// RemoveRepository removes a repository from the workspace
func (wm *WorkspaceManager) RemoveRepository(repoName string) error {
	if !wm.RepositoryExists(repoName) {
		return fmt.Errorf("repository '%s' does not exist", repoName)
	}

	repoPath := wm.GetRepositoryPath(repoName)
	if err := os.RemoveAll(repoPath); err != nil {
		return fmt.Errorf("failed to remove repository: %v", err)
	}

	return nil
}

// isWithinWorkspace checks if the given path is within the workspace directory
func (wm *WorkspaceManager) isWithinWorkspace(path string) bool {
	// Convert both paths to absolute paths for comparison
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	absWorkspace, err := filepath.Abs(wm.workspaceDir)
	if err != nil {
		return false
	}

	// Check if path starts with workspace directory
	rel, err := filepath.Rel(absWorkspace, absPath)
	if err != nil {
		return false
	}

	// Path is within workspace if relative path doesn't start with ".."
	return !strings.HasPrefix(rel, "..") && rel != ".."
}

// Global workspace manager instance
var globalWorkspaceManager *WorkspaceManager

// InitializeWorkspace initializes the global workspace manager
func InitializeWorkspace(workspaceDir string) error {
	var err error
	globalWorkspaceManager, err = NewWorkspaceManager(workspaceDir)
	return err
}

// GetWorkspaceManager returns the global workspace manager
func GetWorkspaceManager() *WorkspaceManager {
	return globalWorkspaceManager
}

// ValidateWorkspacePath validates a path using the global workspace manager
func ValidateWorkspacePath(path string) (string, error) {
	if globalWorkspaceManager == nil {
		return "", fmt.Errorf("workspace not initialized")
	}
	return globalWorkspaceManager.ValidateRepositoryPath(path)
}
