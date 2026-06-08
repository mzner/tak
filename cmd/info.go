package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mzner/tak/internal/config"
	"github.com/mzner/tak/internal/runner"
	"github.com/mzner/tak/internal/state"
	"github.com/mzner/tak/internal/worktree"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info [branch]",
	Short: "Show details about a worktree",
	Long: `Display detailed information about a worktree.

Shows base branch, commits ahead/behind, age, pin status, and dirty files.
Without an argument, shows info for the current worktree.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r := runner.NewExecRunner()
		wtSvc := worktree.NewService(r)

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
			branch, err = wtSvc.CurrentBranch()
			if err != nil {
				fmt.Fprintln(os.Stderr, "error: specify a branch or run from within a worktree")
				os.Exit(1)
			}
		}

		defaultBranch := wtSvc.DefaultBranch()

		// Find path
		var wtPath string
		entries, _ := wtSvc.List()
		for _, e := range entries {
			if e.Branch == branch {
				wtPath = e.Path
				break
			}
		}
		if wtPath == "" {
			fmt.Fprintf(os.Stderr, "error: no worktree for branch '%s'\n", branch)
			os.Exit(1)
		}

		// Commits ahead/behind
		ahead, _ := wtSvc.CommitsAhead(branch, defaultBranch)
		behind, _ := wtSvc.CommitsBehind(branch, defaultBranch)

		// Age from state
		takDir := filepath.Join(repoRoot, ".tak")
		statePath := state.StatePath(takDir)
		st, _ := state.Load(statePath)
		entry, found := state.FindByBranch(st, branch)

		// Dirty
		dirty, _ := wtSvc.IsDirty(wtPath)

		// Print
		fmt.Printf("Branch:   %s\n", branch)
		fmt.Printf("Path:     %s\n", wtPath)
		fmt.Printf("Base:     %s\n", defaultBranch)
		fmt.Printf("Ahead:    %d commit(s)\n", ahead)
		fmt.Printf("Behind:   %d commit(s)\n", behind)
		if found && !entry.CreatedAt.IsZero() {
			fmt.Printf("Age:      %s\n", formatAge(entry.CreatedAt))
		}
		if cfg.IsPinned(branch) {
			fmt.Printf("Pinned:   yes\n")
		} else {
			fmt.Printf("Pinned:   no\n")
		}
		if dirty {
			fmt.Printf("Dirty:    yes\n")
		} else {
			fmt.Printf("Dirty:    no\n")
		}
	},
}

func formatAge(created time.Time) string {
	d := time.Since(created)
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
