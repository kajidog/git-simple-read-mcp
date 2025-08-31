package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// SearchFilesEnhanced searches for files with enhanced features
func SearchFilesEnhanced(repoPath string, keywords []string, searchMode string, includeFilename bool, contextLines int, maxResults int) ([]SearchResult, error) {
	// Validate workspace path
	validPath, err := ValidateWorkspacePath(repoPath)
	if err != nil {
		return nil, err
	}
	repoPath = validPath

	if !isGitRepository(repoPath) {
		return nil, fmt.Errorf("not a git repository: %s", repoPath)
	}

	if len(keywords) == 0 {
		return []SearchResult{}, nil
	}

	var allResults []SearchResult
	
	// Search in file contents
	contentResults, err := searchInContent(repoPath, keywords, searchMode, contextLines)
	if err == nil {
		allResults = append(allResults, contentResults...)
	}
	
	// Search in filenames if requested
	if includeFilename {
		filenameResults, err := searchInFilenames(repoPath, keywords, searchMode)
		if err == nil {
			allResults = append(allResults, filenameResults...)
		}
	}

	// Remove duplicates and apply limit
	uniqueResults := removeDuplicateResults(allResults)
	if maxResults > 0 && len(uniqueResults) > maxResults {
		uniqueResults = uniqueResults[:maxResults]
	}

	return uniqueResults, nil
}

// searchInContent searches for keywords in file contents
func searchInContent(repoPath string, keywords []string, searchMode string, contextLines int) ([]SearchResult, error) {
	if len(keywords) == 0 {
		return []SearchResult{}, nil
	}

	// Build git grep command
	args := []string{"grep"}
	
	// Add context lines if requested
	if contextLines > 0 {
		args = append(args, "-C", strconv.Itoa(contextLines))
	}
	
	// Add line numbers
	args = append(args, "-n")
	
	var results []SearchResult
	
	if searchMode == "or" {
		// For OR mode, search each keyword separately and merge results
		for _, keyword := range keywords {
			keywordArgs := append(args, keyword)
			cmd := exec.Command("git", keywordArgs...)
			cmd.Dir = repoPath
			output, err := cmd.Output()
			
			if err == nil {
				keywordResults := parseGrepOutput(string(output), "content", contextLines > 0)
				results = append(results, keywordResults...)
			}
		}
	} else {
		// For AND mode, use multiple grep commands piped together
		if len(keywords) == 1 {
			// Single keyword
			args = append(args, keywords[0])
			cmd := exec.Command("git", args...)
			cmd.Dir = repoPath
			
			output, err := cmd.Output()
			
			if err == nil {
				results = parseGrepOutput(string(output), "content", contextLines > 0)
			}
		} else {
			// Multiple keywords - implement AND logic by filtering results
			firstKeywordArgs := append(args, keywords[0])
			cmd := exec.Command("git", firstKeywordArgs...)
			cmd.Dir = repoPath
			output, err := cmd.Output()
			
			if err == nil {
				results = parseGrepOutput(string(output), "content", contextLines > 0)
				// Filter results to only include files that contain all keywords
				results = filterResultsByAllKeywords(repoPath, results, keywords[1:])
			}
		}
	}

	return removeDuplicateResults(results), nil
}

// searchInFilenames searches for keywords in filenames
func searchInFilenames(repoPath string, keywords []string, searchMode string) ([]SearchResult, error) {
	// Use git ls-files to get all tracked files
	cmd := exec.Command("git", "ls-files")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %v", err)
	}

	var results []SearchResult
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	
	for scanner.Scan() {
		filePath := strings.TrimSpace(scanner.Text())
		if filePath == "" {
			continue
		}
		
		filename := filepath.Base(filePath)
		matches := false
		
		if searchMode == "or" {
			// OR mode: file matches if any keyword is in filename
			for _, keyword := range keywords {
				if strings.Contains(strings.ToLower(filename), strings.ToLower(keyword)) {
					matches = true
					break
				}
			}
		} else {
			// AND mode: file matches if all keywords are in filename
			matches = true
			for _, keyword := range keywords {
				if !strings.Contains(strings.ToLower(filename), strings.ToLower(keyword)) {
					matches = false
					break
				}
			}
		}
		
		if matches {
			results = append(results, SearchResult{
				Path:      filePath,
				MatchType: "filename",
				Matches: []MatchLine{{
					LineNumber: 0,
					Content:    filename,
				}},
			})
		}
	}

	return results, nil
}

// parseGrepOutput parses git grep output into SearchResult structs
func parseGrepOutput(output string, matchType string, hasContext bool) []SearchResult {
	if output == "" {
		return []SearchResult{}
	}

	lines := strings.Split(output, "\n")
	resultMap := make(map[string]*SearchResult)
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Parse git grep output format: filename:linenumber:content
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}
		
		filePath := parts[0]
		lineNumStr := parts[1]
		content := parts[2]
		
		lineNum, err := strconv.Atoi(lineNumStr)
		if err != nil {
			continue
		}
		
		// Get or create result for this file
		if resultMap[filePath] == nil {
			resultMap[filePath] = &SearchResult{
				Path:      filePath,
				MatchType: matchType,
				Matches:   []MatchLine{},
			}
		}
		
		// Add match line
		resultMap[filePath].Matches = append(resultMap[filePath].Matches, MatchLine{
			LineNumber: lineNum,
			Content:    content,
		})
	}
	
	// Convert map to slice
	var results []SearchResult
	for _, result := range resultMap {
		results = append(results, *result)
	}
	
	return results
}

// filterResultsByAllKeywords filters results to only include files containing all keywords
func filterResultsByAllKeywords(repoPath string, results []SearchResult, keywords []string) []SearchResult {
	var filtered []SearchResult
	
	for _, result := range results {
		content, err := GetFileContent(repoPath, result.Path, 0)
		if err != nil {
			continue
		}
		
		contentLower := strings.ToLower(content)
		allMatch := true
		
		for _, keyword := range keywords {
			if !strings.Contains(contentLower, strings.ToLower(keyword)) {
				allMatch = false
				break
			}
		}
		
		if allMatch {
			filtered = append(filtered, result)
		}
	}
	
	return filtered
}

// removeDuplicateResults removes duplicate search results by path
func removeDuplicateResults(results []SearchResult) []SearchResult {
	seen := make(map[string]*SearchResult)
	
	for _, result := range results {
		if existing, exists := seen[result.Path]; exists {
			// Merge matches from duplicate results
			existing.Matches = append(existing.Matches, result.Matches...)
			// If one result has content matches and another has filename matches, combine them
			if result.MatchType == "filename" && existing.MatchType == "content" {
				existing.MatchType = "both"
			} else if result.MatchType == "content" && existing.MatchType == "filename" {
				existing.MatchType = "both"
			}
		} else {
			// Make a copy to avoid modifying the original
			resultCopy := result
			seen[result.Path] = &resultCopy
		}
	}
	
	// Convert back to slice
	var unique []SearchResult
	for _, result := range seen {
		unique = append(unique, *result)
	}
	
	return unique
}