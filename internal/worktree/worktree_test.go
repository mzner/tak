package worktree

import (
	"fmt"
	"testing"

	"github.com/mzner/tak/internal/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd_NewBranch(t *testing.T) {
	fake := runner.NewFakeRunner(nil)
	svc := NewService(fake)

	err := svc.Add("/tmp/web--feature--auth", "feature/auth", true, "main")
	require.NoError(t, err)

	require.Equal(t, 1, fake.CallCount())
	call := fake.Calls[0]
	assert.Equal(t, "git", call.Name)
	assert.Equal(t, []string{"worktree", "add", "/tmp/web--feature--auth", "-b", "feature/auth", "main"}, call.Args)
}

func TestAdd_NewBranch_NoStartPoint(t *testing.T) {
	fake := runner.NewFakeRunner(nil)
	svc := NewService(fake)

	err := svc.Add("/tmp/web--feature--auth", "feature/auth", true, "")
	require.NoError(t, err)

	call := fake.Calls[0]
	assert.Equal(t, []string{"worktree", "add", "/tmp/web--feature--auth", "-b", "feature/auth"}, call.Args)
}

func TestAdd_ExistingBranch(t *testing.T) {
	fake := runner.NewFakeRunner(nil)
	svc := NewService(fake)

	err := svc.Add("/tmp/web--feature--auth", "feature/auth", false, "")
	require.NoError(t, err)

	call := fake.Calls[0]
	assert.Equal(t, []string{"worktree", "add", "/tmp/web--feature--auth", "feature/auth"}, call.Args)
}

func TestRemove(t *testing.T) {
	fake := runner.NewFakeRunner(nil)
	svc := NewService(fake)

	err := svc.Remove("/tmp/web--feature--auth", false)
	require.NoError(t, err)

	call := fake.Calls[0]
	assert.Equal(t, "git", call.Name)
	assert.Equal(t, []string{"worktree", "remove", "/tmp/web--feature--auth"}, call.Args)
}

func TestRemove_Force(t *testing.T) {
	fake := runner.NewFakeRunner(nil)
	svc := NewService(fake)

	err := svc.Remove("/tmp/web--feature--auth", true)
	require.NoError(t, err)

	call := fake.Calls[0]
	assert.Equal(t, []string{"worktree", "remove", "--force", "/tmp/web--feature--auth"}, call.Args)
}

func TestList(t *testing.T) {
	porcelainOutput := "worktree /Users/dev/projects/web\nHEAD abc123\nbranch refs/heads/main\n\nworktree /Users/dev/projects/web--feature--auth\nHEAD def456\nbranch refs/heads/feature/auth\n\n"
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"git worktree list": {Output: []byte(porcelainOutput)},
	})
	svc := NewService(fake)

	entries, err := svc.List()
	require.NoError(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "/Users/dev/projects/web", entries[0].Path)
	assert.Equal(t, "main", entries[0].Branch)
	assert.Equal(t, "/Users/dev/projects/web--feature--auth", entries[1].Path)
	assert.Equal(t, "feature/auth", entries[1].Branch)
}

func TestList_DetachedHead(t *testing.T) {
	porcelainOutput := "worktree /Users/dev/projects/web\nHEAD abc123\ndetached\n\n"
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"git worktree list": {Output: []byte(porcelainOutput)},
	})
	svc := NewService(fake)

	entries, err := svc.List()
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "(detached)", entries[0].Branch)
}

func TestIsDirty_Clean(t *testing.T) {
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"git -C /tmp/worktree status": {Output: []byte("")},
	})
	svc := NewService(fake)

	dirty, err := svc.IsDirty("/tmp/worktree")
	require.NoError(t, err)
	assert.False(t, dirty)
}

func TestIsDirty_Dirty(t *testing.T) {
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"git -C /tmp/worktree status": {Output: []byte(" M src/main.go\n?? new-file.txt\n")},
	})
	svc := NewService(fake)

	dirty, err := svc.IsDirty("/tmp/worktree")
	require.NoError(t, err)
	assert.True(t, dirty)
}

