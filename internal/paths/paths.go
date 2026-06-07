package paths

import (
	"path/filepath"
	"regexp"
	"strings"
)

var specialChars = regexp.MustCompile(`[^a-z0-9.\-/]`)

// SlugifyBranch converts a branch name into a filesystem-safe slug.
// Slashes become "--", special chars become "-", result is lowercased.
func SlugifyBranch(branch string) string {
	slug := strings.ToLower(branch)
	slug = specialChars.ReplaceAllString(slug, "-")
	slug = strings.ReplaceAll(slug, "/", "--")
	return slug
}

// TmuxSlug converts a branch name into a tmux-friendly window name.
// Slashes become "-" (single dash for readability in tmux status bar).
func TmuxSlug(branch string) string {
	return strings.ReplaceAll(branch, "/", "-")
}

// Resolve returns the filesystem path where a worktree should be created.
//
// If worktreeBase is empty, the worktree is created as a sibling directory
// of the repo (e.g., /projects/web → /projects/web--feature--auth).
//
// If worktreeBase is set, the worktree is created under that directory
// (e.g., ~/worktrees/web--feature--auth).
func Resolve(branch string, repoRoot string, worktreeBase string) string {
	repo := RepoName(repoRoot)
	slug := SlugifyBranch(branch)
	dirName := repo + "--" + slug

	if worktreeBase != "" {
		return filepath.Join(worktreeBase, dirName)
	}

	parent := filepath.Dir(repoRoot)
	return filepath.Join(parent, dirName)
}

// RepoName extracts the repository name from its root path.
func RepoName(repoRoot string) string {
	return filepath.Base(repoRoot)
}
