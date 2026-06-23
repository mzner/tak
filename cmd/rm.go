package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/mzner/tak/internal/config"
	"github.com/mzner/tak/internal/hooks"
	"github.com/mzner/tak/internal/paths"
	"github.com/mzner/tak/internal/runner"
	"github.com/mzner/tak/internal/state"
	"github.com/mzner/tak/internal/tmux"
	"github.com/mzner/tak/internal/worktree"
	"github.com/spf13/cobra"
)

var rmForce bool

var rmCmd = &cobra.Command{
	Use:   "rm [branch]",
	Short: "Remove a worktree",
	Long: `Remove a git worktree for the specified branch.

If no branch is specified, shows an interactive picker.
Refuses to remove pinned worktrees (use tak unpin first).
Refuses to remove dirty worktrees (use -F/--force to override).`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r := runner.NewExecRunner()
		wtSvc := worktree.NewService(r)
		tmuxSvc := tmux.NewService(r)

		repoRoot, err := wtSvc.RepoRoot()
		if err != nil {
			return errNotInRepo
		}

		// Pin git commands to the repo root. `rm` may delete the worktree it
		// was invoked from, which would otherwise leave later git calls running
		// in a CWD that no longer exists — silently failing the branch-keep
		// checks and stranding the branch.
		wtSvc = wtSvc.WithDir(repoRoot)

		cfg, err := config.Load(repoRoot, "")
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		var branches []string
		if len(args) > 0 {
			branches = args
		} else {
			branches, err = selectWorktrees(wtSvc, "Select worktree(s) to remove:")
			if err != nil {
				return err
			}
		}

		fmt.Fprintf(os.Stderr, "Removing branch...\n")

		takDir := filepath.Join(repoRoot, ".tak")
		statePath := state.StatePath(takDir)
		st, _ := state.Load(statePath)
		defaultBranch := wtSvc.DefaultBranch()

		entries, _ := wtSvc.List()

		for _, branch := range branches {
			if cfg.IsPinned(branch) {
				fmt.Fprintf(os.Stderr, "skipping %s: pinned (run `tak unpin %s` first)\n", branch, branch)
				continue
			}

			// Find path from git worktree list (source of truth), then state, then resolve
			var wtPath string
			for _, e := range entries {
				if e.Branch == branch {
					wtPath = e.Path
					break
				}
			}
			stateEntry, stateFound := state.FindByBranch(st, branch)
			if wtPath == "" {
				if stateFound {
					wtPath = stateEntry.Path
				} else {
					wtPath = paths.Resolve(branch, repoRoot, cfg.WorktreeBase)
				}
			}

			if !rmForce {
				dirty, err := wtSvc.IsDirty(wtPath)
				if err == nil && dirty {
					fmt.Fprintf(os.Stderr, "skipping %s: uncommitted changes (use -F to force)\n", branch)
					continue
				}
			}

			windowName := paths.TmuxSlug(branch)
			_ = tmuxSvc.CloseWindow(windowName)

			// Find what branch is actually checked out at this path (may differ from requested branch)
			var checkedOutBranch string
			for _, e := range entries {
				if e.Path == wtPath {
					checkedOutBranch = e.Branch
					break
				}
			}

			// Run pre_remove hooks (abort this worktree on failure)
			if len(cfg.Hooks.PreRemove) > 0 {
				hookCtx := hooks.Context{
					WorktreeName: filepath.Base(wtPath),
					SourceDir:    repoRoot,
					TargetDir:    wtPath,
					Branch:       branch,
					Hook:         "pre_remove",
				}
				actions := toHookActions(cfg.Hooks.PreRemove)
				if err := hooks.Run(actions, repoRoot, wtPath, hookCtx); err != nil {
					fmt.Fprintf(os.Stderr, "skipping %s: pre_remove hook failed: %s\n", branch, err)
					continue
				}
			}

			if err := wtSvc.Remove(wtPath, rmForce); err != nil {
				fmt.Fprintf(os.Stderr, "error removing %s: %s\n", branch, err)
				continue
			}

			// Run post_remove hooks
			if len(cfg.Hooks.PostRemove) > 0 {
				hookCtx := hooks.Context{
					WorktreeName: filepath.Base(wtPath),
					SourceDir:    repoRoot,
					TargetDir:    wtPath,
					Branch:       branch,
					Hook:         "post_remove",
				}
				actions := toHookActions(cfg.Hooks.PostRemove)
				if err := hooks.Run(actions, repoRoot, repoRoot, hookCtx); err != nil {
					fmt.Fprintf(os.Stderr, "warning: post_remove hook failed: %s\n", err)
				}
			}

			// Delete the requested branch (skip if unpushed or has unmerged commits, unless -F)
			deleteBranch := rmForce
			if !deleteBranch {
				compareBranch := defaultBranch
				if stateFound && stateEntry.From != "" {
					compareBranch = stateEntry.From
				}
				hasCommits, err := wtSvc.HasCommitsAhead(branch, compareBranch)
				noLocalWork := err == nil && !hasCommits
				pushed := !wtSvc.HasUnpushedCommits(branch)
				deleteBranch = noLocalWork || pushed
			}
			if deleteBranch {
				if err := wtSvc.DeleteBranch(branch, true); err != nil {
					fmt.Fprintf(os.Stderr, "warning: could not delete branch %s: %s\n", branch, err)
				}
			} else {
				fmt.Fprintf(os.Stderr, "warning: %s has unpushed commits, branch kept (use -F to force delete)\n", branch)
			}

			// Also delete the checked-out branch if it differs (user switched branches inside the worktree)
			if checkedOutBranch != "" && checkedOutBranch != branch && checkedOutBranch != defaultBranch {
				_ = wtSvc.DeleteBranch(checkedOutBranch, true)
			}

			// Also delete the original branch tracked in state for this path
			// (handles case where user selected the current branch from picker but original differs)
			for _, w := range st.Worktrees {
				if w.Path == wtPath && w.Branch != branch && w.Branch != defaultBranch {
					_ = wtSvc.DeleteBranch(w.Branch, true)
					break
				}
			}

			state.Untrack(st, branch)
			// Also untrack by path in case state has it under a different branch name
			for _, w := range st.Worktrees {
				if w.Path == wtPath {
					state.Untrack(st, w.Branch)
					break
				}
			}
			fmt.Printf("Removed worktree %s\n", branch)
		}

		if err := state.Save(statePath, st); err != nil {
			return err
		}
		return nil
	},
}

