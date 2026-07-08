// Package gittest builds throwaway local git repositories for tests.
// Repos live in test temp dirs and are cloneable via file:// URLs, so
// clone and update mechanics can be tested without network access.
package gittest

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Repo is a local git repository rooted at Dir. All helper methods fail
// the test on error.
type Repo struct {
	t   testing.TB
	Dir string
}

// New creates a git repository on branch main with an initial commit
// containing file.txt. The repo is removed automatically when the test
// finishes.
func New(t testing.TB) *Repo {
	t.Helper()
	r := &Repo{t: t, Dir: t.TempDir()}
	r.git("init", "-b", "main")
	r.git("config", "user.name", "Test User")
	r.git("config", "user.email", "test@example.com")
	r.git("config", "commit.gpgsign", "false")
	r.git("config", "tag.gpgsign", "false")
	r.WriteFile("file.txt", "hello\n")
	r.Commit("initial")
	return r
}

// URL returns the file:// URL for cloning the repo.
func (r *Repo) URL() string {
	return "file://" + r.Dir
}

// WriteFile writes a file at the given path relative to the repo root,
// creating parent directories as needed.
func (r *Repo) WriteFile(name, content string) {
	r.t.Helper()
	path := filepath.Join(r.Dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		r.t.Fatalf("gittest: mkdir for %s: %v", name, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		r.t.Fatalf("gittest: write %s: %v", name, err)
	}
}

// Commit stages all changes and commits them, returning the new HEAD SHA.
func (r *Repo) Commit(message string) string {
	r.t.Helper()
	r.git("add", "--all")
	r.git("commit", "-m", message)
	return r.HEAD()
}

// Branch creates a branch at HEAD and checks it out.
func (r *Repo) Branch(name string) {
	r.t.Helper()
	r.git("checkout", "-b", name)
}

// Checkout switches to an existing branch.
func (r *Repo) Checkout(name string) {
	r.t.Helper()
	r.git("checkout", name)
}

// Tag creates a lightweight tag at HEAD.
func (r *Repo) Tag(name string) {
	r.t.Helper()
	r.git("tag", name)
}

// Config sets a git config value on the repo.
func (r *Repo) Config(key, value string) {
	r.t.Helper()
	r.git("config", key, value)
}

// Amend rewrites the last commit, returning the new HEAD SHA. Useful for
// simulating upstream history rewrites.
func (r *Repo) Amend(message string) string {
	r.t.Helper()
	r.git("add", "--all")
	r.git("commit", "--amend", "-m", message)
	return r.HEAD()
}

// HEAD returns the full SHA of the current HEAD commit.
func (r *Repo) HEAD() string {
	r.t.Helper()
	return strings.TrimSpace(r.git("rev-parse", "HEAD"))
}

func (r *Repo) git(args ...string) string {
	r.t.Helper()
	cmd := exec.Command("git", append([]string{"-C", r.Dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("gittest: git %s: %v: %s", strings.Join(args, " "), err, out)
	}
	return string(out)
}
