package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

)

// TestPerformanceWithLargeRepository tests performance with large repositories
func TestPerformanceWithLargeRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	t.Run("large file count", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)
		
		// Create many files in different directories
		repo.CreateManyFiles("src", 100)
		repo.CreateManyFiles("tests", 50)
		repo.CreateManyFiles("docs", 30)
		repo.AddCommit("Add many files")
		
		start := time.Now()
		
		// Test file listing performance
		files, err := ListFiles(repo.Path, ".", true, 200)
		if err != nil {
			t.Fatalf("Failed to list files: %v", err)
		}
		
		elapsed := time.Since(start)
		t.Logf("Listed %d files in %v", len(files), elapsed)
		
		if elapsed > 5*time.Second {
			t.Errorf("File listing took too long: %v", elapsed)
		}
		
		if len(files) == 0 {
			t.Errorf("Expected files to be listed")
		}
	})

	t.Run("pagination performance", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)
		repo.CreateManyFiles("large", 500)
		repo.AddCommit("Add 500 files")
		
		start := time.Now()
		
		// Test pagination with different limits
		limits := []int{10, 50, 100}
		for _, limit := range limits {
			files, err := ListFiles(repo.Path, "large", false, limit)
			if err != nil {
				t.Fatalf("Failed to list files with limit %d: %v", limit, err)
			}
			
			if len(files) > limit {
				t.Errorf("Expected at most %d files, got %d", limit, len(files))
			}
		}
		
		elapsed := time.Since(start)
		t.Logf("Pagination tests completed in %v", elapsed)
	})

	t.Run("search performance", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)
		
		// Create files with searchable content
		for i := 0; i < 100; i++ {
			filename := fmt.Sprintf("search_test_%d.txt", i)
			content := fmt.Sprintf("This is test file %d with database connection", i)
			if i%10 == 0 {
				content += " and postgres backend"
			}
			repo.WriteFile(filename, content)
		}
		repo.AddCommit("Add search test files")
		
		start := time.Now()
		
		// Test search performance
		results, err := SearchFiles(repo.Path, []string{"database"}, "and", false, 0, 50)
		if err != nil {
			t.Fatalf("Failed to search files: %v", err)
		}
		
		elapsed := time.Since(start)
		t.Logf("Found %d files in %v", len(results), elapsed)
		
		if elapsed > 10*time.Second {
			t.Errorf("Search took too long: %v", elapsed)
		}
	})
}

// TestMemoryUsage tests memory usage with large operations
func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	t.Run("large file content", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)
		
		// Create a large file (1MB)
		repo.CreateLargeFile("large_file.txt", 1024)
		repo.AddCommit("Add large file")
		
		// Test reading large file with limits
		content, err := GetFileContent(repo.Path, "large_file.txt", 100)
		if err != nil {
			t.Fatalf("Failed to read large file: %v", err)
		}
		
		// Should be limited to 100 lines
		lines := len(strings.Split(content, "\n"))
		if lines > 101 { // +1 for potential empty line at end
			t.Errorf("Expected at most 101 lines, got %d", lines)
		}
		
		t.Logf("Read %d lines from large file", lines)
	})

	t.Run("repository info with many commits", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)
		
		// Create many commits
		for i := 0; i < 50; i++ {
			filename := fmt.Sprintf("commit_%d.txt", i)
			repo.WriteFile(filename, fmt.Sprintf("Content for commit %d", i))
			repo.AddCommit(fmt.Sprintf("Commit %d", i))
		}
		
		start := time.Now()
		
		info, err := GetRepositoryInfo(repo.Path)
		if err != nil {
			t.Fatalf("Failed to get repository info: %v", err)
		}
		
		elapsed := time.Since(start)
		t.Logf("Got repository info for %d commits in %v", info.CommitCount, elapsed)
		
		if info.CommitCount <= 3 { // Should have more than initial commits
			t.Errorf("Expected more commits, got %d", info.CommitCount)
		}
		
		if elapsed > 5*time.Second {
			t.Errorf("Repository info took too long: %v", elapsed)
		}
	})
}

// TestConcurrentOperations tests concurrent access to repository operations
func TestConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	repo := CreateTestRepositoryWithContent(t)
	
	// Number of concurrent operations
	concurrency := 10
	
	t.Run("concurrent file listing", func(t *testing.T) {
		done := make(chan bool, concurrency)
		errors := make(chan error, concurrency)
		
		for i := 0; i < concurrency; i++ {
			go func(id int) {
				files, err := ListFiles(repo.Path, ".", false, 20)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d failed: %v", id, err)
					return
				}
				if len(files) == 0 {
					errors <- fmt.Errorf("goroutine %d got no files", id)
					return
				}
				done <- true
			}(i)
		}
		
		// Wait for all goroutines
		for i := 0; i < concurrency; i++ {
			select {
			case <-done:
				// Success
			case err := <-errors:
				t.Errorf("Concurrent operation failed: %v", err)
			case <-time.After(10 * time.Second):
				t.Fatalf("Concurrent operation timed out")
			}
		}
	})

	t.Run("concurrent repository info", func(t *testing.T) {
		done := make(chan bool, concurrency)
		errors := make(chan error, concurrency)
		
		for i := 0; i < concurrency; i++ {
			go func(id int) {
				info, err := GetRepositoryInfo(repo.Path)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d failed: %v", id, err)
					return
				}
				if info.CommitCount <= 0 {
					errors <- fmt.Errorf("goroutine %d got invalid commit count", id)
					return
				}
				done <- true
			}(i)
		}
		
		// Wait for all goroutines
		for i := 0; i < concurrency; i++ {
			select {
			case <-done:
				// Success
			case err := <-errors:
				t.Errorf("Concurrent operation failed: %v", err)
			case <-time.After(10 * time.Second):
				t.Fatalf("Concurrent operation timed out")
			}
		}
	})
}

