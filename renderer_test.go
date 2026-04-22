package treeview

import (
	"context"
	"errors"
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
	"github.com/google/go-cmp/cmp"
)

// Mock data type for testing
type mockData struct {
	name string
}

// Mock provider for testing
type mockProvider struct {
	iconFunc   func(*Node[mockData]) string
	formatFunc func(*Node[mockData]) string
	styleFunc  func(*Node[mockData], bool) lipgloss.Style
}

func (p *mockProvider) Icon(n *Node[mockData]) string {
	if p.iconFunc != nil {
		return p.iconFunc(n)
	}
	return "📁"
}

func (p *mockProvider) Format(n *Node[mockData]) string {
	if p.formatFunc != nil {
		return p.formatFunc(n)
	}
	return n.Name()
}

func (p *mockProvider) Style(n *Node[mockData], focused bool) lipgloss.Style {
	if p.styleFunc != nil {
		return p.styleFunc(n, focused)
	}
	return lipgloss.NewStyle()
}

func TestNormalizeIconWidth(t *testing.T) {
	tests := []struct {
		name string
		icon string
		want string
	}{
		{"empty", "", ""},
		{"single_char", "A", "A  "},
		{"two_chars", "AB", "AB "},
		{"three_chars", "ABC", "ABC "},
		{"four_chars", "ABCD", "ABCD "},
		{"emoji", "📁", "📁 "},
		{"emoji_with_space", "📁 ", "📁 "},
		{"wide_emoji", "👨‍👩‍👧‍👦", "👨‍👩‍👧‍👦 "},
		{"single_space", " ", "   "},
		{"two_spaces", "  ", "   "},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := NormalizeIconWidth(test.icon)
			if got != test.want {
				t.Errorf("NormalizeIconWidth(%q) = %q, want %q", test.icon, got, test.want)
			}
		})
	}
}

func TestBuildPrefix(t *testing.T) {
	tests := []struct {
		name                string
		ancestorIsLastChild []bool
		isLast              bool
		want                string
	}{
		{"first_level_not_last", []bool{}, false, "├── "},
		{"first_level_last", []bool{}, true, "└── "},
		{"second_level_parent_not_last_node_not_last", []bool{false}, false, "│   ├── "},
		{"second_level_parent_not_last_node_last", []bool{false}, true, "│   └── "},
		{"second_level_parent_last_node_not_last", []bool{true}, false, "    ├── "},
		{"second_level_parent_last_node_last", []bool{true}, true, "    └── "},
		{"third_level_mixed", []bool{false, true}, true, "│       └── "},
		{"third_level_all_not_last", []bool{false, false}, false, "│   │   ├── "},
		{"third_level_all_last", []bool{true, true}, true, "        └── "},
		{"deep_nesting", []bool{false, true, false, true}, false, "│       │       ├── "},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := buildPrefix(test.ancestorIsLastChild, test.isLast)
			if got != test.want {
				t.Errorf("buildPrefix(%v, %v) = %q, want %q",
					test.ancestorIsLastChild, test.isLast, got, test.want)
			}
		})
	}
}

