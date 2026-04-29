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

	rule := annotation(line, "fpath-rule:")

	var b Bundle
	switch kind(line) {
	case "autoload":
		b = autoloadBundle{Project: proj, FpathRule: rule}
	case "path":
		b = pathBundle{Project: proj}
	case "fpath":
		b = fpathBundle{Project: proj, FpathRule: rule}
	case "clone":
		b = cloneBundle{Project: proj}
	case "defer":
		b = deferBundle{Project: proj}
	default:
		b = zshBundle{Project: proj}
	}

	if subPath := annotation(line, "autoload:"); subPath != "" && kind(line) != "autoload" {
		b = autoloadAnnotationBundle{inner: b, project: proj, subPath: subPath, fpathRule: rule}
	}

	pre := annotation(line, "pre:")
	post := annotation(line, "post:")
	cond := annotation(line, "conditional:")
	if pre != "" || post != "" || cond != "" {
		b = decoratedBundle{inner: b, pre: pre, post: post, conditional: cond}
	}

	return b, nil
}

// decoratedBundle wraps a bundle with optional pre/post commands and a
// conditional guard.
type decoratedBundle struct {
	inner       Bundle
	pre         string
	post        string
	conditional string
}

func (b decoratedBundle) Get() (result string, err error) {
	inner, err := b.inner.Get()
	if err != nil {
		return "", err
	}

	var lines []string
	if b.pre != "" {
		lines = append(lines, b.pre)
	}
	if inner != "" {
		lines = append(lines, inner)
	}
	if b.post != "" {
		lines = append(lines, b.post)
	}
	result = strings.Join(lines, "\n")

	if b.conditional != "" {
		var wrapped []string
		wrapped = append(wrapped, "if "+b.conditional+"; then")
		for line := range strings.SplitSeq(result, "\n") {
			if line != "" {
				wrapped = append(wrapped, "  "+line)
			}
		}
		wrapped = append(wrapped, "fi")
		return strings.Join(wrapped, "\n"), nil
	}

	return result, nil
}

func kind(line string) string {
	for part := range strings.SplitSeq(line, " ") {
		if v, ok := strings.CutPrefix(part, "kind:"); ok {
			return v
		}
	}
	return "zsh"
}

// annotation extracts the value for the first annotation with the given prefix.
func annotation(line, prefix string) string {
	for part := range strings.SplitSeq(line, " ") {
		if v, ok := strings.CutPrefix(part, prefix); ok {
			return v
		}
	}
	return ""
}
