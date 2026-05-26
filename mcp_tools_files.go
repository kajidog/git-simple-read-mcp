package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SearchFilesParams parameters for search_files tool
type SearchFilesParams struct {
	Repository      string   `json:"repository,omitempty"`       // Single repository (uses session default if empty)
	Repositories    []string `json:"repositories,omitempty"`     // Multiple repositories for cross-repo search
	Keywords        []string `json:"keywords"`
	SearchMode      string   `json:"search_mode,omitempty"`      // "and" or "or", defaults to "and"
	IncludeFilename bool     `json:"include_filename,omitempty"` // search in filenames too, defaults to false
	ContextLines    int      `json:"context_lines,omitempty"`    // number of context lines before/after match, 0=no context
	IncludePatterns []string `json:"include_patterns,omitempty"` // file patterns to include (glob)
	ExcludePatterns []string `json:"exclude_patterns,omitempty"` // file patterns to exclude (glob)
	Limit           int      `json:"limit,omitempty"`
}

// ListFilesParams parameters for list_files tool
type ListFilesParams struct {
	Repository      string   `json:"repository"`
	Directory       string   `json:"directory,omitempty"`
	Recursive       bool     `json:"recursive,omitempty"`
	IncludePatterns []string `json:"include_patterns,omitempty"` // file patterns to include (glob)
	ExcludePatterns []string `json:"exclude_patterns,omitempty"` // file patterns to exclude (glob)
	Limit           int      `json:"limit,omitempty"`
}

// GetFileContentParams parameters for get_file_content tool
type GetFileContentParams struct {
	Repository string   `json:"repository"`
	FilePath   string   `json:"file_path,omitempty"`  // Single file path
	FilePaths  []string `json:"file_paths,omitempty"` // Multiple file paths
	StartLine  int      `json:"start_line,omitempty"` // First line to read, 1-based (default: 1)
	EndLine    int      `json:"end_line,omitempty"`   // Last line to read, inclusive (default: start_line + 99)
}

// GetReadmeFilesParams parameters for get_readme_files tool
type GetReadmeFilesParams struct {
	Repository string `json:"repository"`
	Recursive  bool   `json:"recursive,omitempty"` // Search subdirectories
}

// RepoSearchResult result for cross-repository search (used by search_files)
type RepoSearchResult struct {
	Repository string         `json:"repository"`
	Results    []SearchResult `json:"results"`
	TotalCount int            `json:"total_count"`
	Error      string         `json:"error,omitempty"`
}

func handleSearchFiles(ctx context.Context, req *mcp.CallToolRequest, args SearchFilesParams) (*mcp.CallToolResult, any, error) {
	if len(args.Keywords) == 0 {
		return errorResult("Error: at least one keyword is required")
	}

	sc := GetSessionConfig()
	limit := sc.GetSearchLimit(args.Limit)

	searchMode := args.SearchMode
	if searchMode == "" {
		searchMode = "and"
	}

	includePatterns := sc.GetIncludePatterns(args.IncludePatterns)
	excludePatterns := sc.GetExcludePatterns(args.ExcludePatterns)

	if len(args.Repositories) > 0 {
		var allResults []RepoSearchResult
		for _, repoName := range args.Repositories {
			repoResult := RepoSearchResult{Repository: repoName}
			results, err := SearchFiles(repoName, args.Keywords, searchMode, args.IncludeFilename, args.ContextLines, includePatterns, excludePatterns, limit)
			if err != nil {
				repoResult.Error = err.Error()
			} else {
				repoResult.Results = results
				repoResult.TotalCount = len(results)
			}
			allResults = append(allResults, repoResult)
		}
		return textResult(formatMultiRepoSearchResults(allResults, args.Keywords, searchMode))
	}

	repository := sc.GetRepository(args.Repository)
	if repository == "" {
		return errorResult("Error: repository required (no default set)")
	}

	results, err := SearchFiles(repository, args.Keywords, searchMode, args.IncludeFilename, args.ContextLines, includePatterns, excludePatterns, limit)
	if err != nil {
		return errorResult("Search failed: %v", err)
	}

	return textResult(formatSearchResults(results, args.Keywords, searchMode))
}

func handleListFiles(ctx context.Context, req *mcp.CallToolRequest, args ListFilesParams) (*mcp.CallToolResult, any, error) {
	if args.Repository == "" {
		return errorResult("Error: repository path is required")
	}

	directory := args.Directory
	if directory == "" {
		directory = "."
	}

	limit := args.Limit
	if limit == 0 {
		limit = 50
	}

	files, err := ListFiles(args.Repository, directory, args.Recursive, args.IncludePatterns, args.ExcludePatterns, limit)
	if err != nil {
		return errorResult("Failed to list files: %v", err)
	}

	return textResult(formatFileList(files, directory, args.Recursive, limit))
}

