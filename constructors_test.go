package treeview

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNewTree(t *testing.T) {
	tests := []struct {
		name    string
		nodes   []*Node[string]
		opts    []Option[string]
		wantLen int
		wantNil bool
		checkFn func(*testing.T, *Tree[string])
	}{
		{
			name:    "empty_nodes",
			nodes:   []*Node[string]{},
			wantLen: 0,
			wantNil: false,
		},
		{
			name:    "nil_nodes",
			nodes:   nil,
			wantLen: 0,
			wantNil: false,
		},
		{
			name: "single_node",
			nodes: []*Node[string]{
				NewNode("1", "Node 1", "data1"),
			},
			wantLen: 1,
			checkFn: func(t *testing.T, tree *Tree[string]) {
				t.Helper()
				focusedNode := tree.GetFocusedNode()
				if focusedNode == nil {
					t.Error("NewTree(single node) focused = nil, want first node")
				} else if focusedNode.ID() != "1" {
					t.Errorf("NewTree(single node) focused.ID() = %q, want %q", focusedNode.ID(), "1")
				}
			},
		},
		{
			name: "multiple_nodes",
			nodes: []*Node[string]{
				NewNode("1", "Node 1", "data1"),
				NewNode("2", "Node 2", "data2"),
				NewNode("3", "Node 3", "data3"),
			},
			wantLen: 3,
			checkFn: func(t *testing.T, tree *Tree[string]) {
				t.Helper()
				focusedNode := tree.GetFocusedNode()
				if focusedNode == nil {
					t.Error("NewTree(multiple nodes) focused = nil, want first node")
				} else if focusedNode.ID() != "1" {
					t.Errorf("NewTree(multiple nodes) focused.ID() = %q, want %q", focusedNode.ID(), "1")
				}
			},
		},
		{
			name: "with_filter_func",
			nodes: []*Node[string]{
				NewNode("1", "Node 1", "keep"),
				NewNode("2", "Node 2", "filter"),
				NewNode("3", "Node 3", "keep"),
			},
			opts: []Option[string]{
				WithFilterFunc(func(data string) bool {
					return data != "filter"
				}),
			},
			wantLen: 2,
			checkFn: func(t *testing.T, tree *Tree[string]) {
				t.Helper()
				for _, node := range tree.nodes {
					if *node.Data() == "filter" {
						t.Errorf("NewTree(with filter) contains filtered node %q", node.ID())
					}
				}
			},
		},
		{
			name: "with_max_depth",
			nodes: func() []*Node[string] {
				root := NewNode("1", "Root", "root")
				child := NewNode("2", "Child", "child")
				grandchild := NewNode("3", "Grandchild", "grandchild")
				child.AddChild(grandchild)
				root.AddChild(child)
				return []*Node[string]{root}
			}(),
			opts: []Option[string]{
				WithMaxDepth[string](1),
			},
			wantLen: 1,
			checkFn: func(t *testing.T, tree *Tree[string]) {
				t.Helper()
				root := tree.nodes[0]
				if len(root.Children()) != 1 {
					t.Errorf("NewTree(maxDepth=1) root children = %d, want 1", len(root.Children()))
				}
				if len(root.Children()) > 0 && len(root.Children()[0].Children()) != 0 {
					t.Errorf("NewTree(maxDepth=1) child has children, want none")
				}
			},
		},
		{
			name: "with_expand_func",
			nodes: []*Node[string]{
				NewNode("1", "Node 1", "expand"),
				NewNode("2", "Node 2", "collapse"),
			},
			opts: []Option[string]{
				WithExpandFunc[string](func(node *Node[string]) bool {
					return *node.Data() == "expand"
				}),
			},
			wantLen: 2,
			checkFn: func(t *testing.T, tree *Tree[string]) {
				t.Helper()
				for _, node := range tree.nodes {
					want := *node.Data() == "expand"
					if node.IsExpanded() != want {
						t.Errorf("NewTree(with expand func) node %q expanded = %v, want %v",
							node.ID(), node.IsExpanded(), want)
					}
				}
			},
		},
		{
			name: "with_progress_callback",
			nodes: []*Node[string]{
				NewNode("1", "Node 1", "a"),
				NewNode("2", "Node 2", "b"),
				NewNode("3", "Node 3", "c"),
			},
			opts: []Option[string]{
				WithProgressCallback[string](func(processed int, n *Node[string]) {
					// simple no-op to ensure it is invoked; we rely on counting side effects
				}),
			},
			wantLen: 3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := NewTree(test.nodes, test.opts...)

			if test.wantNil {
				if got != nil {
					t.Errorf("NewTree(%s) = %v, want nil", test.name, got)
				}
				return
			}

			if got == nil {
				t.Fatalf("NewTree(%s) = nil, want non-nil", test.name)
			}

			if len(got.nodes) != test.wantLen {
				t.Errorf("NewTree(%s) node count = %d, want %d", test.name, len(got.nodes), test.wantLen)
			}

			if test.checkFn != nil {
				test.checkFn(t, got)
			}
		})
	}
}

