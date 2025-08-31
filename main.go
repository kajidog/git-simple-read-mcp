package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "git-remote-mcp",
		Short: "MCP server for Git remote operations",
		Long: `A Model Context Protocol (MCP) server that provides Git remote operations.
Supports repository information, branch management, file operations, and search capabilities.`,
	}

	// Add MCP server command
	rootCmd.AddCommand(McpCmd)

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}