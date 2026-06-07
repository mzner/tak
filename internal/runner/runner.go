package runner

import (
	"os"
	"os/exec"
)

// CommandRunner defines the interface for executing system commands.
type CommandRunner interface {
	// Run executes a command with the given name and arguments in the current directory.
	Run(name string, args ...string) error

	// RunInDir executes a command with the given name and arguments in the specified directory.
	RunInDir(dir, name string, args ...string) error
}

// ExecRunner is a CommandRunner implementation that executes real system commands.
type ExecRunner struct{}

// Run executes a command with the given name and arguments in the current directory.
func (e *ExecRunner) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// RunInDir executes a command with the given name and arguments in the specified directory.
func (e *ExecRunner) RunInDir(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
