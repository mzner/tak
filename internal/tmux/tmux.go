package tmux

import (
	"fmt"
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

// OpenWindowWithLayout creates a window and splits it into panes with commands.
// If the window already exists, it just switches to it.
// Commands are sent as keystrokes so the shell stays alive after they finish.
func (s *Service) OpenWindowWithLayout(name string, path string, layout string, panes []PaneSpec) error {
	exists, err := s.HasWindow(name)
	if err != nil {
		return err
	}

	if exists {
		_, err = s.runner.Run("tmux", "select-window", "-t", name)
		return err
	}

	// Create first pane
	newArgs := []string{"new-window", "-n", name, "-c", path}
	if len(panes) > 0 && panes[0].Command != "" {
		newArgs = append(newArgs, fmt.Sprintf("%s; exec $SHELL", panes[0].Command))
	}
	if _, err := s.runner.Run("tmux", newArgs...); err != nil {
		return err
	}

	// Split additional panes
	for i := 1; i < len(panes); i++ {
		splitArgs := []string{"split-window", "-t", name, "-v", "-c", path}
		if panes[i].Command != "" {
			splitArgs = append(splitArgs, fmt.Sprintf("%s; exec $SHELL", panes[i].Command))
		}
		if _, err := s.runner.Run("tmux", splitArgs...); err != nil {
			return err
		}
	}

	// Apply layout
	if layout != "" && len(panes) > 1 {
		s.runner.Run("tmux", "select-layout", "-t", name, layout)
	}

	// Select first pane
	if len(panes) > 1 {
		s.runner.Run("tmux", "select-pane", "-t", name+".0")
	}

	return nil
}

// PaneSpec describes a pane to create.
type PaneSpec struct {
	Command string
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
