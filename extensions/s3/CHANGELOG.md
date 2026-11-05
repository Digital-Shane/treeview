# Release Notes
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
## [v2.0.2] - 2026-11-05
### Changed
- Better upward propagation of potential context error.
## [v2.0.1] - 2026-10-13
### Changed
- Bumped the failsafe to its latest version
- New implementation of the prefixes and objects' extraction.
- Removed dependency to `pterm`
### Fixed
- Fixed some missed objects when listing the objects
### Security
- Bumped some vulnerable dependencies.  (OK with `govulncheck` and `snyk`)

## [v2.0.0] - 2026-04-22
### Changed
- Update major version to stay inline with treeview
- Import path to include /v2

## [v1.0.0] - 2026-04-22
### Changed
- Updates to be compatible with treeview v2.

## [v0.1.1] - 2025-10-05
### Changed
- S3 constructor uses concurrency to scan the bucket.  Thus, the constructor is faster than in v0.1.0.

## [v0.1.0] - 2025-09-19
### Added
- Initial version of the s3 extension module