func selectWorktree(wtSvc *worktree.Service, title string) (string, error) {
	options, err := allWorktreeOptions(wtSvc)
	if err != nil {
		return "", err
	}
	if len(options) == 0 {
		return "", fmt.Errorf("no worktrees to select")
	}

	var selected string
	selectField := huh.NewSelect[string]().
		Title(title).
		Options(options...).
		Value(&selected)

	err = huh.NewForm(huh.NewGroup(selectField)).
		WithOutput(os.Stderr).
		Run()
	if err != nil {
		return "", err
	}

	return selected, nil
}

func selectWorktrees(wtSvc *worktree.Service, title string) ([]string, error) {
	options, err := removableWorktreeOptions(wtSvc)
	if err != nil {
		return nil, err
	}
	if len(options) == 0 {
		return nil, fmt.Errorf("no worktrees to select")
	}

	var selected []string
	multiField := huh.NewMultiSelect[string]().
		Title(title).
		Options(options...).
		Value(&selected)

	err = huh.NewForm(huh.NewGroup(multiField)).
		WithOutput(os.Stderr).
		Run()
	if err != nil {
		return nil, err
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("no worktrees selected")
	}

	return selected, nil
}

func allWorktreeOptions(wtSvc *worktree.Service) ([]huh.Option[string], error) {
	entries, err := wtSvc.List()
	if err != nil {
		return nil, err
	}

	var options []huh.Option[string]
	for _, e := range entries {
		if e.Branch == "(detached)" {
			continue
		}
		options = append(options, huh.NewOption(e.Branch, e.Branch))
	}
	return options, nil
}

func removableWorktreeOptions(wtSvc *worktree.Service) ([]huh.Option[string], error) {
	entries, err := wtSvc.List()
	if err != nil {
		return nil, err
	}

	defaultBranch := wtSvc.DefaultBranch()
	var options []huh.Option[string]
	for i, e := range entries {
		// Skip the main worktree (always first) and detached entries
		if i == 0 || e.Branch == defaultBranch || e.Branch == "(detached)" {
			continue
		}
		options = append(options, huh.NewOption(e.Branch, e.Branch))
	}
	return options, nil
}

func completeRemovableBranches(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	r := runner.NewExecRunner()
	wtSvc := worktree.NewService(r)
	entries, err := wtSvc.List()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defaultBranch := wtSvc.DefaultBranch()
	var branches []string
	for i, e := range entries {
		if i == 0 || e.Branch == defaultBranch || e.Branch == "(detached)" {
			continue
		}
		branches = append(branches, e.Branch)
	}
	return branches, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rmCmd.Flags().BoolVarP(&rmForce, "force", "F", false, "remove even with uncommitted changes")
	rmCmd.ValidArgsFunction = completeRemovableBranches
	rootCmd.AddCommand(rmCmd)
}
