package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mattmc3/antibody/internal/gittest"
	"github.com/stretchr/testify/require"
)

// nolint: gochecknoglobals
var binPath string

// TestMain builds the antibody binary once so tests can exercise the real
// command surface: flag parsing, stdin handling, output, and exit codes.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "antibody-cli")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	binPath = filepath.Join(dir, "antibody")
	if out, err := exec.Command("go", "build", "-o", binPath, ".").CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "go build failed: %v: %s\n", err, out)
		os.Exit(1)
	}
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

type cliResult struct {
	stdout   string
	stderr   string
	exitCode int
}

// runCLI invokes the built binary with an isolated home and config dir.
func runCLI(t *testing.T, home, stdin string, args ...string) cliResult {
	t.Helper()
	cmd := exec.Command(binPath, args...)
	cmd.Env = append(os.Environ(),
		"ANTIBODY_HOME="+home,
		"XDG_CONFIG_HOME="+filepath.Join(home, ".config"),
		"GIT_TERMINAL_PROMPT=0",
	)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var out, errb strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err := cmd.Run()
	code := 0
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		require.True(t, ok, "antibody did not run: %v", err)
		code = exitErr.ExitCode()
	}
	return cliResult{stdout: out.String(), stderr: errb.String(), exitCode: code}
}

func pluginFixture(t *testing.T) *gittest.Repo {
	t.Helper()
	r := gittest.New(t)
	r.WriteFile("myplugin.plugin.zsh", "echo myplugin\n")
	r.Commit("add plugin file")
	return r
}

func TestCLIBundleArg(t *testing.T) {
	upstream := pluginFixture(t)
	res := runCLI(t, t.TempDir(), "", "bundle", upstream.URL())
	require.Equal(t, 0, res.exitCode, res.stderr)
	require.Contains(t, res.stdout, "fpath+=( ")
	require.Contains(t, res.stdout, "source ")
	require.Contains(t, res.stdout, "myplugin.plugin.zsh")
}

func TestCLIBundleStdin(t *testing.T) {
	a := pluginFixture(t)
	b := pluginFixture(t)
	input := a.URL() + "\n" + b.URL() + " kind:path\n"
	res := runCLI(t, t.TempDir(), input, "bundle")
	require.Equal(t, 0, res.exitCode, res.stderr)
	require.Contains(t, res.stdout, "source ")
	require.Contains(t, res.stdout, `export PATH="`)
}

func TestCLIBundleError(t *testing.T) {
	res := runCLI(t, t.TempDir(), "", "bundle", "file:///this/path/does/not/exist")
	require.NotEqual(t, 0, res.exitCode)
	require.Contains(t, res.stderr, "antibody: error: failed to bundle")
}

func TestCLIHome(t *testing.T) {
	home := t.TempDir()
	res := runCLI(t, home, "", "home")
	require.Equal(t, 0, res.exitCode, res.stderr)
	require.Equal(t, home+"\n", res.stdout)
}

func TestCLIListAndPath(t *testing.T) {
	upstream := pluginFixture(t)
	home := t.TempDir()
	require.Equal(t, 0, runCLI(t, home, "", "bundle", upstream.URL()).exitCode)

	list := runCLI(t, home, "", "list")
	require.Equal(t, 0, list.exitCode, list.stderr)
	require.Contains(t, list.stdout, upstream.URL())
	require.Contains(t, list.stdout, home)

	path := runCLI(t, home, "", "path", upstream.URL())
	require.Equal(t, 0, path.exitCode, path.stderr)
	require.Contains(t, path.stdout, home)

	missing := runCLI(t, home, "", "path", "not/cloned")
	require.NotEqual(t, 0, missing.exitCode)
	require.Contains(t, missing.stderr, "does not exist in cloned paths")
}

func TestCLIPurge(t *testing.T) {
	upstream := pluginFixture(t)
	home := t.TempDir()
	require.Equal(t, 0, runCLI(t, home, "", "bundle", upstream.URL()).exitCode)

	purge := runCLI(t, home, "", "purge", upstream.URL())
	require.Equal(t, 0, purge.exitCode, purge.stderr)
	require.Contains(t, purge.stdout, "removed!")

	entries, err := os.ReadDir(home)
	require.NoError(t, err)
	for _, e := range entries {
		require.False(t, strings.Contains(e.Name(), "-SLASH-"), "clone left behind: %s", e.Name())
	}

	again := runCLI(t, home, "", "purge", upstream.URL())
	require.NotEqual(t, 0, again.exitCode)
	require.Contains(t, again.stderr, "does not exist")
}

