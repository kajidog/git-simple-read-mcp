package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestEdgeCasesAndErrorHandling tests various edge cases and error conditions
func TestEdgeCasesAndErrorHandling(t *testing.T) {
	t.Run("invalid paths", func(t *testing.T) {
		invalidPaths := []string{
			"/nonexistent/path",
			"",
			"\x00invalid",
			strings.Repeat("a", 1000), // Very long path
		}

		for _, path := range invalidPaths {
			t.Run("path_"+path, func(t *testing.T) {
				_, err := GetRepositoryInfo(path)
				if err == nil {
					t.Errorf("Expected error for invalid path: %s", path)
				}
			})
		}
	})

	t.Run("permission denied", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("Skipping permission test as root")
		}

		repo := CreateTestRepositoryWithContent(t)

		// Create a directory with restricted permissions
		restrictedDir := filepath.Join(repo.Path, "restricted")
		os.Mkdir(restrictedDir, 0000)
		defer os.Chmod(restrictedDir, 0755) // Cleanup

		// Test should handle permission errors gracefully
		files, err := ListFiles(repo.Path, "restricted", false, nil, nil, 10)
		// Should either succeed with empty list or fail gracefully
		if err != nil {
			t.Logf("Permission error handled: %v", err)
		} else {
			t.Logf("Listed %d files in restricted directory", len(files))
		}
	})

	t.Run("special characters in filenames", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)

		specialFiles := []string{
			"file with spaces.txt",
			"file-with-dashes.txt",
			"file_with_underscores.txt",
			"file.with.dots.txt",
			"файл-кириллица.txt",
			"文件-中文.txt",
			"ファイル-日本語.txt",
			"émoji-🎉.txt",
		}

		for _, filename := range specialFiles {
			if !utf8.ValidString(filename) {
				continue // Skip invalid UTF-8
			}

			content := fmt.Sprintf("Content of %s", filename)
			repo.WriteFile(filename, content)
		}
		repo.AddCommit("Add files with special characters")

		// Test file listing
		files, err := ListFiles(repo.Path, ".", false, nil, nil, 50)
		if err != nil {
			t.Fatalf("Failed to list files with special characters: %v", err)
		}

		// Test file content reading
		for _, filename := range specialFiles {
			if !utf8.ValidString(filename) {
				continue
			}

			content, err := GetFileContent(repo.Path, filename, 0)
			if err != nil {
				t.Errorf("Failed to read file %s: %v", filename, err)
				continue
			}

			expectedContent := fmt.Sprintf("Content of %s\n", filename)
			if content != expectedContent {
				t.Errorf("Content mismatch for %s", filename)
			}
		}

		// Verify files appear in listing
		fileMap := make(map[string]bool)
		for _, file := range files {
			fileMap[file.Name] = true
		}

		for _, filename := range specialFiles {
			if !utf8.ValidString(filename) {
				continue
			}
			if !fileMap[filename] {
				t.Errorf("Special file %s not found in listing", filename)
			}
		}
	})

	t.Run("binary files", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)

		// Create binary file content
		binaryContent := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
		binaryFile := "binary.dat"

		fullPath := filepath.Join(repo.Path, binaryFile)
		err := os.WriteFile(fullPath, binaryContent, 0644)
		if err != nil {
			t.Fatalf("Failed to write binary file: %v", err)
		}

		repo.runGitCommand("add", binaryFile)
		repo.runGitCommand("commit", "-m", "Add binary file")

		// Test reading binary file
		content, err := GetFileContent(repo.Path, binaryFile, 0)
		if err != nil {
			t.Errorf("Failed to read binary file: %v", err)
		} else {
			t.Logf("Binary file content length: %d", len(content))
		}
	})

	t.Run("empty files and directories", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)

		// Create empty file
		repo.WriteFile("empty.txt", "")

		// Create empty directory (Git doesn't track empty dirs, so add a placeholder)
		emptyDirPath := filepath.Join(repo.Path, "empty_dir")
		os.Mkdir(emptyDirPath, 0755)
		repo.WriteFile("empty_dir/.gitkeep", "")

		repo.AddCommit("Add empty file and directory")

		// Test reading empty file
		content, err := GetFileContent(repo.Path, "empty.txt", 0)
		if err != nil {
			t.Errorf("Failed to read empty file: %v", err)
		}
		if content != "" {
			t.Errorf("Expected empty content, got: %q", content)
		}

		// Test listing empty directory
		files, err := ListFiles(repo.Path, "empty_dir", false, nil, nil, 10)
		if err != nil {
			t.Errorf("Failed to list empty directory: %v", err)
		}
		if len(files) == 0 {
			t.Logf("Empty directory listing returned no files (expected)")
		}
	})

	t.Run("corrupted repository", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)

		// Corrupt the .git directory by removing critical files
		gitDir := filepath.Join(repo.Path, ".git")
		headFile := filepath.Join(gitDir, "HEAD")

		// Backup and remove HEAD file
		headContent, err := os.ReadFile(headFile)
		if err != nil {
			t.Fatalf("Failed to read HEAD file: %v", err)
		}

		os.Remove(headFile)

		// Test operations on corrupted repository
		_, err = GetRepositoryInfo(repo.Path)
		if err == nil {
			t.Errorf("Expected error for corrupted repository")
		}

		_, err = ListBranches(repo.Path)
		if err == nil {
			t.Errorf("Expected error for corrupted repository")
		}

		// Restore HEAD file for cleanup
		os.WriteFile(headFile, headContent, 0644)
	})

	t.Run("very long file content", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)

		// Create file with many lines
		longContent := strings.Repeat("This is a very long line with lots of text to test line limiting functionality.\n", 1000)
		repo.WriteFile("long_file.txt", longContent)
		repo.AddCommit("Add long file")

		// Test with line limit
		content, err := GetFileContent(repo.Path, "long_file.txt", 10)
		if err != nil {
			t.Errorf("Failed to read long file with limit: %v", err)
		} else {
			lines := strings.Split(content, "\n")
			if len(lines) > 11 { // 10 lines + potential empty line
				t.Errorf("Expected at most 11 lines, got %d", len(lines))
			}
		}

		// Test without limit (should still complete reasonably fast)
		start := time.Now()
		content, err = GetFileContent(repo.Path, "long_file.txt", 0)
		elapsed := time.Since(start)

		if err != nil {
			t.Errorf("Failed to read long file without limit: %v", err)
		} else if elapsed > 5*time.Second {
			t.Errorf("Reading long file took too long: %v", elapsed)
		}
	})

	t.Run("symbolic links", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)

		// Create a symbolic link
		targetFile := "target.txt"
		linkFile := "link.txt"

		repo.WriteFile(targetFile, "Target file content")

		targetPath := filepath.Join(repo.Path, targetFile)
		linkPath := filepath.Join(repo.Path, linkFile)

		err := os.Symlink(targetPath, linkPath)
		if err != nil {
			t.Skip("Symbolic links not supported on this system")
		}

		repo.runGitCommand("add", ".")
		repo.runGitCommand("commit", "-m", "Add symbolic link")

		// Test listing files with symlinks
		files, err := ListFiles(repo.Path, ".", false, nil, nil, 20)
		if err != nil {
			t.Errorf("Failed to list files with symlinks: %v", err)
		} else {
			// Should include both target and link
			foundTarget := false
			foundLink := false
			for _, file := range files {
				if file.Name == targetFile {
					foundTarget = true
				}
				if file.Name == linkFile {
					foundLink = true
				}
			}
			if !foundTarget || !foundLink {
				t.Errorf("Symlink or target not found in file listing")
			}
		}
	})
}

