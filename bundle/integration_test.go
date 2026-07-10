package bundle

import (
	"testing"

	. "github.com/mattmc3/antibody/internal/expect"
)

// Integration tests clone real repos over the network.
// Run with `just test all`; `just test` skips them via -short.

func skipShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test: skipped in -short mode")
	}
}

func TestZshBundleWithNoShFiles(t *testing.T) {
	skipShort(t)
	home := home(t)
	bundle, err := New(home, "mattmc3/antibody")
	Expect(t, NoError(err))
	_, err = bundle.Get()
	Expect(t, NoError(err))
}
