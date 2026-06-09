package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mzner/tak/internal/runner"
	"github.com/mzner/tak/internal/state"
	"github.com/mzner/tak/internal/worktree"
	"github.com/spf13/cobra"
)

const defaultTakYml = `# tak configuration — https://github.com/mzner/tak
#
# worktree_base: ""        # Where to create worktrees (empty = sibling dirs)
# branch_prefix: ""        # Auto-prepend to short branch names

# Pinned worktrees (excluded from tak gc)
pins: []
`

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize tak in the current repository",
	Long:  "Creates .tak.yml config and .tak/ state directory. Adds .tak/ to .gitignore.",
	RunE: func(cmd *cobra.Command, args []string) error {
		r := runner.NewExecRunner()
		wtSvc := worktree.NewService(r)

		repoRoot, err := wtSvc.RepoRoot()
		if err != nil {
			return errNotInRepo
		}

		configPath := filepath.Join(repoRoot, ".tak.yml")
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf(".tak.yml already exists")
		}
		if err := os.WriteFile(configPath, []byte(defaultTakYml), 0644); err != nil {
			return err
		}

		takDir := filepath.Join(repoRoot, ".tak")
		if err := state.EnsureDir(takDir); err != nil {
			return err
		}

		gitignorePath := filepath.Join(repoRoot, ".gitignore")
		if err := addToGitignore(gitignorePath, ".tak/"); err != nil {
			return err
		}
		if err := addToGitignore(gitignorePath, ".tak.yml"); err != nil {
			return err
		}

		fmt.Println("Initialized tak in", repoRoot)
		fmt.Println()
		fmt.Println("  To use `tak cd`, add the following line to your shell config:")
		fmt.Println()
		fmt.Println(`    echo 'eval "$(tak shell-init zsh)"' >> ~/.zshrc`)
		fmt.Println(`    echo 'eval "$(tak shell-init bash)"' >> ~/.bashrc`)
		fmt.Println(`    echo 'tak shell-init fish | source' >> ~/.config/fish/config.fish`)
		fmt.Println()
		fmt.Println("  Without this, `tak cd` prints the path but cannot change your directory.")
		fmt.Println("  After adding, restart your terminal or run: source ~/.zshrc")
		return nil
	},
}

func addToGitignore(path string, entry string) error {
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	for line := range strings.SplitSeq(string(content), "\n") {
		if strings.TrimSpace(line) == entry {
			return nil
		}
	}

	newContent := string(content)
	if len(newContent) > 0 && !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}
	newContent += entry + "\n"
	return os.WriteFile(path, []byte(newContent), 0644)
}

func init() {
	rootCmd.AddCommand(initCmd)
}
