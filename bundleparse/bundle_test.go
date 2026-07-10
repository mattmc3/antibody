package bundleparse

import (
	"strings"
	"testing"

	. "github.com/mattmc3/antibody/internal/expect"
)

func TestParseBundleLine(t *testing.T) {
	bundle, err := ParseBundleLine(`foo/bar kind:zsh branch:main path:plugins/git`)
	Expect(t, NoError(err))

	want := Bundle{
		Name:             "foo/bar",
		Kind:             "zsh",
		Branch:           "main",
		Path:             "plugins/git",
		ExtraAnnotations: map[string]string{},
	}

	Expect(t, DeepEquals(want, bundle))
}

func TestParseBundleLine_AdditionalAnnotations(t *testing.T) {
	bundle, err := ParseBundleLine(`foo/bar kind:zsh pin:v1 branch:main conditional:if-true autoload:yes pre:"echo hi" post:'echo bye' fpath-rule:prepend unknown:yes`)
	Expect(t, NoError(err))

	want := Bundle{
		Name:        "foo/bar",
		Kind:        "zsh",
		Branch:      "main",
		Pin:         "v1",
		Conditional: "if-true",
		Autoload:    "yes",
		Pre:         "echo hi",
		Post:        "echo bye",
		FpathRule:   "prepend",
		ExtraAnnotations: map[string]string{
			"unknown": "yes",
		},
	}

	Expect(t, DeepEquals(want, bundle))
}

func TestParseBundles(t *testing.T) {
	input := strings.Join([]string{
		`# comment line`,
		`foo/bar kind:zsh branch:main path:plugins/git`,
		``,
		`ohmyzsh/ohmyzsh kind:zsh path:plugins/git`,
	}, "\n")

	bundles, err := ParseBundles(input)
	Expect(t, NoError(err))

	want := []Bundle{
		{
			Name:             "foo/bar",
			Kind:             "zsh",
			Branch:           "main",
			Path:             "plugins/git",
			Line:             2,
			ExtraAnnotations: map[string]string{},
		},
		{
			Name:             "ohmyzsh/ohmyzsh",
			Kind:             "zsh",
			Path:             "plugins/git",
			Line:             4,
			ExtraAnnotations: map[string]string{},
		},
	}

	Expect(t, DeepEquals(want, bundles))
}

func TestParseBundles_UsingDirectiveAppliesToBareBundle(t *testing.T) {
	input := strings.Join([]string{
		`using:ohmyzsh/ohmyzsh kind:zsh`,
		`foo`,
	}, "\n")

	bundles, err := ParseBundles(input)
	Expect(t, NoError(err))

	want := []Bundle{
		{
			Name:             "ohmyzsh/ohmyzsh",
			Kind:             "zsh",
			Path:             "ohmyzsh/ohmyzsh/foo",
			Line:             2,
			ExtraAnnotations: map[string]string{},
		},
	}

	Expect(t, DeepEquals(want, bundles))
}

func TestParseBundles_UsingDirectivePreservesExplicitPath(t *testing.T) {
	input := strings.Join([]string{
		`using:ohmyzsh/ohmyzsh kind:zsh`,
		`foo path:plugins/foo`,
	}, "\n")

	bundles, err := ParseBundles(input)
	Expect(t, NoError(err))

	want := []Bundle{
		{
			Name:             "ohmyzsh/ohmyzsh",
			Kind:             "zsh",
			Path:             "plugins/foo",
			Line:             2,
			ExtraAnnotations: map[string]string{},
		},
	}

	Expect(t, DeepEquals(want, bundles))
}

func TestParseBundles_UsingDirectiveAppliesToMultipleBareBundles(t *testing.T) {
	input := strings.Join([]string{
		`using:ohmyzsh/ohmyzsh path:plugins pin:abcdef`,
		`git`,
		`extract`,
	}, "\n")

	bundles, err := ParseBundles(input)
	Expect(t, NoError(err))

	want := []Bundle{
		{
			Name:             "ohmyzsh/ohmyzsh",
			Kind:             "zsh",
			Path:             "plugins/git",
			Pin:              "abcdef",
			Line:             2,
			ExtraAnnotations: map[string]string{},
		},
		{
			Name:             "ohmyzsh/ohmyzsh",
			Kind:             "zsh",
			Path:             "plugins/extract",
			Pin:              "abcdef",
			Line:             3,
			ExtraAnnotations: map[string]string{},
		},
	}

	Expect(t, DeepEquals(want, bundles))
}

func TestParseBundleLine_DefaultsKindToZsh(t *testing.T) {
	bundle, err := ParseBundleLine(`foo`)
	Expect(t, NoError(err))
	Expect(t, Equals(KindZsh, bundle.Kind))
}

func TestParseBundleLine_InvalidAnnotation(t *testing.T) {
	bundle, err := ParseBundleLine(`foo kind:zsh unknown:yes`)
	Expect(t, NoError(err))
	Expect(t, Equals("yes", bundle.ExtraAnnotations["unknown"]))
}

func TestParseBundleLine_InvalidKind(t *testing.T) {
	_, err := ParseBundleLine(`foo kind:invalid`)
	Expect(t, AnError(err), "invalid kind should error")
}

func TestParseBundleLine_AllowedKindValues(t *testing.T) {
	kinds := []string{KindZsh, KindPath, KindFpath, KindDefer, KindClone, KindAutoload}
	for _, kind := range kinds {
		t.Run(kind, func(t *testing.T) {
			bundle, err := ParseBundleLine(`foo kind:` + kind)
			Expect(t, NoError(err), "kind %q", kind)
			Expect(t, Equals(kind, bundle.Kind))
		})
	}
}

func TestParseBundleLine_InvalidFpathRule(t *testing.T) {
	_, err := ParseBundleLine(`foo fpath-rule:invalid`)
	Expect(t, AnError(err), "invalid fpath-rule should error")
}

func TestParseBundleLine_AllowedFpathRules(t *testing.T) {
	rules := []string{FpathRuleAppend, FpathRulePrepend}
	for _, rule := range rules {
		t.Run(rule, func(t *testing.T) {
			bundle, err := ParseBundleLine(`foo fpath-rule:` + rule)
			Expect(t, NoError(err), "fpath-rule %q", rule)
			Expect(t, Equals(rule, bundle.FpathRule))
		})
	}
}

func TestParseBundles_InvalidLine(t *testing.T) {
	_, err := ParseBundles("foo kind\nfoo kind:zsh")
	Expect(t, AnError(err), "malformed line should error")
	Expect(t, Contains(err.Error(), "line 1"))
}

func TestParseBundles_LaterLineError(t *testing.T) {
	_, err := ParseBundles("foo/bar kind:zsh\n# comment\nfoo kind:invalid")
	Expect(t, AnError(err), "invalid kind on later line should error")
	Expect(t, Contains(err.Error(), "line 3"))
}

func TestParseBundles_EmptyAndCommentOnly(t *testing.T) {
	bundles, err := ParseBundles("# only comment\n   \n")
	Expect(t, NoError(err))
	Expect(t, Equals(0, len(bundles)))
}
