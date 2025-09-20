package treeview

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Digital-Shane/treeview/internal/utils"
)

// NewTree creates a new Tree from an existing slice of nodes.
//
// Supported options:
//   - WithFilterFunc:  Filters nodes recursively, keeping parents with matching children
//   - WithMaxDepth:    Limits tree depth (0 = root only, 1 = root + children, etc.)
//   - WithExpandFunc:  Sets initial expansion state for nodes
//   - WithProgressCallback: Reports progress for each provided root node
//   - WithSearcher:    Custom search algorithm
//   - WithFocusPolicy: Custom focus navigation logic
//   - WithProvider:    Custom node rendering provider
//
// Note: WithTraversalCap is not respected at this stage.
//
// Note: WithFilterFunc, WithMaxDepth, and WithExpandFunc are provided for
// convenience, but they are not efficient. It is better to use the filter
// functions provided by the other constructors to filter, limit, and
// expand the tree during construction.
func NewTree[T any](nodes []*Node[T], opts ...Option[T]) *Tree[T] {
	cfg := NewMasterConfig(opts)

	// Report progress for user-supplied nodes so the callback can still be
	// leveraged even when callers pre-assemble the slice.
	if cfg.progressCb != nil {
		for i, n := range nodes {
			cfg.ReportProgress(i+1, n)
		}
	}

	// Apply filtering if configured
	if cfg.filterFunc != nil {
		nodes = filterNodes(nodes, cfg.filterFunc)
	}

	// Apply depth limiting if configured
	if cfg.maxDepth >= 0 {
		nodes = limitDepth(nodes, cfg.maxDepth, 0)
	}

	// Apply expansion function if configured
	if cfg.expandFunc != nil {
		applyExpansion(nodes, cfg)
	}

	return NewTreeFromCfg(nodes, cfg)
}

// NewTreeFromCfg creates a new Tree with the provided nodes and the configuration `cfg`.
func NewTreeFromCfg[T any](nodes []*Node[T], cfg *MasterConfig[T]) *Tree[T] {
	// Initialize focus to the first node if available
	var focusedNodes []*Node[T]
	focusedIDs := make(map[string]bool)
	if len(nodes) > 0 {
		focusedNodes = []*Node[T]{nodes[0]}
		focusedIDs[nodes[0].ID()] = true
	}

	// Create the tree with default components
	t := &Tree[T]{
		nodes:         nodes,
		focusedNodes:  focusedNodes,
		focusedIDs:    focusedIDs,
		searcher:      cfg.searcher,
		focusPol:      cfg.focusPol,
		provider:      cfg.provider,
		truncateWidth: cfg.truncateWidth,
	}
	return t
}

// NestedDataProvider is the counterpart for
// BuildTreeFromNestedData. It exposes just the fundamental
// accessors needed to traverse nested data structures.
//
//   - ID(T) string    unique identifier for the item
//   - Name(T) string  display label for the node
//   - Children(T) []T direct descendants of the item
//
// Optional construction behavior is handled separately via Options.
type NestedDataProvider[T any] interface {
	ID(T) string
	Name(T) string
	Children(T) []T
}

