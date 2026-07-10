package project

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mattmc3/antibody/internal/gittest"
	"github.com/mattmc3/antibody/internal/require"
)

func TestList(t *testing.T) {
	upstream := gittest.New(t)
	home := home(t)
	proj, err := New(home, upstream.URL())
	require.NoError(t, err)
	require.NoError(t, proj.Download())
	list, err := List(home)
	require.NoError(t, err)
	require.Equal(t, 1, len(list))
}

func TestUpdate(t *testing.T) {
	upstream := gittest.New(t)
	home := home(t)
	repo, err := New(home, upstream.URL())
	require.NoError(t, err)
	require.NoError(t, repo.Download())
	require.NoError(t, repo.Update())
}

func TestUpdateHome(t *testing.T) {
	home := home(t)
	for _, tt := range []string{
		gittest.New(t).URL(),
		gittest.New(t).URL(),
		"/tmp",
	} {
		t.Run(tt, func(t *testing.T) {
			proj, err := New(home, tt)
			require.NoError(t, err)
			require.NoError(t, proj.Download())
			require.NoError(t, Update(home, runtime.NumCPU()))
		})
	}
}

func TestUpdateHomeWithNoGitProjects(t *testing.T) {
	upstream := gittest.New(t)
	home := home(t)
	repo, err := New(home, upstream.URL())
	require.NoError(t, err)
	require.NoError(t, repo.Download())
	require.NoError(t, os.RemoveAll(filepath.Join(repo.Path(), ".git")))
	require.Error(t, Update(home, runtime.NumCPU()))
}

func TestSymlinkedHome(t *testing.T) {
	real := t.TempDir()
	link := filepath.Join(t.TempDir(), "antibody-home")
	require.NoError(t, os.Symlink(real, link))
	upstream := gittest.New(t)
	repo := newGitT(t, link, upstream.URL())
	require.NoError(t, repo.Download())
	list, err := List(link)
	require.NoError(t, err)
	require.Equal(t, 1, len(list))
	require.NoError(t, Update(link, 1))
}

func TestListEmptyFolder(t *testing.T) {
	home := home(t)
	list, err := List(home)
	require.NoError(t, err)
	require.Equal(t, 0, len(list))
}

func TestListNonExistentFolder(t *testing.T) {
	list, err := List("/tmp/asdasdadadwhateverwtff")
	require.Error(t, err)
	require.Equal(t, 0, len(list))
}

func TestUpdateNonExistentHome(t *testing.T) {
	require.Error(t, Update("/tmp/asdasdasdasksksksksnopeeeee", runtime.NumCPU()))
}

// Quoted annotation values must map to the same folder the bundle
// pipeline uses; purge and path go through this parser.
func TestCloneRootQuotedPin(t *testing.T) {
	home := home(t)
	sha := strings.Repeat("a", 40)
	root, err := CloneRoot(home, `ohmyzsh/ohmyzsh pin:"`+sha+`"`)
	require.NoError(t, err)
	require.Equal(t,
		filepath.Join(home, "https-COLON--SLASH--SLASH-github.com-SLASH-ohmyzsh-SLASH-ohmyzsh-SLASH-tree-SLASH-"+sha[:7]),
		root,
	)
}

// NewLocal must take the name as-is; it used to split on spaces and
// silently truncate.
func TestNewLocalPathWithSpaces(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "dir with spaces")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	proj, err := NewLocal(dir)
	require.NoError(t, err)
	require.Equal(t, dir, proj.Path())
}

// The bundle format cannot express a name containing spaces; such a
// line must error rather than silently resolve to a truncated path.
func TestNewSpacedPathLineErrors(t *testing.T) {
	_, err := New(t.TempDir(), "/tmp/dir with spaces kind:path")
	require.Error(t, err)
}

func TestCloneRootPinnedRepo(t *testing.T) {
	home := home(t)
	sha := strings.Repeat("a", 40)
	root, err := CloneRoot(home, "ohmyzsh/ohmyzsh pin:"+sha)
	require.NoError(t, err)
	require.Equal(t,
		filepath.Join(home, "https-COLON--SLASH--SLASH-github.com-SLASH-ohmyzsh-SLASH-ohmyzsh-SLASH-tree-SLASH-"+sha[:7]),
		root,
	)
}
