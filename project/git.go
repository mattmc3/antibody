package project

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/mattmc3/antibody/internal/config"
)

// nolint: gochecknoglobals
var gitCmdEnv = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS=0", "SSH_ASKPASS=0")

type gitProject struct {
	URL     string
	Version string
	Pin     string
	folder  string
	inner   string
}

// NewClonedGit is a git project that was already cloned, so, only Update
// will work here.
func NewClonedGit(home, folderName string) Project {
	folderPath := filepath.Join(home, folderName)
	version, err := branch(folderPath)
	if err != nil {
		version = ""
	}
	url := escapedPathToURL(folderName)
	return gitProject{
		folder:  folderPath,
		Version: version,
		URL:     url,
	}
}

const (
	branchMarker = "branch:"
	pathMarker   = "path:"
	pinMarker    = "pin:"
)

func newGit(cwd, repo, version, inner, pin string) Project {
	cfg := config.Get()
	var repoURL string
	if cfg.GitProtocol() == "ssh" {
		repoURL = "git@" + cfg.GitDomain() + ":" + repo
	} else {
		repoURL = "https://" + cfg.GitDomain() + "/" + repo
	}
	switch {
	case strings.HasPrefix(repo, "http://"):
		fallthrough
	case strings.HasPrefix(repo, "https://"):
		fallthrough
	case strings.HasPrefix(repo, "git://"):
		fallthrough
	case strings.HasPrefix(repo, "ssh://"):
		fallthrough
	case strings.HasPrefix(repo, "git@"):
		fallthrough
	case strings.HasPrefix(repo, "file://"):
		repoURL = repo
	}

	// Handle git@ style URLs which url.Parse can't handle
	parseable := repoURL
	if strings.HasPrefix(parseable, "git@") {
		// Convert git@host:path to ssh://git@host/path
		parseable = strings.Replace(parseable, ":", "/", 1)
		parseable = "ssh://" + parseable
	}

	u, err := url.Parse(parseable)
	if err != nil || u == nil {
		log.Printf("failed to parse URL %s: %v", parseable, err)
		u = &url.URL{Host: cfg.GitDomain(), Path: "/unknown"}
	}
	folder := filepath.Join(cwd, escapedPathFromURL(u))
	if pin != "" {
		folder = folder + "-SLASH-tree-SLASH-" + pin
	}
	return gitProject{
		Version: version,
		URL:     repoURL,
		Pin:     pin,
		folder:  folder,
		inner:   inner,
	}
}

// NewGit A git project can be any repository in any given branch. It will
// be downloaded to the provided cwd
func NewGit(cwd, line string) Project {
	version := ""
	inner := ""
	pin := ""
	parts := strings.Split(line, " ")
	for _, part := range parts {
		if strings.HasPrefix(part, branchMarker) {
			version = strings.ReplaceAll(part, branchMarker, "")
		}
		if strings.HasPrefix(part, pathMarker) {
			inner = strings.ReplaceAll(part, pathMarker, "")
		}
		if strings.HasPrefix(part, pinMarker) {
			pin = strings.ReplaceAll(part, pinMarker, "")
		}
	}
	return newGit(cwd, parts[0], version, inner, pin)
}

// NewGitWithAnnotations creates a git project from explicit repo, branch, path, and pin.
func NewGitWithAnnotations(cwd, repo, branch, path, pin string) Project {
	return newGit(cwd, repo, branch, path, pin)
}

// nolint: gochecknoglobals
var locks sync.Map

func (g gitProject) Download() error {
	l, _ := locks.LoadOrStore(g.folder, &sync.Mutex{})
	lock := l.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()
	created := false
	if _, err := os.Stat(g.folder); os.IsNotExist(err) {
		created = true
		// #nosec
		args := []string{
			"clone",
			"--recursive",
			"--depth", "1",
		}
		if g.Version != "" {
			args = append(args, "-b", g.Version)
		}
		args = append(args, g.URL, g.folder)
		var cmd = exec.Command("git", args...)
		cmd.Env = gitCmdEnv

		if bts, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git clone failed: %w: %s", err, strings.TrimSpace(string(bts)))
		}
	}
	if g.Pin != "" {
		err := g.ensurePinned()
		if err != nil && created {
			_ = os.RemoveAll(g.folder)
		}
		return err
	}
	return nil
}

