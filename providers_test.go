package treeview

import (
	"testing"

	"charm.land/lipgloss/v2"
)

// Test helpers for creating test nodes
func makeTestNode[T any](id string, data T) *Node[T] {
	return &Node[T]{
		id:   id,
		data: data,
	}
}

func makeTestNodeWithName[T any](id, name string, data T) *Node[T] {
	n := makeTestNode(id, data)
	n.name = name
	return n
}

func makeExpandedNode[T any](id string, data T) *Node[T] {
	n := makeTestNode(id, data)
	n.expanded = true
	return n
}

func makeCollapsedNode[T any](id string, data T) *Node[T] {
	n := makeTestNode(id, data)
	n.expanded = false
	return n
}

// Mock types for testing
type fileData struct {
	isDir bool
}

func (f fileData) IsDir() bool {
	return f.isDir
}

// Test NodeProvider interface implementation
func TestDefaultNodeProviderImplementsInterface(t *testing.T) {
	// This test ensures DefaultNodeProvider implements NodeProvider interface
	var _ NodeProvider[fileData] = &DefaultNodeProvider[fileData]{}
}

func TestNewDefaultNodeProvider(t *testing.T) {
	tests := []struct {
		name string
		opts []ProviderOption[fileData]
	}{
		{
			name: "no_options",
			opts: nil,
		},
		{
			name: "with_disable_icons",
			opts: []ProviderOption[fileData]{
				func(p *DefaultNodeProvider[fileData]) {
					p.DisableIcons = true
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := NewDefaultNodeProvider(test.opts...)
			if p == nil {
				t.Errorf("NewDefaultNodeProvider() = nil, want non-nil")
			}
		})
	}
}

func TestDefaultNodeProvider_Icon(t *testing.T) {
	tests := []struct {
		name     string
		provider *DefaultNodeProvider[fileData]
		node     *Node[fileData]
		want     string
	}{
		{
			name: "disabled_icons",
			provider: &DefaultNodeProvider[fileData]{
				DisableIcons: true,
			},
			node: makeTestNode("test", fileData{}),
			want: "  ",
		},
		{
			name: "no_rules_no_icon",
			provider: &DefaultNodeProvider[fileData]{
				DisableIcons: false,
			},
			node: makeTestNode("test", fileData{}),
			want: "",
		},
		{
			name: "matching_rule",
			provider: &DefaultNodeProvider[fileData]{
				DisableIcons: false,
				iconRules: []iconRule[fileData]{
					{
						predicate: func(n *Node[fileData]) bool { return true },
						icon:      "📁",
					},
				},
			},
			node: makeTestNode("test", fileData{}),
			want: "📁",
		},
		{
			name: "first_matching_rule_wins",
			provider: &DefaultNodeProvider[fileData]{
				DisableIcons: false,
				iconRules: []iconRule[fileData]{
					{
						predicate: func(n *Node[fileData]) bool { return n.data.isDir },
						icon:      "📁",
					},
					{
						predicate: func(n *Node[fileData]) bool { return true },
						icon:      "📄",
					},
				},
			},
			node: makeTestNode("test", fileData{isDir: true}),
			want: "📁",
		},
		{
			name: "no_match_then_fallback",
			provider: &DefaultNodeProvider[fileData]{
				DisableIcons: false,
				iconRules: []iconRule[fileData]{
					{
						predicate: func(n *Node[fileData]) bool { return n.data.isDir },
						icon:      "📁",
					},
					{
						predicate: func(n *Node[fileData]) bool { return true },
						icon:      "📄",
					},
				},
			},
			node: makeTestNode("test", fileData{isDir: false}),
			want: "📄",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.provider.Icon(test.node)
			if got != test.want {
				t.Errorf("Icon() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestDefaultNodeProvider_Style(t *testing.T) {
	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	focusedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("39"))
	customStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("100"))
	customFocusedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("200")).Background(lipgloss.Color("50"))

	tests := []struct {
		name      string
		provider  *DefaultNodeProvider[fileData]
		node      *Node[fileData]
		isFocused bool
		want      lipgloss.Style
	}{
		{
			name: "default_style_not_focused",
			provider: &DefaultNodeProvider[fileData]{
				defaultStyle: defaultStyle,
				focusedStyle: focusedStyle,
			},
			node:      makeTestNode("test", fileData{}),
			isFocused: false,
			want:      defaultStyle,
		},
		{
			name: "default_style_focused",
			provider: &DefaultNodeProvider[fileData]{
				defaultStyle: defaultStyle,
				focusedStyle: focusedStyle,
			},
			node:      makeTestNode("test", fileData{}),
			isFocused: true,
			want:      focusedStyle,
		},
		{
			name: "custom_rule_not_focused",
			provider: &DefaultNodeProvider[fileData]{
				defaultStyle: defaultStyle,
				focusedStyle: focusedStyle,
				styleRules: []styleRule[fileData]{
					{
						predicate:    func(n *Node[fileData]) bool { return n.data.isDir },
						style:        customStyle,
						focusedStyle: customFocusedStyle,
					},
				},
			},
			node:      makeTestNode("test", fileData{isDir: true}),
			isFocused: false,
			want:      customStyle,
		},
		{
			name: "custom_rule_focused",
			provider: &DefaultNodeProvider[fileData]{
				defaultStyle: defaultStyle,
				focusedStyle: focusedStyle,
				styleRules: []styleRule[fileData]{
					{
						predicate:    func(n *Node[fileData]) bool { return n.data.isDir },
						style:        customStyle,
						focusedStyle: customFocusedStyle,
					},
				},
			},
			node:      makeTestNode("test", fileData{isDir: true}),
			isFocused: true,
			want:      customFocusedStyle,
		},
		{
			name: "no_matching_rule_fallback",
			provider: &DefaultNodeProvider[fileData]{
				defaultStyle: defaultStyle,
				focusedStyle: focusedStyle,
				styleRules: []styleRule[fileData]{
					{
						predicate:    func(n *Node[fileData]) bool { return n.data.isDir },
						style:        customStyle,
						focusedStyle: customFocusedStyle,
					},
				},
			},
			node:      makeTestNode("test", fileData{isDir: false}),
			isFocused: false,
			want:      defaultStyle,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.provider.Style(test.node, test.isFocused)
			// Compare styles by rendering them since lipgloss.Style doesn't have equality
			if got.String() != test.want.String() {
				t.Errorf("Style() = %v, want %v", got.String(), test.want.String())
			}
		})
	}
}

func TestDefaultNodeProvider_Format(t *testing.T) {
	tests := []struct {
		name     string
		provider *DefaultNodeProvider[fileData]
		node     *Node[fileData]
		want     string
	}{
		{
			name:     "no_formatters_uses_node_name",
			provider: &DefaultNodeProvider[fileData]{},
			node:     makeTestNodeWithName("id1", "MyNode", fileData{}),
			want:     "MyNode",
		},
		{
			name: "first_matching_formatter_wins",
			provider: &DefaultNodeProvider[fileData]{
				formatters: []func(node *Node[fileData]) (string, bool){
					func(n *Node[fileData]) (string, bool) {
						if n.data.isDir {
							return "Directory: " + n.Name(), true
						}
						return "", false
					},
					func(n *Node[fileData]) (string, bool) {
						return "File: " + n.Name(), true
					},
				},
			},
			node: makeTestNodeWithName("id1", "MyDir", fileData{isDir: true}),
			want: "Directory: MyDir",
		},
		{
			name: "no_formatter_matches_uses_name",
			provider: &DefaultNodeProvider[fileData]{
				formatters: []func(node *Node[fileData]) (string, bool){
					func(n *Node[fileData]) (string, bool) {
						if n.data.isDir {
							return "Directory: " + n.Name(), true
						}
						return "", false
					},
				},
			},
			node: makeTestNodeWithName("id1", "MyFile", fileData{isDir: false}),
			want: "MyFile",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.provider.Format(test.node)
			if got != test.want {
				t.Errorf("Format() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestSetDefaultStyle(t *testing.T) {
	p := &DefaultNodeProvider[fileData]{}
	newStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("123"))

	p.SetDefaultStyle(newStyle)

	if p.defaultStyle.String() != newStyle.String() {
		t.Errorf("SetDefaultStyle() didn't update style: got %v, want %v",
			p.defaultStyle.String(), newStyle.String())
	}
}

func TestSetFocusedStyle(t *testing.T) {
	p := &DefaultNodeProvider[fileData]{}
	newStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("456")).Background(lipgloss.Color("789"))

	p.SetFocusedStyle(newStyle)

	if p.focusedStyle.String() != newStyle.String() {
		t.Errorf("SetFocusedStyle() didn't update style: got %v, want %v",
			p.focusedStyle.String(), newStyle.String())
	}
}

func TestWithDefaultFolderRules(t *testing.T) {
	tests := []struct {
		name string
		node *Node[fileData]
		want string
	}{
		{
			name: "expanded_folder",
			node: makeExpandedNode("dir", fileData{isDir: true}),
			want: "🔽",
		},
		{
			name: "collapsed_folder",
			node: makeCollapsedNode("dir", fileData{isDir: true}),
			want: "▶️",
		},
		{
			name: "file_no_match",
			node: makeTestNode("file", fileData{isDir: false}),
			want: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := NewDefaultNodeProvider(WithDefaultFolderRules[fileData]())
			got := p.Icon(test.node)
			if got != test.want {
				t.Errorf("Icon() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestWithDefaultFolderRules_Format(t *testing.T) {
	tests := []struct {
		name string
		node *Node[fileData]
		want string
	}{
		{
			name: "folder_with_name",
			node: makeTestNodeWithName("id", "mydir", fileData{isDir: true}),
			want: "mydir/",
		},
		{
			name: "folder_without_name_uses_id",
			node: makeTestNode("folder-id", fileData{isDir: true}),
			want: "folder-id/",
		},
		{
			name: "file_uses_default",
			node: makeTestNodeWithName("id", "myfile", fileData{isDir: false}),
			want: "myfile",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := NewDefaultNodeProvider(WithDefaultFolderRules[fileData]())
			got := p.Format(test.node)
			if got != test.want {
				t.Errorf("Format() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestWithDefaultFileRules(t *testing.T) {
	p := NewDefaultNodeProvider(WithDefaultFileRules[fileData]())

	tests := []struct {
		name string
		node *Node[fileData]
		want string
	}{
		{
			name: "file_gets_icon",
			node: makeTestNode("file", fileData{isDir: false}),
			want: "📄",
		},
		{
			name: "folder_no_match",
			node: makeTestNode("dir", fileData{isDir: true}),
			want: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := p.Icon(test.node)
			if got != test.want {
				t.Errorf("Icon() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestWithDefaultIcon(t *testing.T) {
	p := NewDefaultNodeProvider(WithDefaultIcon[fileData]("🌟"))
	node := makeTestNode("test", fileData{})

	got := p.Icon(node)
	want := "🌟"
	if got != want {
		t.Errorf("Icon() = %q, want %q", got, want)
	}
}

func TestWithIconRule(t *testing.T) {
	p := NewDefaultNodeProvider(
		WithIconRule[fileData](
			func(n *Node[fileData]) bool { return n.id == "special" },
			"✨",
		),
		WithDefaultIcon[fileData]("🌟"),
	)

	tests := []struct {
		name string
		node *Node[fileData]
		want string
	}{
		{
			name: "matching_rule",
			node: makeTestNode("special", fileData{}),
			want: "✨",
		},
		{
			name: "default_icon",
			node: makeTestNode("normal", fileData{}),
			want: "🌟",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := p.Icon(test.node)
			if got != test.want {
				t.Errorf("Icon() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestWithStyleRule(t *testing.T) {
	specialStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("100"))
	specialFocusedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("200"))

	p := NewDefaultNodeProvider(
		WithStyleRule[fileData](
			func(n *Node[fileData]) bool { return n.id == "special" },
			specialStyle,
			specialFocusedStyle,
		),
	)

	tests := []struct {
		name      string
		node      *Node[fileData]
		isFocused bool
		wantStyle string
	}{
		{
			name:      "matching_rule_not_focused",
			node:      makeTestNode("special", fileData{}),
			isFocused: false,
			wantStyle: specialStyle.String(),
		},
		{
			name:      "matching_rule_focused",
			node:      makeTestNode("special", fileData{}),
			isFocused: true,
			wantStyle: specialFocusedStyle.String(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := p.Style(test.node, test.isFocused)
			if got.String() != test.wantStyle {
				t.Errorf("Style() = %v, want %v", got.String(), test.wantStyle)
			}
		})
	}
}

func TestWithFormatter(t *testing.T) {
	p := NewDefaultNodeProvider(
		WithFormatter[fileData](func(n *Node[fileData]) (string, bool) {
			if n.id == "special" {
				return "SPECIAL: " + n.Name(), true
			}
			return "", false
		}),
	)

	tests := []struct {
		name string
		node *Node[fileData]
		want string
	}{
		{
			name: "matching_formatter",
			node: makeTestNodeWithName("special", "test", fileData{}),
			want: "SPECIAL: test",
		},
		{
			name: "no_match_uses_name",
			node: makeTestNodeWithName("normal", "test", fileData{}),
			want: "test",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := p.Format(test.node)
			if got != test.want {
				t.Errorf("Format() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestNewFileNodeProvider(t *testing.T) {
	p := NewFileNodeProvider[fileData]()

	if p == nil {
		t.Errorf("NewFileNodeProvider() = nil, want non-nil")
		return
	}

	// Test that it has some rules
	if len(p.iconRules) == 0 {
		t.Errorf("NewFileNodeProvider() created provider with no icon rules")
	}
}

func TestWithFileExtensionRules(t *testing.T) {
	// Create nodes with names that have extensions
	tests := []struct {
		name string
		node *Node[fileData]
		want string
	}{
		{
			name: "go_file",
			node: makeTestNodeWithName("id", "main.go", fileData{isDir: false}),
			want: "🐹",
		},
		{
			name: "java_file",
			node: makeTestNodeWithName("id", "Main.java", fileData{isDir: false}),
			want: "☕",
		},
		{
			name: "markdown_file",
			node: makeTestNodeWithName("id", "README.md", fileData{isDir: false}),
			want: "📝",
		},
		{
			name: "python_file",
			node: makeTestNodeWithName("id", "script.py", fileData{isDir: false}),
			want: "🐍",
		},
		{
			name: "hidden_file",
			node: makeTestNodeWithName("id", ".hiddenfile", fileData{isDir: false}),
			want: "•",
		},
		{
			name: "gitignore_file",
			node: makeTestNodeWithName("id", "project.gitignore", fileData{isDir: false}),
			want: "🚫",
		},
		{
			name: "hidden_dir_expanded",
			node: func() *Node[fileData] {
				n := makeTestNodeWithName("id", ".config", fileData{isDir: true})
				n.expanded = true
				return n
			}(),
			want: "🔽",
		},
		{
			name: "hidden_dir_collapsed",
			node: func() *Node[fileData] {
				n := makeTestNodeWithName("id", ".config", fileData{isDir: true})
				n.expanded = false
				return n
			}(),
			want: "▶️",
		},
	}

	p := NewDefaultNodeProvider(WithFileExtensionRules[fileData]())

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := p.Icon(test.node)
			if got != test.want {
				t.Errorf("Icon() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestWithFileExtensionRules_Styles(t *testing.T) {
	p := NewDefaultNodeProvider(WithFileExtensionRules[fileData]())

	tests := []struct {
		name      string
		node      *Node[fileData]
		checkFunc func(lipgloss.Style) bool
	}{
		{
			name: "hidden_file_style",
			node: makeTestNodeWithName("id", ".hidden", fileData{isDir: false}),
			checkFunc: func(s lipgloss.Style) bool {
				// Hidden files should have a dimmed color
				return s.GetForeground() == lipgloss.Color("240")
			},
		},
		{
			name: "code_file_style",
			node: makeTestNodeWithName("id", "main.go", fileData{isDir: false}),
			checkFunc: func(s lipgloss.Style) bool {
				// Code files should have a special color
				return s.GetForeground() == lipgloss.Color("205")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := p.Style(test.node, false)
			if !test.checkFunc(got) {
				t.Errorf("Style() returned unexpected style: %v", got)
			}
		})
	}
}

// Test with multiple options combined
func TestProviderWithMultipleOptions(t *testing.T) {
	customIcon := "🎯"
	p := NewDefaultNodeProvider(
		WithIconRule[fileData](
			func(n *Node[fileData]) bool { return n.id == "special" },
			"✨",
		),
		WithDefaultFolderRules[fileData](),
		WithDefaultIcon[fileData](customIcon),
	)

	tests := []struct {
		name string
		node *Node[fileData]
		want string
	}{
		{
			name: "special_rule_takes_precedence",
			node: makeTestNode("special", fileData{}),
			want: "✨",
		},
		{
			name: "folder_rule",
			node: makeExpandedNode("dir", fileData{isDir: true}),
			want: "🔽",
		},
		{
			name: "default_icon_fallback",
			node: makeTestNode("normal", fileData{isDir: false}),
			want: customIcon,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := p.Icon(test.node)
			if got != test.want {
				t.Errorf("Icon() = %q, want %q", got, test.want)
			}
		})
	}
}

// Test order of rule evaluation
func TestRuleEvaluationOrder(t *testing.T) {
	p := &DefaultNodeProvider[fileData]{
		iconRules: []iconRule[fileData]{
			{
				predicate: func(n *Node[fileData]) bool { return n.id == "test" },
				icon:      "first",
			},
			{
				predicate: func(n *Node[fileData]) bool { return n.id == "test" },
				icon:      "second",
			},
			{
				predicate: func(n *Node[fileData]) bool { return true },
				icon:      "default",
			},
		},
	}

	node := makeTestNode("test", fileData{})
	got := p.Icon(node)
	want := "first"

	if got != want {
		t.Errorf("Icon() = %q, want %q (first matching rule should win)", got, want)
	}
}

// Test concurrent access safety (mentioned in interface docs)
func TestConcurrentAccess(t *testing.T) {
	p := NewDefaultNodeProvider(
		WithDefaultIcon[fileData]("🌟"),
		WithFormatter[fileData](func(n *Node[fileData]) (string, bool) {
			return "formatted: " + n.Name(), true
		}),
	)

	node := makeTestNodeWithName("test", "concurrent", fileData{})

	// Run multiple goroutines accessing the provider
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			// Call all methods
			_ = p.Icon(node)
			_ = p.Format(node)
			_ = p.Style(node, false)
			_ = p.Style(node, true)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we get here without panicking, concurrent access is safe
}
