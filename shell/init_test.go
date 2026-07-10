package shell

import (
	"testing"

	"github.com/mattmc3/antibody/internal/require"
)

func TestGeneratesInit(t *testing.T) {
	shell, err := Init()
	require.NoError(t, err)
	require.That(t, shell != "", "init script is empty")
}

// Args must be quoted or arguments with spaces word-split in zsh.
func TestInitQuotesArgs(t *testing.T) {
	shell, err := Init()
	require.NoError(t, err)
	require.Contains(t, shell, `"$@"`)
	require.NotMatch(t, `\$@[^"]`, shell)
}

func TestInitUsesCompletionsSubcommand(t *testing.T) {
	shell, err := Init()
	require.NoError(t, err)
	require.Contains(t, shell, "completions zsh")
	require.Contains(t, shell, "--fpath")
	require.NotContains(t, shell, "compctl")
	// skip completion setup when _antibody is already defined
	require.Contains(t, shell, "$+functions[_antibody]")
}
