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

var rmForce bool

var rmCmd = &cobra.Command{
	Use:   "rm <branch>",
	Short: "Remove a worktree",
	Long: `Remove a git worktree for the specified branch.

Refuses to remove pinned worktrees (use tak unpin first).
Refuses to remove dirty worktrees (use --force to override).`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		branch := args[0]
		r := runner.NewExecRunner()
		wtSvc := worktree.NewService(r)
		tmuxSvc := tmux.NewService(r)

		repoRoot, err := wtSvc.RepoRoot()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error: not in a git repository")
			os.Exit(1)
		}

		cfg, err := config.Load(repoRoot, "")
		if err != nil {
			fmt.Fprintln(os.Stderr, "error: loading config:", err)
			os.Exit(1)
		}

		if cfg.IsPinned(branch) {
			fmt.Fprintf(os.Stderr, "error: worktree is pinned\n\n  Run `tak unpin %s` first, then retry.\n", branch)
			os.Exit(1)
		}

		// Find worktree path
		takDir := filepath.Join(repoRoot, ".tak")
		statePath := state.StatePath(takDir)
		st, _ := state.Load(statePath)

		entry, found := state.FindByBranch(st, branch)
		var wtPath string
		if found {
			wtPath = entry.Path
		} else {
			wtPath = paths.Resolve(branch, repoRoot, cfg.WorktreeBase)
		}

		// Check if dirty
		if !rmForce {
			dirty, err := wtSvc.IsDirty(wtPath)
			if err == nil && dirty {
				fmt.Fprintln(os.Stderr, "error: worktree has uncommitted changes, use --force to remove anyway")
				os.Exit(1)
			}
		}

		// Close tmux window if exists
		windowName := paths.TmuxSlug(branch)
		tmuxSvc.CloseWindow(windowName)

		// Remove worktree
		if err := wtSvc.Remove(wtPath, rmForce); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		// Update state
		state.Untrack(st, branch)
		state.Save(statePath, st)

		fmt.Printf("Removed worktree %s\n", branch)
	},
}

func init() {
	rmCmd.Flags().BoolVar(&rmForce, "force", false, "remove even with uncommitted changes")
	rootCmd.AddCommand(rmCmd)
}
