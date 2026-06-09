package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mzner/tak/internal/runner"
	"github.com/mzner/tak/internal/worktree"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show configuration file paths and contents",
	Long: `Display the active tak configuration.

Shows global and local config file locations and their contents.
Useful for debugging which settings are in effect.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		globalPath := filepath.Join(home, ".config", "tak", "config.yml")
		fmt.Printf("Global: %s\n", globalPath)
		printFileContents(globalPath)

		r := runner.NewExecRunner()
		wtSvc := worktree.NewService(r)
		repoRoot, err := wtSvc.RepoRoot()
		if err != nil {
			fmt.Println("\nLocal:  (not in a git repository)")
			return nil
		}

		localPath := filepath.Join(repoRoot, ".tak.yml")
		fmt.Printf("\nLocal:  %s\n", localPath)
		printFileContents(localPath)
		return nil
	},
}

func printFileContents(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("  (not found)")
		return
	}
	fmt.Println(string(data))
}

func init() {
	rootCmd.AddCommand(configCmd)
}
