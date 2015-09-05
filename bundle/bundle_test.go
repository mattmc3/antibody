package bundle_test

import (
	"os"
	"testing"

	"github.com/caarlos0/antibody/bundle"
	"github.com/caarlos0/antibody/internal"
	"github.com/stretchr/testify/assert"
)

func TestBundleSourceables(t *testing.T) {
	home := internal.TempHome()
	defer os.RemoveAll(home)
	b := bundle.New("caarlos0/zsh-pg", home)
	b.Download()
	assert.NotEmpty(t, b.Sourceables())
}

func TestSourceablesWithoutDownload(t *testing.T) {
	home := internal.TempHome()
	defer os.RemoveAll(home)
	b := bundle.New("caarlos0/zsh-pg", home)
	assert.Empty(t, b.Sourceables())
}

func TestListEmptyFolder(t *testing.T) {
	home := internal.TempHome()
	defer os.RemoveAll(home)
	assert.Empty(t, bundle.List(home))
}

func TestList(t *testing.T) {
	home := internal.TempHome()
	defer os.RemoveAll(home)
	bundle.New("caarlos0/zsh-pg", home).Download()
	assert.NotEmpty(t, bundle.List(home))
}
