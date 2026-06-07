// Package tmux manages tmux windows for worktrees.
//
// It provides operations to create, switch to, and close tmux windows,
// as well as check whether the user is currently inside a tmux session.
// All tmux interaction goes through the runner.CommandRunner interface.
//
// Window naming: branch names are slugified with single dashes
// (e.g., "feature/auth" → "feature-auth") for tmux compatibility.
//
// Example:
//
//	svc := tmux.NewService(runner.NewExecRunner())
//	if !svc.IsInsideTmux() {
//	    return errors.New("not in a tmux session")
//	}
//	err := svc.OpenWindow("feature-auth", "/path/to/worktree")
package tmux
