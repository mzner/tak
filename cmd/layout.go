package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/mzner/tak/internal/config"
	"github.com/mzner/tak/internal/runner"
	"github.com/mzner/tak/internal/worktree"
	"github.com/spf13/cobra"
)

type layoutPreset struct {
	Name   string
	Layout string
	Panes  []config.PaneConfig
}

var presets = []layoutPreset{
	{
		Name:   "editor + shell",
		Layout: "main-horizontal",
		Panes: []config.PaneConfig{
			{Name: "editor", Command: "$EDITOR"},
			{Name: "shell", Command: ""},
		},
	},
	{
		Name:   "editor + dev server",
		Layout: "main-horizontal",
		Panes: []config.PaneConfig{
			{Name: "editor", Command: "$EDITOR"},
			{Name: "dev", Command: ""},
		},
	},
	{
		Name:   "dev + test + shell",
		Layout: "even-vertical",
		Panes: []config.PaneConfig{
			{Name: "dev", Command: ""},
			{Name: "test", Command: ""},
			{Name: "shell", Command: ""},
		},
	},
	{
		Name:   "editor + dev + shell",
		Layout: "main-vertical",
		Panes: []config.PaneConfig{
			{Name: "editor", Command: "$EDITOR"},
			{Name: "dev", Command: ""},
			{Name: "shell", Command: ""},
		},
	},
}

var layoutCmd = &cobra.Command{
	Use:   "layout",
	Short: "Configure tmux pane layout for this repo",
	Long: `Interactively configure the tmux pane layout used by tak open.

Choose from preset layouts or build a custom one. The configuration
is saved to .tak.yml and applies to all worktrees in this repo.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r := runner.NewExecRunner()
		wtSvc := worktree.NewService(r)

		repoRoot, err := wtSvc.RepoRoot()
		if err != nil {
			return errNotInRepo
		}

		cfg, err := config.Load(repoRoot, "")
		if err != nil {
			return err
		}

		tmuxCfg, err := runLayoutWizard()
		if err != nil {
			return err
		}

		if err := cfg.SetTmux(tmuxCfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		fmt.Printf("Layout saved to .tak.yml (%d panes, %s)\n", len(tmuxCfg.Panes), tmuxCfg.Layout)
		return nil
	},
}

func runLayoutWizard() (config.TmuxConfig, error) {
	options := make([]huh.Option[string], 0, len(presets)+1)
	for _, p := range presets {
		options = append(options, huh.NewOption(p.Name, p.Name))
	}
	options = append(options, huh.NewOption("Custom...", "custom"))

	var choice string
	selectField := huh.NewSelect[string]().
		Title("Select a layout:").
		Options(options...).
		Value(&choice)

	err := huh.NewForm(huh.NewGroup(selectField)).
		WithOutput(os.Stderr).
		Run()
	if err != nil {
		return config.TmuxConfig{}, err
	}

	if choice == "custom" {
		return runCustomBuilder()
	}

	for _, p := range presets {
		if p.Name == choice {
			return customizePreset(p)
		}
	}

	return config.TmuxConfig{}, fmt.Errorf("unknown preset")
}

func customizePreset(preset layoutPreset) (config.TmuxConfig, error) {
	panes := make([]config.PaneConfig, len(preset.Panes))
	copy(panes, preset.Panes)

	fields := make([]huh.Field, 0, len(panes))
	for i := range panes {
		fields = append(fields, huh.NewInput().
			Title(fmt.Sprintf("Command for pane %d (%s):", i+1, panes[i].Name)).
			Placeholder("leave empty for shell").
			Value(&panes[i].Command))
	}

	err := huh.NewForm(huh.NewGroup(fields...)).
		WithOutput(os.Stderr).
		Run()
	if err != nil {
		return config.TmuxConfig{}, err
	}

	return config.TmuxConfig{
		Layout: preset.Layout,
		Panes:  panes,
	}, nil
}

func runCustomBuilder() (config.TmuxConfig, error) {
	var numPanes string
	numField := huh.NewSelect[string]().
		Title("How many panes?").
		Options(
			huh.NewOption("2", "2"),
			huh.NewOption("3", "3"),
			huh.NewOption("4", "4"),
		).
		Value(&numPanes)

	var layout string
	layoutField := huh.NewSelect[string]().
		Title("Layout:").
		Options(
			huh.NewOption("even-vertical (stacked)", "even-vertical"),
			huh.NewOption("even-horizontal (side by side)", "even-horizontal"),
			huh.NewOption("main-vertical (big left, small right)", "main-vertical"),
			huh.NewOption("main-horizontal (big top, small bottom)", "main-horizontal"),
			huh.NewOption("tiled", "tiled"),
		).
		Value(&layout)

	err := huh.NewForm(huh.NewGroup(numField, layoutField)).
		WithOutput(os.Stderr).
		Run()
	if err != nil {
		return config.TmuxConfig{}, err
	}

	count := 2
	switch numPanes {
	case "3":
		count = 3
	case "4":
		count = 4
	}

	panes := make([]config.PaneConfig, count)
	fields := make([]huh.Field, count)
	for i := range panes {
		panes[i].Name = fmt.Sprintf("pane%d", i+1)
		fields[i] = huh.NewInput().
			Title(fmt.Sprintf("Command for pane %d:", i+1)).
			Placeholder("leave empty for shell").
			Value(&panes[i].Command)
	}

	err = huh.NewForm(huh.NewGroup(fields...)).
		WithOutput(os.Stderr).
		Run()
	if err != nil {
		return config.TmuxConfig{}, err
	}

	return config.TmuxConfig{
		Layout: layout,
		Panes:  panes,
	}, nil
}

func init() {
	rootCmd.AddCommand(layoutCmd)
}
