package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AddMemoParams parameters for add_memo tool
type AddMemoParams struct {
	Repository string   `json:"repository,omitempty"` // Associated repository name
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Tags       []string `json:"tags,omitempty"`
}

// GetMemoParams parameters for get_memo tool
type GetMemoParams struct {
	ID string `json:"id"`
}

// UpdateMemoParams parameters for update_memo tool
type UpdateMemoParams struct {
	ID         string   `json:"id"`
	Repository string   `json:"repository,omitempty"` // Change associated repository
	Title      string   `json:"title,omitempty"`
	Content    string   `json:"content,omitempty"`
	Tags       []string `json:"tags,omitempty"`
}

// DeleteMemoParams parameters for delete_memo tool
type DeleteMemoParams struct {
	ID string `json:"id"`
}

// ListMemosParams parameters for list_memos tool
type ListMemosParams struct {
	Repository string   `json:"repository,omitempty"` // Filter by repository name
	Query      string   `json:"query,omitempty"`      // Search query for title/content
	Tags       []string `json:"tags,omitempty"`       // Filter by tags
	Limit      int      `json:"limit,omitempty"`      // Maximum number of results (default: 50)
}

// RegisterMemoTools registers all memo-related MCP tools
func RegisterMemoTools(server *mcp.Server) {
	// Add memo tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_memo",
		Description: "Add a new document memo. Options: repository (associate with repo), title (required), content, tags",
	}, handleAddMemo)

	// Get memo tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_memo",
		Description: "Get a specific memo by ID",
	}, handleGetMemo)

	// Update memo tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_memo",
		Description: "Update an existing memo. Options: id (required), repository, title, content, tags",
	}, handleUpdateMemo)

	// Delete memo tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_memo",
		Description: "Delete a memo by ID",
	}, handleDeleteMemo)

	// List/search memos tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_memos",
		Description: "List or search memos. Options: repository (filter by repo), query (search in title/content), tags (filter by tags), limit (max results)",
	}, handleListMemos)

	// Delete all memos tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_all_memos",
		Description: "Delete all memos (use with caution)",
	}, handleDeleteAllMemos)
}

func handleAddMemo(ctx context.Context, req *mcp.CallToolRequest, args AddMemoParams) (*mcp.CallToolResult, any, error) {
	store := GetMemoStore()
	if store == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: memo store not initialized"}},
			IsError: true,
		}, nil, nil
	}

	memo, err := store.AddMemo(args.Repository, args.Title, args.Content, args.Tags)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to add memo: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	var result strings.Builder
	result.WriteString("Memo added successfully\n\n")
	result.WriteString(fmt.Sprintf("ID: %s\n", memo.ID))
	if memo.Repository != "" {
		result.WriteString(fmt.Sprintf("Repository: %s\n", memo.Repository))
	}
	result.WriteString(fmt.Sprintf("Title: %s\n", memo.Title))
	result.WriteString(fmt.Sprintf("Created: %s\n", memo.CreatedAt.Format("2006-01-02 15:04:05")))
	if len(memo.Tags) > 0 {
		result.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(memo.Tags, ", ")))
	}
	result.WriteString(fmt.Sprintf("\nContent:\n%s\n", memo.Content))

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
		IsError: false,
	}, nil, nil
}

func handleGetMemo(ctx context.Context, req *mcp.CallToolRequest, args GetMemoParams) (*mcp.CallToolResult, any, error) {
	store := GetMemoStore()
	if store == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: memo store not initialized"}},
			IsError: true,
		}, nil, nil
	}

	memo, err := store.GetMemo(args.ID)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to get memo: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("ID: %s\n", memo.ID))
	if memo.Repository != "" {
		result.WriteString(fmt.Sprintf("Repository: %s\n", memo.Repository))
	}
	result.WriteString(fmt.Sprintf("Title: %s\n", memo.Title))
	result.WriteString(strings.Repeat("=", 50) + "\n\n")
	if len(memo.Tags) > 0 {
		result.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(memo.Tags, ", ")))
	}
	result.WriteString(fmt.Sprintf("Created: %s\n", memo.CreatedAt.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("Updated: %s\n", memo.UpdatedAt.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("\nContent:\n%s\n", memo.Content))

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
		IsError: false,
	}, nil, nil
}

