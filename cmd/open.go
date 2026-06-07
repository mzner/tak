package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mzner/tak/internal/config"
	"github.com/mzner/tak/internal/paths"
	"github.com/mzner/tak/internal/runner"
	"github.com/mzner/tak/internal/state"
	"github.com/mzner/tak/internal/tmux"
	"github.com/mzner/tak/internal/worktree"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open [branch]",
	Short: "Open or switch to a tmux window for a worktree",
	Long: `Open a tmux window for the specified worktree branch.

If no branch is specified, shows an interactive picker.
If a window already exists, switches to it.
If not, creates a new window cd'd into the worktree.
The worktree must already exist (use tak add -t to create and open).`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r := runner.NewExecRunner()
		wtSvc := worktree.NewService(r)
		tmuxSvc := tmux.NewService(r)

		if !tmuxSvc.IsInsideTmux() {
			fmt.Fprintln(os.Stderr, "error: not in a tmux session, start one with `tmux` first")
			os.Exit(1)
		}

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
			branch, err = selectWorktree(wtSvc, "Select worktree to open:")
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
		}

		// Find worktree path: state → git worktree list → resolved path
		takDir := filepath.Join(repoRoot, ".tak")
		statePath := state.StatePath(takDir)
		st, _ := state.Load(statePath)

		var wtPath string
		entry, found := state.FindByBranch(st, branch)
		if found {
			wtPath = entry.Path
		} else {
			entries, _ := wtSvc.List()
			for _, e := range entries {
				if e.Branch == branch {
					wtPath = e.Path
					break
				}
			}
			if wtPath == "" {
				wtPath = paths.Resolve(branch, repoRoot, cfg.WorktreeBase)
			}
		}

		// Verify worktree exists
		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "error: no worktree for branch '%s', create one with `tak add %s -t`\n", branch, branch)
			os.Exit(1)
		}

		windowName := paths.TmuxSlug(branch)
		if err := tmuxSvc.OpenWindow(windowName, wtPath); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
}
