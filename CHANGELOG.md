# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [7.0.0] - TBD

### Changed
- Moved project to mattmc3/antibody
- Updated to Go 1.26
- Replaced deprecated golang.org/x/crypto/ssh/terminal with golang.org/x/term
- Updated test repositories to use maintained alternatives
- Switched from Make to Just for build tasks
- Removed Travis CI and Hound CI configurations
- Removed Hugo documentation site (moved to docs/ directory)
- Removed GoReleaser config (replaced with GitHub Actions)

### Fixed
- Updated dependency versions for security and compatibility

[Unreleased]: https://github.com/mattmc3/antibody/compare/v7.0.0...HEAD
[7.0.0]: https://github.com/mattmc3/antibody/compare/v6.1.1...v7.0.0
