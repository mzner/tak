package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tak",
	Short: "Git worktree manager",
	Long:  "tak — Git worktree management with pinning, tmux integration, and lifecycle tools.",
	// Errors and usage are printed by Execute, not by cobra, so a runtime
	// failure shows a single "error: ..." line instead of a usage dump.
	SilenceErrors: true,
	SilenceUsage:  true,
}

// exitError lets a command request a specific process exit code without an
// accompanying "error:" message — used by tak exec to forward the exit code
// of the command it runs, whose own output already went to the terminal.
type exitError struct {
	code int
}

func (e *exitError) Error() string {
	return fmt.Sprintf("exit code %d", e.code)
}

// Execute runs the root command. Called from main.go.
func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		if ee, ok := errors.AsType[*exitError](err); ok {
			os.Exit(ee.code)
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func init() {
	// Detect color from stderr so interactive pickers render correctly
	// even when stdout is captured by the shell hook (tak cd).
	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(os.Stderr))
}
