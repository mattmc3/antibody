package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/mattmc3/antibody/internal/expect"
	"github.com/mattmc3/antibody/internal/gittest"
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
		Expect(t, ok, "antibody did not run: %v", err)
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
	Expect(t, Equals(0, res.exitCode), res.stderr)
	Expect(t, Contains(res.stdout, "fpath+=( "))
	Expect(t, Contains(res.stdout, "source "))
	Expect(t, Contains(res.stdout, "myplugin.plugin.zsh"))
}

func TestCLIBundleStdin(t *testing.T) {
	a := pluginFixture(t)
	b := pluginFixture(t)
	input := a.URL() + "\n" + b.URL() + " kind:path\n"
	res := runCLI(t, t.TempDir(), input, "bundle")
	Expect(t, Equals(0, res.exitCode), res.stderr)
	Expect(t, Contains(res.stdout, "source "))
	Expect(t, Contains(res.stdout, `export PATH="`))
}

func TestCLIBundleError(t *testing.T) {
	res := runCLI(t, t.TempDir(), "", "bundle", "file:///this/path/does/not/exist")
	Expect(t, res.exitCode != 0, "expected failure exit code")
	Expect(t, Contains(res.stderr, "antibody: error: failed to bundle"))
}

func TestCLIHome(t *testing.T) {
	home := t.TempDir()
	res := runCLI(t, home, "", "home")
	Expect(t, Equals(0, res.exitCode), res.stderr)
	Expect(t, Equals(home+"\n", res.stdout))
}

func TestCLIListAndPath(t *testing.T) {
	upstream := pluginFixture(t)
	home := t.TempDir()
	Expect(t, Equals(0, runCLI(t, home, "", "bundle", upstream.URL()).exitCode))

	list := runCLI(t, home, "", "list")
	Expect(t, Equals(0, list.exitCode), list.stderr)
	Expect(t, Contains(list.stdout, upstream.URL()))
	Expect(t, Contains(list.stdout, home))

	path := runCLI(t, home, "", "path", upstream.URL())
	Expect(t, Equals(0, path.exitCode), path.stderr)
	Expect(t, Contains(path.stdout, home))

	missing := runCLI(t, home, "", "path", "not/cloned")
	Expect(t, missing.exitCode != 0, "expected failure exit code")
	Expect(t, Contains(missing.stderr, "does not exist in cloned paths"))
}

func TestCLIPurge(t *testing.T) {
	upstream := pluginFixture(t)
	home := t.TempDir()
	Expect(t, Equals(0, runCLI(t, home, "", "bundle", upstream.URL()).exitCode))

	purge := runCLI(t, home, "", "purge", upstream.URL())
	Expect(t, Equals(0, purge.exitCode), purge.stderr)
	Expect(t, Contains(purge.stdout, "removed!"))

	entries, err := os.ReadDir(home)
	Expect(t, NoError(err))
	for _, e := range entries {
		Expect(t, Not(Contains(e.Name(), "-SLASH-")), "clone left behind: %s", e.Name())
	}

	again := runCLI(t, home, "", "purge", upstream.URL())
	Expect(t, again.exitCode != 0, "expected failure exit code")
	Expect(t, Contains(again.stderr, "does not exist"))
}

func TestCLIUpdate(t *testing.T) {
	upstream := pluginFixture(t)
	home := t.TempDir()
	Expect(t, Equals(0, runCLI(t, home, "", "bundle", upstream.URL()).exitCode))

	res := runCLI(t, home, "", "update")
	Expect(t, Equals(0, res.exitCode), res.stderr)
	Expect(t, Contains(res.stdout, "Updating all bundles in "+home))
}

func TestCLIInit(t *testing.T) {
	res := runCLI(t, t.TempDir(), "", "init")
	Expect(t, Equals(0, res.exitCode), res.stderr)
	Expect(t, Contains(res.stdout, "antibody"))
}

func TestCLIVersion(t *testing.T) {
	for _, flag := range []string{"--version", "-v"} {
		res := runCLI(t, t.TempDir(), "", flag)
		Expect(t, Equals(0, res.exitCode))
		Expect(t, Contains(res.stdout+res.stderr, "antibody version"))
	}
}

func TestCLIHelp(t *testing.T) {
	// no args prints usage too
	for _, args := range [][]string{{"-h"}, {"--help"}, {"help"}, {}} {
		res := runCLI(t, t.TempDir(), "", args...)
		Expect(t, Equals(0, res.exitCode), res.stderr)
		out := res.stdout + res.stderr
		Expect(t, Contains(out, "Commands:"))
		Expect(t, Contains(out, "bundle"))
	}
}

func TestCLIHelpCommand(t *testing.T) {
	res := runCLI(t, t.TempDir(), "", "help", "purge")
	Expect(t, Equals(0, res.exitCode), res.stderr)
	Expect(t, Contains(res.stdout+res.stderr, "purge"))

	// ls resolves to list help
	res = runCLI(t, t.TempDir(), "", "help", "ls")
	Expect(t, Equals(0, res.exitCode), res.stderr)
	Expect(t, Contains(res.stdout, "antibody list"))

	// -h works after a subcommand and shows that command's help
	res = runCLI(t, t.TempDir(), "", "bundle", "-h")
	Expect(t, Equals(0, res.exitCode), res.stderr)
	Expect(t, Contains(res.stdout, "antibody bundle"))
}

func TestCLIListVariants(t *testing.T) {
	upstream := pluginFixture(t)
	home := t.TempDir()
	Expect(t, Equals(0, runCLI(t, home, "", "bundle", upstream.URL()).exitCode))

	list := runCLI(t, home, "", "list")
	ls := runCLI(t, home, "", "ls")
	Expect(t, Equals(0, ls.exitCode), ls.stderr)
	Expect(t, Equals(list.stdout, ls.stdout))

	dirs := runCLI(t, home, "", "list", "-d")
	Expect(t, Equals(0, dirs.exitCode), dirs.stderr)
	Expect(t, Contains(dirs.stdout, home))
	Expect(t, Not(Contains(dirs.stdout, upstream.URL())))

	urls := runCLI(t, home, "", "list", "--url")
	Expect(t, Equals(0, urls.exitCode), urls.stderr)
	Expect(t, Contains(urls.stdout, upstream.URL()))
	Expect(t, Not(Contains(urls.stdout, home)))
}

func TestCLIParallelismFlagPosition(t *testing.T) {
	home := t.TempDir()
	// global flag works before and after the subcommand
	before := runCLI(t, home, "", "-p", "2", "home")
	Expect(t, Equals(0, before.exitCode), before.stderr)
	Expect(t, Equals(home+"\n", before.stdout))

	after := runCLI(t, home, "", "home", "-p", "2")
	Expect(t, Equals(0, after.exitCode), after.stderr)
	Expect(t, Equals(home+"\n", after.stdout))
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
		Expect(t, res.exitCode != 0, name)
		Expect(t, Contains(res.stderr, "antibody: error:"), name)
	}
}

func TestCLICompletions(t *testing.T) {
	res := runCLI(t, t.TempDir(), "", "completions", "zsh")
	Expect(t, Equals(0, res.exitCode), res.stderr)
	Expect(t, Contains(res.stdout, "#compdef antibody"))

	bad := runCLI(t, t.TempDir(), "", "completions", "fish")
	Expect(t, bad.exitCode != 0, "expected failure exit code")
	Expect(t, Contains(bad.stderr, "antibody: error:"))
}
