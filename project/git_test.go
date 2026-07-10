package project

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mattmc3/antibody/internal/config"
	. "github.com/mattmc3/antibody/internal/expect"
	"github.com/mattmc3/antibody/internal/gittest"
)

func TestUpdateNonExistentLocalRepo(t *testing.T) {
	home := home(t)
	repo := newGitT(t, home, "zsh-users/zsh-completions")
	Expect(t, AnError(repo.Update()))
}

func TestDownloadFolderNaming(t *testing.T) {
	home := home(t)
	repo := newGitT(t, home, "zsh-users/zsh-completions")
	Expect(t, Equals(home+"/https-COLON--SLASH--SLASH-github.com-SLASH-zsh-users-SLASH-zsh-completions", repo.Path()))
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
			Expect(t, Equals(home+"/"+row.folder, newGitT(t, home, row.line).Path()))
		})
	}
}

// useConfig points the config singleton at a fixture toml, restoring
// defaults when the test ends.
func useConfig(t *testing.T, toml string) {
	t.Helper()
	cfgDir := t.TempDir()
	Expect(t, NoError(os.MkdirAll(filepath.Join(cfgDir, "antibody"), 0o755)))
	// nolint: gosec
	Expect(t, NoError(os.WriteFile(filepath.Join(cfgDir, "antibody", "antibody.toml"), []byte(toml), 0o644)))
	// Cleanup registered before Setenv so it runs after the env is
	// restored, reloading the default config.
	t.Cleanup(func() {
		_, err := config.Load()
		Expect(t, NoError(err))
	})
	t.Setenv("XDG_CONFIG_HOME", cfgDir)
	_, err := config.Load()
	Expect(t, NoError(err))
}

func TestFolderNamingSSHProtocol(t *testing.T) {
	useConfig(t, "[git]\nprotocol = \"ssh\"\n")
	home := home(t)
	Expect(t, Equals(home+"/ssh-COLON--SLASH--SLASH-git-AT-github.com-SLASH-foo-SLASH-bar", newGitT(t, home, "foo/bar").Path()))
}

func TestFolderNamingCustomDomain(t *testing.T) {
	useConfig(t, "[git]\ndomain = \"gitlab.com\"\n")
	home := home(t)
	Expect(t, Equals(home+"/https-COLON--SLASH--SLASH-gitlab.com-SLASH-foo-SLASH-bar", newGitT(t, home, "foo/bar").Path()))
}

// An unparseable URL must error instead of silently cloning into a
// fabricated "unknown" folder.
func TestFolderNamingUnparseableURL(t *testing.T) {
	home := home(t)
	_, err := New(home, "https://github.com/%zz")
	Expect(t, AnError(err))
	Expect(t, Contains(err.Error(), "%zz"))
}

func TestSubFolder(t *testing.T) {
	home := home(t)
	repo := newGitT(t, home, "ohmyzsh/ohmyzsh path:plugins/aws")
	Expect(t, strings.HasSuffix(repo.Path(), "plugins/aws"))
}

func TestPath(t *testing.T) {
	home := home(t)
	repo := newGitT(t, home, "docker/cli path:contrib/completion/zsh/_docker")
	Expect(t, strings.HasSuffix(repo.Path(), "contrib/completion/zsh/_docker"))
}

func TestDownloadPinnedRepo(t *testing.T) {
	upstream := gittest.New(t)
	sha := upstream.HEAD()
	home := home(t)
	repo := newGitT(t, home, fmt.Sprintf("%s pin:%s", upstream.URL(), sha))
	Expect(t, Contains(repo.Path(), "-SLASH-tree-SLASH-"+sha[:7]))
	Expect(t, Not(Contains(repo.Path(), sha[7:])))
	Expect(t, Not(Contains(repo.Path(), "/tree/")))
	if err := repo.Download(); err != nil {
		t.Fatalf("repo.Download failed: %#v", err)
	}

	cmd := exec.Command("git", "-C", repo.Path(), "rev-parse", "HEAD")
	out, err := cmd.Output()
	Expect(t, NoError(err))
	Expect(t, Equals(sha, strings.TrimSpace(string(out))))

	Expect(t, NoError(Update(home, 1)))
}

func TestDownloadAnotherBranch(t *testing.T) {
	upstream := gittest.New(t)
	upstream.Branch("v1")
	upstream.WriteFile("v1.txt", "v1\n")
	branchSHA := upstream.Commit("v1 work")
	upstream.Checkout("main")
	home := home(t)
	repo := newGitT(t, home, upstream.URL()+" branch:v1")
	Expect(t, NoError(repo.Download()))
	cloneSHA, err := commit(repo.Path())
	Expect(t, NoError(err))
	Expect(t, strings.HasPrefix(branchSHA, cloneSHA))
}

