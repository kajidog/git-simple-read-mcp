package main

import "sync"

// SessionConfig holds server-side session configuration
// This allows clients to set defaults that persist across tool calls
type SessionConfig struct {
	mu sync.RWMutex

	// Default repository for all operations (when not specified)
	DefaultRepository string `json:"default_repository,omitempty"`

	// Default search patterns
	DefaultIncludePatterns []string `json:"default_include_patterns,omitempty"`
	DefaultExcludePatterns []string `json:"default_exclude_patterns,omitempty"`

	// Default limits
	DefaultSearchLimit    int `json:"default_search_limit,omitempty"`
	DefaultListFilesLimit int `json:"default_list_files_limit,omitempty"`
	DefaultMaxLines       int `json:"default_max_lines,omitempty"`
	DefaultCommitLimit    int `json:"default_commit_limit,omitempty"`
}

// Global session config instance
var globalSessionConfig = &SessionConfig{}

// GetSessionConfig returns the current session configuration
func GetSessionConfig() *SessionConfig {
	return globalSessionConfig
}

// SetSessionConfigValues sets the session configuration values
func SetSessionConfigValues(config *SessionConfig) {
	globalSessionConfig.mu.Lock()
	defer globalSessionConfig.mu.Unlock()

	if config.DefaultRepository != "" {
		globalSessionConfig.DefaultRepository = config.DefaultRepository
	}
	if len(config.DefaultIncludePatterns) > 0 {
		globalSessionConfig.DefaultIncludePatterns = config.DefaultIncludePatterns
	}
	if len(config.DefaultExcludePatterns) > 0 {
		globalSessionConfig.DefaultExcludePatterns = config.DefaultExcludePatterns
	}
	if config.DefaultSearchLimit > 0 {
		globalSessionConfig.DefaultSearchLimit = config.DefaultSearchLimit
	}
	if config.DefaultListFilesLimit > 0 {
		globalSessionConfig.DefaultListFilesLimit = config.DefaultListFilesLimit
	}
	if config.DefaultMaxLines > 0 {
		globalSessionConfig.DefaultMaxLines = config.DefaultMaxLines
	}
	if config.DefaultCommitLimit > 0 {
		globalSessionConfig.DefaultCommitLimit = config.DefaultCommitLimit
	}
}

// ClearSessionConfig resets the session configuration to defaults
func ClearSessionConfig() {
	globalSessionConfig.mu.Lock()
	defer globalSessionConfig.mu.Unlock()

	globalSessionConfig.DefaultRepository = ""
	globalSessionConfig.DefaultIncludePatterns = nil
	globalSessionConfig.DefaultExcludePatterns = nil
	globalSessionConfig.DefaultSearchLimit = 0
	globalSessionConfig.DefaultListFilesLimit = 0
	globalSessionConfig.DefaultMaxLines = 0
	globalSessionConfig.DefaultCommitLimit = 0
}

// GetRepository returns the provided repository or the default if empty
func (sc *SessionConfig) GetRepository(provided string) string {
	if provided != "" {
		return provided
	}
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.DefaultRepository
}

// GetIncludePatterns returns the provided patterns or the default if empty
func (sc *SessionConfig) GetIncludePatterns(provided []string) []string {
	if len(provided) > 0 {
		return provided
	}
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.DefaultIncludePatterns
}

// GetExcludePatterns returns the provided patterns or the default if empty
func (sc *SessionConfig) GetExcludePatterns(provided []string) []string {
	if len(provided) > 0 {
		return provided
	}
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.DefaultExcludePatterns
}

// GetSearchLimit returns the provided limit or the default if zero
func (sc *SessionConfig) GetSearchLimit(provided int) int {
	if provided > 0 {
		return provided
	}
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	if sc.DefaultSearchLimit > 0 {
		return sc.DefaultSearchLimit
	}
	return 20 // Default fallback
}

// GetListFilesLimit returns the provided limit or the default if zero
func (sc *SessionConfig) GetListFilesLimit(provided int) int {
	if provided > 0 {
		return provided
	}
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	if sc.DefaultListFilesLimit > 0 {
		return sc.DefaultListFilesLimit
	}
	return 50 // Default fallback
}

// GetMaxLines returns the provided max lines or the default if zero
func (sc *SessionConfig) GetMaxLines(provided int) int {
	if provided > 0 {
		return provided
	}
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	if sc.DefaultMaxLines > 0 {
		return sc.DefaultMaxLines
	}
	return 100 // Default fallback
}

// GetCommitLimit returns the provided limit or the default if zero
func (sc *SessionConfig) GetCommitLimit(provided int) int {
	if provided > 0 {
		return provided
	}
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	if sc.DefaultCommitLimit > 0 {
		return sc.DefaultCommitLimit
	}
	return 20 // Default fallback
}

// ToMap returns the session configuration as a map for display
func (sc *SessionConfig) ToMap() map[string]any {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	result := make(map[string]any)

	if sc.DefaultRepository != "" {
		result["default_repository"] = sc.DefaultRepository
	}
	if len(sc.DefaultIncludePatterns) > 0 {
		result["default_include_patterns"] = sc.DefaultIncludePatterns
	}
	if len(sc.DefaultExcludePatterns) > 0 {
		result["default_exclude_patterns"] = sc.DefaultExcludePatterns
	}
	if sc.DefaultSearchLimit > 0 {
		result["default_search_limit"] = sc.DefaultSearchLimit
	}
	if sc.DefaultListFilesLimit > 0 {
		result["default_list_files_limit"] = sc.DefaultListFilesLimit
	}
	if sc.DefaultMaxLines > 0 {
		result["default_max_lines"] = sc.DefaultMaxLines
	}
	if sc.DefaultCommitLimit > 0 {
		result["default_commit_limit"] = sc.DefaultCommitLimit
	}

	return result
}

// IsEmpty returns true if no configuration values are set
func (sc *SessionConfig) IsEmpty() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return sc.DefaultRepository == "" &&
		len(sc.DefaultIncludePatterns) == 0 &&
		len(sc.DefaultExcludePatterns) == 0 &&
		sc.DefaultSearchLimit == 0 &&
		sc.DefaultListFilesLimit == 0 &&
		sc.DefaultMaxLines == 0 &&
		sc.DefaultCommitLimit == 0
}
