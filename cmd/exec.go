package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mzner/tak/internal/config"
	"github.com/mzner/tak/internal/paths"
	"github.com/mzner/tak/internal/runner"
	"github.com/mzner/tak/internal/state"
	"github.com/mzner/tak/internal/worktree"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec <branch> -- <command>",
	Short: "Run a command in a worktree",
	Long: `Run a command inside a worktree without changing your directory.

Everything after -- is passed as the command to execute.

Examples:
  tak exec feature/auth -- git pull
  tak exec feature/auth -- pnpm test
  tak exec feature/auth -- make build`,
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		// Split args at --
		var branch string
		var cmdArgs []string
		for i, a := range args {
			if a == "--" {
				branch = args[0]
				cmdArgs = args[i+1:]
				break
			}
		}
		if branch == "" || len(cmdArgs) == 0 {
			fmt.Fprintln(os.Stderr, "usage: tak exec <branch> -- <command>")
			os.Exit(1)
		}

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

		// Find worktree path
		takDir := filepath.Join(repoRoot, ".tak")
		statePath := state.StatePath(takDir)
		st, _ := state.Load(statePath)

		var wtPath string
		entry, found := state.FindByBranch(st, branch)
		if found {
			wtPath = entry.Path
		} else {
			entries, _ := wtSvc.List()
			for _, e := range entries {
				if e.Branch == branch {
					wtPath = e.Path
					break
				}
			}
			if wtPath == "" {
				wtPath = paths.Resolve(branch, repoRoot, cfg.WorktreeBase)
			}
		}

		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "error: no worktree for branch '%s'\n", branch)
			os.Exit(1)
		}

		// Run the command in the worktree
		c := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		c.Dir = wtPath
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin

		if err := c.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
}
