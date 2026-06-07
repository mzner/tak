// Package worktree manages git worktrees by shelling out to the git binary.
//
// It provides operations to create, remove, and list worktrees, as well as
// check their status (dirty/clean, branch merged). All git interaction goes
// through the runner.CommandRunner interface for testability.
//
// This package does NOT manage tak state (.tak/state.json) or config
// (.tak.yml). It only talks to git. The cmd layer coordinates between
// this package and the state/config packages.
//
// Example:
//
//	svc := worktree.NewService(runner.NewExecRunner())
//	err := svc.Add("/path/to/worktree", "feature/auth", true)
//	entries, err := svc.List()
package worktree
