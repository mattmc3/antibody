package shell

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mattmc3/antibody/internal/config"
	"github.com/stretchr/testify/require"
)

func TestCompletionsZsh(t *testing.T) {
	out, err := Completions("zsh")
	require.NoError(t, err)
	require.Contains(t, out, "#compdef antibody")
	require.Contains(t, out, "_antibody")
}

func TestCompletionsZshListFlags(t *testing.T) {
	out, err := Completions("zsh")
	require.NoError(t, err)
	require.Contains(t, out, "--dirs")
	require.Contains(t, out, "--url")
}

func TestCompletionsZshHelp(t *testing.T) {
	out, err := Completions("zsh")
	require.NoError(t, err)
	// kingpin auto-adds a help subcommand; every subcommand takes -h
	require.Contains(t, out, "help:")
	require.Contains(t, out, "(help)")
	require.Contains(t, out, "Show help for a command")
}

func TestCompletionsZshDualMode(t *testing.T) {
	out, err := Completions("zsh")
	require.NoError(t, err)
	// works as fpath autoload file and as sourced script
	require.Contains(t, out, "zsh_eval_context")
	require.Contains(t, out, "compdef _antibody antibody")
}

func TestCompletionsUnsupportedShell(t *testing.T) {
	_, err := Completions("fish")
	require.Error(t, err)
}

func TestCompletionsFpathZdotdir(t *testing.T) {
	zdotdir := t.TempDir()
	t.Setenv("ZDOTDIR", zdotdir)
	t.Setenv("XDG_CONFIG_HOME", "")

	file, err := CompletionsFpath("zsh")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(zdotdir, "completions", "_antibody"), file)

	content, err := os.ReadFile(file)
	require.NoError(t, err)
	expected, err := Completions("zsh")
	require.NoError(t, err)
	require.Equal(t, expected, string(content))
}

func TestCompletionsFpathXdgFallback(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("ZDOTDIR", "")
	t.Setenv("XDG_CONFIG_HOME", xdg)

	file, err := CompletionsFpath("zsh")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(xdg, "completions", "_antibody"), file)
}

func TestCompletionsFpathConfigDir(t *testing.T) {
	confDir := t.TempDir()
	target := filepath.Join(t.TempDir(), "my-comps")
	require.NoError(t, os.MkdirAll(filepath.Join(confDir, "antibody"), 0o755))
	toml := "[completions]\ndir = \"" + target + "\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(confDir, "antibody", "antibody.toml"), []byte(toml), 0o644))

	t.Setenv("XDG_CONFIG_HOME", confDir)
	t.Setenv("ZDOTDIR", t.TempDir()) // config must win over ZDOTDIR
	_, err := config.Load()
	require.NoError(t, err)
	emptyConf := t.TempDir()
	t.Cleanup(func() {
		// reset singleton to empty config for later tests
		_ = os.Setenv("XDG_CONFIG_HOME", emptyConf)
		_, _ = config.Load()
	})

	file, err := CompletionsFpath("zsh")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(target, "_antibody"), file)
	require.FileExists(t, file)
}

func TestCompletionsFpathNoRewrite(t *testing.T) {
	t.Setenv("ZDOTDIR", t.TempDir())
	file, err := CompletionsFpath("zsh")
	require.NoError(t, err)

	before, err := os.Stat(file)
	require.NoError(t, err)

	_, err = CompletionsFpath("zsh")
	require.NoError(t, err)
	after, err := os.Stat(file)
	require.NoError(t, err)
	require.Equal(t, before.ModTime(), after.ModTime())
}

func TestCompletionsFpathUnsupportedShell(t *testing.T) {
	t.Setenv("ZDOTDIR", t.TempDir())
	_, err := CompletionsFpath("fish")
	require.Error(t, err)
}
