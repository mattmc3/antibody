package project

import (
	"fmt"
	"log"
	"os"

	"github.com/getantidote/bundleparse"
	"golang.org/x/sync/errgroup"
)

// Project is basically any kind of project (git, local, svn, bzr, nfs...)
type Project interface {
	Download() error
	Update() error
	Path() string
}

// New parses a bundle line and returns the project it refers to.
func New(home, line string) (Project, error) {
	parsed, err := bundleparse.ParseBundleLine(line)
	if err != nil {
		return nil, err
	}
	if parsed.Name == "" {
		return nil, fmt.Errorf("not a bundle line: %q", line)
	}
	return NewFromParsed(home, parsed)
}

// NewFromParsed returns the project an already-parsed bundle refers to.
func NewFromParsed(home string, parsed bundleparse.Bundle) (Project, error) {
	if IsLocal(parsed.Name) {
		return NewLocal(parsed.Name)
	}
	return newGit(home, parsed.Name, parsed.Branch, parsed.Path, parsed.Pin)
}

// CloneRoot returns the top-level clone directory for the given bundle line.
func CloneRoot(home, line string) (string, error) {
	proj, err := New(home, line)
	if err != nil {
		return "", err
	}
	if git, ok := proj.(gitProject); ok {
		return git.folder, nil
	}
	return proj.Path(), nil
}

// List all projects in the given folder
func List(home string) (result []string, err error) {
	entries, err := os.ReadDir(home)
	if err != nil {
		return result, err
	}
	for _, entry := range entries {
		if entry.IsDir() && entry.Name()[0] != '.' {
			result = append(result, entry.Name())
		}
	}
	return result, nil
}

// Update all projects in the given folder
func Update(home string, parallelism int) error {
	folders, err := List(home)
	if err != nil {
		return err
	}
	var g errgroup.Group
	g.SetLimit(parallelism)
	for _, folder := range folders {
		if isPinnedFolder(folder) {
			log.Println("skipping pinned repo:", escapedPathToURL(folder))
			continue
		}
		g.Go(func() error {
			return NewClonedGit(home, folder).Update()
		})
	}
	return g.Wait()
}
