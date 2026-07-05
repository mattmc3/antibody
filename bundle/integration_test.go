package bundle

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Integration tests clone real repos over the network.
// Run with `just test all`; `just test` skips them via -short.

func skipShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test: skipped in -short mode")
	}
}

func TestSuccessfullGitBundles(t *testing.T) {
	skipShort(t)
	table := []struct {
		line, result string
	}{
		{
			"zsh-users/zsh-autosuggestions",
			"\nsource ",
		},
		{
			"zsh-users/zsh-autosuggestions kind:path",
			"export PATH=\"",
		},
		{
			"mattmc3/antidote kind:path branch:v1",
			"export PATH=\"",
		},
		{
			"zsh-users/zsh-autosuggestions kind:clone",
			"",
		},
		{
			"zsh-users/zsh-autosuggestions kind:fpath",
			"fpath+=( ",
		},
		{
			"docker/cli path:contrib/completion/zsh/_docker",
			"contrib/completion/zsh/_docker",
		},
		{
			"zsh-users/zsh-autosuggestions kind:defer",
			"zsh-defer source ",
		},
		{
			"sorin-ionescu/prezto kind:autoload path:modules/helper/functions",
			"builtin autoload -Uz ",
		},
	}
	for _, row := range table {
		t.Run(row.line, func(t *testing.T) {
			t.Parallel()
			home := home(t)
			bundle, err := New(home, row.line)
			require.NoError(t, err)
			result, err := bundle.Get()
			require.Contains(t, result, row.result)
			require.NoError(t, err)
		})
	}
}

func TestZshInvalidGitBundle(t *testing.T) {
	skipShort(t)
	home := home(t)
	bundle, err := New(home, "doesnotexist")
	require.NoError(t, err)
	_, err = bundle.Get()
	require.Error(t, err)
}

func TestZshBundleWithNoShFiles(t *testing.T) {
	skipShort(t)
	home := home(t)
	bundle, err := New(home, "mattmc3/antibody")
	require.NoError(t, err)
	_, err = bundle.Get()
	require.NoError(t, err)
}
