package project

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mattmc3/antibody/internal/config"
	"github.com/mattmc3/antibody/internal/gittest"
	"github.com/stretchr/testify/require"
)

func TestUpdateNonExistentLocalRepo(t *testing.T) {
	home := home(t)
	repo := NewGit(home, "zsh-users/zsh-completions")
	require.Error(t, repo.Update())
}

func TestDownloadFolderNaming(t *testing.T) {
	home := home(t)
	repo := NewGit(home, "zsh-users/zsh-completions")
	require.Equal(
		t,
		home+"/https-COLON--SLASH--SLASH-github.com-SLASH-zsh-users-SLASH-zsh-completions",
		repo.Path(),
	)
}

func TestDownloadFolderNamingURLForms(t *testing.T) {
	home := home(t)
	table := []struct{ line, folder string }{
		{
			"zsh-users/zsh-completions",
			"https-COLON--SLASH--SLASH-github.com-SLASH-zsh-users-SLASH-zsh-completions",
		},
		{
			"http://github.com/foo/bar",
			"http-COLON--SLASH--SLASH-github.com-SLASH-foo-SLASH-bar",
		},
		{
			"https://github.com/foo/bar.git",
			"https-COLON--SLASH--SLASH-github.com-SLASH-foo-SLASH-bar.git",
		},
		{
			"git://github.com/foo/bar",
			"git-COLON--SLASH--SLASH-github.com-SLASH-foo-SLASH-bar",
		},
		{
			"ssh://git@github.com/foo/bar",
			"ssh-COLON--SLASH--SLASH-git-AT-github.com-SLASH-foo-SLASH-bar",
		},
		{
			"git@github.com:foo/bar.git",
			"ssh-COLON--SLASH--SLASH-git-AT-github.com-SLASH-foo-SLASH-bar.git",
		},
		{
			"file:///tmp/foo",
			"file-COLON--SLASH--SLASH--SLASH-tmp-SLASH-foo",
		},
	}
	for _, row := range table {
		t.Run(row.line, func(t *testing.T) {
			require.Equal(t, home+"/"+row.folder, NewGit(home, row.line).Path())
		})
	}
}

// useConfig points the config singleton at a fixture toml, restoring
// defaults when the test ends.
func useConfig(t *testing.T, toml string) {
	t.Helper()
	cfgDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(cfgDir, "antibody"), 0o755))
	// nolint: gosec
	require.NoError(t, os.WriteFile(filepath.Join(cfgDir, "antibody", "antibody.toml"), []byte(toml), 0o644))
	// Cleanup registered before Setenv so it runs after the env is
	// restored, reloading the default config.
	t.Cleanup(func() {
		_, err := config.Load()
		require.NoError(t, err)
	})
	t.Setenv("XDG_CONFIG_HOME", cfgDir)
	_, err := config.Load()
	require.NoError(t, err)
}

func TestFolderNamingSSHProtocol(t *testing.T) {
	useConfig(t, "[git]\nprotocol = \"ssh\"\n")
	home := home(t)
	require.Equal(
		t,
		home+"/ssh-COLON--SLASH--SLASH-git-AT-github.com-SLASH-foo-SLASH-bar",
		NewGit(home, "foo/bar").Path(),
	)
}

func TestFolderNamingCustomDomain(t *testing.T) {
	useConfig(t, "[git]\ndomain = \"gitlab.com\"\n")
	home := home(t)
	require.Equal(
		t,
		home+"/https-COLON--SLASH--SLASH-gitlab.com-SLASH-foo-SLASH-bar",
		NewGit(home, "foo/bar").Path(),
	)
}

func TestFolderNamingUnparseableURL(t *testing.T) {
	home := home(t)
	repo := NewGit(home, "https://github.com/%zz")
	require.Contains(t, repo.Path(), "-SLASH-unknown")
}

func TestSubFolder(t *testing.T) {
	home := home(t)
	repo := NewGit(home, "ohmyzsh/ohmyzsh path:plugins/aws")
	require.True(t, strings.HasSuffix(repo.Path(), "plugins/aws"))
}

func TestPath(t *testing.T) {
	home := home(t)
	repo := NewGit(home, "docker/cli path:contrib/completion/zsh/_docker")
	require.True(t, strings.HasSuffix(repo.Path(), "contrib/completion/zsh/_docker"))
}

