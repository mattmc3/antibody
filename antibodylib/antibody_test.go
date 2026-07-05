package antibodylib

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mattmc3/antibody/internal/config"
	"github.com/mattmc3/antibody/internal/gittest"
	"github.com/stretchr/testify/require"
)

// escapedDir returns the clone folder a URL lands in under home.
func escapedDir(home, url string) string {
	return filepath.Join(home, strings.NewReplacer(
		":", "-COLON-", "/", "-SLASH-", "@", "-AT-",
	).Replace(url))
}

// pluginRepo builds a repo containing a sourceable plugin init file.
func pluginRepo(t *testing.T) *gittest.Repo {
	t.Helper()
	r := gittest.New(t)
	r.WriteFile("myplugin.plugin.zsh", "echo myplugin\n")
	r.Commit("add plugin file")
	return r
}

func TestAntibody(t *testing.T) {
	rPath := gittest.New(t)
	rBranch := gittest.New(t)
	rBranch.Branch("v1")
	rBranch.WriteFile("v1.txt", "v1\n")
	rBranch.Commit("v1 work")
	rBranch.Checkout("main")
	rZsh := pluginRepo(t)

	home := home(t)
	bundles := []string{
		"# comments also are allowed",
		rPath.URL() + " kind:path # comment at the end of the line",
		rBranch.URL() + " kind:path branch:v1",
		rZsh.URL() + "     kind:zsh",
		"",
		"        ",
		"  # trick play",
		"/tmp kind:path",
	}
	sh, err := New(
		home,
		bytes.NewBufferString(strings.Join(bundles, "\n")),
		runtime.NumCPU(),
	).Bundle()
	require.NoError(t, err)
	files, err := os.ReadDir(home)
	require.NoError(t, err)
	require.Len(t, files, 3)
	require.Contains(t, sh, `export PATH="/tmp:$PATH"`)
	require.Contains(t, sh, `export PATH="`+escapedDir(home, rPath.URL())+`:$PATH"`)
	require.Contains(t, sh, `export PATH="`+escapedDir(home, rBranch.URL())+`:$PATH"`)
	require.Contains(t, sh, `source "`+filepath.Join(escapedDir(home, rZsh.URL()), "myplugin.plugin.zsh")+`"`)
}

func TestAntibodyError(t *testing.T) {
	home := home(t)
	bundles := bytes.NewBufferString("file:///this/path/does/not/exist")
	sh, err := New(home, bundles, runtime.NumCPU()).Bundle()
	require.Error(t, err)
	require.Empty(t, sh)
}

func TestMultipleRepositories(t *testing.T) {
	rPath := gittest.New(t)
	rDupe := pluginRepo(t)
	rInner := gittest.New(t)
	rInner.WriteFile("plugins/a/a.plugin.zsh", "echo a\n")
	rInner.WriteFile("plugins/b/b.plugin.zsh", "echo b\n")
	rInner.Commit("add plugins")
	rLast := pluginRepo(t)

	home := home(t)
	bundles := []string{
		"# this block is in alphabetic order",
		rPath.URL() + " kind:path",
		rDupe.URL(),
		rDupe.URL(),
		"",
		rInner.URL() + " path:plugins/a",
		rInner.URL() + " path:plugins/b",
		"# these should be at last!",
		rLast.URL(),
	}
	sh, err := New(
		home,
		bytes.NewBufferString(strings.Join(bundles, "\n")),
		runtime.NumCPU(),
	).Bundle()
	require.NoError(t, err)
	// path repo: 1 line; dupe twice: 4; two inner paths: 4; last: 2
	require.Len(t, strings.Split(sh, "\n"), 11)
	require.True(
		t,
		strings.HasSuffix(sh, filepath.Join(escapedDir(home, rLast.URL()), "myplugin.plugin.zsh")+`"`),
		"last bundle should come last, got: %s", sh,
	)
	files, err := os.ReadDir(home)
	require.NoError(t, err)
	require.Len(t, files, 4)
}

// useDeferFixture points the config singleton at a local defer bundle so
// the ensure block clones offline, restoring defaults when the test ends.
func useDeferFixture(t *testing.T) *gittest.Repo {
	t.Helper()
	deferRepo := gittest.New(t)
	deferRepo.WriteFile("zsh-defer.plugin.zsh", "echo defer\n")
	deferRepo.Commit("add plugin file")

	cfgDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(cfgDir, "antibody"), 0o755))
	toml := "[defer]\nbundle = \"" + deferRepo.URL() + "\"\n"
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
	return deferRepo
}

func TestDeferEnsureInjectedOnce(t *testing.T) {
	useDeferFixture(t)
	rA := pluginRepo(t)
	rB := pluginRepo(t)
	home := home(t)
	bundles := []string{
		rA.URL() + " kind:defer",
		rB.URL() + " kind:defer",
	}
	sh, err := New(
		home,
		bytes.NewBufferString(strings.Join(bundles, "\n")),
		runtime.NumCPU(),
	).Bundle()
	require.NoError(t, err)
	require.Equal(t, 1, strings.Count(sh, "if ! (( $+functions[zsh-defer] )); then"))
	require.Equal(t, 2, strings.Count(sh, "zsh-defer source "))
}

func TestUsingDirective(t *testing.T) {
	home := home(t)
	repo := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(repo, "myplugin.plugin.zsh"), []byte("echo hi"), 0644))

	bundles := []string{
		"using:" + repo,
		"git",
		"extract",
	}

	sh, err := New(home, bytes.NewBufferString(strings.Join(bundles, "\n")), runtime.NumCPU()).Bundle()
	require.NoError(t, err)
	require.Contains(t, sh, `source "`+filepath.Join(repo, "myplugin.plugin.zsh")+`"`)
}

func TestBareCarriageReturnLineEndings(t *testing.T) {
	home := home(t)
	p1 := t.TempDir()
	p2 := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(p1, "a.plugin.zsh"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(p2, "b.plugin.zsh"), []byte(""), 0644))

	sh, err := New(home, bytes.NewBufferString(p1+"\r"+p2+"\r"), 1).Bundle()
	require.NoError(t, err)
	require.Contains(t, sh, `source "`+filepath.Join(p1, "a.plugin.zsh")+`"`)
	require.Contains(t, sh, `source "`+filepath.Join(p2, "b.plugin.zsh")+`"`)
}

func TestHome(t *testing.T) {
	h, err := Home()
	require.NoError(t, err)
	require.Contains(t, h, "antibody")
}

func TestHomeFromEnvironmentVariable(t *testing.T) {
	require.NoError(t, os.Setenv("ANTIBODY_HOME", "/tmp"))
	h, err := Home()
	require.NoError(t, err)
	require.Equal(t, "/tmp", h)
}

func home(tb testing.TB) string {
	tb.Helper()
	return tb.TempDir()
}
