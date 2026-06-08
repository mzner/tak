//go:build integration

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_FullWorkflow(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "tak")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, ".")
	buildCmd.Dir = projectRoot(t)
	require.NoError(t, buildCmd.Run(), "failed to build tak")

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
	assert.True(t, strings.Contains(output, "Nothing to clean up") || strings.Contains(output, "pinned"))

	// tak unpin feature/test
	output = runTak(t, tmpBin, repoDir, "unpin", "feature/test")
	assert.Contains(t, output, "Unpinned feature/test")

	// tak rm feature/test
	output = runTak(t, tmpBin, repoDir, "rm", "feature/test")
	assert.Contains(t, output, "Removed worktree feature/test")
	assert.NoDirExists(t, wtPath)

	t.Cleanup(func() {
		os.RemoveAll(wtPath)
	})
}

func TestIntegration_RmDeletesBranch(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "tak")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, ".")
	buildCmd.Dir = projectRoot(t)
	require.NoError(t, buildCmd.Run())

	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "initial")
	runTak(t, tmpBin, repoDir, "init")

	// Add and remove a worktree with no commits (should delete branch)
	runTak(t, tmpBin, repoDir, "add", "feature/empty")
	runTak(t, tmpBin, repoDir, "rm", "feature/empty")

	branches := runGitOutput(t, repoDir, "branch")
	assert.NotContains(t, branches, "feature/empty")
}

func TestIntegration_RmKeepsBranchWithCommits(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "tak")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, ".")
	buildCmd.Dir = projectRoot(t)
	require.NoError(t, buildCmd.Run())

	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "initial")
	runTak(t, tmpBin, repoDir, "init")

	// Add worktree and make a commit in it
	runTak(t, tmpBin, repoDir, "add", "feature/work")
	wtPath := filepath.Join(filepath.Dir(repoDir), filepath.Base(repoDir)+"--feature--work")
	runGit(t, wtPath, "commit", "--allow-empty", "-m", "work in progress")

	// Remove without --force: branch should be kept
	runTak(t, tmpBin, repoDir, "rm", "feature/work")

	branches := runGitOutput(t, repoDir, "branch")
	assert.Contains(t, branches, "feature/work")

	t.Cleanup(func() { os.RemoveAll(wtPath) })
}

func TestIntegration_RmForceDeletesBranchWithCommits(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "tak")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, ".")
	buildCmd.Dir = projectRoot(t)
	require.NoError(t, buildCmd.Run())

	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "initial")
	runTak(t, tmpBin, repoDir, "init")

	runTak(t, tmpBin, repoDir, "add", "feature/force")
	wtPath := filepath.Join(filepath.Dir(repoDir), filepath.Base(repoDir)+"--feature--force")
	runGit(t, wtPath, "commit", "--allow-empty", "-m", "unmerged work")

	// Remove with --force: branch should be deleted
	runTak(t, tmpBin, repoDir, "rm", "--force", "feature/force")

	branches := runGitOutput(t, repoDir, "branch")
	assert.NotContains(t, branches, "feature/force")

	t.Cleanup(func() { os.RemoveAll(wtPath) })
}

func TestIntegration_RmRefusesPinned(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "tak")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, ".")
	buildCmd.Dir = projectRoot(t)
	require.NoError(t, buildCmd.Run())

	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "initial")
	runTak(t, tmpBin, repoDir, "init")

	runTak(t, tmpBin, repoDir, "add", "feature/pinned", "--pin")
	wtPath := filepath.Join(filepath.Dir(repoDir), filepath.Base(repoDir)+"--feature--pinned")

	// Try to remove pinned worktree
	output := runTak(t, tmpBin, repoDir, "rm", "feature/pinned")
	assert.Contains(t, output, "pinned")
	assert.DirExists(t, wtPath)

	t.Cleanup(func() { os.RemoveAll(wtPath) })
}

func TestIntegration_RmFromWorktree(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "tak")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, ".")
	buildCmd.Dir = projectRoot(t)
	require.NoError(t, buildCmd.Run())

	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "initial")
	runTak(t, tmpBin, repoDir, "init")

	// Pin a worktree
	runTak(t, tmpBin, repoDir, "add", "feature/pintest", "--pin")
	wtPath := filepath.Join(filepath.Dir(repoDir), filepath.Base(repoDir)+"--feature--pintest")

	// Try rm from inside the worktree — should still find pin
	output := runTakDir(t, tmpBin, wtPath, "rm", "feature/pintest")
	assert.Contains(t, output, "pinned")
	assert.DirExists(t, wtPath)

	t.Cleanup(func() { os.RemoveAll(wtPath) })
}

