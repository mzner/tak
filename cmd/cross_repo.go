package cmd

import (
	"fmt"
	"strings"

	"github.com/mzner/tak/internal/config"
	"github.com/mzner/tak/internal/runner"
)

// resolveRemoteWorktree resolves a "repo:branch" reference to a filesystem path.
func resolveRemoteWorktree(r *runner.ExecRunner, ref string) (string, error) {
	parts := strings.SplitN(ref, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid cross-repo reference: %s (use repo:branch)", ref)
	}
	repoName := parts[0]
	branch := parts[1]

	repos, err := config.LoadGlobal()
	if err != nil {
		return "", fmt.Errorf("could not load global config: %w", err)
	}

	repoPath, ok := repos[repoName]
	if !ok {
		available := make([]string, 0, len(repos))
		for name := range repos {
			available = append(available, name)
		}
		return "", fmt.Errorf("unknown repo '%s' (available: %s)", repoName, strings.Join(available, ", "))
	}

	// If no branch specified (just "web:"), return the main repo root
	if branch == "" {
		return repoPath, nil
	}

	// Find the worktree path by listing worktrees in the target repo
	output, err := r.RunInDir(repoPath, "git", "worktree", "list", "--porcelain")
	if err != nil {
		return "", fmt.Errorf("could not list worktrees in %s: %w", repoName, err)
	}

	var currentPath string
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "worktree ") {
			currentPath = strings.TrimPrefix(line, "worktree ")
		}
		if strings.HasPrefix(line, "branch refs/heads/") {
			if strings.TrimPrefix(line, "branch refs/heads/") == branch {
				return currentPath, nil
			}
		}
	}

	return "", fmt.Errorf("no worktree for branch '%s' in repo '%s'", branch, repoName)
}
