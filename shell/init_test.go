package shell

import (
	"os"
	"strings"
	"testing"

	. "github.com/mattmc3/antibody/internal/expect"
)

func TestGeneratesInit(t *testing.T) {
	shell, err := Init()
	Expect(t, NoError(err))
	Expect(t, shell != "", "init script is empty")
}

// Args must be quoted or arguments with spaces word-split in zsh.
func TestInitQuotesArgs(t *testing.T) {
	shell, err := Init()
	Expect(t, NoError(err))
	Expect(t, Contains(shell, `"$@"`))
	Expect(t, Not(Matches(`\$@[^"]`, shell)))
}

func TestInitUsesCompletionsSubcommand(t *testing.T) {
	shell, err := Init()
	Expect(t, NoError(err))
	Expect(t, Contains(shell, "completions zsh"))
	Expect(t, Contains(shell, "--fpath"))
	Expect(t, Not(Contains(shell, "compctl")))
	// skip completion setup when _antibody is already defined
	Expect(t, Contains(shell, "$+functions[_antibody]"))
}

func TestInitMarksBundleCallsDynamic(t *testing.T) {
	shell, err := Init()
	Expect(t, NoError(err))
	Expect(t, Contains(shell, "ANTIBODY_DYNAMIC=true"))
}

func TestInitNamesBinaryOnce(t *testing.T) {
	shell, err := Init()
	Expect(t, NoError(err))
	bin, err := os.Executable()
	Expect(t, NoError(err))
	Expect(t, Equals(1, strings.Count(shell, bin)))
	Expect(t, Contains(shell, `_ANTIBODY_BIN="`+bin+`"`))
}
