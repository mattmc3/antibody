package bundleparse

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mattmc3/antibody/internal/config"
)

// Presets holds fallback annotations recorded by preset: directives, keyed by
// the identity of the bundle they apply to.
type Presets map[string]map[string]string

// set records the annotations for a bundle, replacing any earlier preset.
func (p Presets) set(name string, annotations map[string]string) {
	p[presetKey(name)] = annotations
}

// apply fills in annotations the line does not already carry.
func (p Presets) apply(parsed ParsedLine) {
	for key, value := range p[presetKey(parsed.Name)] {
		if _, ok := parsed.Annotations[key]; !ok {
			parsed.Annotations[key] = value
		}
	}
}

// presetKey identifies the bundle a preset applies to. Every spelling of one
// repo shares a key; local bundles key on their resolved path.
func presetKey(name string) string {
	name = strings.TrimRight(name, "/")
	if isLocalName(name) {
		return localPresetKey(name)
	}
	return repoPresetKey(strings.TrimSuffix(name, ".git"))
}

func isLocalName(name string) bool {
	return strings.HasPrefix(name, "/") ||
		strings.HasPrefix(name, "~") ||
		strings.HasPrefix(name, "$") ||
		strings.HasPrefix(name, ".")
}

// localPresetKey resolves the spellings of a local path that name one place,
// so ~/foo, $HOME/foo, and /home/you/foo share a key.
func localPresetKey(name string) string {
	name = os.ExpandEnv(name)
	if name == "~" || strings.HasPrefix(name, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return name
		}
		name = filepath.Join(home, strings.TrimPrefix(name, "~"))
	}
	return filepath.Clean(name)
}

func repoPresetKey(name string) string {
	if i := strings.Index(name, "://"); i >= 0 {
		// a file: clone lives under the bundle home, not at the path it names
		if scheme := name[:i]; scheme == "file" {
			return name
		}
		return stripUserInfo(name[i+3:])
	}
	if i := strings.Index(name, "@"); i >= 0 {
		host := name[i+1:]
		return strings.Replace(host, ":", "/", 1)
	}
	return config.Get().GitDomain() + "/" + name
}

func stripUserInfo(host string) string {
	if i := strings.Index(host, "@"); i >= 0 {
		return host[i+1:]
	}
	return host
}
