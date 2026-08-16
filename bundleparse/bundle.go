package bundleparse

import (
	"fmt"
	"maps"
	"regexp"
	"strings"
)

type Bundle struct {
	Name             string
	Kind             string
	Branch           string
	Path             string
	Pin              string
	Conditional      string
	Autoload         string
	Pre              string
	Post             string
	FpathRule        string
	Line             int
	ExtraAnnotations map[string]string
}

const (
	KeyKind        = "kind"
	KeyBranch      = "branch"
	KeyPath        = "path"
	KeyPin         = "pin"
	KeyConditional = "conditional"
	KeyAutoload    = "autoload"
	KeyPre         = "pre"
	KeyPost        = "post"
	KeyFpathRule   = "fpath-rule"

	KindZsh      = "zsh"
	KindPath     = "path"
	KindFpath    = "fpath"
	KindDefer    = "defer"
	KindClone    = "clone"
	KindAutoload = "autoload"

	FpathRuleAppend  = "append"
	FpathRulePrepend = "prepend"
)

var validKindValues = map[string]struct{}{
	KindZsh:      {},
	KindPath:     {},
	KindFpath:    {},
	KindDefer:    {},
	KindClone:    {},
	KindAutoload: {},
}

var validFpathRules = map[string]struct{}{
	FpathRuleAppend:  {},
	FpathRulePrepend: {},
}

type ParseError struct {
	Line int
	Err  error
}

func (e ParseError) Error() string {
	return fmt.Sprintf("line %d: %v", e.Line, e.Err)
}

func (e ParseError) Unwrap() error {
	return e.Err
}

// ParseBundles parses a multi-line bundle specification.
// Each non-empty line is passed through ParseLine, then validated and
// converted into a Bundle struct.
func ParseBundles(input string) ([]Bundle, error) {
	bundles, _, err := ParseBundlesWith(input, nil)
	return bundles, err
}

// ParseBundlesWith parses input starting from the given presets and returns
// the presets in effect afterwards, so a caller bundling a plugins file one
// call at a time can carry them across calls.
func ParseBundlesWith(input string, presets Presets) ([]Bundle, Presets, error) {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	lines := strings.Split(input, "\n")
	bundles := make([]Bundle, 0, len(lines))
	var usingDirective *ParsedLine

	carried := Presets{}
	maps.Copy(carried, presets)

	for idx, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parsed, err := ParseLine(line)
		if err != nil {
			return nil, nil, ParseError{Line: idx + 1, Err: err}
		}

		switch parsed.Directive {
		case UsingDirective:
			usingDirective = &parsed
			continue
		case PresetDirective:
			if err := validateAnnotations(parsed.Name, parsed.Annotations); err != nil {
				return nil, nil, ParseError{Line: idx + 1, Err: err}
			}
			carried.set(parsed.Name, parsed.Annotations)
			continue
		}

		if parsed.Directive != BundleDirective || parsed.Name == "" {
			continue
		}

		// a bare word under a path-style using: resolves to its own
		// subdirectory, so the target's preset is not its preset
		presetable := true
		if usingDirective != nil && !strings.Contains(parsed.Name, "/") {
			presetable = !isLocalName(usingDirective.Name)
			parsed = applyUsingDirective(parsed, *usingDirective)
		}
		if presetable {
			carried.apply(parsed)
		}

		bundle, err := bundleFromParsed(parsed)
		if err != nil {
			return nil, nil, ParseError{Line: idx + 1, Err: err}
		}
		bundle.Line = idx + 1
		bundles = append(bundles, bundle)
	}

	return bundles, carried, nil
}

// ParseBundleLine parses a single bundle definition line.
func ParseBundleLine(line string) (Bundle, error) {
	parsed, err := ParseLine(line)
	if err != nil {
		return Bundle{}, err
	}
	if parsed.Directive != BundleDirective {
		return Bundle{}, nil
	}
	if parsed.Name == "" {
		return Bundle{}, nil
	}

	bundle, err := bundleFromParsed(parsed)
	if err != nil {
		return Bundle{}, err
	}
	return bundle, nil
}

