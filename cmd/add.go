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

var (
	addTmux bool
	addPin  bool
)

var addCmd = &cobra.Command{
	Use:   "add <branch>",
	Short: "Create a new worktree",
	Long: `Create a new git worktree for the specified branch.

If the branch doesn't exist, it is created from the current HEAD.
If the branch exists (locally or remotely), it is checked out.`,
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

		// Apply branch prefix if configured
		if cfg.BranchPrefix != "" && !hasPrefix(branch, cfg.BranchPrefix) {
			branch = cfg.BranchPrefix + branch
		}

		// Resolve worktree path
		wtPath := paths.Resolve(branch, repoRoot, cfg.WorktreeBase)

		// Check if branch already has a worktree
		entries, _ := wtSvc.List()
		for _, e := range entries {
			if e.Branch == branch {
				fmt.Fprintf(os.Stderr, "error: '%s' already has a worktree at %s\n\n", branch, e.Path)
				fmt.Fprintf(os.Stderr, "  Use `tak cd %s` or `tak open %s` to switch to it.\n", branch, branch)
				os.Exit(1)
			}
		}

		// Determine if branch is new or existing
		newBranch := !wtSvc.BranchExists(branch)

		// Create worktree
		if err := wtSvc.Add(wtPath, branch, newBranch); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		// Track in state
		takDir := filepath.Join(repoRoot, ".tak")
		state.EnsureDir(takDir)
		statePath := state.StatePath(takDir)
		st, _ := state.Load(statePath)
		state.Track(st, branch, wtPath)
		state.Save(statePath, st)

		// Pin if requested
		if addPin {
			if err := cfg.AddPin(branch); err != nil {
				fmt.Fprintln(os.Stderr, "warning: could not save pin:", err)
			}
		}

		// Relative path for display
		relPath, _ := filepath.Rel(filepath.Dir(repoRoot), wtPath)
		if relPath == "" {
			relPath = wtPath
		}
		fmt.Printf("Created worktree %s at %s\n", branch, relPath)

		// Open tmux window if requested
		if addTmux {
			if !tmuxSvc.IsInstalled() {
				fmt.Fprintln(os.Stderr, "warning: tmux is not installed, skipping -t")
				return
			}
			if !tmuxSvc.IsInsideTmux() {
				fmt.Fprintln(os.Stderr, "warning: not in a tmux session, skipping -t")
				return
			}
			windowName := paths.TmuxSlug(branch)
			if err := tmuxSvc.OpenWindow(windowName, wtPath); err != nil {
				fmt.Fprintln(os.Stderr, "warning: could not open tmux window:", err)
			}
		}
	},
}

func hasPrefix(branch string, prefix string) bool {
	return len(branch) > len(prefix) && branch[:len(prefix)] == prefix
}

func init() {
	addCmd.Flags().BoolVarP(&addTmux, "tmux", "t", false, "open a tmux window for the worktree")
	addCmd.Flags().BoolVar(&addPin, "pin", false, "pin the worktree (exclude from gc)")
	rootCmd.AddCommand(addCmd)
}
