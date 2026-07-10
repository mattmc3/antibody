// Package expect provides test assertions built from composable
// checks: predicates like Equals and Contains return a Check, Expect
// asserts one, and Not negates any check without a mirrored helper.
// Test files dot-import this package so assertions read as sentences:
//
//	Expect(t, Contains(out, "--dirs"))
//	Expect(t, Not(Matches(`\$@[^"]`, script)))
//
// Failures stop the test via t.Fatal.
package expect

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

// Check is the outcome of a predicate: whether it held, plus failure
// text for both polarities so Not can flip it losslessly.
type Check struct {
	ok       bool
	pos, neg string
}

// Expect fails the test when the condition does not hold. It accepts a
// predicate Check or a raw bool; pair raw bools with a message that
// prints the values under test, since a bool carries no context.
func Expect[C bool | Check](t testing.TB, c C, msgAndArgs ...any) {
	t.Helper()
	chk, isCheck := any(c).(Check)
	if !isCheck {
		chk = Check{ok: any(c).(bool), pos: "condition failed"}
	}
	if !chk.ok {
		fail(t, chk.pos, msgAndArgs)
	}
}

// Not inverts a check, swapping its failure messages.
func Not(c Check) Check {
	return Check{ok: !c.ok, pos: c.neg, neg: c.pos}
}

func NoError(err error) Check {
	return Not(AnError(err))
}

// AnError holds when err is non-nil; Expect(t, AnError(err)) reads as
// "expect an error".
func AnError(err error) Check {
	return Check{
		ok:  err != nil,
		pos: "expected an error, got nil",
		neg: fmt.Sprintf("unexpected error: %v", err),
	}
}

func Equals[T comparable](want, got T) Check {
	return Check{
		ok:  want == got,
		pos: "not equal:\n" + describe(want, got),
		neg: fmt.Sprintf("should not be equal: %#v", got),
	}
}

func Contains(s, substr string) Check {
	return Check{
		ok:  strings.Contains(s, substr),
		pos: fmt.Sprintf("%q\nshould contain %q", s, substr),
		neg: fmt.Sprintf("%q\nshould not contain %q", s, substr),
	}
}

func Matches(pattern, s string) Check {
	return Check{
		ok:  regexp.MustCompile(pattern).MatchString(s),
		pos: fmt.Sprintf("%q\nshould match %q", s, pattern),
		neg: fmt.Sprintf("%q\nshould not match %q", s, pattern),
	}
}

func FileExists(path string) Check {
	info, err := os.Stat(path)
	c := Check{
		ok:  err == nil && !info.IsDir(),
		neg: "file should not exist: " + path,
	}
	if err != nil {
		c.pos = fmt.Sprintf("file does not exist: %s (%v)", path, err)
	} else if info.IsDir() {
		c.pos = "expected a file, found a directory: " + path
	}
	return c
}

func DirExists(path string) Check {
	info, err := os.Stat(path)
	c := Check{
		ok:  err == nil && info.IsDir(),
		neg: "directory should not exist: " + path,
	}
	if err != nil {
		c.pos = fmt.Sprintf("directory does not exist: %s (%v)", path, err)
	} else if !info.IsDir() {
		c.pos = "expected a directory, found a file: " + path
	}
	return c
}

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
