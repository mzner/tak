package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
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
	Use:   "rm [branch]",
	Short: "Remove a worktree",
	Long: `Remove a git worktree for the specified branch.

If no branch is specified, shows an interactive picker.
Refuses to remove pinned worktrees (use tak unpin first).
Refuses to remove dirty worktrees (use --force to override).`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
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

		var branches []string
		if len(args) > 0 {
			branches = args
		} else {
			branches, err = selectWorktrees(wtSvc, "Select worktree(s) to remove:")
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
		}

		takDir := filepath.Join(repoRoot, ".tak")
		statePath := state.StatePath(takDir)
		st, _ := state.Load(statePath)
		defaultBranch := wtSvc.DefaultBranch()

		for _, branch := range branches {
			if cfg.IsPinned(branch) {
				fmt.Fprintf(os.Stderr, "skipping %s: pinned (run `tak unpin %s` first)\n", branch, branch)
				continue
			}

			entry, found := state.FindByBranch(st, branch)
			var wtPath string
			if found {
				wtPath = entry.Path
			} else {
				wtPath = paths.Resolve(branch, repoRoot, cfg.WorktreeBase)
			}

			if !rmForce {
				dirty, err := wtSvc.IsDirty(wtPath)
				if err == nil && dirty {
					fmt.Fprintf(os.Stderr, "skipping %s: uncommitted changes (use --force)\n", branch)
					continue
				}
			}

			windowName := paths.TmuxSlug(branch)
			tmuxSvc.CloseWindow(windowName)

			if err := wtSvc.Remove(wtPath, rmForce); err != nil {
				fmt.Fprintf(os.Stderr, "error removing %s: %s\n", branch, err)
				continue
			}

			hasCommits, err := wtSvc.HasCommitsAhead(branch, defaultBranch)
			canDelete := err == nil && !hasCommits
			if canDelete || rmForce {
				if err := wtSvc.DeleteBranch(branch, true); err != nil {
					fmt.Fprintf(os.Stderr, "warning: could not delete branch %s: %s\n", branch, err)
				}
			}

			state.Untrack(st, branch)
			fmt.Printf("Removed worktree %s\n", branch)
		}

		state.Save(statePath, st)
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
	for _, e := range entries {
		if e.Branch == defaultBranch || e.Branch == "(detached)" {
			continue
		}
		options = append(options, huh.NewOption(e.Branch, e.Branch))
	}
	return options, nil
}

func init() {
	rmCmd.Flags().BoolVarP(&rmForce, "force", "F", false, "remove even with uncommitted changes")
	rootCmd.AddCommand(rmCmd)
}
