package antibodylib

import (
	"bytes"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/mattmc3/antibody/internal/require"
)

// Integration tests clone real repos over the network.
// Run with `just test all`; `just test` skips them via -short.

func skipShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test: skipped in -short mode")
	}
}

// TestBundleLargeCorpus clones every repo in scripts/profiling/bundles.txt
// and verifies each active line produces at least one source statement.
// Repos without a canonical init file source every glob match, so the
// number of source lines can exceed the number of bundles.
func TestBundleLargeCorpus(t *testing.T) {
	skipShort(t)
	corpus, err := os.ReadFile("../scripts/profiling/bundles.txt")
	require.NoError(t, err)

	active := 0
	for _, line := range strings.Split(string(corpus), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			active++
		}
	}
	require.That(t, active > 100, "corpus unexpectedly small: %d", active)

	home := home(t)
	sh, err := New(
		home,
		bytes.NewBuffer(corpus),
		runtime.NumCPU(),
	).Bundle()
	require.NoError(t, err)

	sources := 0
	for _, line := range strings.Split(sh, "\n") {
		if strings.HasPrefix(line, "source ") {
			sources++
		}
	}
	require.That(t, sources >= active, "sources %d < active %d", sources, active)

	cloned, err := os.ReadDir(home)
	require.NoError(t, err)
	require.Equal(t, active, len(cloned))
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
		home := home(b)
		_, err := New(
			home,
			bytes.NewBufferString(bundles),
			runtime.NumCPU(),
		).Bundle()
		require.NoError(b, err)
	}
}