func TestIntegration_GcMerged(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "tak")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, ".")
	buildCmd.Dir = projectRoot(t)
	require.NoError(t, buildCmd.Run())

	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "initial")
	runTak(t, tmpBin, repoDir, "init")

	// Create a worktree, merge it, then gc
	runTak(t, tmpBin, repoDir, "add", "feature/merged")
	runGit(t, repoDir, "merge", "feature/merged")

	// gc without --merged should NOT remove it
	output := runTak(t, tmpBin, repoDir, "gc", "--dry-run")
	assert.Contains(t, output, "Nothing to clean up")

	// gc --merged should remove it
	output = runTak(t, tmpBin, repoDir, "gc", "--merged")
	assert.Contains(t, output, "Removed")

	wtPath := filepath.Join(filepath.Dir(repoDir), filepath.Base(repoDir)+"--feature--merged")
	assert.NoDirExists(t, wtPath)

	// Branch should also be deleted
	branches := runGitOutput(t, repoDir, "branch")
	assert.NotContains(t, branches, "feature/merged")

	t.Cleanup(func() { os.RemoveAll(wtPath) })
}

func TestIntegration_CdNotFound(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "tak")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, ".")
	buildCmd.Dir = projectRoot(t)
	require.NoError(t, buildCmd.Run())

	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "initial")
	runTak(t, tmpBin, repoDir, "init")

	output := runTak(t, tmpBin, repoDir, "cd", "nonexistent")
	assert.Contains(t, output, "no worktree for branch")
}

func TestIntegration_InitAlreadyExists(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "tak")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, ".")
	buildCmd.Dir = projectRoot(t)
	require.NoError(t, buildCmd.Run())

	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "initial")

	runTak(t, tmpBin, repoDir, "init")
	output := runTak(t, tmpBin, repoDir, "init")
	assert.Contains(t, output, "already exists")
}

func TestIntegration_AddDuplicate(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "tak")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, ".")
	buildCmd.Dir = projectRoot(t)
	require.NoError(t, buildCmd.Run())

	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "initial")
	runTak(t, tmpBin, repoDir, "init")

	runTak(t, tmpBin, repoDir, "add", "feature/dup")
	output := runTak(t, tmpBin, repoDir, "add", "feature/dup")
	assert.Contains(t, output, "already has a worktree")

	wtPath := filepath.Join(filepath.Dir(repoDir), filepath.Base(repoDir)+"--feature--dup")
	t.Cleanup(func() { os.RemoveAll(wtPath) })
}

func TestIntegration_DoctorBroken(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "tak")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, ".")
	buildCmd.Dir = projectRoot(t)
	require.NoError(t, buildCmd.Run())

	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "initial")
	runTak(t, tmpBin, repoDir, "init")

	runTak(t, tmpBin, repoDir, "add", "feature/broken")
	wtPath := filepath.Join(filepath.Dir(repoDir), filepath.Base(repoDir)+"--feature--broken")

	// Manually delete the worktree directory to simulate a broken state
	os.RemoveAll(wtPath)

	output := runTak(t, tmpBin, repoDir, "doctor")
	assert.Contains(t, output, "path does not exist")
}

func TestIntegration_ShellInit(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "tak")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, ".")
	buildCmd.Dir = projectRoot(t)
	require.NoError(t, buildCmd.Run())

	output := runTak(t, tmpBin, t.TempDir(), "shell-init", "zsh")
	assert.Contains(t, output, "tak()")
	assert.Contains(t, output, "command tak cd")

	output = runTak(t, tmpBin, t.TempDir(), "shell-init", "fish")
	assert.Contains(t, output, "function tak")
}

func TestIntegration_BranchPrefix(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "tak")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, ".")
	buildCmd.Dir = projectRoot(t)
	require.NoError(t, buildCmd.Run())

	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "initial")

	// Write config with branch prefix
	configContent := "worktree_base: \"\"\nbranch_prefix: \"feature/\"\npins: []\n"
	os.WriteFile(filepath.Join(repoDir, ".tak.yml"), []byte(configContent), 0644)
	os.MkdirAll(filepath.Join(repoDir, ".tak"), 0755)

	// tak add auth → should create feature/auth
	output := runTak(t, tmpBin, repoDir, "add", "auth")
	assert.Contains(t, output, "feature/auth")

	wtPath := filepath.Join(filepath.Dir(repoDir), filepath.Base(repoDir)+"--feature--auth")
	assert.DirExists(t, wtPath)

	t.Cleanup(func() { os.RemoveAll(wtPath) })
}

func TestIntegration_Exec(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "tak")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, ".")
	buildCmd.Dir = projectRoot(t)
	require.NoError(t, buildCmd.Run())

	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "initial")
	initTakWithLocalBase(t, repoDir)

	runTak(t, tmpBin, repoDir, "add", "feature/exec-test")
	wtPath := filepath.Join(filepath.Dir(repoDir), filepath.Base(repoDir)+"--feature--exec-test")

	// Exec pwd in the worktree
	output := runTak(t, tmpBin, repoDir, "exec", "feature/exec-test", "--", "pwd")
	assert.Contains(t, output, wtPath)

	t.Cleanup(func() { os.RemoveAll(wtPath) })
}

