package main

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestMemoStore(t *testing.T) (*MemoStore, string) {
	// Reset global memo store
	globalMemoStore = nil

	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "memo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Initialize memo store
	if err := InitializeMemoStore(tmpDir); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to initialize memo store: %v", err)
	}

	return GetMemoStore(), tmpDir
}

func cleanupTestMemoStore(tmpDir string) {
	os.RemoveAll(tmpDir)
}

func TestAddMemo(t *testing.T) {
	store, tmpDir := setupTestMemoStore(t)
	defer cleanupTestMemoStore(tmpDir)

	memo, err := store.AddMemo("Test Title", "Test Content", []string{"tag1", "tag2"})
	if err != nil {
		t.Fatalf("Failed to add memo: %v", err)
	}

	if memo.ID == "" {
		t.Error("Memo ID should not be empty")
	}
	if memo.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", memo.Title)
	}
	if memo.Content != "Test Content" {
		t.Errorf("Expected content 'Test Content', got '%s'", memo.Content)
	}
	if len(memo.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(memo.Tags))
	}
	if memo.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if memo.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestAddMemoWithoutTitle(t *testing.T) {
	store, tmpDir := setupTestMemoStore(t)
	defer cleanupTestMemoStore(tmpDir)

	_, err := store.AddMemo("", "Test Content", nil)
	if err == nil {
		t.Error("Expected error when adding memo without title")
	}
}

func TestGetMemo(t *testing.T) {
	store, tmpDir := setupTestMemoStore(t)
	defer cleanupTestMemoStore(tmpDir)

	// Add a memo
	original, err := store.AddMemo("Test Title", "Test Content", []string{"tag1"})
	if err != nil {
		t.Fatalf("Failed to add memo: %v", err)
	}

	// Get the memo
	retrieved, err := store.GetMemo(original.ID)
	if err != nil {
		t.Fatalf("Failed to get memo: %v", err)
	}

	if retrieved.ID != original.ID {
		t.Errorf("Expected ID '%s', got '%s'", original.ID, retrieved.ID)
	}
	if retrieved.Title != original.Title {
		t.Errorf("Expected title '%s', got '%s'", original.Title, retrieved.Title)
	}
	if retrieved.Content != original.Content {
		t.Errorf("Expected content '%s', got '%s'", original.Content, retrieved.Content)
	}
}

func TestGetNonExistentMemo(t *testing.T) {
	store, tmpDir := setupTestMemoStore(t)
	defer cleanupTestMemoStore(tmpDir)

	_, err := store.GetMemo("non-existent-id")
	if err == nil {
		t.Error("Expected error when getting non-existent memo")
	}
}

func TestUpdateMemo(t *testing.T) {
	store, tmpDir := setupTestMemoStore(t)
	defer cleanupTestMemoStore(tmpDir)

	// Add a memo
	original, err := store.AddMemo("Original Title", "Original Content", []string{"tag1"})
	if err != nil {
		t.Fatalf("Failed to add memo: %v", err)
	}

	// Update the memo
	updated, err := store.UpdateMemo(original.ID, "Updated Title", "Updated Content", []string{"tag2", "tag3"})
	if err != nil {
		t.Fatalf("Failed to update memo: %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got '%s'", updated.Title)
	}
	if updated.Content != "Updated Content" {
		t.Errorf("Expected content 'Updated Content', got '%s'", updated.Content)
	}
	if len(updated.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(updated.Tags))
	}
	if updated.UpdatedAt.Before(original.UpdatedAt) {
		t.Error("UpdatedAt should not be before original UpdatedAt")
	}
}

func TestDeleteMemo(t *testing.T) {
	store, tmpDir := setupTestMemoStore(t)
	defer cleanupTestMemoStore(tmpDir)

	// Add a memo
	memo, err := store.AddMemo("Test Title", "Test Content", nil)
	if err != nil {
		t.Fatalf("Failed to add memo: %v", err)
	}

	// Delete the memo
	if err := store.DeleteMemo(memo.ID); err != nil {
		t.Fatalf("Failed to delete memo: %v", err)
	}

	// Try to get the deleted memo
	_, err = store.GetMemo(memo.ID)
	if err == nil {
		t.Error("Expected error when getting deleted memo")
	}
}

