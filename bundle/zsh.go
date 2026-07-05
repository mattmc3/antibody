package bundle

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mattmc3/antibody/project"
)

type zshBundle struct {
	Project project.Project
}

func (bundle zshBundle) Get() (result string, err error) {
	if err = bundle.Project.Download(); err != nil {
		return result, err
	}
	dir := bundle.Project.Path()
	info, err := os.Stat(dir)
	if err != nil {
		return "", err
	}
	// it is a file, not a folder, so just return it
	if info.Mode().IsRegular() {
		// XXX: should we add the parent folder to fpath too?
		return "source " + dir, nil
	}
	files, err := initFiles(dir)
	if err != nil {
		return "", err
	}
	lines := []string{fpathLine(dir, "")}
	for _, file := range files {
		lines = append(lines, "source "+file)
	}
	return strings.Join(lines, "\n"), nil
}

// initFiles picks the files to source for a plugin directory:
// <dir>/<name>.plugin.zsh first, then the glob fallbacks, and if
// nothing matches, assume the default plugin file.
func initFiles(dir string) ([]string, error) {
	candidate := filepath.Join(dir, initFileBase(dir)+".plugin.zsh")
	if _, err := os.Stat(candidate); err == nil {
		return []string{candidate}, nil
	}
	for _, glob := range []string{"*.plugin.zsh", "*.zsh", "*.sh", "*.zsh-theme"} {
		files, err := filepath.Glob(filepath.Join(dir, glob))
		if err != nil {
			return nil, err
		}
		if len(files) > 0 {
			return files, nil
		}
	}
	return []string{candidate}, nil
}

// initFileBase returns the plugin name a directory's default init file is
// named after. Clone folders are URL-escaped, so unescape before taking the
// last path segment.
func initFileBase(dir string) string {
	base := project.EscapedPathToURL(filepath.Base(dir))
	if i := strings.LastIndex(base, "/"); i >= 0 {
		base = base[i+1:]
	}
	return base
}
