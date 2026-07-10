package antibodylib

import (
	"testing"

	. "github.com/mattmc3/antibody/internal/expect"
)

// Output lines must come back in input order regardless of the order
// parallel workers append them; index -1 (the defer ensure block) sorts
// first.
func TestIndexedLinesOrder(t *testing.T) {
	var s safeIndexedLines
	s.Append(indexedLine{idx: 2, line: "third"})
	s.Append(indexedLine{idx: -1, line: "first"})
	s.Append(indexedLine{idx: 0, line: "second"})
	Expect(t, Equals("first\nsecond\nthird", s.Items().String()))
}

func TestIndexedLinesEmpty(t *testing.T) {
	var s safeIndexedLines
	Expect(t, Equals("", s.Items().String()))
}
