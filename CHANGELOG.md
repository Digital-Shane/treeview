# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
