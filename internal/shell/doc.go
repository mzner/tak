// Package shell generates shell hook functions for tak cd integration.
//
// Since a child process (the tak binary) cannot change the parent shell's
// working directory, tak provides a shell function that wraps the `tak cd`
// subcommand. The function intercepts `tak cd`, gets the path from the
// binary, and performs the actual `cd` in the shell.
//
// Supported shells: zsh, bash, fish
//
// Usage: add to shell rc file:
//
//	eval "$(tak shell-init zsh)"
package shell
