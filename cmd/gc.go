package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mzner/tak/internal/config"
	"github.com/mzner/tak/internal/doctor"
	"github.com/mzner/tak/internal/paths"
	"github.com/mzner/tak/internal/runner"
	"github.com/mzner/tak/internal/state"
	"github.com/mzner/tak/internal/tmux"
	"github.com/mzner/tak/internal/worktree"
	"github.com/spf13/cobra"
)

var (
	gcMerged bool
	gcDryRun bool
)

var gcCmd = &cobra.Command{
	Use:   "gc",
	Short: "Remove stale and merged worktrees",
	Long: `Garbage-collect worktrees that are no longer needed.

By default, removes only broken worktrees (path missing).
With --merged, also removes worktrees whose branch is merged.
Always skips pinned worktrees.`,
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
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		entries, err := wtSvc.List()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		d := doctor.New(wtSvc)
		findings := d.Check(entries, cfg.Pins, wtSvc.DefaultBranch())

		var toRemove []doctor.Finding
		var skipped []doctor.Finding

		for _, f := range findings {
			if f.Pinned {
				skipped = append(skipped, f)
				continue
			}
			switch f.Check {
			case doctor.CheckBroken:
				toRemove = append(toRemove, f)
			case doctor.CheckMerged:
				if gcMerged {
					toRemove = append(toRemove, f)
				}
			}
		}

		if len(toRemove) == 0 {
			fmt.Println("Nothing to clean up.")
			return
		}

		if gcDryRun {
			fmt.Println("Would remove:")
			for _, f := range toRemove {
				fmt.Printf("  %-24s (%s)\n", f.Branch, f.Message)
			}
			if len(skipped) > 0 {
				fmt.Println("\nSkipped (pinned):")
				for _, f := range skipped {
					fmt.Printf("  %s\n", f.Branch)
				}
			}
			fmt.Println("\nRun without --dry-run to remove.")
			return
		}

		// Perform removals
		takDir := filepath.Join(repoRoot, ".tak")
		statePath := state.StatePath(takDir)
		st, _ := state.Load(statePath)

		removed := 0
		for _, f := range toRemove {
			windowName := paths.TmuxSlug(f.Branch)
			_ = tmuxSvc.CloseWindow(windowName)

			if f.Check != doctor.CheckBroken {
				if err := wtSvc.Remove(f.Path, true); err != nil {
					fmt.Fprintf(os.Stderr, "warning: could not remove %s: %s\n", f.Branch, err)
					continue
				}
			}

			_ = wtSvc.DeleteBranch(f.Branch, true)
			state.Untrack(st, f.Branch)
			removed++
			fmt.Printf("Removed %s (%s)\n", f.Branch, f.Message)
		}

		// Sync state: add worktrees git knows about but state doesn't
		for _, entry := range entries {
			if entry.Branch == wtSvc.DefaultBranch() || entry.Branch == "(detached)" {
				continue
			}
			if _, found := state.FindByBranch(st, entry.Branch); !found {
				state.Track(st, entry.Branch, entry.Path, "")
			}
		}

		if err := state.Save(statePath, st); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		// Prune git's worktree registry for paths that no longer exist
		wtSvc.Prune()

		if len(skipped) > 0 {
			fmt.Printf("\nSkipped %d pinned worktree(s).\n", len(skipped))
		}
		fmt.Printf("\nCleaned up %d worktree(s).\n", removed)
	},
}

func init() {
	gcCmd.Flags().BoolVarP(&gcMerged, "merged", "m", false, "remove worktrees whose branch is merged into main")
	gcCmd.Flags().BoolVarP(&gcDryRun, "dry-run", "n", false, "show what would be removed without acting")
	rootCmd.AddCommand(gcCmd)
}
