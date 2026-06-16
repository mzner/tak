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
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := args[0]
		r := runner.NewExecRunner()
		wtSvc := worktree.NewService(r)
		tmuxSvc := tmux.NewService(r)

		repoRoot, err := wtSvc.RepoRoot()
		if err != nil {
			return errNotInRepo
		}

		cfg, err := config.Load(repoRoot, "")
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
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
				return fmt.Errorf("'%s' already has a worktree at %s\n\n  Use `tak cd %s` or `tak open %s` to switch to it", branch, e.Path, branch, branch)
			}
		}

		// Determine if branch is new or existing
		newBranch := !wtSvc.BranchExists(branch)

		if !newBranch && addFrom != "" {
			return fmt.Errorf("branch '%s' already exists, --from is ignored for existing branches\n  To recreate from %s: tak rm %s && tak add %s -f %s", branch, addFrom, branch, branch, addFrom)
		}

		// Resolve start point for new branches
		startPoint := addFrom
		if newBranch && startPoint == "" {
			startPoint = wtSvc.DefaultBranch()
			if !wtSvc.BranchExists(startPoint) {
				return fmt.Errorf("repository has no commits yet\n\n  Create an initial commit first:\n    git commit --allow-empty -m \"initial\"")
			}
		}

		// Run pre_create hooks (abort on failure)
		if len(cfg.Hooks.PreCreate) > 0 {
			fmt.Fprintf(os.Stderr, "Running pre_create hooks...\n")
			hookCtx := hooks.Context{
				WorktreeName: filepath.Base(wtPath),
				SourceDir:    repoRoot,
				TargetDir:    wtPath,
				Branch:       branch,
				Hook:         "pre_create",
			}
			actions := toHookActions(cfg.Hooks.PreCreate)
			if err := hooks.Run(actions, repoRoot, repoRoot, hookCtx); err != nil {
				return fmt.Errorf("pre_create hook failed: %w", err)
			}
		}

		// Create worktree
		fmt.Fprintf(os.Stderr, "Creating worktree %s...\n", branch)
		if err := wtSvc.Add(wtPath, branch, newBranch, startPoint); err != nil {
			return err
		}

		// Track in state
		takDir := filepath.Join(repoRoot, ".tak")
		if err := state.EnsureDir(takDir); err != nil {
			return err
		}
		statePath := state.StatePath(takDir)
		st, _ := state.Load(statePath)
		state.Track(st, branch, wtPath, startPoint)
		if err := state.Save(statePath, st); err != nil {
			return err
		}

		// Pin if requested
		if addPin {
			if err := cfg.AddPin(branch); err != nil {
				fmt.Fprintln(os.Stderr, "warning: could not save pin:", err)
			}
		}

		// Run post_create hooks
		if len(cfg.Hooks.PostCreate) > 0 {
			fmt.Fprintf(os.Stderr, "Running post_create hooks...\n")
			hookCtx := hooks.Context{
				WorktreeName: filepath.Base(wtPath),
				SourceDir:    repoRoot,
				TargetDir:    wtPath,
				Branch:       branch,
				Hook:         "post_create",
			}
			actions := toHookActions(cfg.Hooks.PostCreate)
			if err := hooks.Run(actions, repoRoot, wtPath, hookCtx); err != nil {
				fmt.Fprintf(os.Stderr, "warning: post_create hook failed: %s\n", err)
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
				return nil
			}
			if !tmuxSvc.IsInsideTmux() {
				fmt.Fprintln(os.Stderr, "warning: not in a tmux session, skipping -o")
				return nil
			}
			windowName := paths.TmuxSlug(branch)
			if err := openTmuxWindow(tmuxSvc, cfg, windowName, wtPath); err != nil {
				fmt.Fprintln(os.Stderr, "warning: could not open tmux window:", err)
			}
		}
		return nil
	},
}

func hasPrefix(branch string, prefix string) bool {
	return len(branch) > len(prefix) && branch[:len(prefix)] == prefix
}

func toHookActions(cfgActions []config.HookAction) []hooks.Action {
	actions := make([]hooks.Action, len(cfgActions))
	for i, h := range cfgActions {
		actions[i] = hooks.Action{
			Type:    h.Type,
			From:    h.From,
			To:      h.To,
			Command: h.Command,
			Env:     h.Env,
			WorkDir: h.WorkDir,
		}
	}
	return actions
}

func init() {
	addCmd.Flags().BoolVarP(&addOpen, "open", "o", false, "open a tmux window for the worktree")
	addCmd.Flags().BoolVarP(&addPin, "pin", "p", false, "pin the worktree (exclude from gc)")
	addCmd.Flags().StringVarP(&addFrom, "from", "f", "", "base branch or commit for new branches (default: main)")
	rootCmd.AddCommand(addCmd)
}
