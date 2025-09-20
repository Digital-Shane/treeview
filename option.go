package treeview

import (
	"context"
	"slices"
	"strings"
)

// SearchFn returns true if the node matches the search term.
type SearchFn[T any] func(ctx context.Context, node *Node[T], term string) bool

// FilterFn returns true if the item should be included in the tree.
type FilterFn[T any] func(item T) bool

// ExpandFn returns true if the node should be expanded during the build process.
type ExpandFn[T any] func(node *Node[T]) bool

// ProgressCallback is invoked during tree construction every time a node is created.
type ProgressCallback[T any] func(processed int, node *Node[T])

// FocusPolicyFn selects the next node to focus when the user moves up/down the
// list. The offset is usually ±1 but can be any integer.
type FocusPolicyFn[T any] func(ctx context.Context, visible []*Node[T], current *Node[T], offset int) (*Node[T], error)

// Option is the unified functional Option type used by all tree constructors.
// It allows callers to provide build-time and run-time configurations in a
// single, flat list.
type Option[T any] func(*MasterConfig[T])

// WithProvider specifies the NodeProvider to use for rendering nodes.
// The provider controls how nodes are formatted, styled, and displayed.
func WithProvider[T any](p NodeProvider[T]) Option[T] {
	return func(cfg *MasterConfig[T]) {
		cfg.provider = p
	}
}

// WithSearcher overwrites the algorithm used when the search feature is invoked.
func WithSearcher[T any](fn SearchFn[T]) Option[T] {
	return func(cfg *MasterConfig[T]) {
		cfg.searcher = fn
	}
}

// WithFocusPolicy replaces the logic that decides which node should be focused
// after search or navigation.
func WithFocusPolicy[T any](fn FocusPolicyFn[T]) Option[T] {
	return func(cfg *MasterConfig[T]) {
		cfg.focusPol = fn
	}
}

// WithExpandFunc installs a predicate that decides for each node whether it
// should start expanded.
func WithExpandFunc[T any](fn ExpandFn[T]) Option[T] {
	return func(cfg *MasterConfig[T]) {
		cfg.expandFunc = fn
	}
}

// WithExpandAll expands all nodes by default.
func WithExpandAll[T any]() Option[T] {
	return func(cfg *MasterConfig[T]) {
		cfg.expandFunc = func(*Node[T]) bool { return true }
	}
}

// WithFilterFunc installs a predicate that decides for each node whether it
// should be included in the tree.
func WithFilterFunc[T any](filter FilterFn[T]) Option[T] {
	return func(cfg *MasterConfig[T]) {
		cfg.filterFunc = filter
	}
}

// WithMaxDepth limits how deep the walker descends into directories or other
// data structures. Use a negative depth for no limit (default).
func WithMaxDepth[T any](d int) Option[T] {
	return func(c *MasterConfig[T]) {
		c.maxDepth = d
	}
}

// WithTraversalCap sets an upper bound on the total number of entries visited
// during tree construction. A value ≤ 0 is interpreted as no limit.
func WithTraversalCap[T any](cap int) Option[T] {
	return func(c *MasterConfig[T]) {
		c.traversalCap = cap
	}
}

// WithProgressCallback registers a callback that is invoked each time a new
// node is created during any of the constructor build phases. It is not
// invoked by NewTree (which accepts pre-built nodes) except once per root
// node supplied so callers have a consistent hook.
func WithProgressCallback[T any](cb ProgressCallback[T]) Option[T] {
	return func(c *MasterConfig[T]) {
		c.progressCb = cb
	}
}

// WithTruncate sets the maximum width for rendered lines. Lines longer than
// this width will be truncated with an ellipsis. A width of 0 disables
// truncation (default).
func WithTruncate[T any](width int) Option[T] {
	return func(c *MasterConfig[T]) {
		c.truncateWidth = width
	}
}

// MasterConfig is the structure that aggregates options from
// different domains (build, filesystem, tree). It is used by the unified
// constructors to collect and dispatch options to the appropriate internal
// functions.
type MasterConfig[T any] struct {
	// Options used during the build process.
	maxDepth     int
	traversalCap int
	expandFunc   ExpandFn[T]         // If the function returns true, the node is expanded immediately during the build process.
	filterFunc   FilterFn[T]         // If the function returns true, the node is included in the tree.
	progressCb   ProgressCallback[T] // Optional progress reporting during construction.

	// Options passed to the final tree.
	searcher      SearchFn[T]
	focusPol      FocusPolicyFn[T]
	provider      NodeProvider[T]
	truncateWidth int // Maximum width for rendered lines (0 = no truncation)
}

