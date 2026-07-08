package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalProject(t *testing.T) {
	proj, err := NewLocal("/tmp")
	require.NoError(t, err)
	require.NoError(t, proj.Download())
	require.NoError(t, proj.Update())
	require.Equal(t, "/tmp", proj.Path())
}

func TestLocalProjectRelativeToHome(t *testing.T) {
	proj, err := NewLocal("~/tmp")
	require.NoError(t, err)
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(home, "tmp"), proj.Path())
}

func TestLocalProjectEnvVar(t *testing.T) {
	// $VAR names are never expanded and never hit the filesystem
	proj, err := New(t.TempDir(), "$MYPLUGS/myplug")
	require.NoError(t, err)
	require.Equal(t, "$MYPLUGS/myplug", proj.Path())
	require.NoError(t, proj.Download())
	require.NoError(t, proj.Update())
	_, ok := proj.(localProject)
	require.True(t, ok)
}

func TestLocalProjectRelativePath(t *testing.T) {
	proj, err := New(t.TempDir(), "./myplug")
	require.NoError(t, err)
	require.Equal(t, "./myplug", proj.Path())
	_, ok := proj.(localProject)
	require.True(t, ok)
}
