package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mzner/tak/internal/config"
	"github.com/mzner/tak/internal/paths"
	"github.com/mzner/tak/internal/runner"
	"github.com/mzner/tak/internal/state"
	"github.com/mzner/tak/internal/worktree"
	"github.com/spf13/cobra"
)

var cdCmd = &cobra.Command{
	Use:   "cd [branch]",
	Short: "Print the worktree path (use with shell hook for actual cd)",
	Long: `Print the filesystem path of the worktree for the given branch.

If no branch is specified, shows an interactive picker.
With the shell hook installed (eval "$(tak shell-init zsh)"),
this command changes your working directory directly.
Without the hook, use: cd $(tak cd <branch>)`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r := runner.NewExecRunner()
		wtSvc := worktree.NewService(r)

		repoRoot, err := wtSvc.RepoRoot()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error: not in a git repository")
			os.Exit(1)
		}

		cfg, err := config.Load(repoRoot, "")
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		var branch string
		if len(args) > 0 {
			branch = args[0]
		} else {
			branch, err = selectWorktree(wtSvc, "Select worktree:")
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
		}

		// Look up in state first
		takDir := filepath.Join(repoRoot, ".tak")
		statePath := state.StatePath(takDir)
		st, _ := state.Load(statePath)

		entry, found := state.FindByBranch(st, branch)
		if found {
			fmt.Println(entry.Path)
			return
		}

		// Fall back to resolving the path
		wtPath := paths.Resolve(branch, repoRoot, cfg.WorktreeBase)
		if _, err := os.Stat(wtPath); err == nil {
			fmt.Println(wtPath)
			return
		}

		// Not found
		fmt.Fprintf(os.Stderr, "error: no worktree for branch '%s'\n\n", branch)
		fmt.Fprintln(os.Stderr, "Available worktrees:")
		entries, _ := wtSvc.List()
		for _, e := range entries {
			fmt.Fprintf(os.Stderr, "  %s\n", e.Branch)
		}
		os.Exit(1)
	},
}

func init() {
	rootCmd.AddCommand(cdCmd)
}
