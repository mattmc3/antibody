package antibodylib

import (
	"bytes"
	"io/ioutil"
	"os"
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
	files, err := ioutil.ReadDir(home)
	require.NoError(t, err)
	require.Len(t, files, 3)
	require.Contains(t, sh, `export PATH="/tmp:$PATH"`)
	require.Contains(t, sh, `export PATH="`+home+`/https-COLON--SLASH--SLASH-github.com-SLASH-zsh-users-SLASH-zsh-completions:$PATH"`)
	require.Contains(t, sh, `export PATH="`+home+`/https-COLON--SLASH--SLASH-github.com-SLASH-mattmc3-SLASH-antidote:$PATH"`)
	// nolint: lll
	require.Contains(t, sh, `source `+home+`/https-COLON--SLASH--SLASH-github.com-SLASH-zsh-users-SLASH-zsh-syntax-highlighting/zsh-syntax-highlighting.plugin.zsh`)
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
	home, err := ioutil.TempDir(os.TempDir(), "antibody")
	if err != nil {
		panic(err.Error())
	}
	return home
}