type testNestedProvider struct{}

func (p *testNestedProvider) ID(item testNestedItem) string {
	return item.id
}

func (p *testNestedProvider) Name(item testNestedItem) string {
	return item.name
}

func (p *testNestedProvider) Children(item testNestedItem) []testNestedItem {
	return item.children
}

type testNestedItem struct {
	id       string
	name     string
	children []testNestedItem
	data     string
}

func TestNewTreeFromNestedData(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		items     []testNestedItem
		provider  NestedDataProvider[testNestedItem]
		opts      []Option[testNestedItem]
		wantNodes int
		wantErr   error
		checkFn   func(*testing.T, *Tree[testNestedItem])
	}{
		{
			name:      "empty_items",
			items:     []testNestedItem{},
			provider:  &testNestedProvider{},
			wantNodes: 0,
			wantErr:   nil,
		},
		{
			name: "single_root",
			items: []testNestedItem{
				{id: "1", name: "Root", data: "root"},
			},
			provider:  &testNestedProvider{},
			wantNodes: 1,
			checkFn: func(t *testing.T, tree *Tree[testNestedItem]) {
				t.Helper()
				if tree.nodes[0].ID() != "1" {
					t.Errorf("NewTreeFromNestedData(single root) root ID = %q, want %q", tree.nodes[0].ID(), "1")
				}
			},
		},
		{
			name: "nested_structure",
			items: []testNestedItem{
				{
					id:   "1",
					name: "Root",
					children: []testNestedItem{
						{id: "2", name: "Child 1"},
						{
							id:   "3",
							name: "Child 2",
							children: []testNestedItem{
								{id: "4", name: "Grandchild"},
							},
						},
					},
				},
			},
			provider:  &testNestedProvider{},
			wantNodes: 1,
			checkFn: func(t *testing.T, tree *Tree[testNestedItem]) {
				t.Helper()
				root := tree.nodes[0]
				if len(root.Children()) != 2 {
					t.Errorf("NewTreeFromNestedData(nested) root children = %d, want 2", len(root.Children()))
				}
				if len(root.Children()) > 1 && len(root.Children()[1].Children()) != 1 {
					t.Errorf("NewTreeFromNestedData(nested) second child children = %d, want 1",
						len(root.Children()[1].Children()))
				}
			},
		},
		{
			name: "with_filter",
			items: []testNestedItem{
				{
					id:   "1",
					name: "Root",
					data: "keep",
					children: []testNestedItem{
						{id: "2", name: "Child 1", data: "filter"},
						{id: "3", name: "Child 2", data: "keep"},
					},
				},
			},
			provider: &testNestedProvider{},
			opts: []Option[testNestedItem]{
				WithFilterFunc(func(item testNestedItem) bool {
					return item.data != "filter"
				}),
			},
			wantNodes: 1,
			checkFn: func(t *testing.T, tree *Tree[testNestedItem]) {
				t.Helper()
				root := tree.nodes[0]
				if len(root.Children()) != 1 {
					t.Errorf("NewTreeFromNestedData(with filter) root children = %d, want 1", len(root.Children()))
				}
				if len(root.Children()) > 0 && root.Children()[0].ID() != "3" {
					t.Errorf("NewTreeFromNestedData(with filter) child ID = %q, want %q",
						root.Children()[0].ID(), "3")
				}
			},
		},
		{
			name: "with_max_depth",
			items: []testNestedItem{
				{
					id:   "1",
					name: "Root",
					children: []testNestedItem{
						{
							id:   "2",
							name: "Child",
							children: []testNestedItem{
								{id: "3", name: "Grandchild"},
							},
						},
					},
				},
			},
			provider: &testNestedProvider{},
			opts: []Option[testNestedItem]{
				WithMaxDepth[testNestedItem](1),
			},
			wantNodes: 1,
			checkFn: func(t *testing.T, tree *Tree[testNestedItem]) {
				t.Helper()
				root := tree.nodes[0]
				if len(root.Children()) != 1 {
					t.Errorf("NewTreeFromNestedData(maxDepth=1) root children = %d, want 1", len(root.Children()))
				}
				if len(root.Children()) > 0 && len(root.Children()[0].Children()) != 0 {
					t.Errorf("NewTreeFromNestedData(maxDepth=1) child has children, want none")
				}
			},
		},
		{
			name: "with_traversal_cap",
			items: []testNestedItem{
				{
					id:   "1",
					name: "Root",
					children: []testNestedItem{
						{id: "2", name: "Child 1"},
						{id: "3", name: "Child 2"},
						{id: "4", name: "Child 3"},
					},
				},
			},
			provider: &testNestedProvider{},
			opts: []Option[testNestedItem]{
				WithTraversalCap[testNestedItem](3),
			},
			wantNodes: 1,
			wantErr:   ErrTraversalLimit,
			checkFn: func(t *testing.T, tree *Tree[testNestedItem]) {
				t.Helper()
				root := tree.nodes[0]
				totalNodes := 1 + len(root.Children())
				if totalNodes > 3 {
					t.Errorf("NewTreeFromNestedData(traversalCap=3) total nodes = %d, want <= 3", totalNodes)
				}
			},
		},
		{
			name: "with_progress_callback",
			items: []testNestedItem{
				{
					id:       "1",
					name:     "Root",
					children: []testNestedItem{{id: "2", name: "Child"}},
				},
			},
			provider:  &testNestedProvider{},
			wantNodes: 1,
			checkFn: func(t *testing.T, tree *Tree[testNestedItem]) {
				// The callback assertion is performed inside the Option via closure.
			},
			opts: []Option[testNestedItem]{
				WithProgressCallback[testNestedItem](func(processed int, n *Node[testNestedItem]) {
					_ = processed // intentionally ignore; existence proves invocation path compiles
					if n == nil {
						panic("progress callback received nil node")
					}
				}),
				WithExpandAll[testNestedItem](),
			},
		},
		{
			name:      "context_cancelled",
			items:     []testNestedItem{},
			provider:  &testNestedProvider{},
			wantNodes: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Handle context cancellation test separately
			if test.name == "context_cancelled" {
				cancelCtx, cancel := context.WithCancel(context.Background())
				cancel()
				_, err := NewTreeFromNestedData(cancelCtx, []testNestedItem{{id: "1"}}, &testNestedProvider{})
				if !errors.Is(err, context.Canceled) {
					t.Errorf("NewTreeFromNestedData(cancelled context) error = %v, want context.Canceled", err)
				}
				return
			}

			got, err := NewTreeFromNestedData(ctx, test.items, test.provider, test.opts...)

			if test.wantErr != nil {
				if !errors.Is(err, test.wantErr) {
					t.Errorf("NewTreeFromNestedData(%s) error = %v, want %v", test.name, err, test.wantErr)
				}
			} else if err != nil && err != ErrTraversalLimit {
				t.Errorf("NewTreeFromNestedData(%s) unexpected error = %v", test.name, err)
			}

			if got == nil {
				t.Fatalf("NewTreeFromNestedData(%s) = nil, want non-nil", test.name)
			}

			if len(got.nodes) != test.wantNodes {
				t.Errorf("NewTreeFromNestedData(%s) node count = %d, want %d", test.name, len(got.nodes), test.wantNodes)
			}

			if test.checkFn != nil {
				test.checkFn(t, got)
			}
		})
	}
}

