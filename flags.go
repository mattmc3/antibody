package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

// flagSet wraps flag.FlagSet with short-flag aliases and help and error
// handling wired to the CLI's usage text.
type flagSet struct {
	*flag.FlagSet
}

// newFlagSet returns a flagSet named for its command, with the global
// -p/--parallelism flag registered so it works before or after the
// subcommand.
func newFlagSet(name string) *flagSet {
	fs := &flagSet{flag.NewFlagSet(name, flag.ContinueOnError)}
	fs.SetOutput(io.Discard)
	fs.IntVar(&opts.parallelism, "parallelism", opts.parallelism, parallelismUsage)
	fs.Alias("p", "parallelism")
	return fs
}

// Alias registers short as another name for an already registered flag.
// Sharing the flag.Value keeps bool semantics and makes either spelling
// set the same variable.
func (fs *flagSet) Alias(short, long string) {
	f := fs.Lookup(long)
	if f == nil {
		panic("alias of unregistered flag: " + long)
	}
	fs.Var(f.Value, short, f.Usage)
}

// parse handles -h/--help (print usage, exit 0) and parse errors (fatal),
// and returns the remaining positional arguments.
func (fs *flagSet) parse(args []string) []string {
	err := fs.Parse(args)
	if errors.Is(err, flag.ErrHelp) {
		fmt.Print(usageFor(fs.Name()))
		os.Exit(0)
	}
	if err != nil {
		fatalf("%v, try --help", err)
	}
	return fs.Args()
}
