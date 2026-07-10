package project

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	. "github.com/mattmc3/antibody/internal/expect"
	"github.com/mattmc3/antibody/internal/gittest"
)

func TestList(t *testing.T) {
	upstream := gittest.New(t)
	home := home(t)
	proj, err := New(home, upstream.URL())
	Expect(t, NoError(err))
	Expect(t, NoError(proj.Download()))
	list, err := List(home)
	Expect(t, NoError(err))
	Expect(t, Equals(1, len(list)))
}

func TestUpdate(t *testing.T) {
	upstream := gittest.New(t)
	home := home(t)
	repo, err := New(home, upstream.URL())
	Expect(t, NoError(err))
	Expect(t, NoError(repo.Download()))
	Expect(t, NoError(repo.Update()))
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
			Expect(t, NoError(err))
			Expect(t, NoError(proj.Download()))
			Expect(t, NoError(Update(home, runtime.NumCPU())))
		})
	}
}

func TestUpdateHomeWithNoGitProjects(t *testing.T) {
	upstream := gittest.New(t)
	home := home(t)
	repo, err := New(home, upstream.URL())
	Expect(t, NoError(err))
	Expect(t, NoError(repo.Download()))
	Expect(t, NoError(os.RemoveAll(filepath.Join(repo.Path(), ".git"))))
	Expect(t, AnError(Update(home, runtime.NumCPU())))
}

func TestSymlinkedHome(t *testing.T) {
	real := t.TempDir()
	link := filepath.Join(t.TempDir(), "antibody-home")
	Expect(t, NoError(os.Symlink(real, link)))
	upstream := gittest.New(t)
	repo := newGitT(t, link, upstream.URL())
	Expect(t, NoError(repo.Download()))
	list, err := List(link)
	Expect(t, NoError(err))
	Expect(t, Equals(1, len(list)))
	Expect(t, NoError(Update(link, 1)))
}

func TestListEmptyFolder(t *testing.T) {
	home := home(t)
	list, err := List(home)
	Expect(t, NoError(err))
	Expect(t, Equals(0, len(list)))
}

func TestListNonExistentFolder(t *testing.T) {
	list, err := List("/tmp/asdasdadadwhateverwtff")
	Expect(t, AnError(err))
	Expect(t, Equals(0, len(list)))
}

func TestUpdateNonExistentHome(t *testing.T) {
	Expect(t, AnError(Update("/tmp/asdasdasdasksksksksnopeeeee", runtime.NumCPU())))
}

// Quoted annotation values must map to the same folder the bundle
// pipeline uses; purge and path go through this parser.
func TestCloneRootQuotedPin(t *testing.T) {
	home := home(t)
	sha := strings.Repeat("a", 40)
	root, err := CloneRoot(home, `ohmyzsh/ohmyzsh pin:"`+sha+`"`)
	Expect(t, NoError(err))
	Expect(t, Equals(filepath.Join(home, "https-COLON--SLASH--SLASH-github.com-SLASH-ohmyzsh-SLASH-ohmyzsh-SLASH-tree-SLASH-"+sha[:7]), root))
}

// NewLocal must take the name as-is; it used to split on spaces and
// silently truncate.
func TestNewLocalPathWithSpaces(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "dir with spaces")
	Expect(t, NoError(os.MkdirAll(dir, 0o755)))
	proj, err := NewLocal(dir)
	Expect(t, NoError(err))
	Expect(t, Equals(dir, proj.Path()))
}

// The bundle format cannot express a name containing spaces; such a
// line must error rather than silently resolve to a truncated path.
func TestNewSpacedPathLineErrors(t *testing.T) {
	_, err := New(t.TempDir(), "/tmp/dir with spaces kind:path")
	Expect(t, AnError(err))
}

func TestCloneRootPinnedRepo(t *testing.T) {
	home := home(t)
	sha := strings.Repeat("a", 40)
	root, err := CloneRoot(home, "ohmyzsh/ohmyzsh pin:"+sha)
	Expect(t, NoError(err))
	Expect(t, Equals(filepath.Join(home, "https-COLON--SLASH--SLASH-github.com-SLASH-ohmyzsh-SLASH-ohmyzsh-SLASH-tree-SLASH-"+sha[:7]), root))
}
