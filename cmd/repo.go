package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/mzner/tak/internal/config"
	"github.com/mzner/tak/internal/runner"
	"github.com/spf13/cobra"
)

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
	Run: func(cmd *cobra.Command, args []string) {
		var paths []string
		if len(args) == 0 {
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
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
	},
}

var repoRmCmd = &cobra.Command{
	Use:   "rm [name...]",
	Short: "Unregister repo(s)",
	Long:  "Unregister one or more repos. Without arguments, shows an interactive picker.",
	ValidArgsFunction: completeRepoNames,
	Run: func(cmd *cobra.Command, args []string) {
		var names []string
		if len(args) > 0 {
			names = args
		} else {
			repos, err := config.LoadGlobal()
			if err != nil || len(repos) == 0 {
				fmt.Fprintln(os.Stderr, "No repos registered.")
				os.Exit(1)
			}

			var options []huh.Option[string]
			for name := range repos {
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
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			if len(selected) == 0 {
				fmt.Fprintln(os.Stderr, "No repos selected.")
				os.Exit(1)
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
	},
}

func completeRepoNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	repos, err := config.LoadGlobal()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var names []string
	for name := range repos {
		names = append(names, name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

var repoLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List registered repos",
	Run: func(cmd *cobra.Command, args []string) {
		repos, err := config.LoadGlobal()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		if len(repos) == 0 {
			fmt.Println("No repos registered. Run `tak repo add` inside a git repo.")
			return
		}
		for name, path := range repos {
			fmt.Printf("%s → %s\n", name, path)
		}
	},
}

func init() {
	repoCmd.AddCommand(repoAddCmd)
	repoCmd.AddCommand(repoRmCmd)
	repoCmd.AddCommand(repoLsCmd)
	rootCmd.AddCommand(repoCmd)
}
