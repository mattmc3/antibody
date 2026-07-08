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
	// $VAR names are emitted verbatim for the shell to evaluate at load
	// time. The env only decides file-vs-dir; nothing is expanded or
	// globbed in the output.
	if strings.HasPrefix(dir, "$") {
		if info, err := os.Stat(os.ExpandEnv(dir)); err == nil && info.Mode().IsRegular() {
			return "source " + dir, nil
		}
		return fpathLine(dir, "") + "\nsource " + dir + "/" + initFileBase(dir) + ".plugin.zsh", nil
	}
	info, err := os.Stat(dir)
	if err != nil {
		return "", err
	}
	// it is a file, not a folder, so just return it
	if info.Mode().IsRegular() {
		// XXX: should we add the parent folder to fpath too?
		return "source " + display(bundle.Project, dir), nil
	}
	files, err := initFiles(dir)
	if err != nil {
		return "", err
	}
	lines := []string{fpathLine(display(bundle.Project, dir), "")}
	for _, file := range files {
		lines = append(lines, "source "+display(bundle.Project, file))
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

// display returns the output form of a filesystem path: mapped back to the
// prefix the bundle line used (eg $VAR) when the project supports it, then
// $HOME-substituted.
func display(proj project.Project, p string) string {
	type displayer interface{ Display(string) string }
	if d, ok := proj.(displayer); ok {
		p = d.Display(p)
	}
	return displayPath(p)
}

// displayPath replaces the user's home dir prefix with a literal $HOME so
// generated static files are portable.
func displayPath(p string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return p
	}
	if p == home {
		return "$HOME"
	}
	if strings.HasPrefix(p, home+string(filepath.Separator)) {
		return "$HOME" + p[len(home):]
	}
	return p
}
