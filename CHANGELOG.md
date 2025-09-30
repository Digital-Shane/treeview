# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased
### Added
- Golang CI lint config and workflow

## [v1.8.2] - 2025-09-20
### Updated
- Expose Option for use in the sub S3 extension

## [v1.8.1] - 2025-09-03
### Fixed
- Windows compile error due to `syscall.Stat_t` being unable in the window build env.

## [v1.8.0] - 2025-09-02
### Added
- `WithTruncate` option for tree to allow defining max line width. 

## [v1.7.0] - 2025-08-19
### Added
- `SetNodes` option for tree to allow updates to root nodes without full tree reconstruction.

## [v1.6.0] - 2025-08-19
### Added
- `WithProgressCallback` option for all constructors to report node build progress.

## [v1.5.1] - 2025-08-17
### Fixed
- Fix focus node falling outside of viewport.

## [v1.5.0] - 2025-08-17
### Updated
- Significantly improve render performance when using a viewport.

## [v1.4.1] - 2025-08-04
### Fixed
- Default search function data casting

## [v1.4.0] - 2025-08-03
### Added
- New `AllBottomUp()` iterator for bottom-up tree traversal

## [v1.3.0] - 2025-08-03
### Updated
- Updated data returned to a pointer

## [v1.2.0] - 2025-07-30
### Added
- Ability to add arbitrary data to the `FileInfo` pre-build node type.
- Star history to readme. 

## [v1.1.0] - 2025-07-25
### Added
- Multi-Focus Support
  - Added new methods to Tree for programmatic multi-node selection.
  - New `AllFocused()` iterator to traverse all focused nodes.
  - Added shift+up/down key bindings (`ExtendUp`, `ExtendDown`) for interactive range selection in terminal UI
  - Updated renderer to highlight all focused nodes simultaneously in tree output
  - Enhanced search functionality: `SearchAndExpand()` now focuses all directly matching nodes (not just the first match)
### Updated
- `Tree.GetFocusedID()` and `Tree.GetFocusedNode()` now return the primary (first) focused node to maintain API compatibility
- `Tree.SearchAndExpand()` now highlights all matching nodes
- `Tree.ToggleFocused()` now toggles all focused nodes
- File browser example showcases multi-focus capabilities with an updated metadata panel during multi-focus

## [v1.0.0] - 2025-07-24
### Released
- My feature-rich Go library for displaying and navigating tree structures in the terminal. 😁
