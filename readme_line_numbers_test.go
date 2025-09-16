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

// TestGetFileContentWithLineNumbers tests the line numbers and start line functionality
func TestGetFileContentWithLineNumbers(t *testing.T) {
	repo := CreateTestRepositoryWithContent(t)
	testFile := "test_file_for_reading.txt"
	var fileContent strings.Builder
	for i := 1; i <= 10; i++ {
		fileContent.WriteString(fmt.Sprintf("Line %d\n", i))
	}
	repo.WriteFile(testFile, fileContent.String())
	repo.AddCommit("Add test file for reading")

	tests := []struct {
		name               string
		filePath           string
		maxLines           int
		startLine          int
		showLineNumbers    bool
		expectError        bool
		expectedTotalLines int
		expectedStartLine  int
		expectedEndLine    int
		expectedContent    string
	}{
		{
			name:               "read from start",
			filePath:           testFile,
			maxLines:           3,
			startLine:          1,
			showLineNumbers:    false,
			expectError:        false,
			expectedTotalLines: 10,
			expectedStartLine:  1,
			expectedEndLine:    3,
			expectedContent:    "Line 1\nLine 2\nLine 3\n",
		},
		{
			name:               "read from middle",
			filePath:           testFile,
			maxLines:           2,
			startLine:          5,
			showLineNumbers:    false,
			expectError:        false,
			expectedTotalLines: 10,
			expectedStartLine:  5,
			expectedEndLine:    6,
			expectedContent:    "Line 5\nLine 6\n",
		},
		{
			name:               "read with line numbers",
			filePath:           testFile,
			maxLines:           2,
			startLine:          8,
			showLineNumbers:    true,
			expectError:        false,
			expectedTotalLines: 10,
			expectedStartLine:  8,
			expectedEndLine:    9,
			expectedContent:    "   8: Line 8\n   9: Line 9\n",
		},
		{
			name:               "read past end of file",
			filePath:           testFile,
			maxLines:           5,
			startLine:          9,
			showLineNumbers:    false,
			expectError:        false,
			expectedTotalLines: 10,
			expectedStartLine:  9,
			expectedEndLine:    10,
			expectedContent:    "Line 9\nLine 10\n",
		},
		{
			name:               "start line out of bounds",
			filePath:           testFile,
			maxLines:           5,
			startLine:          11,
			showLineNumbers:    false,
			expectError:        false,
			expectedTotalLines: 10,
			expectedStartLine:  11,
			expectedEndLine:    10, // Should report the last line number of the file
			expectedContent:    "",
		},
		{
			name:        "non-existent file",
			filePath:    "nonexistent.txt",
			maxLines:    0,
			startLine:   1,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetFileContentWithLineNumbers(repo.Path, tt.filePath, tt.maxLines, tt.startLine, tt.showLineNumbers)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				// Even on error, a partial result might be returned
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.TotalLines != tt.expectedTotalLines {
				t.Errorf("Expected TotalLines %d, got %d", tt.expectedTotalLines, result.TotalLines)
			}
			if result.StartLine != tt.expectedStartLine {
				t.Errorf("Expected StartLine %d, got %d", tt.expectedStartLine, result.StartLine)
			}
			if result.EndLine != tt.expectedEndLine {
				t.Errorf("Expected EndLine %d, got %d", tt.expectedEndLine, result.EndLine)
			}
			if result.Content != tt.expectedContent {
				t.Errorf("Content mismatch.\nExpected: %q\nActual:   %q", tt.expectedContent, result.Content)
			}
		})
	}
}

// TestGetMultipleFileContentsWithLineNumbers tests multiple file content with line numbers and start line
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

	t.Run("Get multiple files with line numbers and start line", func(t *testing.T) {
		filePaths := []string{"file1.txt", "file2.txt", "file3.txt"}
		results, err := GetMultipleFileContentsWithLineNumbers(repo.Path, filePaths, 1, 2, true)
		if err != nil {
			t.Fatalf("GetMultipleFileContentsWithLineNumbers failed: %v", err)
		}

		if len(results) != 3 {
			t.Fatalf("Expected 3 results, got %d", len(results))
		}

		// Check file1.txt (2 lines total, start at 2, max 1)
		res1 := results[0]
		if res1.TotalLines != 2 || res1.StartLine != 2 || res1.EndLine != 2 {
			t.Errorf("File1 metadata incorrect: Total %d, Start %d, End %d", res1.TotalLines, res1.StartLine, res1.EndLine)
		}
		if !strings.Contains(res1.Content, "   2: Second line") {
			t.Errorf("File1 content incorrect: %q", res1.Content)
		}

		// Check file2.txt (1 line total, start at 2 -> no content)
		res2 := results[1]
		if res2.TotalLines != 1 || res2.StartLine != 2 || res2.Content != "" {
			t.Errorf("File2 incorrect: Total %d, Start %d, Content %q", res2.TotalLines, res2.StartLine, res2.Content)
		}

		// Check file3.txt (3 lines total, start at 2, max 1)
		res3 := results[2]
		if res3.TotalLines != 3 || res3.StartLine != 2 || res3.EndLine != 2 {
			t.Errorf("File3 metadata incorrect: Total %d, Start %d, End %d", res3.TotalLines, res3.StartLine, res3.EndLine)
		}
		if !strings.Contains(res3.Content, "   2: Line B") {
			t.Errorf("File3 content incorrect: %q", res3.Content)
		}
	})

	t.Run("Get multiple files without line numbers", func(t *testing.T) {
		filePaths := []string{"file1.txt", "file2.txt"}
		results, err := GetMultipleFileContentsWithLineNumbers(repo.Path, filePaths, 0, 1, false)
		if err != nil {
			t.Fatalf("GetMultipleFileContentsWithLineNumbers failed: %v", err)
		}

		if len(results) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(results))
		}

		// Check that no line numbers are present
		for _, result := range results {
			if strings.Contains(result.Content, ": ") {
				// A simple check, might not be robust for all file contents
				if len(result.Content) > 4 && result.Content[4] == ':' {
					t.Errorf("Found unexpected line number prefix in content without line numbers")
				}
			}
		}
	})
}