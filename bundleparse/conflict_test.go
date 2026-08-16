package bundleparse

import (
	"strings"
	"testing"

	. "github.com/mattmc3/antibody/internal/expect"
)

func TestParseBundles_ConflictingBranch(t *testing.T) {
	_, err := ParseBundles("foo/bar branch:main\nfoo/bar branch:next")
	Expect(t, AnError(err))
	Expect(t, Contains(err.Error(), "line 2"))
	Expect(t, Contains(err.Error(), "conflicting branch"))
}

func TestParseBundles_ConflictingBranchAcrossSpellings(t *testing.T) {
	_, err := ParseBundles("foo/bar branch:main\nhttps://github.com/foo/bar branch:next")
	Expect(t, AnError(err))
	Expect(t, Contains(err.Error(), "conflicting branch"))
}

func TestParseBundles_InconsistentBranch(t *testing.T) {
	inputs := []string{
		"foo/bar branch:main\nfoo/bar path:lib",
		"foo/bar path:lib\nfoo/bar branch:main",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			_, err := ParseBundles(input)
			Expect(t, AnError(err))
			Expect(t, Contains(err.Error(), "inconsistent branch"))
		})
	}
}

func TestParseBundles_SameBranchAcrossSubpaths(t *testing.T) {
	bundles, err := ParseBundles("foo/bar branch:main path:lib\nfoo/bar branch:main path:plugins/git")
	Expect(t, NoError(err))
	Expect(t, Equals(2, len(bundles)))
}

// Two pins for one repo would load the same plugin twice at different
// revisions, however separate their clone directories are.
func TestParseBundles_ConflictingPin(t *testing.T) {
	other := strings.Repeat("a", 40)
	_, err := ParseBundles("foo/bar pin:" + sha + "\nfoo/bar pin:" + other)
	Expect(t, AnError(err))
	Expect(t, Contains(err.Error(), "line 2"))
	Expect(t, Contains(err.Error(), "conflicting pin"))
}

func TestParseBundles_InconsistentPin(t *testing.T) {
	inputs := []string{
		"foo/bar\nfoo/bar pin:" + sha,
		"foo/bar pin:" + sha + "\nfoo/bar",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			_, err := ParseBundles(input)
			Expect(t, AnError(err))
			Expect(t, Contains(err.Error(), "inconsistent pin"))
		})
	}
}

func TestParseBundles_SamePinAcrossSubpaths(t *testing.T) {
	input := "foo/bar pin:" + sha + " path:lib\nfoo/bar pin:" + sha + " path:plugins/git"
	bundles, err := ParseBundles(input)
	Expect(t, NoError(err))
	Expect(t, Equals(2, len(bundles)))
}

// A preset is what keeps a multi-subpath repo consistent without repeating
// the pin on every line.
func TestParseBundles_PresetSatisfiesPinConsistency(t *testing.T) {
	input := strings.Join([]string{
		`preset:foo/bar pin:` + sha,
		`foo/bar path:lib`,
		`foo/bar path:plugins/git`,
	}, "\n")

	bundles, err := ParseBundles(input)
	Expect(t, NoError(err))
	Expect(t, Equals(2, len(bundles)))
	Expect(t, Equals(sha, bundles[1].Pin))
}

func TestParseBundles_BranchUnderUsingDirective(t *testing.T) {
	input := strings.Join([]string{
		`using:foo/bar path:plugins branch:main`,
		`git`,
		`extract`,
	}, "\n")

	bundles, err := ParseBundles(input)
	Expect(t, NoError(err))
	Expect(t, Equals(2, len(bundles)))
}

func TestParseBundles_LocalBundlesIgnoreBranch(t *testing.T) {
	bundles, err := ParseBundles("~/myplugin branch:main\n~/myplugin path:lib")
	Expect(t, NoError(err))
	Expect(t, Equals(2, len(bundles)))
}
