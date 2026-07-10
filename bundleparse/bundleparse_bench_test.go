package bundleparse

import (
	"os"
	"strings"
	"testing"
	"unicode"
)

func parseLineWithUnicode(line string) (map[string]string, error) {
	result := make(map[string]string)

	var i int
	n := len(line)

	for i < n && unicode.IsSpace(rune(line[i])) {
		i++
	}

	if i >= n || line[i] == '#' {
		return result, nil
	}

	start := i
	for i < n && !unicode.IsSpace(rune(line[i])) && line[i] != '#' {
		i++
	}

	bundle := line[start:i]
	if bundle == "" {
		return result, nil
	}

	result["bundle"] = bundle

	for i < n {
		for i < n && unicode.IsSpace(rune(line[i])) {
			i++
		}

		if i >= n {
			break
		}

		if line[i] == '#' {
			break
		}

		start = i
		for i < n && line[i] != ':' && !unicode.IsSpace(rune(line[i])) {
			i++
		}

		if i >= n || line[i] != ':' {
			return nil, nil
		}

		key := line[start:i]
		i++
		if i >= n {
			return nil, nil
		}

		var val strings.Builder
		switch line[i] {
		case '"':
			i++
			for i < n {
				if line[i] == '\\' {
					i++
					if i >= n {
						return nil, nil
					}
					val.WriteByte(line[i])
					i++
					continue
				}
				if line[i] == '"' {
					i++
					break
				}
				val.WriteByte(line[i])
				i++
			}
		case '\'':
			i++
			for i < n && line[i] != '\'' {
				val.WriteByte(line[i])
				i++
			}
			if i >= n {
				return nil, nil
			}
			i++
		default:
			for i < n && !unicode.IsSpace(rune(line[i])) && line[i] != '#' {
				if line[i] == '\\' {
					i++
					if i >= n {
						return nil, nil
					}
					val.WriteByte(line[i])
					i++
					continue
				}
				val.WriteByte(line[i])
				i++
			}
		}

		result[key] = val.String()
	}

	return result, nil
}

func BenchmarkParseLineMap(b *testing.B) {
	line := `foo/bar kind:zsh pin:v1 branch:main conditional:if-true autoload:yes pre:"echo hi" post:'echo bye' fpath-rule:prepend unknown:yes`
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ParseLine(line)
	}
}

func BenchmarkParseLineUnicode(b *testing.B) {
	line := `foo/bar kind:zsh pin:v1 branch:main conditional:if-true autoload:yes pre:"echo hi" post:'echo bye' fpath-rule:prepend unknown:yes`
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = parseLineWithUnicode(line)
	}
}

func BenchmarkParseLargeBundleFile(b *testing.B) {
	data, err := os.ReadFile("../scripts/profiling/bundles.txt")
	if err != nil {
		b.Fatalf("read test data: %v", err)
	}
	input := string(data)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := ParseBundles(input)
		if err != nil {
			b.Fatalf("parse bundles: %v", err)
		}
	}
}
