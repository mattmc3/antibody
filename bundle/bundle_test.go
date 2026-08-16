package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mattmc3/antibody/bundleparse"
	. "github.com/mattmc3/antibody/internal/expect"
	"github.com/mattmc3/antibody/internal/gittest"
)

func TestSuccessfullGitBundles(t *testing.T) {
	pluginSetup := func(r *gittest.Repo) {
		r.WriteFile("myplugin.plugin.zsh", "echo myplugin\n")
		r.Commit("add plugin file")
	}
	table := []struct {
		name   string
		args   string
		setup  func(r *gittest.Repo)
		result string
	}{
		{"zsh", "", pluginSetup, "\nsource "},
		{"path", " kind:path", nil, "export PATH=\""},
		{"path-branch", " kind:path branch:v1", func(r *gittest.Repo) {
			r.Branch("v1")
			r.WriteFile("v1.txt", "v1\n")
			r.Commit("v1 work")
			r.Checkout("main")
		}, "export PATH=\""},
		{"clone", " kind:clone", nil, ""},
		{"fpath", " kind:fpath", nil, "fpath+=( "},
		{"inner-path", " path:completions/_myfunc", func(r *gittest.Repo) {
			r.WriteFile("completions/_myfunc", "#compdef myfunc\n")
			r.Commit("add completion")
		}, "completions/_myfunc"},
		{"defer", " kind:defer", pluginSetup, "zsh-defer source "},
		{"autoload", " kind:autoload path:functions", func(r *gittest.Repo) {
			r.WriteFile("functions/myfunc", "echo myfunc\n")
			r.Commit("add function")
		}, "builtin autoload -Uz "},
	}
	for _, row := range table {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			upstream := gittest.New(t)
			if row.setup != nil {
				row.setup(upstream)
			}
			home := home(t)
			bundle, err := New(home, upstream.URL()+row.args)
			Expect(t, NoError(err))
			result, err := bundle.Get()
			Expect(t, NoError(err))
			Expect(t, Contains(result, row.result))
		})
	}
}

func TestZshInvalidGitBundle(t *testing.T) {
	home := home(t)
	bundle, err := New(home, "file:///this/path/does/not/exist")
	Expect(t, NoError(err))
	_, err = bundle.Get()
	Expect(t, AnError(err))
}

// A failed clone must propagate through every bundle kind and wrapper.
func TestInvalidGitBundleAllKinds(t *testing.T) {
	for _, args := range []string{
		"kind:path",
		"kind:fpath",
		"kind:clone",
		"kind:defer",
		"kind:autoload",
		"autoload:functions",
	} {
		t.Run(args, func(t *testing.T) {
			b, err := New(home(t), "file:///this/path/does/not/exist "+args)
			Expect(t, NoError(err))
			_, err = b.Get()
			Expect(t, AnError(err))
		})
	}
}

func TestZshLocalBundle(t *testing.T) {
	home := home(t)
	// nolint: gosec
	Expect(t, NoError(os.WriteFile(home+"/a.sh", []byte("echo 9"), 0644)))
	bundle, err := New(home, home)
	Expect(t, NoError(err))
	result, err := bundle.Get()
	Expect(t, Contains(result, "a.sh"))
	Expect(t, NoError(err))
}

func TestZshInvalidLocalBundle(t *testing.T) {
	home := home(t)
	bundle, err := New(home, "/asduhasd/asdasda")
	Expect(t, NoError(err))
	_, err = bundle.Get()
	Expect(t, AnError(err))
}

func TestPathInvalidLocalBundle(t *testing.T) {
	home := home(t)
	bundle, err := New(home, "/asduhasd/asdasda kind:path")
	Expect(t, NoError(err))
	_, err = bundle.Get()
	Expect(t, AnError(err))
}