func TestRenderNode(t *testing.T) {
	tests := []struct {
		name      string
		node      *Node[mockData]
		prefix    string
		isFocused bool
		provider  *mockProvider
		want      string
		wantErr   bool
	}{
		{
			name:      "basic_node",
			node:      NewNode("test", "test", mockData{name: "test"}),
			prefix:    "├── ",
			isFocused: false,
			provider:  &mockProvider{},
			want:      "├── 📁 test",
			wantErr:   false,
		},
		{
			name:      "focused_node",
			node:      NewNode("focused", "focused", mockData{name: "focused"}),
			prefix:    "└── ",
			isFocused: true,
			provider: &mockProvider{
				styleFunc: func(n *Node[mockData], focused bool) lipgloss.Style {
					if focused {
						return lipgloss.NewStyle().Bold(true)
					}
					return lipgloss.NewStyle()
				},
			},
			want:    "\x1b[1m└── 📁 focused\x1b[m",
			wantErr: false,
		},
		{
			name:      "custom_icon_and_format",
			node:      NewNode("custom", "custom", mockData{name: "custom"}),
			prefix:    "│   ├── ",
			isFocused: false,
			provider: &mockProvider{
				iconFunc:   func(n *Node[mockData]) string { return "🔧" },
				formatFunc: func(n *Node[mockData]) string { return "custom-format" },
			},
			want:    "│   ├── 🔧 custom-format",
			wantErr: false,
		},
		{
			name:      "empty_prefix",
			node:      NewNode("root", "root", mockData{name: "root"}),
			prefix:    "",
			isFocused: false,
			provider:  &mockProvider{},
			want:      "📁 root",
			wantErr:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := renderNode(test.provider, test.node, test.prefix, test.isFocused, 0)

			if (err != nil) != test.wantErr {
				t.Errorf("renderNode() error = %v, wantErr %v", err, test.wantErr)
				return
			}

			if got != test.want {
				t.Errorf("renderNode() = %q, want %q", got, test.want)
			}
		})
	}
}

// Create fresh nodes for each test case to avoid state sharing
func createTestTree() (*Node[mockData], *Node[mockData], *Node[mockData], *Node[mockData], *Node[mockData]) {
	rootNode := NewNode("root", "root", mockData{name: "root"})
	child1 := NewNode("child1", "child1", mockData{name: "child1"})
	child2 := NewNode("child2", "child2", mockData{name: "child2"})
	grandchild1 := NewNode("grandchild1", "grandchild1", mockData{name: "grandchild1"})
	grandchild2 := NewNode("grandchild2", "grandchild2", mockData{name: "grandchild2"})

	// Build tree structure
	rootNode.AddChild(child1)
	rootNode.AddChild(child2)
	child1.AddChild(grandchild1)
	child1.AddChild(grandchild2)

	return rootNode, child1, child2, grandchild1, grandchild2
}

func TestRenderTree(t *testing.T) {
	tests := []struct {
		name             string
		tree             *Tree[mockData]
		want             string
		wantFocusedIndex int
		wantErr          bool
	}{
		{
			name: "simple_tree",
			tree: func() *Tree[mockData] {
				rootNode, _, _, _, _ := createTestTree()
				tree := NewTree([]*Node[mockData]{rootNode}, WithProvider[mockData](&mockProvider{}))
				ctx := context.Background()
				tree.SetExpanded(ctx, "root", true)
				tree.SetExpanded(ctx, "child1", true)
				return tree
			}(),
			want: `📁 root
    ├── 📁 child1
    │   ├── 📁 grandchild1
    │   └── 📁 grandchild2
    └── 📁 child2`,
			wantFocusedIndex: 0,
			wantErr:          false,
		},
		{
			name: "tree_with_focus",
			tree: func() *Tree[mockData] {
				rootNode, _, _, _, _ := createTestTree()
				tree := NewTree([]*Node[mockData]{rootNode}, WithProvider[mockData](&mockProvider{}))
				ctx := context.Background()
				tree.SetExpanded(ctx, "root", true)
				tree.SetExpanded(ctx, "child1", true)
				tree.SetFocusedID(ctx, "grandchild1")
				return tree
			}(),
			want: `📁 root
    ├── 📁 child1
    │   ├── 📁 grandchild1
    │   └── 📁 grandchild2
    └── 📁 child2`,
			wantFocusedIndex: 2,
			wantErr:          false,
		},
		{
			name: "collapsed_tree",
			tree: func() *Tree[mockData] {
				rootNode, _, _, _, _ := createTestTree()
				tree := NewTree([]*Node[mockData]{rootNode}, WithProvider[mockData](&mockProvider{}))
				ctx := context.Background()
				tree.SetExpanded(ctx, "root", true)
				return tree
			}(),
			want: `📁 root
    ├── 📁 child1
    └── 📁 child2`,
			wantFocusedIndex: 0,
			wantErr:          false,
		},
		{
			name: "single_node",
			tree: func() *Tree[mockData] {
				singleNode := NewNode("single", "single", mockData{name: "single"})
				return NewTree([]*Node[mockData]{singleNode}, WithProvider[mockData](&mockProvider{}))
			}(),
			want:             `📁 single`,
			wantFocusedIndex: 0,
			wantErr:          false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			got, gotFocusedIndex, err := renderTree(ctx, test.tree)

			if (err != nil) != test.wantErr {
				t.Errorf("renderTree() error = %v, wantErr %v", err, test.wantErr)
				return
			}

			if got != test.want {
				t.Errorf("renderTree() content = %q, want %q", got, test.want)
				if diff := cmp.Diff(test.want, got); diff != "" {
					t.Errorf("renderTree() diff (-want +got):\n%s", diff)
				}
			}

			if gotFocusedIndex != test.wantFocusedIndex {
				t.Errorf("renderTree() focusedIndex = %v, want %v", gotFocusedIndex, test.wantFocusedIndex)
			}
		})
	}
}