func TestDeleteNonExistentMemo(t *testing.T) {
	store, tmpDir := setupTestMemoStore(t)
	defer cleanupTestMemoStore(tmpDir)

	err := store.DeleteMemo("non-existent-id")
	if err == nil {
		t.Error("Expected error when deleting non-existent memo")
	}
}

func TestDeleteAllMemos(t *testing.T) {
	store, tmpDir := setupTestMemoStore(t)
	defer cleanupTestMemoStore(tmpDir)

	// Add multiple memos
	store.AddMemo("Title 1", "Content 1", nil)
	store.AddMemo("Title 2", "Content 2", nil)
	store.AddMemo("Title 3", "Content 3", nil)

	if store.Count() != 3 {
		t.Errorf("Expected 3 memos, got %d", store.Count())
	}

	// Delete all memos
	if err := store.DeleteAllMemos(); err != nil {
		t.Fatalf("Failed to delete all memos: %v", err)
	}

	if store.Count() != 0 {
		t.Errorf("Expected 0 memos after delete all, got %d", store.Count())
	}
}

func TestSearchMemos(t *testing.T) {
	store, tmpDir := setupTestMemoStore(t)
	defer cleanupTestMemoStore(tmpDir)

	// Add test memos
	store.AddMemo("Go Programming", "Learning Go language", []string{"go", "programming"})
	store.AddMemo("Python Basics", "Python tutorial", []string{"python", "programming"})
	store.AddMemo("Web Development", "Building web apps with Go", []string{"go", "web"})

	// Search by query
	results := store.SearchMemos("Go", nil, 10)
	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'Go', got %d", len(results))
	}

	// Search by tag
	results = store.SearchMemos("", []string{"programming"}, 10)
	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'programming' tag, got %d", len(results))
	}

	// Search by query and tag
	results = store.SearchMemos("Go", []string{"web"}, 10)
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'Go' with 'web' tag, got %d", len(results))
	}

	// Search with limit
	results = store.SearchMemos("", nil, 2)
	if len(results) != 2 {
		t.Errorf("Expected 2 results with limit=2, got %d", len(results))
	}
}

func TestListAllMemos(t *testing.T) {
	store, tmpDir := setupTestMemoStore(t)
	defer cleanupTestMemoStore(tmpDir)

	// Add test memos
	store.AddMemo("Title 1", "Content 1", nil)
	store.AddMemo("Title 2", "Content 2", nil)
	store.AddMemo("Title 3", "Content 3", nil)

	memos := store.ListAllMemos()
	if len(memos) != 3 {
		t.Errorf("Expected 3 memos, got %d", len(memos))
	}
}

func TestMemoPersistence(t *testing.T) {
	// Reset global store
	globalMemoStore = nil

	tmpDir, err := os.MkdirTemp("", "memo-persist-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize first store and add memo
	if err := InitializeMemoStore(tmpDir); err != nil {
		t.Fatalf("Failed to initialize memo store: %v", err)
	}
	store1 := GetMemoStore()
	memo, err := store1.AddMemo("Persistent Memo", "This should persist", []string{"test"})
	if err != nil {
		t.Fatalf("Failed to add memo: %v", err)
	}

	// Reset global store
	globalMemoStore = nil

	// Initialize second store (should load from file)
	if err := InitializeMemoStore(tmpDir); err != nil {
		t.Fatalf("Failed to reinitialize memo store: %v", err)
	}
	store2 := GetMemoStore()

	// Verify memo was loaded
	retrieved, err := store2.GetMemo(memo.ID)
	if err != nil {
		t.Fatalf("Failed to get memo from reloaded store: %v", err)
	}

	if retrieved.Title != "Persistent Memo" {
		t.Errorf("Expected title 'Persistent Memo', got '%s'", retrieved.Title)
	}
	if retrieved.Content != "This should persist" {
		t.Errorf("Expected content 'This should persist', got '%s'", retrieved.Content)
	}

	// Verify file exists
	filePath := filepath.Join(tmpDir, "memos.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Memos file should exist")
	}
}

func TestConcurrentAccess(t *testing.T) {
	store, tmpDir := setupTestMemoStore(t)
	defer cleanupTestMemoStore(tmpDir)

	// Add memos concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(index int) {
			store.AddMemo("Title", "Content", nil)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	if store.Count() != 10 {
		t.Errorf("Expected 10 memos, got %d", store.Count())
	}
}
