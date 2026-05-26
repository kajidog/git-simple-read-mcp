package main

import (
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// errorResult builds an MCP CallToolResult with IsError=true and the given
// formatted message. Returned as the standard three-value tuple expected by
// mcp.AddTool handlers so callers can simply `return errorResult(...)`.
func errorResult(format string, args ...any) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, args...)}},
		IsError: true,
	}, nil, nil
}

// textResult builds a non-error CallToolResult with the given text content.
func textResult(text string) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}, nil, nil
}
