package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/mattmc3/antibody/antibodylib"
	"github.com/mattmc3/antibody/internal/config"
	"github.com/mattmc3/antibody/project"
	"github.com/mattmc3/antibody/shell"
	"golang.org/x/term"
	"gopkg.in/alecthomas/kingpin.v2"
)

// nolint: gochecknoglobals
var (
	version = "7.0.0-dev"

	app         = kingpin.New("antibody", "The fastest shell plugin manager")
	parallelism = app.Flag("parallelism", "max amount of tasks to launch in parallel").
			Short('p').
			Default(strconv.Itoa(runtime.NumCPU())).
			Int()
	bundleCmd = app.Command("bundle", "downloads a bundle and prints its source line")
	bundles   = bundleCmd.Arg("bundles", "bundle list").Strings()
	updateCmd = app.Command("update", "updates all previously bundled bundles")
	homeCmd   = app.Command("home", "prints where antibody is cloning the bundles")
	purgeCmd  = app.Command("purge", "purges a bundle from your computer")
	purgee    = purgeCmd.Arg("bundle", "bundle to be purged").Required().String()
	listCmd   = app.Command("list", "lists all currently installed bundles").Alias("ls")
	listDirs  = listCmd.Flag("dirs", "show bundle directory paths").Short('d').Bool()
	listURL   = listCmd.Flag("url", "show bundle URLs only").Short('u').Bool()
	pathCmd   = app.Command("path", "prints the path of a currently cloned bundle")
	pathee    = pathCmd.Arg("bundle", "bundle in which to find and print cloned path").Required().String()
	initCmd   = app.Command("init", "initializes the shell so Antibody can work as expected")

	completionsCmd   = app.Command("completions", "generates shell completion scripts")
	completionsFor   = completionsCmd.Arg("shell", "shell to generate completions for").Required().String()
	completionsFpath = completionsCmd.Flag("fpath", "write the script to the completions dir and print the file path").Bool()
)

// nolint: gochecknoinits
func init() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("antibody: ")
	log.SetFlags(0)
}

func main() {
	_, err := config.Load()
	app.FatalIfError(err, "failed to parse config")

	app.Author("Carlos Alexandro Becker <caarlos0@gmail.com>")
	app.Version("antibody version " + version)
	app.VersionFlag.Short('v')
	app.HelpFlag.Short('h')

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case bundleCmd.FullCommand():
		bundle()
	case updateCmd.FullCommand():
		update()
	case homeCmd.FullCommand():
		fmt.Println(findHome())
	case purgeCmd.FullCommand():
		purge()
	case listCmd.FullCommand():
		list()
	case pathCmd.FullCommand():
		path()
	case initCmd.FullCommand():
		sh, err := shell.Init()
		app.FatalIfError(err, "failed to init")
		fmt.Println(sh)
	case completionsCmd.FullCommand():
		if *completionsFpath {
			file, err := shell.CompletionsFpath(*completionsFor)
			app.FatalIfError(err, "failed to generate completions")
			fmt.Println(file)
		} else {
			sh, err := shell.Completions(*completionsFor)
			app.FatalIfError(err, "failed to generate completions")
			fmt.Println(sh)
		}
	}
}

func bundle() {
	var input io.Reader
	if !term.IsTerminal(int(os.Stdin.Fd())) && len(*bundles) == 0 {
		input = os.Stdin
	} else {
		input = bytes.NewBufferString(strings.Join(*bundles, " "))
	}
	sh, err := antibodylib.New(findHome(), input, *parallelism).Bundle()
	app.FatalIfError(err, "failed to bundle")
	fmt.Println(sh)
}

func update() {
	var home = findHome()
	fmt.Printf("Updating all bundles in %v...\n", home)
	var err = project.Update(home, *parallelism)
	app.FatalIfError(err, "failed to update")
}

func purge() {
	home := findHome()
	fmt.Printf("Removing %s...\n", *purgee)
	root := project.CloneRoot(home, *purgee)
	candidates := []string{root}

	pinnedGlob := root + "-SLASH-*"
	if matches, err := filepath.Glob(pinnedGlob); err == nil {
		candidates = append(candidates, matches...)
	}

	removed := false
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			app.FatalIfError(os.RemoveAll(path), "failed to purge")
			removed = true
		}
	}
	if !removed {
		app.Fatalf("%s does not exist on expected location: %s", *purgee, root)
	}
	fmt.Println("removed!")
}

func list() {
	var home = findHome()
	projects, err := project.List(home)
	app.FatalIfError(err, "failed to list bundles")

	switch {
	case *listDirs:
		for _, b := range projects {
			fmt.Println(filepath.Join(home, b))
		}
	case *listURL:
		for _, b := range projects {
			fmt.Println(project.EscapedPathToURL(b))
		}
	default:
		w := tabwriter.NewWriter(os.Stdout, 0, 1, 4, ' ', tabwriter.TabIndent)
		for _, b := range projects {
			if _, err := fmt.Fprintf(w, "%s\t%s\n", project.EscapedPathToURL(b), filepath.Join(home, b)); err != nil {
				app.FatalIfError(err, "failed to write")
			}
		}
		app.FatalIfError(w.Flush(), "failed to flush")
	}
}

func path() {
	home := findHome()
	root := project.CloneRoot(home, *pathee)
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
		app.Fatalf("%s does not exist in cloned paths", *pathee)
	}

	for _, path := range existing {
		fmt.Println(path)
	}
}

func findHome() string {
	h, err := antibodylib.Home()
	if err != nil {
		app.Fatalf("could't get cache folder: %v", err)
	}
	return h
}
