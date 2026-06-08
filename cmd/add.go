package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mzner/tak/internal/config"
	"github.com/mzner/tak/internal/hooks"
	"github.com/mzner/tak/internal/paths"
	"github.com/mzner/tak/internal/runner"
	"github.com/mzner/tak/internal/state"
	"github.com/mzner/tak/internal/tmux"
	"github.com/mzner/tak/internal/worktree"
	"github.com/spf13/cobra"
)

var (
	addOpen bool
	addPin  bool
	addFrom string
)

var addCmd = &cobra.Command{
	Use:   "add <branch>",
	Short: "Create a new worktree",
	Long: `Create a new git worktree for the specified branch.

If the branch doesn't exist, it is created from the default branch (or --from).
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

		if !newBranch && addFrom != "" {
			fmt.Fprintf(os.Stderr, "error: branch '%s' already exists, --from is ignored for existing branches\n", branch)
			fmt.Fprintf(os.Stderr, "  To recreate from %s: tak rm %s && tak add %s -f %s\n", addFrom, branch, branch, addFrom)
			os.Exit(1)
		}

		// Resolve start point for new branches
		startPoint := addFrom
		if newBranch && startPoint == "" {
			startPoint = wtSvc.DefaultBranch()
		}

		// Create worktree
		fmt.Fprintf(os.Stderr, "Creating worktree %s...\n", branch)
		if err := wtSvc.Add(wtPath, branch, newBranch, startPoint); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		// Track in state
		takDir := filepath.Join(repoRoot, ".tak")
		state.EnsureDir(takDir)
		statePath := state.StatePath(takDir)
		st, _ := state.Load(statePath)
		state.Track(st, branch, wtPath, startPoint)
		state.Save(statePath, st)

		// Pin if requested
		if addPin {
			if err := cfg.AddPin(branch); err != nil {
				fmt.Fprintln(os.Stderr, "warning: could not save pin:", err)
			}
		}

		// Run post_create hooks
		if len(cfg.Hooks.PostCreate) > 0 {
			fmt.Fprintf(os.Stderr, "Running post_create hooks...\n")
			var actions []hooks.Action
			for _, h := range cfg.Hooks.PostCreate {
				actions = append(actions, hooks.Action{
					Type:    h.Type,
					From:    h.From,
					To:      h.To,
					Command: h.Command,
					Env:     h.Env,
					WorkDir: h.WorkDir,
				})
			}
			if err := hooks.RunPostCreate(actions, repoRoot, wtPath); err != nil {
				fmt.Fprintf(os.Stderr, "warning: hook failed: %s\n", err)
			}
		}

		// Relative path for display
		relPath, _ := filepath.Rel(filepath.Dir(repoRoot), wtPath)
		if relPath == "" {
			relPath = wtPath
		}
		fmt.Printf("Created worktree %s at %s\n", branch, relPath)

		// Open tmux window if requested
		if addOpen {
			if !tmuxSvc.IsInstalled() {
				fmt.Fprintln(os.Stderr, "warning: tmux is not installed, skipping -o")
				return
			}
			if !tmuxSvc.IsInsideTmux() {
				fmt.Fprintln(os.Stderr, "warning: not in a tmux session, skipping -o")
				return
			}
			windowName := paths.TmuxSlug(branch)
			if err := openTmuxWindow(tmuxSvc, cfg, windowName, wtPath); err != nil {
				fmt.Fprintln(os.Stderr, "warning: could not open tmux window:", err)
			}
		}
	},
}

func hasPrefix(branch string, prefix string) bool {
	return len(branch) > len(prefix) && branch[:len(prefix)] == prefix
}

func init() {
	addCmd.Flags().BoolVarP(&addOpen, "open", "o", false, "open a tmux window for the worktree")
	addCmd.Flags().BoolVarP(&addPin, "pin", "p", false, "pin the worktree (exclude from gc)")
	addCmd.Flags().StringVarP(&addFrom, "from", "f", "", "base branch or commit for new branches (default: main)")
	rootCmd.AddCommand(addCmd)
}
