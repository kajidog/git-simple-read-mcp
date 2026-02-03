package main

import (
	"os"
	"strings"
	"testing"
)

// TestPatternFilteringListFiles tests file pattern filtering in ListFiles
func TestPatternFilteringListFiles(t *testing.T) {
	// Initialize workspace for testing
	if err := InitializeWorkspace("./test_workspace_patterns"); err != nil {
		t.Fatalf("Failed to initialize workspace: %v", err)
	}
	defer os.RemoveAll("./test_workspace_patterns")

	repo := CreateTestRepository(t)

	// Create test files with different extensions
	testFiles := map[string]string{
		"main.go":           "package main\nfunc main() {}\n",
		"utils.go":          "package utils\nfunc Helper() {}\n",
		"config.json":       `{"name": "test"}`,
		"README.md":         "# Test Project",
		"src/app.js":        "console.log('hello');",
		"src/styles.css":    "body { margin: 0; }",
		"test/main_test.go": "package main\nimport \"testing\"",
	}

	for path, content := range testFiles {
		repo.WriteFile(path, content)
	}
	repo.AddCommit("Add test files")

	tests := []struct {
		name             string
		includePatterns  []string
		excludePatterns  []string
		expectedFiles    []string
		notExpectedFiles []string
	}{
		{
			name:             "Include Go files only",
			includePatterns:  []string{"*.go"},
			expectedFiles:    []string{"main.go", "utils.go", "test/main_test.go"},
			notExpectedFiles: []string{"config.json", "README.md", "src/app.js"},
		},
		{
			name:             "Exclude test files",
			excludePatterns:  []string{"*_test.go", "test/*"},
			expectedFiles:    []string{"main.go", "utils.go", "config.json"},
			notExpectedFiles: []string{"test/main_test.go"},
		},
		{
			name:             "Include JS and CSS, exclude test directory",
			includePatterns:  []string{"*.js", "*.css"},
			excludePatterns:  []string{"test/*"},
			expectedFiles:    []string{"src/app.js", "src/styles.css"},
			notExpectedFiles: []string{"main.go", "test/main_test.go"},
		},
		{
			name:             "Include src directory files",
			includePatterns:  []string{"src/*"},
			expectedFiles:    []string{"src/app.js", "src/styles.css"},
			notExpectedFiles: []string{"main.go", "config.json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := ListFiles(repo.Path, ".", true, tt.includePatterns, tt.excludePatterns, 100)
			if err != nil {
				t.Fatalf("ListFiles failed: %v", err)
			}

			// Check expected files are included
			for _, expectedFile := range tt.expectedFiles {
				found := false
				for _, file := range files {
					if file.Path == expectedFile {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected file %s not found in results", expectedFile)
				}
			}

			// Check not expected files are excluded
			for _, notExpectedFile := range tt.notExpectedFiles {
				for _, file := range files {
					if file.Path == notExpectedFile {
						t.Errorf("Unexpected file %s found in results", notExpectedFile)
					}
				}
			}
		})
	}
}

// TestCharacterCountListFiles tests character and line count functionality
func TestCharacterCountListFiles(t *testing.T) {
	// Initialize workspace for testing
	if err := InitializeWorkspace("./test_workspace_charcount"); err != nil {
		t.Fatalf("Failed to initialize workspace: %v", err)
	}
	defer os.RemoveAll("./test_workspace_charcount")

	repo := CreateTestRepository(t)

	// Create test files with known content
	testFiles := map[string]struct {
		content       string
		expectedChars int
		expectedLines int
	}{
		"small.txt": {
			content:       "Hello\nWorld",
			expectedChars: 11, // "Hello\nWorld" = 11 chars
			expectedLines: 2,
		},
		"empty.txt": {
			content:       "",
			expectedChars: 0,
			expectedLines: 0,
		},
		"single_line.txt": {
			content:       "Single line without newline",
			expectedChars: 28,
			expectedLines: 1,
		},
		"multi_line.txt": {
			content:       "Line 1\nLine 2\nLine 3\n",
			expectedChars: 20,
			expectedLines: 3,
		},
	}

	for path, fileData := range testFiles {
		repo.WriteFile(path, fileData.content)
	}
	repo.AddCommit("Add test files for character count")

	files, err := ListFiles(repo.Path, ".", false, nil, nil, 100)
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}

	for _, file := range files {
		if expected, exists := testFiles[file.Name]; exists {
			if file.LineCount != expected.expectedLines {
				t.Errorf("File %s: expected %d lines, got %d", file.Name, expected.expectedLines, file.LineCount)
			}
		}
	}
}

