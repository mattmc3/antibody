package gittest

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRepo(t *testing.T) {
	r := New(t)
	require.DirExists(t, filepath.Join(r.Dir, ".git"))
	require.FileExists(t, filepath.Join(r.Dir, "file.txt"))
	require.Len(t, r.HEAD(), 40)
	require.Equal(t, "file://"+r.Dir, r.URL())
}

func TestCommitAdvancesHead(t *testing.T) {
	r := New(t)
	first := r.HEAD()
	r.WriteFile("myplugin.plugin.zsh", "echo myplugin\n")
	second := r.Commit("add plugin file")
	require.NotEqual(t, first, second)
	require.Equal(t, second, r.HEAD())
}

func TestWriteFileCreatesSubdirs(t *testing.T) {
	r := New(t)
	r.WriteFile("plugins/aws/aws.plugin.zsh", "echo aws\n")
	require.FileExists(t, filepath.Join(r.Dir, "plugins", "aws", "aws.plugin.zsh"))
}

func TestBranchAndCheckout(t *testing.T) {
	r := New(t)
	main := r.HEAD()
	r.Branch("v1")
	r.WriteFile("v1.txt", "v1\n")
	branched := r.Commit("v1 work")
	require.NotEqual(t, main, branched)
	r.Checkout("main")
	require.Equal(t, main, r.HEAD())
}

func TestTag(t *testing.T) {
	r := New(t)
	r.Tag("v1.0.0")
	sha := r.git("rev-parse", "v1.0.0^{commit}")
	require.Contains(t, sha, r.HEAD())
}
