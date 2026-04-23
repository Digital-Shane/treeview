# Updating to TreeView v2

TreeView v2 is mostly a cleanup release, but there are a few changes worth
paying attention to before you bump the version in a real project. If you use
the standard constructors and the exported `With...` helpers, the upgrade will
probably be pretty small. If you built a custom extension or relied on internal
construction details from the root package, this is the release where you will
need to touch code.

The short version is that v2 moves TreeView onto the Charm v2 ecosystem (Thanks darkhz!),
adds a small compatibility option for TUI apps, cleans up the extension story
around a new walker-based API, and fixes a viewport scrolling bug that some
apps may have worked around manually.

## Charm v2

TreeView now uses the Charm v2 packages. If your project imports Bubble Tea,
Bubbles, or Lip Gloss directly alongside TreeView, you should update those
imports and versions at the same time so everything is on the same stack.

This release also switches to the proper Go semantic import versioning path, so
the module is now imported as `github.com/Digital-Shane/treeview/v2`. If you
use the S3 extension module, that import path also picks up the version suffix
and becomes `github.com/Digital-Shane/treeview/extensions/s3/v2`.

```go
import "github.com/Digital-Shane/treeview/v2"
import tea "charm.land/bubbletea/v2"
import "charm.land/bubbles/v2/viewport"
import "charm.land/lipgloss/v2"
```

If you only use TreeView as a library and do not import the Charm packages
yourself, this part may not require much more than updating your dependency and
making sure your local toolchain and CI are running Go 1.25.

## TUI behavior and `WithTuiAltScreen`

v2 adds `WithTuiAltScreen(true)` to restore the older alt screen behavior. If
your TUI felt right before and still feels right after the upgrade, you can
ignore this completely. If your app suddenly has the terminal command used to
launch your app persisting at the top add:

```go
model := treeview.NewTuiTreeModel(
	tree,
	treeview.WithTuiAltScreen(true),
)
```

## The extension system

This is the biggest API change in v2. Earlier extension support worked by
exposing a few things from the root package that were really supposed to stay
internal. That made the S3 extension possible, but it also meant external code
could depend on internal tree construction details. It worked, but it was not a
good long-term pattern.

v2 fixes that by introducing a real extension seam built around `Walker`,
`WalkItem`, and `NewTreeFromWalker`. Instead of recreating tree construction in
an extension, you now provide a root item and the children for a given item, and
TreeView handles the shared behavior like filtering, max depth, traversal caps,
expansion, progress callbacks, and final tree assembly.

```go
type Walker[T any] interface {
	Root(context.Context) (treeview.WalkItem[T], error)
	Children(context.Context, treeview.WalkItem[T]) ([]treeview.WalkItem[T], error)
}
```

If you were using internal root types like `MasterConfig`, `NewMasterConfig`,
or `NewTreeFromCfg`, the important change is those path is gone. The new
path is `NewTreeFromWalker`.

One practical note here: walker-based filtering happens before descent. If a
parent item is filtered out, its descendants are pruned with it. For most data
sources that is exactly what you want, but it is worth knowing if your older
extension logic assumed it could still descend through filtered parents.

## `Option` still works, but the internals do not leak anymore

`Option` is still the public way to configure TreeView constructors, so normal
app code using `WithProvider`, `WithFilterFunc`, `WithExpandFunc`,
`WithMaxDepth`, `WithTraversalCap`, or `WithProgressCallback` should feel the
same. What changed is that extension support no longer depends on exposing the
internal config types that power those options.

In other words, if you were just passing TreeView options around, you are fine.
If you were building extension code that depended on root-package config
internals, move that logic into your walker or constructor and keep TreeView
configuration on the public `With...` side.

The filesystem constructor was updated to use the same walker-based path that
extensions use, and the S3 extension was moved over to that model too. 

## Viewport autoscroll was fixed

v2 also fixes a viewport autoscroll issue where the scroll position could clamp
back to the top because the content length was being set after the scroll
offset. If you added a workaround for odd viewport jumps or the tree snapping back to
the top during navigation, test your app without that workaround first. There is
a good chance you can delete code now.