func TestUpdateAnotherBranch(t *testing.T) {
	upstream := gittest.New(t)
	upstream.Branch("v1")
	upstream.WriteFile("v1.txt", "v1\n")
	upstream.Commit("v1 work")
	upstream.Checkout("main")
	home := home(t)
	repo := newGitT(t, home, upstream.URL()+" branch:v1")
	Expect(t, NoError(repo.Download()))
	folders, err := List(home)
	Expect(t, NoError(err))
	Expect(t, Equals(1, len(folders)))
	Expect(t, NoError(NewClonedGit(home, folders[0]).Update()))
}

func TestUpdateExistentLocalRepo(t *testing.T) {
	upstream := gittest.New(t)
	home := home(t)
	repo := newGitT(t, home, upstream.URL())
	Expect(t, NoError(repo.Download()))
	folders, err := List(home)
	Expect(t, NoError(err))
	Expect(t, Equals(1, len(folders)))
	Expect(t, NoError(NewClonedGit(home, folders[0]).Update()))
}

func TestDownloadMultipleTimes(t *testing.T) {
	upstream := gittest.New(t)
	home := home(t)
	repo := newGitT(t, home, upstream.URL())
	Expect(t, NoError(repo.Download()))
	Expect(t, NoError(repo.Download()))
	Expect(t, NoError(repo.Update()))
}

func TestMultipleSubFolders(t *testing.T) {
	upstream := gittest.New(t)
	upstream.WriteFile("plugins/aws/aws.plugin.zsh", "echo aws\n")
	upstream.WriteFile("plugins/battery/battery.plugin.zsh", "echo battery\n")
	upstream.Commit("add plugins")
	home := home(t)
	for _, line := range []string{
		upstream.URL() + " path:plugins/aws",
		upstream.URL() + " path:plugins/battery",
	} {
		Expect(t, NoError(newGitT(t, home, line).Download()))
	}
}

func TestUpdatePullsUpstreamCommit(t *testing.T) {
	upstream := gittest.New(t)
	home := home(t)
	repo := newGitT(t, home, upstream.URL())
	Expect(t, NoError(repo.Download()))
	upstream.WriteFile("new.txt", "new\n")
	newSHA := upstream.Commit("upstream work")
	Expect(t, NoError(repo.Update()))
	cloneSHA, err := commit(repo.Path())
	Expect(t, NoError(err))
	Expect(t, strings.HasPrefix(newSHA, cloneSHA))
}

func TestUpdateSkipsPinnedRepo(t *testing.T) {
	upstream := gittest.New(t)
	sha := upstream.HEAD()
	home := home(t)
	repo := newGitT(t, home, fmt.Sprintf("%s pin:%s", upstream.URL(), sha))
	Expect(t, NoError(repo.Download()))
	upstream.WriteFile("new.txt", "new\n")
	upstream.Commit("upstream work")
	Expect(t, NoError(Update(home, 1)))
	cloneSHA, err := commit(repo.Path())
	Expect(t, NoError(err))
	Expect(t, strings.HasPrefix(sha, cloneSHA))
}

func TestDownloadPinnedToOlderCommit(t *testing.T) {
	upstream := gittest.New(t)
	upstream.Config("uploadpack.allowAnySHA1InWant", "true")
	oldSHA := upstream.HEAD()
	upstream.WriteFile("new.txt", "new\n")
	upstream.Commit("newer work")
	home := home(t)
	repo := newGitT(t, home, fmt.Sprintf("%s pin:%s", upstream.URL(), oldSHA))
	Expect(t, NoError(repo.Download()))
	cloneSHA, err := commit(repo.Path())
	Expect(t, NoError(err))
	Expect(t, strings.HasPrefix(oldSHA, cloneSHA))
}

func TestDownloadBranchTag(t *testing.T) {
	upstream := gittest.New(t)
	tagSHA := upstream.HEAD()
	upstream.Tag("v1.0.0")
	upstream.WriteFile("new.txt", "new\n")
	upstream.Commit("after tag")
	home := home(t)
	repo := newGitT(t, home, upstream.URL()+" branch:v1.0.0")
	Expect(t, NoError(repo.Download()))
	cloneSHA, err := commit(repo.Path())
	Expect(t, NoError(err))
	Expect(t, strings.HasPrefix(tagSHA, cloneSHA))
}