func TestDownloadPinnedRepo(t *testing.T) {
	upstream := gittest.New(t)
	sha := upstream.HEAD()
	home := home(t)
	repo := NewGit(home, fmt.Sprintf("%s pin:%s", upstream.URL(), sha))
	require.Contains(t, repo.Path(), "-SLASH-tree-SLASH-"+sha)
	require.NotContains(t, repo.Path(), "/tree/")
	if err := repo.Download(); err != nil {
		t.Fatalf("repo.Download failed: %#v", err)
	}

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
	home := home(t)
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
	home := home(t)
	repo := NewGit(home, upstream.URL()+" branch:v1")
	require.NoError(t, repo.Download())
	folders, err := List(home)
	require.NoError(t, err)
	require.Len(t, folders, 1)
	require.NoError(t, NewClonedGit(home, folders[0]).Update())
}

func TestUpdateExistentLocalRepo(t *testing.T) {
	upstream := gittest.New(t)
	home := home(t)
	repo := NewGit(home, upstream.URL())
	require.NoError(t, repo.Download())
	folders, err := List(home)
	require.NoError(t, err)
	require.Len(t, folders, 1)
	require.NoError(t, NewClonedGit(home, folders[0]).Update())
}

func TestDownloadMultipleTimes(t *testing.T) {
	upstream := gittest.New(t)
	home := home(t)
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
	home := home(t)
	require.NoError(t, NewGit(home, strings.Join([]string{
		upstream.URL() + " path:plugins/aws",
		upstream.URL() + " path:plugins/battery",
	}, "\n")).Download())
}

func TestUpdatePullsUpstreamCommit(t *testing.T) {
	upstream := gittest.New(t)
	home := home(t)
	repo := NewGit(home, upstream.URL())
	require.NoError(t, repo.Download())
	upstream.WriteFile("new.txt", "new\n")
	newSHA := upstream.Commit("upstream work")
	require.NoError(t, repo.Update())
	cloneSHA, err := commit(repo.Path())
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(newSHA, cloneSHA))
}

func TestUpdateSkipsPinnedRepo(t *testing.T) {
	upstream := gittest.New(t)
	sha := upstream.HEAD()
	home := home(t)
	repo := NewGit(home, fmt.Sprintf("%s pin:%s", upstream.URL(), sha))
	require.NoError(t, repo.Download())
	upstream.WriteFile("new.txt", "new\n")
	upstream.Commit("upstream work")
	require.NoError(t, Update(home, 1))
	cloneSHA, err := commit(repo.Path())
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(sha, cloneSHA))
}

func TestDownloadPinnedToOlderCommit(t *testing.T) {
	upstream := gittest.New(t)
	upstream.Config("uploadpack.allowAnySHA1InWant", "true")
	oldSHA := upstream.HEAD()
	upstream.WriteFile("new.txt", "new\n")
	upstream.Commit("newer work")
	home := home(t)
	repo := NewGit(home, fmt.Sprintf("%s pin:%s", upstream.URL(), oldSHA))
	require.NoError(t, repo.Download())
	cloneSHA, err := commit(repo.Path())
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(oldSHA, cloneSHA))
}

func TestDownloadBranchTag(t *testing.T) {
	upstream := gittest.New(t)
	tagSHA := upstream.HEAD()
	upstream.Tag("v1.0.0")
	upstream.WriteFile("new.txt", "new\n")
	upstream.Commit("after tag")
	home := home(t)
	repo := NewGit(home, upstream.URL()+" branch:v1.0.0")
	require.NoError(t, repo.Download())
	cloneSHA, err := commit(repo.Path())
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(tagSHA, cloneSHA))
}

func TestDownloadAdvancePin(t *testing.T) {
	upstream := gittest.New(t)
	upstream.Config("uploadpack.allowAnySHA1InWant", "true")
	oldSHA := upstream.HEAD()
	upstream.WriteFile("new.txt", "new\n")
	newSHA := upstream.Commit("second")
	home := home(t)

	oldPin := NewGit(home, fmt.Sprintf("%s pin:%s", upstream.URL(), oldSHA))
	require.NoError(t, oldPin.Download())
	newPin := NewGit(home, fmt.Sprintf("%s pin:%s", upstream.URL(), newSHA))
	require.NoError(t, newPin.Download())

	require.NotEqual(t, oldPin.Path(), newPin.Path())
	for repo, sha := range map[Project]string{oldPin: oldSHA, newPin: newSHA} {
		cloneSHA, err := commit(repo.Path())
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(sha, cloneSHA))
	}
	require.NoError(t, Update(home, 1))
}

