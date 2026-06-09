package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/mzner/tak/internal/config"
	"github.com/mzner/tak/internal/paths"
	"github.com/mzner/tak/internal/runner"
	"github.com/mzner/tak/internal/state"
	"github.com/mzner/tak/internal/tmux"
	"github.com/mzner/tak/internal/worktree"
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:   "rename <old-branch> <new-branch>",
	Short: "Rename a worktree's branch",
	Long: `Rename the branch of an existing worktree.

Updates the git branch name, tak state, tmux window name, and pins.`,
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: completeWorktreeBranches,
	RunE: func(cmd *cobra.Command, args []string) error {
		oldBranch := args[0]
		newBranch := args[1]

		r := runner.NewExecRunner()
		wtSvc := worktree.NewService(r)
		tmuxSvc := tmux.NewService(r)

		repoRoot, err := wtSvc.RepoRoot()
		if err != nil {
			return errNotInRepo
		}

		cfg, err := config.Load(repoRoot, "")
		if err != nil {
			return err
		}

		// Prevent renaming the default branch
		if oldBranch == wtSvc.DefaultBranch() {
			return fmt.Errorf("cannot rename the default branch")
		}

		// Verify old branch has a worktree
		entries, _ := wtSvc.List()
		found := false
		for _, e := range entries {
			if e.Branch == oldBranch {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("no worktree for branch '%s'", oldBranch)
		}

		// Check new branch name doesn't already exist
		if wtSvc.BranchExists(newBranch) {
			return fmt.Errorf("branch '%s' already exists", newBranch)
		}

		// Rename git branch
		if err := wtSvc.RenameBranch(oldBranch, newBranch); err != nil {
			return err
		}

		// Update state
		takDir := filepath.Join(repoRoot, ".tak")
		statePath := state.StatePath(takDir)
		st, _ := state.Load(statePath)
		state.Rename(st, oldBranch, newBranch)
		if err := state.Save(statePath, st); err != nil {
			return err
		}

		// Update pin if pinned
		if cfg.IsPinned(oldBranch) {
			_ = cfg.RemovePin(oldBranch)
			_ = cfg.AddPin(newBranch)
		}

		// Rename tmux window if exists
		oldWindow := paths.TmuxSlug(oldBranch)
		newWindow := paths.TmuxSlug(newBranch)
		tmuxSvc.RenameWindow(oldWindow, newWindow)

		fmt.Printf("Renamed %s → %s\n", oldBranch, newBranch)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(renameCmd)
}