// NewTreeFromNestedData builds a Tree from a data source where items are already
// structured in a parent-child hierarchy. The provider interface tells the
// builder how to access the ID, name, and children for each item.
// Returns context errors unwrapped.
//
// Example:
//
//	tree, err := treeview.NewTreeFromNestedData(
//	    ctx,
//	    menuItems,
//	    &MenuProvider{},
//	    treeview.WithExpandFunc(expandTopLevel),
//	    treeview.WithMaxDepth(3),
//	)
//
// Supported options:
// Build options:
//   - WithFilterFunc:   Filters items during tree building
//   - WithMaxDepth:     Limits tree depth during construction
//   - WithExpandFunc:   Sets initial expansion state for nodes
//   - WithTraversalCap: Limits total nodes processed (returns partial tree + error if exceeded)
//   - WithProgressCallback: Invoked after each node creation (depth-first)
//
// Options used during a tree's runtime:
//   - WithSearcher:     Custom search algorithm
//   - WithFocusPolicy:  Custom focus navigation logic
//   - WithProvider:     Custom node rendering provider
func NewTreeFromNestedData[T any](
	ctx context.Context,
	items []T,
	provider NestedDataProvider[T],
	opts ...Option[T],
) (*Tree[T], error) {
	cfg := NewMasterConfig(opts)

	// Build the node hierarchy using the collected build options.
	nodes, err := buildTreeFromNestedData(ctx, items, provider, cfg)

	// Create the final tree
	if err != nil && err != ErrTraversalLimit {
		err = fmt.Errorf("%w: %w", ErrTreeConstruction, err)
	}

	tree := NewTreeFromCfg(nodes, cfg)
	return tree, err
}

func buildTreeFromNestedData[T any](ctx context.Context, items []T, provider NestedDataProvider[T], cfg *MasterConfig[T]) ([]*Node[T], error) {
	nodeCount := 0
	hitTraversalCap := false

	// Helper function to recursively convert an item (and its descendants) into
	// *Node values, wiring up parent / child relationships on the fly.
	var buildSubtree func(context.Context, T, int) (*Node[T], error)
	buildSubtree = func(ctx context.Context, item T, depth int) (*Node[T], error) {
		if err := ctx.Err(); err != nil {
			return nil, err // Context has ended
		}
		if cfg.ShouldFilter(item) {
			return nil, nil // Item was filtered out
		}
		if cfg.HasTraversalCapBeenReached(nodeCount) {
			hitTraversalCap = true // We've hit the traversal cap
			return nil, nil
		}

		// Create node for this item using provider
		id := provider.ID(item)
		if id == "" {
			return nil, ErrEmptyID
		}
		n := NewNode(id, provider.Name(item), item)

		// Increment node count & report progress
		nodeCount++
		cfg.ReportProgress(nodeCount, n)

		// Check depth limit
		if cfg.HasDepthLimitBeenReached(depth) {
			cfg.HandleExpansion(n)
			// Return the node without children.
			return n, nil
		}

		// Recursively convert children, if any.
		rawChildren := provider.Children(item)
		for _, childItem := range rawChildren {
			childNode, err := buildSubtree(ctx, childItem, depth+1)
			if err != nil {
				return n, err
			}
			if childNode != nil {
				n.AddChild(childNode)
			}
		}

		cfg.HandleExpansion(n)
		return n, nil
	}

	// Initialize the recursive build of the tree
	roots := make([]*Node[T], 0, len(items))
	for _, item := range items {
		root, err := buildSubtree(ctx, item, 0)
		if err != nil {
			return nil, err
		}
		if root != nil {
			roots = append(roots, root)
		}
		// Stop processing more root items if we hit the cap
		if hitTraversalCap {
			return roots, ErrTraversalLimit
		}
	}

	return roots, nil
}

// FlatDataProvider is the counterpart for
// BuildTreeFromFlatData. It exposes just the fundamental
// accessors needed to build nested data structures from flat data.
//
//   - ID(T) string       unique identifier for the item
//   - Name(T) string     display label for the node
//   - ParentID(T) string identifier of the item's parent ("" for roots)
//
// Optional construction behavior is handled separately via Options.
type FlatDataProvider[T any] interface {
	ID(T) string
	Name(T) string
	ParentID(T) string
}

