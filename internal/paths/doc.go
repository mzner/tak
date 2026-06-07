// Package paths resolves filesystem paths for git worktrees.
//
// It takes a branch name and configuration (worktree base directory, repo root)
// and produces the directory path where the worktree should be created.
//
// Slugification rules for directory names:
//   - "/" becomes "--" (double dash)
//   - Spaces and special characters become "-"
//   - Lowercased
//
// Tmux window names use a different slug (single dash for readability):
//   - "/" becomes "-"
//
// Examples:
//
//	Resolve("feature/auth", "/Users/dev/projects/web", "")
//	// => "/Users/dev/projects/web--feature--auth"
//
//	TmuxSlug("feature/auth")
//	// => "feature-auth"
package paths
