package treeview

import (
	"context"
	"fmt"
	"testing"

	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
)

// benchProvider is a minimal provider to minimize noise in benchmarks.
// We implement generic methods by wrapping with a generic struct.
type benchNodeProvider[T any] struct{}

func (p *benchNodeProvider[T]) Icon(n *Node[T]) string   { return "📁" }
func (p *benchNodeProvider[T]) Format(n *Node[T]) string { return n.Name() }
func (p *benchNodeProvider[T]) Style(n *Node[T], focused bool) lipgloss.Style {
	if focused {
		return lipgloss.NewStyle().Bold(true)
	}
	return lipgloss.NewStyle()
}

// buildWideTree creates a root with numChildren direct leaf children.
func buildWideTree(numChildren int) *Tree[string] {
	root := NewNode("root", "root", "root")
	for i := 0; i < numChildren; i++ {
		child := NewNode(fmt.Sprintf("c%d", i), fmt.Sprintf("c%d", i), "child")
		root.AddChild(child)
	}
	tree := NewTree([]*Node[string]{root}, WithProvider[string](&benchNodeProvider[string]{}))
	ctx := context.Background()
	tree.SetExpanded(ctx, "root", true)
	return tree
}

// buildDeepTree creates a multi-level tree with branching factor per depth.
func buildDeepTree(depth, branching int) *Tree[string] {
	root := NewNode("root", "root", "root")
	curLevel := []*Node[string]{root}
	nodeCounter := 0
	for d := 0; d < depth; d++ {
		var next []*Node[string]
		for _, parent := range curLevel {
			for i := 0; i < branching; i++ {
				id := fmt.Sprintf("d%dn%d_%d", d, i, nodeCounter)
				nodeCounter++
				child := NewNode(id, id, "n")
				parent.AddChild(child)
				next = append(next, child)
			}
		}
		curLevel = next
	}
	tree := NewTree([]*Node[string]{root}, WithProvider[string](&benchNodeProvider[string]{}))
	ctx := context.Background()
	tree.ExpandAll(ctx)
	return tree
}

var benchSink string // global sink to avoid compiler elimination

func benchmarkRenderer(b *testing.B, fn func(context.Context, *Tree[string], *viewport.Model) (string, error), tree *Tree[string]) {
	ctx := context.Background()
	vp := viewport.New(viewport.WithWidth(100), viewport.WithHeight(30))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vpcopy := vp // copy so offsets reset
		s, err := fn(ctx, tree, &vpcopy)
		if err != nil {
			b.Fatalf("render error: %v", err)
		}
		benchSink = s
	}
}

func BenchmarkRenderTreeWithViewport_Wide(b *testing.B) {
	for _, n := range []int{1000, 5000, 10000} {
		b.Run(fmt.Sprintf("%d", n), func(sb *testing.B) {
			tree := buildWideTree(n)
			benchmarkRenderer(sb, renderTreeWithViewport[string], tree)
		})
	}
}

func BenchmarkRenderTreeWithViewport_Deep(b *testing.B) {
	cases := []struct{ depth, branching int }{{5, 3}, {6, 3}}
	for _, c := range cases {
		label := fmt.Sprintf("d%d_b%d", c.depth, c.branching)
		b.Run(label, func(sb *testing.B) {
			tree := buildDeepTree(c.depth, c.branching)
			benchmarkRenderer(sb, renderTreeWithViewport[string], tree)
		})
	}
}