func TestIntegration_Rename(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "tak")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, ".")
	buildCmd.Dir = projectRoot(t)
	require.NoError(t, buildCmd.Run())

	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "initial")
	initTakWithLocalBase(t, repoDir)

	runTak(t, tmpBin, repoDir, "add", "feature/old-name")
	wtPath := filepath.Join(filepath.Dir(repoDir), filepath.Base(repoDir)+"--feature--old-name")

	// Rename
	output := runTak(t, tmpBin, repoDir, "rename", "feature/old-name", "feature/new-name")
	assert.Contains(t, output, "Renamed")

	// ls should show new name
	output = runTak(t, tmpBin, repoDir, "ls")
	assert.Contains(t, output, "feature/new-name")
	assert.NotContains(t, output, "feature/old-name")

	// git branch should show new name
	branches := runGitOutput(t, repoDir, "branch")
	assert.Contains(t, branches, "feature/new-name")
	assert.NotContains(t, branches, "feature/old-name")

	t.Cleanup(func() { os.RemoveAll(wtPath) })
}

func TestIntegration_Hooks(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "tak")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, ".")
	buildCmd.Dir = projectRoot(t)
	require.NoError(t, buildCmd.Run())

	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "initial")

	// Write config with hooks and explicit worktree_base
	parentDir := filepath.Dir(repoDir)
	configContent := fmt.Sprintf(`worktree_base: %s
pins: []
hooks:
  post_create:
    - type: copy
      from: .env
      to: .env
    - type: command
      command: echo "hook-ran" > hook-proof.txt
`, parentDir)
	os.WriteFile(filepath.Join(repoDir, ".tak.yml"), []byte(configContent), 0644)
	os.MkdirAll(filepath.Join(repoDir, ".tak"), 0755)
	os.WriteFile(filepath.Join(repoDir, ".env"), []byte("SECRET=test123"), 0644)

	runTak(t, tmpBin, repoDir, "add", "feature/hooks")
	wtPath := filepath.Join(filepath.Dir(repoDir), filepath.Base(repoDir)+"--feature--hooks")

	// Verify copy hook worked
	envContent, err := os.ReadFile(filepath.Join(wtPath, ".env"))
	require.NoError(t, err)
	assert.Equal(t, "SECRET=test123", string(envContent))

	// Verify command hook worked
	proofContent, err := os.ReadFile(filepath.Join(wtPath, "hook-proof.txt"))
	require.NoError(t, err)
	assert.Contains(t, string(proofContent), "hook-ran")

	t.Cleanup(func() { os.RemoveAll(wtPath) })
}

func TestIntegration_RepoAddAndLs(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "tak")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, ".")
	buildCmd.Dir = projectRoot(t)
	require.NoError(t, buildCmd.Run())

	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "initial")

	// Register the repo
	output := runTak(t, tmpBin, repoDir, "repo", "add")
	assert.Contains(t, output, "Registered")

	// List should show it
	output = runTak(t, tmpBin, repoDir, "repo", "ls")
	assert.Contains(t, output, filepath.Base(repoDir))

	// Remove it
	output = runTak(t, tmpBin, repoDir, "repo", "rm", filepath.Base(repoDir))
	assert.Contains(t, output, "Unregistered")
}

func TestIntegration_RepoAddInvalidPath(t *testing.T) {
	tmpBin := filepath.Join(t.TempDir(), "tak")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, ".")
	buildCmd.Dir = projectRoot(t)
	require.NoError(t, buildCmd.Run())

	output := runTak(t, tmpBin, t.TempDir(), "repo", "add", "/nonexistent/path")
	assert.Contains(t, output, "not a directory")
}

// Helper functions

func initTakWithLocalBase(t *testing.T, repoDir string) {
	t.Helper()
	// Use the parent dir of the repo as worktree_base to avoid global config override
	parentDir := filepath.Dir(repoDir)
	configContent := fmt.Sprintf("worktree_base: %s\npins: []\n", parentDir)
	os.WriteFile(filepath.Join(repoDir, ".tak.yml"), []byte(configContent), 0644)
	os.MkdirAll(filepath.Join(repoDir, ".tak"), 0755)
}

var testHome string

func TestMain(m *testing.M) {
	// Use a temp HOME so tests don't read the user's global tak config
	dir, _ := os.MkdirTemp("", "tak-test-home")
	testHome = dir
	os.Exit(m.Run())
}

func runTak(t *testing.T, bin string, dir string, args ...string) string {
	t.Helper()
	return runTakDir(t, bin, dir, args...)
}

func runTakDir(t *testing.T, bin string, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "HOME="+testHome)
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
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Test",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %s failed: %s", strings.Join(args, " "), output)
}

func runGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Test",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %s failed: %s", strings.Join(args, " "), output)
	return string(output)
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
