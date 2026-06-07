package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecRunner_Run_Success(t *testing.T) {
	r := NewExecRunner()
	output, err := r.Run("echo", "hello")
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(output))
}

func TestExecRunner_Run_Failure(t *testing.T) {
	r := NewExecRunner()
	_, err := r.Run("git", "status", "--nonexistent-flag-xyz")
	assert.Error(t, err)
}

func TestExecRunner_RunInDir(t *testing.T) {
	r := NewExecRunner()
	output, err := r.RunInDir("/tmp", "pwd")
	require.NoError(t, err)
	assert.Contains(t, string(output), "tmp")
}

func TestFakeRunner_RecordsCalls(t *testing.T) {
	fake := NewFakeRunner(nil)

	_, _ = fake.Run("git", "worktree", "list")
	_, _ = fake.RunInDir("/some/path", "git", "status")

	assert.Equal(t, 2, fake.CallCount())
	assert.Equal(t, "git", fake.Calls[0].Name)
	assert.Equal(t, []string{"worktree", "list"}, fake.Calls[0].Args)
	assert.Equal(t, "/some/path", fake.Calls[1].Dir)
}

func TestFakeRunner_ReturnsPresetResponse(t *testing.T) {
	fake := NewFakeRunner(map[string]Response{
		"git worktree list": {Output: []byte("worktree1\nworktree2\n")},
		"git status":        {Err: assert.AnError},
	})

	output, err := fake.Run("git", "worktree", "list")
	require.NoError(t, err)
	assert.Equal(t, "worktree1\nworktree2\n", string(output))

	_, err = fake.Run("git", "status")
	assert.Error(t, err)
}

func TestFakeRunner_UnknownCommandReturnsNil(t *testing.T) {
	fake := NewFakeRunner(nil)

	output, err := fake.Run("git", "whatever")
	assert.NoError(t, err)
	assert.Nil(t, output)
}

func TestCall_String(t *testing.T) {
	c := Call{Name: "git", Args: []string{"worktree", "add"}}
	assert.Equal(t, "git worktree add", c.String())

	c = Call{Dir: "/tmp", Name: "git", Args: []string{"status"}}
	assert.Equal(t, "[/tmp] git status", c.String())
}
