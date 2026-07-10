package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/mattmc3/antibody/internal/expect"
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
	Expect(t, NoError(err))
	Expect(t, Equals("/xdg/antibody/antibody.toml", p))
}

func TestConfigPath_NoXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home, err := os.UserHomeDir()
	Expect(t, NoError(err))
	p, err := configPath()
	Expect(t, NoError(err))
	Expect(t, Equals(filepath.Join(home, ".config", "antibody", "antibody.toml"), p))
}

func TestLoadFile_Missing(t *testing.T) {
	cfg, err := loadFile(testdata("nonexistent.toml"))
	Expect(t, NoError(err))
	Expect(t, cfg != nil)
}

func TestLoadFile_Valid(t *testing.T) {
	cfg, err := loadFile(testdata("antibody.toml"))
	Expect(t, NoError(err))
	Expect(t, Equals("romkatv/zsh-defer", cfg.Defer.Bundle))
	Expect(t, Equals("append", cfg.Fpath.Rule))
	Expect(t, Equals("github.com", cfg.Git.Domain))
	Expect(t, Equals("https", cfg.Git.Protocol))
}

func TestLoadFile_Malformed(t *testing.T) {
	_, err := loadFile(testdata("malformed.toml"))
	Expect(t, AnError(err))
	Expect(t, Contains(err.Error(), "failed to parse config"))
}

func TestGet_BeforeLoad(t *testing.T) {
	resetSingleton(t)
	cfg := Get()
	Expect(t, cfg != nil)
}

func TestGet_AfterLoad(t *testing.T) {
	resetSingleton(t)
	instance, _ = loadFile(testdata("antibody.toml"))
	cfg := Get()
	Expect(t, cfg != nil)
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
		Expect(t, Equals(tt.want, cfg.GitDomain()), "input: %q", tt.input)
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
		Expect(t, Equals(tt.want, cfg.GitProtocol()), "input: %q", tt.input)
	}

	// invalid warns but returns https
	cfg := &Config{Git: gitConfig{Protocol: "ftp"}}
	Expect(t, Equals("https", cfg.GitProtocol()))
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
		Expect(t, Equals(tt.want, cfg.FpathRule()), "input: %q", tt.input)
	}

	// invalid warns but returns append
	cfg := &Config{Fpath: fpathConfig{Rule: "bogus"}}
	Expect(t, Equals("append", cfg.FpathRule()))
}

func TestConfig_HomeDir_EnvVar(t *testing.T) {
	t.Setenv("ANTIBODY_HOME", "/tmp/test-antibody")
	cfg := &Config{}
	dir, err := cfg.HomeDir()
	Expect(t, NoError(err))
	Expect(t, Equals("/tmp/test-antibody", dir))
}

func TestConfig_HomeDir_Config(t *testing.T) {
	t.Setenv("ANTIBODY_HOME", "")
	cfg := &Config{Home: homeConfig{Dir: "/tmp/config-antibody"}}
	dir, err := cfg.HomeDir()
	Expect(t, NoError(err))
	Expect(t, Equals("/tmp/config-antibody", dir))
}

func TestConfig_HomeDir_Tilde(t *testing.T) {
	t.Setenv("ANTIBODY_HOME", "")
	cfg := &Config{Home: homeConfig{Dir: "~/antibody"}}
	dir, err := cfg.HomeDir()
	Expect(t, NoError(err))
	home, _ := os.UserHomeDir()
	Expect(t, Equals(filepath.Join(home, "antibody"), dir))
}

func TestConfig_DeferBundle(t *testing.T) {
	Expect(t, Equals("romkatv/zsh-defer", (&Config{}).DeferBundle()))
	Expect(t, Equals("romkatv/zsh-defer", (&Config{Defer: deferConfig{Bundle: "romkatv/zsh-defer"}}).DeferBundle()))
	Expect(t, Equals("myorg/my-defer", (&Config{Defer: deferConfig{Bundle: "myorg/my-defer"}}).DeferBundle()))
}

func TestConfig_Compdir(t *testing.T) {
	Expect(t, Equals("", (&Config{}).Compdir()))
	cfg := &Config{Completions: completionsConfig{Dir: "/tmp/comps"}}
	Expect(t, Equals("/tmp/comps", cfg.Compdir()))
}

func TestConfig_Compdir_Tilde(t *testing.T) {
	cfg := &Config{Completions: completionsConfig{Dir: "~/comps"}}
	home, _ := os.UserHomeDir()
	Expect(t, Equals(filepath.Join(home, "comps"), cfg.Compdir()))
}

func TestConfig_HomeDir_Default(t *testing.T) {
	t.Setenv("ANTIBODY_HOME", "")
	cfg := &Config{}
	dir, err := cfg.HomeDir()
	Expect(t, NoError(err))
	Expect(t, strings.HasSuffix(dir, "antibody"), "got: %s", dir)
}
