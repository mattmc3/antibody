package project

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mattmc3/antibody/internal/config"
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
	if line[0] == '/' || strings.HasPrefix(line, "~/") {
		return NewLocal(line)
	}
	return NewGit(home, line), nil
}

// List all projects in the given folder
func List(home string) (result []string, err error) {
	if _, err := os.Stat(home); err != nil {
		return result, err
	}

	segments := config.Get().PathStyle().Segments()

	pattern := filepath.Join(home, strings.Repeat("*"+string(filepath.Separator), segments))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return result, err
	}

	for _, path := range matches {
		gitPath := filepath.Join(path, ".git")
		info, err := os.Stat(gitPath)
		if err != nil || !info.IsDir() {
			continue
		}
		rel, err := filepath.Rel(home, path)
		if err != nil {
			return result, err
		}
		result = append(result, rel)
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
