package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mzner/tak/internal/config"
	"github.com/mzner/tak/internal/runner"
	"github.com/mzner/tak/internal/state"
	"github.com/mzner/tak/internal/worktree"
	"github.com/spf13/cobra"
)

var pinCmd = &cobra.Command{
	Use:   "pin [branch]",
	Short: "Pin a worktree (exclude from gc)",
	Long: `Pin a worktree to prevent it from being removed by tak gc.

Without an argument, pins the worktree for the current working directory.
With an argument, pins the specified branch.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r := runner.NewExecRunner()
		wtSvc := worktree.NewService(r)

		repoRoot, err := wtSvc.RepoRoot()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error: not in a git repository")
			os.Exit(1)
		}

		branch, err := resolveBranchArg(args, wtSvc, repoRoot)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		cfg, err := config.Load(repoRoot, "")
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		if err := cfg.AddPin(branch); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		fmt.Printf("Pinned %s\n", branch)
	},
}

var unpinCmd = &cobra.Command{
	Use:   "unpin [branch]",
	Short: "Unpin a worktree (allow gc to remove it)",
	Long: `Remove the pin from a worktree, allowing tak gc to clean it up.

Without an argument, unpins the worktree for the current working directory.
With an argument, unpins the specified branch.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r := runner.NewExecRunner()
		wtSvc := worktree.NewService(r)

		repoRoot, err := wtSvc.RepoRoot()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error: not in a git repository")
			os.Exit(1)
		}

		branch, err := resolveBranchArg(args, wtSvc, repoRoot)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		cfg, err := config.Load(repoRoot, "")
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		if err := cfg.RemovePin(branch); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		fmt.Printf("Unpinned %s\n", branch)
	},
}

func resolveBranchArg(args []string, wtSvc *worktree.Service, repoRoot string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("specify a branch or run from within a worktree")
	}

	takDir := filepath.Join(repoRoot, ".tak")
	statePath := state.StatePath(takDir)
	st, _ := state.Load(statePath)

	for _, entry := range st.Worktrees {
		if cwd == entry.Path || isSubdir(cwd, entry.Path) {
			return entry.Branch, nil
		}
	}

	branch, err := wtSvc.CurrentBranch()
	if err != nil {
		return "", fmt.Errorf("specify a branch or run from within a worktree")
	}
	return branch, nil
}

func isSubdir(child, parent string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel != ".." && len(rel) > 0 && rel[0] != '.'
}

func init() {
	pinCmd.ValidArgsFunction = completeWorktreeBranches
	unpinCmd.ValidArgsFunction = completeWorktreeBranches
	rootCmd.AddCommand(pinCmd)
	rootCmd.AddCommand(unpinCmd)
}