// A leftover antibody.pin config from older versions must not block
// updates; pin state lives in the folder name now.
func TestUpdateIgnoresStalePinConfig(t *testing.T) {
	upstream := gittest.New(t)
	home := home(t)
	repo := NewGit(home, upstream.URL())
	require.NoError(t, repo.Download())
	cmd := exec.Command("git", "-C", repo.Path(), "config", "antibody.pin", upstream.HEAD())
	require.NoError(t, cmd.Run())
	require.NoError(t, repo.Download())
	upstream.WriteFile("new.txt", "new\n")
	newSHA := upstream.Commit("upstream work")
	require.NoError(t, Update(home, 1))
	cloneSHA, err := commit(repo.Path())
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(newSHA, cloneSHA))
}

// Characterizes current behavior: an existing clone is never re-cloned or
// switched, so a changed branch: annotation is ignored.
func TestDownloadExistingCloneIgnoresBranchChange(t *testing.T) {
	upstream := gittest.New(t)
	mainSHA := upstream.HEAD()
	upstream.Branch("v1")
	upstream.WriteFile("v1.txt", "v1\n")
	upstream.Commit("v1 work")
	upstream.Checkout("main")
	home := home(t)
	require.NoError(t, NewGit(home, upstream.URL()).Download())
	repo := NewGit(home, upstream.URL()+" branch:v1")
	require.NoError(t, repo.Download())
	cloneSHA, err := commit(repo.Path())
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(mainSHA, cloneSHA))
}

// Characterizes a latent shallow-clone bug: after an upstream history
// rewrite, a depth-1 clone cannot pull and Update errors.
func TestUpdateAfterUpstreamRewriteFails(t *testing.T) {
	upstream := gittest.New(t)
	home := home(t)
	repo := NewGit(home, upstream.URL())
	require.NoError(t, repo.Download())
	upstream.WriteFile("new.txt", "new\n")
	upstream.Amend("rewritten history")
	require.Error(t, repo.Update())
}

func TestDownloadNonExistentRepo(t *testing.T) {
	home := home(t)
	repo := NewGit(home, "file:///this/path/does/not/exist")
	require.Error(t, repo.Download())
}

func TestDownloadNonExistentBranch(t *testing.T) {
	upstream := gittest.New(t)
	home := home(t)
	repo := NewGit(home, upstream.URL()+" branch:also-nope")
	require.Error(t, repo.Download())
}

func TestDownloadPinnedRepoInvalidPinCleansUp(t *testing.T) {
	upstream := gittest.New(t)
	home := home(t)
	sha := "lmnop"
	repo := NewGit(home, fmt.Sprintf("%s pin:%s", upstream.URL(), sha))
	require.Error(t, repo.Download())
	require.NoDirExists(t, repo.Path())
}

// Existing clones must load without spawning git at all; bundle runs on
// every shell startup and process spawns dominate its cost.
func TestDownloadExistingCloneNeedsNoGit(t *testing.T) {
	upstream := gittest.New(t)
	sha := upstream.HEAD()
	home := home(t)
	pinned := NewGit(home, fmt.Sprintf("%s pin:%s", upstream.URL(), sha))
	require.NoError(t, pinned.Download())
	plain := NewGit(home, upstream.URL())
	require.NoError(t, plain.Download())

	t.Setenv("PATH", "")
	require.NoError(t, plain.Download())
	require.NoError(t, pinned.Download())
}

func BenchmarkDownloadExistingClone(b *testing.B) {
	upstream := gittest.New(b)
	home := b.TempDir()
	repo := NewGit(home, upstream.URL())
	if err := repo.Download(); err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if err := repo.Download(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDownloadExistingPinnedClone(b *testing.B) {
	upstream := gittest.New(b)
	home := b.TempDir()
	repo := NewGit(home, fmt.Sprintf("%s pin:%s", upstream.URL(), upstream.HEAD()))
	if err := repo.Download(); err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if err := repo.Download(); err != nil {
			b.Fatal(err)
		}
	}
}

func home(t *testing.T) string {
	return t.TempDir()
}
