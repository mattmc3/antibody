package expect

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Predicates are pure, so their pass/fail logic and failure text are
// testable directly; Expect's fatal path is exercised by the whole suite.

func TestEquals(t *testing.T) {
	if !Equals(1, 1).ok {
		t.Fatal("1 == 1 should hold")
	}
	c := Equals("a", "b")
	if c.ok {
		t.Fatal("a == b should not hold")
	}
	if !strings.Contains(c.pos, `"a"`) || !strings.Contains(c.pos, `"b"`) {
		t.Fatalf("failure text missing operands: %s", c.pos)
	}
}

func TestContains(t *testing.T) {
	if !Contains("hello world", "world").ok {
		t.Fatal("should hold")
	}
	c := Contains("hello", "bye")
	if c.ok {
		t.Fatal("should not hold")
	}
	if !strings.Contains(c.pos, "hello") || !strings.Contains(c.pos, "bye") {
		t.Fatalf("failure text missing operands: %s", c.pos)
	}
}

func TestMatches(t *testing.T) {
	if !Matches(`^a+$`, "aaa").ok {
		t.Fatal("should hold")
	}
	if Matches(`^a+$`, "bbb").ok {
		t.Fatal("should not hold")
	}
}

func TestNotSwapsMessages(t *testing.T) {
	c := Contains("hello", "he")
	n := Not(c)
	if n.ok {
		t.Fatal("negation should not hold")
	}
	if !strings.Contains(n.pos, "should not contain") {
		t.Fatalf("negated failure text wrong: %s", n.pos)
	}
	if Not(n) != c {
		t.Fatal("double negation should restore the check")
	}
}

func TestFileAndDirExists(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f.txt")
	if FileExists(file).ok {
		t.Fatal("missing file should not hold")
	}
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !FileExists(file).ok {
		t.Fatal("existing file should hold")
	}
	if FileExists(dir).ok {
		t.Fatal("dir is not a file")
	}
	if !DirExists(dir).ok {
		t.Fatal("existing dir should hold")
	}
	if DirExists(file).ok {
		t.Fatal("file is not a dir")
	}
}

func TestDescribeMultilineFirstDiff(t *testing.T) {
	out := describe("a\nb\nc", "a\nX\nc")
	if !strings.Contains(out, "line 2") || !strings.Contains(out, `"b"`) || !strings.Contains(out, `"X"`) {
		t.Fatalf("bad diff output: %s", out)
	}
}

func TestDescribeLengthMismatch(t *testing.T) {
	out := describe("a\nb", "a\nb\nextra")
	if !strings.Contains(out, "line 3") {
		t.Fatalf("bad diff output: %s", out)
	}
}

func TestDescribeNonString(t *testing.T) {
	out := describe(1, 2)
	if !strings.Contains(out, "want: 1") || !strings.Contains(out, "got:  2") {
		t.Fatalf("bad output: %s", out)
	}
}
