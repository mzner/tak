package tmux

import (
	"testing"

	"github.com/mzner/tak/internal/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenWindowWithLayout_NewWindow(t *testing.T) {
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"tmux list-windows": {Output: []byte("main\n")},
	})
	svc := NewService(fake)

	panes := []PaneSpec{
		{Command: "nvim ."},
		{Command: "pnpm dev"},
		{Command: ""},
	}

	err := svc.OpenWindowWithLayout("feature-auth", "/tmp/wt", "main-vertical", panes)
	require.NoError(t, err)

	// list-windows + new-window + 2 split-windows + select-layout + select-pane
	assert.GreaterOrEqual(t, fake.CallCount(), 5)

	// First call: list-windows
	assert.Equal(t, []string{"list-windows", "-F", "#{window_name}"}, fake.Calls[0].Args)

	// Second call: new-window with first pane's command
	call := fake.Calls[1]
	assert.Equal(t, "tmux", call.Name)
	assert.Contains(t, call.Args, "new-window")
	assert.Contains(t, call.Args, "-n")
	assert.Contains(t, call.Args, "feature-auth")
	assert.Contains(t, call.Args, "-c")
	assert.Contains(t, call.Args, "/tmp/wt")
	// Should have command with exec $SHELL
	assert.Contains(t, call.Args[len(call.Args)-1], "nvim .")
	assert.Contains(t, call.Args[len(call.Args)-1], "exec $SHELL")

	// Third call: first split-window (pnpm dev)
	call = fake.Calls[2]
	assert.Contains(t, call.Args, "split-window")
	assert.Contains(t, call.Args[len(call.Args)-1], "pnpm dev")

	// Fourth call: second split-window (empty — no command arg beyond standard flags)
	call = fake.Calls[3]
	assert.Contains(t, call.Args, "split-window")
	assert.Equal(t, []string{"split-window", "-t", "feature-auth", "-v", "-c", "/tmp/wt"}, call.Args)
}

func TestOpenWindowWithLayout_ExistingWindow(t *testing.T) {
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"tmux list-windows": {Output: []byte("main\nfeature-auth\n")},
	})
	svc := NewService(fake)

	panes := []PaneSpec{{Command: "nvim ."}, {Command: ""}}

	err := svc.OpenWindowWithLayout("feature-auth", "/tmp/wt", "even-vertical", panes)
	require.NoError(t, err)

	// Should only list-windows + select-window (switch to existing)
	assert.Equal(t, 2, fake.CallCount())
	assert.Equal(t, []string{"select-window", "-t", "feature-auth"}, fake.Calls[1].Args)
}

func TestOpenWindowWithLayout_SinglePane(t *testing.T) {
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"tmux list-windows": {Output: []byte("main\n")},
	})
	svc := NewService(fake)

	panes := []PaneSpec{{Command: "nvim ."}}

	err := svc.OpenWindowWithLayout("feature-auth", "/tmp/wt", "even-vertical", panes)
	require.NoError(t, err)

	// list-windows + new-window only (no splits, no layout, no select-pane for single pane)
	assert.Equal(t, 2, fake.CallCount())
}

func TestOpenWindowWithLayout_EmptyCommand(t *testing.T) {
	fake := runner.NewFakeRunner(map[string]runner.Response{
		"tmux list-windows": {Output: []byte("main\n")},
	})
	svc := NewService(fake)

	panes := []PaneSpec{{Command: ""}, {Command: ""}}

	err := svc.OpenWindowWithLayout("feature-auth", "/tmp/wt", "even-vertical", panes)
	require.NoError(t, err)

	// new-window should NOT have a command argument
	newWindowCall := fake.Calls[1]
	assert.Equal(t, []string{"new-window", "-n", "feature-auth", "-c", "/tmp/wt"}, newWindowCall.Args)

	// split-window should NOT have a command argument
	splitCall := fake.Calls[2]
	assert.Equal(t, []string{"split-window", "-t", "feature-auth", "-v", "-c", "/tmp/wt"}, splitCall.Args)
}
