package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mattmc3/antibody/internal/gittest"
	"github.com/stretchr/testify/require"
)

// writeConfig writes an antibody.toml under the XDG config dir runCLI uses.
func writeConfig(t *testing.T, home, content string) {
	t.Helper()
	dir := filepath.Join(home, ".config", "antibody")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	// nolint: gosec
	require.NoError(t, os.WriteFile(filepath.Join(dir, "antibody.toml"), []byte(content), 0o644))
}

func TestCLIConfigDeferBundle(t *testing.T) {
	deferRepo := gittest.New(t)
	deferRepo.WriteFile("my-defer.plugin.zsh", "echo defer\n")
	deferRepo.Commit("add plugin file")
	plugin := pluginFixture(t)
	home := t.TempDir()
	writeConfig(t, home, "[defer]\nbundle = \""+deferRepo.URL()+"\"\n")

	res := runCLI(t, home, "", "bundle", plugin.URL()+" kind:defer")
	require.Equal(t, 0, res.exitCode, res.stderr)
	require.Contains(t, res.stdout, "if ! (( $+functions[zsh-defer] )); then")
	require.Contains(t, res.stdout, "my-defer.plugin.zsh")
	require.Contains(t, res.stdout, "zsh-defer source ")
}

func TestCLIConfigFpathRule(t *testing.T) {
	plugin := pluginFixture(t)
	home := t.TempDir()
	writeConfig(t, home, "[fpath]\nrule = \"prepend\"\n")

	res := runCLI(t, home, "", "bundle", plugin.URL())
	require.Equal(t, 0, res.exitCode, res.stderr)
	require.Contains(t, res.stdout, "fpath=( ")
	require.Contains(t, res.stdout, " $fpath )")
	require.NotContains(t, res.stdout, "fpath+=( ")
}

func TestCLIConfigGitDomain(t *testing.T) {
	home := t.TempDir()
	writeConfig(t, home, "[git]\ndomain = \"localhost:1\"\n")

	res := runCLI(t, home, "", "bundle", "foo/bar")
	require.NotEqual(t, 0, res.exitCode)
	require.Contains(t, res.stderr, "https://localhost:1/foo/bar")
}

func TestCLIConfigMalformed(t *testing.T) {
	home := t.TempDir()
	writeConfig(t, home, "not [valid toml\n")

	res := runCLI(t, home, "", "home")
	require.NotEqual(t, 0, res.exitCode)
	require.Contains(t, res.stderr, "failed to parse config")
}
