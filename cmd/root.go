package cmd

import (
	"github.com/spf13/cobra"
)

var (
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tak",
	Short: "Terminal multiplexer and worktree manager",
	Long: `tak is a CLI tool for managing terminal multiplexers and worktrees.
It provides utilities for managing workspaces across multiple terminals.`,
}

// Execute executes the root command with the given version string.
func Execute(version string) error {
	rootCmd.Version = version
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable verbose output")
}
