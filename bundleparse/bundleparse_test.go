package bundleparse

import (
	"testing"

	. "github.com/mattmc3/antibody/internal/expect"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ParsedLine
		wantErr bool
	}{
		{
			name:  "basic bundle",
			input: `zsh-users/zsh-autosuggestions kind:zsh`,
			want: ParsedLine{
				Directive: BundleDirective,
				Name:      "zsh-users/zsh-autosuggestions",
				Annotations: map[string]string{
					"kind": "zsh",
				},
			},
		},
		{
			name:  "only bundle",
			input: `zsh-users/zsh-autosuggestions`,
			want: ParsedLine{
				Directive: BundleDirective,
				Name:      "zsh-users/zsh-autosuggestions",
			},
		},
		{
			name:  "double quoted value",
			input: `foo pre:"echo hello world"`,
			want: ParsedLine{
				Directive: BundleDirective,
				Name:      "foo",
				Annotations: map[string]string{
					"pre": "echo hello world",
				},
			},
		},
		{
			name:  "single quoted value",
			input: `foo post:'echo goodbye world'`,
			want: ParsedLine{
				Directive: BundleDirective,
				Name:      "foo",
				Annotations: map[string]string{
					"post": "echo goodbye world",
				},
			},
		},
		{
			name:  "mixed quotes",
			input: `rupa/z pre:"echo 'hello' world" post:'echo "goodbye" world'`,
			want: ParsedLine{
				Directive: BundleDirective,
				Name:      "rupa/z",
				Annotations: map[string]string{
					"pre":  "echo 'hello' world",
					"post": `echo "goodbye" world`,
				},
			},
		},
		{
			name:  "escaped characters in double quotes",
			input: `foo pre:"echo \"hello\" world"`,
			want: ParsedLine{
				Directive: BundleDirective,
				Name:      "foo",
				Annotations: map[string]string{
					"pre": `echo "hello" world`,
				},
			},
		},
		{
			name:  "comment ignored",
			input: `foo kind:zsh # this is a comment`,
			want: ParsedLine{
				Directive: BundleDirective,
				Name:      "foo",
				Annotations: map[string]string{
					"kind": "zsh",
				},
			},
		},
		{
			name:  "trailing whitespace",
			input: "   foo    kind:zsh   ",
			want: ParsedLine{
				Directive: BundleDirective,
				Name:      "foo",
				Annotations: map[string]string{
					"kind": "zsh",
				},
			},
		},
		{
			name:  "backslash escape",
			input: "foo/bar pre:echo\\ here",
			want: ParsedLine{
				Directive: BundleDirective,
				Name:      "foo/bar",
				Annotations: map[string]string{
					"pre": "echo here",
				},
			},
		},
		{
			name:  "path with slashes",
			input: `ohmyzsh/ohmyzsh path:plugins/git`,
			want: ParsedLine{
				Directive: BundleDirective,
				Name:      "ohmyzsh/ohmyzsh",
				Annotations: map[string]string{
					"path": "plugins/git",
				},
			},
		},
		{
			name:  "empty line",
			input: ``,
			want:  ParsedLine{},
		},
		{
			name:    "missing key value colon",
			input:   `foo kind`,
			wantErr: true,
		},
		{
			name:    "missing value",
			input:   `foo kind:`,
			wantErr: true,
		},
		{
			name:    "unterminated double quote",
			input:   `foo pre:"echo hello`,
			wantErr: true,
		},
		{
			name:    "unterminated single quote",
			input:   `foo pre:'echo hello`,
			wantErr: true,
		},
		{
			name:    "unterminated escape",
			input:   `foo pre:"echo hello\`,
			wantErr: true,
		},
		{
			name:  "multiple annotations",
			input: `foo/bar kind:zsh branch:main path:plugins/git`,
			want: ParsedLine{
				Directive: BundleDirective,
				Name:      "foo/bar",
				Annotations: map[string]string{
					"kind":   "zsh",
					"branch": "main",
					"path":   "plugins/git",
				},
			},
		},
		{
			name:  "bundle-like token",
			input: `kind:zsh`,
			want: ParsedLine{
				Directive: BundleDirective,
				Name:      "kind:zsh",
			},
		},
		{
			name:  "Full SSH URL",
			input: `git@github.com:zsh-users/zsh-autosuggestions kind:zsh post:a:b:c  # comment`,
			want: ParsedLine{
				Directive: BundleDirective,
				Name:      "git@github.com:zsh-users/zsh-autosuggestions",
				Annotations: map[string]string{
					"kind": "zsh",
					"post": "a:b:c",
				},
			},
		},
		{
			name:  "Full URL",
			input: `https://github.com/zsh-users/zsh-autosuggestions`,
			want: ParsedLine{
				Directive: BundleDirective,
				Name:      "https://github.com/zsh-users/zsh-autosuggestions",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLine(tt.input)

			if tt.wantErr {
				Expect(t, AnError(err), "result=%v", got)
				return
			}
			Expect(t, NoError(err))

			if got.Annotations == nil {
				got.Annotations = map[string]string{}
			}
			want := tt.want
			if want.Annotations == nil {
				want.Annotations = map[string]string{}
			}

			Expect(t, DeepEquals(want, got))
		})
	}
}

func TestParseLine_UsingDirective(t *testing.T) {
	got, err := ParseLine(`using:foo/bar kind:zsh`)
	Expect(t, NoError(err))

	want := ParsedLine{
		Directive: UsingDirective,
		Name:      "foo/bar",
		Annotations: map[string]string{
			"kind": "zsh",
		},
	}

	Expect(t, DeepEquals(want, got))
}
