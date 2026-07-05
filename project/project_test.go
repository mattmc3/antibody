package project

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mattmc3/antibody/internal/gittest"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	upstream := gittest.New(t)
	home := home(t)
	proj, err := New(home, upstream.URL())
	require.NoError(t, err)
	require.NoError(t, proj.Download())
	list, err := List(home)
	require.NoError(t, err)
	require.Len(t, list, 1)
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
	repo := NewGit(link, upstream.URL())
	require.NoError(t, repo.Download())
	list, err := List(link)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.NoError(t, Update(link, 1))
}

func TestListEmptyFolder(t *testing.T) {
	home := home(t)
	list, err := List(home)
	require.NoError(t, err)
	require.Len(t, list, 0)
}

func TestListNonExistentFolder(t *testing.T) {
	list, err := List("/tmp/asdasdadadwhateverwtff")
	require.Error(t, err)
	require.Len(t, list, 0)
}

func TestUpdateNonExistentHome(t *testing.T) {
	require.Error(t, Update("/tmp/asdasdasdasksksksksnopeeeee", runtime.NumCPU()))
}

func TestCloneRootPinnedRepo(t *testing.T) {
	home := home(t)
	sha := strings.Repeat("a", 40)
	root := CloneRoot(home, "ohmyzsh/ohmyzsh pin:"+sha)
	require.Equal(t,
		filepath.Join(home, "https-COLON--SLASH--SLASH-github.com-SLASH-ohmyzsh-SLASH-ohmyzsh-SLASH-tree-SLASH-"+sha),
		root,
	)
}