// TestPatternFilteringSearchFiles tests file pattern filtering in SearchFiles
func TestPatternFilteringSearchFiles(t *testing.T) {
	// Initialize workspace for testing
	if err := InitializeWorkspace("./test_workspace_search_patterns"); err != nil {
		t.Fatalf("Failed to initialize workspace: %v", err)
	}
	defer os.RemoveAll("./test_workspace_search_patterns")

	repo := CreateTestRepository(t)

	// Create test files with search content
	testFiles := map[string]string{
		"main.go":           "package main\nfunc main() {\n\tfmt.Println(\"hello world\")\n}",
		"utils.go":          "package utils\nfunc Hello() string {\n\treturn \"hello\"\n}",
		"config.json":       `{"message": "hello world"}`,
		"README.md":         "# Hello World\nThis is a hello world example",
		"src/app.js":        "console.log('hello world');",
		"test/main_test.go": "package main\nimport \"testing\"\nfunc TestHello(t *testing.T) {}",
	}

	for path, content := range testFiles {
		repo.WriteFile(path, content)
	}
	repo.AddCommit("Add test files for search patterns")

	tests := []struct {
		name             string
		keywords         []string
		includePatterns  []string
		excludePatterns  []string
		expectedFiles    []string
		notExpectedFiles []string
	}{
		{
			name:             "Search 'hello' in Go files only",
			keywords:         []string{"hello"},
			includePatterns:  []string{"*.go"},
			expectedFiles:    []string{"main.go", "utils.go", "test/main_test.go"},
			notExpectedFiles: []string{"config.json", "README.md", "src/app.js"},
		},
		{
			name:             "Search 'hello' excluding test files",
			keywords:         []string{"hello"},
			excludePatterns:  []string{"*_test.go", "test/*"},
			expectedFiles:    []string{"main.go", "utils.go", "config.json", "README.md", "src/app.js"},
			notExpectedFiles: []string{"test/main_test.go"},
		},
		{
			name:             "Search 'world' in JSON and JS files",
			keywords:         []string{"world"},
			includePatterns:  []string{"*.json", "*.js"},
			expectedFiles:    []string{"config.json", "src/app.js"},
			notExpectedFiles: []string{"main.go", "README.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := SearchFiles(repo.Path, tt.keywords, "and", false, 0, tt.includePatterns, tt.excludePatterns, 100)
			if err != nil {
				t.Fatalf("SearchFiles failed: %v", err)
			}

			// Check expected files are included
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

			// Check not expected files are excluded
			for _, notExpectedFile := range tt.notExpectedFiles {
				for _, result := range results {
					if result.Path == notExpectedFile {
						t.Errorf("Unexpected file %s found in search results", notExpectedFile)
					}
				}
			}
		})
	}
}