func TestPathLocalBundle(t *testing.T) {
	home := home(t)
	// nolint: gosec
	Expect(t, NoError(os.WriteFile(filepath.Join(home, "whatever.sh"), []byte(""), 0644)))
	bundle, err := New(home, home+" kind:path")
	Expect(t, NoError(err))
	result, err := bundle.Get()
	Expect(t, NoError(err))
	Expect(t, Equals("export PATH=\""+home+":$PATH\"", result))
	Expect(t, NoError(err))
}

func TestDecoratedLocalBundle(t *testing.T) {
	home := home(t)
	// nolint: gosec
	Expect(t, NoError(os.WriteFile(filepath.Join(home, "p.plugin.zsh"), []byte(""), 0644)))

	t.Run("pre", func(t *testing.T) {
		b, err := New(home, home+" pre:my_pre_cmd")
		Expect(t, NoError(err))
		result, err := b.Get()
		Expect(t, NoError(err))
		Expect(t, strings.HasPrefix(result, "my_pre_cmd\n"))
	})

	t.Run("post", func(t *testing.T) {
		b, err := New(home, home+" post:my_post_cmd")
		Expect(t, NoError(err))
		result, err := b.Get()
		Expect(t, NoError(err))
		Expect(t, strings.HasSuffix(result, "\nmy_post_cmd"))
	})

	t.Run("conditional", func(t *testing.T) {
		b, err := New(home, home+" conditional:is_mac")
		Expect(t, NoError(err))
		result, err := b.Get()
		Expect(t, NoError(err))
		Expect(t, strings.HasPrefix(result, "if is_mac; then\n"))
		Expect(t, strings.HasSuffix(result, "\nfi"))
	})
}

// A clone home containing spaces must produce quoted, sourceable output.
func TestPathsWithSpaces(t *testing.T) {
	upstream := gittest.New(t)
	upstream.WriteFile("myplugin.plugin.zsh", "echo myplugin\n")
	upstream.Commit("add plugin file")
	spacedHome := filepath.Join(t.TempDir(), "antibody with spaces")
	Expect(t, NoError(os.MkdirAll(spacedHome, 0755)))

	t.Run("zsh", func(t *testing.T) {
		b, err := New(spacedHome, upstream.URL())
		Expect(t, NoError(err))
		result, err := b.Get()
		Expect(t, NoError(err))
		lines := strings.Split(result, "\n")
		Expect(t, Equals(2, len(lines)))
		Expect(t, Matches(`^fpath\+=\( ".* with spaces.*" \)$`, lines[0]))
		Expect(t, Matches(`^source ".* with spaces.*/myplugin\.plugin\.zsh"$`, lines[1]))
	})

	t.Run("fpath", func(t *testing.T) {
		b, err := New(spacedHome, upstream.URL()+" kind:fpath")
		Expect(t, NoError(err))
		result, err := b.Get()
		Expect(t, NoError(err))
		Expect(t, Matches(`^fpath\+=\( ".* with spaces.*" \)$`, result))
	})

	t.Run("defer", func(t *testing.T) {
		b, err := New(spacedHome, upstream.URL()+" kind:defer")
		Expect(t, NoError(err))
		result, err := b.Get()
		Expect(t, NoError(err))
		Expect(t, Matches(`zsh-defer source ".* with spaces.*/myplugin\.plugin\.zsh"`, result))
	})
}

func TestQuotedAnnotationValues(t *testing.T) {
	home := home(t)
	// nolint: gosec
	Expect(t, NoError(os.WriteFile(filepath.Join(home, "p.plugin.zsh"), []byte(""), 0644)))

	t.Run("double quotes", func(t *testing.T) {
		b, err := New(home, home+` pre:"echo hello world"`)
		Expect(t, NoError(err))
		result, err := b.Get()
		Expect(t, NoError(err))
		Expect(t, strings.HasPrefix(result, "echo hello world\n"), "got: %s", result)
	})

	t.Run("single quotes", func(t *testing.T) {
		b, err := New(home, home+" post:'echo bye bye'")
		Expect(t, NoError(err))
		result, err := b.Get()
		Expect(t, NoError(err))
		Expect(t, strings.HasSuffix(result, "\necho bye bye"), "got: %s", result)
	})
}

