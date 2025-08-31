package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "git-simple-read-mcp",
		Short: "MCP server for Git read operations",
		Long: `A Model Context Protocol (MCP) server that provides Git read operations.
Supports repository information, branch listing, file operations, and search capabilities.`,
	}

	// Add MCP server command
	rootCmd.AddCommand(McpCmd)

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
