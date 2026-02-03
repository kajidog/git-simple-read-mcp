package main

import (
	"strings"
	"testing"
)

func TestGetRepositoryInfo(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(t *testing.T) *TestRepository
		expectError bool
	}{
		{
			name: "valid repository with content",
			setupRepo: func(t *testing.T) *TestRepository {
				return CreateTestRepositoryWithContent(t)
			},
			expectError: false,
		},
		{
			name: "non-git directory",
			setupRepo: func(t *testing.T) *TestRepository {
				tempDir := t.TempDir()
				return &TestRepository{Path: tempDir, T: t}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo(t)

			info, err := GetRepositoryInfo(repo.Path)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify repository info
			if info.Path != repo.Path {
				t.Errorf("Expected path %s, got %s", repo.Path, info.Path)
			}

			if info.CurrentBranch == "" {
				t.Errorf("Expected current branch to be set")
			}

			if info.License == "" {
				t.Errorf("Expected license file to be detected")
			}

			if info.ReadmeContent == "" {
				t.Errorf("Expected README content to be read")
			}

			if info.LastUpdate.IsZero() {
				t.Errorf("Expected last update time to be set")
			}
		})
	}
}

func TestListBranches(t *testing.T) {
	tests := []struct {
		name             string
		setupRepo        func(t *testing.T) *TestRepository
		expectError      bool
		expectedBranches []string
	}{
		{
			name: "repository with multiple branches",
			setupRepo: func(t *testing.T) *TestRepository {
				return CreateTestRepositoryWithContent(t)
			},
			expectError:      false,
			expectedBranches: []string{"main", "feature/test", "develop"},
		},
		{
			name: "non-git directory",
			setupRepo: func(t *testing.T) *TestRepository {
				tempDir := t.TempDir()
				return &TestRepository{Path: tempDir, T: t}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo(t)

			branches, err := ListBranches(repo.Path)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check that we have the expected branches
			branchNames := make([]string, len(branches))
			currentBranch := ""
			for i, branch := range branches {
				branchNames[i] = branch.Name
				if branch.IsCurrent {
					currentBranch = branch.Name
				}
			}

			// Verify current branch is set
			if currentBranch == "" {
				t.Errorf("Expected one branch to be marked as current")
			}

			// Verify expected branches exist
			for _, expectedBranch := range tt.expectedBranches {
				found := false
				for _, actualBranch := range branchNames {
					if actualBranch == expectedBranch {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected branch %s not found in %v", expectedBranch, branchNames)
				}
			}
		})
	}
}

func TestSwitchBranch(t *testing.T) {
	tests := []struct {
		name         string
		targetBranch string
		expectError  bool
	}{
		{
			name:         "switch to existing branch",
			targetBranch: "develop",
			expectError:  false,
		},
		{
			name:         "switch to non-existing branch",
			targetBranch: "non-existent",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := CreateTestRepositoryWithContent(t)

			output, err := SwitchBranch(repo.Path, tt.targetBranch)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if output == "" {
				t.Errorf("Expected output from branch switch command")
			}

			// Verify the branch was actually switched
			branches, err := ListBranches(repo.Path)
			if err != nil {
				t.Fatalf("Failed to list branches after switch: %v", err)
			}

			currentBranch := ""
			for _, branch := range branches {
				if branch.IsCurrent {
					currentBranch = branch.Name
					break
				}
			}

			if currentBranch != tt.targetBranch {
				t.Errorf("Expected current branch %s, got %s", tt.targetBranch, currentBranch)
			}
		})
	}
}

func TestSearchFiles(t *testing.T) {
	tests := []struct {
		name          string
		keywords      []string
		maxResults    int
		expectError   bool
		expectedFiles []string
	}{
		{
			name:          "search for single keyword",
			keywords:      []string{"database"},
			maxResults:    10,
			expectError:   false,
			expectedFiles: []string{"config.json"},
		},
		{
			name:          "search for multiple keywords (AND logic)",
			keywords:      []string{"database", "postgres"},
			maxResults:    10,
			expectError:   false,
			expectedFiles: []string{"config.json"},
		},
		{
			name:          "search with no matches",
			keywords:      []string{"nonexistent"},
			maxResults:    10,
			expectError:   false,
			expectedFiles: []string{},
		},
		{
			name:          "empty keywords",
			keywords:      []string{},
			maxResults:    10,
			expectError:   false,
			expectedFiles: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := CreateTestRepositoryWithContent(t)

			results, err := SearchFiles(repo.Path, tt.keywords, "and", false, 0, nil, nil, tt.maxResults)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify result count
			if len(results) != len(tt.expectedFiles) {
				t.Errorf("Expected %d results, got %d", len(tt.expectedFiles), len(results))
			}

			// Verify expected files are found
			for _, expectedFile := range tt.expectedFiles {
				found := false
				for _, result := range results {
					if result.Path == expectedFile {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected file %s not found in search results", expectedFile)
				}
			}
		})
	}
}

