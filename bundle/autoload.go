package bundle

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mattmc3/antibody/internal/config"
	"github.com/mattmc3/antibody/project"
)

type autoloadBundle struct {
	Project project.Project
}

func (b autoloadBundle) Get() (result string, err error) {
	if err = b.Project.Download(); err != nil {
		return result, err
	}
	return autoloadLines(b.Project.Path()), nil
}

// autoloadAnnotationBundle wraps an inner bundle, prepending autoload lines
// for the given subpath before the inner bundle's output.
type autoloadAnnotationBundle struct {
	inner   Bundle
	project project.Project
	subPath string
}

func (b autoloadAnnotationBundle) Get() (result string, err error) {
	inner, err := b.inner.Get()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(b.project.Path(), b.subPath)
	prefix := autoloadLines(dir)
	if inner == "" {
		return prefix, nil
	}
	return prefix + "\n" + inner, nil
}

// autoloadLines builds the fpath + autoload lines for a directory,
// respecting the configured fpath rule.
func autoloadLines(dir string) string {
	var lines []string
	if config.Get().FpathRule() == "prepend" {
		lines = append(lines, fmt.Sprintf("fpath=( %s $fpath )", dir))
		lines = append(lines, "builtin autoload -Uz $fpath[1]/*(N.:t)")
	} else {
		lines = append(lines, fmt.Sprintf("fpath+=( %s )", dir))
		lines = append(lines, "builtin autoload -Uz $fpath[-1]/*(N.:t)")
	}
	return strings.Join(lines, "\n")
}
