package main

import (
	"fmt"
	"strings"
	"testing"
)

// TestGetReadmeFiles tests the README file discovery functionality
func TestGetReadmeFiles(t *testing.T) {
	repo := CreateTestRepositoryWithContent(t)

	// Create various README files
	testFiles := map[string]string{
		"README.md":         "# Main README\nThis is the main README file.",
		"README.txt":        "Main README in txt format",
		"docs/README.md":    "# Documentation README\nDocs specific README.",
		"src/readme.txt":    "Source code readme file",
		"test/Readme.md":    "# Test README\nTesting documentation.",
		"other.md":          "This is not a README file",
		"READMENOT.md":      "This should not be matched",
		"sub/dir/README.md": "# Sub directory README",
	}

	for path, content := range testFiles {
		repo.WriteFile(path, content)
	}
	repo.AddCommit("Add test files for README discovery")

	t.Run("Get README files non-recursive", func(t *testing.T) {
		readmeFiles, err := GetReadmeFiles(repo.Path, false)
		if err != nil {
			t.Fatalf("GetReadmeFiles failed: %v", err)
		}

		// CreateTestRepositoryWithContent already creates README.md, plus we add our test files
		expectedFiles := []string{"README.md", "README.txt"}
		if len(readmeFiles) < len(expectedFiles) {
			t.Fatalf("Expected at least %d README files, got %d", len(expectedFiles), len(readmeFiles))
		}

		for _, expected := range expectedFiles {
			found := false
			for _, readme := range readmeFiles {
				if readme.Path == expected {
					found = true
					if readme.Size == 0 {
						t.Errorf("Expected non-zero size for %s", expected)
					}
					if readme.LineCount == 0 {
						t.Errorf("Expected non-zero line count for %s", expected)
					}
					break
				}
			}
			if !found {
				t.Errorf("Expected README file %s not found", expected)
			}
		}
	})

	t.Run("Get README files recursive", func(t *testing.T) {
		readmeFiles, err := GetReadmeFiles(repo.Path, true)
		if err != nil {
			t.Fatalf("GetReadmeFiles failed: %v", err)
		}

		// Should find all our test README files
		expectedFiles := []string{"README.md", "README.txt", "docs/README.md", "src/readme.txt", "test/Readme.md", "sub/dir/README.md"}
		if len(readmeFiles) < len(expectedFiles) {
			t.Fatalf("Expected at least %d README files, got %d", len(expectedFiles), len(readmeFiles))
		}

		for _, expected := range expectedFiles {
			found := false
			for _, readme := range readmeFiles {
				if readme.Path == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected README file %s not found", expected)
			}
		}

		// Check that non-README files are not included
		notExpectedFiles := []string{"other.md", "READMENOT.md"}
		for _, notExpected := range notExpectedFiles {
			for _, readme := range readmeFiles {
				if readme.Path == notExpected {
					t.Errorf("Unexpected file %s found in README results", notExpected)
				}
			}
		}
	})
}

// TestGetFileContentWithLineNumbers tests the line numbers functionality
func TestGetFileContentWithLineNumbers(t *testing.T) {
	repo := CreateTestRepositoryWithContent(t)

	// Create test files
	testContent := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	repo.WriteFile("test.txt", testContent)
	repo.AddCommit("Add test file for line numbers")

	t.Run("Get content without line numbers", func(t *testing.T) {
		content, totalLines, startLine, endLine, err := GetFileContentWithLineNumbers(repo.Path, "test.txt", 1, 0, false)
		if err != nil {
			t.Fatalf("GetFileContentWithLineNumbers failed: %v", err)
		}

		if totalLines != 5 {
			t.Errorf("Expected totalLines=5, got %d", totalLines)
		}
		if startLine != 1 || endLine != 5 {
			t.Errorf("Expected startLine=1, endLine=5, got %d-%d", startLine, endLine)
		}

		lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
		if len(lines) != 5 {
			t.Errorf("Expected 5 lines, got %d", len(lines))
		}

		// Check that no line numbers are present
		for i, line := range lines {
			expected := strings.Split(testContent, "\n")[i]
			if line != expected {
				t.Errorf("Line %d: expected %q, got %q", i+1, expected, line)
			}
		}
	})

	t.Run("Get content with line numbers", func(t *testing.T) {
		content, _, _, _, err := GetFileContentWithLineNumbers(repo.Path, "test.txt", 1, 0, true)
		if err != nil {
			t.Fatalf("GetFileContentWithLineNumbers failed: %v", err)
		}

		lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
		if len(lines) != 5 {
			t.Errorf("Expected 5 lines, got %d", len(lines))
		}

		// Check that line numbers are present and correct
		for i, line := range lines {
			expectedPrefix := fmt.Sprintf("%4d: ", i+1)
			if !strings.HasPrefix(line, expectedPrefix) {
				if len(line) > 10 {
					t.Errorf("Line %d: expected prefix %q, got %q", i+1, expectedPrefix, line[:10])
				} else {
					t.Errorf("Line %d: expected prefix %q, got %q", i+1, expectedPrefix, line)
				}
			}
		}
	})

	t.Run("Get content with line numbers and limit", func(t *testing.T) {
		content, totalLines, startLine, endLine, err := GetFileContentWithLineNumbers(repo.Path, "test.txt", 1, 3, true)
		if err != nil {
			t.Fatalf("GetFileContentWithLineNumbers failed: %v", err)
		}

		if totalLines != 5 {
			t.Errorf("Expected totalLines=5, got %d", totalLines)
		}
		if startLine != 1 || endLine != 3 {
			t.Errorf("Expected startLine=1, endLine=3, got %d-%d", startLine, endLine)
		}

		lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
		if len(lines) != 3 {
			t.Errorf("Expected 3 lines due to limit, got %d", len(lines))
		}

		// Check that line numbers are correct
		for i, line := range lines {
			expectedPrefix := fmt.Sprintf("%4d: ", i+1)
			if !strings.HasPrefix(line, expectedPrefix) {
				t.Errorf("Line %d: expected prefix %q, got %q", i+1, expectedPrefix, line[:5])
			}
		}
	})

	t.Run("Get content with startLine offset", func(t *testing.T) {
		content, totalLines, startLine, endLine, err := GetFileContentWithLineNumbers(repo.Path, "test.txt", 3, 0, true)
		if err != nil {
			t.Fatalf("GetFileContentWithLineNumbers failed: %v", err)
		}

		if totalLines != 5 {
			t.Errorf("Expected totalLines=5, got %d", totalLines)
		}
		if startLine != 3 || endLine != 5 {
			t.Errorf("Expected startLine=3, endLine=5, got %d-%d", startLine, endLine)
		}

		lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
		if len(lines) != 3 {
			t.Errorf("Expected 3 lines (from line 3 to 5), got %d", len(lines))
		}

		// Check that line numbers start from 3
		if !strings.HasPrefix(lines[0], "   3: ") {
			t.Errorf("First line should start with '   3: ', got %q", lines[0][:6])
		}
	})
}