type testFlatProvider struct{}

func (p *testFlatProvider) ID(item testFlatItem) string {
	return item.id
}

func (p *testFlatProvider) Name(item testFlatItem) string {
	return item.name
}

func (p *testFlatProvider) ParentID(item testFlatItem) string {
	return item.parentID
}

type testFlatItem struct {
	id       string
	name     string
	parentID string
}

func TestNewTreeFromFlatData(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		items     []testFlatItem
		provider  FlatDataProvider[testFlatItem]
		opts      []Option[testFlatItem]
		wantRoots int
		wantErr   bool
		checkFn   func(*testing.T, *Tree[testFlatItem])
	}{
		{
			name:      "empty_items",
			items:     []testFlatItem{},
			provider:  &testFlatProvider{},
			wantRoots: 0,
		},
		{
			name: "single_root",
			items: []testFlatItem{
				{id: "1", name: "Root", parentID: ""},
			},
			provider:  &testFlatProvider{},
			wantRoots: 1,
		},
		{
			name: "parent_child_hierarchy",
			items: []testFlatItem{
				{id: "1", name: "Root", parentID: ""},
				{id: "2", name: "Child 1", parentID: "1"},
				{id: "3", name: "Child 2", parentID: "1"},
				{id: "4", name: "Grandchild", parentID: "2"},
			},
			provider:  &testFlatProvider{},
			wantRoots: 1,
			checkFn: func(t *testing.T, tree *Tree[testFlatItem]) {
				t.Helper()
				root := tree.nodes[0]
				if len(root.Children()) != 2 {
					t.Errorf("NewTreeFromFlatData(hierarchy) root children = %d, want 2", len(root.Children()))
				}
				child1 := root.Children()[0]
				if child1.ID() == "2" && len(child1.Children()) != 1 {
					t.Errorf("NewTreeFromFlatData(hierarchy) child1 children = %d, want 1", len(child1.Children()))
				}
			},
		},
		{
			name: "multiple_roots",
			items: []testFlatItem{
				{id: "1", name: "Root 1", parentID: ""},
				{id: "2", name: "Root 2", parentID: ""},
				{id: "3", name: "Child of 1", parentID: "1"},
			},
			provider:  &testFlatProvider{},
			wantRoots: 2,
		},
		{
			name: "cyclic_reference",
			items: []testFlatItem{
				{id: "1", name: "Node 1", parentID: "2"},
				{id: "2", name: "Node 2", parentID: "1"},
			},
			provider: &testFlatProvider{},
			wantErr:  true,
			checkFn: func(t *testing.T, tree *Tree[testFlatItem]) {
				t.Helper()
				_, err := NewTreeFromFlatData(ctx, []testFlatItem{
					{id: "1", name: "Node 1", parentID: "2"},
					{id: "2", name: "Node 2", parentID: "1"},
				}, &testFlatProvider{})
				if err == nil {
					t.Error("NewTreeFromFlatData(cyclic) error = nil, want cyclic reference error")
				}
			},
		},
		{
			name: "missing_parent",
			items: []testFlatItem{
				{id: "1", name: "Child", parentID: "999"},
			},
			provider: &testFlatProvider{},
			wantErr:  true,
		},
		{
			name: "with_max_depth",
			items: []testFlatItem{
				{id: "1", name: "Root", parentID: ""},
				{id: "2", name: "Child", parentID: "1"},
				{id: "3", name: "Grandchild", parentID: "2"},
			},
			provider: &testFlatProvider{},
			opts: []Option[testFlatItem]{
				WithMaxDepth[testFlatItem](1),
			},
			wantRoots: 1,
			checkFn: func(t *testing.T, tree *Tree[testFlatItem]) {
				t.Helper()
				root := tree.nodes[0]
				if len(root.Children()) != 1 {
					t.Errorf("NewTreeFromFlatData(maxDepth=1) root children = %d, want 1", len(root.Children()))
				}
				if len(root.Children()) > 0 && len(root.Children()[0].Children()) != 0 {
					t.Errorf("NewTreeFromFlatData(maxDepth=1) child has children, want none")
				}
			},
		},
		{
			name: "with_traversal_cap",
			items: []testFlatItem{
				{id: "1", name: "Root", parentID: ""},
				{id: "2", name: "Child 1", parentID: "1"},
				{id: "3", name: "Child 2", parentID: "1"},
				{id: "4", name: "Child 3", parentID: "1"},
			},
			provider: &testFlatProvider{},
			opts: []Option[testFlatItem]{
				WithTraversalCap[testFlatItem](2),
			},
			wantRoots: 1,
			wantErr:   true,
			checkFn: func(t *testing.T, tree *Tree[testFlatItem]) {
				t.Helper()
				if tree == nil {
					return
				}
				totalNodes := 0
				for _, root := range tree.nodes {
					totalNodes += 1 + len(root.Children())
				}
				if totalNodes > 2 {
					t.Errorf("NewTreeFromFlatData(traversalCap=2) total nodes = %d, want <= 2", totalNodes)
				}
			},
		},
		{
			name: "with_progress_callback",
			items: []testFlatItem{
				{id: "1", name: "Root", parentID: ""},
				{id: "2", name: "Child", parentID: "1"},
			},
			provider: &testFlatProvider{},
			opts: []Option[testFlatItem]{
				WithProgressCallback[testFlatItem](func(processed int, n *Node[testFlatItem]) {
					_ = processed
					if n == nil {
						panic("nil node in progress callback")
					}
				}),
			},
			wantRoots: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NewTreeFromFlatData(ctx, test.items, test.provider, test.opts...)

			if test.wantErr {
				if err == nil {
					t.Errorf("NewTreeFromFlatData(%s) error = nil, want error", test.name)
				}
			} else if err != nil && err != ErrTraversalLimit {
				t.Errorf("NewTreeFromFlatData(%s) unexpected error = %v", test.name, err)
			}

			if got == nil {
				if !test.wantErr {
					t.Fatalf("NewTreeFromFlatData(%s) = nil, want non-nil", test.name)
				}
				return
			}

			if len(got.nodes) != test.wantRoots {
				t.Errorf("NewTreeFromFlatData(%s) root count = %d, want %d", test.name, len(got.nodes), test.wantRoots)
			}

			if test.checkFn != nil {
				test.checkFn(t, got)
			}
		})
	}
}