func TestDownloadAdvancePin(t *testing.T) {
	upstream := gittest.New(t)
	upstream.Config("uploadpack.allowAnySHA1InWant", "true")
	oldSHA := upstream.HEAD()
	upstream.WriteFile("new.txt", "new\n")
	newSHA := upstream.Commit("second")
	home := home(t)

	oldPin := newGitT(t, home, fmt.Sprintf("%s pin:%s", upstream.URL(), oldSHA))
	Expect(t, NoError(oldPin.Download()))
	newPin := newGitT(t, home, fmt.Sprintf("%s pin:%s", upstream.URL(), newSHA))
	Expect(t, NoError(newPin.Download()))

	Expect(t, oldPin.Path() != newPin.Path(), "pins should clone to distinct dirs: %s", oldPin.Path())
	for repo, sha := range map[Project]string{oldPin: oldSHA, newPin: newSHA} {
		cloneSHA, err := commit(repo.Path())
		Expect(t, NoError(err))
		Expect(t, strings.HasPrefix(sha, cloneSHA))
	}
	Expect(t, NoError(Update(home, 1)))
}

// A leftover antibody.pin config from older versions must not block
// updates; pin state lives in the folder name now.
func TestUpdateIgnoresStalePinConfig(t *testing.T) {
	upstream := gittest.New(t)
	home := home(t)
	repo := newGitT(t, home, upstream.URL())
	Expect(t, NoError(repo.Download()))
	cmd := exec.Command("git", "-C", repo.Path(), "config", "antibody.pin", upstream.HEAD())
	Expect(t, NoError(cmd.Run()))
	Expect(t, NoError(repo.Download()))
	upstream.WriteFile("new.txt", "new\n")
	newSHA := upstream.Commit("upstream work")
	Expect(t, NoError(Update(home, 1)))
	cloneSHA, err := commit(repo.Path())
	Expect(t, NoError(err))
	Expect(t, strings.HasPrefix(newSHA, cloneSHA))
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
	Expect(t, NoError(newGitT(t, home, upstream.URL()).Download()))
	repo := newGitT(t, home, upstream.URL()+" branch:v1")
	Expect(t, NoError(repo.Download()))
	cloneSHA, err := commit(repo.Path())
	Expect(t, NoError(err))
	Expect(t, strings.HasPrefix(mainSHA, cloneSHA))
}

// Characterizes a latent shallow-clone bug: after an upstream history
// rewrite, a depth-1 clone cannot pull and Update errors.
func TestUpdateAfterUpstreamRewriteFails(t *testing.T) {
	upstream := gittest.New(t)
	home := home(t)
	repo := newGitT(t, home, upstream.URL())
	Expect(t, NoError(repo.Download()))
	upstream.WriteFile("new.txt", "new\n")
	upstream.Amend("rewritten history")
	Expect(t, AnError(repo.Update()))
}

func TestDownloadNonExistentRepo(t *testing.T) {
	home := home(t)
	repo := newGitT(t, home, "file:///this/path/does/not/exist")
	Expect(t, AnError(repo.Download()))
}

func TestDownloadNonExistentBranch(t *testing.T) {
	upstream := gittest.New(t)
	home := home(t)
	repo := newGitT(t, home, upstream.URL()+" branch:also-nope")
	Expect(t, AnError(repo.Download()))
}

func TestDownloadPinnedRepoInvalidPinCleansUp(t *testing.T) {
	upstream := gittest.New(t)
	home := home(t)
	sha := "lmnop"
	repo := newGitT(t, home, fmt.Sprintf("%s pin:%s", upstream.URL(), sha))
	Expect(t, AnError(repo.Download()))
	_, statErr := os.Stat(repo.Path())
	Expect(t, os.IsNotExist(statErr), "clone dir should not exist: %s", repo.Path())
}

// Existing clones must load without spawning git at all; bundle runs on
// every shell startup and process spawns dominate its cost.
func TestDownloadExistingCloneNeedsNoGit(t *testing.T) {
	upstream := gittest.New(t)
	sha := upstream.HEAD()
	home := home(t)
	pinned := newGitT(t, home, fmt.Sprintf("%s pin:%s", upstream.URL(), sha))
	Expect(t, NoError(pinned.Download()))
	plain := newGitT(t, home, upstream.URL())
	Expect(t, NoError(plain.Download()))

	t.Setenv("PATH", "")
	Expect(t, NoError(plain.Download()))
	Expect(t, NoError(pinned.Download()))
}

func BenchmarkDownloadExistingClone(b *testing.B) {
	upstream := gittest.New(b)
	home := b.TempDir()
	repo := newGitT(b, home, upstream.URL())
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
	repo := newGitT(b, home, fmt.Sprintf("%s pin:%s", upstream.URL(), upstream.HEAD()))
	if err := repo.Download(); err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if err := repo.Download(); err != nil {
			b.Fatal(err)
		}
	}
}

// newGitT parses a bundle line into its project, failing the test on
// parse errors. Replaces the removed NewGit constructor in tests.
func newGitT(tb testing.TB, home, line string) Project {
	tb.Helper()
	proj, err := New(home, line)
	Expect(tb, NoError(err))
	return proj
}

func home(t *testing.T) string {
	return t.TempDir()
}