// TestGetMultipleFileContentsWithLineNumbers tests multiple file content with line numbers
func TestGetMultipleFileContentsWithLineNumbers(t *testing.T) {
	repo := CreateTestRepositoryWithContent(t)

	// Create test files
	testFiles := map[string]string{
		"file1.txt": "First line\nSecond line",
		"file2.txt": "Only one line",
		"file3.txt": "Line A\nLine B\nLine C",
	}

	for path, content := range testFiles {
		repo.WriteFile(path, content)
	}
	repo.AddCommit("Add test files for multiple file line numbers")

	t.Run("Get multiple files with line numbers", func(t *testing.T) {
		filePaths := []string{"file1.txt", "file2.txt", "file3.txt"}
		results, err := GetMultipleFileContentsWithLineNumbers(repo.Path, filePaths, 1, 0, true)
		if err != nil {
			t.Fatalf("GetMultipleFileContentsWithLineNumbers failed: %v", err)
		}

		if len(results) != 3 {
			t.Fatalf("Expected 3 results, got %d", len(results))
		}

		// Check file1.txt (2 lines)
		lines1 := strings.Split(strings.TrimSuffix(results[0].Content, "\n"), "\n")
		if len(lines1) != 2 {
			t.Errorf("File1: expected 2 lines, got %d", len(lines1))
		}
		if !strings.HasPrefix(lines1[0], "   1: ") {
			t.Errorf("File1 line 1: expected line number prefix")
		}
		if results[0].TotalLines != 2 {
			t.Errorf("File1: expected TotalLines=2, got %d", results[0].TotalLines)
		}

		// Check file2.txt (1 line)
		lines2 := strings.Split(strings.TrimSuffix(results[1].Content, "\n"), "\n")
		if len(lines2) != 1 {
			t.Errorf("File2: expected 1 line, got %d", len(lines2))
		}
		if !strings.HasPrefix(lines2[0], "   1: ") {
			t.Errorf("File2 line 1: expected line number prefix")
		}
		if results[1].TotalLines != 1 {
			t.Errorf("File2: expected TotalLines=1, got %d", results[1].TotalLines)
		}

		// Check file3.txt (3 lines)
		lines3 := strings.Split(strings.TrimSuffix(results[2].Content, "\n"), "\n")
		if len(lines3) != 3 {
			t.Errorf("File3: expected 3 lines, got %d", len(lines3))
		}
		if !strings.HasPrefix(lines3[2], "   3: ") {
			t.Errorf("File3 line 3: expected line number prefix")
		}
		if results[2].TotalLines != 3 {
			t.Errorf("File3: expected TotalLines=3, got %d", results[2].TotalLines)
		}
	})

	t.Run("Get multiple files without line numbers", func(t *testing.T) {
		filePaths := []string{"file1.txt", "file2.txt"}
		results, err := GetMultipleFileContentsWithLineNumbers(repo.Path, filePaths, 1, 0, false)
		if err != nil {
			t.Fatalf("GetMultipleFileContentsWithLineNumbers failed: %v", err)
		}

		if len(results) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(results))
		}

		// Check that no line numbers are present
		for _, result := range results {
			lines := strings.Split(strings.TrimSuffix(result.Content, "\n"), "\n")
			for _, line := range lines {
				if strings.Contains(line, ":") && len(line) > 4 && line[4] == ':' {
					t.Errorf("Found unexpected line number prefix in content without line numbers")
				}
			}
		}
	})
}