package project

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListEmptyFolder(t *testing.T) {
	home := home()
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
	home := home()
	sha := strings.Repeat("a", 40)
	root := CloneRoot(home, "ohmyzsh/ohmyzsh pin:"+sha)
	require.Equal(t,
		filepath.Join(home, "https-COLON--SLASH--SLASH-github.com-SLASH-ohmyzsh-SLASH-ohmyzsh-SLASH-tree-SLASH-"+sha),
		root,
	)
}
