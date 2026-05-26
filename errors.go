package main

import "errors"

// Sentinel errors that callers can classify via errors.Is. Returned from the
// business-logic layer and used at the MCP handler boundary to produce the
// right user-facing message without string matching.
var (
	// ErrWorkspaceNotInitialized indicates the workspace manager has not yet
	// been initialized at process start.
	ErrWorkspaceNotInitialized = errors.New("workspace not initialized")

	// ErrPathOutsideWorkspace indicates a requested path resolves outside of
	// the configured workspace directory.
	ErrPathOutsideWorkspace = errors.New("path outside workspace")

	// ErrRepositoryNotFound indicates the named repository does not exist in
	// the workspace.
	ErrRepositoryNotFound = errors.New("repository not found")

	// ErrMemoNotFound indicates the requested memo ID does not exist in the
	// memo store.
	ErrMemoNotFound = errors.New("memo not found")
)
