package project

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func TestList(t *testing.T) {
	skipShort(t)
	home := home()
	proj, err := New(home, "mattmc3/antidote branch:v1")
	require.NoError(t, err)
	require.NoError(t, proj.Download())
	list, err := List(home)
	require.NoError(t, err)
	require.Len(t, list, 1)
}

func TestUpdate(t *testing.T) {
	skipShort(t)
	home := home()
	repo, err := New(home, "zsh-users/zsh-completions")
	require.NoError(t, err)
	require.NoError(t, repo.Download())
	require.NoError(t, repo.Update())
}

func TestUpdateHome(t *testing.T) {
	skipShort(t)
	home := home()
	for _, tt := range []string{
		"zsh-users/zsh-autosuggestions",
		"zsh-users/zsh-completions",
		"/tmp",
	} {
		tt := tt
		t.Run(tt, func(t *testing.T) {
			proj, err := New(home, tt)
			require.NoError(t, err)
			require.NoError(t, proj.Download())
			require.NoError(t, Update(home, runtime.NumCPU()))
		})
	}
}

func TestUpdateHomeWithNoGitProjects(t *testing.T) {
	skipShort(t)
	home := home()
	repo, err := New(home, "zsh-users/zsh-autosuggestions")
	require.NoError(t, err)
	require.NoError(t, repo.Download())
	require.NoError(t, os.RemoveAll(filepath.Join(repo.Path(), ".git")))
	require.Error(t, Update(home, runtime.NumCPU()))
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

func TestDownloadAnotherBranch(t *testing.T) {
	skipShort(t)
	home := home()
	require.NoError(t, NewGit(home, "mattmc3/antidote branch:v1").Download())
}

func TestUpdateAnotherBranch(t *testing.T) {
	skipShort(t)
	home := home()
	repo := NewGit(home, "mattmc3/antidote branch:v1")
	require.NoError(t, repo.Download())
	alreadyClonedRepo := NewClonedGit(home, "https-COLON--SLASH--SLASH-github.com-SLASH-mattmc3-SLASH-antidote")
	require.NoError(t, alreadyClonedRepo.Update())
}

func TestUpdateExistentLocalRepo(t *testing.T) {
	skipShort(t)
	home := home()
	repo := NewGit(home, "zsh-users/zsh-completions")
	require.NoError(t, repo.Download())
	alreadyClonedRepo := NewClonedGit(home, "https-COLON--SLASH--SLASH-github.com-SLASH-zsh-users-SLASH-zsh-completions")
	require.NoError(t, alreadyClonedRepo.Update())
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

func TestDownloadMultipleTimes(t *testing.T) {
	skipShort(t)
	home := home()
	repo := NewGit(home, "zsh-users/zsh-completions")
	require.NoError(t, repo.Download())
	require.NoError(t, repo.Download())
	require.NoError(t, repo.Update())
}

func TestMultipleSubFolders(t *testing.T) {
	skipShort(t)
	home := home()
	require.NoError(t, NewGit(home, strings.Join([]string{
		"ohmyzsh/ohmyzsh path:plugins/aws",
		"ohmyzsh/ohmyzsh path:plugins/battery",
	}, "\n")).Download())
}