// TestMultipleFileContentRetrieval tests getting content from multiple files
func TestMultipleFileContentRetrieval(t *testing.T) {
	// Initialize workspace for testing
	if err := InitializeWorkspace("./test_workspace_multifile"); err != nil {
		t.Fatalf("Failed to initialize workspace: %v", err)
	}
	defer os.RemoveAll("./test_workspace_multifile")

	repo := CreateTestRepository(t)

	// Create test files
	testFiles := map[string]string{
		"file1.txt":       "Content of file 1\nSecond line",
		"file2.txt":       "Content of file 2",
		"file3.txt":       "Content of file 3\nLine 2\nLine 3",
		"nonexistent.txt": "", // This will not be created
	}

	for path, content := range testFiles {
		if path != "nonexistent.txt" {
			repo.WriteFile(path, content)
		}
	}
	repo.AddCommit("Add test files for multi-file content")

	t.Run("Get multiple existing files", func(t *testing.T) {
		filePaths := []string{"file1.txt", "file2.txt", "file3.txt"}
		results, err := GetMultipleFileContents(repo.Path, filePaths, 0)
		if err != nil {
			t.Fatalf("GetMultipleFileContents failed: %v", err)
		}

		if len(results) != 3 {
			t.Fatalf("Expected 3 results, got %d", len(results))
		}

		for i, result := range results {
			expectedPath := filePaths[i]
			if result.FilePath != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, result.FilePath)
			}

			if result.Error != "" {
				t.Errorf("Unexpected error for %s: %s", result.FilePath, result.Error)
			}

			expectedContent := testFiles[expectedPath]
			if !strings.Contains(result.Content, expectedContent) {
				t.Errorf("Content mismatch for %s", result.FilePath)
			}
		}
	})

	t.Run("Get files with some non-existent", func(t *testing.T) {
		filePaths := []string{"file1.txt", "nonexistent.txt", "file2.txt"}
		results, err := GetMultipleFileContents(repo.Path, filePaths, 0)
		if err != nil {
			t.Fatalf("GetMultipleFileContents failed: %v", err)
		}

		if len(results) != 3 {
			t.Fatalf("Expected 3 results, got %d", len(results))
		}

		// Check first file (should succeed)
		if results[0].Error != "" {
			t.Errorf("Unexpected error for file1.txt: %s", results[0].Error)
		}

		// Check non-existent file (should have error)
		if results[1].Error == "" {
			t.Errorf("Expected error for nonexistent.txt but got none")
		}

		// Check third file (should succeed)
		if results[2].Error != "" {
			t.Errorf("Unexpected error for file2.txt: %s", results[2].Error)
		}
	})

	t.Run("Get files with line limit", func(t *testing.T) {
		filePaths := []string{"file3.txt"}
		results, err := GetMultipleFileContents(repo.Path, filePaths, 2)
		if err != nil {
			t.Fatalf("GetMultipleFileContents failed: %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}

		result := results[0]
		lines := strings.Split(strings.TrimSpace(result.Content), "\n")
		if len(lines) > 2 {
			t.Errorf("Expected maximum 2 lines, got %d", len(lines))
		}
	})
}

// TestPatternHelperFunctions tests the new helper functions
func TestPatternHelperFunctions(t *testing.T) {
	t.Run("matchesPatterns", func(t *testing.T) {
		tests := []struct {
			filePath string
			patterns []string
			expected bool
		}{
			{"main.go", []string{"*.go"}, true},
			{"test.js", []string{"*.go"}, false},
			{"src/app.js", []string{"src/*"}, true},
			{"src/app.js", []string{"*.js"}, true},
			{"test/main_test.go", []string{"*_test.go"}, true},
			{"main.go", []string{"*.js", "*.go"}, true},
			{"README.md", []string{}, true}, // Empty patterns should match all
		}

		for _, tt := range tests {
			result := matchesPatterns(tt.filePath, tt.patterns)
			if result != tt.expected {
				t.Errorf("matchesPatterns(%q, %v) = %v, want %v",
					tt.filePath, tt.patterns, result, tt.expected)
			}
		}
	})

	t.Run("shouldIncludeFile", func(t *testing.T) {
		tests := []struct {
			filePath        string
			includePatterns []string
			excludePatterns []string
			expected        bool
		}{
			{"main.go", []string{"*.go"}, []string{}, true},
			{"main.go", []string{"*.js"}, []string{}, false},
			{"test.go", []string{"*.go"}, []string{"test*"}, false},
			{"main.go", []string{}, []string{"test*"}, true},
			{"test.go", []string{}, []string{"test*"}, false},
		}

		for _, tt := range tests {
			result := shouldIncludeFile(tt.filePath, tt.includePatterns, tt.excludePatterns)
			if result != tt.expected {
				t.Errorf("shouldIncludeFile(%q, %v, %v) = %v, want %v",
					tt.filePath, tt.includePatterns, tt.excludePatterns, result, tt.expected)
			}
		}
	})
}

// TestDirectoryPatternExclusion tests directory patterns with trailing slash
func TestDirectoryPatternExclusion(t *testing.T) {
	t.Run("matchesPattern with directory patterns", func(t *testing.T) {
		tests := []struct {
			filePath string
			pattern  string
			expected bool
		}{
			// Directory patterns with trailing /
			{"vendor/foo/bar.go", "vendor/", true},
			{"vendor/package/main.go", "vendor/", true},
			{"node_modules/react/index.js", "node_modules/", true},
			{"src/vendor/local.go", "vendor/", false}, // nested vendor should not match
			{"vendorfile.go", "vendor/", false},       // file starting with vendor should not match

			// Recursive patterns with **
			{"vendor/foo/bar.go", "vendor/**", true},
			{"vendor/deep/nested/file.go", "vendor/**", true},
			{"src/test/file.go", "**/test/**", true},
			{"test/file.go", "**/test/**", true},
			{"vendor/foo.go", "vendor/**/*.go", true},
			{"vendor/sub/bar.go", "vendor/**/*.go", true},

			// Directory boundary tests - must not match similar prefixes
			{"vendor2/foo.go", "vendor/**", false},        // vendor2 should NOT match vendor/**
			{"vendor2/sub/bar.go", "vendor/**/*.go", false}, // vendor2 should NOT match vendor/**/*.go
			{"node_modules2/pkg/index.js", "node_modules/**", false},
			{"vendors/dep/main.go", "vendor/**", false},   // vendors should NOT match vendor/**

			// Simple patterns should still work
			{"vendor/foo.go", "vendor/*", true},
			{"main.go", "*.go", true},
		}

		for _, tt := range tests {
			result := matchesPattern(tt.filePath, tt.pattern)
			if result != tt.expected {
				t.Errorf("matchesPattern(%q, %q) = %v, want %v",
					tt.filePath, tt.pattern, result, tt.expected)
			}
		}
	})

	t.Run("shouldSkipDirectory", func(t *testing.T) {
		tests := []struct {
			dirPath  string
			patterns []string
			expected bool
		}{
			// Directory patterns with trailing /
			{"vendor", []string{"vendor/"}, true},
			{"vendor/sub", []string{"vendor/"}, true},
			{"node_modules", []string{"node_modules/"}, true},
			{"src", []string{"vendor/"}, false},

			// Recursive patterns
			{"vendor", []string{"vendor/**"}, true},
			{"vendor/sub", []string{"vendor/**"}, true},

			// Simple directory name
			{"vendor", []string{"vendor"}, true},
			{"vendor/sub", []string{"vendor"}, true},
			{"vendorfile", []string{"vendor"}, false},

			// Path patterns
			{"src/vendor", []string{"src/vendor/*"}, true},
			{"vendor", []string{"src/vendor/*"}, false},

			// Multiple patterns
			{"vendor", []string{"node_modules/", "vendor/"}, true},
			{"node_modules", []string{"node_modules/", "vendor/"}, true},
			{"src", []string{"node_modules/", "vendor/"}, false},

			// Empty patterns
			{"vendor", []string{}, false},
		}

		for _, tt := range tests {
			result := shouldSkipDirectory(tt.dirPath, tt.patterns)
			if result != tt.expected {
				t.Errorf("shouldSkipDirectory(%q, %v) = %v, want %v",
					tt.dirPath, tt.patterns, result, tt.expected)
			}
		}
	})
}

// TestListFilesWithDirectoryExclusion tests that ListFiles properly skips excluded directories
func TestListFilesWithDirectoryExclusion(t *testing.T) {
	// Initialize workspace for testing
	if err := InitializeWorkspace("./test_workspace_dir_exclusion"); err != nil {
		t.Fatalf("Failed to initialize workspace: %v", err)
	}
	defer os.RemoveAll("./test_workspace_dir_exclusion")

	repo := CreateTestRepository(t)

	// Create test files including vendor and node_modules directories
	testFiles := map[string]string{
		"main.go":                          "package main",
		"utils.go":                         "package utils",
		"vendor/dep/main.go":               "package dep",
		"vendor/dep/sub/util.go":           "package sub",
		"node_modules/react/index.js":      "module.exports = {}",
		"node_modules/react/lib/core.js":   "exports.core = {}",
		"src/app.go":                       "package app",
		"src/vendor/local.go":              "package local", // nested vendor should be included
	}

	for path, content := range testFiles {
		repo.WriteFile(path, content)
	}
	repo.AddCommit("Add test files with vendor and node_modules")

	tests := []struct {
		name             string
		excludePatterns  []string
		expectedFiles    []string
		notExpectedFiles []string
	}{
		{
			name:             "Exclude vendor directory with trailing slash",
			excludePatterns:  []string{"vendor/"},
			expectedFiles:    []string{"main.go", "utils.go", "src/app.go", "src/vendor/local.go"},
			notExpectedFiles: []string{"vendor/dep/main.go", "vendor/dep/sub/util.go"},
		},
		{
			name:             "Exclude node_modules with recursive pattern",
			excludePatterns:  []string{"node_modules/**"},
			expectedFiles:    []string{"main.go", "src/app.go"},
			notExpectedFiles: []string{"node_modules/react/index.js", "node_modules/react/lib/core.js"},
		},
		{
			name:             "Exclude multiple directories",
			excludePatterns:  []string{"vendor/", "node_modules/"},
			expectedFiles:    []string{"main.go", "utils.go", "src/app.go", "src/vendor/local.go"},
			notExpectedFiles: []string{"vendor/dep/main.go", "node_modules/react/index.js"},
		},
		{
			name:             "Exclude with simple directory name",
			excludePatterns:  []string{"vendor"},
			expectedFiles:    []string{"main.go", "src/app.go", "src/vendor/local.go"},
			notExpectedFiles: []string{"vendor/dep/main.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := ListFiles(repo.Path, ".", true, nil, tt.excludePatterns, 100)
			if err != nil {
				t.Fatalf("ListFiles failed: %v", err)
			}

			// Check expected files are included
			for _, expectedFile := range tt.expectedFiles {
				found := false
				for _, file := range files {
					if file.Path == expectedFile {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected file %s not found in results", expectedFile)
				}
			}

			// Check not expected files are excluded
			for _, notExpectedFile := range tt.notExpectedFiles {
				for _, file := range files {
					if file.Path == notExpectedFile {
						t.Errorf("Unexpected file %s found in results", notExpectedFile)
					}
				}
			}
		})
	}
}
