package pathstyle

import (
	"fmt"
	"net/url"
	"testing"
)

func TestEscapedStyle(t *testing.T) {
	style := &EscapedStyle{}
	u, _ := url.Parse("https://github.com/user/repo")

	path := style.FromURL(u)
	expected := "https-COLON--SLASH--SLASH-github.com-SLASH-user-SLASH-repo"
	if path != expected {
		t.Errorf("FromURL got %q, want %q", path, expected)
	}

	urlStr := style.ToURL(path)
	u2, _ := url.Parse(urlStr)
	if u2.Host != "github.com" {
		t.Errorf("ToURL host got %q, want %q", u2.Host, "github.com")
	}
}

func TestShortStyle(t *testing.T) {
	style := &ShortStyle{}
	u, _ := url.Parse("https://github.com/zsh-users/zsh-syntax-highlighting")

	path := style.FromURL(u)
	expected := "zsh-users/zsh-syntax-highlighting"
	if path != expected {
		t.Errorf("FromURL got %q, want %q", path, expected)
	}
}

func TestFullStyle(t *testing.T) {
	style := &FullStyle{}
	u, _ := url.Parse("https://github.com/zsh-users/zsh-syntax-highlighting")

	path := style.FromURL(u)
	expected := "github.com/zsh-users/zsh-syntax-highlighting"
	if path != expected {
		t.Errorf("FromURL got %q, want %q", path, expected)
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"escaped", "*pathstyle.EscapedStyle"},
		{"short", "*pathstyle.ShortStyle"},
		{"full", "*pathstyle.FullStyle"},
		{"ESCAPED", "*pathstyle.EscapedStyle"},
		{"unknown", "*pathstyle.EscapedStyle"},
	}

	for _, tt := range tests {
		got := New(tt.input)
		gotType := fmt.Sprintf("%T", got)
		if gotType != tt.want {
			t.Errorf("New(%q) = %v, want %v", tt.input, gotType, tt.want)
		}
	}
}
