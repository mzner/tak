package runner

import (
	"fmt"
	"os/exec"
)

// CommandRunner abstracts command execution so tests can provide
// fake implementations that don't actually shell out.
type CommandRunner interface {
	// Run executes a command and returns its stdout output.
	// Returns an error if the command exits non-zero.
	Run(name string, args ...string) ([]byte, error)

	// RunInDir executes a command in a specific working directory.
	RunInDir(dir string, name string, args ...string) ([]byte, error)
}

// ExecRunner implements CommandRunner using os/exec.
type ExecRunner struct{}

// NewExecRunner creates a new ExecRunner.
func NewExecRunner() *ExecRunner {
	return &ExecRunner{}
}

// Run executes a command in the current working directory.
func (r *ExecRunner) Run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("%s: %s", err, string(exitErr.Stderr))
		}
		return nil, err
	}
	return output, nil
}

// RunInDir executes a command in the specified directory.
func (r *ExecRunner) RunInDir(dir string, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("%s: %s", err, string(exitErr.Stderr))
		}
		return nil, err
	}
	return output, nil
}
