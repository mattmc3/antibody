package project

import (
	"fmt"
	"os"
	"os/exec"
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

func TestDownloadPinnedRepo(t *testing.T) {
	repoPath, sha := createTempGitRepo(t)
	home := home()
	repo := NewGit(home, fmt.Sprintf("file://%s pin:%s", repoPath, sha))
	require.Contains(t, repo.Path(), "-SLASH-tree-SLASH-"+sha)
	require.NotContains(t, repo.Path(), "/tree/")
	if err := repo.Download(); err != nil {
		t.Fatalf("repo.Download failed: %#v", err)
	}

	configValue, err := gitConfigGet(repo.Path(), "antibody.pin")
	if err != nil {
		out, _ := exec.Command("git", "-C", repo.Path(), "config", "--list").CombinedOutput()
		t.Fatalf("gitConfigGet failed: %#v, config list: %q", err, string(out))
	}
	require.Equal(t, sha, configValue)

	cmd := exec.Command("git", "-C", repo.Path(), "rev-parse", "HEAD")
	out, err := cmd.Output()
	require.NoError(t, err)
	require.Equal(t, sha, strings.TrimSpace(string(out)))

	require.NoError(t, Update(home, 1))
}

func TestDownloadPinnedRepoInvalidPinCleansUp(t *testing.T) {
	repoPath, _ := createTempGitRepo(t)
	home := home()
	sha := "lmnop"
	repo := NewGit(home, fmt.Sprintf("file://%s pin:%s", repoPath, sha))
	require.Error(t, repo.Download())
	require.NoError(t, os.RemoveAll(repoPath))
	require.NoDirExists(t, repo.Path())
}

func createTempGitRepo(t *testing.T) (string, string) {
	dir, err := os.MkdirTemp(os.TempDir(), "gitrepo")
	require.NoError(t, err)

	cmd := exec.Command("git", "-C", dir, "init")
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-C", dir, "config", "user.name", "Test User")
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "-C", dir, "config", "user.email", "test@example.com")
	require.NoError(t, cmd.Run())

	filePath := filepath.Join(dir, "file.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("hello\n"), 0o644))

	cmd = exec.Command("git", "-C", dir, "add", "file.txt")
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "-C", dir, "commit", "-m", "initial")
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "-C", dir, "rev-parse", "HEAD")
	shaBytes, err := cmd.Output()
	require.NoError(t, err)

	return dir, strings.TrimSpace(string(shaBytes))
}

func home() string {
	home, err := os.MkdirTemp(os.TempDir(), "antibody")
	if err != nil {
		panic(err.Error())
	}
	return home
}
