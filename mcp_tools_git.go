package main

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterGitTools registers all Git-related MCP tools.
//
// Tool implementations live in dedicated files by area:
//   - mcp_tools_repo.go    repository lifecycle and info
//   - mcp_tools_files.go   file listing, content, and search
//   - mcp_tools_history.go commits and diffs
//   - mcp_tools_config.go  session defaults and batch ops
//
// Handlers use the errorResult/textResult helpers from mcp_errors.go to
// build CallToolResult values uniformly.
func RegisterGitTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "get_repository_info",
		Description: "Summarize a workspace repo: current branch, last update, remote URL, license, " +
			"file/dir counts, top file extensions, and README content. " +
			"Set include_memos=true to also list associated memos (memo_limit caps the count). " +
			"Use this first when exploring an unfamiliar repo.",
	}, handleGetRepositoryInfo)

	mcp.AddTool(server, &mcp.Tool{
		Name: "pull_repository",
		Description: "Run git pull on a workspace repo. Use after clone_repository when the remote " +
			"has new commits. Returns the git output verbatim.",
	}, handlePullRepository)

	mcp.AddTool(server, &mcp.Tool{
		Name: "list_branches",
		Description: "List local and remote-tracking branches of a workspace repo. " +
			"Use before switch_branch to discover available targets.",
	}, handleListBranches)

	mcp.AddTool(server, &mcp.Tool{
		Name: "switch_branch",
		Description: "Check out a branch in a workspace repo. The branch must already exist " +
			"(use list_branches to verify). This is a read-only inspection switch; do not use to " +
			"create branches.",
	}, handleSwitchBranch)

	mcp.AddTool(server, &mcp.Tool{
		Name: "search_files",
		Description: "Search file contents (and optionally filenames) by keywords across one or more " +
			"workspace repos. Set repositories=[...] for cross-repo search, or repository=name for one. " +
			"search_mode=\"and\"|\"or\" (default and); context_lines>0 includes surrounding lines; " +
			"include_patterns/exclude_patterns filter by glob (\"*.go\", \"vendor/**\").",
	}, handleSearchFiles)

	mcp.AddTool(server, &mcp.Tool{
		Name: "list_files",
		Description: "List files in a workspace repo with size, line count, and modification time. " +
			"directory=\"subdir\" scopes the listing; recursive=true descends. " +
			"include_patterns/exclude_patterns filter by glob (directory excludes like \"vendor/\" " +
			"prune entire subtrees).",
	}, handleListFiles)

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_file_content",
		Description: "Read one file (file_path) or several files (file_paths) from a workspace repo. " +
			"Line numbers are always emitted. start_line is 1-based; end_line is inclusive (defaults " +
			"to start_line + line_limit - 1, where line_limit comes from the session default, " +
			"falling back to 100).",
	}, handleGetFileContent)

	mcp.AddTool(server, &mcp.Tool{
		Name: "clone_repository",
		Description: "Clone a Git repository into the workspace. The name is derived from the URL " +
			"if omitted. If the repo already exists, runs git pull instead. " +
			"include_info=true to attach get_repository_info output; include_branches=true to attach " +
			"the branch list.",
	}, handleCloneRepository)

	mcp.AddTool(server, &mcp.Tool{
		Name: "list_repositories",
		Description: "List all repositories in the workspace. include_status=true adds git status " +
			"(branch, dirty flag) per repo; include_commits=true adds recent commits per repo, " +
			"capped at commit_limit (default 5).",
	}, handleListWorkspaceRepositories)

	mcp.AddTool(server, &mcp.Tool{
		Name: "remove_repository",
		Description: "Delete a workspace repo from disk. Destructive: there is no undo. The on-disk " +
			"clone is removed, but associated memos are preserved.",
	}, handleRemoveRepository)

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_readme_files",
		Description: "Find README files in a workspace repo. By default checks the repo root only; " +
			"recursive=true scans all subdirectories. Matches README, README.md, readme.rst, etc.",
	}, handleGetReadmeFiles)

	mcp.AddTool(server, &mcp.Tool{
		Name: "list_commits",
		Description: "List the most recent commits on the current branch of a workspace repo. " +
			"limit caps the number returned (default 20).",
	}, handleListCommits)

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_commit_diff",
		Description: "Show the diff for a single commit in a workspace repo. commit_hash accepts the " +
			"full SHA or any prefix git recognizes. Pair with list_commits to discover hashes.",
	}, handleGetCommitDiff)

	mcp.AddTool(server, &mcp.Tool{
		Name: "session",
		Description: "Manage server-side session defaults to reduce per-call parameter repetition. " +
			"action=\"set\" stores defaults (default_repository, default_include_patterns, " +
			"default_exclude_patterns, default_search_limit, default_list_files_limit, " +
			"default_line_limit, default_commit_limit); action=\"get\" returns the current state; " +
			"action=\"clear\" resets everything.",
	}, handleSession)

	mcp.AddTool(server, &mcp.Tool{
		Name: "batch",
		Description: "Run the same operation across multiple repos in one call. " +
			"operation=\"clone\" with urls=[...] clones each URL; \"pull\" updates listed repos " +
			"(or all workspace repos when repositories is empty); \"status\" reports current branch " +
			"and dirty flag.",
	}, handleBatch)
}