// NewTreeFromFlatData builds a Tree from a flat list of items, where each item
// references its parent by ID. This is useful for data from databases or APIs.
// Returns context errors unwrapped.
//
// Example:
//
//	tree, err := treeview.NewTreeFromFlatData(
//	    ctx,
//	    employees,
//	    &EmployeeProvider{},
//	    treeview.WithExpandFunc(expandManagers),
//	    treeview.WithFilterFunc(filterActiveEmployees),
//	)
//
// Supported options:
// Build options:
//   - WithFilterFunc:   Filters items during tree building
//   - WithMaxDepth:     Limits tree depth after hierarchy is built
//   - WithExpandFunc:   Sets initial expansion state for nodes
//   - WithTraversalCap: Limits total nodes processed (returns partial tree + error if exceeded)
//   - WithProgressCallback: Invoked after each node creation (flat pass order)
//
// Options used during a tree's runtime:
//   - WithSearcher:     Custom search algorithm
//   - WithFocusPolicy:  Custom focus navigation logic
//   - WithProvider:     Custom node rendering provider
func NewTreeFromFlatData[T any](
	ctx context.Context,
	items []T,
	provider FlatDataProvider[T],
	opts ...Option[T],
) (*Tree[T], error) {
	// 1. Create config from options.
	cfg := NewMasterConfig(opts)

	// 2. Build the node hierarchy.
	nodes, err := buildTreeFromFlatData(ctx, items, provider, cfg)

	// 3. Create the final tree
	if err != nil && err != ErrTraversalLimit {
		err = fmt.Errorf("%w: %w", ErrTreeConstruction, err)
	}

	tree := NewTreeFromCfg(nodes, cfg)
	return tree, err
}

func buildTreeFromFlatData[T any](ctx context.Context, items []T, provider FlatDataProvider[T], cfg *MasterConfig[T]) ([]*Node[T], error) {
	// Pass 1: Convert all raw items to *Node values, and cache relationship map

	// parentLookup maps a node ID to the ID of its parent, so we can wire the
	// hierarchy in a second pass once all nodes have been created.
	parentLookup := make(map[string]string, len(items))
	// idToNode lets us resolve a parent ID to the corresponding *Node.
	idToNode := make(map[string]*Node[T], len(items))

	// Track if we hit the traversal cap
	nodeCount := 0
	hitTraversalCap := false

	for _, item := range items {
		if err := ctx.Err(); err != nil {
			return nil, err // Context has ended
		}
		if cfg.ShouldFilter(item) {
			return nil, nil // Item was filtered out
		}
		if cfg.HasTraversalCapBeenReached(nodeCount) {
			hitTraversalCap = true // We've hit the traversal cap
			break
		}

		// Extract ID, name, and parent ID using configured provider
		id := provider.ID(item)
		if id == "" {
			return nil, ErrEmptyID
		}
		n := NewNode(id, provider.Name(item), item)

		// Add to tracking collections
		parentLookup[id] = provider.ParentID(item)
		idToNode[id] = n
		nodeCount++
		cfg.ReportProgress(nodeCount, n)
	}

	// Pass 2: Establish parent/child relationships and validate tree has no cycles
	// Collect root nodes to return
	roots := make([]*Node[T], 0)
	for id, parentID := range parentLookup {
		if err := ctx.Err(); err != nil {
			return nil, err // Context has ended
		}
		node := idToNode[id]
		cfg.HandleExpansion(node)

		// Gather root notes
		if parentID == "" {
			roots = append(roots, node)
			continue // root node
		}

		// Check for cycles
		if detectCycle(id, parentID, parentLookup) {
			return nil, cyclicReferenceError(id, parentID)
		}

		// Set up hierarchical relationships
		parent, ok := idToNode[parentID]
		if !ok {
			return nil, errors.Join(ErrTreeConstruction, fmt.Errorf("treeview: parent id %q not found for node %q", parentID, id))
		}
		parent.AddChild(node)
	}

	// Pass 3: Apply depth limiting if configured
	if cfg.maxDepth >= 0 {
		roots = limitDepth(roots, cfg.maxDepth, 0)
	}

	// Return partial tree with error if we hit the traversal cap
	if hitTraversalCap {
		return roots, ErrTraversalLimit
	}

	return roots, nil
}

