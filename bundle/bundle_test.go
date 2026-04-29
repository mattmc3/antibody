package bundle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSuccessfullGitBundles(t *testing.T) {
	table := []struct {
		line, result string
	}{
		{
			"zsh-users/zsh-autosuggestions",
			"zsh-autosuggestions.plugin.zsh\nfpath+=( ",
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
	home := home(t)
	bundle, err := New(home, "does not exist")
	require.NoError(t, err)
	_, err = bundle.Get()
	require.Error(t, err)
}

func TestZshLocalBundle(t *testing.T) {
	home := home(t)
	// nolint: gosec
	require.NoError(t, os.WriteFile(home+"/a.sh", []byte("echo 9"), 0644))
	bundle, err := New(home, home)
	require.NoError(t, err)
	result, err := bundle.Get()
	require.Contains(t, result, "a.sh")
	require.NoError(t, err)
}

func TestZshInvalidLocalBundle(t *testing.T) {
	home := home(t)
	bundle, err := New(home, "/asduhasd/asdasda")
	require.NoError(t, err)
	_, err = bundle.Get()
	require.Error(t, err)
}

func TestZshBundleWithNoShFiles(t *testing.T) {
	home := home(t)
	bundle, err := New(home, "mattmc3/antibody")
	require.NoError(t, err)
	_, err = bundle.Get()
	require.NoError(t, err)
}

func TestPathInvalidLocalBundle(t *testing.T) {
	home := home(t)
	bundle, err := New(home, "/asduhasd/asdasda kind:path")
	require.NoError(t, err)
	_, err = bundle.Get()
	require.Error(t, err)
}

func TestPathLocalBundle(t *testing.T) {
	home := home(t)
	// nolint: gosec
	require.NoError(t, os.WriteFile(filepath.Join(home, "whatever.sh"), []byte(""), 0644))
	bundle, err := New(home, home+" kind:path")
	require.NoError(t, err)
	result, err := bundle.Get()
	require.NoError(t, err)
	require.Equal(t, "export PATH=\""+home+":$PATH\"", result)
	require.NoError(t, err)
}

func TestDeferLocalBundle(t *testing.T) {
	home := home(t)
	// nolint: gosec
	require.NoError(t, os.WriteFile(filepath.Join(home, "myplugin.plugin.zsh"), []byte(""), 0644))
	bundle, err := New(home, home+" kind:defer")
	require.NoError(t, err)
	result, err := bundle.Get()
	require.NoError(t, err)
	require.Contains(t, result, "if ! (( $+functions[zsh-defer] )); then")
	require.Contains(t, result, "zsh-defer source ")
	for line := range strings.SplitSeq(result, "\n") {
		if strings.HasPrefix(line, "fpath") {
			require.False(t, strings.HasPrefix(line, "zsh-defer"), "fpath line should not be deferred: %q", line)
		}
	}
}

func home(t *testing.T) string {
	home, err := os.MkdirTemp(os.TempDir(), "antibody")
	require.NoError(t, err)
	return home
}
