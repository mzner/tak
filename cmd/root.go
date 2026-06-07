package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "tak",
	Short: "Git worktree manager",
	Long:  "tak — Git worktree management with pinning, tmux integration, and lifecycle tools.",
}

// Execute runs the root command. Called from main.go.
func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Detect color from stderr so interactive pickers render correctly
	// even when stdout is captured by the shell hook (tak cd).
	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(os.Stderr))

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
