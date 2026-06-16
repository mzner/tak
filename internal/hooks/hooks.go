package hooks

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Action represents a single hook step.
type Action struct {
	Type    string
	From    string
	To      string
	Command string
	Env     map[string]string
	WorkDir string
}

// Context provides metadata passed to hook commands as environment variables.
type Context struct {
	WorktreeName string
	SourceDir    string
	TargetDir    string
	Branch       string
	Hook         string
}

// Environ returns the context as TAK_* environment variables.
func (c Context) Environ() []string {
	return []string{
		"TAK_WORKTREE_NAME=" + c.WorktreeName,
		"TAK_SOURCE_DIR=" + c.SourceDir,
		"TAK_TARGET_DIR=" + c.TargetDir,
		"TAK_BRANCH=" + c.Branch,
		"TAK_HOOK=" + c.Hook,
	}
}

// Run executes a list of hook actions with the given context.
// mainRoot is the main worktree root (source for copy/symlink).
// wtPath is the worktree path (destination for copy/symlink, working dir for commands).
func Run(actions []Action, mainRoot string, wtPath string, ctx Context) error {
	total := len(actions)
	for i, a := range actions {
		label := actionLabel(a)
		fmt.Fprintf(os.Stderr, "  [%d/%d] %s\n", i+1, total, label)
		switch a.Type {
		case "copy":
			if err := runCopy(a, mainRoot, wtPath); err != nil {
				return fmt.Errorf("hook copy %s: %w", a.From, err)
			}
		case "symlink":
			if err := runSymlink(a, mainRoot, wtPath); err != nil {
				return fmt.Errorf("hook symlink %s: %w", a.From, err)
			}
		case "command":
			if err := runCommand(a, wtPath, ctx); err != nil {
				return fmt.Errorf("hook command '%s': %w", a.Command, err)
			}
		default:
			return fmt.Errorf("unknown hook type: %s", a.Type)
		}
	}
	return nil
}

func actionLabel(a Action) string {
	switch a.Type {
	case "copy":
		if a.To != "" && a.To != a.From {
			return fmt.Sprintf("copy %s → %s", a.From, a.To)
		}
		return fmt.Sprintf("copy %s", a.From)
	case "symlink":
		return fmt.Sprintf("symlink %s", a.From)
	case "command":
		cmd := a.Command
		if i := strings.IndexByte(cmd, '\n'); i >= 0 {
			cmd = cmd[:i]
		}
		if len(cmd) > 60 {
			cmd = cmd[:57] + "..."
		}
		return fmt.Sprintf("run: %s", cmd)
	default:
		return a.Type
	}
}

// RunPostCreate executes post_create hooks (legacy wrapper).
// mainRoot is the main worktree root (source for copy/symlink).
// wtPath is the newly created worktree path (destination).
func RunPostCreate(actions []Action, mainRoot string, wtPath string) error {
	return Run(actions, mainRoot, wtPath, Context{})
}

func runCopy(a Action, mainRoot string, wtPath string) error {
	src := filepath.Join(mainRoot, a.From)
	dst := a.To
	if dst == "" {
		dst = a.From
	}
	dst = filepath.Join(wtPath, dst)

	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("source not found: %s", src)
	}

	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

func runSymlink(a Action, mainRoot string, wtPath string) error {
	src := filepath.Join(mainRoot, a.From)
	dst := a.To
	if dst == "" {
		dst = a.From
	}
	dst = filepath.Join(wtPath, dst)

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// Remove existing file/dir at destination
	_ = os.Remove(dst)

	return os.Symlink(src, dst)
}

func runCommand(a Action, wtPath string, ctx Context) error {
	dir := wtPath
	if a.WorkDir != "" {
		dir = filepath.Join(wtPath, a.WorkDir)
	}

	cmd := exec.Command("sh", "-c", a.Command)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, ctx.Environ()...)
	for k, v := range a.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	return cmd.Run()
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	_, err = io.Copy(out, in)
	return err
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target)
	})
}
