package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadGlobal reads only the global config file. Used for cross-repo access.
func LoadGlobal() (map[string]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	globalPath := filepath.Join(home, ".config", "tak", "config.yml")
	global, err := loadGlobalFile(globalPath)
	if err != nil {
		return nil, err
	}
	repos := make(map[string]string)
	for name, path := range global.Repos {
		repos[name] = expandHome(path)
	}
	return repos, nil
}

// AddRepo registers a repo in the global config.
func AddRepo(name string, path string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	globalPath := filepath.Join(home, ".config", "tak", "config.yml")

	global, err := loadGlobalFile(globalPath)
	if err != nil {
		return err
	}

	if global.Repos == nil {
		global.Repos = make(map[string]string)
	}
	global.Repos[name] = path

	if err := os.MkdirAll(filepath.Dir(globalPath), 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(global)
	if err != nil {
		return err
	}
	return os.WriteFile(globalPath, data, 0644)
}

// RemoveRepo unregisters a repo from the global config.
func RemoveRepo(name string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	globalPath := filepath.Join(home, ".config", "tak", "config.yml")

	global, err := loadGlobalFile(globalPath)
	if err != nil {
		return err
	}

	if global.Repos == nil {
		return fmt.Errorf("repo '%s' not found", name)
	}
	if _, ok := global.Repos[name]; !ok {
		return fmt.Errorf("repo '%s' not found", name)
	}
	delete(global.Repos, name)

	data, err := yaml.Marshal(global)
	if err != nil {
		return err
	}
	return os.WriteFile(globalPath, data, 0644)
}

// Load reads and merges the global and local config files.
func Load(repoRoot string, globalConfigPath string) (*Config, error) {
	if globalConfigPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		globalConfigPath = filepath.Join(home, ".config", "tak", "config.yml")
	}

	global, err := loadGlobalFile(globalConfigPath)
	if err != nil {
		return nil, err
	}

	localPath := filepath.Join(repoRoot, ".tak.yml")
	local, err := loadLocalFile(localPath)
	if err != nil {
		return nil, err
	}

	cfg := merge(global, local, repoRoot)
	cfg.LocalConfigPath = localPath
	return &cfg, nil
}

// IsPinned returns true if the given branch is in the pins list.
func (c *Config) IsPinned(branch string) bool {
	for _, p := range c.Pins {
		if p == branch {
			return true
		}
	}
	return false
}

// AddPin adds a branch to the pins list and persists to .tak.yml.
func (c *Config) AddPin(branch string) error {
	if c.IsPinned(branch) {
		return nil
	}
	c.Pins = append(c.Pins, branch)
	return c.persistPins()
}

// RemovePin removes a branch from the pins list and persists to .tak.yml.
func (c *Config) RemovePin(branch string) error {
	idx := -1
	for i, p := range c.Pins {
		if p == branch {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil
	}
	c.Pins = append(c.Pins[:idx], c.Pins[idx+1:]...)
	return c.persistPins()
}

// SetTmux updates the tmux layout config and persists to .tak.yml.
func (c *Config) SetTmux(tmuxCfg TmuxConfig) error {
	c.Tmux = tmuxCfg
	local, err := loadLocalFile(c.LocalConfigPath)
	if err != nil {
		return err
	}
	local.Tmux = tmuxCfg

	data, err := yaml.Marshal(local)
	if err != nil {
		return err
	}
	return os.WriteFile(c.LocalConfigPath, data, 0644)
}

func (c *Config) persistPins() error {
	local, err := loadLocalFile(c.LocalConfigPath)
	if err != nil {
		return err
	}
	local.Pins = c.Pins

	data, err := yaml.Marshal(local)
	if err != nil {
		return err
	}
	return os.WriteFile(c.LocalConfigPath, data, 0644)
}

func loadGlobalFile(path string) (globalFile, error) {
	var cfg globalFile
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	err = yaml.Unmarshal(data, &cfg)
	return cfg, err
}

func loadLocalFile(path string) (localFile, error) {
	var cfg localFile
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	err = yaml.Unmarshal(data, &cfg)
	return cfg, err
}

func expandHome(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func merge(global globalFile, local localFile, repoRoot string) Config {
	cfg := Config{
		WorktreeBase: expandHome(global.WorktreeBase),
		Repos:        global.Repos,
		RepoRoot:     repoRoot,
	}

	if local.WorktreeBase != "" {
		cfg.WorktreeBase = expandHome(local.WorktreeBase)
	}
	if local.BranchPrefix != "" {
		cfg.BranchPrefix = local.BranchPrefix
	}
	if local.Pins != nil {
		cfg.Pins = local.Pins
	}
	if len(local.Tmux.Panes) > 0 {
		cfg.Tmux = local.Tmux
	}
	if len(local.Hooks.PostCreate) > 0 {
		cfg.Hooks = local.Hooks
	}

	return cfg
}
