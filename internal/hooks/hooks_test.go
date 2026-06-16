package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setup creates a fake (mainRoot, wtPath) pair under t.TempDir() and returns
// both. The main root is pre-populated with the files the test will reference
// via Action.From.
func setup(t *testing.T, mainFiles map[string]string) (mainRoot, wtPath string) {
	t.Helper()
	root := t.TempDir()
	mainRoot = filepath.Join(root, "main")
	wtPath = filepath.Join(root, "wt")
	if err := os.MkdirAll(mainRoot, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(wtPath, 0755); err != nil {
		t.Fatal(err)
	}
	for rel, content := range mainFiles {
		full := filepath.Join(mainRoot, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return mainRoot, wtPath
}

func TestRunPostCreate_CopyFile(t *testing.T) {
	mainRoot, wtPath := setup(t, map[string]string{".env": "SECRET=abc"})

	err := RunPostCreate([]Action{{Type: "copy", From: ".env"}}, mainRoot, wtPath)
	if err != nil {
		t.Fatalf("RunPostCreate failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(wtPath, ".env"))
	if err != nil {
		t.Fatalf("destination not created: %v", err)
	}
	if string(got) != "SECRET=abc" {
		t.Errorf("content = %q, want %q", got, "SECRET=abc")
	}
}

func TestRunPostCreate_CopyFileWithDifferentTo(t *testing.T) {
	mainRoot, wtPath := setup(t, map[string]string{".env.example": "X=1"})

	err := RunPostCreate(
		[]Action{{Type: "copy", From: ".env.example", To: ".env"}},
		mainRoot, wtPath,
	)
	if err != nil {
		t.Fatalf("RunPostCreate failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(wtPath, ".env")); err != nil {
		t.Errorf("expected .env at destination: %v", err)
	}
	if _, err := os.Stat(filepath.Join(wtPath, ".env.example")); !os.IsNotExist(err) {
		t.Error("source filename should not exist at destination when To differs")
	}
}

func TestRunPostCreate_CopyDir(t *testing.T) {
	mainRoot, wtPath := setup(t, map[string]string{
		"config/a.yml":     "a: 1",
		"config/sub/b.yml": "b: 2",
	})

	err := RunPostCreate([]Action{{Type: "copy", From: "config"}}, mainRoot, wtPath)
	if err != nil {
		t.Fatalf("RunPostCreate failed: %v", err)
	}

	for _, rel := range []string{"config/a.yml", "config/sub/b.yml"} {
		if _, err := os.Stat(filepath.Join(wtPath, rel)); err != nil {
			t.Errorf("missing %s: %v", rel, err)
		}
	}
}

func TestRunPostCreate_CopyMissingSourceErrors(t *testing.T) {
	mainRoot, wtPath := setup(t, nil)

	err := RunPostCreate([]Action{{Type: "copy", From: "nope.txt"}}, mainRoot, wtPath)
	if err == nil {
		t.Fatal("expected error for missing source")
	}
	if !strings.Contains(err.Error(), "source not found") {
		t.Errorf("error = %v, want 'source not found'", err)
	}
}

func TestRunPostCreate_CopyPreservesFileMode(t *testing.T) {
	mainRoot, wtPath := setup(t, nil)
	src := filepath.Join(mainRoot, "run.sh")
	if err := os.WriteFile(src, []byte("#!/bin/sh\necho hi\n"), 0755); err != nil {
		t.Fatal(err)
	}

	err := RunPostCreate([]Action{{Type: "copy", From: "run.sh"}}, mainRoot, wtPath)
	if err != nil {
		t.Fatalf("RunPostCreate failed: %v", err)
	}

	info, err := os.Stat(filepath.Join(wtPath, "run.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("mode = %o, want 0755", info.Mode().Perm())
	}
}

func TestRunPostCreate_Symlink(t *testing.T) {
	mainRoot, wtPath := setup(t, map[string]string{"node_modules/pkg/index.js": "// pkg"})

	err := RunPostCreate(
		[]Action{{Type: "symlink", From: "node_modules"}},
		mainRoot, wtPath,
	)
	if err != nil {
		t.Fatalf("RunPostCreate failed: %v", err)
	}

	link := filepath.Join(wtPath, "node_modules")
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("expected symlink at %s: %v", link, err)
	}
	want := filepath.Join(mainRoot, "node_modules")
	if target != want {
		t.Errorf("symlink target = %q, want %q", target, want)
	}

	// And the symlink resolves to a real file in the source.
	if _, err := os.Stat(filepath.Join(link, "pkg/index.js")); err != nil {
		t.Errorf("symlinked content not reachable: %v", err)
	}
}

func TestRunPostCreate_SymlinkOverwritesExisting(t *testing.T) {
	mainRoot, wtPath := setup(t, map[string]string{"shared/x": "shared"})
	// Pre-create something at the destination so the action must replace it.
	if err := os.WriteFile(filepath.Join(wtPath, "shared"), []byte("stale"), 0644); err != nil {
		t.Fatal(err)
	}

	err := RunPostCreate([]Action{{Type: "symlink", From: "shared"}}, mainRoot, wtPath)
	if err != nil {
		t.Fatalf("RunPostCreate failed: %v", err)
	}

	info, err := os.Lstat(filepath.Join(wtPath, "shared"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("destination should be a symlink, not a regular file")
	}
}

func TestRunPostCreate_Command(t *testing.T) {
	_, wtPath := setup(t, nil)

	err := RunPostCreate(
		[]Action{{Type: "command", Command: "echo ran > marker.txt"}},
		"", wtPath,
	)
	if err != nil {
		t.Fatalf("RunPostCreate failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(wtPath, "marker.txt"))
	if err != nil {
		t.Fatalf("marker.txt missing: %v", err)
	}
	if !strings.Contains(string(got), "ran") {
		t.Errorf("marker content = %q, want to contain 'ran'", got)
	}
}

func TestRunPostCreate_CommandRunsInWorktreeDir(t *testing.T) {
	_, wtPath := setup(t, nil)

	err := RunPostCreate(
		[]Action{{Type: "command", Command: "pwd > where.txt"}},
		"", wtPath,
	)
	if err != nil {
		t.Fatalf("RunPostCreate failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(wtPath, "where.txt"))
	if err != nil {
		t.Fatal(err)
	}
	// macOS resolves /var/folders to /private/var/folders for pwd, so compare
	// by suffix rather than exact path.
	if !strings.Contains(strings.TrimSpace(string(got)), filepath.Base(wtPath)) {
		t.Errorf("pwd output = %q, expected to contain %q", got, filepath.Base(wtPath))
	}
}

func TestRunPostCreate_CommandRunsInWorkDirSubdir(t *testing.T) {
	_, wtPath := setup(t, nil)
	if err := os.MkdirAll(filepath.Join(wtPath, "frontend"), 0755); err != nil {
		t.Fatal(err)
	}

	err := RunPostCreate(
		[]Action{{Type: "command", Command: "pwd > pwd.txt", WorkDir: "frontend"}},
		"", wtPath,
	)
	if err != nil {
		t.Fatalf("RunPostCreate failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(wtPath, "frontend", "pwd.txt"))
	if err != nil {
		t.Fatalf("pwd.txt should be in frontend/, got: %v", err)
	}
	if !strings.Contains(string(got), "frontend") {
		t.Errorf("expected pwd output under frontend/, got %q", got)
	}
}

func TestRunPostCreate_CommandReceivesEnv(t *testing.T) {
	_, wtPath := setup(t, nil)

	err := RunPostCreate(
		[]Action{{
			Type:    "command",
			Command: `echo "$TAK_TEST_VAR" > out.txt`,
			Env:     map[string]string{"TAK_TEST_VAR": "from-config"},
		}},
		"", wtPath,
	)
	if err != nil {
		t.Fatalf("RunPostCreate failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(wtPath, "out.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "from-config") {
		t.Errorf("env not propagated, got %q", got)
	}
}

func TestRunPostCreate_CommandNonzeroExitErrors(t *testing.T) {
	_, wtPath := setup(t, nil)

	err := RunPostCreate(
		[]Action{{Type: "command", Command: "exit 7"}},
		"", wtPath,
	)
	if err == nil {
		t.Fatal("expected error for nonzero command exit")
	}
	if !strings.Contains(err.Error(), "hook command") {
		t.Errorf("error = %v, want to mention 'hook command'", err)
	}
}

func TestRunPostCreate_UnknownTypeErrors(t *testing.T) {
	_, wtPath := setup(t, nil)

	err := RunPostCreate([]Action{{Type: "magical"}}, "", wtPath)
	if err == nil {
		t.Fatal("expected error for unknown hook type")
	}
	if !strings.Contains(err.Error(), "unknown hook type") {
		t.Errorf("error = %v, want 'unknown hook type'", err)
	}
}

func TestRunPostCreate_StopsOnFirstError(t *testing.T) {
	mainRoot, wtPath := setup(t, nil)

	err := RunPostCreate([]Action{
		{Type: "copy", From: "missing.txt"},                    // fails
		{Type: "command", Command: "echo should-not-run > x"}, // must not run
	}, mainRoot, wtPath)
	if err == nil {
		t.Fatal("expected error from first action")
	}
	if _, statErr := os.Stat(filepath.Join(wtPath, "x")); !os.IsNotExist(statErr) {
		t.Error("subsequent action ran despite earlier failure")
	}
}

func TestRunPostCreate_EmptyActionsIsNoop(t *testing.T) {
	mainRoot, wtPath := setup(t, nil)
	if err := RunPostCreate(nil, mainRoot, wtPath); err != nil {
		t.Errorf("nil actions should be a no-op, got: %v", err)
	}
	if err := RunPostCreate([]Action{}, mainRoot, wtPath); err != nil {
		t.Errorf("empty actions should be a no-op, got: %v", err)
	}
}

func TestRun_CommandReceivesContext(t *testing.T) {
	_, wtPath := setup(t, nil)

	ctx := Context{
		WorktreeName: "myrepo--feature--x",
		SourceDir:    "/src",
		TargetDir:    "/dst",
		Branch:       "feature/x",
		Hook:         "pre_create",
	}

	err := Run(
		[]Action{{Type: "command", Command: `printenv | grep ^TAK_ | sort > ctx.txt`}},
		"", wtPath, ctx,
	)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(wtPath, "ctx.txt"))
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"TAK_WORKTREE_NAME=myrepo--feature--x",
		"TAK_SOURCE_DIR=/src",
		"TAK_TARGET_DIR=/dst",
		"TAK_BRANCH=feature/x",
		"TAK_HOOK=pre_create",
	} {
		if !strings.Contains(string(got), want) {
			t.Errorf("missing %q in output:\n%s", want, got)
		}
	}
}

func TestRun_CommandEnvOverridesContext(t *testing.T) {
	_, wtPath := setup(t, nil)

	ctx := Context{
		WorktreeName: "wt",
		Hook:         "post_create",
	}

	err := Run(
		[]Action{{
			Type:    "command",
			Command: `echo "$CUSTOM_VAR" > out.txt`,
			Env:     map[string]string{"CUSTOM_VAR": "hello"},
		}},
		"", wtPath, ctx,
	)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(wtPath, "out.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "hello") {
		t.Errorf("custom env not set, got %q", got)
	}
}

func TestRun_NonzeroExitBlocksOperation(t *testing.T) {
	_, wtPath := setup(t, nil)

	ctx := Context{Hook: "pre_remove"}

	err := Run(
		[]Action{{Type: "command", Command: "exit 1"}},
		"", wtPath, ctx,
	)
	if err == nil {
		t.Fatal("expected error from failing hook")
	}
}

func TestRun_EmptyContextStillWorks(t *testing.T) {
	_, wtPath := setup(t, nil)

	err := Run(
		[]Action{{Type: "command", Command: "echo ok > out.txt"}},
		"", wtPath, Context{},
	)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(wtPath, "out.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "ok") {
		t.Errorf("got %q, want ok", got)
	}
}
