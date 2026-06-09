package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		r := runner.NewExecRunner()
		wtSvc := worktree.NewService(r)

		var branch string
		if len(args) > 0 {
			branch = args[0]
		}

		// Handle repo:branch syntax (cross-repo access)
		if branch != "" && strings.Contains(branch, ":") {
			path, err := resolveRemoteWorktree(r, branch)
			if err != nil {
				return err
			}
			fmt.Println(path)
			return nil
		}

		repoRoot, err := wtSvc.RepoRoot()
		if err != nil {
			return errNotInRepo
		}

		cfg, err := config.Load(repoRoot, "")
		if err != nil {
			return err
		}

		if branch == "" {
			branch, err = selectWorktree(wtSvc, "Select worktree:")
			if err != nil {
				return err
			}
		}

		// Look up in state first
		takDir := filepath.Join(repoRoot, ".tak")
		statePath := state.StatePath(takDir)
		st, _ := state.Load(statePath)

		entry, found := state.FindByBranch(st, branch)
		if found {
			fmt.Println(entry.Path)
			return nil
		}

		// Check git worktree list (covers main and untracked worktrees)
		entries, _ := wtSvc.List()
		for _, e := range entries {
			if e.Branch == branch {
				fmt.Println(e.Path)
				return nil
			}
		}

		// Fall back to resolving the path
		wtPath := paths.Resolve(branch, repoRoot, cfg.WorktreeBase)
		if _, err := os.Stat(wtPath); err == nil {
			fmt.Println(wtPath)
			return nil
		}

		// Not found — list what is available to help the user
		var available strings.Builder
		for _, e := range entries {
			fmt.Fprintf(&available, "\n  %s", e.Branch)
		}
		return fmt.Errorf("no worktree for branch '%s'\n\nAvailable worktrees:%s", branch, available.String())
	},
}

func init() {
	cdCmd.ValidArgsFunction = completeWorktreeBranches
	rootCmd.AddCommand(cdCmd)
}
