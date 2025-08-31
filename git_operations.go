package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// RepositoryInfo contains basic repository information
type RepositoryInfo struct {
	Path            string    `json:"path"`
	CommitCount     int       `json:"commit_count"`
	LastUpdate      time.Time `json:"last_update"`
	CurrentBranch   string    `json:"current_branch"`
	License         string    `json:"license,omitempty"`
	ReadmeContent   string    `json:"readme_content,omitempty"`
	RemoteURL       string    `json:"remote_url,omitempty"`
}

// Branch represents a git branch
type Branch struct {
	Name      string `json:"name"`
	IsCurrent bool   `json:"is_current"`
	LastCommit string `json:"last_commit,omitempty"`
}

// SearchResult represents a file search result
type SearchResult struct {
	Path        string      `json:"path"`
	MatchType   string      `json:"match_type,omitempty"`   // "content" or "filename"
	Matches     []MatchLine `json:"matches,omitempty"`      // detailed match information
}

// MatchLine represents a single match within a file
type MatchLine struct {
	LineNumber int    `json:"line_number"`           // line number (0 for filename matches)
	Content    string `json:"content"`               // the matching line content
	Context    []string `json:"context,omitempty"`   // surrounding context lines
}

// FileInfo represents file or directory information
type FileInfo struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	IsDir     bool      `json:"is_dir"`
	Size      int64     `json:"size,omitempty"`
	ModTime   time.Time `json:"mod_time,omitempty"`
	CharCount int       `json:"char_count,omitempty"`  // Character count for text files
	LineCount int       `json:"line_count,omitempty"`  // Line count for text files
}

// GetRepositoryInfo retrieves basic repository information
func GetRepositoryInfo(repoPath string) (*RepositoryInfo, error) {
	// Validate workspace path
	validPath, err := ValidateWorkspacePath(repoPath)
	if err != nil {
		return nil, err
	}
	repoPath = validPath

	if !isGitRepository(repoPath) {
		return nil, fmt.Errorf("not a git repository: %s", repoPath)
	}

	info := &RepositoryInfo{Path: repoPath}

	// Get commit count
	if count, err := getCommitCount(repoPath); err == nil {
		info.CommitCount = count
	}

	// Get last update
	if lastUpdate, err := getLastCommit(repoPath); err == nil {
		info.LastUpdate = lastUpdate
	}

	// Get current branch
	if branch, err := getCurrentBranch(repoPath); err == nil {
		info.CurrentBranch = branch
	}

	// Get remote URL
	if remoteURL, err := getRemoteURL(repoPath); err == nil {
		info.RemoteURL = remoteURL
	}

	// Try to find license file
	if license, err := findLicenseFile(repoPath); err == nil {
		info.License = license
	}

	// Try to find and read README
	if readme, err := findAndReadReadme(repoPath); err == nil {
		info.ReadmeContent = readme
	}

	return info, nil
}

// PullRepository executes git pull on the specified repository
func PullRepository(repoPath string) (string, error) {
	// Validate workspace path
	validPath, err := ValidateWorkspacePath(repoPath)
	if err != nil {
		return "", err
	}
	repoPath = validPath

	if !isGitRepository(repoPath) {
		return "", fmt.Errorf("not a git repository: %s", repoPath)
	}

	cmd := exec.Command("git", "pull")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("git pull failed: %v", err)
	}

	return string(output), nil
}

// ListBranches lists all branches in the repository
func ListBranches(repoPath string) ([]Branch, error) {
	// Validate workspace path
	validPath, err := ValidateWorkspacePath(repoPath)
	if err != nil {
		return nil, err
	}
	repoPath = validPath

	if !isGitRepository(repoPath) {
		return nil, fmt.Errorf("not a git repository: %s", repoPath)
	}

	cmd := exec.Command("git", "branch", "-a")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %v", err)
	}

	var branches []Branch
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		branch := Branch{}
		if strings.HasPrefix(line, "* ") {
			branch.IsCurrent = true
			branch.Name = strings.TrimSpace(line[2:])
		} else {
			branch.Name = strings.TrimSpace(line)
		}

		// Skip remote tracking info
		if strings.Contains(branch.Name, " -> ") {
			continue
		}

		branches = append(branches, branch)
	}

	return branches, nil
}

