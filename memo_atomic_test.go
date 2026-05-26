package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestMemoSaveAtomicity verifies that the on-disk memos.json contains the
// fully-encoded final state and not a partial write that would leave invalid
// JSON.
func TestMemoSaveAtomicity(t *testing.T) {
	workspaceDir := t.TempDir()
	resetMemoStoreForTest()
	if err := InitializeMemoStore(workspaceDir); err != nil {
		t.Fatalf("InitializeMemoStore: %v", err)
	}
	store := GetMemoStore()

	for i := 0; i < 5; i++ {
		if _, err := store.AddMemo("", "title", "content", nil); err != nil {
			t.Fatalf("AddMemo: %v", err)
		}
	}

	data, err := os.ReadFile(filepath.Join(workspaceDir, "memos.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var memos []*Memo
	if err := json.Unmarshal(data, &memos); err != nil {
		t.Fatalf("memos.json is not valid JSON: %v\ncontent: %s", err, string(data))
	}
	if len(memos) != 5 {
		t.Fatalf("expected 5 memos, got %d", len(memos))
	}
}

// TestMemoSaveLeavesNoTempArtifacts verifies that successful writes do not
// leave temporary files behind in the workspace directory.
func TestMemoSaveLeavesNoTempArtifacts(t *testing.T) {
	workspaceDir := t.TempDir()
	resetMemoStoreForTest()
	if err := InitializeMemoStore(workspaceDir); err != nil {
		t.Fatalf("InitializeMemoStore: %v", err)
	}
	store := GetMemoStore()

	for i := 0; i < 10; i++ {
		if _, err := store.AddMemo("", "title", "content", nil); err != nil {
			t.Fatalf("AddMemo: %v", err)
		}
	}

	entries, err := os.ReadDir(workspaceDir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".memos-") && strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("temp artifact left behind: %s", e.Name())
		}
	}
}

// TestMemoSaveRollbackOnError verifies that if a save fails mid-flight, the
// in-memory state is rolled back so it stays consistent with disk.
func TestMemoSaveRollbackOnError(t *testing.T) {
	workspaceDir := t.TempDir()
	resetMemoStoreForTest()
	if err := InitializeMemoStore(workspaceDir); err != nil {
		t.Fatalf("InitializeMemoStore: %v", err)
	}
	store := GetMemoStore()
	if _, err := store.AddMemo("", "first", "content", nil); err != nil {
		t.Fatalf("AddMemo: %v", err)
	}
	before := store.Count()

	// Force subsequent saves to fail by pointing filePath at a directory
	// that does not exist. os.CreateTemp in that directory will fail.
	store.filePath = filepath.Join(workspaceDir, "no-such-dir", "memos.json")

	if _, err := store.AddMemo("", "second", "content", nil); err == nil {
		t.Fatal("expected AddMemo to fail when save target is unwritable")
	}

	if store.Count() != before {
		t.Fatalf("expected in-memory count to roll back to %d, got %d", before, store.Count())
	}
}

// TestMemoConcurrentAddDoesNotCorrupt verifies that concurrent AddMemo calls
// do not produce corrupt JSON files. The exact number of memos is not what
// we care about; we care that the file always parses cleanly.
func TestMemoConcurrentAddDoesNotCorrupt(t *testing.T) {
	workspaceDir := t.TempDir()
	resetMemoStoreForTest()
	if err := InitializeMemoStore(workspaceDir); err != nil {
		t.Fatalf("InitializeMemoStore: %v", err)
	}
	store := GetMemoStore()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = store.AddMemo("", "t", "c", nil)
		}()
	}
	wg.Wait()

	data, err := os.ReadFile(filepath.Join(workspaceDir, "memos.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var memos []*Memo
	if err := json.Unmarshal(data, &memos); err != nil {
		t.Fatalf("memos.json is not valid JSON after concurrent writes: %v", err)
	}
	if len(memos) != 20 {
		t.Fatalf("expected 20 memos, got %d", len(memos))
	}
}

// resetMemoStoreForTest clears the global memo store so each test can
// initialize a fresh one. This is necessary because InitializeMemoStore
// is a no-op when the global is already set.
func resetMemoStoreForTest() {
	globalMemoStore = nil
}
