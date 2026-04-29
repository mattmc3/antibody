package bundle

import (
	"fmt"

	"github.com/mattmc3/antibody/internal/config"
	"github.com/mattmc3/antibody/project"
)

type fpathBundle struct {
	Project project.Project
}

func (bundle fpathBundle) Get() (result string, err error) {
	if err = bundle.Project.Download(); err != nil {
		return result, err
	}
	path := bundle.Project.Path()
	if config.Get().FpathRule() == "prepend" {
		return fmt.Sprintf("fpath=( %s $fpath )", path), nil
	}
	return fmt.Sprintf("fpath+=( %s )", path), nil
}
