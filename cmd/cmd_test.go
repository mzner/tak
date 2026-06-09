package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runCmd executes the root command with the given args and captures what the
// command writes to stdout and stderr, plus the error RunE returns. Because
// cobra retains package-level flag variables between runs, resetFlags is
// called first so each invocation starts from a clean slate.
func runCmd(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	resetFlags()

	origOut, origErr := os.Stdout, os.Stderr
	outR, outW, _ := os.Pipe()
	errR, errW, _ := os.Pipe()
	os.Stdout, os.Stderr = outW, errW

	outDone := make(chan string)
	errDone := make(chan string)
	go func() { var b bytes.Buffer; _, _ = io.Copy(&b, outR); outDone <- b.String() }()
	go func() { var b bytes.Buffer; _, _ = io.Copy(&b, errR); errDone <- b.String() }()

	rootCmd.SetArgs(args)
	err = rootCmd.Execute()

	_ = outW.Close()
	_ = errW.Close()
	os.Stdout, os.Stderr = origOut, origErr
	return <-outDone, <-errDone, err
}

// resetFlags restores every command's bound flag variable to its zero value so
// state does not leak between test cases sharing the global rootCmd.
func resetFlags() {
	rmForce = false
	addOpen, addPin, addFrom = false, false, ""
	lsStatus, lsJSON = false, false
	gcMerged, gcDryRun = false, false
}

