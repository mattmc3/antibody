package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mattmc3/antibody/internal/require"
)

// Integration tests clone real repos over the network.
// Run with `just test all`; `just test` skips them via -short.

func skipShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test: skipped in -short mode")
	}
}

func TestDownloadAllKinds(t *testing.T) {
	skipShort(t)
	urls := []string{
		"zsh-users/zsh-completions",
		"http://github.com/zsh-users/zsh-completions",
		"http://github.com/zsh-users/zsh-completions.git",
		"https://github.com/zsh-users/zsh-completions",
		"https://github.com/zsh-users/zsh-completions.git",
	}
	// Skip SSH URL test in CI (requires SSH keys)
	if os.Getenv("CI") == "" {
		urls = append(urls, "git@github.com:zsh-users/zsh-completions.git")
	}
	for _, url := range urls {
		home := home(t)
		require.NoError(
			t,
			newGitT(t, home, url).Download(),
			"Repo "+url+" failed to download",
		)
	}
}

func TestDownloadSubmodules(t *testing.T) {
	skipShort(t)
	t.Skip("Skipping submodule test - can hang on CI")
	var home = home(t)
	var proj = newGitT(t, home, "ohmyzsh/ohmyzsh branch:master")
	var module = filepath.Join(proj.Path(), "plugins")
	require.NoError(t, proj.Download())
	require.NoError(t, proj.Update())
	files, err := os.ReadDir(module)
	require.NoError(t, err)
	require.That(t, len(files) > 1)
}
