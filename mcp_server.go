package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// McpCmd is the command for starting MCP server
var McpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server for Git remote operations",
	Long: `Start a Model Context Protocol (MCP) server that provides Git remote operations.

Supports repository information, branch management, file operations, and search capabilities.
Supports both stdio (default) and HTTP transports for maximum compatibility.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		transport, _ := cmd.Flags().GetString("transport")
		port, _ := cmd.Flags().GetInt("port")
		host, _ := cmd.Flags().GetString("host")
		workspace, _ := cmd.Flags().GetString("workspace")

		// For stdio mode, logs are automatically redirected to stderr
		// to avoid protocol contamination on stdout

		// Initialize workspace
		if workspace == "" {
			workspace = "./workspace" // Default workspace directory
		}
		if err := InitializeWorkspace(workspace); err != nil {
			return fmt.Errorf("failed to initialize workspace: %v", err)
		}

		// Create MCP server
		server := CreateMCPServer()

		// Handle different transport modes
		switch transport {
		case "stdio":
			// Use stdio transport (default) with minimal logging
			fmt.Fprintf(os.Stderr, "Starting Git Remote MCP server over stdio\n")
			return server.Run(context.Background(), &mcp.StdioTransport{})

		case "http":
			// Use HTTP transport
			handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
				return server
			}, nil)
			
			address := fmt.Sprintf("%s:%d", host, port)
			fmt.Printf("Starting Git Remote MCP server on %s\n", address)
			return http.ListenAndServe(address, handler)

		default:
			fmt.Fprintf(cmd.ErrOrStderr(), "Error: Unsupported transport: %s\n", transport)
			fmt.Fprintf(cmd.ErrOrStderr(), "Supported transports: stdio, http\n")
			return fmt.Errorf("unsupported transport: %s", transport)
		}
	},
}

// CreateMCPServer creates a new MCP server with all tools registered
func CreateMCPServer() *mcp.Server {
	// Create server with Git Remote implementation info and options
	opts := &mcp.ServerOptions{
		Instructions: "Use this Git Remote MCP server for repository operations!",
	}
	
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "git-remote-mcp",
		Version: "1.0.0",
	}, opts)

	// Register all Git tools
	RegisterGitTools(server)

	return server
}

func init() {
	McpCmd.Flags().String("transport", "stdio", "Transport protocol (stdio or http)")
	McpCmd.Flags().Int("port", 8080, "Port for HTTP transport (ignored for stdio)")
	McpCmd.Flags().String("host", "localhost", "Host address for HTTP transport (use 0.0.0.0 for all interfaces)")
	McpCmd.Flags().String("workspace", "./workspace", "Workspace directory for Git repositories")
}