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
	Run: func(cmd *cobra.Command, args []string) {
		r := runner.NewExecRunner()
		wtSvc := worktree.NewService(r)

		repoRoot, err := wtSvc.RepoRoot()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error: not in a git repository")
			os.Exit(1)
		}

		configPath := filepath.Join(repoRoot, ".tak.yml")
		if _, err := os.Stat(configPath); err == nil {
			fmt.Fprintln(os.Stderr, "error: .tak.yml already exists")
			os.Exit(1)
		}
		if err := os.WriteFile(configPath, []byte(defaultTakYml), 0644); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		takDir := filepath.Join(repoRoot, ".tak")
		if err := state.EnsureDir(takDir); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		gitignorePath := filepath.Join(repoRoot, ".gitignore")
		addToGitignore(gitignorePath, ".tak/")
		addToGitignore(gitignorePath, ".tak.yml")

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
	},
}

func addToGitignore(path string, entry string) {
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == entry {
			return
		}
	}

	newContent := string(content)
	if len(newContent) > 0 && !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}
	newContent += entry + "\n"
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
}
