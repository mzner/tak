// Package config handles loading and merging tak configuration files.
//
// tak uses two optional config files that are merged together:
//
//   - Global: ~/.config/tak/config.yml (user-wide defaults)
//   - Local: .tak.yml in the git repo root (per-repo overrides)
//
// Merge logic: local values override global values.
// Missing values fall back to hardcoded defaults.
//
// Pins (persistent worktrees excluded from gc) are stored in the
// local config file (.tak.yml), making them recoverable if the
// state cache (.tak/state.json) is deleted.
package config
