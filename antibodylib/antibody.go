package antibodylib

import (
	"bufio"
	"io"
	"strings"

	"github.com/getantidote/bundleparse"
	"github.com/mattmc3/antibody/bundle"
	"github.com/mattmc3/antibody/internal/config"
	"golang.org/x/sync/errgroup"
)

// Antibody the main thing
type Antibody struct {
	r           io.Reader
	parallelism int
	Home        string
}

// New creates a new Antibody instance with the given parameters
func New(home string, r io.Reader, p int) *Antibody {
	return &Antibody{
		r:           r,
		parallelism: p,
		Home:        home,
	}
}

// Bundle processes all given lines and returns the shell content to execute
func (a *Antibody) Bundle() (string, error) {
	var lines []string
	scanner := bufio.NewScanner(a.r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	hasDefer := false
	input := strings.Join(lines, "\n")
	parsedBundles, err := bundleparse.ParseBundles(input)
	if err != nil {
		return "", err
	}
	for _, parsed := range parsedBundles {
		if parsed.Kind == bundleparse.KindDefer {
			hasDefer = true
			break
		}
	}

	var g errgroup.Group
	var shs safeIndexedLines
	var sem = make(chan bool, a.parallelism)

	if hasDefer {
		sem <- true
		g.Go(func() error {
			defer func() { <-sem }()
			sh, err := deferEnsure(a.Home)
			if err != nil {
				return err
			}
			shs.Append(indexedLine{idx: -1, line: sh})
			return nil
		})
	}

	for i, parsed := range parsedBundles {
		i := i
		parsed := parsed
		sem <- true
		g.Go(func() error {
			defer func() { <-sem }()
			lineBundle, berr := bundle.NewFromParsed(a.Home, parsed)
			if berr != nil {
				return berr
			}
			sh, berr := lineBundle.Get()
			shs.Append(indexedLine{idx: i, line: sh})
			return berr
		})
	}

	if err := g.Wait(); err != nil {
		return "", err
	}
	return shs.Items().String(), nil
}

// deferEnsure builds the load-once ensure block for the zsh-defer tool.
func deferEnsure(home string) (string, error) {
	cfg := config.Get()
	deferBundle, err := bundle.New(home, cfg.DeferBundle())
	if err != nil {
		return "", err
	}
	deferSrc, err := deferBundle.Get()
	if err != nil {
		return "", err
	}
	var out []string
	out = append(out, "if ! (( $+functions[zsh-defer] )); then")
	for _, line := range strings.Split(deferSrc, "\n") {
		if line != "" {
			out = append(out, "  "+line)
		}
	}
	out = append(out, "fi")
	return strings.Join(out, "\n"), nil
}

// Home returns the directory where bundles are cloned.
func Home() (string, error) {
	return config.Get().HomeDir()
}