// detectCycle checks if adding a parent-child relationship would create a cycle.
// It traverses the parent chain starting from parentID to see if we would eventually
// reach childID, which would indicate a cycle.
func detectCycle(childID, parentID string, parentLookup map[string]string) bool {
	visited := make(map[string]bool, len(parentLookup))
	current := parentID
	for current != "" {
		if visited[current] {
			return true // Found a cycle
		}
		if current == childID {
			return true // Adding this relationship would create a cycle
		}
		visited[current] = true
		current = parentLookup[current]
	}
	return false
}

// NewTreeFromFileSystem creates a Tree that represents the filesystem hierarchy
// starting from the given root path. It is specialized for os.FileInfo data.
// Returns context errors unwrapped, or ErrFileSystem for filesystem errors.
//
// Supported options:
// Build options:
//   - WithFilterFunc:   Filters items during tree building
//   - WithMaxDepth:     Limits tree depth during construction
//   - WithExpandFunc:   Sets initial expansion state for nodes
//   - WithTraversalCap: Limits total nodes processed (returns partial tree + error if exceeded)
//   - WithProgressCallback: Invoked after each filesystem entry is processed (breadth-first per directory)
//
// Options used during a tree's runtime:
//   - WithSearcher:     Custom search algorithm
//   - WithFocusPolicy:  Custom focus navigation logic
//   - WithProvider:     Custom node rendering provider
func NewTreeFromFileSystem(
	ctx context.Context,
	path string,
	followSymlinks bool,
	opts ...Option[FileInfo],
) (*Tree[FileInfo], error) {
	// 1. Create config with a default provider for the filesystem.
	cfg := NewMasterConfig(opts, WithProvider[FileInfo](NewDefaultNodeProvider(
		WithFileExtensionRules[FileInfo](),
	)))

	// 2. Build the node hierarchy from the filesystem.
	nodes, err := buildFileSystemTree(ctx, path, followSymlinks, cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFileSystem, err)
	}

	// 3. Create the final tree with a specialized filesystem provider.
	tree := NewTreeFromCfg(nodes, cfg)
	return tree, nil
}

func buildFileSystemTree(ctx context.Context, path string, followSymlinks bool, cfg *MasterConfig[FileInfo]) ([]*Node[FileInfo], error) {
	// Resolve the path to absolute form, handling `~`, `..`, `.` expansion
	absPath, err := utils.ResolvePath(path)
	if err != nil {
		return nil, pathError(ErrPathResolution, path, err)
	}

	// Track visited inodes to detect symlink loops
	// Key is device:inode pair (Unix) or approximation (Windows)
	visited := make(map[string]struct{})

	// Get file info for the root path
	// This handles symlinks based on config settings
	info, err := utils.SafeStat(absPath, followSymlinks, visited)
	if err != nil {
		return nil, pathError(ErrFileSystem, absPath, err)
	}

	// Initialize traversal counter
	total := 1
	if cfg.HasTraversalCapBeenReached(total) {
		return nil, pathError(ErrTraversalLimit, absPath, nil)
	}

	// Create the root node
	rootNode := NewFileSystemNode(absPath, info)
	cfg.ReportProgress(1, rootNode)

	// Apply initial expansion state if configured
	cfg.HandleExpansion(rootNode)

	// If root is a directory, recursively scan its contents
	if info.IsDir() {
		if err := scanDir(ctx, rootNode, 0, followSymlinks, cfg, visited, &total); err != nil {
			return nil, err
		}
	}

	// Return single-element slice containing the root
	return []*Node[FileInfo]{rootNode}, nil
}