// SwitchBranch switches to the specified branch
func SwitchBranch(repoPath, branchName string) (string, error) {
	// Validate workspace path
	validPath, err := ValidateWorkspacePath(repoPath)
	if err != nil {
		return "", err
	}
	repoPath = validPath

	if !isGitRepository(repoPath) {
		return "", fmt.Errorf("not a git repository: %s", repoPath)
	}

	cmd := exec.Command("git", "checkout", branchName)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("failed to switch branch: %v", err)
	}

	return string(output), nil
}

// SearchFiles searches for files containing the specified keywords
func SearchFiles(repoPath string, keywords []string, searchMode string, includeFilename bool, contextLines int, includePatterns, excludePatterns []string, maxResults int) ([]SearchResult, error) {
	return SearchFilesEnhanced(repoPath, keywords, searchMode, includeFilename, contextLines, includePatterns, excludePatterns, maxResults)
}

// ListFiles lists files in the specified directory
func ListFiles(repoPath, dirPath string, recursive bool, includePatterns, excludePatterns []string, maxResults int) ([]FileInfo, error) {
	// Validate workspace path
	validPath, err := ValidateWorkspacePath(repoPath)
	if err != nil {
		return nil, err
	}
	repoPath = validPath

	fullPath := filepath.Join(repoPath, dirPath)
	
	var files []FileInfo
	count := 0

	if recursive {
		err := filepath.WalkDir(fullPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // Skip errors, continue walking
			}

			// Skip .git directory
			if d.IsDir() && d.Name() == ".git" {
				return fs.SkipDir
			}

			// Get relative path
			relPath, err := filepath.Rel(repoPath, path)
			if err != nil {
				return nil
			}

			// Check if file should be included based on patterns
			if !shouldIncludeFile(relPath, includePatterns, excludePatterns) {
				if d.IsDir() {
					return nil // Skip directory contents but don't return SkipDir
				}
				return nil
			}

			info, err := d.Info()
			if err != nil {
				return nil
			}

			fileInfo := FileInfo{
				Name:    d.Name(),
				Path:    relPath,
				IsDir:   d.IsDir(),
				Size:    info.Size(),
				ModTime: info.ModTime(),
			}

			// Add character and line count for text files
			if !d.IsDir() {
				charCount, lineCount := countFileCharacters(path)
				fileInfo.CharCount = charCount
				fileInfo.LineCount = lineCount
			}

			files = append(files, fileInfo)

			count++
			if maxResults > 0 && count >= maxResults {
				return fmt.Errorf("max results reached")
			}

			return nil
		})

		if err != nil && err.Error() != "max results reached" {
			return nil, fmt.Errorf("failed to walk directory: %v", err)
		}
	} else {
		entries, err := os.ReadDir(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %v", err)
		}

		for _, entry := range entries {
			if maxResults > 0 && count >= maxResults {
				break
			}

			relPath := filepath.Join(dirPath, entry.Name())
			
			// Check if file should be included based on patterns
			if !shouldIncludeFile(relPath, includePatterns, excludePatterns) {
				continue
			}

			info, err := entry.Info()
			if err != nil {
				continue
			}

			fileInfo := FileInfo{
				Name:    entry.Name(),
				Path:    relPath,
				IsDir:   entry.IsDir(),
				Size:    info.Size(),
				ModTime: info.ModTime(),
			}

			// Add character and line count for text files
			if !entry.IsDir() {
				fullPath := filepath.Join(repoPath, relPath)
				charCount, lineCount := countFileCharacters(fullPath)
				fileInfo.CharCount = charCount
				fileInfo.LineCount = lineCount
			}

			files = append(files, fileInfo)
			count++
		}
	}

	return files, nil
}

