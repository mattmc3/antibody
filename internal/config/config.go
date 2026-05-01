package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
)

type deferConfig struct {
	Bundle string `toml:"bundle"`
}

type fpathConfig struct {
	Rule string `toml:"rule"`
}

type gitConfig struct {
	Domain   string `toml:"domain"`
	Protocol string `toml:"protocol"`
}

type homeConfig struct {
	Dir string `toml:"dir"`
}

// Config holds the antibody configuration.
type Config struct {
	Defer deferConfig `toml:"defer"`
	Fpath fpathConfig `toml:"fpath"`
	Git   gitConfig   `toml:"git"`
	Home  homeConfig  `toml:"home"`
}

// nolint: gochecknoglobals
var (
	instance   *Config
	instanceMu sync.Mutex
)

// Load reads ~/.config/antibody/antibody.toml and stores it as the singleton
// returned by Get. Call once at startup. A missing file is not an error.
func Load() (*Config, error) {
	instanceMu.Lock()
	defer instanceMu.Unlock()

	path, err := configPath()
	if err != nil {
		instance = &Config{}
		return instance, nil
	}
	cfg, err := loadFile(path)
	if err != nil {
		return nil, err
	}
	instance = cfg
	return cfg, nil
}

func loadFile(path string) (*Config, error) {
	cfg := &Config{}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("antibody: failed to parse config %s: %w", path, err)
	}
	return cfg, nil
}

// Get returns the singleton config. Returns defaults if Load has not been called.
func Get() *Config {
	instanceMu.Lock()
	defer instanceMu.Unlock()
	if instance == nil {
		instance = &Config{}
	}
	return instance
}

// DeferBundle returns the bundle spec for the zsh-defer tool,
// defaulting to romkatv/zsh-defer.
func (c *Config) DeferBundle() string {
	if c.Defer.Bundle == "" {
		return "romkatv/zsh-defer"
	}
	return c.Defer.Bundle
}

// FpathRule returns the fpath rule, either "append" (default) or "prepend".
// Warns to stderr if the configured value is not recognized.
func (c *Config) FpathRule() string {
	switch strings.ToLower(c.Fpath.Rule) {
	case "append", "":
		return "append"
	case "prepend":
		return "prepend"
	default:
		fmt.Fprintf(os.Stderr, "antibody: unknown fpath rule %q, using \"append\"\n", c.Fpath.Rule)
		return "append"
	}
}

// GitDomain returns the configured git hosting domain, defaulting to github.com.
func (c *Config) GitDomain() string {
	d := strings.TrimPrefix(c.Git.Domain, "https://")
	d = strings.TrimPrefix(d, "http://")
	d = strings.TrimRight(d, "/")
	if d == "" {
		return "github.com"
	}
	return d
}

// GitProtocol returns the configured git protocol, either "https" (default) or "ssh".
// Warns to stderr if the configured value is not recognized.
func (c *Config) GitProtocol() string {
	switch strings.ToLower(c.Git.Protocol) {
	case "https", "":
		return "https"
	case "ssh":
		return "ssh"
	default:
		fmt.Fprintf(os.Stderr, "antibody: unknown git protocol %q, using \"https\"\n", c.Git.Protocol)
		return "https"
	}
}

// HomeDir returns the directory where bundles are cloned.
// Priority: $ANTIBODY_HOME env var > config [home] dir > OS cache dir.
func (c *Config) HomeDir() (string, error) {
	if dir := os.Getenv("ANTIBODY_HOME"); dir != "" {
		return dir, nil
	}
	if dir := c.configHomeDir(); dir != "" {
		return dir, nil
	}
	dir, err := os.UserCacheDir()
	return filepath.Join(dir, "antibody"), err
}

// configHomeDir returns the config [home] dir with ~ expanded.
func (c *Config) configHomeDir() string {
	dir := c.Home.Dir
	if dir == "" {
		return ""
	}
	if strings.HasPrefix(dir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return dir
		}
		return filepath.Join(home, dir[2:])
	}
	return dir
}

func configPath() (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "antibody", "antibody.toml"), nil
}