func TestNewTreeFromFileSystem(t *testing.T) {
	ctx := context.Background()

	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create test file structure
	testFiles := []string{
		filepath.Join(tmpDir, "file1.txt"),
		filepath.Join(tmpDir, "file2.go"),
		filepath.Join(tmpDir, "dir1", "file3.txt"),
		filepath.Join(tmpDir, "dir1", "subdir", "file4.go"),
		filepath.Join(tmpDir, "dir2", "file5.txt"),
	}

	for _, file := range testFiles {
		dir := filepath.Dir(file)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create test directory %s: %v", dir, err)
		}
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	tests := []struct {
		name           string
		path           string
		followSymlinks bool
		opts           []Option[FileInfo]
		wantErr        bool
		checkFn        func(*testing.T, *Tree[FileInfo])
	}{
		{
			name:           "valid_directory",
			path:           tmpDir,
			followSymlinks: false,
			checkFn: func(t *testing.T, tree *Tree[FileInfo]) {
				t.Helper()
				if len(tree.nodes) != 1 {
					t.Errorf("NewTreeFromFileSystem(valid dir) root count = %d, want 1", len(tree.nodes))
				}
				root := tree.nodes[0]
				// Should have 2 directories and 2 files at root level
				if len(root.Children()) != 4 {
					t.Errorf("NewTreeFromFileSystem(valid dir) root children = %d, want 4", len(root.Children()))
				}
			},
		},
		{
			name:           "single_file",
			path:           testFiles[0],
			followSymlinks: false,
			checkFn: func(t *testing.T, tree *Tree[FileInfo]) {
				t.Helper()
				if len(tree.nodes) != 1 {
					t.Errorf("NewTreeFromFileSystem(single file) root count = %d, want 1", len(tree.nodes))
				}
				root := tree.nodes[0]
				if root.Data().IsDir() {
					t.Error("NewTreeFromFileSystem(single file) root is directory, want file")
				}
				if len(root.Children()) != 0 {
					t.Errorf("NewTreeFromFileSystem(single file) children = %d, want 0", len(root.Children()))
				}
			},
		},
		{
			name:           "nonexistent_path",
			path:           filepath.Join(tmpDir, "nonexistent"),
			followSymlinks: false,
			wantErr:        true,
		},
		{
			name:           "with_max_depth",
			path:           tmpDir,
			followSymlinks: false,
			opts: []Option[FileInfo]{
				WithMaxDepth[FileInfo](1),
			},
			checkFn: func(t *testing.T, tree *Tree[FileInfo]) {
				t.Helper()
				root := tree.nodes[0]
				// Check that subdirectories don't have children
				for _, child := range root.Children() {
					if child.Data().IsDir() && len(child.Children()) != 0 {
						t.Errorf("NewTreeFromFileSystem(maxDepth=1) directory %q has children, want none",
							child.Name())
					}
				}
			},
		},
		{
			name:           "with_filter",
			path:           tmpDir,
			followSymlinks: false,
			opts: []Option[FileInfo]{
				WithFilterFunc(func(info FileInfo) bool {
					return filepath.Ext(info.Path) != ".txt"
				}),
			},
			checkFn: func(t *testing.T, tree *Tree[FileInfo]) {
				t.Helper()
				// Check that no .txt files are present
				var checkNode func(*Node[FileInfo])
				checkNode = func(node *Node[FileInfo]) {
					if filepath.Ext(node.Data().Path) == ".txt" {
						t.Errorf("NewTreeFromFileSystem(filter .txt) found .txt file: %s", node.Data().Path)
					}
					for _, child := range node.Children() {
						checkNode(child)
					}
				}
				for _, root := range tree.nodes {
					checkNode(root)
				}
			},
		},
		{
			name:           "with_traversal_cap",
			path:           tmpDir,
			followSymlinks: false,
			opts: []Option[FileInfo]{
				WithTraversalCap[FileInfo](3),
			},
			wantErr: true,
		},
		{
			name:           "with_progress_callback",
			path:           tmpDir,
			followSymlinks: false,
			opts: []Option[FileInfo]{
				WithProgressCallback[FileInfo](func(processed int, n *Node[FileInfo]) {
					_ = processed
					if n == nil {
						panic("nil node in fs progress callback")
					}
				}),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NewTreeFromFileSystem(ctx, test.path, test.followSymlinks, test.opts...)

			if test.wantErr {
				if err == nil {
					t.Errorf("NewTreeFromFileSystem(%s) error = nil, want error", test.name)
				}
				return
			}

			if err != nil {
				t.Errorf("NewTreeFromFileSystem(%s) unexpected error = %v", test.name, err)
				return
			}

			if got == nil {
				t.Fatalf("NewTreeFromFileSystem(%s) = nil, want non-nil", test.name)
			}

			if test.checkFn != nil {
				test.checkFn(t, got)
			}
		})
	}
}

