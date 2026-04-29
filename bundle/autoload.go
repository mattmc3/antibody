package bundle

import (
	"path/filepath"
	"strings"

	"github.com/mattmc3/antibody/project"
)

type autoloadBundle struct {
	Project   project.Project
	FpathRule string
}

func (b autoloadBundle) Get() (result string, err error) {
	if err = b.Project.Download(); err != nil {
		return result, err
	}
	return autoloadLines(b.Project.Path(), b.FpathRule), nil
}

// autoloadAnnotationBundle wraps an inner bundle, prepending autoload lines
// for the given subpath before the inner bundle's output.
type autoloadAnnotationBundle struct {
	inner     Bundle
	project   project.Project
	subPath   string
	fpathRule string
}

func (b autoloadAnnotationBundle) Get() (result string, err error) {
	inner, err := b.inner.Get()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(b.project.Path(), b.subPath)
	prefix := autoloadLines(dir, b.fpathRule)
	if inner == "" {
		return prefix, nil
	}
	return prefix + "\n" + inner, nil
}

// autoloadLines builds fpath + autoload -Uz lines for a directory.
func autoloadLines(dir, fpathRule string) string {
	var lines []string
	lines = append(lines, fpathLine(dir, fpathRule))
	if resolvedFpathRule(fpathRule) == "prepend" {
		lines = append(lines, "builtin autoload -Uz $fpath[1]/*(N.:t)")
	} else {
		lines = append(lines, "builtin autoload -Uz $fpath[-1]/*(N.:t)")
	}
	return strings.Join(lines, "\n")
}
