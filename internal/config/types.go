package config

// Config holds the merged configuration used at runtime.
type Config struct {
	WorktreeBase    string
	BranchPrefix    string
	Pins            []string
	Repos           map[string]string
	RepoRoot        string
	LocalConfigPath string
	Tmux            TmuxConfig
	Hooks           HooksConfig
}

// HooksConfig holds lifecycle hook definitions.
type HooksConfig struct {
	PostCreate []HookAction `yaml:"post_create,omitempty"`
}

// HookAction represents a single hook step.
type HookAction struct {
	Type    string            `yaml:"type"`
	From    string            `yaml:"from,omitempty"`
	To      string            `yaml:"to,omitempty"`
	Command string            `yaml:"command,omitempty"`
	Env     map[string]string `yaml:"env,omitempty"`
	WorkDir string            `yaml:"work_dir,omitempty"`
}

// TmuxConfig describes the pane layout for tak open.
type TmuxConfig struct {
	Layout string       `yaml:"layout,omitempty"`
	Panes  []PaneConfig `yaml:"panes,omitempty"`
}

// PaneConfig describes a single tmux pane.
type PaneConfig struct {
	Name    string `yaml:"name,omitempty"`
	Command string `yaml:"command"`
}

// globalFile represents the structure of ~/.config/tak/config.yml.
type globalFile struct {
	WorktreeBase string            `yaml:"worktree_base"`
	Repos        map[string]string `yaml:"repos"`
}

// localFile represents the structure of .tak.yml in a repo root.
type localFile struct {
	WorktreeBase string      `yaml:"worktree_base"`
	BranchPrefix string      `yaml:"branch_prefix"`
	Pins         []string    `yaml:"pins"`
	Tmux         TmuxConfig  `yaml:"tmux,omitempty"`
	Hooks        HooksConfig `yaml:"hooks,omitempty"`
}