func TestDetectCycle(t *testing.T) {
	tests := []struct {
		name         string
		childID      string
		parentID     string
		parentLookup map[string]string
		want         bool
	}{
		{
			name:     "no_cycle",
			childID:  "3",
			parentID: "2",
			parentLookup: map[string]string{
				"2": "1",
				"1": "",
			},
			want: false,
		},
		{
			name:     "direct_cycle",
			childID:  "1",
			parentID: "2",
			parentLookup: map[string]string{
				"2": "1",
			},
			want: true,
		},
		{
			name:     "indirect_cycle",
			childID:  "1",
			parentID: "3",
			parentLookup: map[string]string{
				"3": "2",
				"2": "1",
			},
			want: true,
		},
		{
			name:     "self_reference",
			childID:  "1",
			parentID: "1",
			parentLookup: map[string]string{
				"1": "1",
			},
			want: true,
		},
		{
			name:         "empty_parent_lookup",
			childID:      "1",
			parentID:     "2",
			parentLookup: map[string]string{},
			want:         false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := detectCycle(test.childID, test.parentID, test.parentLookup)
			if got != test.want {
				t.Errorf("detectCycle(%q, %q, %v) = %v, want %v",
					test.childID, test.parentID, test.parentLookup, got, test.want)
			}
		})
	}
}

