package main

import "fmt"

// one-line flag descriptions, shared by flag registration and help text
const (
	parallelismUsage = "max amount of tasks to launch in parallel"
	versionUsage     = "Show application version."
	dirsUsage        = "show bundle directory paths"
	urlUsage         = "show bundle URLs only"
	fpathUsage       = "write the script to the completions dir and print the file path"
)

const parallelismHelp = "  -p, --parallelism=N  " + parallelismUsage + "\n"

const listHelp = `usage: antibody list [<flags>]

lists all currently installed bundles

Flags:
  -d, --dirs  ` + dirsUsage + `
  -u, --url   ` + urlUsage + `
`

const mainUsage = `usage: antibody [<flags>] <command> [<args> ...]

The fastest shell plugin manager

Flags:
  -h, --help           Show context-sensitive help.
  -v, --version        ` + versionUsage + `
` + parallelismHelp + `
Commands:
  help [<command>]
    Show help.

  bundle [<bundles>...]
    downloads a bundle and prints its source line

  update
    updates all previously bundled bundles

  home
    prints where antibody is cloning the bundles

  purge <bundle>
    purges a bundle from your computer

  list [<flags>]
    lists all currently installed bundles

  path <bundle>
    prints the path of a currently cloned bundle

  init
    initializes the shell so Antibody can work as expected

  completions [<flags>] <shell>
    generates shell completion scripts
`

// nolint: gochecknoglobals
var cmdHelp = map[string]string{
	"help": `usage: antibody help [<command>]

Show help.
`,
	"bundle": `usage: antibody bundle [<bundles>...]

downloads a bundle and prints its source line; reads stdin when no
bundles are given

Flags:
` + parallelismHelp,
	"update": `usage: antibody update

updates all previously bundled bundles

Flags:
` + parallelismHelp,
	"home": `usage: antibody home

prints where antibody is cloning the bundles
`,
	"purge": `usage: antibody purge <bundle>

purges a bundle from your computer
`,
	"list": listHelp,
	"ls":   listHelp,
	"path": `usage: antibody path <bundle>

prints the path of a currently cloned bundle
`,
	"init": `usage: antibody init

initializes the shell so Antibody can work as expected
`,
	"completions": `usage: antibody completions [<flags>] <shell>

generates shell completion scripts

Flags:
      --fpath  ` + fpathUsage + `
`,
}

func usageFor(cmd string) string {
	if u, ok := cmdHelp[cmd]; ok {
		return u
	}
	return mainUsage
}

// help implements the help subcommand: no args prints the main usage,
// otherwise the named command's usage.
func help(args []string) {
	if len(args) == 0 {
		fmt.Print(mainUsage)
		return
	}
	topic := args[0]
	u, ok := cmdHelp[topic]
	if !ok {
		fatalf("unknown help topic %q, try --help", topic)
	}
	fmt.Print(u)
}