func (g gitProject) Update() error {
	log.Println("updating:", g.URL)
	oldRev, err := commit(g.folder)
	if err != nil {
		return err
	}
	// #nosec
	args := []string{
		"pull",
		"--recurse-submodules",
		"origin",
	}
	if g.Version != "" {
		args = append(args, g.Version)
	}
	cmd := exec.Command("git", args...)
	cmd.Env = gitCmdEnv

	cmd.Dir = g.folder
	if bts, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git update failed: %w: %s", err, strings.TrimSpace(string(bts)))
	}
	rev, err := commit(g.folder)
	if err != nil {
		return err
	}
	if rev != oldRev {
		log.Println("updated:", g.URL, oldRev, "->", rev)
	}
	return nil
}

func commit(folder string) (string, error) {
	// #nosec
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = folder
	rev, err := cmd.Output()
	return strings.ReplaceAll(string(rev), "\n", ""), err
}

func gitCheckoutDetach(folder, sha string) error {
	cmd := exec.Command("git", "checkout", "--quiet", "--detach", sha)
	cmd.Dir = folder
	cmd.Env = gitCmdEnv
	if bts, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout failed: %w: %s", err, strings.TrimSpace(string(bts)))
	}
	return nil
}

func (g gitProject) ensurePinned() error {
	if headSHA(g.folder) == g.Pin {
		return nil
	}
	if err := gitCheckoutDetach(g.folder, g.Pin); err != nil {
		// Try fetching the commit if it's not present locally.
		cmd := exec.Command("git", "fetch", "--depth", "1", "origin", g.Pin)
		cmd.Dir = g.folder
		cmd.Env = gitCmdEnv
		if bts, ferr := cmd.CombinedOutput(); ferr != nil {
			return fmt.Errorf("git fetch failed: %w: %s", ferr, strings.TrimSpace(string(bts)))
		}
		return gitCheckoutDetach(g.folder, g.Pin)
	}
	return nil
}

// headSHA reads .git/HEAD directly so a correctly pinned clone can be
// verified without spawning git. A detached HEAD holds the bare SHA;
// anything else ("ref: ..." or a read error) fails the match and takes
// the checkout path.
func headSHA(folder string) string {
	head, err := os.ReadFile(filepath.Join(folder, ".git", "HEAD"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(head))
}

func branch(folder string) (string, error) {
	// #nosec
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = folder
	branch, err := cmd.Output()
	return strings.ReplaceAll(string(branch), "\n", ""), err
}

func escapedPathFromURL(u *url.URL) string {
	result := u.String()
	result = strings.ReplaceAll(result, ":", "-COLON-")
	result = strings.ReplaceAll(result, "/", "-SLASH-")
	result = strings.ReplaceAll(result, "@", "-AT-")
	return result
}

// nolint: gochecknoglobals
var pinnedFolderPattern = regexp.MustCompile(`-SLASH-tree-SLASH-[0-9a-f]{40}$`)

// isPinnedFolder reports whether a clone folder name is a pinned clone.
// The pin SHA is part of the folder name, so no git state is consulted.
func isPinnedFolder(folder string) bool {
	return pinnedFolderPattern.MatchString(folder)
}

func escapedPathToURL(path string) string {
	result := path
	result = strings.ReplaceAll(result, "-AT-", "@")
	result = strings.ReplaceAll(result, "-SLASH-", "/")
	result = strings.ReplaceAll(result, "-COLON-", ":")
	return result
}

func EscapedPathToURL(path string) string {
	return escapedPathToURL(path)
}

func (g gitProject) Path() string {
	return filepath.Join(g.folder, g.inner)
}