func TestCLIUpdate(t *testing.T) {
	upstream := pluginFixture(t)
	home := t.TempDir()
	require.Equal(t, 0, runCLI(t, home, "", "bundle", upstream.URL()).exitCode)

	res := runCLI(t, home, "", "update")
	require.Equal(t, 0, res.exitCode, res.stderr)
	require.Contains(t, res.stdout, "Updating all bundles in "+home)
}

func TestCLIInit(t *testing.T) {
	res := runCLI(t, t.TempDir(), "", "init")
	require.Equal(t, 0, res.exitCode, res.stderr)
	require.Contains(t, res.stdout, "antibody")
}

func TestCLIVersion(t *testing.T) {
	for _, flag := range []string{"--version", "-v"} {
		res := runCLI(t, t.TempDir(), "", flag)
		require.Equal(t, 0, res.exitCode)
		require.Contains(t, res.stdout+res.stderr, "antibody version")
	}
}

func TestCLIHelp(t *testing.T) {
	// no args prints usage too
	for _, args := range [][]string{{"-h"}, {"--help"}, {"help"}, {}} {
		res := runCLI(t, t.TempDir(), "", args...)
		require.Equal(t, 0, res.exitCode, res.stderr)
		out := res.stdout + res.stderr
		require.Contains(t, out, "Commands:")
		require.Contains(t, out, "bundle")
	}
}

func TestCLIHelpCommand(t *testing.T) {
	res := runCLI(t, t.TempDir(), "", "help", "purge")
	require.Equal(t, 0, res.exitCode, res.stderr)
	require.Contains(t, res.stdout+res.stderr, "purge")

	// ls resolves to list help
	res = runCLI(t, t.TempDir(), "", "help", "ls")
	require.Equal(t, 0, res.exitCode, res.stderr)
	require.Contains(t, res.stdout, "antibody list")

	// -h works after a subcommand and shows that command's help
	res = runCLI(t, t.TempDir(), "", "bundle", "-h")
	require.Equal(t, 0, res.exitCode, res.stderr)
	require.Contains(t, res.stdout, "antibody bundle")
}

func TestCLIListVariants(t *testing.T) {
	upstream := pluginFixture(t)
	home := t.TempDir()
	require.Equal(t, 0, runCLI(t, home, "", "bundle", upstream.URL()).exitCode)

	list := runCLI(t, home, "", "list")
	ls := runCLI(t, home, "", "ls")
	require.Equal(t, 0, ls.exitCode, ls.stderr)
	require.Equal(t, list.stdout, ls.stdout)

	dirs := runCLI(t, home, "", "list", "-d")
	require.Equal(t, 0, dirs.exitCode, dirs.stderr)
	require.Contains(t, dirs.stdout, home)
	require.NotContains(t, dirs.stdout, upstream.URL())

	urls := runCLI(t, home, "", "list", "--url")
	require.Equal(t, 0, urls.exitCode, urls.stderr)
	require.Contains(t, urls.stdout, upstream.URL())
	require.NotContains(t, urls.stdout, home)
}

func TestCLIParallelismFlagPosition(t *testing.T) {
	home := t.TempDir()
	// global flag works before and after the subcommand
	before := runCLI(t, home, "", "-p", "2", "home")
	require.Equal(t, 0, before.exitCode, before.stderr)
	require.Equal(t, home+"\n", before.stdout)

	after := runCLI(t, home, "", "home", "-p", "2")
	require.Equal(t, 0, after.exitCode, after.stderr)
	require.Equal(t, home+"\n", after.stdout)
}

func TestCLIUsageErrors(t *testing.T) {
	cases := map[string][]string{
		"unknown command": {"frobnicate"},
		"unknown flag":    {"--bogus"},
		"missing arg":     {"purge"},
		"missing shell":   {"completions"},
	}
	for name, args := range cases {
		res := runCLI(t, t.TempDir(), "", args...)
		require.NotEqual(t, 0, res.exitCode, name)
		require.Contains(t, res.stderr, "antibody: error:", name)
	}
}

func TestCLICompletions(t *testing.T) {
	res := runCLI(t, t.TempDir(), "", "completions", "zsh")
	require.Equal(t, 0, res.exitCode, res.stderr)
	require.Contains(t, res.stdout, "#compdef antibody")

	bad := runCLI(t, t.TempDir(), "", "completions", "fish")
	require.NotEqual(t, 0, bad.exitCode)
	require.Contains(t, bad.stderr, "antibody: error:")
}