func TestListFiles(t *testing.T) {
	tests := []struct {
		name        string
		dirPath     string
		recursive   bool
		maxResults  int
		expectError bool
		minFiles    int
	}{
		{
			name:        "list root directory non-recursive",
			dirPath:     ".",
			recursive:   false,
			maxResults:  50,
			expectError: false,
			minFiles:    5, // README.md, LICENSE, main.go, etc.
		},
		{
			name:        "list src directory",
			dirPath:     "src",
			recursive:   false,
			maxResults:  50,
			expectError: false,
			minFiles:    1, // utils.go
		},
		{
			name:        "list root directory recursive",
			dirPath:     ".",
			recursive:   true,
			maxResults:  50,
			expectError: false,
			minFiles:    6, // all files including subdirectories
		},
		{
			name:        "list non-existent directory",
			dirPath:     "nonexistent",
			recursive:   false,
			maxResults:  50,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := CreateTestRepositoryWithContent(t)

			files, err := ListFiles(repo.Path, tt.dirPath, tt.recursive, nil, nil, tt.maxResults)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(files) < tt.minFiles {
				t.Errorf("Expected at least %d files, got %d", tt.minFiles, len(files))
			}

			// Verify file info is populated
			for _, file := range files {
				if file.Name == "" {
					t.Errorf("File name should not be empty")
				}
				if file.Path == "" {
					t.Errorf("File path should not be empty")
				}
				// ModTime should be set for existing files
				if file.ModTime.IsZero() {
					t.Errorf("File mod time should be set")
				}
			}
		})
	}
}

func TestGetFileContent(t *testing.T) {
	tests := []struct {
		name            string
		filePath        string
		maxLines        int
		expectError     bool
		expectedContent string
	}{
		{
			name:            "read existing file",
			filePath:        "README.md",
			maxLines:        0,
			expectError:     false,
			expectedContent: "# Test Repository\n\nThis is a test repository for Git Simple Read MCP.\n",
		},
		{
			name:            "read file with line limit",
			filePath:        "README.md",
			maxLines:        1,
			expectError:     false,
			expectedContent: "# Test Repository\n",
		},
		{
			name:        "read non-existent file",
			filePath:    "nonexistent.txt",
			maxLines:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := CreateTestRepositoryWithContent(t)

			content, err := GetFileContent(repo.Path, tt.filePath, tt.maxLines)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if content != tt.expectedContent {
				t.Errorf("Content mismatch.\nExpected: %q\nActual: %q", tt.expectedContent, content)
			}
		})
	}
}

func TestPullRepository(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(t *testing.T) *TestRepository
		expectError bool
	}{
		{
			name: "pull in repository without remote",
			setupRepo: func(t *testing.T) *TestRepository {
				return CreateTestRepositoryWithContent(t)
			},
			expectError: true, // No remote configured
		},
		{
			name: "pull in non-git directory",
			setupRepo: func(t *testing.T) *TestRepository {
				tempDir := t.TempDir()
				return &TestRepository{Path: tempDir, T: t}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo(t)

			output, err := PullRepository(repo.Path)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				// Output should still contain error information
				if output == "" {
					t.Errorf("Expected output even on error")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if output == "" {
				t.Errorf("Expected output from pull command")
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("isGitRepository", func(t *testing.T) {
		// Test with git repository
		repo := CreateTestRepositoryWithContent(t)
		if !isGitRepository(repo.Path) {
			t.Errorf("Expected %s to be recognized as git repository", repo.Path)
		}

		// Test with non-git directory
		tempDir := t.TempDir()
		if isGitRepository(tempDir) {
			t.Errorf("Expected %s to not be recognized as git repository", tempDir)
		}
	})

	t.Run("findLicenseFile", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)
		license, err := findLicenseFile(repo.Path)
		if err != nil {
			t.Fatalf("Expected to find license file, got error: %v", err)
		}
		if license != "LICENSE" {
			t.Errorf("Expected license file 'LICENSE', got '%s'", license)
		}
	})

	t.Run("findAndReadReadme", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)
		readme, err := findAndReadReadme(repo.Path)
		if err != nil {
			t.Fatalf("Expected to find and read README, got error: %v", err)
		}
		if !strings.Contains(readme, "Test Repository") {
			t.Errorf("Expected README to contain 'Test Repository', got: %s", readme)
		}
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("very large repository", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)

		// Create many files
		repo.CreateManyFiles("large", 100)
		repo.AddCommit("Add many files")

		// Test listing files with limit
		files, err := ListFiles(repo.Path, "large", false, nil, nil, 10)
		if err != nil {
			t.Fatalf("Failed to list files: %v", err)
		}

		if len(files) > 10 {
			t.Errorf("Expected at most 10 files, got %d", len(files))
		}
	})

	t.Run("file with special characters", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)

		// Create file with special characters
		specialFile := "special-file_123.txt"
		specialContent := "Content with special characters: åäö, 中文, 🎉\n"
		repo.WriteFile(specialFile, specialContent)
		repo.AddCommit("Add special file")

		// Test reading the file
		content, err := GetFileContent(repo.Path, specialFile, 0)
		if err != nil {
			t.Fatalf("Failed to read special file: %v", err)
		}

		if content != specialContent {
			t.Errorf("Special character content mismatch")
		}
	})

	t.Run("empty repository", func(t *testing.T) {
		repo := CreateTestRepository(t)

		// Test operations on empty repository
		info, err := GetRepositoryInfo(repo.Path)
		if err != nil {
			t.Fatalf("Failed to get info for empty repo: %v", err)
		}

		// Empty repo should have a valid path
		if info.Path == "" {
			t.Errorf("Expected path to be set for empty repo")
		}
	})
}