// BenchmarkOperations provides benchmarks for core operations
func BenchmarkOperations(b *testing.B) {
	repo := CreateTestRepositoryWithContent(&testing.T{})
	
	b.Run("GetRepositoryInfo", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := GetRepositoryInfo(repo.Path)
			if err != nil {
				b.Fatalf("Benchmark failed: %v", err)
			}
		}
	})

	b.Run("ListBranches", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := ListBranches(repo.Path)
			if err != nil {
				b.Fatalf("Benchmark failed: %v", err)
			}
		}
	})

	b.Run("ListFiles", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := ListFiles(repo.Path, ".", false, 20)
			if err != nil {
				b.Fatalf("Benchmark failed: %v", err)
			}
		}
	})

	b.Run("SearchFiles", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := SearchFiles(repo.Path, []string{"database"}, "and", false, 0, 10)
			if err != nil {
				b.Fatalf("Benchmark failed: %v", err)
			}
		}
	})

	b.Run("GetFileContent", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := GetFileContent(repo.Path, "README.md", 50)
			if err != nil {
				b.Fatalf("Benchmark failed: %v", err)
			}
		}
	})
}

// BenchmarkMCPHandlers benchmarks MCP tool handlers
func BenchmarkMCPHandlers(b *testing.B) {
	repo := CreateTestRepositoryWithContent(&testing.T{})
	ctx := context.Background()
	
	b.Run("handleGetRepositoryInfo", func(b *testing.B) {
		params := GetRepositoryInfoParams{Repository: repo.Path}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := handleGetRepositoryInfo(ctx, nil, params)
			if err != nil {
				b.Fatalf("Benchmark failed: %v", err)
			}
		}
	})

	b.Run("handleListFiles", func(b *testing.B) {
		params := ListFilesParams{
			Repository:      repo.Path,
			Directory: ".",
			Recursive: false,
			Limit:     20,
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := handleListFiles(ctx, nil, params)
			if err != nil {
				b.Fatalf("Benchmark failed: %v", err)
			}
		}
	})

	b.Run("handleSearchFiles", func(b *testing.B) {
		params := SearchFilesParams{
			Repository:     repo.Path,
			Keywords: []string{"database"},
			Limit:    10,
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := handleSearchFiles(ctx, nil, params)
			if err != nil {
				b.Fatalf("Benchmark failed: %v", err)
			}
		}
	})
}

// TestResourceLimits tests behavior under resource constraints
func TestResourceLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resource limit test in short mode")
	}

	t.Run("very large file listing", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)
		
		// Create a very large number of files
		repo.CreateManyFiles("massive", 1000)
		repo.AddCommit("Add massive file set")
		
		// Test with different limits to ensure pagination works
		limits := []int{10, 100, 500}
		
		for _, limit := range limits {
			start := time.Now()
			files, err := ListFiles(repo.Path, "massive", false, limit)
			elapsed := time.Since(start)
			
			if err != nil {
				t.Errorf("Failed to list files with limit %d: %v", limit, err)
				continue
			}
			
			if len(files) > limit {
				t.Errorf("Limit %d exceeded: got %d files", limit, len(files))
			}
			
			t.Logf("Listed %d files with limit %d in %v", len(files), limit, elapsed)
			
			// Should complete within reasonable time even for large directories
			if elapsed > 30*time.Second {
				t.Errorf("Listing took too long with limit %d: %v", limit, elapsed)
			}
		}
	})

	t.Run("deep directory structure", func(t *testing.T) {
		repo := CreateTestRepositoryWithContent(t)
		
		// Create deep directory structure
		deepPath := "level1"
		for i := 2; i <= 10; i++ {
			deepPath = filepath.Join(deepPath, fmt.Sprintf("level%d", i))
		}
		
		repo.WriteFile(filepath.Join(deepPath, "deep_file.txt"), "Deep file content")
		repo.AddCommit("Add deep directory structure")
		
		// Test recursive listing
		start := time.Now()
		files, err := ListFiles(repo.Path, ".", true, 100)
		elapsed := time.Since(start)
		
		if err != nil {
			t.Errorf("Failed to list files recursively: %v", err)
		} else {
			t.Logf("Listed %d files recursively in %v", len(files), elapsed)
			
			// Should find the deep file
			found := false
			for _, file := range files {
				if strings.Contains(file.Path, "deep_file.txt") {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Deep file not found in recursive listing")
			}
		}
	})
}