func TestFpathRuleAnnotation(t *testing.T) {
	home := home(t)
	b, err := New(home, home+" kind:fpath fpath-rule:prepend")
	Expect(t, NoError(err))
	// nolint: gosec
	Expect(t, NoError(os.WriteFile(filepath.Join(home, "_func"), []byte(""), 0644)))
	result, err := b.Get()
	Expect(t, NoError(err))
	Expect(t, strings.HasPrefix(result, "fpath=( "), "got: %s", result)
}

func TestAutoloadLocalBundle(t *testing.T) {
	home := home(t)
	// nolint: gosec
	Expect(t, NoError(os.WriteFile(filepath.Join(home, "_myfunc"), []byte(""), 0644)))
	bundle, err := New(home, home+" kind:autoload")
	Expect(t, NoError(err))
	result, err := bundle.Get()
	Expect(t, NoError(err))
	Expect(t, Contains(result, "fpath+=( "))
	Expect(t, Contains(result, "builtin autoload -Uz "+quote(displayPath(home))+"/*(N.:t)"))
}

func TestAutoloadLocalBundlePrepend(t *testing.T) {
	home := home(t)
	// nolint: gosec
	Expect(t, NoError(os.WriteFile(filepath.Join(home, "_myfunc"), []byte(""), 0644)))
	bundle, err := New(home, home+" kind:autoload fpath-rule:prepend")
	Expect(t, NoError(err))
	result, err := bundle.Get()
	Expect(t, NoError(err))
	Expect(t, Contains(result, "fpath=( "))
	Expect(t, Contains(result, "builtin autoload -Uz "+quote(displayPath(home))+"/*(N.:t)"))
}

func TestAutoloadAnnotationLocalBundle(t *testing.T) {
	home := home(t)
	Expect(t, NoError(os.MkdirAll(filepath.Join(home, "functions"), 0755)))
	// nolint: gosec
	Expect(t, NoError(os.WriteFile(filepath.Join(home, "myplugin.plugin.zsh"), []byte(""), 0644)))
	bundle, err := New(home, home+" autoload:functions")
	Expect(t, NoError(err))
	result, err := bundle.Get()
	Expect(t, NoError(err))
	dir := displayPath(filepath.Join(home, "functions"))
	Expect(t, Contains(result, "builtin autoload -Uz "+quote(dir)+"/*(N.:t)"))
	Expect(t, Contains(result, "source "))
}

func TestDeferLocalBundle(t *testing.T) {
	home := home(t)
	// nolint: gosec
	Expect(t, NoError(os.WriteFile(filepath.Join(home, "myplugin.plugin.zsh"), []byte(""), 0644)))
	bundle, err := New(home, home+" kind:defer")
	Expect(t, NoError(err))
	result, err := bundle.Get()
	Expect(t, NoError(err))
	Expect(t, Contains(result, "zsh-defer source "))
	for line := range strings.SplitSeq(result, "\n") {
		if strings.HasPrefix(line, "fpath") {
			Expect(t, !strings.HasPrefix(line, "zsh-defer"), "fpath line should not be deferred: %q", line)
		}
	}
}

func TestFpathBeforeSource(t *testing.T) {
	home := home(t)
	// nolint: gosec
	Expect(t, NoError(os.WriteFile(filepath.Join(home, "p.plugin.zsh"), []byte(""), 0644)))
	b, err := New(home, home)
	Expect(t, NoError(err))
	result, err := b.Get()
	Expect(t, NoError(err))
	lines := strings.Split(result, "\n")
	Expect(t, strings.HasPrefix(lines[0], "fpath+=( "), "want fpath line first, got: %s", result)
	Expect(t, strings.HasPrefix(lines[1], "source "), "want source line second, got: %s", result)
}

