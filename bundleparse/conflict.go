package bundleparse

import "fmt"

// clones tracks the revision each repo is first asked for, so every entry of
// one repo is held to one answer. Subpaths of a repo share its clone, and a
// shell loading one plugin at two revisions is incoherent however the clones
// are laid out on disk.
type clones map[string]string

// check reports whether a bundle asks for a different pin or branch than an
// earlier entry of the same repo.
func (c clones) check(bundle Bundle) error {
	if isLocalName(bundle.Name) {
		return nil
	}
	repo := presetKey(bundle.Name)
	for _, field := range []struct {
		key   string
		value string
	}{
		{KeyPin, bundle.Pin},
		{KeyBranch, bundle.Branch},
	} {
		seen, ok := c[repo+"\x00"+field.key]
		c[repo+"\x00"+field.key] = field.value
		switch {
		case !ok, seen == field.value:
			continue
		case seen == "" || field.value == "":
			return fmt.Errorf("inconsistent %s for %q: some entries have %s:%s, others do not",
				field.key, bundle.Name, field.key, seen+field.value)
		default:
			return fmt.Errorf("conflicting %s for %q: %s:%s vs %s:%s",
				field.key, bundle.Name, field.key, field.value, field.key, seen)
		}
	}
	return nil
}