// CloneRepository clones a Git repository into the workspace
// If repoName is empty, it will be extracted from the URL
func CloneRepository(repoURL, repoName string) (string, string, error) {
	wm := GetWorkspaceManager()
	if wm == nil {
		return "", "", fmt.Errorf("workspace not initialized")
	}

	// Extract repository name from URL if not provided
	if repoName == "" {
		var err error
		repoName, err = extractRepoNameFromURL(repoURL)
		if err != nil {
			return "", "", fmt.Errorf("failed to extract repository name from URL: %v", err)
		}
	}

	// Check if repository already exists
	if wm.RepositoryExists(repoName) {
		return "", "", fmt.Errorf("repository '%s' already exists in workspace", repoName)
	}

	// Get target path for clone
	targetPath := wm.GetRepositoryPath(repoName)

	// Execute git clone
	cmd := exec.Command("git", "clone", repoURL, targetPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), repoName, fmt.Errorf("git clone failed: %v", err)
	}

	return string(output), repoName, nil
}

// extractRepoNameFromURL extracts the repository name from a Git URL
func extractRepoNameFromURL(repoURL string) (string, error) {
	if repoURL == "" {
		return "", fmt.Errorf("repository URL cannot be empty")
	}

	// Handle different URL formats:
	// https://github.com/user/repo.git
	// git@github.com:user/repo.git
	// https://github.com/user/repo
	// etc.
	
	// Remove .git suffix if present
	url := strings.TrimSuffix(repoURL, ".git")

	// Split by / and get the last part
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid repository URL format")
	}

	repoName := parts[len(parts)-1]
	
	// Handle SSH URLs like git@github.com:user/repo
	if strings.Contains(repoName, ":") {
		colonParts := strings.Split(repoName, ":")
		if len(colonParts) > 1 {
			pathParts := strings.Split(colonParts[len(colonParts)-1], "/")
			repoName = pathParts[len(pathParts)-1]
		}
	}

	// Validate repository name
	if repoName == "" {
		return "", fmt.Errorf("could not extract repository name from URL")
	}

	// Clean repository name (remove invalid characters for directory names)
	repoName = strings.ReplaceAll(repoName, " ", "-")
	repoName = strings.ReplaceAll(repoName, ":", "-")
	
	return repoName, nil
}

// FileContentResult represents the content of a single file
type FileContentResult struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
	Error    string `json:"error,omitempty"`
}

// GetFileContent reads the content of a file
func GetFileContent(repoPath, filePath string, maxLines int) (string, error) {
	// Validate workspace path
	validPath, err := ValidateWorkspacePath(repoPath)
	if err != nil {
		return "", err
	}
	repoPath = validPath

	fullPath := filepath.Join(repoPath, filePath)
	
	file, err := os.Open(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var content strings.Builder
	scanner := bufio.NewScanner(file)
	lineCount := 0

	for scanner.Scan() && (maxLines == 0 || lineCount < maxLines) {
		content.WriteString(scanner.Text())
		content.WriteString("\n")
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	return content.String(), nil
}

// GetMultipleFileContents reads the content of multiple files
func GetMultipleFileContents(repoPath string, filePaths []string, maxLines int) ([]FileContentResult, error) {
	// Validate workspace path
	validPath, err := ValidateWorkspacePath(repoPath)
	if err != nil {
		return nil, err
	}
	repoPath = validPath

	var results []FileContentResult

	for _, filePath := range filePaths {
		result := FileContentResult{
			FilePath: filePath,
		}

		content, err := GetFileContent(repoPath, filePath, maxLines)
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Content = content
		}

		results = append(results, result)
	}

	return results, nil
}

// Helper functions

func isGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	if stat, err := os.Stat(gitDir); err == nil {
		return stat.IsDir()
	}
	return false
}

func getCommitCount(repoPath string) (int, error) {
	cmd := exec.Command("git", "rev-list", "--all", "--count")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	
	count, err := strconv.Atoi(strings.TrimSpace(string(output)))
	return count, err
}

func getLastCommit(repoPath string) (time.Time, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%ci")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return time.Time{}, err
	}
	
	return time.Parse("2006-01-02 15:04:05 -0700", strings.TrimSpace(string(output)))
}

func getCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	
	return strings.TrimSpace(string(output)), nil
}

func getRemoteURL(repoPath string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	
	return strings.TrimSpace(string(output)), nil
}

func findLicenseFile(repoPath string) (string, error) {
	licenseFiles := []string{"LICENSE", "LICENSE.txt", "LICENSE.md", "LICENCE", "LICENCE.txt", "LICENCE.md"}
	
	for _, filename := range licenseFiles {
		path := filepath.Join(repoPath, filename)
		if _, err := os.Stat(path); err == nil {
			return filename, nil
		}
	}
	
	return "", fmt.Errorf("no license file found")
}

func findAndReadReadme(repoPath string) (string, error) {
	readmeFiles := []string{"README.md", "README.txt", "README", "readme.md", "readme.txt", "readme"}
	
	for _, filename := range readmeFiles {
		path := filepath.Join(repoPath, filename)
		if _, err := os.Stat(path); err == nil {
			content, err := GetFileContent(repoPath, filename, 50) // Limit to 50 lines
			if err == nil {
				return content, nil
			}
		}
	}
	
	return "", fmt.Errorf("no readme file found")
}

func filterResultsByKeywords(repoPath string, results []SearchResult, keywords []string) []SearchResult {
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

// mergeResultsWithOR searches for files containing any of the additional keywords (OR logic)
func mergeResultsWithOR(repoPath string, existingResults []SearchResult, keywords []string) []SearchResult {
	// Create a map to avoid duplicates
	resultMap := make(map[string]SearchResult)
	
	// Add existing results
	for _, result := range existingResults {
		resultMap[result.Path] = result
	}
	
	// Search for each additional keyword separately
	for _, keyword := range keywords {
		args := []string{"grep", "-l", "-r", "--exclude-dir=.git", keyword}
		cmd := exec.Command("git", args...)
		cmd.Dir = repoPath
		output, err := cmd.Output()
		
		if err != nil {
			// No matches found is not an error for OR logic
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				continue
			}
			// Continue with other keywords even if one fails
			continue
		}
		
		scanner := bufio.NewScanner(strings.NewReader(string(output)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				resultMap[line] = SearchResult{Path: line}
			}
		}
	}
	
	// Convert map back to slice
	var results []SearchResult
	for _, result := range resultMap {
		results = append(results, result)
	}
	
	return results
}

// File pattern matching helper functions

// matchesPatterns checks if a file path matches any of the given patterns
func matchesPatterns(filePath string, patterns []string) bool {
	if len(patterns) == 0 {
		return true // No patterns means match all
	}
	
	for _, pattern := range patterns {
		if matched, err := filepath.Match(pattern, filepath.Base(filePath)); err == nil && matched {
			return true
		}
		// Also try matching against the full path
		if matched, err := filepath.Match(pattern, filePath); err == nil && matched {
			return true
		}
		// Support directory patterns like "*.go" or "src/*.go"
		if strings.Contains(pattern, "/") {
			if matched, err := filepath.Match(pattern, filePath); err == nil && matched {
				return true
			}
		}
	}
	return false
}

// shouldIncludeFile determines if a file should be included based on include/exclude patterns
func shouldIncludeFile(filePath string, includePatterns, excludePatterns []string) bool {
	// Check exclude patterns first
	if len(excludePatterns) > 0 && matchesPatterns(filePath, excludePatterns) {
		return false
	}
	
	// Check include patterns
	return matchesPatterns(filePath, includePatterns)
}

// countFileCharacters counts characters and lines in a text file
func countFileCharacters(fullPath string) (int, int) {
	file, err := os.Open(fullPath)
	if err != nil {
		return 0, 0
	}
	defer file.Close()

	var charCount, lineCount int
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := scanner.Text()
		charCount += len(line) + 1 // +1 for newline character
		lineCount++
	}
	
	// If file doesn't end with newline, don't count the extra character
	if lineCount > 0 && charCount > 0 {
		charCount-- // Remove the last extra newline count
	}
	
	return charCount, lineCount
}