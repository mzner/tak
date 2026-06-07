package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadLocalConfig(t *testing.T) {
	cfg, err := loadLocalFile("../../testdata/config/local.yml")
	require.NoError(t, err)
	assert.Equal(t, "", cfg.WorktreeBase)
	assert.Equal(t, "feature/", cfg.BranchPrefix)
	assert.Equal(t, []string{"feature/auth", "long-running/experiment"}, cfg.Pins)
}

func TestLoadGlobalConfig(t *testing.T) {
	cfg, err := loadGlobalFile("../../testdata/config/global.yml")
	require.NoError(t, err)
	assert.Equal(t, "/Users/dev/worktrees", cfg.WorktreeBase)
	assert.Equal(t, "~/projects/web", cfg.Repos["web"])
}

func TestLoadLocalConfig_FileNotFound(t *testing.T) {
	cfg, err := loadLocalFile("/nonexistent/path/.tak.yml")
	require.NoError(t, err)
	assert.Equal(t, localFile{}, cfg)
}

func TestLoadGlobalConfig_FileNotFound(t *testing.T) {
	cfg, err := loadGlobalFile("/nonexistent/path/config.yml")
	require.NoError(t, err)
	assert.Equal(t, globalFile{}, cfg)
}

func TestMerge_LocalOverridesGlobal(t *testing.T) {
	global := globalFile{
		WorktreeBase: "/global/worktrees",
		Repos:        map[string]string{"web": "~/projects/web"},
	}
	local := localFile{
		WorktreeBase: "/local/override",
		BranchPrefix: "feature/",
		Pins:         []string{"feature/auth"},
	}

	cfg := merge(global, local, "/Users/dev/projects/web")
	assert.Equal(t, "/local/override", cfg.WorktreeBase)
	assert.Equal(t, "feature/", cfg.BranchPrefix)
	assert.Equal(t, []string{"feature/auth"}, cfg.Pins)
	assert.Equal(t, "~/projects/web", cfg.Repos["web"])
}

func TestMerge_GlobalUsedWhenLocalEmpty(t *testing.T) {
	global := globalFile{
		WorktreeBase: "/global/worktrees",
	}
	local := localFile{
		WorktreeBase: "",
	}

	cfg := merge(global, local, "/Users/dev/projects/web")
	assert.Equal(t, "/global/worktrees", cfg.WorktreeBase)
}

func TestIsPinned(t *testing.T) {
	cfg := &Config{
		Pins: []string{"feature/auth", "long-running/experiment"},
	}
	assert.True(t, cfg.IsPinned("feature/auth"))
	assert.True(t, cfg.IsPinned("long-running/experiment"))
	assert.False(t, cfg.IsPinned("feature/other"))
}

func TestAddPin(t *testing.T) {
	dir := t.TempDir()
	localPath := filepath.Join(dir, ".tak.yml")
	err := os.WriteFile(localPath, []byte("pins: []\n"), 0644)
	require.NoError(t, err)

	cfg := &Config{
		Pins:            []string{},
		LocalConfigPath: localPath,
	}

	err = cfg.AddPin("feature/auth")
	require.NoError(t, err)
	assert.Equal(t, []string{"feature/auth"}, cfg.Pins)

	reloaded, err := loadLocalFile(localPath)
	require.NoError(t, err)
	assert.Contains(t, reloaded.Pins, "feature/auth")
}

func TestAddPin_AlreadyPinned(t *testing.T) {
	cfg := &Config{
		Pins:            []string{"feature/auth"},
		LocalConfigPath: "/tmp/doesnt-matter.yml",
	}
	err := cfg.AddPin("feature/auth")
	assert.NoError(t, err)
	assert.Equal(t, []string{"feature/auth"}, cfg.Pins)
}

func TestRemovePin(t *testing.T) {
	dir := t.TempDir()
	localPath := filepath.Join(dir, ".tak.yml")
	err := os.WriteFile(localPath, []byte("pins:\n  - feature/auth\n  - fix/bug\n"), 0644)
	require.NoError(t, err)

	cfg := &Config{
		Pins:            []string{"feature/auth", "fix/bug"},
		LocalConfigPath: localPath,
	}

	err = cfg.RemovePin("feature/auth")
	require.NoError(t, err)
	assert.Equal(t, []string{"fix/bug"}, cfg.Pins)

	reloaded, err := loadLocalFile(localPath)
	require.NoError(t, err)
	assert.Equal(t, []string{"fix/bug"}, reloaded.Pins)
}

func TestRemovePin_NotPinned(t *testing.T) {
	cfg := &Config{
		Pins:            []string{"feature/auth"},
		LocalConfigPath: "/tmp/doesnt-matter.yml",
	}
	err := cfg.RemovePin("not-pinned")
	assert.NoError(t, err)
	assert.Equal(t, []string{"feature/auth"}, cfg.Pins)
}
