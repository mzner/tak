package tmux

import (
	"os"
	"os/exec"
	"strings"

	"github.com/mzner/tak/internal/runner"
)

// Service provides tmux window management operations.
type Service struct {
	runner runner.CommandRunner
}

// NewService creates a Service with the given command runner.
func NewService(r runner.CommandRunner) *Service {
	return &Service{runner: r}
}

// IsInstalled returns true if the tmux binary is available on PATH.
func (s *Service) IsInstalled() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

// IsInsideTmux returns true if the current process is running inside tmux.
func (s *Service) IsInsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

// HasWindow checks if a tmux window with the given name exists.
func (s *Service) HasWindow(name string) (bool, error) {
	output, err := s.runner.Run("tmux", "list-windows", "-F", "#{window_name}")
	if err != nil {
		return false, err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == name {
			return true, nil
		}
	}
	return false, nil
}

// OpenWindow creates a new tmux window or switches to an existing one.
func (s *Service) OpenWindow(name string, path string) error {
	exists, err := s.HasWindow(name)
	if err != nil {
		return err
	}

	if exists {
		_, err = s.runner.Run("tmux", "select-window", "-t", name)
		return err
	}

	_, err = s.runner.Run("tmux", "new-window", "-n", name, "-c", path)
	return err
}

// CloseWindow kills a tmux window by name. No-op if window doesn't exist.
func (s *Service) CloseWindow(name string) error {
	exists, err := s.HasWindow(name)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	_, err = s.runner.Run("tmux", "kill-window", "-t", name)
	return err
}
