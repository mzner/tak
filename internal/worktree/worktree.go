package worktree

import (
	"fmt"
	"strings"

	"github.com/mzner/tak/internal/runner"
)

// Entry represents a single git worktree as reported by `git worktree list`.
type Entry struct {
	Path   string
	Branch string
}

// Service provides git worktree operations.
type Service struct {
	runner runner.CommandRunner
}

// NewService creates a Service with the given command runner.
func NewService(r runner.CommandRunner) *Service {
	return &Service{runner: r}
}

// Add creates a new worktree. If newBranch is true, it creates the branch
// with -b. If false, it checks out an existing branch.
// startPoint optionally specifies the commit/branch to start from (only for new branches).
func (s *Service) Add(path string, branch string, newBranch bool, startPoint string) error {
	args := []string{"worktree", "add", path}
	if newBranch {
		args = append(args, "-b", branch)
		if startPoint != "" {
			args = append(args, startPoint)
		}
	} else {
		args = append(args, branch)
	}
	_, err := s.runner.Run("git", args...)
	return err
}

// Remove deletes a worktree. If force is true, removes even with changes.
func (s *Service) Remove(path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)
	_, err := s.runner.Run("git", args...)
	return err
}

// List returns all worktrees by parsing `git worktree list --porcelain`.
func (s *Service) List() ([]Entry, error) {
	output, err := s.runner.Run("git", "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	return parsePorcelain(string(output)), nil
}

// IsDirty checks if a worktree has uncommitted changes.
func (s *Service) IsDirty(path string) (bool, error) {
	output, err := s.runner.Run("git", "-C", path, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// IsMerged checks if a branch has been merged into the target branch.
// It uses two strategies:
//  1. Ancestry: the branch tip is reachable from origin/<target> (handles
//     regular merges where commit hashes match).
//  2. Gone tracking ref: the branch had a remote tracking ref that has since
//     been deleted (handles squash-merges and PR merges where hashes differ).
func (s *Service) IsMerged(branch string, target string) (bool, error) {
	// Strategy 1: ancestry check against local and remote target.
	if ok, _ := s.isAncestorOf(branch, target); ok {
		return true, nil
	}
	if ok, _ := s.isAncestorOf(branch, "origin/"+target); ok {
		return true, nil
	}
	// Strategy 2: tracking ref gone (squash-merge / PR-delete pattern).
	return s.trackingRefGone(branch)
}

// isAncestorOf returns true if commit is a strict ancestor of base.
func (s *Service) isAncestorOf(commit, base string) (bool, error) {
	_, err := s.runner.Run("git", "merge-base", "--is-ancestor", commit, base)
	if err != nil {
		return false, err
	}
	return true, nil
}

// trackingRefGone returns true when the branch has a configured remote
// tracking ref that no longer exists (i.e. was deleted on the remote).
// A purely local branch with no upstream configured returns false, nil.
func (s *Service) trackingRefGone(branch string) (bool, error) {
	output, err := s.runner.Run("git", "for-each-ref",
		"--format=%(upstream:short) %(upstream:track)",
		"refs/heads/"+branch)
	if err != nil {
		return false, err
	}
	return strings.Contains(string(output), "[gone]"), nil
}

// RepoRoot returns the root directory of the main working tree.
// This always resolves to the primary repo root, even when called
// from inside a linked worktree.
func (s *Service) RepoRoot() (string, error) {
	output, err := s.runner.Run("git", "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	gitDir := strings.TrimSpace(string(output))
	// For main worktree: /path/to/repo/.git → /path/to/repo
	// For linked worktree: still returns /path/to/repo/.git
	return strings.TrimSuffix(gitDir, "/.git"), nil
}

// CurrentBranch returns the current branch name.
func (s *Service) CurrentBranch() (string, error) {
	output, err := s.runner.Run("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// RenameBranch renames a local branch.
func (s *Service) RenameBranch(oldName string, newName string) error {
	_, err := s.runner.Run("git", "branch", "-m", oldName, newName)
	return err
}

// DeleteBranch deletes a local branch. If force is true, uses -D (even if unmerged).
func (s *Service) DeleteBranch(branch string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	_, err := s.runner.Run("git", "branch", flag, branch)
	return err
}

// HasCommitsAhead returns true if branch has commits not in target.
func (s *Service) HasCommitsAhead(branch string, target string) (bool, error) {
	ahead, err := s.CommitsAhead(branch, target)
	return ahead > 0, err
}

// CommitsAhead returns the number of commits in branch that are not in target.
func (s *Service) CommitsAhead(branch string, target string) (int, error) {
	output, err := s.runner.Run("git", "rev-list", "--count", target+".."+branch)
	if err != nil {
		return 0, err
	}
	count := strings.TrimSpace(string(output))
	n := 0
	_, _ = fmt.Sscanf(count, "%d", &n)
	return n, nil
}

// CommitsBehind returns the number of commits in target that are not in branch.
func (s *Service) CommitsBehind(branch string, target string) (int, error) {
	output, err := s.runner.Run("git", "rev-list", "--count", branch+".."+target)
	if err != nil {
		return 0, err
	}
	count := strings.TrimSpace(string(output))
	n := 0
	_, _ = fmt.Sscanf(count, "%d", &n)
	return n, nil
}

// MergeBase returns the common ancestor commit of two branches.
func (s *Service) MergeBase(branch string, target string) (string, error) {
	output, err := s.runner.Run("git", "merge-base", branch, target)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// Prune removes stale worktree entries from git's registry.
func (s *Service) Prune() {
	_, _ = s.runner.Run("git", "worktree", "prune")
}

// HasUnpushedCommits returns true if the branch has commits not pushed to its remote.
func (s *Service) HasUnpushedCommits(branch string) bool {
	remote := "origin/" + branch
	output, err := s.runner.Run("git", "rev-list", "--count", remote+".."+branch)
	if err != nil {
		// No remote tracking branch — check if branch has any commits at all
		return true
	}
	return strings.TrimSpace(string(output)) != "0"
}

// BranchExists checks if a branch exists (local or remote tracking).
func (s *Service) BranchExists(branch string) bool {
	_, err := s.runner.Run("git", "rev-parse", "--verify", branch)
	return err == nil
}

// DefaultBranch returns the repository's default branch (main, master, etc.)
// by checking what origin/HEAD points to, falling back to common names.
func (s *Service) DefaultBranch() string {
	output, err := s.runner.Run("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil {
		ref := strings.TrimSpace(string(output))
		parts := strings.Split(ref, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}

	// Fallback: check if main or master exists
	if s.BranchExists("main") {
		return "main"
	}
	if s.BranchExists("master") {
		return "master"
	}
	return "main"
}

// parsePorcelain parses the output of `git worktree list --porcelain`.
func parsePorcelain(output string) []Entry {
	var entries []Entry
	var current Entry

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "worktree "):
			current = Entry{Path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "branch refs/heads/"):
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		case line == "detached":
			current.Branch = "(detached)"
		case line == "":
			if current.Path != "" {
				entries = append(entries, current)
				current = Entry{}
			}
		}
	}
	return entries
}