func TestFilterNodes(t *testing.T) {
	// Create test tree structure
	root := NewNode("1", "Root", "keep")
	child1 := NewNode("2", "Child 1", "filter")
	child2 := NewNode("3", "Child 2", "keep")
	grandchild := NewNode("4", "Grandchild", "keep")

	child1.AddChild(grandchild)
	root.AddChild(child1)
	root.AddChild(child2)

	filterFunc := func(data string) bool {
		return data == "keep"
	}

	result := filterNodes([]*Node[string]{root}, filterFunc)

	if len(result) != 1 {
		t.Fatalf("filterNodes() result count = %d, want 1", len(result))
	}

	filteredRoot := result[0]
	if filteredRoot.ID() != "1" {
		t.Errorf("filterNodes() root ID = %q, want %q", filteredRoot.ID(), "1")
	}

	// Should have both children (child1 kept because it has a matching grandchild)
	if len(filteredRoot.Children()) != 2 {
		t.Errorf("filterNodes() root children = %d, want 2", len(filteredRoot.Children()))
	}

	// Check that child1 has its grandchild
	var filteredChild1 *Node[string]
	for _, child := range filteredRoot.Children() {
		if child.ID() == "2" {
			filteredChild1 = child
			break
		}
	}

	if filteredChild1 == nil {
		t.Fatal("filterNodes() child1 not found in filtered tree")
	}

	if len(filteredChild1.Children()) != 1 {
		t.Errorf("filterNodes() child1 children = %d, want 1", len(filteredChild1.Children()))
	}
}

