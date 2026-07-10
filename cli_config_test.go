package main

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/mattmc3/antibody/internal/expect"
	"github.com/mattmc3/antibody/internal/gittest"
)

// writeConfig writes an antibody.toml under the XDG config dir runCLI uses.
func writeConfig(t *testing.T, home, content string) {
	t.Helper()
	dir := filepath.Join(home, ".config", "antibody")
	Expect(t, NoError(os.MkdirAll(dir, 0o755)))
	// nolint: gosec
	Expect(t, NoError(os.WriteFile(filepath.Join(dir, "antibody.toml"), []byte(content), 0o644)))
}

func TestCLIConfigDeferBundle(t *testing.T) {
	deferRepo := gittest.New(t)
	deferRepo.WriteFile("my-defer.plugin.zsh", "echo defer\n")
	deferRepo.Commit("add plugin file")
	plugin := pluginFixture(t)
	home := t.TempDir()
	writeConfig(t, home, "[defer]\nbundle = \""+deferRepo.URL()+"\"\n")

	res := runCLI(t, home, "", "bundle", plugin.URL()+" kind:defer")
	Expect(t, Equals(0, res.exitCode), res.stderr)
	Expect(t, Contains(res.stdout, "if ! (( $+functions[zsh-defer] )); then"))
	Expect(t, Contains(res.stdout, "my-defer.plugin.zsh"))
	Expect(t, Contains(res.stdout, "zsh-defer source "))
}

func TestCLIConfigFpathRule(t *testing.T) {
	plugin := pluginFixture(t)
	home := t.TempDir()
	writeConfig(t, home, "[fpath]\nrule = \"prepend\"\n")

	res := runCLI(t, home, "", "bundle", plugin.URL())
	Expect(t, Equals(0, res.exitCode), res.stderr)
	Expect(t, Contains(res.stdout, "fpath=( "))
	Expect(t, Contains(res.stdout, " $fpath )"))
	Expect(t, Not(Contains(res.stdout, "fpath+=( ")))
}

func TestCLIConfigGitDomain(t *testing.T) {
	home := t.TempDir()
	writeConfig(t, home, "[git]\ndomain = \"localhost:1\"\n")

	res := runCLI(t, home, "", "bundle", "foo/bar")
	Expect(t, res.exitCode != 0, "expected failure exit code")
	Expect(t, Contains(res.stderr, "https://localhost:1/foo/bar"))
}

func TestCLIConfigMalformed(t *testing.T) {
	home := t.TempDir()
	writeConfig(t, home, "not [valid toml\n")

	res := runCLI(t, home, "", "home")
	Expect(t, res.exitCode != 0, "expected failure exit code")
	Expect(t, Contains(res.stderr, "failed to parse config"))
}
