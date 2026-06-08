package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Load reads the state file from disk.
// If the file doesn't exist, returns an empty state (not an error).
func Load(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{}, nil
		}
		return nil, err
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return &State{}, nil
	}
	return &s, nil
}

// Save writes the state to disk as formatted JSON.
func Save(path string, s *State) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Track adds a worktree entry to the state.
// If the branch is already tracked, this is a no-op.
func Track(s *State, branch string, path string, from string) {
	for _, w := range s.Worktrees {
		if w.Branch == branch {
			return
		}
	}
	s.Worktrees = append(s.Worktrees, WorktreeEntry{
		Branch:    branch,
		Path:      path,
		CreatedAt: time.Now().UTC(),
		From:      from,
	})
}

// Untrack removes a worktree entry from the state by branch name.
func Untrack(s *State, branch string) {
	for i, w := range s.Worktrees {
		if w.Branch == branch {
			s.Worktrees = append(s.Worktrees[:i], s.Worktrees[i+1:]...)
			return
		}
	}
}

// FindByBranch looks up a worktree entry by branch name.
func FindByBranch(s *State, branch string) (WorktreeEntry, bool) {
	for _, w := range s.Worktrees {
		if w.Branch == branch {
			return w, true
		}
	}
	return WorktreeEntry{}, false
}

// EnsureDir creates the .tak directory if it doesn't exist.
func EnsureDir(takDir string) error {
	return os.MkdirAll(takDir, 0755)
}

// StatePath returns the path to state.json given the .tak directory.
func StatePath(takDir string) string {
	return filepath.Join(takDir, "state.json")
}
