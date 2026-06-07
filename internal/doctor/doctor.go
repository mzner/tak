package doctor

import (
	"os"

	"github.com/mzner/tak/internal/worktree"
)

// Severity indicates how serious a finding is.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// CheckType identifies what kind of issue was found.
type CheckType string

const (
	CheckMerged CheckType = "merged"
	CheckBroken CheckType = "broken"
	CheckDirty  CheckType = "dirty"
)

// Finding represents a single health check result.
type Finding struct {
	Branch   string
	Path     string
	Check    CheckType
	Severity Severity
	Message  string
	Pinned   bool
}

// Doctor performs health checks on worktrees.
type Doctor struct {
	wt *worktree.Service
}

// New creates a Doctor with the given worktree service.
func New(wt *worktree.Service) *Doctor {
	return &Doctor{wt: wt}
}

// Check runs all health checks against the given worktree entries.
// It skips the main branch worktree. Pinned branches are still checked
// but marked as pinned in the finding.
func (d *Doctor) Check(entries []worktree.Entry, pins []string, mainBranch string) []Finding {
	var findings []Finding

	pinSet := make(map[string]bool, len(pins))
	for _, p := range pins {
		pinSet[p] = true
	}

	for _, entry := range entries {
		if entry.Branch == mainBranch {
			continue
		}

		isPinned := pinSet[entry.Branch]

		// Check if path exists on disk
		if _, err := os.Stat(entry.Path); os.IsNotExist(err) {
			findings = append(findings, Finding{
				Branch:   entry.Branch,
				Path:     entry.Path,
				Check:    CheckBroken,
				Severity: SeverityError,
				Message:  "path does not exist",
				Pinned:   isPinned,
			})
			continue
		}

		// Check if branch is merged
		merged, err := d.wt.IsMerged(entry.Branch, mainBranch)
		if err == nil && merged {
			findings = append(findings, Finding{
				Branch:   entry.Branch,
				Path:     entry.Path,
				Check:    CheckMerged,
				Severity: SeverityWarning,
				Message:  "branch merged into " + mainBranch,
				Pinned:   isPinned,
			})
			continue
		}

		// Check if worktree is dirty
		dirty, err := d.wt.IsDirty(entry.Path)
		if err == nil && dirty {
			findings = append(findings, Finding{
				Branch:   entry.Branch,
				Path:     entry.Path,
				Check:    CheckDirty,
				Severity: SeverityInfo,
				Message:  "uncommitted changes",
				Pinned:   isPinned,
			})
		}
	}

	return findings
}
