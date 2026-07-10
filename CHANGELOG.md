# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [7.0.1] - 2026-07-10

### Changed
- Replaced the kingpin CLI library with Go standard library flag parsing;
  help output is reformatted slightly and stacked short flags (eg: `-du`)
  are no longer accepted
- Replaced testify with an internal assertion package; antibody now has
  no test-only dependencies
- Absorbed the bundleparse library into this repo as an importable
  package; go.mod is down to x/sync and BurntSushi/toml

### Fixed
- `antibody help <command>` and `-h` after a subcommand show that
  command's usage

## [7.0.0] - 2026-07-05

### Added
- Config file support via `~/.config/antibody/antibody.toml` (git domain
  and protocol, fpath rule, defer bundle, home dir, completions dir)
- `kind:defer` bundles, deferred with romkatv/zsh-defer
- `kind:autoload` bundles and the `autoload:` annotation
- `pre:` and `post:` hook annotations
- `if:` conditional annotation
- `fpath-rule:` annotation with append/prepend rules
- `pin:` annotation to pin a bundle to a specific commit (full 40-char SHA)
- Quoted values in bundle annotations
- `antibody completions zsh` subcommand; `--fpath` writes the `_antibody`
  file to the completions dir and prints its path
- `antibody list --dirs` and `--url` flags

### Changed
- Moved project to mattmc3/antibody; this fork is actively maintained and
  targets feature parity with antidote
- Tab completion for `antibody` now uses zsh's completion system, with
  subcommand descriptions and bundle name completion
- Bundle lines are parsed with the same parser as antidote, so
  `.zsh_plugins.txt` files work in both
- Updated to Go 1.26 and refreshed all dependencies

### Fixed
- Bundle outputs `$HOME` instead of absolute paths, so generated files are
  portable across machines
- Plugins with multiple .zsh files source only their main init file, the
  same one antidote picks
- fpath lines are emitted before source lines so completions load reliably
- Plugins files with Windows or bare carriage-return line endings parse
  correctly

[Unreleased]: https://github.com/mattmc3/antibody/compare/v7.0.0...HEAD
[7.0.0]: https://github.com/mattmc3/antibody/compare/v6.1.1...v7.0.0
