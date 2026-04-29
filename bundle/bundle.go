package bundle

import (
	"strings"

	"github.com/mattmc3/antibody/project"
)

// Bundle main interface.
type Bundle interface {
	Get() (result string, err error)
}

// New bundle with at the given home (when apply) and line.
//
// Accepted line formats:
//
//   - Local bundle (download and update do nothing):
//     /home/carlos/Code/my-local-bundle
//   - Github repo in the owner/repo format:
//     caarlos0/github-repo
//   - Git repo in any valid URL form:
//     https://github.com/caarlos0/other-github-repo.git
//   - Any type of repo, specifying the kind of resource:
//     caarlos0/add-to-path-style kind:path
//   - Any git repo, specifying a branch:
//     caarlos0/versioned-with-branch branch:v1.0 kind:zsh
//   - Any git repo, autoloading functions from a subpath:
//     caarlos0/my-plugin autoload:functions
func New(home, line string) (Bundle, error) {
	proj, err := project.New(home, line)
	if err != nil {
		return nil, err
	}

	var b Bundle
	switch kind(line) {
	case "autoload":
		b = autoloadBundle{Project: proj}
	case "path":
		b = pathBundle{Project: proj}
	case "fpath":
		b = fpathBundle{Project: proj}
	case "clone":
		b = cloneBundle{Project: proj}
	case "defer":
		b = deferBundle{Project: proj}
	default:
		b = zshBundle{Project: proj}
	}

	if subPath := autoloadAnnotation(line); subPath != "" && kind(line) != "autoload" {
		b = autoloadAnnotationBundle{inner: b, project: proj, subPath: subPath}
	}

	return b, nil
}

func kind(line string) string {
	for _, part := range strings.Split(line, " ") {
		if strings.HasPrefix(part, "kind:") {
			return strings.ReplaceAll(part, "kind:", "")
		}
	}
	return "zsh"
}

func autoloadAnnotation(line string) string {
	for _, part := range strings.Split(line, " ") {
		if strings.HasPrefix(part, "autoload:") {
			return strings.TrimPrefix(part, "autoload:")
		}
	}
	return ""
}
