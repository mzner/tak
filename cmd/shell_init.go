package cmd

import (
	"fmt"
	"os"

	"github.com/mzner/tak/internal/shell"
	"github.com/spf13/cobra"
)

var shellInitCmd = &cobra.Command{
	Use:   "shell-init <zsh|bash|fish>",
	Short: "Print shell hook for tak cd integration",
	Long: `Print a shell function that wraps tak cd to change your working directory.

Add to your shell rc file:
  eval "$(tak shell-init zsh)"    # for .zshrc
  eval "$(tak shell-init bash)"   # for .bashrc
  tak shell-init fish | source    # for config.fish`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		output, err := shell.Generate(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		fmt.Print(output)
	},
}

func init() {
	rootCmd.AddCommand(shellInitCmd)
}
