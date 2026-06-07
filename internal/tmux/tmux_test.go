package tmux

import (
	"testing"

	"github.com/mzner/tak/internal/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasWindow_Found(t *testing.T) {
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"tmux list-windows": {Output: []byte("main\nfeature-auth\nfix-bug\n")},
	})
	svc := NewService(fake)

	has, err := svc.HasWindow("feature-auth")
	require.NoError(t, err)
	assert.True(t, has)
}

func TestHasWindow_NotFound(t *testing.T) {
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"tmux list-windows": {Output: []byte("main\nfix-bug\n")},
	})
	svc := NewService(fake)

	has, err := svc.HasWindow("feature-auth")
	require.NoError(t, err)
	assert.False(t, has)
}

func TestOpenWindow_NewWindow(t *testing.T) {
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"tmux list-windows": {Output: []byte("main\n")},
	})
	svc := NewService(fake)

	err := svc.OpenWindow("feature-auth", "/tmp/worktree")
	require.NoError(t, err)

	assert.Equal(t, 2, fake.CallCount())
	call := fake.Calls[1]
	assert.Equal(t, "tmux", call.Name)
	assert.Equal(t, []string{"new-window", "-n", "feature-auth", "-c", "/tmp/worktree"}, call.Args)
}

func TestOpenWindow_ExistingWindow(t *testing.T) {
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"tmux list-windows": {Output: []byte("main\nfeature-auth\n")},
	})
	svc := NewService(fake)

	err := svc.OpenWindow("feature-auth", "/tmp/worktree")
	require.NoError(t, err)

	assert.Equal(t, 2, fake.CallCount())
	call := fake.Calls[1]
	assert.Equal(t, "tmux", call.Name)
	assert.Equal(t, []string{"select-window", "-t", "feature-auth"}, call.Args)
}

func TestCloseWindow(t *testing.T) {
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"tmux list-windows": {Output: []byte("feature-auth\n")},
	})
	svc := NewService(fake)

	err := svc.CloseWindow("feature-auth")
	require.NoError(t, err)

	call := fake.Calls[1]
	assert.Equal(t, []string{"kill-window", "-t", "feature-auth"}, call.Args)
}

func TestCloseWindow_NotExists(t *testing.T) {
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"tmux list-windows": {Output: []byte("main\n")},
	})
	svc := NewService(fake)

	err := svc.CloseWindow("feature-auth")
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.CallCount())
}