// TestMCPToolEdgeCases tests edge cases in MCP tool handlers
func TestMCPToolEdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("nil parameters", func(t *testing.T) {
		// Test handlers with nil parameters should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Handler panicked with nil params: %v", r)
			}
		}()

		// This should fail gracefully, not panic
		params := GetRepositoryInfoParams{}

		result, _, err := handleGetRepositoryInfo(ctx, nil, params)
		if err != nil {
			t.Errorf("Handler returned unexpected error: %v", err)
		}
		if !result.IsError {
			t.Errorf("Expected error result for empty params")
		}
	})

	t.Run("extreme parameter values", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)

		// Test with very large limits
		params := ListFilesParams{
			Repository: repo.Path,
			Directory:  ".",
			Recursive:  true,
			Limit:      1000000, // Very large limit
		}

		result, _, err := handleListFiles(ctx, nil, params)
		if err != nil {
			t.Errorf("Handler failed with large limit: %v", err)
		}
		if result.IsError {
			t.Errorf("Handler returned error with large limit")
		}

		// Test with zero limit (should use default)
		params.Limit = 0
		result, _, err = handleListFiles(ctx, nil, params)
		if err != nil {
			t.Errorf("Handler failed with zero limit: %v", err)
		}
		if result.IsError {
			t.Errorf("Handler returned error with zero limit")
		}

		// Test with negative limit
		params.Limit = -1
		result, _, err = handleListFiles(ctx, nil, params)
		if err != nil {
			t.Errorf("Handler failed with negative limit: %v", err)
		}
		// Should handle gracefully (may treat as unlimited or default)
	})

	t.Run("malformed search keywords", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)

		malformedKeywords := [][]string{
			{""},                        // Empty keyword
			{"", "valid"},               // Mixed empty and valid
			{strings.Repeat("a", 1000)}, // Very long keyword
			{"keyword1", "keyword2", "keyword3", "keyword4", "keyword5"}, // Many keywords
		}

		for i, keywords := range malformedKeywords {
			t.Run(fmt.Sprintf("keywords_%d", i), func(t *testing.T) {
				params := SearchFilesParams{
					Repository: repo.Path,
					Keywords:   keywords,
					Limit:      10,
				}

				result, _, err := handleSearchFiles(ctx, nil, params)
				if err != nil {
					t.Errorf("Handler returned unexpected error: %v", err)
				}

				// Should handle gracefully, either succeeding or failing with proper error
				if result.IsError {
					content := result.Content[0].(*mcp.TextContent).Text
					t.Logf("Expected error for malformed keywords: %s", content)
				}
			})
		}
	})

	t.Run("invalid branch names", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)

		invalidBranches := []string{
			"",
			"invalid/branch/name",
			"branch with spaces",
			".invalid",
			"-invalid",
			"invalid.",
			strings.Repeat("a", 300), // Very long branch name
		}

		for _, branchName := range invalidBranches {
			t.Run("branch_"+branchName, func(t *testing.T) {
				params := SwitchBranchParams{
					Repository: repo.Path,
					Branch:     branchName,
				}

				result, _, err := handleSwitchBranch(ctx, nil, params)
				if err != nil {
					t.Errorf("Handler returned unexpected error: %v", err)
				}

				// Should return error for invalid branch names
				if !result.IsError && branchName != "" {
					t.Errorf("Expected error for invalid branch name: %s", branchName)
				}
			})
		}
	})

	t.Run("deeply nested paths", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)

		// Create deeply nested directory structure
		deepPath := "level1"
		for i := 2; i <= 20; i++ {
			deepPath = filepath.Join(deepPath, fmt.Sprintf("level%d", i))
		}

		repo.WriteFile(filepath.Join(deepPath, "deep.txt"), "Deep content")
		repo.AddCommit("Add deep structure")

		// Test listing deeply nested directory
		params := ListFilesParams{
			Repository: repo.Path,
			Directory:  deepPath,
			Recursive:  false,
			Limit:      10,
		}

		result, _, err := handleListFiles(ctx, nil, params)
		if err != nil {
			t.Errorf("Handler failed with deep path: %v", err)
		}

		if result.IsError {
			// May fail if path is too long for filesystem
			t.Logf("Deep path failed as expected: %s", result.Content[0].(*mcp.TextContent).Text)
		}
	})
}

