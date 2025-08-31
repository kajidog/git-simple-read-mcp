package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRepository represents a test Git repository
type TestRepository struct {
	Path     string
	T        *testing.T
	branches []string
}

// CreateTestRepository creates a temporary Git repository for testing
func CreateTestRepository(t *testing.T) *TestRepository {
	tempDir := t.TempDir()
	
	repo := &TestRepository{
		Path: tempDir,
		T:    t,
	}
	
	// Initialize git repository
	repo.runGitCommand("init")
	repo.runGitCommand("config", "user.name", "Test User")
	repo.runGitCommand("config", "user.email", "test@example.com")
	repo.runGitCommand("config", "init.defaultBranch", "main")
	
	return repo
}

// CreateTestRepositoryWithContent creates a test repository with sample content
func CreateTestRepositoryWithContent(t *testing.T) *TestRepository {
	repo := CreateTestRepository(t)
	
	// Create initial files
	repo.WriteFile("README.md", "# Test Repository\n\nThis is a test repository for Git Remote MCP.")
	repo.WriteFile("LICENSE", "MIT License\n\nCopyright (c) 2024 Test")
	repo.WriteFile("main.go", `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`)
	repo.WriteFile("src/utils.go", `package src

func Add(a, b int) int {
	return a + b
}

func Multiply(a, b int) int {
	return a * b
}
`)
	repo.WriteFile("docs/api.md", "# API Documentation\n\nThis is the API documentation.")
	
	// Create initial commit
	repo.runGitCommand("add", ".")
	repo.runGitCommand("commit", "-m", "Initial commit")
	
	// Ensure we're on main branch (rename master to main if needed)
	currentBranch := repo.getCurrentBranch()
	if currentBranch == "master" {
		repo.runGitCommand("branch", "-m", "master", "main")
	}
	
	// Create additional branches
	repo.CreateBranch("feature/test")
	repo.CreateBranch("develop")
	
	// Switch back to main
	repo.runGitCommand("checkout", "main")
	
	// Add more commits
	repo.WriteFile("version.txt", "1.0.0")
	repo.runGitCommand("add", "version.txt")
	repo.runGitCommand("commit", "-m", "Add version file")
	
	// Add file with search keywords
	repo.WriteFile("config.json", `{
	"database": "postgres",
	"redis": true,
	"logging": {
		"level": "info",
		"file": "app.log"
	}
}`)
	repo.runGitCommand("add", "config.json")
	repo.runGitCommand("commit", "-m", "Add configuration")
	
	return repo
}

// WriteFile writes content to a file in the test repository
func (tr *TestRepository) WriteFile(filename, content string) {
	fullPath := filepath.Join(tr.Path, filename)
	dir := filepath.Dir(fullPath)
	
	if err := os.MkdirAll(dir, 0755); err != nil {
		tr.T.Fatalf("Failed to create directory %s: %v", dir, err)
	}
	
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		tr.T.Fatalf("Failed to write file %s: %v", fullPath, err)
	}
}

// CreateBranch creates a new branch in the test repository
func (tr *TestRepository) CreateBranch(branchName string) {
	tr.runGitCommand("checkout", "-b", branchName)
	tr.branches = append(tr.branches, branchName)
}

// SwitchBranch switches to the specified branch
func (tr *TestRepository) SwitchBranch(branchName string) {
	tr.runGitCommand("checkout", branchName)
}

// AddCommit adds a commit with the specified message
func (tr *TestRepository) AddCommit(message string) {
	tr.runGitCommand("add", ".")
	tr.runGitCommand("commit", "-m", message)
}

// CreateLargeFile creates a large file for testing file size limits
func (tr *TestRepository) CreateLargeFile(filename string, sizeKB int) {
	content := make([]byte, sizeKB*1024)
	for i := range content {
		content[i] = byte('A' + (i % 26))
	}
	tr.WriteFile(filename, string(content))
}

// CreateManyFiles creates many files for testing pagination
func (tr *TestRepository) CreateManyFiles(dir string, count int) {
	for i := 0; i < count; i++ {
		filename := filepath.Join(dir, fmt.Sprintf("file_%03d.txt", i))
		content := fmt.Sprintf("Content of file %d\nLine 2 of file %d\n", i, i)
		tr.WriteFile(filename, content)
	}
}

// GetBranches returns the list of branches created in this test repository
func (tr *TestRepository) GetBranches() []string {
	return append([]string{"main"}, tr.branches...)
}

// runGitCommand runs a git command in the test repository
func (tr *TestRepository) runGitCommand(args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = tr.Path
	if err := cmd.Run(); err != nil {
		tr.T.Fatalf("Git command failed: git %v, error: %v", args, err)
	}
}

// getCurrentBranch returns the current branch name
func (tr *TestRepository) getCurrentBranch() string {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = tr.Path
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// AssertFileExists asserts that a file exists in the repository
func (tr *TestRepository) AssertFileExists(filename string) {
	fullPath := filepath.Join(tr.Path, filename)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		tr.T.Fatalf("Expected file %s to exist, but it doesn't", filename)
	}
}

// AssertFileContent asserts that a file has the expected content
func (tr *TestRepository) AssertFileContent(filename, expectedContent string) {
	fullPath := filepath.Join(tr.Path, filename)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		tr.T.Fatalf("Failed to read file %s: %v", filename, err)
	}
	if string(content) != expectedContent {
		tr.T.Fatalf("File %s content mismatch.\nExpected: %s\nActual: %s", filename, expectedContent, string(content))
	}
}

// WaitForFileSystem waits a bit for filesystem operations to complete
func (tr *TestRepository) WaitForFileSystem() {
	time.Sleep(10 * time.Millisecond)
}