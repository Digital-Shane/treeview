package treeview

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"charm.land/bubbles/v2/viewport"
	"github.com/mattn/go-runewidth"
)

var sbPool = sync.Pool{New: func() any { return new(strings.Builder) }}

// ansiRegex matches ANSI escape sequences
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// stripANSI removes ANSI escape sequences from a string
func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// visualWidth calculates the display width of a string, accounting for ANSI codes
func visualWidth(s string) int {
	// Remove ANSI codes before calculating width
	plain := stripANSI(s)
	return runewidth.StringWidth(plain)
}

// truncateLine truncates a line to fit within maxWidth, adding ellipsis if needed
func truncateLine(line string, maxWidth int) string {
	if maxWidth <= 0 {
		return line
	}

	width := visualWidth(line)
	if width <= maxWidth {
		return line
	}

	// We need to truncate. Reserve 3 characters for ellipsis
	if maxWidth <= 3 {
		return "..."
	}

	targetWidth := maxWidth - 3

	// Find ANSI sequences to preserve styling (work with byte indices)
	ansiParts := ansiRegex.FindAllStringIndex(line, -1)

	// Build the truncated string byte by byte
	var result strings.Builder
	currentWidth := 0
	bytes := []byte(line)
	ansiIdx := 0
	i := 0

	for i < len(bytes) {
		// Check if we're at an ANSI sequence
		if ansiIdx < len(ansiParts) && i == ansiParts[ansiIdx][0] {
			// Add the ANSI sequence (doesn't affect width)
			end := ansiParts[ansiIdx][1]
			result.Write(bytes[i:end])
			i = end
			ansiIdx++
			continue
		}

		// Decode the next rune
		r, size := utf8.DecodeRune(bytes[i:])
		charWidth := runewidth.RuneWidth(r)

		if currentWidth+charWidth > targetWidth {
			// Stop here and add ellipsis
			break
		}

		result.Write(bytes[i : i+size])
		currentWidth += charWidth
		i += size
	}

	// Add ellipsis
	result.WriteString("...")

	// Add any remaining ANSI reset codes if they exist at the end of the original string
	if strings.HasSuffix(line, "\x1b[0m") {
		result.WriteString("\x1b[0m")
	}

	return result.String()
}

// renderNode implements the NodeRenderer interface. It asks the NodeProvider
// for icon, label, and style information, then returns the final string for a
// single line including tree-branch glyphs.
//
// The function is fast and does not allocate beyond what the provider allocates.
func renderNode[T any](provider NodeProvider[T], node *Node[T], prefix string, isFocused bool, maxWidth int) (string, error) {
	// Get the icon from the provider and ensure consistent width
	// This keeps the tree aligned even with different icon widths
	icon := NormalizeIconWidth(provider.Icon(node))

	// Get the human-readable text for this node
	displayText := provider.Format(node)

	// Get the appropriate style based on focus state
	style := provider.Style(node, isFocused)

	// Combine all parts and apply the style
	// Result: "│   └── 📁 folder-name/" (styled)
	result := style.Render(prefix + icon + displayText)

	// Apply truncation if maxWidth is set
	return truncateLine(result, maxWidth), nil
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
		line, err := renderNode(tree.provider, node, prefix, isFocused, tree.truncateWidth)
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
	// First, find the focused line position to determine if we need to adjust the viewport
	focusedLineIndex := findFocusedLineIndex(ctx, tree)

	// Auto-scroll to keep focused line visible BEFORE rendering
	if focusedLineIndex >= 0 && vp.Height() > 0 {
		// If focused line is above viewport, scroll up
		if focusedLineIndex < vp.YOffset() {
			vp.SetYOffset(focusedLineIndex)
		} else if focusedLineIndex >= vp.YOffset()+vp.Height() {
			// If focused line is below viewport, scroll down
			// Keep one line of context if possible
			offset := focusedLineIndex - vp.Height() + 1
			offset = max(offset, 0)
			vp.SetYOffset(offset)
		}
	}

	// Now render only the visible portion with the correct viewport offset
	content, totalLines, err := renderViewportOnly(ctx, tree, vp)

	// Update viewport's understanding of total content for scrollbar
	// We use empty lines to set the height without the memory cost of actual content
	vp.SetContent(strings.Repeat("\n", max(0, totalLines-1)))

	// Return just the visible content
	return content, err
}

// findFocusedLineIndex quickly scans through the tree to find the focused line's position.
// This is a lightweight operation that doesn't render anything.
func findFocusedLineIndex[T any](ctx context.Context, tree *Tree[T]) int {
	lineIdx := 0
	for info, err := range tree.AllVisible(ctx) {
		if err != nil {
			return -1
		}
		if tree.IsFocused(info.Node.ID()) {
			return lineIdx
		}
		lineIdx++

		// Check for context cancellation periodically
		if lineIdx%100 == 0 {
			if ctx.Err() != nil {
				return -1
			}
		}
	}
	return -1
}

// renderViewportOnly efficiently renders only the visible lines in the viewport
// in a single pass through the tree. Returns the rendered content, total line count, and any error.
func renderViewportOnly[T any](ctx context.Context, tree *Tree[T], vp *viewport.Model) (string, int, error) {
	// Get a string builder from the pool for efficiency
	sb := sbPool.Get().(*strings.Builder)
	defer func() {
		sb.Reset()
		sbPool.Put(sb)
	}()

	// Calculate the range of lines we need to render
	startLine := vp.YOffset()
	endLine := vp.YOffset() + vp.Height()

	// Track state for single-pass rendering
	currentLine := 0
	renderBuffer := make([]string, 0, vp.Height()) // Pre-allocate for viewport height

	// ancestorIsLastChild tracks whether each ancestor (at each depth level) was the last
	// child among its siblings. This determines whether we draw a vertical continuation
	// line (│) or a space when building the tree prefix.
	var ancestorIsLastChild []bool

	for info, err := range tree.AllVisible(ctx) {
		if err != nil {
			return "", currentLine, err
		}

		node := info.Node
		depth := info.Depth
		isLast := info.IsLast

		// Update our tracking of which ancestors are last children at each depth.
		// We need to maintain this for all nodes to get correct prefixes
		if depth >= len(ancestorIsLastChild) {
			ancestorIsLastChild = append(ancestorIsLastChild, isLast)
		} else {
			// At same depth or going shallower. Update current depth and trim deeper levels.
			ancestorIsLastChild[depth] = isLast
			ancestorIsLastChild = ancestorIsLastChild[:depth+1]
		}

		// Only render lines that will be visible in the viewport
		if currentLine >= startLine && currentLine < endLine {
			// Build the tree branch prefix based on ancestor positions
			// Root nodes (depth 0) get no prefix at all
			var prefix string
			if depth > 0 {
				prefix = buildPrefix(ancestorIsLastChild[:depth], isLast)
			}

			// Check if this node is focused
			isFocused := tree.IsFocused(node.ID())

			// Render the actual node content
			line, err := renderNode(tree.provider, node, prefix, isFocused, tree.truncateWidth)
			if err != nil {
				return sb.String(), currentLine, err
			}
			renderBuffer = append(renderBuffer, line)
		}

		currentLine++

		if err := ctx.Err(); err != nil {
			return sb.String(), currentLine, err
		}
	}

	// Join the visible lines
	for i, line := range renderBuffer {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(line)
	}

	return sb.String(), currentLine, nil
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
