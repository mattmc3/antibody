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

func TestUpdateNonExistentLocalRepo(t *testing.T) {
	home := home()
	repo := NewGit(home, "zsh-users/zsh-completions")
	require.Error(t, repo.Update())
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
