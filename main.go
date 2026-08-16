package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/mattmc3/antibody/antibodylib"
	"github.com/mattmc3/antibody/bundleparse"
	"github.com/mattmc3/antibody/internal/config"
	"github.com/mattmc3/antibody/project"
	"github.com/mattmc3/antibody/shell"
)

// nolint: gochecknoglobals
var version = "7.0.1-dev"

// cliOptions holds the values of all command line flags.
type cliOptions struct {
	parallelism int
	showVersion bool
	dirs        bool
	urls        bool
	fpath       bool
}

// nolint: gochecknoglobals
var opts = cliOptions{parallelism: runtime.NumCPU()}

// nolint: gochecknoinits
func init() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("antibody: ")
	log.SetFlags(0)
}

func main() {
	_, err := config.Load()
	fatalIfError(err, "failed to parse config")

	rootFlags := newFlagSet("antibody")
	rootFlags.BoolVar(&opts.showVersion, "version", false, versionUsage)
	rootFlags.Alias("v", "version")
	rest := rootFlags.parse(os.Args[1:])

	if opts.showVersion {
		fmt.Println("antibody version " + version)
		return
	}
	if len(rest) == 0 {
		help(nil)
		return
	}

	cmd, args := rest[0], rest[1:]
	switch cmd {
	case "help":
		help(args)
	case "bundle":
		bundle(newFlagSet("bundle").parse(args))
	case "update":
		newFlagSet("update").parse(args)
		update()
	case "home":
		newFlagSet("home").parse(args)
		fmt.Println(findHome())
	case "purge":
		purge(requiredArg(newFlagSet("purge").parse(args), "bundle"))
	case "list", "ls":
		fs := newFlagSet("list")
		fs.BoolVar(&opts.dirs, "dirs", false, dirsUsage)
		fs.Alias("d", "dirs")
		fs.BoolVar(&opts.urls, "url", false, urlUsage)
		fs.Alias("u", "url")
		fs.parse(args)
		list(opts.dirs, opts.urls)
	case "path":
		path(requiredArg(newFlagSet("path").parse(args), "bundle"))
	case "init":
		newFlagSet("init").parse(args)
		sh, err := shell.Init()
		fatalIfError(err, "failed to init")
		fmt.Println(sh)
	case "completions":
		fs := newFlagSet("completions")
		fs.BoolVar(&opts.fpath, "fpath", false, fpathUsage)
		completions(requiredArg(fs.parse(args), "shell"), opts.fpath)
	default:
		fatalf("expected command but got %q, try --help", cmd)
	}
}

func requiredArg(rest []string, name string) string {
	if len(rest) == 0 {
		fatalf("required argument '%s' not provided, try --help", name)
	}
	return rest[0]
}

func bundle(bundles []string) {
	var input io.Reader
	if stdinPiped() && len(bundles) == 0 {
		input = os.Stdin
	} else {
		input = bytes.NewBufferString(strings.Join(bundles, " "))
	}
	ab := antibodylib.New(findHome(), input, opts.parallelism)
	ab.Presets = presetsFromEnv()
	sh, err := ab.Bundle()
	fatalIfError(err, "failed to bundle")
	fmt.Println(sh)
	if export := presetExport(ab.Presets); export != "" {
		fmt.Println(export)
	}
}

// presetsFromEnv reads the presets an earlier dynamic-mode call exported. The
// export outlives that call, so a static bundling must not pick it up.
func presetsFromEnv() bundleparse.Presets {
	raw := os.Getenv("ANTIBODY_PRESETS")
	if raw == "" || !dynamicMode() {
		return nil
	}
	var presets bundleparse.Presets
	if err := json.Unmarshal([]byte(raw), &presets); err != nil {
		log.Printf("ignoring unreadable $ANTIBODY_PRESETS: %v", err)
		return nil
	}
	return presets
}

// presetExport returns the line that hands presets to the next dynamic-mode
// call, which the shell picks up by sourcing this output.
func presetExport(presets bundleparse.Presets) string {
	if len(presets) == 0 || !dynamicMode() {
		return ""
	}
	data, err := json.Marshal(presets)
	if err != nil {
		log.Printf("could not export presets: %v", err)
		return ""
	}
	return "export ANTIBODY_PRESETS=" + singleQuote(string(data))
}

// dynamicMode reports whether antibody init's shell function is sourcing this
// output, which is the only case where preset state carries between calls.
func dynamicMode() bool {
	return os.Getenv("ANTIBODY_DYNAMIC") == "true"
}

func singleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func update() {
	var home = findHome()
	fmt.Printf("Updating all bundles in %v...\n", home)
	var err = project.Update(home, opts.parallelism)
	fatalIfError(err, "failed to update")
}

func purge(bundle string) {
	home := findHome()
	fmt.Printf("Removing %s...\n", bundle)
	root, err := project.CloneRoot(home, bundle)
	fatalIfError(err, "failed to purge")
	candidates := []string{root}

	pinnedGlob := root + "-SLASH-*"
	if matches, err := filepath.Glob(pinnedGlob); err == nil {
		candidates = append(candidates, matches...)
	}

	removed := false
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			fatalIfError(os.RemoveAll(path), "failed to purge")
			removed = true
		}
	}
	if !removed {
		fatalf("%s does not exist on expected location: %s", bundle, root)
	}
	fmt.Println("removed!")
}

func list(dirs, urls bool) {
	var home = findHome()
	projects, err := project.List(home)
	fatalIfError(err, "failed to list bundles")

	switch {
	case dirs:
		for _, b := range projects {
			fmt.Println(filepath.Join(home, b))
		}
	case urls:
		for _, b := range projects {
			fmt.Println(project.EscapedPathToURL(b))
		}
	default:
		w := tabwriter.NewWriter(os.Stdout, 0, 1, 4, ' ', tabwriter.TabIndent)
		for _, b := range projects {
			if _, err := fmt.Fprintf(w, "%s\t%s\n", project.EscapedPathToURL(b), filepath.Join(home, b)); err != nil {
				fatalIfError(err, "failed to write")
			}
		}
		fatalIfError(w.Flush(), "failed to flush")
	}
}

func path(bundle string) {
	home := findHome()
	root, err := project.CloneRoot(home, bundle)
	fatalIfError(err, "failed to find path")
	paths := []string{root}
	pinnedGlob := root + "-SLASH-*"
	if matches, err := filepath.Glob(pinnedGlob); err == nil {
		paths = append(paths, matches...)
	}

	existing := []string{}
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			existing = append(existing, path)
		}
	}

	if len(existing) == 0 {
		fatalf("%s does not exist in cloned paths", bundle)
	}

	for _, path := range existing {
		fmt.Println(path)
	}
}

func completions(forShell string, fpath bool) {
	if fpath {
		file, err := shell.CompletionsFpath(forShell)
		fatalIfError(err, "failed to generate completions")
		fmt.Println(file)
	} else {
		sh, err := shell.Completions(forShell)
		fatalIfError(err, "failed to generate completions")
		fmt.Println(sh)
	}
}

// stdinPiped reports whether stdin is not a terminal.
func stdinPiped() bool {
	stat, err := os.Stdin.Stat()
	return err == nil && stat.Mode()&os.ModeCharDevice == 0
}

func findHome() string {
	h, err := antibodylib.Home()
	if err != nil {
		fatalf("could't get cache folder: %v", err)
	}
	return h
}

func fatalf(format string, args ...any) {
	log.Fatalf("error: "+format, args...)
}

func fatalIfError(err error, msg string) {
	if err != nil {
		fatalf("%s: %v", msg, err)
	}
}