func applyUsingDirective(parsed ParsedLine, using ParsedLine) ParsedLine {
	originalName := parsed.Name
	parsed.Name = using.Name

	if _, ok := parsed.Annotations[KeyPath]; !ok {
		pathBase := using.Name
		if usingPath, ok2 := using.Annotations[KeyPath]; ok2 && usingPath != "" {
			pathBase = usingPath
		}
		if pathBase != "" {
			pathBase = strings.TrimSuffix(pathBase, "/")
			parsed.Annotations[KeyPath] = pathBase + "/" + originalName
		}
	}

	for key, value := range using.Annotations {
		if key == KeyPath {
			continue
		}
		if _, ok := parsed.Annotations[key]; ok {
			continue
		}
		parsed.Annotations[key] = value
	}

	return parsed
}

func bundleFromParsed(parsed ParsedLine) (Bundle, error) {
	if parsed.Directive != BundleDirective {
		return Bundle{}, fmt.Errorf("parsed line is not a bundle")
	}
	if strings.TrimSpace(parsed.Name) == "" {
		return Bundle{}, fmt.Errorf("missing bundle name")
	}

	bundle := Bundle{Name: parsed.Name}
	var extras map[string]string

	if value, ok := parsed.Annotations[KeyKind]; ok {
		if err := validateKind(value); err != nil {
			return Bundle{}, err
		}
		bundle.Kind = value
	}
	if value, ok := parsed.Annotations[KeyBranch]; ok {
		bundle.Branch = value
	}
	if value, ok := parsed.Annotations[KeyPath]; ok {
		bundle.Path = value
	}
	if value, ok := parsed.Annotations[KeyPin]; ok {
		bundle.Pin = value
	}
	if value, ok := parsed.Annotations[KeyConditional]; ok {
		bundle.Conditional = value
	}
	if value, ok := parsed.Annotations[KeyAutoload]; ok {
		bundle.Autoload = value
	}
	if value, ok := parsed.Annotations[KeyPre]; ok {
		bundle.Pre = value
	}
	if value, ok := parsed.Annotations[KeyPost]; ok {
		bundle.Post = value
	}
	if value, ok := parsed.Annotations[KeyFpathRule]; ok {
		if err := validateFpathRule(value); err != nil {
			return Bundle{}, err
		}
		bundle.FpathRule = value
	}

	for key, value := range parsed.Annotations {
		switch key {
		case KeyKind, KeyBranch, KeyPath, KeyPin, KeyConditional, KeyAutoload, KeyPre, KeyPost, KeyFpathRule:
			continue
		default:
			if extras == nil {
				extras = make(map[string]string, len(parsed.Annotations))
			}
			extras[key] = value
		}
	}

	if extras == nil {
		extras = map[string]string{}
	}
	bundle.ExtraAnnotations = extras

	if bundle.Kind == "" {
		bundle.Kind = KindZsh
	}

	return bundle, nil
}

// validateAnnotations checks the values a directive line carries, which is
// otherwise only done when a bundle line is turned into a Bundle.
func validateAnnotations(name string, annotations map[string]string) error {
	if value, ok := annotations[KeyKind]; ok {
		if err := validateKind(value); err != nil {
			return err
		}
	}
	if value, ok := annotations[KeyPin]; ok {
		if err := ValidatePin(name, value); err != nil {
			return err
		}
	}
	if value, ok := annotations[KeyFpathRule]; ok {
		return validateFpathRule(value)
	}
	return nil
}

var pinSHAPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

// ValidatePin requires full 40-character commit SHAs for repo bundles.
// Local bundles ignore pins.
func ValidatePin(name, pin string) error {
	if pin == "" || isLocalName(name) {
		return nil
	}
	if !pinSHAPattern.MatchString(pin) {
		return fmt.Errorf("pin requires a full 40-character commit SHA, got %q", pin)
	}
	return nil
}

func validateKind(value string) error {
	if _, ok := validKindValues[value]; !ok {
		return fmt.Errorf("invalid kind %q", value)
	}
	return nil
}

func validateFpathRule(value string) error {
	if _, ok := validFpathRules[value]; !ok {
		return fmt.Errorf("invalid fpath-rule %q", value)
	}
	return nil
}
