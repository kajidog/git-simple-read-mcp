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
	mcp.AddTool(server, &mcp.Tool{
		Name: "add_memo",
		Description: "Create a persistent memo. title is required; repository (optional) associates " +
			"the memo with a workspace repo so it shows up in get_repository_info. " +
			"Returns the generated UUID, which other memo tools use to address it.",
	}, handleAddMemo)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_memo",
		Description: "Read one memo by ID. Use list_memos to discover IDs.",
	}, handleGetMemo)

	mcp.AddTool(server, &mcp.Tool{
		Name: "update_memo",
		Description: "Modify an existing memo by ID. Only non-empty fields are applied; " +
			"omitted fields are left unchanged. UpdatedAt is refreshed automatically.",
	}, handleUpdateMemo)

	mcp.AddTool(server, &mcp.Tool{
		Name: "delete_memo",
		Description: "Delete one memo by ID. Destructive: there is no undo. " +
			"Use list_memos to confirm the ID first.",
	}, handleDeleteMemo)

	mcp.AddTool(server, &mcp.Tool{
		Name: "list_memos",
		Description: "Search and list memos. query searches title and content (case-insensitive); " +
			"repository filters to one repo; tags filters by tag (any match). " +
			"limit caps results (default 50).",
	}, handleListMemos)
}

func handleAddMemo(ctx context.Context, req *mcp.CallToolRequest, args AddMemoParams) (*mcp.CallToolResult, any, error) {
	store := GetMemoStore()
	if store == nil {
		return errorResult("Error: memo store not initialized")
	}

	memo, err := store.AddMemo(args.Repository, args.Title, args.Content, args.Tags)
	if err != nil {
		return errorResult("Failed to add memo: %v", err)
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

	return textResult(result.String())
}

func handleGetMemo(ctx context.Context, req *mcp.CallToolRequest, args GetMemoParams) (*mcp.CallToolResult, any, error) {
	store := GetMemoStore()
	if store == nil {
		return errorResult("Error: memo store not initialized")
	}

	memo, err := store.GetMemo(args.ID)
	if err != nil {
		return errorResult("Failed to get memo: %v", err)
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

	return textResult(result.String())
}

func handleUpdateMemo(ctx context.Context, req *mcp.CallToolRequest, args UpdateMemoParams) (*mcp.CallToolResult, any, error) {
	store := GetMemoStore()
	if store == nil {
		return errorResult("Error: memo store not initialized")
	}

	memo, err := store.UpdateMemo(args.ID, args.Repository, args.Title, args.Content, args.Tags)
	if err != nil {
		return errorResult("Failed to update memo: %v", err)
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

	return textResult(result.String())
}

func handleDeleteMemo(ctx context.Context, req *mcp.CallToolRequest, args DeleteMemoParams) (*mcp.CallToolResult, any, error) {
	store := GetMemoStore()
	if store == nil {
		return errorResult("Error: memo store not initialized")
	}

	if err := store.DeleteMemo(args.ID); err != nil {
		return errorResult("Failed to delete memo: %v", err)
	}

	return textResult(fmt.Sprintf("Memo deleted successfully: %s", args.ID))
}

func handleListMemos(ctx context.Context, req *mcp.CallToolRequest, args ListMemosParams) (*mcp.CallToolResult, any, error) {
	store := GetMemoStore()
	if store == nil {
		return errorResult("Error: memo store not initialized")
	}

	limit := args.Limit
	if limit == 0 {
		limit = 50
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

	return textResult(result.String())
}
