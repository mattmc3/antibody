package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mattmc3/antibody/internal/require"
)

func resetSingleton(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { instance = nil })
	instance = nil
}

func testdata(name string) string {
	return filepath.Join("testdata", name)
}

func TestConfigPath_XDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/xdg")
	p, err := configPath()
	require.NoError(t, err)
	require.Equal(t, "/xdg/antibody/antibody.toml", p)
}

func TestConfigPath_NoXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	p, err := configPath()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(home, ".config", "antibody", "antibody.toml"), p)
}

func TestLoadFile_Missing(t *testing.T) {
	cfg, err := loadFile(testdata("nonexistent.toml"))
	require.NoError(t, err)
	require.That(t, cfg != nil)
}

func TestLoadFile_Valid(t *testing.T) {
	cfg, err := loadFile(testdata("antibody.toml"))
	require.NoError(t, err)
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
	require.That(t, cfg != nil)
}

func TestGet_AfterLoad(t *testing.T) {
	resetSingleton(t)
	instance, _ = loadFile(testdata("antibody.toml"))
	cfg := Get()
	require.That(t, cfg != nil)
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

func TestConfig_Compdir(t *testing.T) {
	require.Equal(t, "", (&Config{}).Compdir())
	cfg := &Config{Completions: completionsConfig{Dir: "/tmp/comps"}}
	require.Equal(t, "/tmp/comps", cfg.Compdir())
}

func TestConfig_Compdir_Tilde(t *testing.T) {
	cfg := &Config{Completions: completionsConfig{Dir: "~/comps"}}
	home, _ := os.UserHomeDir()
	require.Equal(t, filepath.Join(home, "comps"), cfg.Compdir())
}

func TestConfig_HomeDir_Default(t *testing.T) {
	t.Setenv("ANTIBODY_HOME", "")
	cfg := &Config{}
	dir, err := cfg.HomeDir()
	require.NoError(t, err)
	require.That(t, strings.HasSuffix(dir, "antibody"), "got: %s", dir)
}
