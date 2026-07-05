package antibodylib

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAntibody(t *testing.T) {
	home := home()
	bundles := []string{
		"# comments also are allowed",
		"zsh-users/zsh-completions kind:path # comment at the end of the line",
		"mattmc3/antidote kind:path branch:v1",
		"zsh-users/zsh-syntax-highlighting     kind:zsh",
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
	require.Contains(t, sh, `export PATH="`+home+`/https-COLON--SLASH--SLASH-github.com-SLASH-zsh-users-SLASH-zsh-completions:$PATH"`)
	require.Contains(t, sh, `export PATH="`+home+`/https-COLON--SLASH--SLASH-github.com-SLASH-mattmc3-SLASH-antidote:$PATH"`)
	// nolint: lll
	require.Contains(t, sh, `source `+home+`/https-COLON--SLASH--SLASH-github.com-SLASH-zsh-users-SLASH-zsh-syntax-highlighting/zsh-syntax-highlighting.plugin.zsh`)
}

func TestUsingDirective(t *testing.T) {
	home := home()
	repo, err := os.MkdirTemp(os.TempDir(), "antibody-using")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(repo))
	}()

	require.NoError(t, os.WriteFile(filepath.Join(repo, "myplugin.plugin.zsh"), []byte("echo hi"), 0644))

	bundles := []string{
		"using:" + repo,
		"git",
		"extract",
	}

	sh, err := New(home, bytes.NewBufferString(strings.Join(bundles, "\n")), runtime.NumCPU()).Bundle()
	require.NoError(t, err)
	require.Contains(t, sh, "source "+filepath.Join(repo, "myplugin.plugin.zsh"))
}

func TestAntibodyError(t *testing.T) {
	home := home()
	bundles := bytes.NewBufferString("invalid-repo")
	sh, err := New(home, bundles, runtime.NumCPU()).Bundle()
	require.Error(t, err)
	require.Empty(t, sh)
}

func TestMultipleRepositories(t *testing.T) {
	home := home()
	bundles := []string{
		"# this block is in alphabetic order",
		"unixorn/git-extra-commands kind:path",
		"zsh-users/zsh-autosuggestions",
		"zsh-users/zsh-completions kind:path",
		"mafredri/zsh-async",
		"rupa/z",
		"Tarrasch/zsh-bd",
		"zsh-users/zsh-completions",
		"zsh-users/zsh-autosuggestions",
		"",
		"ohmyzsh/ohmyzsh path:plugins/asdf",
		"ohmyzsh/ohmyzsh path:plugins/autoenv",
		"# these should be at last!",
		"sindresorhus/pure",
		"zsh-users/zsh-syntax-highlighting",
		"zsh-users/zsh-history-substring-search",
	}
	sh, err := New(
		home,
		bytes.NewBufferString(strings.Join(bundles, "\n")),
		runtime.NumCPU(),
	).Bundle()
	require.NoError(t, err)
	require.Len(t, strings.Split(sh, "\n"), 24)
}

// BenchmarkDownload-8   	       1	2907868713 ns/op	  480296 B/op	    2996 allocs/op v1
// BenchmarkDownload-8   	       1	2708120385 ns/op	  475904 B/op	    3052 allocs/op v2
func BenchmarkDownload(b *testing.B) {
	var bundles = strings.Join([]string{
		"ohmyzsh/ohmyzsh path:plugins/aws",
		"romkatv/gitstatus kind:path",
		"mattmc3/antidote branch:v1",
		"romkatv/zsh-bench kind:path",
		"",
		"# comment whatever",
		"unixorn/git-extra-commands kind:path",
		"ohmyzsh/ohmyzsh path:plugins/battery",
		"trystan2k/zsh-tab-title",
		"changyuheng/zsh-interactive-cd",
		"ohmyzsh/ohmyzsh path:plugins/asdf",
		"mafredri/zsh-async",
		"rupa/z",
		"Tarrasch/zsh-bd",
		"",
		"Linuxbrew/brew path:completions/zsh kind:fpath",
		"wbinglee/zsh-wakatime",
		"zsh-users/zsh-completions",
		"zsh-users/zsh-autosuggestions",
		"ohmyzsh/ohmyzsh path:plugins/autoenv",
		"# these should be at last!",
		"sindresorhus/pure",
		"zsh-users/zsh-syntax-highlighting",
		"zsh-users/zsh-history-substring-search",
	}, "\n")
	for i := 0; i < b.N; i++ {
		home := home()
		_, err := New(
			home,
			bytes.NewBufferString(bundles),
			runtime.NumCPU(),
		).Bundle()
		require.NoError(b, err)
	}
}

func TestDeferEnsureInjectedOnce(t *testing.T) {
	home := home()
	bundles := []string{
		"zsh-users/zsh-autosuggestions kind:defer",
		"sindresorhus/pure kind:defer",
	}
	sh, err := New(
		home,
		bytes.NewBufferString(strings.Join(bundles, "\n")),
		runtime.NumCPU(),
	).Bundle()
	require.NoError(t, err)
	require.Equal(t, 1, strings.Count(sh, "if ! (( $+functions[zsh-defer] )); then"))
	require.Contains(t, sh, "zsh-defer source ")
}

func TestBareCarriageReturnLineEndings(t *testing.T) {
	home := home()
	p1, err := os.MkdirTemp(os.TempDir(), "antibody-cr1")
	require.NoError(t, err)
	p2, err := os.MkdirTemp(os.TempDir(), "antibody-cr2")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(p1, "a.plugin.zsh"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(p2, "b.plugin.zsh"), []byte(""), 0644))

	sh, err := New(home, bytes.NewBufferString(p1+"\r"+p2+"\r"), 1).Bundle()
	require.NoError(t, err)
	require.Contains(t, sh, "source "+filepath.Join(p1, "a.plugin.zsh"))
	require.Contains(t, sh, "source "+filepath.Join(p2, "b.plugin.zsh"))
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

func home() string {
	home, err := os.MkdirTemp(os.TempDir(), "antibody")
	if err != nil {
		panic(err.Error())
	}
	return home
}
