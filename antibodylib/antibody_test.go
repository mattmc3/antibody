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
