// Package doctor provides health checks for git worktrees.
//
// It inspects all known worktrees and reports issues:
//   - Merged branches: branch has been merged into main
//   - Stale worktrees: uncommitted changes older than 7 days
//   - Broken worktrees: path no longer exists on disk
//
// Doctor only reports findings — it never takes action.
// The gc command acts on these findings.
//
// Example:
//
//	d := doctor.New(worktreeSvc)
//	findings := d.Check(entries, pins, "main")
//	for _, f := range findings {
//	    fmt.Printf("%s: %s\n", f.Severity, f.Message)
//	}
package doctor
