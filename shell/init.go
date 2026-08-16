package shell

import (
	"bytes"
	"os"
	"text/template"
)

const tmpl = `#!/usr/bin/env zsh
typeset -g _ANTIBODY_BIN="{{ . }}"

antibody() {
	case "$1" in
	bundle)
		source <( ANTIBODY_DYNAMIC=true "$_ANTIBODY_BIN" "$@" ) || "$_ANTIBODY_BIN" "$@"
		;;
	*)
		"$_ANTIBODY_BIN" "$@"
		;;
	esac
}

if ! (( $+functions[_antibody] )); then
	if (( $+functions[compdef] )); then
		source <( "$_ANTIBODY_BIN" completions zsh )
	else
		fpath=("${$( "$_ANTIBODY_BIN" completions --fpath zsh ):h}" $fpath)
	fi
fi
`

// Init returns the shell that should be loaded to antibody to work correctly.
func Init() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	initTmpl := template.Must(template.New("init").Parse(tmpl))
	var out bytes.Buffer
	err = initTmpl.Execute(&out, executable)
	return out.String(), err
}
