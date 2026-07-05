package shell

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mattmc3/antibody/internal/config"
)

// zshCompletionFunction defines _antibody, the compsys completion function
// (uses _arguments/_describe, so it must be registered with compdef, not
// the legacy compctl). Shared by Completions (fpath file form) and Init
// (inline eval form).
const zshCompletionFunction = `function _antibody_installed_bundles {
	local -a bundles=("${(@f)$(antibody list --url 2>/dev/null | awk -F'/' '{print $(NF-1)"/"$NF}')}")
	_describe 'installed bundles' bundles
}

function _antibody {
	local curcontext="$curcontext" state line ret=1
	typeset -A opt_args

	local -a subcommands=(
		'bundle:downloads a bundle and prints its source line'
		'update:updates all previously bundled bundles'
		'home:prints where antibody is cloning the bundles'
		'purge:purges a bundle from your computer'
		'list:lists all currently installed bundles'
		'path:prints the path of a currently cloned bundle'
		'init:initializes the shell so Antibody can work as expected'
		'completions:generates shell completion scripts'
	)

	_arguments -C \
		'(- *)'{-v,--version}'[Show application version]' \
		'(- *)'{-h,--help}'[Show help]' \
		'(-p --parallelism)'{-p,--parallelism}'[max amount of tasks to launch in parallel]:parallelism' \
		'1: :->command' \
		'*:: :->args' && return 0

	case "$state" in
	(command)
		_describe 'command' subcommands && ret=0
		;;
	(args)
		case $words[1] in
		(purge|path)
			_antibody_installed_bundles && ret=0
			;;
		(completions)
			local -a shells=('zsh:generate zsh completions')
			_describe 'shell' shells && ret=0
			;;
		(*)
			ret=0
			;;
		esac
		;;
	esac

	return ret
}
`

// dual-mode tail: as an fpath autoload file the function runs directly;
// sourced or eval'd, it registers with compdef instead.
const zshCompletionTail = `if [[ $zsh_eval_context[-1] == loadautofunc ]]; then
	_antibody "$@"
else
	compdef _antibody antibody
fi
`

// Completions returns a completion script for the given shell. The zsh script
// works both as an fpath file (antibody completions zsh > .../_antibody) and
// sourced directly (source <(antibody completions zsh)).
func Completions(shell string) (string, error) {
	switch shell {
	case "zsh":
		return "#compdef antibody\n\n" + zshCompletionFunction + "\n" + zshCompletionTail, nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}
}

// CompletionsFpath writes the completion script to the config [completions]
// dir, or ${ZDOTDIR:-${XDG_CONFIG_HOME:-$HOME/.config}}/completions when not
// configured, and returns the full path of the written file. The file is
// only rewritten when its content changes.
func CompletionsFpath(shell string) (string, error) {
	script, err := Completions(shell)
	if err != nil {
		return "", err
	}
	dir, err := completionsDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	file := filepath.Join(dir, "_antibody")
	if existing, err := os.ReadFile(file); err == nil && bytes.Equal(existing, []byte(script)) {
		return file, nil
	}
	if err := os.WriteFile(file, []byte(script), 0o644); err != nil {
		return "", err
	}
	return file, nil
}

// completionsDir resolves where the completion file lives.
// Priority: config [completions] dir > $ZDOTDIR > $XDG_CONFIG_HOME > ~/.config.
func completionsDir() (string, error) {
	if dir := config.Get().Compdir(); dir != "" {
		return dir, nil
	}
	base := os.Getenv("ZDOTDIR")
	if base == "" {
		base = os.Getenv("XDG_CONFIG_HOME")
	}
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "completions"), nil
}
