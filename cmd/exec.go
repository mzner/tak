package cmd

import (
	"errors"
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
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: completeWorktreeBranches,
	RunE: func(cmd *cobra.Command, args []string) error {
		dashIdx := cmd.ArgsLenAtDash()
		if dashIdx < 1 || dashIdx >= len(args) {
			return fmt.Errorf("usage: tak exec <branch> -- <command>")
		}
		branch := args[0]
		cmdArgs := args[dashIdx:]

		r := runner.NewExecRunner()
		wtSvc := worktree.NewService(r)

		repoRoot, err := wtSvc.RepoRoot()
		if err != nil {
			return errNotInRepo
		}

		cfg, err := config.Load(repoRoot, "")
		if err != nil {
			return err
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
			return fmt.Errorf("no worktree for branch '%s'", branch)
		}

		// Run the command in the worktree
		c := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		c.Dir = wtPath
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin

		if err := c.Run(); err != nil {
			// Forward the child's exit code without printing an "error:" line —
			// its own output already went to the terminal.
			if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
				return &exitError{code: exitErr.ExitCode()}
			}
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
}