// newRepo creates a git repo in a temp dir, makes an initial commit, runs
// `tak init`-equivalent setup with a contained worktree_base, and chdirs into
// it. HOME is pointed at a temp dir so the user's global config is never read.
func newRepo(t *testing.T) (repoDir, wtBase string) {
	t.Helper()

	repoDir = t.TempDir()
	wtBase = t.TempDir()

	t.Setenv("HOME", t.TempDir())
	t.Setenv("GIT_AUTHOR_NAME", "Test")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "Test")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")

	git(t, repoDir, "init")
	git(t, repoDir, "commit", "--allow-empty", "-m", "initial")
	// Normalize the default branch name so tests don't depend on the host's
	// init.defaultBranch (CI may default to "master" instead of "main").
	git(t, repoDir, "branch", "-M", "main")

	// Contain worktrees in wtBase instead of creating sibling dirs.
	cfg := "worktree_base: " + wtBase + "\npins: []\n"
	if err := os.WriteFile(filepath.Join(repoDir, ".tak.yml"), []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repoDir, ".tak"), 0755); err != nil {
		t.Fatal(err)
	}

	t.Chdir(repoDir)
	return repoDir, wtBase
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	c := exec.Command("git", args...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %s\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

// wtPath returns the worktree path tak would compute for a branch under wtBase.
func wtPath(repoDir, wtBase, branch string) string {
	slug := strings.ReplaceAll(branch, "/", "--")
	return filepath.Join(wtBase, filepath.Base(repoDir)+"--"+slug)
}

func TestAdd_CreatesWorktreeAndBranch(t *testing.T) {
	repoDir, wtBase := newRepo(t)

	stdout, _, err := runCmd(t, "add", "feature/auth")
	if err != nil {
		t.Fatalf("add returned error: %v", err)
	}
	if !strings.Contains(stdout, "Created worktree feature/auth") {
		t.Errorf("unexpected stdout: %q", stdout)
	}

	if _, statErr := os.Stat(wtPath(repoDir, wtBase, "feature/auth")); statErr != nil {
		t.Errorf("worktree dir missing: %v", statErr)
	}
	if !strings.Contains(git(t, repoDir, "branch"), "feature/auth") {
		t.Error("branch feature/auth was not created")
	}
}

func TestAdd_DuplicateIsError(t *testing.T) {
	newRepo(t)

	if _, _, err := runCmd(t, "add", "feature/dup"); err != nil {
		t.Fatalf("first add failed: %v", err)
	}
	_, _, err := runCmd(t, "add", "feature/dup")
	if err == nil {
		t.Fatal("expected error adding a duplicate worktree, got nil")
	}
	if !strings.Contains(err.Error(), "already has a worktree") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAdd_FromOnExistingBranchIsError(t *testing.T) {
	repoDir, _ := newRepo(t)
	git(t, repoDir, "branch", "existing")

	_, _, err := runCmd(t, "add", "existing", "--from", "main")
	if err == nil {
		t.Fatal("expected error using --from on an existing branch, got nil")
	}
	if !strings.Contains(err.Error(), "--from is ignored") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAdd_BranchPrefixApplied(t *testing.T) {
	repoDir, wtBase := newRepo(t)

	cfg := "worktree_base: " + wtBase + "\nbranch_prefix: \"feature/\"\npins: []\n"
	if err := os.WriteFile(filepath.Join(repoDir, ".tak.yml"), []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, _, err := runCmd(t, "add", "auth")
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}
	if !strings.Contains(stdout, "feature/auth") {
		t.Errorf("prefix not applied, stdout: %q", stdout)
	}
}

func TestRm_DeletesBranchWithNoCommits(t *testing.T) {
	repoDir, _ := newRepo(t)

	if _, _, err := runCmd(t, "add", "feature/empty"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runCmd(t, "rm", "feature/empty"); err != nil {
		t.Fatalf("rm failed: %v", err)
	}
	if strings.Contains(git(t, repoDir, "branch"), "feature/empty") {
		t.Error("branch should have been deleted")
	}
}

func TestRm_KeepsBranchWithUnpushedCommits(t *testing.T) {
	repoDir, wtBase := newRepo(t)

	if _, _, err := runCmd(t, "add", "feature/work"); err != nil {
		t.Fatal(err)
	}
	git(t, wtPath(repoDir, wtBase, "feature/work"), "commit", "--allow-empty", "-m", "wip")

	_, stderr, err := runCmd(t, "rm", "feature/work")
	if err != nil {
		t.Fatalf("rm failed: %v", err)
	}
	if !strings.Contains(git(t, repoDir, "branch"), "feature/work") {
		t.Error("branch with unpushed commits should be kept")
	}
	if !strings.Contains(stderr, "unpushed commits") {
		t.Errorf("expected warning about unpushed commits, stderr: %q", stderr)
	}
}

func TestRm_ForceDeletesBranchWithCommits(t *testing.T) {
	repoDir, wtBase := newRepo(t)

	if _, _, err := runCmd(t, "add", "feature/force"); err != nil {
		t.Fatal(err)
	}
	git(t, wtPath(repoDir, wtBase, "feature/force"), "commit", "--allow-empty", "-m", "wip")

	if _, _, err := runCmd(t, "rm", "--force", "feature/force"); err != nil {
		t.Fatalf("rm --force failed: %v", err)
	}
	if strings.Contains(git(t, repoDir, "branch"), "feature/force") {
		t.Error("branch should have been force-deleted")
	}
}

func TestRm_RefusesPinned(t *testing.T) {
	repoDir, wtBase := newRepo(t)

	if _, _, err := runCmd(t, "add", "feature/pinned", "--pin"); err != nil {
		t.Fatal(err)
	}
	_, stderr, err := runCmd(t, "rm", "feature/pinned")
	if err != nil {
		t.Fatalf("rm of pinned should not error, got: %v", err)
	}
	if !strings.Contains(stderr, "pinned") {
		t.Errorf("expected pinned skip message, stderr: %q", stderr)
	}
	if _, statErr := os.Stat(wtPath(repoDir, wtBase, "feature/pinned")); statErr != nil {
		t.Error("pinned worktree should still exist")
	}
}

func TestPinUnpin_Roundtrip(t *testing.T) {
	newRepo(t)

	if _, _, err := runCmd(t, "add", "feature/x"); err != nil {
		t.Fatal(err)
	}

	stdout, _, err := runCmd(t, "pin", "feature/x")
	if err != nil || !strings.Contains(stdout, "Pinned feature/x") {
		t.Fatalf("pin failed: %v / %q", err, stdout)
	}

	stdout, _, err = runCmd(t, "unpin", "feature/x")
	if err != nil || !strings.Contains(stdout, "Unpinned feature/x") {
		t.Fatalf("unpin failed: %v / %q", err, stdout)
	}
}

func TestCd_PrintsPath(t *testing.T) {
	repoDir, wtBase := newRepo(t)

	if _, _, err := runCmd(t, "add", "feature/nav"); err != nil {
		t.Fatal(err)
	}
	stdout, _, err := runCmd(t, "cd", "feature/nav")
	if err != nil {
		t.Fatalf("cd failed: %v", err)
	}
	if strings.TrimSpace(stdout) != wtPath(repoDir, wtBase, "feature/nav") {
		t.Errorf("cd printed %q, want %q", strings.TrimSpace(stdout), wtPath(repoDir, wtBase, "feature/nav"))
	}
}

func TestCd_NotFoundIsError(t *testing.T) {
	newRepo(t)

	_, _, err := runCmd(t, "cd", "ghost")
	if err == nil {
		t.Fatal("expected error for nonexistent branch")
	}
	if !strings.Contains(err.Error(), "no worktree for branch") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRename_BlocksDefaultBranch(t *testing.T) {
	newRepo(t)

	_, _, err := runCmd(t, "rename", "main", "trunk")
	if err == nil {
		t.Fatal("expected error renaming the default branch")
	}
	if !strings.Contains(err.Error(), "default branch") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRename_RenamesBranch(t *testing.T) {
	repoDir, _ := newRepo(t)

	if _, _, err := runCmd(t, "add", "feature/old"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runCmd(t, "rename", "feature/old", "feature/new"); err != nil {
		t.Fatalf("rename failed: %v", err)
	}

	branches := git(t, repoDir, "branch")
	if !strings.Contains(branches, "feature/new") || strings.Contains(branches, "feature/old") {
		t.Errorf("rename did not update branch: %q", branches)
	}
}

func TestLs_JSONOutput(t *testing.T) {
	newRepo(t)

	if _, _, err := runCmd(t, "add", "feature/json"); err != nil {
		t.Fatal(err)
	}
	stdout, _, err := runCmd(t, "ls", "--json")
	if err != nil {
		t.Fatalf("ls --json failed: %v", err)
	}
	if !strings.Contains(stdout, `"branch": "feature/json"`) {
		t.Errorf("json output missing branch, got: %q", stdout)
	}
	if !strings.Contains(stdout, `"pinned": false`) {
		t.Errorf("json output missing pinned field, got: %q", stdout)
	}
}

func TestInit_AlreadyExistsIsError(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GIT_AUTHOR_NAME", "Test")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "Test")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")
	git(t, repoDir, "init")
	git(t, repoDir, "commit", "--allow-empty", "-m", "initial")
	t.Chdir(repoDir)

	if _, _, err := runCmd(t, "init"); err != nil {
		t.Fatalf("first init failed: %v", err)
	}
	_, _, err := runCmd(t, "init")
	if err == nil {
		t.Fatal("expected error re-initializing")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExec_ForwardsChildExitCode(t *testing.T) {
	newRepo(t)

	if _, _, err := runCmd(t, "add", "feature/exec"); err != nil {
		t.Fatal(err)
	}

	_, _, err := runCmd(t, "exec", "feature/exec", "--", "sh", "-c", "exit 3")
	if err == nil {
		t.Fatal("expected exec to surface a nonzero exit")
	}
	ee, ok := errors.AsType[*exitError](err)
	if !ok {
		t.Fatalf("expected *exitError, got %T: %v", err, err)
	}
	if ee.code != 3 {
		t.Errorf("forwarded exit code = %d, want 3", ee.code)
	}
}
