package state

import "time"

// State holds the cached list of known worktrees.
type State struct {
	Worktrees []WorktreeEntry `json:"worktrees"`
}

// WorktreeEntry represents a single tracked worktree.
type WorktreeEntry struct {
	Branch    string    `json:"branch"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}