func handleGetFileContent(ctx context.Context, req *mcp.CallToolRequest, args GetFileContentParams) (*mcp.CallToolResult, any, error) {
	if args.Repository == "" {
		return errorResult("Error: repository path is required")
	}

	var filePaths []string
	if len(args.FilePaths) > 0 {
		filePaths = args.FilePaths
	} else if args.FilePath != "" {
		filePaths = []string{args.FilePath}
	} else {
		return errorResult("Error: file path(s) required")
	}

	startLine := args.StartLine
	if startLine < 1 {
		startLine = 1
	}

	// Resolve the line span. Explicit end_line wins; otherwise fall back to
	// the session-configured default span (100 lines by default).
	var maxLines int
	if args.EndLine > 0 {
		if args.EndLine < startLine {
			return errorResult("Error: end_line (%d) must be >= start_line (%d)", args.EndLine, startLine)
		}
		maxLines = args.EndLine - startLine + 1
	} else {
		maxLines = GetSessionConfig().GetLineLimit(0)
	}

	showLineNumbers := true

	if len(filePaths) == 1 {
		content, totalLines, actualStart, actualEnd, err := GetFileContentWithLineNumbers(args.Repository, filePaths[0], startLine, maxLines, showLineNumbers)
		if err != nil {
			return errorResult("[%s ERR:%v]", filePaths[0], err)
		}
		return textResult(fmt.Sprintf("[%s L%d-%d/%d]\n%s", filePaths[0], actualStart, actualEnd, totalLines, content))
	}

	results, err := GetMultipleFileContentsWithLineNumbers(args.Repository, filePaths, startLine, maxLines, showLineNumbers)
	if err != nil {
		return errorResult("ERR:%v", err)
	}
	return textResult(formatMultipleFileContents(results))
}

func handleGetReadmeFiles(ctx context.Context, req *mcp.CallToolRequest, args GetReadmeFilesParams) (*mcp.CallToolResult, any, error) {
	if args.Repository == "" {
		return errorResult("Error: repository path is required")
	}

	readmeFiles, err := GetReadmeFiles(args.Repository, args.Recursive)
	if err != nil {
		return errorResult("Failed to find README files: %v", err)
	}

	return textResult(formatReadmeFiles(readmeFiles, args.Recursive))
}

func formatSearchResults(results []SearchResult, keywords []string, searchMode string) string {
	var result strings.Builder

	var keywordStr string
	if searchMode == "or" {
		keywordStr = strings.Join(keywords, " OR ")
	} else {
		keywordStr = strings.Join(keywords, " AND ")
	}
	result.WriteString(fmt.Sprintf("Search Results for: %s (%d files found)\n", keywordStr, len(results)))
	result.WriteString(strings.Repeat("-", 50) + "\n")

	if len(results) == 0 {
		result.WriteString("No files found matching the specified keywords.\n")
		return result.String()
	}

	for i, searchResult := range results {
		if i > 0 {
			result.WriteString("\n")
		}

		matchTypeStr := ""
		switch searchResult.MatchType {
		case "filename":
			matchTypeStr = " [filename match]"
		case "content":
			matchTypeStr = " [content match]"
		case "both":
			matchTypeStr = " [filename + content match]"
		}

		result.WriteString(fmt.Sprintf("📄 %s%s\n", searchResult.Path, matchTypeStr))

		if len(searchResult.Matches) > 0 {
			for _, match := range searchResult.Matches {
				if match.LineNumber == 0 {
					result.WriteString(fmt.Sprintf("   └─ Filename: %s\n", match.Content))
				} else {
					result.WriteString(fmt.Sprintf("   └─ Line %d: %s\n", match.LineNumber, strings.TrimSpace(match.Content)))
				}
			}
		}
	}

	return result.String()
}

