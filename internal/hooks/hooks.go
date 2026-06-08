package hooks

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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

// RunPostCreate executes post_create hooks.
// mainRoot is the main worktree root (source for copy/symlink).
// wtPath is the newly created worktree path (destination).
func RunPostCreate(actions []Action, mainRoot string, wtPath string) error {
	for _, a := range actions {
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
			if err := runCommand(a, wtPath); err != nil {
				return fmt.Errorf("hook command '%s': %w", a.Command, err)
			}
		default:
			return fmt.Errorf("unknown hook type: %s", a.Type)
		}
	}
	return nil
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
	os.Remove(dst)

	return os.Symlink(src, dst)
}

func runCommand(a Action, wtPath string) error {
	dir := wtPath
	if a.WorkDir != "" {
		dir = filepath.Join(wtPath, a.WorkDir)
	}

	cmd := exec.Command("sh", "-c", a.Command)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if len(a.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range a.Env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
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
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

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
