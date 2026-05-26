package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSymlinkEscapeRejection verifies that a symlink inside the workspace
// cannot be used to access paths outside of the workspace.
func TestSymlinkEscapeRejection(t *testing.T) {
	workspaceDir := t.TempDir()
	outsideDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(outsideDir, "secret.txt"), []byte("secret"), 0644); err != nil {
		t.Fatalf("failed to create outside file: %v", err)
	}

	wm, err := NewWorkspaceManager(workspaceDir)
	if err != nil {
		t.Fatalf("NewWorkspaceManager failed: %v", err)
	}

	linkPath := filepath.Join(workspaceDir, "escape")
	if err := os.Symlink(outsideDir, linkPath); err != nil {
		t.Skipf("symlink not supported on this platform: %v", err)
	}

	if _, err := wm.ValidateRepositoryPath("escape"); err == nil {
		t.Fatal("expected ValidateRepositoryPath to reject symlink escape, but it succeeded")
	} else if !errors.Is(err, ErrPathOutsideWorkspace) {
		t.Fatalf("expected ErrPathOutsideWorkspace, got: %v", err)
	}

	if _, err := wm.ValidateRepositoryPath("escape/secret.txt"); err == nil {
		t.Fatal("expected ValidateRepositoryPath to reject symlink-traversed file access, but it succeeded")
	}
}

// TestWorkspaceSymlinkResolved verifies that when the workspace directory
// itself is a symlink to a real directory, paths still validate correctly.
func TestWorkspaceSymlinkResolved(t *testing.T) {
	parent := t.TempDir()
	realDir := filepath.Join(parent, "real")
	if err := os.MkdirAll(realDir, 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	linkDir := filepath.Join(parent, "link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Skipf("symlink not supported on this platform: %v", err)
	}

	wm, err := NewWorkspaceManager(linkDir)
	if err != nil {
		t.Fatalf("NewWorkspaceManager failed: %v", err)
	}

	repoDir := filepath.Join(realDir, "repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("mkdir repo failed: %v", err)
	}

	resolved, err := wm.ValidateRepositoryPath("repo")
	if err != nil {
		t.Fatalf("ValidateRepositoryPath rejected legitimate path: %v", err)
	}
	if !strings.HasPrefix(resolved, realDir) {
		t.Fatalf("expected resolved path under %s, got %s", realDir, resolved)
	}
}

// TestValidateNonExistentClonedTarget verifies that paths which do not yet
// exist (e.g. clone targets) are still validated against workspace bounds.
func TestValidateNonExistentClonedTarget(t *testing.T) {
	workspaceDir := t.TempDir()
	wm, err := NewWorkspaceManager(workspaceDir)
	if err != nil {
		t.Fatalf("NewWorkspaceManager failed: %v", err)
	}

	resolved, err := wm.ValidateRepositoryPath("not-yet-cloned")
	if err != nil {
		t.Fatalf("ValidateRepositoryPath should accept non-existent paths within workspace: %v", err)
	}
	expected := filepath.Join(wm.GetWorkspaceDir(), "not-yet-cloned")
	if resolved != expected {
		t.Fatalf("expected %s, got %s", expected, resolved)
	}

	if _, err := wm.ValidateRepositoryPath("../outside"); err == nil {
		t.Fatal("expected rejection of ../outside")
	}
}
