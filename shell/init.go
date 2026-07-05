package shell

import (
	"bytes"
	"os"
	"text/template"
)

const tmpl = `#!/usr/bin/env zsh
antibody() {
	case "$1" in
	bundle)
		source <( {{ . }} $@ ) || {{ . }} $@
		;;
	*)
		{{ . }} $@
		;;
	esac
}

if ! (( $+functions[_antibody] )); then
	if (( $+functions[compdef] )); then
		source <( {{ . }} completions zsh )
	else
		fpath=("${$( {{ . }} completions --fpath zsh ):h}" $fpath)
	fi
fi
`

// Init returns the shell that should be loaded to antibody to work correctly.
func Init() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	var template = template.Must(template.New("init").Parse(tmpl))
	var out bytes.Buffer
	err = template.Execute(&out, executable)
	return out.String(), err
}