// scanDir scans a directory and its subdirectories, creating Node[FileInfo] for each entry.
// It returns an error if the traversal cap is exceeded or if there is an error.
func scanDir(ctx context.Context, parent *Node[FileInfo], depth int, followSymlinks bool, cfg *MasterConfig[FileInfo], visited map[string]struct{}, count *int) error {
	// Enforce depth limit if configured
	if cfg.HasDepthLimitBeenReached(depth) {
		return nil
	}

	// Read all entries in the directory
	entries, err := os.ReadDir(parent.Data().Path)
	if err != nil {
		return pathError(ErrDirectoryScan, parent.Data().Path, err)
	}

	// Collect child nodes as we process entries
	var children []*Node[FileInfo]
	for _, entry := range entries {
		// Check for cancellation between entries
		if err := ctx.Err(); err != nil {
			return err
		}

		// Build full path for the child entry
		childPath := filepath.Join(parent.Data().Path, entry.Name())

		// Get file info, following symlinks if configured
		// This also updates the visited map to detect loops
		info, err := utils.SafeStat(childPath, followSymlinks, visited)
		if err != nil {
			return pathError(ErrFileSystem, childPath, err)
		}

		// Apply filter function if provided
		if cfg.ShouldFilter(FileInfo{
			FileInfo: info,
			Path:     childPath,
		}) {
			continue // Item was filtered out
		}

		// Create node for this entry
		childNode := NewFileSystemNode(childPath, info)

		// Apply expansion state if configured
		cfg.HandleExpansion(childNode)

		// Increment and check traversal count
		// This prevents runaway scans of huge directories
		*count++
		cfg.ReportProgress(*count, childNode)
		if cfg.HasTraversalCapBeenReached(*count) {
			return pathError(ErrTraversalLimit, childPath, nil)
		}

		// Recursively scan subdirectories
		if info.IsDir() {
			if err := scanDir(ctx, childNode, depth+1, followSymlinks, cfg, visited, count); err != nil {
				return err
			}
		}

		// Add to children list
		children = append(children, childNode)
	}

	// Attach all children to the parent node
	if len(children) > 0 {
		parent.SetChildren(children)
	}
	return nil
}

// filterNodes recursively filters nodes based on the provided filter function.
// It maintains the tree structure by keeping parent nodes if they have matching children.
func filterNodes[T any](nodes []*Node[T], filterFunc FilterFn[T]) []*Node[T] {
	var filtered []*Node[T]
	for _, node := range nodes {
		// First, recursively filter children
		children := node.Children()
		filteredChildren := filterNodes(children, filterFunc)

		// Check if the current node should be included
		shouldInclude := filterFunc(*node.Data())

		// Include the node if:
		// 1. It passes the filter itself, OR
		// 2. It has children that passed the filter (to maintain tree structure)
		if shouldInclude || len(filteredChildren) > 0 {
			// Create a copy of the node to avoid modifying the original
			nodeCopy := NewNodeClone(node)

			// Set the filtered children
			nodeCopy.SetChildren(filteredChildren)
			filtered = append(filtered, nodeCopy)
		}
	}

	return filtered
}

// limitDepth recursively limits the tree depth to maxDepth levels.
// currentDepth tracks how deep we are in the tree (0 for root level).
func limitDepth[T any](nodes []*Node[T], maxDepth int, currentDepth int) []*Node[T] {
	// If we're at the max depth, return nodes without children
	if currentDepth >= maxDepth {
		var limited []*Node[T]
		for _, node := range nodes {
			// Create a copy without children
			limited = append(limited, NewNodeClone(node))
		}
		return limited
	}

	// Otherwise, recursively limit depth of children
	var limited []*Node[T]
	for _, node := range nodes {
		// Create a copy of the node
		nodeCopy := NewNodeClone(node)

		// Recursively limit children at the next depth level
		children := node.Children()
		limitedChildren := limitDepth(children, maxDepth, currentDepth+1)
		nodeCopy.SetChildren(limitedChildren)

		limited = append(limited, nodeCopy)
	}

	return limited
}

// applyExpansion recursively applies the expansion function to all nodes in the tree.
func applyExpansion[T any](nodes []*Node[T], cfg *MasterConfig[T]) {
	for _, node := range nodes {
		// Apply expansion function to this node
		cfg.HandleExpansion(node)

		// Recursively apply to children
		applyExpansion(node.Children(), cfg)
	}
}
