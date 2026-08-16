package bundle

import (
	"fmt"
	"strings"

	"github.com/mattmc3/antibody/bundleparse"
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
	parsed, err := bundleparse.ParseBundleLine(line)
	if err != nil {
		return nil, err
	}
	if parsed.Name == "" {
		return nil, fmt.Errorf("not a bundle line: %q", line)
	}
	return NewFromParsed(home, parsed)
}

func NewFromParsed(home string, parsed bundleparse.Bundle) (Bundle, error) {
	if parsed.Name == "" {
		return nil, fmt.Errorf("empty bundle name")
	}
	if err := bundleparse.ValidatePin(parsed.Name, parsed.Pin); err != nil {
		return nil, err
	}

	proj, err := project.NewFromParsed(home, parsed)
	if err != nil {
		return nil, err
	}

	if parsed.Kind == "" {
		parsed.Kind = bundleparse.KindZsh
	}

	return bundleFromParsed(parsed, proj)
}

func bundleFromParsed(parsed bundleparse.Bundle, proj project.Project) (Bundle, error) {
	var b Bundle
	switch parsed.Kind {
	case bundleparse.KindAutoload:
		b = autoloadBundle{Project: proj, FpathRule: parsed.FpathRule}
	case bundleparse.KindPath:
		b = pathBundle{Project: proj}
	case bundleparse.KindFpath:
		b = fpathBundle{Project: proj, FpathRule: parsed.FpathRule}
	case bundleparse.KindClone:
		b = cloneBundle{Project: proj}
	case bundleparse.KindDefer:
		b = deferBundle{Project: proj}
	default:
		b = zshBundle{Project: proj}
	}

	if parsed.Autoload != "" && parsed.Kind != bundleparse.KindAutoload && parsed.Kind != bundleparse.KindClone {
		b = autoloadAnnotationBundle{inner: b, project: proj, subPath: parsed.Autoload, fpathRule: parsed.FpathRule}
	}

	if parsed.Pre != "" || parsed.Post != "" || parsed.Conditional != "" {
		b = decoratedBundle{
			inner:       b,
			pre:         parsed.Pre,
			post:        parsed.Post,
			conditional: parsed.Conditional,
			deferPost:   parsed.Kind == bundleparse.KindDefer,
		}
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
	deferPost   bool
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
		post := b.post
		if b.deferPost {
			post = "zsh-defer " + post
		}
		lines = append(lines, post)
	}
	result = strings.Join(lines, "\n")
	if result == "" {
		return "", nil
	}

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