func handleUpdateMemo(ctx context.Context, req *mcp.CallToolRequest, args UpdateMemoParams) (*mcp.CallToolResult, any, error) {
	store := GetMemoStore()
	if store == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: memo store not initialized"}},
			IsError: true,
		}, nil, nil
	}

	memo, err := store.UpdateMemo(args.ID, args.Repository, args.Title, args.Content, args.Tags)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to update memo: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	var result strings.Builder
	result.WriteString("Memo updated successfully\n\n")
	result.WriteString(fmt.Sprintf("ID: %s\n", memo.ID))
	if memo.Repository != "" {
		result.WriteString(fmt.Sprintf("Repository: %s\n", memo.Repository))
	}
	result.WriteString(fmt.Sprintf("Title: %s\n", memo.Title))
	result.WriteString(fmt.Sprintf("Updated: %s\n", memo.UpdatedAt.Format("2006-01-02 15:04:05")))
	if len(memo.Tags) > 0 {
		result.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(memo.Tags, ", ")))
	}
	result.WriteString(fmt.Sprintf("\nContent:\n%s\n", memo.Content))

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
		IsError: false,
	}, nil, nil
}

func handleDeleteMemo(ctx context.Context, req *mcp.CallToolRequest, args DeleteMemoParams) (*mcp.CallToolResult, any, error) {
	store := GetMemoStore()
	if store == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: memo store not initialized"}},
			IsError: true,
		}, nil, nil
	}

	if err := store.DeleteMemo(args.ID); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to delete memo: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Memo deleted successfully: %s", args.ID)}},
		IsError: false,
	}, nil, nil
}

func handleListMemos(ctx context.Context, req *mcp.CallToolRequest, args ListMemosParams) (*mcp.CallToolResult, any, error) {
	store := GetMemoStore()
	if store == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: memo store not initialized"}},
			IsError: true,
		}, nil, nil
	}

	limit := args.Limit
	if limit == 0 {
		limit = 50 // Default limit
	}

	memos := store.SearchMemos(args.Query, args.Repository, args.Tags, limit)

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d memo(s)", len(memos)))
	if args.Repository != "" {
		result.WriteString(fmt.Sprintf(" for repository: %s", args.Repository))
	}
	result.WriteString("\n")
	result.WriteString(strings.Repeat("=", 50) + "\n\n")

	for i, memo := range memos {
		// Show full ID for AI usability (was truncated to 8 chars before)
		result.WriteString(fmt.Sprintf("%d. %s\n", i+1, memo.Title))
		result.WriteString(fmt.Sprintf("   ID: %s\n", memo.ID))
		if memo.Repository != "" {
			result.WriteString(fmt.Sprintf("   Repository: %s\n", memo.Repository))
		}
		if len(memo.Tags) > 0 {
			result.WriteString(fmt.Sprintf("   Tags: %s\n", strings.Join(memo.Tags, ", ")))
		}
		result.WriteString(fmt.Sprintf("   Created: %s | Updated: %s\n",
			memo.CreatedAt.Format("2006-01-02 15:04"),
			memo.UpdatedAt.Format("2006-01-02 15:04")))

		// Show content preview (first 100 characters)
		contentPreview := memo.Content
		if len(contentPreview) > 100 {
			contentPreview = contentPreview[:100] + "..."
		}
		result.WriteString(fmt.Sprintf("   Preview: %s\n", strings.ReplaceAll(contentPreview, "\n", " ")))
		result.WriteString("\n")
	}

	if len(memos) == 0 {
		result.WriteString("No memos found")
		if args.Query != "" || len(args.Tags) > 0 || args.Repository != "" {
			result.WriteString(" matching the search criteria")
		}
		result.WriteString(".\n")
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: result.String()}},
		IsError: false,
	}, nil, nil
}

func handleDeleteAllMemos(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
	store := GetMemoStore()
	if store == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: memo store not initialized"}},
			IsError: true,
		}, nil, nil
	}

	count := store.Count()
	if err := store.DeleteAllMemos(); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to delete all memos: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("All memos deleted successfully (%d memos removed)", count)}},
		IsError: false,
	}, nil, nil
}
