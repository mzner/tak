package worktree

import (
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
func (s *Service) Add(path string, branch string, newBranch bool) error {
	args := []string{"worktree", "add", path}
	if newBranch {
		args = append(args, "-b", branch)
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
func (s *Service) IsMerged(branch string, target string) (bool, error) {
	output, err := s.runner.Run("git", "branch", "--merged", target)
	if err != nil {
		return false, err
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		name := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "*"))
		if name == branch {
			return true, nil
		}
	}
	return false, nil
}

// RepoRoot returns the root directory of the current git repository.
func (s *Service) RepoRoot() (string, error) {
	output, err := s.runner.Run("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// CurrentBranch returns the current branch name.
func (s *Service) CurrentBranch() (string, error) {
	output, err := s.runner.Run("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// BranchExists checks if a branch exists (local or remote tracking).
func (s *Service) BranchExists(branch string) bool {
	_, err := s.runner.Run("git", "rev-parse", "--verify", branch)
	return err == nil
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
