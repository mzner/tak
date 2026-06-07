package doctor

import (
	"testing"

	"github.com/mzner/tak/internal/runner"
	"github.com/mzner/tak/internal/worktree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheck_MergedBranch(t *testing.T) {
	tmpDir := t.TempDir()
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"git branch --merged": {Output: []byte("  feature/old\n* main\n")},
		"git -C " + tmpDir + " status": {Output: []byte("")},
	})
	wtSvc := worktree.NewService(fake)
	d := New(wtSvc)

	entries := []worktree.Entry{
		{Path: tmpDir, Branch: "feature/old"},
	}

	findings := d.Check(entries, nil, "main")
	require.Len(t, findings, 1)
	assert.Equal(t, SeverityWarning, findings[0].Severity)
	assert.Equal(t, CheckMerged, findings[0].Check)
	assert.Equal(t, "feature/old", findings[0].Branch)
}

func TestCheck_BrokenWorktree(t *testing.T) {
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"git branch --merged": {Output: []byte("* main\n")},
	})
	wtSvc := worktree.NewService(fake)
	d := New(wtSvc)

	entries := []worktree.Entry{
		{Path: "/nonexistent/path/worktree", Branch: "feature/gone"},
	}

	findings := d.Check(entries, nil, "main")
	require.Len(t, findings, 1)
	assert.Equal(t, SeverityError, findings[0].Severity)
	assert.Equal(t, CheckBroken, findings[0].Check)
}

func TestCheck_SkipsMainWorktree(t *testing.T) {
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"git branch --merged": {Output: []byte("* main\n")},
	})
	wtSvc := worktree.NewService(fake)
	d := New(wtSvc)

	entries := []worktree.Entry{
		{Path: "/tmp/web", Branch: "main"},
	}

	findings := d.Check(entries, nil, "main")
	assert.Empty(t, findings)
}

func TestCheck_SkipsPinnedBranches(t *testing.T) {
	tmpDir := t.TempDir()
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"git branch --merged": {Output: []byte("  feature/auth\n* main\n")},
		"git -C " + tmpDir + " status": {Output: []byte("")},
	})
	wtSvc := worktree.NewService(fake)
	d := New(wtSvc)

	entries := []worktree.Entry{
		{Path: tmpDir, Branch: "feature/auth"},
	}
	pins := []string{"feature/auth"}

	findings := d.Check(entries, pins, "main")
	require.Len(t, findings, 1)
	assert.True(t, findings[0].Pinned)
}

func TestCheck_AllClean(t *testing.T) {
	tmpDir := t.TempDir()
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"git branch --merged": {Output: []byte("* main\n")},
		"git -C " + tmpDir + " status": {Output: []byte("")},
	})
	wtSvc := worktree.NewService(fake)
	d := New(wtSvc)

	entries := []worktree.Entry{
		{Path: tmpDir, Branch: "feature/active"},
	}

	findings := d.Check(entries, nil, "main")
	assert.Empty(t, findings)
}
