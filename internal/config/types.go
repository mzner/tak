package config

// Config holds the merged configuration used at runtime.
type Config struct {
	WorktreeBase    string
	BranchPrefix    string
	Pins            []string
	Repos           map[string]string
	RepoRoot        string
	LocalConfigPath string
}

// globalFile represents the structure of ~/.config/tak/config.yml.
type globalFile struct {
	WorktreeBase string            `yaml:"worktree_base"`
	Repos        map[string]string `yaml:"repos"`
}

// localFile represents the structure of .tak.yml in a repo root.
type localFile struct {
	WorktreeBase string   `yaml:"worktree_base"`
	BranchPrefix string   `yaml:"branch_prefix"`
	Pins         []string `yaml:"pins"`
}
