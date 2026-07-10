package gittest

import (
	"path/filepath"
	"testing"

	. "github.com/mattmc3/antibody/internal/expect"
)

func TestNewRepo(t *testing.T) {
	r := New(t)
	Expect(t, DirExists(filepath.Join(r.Dir, ".git")))
	Expect(t, FileExists(filepath.Join(r.Dir, "file.txt")))
	Expect(t, Equals(40, len(r.HEAD())))
	Expect(t, Equals("file://"+r.Dir, r.URL()))
}

func TestCommitAdvancesHead(t *testing.T) {
	r := New(t)
	first := r.HEAD()
	r.WriteFile("myplugin.plugin.zsh", "echo myplugin\n")
	second := r.Commit("add plugin file")
	Expect(t, first != second, "commit should advance HEAD: %s", first)
	Expect(t, Equals(second, r.HEAD()))
}

func TestWriteFileCreatesSubdirs(t *testing.T) {
	r := New(t)
	r.WriteFile("plugins/aws/aws.plugin.zsh", "echo aws\n")
	Expect(t, FileExists(filepath.Join(r.Dir, "plugins", "aws", "aws.plugin.zsh")))
}

func TestBranchAndCheckout(t *testing.T) {
	r := New(t)
	main := r.HEAD()
	r.Branch("v1")
	r.WriteFile("v1.txt", "v1\n")
	branched := r.Commit("v1 work")
	Expect(t, main != branched, "branch should diverge from main: %s", main)
	r.Checkout("main")
	Expect(t, Equals(main, r.HEAD()))
}

func TestTag(t *testing.T) {
	r := New(t)
	r.Tag("v1.0.0")
	sha := r.git("rev-parse", "v1.0.0^{commit}")
	Expect(t, Contains(sha, r.HEAD()))
}