func TestIsMerged_AncestryLocal(t *testing.T) {
	// Branch tip is reachable from local target.
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"git merge-base --is-ancestor feature/auth main": {Output: []byte("")},
	})
	svc := NewService(fake)

	merged, err := svc.IsMerged("feature/auth", "main")
	require.NoError(t, err)
	assert.True(t, merged)
}

func TestIsMerged_AncestryRemote(t *testing.T) {
	// Not in local main but reachable from origin/main (regular remote merge).
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"git merge-base --is-ancestor feature/auth main":        {Err: fmt.Errorf("exit status 1")},
		"git merge-base --is-ancestor feature/auth origin/main": {Output: []byte("")},
	})
	svc := NewService(fake)

	merged, err := svc.IsMerged("feature/auth", "main")
	require.NoError(t, err)
	assert.True(t, merged)
}

func TestIsMerged_SquashMerge(t *testing.T) {
	// Squash-merge: ancestry check fails, but tracking ref is gone.
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"git merge-base --is-ancestor fix/thing main":        {Err: fmt.Errorf("exit status 1")},
		"git merge-base --is-ancestor fix/thing origin/main": {Err: fmt.Errorf("exit status 1")},
		"git for-each-ref --format=%(upstream:short) %(upstream:track) refs/heads/fix/thing": {
			Output: []byte("origin/fix/thing [gone]\n"),
		},
	})
	svc := NewService(fake)

	merged, err := svc.IsMerged("fix/thing", "main")
	require.NoError(t, err)
	assert.True(t, merged)
}

func TestIsMerged_ActiveBranch(t *testing.T) {
	// Branch is still open — ancestry fails and tracking ref is alive.
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"git merge-base --is-ancestor feature/wip main":        {Err: fmt.Errorf("exit status 1")},
		"git merge-base --is-ancestor feature/wip origin/main": {Err: fmt.Errorf("exit status 1")},
		"git for-each-ref --format=%(upstream:short) %(upstream:track) refs/heads/feature/wip": {
			Output: []byte("origin/feature/wip\n"),
		},
	})
	svc := NewService(fake)

	merged, err := svc.IsMerged("feature/wip", "main")
	require.NoError(t, err)
	assert.False(t, merged)
}

func TestIsMerged_LocalOnlyBranch(t *testing.T) {
	// No upstream configured — purely local branch, not flagged as merged.
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"git merge-base --is-ancestor local/exp main":        {Err: fmt.Errorf("exit status 1")},
		"git merge-base --is-ancestor local/exp origin/main": {Err: fmt.Errorf("exit status 1")},
		"git for-each-ref --format=%(upstream:short) %(upstream:track) refs/heads/local/exp": {
			Output: []byte("\n"),
		},
	})
	svc := NewService(fake)

	merged, err := svc.IsMerged("local/exp", "main")
	require.NoError(t, err)
	assert.False(t, merged)
}

func TestRepoRoot(t *testing.T) {
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"git rev-parse --path-format=absolute --git-common-dir": {Output: []byte("/Users/dev/projects/web/.git\n")},
	})
	svc := NewService(fake)

	root, err := svc.RepoRoot()
	require.NoError(t, err)
	assert.Equal(t, "/Users/dev/projects/web", root)
}

func TestRepoRoot_FromWorktree(t *testing.T) {
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"git rev-parse --path-format=absolute --git-common-dir": {Output: []byte("/Users/dev/projects/web/.git\n")},
	})
	svc := NewService(fake)

	root, err := svc.RepoRoot()
	require.NoError(t, err)
	assert.Equal(t, "/Users/dev/projects/web", root)
}

func TestCurrentBranch(t *testing.T) {
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"git rev-parse": {Output: []byte("feature/auth\n")},
	})
	svc := NewService(fake)

	branch, err := svc.CurrentBranch()
	require.NoError(t, err)
	assert.Equal(t, "feature/auth", branch)
}

func TestBranchExists(t *testing.T) {
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"git rev-parse": {Output: []byte("abc123\n")},
	})
	svc := NewService(fake)

	exists := svc.BranchExists("feature/auth")
	assert.True(t, exists)
}

func TestBranchExists_NotFound(t *testing.T) {
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"git rev-parse": {Err: assert.AnError},
	})
	svc := NewService(fake)

	exists := svc.BranchExists("nonexistent")
	assert.False(t, exists)
}
