<p align="center">
  <img alt="Antibody Logo" src="logo.png" height="140" />
  <h3 align="center">Antibody</h3>
  <p align="center">The fastest shell plugin manager.</p>
  <p align="center">
    <a href="/LICENSE.md">
      <img alt="Software License" src="https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square">
    </a>
    <a href="https://goreportcard.com/report/github.com/mattmc3/antibody">
      <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/mattmc3/antibody?style=flat-square">
    </a>
    <a href="http://godoc.org/github.com/mattmc3/antibody">
      <img alt="Go Doc" src="https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square">
    </a>
  </p>
</p>

---

Antibody is a shell plugin manager made from the ground up thinking about
performance.

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
- `antibody init` - Initialize shell integration

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Thanks

- [@caarlos0](https://github.com/caarlos0), original author
- [@pragmaticivan](https://github.com/pragmaticivan), for the logo design
- All the amazing [contributors](https://github.com/mattmc3/antibody/graphs/contributors)
