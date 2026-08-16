package bundleparse

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/mattmc3/antibody/internal/expect"
)

const sha = "0123456789abcdef0123456789abcdef01234567"

func TestParseBundles_PresetEmitsNoEntryAndSuppliesFallback(t *testing.T) {
	input := strings.Join([]string{
		`preset:foo/bar pin:` + sha,
		`foo/bar path:plugins/git`,
	}, "\n")

	bundles, err := ParseBundles(input)
	Expect(t, NoError(err))

	want := []Bundle{
		{
			Name:             "foo/bar",
			Kind:             "zsh",
			Path:             "plugins/git",
			Pin:              sha,
			Line:             2,
			ExtraAnnotations: map[string]string{},
		},
	}

	Expect(t, DeepEquals(want, bundles))
}

func TestParseBundles_PresetLosesToLineAnnotation(t *testing.T) {
	input := strings.Join([]string{
		`preset:foo/bar branch:main kind:fpath`,
		`foo/bar branch:next`,
	}, "\n")

	bundles, err := ParseBundles(input)
	Expect(t, NoError(err))
	Expect(t, Equals(1, len(bundles)))
	Expect(t, Equals("next", bundles[0].Branch))
	Expect(t, Equals(KindFpath, bundles[0].Kind))
}

func TestParseBundles_PresetLosesToUsingDirective(t *testing.T) {
	input := strings.Join([]string{
		`preset:foo/bar branch:preset`,
		`using:foo/bar branch:using`,
		`baz`,
	}, "\n")

	bundles, err := ParseBundles(input)
	Expect(t, NoError(err))
	Expect(t, Equals(1, len(bundles)))
	Expect(t, Equals("using", bundles[0].Branch))
}

func TestParseBundles_PresetReachesBareWordsUnderUsing(t *testing.T) {
	input := strings.Join([]string{
		`preset:foo/bar pin:` + sha,
		`using:foo/bar path:plugins`,
		`git`,
	}, "\n")

	bundles, err := ParseBundles(input)
	Expect(t, NoError(err))
	Expect(t, Equals(1, len(bundles)))
	Expect(t, Equals(sha, bundles[0].Pin))
	Expect(t, Equals("plugins/git", bundles[0].Path))
}

func TestParseBundles_PresetAppliesOnlyToLaterEntries(t *testing.T) {
	input := strings.Join([]string{
		`foo/bar`,
		`preset:foo/bar kind:fpath`,
		`foo/bar path:lib`,
	}, "\n")

	bundles, err := ParseBundles(input)
	Expect(t, NoError(err))
	Expect(t, Equals(2, len(bundles)))
	Expect(t, Equals(KindZsh, bundles[0].Kind))
	Expect(t, Equals(KindFpath, bundles[1].Kind))
}

func TestParseBundles_PresetReplacesEarlierPreset(t *testing.T) {
	input := strings.Join([]string{
		`preset:foo/bar branch:next kind:fpath`,
		`preset:foo/bar branch:main`,
		`foo/bar`,
	}, "\n")

	bundles, err := ParseBundles(input)
	Expect(t, NoError(err))
	Expect(t, Equals(1, len(bundles)))
	Expect(t, Equals("main", bundles[0].Branch))
	Expect(t, Equals(KindZsh, bundles[0].Kind))
}

func TestParseBundles_PresetsForDifferentBundlesCoexist(t *testing.T) {
	input := strings.Join([]string{
		`preset:foo/bar branch:foo`,
		`preset:baz/qux branch:baz`,
		`foo/bar`,
		`baz/qux`,
	}, "\n")

	bundles, err := ParseBundles(input)
	Expect(t, NoError(err))
	Expect(t, Equals(2, len(bundles)))
	Expect(t, Equals("foo", bundles[0].Branch))
	Expect(t, Equals("baz", bundles[1].Branch))
}

func TestParseBundles_PresetSharedAcrossRepoSpellings(t *testing.T) {
	spellings := []string{
		"foo/bar",
		"https://github.com/foo/bar",
		"https://github.com/foo/bar.git",
		"git@github.com:foo/bar",
		"ssh://git@github.com/foo/bar",
	}
	for _, spelling := range spellings {
		t.Run(spelling, func(t *testing.T) {
			bundles, err := ParseBundles("preset:foo/bar branch:main\n" + spelling)
			Expect(t, NoError(err))
			Expect(t, Equals(1, len(bundles)))
			Expect(t, Equals("main", bundles[0].Branch), "spelling %q", spelling)
		})
	}
}

