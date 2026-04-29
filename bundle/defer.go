package bundle

import (
	"strings"

	"github.com/mattmc3/antibody/project"
)

type deferBundle struct {
	Project project.Project
}

func (bundle deferBundle) Get() (result string, err error) {
	result, err = zshBundle{Project: bundle.Project}.Get()
	if err != nil || result == "" {
		return result, err
	}
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "source ") {
			lines[i] = "zsh-defer " + line
		}
	}
	return strings.Join(lines, "\n"), nil
}
