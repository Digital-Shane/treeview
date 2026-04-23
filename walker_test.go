package treeview_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Digital-Shane/treeview"
)

type fixtureWalker struct {
	root     treeview.WalkItem[string]
	children map[string][]treeview.WalkItem[string]
}

func (w *fixtureWalker) Root(context.Context) (treeview.WalkItem[string], error) {
	return w.root, nil
}

func (w *fixtureWalker) Children(_ context.Context, parent treeview.WalkItem[string]) ([]treeview.WalkItem[string], error) {
	return w.children[parent.ID], nil
}

func TestNewTreeFromWalker(t *testing.T) {
	ctx := context.Background()
	walker := &fixtureWalker{
		root: treeview.WalkItem[string]{ID: "root", Name: "Root", Data: "root"},
		children: map[string][]treeview.WalkItem[string]{
			"root": {
				{ID: "keep", Name: "Keep", Data: "keep"},
				{ID: "skip", Name: "Skip", Data: "skip"},
			},
			"keep": {
				{ID: "grandchild", Name: "Grandchild", Data: "grandchild"},
			},
		},
	}

	progressCalls := 0
	tree, err := treeview.NewTreeFromWalker(ctx, walker,
		treeview.WithFilterFunc(func(item string) bool { return item != "skip" }),
		treeview.WithMaxDepth[string](1),
		treeview.WithExpandFunc[string](func(node *treeview.Node[string]) bool {
			return node.ID() == "root"
		}),
		treeview.WithProgressCallback[string](func(processed int, node *treeview.Node[string]) {
			if processed <= 0 || node == nil {
				t.Fatalf("progress callback received invalid values: processed=%d node=%v", processed, node)
			}
			progressCalls++
		}),
	)
	if err != nil {
		t.Fatalf("NewTreeFromWalker() error = %v, want nil", err)
	}

	if len(tree.Nodes()) != 1 {
		t.Fatalf("NewTreeFromWalker() root count = %d, want 1", len(tree.Nodes()))
	}

	root := tree.Nodes()[0]
	if !root.IsExpanded() {
		t.Fatal("NewTreeFromWalker() root is not expanded")
	}
	if len(root.Children()) != 1 {
		t.Fatalf("NewTreeFromWalker() root children = %d, want 1", len(root.Children()))
	}
	if got := root.Children()[0].ID(); got != "keep" {
		t.Fatalf("NewTreeFromWalker() child ID = %q, want %q", got, "keep")
	}
	if len(root.Children()[0].Children()) != 0 {
		t.Fatalf("NewTreeFromWalker() child grandchildren = %d, want 0 due to max depth", len(root.Children()[0].Children()))
	}
	if progressCalls != 2 {
		t.Fatalf("NewTreeFromWalker() progress calls = %d, want 2", progressCalls)
	}
}

func TestNewTreeFromWalkerMaxDepthZero(t *testing.T) {
	ctx := context.Background()
	walker := &fixtureWalker{
		root: treeview.WalkItem[string]{ID: "root", Name: "Root", Data: "root"},
		children: map[string][]treeview.WalkItem[string]{
			"root": {
				{ID: "child", Name: "Child", Data: "child"},
			},
		},
	}

	tree, err := treeview.NewTreeFromWalker(ctx, walker, treeview.WithMaxDepth[string](0))
	if err != nil {
		t.Fatalf("NewTreeFromWalker() error = %v, want nil", err)
	}
	if len(tree.Nodes()) != 1 {
		t.Fatalf("NewTreeFromWalker() root count = %d, want 1", len(tree.Nodes()))
	}
	if len(tree.Nodes()[0].Children()) != 0 {
		t.Fatalf("NewTreeFromWalker() root children = %d, want 0 with max depth 0", len(tree.Nodes()[0].Children()))
	}
}

func TestNewTreeFromWalkerTraversalCap(t *testing.T) {
	ctx := context.Background()
	walker := &fixtureWalker{
		root: treeview.WalkItem[string]{ID: "root", Name: "Root", Data: "root"},
		children: map[string][]treeview.WalkItem[string]{
			"root": {
				{ID: "a", Name: "A", Data: "a"},
				{ID: "b", Name: "B", Data: "b"},
			},
		},
	}

	tree, err := treeview.NewTreeFromWalker(ctx, walker, treeview.WithTraversalCap[string](2))
	if !errors.Is(err, treeview.ErrTraversalLimit) {
		t.Fatalf("NewTreeFromWalker() error = %v, want ErrTraversalLimit", err)
	}
	if tree == nil {
		t.Fatal("NewTreeFromWalker() tree = nil, want partial tree")
	}
	if len(tree.Nodes()) != 1 {
		t.Fatalf("NewTreeFromWalker() root count = %d, want 1", len(tree.Nodes()))
	}
	if len(tree.Nodes()[0].Children()) != 1 {
		t.Fatalf("NewTreeFromWalker() partial child count = %d, want 1", len(tree.Nodes()[0].Children()))
	}
}