// TestErrorRecovery tests error recovery and cleanup
func TestErrorRecovery(t *testing.T) {
	t.Run("cleanup after errors", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)

		// Cause an error and ensure state is not corrupted
		_, err := SwitchBranch(repo.Path, "nonexistent")
		if err == nil {
			t.Errorf("Expected error for nonexistent branch")
		}

		// Subsequent operations should still work
		info, err := GetRepositoryInfo(repo.Path)
		if err != nil {
			t.Errorf("Repository info failed after error: %v", err)
		}

		if info.CurrentBranch == "" {
			t.Errorf("Current branch should still be available after error")
		}
	})

	t.Run("partial failures", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)

		// Create some accessible and some inaccessible files
		repo.WriteFile("accessible.txt", "Accessible content")

		inaccessibleDir := filepath.Join(repo.Path, "inaccessible")
		os.Mkdir(inaccessibleDir, 0755)
		repo.WriteFile("inaccessible/file.txt", "Inaccessible content")

		if os.Getuid() != 0 {
			// Make directory inaccessible
			os.Chmod(inaccessibleDir, 0000)
			defer os.Chmod(inaccessibleDir, 0755) // Cleanup
		}

		// Should still list accessible files
		files, err := ListFiles(repo.Path, ".", true, nil, nil, 50)
		// May succeed with partial results or fail completely
		if err != nil {
			t.Logf("Partial failure handled: %v", err)
		} else {
			t.Logf("Listed %d files despite inaccessible directory", len(files))
		}
	})

	t.Run("git command failures", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)

		// Temporarily break git by moving .git directory
		gitDir := filepath.Join(repo.Path, ".git")
		backupDir := filepath.Join(repo.Path, ".git.backup")

		os.Rename(gitDir, backupDir)

		// Operations should fail gracefully
		_, err := GetRepositoryInfo(repo.Path)
		if err == nil {
			t.Errorf("Expected error when git directory is missing")
		}

		_, err = ListBranches(repo.Path)
		if err == nil {
			t.Errorf("Expected error when git directory is missing")
		}

		// Restore git directory
		os.Rename(backupDir, gitDir)

		// Operations should work again
		_, err = GetRepositoryInfo(repo.Path)
		if err != nil {
			t.Errorf("Repository info should work after restoring git directory: %v", err)
		} else {
			t.Logf("Successfully recovered after git directory restoration")
		}
	})
}

