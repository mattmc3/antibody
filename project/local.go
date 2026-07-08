package project

import (
	"os"
	"path/filepath"
	"strings"
)

// NewLocal Returns a local project, which can be any folder you want to
func NewLocal(line string) (Project, error) {
	name := strings.Split(line, " ")[0]
	folder, err := expandFolder(name)
	if err != nil {
		return localProject{}, err
	}
	display := name
	if strings.HasPrefix(name, "~/") {
		display = folder
	}
	return localProject{folder: folder, display: display}, nil
}

// IsLocal reports whether a bundle name is a filesystem path rather than
// a repo: absolute, home-relative, $VAR-prefixed, or dot-relative.
func IsLocal(name string) bool {
	return strings.HasPrefix(name, "/") ||
		strings.HasPrefix(name, "~") ||
		strings.HasPrefix(name, "$") ||
		strings.HasPrefix(name, ".")
}

func expandFolder(folder string) (string, error) {
	if strings.HasPrefix(folder, "~/") {
		dir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return strings.Replace(folder, "~", dir, 1), nil
	}
	return folder, nil
}

type localProject struct {
	folder  string // expanded path used for filesystem access
	display string // form written to generated output
}

func (l localProject) Download() error {
	// $VAR names are emitted verbatim and evaluated by the shell at load
	// time, so there is nothing to check here.
	if strings.HasPrefix(l.folder, "$") {
		return nil
	}
	_, err := os.Stat(l.folder)
	return err
}

func (l localProject) Update() error {
	return l.Download()
}

func (l localProject) Path() string {
	return l.folder
}

// Display maps a filesystem path under the project folder back to the
// form the bundle line used, eg a $VAR-prefixed path. Globbed file paths
// come back cleaned, so the cleaned folder is matched too.
func (l localProject) Display(p string) string {
	for _, prefix := range []string{l.folder, filepath.Clean(l.folder)} {
		if p == prefix {
			return l.display
		}
		if strings.HasPrefix(p, prefix) && p[len(prefix)] == '/' {
			return l.display + p[len(prefix):]
		}
	}
	return p
}
