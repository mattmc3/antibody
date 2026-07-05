package project

import (
	"os"
	"path/filepath"
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
		home := home()
		require.NoError(
			t,
			NewGit(home, url).Download(),
			"Repo "+url+" failed to download",
		)
	}
}

func TestDownloadSubmodules(t *testing.T) {
	skipShort(t)
	t.Skip("Skipping submodule test - can hang on CI")
	var home = home()
	var proj = NewGit(home, "ohmyzsh/ohmyzsh branch:master")
	var module = filepath.Join(proj.Path(), "plugins")
	require.NoError(t, proj.Download())
	require.NoError(t, proj.Update())
	files, err := os.ReadDir(module)
	require.NoError(t, err)
	require.True(t, len(files) > 1)
}

func TestDownloadNonExistentRepo(t *testing.T) {
	skipShort(t)
	home := home()
	repo := NewGit(home, "zsh-users/not-a-real-repo")
	require.Error(t, repo.Download())
}

func TestDownloadMalformedRepo(t *testing.T) {
	skipShort(t)
	home := home()
	repo := NewGit(home, "doesn-not-exist-really branch:also-nope")
	require.Error(t, repo.Download())
}
