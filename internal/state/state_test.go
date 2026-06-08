package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_ValidFile(t *testing.T) {
	s, err := Load("../../testdata/state/valid.json")
	require.NoError(t, err)
	assert.Len(t, s.Worktrees, 2)
	assert.Equal(t, "feature/auth", s.Worktrees[0].Branch)
	assert.Equal(t, "/Users/dev/projects/web--feature--auth", s.Worktrees[0].Path)
}

func TestLoad_FileNotFound(t *testing.T) {
	s, err := Load("/nonexistent/state.json")
	require.NoError(t, err)
	assert.NotNil(t, s)
	assert.Empty(t, s.Worktrees)
}

func TestSave_And_Reload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	s := &State{
		Worktrees: []WorktreeEntry{
			{
				Branch:    "feature/test",
				Path:      "/tmp/repo--feature--test",
				CreatedAt: time.Date(2026, 6, 7, 10, 0, 0, 0, time.UTC),
			},
		},
	}

	err := Save(path, s)
	require.NoError(t, err)

	reloaded, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, s.Worktrees, reloaded.Worktrees)
}

func TestTrack(t *testing.T) {
	s := &State{}
	Track(s, "feature/auth", "/tmp/web--feature--auth", "main")

	assert.Len(t, s.Worktrees, 1)
	assert.Equal(t, "feature/auth", s.Worktrees[0].Branch)
	assert.Equal(t, "/tmp/web--feature--auth", s.Worktrees[0].Path)
	assert.Equal(t, "main", s.Worktrees[0].From)
	assert.False(t, s.Worktrees[0].CreatedAt.IsZero())
}

func TestTrack_DuplicateIgnored(t *testing.T) {
	s := &State{
		Worktrees: []WorktreeEntry{
			{Branch: "feature/auth", Path: "/tmp/path"},
		},
	}
	Track(s, "feature/auth", "/tmp/path", "main")
	assert.Len(t, s.Worktrees, 1)
}

func TestUntrack(t *testing.T) {
	s := &State{
		Worktrees: []WorktreeEntry{
			{Branch: "feature/auth", Path: "/tmp/a"},
			{Branch: "fix/bug", Path: "/tmp/b"},
		},
	}
	Untrack(s, "feature/auth")
	assert.Len(t, s.Worktrees, 1)
	assert.Equal(t, "fix/bug", s.Worktrees[0].Branch)
}

func TestUntrack_NotFound(t *testing.T) {
	s := &State{
		Worktrees: []WorktreeEntry{
			{Branch: "feature/auth", Path: "/tmp/a"},
		},
	}
	Untrack(s, "nonexistent")
	assert.Len(t, s.Worktrees, 1)
}

func TestFindByBranch(t *testing.T) {
	s := &State{
		Worktrees: []WorktreeEntry{
			{Branch: "feature/auth", Path: "/tmp/a"},
			{Branch: "fix/bug", Path: "/tmp/b"},
		},
	}

	entry, found := FindByBranch(s, "feature/auth")
	assert.True(t, found)
	assert.Equal(t, "/tmp/a", entry.Path)

	_, found = FindByBranch(s, "nonexistent")
	assert.False(t, found)
}

func TestEnsureDir(t *testing.T) {
	dir := t.TempDir()
	takDir := filepath.Join(dir, ".tak")

	err := EnsureDir(takDir)
	require.NoError(t, err)

	info, err := os.Stat(takDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	err = EnsureDir(takDir)
	assert.NoError(t, err)
}
