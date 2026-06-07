package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate_Zsh(t *testing.T) {
	output, err := Generate("zsh")
	require.NoError(t, err)
	assert.Contains(t, output, "tak()")
	assert.Contains(t, output, `"$1" = "cd"`)
	assert.Contains(t, output, "command tak cd")
	assert.Contains(t, output, `cd "$dir"`)
}

func TestGenerate_Bash(t *testing.T) {
	output, err := Generate("bash")
	require.NoError(t, err)
	assert.Contains(t, output, "tak()")
	assert.Contains(t, output, "command tak cd")
}

func TestGenerate_Fish(t *testing.T) {
	output, err := Generate("fish")
	require.NoError(t, err)
	assert.Contains(t, output, "function tak")
	assert.Contains(t, output, "command tak cd")
	assert.Contains(t, output, "cd $dir")
}

func TestGenerate_Unknown(t *testing.T) {
	_, err := Generate("powershell")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported shell")
}

func TestGenerate_ZshAndBashSameOutput(t *testing.T) {
	zsh, _ := Generate("zsh")
	bash, _ := Generate("bash")
	assert.Equal(t, zsh, bash)
}
