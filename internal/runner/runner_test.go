package runner

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExecRunnerEcho tests that ExecRunner can execute a simple echo command.
func TestExecRunnerEcho(t *testing.T) {
	runner := &ExecRunner{}
	err := runner.Run("echo", "hello")
	assert.NoError(t, err)
}

// TestExecRunnerInvalidFlag tests that ExecRunner properly returns an error for invalid flags.
func TestExecRunnerInvalidFlag(t *testing.T) {
	runner := &ExecRunner{}
	err := runner.Run("ls", "--invalid-flag-that-does-not-exist")
	assert.Error(t, err)
}

// TestExecRunnerRunInDir tests that ExecRunner can execute a command in a specific directory.
func TestExecRunnerRunInDir(t *testing.T) {
	runner := &ExecRunner{}
	// Test with temp directory - use pwd which works everywhere
	err := runner.RunInDir("/tmp", "pwd")
	assert.NoError(t, err)
}

// TestFakeRunnerRecordsCalls tests that FakeRunner records all calls.
func TestFakeRunnerRecordsCalls(t *testing.T) {
	runner := NewFakeRunner()
	err := runner.Run("git", "status")
	require.NoError(t, err)
	assert.Len(t, runner.Calls, 1)
	assert.Equal(t, "git", runner.Calls[0].Name)
	assert.Equal(t, []string{"status"}, runner.Calls[0].Args)
	assert.Equal(t, "", runner.Calls[0].Dir)
}

// TestFakeRunnerRecordsCallsWithDir tests that FakeRunner records calls with directory.
func TestFakeRunnerRecordsCallsWithDir(t *testing.T) {
	runner := NewFakeRunner()
	err := runner.RunInDir("/tmp", "make", "build")
	require.NoError(t, err)
	assert.Len(t, runner.Calls, 1)
	assert.Equal(t, "/tmp", runner.Calls[0].Dir)
	assert.Equal(t, "make", runner.Calls[0].Name)
	assert.Equal(t, []string{"build"}, runner.Calls[0].Args)
}

// TestFakeRunnerReturnsPresets tests that FakeRunner returns preset responses.
func TestFakeRunnerReturnsPresets(t *testing.T) {
	runner := NewFakeRunner()
	testErr := errors.New("test error")
	runner.SetResponse("git status", &Response{Error: testErr})

	err := runner.Run("git", "status")
	assert.Equal(t, testErr, err)
}

// TestFakeRunnerUnknownReturnsNil tests that FakeRunner returns nil for unknown commands.
func TestFakeRunnerUnknownReturnsNil(t *testing.T) {
	runner := NewFakeRunner()
	err := runner.Run("unknown", "command")
	assert.NoError(t, err)
}

// TestFakeRunnerPresetByCommand tests that FakeRunner can match on command name alone.
func TestFakeRunnerPresetByCommand(t *testing.T) {
	runner := NewFakeRunner()
	testErr := errors.New("git error")
	runner.SetResponse("git", &Response{Error: testErr})

	err := runner.Run("git", "status")
	assert.Equal(t, testErr, err)
}

// TestFakeRunnerPresetByDir tests that FakeRunner can match on directory.
func TestFakeRunnerPresetByDir(t *testing.T) {
	runner := NewFakeRunner()
	testErr := errors.New("dir error")
	runner.SetResponse("/tmp", &Response{Error: testErr})

	err := runner.RunInDir("/tmp", "make", "build")
	assert.Equal(t, testErr, err)
}

// TestCallString tests the Call.String() method.
func TestCallString(t *testing.T) {
	tests := []struct {
		name     string
		call     *Call
		expected string
	}{
		{
			name:     "simple command",
			call:     &Call{Dir: "", Name: "git", Args: []string{"status"}},
			expected: "git status",
		},
		{
			name:     "command with multiple args",
			call:     &Call{Dir: "", Name: "git", Args: []string{"commit", "-m", "message"}},
			expected: "git commit -m message",
		},
		{
			name:     "command with no args",
			call:     &Call{Dir: "", Name: "git", Args: []string{}},
			expected: "git",
		},
		{
			name:     "with directory",
			call:     &Call{Dir: "/tmp", Name: "make", Args: []string{"build"}},
			expected: "/tmp|make build",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.call.String())
		})
	}
}
