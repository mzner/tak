//go:build integration

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_FullWorkflow(t *testing.T) {
	// Build tak binary
	tmpBin := filepath.Join(t.TempDir(), "tak")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, ".")
	buildCmd.Dir = projectRoot(t)
	require.NoError(t, buildCmd.Run(), "failed to build tak")

	// Create a temporary git repo
	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "initial")

	// tak init
	output := runTak(t, tmpBin, repoDir, "init")
	assert.Contains(t, output, "Initialized tak")
	assert.FileExists(t, filepath.Join(repoDir, ".tak.yml"))
	assert.DirExists(t, filepath.Join(repoDir, ".tak"))

	// tak add feature/test
	output = runTak(t, tmpBin, repoDir, "add", "feature/test")
	assert.Contains(t, output, "Created worktree feature/test")

	// Verify worktree exists
	wtPath := filepath.Join(filepath.Dir(repoDir), filepath.Base(repoDir)+"--feature--test")
	assert.DirExists(t, wtPath)

	// tak ls
	output = runTak(t, tmpBin, repoDir, "ls")
	assert.Contains(t, output, "feature/test")

	// tak pin feature/test
	output = runTak(t, tmpBin, repoDir, "pin", "feature/test")
	assert.Contains(t, output, "Pinned feature/test")

	// tak cd feature/test
	output = runTak(t, tmpBin, repoDir, "cd", "feature/test")
	assert.Contains(t, output, wtPath)

	// tak doctor
	output = runTak(t, tmpBin, repoDir, "doctor")
	assert.Contains(t, output, "Checking")

	// tak gc --dry-run (pinned should be skipped)
	output = runTak(t, tmpBin, repoDir, "gc", "--dry-run")
	// Either "Nothing to clean up" or shows pinned in skipped
	assert.True(t, strings.Contains(output, "Nothing to clean up") || strings.Contains(output, "pinned"))

	// tak unpin feature/test
	output = runTak(t, tmpBin, repoDir, "unpin", "feature/test")
	assert.Contains(t, output, "Unpinned feature/test")

	// tak rm feature/test
	output = runTak(t, tmpBin, repoDir, "rm", "feature/test")
	assert.Contains(t, output, "Removed worktree feature/test")
	assert.NoDirExists(t, wtPath)

	// Cleanup
	t.Cleanup(func() {
		os.RemoveAll(wtPath)
	})
}

func runTak(t *testing.T, bin string, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("tak %s failed: %s\nOutput: %s", strings.Join(args, " "), err, output)
	}
	return string(output)
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %s failed: %s", strings.Join(args, " "), output)
}

func projectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root")
		}
		dir = parent
	}
}
