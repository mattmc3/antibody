package shell

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeneratesInit(t *testing.T) {
	shell, err := Init()
	require.NoError(t, err)
	require.NotEmpty(t, shell)
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