func TestEnvVarLocalBundle(t *testing.T) {
	t.Run("dir", func(t *testing.T) {
		// var need not be set: emitted literally, init file assumed
		b, err := New(home(t), "$MYPLUGS/myplug")
		Expect(t, NoError(err))
		result, err := b.Get()
		Expect(t, NoError(err))
		Expect(t, Equals(`fpath+=( "$MYPLUGS/myplug" )`+"\n"+`source "$MYPLUGS/myplug/myplug.plugin.zsh"`, result))
	})

	t.Run("path", func(t *testing.T) {
		b, err := New(home(t), "$MYPLUGS/myplug kind:path")
		Expect(t, NoError(err))
		result, err := b.Get()
		Expect(t, NoError(err))
		Expect(t, Equals(`export PATH="$MYPLUGS/myplug:$PATH"`, result))
	})

	t.Run("file", func(t *testing.T) {
		plugins := t.TempDir()
		t.Setenv("MYPLUGS", plugins)
		// nolint: gosec
		Expect(t, NoError(os.WriteFile(filepath.Join(plugins, "single.zsh"), []byte(""), 0644)))
		b, err := New(home(t), "$MYPLUGS/single.zsh")
		Expect(t, NoError(err))
		result, err := b.Get()
		Expect(t, NoError(err))
		Expect(t, Equals(`source "$MYPLUGS/single.zsh"`, result))
	})

	t.Run("no expansion or globbing", func(t *testing.T) {
		plugins := t.TempDir()
		t.Setenv("MYPLUGS", plugins)
		dir := filepath.Join(plugins, "oddplug")
		Expect(t, NoError(os.MkdirAll(dir, 0755)))
		// nolint: gosec
		Expect(t, NoError(os.WriteFile(filepath.Join(dir, "other.plugin.zsh"), []byte(""), 0644)))
		b, err := New(home(t), "$MYPLUGS/oddplug")
		Expect(t, NoError(err))
		result, err := b.Get()
		Expect(t, NoError(err))
		Expect(t, Equals(`fpath+=( "$MYPLUGS/oddplug" )`+"\n"+`source "$MYPLUGS/oddplug/oddplug.plugin.zsh"`, result))
	})
}

func TestRelativeLocalBundle(t *testing.T) {
	base := t.TempDir()
	Expect(t, NoError(os.MkdirAll(filepath.Join(base, "myplug"), 0755)))
	// nolint: gosec
	Expect(t, NoError(os.WriteFile(filepath.Join(base, "myplug", "myplug.plugin.zsh"), []byte(""), 0644)))
	t.Chdir(base)

	b, err := New(home(t), "./myplug")
	Expect(t, NoError(err))
	result, err := b.Get()
	Expect(t, NoError(err))
	Expect(t, Contains(result, `fpath+=( "./myplug" )`))
	Expect(t, Contains(result, `source "./myplug/myplug.plugin.zsh"`))
}

func TestHomeSubstitution(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)
	plugin := filepath.Join(fakeHome, "myplugin")
	Expect(t, NoError(os.MkdirAll(plugin, 0755)))
	// nolint: gosec
	Expect(t, NoError(os.WriteFile(filepath.Join(plugin, "myplugin.plugin.zsh"), []byte(""), 0644)))

	t.Run("zsh", func(t *testing.T) {
		b, err := New(fakeHome, plugin)
		Expect(t, NoError(err))
		result, err := b.Get()
		Expect(t, NoError(err))
		Expect(t, Contains(result, `fpath+=( "$HOME/myplugin" )`))
		Expect(t, Contains(result, `source "$HOME/myplugin/myplugin.plugin.zsh"`))
	})

	t.Run("path", func(t *testing.T) {
		b, err := New(fakeHome, plugin+" kind:path")
		Expect(t, NoError(err))
		result, err := b.Get()
		Expect(t, NoError(err))
		Expect(t, Equals(`export PATH="$HOME/myplugin:$PATH"`, result))
	})
}

