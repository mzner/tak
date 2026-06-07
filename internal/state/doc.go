// Package state manages the .tak/state.json cache file.
//
// The state file tracks known worktrees with their paths and creation
// timestamps. It is a cache — if deleted, tak rebuilds it from
// `git worktree list --porcelain`.
//
// Pins are NOT stored here (they live in .tak.yml config).
// The state file only tracks what worktrees exist and when they were created.
//
// File location: <repo-root>/.tak/state.json
package state
