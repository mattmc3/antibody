package project

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/mattmc3/antibody/internal/expect"
)

func TestLocalProject(t *testing.T) {
	proj, err := NewLocal("/tmp")
	Expect(t, NoError(err))
	Expect(t, NoError(proj.Download()))
	Expect(t, NoError(proj.Update()))
	Expect(t, Equals("/tmp", proj.Path()))
}

func TestLocalProjectRelativeToHome(t *testing.T) {
	proj, err := NewLocal("~/tmp")
	Expect(t, NoError(err))
	home, err := os.UserHomeDir()
	Expect(t, NoError(err))
	Expect(t, Equals(filepath.Join(home, "tmp"), proj.Path()))
}

func TestLocalProjectEnvVar(t *testing.T) {
	// $VAR names are never expanded and never hit the filesystem
	proj, err := New(t.TempDir(), "$MYPLUGS/myplug")
	Expect(t, NoError(err))
	Expect(t, Equals("$MYPLUGS/myplug", proj.Path()))
	Expect(t, NoError(proj.Download()))
	Expect(t, NoError(proj.Update()))
	_, ok := proj.(localProject)
	Expect(t, ok)
}

func TestLocalProjectRelativePath(t *testing.T) {
	proj, err := New(t.TempDir(), "./myplug")
	Expect(t, NoError(err))
	Expect(t, Equals("./myplug", proj.Path()))
	_, ok := proj.(localProject)
	Expect(t, ok)
}