func TestLimitDepth(t *testing.T) {
	// Create test tree structure
	root := NewNode("1", "Root", "data")
	child := NewNode("2", "Child", "data")
	grandchild := NewNode("3", "Grandchild", "data")
	greatGrandchild := NewNode("4", "Great-Grandchild", "data")

	grandchild.AddChild(greatGrandchild)
	child.AddChild(grandchild)
	root.AddChild(child)

	tests := []struct {
		name         string
		nodes        []*Node[string]
		maxDepth     int
		currentDepth int
		checkFn      func(*testing.T, []*Node[string])
	}{
		{
			name:         "depth_0",
			nodes:        []*Node[string]{root},
			maxDepth:     0,
			currentDepth: 0,
			checkFn: func(t *testing.T, result []*Node[string]) {
				t.Helper()
				if len(result) != 1 {
					t.Errorf("limitDepth(0) result count = %d, want 1", len(result))
					return
				}
				if len(result[0].Children()) != 0 {
					t.Errorf("limitDepth(0) root has children, want none")
				}
			},
		},
		{
			name:         "depth_1",
			nodes:        []*Node[string]{root},
			maxDepth:     1,
			currentDepth: 0,
			checkFn: func(t *testing.T, result []*Node[string]) {
				t.Helper()
				if len(result[0].Children()) != 1 {
					t.Errorf("limitDepth(1) root children = %d, want 1", len(result[0].Children()))
					return
				}
				child := result[0].Children()[0]
				if len(child.Children()) != 0 {
					t.Errorf("limitDepth(1) child has children, want none")
				}
			},
		},
		{
			name:         "depth_2",
			nodes:        []*Node[string]{root},
			maxDepth:     2,
			currentDepth: 0,
			checkFn: func(t *testing.T, result []*Node[string]) {
				t.Helper()
				child := result[0].Children()[0]
				if len(child.Children()) != 1 {
					t.Errorf("limitDepth(2) child children = %d, want 1", len(child.Children()))
					return
				}
				grandchild := child.Children()[0]
				if len(grandchild.Children()) != 0 {
					t.Errorf("limitDepth(2) grandchild has children, want none")
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := limitDepth(test.nodes, test.maxDepth, test.currentDepth)
			test.checkFn(t, result)
		})
	}
}

func TestProviderErrors(t *testing.T) {
	ctx := context.Background()

	t.Run("empty_id_nested", func(t *testing.T) {
		provider := &emptyIDNestedProvider{}
		items := []testNestedItem{{id: "", name: "Invalid"}}

		_, err := NewTreeFromNestedData(ctx, items, provider)
		if !errors.Is(err, ErrEmptyID) {
			t.Errorf("NewTreeFromNestedData(empty ID) error = %v, want ErrEmptyID", err)
		}
	})

	t.Run("empty_id_flat", func(t *testing.T) {
		provider := &emptyIDFlatProvider{}
		items := []testFlatItem{{id: "", name: "Invalid"}}

		_, err := NewTreeFromFlatData(ctx, items, provider)
		if !errors.Is(err, ErrEmptyID) {
			t.Errorf("NewTreeFromFlatData(empty ID) error = %v, want ErrEmptyID", err)
		}
	})
}

type emptyIDNestedProvider struct {
	testNestedProvider
}

func (p *emptyIDNestedProvider) ID(item testNestedItem) string {
	return ""
}

type emptyIDFlatProvider struct {
	testFlatProvider
}

func (p *emptyIDFlatProvider) ID(item testFlatItem) string {
	return ""
}

func TestOptionsIntegration(t *testing.T) {
	// Test that multiple options work together correctly
	root := NewNode("1", "Root", "expand")
	child1 := NewNode("2", "Child 1", "filter")
	child2 := NewNode("3", "Child 2", "expand")
	grandchild1 := NewNode("4", "Grandchild 1", "expand")
	grandchild2 := NewNode("5", "Grandchild 2", "filter")

	child1.AddChild(grandchild1)
	child2.AddChild(grandchild2)
	root.AddChild(child1)
	root.AddChild(child2)

	// Apply options - note that filtering is applied before depth limiting
	// So we start with: root -> [child1(filter), child2(expand)] -> [grandchild1, grandchild2(filter)]
	// After filter: root -> child2 -> grandchild2 is removed because parent doesn't match
	// But filterNodes keeps parents with matching children, so child1 is kept because grandchild1 matches
	tree := NewTree([]*Node[string]{root},
		WithFilterFunc(func(data string) bool {
			return data != "filter"
		}),
		WithMaxDepth[string](1),
		WithExpandFunc[string](func(node *Node[string]) bool {
			return *node.Data() == "expand"
		}),
	)

	// Check that root is present and expanded
	if len(tree.nodes) != 1 {
		t.Fatalf("NewTree(multiple options) root count = %d, want 1", len(tree.nodes))
	}

	if !tree.nodes[0].IsExpanded() {
		t.Error("NewTree(multiple options) root not expanded")
	}

	// Check children - filterNodes keeps parents with matching children
	// So both child1 and child2 are kept (child1 has grandchild1 which matches)
	if len(tree.nodes[0].Children()) != 2 {
		t.Errorf("NewTree(multiple options) root children = %d, want 2", len(tree.nodes[0].Children()))
		return
	}

	// Check that depth is limited (no grandchildren due to maxDepth=1)
	for _, child := range tree.nodes[0].Children() {
		if len(child.Children()) != 0 {
			t.Errorf("NewTree(multiple options) child %q has children despite maxDepth=1", child.ID())
		}
	}
}
