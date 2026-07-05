package bundle

import (
	"testing"

	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)
	_, err = bundle.Get()
	require.NoError(t, err)
}
