package project

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/mattmc3/antibody/internal/gittest"
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
	upstream := gittest.New(t)
	sha := upstream.HEAD()
	home := home()
	repo := NewGit(home, fmt.Sprintf("%s pin:%s", upstream.URL(), sha))
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

func TestDownloadAnotherBranch(t *testing.T) {
	upstream := gittest.New(t)
	upstream.Branch("v1")
	upstream.WriteFile("v1.txt", "v1\n")
	branchSHA := upstream.Commit("v1 work")
	upstream.Checkout("main")
	home := home()
	repo := NewGit(home, upstream.URL()+" branch:v1")
	require.NoError(t, repo.Download())
	cloneSHA, err := commit(repo.Path())
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(branchSHA, cloneSHA))
}

func TestUpdateAnotherBranch(t *testing.T) {
	upstream := gittest.New(t)
	upstream.Branch("v1")
	upstream.WriteFile("v1.txt", "v1\n")
	upstream.Commit("v1 work")
	upstream.Checkout("main")
	home := home()
	repo := NewGit(home, upstream.URL()+" branch:v1")
	require.NoError(t, repo.Download())
	folders, err := List(home)
	require.NoError(t, err)
	require.Len(t, folders, 1)
	require.NoError(t, NewClonedGit(home, folders[0]).Update())
}

func TestUpdateExistentLocalRepo(t *testing.T) {
	upstream := gittest.New(t)
	home := home()
	repo := NewGit(home, upstream.URL())
	require.NoError(t, repo.Download())
	folders, err := List(home)
	require.NoError(t, err)
	require.Len(t, folders, 1)
	require.NoError(t, NewClonedGit(home, folders[0]).Update())
}

func TestDownloadMultipleTimes(t *testing.T) {
	upstream := gittest.New(t)
	home := home()
	repo := NewGit(home, upstream.URL())
	require.NoError(t, repo.Download())
	require.NoError(t, repo.Download())
	require.NoError(t, repo.Update())
}

func TestMultipleSubFolders(t *testing.T) {
	upstream := gittest.New(t)
	upstream.WriteFile("plugins/aws/aws.plugin.zsh", "echo aws\n")
	upstream.WriteFile("plugins/battery/battery.plugin.zsh", "echo battery\n")
	upstream.Commit("add plugins")
	home := home()
	require.NoError(t, NewGit(home, strings.Join([]string{
		upstream.URL() + " path:plugins/aws",
		upstream.URL() + " path:plugins/battery",
	}, "\n")).Download())
}

func TestDownloadPinnedRepoInvalidPinCleansUp(t *testing.T) {
	upstream := gittest.New(t)
	home := home()
	sha := "lmnop"
	repo := NewGit(home, fmt.Sprintf("%s pin:%s", upstream.URL(), sha))
	require.Error(t, repo.Download())
	require.NoDirExists(t, repo.Path())
}

func home() string {
	home, err := os.MkdirTemp(os.TempDir(), "antibody")
	if err != nil {
		panic(err.Error())
	}
	return home
}
