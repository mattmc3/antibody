// Package require provides the subset of testify's require API these
// tests actually use, backed only by the standard library. Failures
// stop the test via t.Fatal.
package require

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

func fail(t testing.TB, msg string, msgAndArgs []any) {
	t.Helper()
	if len(msgAndArgs) > 0 {
		if format, ok := msgAndArgs[0].(string); ok {
			msg += "\nmessage: " + fmt.Sprintf(format, msgAndArgs[1:]...)
		} else {
			msg += "\nmessage: " + fmt.Sprint(msgAndArgs...)
		}
	}
	t.Fatal(msg)
}

func NoError(t testing.TB, err error, msgAndArgs ...any) {
	t.Helper()
	if err != nil {
		fail(t, fmt.Sprintf("unexpected error: %v", err), msgAndArgs)
	}
}

func Error(t testing.TB, err error, msgAndArgs ...any) {
	t.Helper()
	if err == nil {
		fail(t, "expected an error, got nil", msgAndArgs)
	}
}

func Equal[T comparable](t testing.TB, want, got T, msgAndArgs ...any) {
	t.Helper()
	if want != got {
		fail(t, "not equal:\n"+describe(want, got), msgAndArgs)
	}
}

func Contains(t testing.TB, s, substr string, msgAndArgs ...any) {
	t.Helper()
	if !strings.Contains(s, substr) {
		fail(t, fmt.Sprintf("%q\nshould contain %q", s, substr), msgAndArgs)
	}
}

func NotContains(t testing.TB, s, substr string, msgAndArgs ...any) {
	t.Helper()
	if strings.Contains(s, substr) {
		fail(t, fmt.Sprintf("%q\nshould not contain %q", s, substr), msgAndArgs)
	}
}

func FileExists(t testing.TB, path string, msgAndArgs ...any) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		fail(t, fmt.Sprintf("file does not exist: %s (%v)", path, err), msgAndArgs)
	} else if info.IsDir() {
		fail(t, fmt.Sprintf("expected a file, found a directory: %s", path), msgAndArgs)
	}
}

func DirExists(t testing.TB, path string, msgAndArgs ...any) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		fail(t, fmt.Sprintf("directory does not exist: %s (%v)", path, err), msgAndArgs)
	} else if !info.IsDir() {
		fail(t, fmt.Sprintf("expected a directory, found a file: %s", path), msgAndArgs)
	}
}

// That asserts an arbitrary condition; pair it with a message that
// prints the values under test.
func That(t testing.TB, ok bool, msgAndArgs ...any) {
	t.Helper()
	if !ok {
		fail(t, "condition failed", msgAndArgs)
	}
}

func Match(t testing.TB, pattern, s string, msgAndArgs ...any) {
	t.Helper()
	if !regexp.MustCompile(pattern).MatchString(s) {
		fail(t, fmt.Sprintf("%q\nshould match %q", s, pattern), msgAndArgs)
	}
}

func NotMatch(t testing.TB, pattern, s string, msgAndArgs ...any) {
	t.Helper()
	if regexp.MustCompile(pattern).MatchString(s) {
		fail(t, fmt.Sprintf("%q\nshould not match %q", s, pattern), msgAndArgs)
	}
}

// describe renders a want/got pair; multiline strings report the first
// differing line instead of two opaque blobs.
func describe(want, got any) string {
	ws, wok := want.(string)
	gs, gok := got.(string)
	if wok && gok && (strings.Contains(ws, "\n") || strings.Contains(gs, "\n")) {
		wl, gl := strings.Split(ws, "\n"), strings.Split(gs, "\n")
		for i := 0; i < len(wl) || i < len(gl); i++ {
			var w, g string
			if i < len(wl) {
				w = wl[i]
			}
			if i < len(gl) {
				g = gl[i]
			}
			if w != g {
				return fmt.Sprintf("first difference at line %d:\nwant: %q\ngot:  %q", i+1, w, g)
			}
		}
	}
	return fmt.Sprintf("want: %#v\ngot:  %#v", want, got)
}
