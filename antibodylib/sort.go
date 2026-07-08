package antibodylib

import (
	"cmp"
	"slices"
	"strings"
	"sync"
)

type indexedLine struct {
	idx  int
	line string
}

type indexedLines []indexedLine

type safeIndexedLines struct {
	mutex sync.Mutex
	data  indexedLines
}

// Append safely appends items to the slice
func (slice *safeIndexedLines) Append(item indexedLine) {
	slice.mutex.Lock()
	defer slice.mutex.Unlock()

	slice.data = append(slice.data, item)
}

func (slice *safeIndexedLines) Items() indexedLines {
	slice.mutex.Lock()
	defer slice.mutex.Unlock()
	return slice.data
}

// Sort all lines and join them in a string
func (slice indexedLines) String() string {
	slices.SortFunc(slice, func(a, b indexedLine) int {
		return cmp.Compare(a.idx, b.idx)
	})
	lines := make([]string, 0, len(slice))
	for _, line := range slice {
		lines = append(lines, line.line)
	}
	return strings.Join(lines, "\n")
}
