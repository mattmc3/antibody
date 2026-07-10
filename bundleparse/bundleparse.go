package bundleparse

import (
	"fmt"
	"strings"
)

type Directive string

const (
	BundleDirective Directive = "bundle"
	UsingDirective  Directive = "using"
)

type ParsedLine struct {
	Directive   Directive
	Name        string
	Annotations map[string]string
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\f' || b == '\v'
}

func ParseLine(line string) (ParsedLine, error) {
	result := ParsedLine{Annotations: make(map[string]string, 8)}

	var i int
	n := len(line)

	// --- skip leading whitespace ---
	for i < n && isSpace(line[i]) {
		i++
	}

	// empty or comment-only line
	if i >= n || line[i] == '#' {
		return result, nil
	}

	// --- parse first token as bundle ---
	start := i
	for i < n && !isSpace(line[i]) && line[i] != '#' {
		i++
	}

	name := line[start:i]
	if name == "" {
		return result, nil
	}

	if strings.HasPrefix(name, "using:") {
		usingName := strings.TrimPrefix(name, "using:")
		if usingName == "" {
			return ParsedLine{}, fmt.Errorf("missing using target")
		}
		result.Directive = UsingDirective
		result.Name = usingName
	} else {
		result.Directive = BundleDirective
		result.Name = name
	}

	// --- parse remaining key:value tokens ---
	for i < n {
		// skip whitespace
		for i < n && isSpace(line[i]) {
			i++
		}

		if i >= n {
			break
		}

		// comment
		if line[i] == '#' {
			break
		}

		// --- parse key ---
		start = i
		for i < n && line[i] != ':' && !isSpace(line[i]) {
			i++
		}

		if i >= n || line[i] != ':' {
			return ParsedLine{}, fmt.Errorf("expected ':' after key near %q", line[start:])
		}

		key := line[start:i]
		i++ // skip ':'

		if key == "" {
			return ParsedLine{}, fmt.Errorf("empty key before ':'")
		}

		if i >= n {
			return ParsedLine{}, fmt.Errorf("missing value for key %q", key)
		}

		var val strings.Builder
		switch line[i] {
		case '"':
			i++
			closed := false

			for i < n {
				if line[i] == '\\' {
					i++
					if i >= n {
						return ParsedLine{}, fmt.Errorf("unterminated escape in quoted value for key %q", key)
					}
					val.WriteByte(line[i])
					i++
					continue
				}
				if line[i] == '"' {
					i++
					closed = true
					break
				}
				val.WriteByte(line[i])
				i++
			}

			if !closed {
				return ParsedLine{}, fmt.Errorf("unterminated double-quoted value for key %q", key)
			}

		case '\'':
			// --- single quoted ---
			i++
			for i < n && line[i] != '\'' {
				val.WriteByte(line[i])
				i++
			}
			if i >= n {
				return ParsedLine{}, fmt.Errorf("unterminated single-quoted value for key %q", key)
			}
			i++ // skip closing quote

		default:
			// --- unquoted ---
			for i < n && !isSpace(line[i]) && line[i] != '#' {
				if line[i] == '\\' {
					i++
					if i >= n {
						return ParsedLine{}, fmt.Errorf("unterminated escape in unquoted value for key %q", key)
					}
					val.WriteByte(line[i])
					i++
					continue
				}
				val.WriteByte(line[i])
				i++
			}
		}

		result.Annotations[key] = val.String()
	}

	return result, nil
}
