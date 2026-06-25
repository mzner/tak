package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/mzner/tak/internal/config"
	"github.com/mzner/tak/internal/runner"
	"github.com/mzner/tak/internal/worktree"
	"github.com/spf13/cobra"
)

var (
	lsStatus bool
	lsJSON   bool
)

var lsCmd = &cobra.Command{
	Use:   "ls [repo]",
	Short: "List all worktrees",
	Long:  "List all git worktrees with their branch, path, and pin status. Use -s to include dirty/clean state. Pass a repo name to list a different repo's worktrees.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r := runner.NewExecRunner()

		// Cross-repo: tak ls web
		if len(args) > 0 {
			listRemoteRepo(r, args[0])
			return
		}

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

		// Drop the main worktree (the repo root itself). It's the bare repo,
		// not a tak-managed branch, and including it would surface whatever
		// branch HEAD happens to point to there.
		entries = filterMainWorktree(entries, repoRoot)

		if lsJSON {
			type jsonEntry struct {
				Branch string `json:"branch"`
				Path   string `json:"path"`
				Pinned bool   `json:"pinned"`
			}
			var result []jsonEntry
			for _, e := range entries {
				result = append(result, jsonEntry{
					Branch: e.Branch,
					Path:   e.Path,
					Pinned: cfg.IsPinned(e.Branch),
				})
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
		if lsStatus {
			_, _ = fmt.Fprintln(w, "BRANCH\tPATH\tSTATUS")
		} else {
			_, _ = fmt.Fprintln(w, "BRANCH\tPATH")
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
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", entry.Branch, displayPath, strings.Join(status, "  "))
			} else {
				pinLabel := ""
				if cfg.IsPinned(entry.Branch) {
					pinLabel = " (pinned)"
				}
				_, _ = fmt.Fprintf(w, "%s\t%s%s\n", entry.Branch, displayPath, pinLabel)
			}
		}
		_ = w.Flush()
	},
}

func filterMainWorktree(entries []worktree.Entry, repoRoot string) []worktree.Entry {
	out := entries[:0]
	for _, e := range entries {
		if e.Path == repoRoot {
			continue
		}
		out = append(out, e)
	}
	return out
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

func listRemoteRepo(r *runner.ExecRunner, repoName string) {
	repos, err := config.LoadGlobal()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	repoPath, ok := repos[repoName]
	if !ok {
		available := make([]string, 0, len(repos))
		for name := range repos {
			available = append(available, name)
		}
		fmt.Fprintf(os.Stderr, "error: unknown repo '%s' (available: %s)\n", repoName, strings.Join(available, ", "))
		os.Exit(1)
	}

	output, err := r.RunInDir(repoPath, "git", "worktree", "list", "--porcelain")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	_, _ = fmt.Fprintln(w, "BRANCH\tPATH")

	var currentPath string
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "worktree ") {
			currentPath = strings.TrimPrefix(line, "worktree ")
		}
		if strings.HasPrefix(line, "branch refs/heads/") {
			branch := strings.TrimPrefix(line, "branch refs/heads/")
			_, _ = fmt.Fprintf(w, "%s\t%s\n", branch, shortenHome(currentPath))
		}
	}
	_ = w.Flush()
}

func init() {
	lsCmd.Flags().BoolVarP(&lsStatus, "status", "s", false, "include dirty/clean status (slower)")
	lsCmd.Flags().BoolVar(&lsJSON, "json", false, "output as JSON")
	rootCmd.AddCommand(lsCmd)
}
