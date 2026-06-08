package cmd

import (
	"os"
	"strings"

	"github.com/mzner/tak/internal/config"
	"github.com/mzner/tak/internal/runner"
	"github.com/mzner/tak/internal/worktree"
	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion <zsh|bash|fish>",
	Short: "Generate shell completion script",
	Long: `Generate a completion script for your shell.

Add to your shell config:

  # zsh (add to ~/.zshrc)
  source <(tak completion zsh)

  # bash (add to ~/.bashrc)
  source <(tak completion bash)

  # fish (add to ~/.config/fish/config.fish)
  tak completion fish | source`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"zsh", "bash", "fish"},
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "zsh":
			rootCmd.GenZshCompletion(os.Stdout)
		case "bash":
			rootCmd.GenBashCompletion(os.Stdout)
		case "fish":
			rootCmd.GenFishCompletion(os.Stdout, true)
		}
	},
}

func completeWorktreeBranches(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// If user is typing repo:branch, complete branches from that repo
	if parts := strings.SplitN(toComplete, ":", 2); len(parts) == 2 {
		return completeRemoteRepoBranches(parts[0], parts[1])
	}

	r := runner.NewExecRunner()
	wtSvc := worktree.NewService(r)

	// Try local worktree branches first
	entries, err := wtSvc.List()
	if err == nil && len(entries) > 0 {
		var completions []string
		for _, e := range entries {
			if e.Branch != "(detached)" {
				completions = append(completions, e.Branch)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}

	// Not in a repo — show repo names only
	repos, err := config.LoadGlobal()
	if err == nil {
		var completions []string
		for name := range repos {
			completions = append(completions, name+":")
		}
		return completions, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
	}

	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeRemoteRepoBranches(repoName string, prefix string) ([]string, cobra.ShellCompDirective) {
	repos, err := config.LoadGlobal()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	repoPath, ok := repos[repoName]
	if !ok {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	r := runner.NewExecRunner()
	wtSvc := worktree.NewService(r)

	// Run git worktree list in the target repo
	output, err := r.RunInDir(repoPath, "git", "worktree", "list", "--porcelain")
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	_ = wtSvc // unused but keeps import clean
	var completions []string
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "branch refs/heads/") {
			branch := strings.TrimPrefix(line, "branch refs/heads/")
			completions = append(completions, repoName+":"+branch)
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
