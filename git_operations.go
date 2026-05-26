package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// RepositoryInfo contains basic repository information
type RepositoryInfo struct {
	Path          string    `json:"path"`
	LastUpdate    time.Time `json:"last_update"`
	CurrentBranch string    `json:"current_branch"`
	License       string    `json:"license,omitempty"`
	ReadmeContent string    `json:"readme_content,omitempty"`
	RemoteURL     string    `json:"remote_url,omitempty"`
}

// Branch represents a git branch
type Branch struct {
	Name       string `json:"name"`
	IsCurrent  bool   `json:"is_current"`
	LastCommit string `json:"last_commit,omitempty"`
}

// Commit represents a git commit
type Commit struct {
	Hash    string `json:"hash"`
	Author  string `json:"author"`
	Date    string `json:"date"`
	Message string `json:"message"`
}

// SearchResult represents a file search result
type SearchResult struct {
	Path      string      `json:"path"`
	MatchType string      `json:"match_type,omitempty"` // "content" or "filename"
	Matches   []MatchLine `json:"matches,omitempty"`    // detailed match information
}

// MatchLine represents a single match within a file
type MatchLine struct {
	LineNumber int      `json:"line_number"`       // line number (0 for filename matches)
	Content    string   `json:"content"`           // the matching line content
	Context    []string `json:"context,omitempty"` // surrounding context lines
}

// FileInfo represents file information
type FileInfo struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Size      int64     `json:"size,omitempty"`
	ModTime   time.Time `json:"mod_time,omitempty"`
	LineCount int       `json:"line_count,omitempty"` // Line count for text files
}

// RepositoryStatus represents the current status of a repository
type RepositoryStatus struct {
	CurrentBranch string `json:"current_branch"`
	HasChanges    bool   `json:"has_changes"`
	StatusOutput  string `json:"status_output,omitempty"`
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

// GetRepositoryStatus returns the current status of a repository
func GetRepositoryStatus(repoPath string) (*RepositoryStatus, error) {
	// Validate workspace path
	validPath, err := ValidateWorkspacePath(repoPath)
	if err != nil {
		return nil, err
	}
	repoPath = validPath

	if !isGitRepository(repoPath) {
		return nil, fmt.Errorf("not a git repository: %s", repoPath)
	}

	status := &RepositoryStatus{}

	// Get current branch
	if branch, err := getCurrentBranch(repoPath); err == nil {
		status.CurrentBranch = branch
	}

	// Get git status (porcelain format for easy parsing)
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git status: %w", err)
	}

	statusOutput := strings.TrimSpace(string(output))
	status.StatusOutput = statusOutput
	status.HasChanges = len(statusOutput) > 0

	return status, nil
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
		return string(output), fmt.Errorf("git pull failed: %w", err)
	}

	return string(output), nil
}

// ListCommits lists commits in the repository
func ListCommits(repoPath string, limit int) ([]Commit, error) {
	validPath, err := ValidateWorkspacePath(repoPath)
	if err != nil {
		return nil, err
	}
	repoPath = validPath

	if !isGitRepository(repoPath) {
		return nil, fmt.Errorf("not a git repository: %s", repoPath)
	}

	args := []string{"log", "--pretty=format:%H|%an|%ad|%s", "--date=iso"}
	if limit > 0 {
		args = append(args, fmt.Sprintf("--max-count=%d", limit))
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list commits: %w", err)
	}

	var commits []Commit
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "|", 4)
		if len(parts) == 4 {
			commit := Commit{
				Hash:    parts[0],
				Author:  parts[1],
				Date:    parts[2],
				Message: parts[3],
			}
			commits = append(commits, commit)
		}
	}

	return commits, nil
}

