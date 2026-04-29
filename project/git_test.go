package project

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDownloadAllKinds(t *testing.T) {
	urls := []string{
		"zsh-users/zsh-completions",
		"http://github.com/zsh-users/zsh-completions",
		"http://github.com/zsh-users/zsh-completions.git",
		"https://github.com/zsh-users/zsh-completions",
		"https://github.com/zsh-users/zsh-completions.git",
		// FIXME: those fail on travis:
		// "git://github.com/zsh-users/zsh-completions.git", // git:// protocol deprecated/blocked
		// "git@gitlab.com:zsh-users/test.git",
		// "ssh://git@github.com/zsh-users/zsh-completions.git",
		// "git@github.com:zsh-users/zsh-completions.git",
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
	t.Skip("Skipping submodule test - can hang on CI")
	var home = home()
	var proj = NewGit(home, "ohmyzsh/ohmyzsh branch:master")
	var module = filepath.Join(proj.Path(), "plugins")
	require.NoError(t, proj.Download())
	require.NoError(t, proj.Update())
	files, err := ioutil.ReadDir(module)
	require.NoError(t, err)
	require.True(t, len(files) > 1)
}

func TestDownloadAnotherBranch(t *testing.T) {
	home := home()
	require.NoError(t, NewGit(home, "mattmc3/antidote branch:v1").Download())
}

func TestUpdateAnotherBranch(t *testing.T) {
	home := home()
	repo := NewGit(home, "mattmc3/antidote branch:v1")
	require.NoError(t, repo.Download())
	alreadyClonedRepo := NewClonedGit(home, "https-COLON--SLASH--SLASH-github.com-SLASH-mattmc3-SLASH-antidote")
	require.NoError(t, alreadyClonedRepo.Update())
}

func TestUpdateExistentLocalRepo(t *testing.T) {
	home := home()
	repo := NewGit(home, "zsh-users/zsh-completions")
	require.NoError(t, repo.Download())
	alreadyClonedRepo := NewClonedGit(home, "https-COLON--SLASH--SLASH-github.com-SLASH-zsh-users-SLASH-zsh-completions")
	require.NoError(t, alreadyClonedRepo.Update())
}

func TestUpdateNonExistentLocalRepo(t *testing.T) {
	home := home()
	repo := NewGit(home, "zsh-users/zsh-completions")
	require.Error(t, repo.Update())
}

func TestDownloadNonExistentRepo(t *testing.T) {
	home := home()
	repo := NewGit(home, "zsh-users/not-a-real-repo")
	require.Error(t, repo.Download())
}

func TestDownloadMalformedRepo(t *testing.T) {
	home := home()
	repo := NewGit(home, "doesn-not-exist-really branch:also-nope")
	require.Error(t, repo.Download())
}

func TestDownloadMultipleTimes(t *testing.T) {
	home := home()
	repo := NewGit(home, "zsh-users/zsh-completions")
	require.NoError(t, repo.Download())
	require.NoError(t, repo.Download())
	require.NoError(t, repo.Update())
}

func TestDownloadFolderNaming(t *testing.T) {
	home := home()
	repo := NewGit(home, "zsh-users/zsh-completions")
	require.Equal(
		t,
		home+"/https-COLON--SLASH--SLASH-github.com-SLASH-zsh-users-SLASH-zsh-completions",
		repo.Path(),
	)
}

func TestSubFolder(t *testing.T) {
	home := home()
	repo := NewGit(home, "ohmyzsh/ohmyzsh path:plugins/aws")
	require.True(t, strings.HasSuffix(repo.Path(), "plugins/aws"))
}

func TestPath(t *testing.T) {
	home := home()
	repo := NewGit(home, "docker/cli path:contrib/completion/zsh/_docker")
	require.True(t, strings.HasSuffix(repo.Path(), "contrib/completion/zsh/_docker"))
}

func TestMultipleSubFolders(t *testing.T) {
	home := home()
	require.NoError(t, NewGit(home, strings.Join([]string{
		"ohmyzsh/ohmyzsh path:plugins/aws",
		"ohmyzsh/ohmyzsh path:plugins/battery",
	}, "\n")).Download())
}

func home() string {
	home, err := ioutil.TempDir(os.TempDir(), "antibody")
	if err != nil {
		panic(err.Error())
	}
	return home
}
