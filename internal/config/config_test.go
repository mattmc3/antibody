package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func resetSingleton(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { instance = nil })
	instance = nil
}

func testdata(name string) string {
	return filepath.Join("testdata", name)
}

func TestLoadFile_Missing(t *testing.T) {
	cfg, err := loadFile(testdata("nonexistent.toml"))
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, "", cfg.Bundle.PathStyle)
}

func TestLoadFile_Valid(t *testing.T) {
	cfg, err := loadFile(testdata("antibody.toml"))
	require.NoError(t, err)
	require.Equal(t, "escaped", cfg.Bundle.PathStyle)
	require.Equal(t, "romkatv/zsh-defer", cfg.Defer.Bundle)
	require.Equal(t, "append", cfg.Fpath.Rule)
	require.Equal(t, "github.com", cfg.Git.Domain)
	require.Equal(t, "https", cfg.Git.Protocol)
}

func TestLoadFile_Malformed(t *testing.T) {
	_, err := loadFile(testdata("malformed.toml"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse config")
}

func TestGet_BeforeLoad(t *testing.T) {
	resetSingleton(t)
	cfg := Get()
	require.NotNil(t, cfg)
	require.Equal(t, "", cfg.Bundle.PathStyle)
}

func TestGet_AfterLoad(t *testing.T) {
	resetSingleton(t)
	instance, _ = loadFile(testdata("antibody.toml"))
	cfg := Get()
	require.Equal(t, "escaped", cfg.Bundle.PathStyle)
}

func TestConfig_PathStyle(t *testing.T) {
	tests := []struct {
		input    string
		wantType string
	}{
		{"escaped", "*pathstyle.EscapedStyle"},
		{"short", "*pathstyle.ShortStyle"},
		{"full", "*pathstyle.FullStyle"},
		{"", "*pathstyle.EscapedStyle"},
	}
	for _, tt := range tests {
		cfg := &Config{Bundle: bundleConfig{PathStyle: tt.input}}
		got := cfg.PathStyle()
		require.NotNil(t, got)
	}

	// invalid value warns but returns escaped
	cfg := &Config{Bundle: bundleConfig{PathStyle: "bogus"}}
	got := cfg.PathStyle()
	require.NotNil(t, got)
}

func TestConfig_GitDomain(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "github.com"},
		{"github.com", "github.com"},
		{"https://github.com", "github.com"},
		{"http://github.com", "github.com"},
		{"gitlab.com/", "gitlab.com"},
		{"https://gitlab.com/", "gitlab.com"},
	}
	for _, tt := range tests {
		cfg := &Config{Git: gitConfig{Domain: tt.input}}
		require.Equal(t, tt.want, cfg.GitDomain(), "input: %q", tt.input)
	}
}

func TestConfig_GitProtocol(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "https"},
		{"https", "https"},
		{"HTTPS", "https"},
		{"ssh", "ssh"},
		{"SSH", "ssh"},
	}
	for _, tt := range tests {
		cfg := &Config{Git: gitConfig{Protocol: tt.input}}
		require.Equal(t, tt.want, cfg.GitProtocol(), "input: %q", tt.input)
	}

	// invalid warns but returns https
	cfg := &Config{Git: gitConfig{Protocol: "ftp"}}
	require.Equal(t, "https", cfg.GitProtocol())
}

func TestConfig_FpathRule(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "append"},
		{"append", "append"},
		{"APPEND", "append"},
		{"prepend", "prepend"},
		{"PREPEND", "prepend"},
	}
	for _, tt := range tests {
		cfg := &Config{Fpath: fpathConfig{Rule: tt.input}}
		require.Equal(t, tt.want, cfg.FpathRule(), "input: %q", tt.input)
	}

	// invalid warns but returns append
	cfg := &Config{Fpath: fpathConfig{Rule: "bogus"}}
	require.Equal(t, "append", cfg.FpathRule())
}

func TestConfig_HomeDir_EnvVar(t *testing.T) {
	t.Setenv("ANTIBODY_HOME", "/tmp/test-antibody")
	cfg := &Config{}
	dir, err := cfg.HomeDir()
	require.NoError(t, err)
	require.Equal(t, "/tmp/test-antibody", dir)
}

func TestConfig_HomeDir_Config(t *testing.T) {
	t.Setenv("ANTIBODY_HOME", "")
	cfg := &Config{Home: homeConfig{Dir: "/tmp/config-antibody"}}
	dir, err := cfg.HomeDir()
	require.NoError(t, err)
	require.Equal(t, "/tmp/config-antibody", dir)
}

func TestConfig_HomeDir_Tilde(t *testing.T) {
	t.Setenv("ANTIBODY_HOME", "")
	cfg := &Config{Home: homeConfig{Dir: "~/antibody"}}
	dir, err := cfg.HomeDir()
	require.NoError(t, err)
	home, _ := os.UserHomeDir()
	require.Equal(t, filepath.Join(home, "antibody"), dir)
}

func TestConfig_DeferBundle(t *testing.T) {
	require.Equal(t, "romkatv/zsh-defer", (&Config{}).DeferBundle())
	require.Equal(t, "romkatv/zsh-defer", (&Config{Defer: deferConfig{Bundle: "romkatv/zsh-defer"}}).DeferBundle())
	require.Equal(t, "myorg/my-defer", (&Config{Defer: deferConfig{Bundle: "myorg/my-defer"}}).DeferBundle())
}

func TestConfig_HomeDir_Default(t *testing.T) {
	t.Setenv("ANTIBODY_HOME", "")
	cfg := &Config{}
	dir, err := cfg.HomeDir()
	require.NoError(t, err)
	require.True(t, strings.HasSuffix(dir, "antibody"), "got: %s", dir)
}
