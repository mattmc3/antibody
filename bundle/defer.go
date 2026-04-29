package bundle

import (
	"fmt"
	"strings"

	"github.com/mattmc3/antibody/internal/config"
	"github.com/mattmc3/antibody/project"
)

type deferBundle struct {
	Project project.Project
}

func (bundle deferBundle) Get() (result string, err error) {
	cfg := config.Get()

	home, err := cfg.HomeDir()
	if err != nil {
		return "", err
	}

	// Build the zsh-defer tool source, wrapped in a load-once ensure block.
	deferProj, err := project.New(home, cfg.DeferBundle())
	if err != nil {
		return "", err
	}
	deferSrc, err := zshBundle{Project: deferProj}.Get()
	if err != nil {
		return "", err
	}
	var ensureLines []string
	ensureLines = append(ensureLines, "if ! (( $+functions[zsh-defer] )); then")
	for _, line := range strings.Split(deferSrc, "\n") {
		if line != "" {
			ensureLines = append(ensureLines, "  "+line)
		}
	}
	ensureLines = append(ensureLines, "fi")

	// Build the deferred plugin source, wrapping source lines.
	pluginSrc, err := zshBundle{Project: bundle.Project}.Get()
	if err != nil {
		return "", err
	}
	var pluginLines []string
	for _, line := range strings.Split(pluginSrc, "\n") {
		if strings.HasPrefix(line, "source ") {
			pluginLines = append(pluginLines, "zsh-defer "+line)
		} else {
			pluginLines = append(pluginLines, line)
		}
	}

	return fmt.Sprintf("%s\n%s", strings.Join(ensureLines, "\n"), strings.Join(pluginLines, "\n")), nil
}
