<p align="center">
  <img alt="Antibody Logo" src="logo.png" height="140" />
  <h3 align="center">Antibody</h3>
  <p align="center">The fastest shell plugin manager.</p>
  <p align="center">
    <a href="/LICENSE.md">
      <img alt="Software License" src="https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square">
    </a>
    <a href="https://github.com/mattmc3/antibody/actions/workflows/ci.yml">
      <img alt="CI Status" src="https://img.shields.io/github/actions/workflow/status/mattmc3/antibody/ci.yml?style=flat-square&label=CI">
    </a>
    <a href="https://pkg.go.dev/github.com/mattmc3/antibody">
      <img alt="Go Reference" src="https://img.shields.io/badge/go.dev-reference-blue.svg?style=flat-square">
    </a>
  </p>
</p>

---

Antibody is a shell plugin manager made from the ground up thinking about
performance.

> [!NOTE]
> This fork of [getantibody/antibody](https://github.com/getantibody/antibody) is
> **ACTIVELY MAINTAINED**. While the upstream project is deprecated, this fork strives
> for feature parity with [antidote](https://github.com/mattmc3/antidote), which is its
> pure-Zsh successor. If you prefer a Go binary over a Zsh script, this is
> the antibody version you're looking for.

## Quick Start

```sh
# Install from GitHub releases
# macOS
curl -sfL https://github.com/mattmc3/antibody/releases/latest/download/antibody_darwin_arm64.tar.gz | tar -xz
sudo mv antibody /usr/local/bin/

# Linux
curl -sfL https://github.com/mattmc3/antibody/releases/latest/download/antibody_linux_amd64.tar.gz | tar -xz
sudo mv antibody /usr/local/bin/

# Or install with Go
go install github.com/mattmc3/antibody@latest

# Bundle plugins
antibody bundle < ${ZDOTDIR:-$HOME}/.zsh_plugins.txt > ${ZDOTDIR:-$HOME}/.zsh_plugins.zsh

# Load in your .zshrc
source ${ZDOTDIR:-$HOME}/.zsh_plugins.zsh
```

## Usage

Create a `${ZDOTDIR:-$HOME}/.zsh_plugins.txt` file with your desired plugins:

```
# Plugins
zsh-users/zsh-completions
zsh-users/zsh-autosuggestions
zsh-users/zsh-syntax-highlighting

# Themes
ohmyzsh/ohmyzsh path:plugins/git
```

Then run:

```sh
antibody bundle < ${ZDOTDIR:-$HOME}/.zsh_plugins.txt > ${ZDOTDIR:-$HOME}/.zsh_plugins.zsh
```

Source the generated file in your `.zshrc`:

```sh
source ${ZDOTDIR:-$HOME}/.zsh_plugins.zsh
```

## Commands

- `antibody bundle` - Download and bundle plugins
- `antibody update` - Update all plugins
- `antibody home` - Show antibody's home directory
- `antibody list` - List installed plugins
- `antibody path` - Show plugin path
- `antibody purge` - Remove a plugin
- `antibody init` - Initialize shell integration
- `antibody completions` - Generate shell completion scripts

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Thanks

- [@caarlos0](https://github.com/caarlos0), original author
- [@pragmaticivan](https://github.com/pragmaticivan), for the logo design
- All the amazing [contributors](https://github.com/mattmc3/antibody/graphs/contributors)
