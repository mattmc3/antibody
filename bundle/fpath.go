package bundle

import (
	"fmt"

	"github.com/mattmc3/antibody/internal/config"
	"github.com/mattmc3/antibody/project"
)

type fpathBundle struct {
	Project   project.Project
	FpathRule string
}

func (bundle fpathBundle) Get() (result string, err error) {
	if err = bundle.Project.Download(); err != nil {
		return result, err
	}
	return fpathLine(display(bundle.Project, bundle.Project.Path()), bundle.FpathRule), nil
}

// fpathLine returns the appropriate fpath assignment for a directory,
// already in display form. If rule is empty the global config value is used.
func fpathLine(dir, rule string) string {
	if resolvedFpathRule(rule) == "prepend" {
		return fmt.Sprintf("fpath=( %s $fpath )", quote(dir))
	}
	return fmt.Sprintf("fpath+=( %s )", quote(dir))
}

// resolvedFpathRule returns rule if non-empty, otherwise the config default.
func resolvedFpathRule(rule string) string {
	if rule == "" {
		return config.Get().FpathRule()
	}
	return rule
}
