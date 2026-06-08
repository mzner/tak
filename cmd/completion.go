package cmd

import (
	"os"

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
	r := runner.NewExecRunner()
	wtSvc := worktree.NewService(r)
	entries, err := wtSvc.List()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var branches []string
	for _, e := range entries {
		if e.Branch != "(detached)" {
			branches = append(branches, e.Branch)
		}
	}
	return branches, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