// GetCommitDiff gets the diff for a specific commit
func GetCommitDiff(repoPath, commitHash string) (string, error) {
	validPath, err := ValidateWorkspacePath(repoPath)
	if err != nil {
		return "", err
	}
	repoPath = validPath

	if !isGitRepository(repoPath) {
		return "", fmt.Errorf("not a git repository: %s", repoPath)
	}

	// Basic validation for commit hash
	if len(commitHash) < 4 || len(commitHash) > 40 {
		return "", fmt.Errorf("invalid commit hash format")
	}

	cmd := exec.Command("git", "show", commitHash)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("git show failed for commit '%s': %w", commitHash, err)
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
		return nil, fmt.Errorf("failed to list branches: %w", err)
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
		return string(output), fmt.Errorf("failed to switch branch: %w", err)
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

			// For directories, check if we should skip the entire subtree
			if d.IsDir() {
				if shouldSkipDirectory(relPath, excludePatterns) {
					return fs.SkipDir // Skip this directory and all its contents
				}
				return nil // Skip directory entries in output
			}

			// Check if file should be included based on patterns
			if !shouldIncludeFile(relPath, includePatterns, excludePatterns) {
				return nil
			}

			info, err := d.Info()
			if err != nil {
				return nil
			}

			_, lineCount := countFileCharacters(path)
			fileInfo := FileInfo{
				Name:      d.Name(),
				Path:      relPath,
				Size:      info.Size(),
				ModTime:   info.ModTime(),
				LineCount: lineCount,
			}

			files = append(files, fileInfo)

			count++
			if maxResults > 0 && count >= maxResults {
				return fmt.Errorf("max results reached")
			}

			return nil
		})

		if err != nil && err.Error() != "max results reached" {
			return nil, fmt.Errorf("failed to walk directory: %w", err)
		}
	} else {
		entries, err := os.ReadDir(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %w", err)
		}

		for _, entry := range entries {
			if maxResults > 0 && count >= maxResults {
				break
			}

			// Skip directories in output
			if entry.IsDir() {
				continue
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

			entryFullPath := filepath.Join(repoPath, relPath)
			_, lineCount := countFileCharacters(entryFullPath)
			fileInfo := FileInfo{
				Name:      entry.Name(),
				Path:      relPath,
				Size:      info.Size(),
				ModTime:   info.ModTime(),
				LineCount: lineCount,
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
		return "", "", ErrWorkspaceNotInitialized
	}

	// Extract repository name from URL if not provided
	if repoName == "" {
		var err error
		repoName, err = extractRepoNameFromURL(repoURL)
		if err != nil {
			return "", "", fmt.Errorf("failed to extract repository name from URL: %w", err)
		}
	}

	// Check if repository already exists
	if wm.RepositoryExists(repoName) {
		return "", repoName, fmt.Errorf("repository '%s' already exists in workspace", repoName)
	}

	// Get target path for clone
	targetPath := wm.GetRepositoryPath(repoName)

	// Execute git clone
	cmd := exec.Command("git", "clone", repoURL, targetPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), repoName, fmt.Errorf("git clone failed: %w", err)
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
	FilePath   string `json:"file_path"`
	Content    string `json:"content"`
	Error      string `json:"error,omitempty"`
	TotalLines int    `json:"total_lines,omitempty"`
	StartLine  int    `json:"start_line,omitempty"`
	EndLine    int    `json:"end_line,omitempty"`
}

// ReadmeFileInfo represents information about a README file
type ReadmeFileInfo struct {
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	ModTime   time.Time `json:"mod_time"`
	LineCount int       `json:"line_count,omitempty"`
}

// FileStatistics contains file statistics for a repository
type FileStatistics struct {
	TotalFiles      int            `json:"total_files"`
	TotalDirs       int            `json:"total_dirs"`
	ExtensionCounts map[string]int `json:"extension_counts"` // extension -> count
}

// GetFileStatistics returns file statistics for a repository
func GetFileStatistics(repoPath string, excludePatterns []string) (*FileStatistics, error) {
	// Validate workspace path
	validPath, err := ValidateWorkspacePath(repoPath)
	if err != nil {
		return nil, err
	}
	repoPath = validPath

	stats := &FileStatistics{
		ExtensionCounts: make(map[string]int),
	}

	err = filepath.WalkDir(repoPath, func(path string, d fs.DirEntry, err error) error {
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

		// Skip root
		if relPath == "." {
			return nil
		}

		// For directories, check if we should skip the entire subtree
		if d.IsDir() {
			if shouldSkipDirectory(relPath, excludePatterns) {
				return fs.SkipDir
			}
			stats.TotalDirs++
			return nil
		}

		// Check exclude patterns for files
		if len(excludePatterns) > 0 && matchesPatterns(relPath, excludePatterns) {
			return nil
		}

		stats.TotalFiles++

		// Count by extension
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if ext == "" {
			ext = "(no ext)"
		}
		stats.ExtensionCounts[ext]++

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return stats, nil
}

// GetFileContent reads the content of a file, optionally limited to the
// first maxLines lines (0 = unlimited). Long lines are supported without
// the 64KB bufio.Scanner token limit: the file is loaded into memory and
// split on '\n'.
//
// Output normalizes line endings: each returned line is followed by '\n',
// regardless of whether the source file had a trailing newline.
func GetFileContent(repoPath, filePath string, maxLines int) (string, error) {
	// Validate workspace path
	validPath, err := ValidateWorkspacePath(repoPath)
	if err != nil {
		return "", err
	}
	repoPath = validPath

	fullPath := filepath.Join(repoPath, filePath)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	if maxLines > 0 && len(lines) > maxLines {
		lines = lines[:maxLines]
	}

	var content strings.Builder
	for _, line := range lines {
		content.WriteString(line)
		content.WriteString("\n")
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

// GetFileContentWithLineNumbers reads the content of a file with optional line numbers.
// Returns: content, totalLines, actualStartLine, actualEndLine, error.
//
// Reads the file in a single pass and builds an index of newline positions, so
// long lines (which would overflow bufio.Scanner's default 64KB token buffer)
// are handled without error. The line span [startLine, startLine+maxLines) is
// then sliced from the in-memory bytes.
func GetFileContentWithLineNumbers(repoPath, filePath string, startLine, maxLines int, showLineNumbers bool) (string, int, int, int, error) {
	// Validate workspace path
	validPath, err := ValidateWorkspacePath(repoPath)
	if err != nil {
		return "", 0, 0, 0, err
	}
	repoPath = validPath

	fullPath := filepath.Join(repoPath, filePath)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", 0, 0, 0, fmt.Errorf("failed to open file: %w", err)
	}

	// Split into lines without losing the (lack of) trailing newline. Using
	// strings.Split on "\n" gives us len(lines) = (number of newlines) + 1,
	// where a trailing newline produces an empty final element. We trim that
	// to make totalLines match the human-visible line count.
	lines := strings.Split(string(data), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	totalLines := len(lines)

	// Normalize startLine
	if startLine < 1 {
		startLine = 1
	}
	if startLine > totalLines {
		return "", totalLines, startLine, startLine, nil
	}

	end := totalLines
	if maxLines > 0 && startLine+maxLines-1 < end {
		end = startLine + maxLines - 1
	}

	var content strings.Builder
	for i := startLine - 1; i < end; i++ {
		if showLineNumbers {
			fmt.Fprintf(&content, "%4d: %s\n", i+1, lines[i])
		} else {
			content.WriteString(lines[i])
			content.WriteString("\n")
		}
	}

	return content.String(), totalLines, startLine, end, nil
}

// GetMultipleFileContentsWithLineNumbers reads the content of multiple files with optional line numbers
func GetMultipleFileContentsWithLineNumbers(repoPath string, filePaths []string, startLine, maxLines int, showLineNumbers bool) ([]FileContentResult, error) {
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

		content, totalLines, actualStart, actualEnd, err := GetFileContentWithLineNumbers(repoPath, filePath, startLine, maxLines, showLineNumbers)
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Content = content
			result.TotalLines = totalLines
			result.StartLine = actualStart
			result.EndLine = actualEnd
		}

		results = append(results, result)
	}

	return results, nil
}

// GetReadmeFiles finds all README files in the repository
func GetReadmeFiles(repoPath string, recursive bool) ([]ReadmeFileInfo, error) {
	// Validate workspace path
	validPath, err := ValidateWorkspacePath(repoPath)
	if err != nil {
		return nil, err
	}
	repoPath = validPath

	var readmeFiles []ReadmeFileInfo
	readmePatterns := []string{"README", "README.*", "readme", "readme.*", "Readme", "Readme.*"}

	if recursive {
		err := filepath.WalkDir(repoPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // Skip errors, continue walking
			}

			// Skip .git directory
			if d.IsDir() && d.Name() == ".git" {
				return fs.SkipDir
			}

			// Skip directories
			if d.IsDir() {
				return nil
			}

			// Check if file matches README pattern
			fileName := d.Name()
			for _, pattern := range readmePatterns {
				if matched, err := filepath.Match(pattern, fileName); err == nil && matched {
					info, err := d.Info()
					if err != nil {
						return nil
					}

					// Get relative path
					relPath, err := filepath.Rel(repoPath, path)
					if err != nil {
						return nil
					}

					// Count lines for text files
					_, lineCount := countFileCharacters(path)

					readmeInfo := ReadmeFileInfo{
						Path:      relPath,
						Size:      info.Size(),
						ModTime:   info.ModTime(),
						LineCount: lineCount,
					}

					readmeFiles = append(readmeFiles, readmeInfo)
					break // Don't check other patterns for this file
				}
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to walk directory: %w", err)
		}
	} else {
		// Only search in root directory
		entries, err := os.ReadDir(repoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			fileName := entry.Name()
			for _, pattern := range readmePatterns {
				if matched, err := filepath.Match(pattern, fileName); err == nil && matched {
					info, err := entry.Info()
					if err != nil {
						continue
					}

					// Count lines for text files
					fullPath := filepath.Join(repoPath, fileName)
					_, lineCount := countFileCharacters(fullPath)

					readmeInfo := ReadmeFileInfo{
						Path:      fileName,
						Size:      info.Size(),
						ModTime:   info.ModTime(),
						LineCount: lineCount,
					}

					readmeFiles = append(readmeFiles, readmeInfo)
					break // Don't check other patterns for this file
				}
			}
		}
	}

	return readmeFiles, nil
}

// Helper functions

func isGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	if stat, err := os.Stat(gitDir); err == nil {
		return stat.IsDir()
	}
	return false
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
			// Count total lines in the README
			fullPath := filepath.Join(repoPath, filename)
			_, totalLines := countFileCharacters(fullPath)

			// Get content (no limit to show full README)
			content, err := GetFileContent(repoPath, filename, 0) // 0 = no limit
			if err == nil {
				// If README is very long, add a note about total lines
				if totalLines > 100 {
					content += fmt.Sprintf("\n[README: %d lines total]\n", totalLines)
				}
				return content, nil
			}
		}
	}

	return "", fmt.Errorf("no readme file found")
}

// File pattern matching helper functions

// matchesPatterns checks if a file path matches any of the given patterns
func matchesPatterns(filePath string, patterns []string) bool {
	if len(patterns) == 0 {
		return true // No patterns means match all
	}

	for _, pattern := range patterns {
		if matchesPattern(filePath, pattern) {
			return true
		}
	}
	return false
}

// matchesPattern checks if a file path matches a single pattern
// Supports:
// - Simple glob: *.go, *.js
// - Directory patterns: vendor/, node_modules/
// - Path patterns: src/*.go, test/*
// - Recursive patterns: vendor/**, **/test/**
func matchesPattern(filePath, pattern string) bool {
	// Handle directory patterns ending with /
	// e.g., "vendor/" should match "vendor/foo/bar.go"
	if strings.HasSuffix(pattern, "/") {
		dirPattern := strings.TrimSuffix(pattern, "/")
		if filePath == dirPattern || strings.HasPrefix(filePath, dirPattern+"/") {
			return true
		}
	}

	// Handle recursive patterns with **
	// e.g., "vendor/**" should match "vendor/foo/bar.go"
	if strings.Contains(pattern, "**") {
		return matchesRecursivePattern(filePath, pattern)
	}

	// Try matching against basename
	if matched, err := filepath.Match(pattern, filepath.Base(filePath)); err == nil && matched {
		return true
	}

	// Try matching against the full path
	if matched, err := filepath.Match(pattern, filePath); err == nil && matched {
		return true
	}

	// For patterns with /, check if any path prefix matches
	// e.g., "vendor/*" should match "vendor/foo" but also check parent dirs
	if strings.Contains(pattern, "/") {
		// Check if pattern matches any parent directory path
		parts := strings.Split(filePath, "/")
		for i := 1; i <= len(parts); i++ {
			subPath := strings.Join(parts[:i], "/")
			if matched, err := filepath.Match(pattern, subPath); err == nil && matched {
				return true
			}
		}
	}

	return false
}

// matchesRecursivePattern handles ** glob patterns
// ** matches zero or more directory levels
func matchesRecursivePattern(filePath, pattern string) bool {
	// Handle patterns like **/test/** (directory anywhere in path)
	if strings.HasPrefix(pattern, "**/") && strings.HasSuffix(pattern, "/**") {
		// Extract the middle part, e.g., "test" from "**/test/**"
		middle := strings.TrimPrefix(pattern, "**/")
		middle = strings.TrimSuffix(middle, "/**")
		// Check if any path component matches the middle part
		parts := strings.Split(filePath, "/")
		for _, part := range parts {
			if part == middle {
				return true
			}
		}
		return false
	}

	// Handle patterns starting with **/ (match any prefix)
	if strings.HasPrefix(pattern, "**/") {
		suffix := strings.TrimPrefix(pattern, "**/")
		// Try matching suffix against the path and all subpaths
		if matchesPattern(filePath, suffix) {
			return true
		}
		parts := strings.Split(filePath, "/")
		for i := range parts {
			subPath := strings.Join(parts[i:], "/")
			if matchesPattern(subPath, suffix) {
				return true
			}
		}
		return false
	}

	// Split pattern by **
	parts := strings.Split(pattern, "**")

	if len(parts) == 2 {
		prefix := strings.TrimSuffix(parts[0], "/")
		suffix := strings.TrimPrefix(parts[1], "/")

		// Check prefix - must match exact directory boundary
		// e.g., "vendor/**" should match "vendor/foo" but NOT "vendor2/foo"
		if prefix != "" {
			if filePath != prefix && !strings.HasPrefix(filePath, prefix+"/") {
				return false
			}
		}

		// Check suffix
		if suffix != "" {
			// Suffix can be a pattern like "*.go"
			if strings.Contains(suffix, "*") && !strings.Contains(suffix, "**") {
				// Match suffix pattern against basename or remaining path
				remaining := filePath
				if prefix != "" {
					remaining = strings.TrimPrefix(filePath, prefix)
					remaining = strings.TrimPrefix(remaining, "/")
				}
				// Try matching against basename
				if matched, err := filepath.Match(suffix, filepath.Base(remaining)); err == nil && matched {
					return true
				}
				// Try matching against each path segment
				segments := strings.Split(remaining, "/")
				for i := range segments {
					subPath := strings.Join(segments[i:], "/")
					if matched, err := filepath.Match(suffix, subPath); err == nil && matched {
						return true
					}
				}
				return false
			}
			return strings.HasSuffix(filePath, suffix) || strings.Contains(filePath, suffix+"/")
		}

		// No suffix means match everything under prefix
		return true
	}

	// Fallback for complex patterns - try simple prefix/suffix matching
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

// shouldSkipDirectory checks if a directory should be completely skipped during traversal
// This is an optimization to avoid walking into large excluded directories like vendor/ or node_modules/
func shouldSkipDirectory(dirPath string, excludePatterns []string) bool {
	if len(excludePatterns) == 0 {
		return false
	}

	for _, pattern := range excludePatterns {
		// Check directory patterns ending with /
		if strings.HasSuffix(pattern, "/") {
			dirPattern := strings.TrimSuffix(pattern, "/")
			if dirPath == dirPattern || strings.HasPrefix(dirPath, dirPattern+"/") {
				return true
			}
		}

		// Check recursive patterns like vendor/**
		if strings.HasSuffix(pattern, "/**") {
			dirPattern := strings.TrimSuffix(pattern, "/**")
			if dirPath == dirPattern || strings.HasPrefix(dirPath, dirPattern+"/") {
				return true
			}
		}

		// Check simple directory name match (e.g., "vendor" matches "vendor" directory)
		if !strings.Contains(pattern, "*") && !strings.Contains(pattern, "/") {
			if dirPath == pattern || strings.HasPrefix(dirPath, pattern+"/") {
				return true
			}
		}

		// Check path patterns like "src/vendor/*"
		if strings.HasSuffix(pattern, "/*") {
			dirPattern := strings.TrimSuffix(pattern, "/*")
			if dirPath == dirPattern {
				return true
			}
		}
	}

	return false
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
