package treeview

import (
	"context"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/mattn/go-runewidth"
)

var sbPool = sync.Pool{New: func() any { return new(strings.Builder) }}

// renderNode implements the NodeRenderer interface. It asks the NodeProvider
// for icon, label, and style information, then returns the final string for a
// single line including tree-branch glyphs.
//
// The function is fast and does not allocate beyond what the provider allocates.
func renderNode[T any](provider NodeProvider[T], node *Node[T], prefix string, isFocused bool) (string, error) {
	// Get the icon from the provider and ensure consistent width
	// This keeps the tree aligned even with different icon widths
	icon := NormalizeIconWidth(provider.Icon(node))

	// Get the human-readable text for this node
	displayText := provider.Format(node)

	// Get the appropriate style based on focus state
	style := provider.Style(node, isFocused)

	// Combine all parts and apply the style
	// Result: "│   └── 📁 folder-name/" (styled)
	return style.Render(prefix + icon + displayText), nil
}

// renderTree walks the tree, turns every visible node into a line.
func renderTree[T any](ctx context.Context, tree *Tree[T]) (string, int, error) {
	// Get a string builder from the pool for efficiency
	sb := sbPool.Get().(*strings.Builder)
	defer func() {
		sb.Reset()
		sbPool.Put(sb)
	}()

	// Track state for single-pass rendering
	lineIdx := 0
	focusedLineIndex := -1

	// ancestorIsLastChild tracks whether each ancestor (at each depth level) was the last
	// child among its siblings. This determines whether we draw a vertical continuation
	// line (│) or a space when building the tree prefix.
	//
	// For example, with this tree:
	//   ├── folder1        <- Not last child, so children get "│   " prefix
	//   │   ├── file1.txt  <- These lines connect back to folder1
	//   └───└── file2.txt  <- No vertical line because parent was last
	//
	// The slice index corresponds to the depth level
	var ancestorIsLastChild []bool

	for info, err := range tree.AllVisible(ctx) {
		if err != nil {
			return "", 0, err
		}
		node := info.Node
		depth := info.Depth
		isLast := info.IsLast

		// Update our tracking of which ancestors are last children at each depth.
		if depth >= len(ancestorIsLastChild) {
			ancestorIsLastChild = append(ancestorIsLastChild, isLast)
		} else {
			// At same depth or going shallower. Update current depth and trim deeper levels.
			ancestorIsLastChild[depth] = isLast
			ancestorIsLastChild = ancestorIsLastChild[:depth+1]
		}

		// Build the tree branch prefix based on ancestor positions
		// Root nodes (depth 0) get no prefix at all
		var prefix string
		if depth > 0 {
			prefix = buildPrefix(ancestorIsLastChild[:depth], isLast)
		}

		// Check if this node should be highlighted as focused
		isFocused := tree.IsFocused(node.ID())
		if isFocused && focusedLineIndex == -1 {
			// Set focused line index to the first focused node for viewport positioning
			focusedLineIndex = lineIdx
		}

		// Render the actual node content
		line, err := renderNode(tree.provider, node, prefix, isFocused)
		if err != nil {
			return sb.String(), focusedLineIndex, err
		}

		// Add newline before each line except the first
		if lineIdx > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(line)
		lineIdx++

		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			return sb.String(), focusedLineIndex, err
		}
	}

	return sb.String(), focusedLineIndex, nil
}

// renderTreeWithViewport combines tree rendering with viewport scrolling.
// It automatically positions the viewport to keep the focused line visible.
func renderTreeWithViewport[T any](ctx context.Context, tree *Tree[T], vp *viewport.Model) (string, error) {
	content, focusedLineIndex, err := renderTree(ctx, tree)
	// Error can't impact the viewport, so we ignore it during vp setup,
	// but we return it so callers can handle it if they want.
	vp.SetContent(content)

	// Auto-scroll to keep focused line visible
	if focusedLineIndex >= 0 && vp.Height > 0 {
		// If focused line is above viewport, scroll up
		if focusedLineIndex < vp.YOffset {
			vp.YOffset = focusedLineIndex
		} else if focusedLineIndex >= vp.YOffset+vp.Height {
			// If focused line is below viewport, scroll down
			// Keep one line of context if possible
			vp.YOffset = focusedLineIndex - vp.Height + 1
			vp.YOffset = max(vp.YOffset, 0)
		}
	}

	// Return the viewport's view of the content
	// This shows only the visible portion based on scroll position
	return vp.View(), err
}

// buildPrefix constructs the complete tree branch prefix string that visually connects
// nodes to their ancestors. This function should only be called for non-root nodes (depth > 0).
// Each boolean in ancestorIsLastChild represents whether an ancestor at that depth was the
// last among its siblings. The isLast parameter indicates whether the current node is the
// last among its siblings.
//
// Examples:
//
//	ancestorIsLastChild = [],             isLast = false → "├── "         (first level child, has siblings)
//	ancestorIsLastChild = [],             isLast = true  → "└── "         (first level child, last sibling)
//	ancestorIsLastChild = [false],        isLast = false → "│   ├── "     (parent and node have siblings)
//	ancestorIsLastChild = [false],        isLast = true  → "│   └── "     (parent has siblings, node is last)
//	ancestorIsLastChild = [true],         isLast = false → "    ├── "     (parent was last, node has siblings)
//	ancestorIsLastChild = [true],         isLast = true  → "    └── "     (parent was last, node is last)
//	ancestorIsLastChild = [false,  true], isLast = true  → "│       └── " (complex nesting)
//
// This creates the complete visual tree structure including vertical lines and branch characters.
func buildPrefix(ancestorIsLastChild []bool, isLast bool) string {
	var prefixBuilder strings.Builder

	// Add vertical lines for ancestors
	for _, isLastChild := range ancestorIsLastChild {
		if isLastChild {
			// Parent was last child
			prefixBuilder.WriteString("    ")
		} else {
			// Parent has more siblings
			prefixBuilder.WriteString("│   ")
		}
	}

	// Add the final branch character
	if isLast {
		// Last child gets └── branch
		prefixBuilder.WriteString("└── ")
	} else {
		// Other children get ├── branch
		prefixBuilder.WriteString("├── ")
	}

	return prefixBuilder.String()
}

// NormalizeIconWidth pads or trims the icon so that the combined width of icon
// plus trailing space is at least targetWidth runes. This keeps labels neatly
// aligned under each other.
//
// This is called during the rendering process, but it is more efficient to
// call it once for each icon before storing it in your provider to cache the result.
// No string concatenation is performed if the icon is already the correct width.
func NormalizeIconWidth(icon string) string {
	if icon == "" {
		return ""
	}

	const targetWidth = 3
	width := runewidth.StringWidth(icon)

	if width >= targetWidth {
		if strings.HasSuffix(icon, " ") {
			return icon
		}
		return icon + " "
	}

	return icon + strings.Repeat(" ", targetWidth-width)
}
