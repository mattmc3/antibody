package require

import (
	"strings"
	"testing"
)

// Fatal paths can't run under the same t; cover the pure diff renderer
// and rely on the repo's suite exercising the passing paths.
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
