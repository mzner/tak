package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/charmbracelet/huh"
	"github.com/mzner/tak/internal/config"
	"github.com/mzner/tak/internal/runner"
	"github.com/spf13/cobra"
)

// sortedRepoNames returns the repo names from a config map in alphabetical order.
func sortedRepoNames(repos map[string]string) []string {
	names := make([]string, 0, len(repos))
	for name := range repos {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage registered repos for cross-repo access",
	Long: `Register repos so you can use tak cd web:branch and tak ls web from anywhere.

Examples:
  tak repo add                        # register current directory
  tak repo add ~/projects/api         # register a specific path
  tak repo rm web                     # unregister a repo
  tak repo ls                         # list registered repos`,
}

var repoAddCmd = &cobra.Command{
	Use:   "add [path...]",
	Short: "Register a repo",
	Long: `Register one or more repos for cross-repo access.

Without arguments, registers the current directory.
With paths, registers each one. The repo name is the directory name.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var paths []string
		if len(args) == 0 {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			paths = append(paths, cwd)
		} else {
			paths = args
		}

		r := runner.NewExecRunner()

		for _, p := range paths {
			absPath, err := filepath.Abs(p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: invalid path %s: %s\n", p, err)
				continue
			}

			// Verify it's a directory
			info, err := os.Stat(absPath)
			if err != nil || !info.IsDir() {
				fmt.Fprintf(os.Stderr, "error: %s is not a directory\n", absPath)
				continue
			}

			// Verify it's a git repo
			_, err = r.RunInDir(absPath, "git", "rev-parse", "--git-dir")
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %s is not a git repository\n", absPath)
				continue
			}

			name := filepath.Base(absPath)
			if err := config.AddRepo(name, absPath); err != nil {
				fmt.Fprintf(os.Stderr, "error: %s\n", err)
				continue
			}
			fmt.Printf("Registered %s → %s\n", name, absPath)
		}
		return nil
	},
}

var repoRmCmd = &cobra.Command{
	Use:               "rm [name...]",
	Short:             "Unregister repo(s)",
	Long:              "Unregister one or more repos. Without arguments, shows an interactive picker.",
	ValidArgsFunction: completeRepoNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		var names []string
		if len(args) > 0 {
			names = args
		} else {
			repos, err := config.LoadGlobal()
			if err != nil || len(repos) == 0 {
				return fmt.Errorf("no repos registered")
			}

			var options []huh.Option[string]
			for _, name := range sortedRepoNames(repos) {
				options = append(options, huh.NewOption(name, name))
			}

			var selected []string
			multiField := huh.NewMultiSelect[string]().
				Title("Select repo(s) to unregister:").
				Options(options...).
				Value(&selected)

			err = huh.NewForm(huh.NewGroup(multiField)).
				WithOutput(os.Stderr).
				Run()
			if err != nil {
				return err
			}
			if len(selected) == 0 {
				return fmt.Errorf("no repos selected")
			}
			names = selected
		}

		for _, name := range names {
			if err := config.RemoveRepo(name); err != nil {
				fmt.Fprintf(os.Stderr, "error: %s\n", err)
				continue
			}
			fmt.Printf("Unregistered %s\n", name)
		}
		return nil
	},
}

func completeRepoNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	repos, err := config.LoadGlobal()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return sortedRepoNames(repos), cobra.ShellCompDirectiveNoFileComp
}

var repoLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List registered repos",
	RunE: func(cmd *cobra.Command, args []string) error {
		repos, err := config.LoadGlobal()
		if err != nil {
			return err
		}
		if len(repos) == 0 {
			fmt.Println("No repos registered. Run `tak repo add` inside a git repo.")
			return nil
		}
		for _, name := range sortedRepoNames(repos) {
			fmt.Printf("%s → %s\n", name, repos[name])
		}
		return nil
	},
}

func init() {
	repoCmd.AddCommand(repoAddCmd)
	repoCmd.AddCommand(repoRmCmd)
	repoCmd.AddCommand(repoLsCmd)
	rootCmd.AddCommand(repoCmd)
}