func TestParseBundles_PresetLocalPath(t *testing.T) {
	home, err := os.UserHomeDir()
	Expect(t, NoError(err))
	dir := filepath.Join(home, "myplugin")

	bundles, err := ParseBundles("preset:~/myplugin kind:path\n" + dir)
	Expect(t, NoError(err))
	Expect(t, Equals(1, len(bundles)))
	Expect(t, Equals(KindPath, bundles[0].Kind))
}

func TestParseBundles_PresetLocalPathSpellings(t *testing.T) {
	home, err := os.UserHomeDir()
	Expect(t, NoError(err))
	spellings := []string{"~/myplugin", "$HOME/myplugin", filepath.Join(home, "myplugin")}

	for _, preset := range spellings {
		for _, bundle := range spellings {
			t.Run(preset+" -> "+bundle, func(t *testing.T) {
				bundles, err := ParseBundles("preset:" + preset + " kind:path\n" + bundle)
				Expect(t, NoError(err))
				Expect(t, Equals(1, len(bundles)))
				Expect(t, Equals(KindPath, bundles[0].Kind))
			})
		}
	}
}

func TestParseBundles_PresetDoesNotReachOtherBundles(t *testing.T) {
	bundles, err := ParseBundles("preset:foo/bar branch:main\nbaz/qux")
	Expect(t, NoError(err))
	Expect(t, Equals(1, len(bundles)))
	Expect(t, Equals("", bundles[0].Branch))
}

func TestParseBundles_PresetInvalidAnnotationErrors(t *testing.T) {
	_, err := ParseBundles("preset:foo/bar kind:nope\nfoo/bar")
	Expect(t, AnError(err), "invalid preset annotation should error")
}

func TestParseBundles_PresetInvalidPinErrors(t *testing.T) {
	_, err := ParseBundles("preset:foo/bar pin:abc123\nfoo/bar")
	Expect(t, AnError(err), "short preset pin should error")
	Expect(t, Contains(err.Error(), "line 1"))
}

func TestParseBundles_PresetLocalPinNotValidated(t *testing.T) {
	_, err := ParseBundles("preset:~/myplugin pin:abc123\n~/myplugin")
	Expect(t, NoError(err))
}

// a file:// bundle clones into ANTIBODY_HOME, so it is not the local path
func TestParseBundles_PresetLocalPathDoesNotReachFileURL(t *testing.T) {
	bundles, err := ParseBundles("preset:/tmp/myplugin kind:path\nfile:///tmp/myplugin")
	Expect(t, NoError(err))
	Expect(t, Equals(1, len(bundles)))
	Expect(t, Equals(KindZsh, bundles[0].Kind))
}

func TestParseBundles_PresetFileURLDoesNotReachLocalPath(t *testing.T) {
	bundles, err := ParseBundles("preset:file:///tmp/myplugin kind:path\n/tmp/myplugin")
	Expect(t, NoError(err))
	Expect(t, Equals(1, len(bundles)))
	Expect(t, Equals(KindZsh, bundles[0].Kind))
}

// bare words under a path-style using: resolve to their own subdirectories
func TestParseBundles_PresetSkipsBareWordsUnderPathStyleUsing(t *testing.T) {
	input := strings.Join([]string{
		`preset:~/myplugins branch:next`,
		`using:~/myplugins`,
		`git`,
	}, "\n")

	bundles, err := ParseBundles(input)
	Expect(t, NoError(err))
	Expect(t, Equals(1, len(bundles)))
	Expect(t, Equals("", bundles[0].Branch))
}

func TestParseBundlesWith_CarriesPresetsInAndOut(t *testing.T) {
	_, presets, err := ParseBundlesWith("preset:foo/bar branch:main", nil)
	Expect(t, NoError(err))

	bundles, presets, err := ParseBundlesWith("foo/bar", presets)
	Expect(t, NoError(err))
	Expect(t, Equals(1, len(bundles)))
	Expect(t, Equals("main", bundles[0].Branch))

	bundles, _, err = ParseBundlesWith("foo/bar path:lib", presets)
	Expect(t, NoError(err))
	Expect(t, Equals("main", bundles[0].Branch))
}

func TestParseBundles_PresetLocalPathKeepsGitSuffix(t *testing.T) {
	home, err := os.UserHomeDir()
	Expect(t, NoError(err))

	bundles, err := ParseBundles("preset:~/myplugin kind:path\n" + filepath.Join(home, "myplugin.git"))
	Expect(t, NoError(err))
	Expect(t, Equals(1, len(bundles)))
	Expect(t, Equals(KindZsh, bundles[0].Kind))
}