// NewMasterConfig is a helper that creates a MasterConfig, applies defaults, and then user-provided options.
func NewMasterConfig[T any](opts []Option[T], defaults ...Option[T]) *MasterConfig[T] {
	cfg := &MasterConfig[T]{
		searcher:     defaultSearchFn[T],
		focusPol:     defaultFocusPolicy[T],
		provider:     NewDefaultNodeProvider[T](),
		expandFunc:   nil,
		filterFunc:   nil,
		maxDepth:     -1,
		traversalCap: 10000,
		progressCb:   nil,
	}

	// Apply provided defaults first
	for _, opt := range defaults {
		if opt != nil {
			opt(cfg)
		}
	}

	// Apply user-provided options.
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	return cfg
}

// ShouldFilter evaluates whether the given item should be excluded based on the configured filter function.
func (cfg *MasterConfig[T]) ShouldFilter(item T) bool {
	if cfg.filterFunc == nil {
		return false
	}
	return !cfg.filterFunc(item)
}

// HandleExpansion checks if a node should be expanded based on the configured expand function and expands it if true.
func (cfg *MasterConfig[T]) HandleExpansion(node *Node[T]) {
	if cfg.expandFunc == nil {
		return
	}
	if cfg.expandFunc(node) {
		node.Expand()
	}
}

// HasTraversalCapBeenReached checks if the current node count has reached or exceeded the configured traversal cap.
func (cfg *MasterConfig[T]) HasTraversalCapBeenReached(nodeCount int) bool {
	if cfg.traversalCap <= 0 {
		return false
	}
	return nodeCount >= cfg.traversalCap
}

// HasDepthLimitBeenReached checks if the given current depth has reached or exceeded the configured maximum depth limit.
func (cfg *MasterConfig[T]) HasDepthLimitBeenReached(currentDepth int) bool {
	if cfg.maxDepth <= 0 {
		return false
	}
	return currentDepth >= cfg.maxDepth
}

// ReportProgress invokes the configured progress callback (if any).
func (cfg *MasterConfig[T]) ReportProgress(processed int, node *Node[T]) {
	if cfg.progressCb == nil || node == nil {
		return
	}
	cfg.progressCb(processed, node)
}

// defaultSearchFn implements a simple case-insensitive search strategy that matches any substring
// in the node's ID, name, or data (if it's a string or has a String() method).
func defaultSearchFn[T any](_ context.Context, node *Node[T], term string) bool {
	// Empty term means no matches
	if term == "" {
		return false
	}

	// First, prepare the search comparison values
	// We'll check both the node's ID and Name definitely
	fields := []string{
		strings.ToLower(node.ID()),
		strings.ToLower(node.Name()),
		"",
		"",
	}
	// Then the data if it's a string or has a String() method.
	if nodeData, ok := any(*node.Data()).(string); ok {
		fields[2] = strings.ToLower(nodeData)
	}
	if nodeData, ok := any(*node.Data()).(interface{ String() string }); ok {
		fields[3] = strings.ToLower(nodeData.String())
	}

	searchTerm := strings.ToLower(term)

	// Substring match of any field
	if slices.ContainsFunc(fields, func(field string) bool {
		return strings.Contains(field, searchTerm)
	}) {
		return true
	}

	return false
}

// defaultFocusPolicy implements a simple focus navigation strategy.
// It moves through the visible nodes linearly, wrapping at boundaries.
func defaultFocusPolicy[T any](_ context.Context, visible []*Node[T], current *Node[T], offset int) (*Node[T], error) {
	// Handle empty list
	if len(visible) == 0 {
		return nil, nil
	}

	// If no current focus, start at the beginning for positive offset
	// or end for negative offset
	if current == nil {
		if offset > 0 {
			return visible[0], nil
		}
		return visible[len(visible)-1], nil
	}

	// Find current position
	currentIdx := -1
	for i, node := range visible {
		if node == current {
			currentIdx = i
			break
		}
	}

	// If current node not in visible list, start from beginning/end
	if currentIdx == -1 {
		if offset > 0 {
			return visible[0], nil
		}
		return visible[len(visible)-1], nil
	}

	// Calculate new position with wrapping
	newIdx := currentIdx + offset
	if newIdx < 0 {
		newIdx = len(visible) - 1 // Wrap to end
	} else if newIdx >= len(visible) {
		newIdx = 0 // Wrap to beginning
	}

	return visible[newIdx], nil
}
