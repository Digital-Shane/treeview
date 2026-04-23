package treeview

import "context"

// WalkItem describes a single source item that can be converted into a tree node.
type WalkItem[T any] struct {
	ID   string
	Name string
	Data T
}

// Walker provides a source-specific traversal adapter for NewTreeFromWalker.
//
// Implementations provide the root item and enumerate children for any given
// parent. The treeview package owns filtering, depth limiting, traversal caps,
// expansion, progress callbacks, and parent/child wiring. Filtering is applied
// before descent, so filtered parents are pruned with their descendants.
type Walker[T any] interface {
	Root(context.Context) (WalkItem[T], error)
	Children(context.Context, WalkItem[T]) ([]WalkItem[T], error)
}

// NewTreeFromWalker builds a tree from a Walker.
//
// Walker errors are returned as-is. Context errors are returned unwrapped.
func NewTreeFromWalker[T any](ctx context.Context, walker Walker[T], opts ...Option[T]) (*Tree[T], error) {
	cfg := newMasterConfig(opts)
	nodes, err := buildTreeFromWalker(ctx, walker, cfg)
	return newTree(nodes, cfg), err
}

func buildTreeFromWalker[T any](ctx context.Context, walker Walker[T], cfg *masterConfig[T]) ([]*Node[T], error) {
	rootItem, err := walker.Root(ctx)
	if err != nil {
		return nil, err
	}

	nodeCount := 0
	hitTraversalCap := false

	var buildSubtree func(WalkItem[T], int) (*Node[T], error)
	buildSubtree = func(item WalkItem[T], depth int) (*Node[T], error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if cfg.shouldFilter(item.Data) {
			return nil, nil
		}
		if cfg.hasTraversalCapBeenReached(nodeCount) {
			hitTraversalCap = true
			return nil, nil
		}
		if item.ID == "" {
			return nil, ErrEmptyID
		}

		node := NewNode(item.ID, item.Name, item.Data)
		nodeCount++
		cfg.reportProgress(nodeCount, node)
		cfg.handleExpansion(node)

		if cfg.hasDepthLimitBeenReached(depth) {
			return node, nil
		}

		children, err := walker.Children(ctx, item)
		if err != nil {
			return nil, err
		}

		childNodes := make([]*Node[T], 0, len(children))
		for _, child := range children {
			childNode, err := buildSubtree(child, depth+1)
			if err != nil {
				return nil, err
			}
			if childNode != nil {
				childNodes = append(childNodes, childNode)
			}
			if hitTraversalCap {
				break
			}
		}

		if len(childNodes) > 0 {
			node.SetChildren(childNodes)
		}

		return node, nil
	}

	rootNode, err := buildSubtree(rootItem, 0)
	if err != nil {
		return nil, err
	}
	if rootNode == nil {
		return nil, nil
	}
	if hitTraversalCap {
		return []*Node[T]{rootNode}, ErrTraversalLimit
	}

	return []*Node[T]{rootNode}, nil
}