func TestRenderTreeWithViewport(t *testing.T) {
	// Create a tree for viewport testing
	rootNode := NewNode("root", "root", mockData{name: "root"})
	childNames := []string{"child1", "child2", "child3", "child4", "child5", "child6", "child7", "child8", "child9", "child10"}
	for _, name := range childNames {
		child := NewNode(name, name, mockData{name: name})
		rootNode.AddChild(child)
	}

	tests := []struct {
		name           string
		tree           *Tree[mockData]
		viewportHeight int
		viewportWidth  int
		wantContains   []string
		wantErr        bool
	}{
		{
			name: "viewport_shows_top",
			tree: func() *Tree[mockData] {
				tree := NewTree([]*Node[mockData]{rootNode}, WithProvider[mockData](&mockProvider{}))
				ctx := context.Background()
				tree.SetExpanded(ctx, "root", true)
				return tree
			}(),
			viewportHeight: 5,
			viewportWidth:  30,
			wantContains:   []string{"root", "child1", "child2", "child3", "child4"},
			wantErr:        false,
		},
		{
			name: "viewport_scrolls_to_focus",
			tree: func() *Tree[mockData] {
				tree := NewTree([]*Node[mockData]{rootNode}, WithProvider[mockData](&mockProvider{}))
				ctx := context.Background()
				tree.SetExpanded(ctx, "root", true)
				tree.SetFocusedID(ctx, "child8")
				return tree
			}(),
			viewportHeight: 5,
			viewportWidth:  30,
			wantContains:   []string{"child8"},
			wantErr:        false,
		},
		{
			name: "empty_viewport",
			tree: func() *Tree[mockData] {
				tree := NewTree([]*Node[mockData]{rootNode}, WithProvider[mockData](&mockProvider{}))
				ctx := context.Background()
				tree.SetExpanded(ctx, "root", true)
				return tree
			}(),
			viewportHeight: 0,
			viewportWidth:  30,
			wantContains:   []string{},
			wantErr:        false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			vp := viewport.New(viewport.WithHeight(test.viewportHeight), viewport.WithWidth(test.viewportWidth))

			got, err := renderTreeWithViewport(ctx, test.tree, &vp)

			if (err != nil) != test.wantErr {
				t.Errorf("renderTreeWithViewport() error = %v, wantErr %v", err, test.wantErr)
				return
			}

			for _, want := range test.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("renderTreeWithViewport() = %q, want to contain %q", got, want)
				}
			}
		})
	}
}

func TestRenderTreeContextCancellation(t *testing.T) {
	// Create a large tree
	rootNode := NewNode("root", "root", mockData{name: "root"})
	for i := 0; i < 100; i++ {
		child := NewNode("child"+string(rune(i)), "child"+string(rune(i)), mockData{name: "child" + string(rune(i))})
		rootNode.AddChild(child)
	}

	tree := NewTree([]*Node[mockData]{rootNode}, WithProvider[mockData](&mockProvider{}))
	ctx := context.Background()
	tree.SetExpanded(ctx, "root", true)

	// Cancel context immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := renderTree(ctx, tree)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("renderTree() with cancelled context = %v, want context.Canceled", err)
	}
}
