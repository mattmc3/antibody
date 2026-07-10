package shell

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mattmc3/antibody/internal/config"
	. "github.com/mattmc3/antibody/internal/expect"
)

func TestCompletionsZsh(t *testing.T) {
	out, err := Completions("zsh")
	Expect(t, NoError(err))
	Expect(t, Contains(out, "#compdef antibody"))
	Expect(t, Contains(out, "_antibody"))
}

func TestCompletionsZshListFlags(t *testing.T) {
	out, err := Completions("zsh")
	Expect(t, NoError(err))
	Expect(t, Contains(out, "--dirs"))
	Expect(t, Contains(out, "--url"))
}

func TestCompletionsZshHelp(t *testing.T) {
	out, err := Completions("zsh")
	Expect(t, NoError(err))
	// the CLI has a help subcommand; every subcommand takes -h
	Expect(t, Contains(out, "help:"))
	Expect(t, Contains(out, "(help)"))
	Expect(t, Contains(out, "Show help for a command"))
}

func TestCompletionsZshDualMode(t *testing.T) {
	out, err := Completions("zsh")
	Expect(t, NoError(err))
	// works as fpath autoload file and as sourced script
	Expect(t, Contains(out, "zsh_eval_context"))
	Expect(t, Contains(out, "compdef _antibody antibody"))
}

func TestCompletionsUnsupportedShell(t *testing.T) {
	_, err := Completions("fish")
	Expect(t, AnError(err))
}

func TestCompletionsFpathZdotdir(t *testing.T) {
	zdotdir := t.TempDir()
	t.Setenv("ZDOTDIR", zdotdir)
	t.Setenv("XDG_CONFIG_HOME", "")

	file, err := CompletionsFpath("zsh")
	Expect(t, NoError(err))
	Expect(t, Equals(filepath.Join(zdotdir, "completions", "_antibody"), file))

	content, err := os.ReadFile(file)
	Expect(t, NoError(err))
	expected, err := Completions("zsh")
	Expect(t, NoError(err))
	Expect(t, Equals(expected, string(content)))
}

func TestCompletionsFpathXdgFallback(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("ZDOTDIR", "")
	t.Setenv("XDG_CONFIG_HOME", xdg)

	file, err := CompletionsFpath("zsh")
	Expect(t, NoError(err))
	Expect(t, Equals(filepath.Join(xdg, "completions", "_antibody"), file))
}

func TestCompletionsFpathConfigDir(t *testing.T) {
	confDir := t.TempDir()
	target := filepath.Join(t.TempDir(), "my-comps")
	Expect(t, NoError(os.MkdirAll(filepath.Join(confDir, "antibody"), 0o755)))
	toml := "[completions]\ndir = \"" + target + "\"\n"
	Expect(t, NoError(os.WriteFile(filepath.Join(confDir, "antibody", "antibody.toml"), []byte(toml), 0o644)))

	t.Setenv("XDG_CONFIG_HOME", confDir)
	t.Setenv("ZDOTDIR", t.TempDir()) // config must win over ZDOTDIR
	_, err := config.Load()
	Expect(t, NoError(err))
	emptyConf := t.TempDir()
	t.Cleanup(func() {
		// reset singleton to empty config for later tests
		_ = os.Setenv("XDG_CONFIG_HOME", emptyConf)
		_, _ = config.Load()
	})

	file, err := CompletionsFpath("zsh")
	Expect(t, NoError(err))
	Expect(t, Equals(filepath.Join(target, "_antibody"), file))
	Expect(t, FileExists(file))
}

func TestCompletionsFpathNoRewrite(t *testing.T) {
	t.Setenv("ZDOTDIR", t.TempDir())
	file, err := CompletionsFpath("zsh")
	Expect(t, NoError(err))

	before, err := os.Stat(file)
	Expect(t, NoError(err))

	_, err = CompletionsFpath("zsh")
	Expect(t, NoError(err))
	after, err := os.Stat(file)
	Expect(t, NoError(err))
	Expect(t, Equals(before.ModTime(), after.ModTime()))
}

func TestCompletionsFpathUnsupportedShell(t *testing.T) {
	t.Setenv("ZDOTDIR", t.TempDir())
	_, err := CompletionsFpath("fish")
	Expect(t, AnError(err))
}