// TestBoundaryConditions tests boundary conditions
func TestBoundaryConditions(t *testing.T) {
	t.Run("zero-byte files", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)

		repo.WriteFile("zero_byte.txt", "")
		repo.AddCommit("Add zero byte file")

		content, err := GetFileContent(repo.Path, "zero_byte.txt", 0)
		if err != nil {
			t.Errorf("Failed to read zero-byte file: %v", err)
		}
		if content != "" {
			t.Errorf("Expected empty content, got: %q", content)
		}
	})

	t.Run("single character files", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)

		repo.WriteFile("single_char.txt", "a")
		repo.AddCommit("Add single character file")

		content, err := GetFileContent(repo.Path, "single_char.txt", 1)
		if err != nil {
			t.Errorf("Failed to read single character file: %v", err)
		}
		if content != "a\n" {
			t.Errorf("Expected 'a\\n', got: %q", content)
		}
	})

	t.Run("exactly at limits", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)

		// Create exactly 50 files to test pagination boundary
		repo.CreateManyFiles("boundary", 50)
		repo.AddCommit("Add boundary test files")

		// Test with limit exactly matching file count
		files, err := ListFiles(repo.Path, "boundary", false, nil, nil, 50)
		if err != nil {
			t.Errorf("Failed to list files at boundary: %v", err)
		}
		if len(files) != 50 {
			t.Errorf("Expected exactly 50 files, got %d", len(files))
		}

		// Test with limit one less than file count
		files, err = ListFiles(repo.Path, "boundary", false, nil, nil, 49)
		if err != nil {
			t.Errorf("Failed to list files below boundary: %v", err)
		}
		if len(files) != 49 {
			t.Errorf("Expected exactly 49 files, got %d", len(files))
		}
	})
}
