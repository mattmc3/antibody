package project

import (
	"os"

	"golang.org/x/sync/errgroup"
)

// Project is basically any kind of project (git, local, svn, bzr, nfs...)
type Project interface {
	Download() error
	Update() error
	Path() string
}

// New decides what kind of project it is, based on the given line
func New(home, line string) (Project, error) {
	if IsLocal(line) {
		return NewLocal(line)
	}
	return NewGit(home, line), nil
}

// CloneRoot returns the top-level clone directory for the given bundle line.
func CloneRoot(home, line string) string {
	proj, _ := New(home, line)
	if git, ok := proj.(gitProject); ok {
		return git.folder
	}
	return proj.Path()
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
	sem := make(chan bool, parallelism)
	var g errgroup.Group
	for _, folder := range folders {
		folder := folder
		sem <- true
		g.Go(func() error {
			defer func() {
				<-sem
			}()
			return NewClonedGit(home, folder).Update()
		})
	}
	return g.Wait()
}
