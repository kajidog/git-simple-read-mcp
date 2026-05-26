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
		return nil, fmt.Errorf("invalid workspace path: %w", err)
	}

	// Create workspace directory if it doesn't exist
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workspace directory: %w", err)
	}

	// Resolve symlinks so all later comparisons happen against the real path.
	// This is required to prevent symlink-based path traversal: without this,
	// a symlink inside the workspace pointing outside would satisfy the prefix
	// check but resolve to an external location at the OS level.
	resolved, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve workspace symlinks: %w", err)
	}

	return &WorkspaceManager{
		workspaceDir: resolved,
	}, nil
}

// GetWorkspaceDir returns the absolute path to the workspace directory
func (wm *WorkspaceManager) GetWorkspaceDir() string {
	return wm.workspaceDir
}

// ValidateRepositoryPath validates that the given path is within the workspace
// and converts relative paths to absolute paths within the workspace.
//
// Symlinks are resolved before the containment check so that a symlink inside
// the workspace pointing outside (e.g. workspace/evil -> /etc) cannot be used
// to escape the workspace. For paths that do not yet exist (e.g. a clone
// target), the closest existing ancestor is resolved and the remaining
// components are appended.
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
		return "", fmt.Errorf("invalid repository path: %w", err)
	}

	// Resolve symlinks. The path may not exist yet (clone target etc.), so we
	// walk up to the closest existing ancestor, resolve that, and re-attach
	// the missing tail.
	resolved, err := resolveSymlinksAllowMissing(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path symlinks: %w", err)
	}

	// Check if the resolved path is within workspace
	if !wm.isWithinWorkspace(resolved) {
		return "", fmt.Errorf("repository path must be within workspace directory: %s", wm.workspaceDir)
	}

	return resolved, nil
}

// resolveSymlinksAllowMissing resolves symlinks on path. If the leaf or
// intermediate components do not exist, the closest existing ancestor is
// resolved and the remaining components are appended.
func resolveSymlinksAllowMissing(path string) (string, error) {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	// Walk up to find the closest existing ancestor.
	dir := path
	tail := ""
	for {
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the root without finding an existing ancestor.
			return path, nil
		}
		tail = filepath.Join(filepath.Base(dir), tail)
		dir = parent
		resolved, err := filepath.EvalSymlinks(dir)
		if err == nil {
			return filepath.Join(resolved, tail), nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
	}
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
		return "", fmt.Errorf("failed to get relative path: %w", err)
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
		return nil, fmt.Errorf("failed to read workspace directory: %w", err)
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
		return fmt.Errorf("failed to remove repository: %w", err)
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
