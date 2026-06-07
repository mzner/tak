package cmd

import (
	"fmt"
	"os"

	"github.com/mzner/tak/internal/config"
	"github.com/mzner/tak/internal/doctor"
	"github.com/mzner/tak/internal/runner"
	"github.com/mzner/tak/internal/worktree"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Health check all worktrees",
	Long: `Run health checks on all known worktrees and report issues.

Checks performed:
  - Branch merged into main: suggests removal
  - Uncommitted changes: warns about dirty worktrees
  - Broken paths: flags worktrees whose path doesn't exist

Doctor only reports — it never removes anything.`,
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

		fmt.Printf("Checking %d worktrees...\n\n", len(entries))

		d := doctor.New(wtSvc)
		findings := d.Check(entries, cfg.Pins, "main")

		if len(findings) == 0 {
			fmt.Println("All worktrees healthy.")
			return
		}

		for _, f := range findings {
			icon := severityIcon(f.Severity)
			pinLabel := ""
			if f.Pinned {
				pinLabel = " (pinned)"
			}
			fmt.Printf("%s %-24s %s%s\n", icon, f.Branch, f.Message, pinLabel)
		}

		fmt.Printf("\n%d issue(s) found.", len(findings))

		hasMerged := false
		for _, f := range findings {
			if f.Check == doctor.CheckMerged {
				hasMerged = true
				break
			}
		}
		if hasMerged {
			fmt.Print(" Run `tak gc --merged` to clean up merged branches.")
		}
		fmt.Println()
	},
}

func severityIcon(s doctor.Severity) string {
	switch s {
	case doctor.SeverityError:
		return "✗"
	case doctor.SeverityWarning:
		return "⚠"
	case doctor.SeverityInfo:
		return "ℹ"
	default:
		return " "
	}
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