func formatFileList(files []FileInfo, directory string, recursive bool, limit int) string {
	var result strings.Builder

	modeStr := "non-recursive"
	if recursive {
		modeStr = "recursive"
	}

	result.WriteString(fmt.Sprintf("Files in '%s' (%s, %d files):\n", directory, modeStr, len(files)))
	result.WriteString(strings.Repeat("-", 50) + "\n")

	for _, file := range files {
		infoStr := ""
		if file.Size > 0 || file.LineCount > 0 {
			var parts []string

			if file.Size > 0 {
				if file.Size < 1024 {
					parts = append(parts, fmt.Sprintf("%dB", file.Size))
				} else if file.Size < 1024*1024 {
					parts = append(parts, fmt.Sprintf("%.1fKB", float64(file.Size)/1024))
				} else {
					parts = append(parts, fmt.Sprintf("%.1fMB", float64(file.Size)/(1024*1024)))
				}
			}

			if file.LineCount > 0 {
				parts = append(parts, fmt.Sprintf("%dL", file.LineCount))
			}

			if len(parts) > 0 {
				infoStr = fmt.Sprintf(" (%s)", strings.Join(parts, ", "))
			}
		}
		result.WriteString(fmt.Sprintf("%s%s\n", file.Path, infoStr))
	}

	if len(files) == limit {
		result.WriteString(fmt.Sprintf("\n(Limited to %d results)", limit))
	}

	return result.String()
}

func formatMultipleFileContents(results []FileContentResult) string {
	var result strings.Builder

	for i, fileResult := range results {
		if i > 0 {
			result.WriteString("\n")
		}

		if fileResult.Error != "" {
			result.WriteString(fmt.Sprintf("[%s ERR:%s]\n", fileResult.FilePath, fileResult.Error))
		} else {
			result.WriteString(fmt.Sprintf("[%s L%d-%d/%d]\n", fileResult.FilePath, fileResult.StartLine, fileResult.EndLine, fileResult.TotalLines))
			result.WriteString(fileResult.Content)
		}
	}

	return result.String()
}

func formatReadmeFiles(readmeFiles []ReadmeFileInfo, recursive bool) string {
	var result strings.Builder

	searchMode := "root directory only"
	if recursive {
		searchMode = "recursive search"
	}

	result.WriteString(fmt.Sprintf("README Files (%s, %d found):\n", searchMode, len(readmeFiles)))
	result.WriteString(strings.Repeat("=", 50) + "\n\n")

	if len(readmeFiles) == 0 {
		result.WriteString("No README files found in the repository.\n")
		return result.String()
	}

	for _, readme := range readmeFiles {
		result.WriteString(fmt.Sprintf("📄 %s\n", readme.Path))
		if readme.Size > 0 {
			var sizeStr string
			if readme.Size < 1024 {
				sizeStr = fmt.Sprintf("%d bytes", readme.Size)
			} else if readme.Size < 1024*1024 {
				sizeStr = fmt.Sprintf("%.1f KB", float64(readme.Size)/1024)
			} else {
				sizeStr = fmt.Sprintf("%.1f MB", float64(readme.Size)/(1024*1024))
			}
			result.WriteString(fmt.Sprintf("   Size: %s", sizeStr))

			if readme.LineCount > 0 {
				result.WriteString(fmt.Sprintf(" | Lines: %d", readme.LineCount))
			}
			result.WriteString("\n")
		}
		if !readme.ModTime.IsZero() {
			result.WriteString(fmt.Sprintf("   Modified: %s\n", readme.ModTime.Format("2006-01-02 15:04:05")))
		}
		result.WriteString("\n")
	}

	return result.String()
}

func formatMultiRepoSearchResults(results []RepoSearchResult, keywords []string, searchMode string) string {
	var sb strings.Builder

	totalMatches := 0
	reposWithMatches := 0
	for _, r := range results {
		if r.TotalCount > 0 {
			totalMatches += r.TotalCount
			reposWithMatches++
		}
	}

	var keywordStr string
	if searchMode == "or" {
		keywordStr = strings.Join(keywords, " OR ")
	} else {
		keywordStr = strings.Join(keywords, " AND ")
	}

	sb.WriteString(fmt.Sprintf("Multi-Repository Search: %s\n", keywordStr))
	sb.WriteString(strings.Repeat("=", 50) + "\n")
	sb.WriteString(fmt.Sprintf("Found %d matches in %d/%d repositories\n\n", totalMatches, reposWithMatches, len(results)))

	for _, r := range results {
		if r.Error != "" {
			sb.WriteString(fmt.Sprintf("📁 %s: Error - %s\n\n", r.Repository, r.Error))
			continue
		}

		if r.TotalCount == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("📁 %s (%d matches)\n", r.Repository, r.TotalCount))

		for _, searchResult := range r.Results {
			sb.WriteString(fmt.Sprintf("  📄 %s\n", searchResult.Path))
			if len(searchResult.Matches) > 0 {
				for _, match := range searchResult.Matches {
					if match.LineNumber > 0 {
						sb.WriteString(fmt.Sprintf("     L%d: %s\n", match.LineNumber, strings.TrimSpace(match.Content)))
					}
				}
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
