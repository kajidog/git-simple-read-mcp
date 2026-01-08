package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Memo represents a document memo
type Memo struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MemoStore manages memo storage and operations
type MemoStore struct {
	mu       sync.RWMutex
	memos    map[string]*Memo
	filePath string
}

var globalMemoStore *MemoStore

// InitializeMemoStore initializes the global memo store
func InitializeMemoStore(workspaceDir string) error {
	if globalMemoStore != nil {
		return nil // Already initialized
	}

	filePath := filepath.Join(workspaceDir, "memos.json")
	store := &MemoStore{
		memos:    make(map[string]*Memo),
		filePath: filePath,
	}

	// Load existing memos from file
	if err := store.load(); err != nil {
		// If file doesn't exist, that's okay - we'll create it on first save
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to load memos: %v", err)
		}
	}

	globalMemoStore = store
	return nil
}

// GetMemoStore returns the global memo store instance
func GetMemoStore() *MemoStore {
	return globalMemoStore
}

// load reads memos from the JSON file
func (ms *MemoStore) load() error {
	data, err := os.ReadFile(ms.filePath)
	if err != nil {
		return err
	}

	var memos []*Memo
	if err := json.Unmarshal(data, &memos); err != nil {
		return fmt.Errorf("failed to unmarshal memos: %v", err)
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.memos = make(map[string]*Memo)
	for _, memo := range memos {
		ms.memos[memo.ID] = memo
	}

	return nil
}

// saveUnlocked writes memos to the JSON file (assumes lock is already held)
func (ms *MemoStore) saveUnlocked() error {
	memos := make([]*Memo, 0, len(ms.memos))
	for _, memo := range ms.memos {
		memos = append(memos, memo)
	}

	data, err := json.MarshalIndent(memos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal memos: %v", err)
	}

	if err := os.WriteFile(ms.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write memos file: %v", err)
	}

	return nil
}

// save writes memos to the JSON file (acquires lock)
func (ms *MemoStore) save() error {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.saveUnlocked()
}

// AddMemo adds a new memo
func (ms *MemoStore) AddMemo(title, content string, tags []string) (*Memo, error) {
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	memo := &Memo{
		ID:        uuid.New().String(),
		Title:     title,
		Content:   content,
		Tags:      tags,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ms.mu.Lock()
	ms.memos[memo.ID] = memo
	ms.mu.Unlock()

	if err := ms.save(); err != nil {
		return nil, err
	}

	return memo, nil
}

// GetMemo retrieves a memo by ID
func (ms *MemoStore) GetMemo(id string) (*Memo, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	memo, exists := ms.memos[id]
	if !exists {
		return nil, fmt.Errorf("memo not found: %s", id)
	}

	return memo, nil
}

// UpdateMemo updates an existing memo
func (ms *MemoStore) UpdateMemo(id, title, content string, tags []string) (*Memo, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	memo, exists := ms.memos[id]
	if !exists {
		return nil, fmt.Errorf("memo not found: %s", id)
	}

	if title != "" {
		memo.Title = title
	}
	if content != "" {
		memo.Content = content
	}
	if tags != nil {
		memo.Tags = tags
	}
	memo.UpdatedAt = time.Now()

	if err := ms.saveUnlocked(); err != nil {
		return nil, err
	}

	return memo, nil
}

// DeleteMemo deletes a memo by ID
func (ms *MemoStore) DeleteMemo(id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, exists := ms.memos[id]; !exists {
		return fmt.Errorf("memo not found: %s", id)
	}

	delete(ms.memos, id)

	if err := ms.saveUnlocked(); err != nil {
		return err
	}

	return nil
}

// DeleteAllMemos deletes all memos
func (ms *MemoStore) DeleteAllMemos() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.memos = make(map[string]*Memo)

	if err := ms.saveUnlocked(); err != nil {
		return err
	}

	return nil
}

// SearchMemos searches for memos matching the criteria
func (ms *MemoStore) SearchMemos(query string, tags []string, limit int) []*Memo {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var results []*Memo
	queryLower := strings.ToLower(query)

	for _, memo := range ms.memos {
		// Check if memo matches search criteria
		matches := false

		// If query is empty, match all (useful for listing all memos)
		if query == "" {
			matches = true
		} else {
			// Search in title and content
			if strings.Contains(strings.ToLower(memo.Title), queryLower) ||
				strings.Contains(strings.ToLower(memo.Content), queryLower) {
				matches = true
			}
		}

		// Filter by tags if specified
		if len(tags) > 0 {
			tagMatch := false
			for _, searchTag := range tags {
				for _, memoTag := range memo.Tags {
					if strings.EqualFold(memoTag, searchTag) {
						tagMatch = true
						break
					}
				}
				if tagMatch {
					break
				}
			}
			if !tagMatch {
				matches = false
			}
		}

		if matches {
			results = append(results, memo)
		}

		// Apply limit
		if limit > 0 && len(results) >= limit {
			break
		}
	}

	return results
}

// ListAllMemos returns all memos
func (ms *MemoStore) ListAllMemos() []*Memo {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	memos := make([]*Memo, 0, len(ms.memos))
	for _, memo := range ms.memos {
		memos = append(memos, memo)
	}

	return memos
}

// Count returns the total number of memos
func (ms *MemoStore) Count() int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return len(ms.memos)
}