func TestInitFileDirNamePriority(t *testing.T) {
	home := home(t)
	dir := filepath.Join(home, "myplug")
	Expect(t, NoError(os.MkdirAll(dir, 0755)))
	// nolint: gosec
	Expect(t, NoError(os.WriteFile(filepath.Join(dir, "myplug.plugin.zsh"), []byte(""), 0644)))
	// nolint: gosec
	Expect(t, NoError(os.WriteFile(filepath.Join(dir, "other.plugin.zsh"), []byte(""), 0644)))
	b, err := New(home, dir)
	Expect(t, NoError(err))
	result, err := b.Get()
	Expect(t, NoError(err))
	Expect(t, Contains(result, `source "`+filepath.Join(dir, "myplug.plugin.zsh")+`"`))
	Expect(t, Not(Contains(result, "other.plugin.zsh")))
}

func TestInitFileAssumeDefault(t *testing.T) {
	home := home(t)
	dir := filepath.Join(home, "myplug")
	Expect(t, NoError(os.MkdirAll(dir, 0755)))
	// nolint: gosec
	Expect(t, NoError(os.WriteFile(filepath.Join(dir, "README.md"), []byte(""), 0644)))
	b, err := New(home, dir)
	Expect(t, NoError(err))
	result, err := b.Get()
	Expect(t, NoError(err))
	Expect(t, Contains(result, `fpath+=( "`+dir+`" )`))
	Expect(t, Contains(result, `source "`+filepath.Join(dir, "myplug.plugin.zsh")+`"`))
}

func TestDeferredPost(t *testing.T) {
	home := home(t)
	// nolint: gosec
	Expect(t, NoError(os.WriteFile(filepath.Join(home, "p.plugin.zsh"), []byte(""), 0644)))
	b, err := New(home, home+" kind:defer post:my_post_cmd")
	Expect(t, NoError(err))
	result, err := b.Get()
	Expect(t, NoError(err))
	Expect(t, strings.HasSuffix(result, "\nzsh-defer my_post_cmd"), "post should be deferred, got: %s", result)
}

// Non-bundle lines must error, not build a bundle around a garbage
// project or return a nil that panics at Get.
func TestNewRejectsNonBundleLines(t *testing.T) {
	for _, line := range []string{"", "   ", "# comment"} {
		t.Run("line "+line, func(t *testing.T) {
			b, err := New(home(t), line)
			Expect(t, AnError(err))
			Expect(t, b == nil)
		})
	}
}

func TestNewFromParsedRejectsEmptyName(t *testing.T) {
	b, err := NewFromParsed(home(t), bundleparse.Bundle{})
	Expect(t, AnError(err))
	Expect(t, b == nil)
}

// A pinned clone folder carries a /tree/<sha> suffix; the sha must not
// leak into the init file name the zsh kind gives priority to.
func TestPinnedZshBundleInitFilePriority(t *testing.T) {
	upstream := gittest.New(t)
	name := filepath.Base(upstream.Dir)
	upstream.WriteFile(name+".plugin.zsh", "echo main\n")
	upstream.WriteFile("zzz.plugin.zsh", "echo decoy\n")
	sha := upstream.Commit("add plugins")

	b, err := New(home(t), fmt.Sprintf("%s pin:%s", upstream.URL(), sha))
	Expect(t, NoError(err))
	result, err := b.Get()
	Expect(t, NoError(err))
	Expect(t, Contains(result, name+".plugin.zsh"))
	Expect(t, Not(Contains(result, "zzz.plugin.zsh")))
}

func TestPinRequiresFullSHA(t *testing.T) {
	home := home(t)
	_, err := New(home, "owner/repo pin:abc123")
	Expect(t, AnError(err))
	Expect(t, Contains(err.Error(), "40-character"))
}

func home(t *testing.T) string {
	return t.TempDir()
}
