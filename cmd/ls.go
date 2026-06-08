package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/mzner/tak/internal/config"
	"github.com/mzner/tak/internal/runner"
	"github.com/mzner/tak/internal/worktree"
	"github.com/spf13/cobra"
)

var lsStatus bool

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all worktrees",
	Long:  "List all git worktrees with their branch, path, and pin status. Use -s to include dirty/clean state.",
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

		entries, err := wtSvc.List()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
		if lsStatus {
			fmt.Fprintln(w, "BRANCH\tPATH\tSTATUS")
		} else {
			fmt.Fprintln(w, "BRANCH\tPATH")
		}

		for _, entry := range entries {
			displayPath := shortenHome(entry.Path)

			if lsStatus {
				var status []string
				if cfg.IsPinned(entry.Branch) {
					status = append(status, "pinned")
				}
				dirty, err := wtSvc.IsDirty(entry.Path)
				if err == nil && dirty {
					status = append(status, "dirty")
				} else if err == nil {
					status = append(status, "clean")
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", entry.Branch, displayPath, strings.Join(status, "  "))
			} else {
				pinLabel := ""
				if cfg.IsPinned(entry.Branch) {
					pinLabel = " (pinned)"
				}
				fmt.Fprintf(w, "%s\t%s%s\n", entry.Branch, displayPath, pinLabel)
			}
		}
		w.Flush()
	},
}

func shortenHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}

func init() {
	lsCmd.Flags().BoolVarP(&lsStatus, "status", "s", false, "include dirty/clean status (slower)")
	rootCmd.AddCommand(lsCmd)
}
