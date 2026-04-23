# TreeView

**TreeView** is a feature-rich Go library for displaying and navigating tree structures in the terminal.
TreeView has full [Bubble Tea][bt] and [Lipgloss][lg] support, allowing you to build glamorous, interactive terminal applications.

If you're upgrading from v1, the [v2 update guide](./v2_updates.md) walks
through the important changes and what to watch for.

## Why TreeView?

What sets TreeView apart is its flexibility. You can build trees from any data source: flat lists with parent relationships, 
pre-nested hierarchical data, or create trees from your filesystem. The search and filtering system uses customizable
predicates, giving you complete control over how users find what they're looking for. Search results automatically expand
ancestor nodes so matches are visible, and the real-time filtering keeps the tree structure intact while hiding irrelevant
items. When dealing with large trees, the bubbletea viewport support ensures smooth scrolling and focus.
Navigation feels natural with keyboard bindings that you can customize to match your app's needs.

The styling system is particularly powerful. TreeView automatically detects file types and applies appropriate
icons, but you can customize everything through the provider system to style any hierarchical with glamor. Whether you
want subtle themes or vibrant, emoji-heavy displays, the Lipgloss integration makes it straightforward.

TreeView is ready for production. It provides proper error handling with sentinel errors, operations respect context
cancellation and timeouts, and the library is thread-safe where it matters.

---
## Examples

The examples directory contains twelve progressively complex demonstrations, from basic tree rendering to a fully-featured
file browser. Each example includes comments and animated demos, so you can see exactly how to implement
specific features in your own applications.

### Basic Examples

Basic examples only demonstrate building and rendering a tree in the console.

**[Simple Tree](examples/basic/01-basic-simple/)**  
This example demonstrates the fundamental tree creation and rendering capabilities of the TreeView library.

<img src="https://vhs.charm.sh/vhs-6m0PTqHHwRYCffGFDIwrNq.gif" alt="Demo" height="400">

**[Iterators](examples/basic/02-iterators/)**  
This example demonstrates how to range over a tree's nodes using the provided iterators.
  
<img src="https://vhs.charm.sh/vhs-2s5f1uZvkwedoZjRFGu8Si.gif" alt="Demo" height="400">

**[Custom Styling](examples/basic/03-styling/)**  
A showcase of the styling capabilities provided via [lipgloss][lg] support
  
<img src="https://vhs.charm.sh/vhs-K3MxAtpw9ayuGwvQwOMbH.gif" alt="Demo" height="400">

**[Comparison Display](examples/basic/04-comparison-display/)**  
This example demonstrates visualizing comparison operations with before and after displays.

<img src="https://vhs.charm.sh/vhs-2UxOtkSnq8DgGpriGioQaL.gif" alt="Demo" height="400">

**[Workflow Visualization](examples/basic/05-workflow-visualization/)**  
Example showing how to create domain-specific tree nodes for process and workflow tracking.
  
<img src="https://vhs.charm.sh/vhs-4EIxjwVxC1iHWA8htLBIja.gif" alt="Demo" height="400">

**[Build Trees From Parent Relationships](examples/basic/06-builders-flat/)**  
Shows how to use BuildTree to create trees from flat data structures where relationships are defined by parent ID references.

<img src="https://vhs.charm.sh/vhs-7BrzyAjGy08PWxZrmY4dVb.gif" alt="Demo" height="400">

**[Build Trees From Hierarchical Data](examples/basic/07-builders-nested/)**  
Demonstrates using BuildTree for constructing trees from naturally hierarchical data structures.
  
<img src="https://vhs.charm.sh/vhs-7zo22mFV7SOdiyVCQhemof.gif" alt="Demo" height="400">

**[Filesystem Builder](examples/basic/08-builders-filesystem/)**  
Demonstrates creating filesystem tree structures with NewTreeFromFileSystem constructor with support for special path characters like `.`, `..`, and `~`. Shows path resolution, configurable depth limits, hidden file handling, and automatic file type detection for styling.
  
<img src="https://vhs.charm.sh/vhs-5tyMF6uJvLTqLpDaa75oIr.gif" alt="Demo" height="600">

---
### Intermediate Examples

Intermediate examples demonstrate treeview's full TUI support via BubbleTea.

**[Keyboard Controls](examples/intermediate/01-keyboard-controls/)**  
Advanced interactive example demonstrating comprehensive keyboard-based navigation within tree
structures. Shows how to implement arrow key navigation, selection management, and basic searching.
  
<img src="https://vhs.charm.sh/vhs-1yQdtGnU6NSeAlzIyLTp7Q.gif" alt="Demo" height="400">

**[Search & Filtering](examples/intermediate/02-search/)**  
This example demonstrates extending the search functionality to handle domain data.
  
<img src="https://vhs.charm.sh/vhs-X9Ny1urVxOcAxadiO9lLY.gif" alt="Demo" height="400">

**[Viewport Rendering](examples/intermediate/03-viewport/)**  
Demonstrates viewport-aware rendering for handling large tree structures\. Solves the common problem of cursor navigation going off-screen in large directory trees by implementing automatic scrolling and responsive viewport management.
  
<img src="https://vhs.charm.sh/vhs-38hvhngDH75MGBhVPcRDBD.gif" alt="Demo" height="400">

**[File Browser](examples/intermediate/04-file-browser/)**  
Complete, terminal file browser application built with TreeView and Bubble Tea.

<img src="https://vhs.charm.sh/vhs-5XfivVFWrVQQODOA3d6ysd.gif" alt="Demo" height="400">

## Contributing

Contributions are welcome! If you have any suggestions or encounter a bug, please open an
[issue](https://github.com/Digital-Shane/treeview/issues) or submit a pull request.

When contributing:

1. Fork the repository and create a new feature branch
2. Make your changes in a well-structured commit history
3. Include tests (when applicable)
4. Submit a pull request with a clear description of your changes

## License

This project is licensed under the GNU Version 3 - see the [LICENSE](./LICENSE) file for details.

## Acknowledgments

TreeView is built on top of the excellent [Charm](https://charm.sh) ecosystem:
- [Bubble Tea][bt] for TUI framework
- [Lipgloss][lg] for styling and layout
- [VHS](https://github.com/charmbracelet/vhs) for epic scriptable demos
- [Charm](https://github.com/charmbracelet) community for inspiration and gif hosting!

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=Digital-Shane/treeview&type=Date)](https://www.star-history.com/#Digital-Shane/treeview&Date)

---

[lg]: https://github.com/charmbracelet/lipgloss
[bt]: https://github.com/charmbracelet/bubbletea
