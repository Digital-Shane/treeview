package treeview

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestNewTuiTreeModel(t *testing.T) {
	nodes := []*Node[string]{NewNode("root", "root", "root")}
	tree := NewTree(nodes)

	tests := []struct {
		name string
		opts []TuiTreeModelOption[string]
		want *TuiTreeModel[string]
	}{
		{
			name: "default_config",
			opts: nil,
			want: &TuiTreeModel[string]{
				Tree:              tree,
				keyMap:            DefaultKeyMap(),
				width:             80,
				height:            24,
				allowResize:       true,
				searchTerm:        "",
				showSearch:        false,
				navigationTimeout: 100 * time.Millisecond,
				searchTimeout:     300 * time.Millisecond,
				disableNavBar:     false,
			},
		},
		{
			name: "with_width_height",
			opts: []TuiTreeModelOption[string]{
				WithTuiWidth[string](120),
				WithTuiHeight[string](30),
			},
			want: &TuiTreeModel[string]{
				Tree:              tree,
				keyMap:            DefaultKeyMap(),
				width:             120,
				height:            30,
				allowResize:       true,
				searchTerm:        "",
				showSearch:        false,
				navigationTimeout: 100 * time.Millisecond,
				searchTimeout:     300 * time.Millisecond,
				disableNavBar:     false,
			},
		},
		{
			name: "disable_resize_and_navbar",
			opts: []TuiTreeModelOption[string]{
				WithTuiAllowResize[string](false),
				WithTuiDisableNavBar[string](true),
			},
			want: &TuiTreeModel[string]{
				Tree:              tree,
				keyMap:            DefaultKeyMap(),
				width:             80,
				height:            24,
				allowResize:       false,
				searchTerm:        "",
				showSearch:        false,
				navigationTimeout: 100 * time.Millisecond,
				searchTimeout:     300 * time.Millisecond,
				disableNavBar:     true,
			},
		},
		{
			name: "custom_timeouts",
			opts: []TuiTreeModelOption[string]{
				WithTuiNavigationTimeout[string](200 * time.Millisecond),
				WithTuiSearchTimeout[string](500 * time.Millisecond),
			},
			want: &TuiTreeModel[string]{
				Tree:              tree,
				keyMap:            DefaultKeyMap(),
				width:             80,
				height:            24,
				allowResize:       true,
				searchTerm:        "",
				showSearch:        false,
				navigationTimeout: 200 * time.Millisecond,
				searchTimeout:     500 * time.Millisecond,
				disableNavBar:     false,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := NewTuiTreeModel(tree, test.opts...)

			if diff := cmp.Diff(test.want.Tree, got.Tree, cmpopts.IgnoreUnexported(Tree[string]{})); diff != "" {
				t.Errorf("NewTuiTreeModel(%v, %v) Tree field mismatch (-want +got):\n%s", tree, test.opts, diff)
			}

			if diff := cmp.Diff(test.want.keyMap, got.keyMap); diff != "" {
				t.Errorf("NewTuiTreeModel(%v, %v) keyMap field mismatch (-want +got):\n%s", tree, test.opts, diff)
			}

			if got.width != test.want.width {
				t.Errorf("NewTuiTreeModel(%v, %v) width = %v, want %v", tree, test.opts, got.width, test.want.width)
			}

			if got.height != test.want.height {
				t.Errorf("NewTuiTreeModel(%v, %v) height = %v, want %v", tree, test.opts, got.height, test.want.height)
			}

			if got.allowResize != test.want.allowResize {
				t.Errorf("NewTuiTreeModel(%v, %v) allowResize = %v, want %v", tree, test.opts, got.allowResize, test.want.allowResize)
			}

			if got.searchTerm != test.want.searchTerm {
				t.Errorf("NewTuiTreeModel(%v, %v) searchTerm = %q, want %q", tree, test.opts, got.searchTerm, test.want.searchTerm)
			}

			if got.showSearch != test.want.showSearch {
				t.Errorf("NewTuiTreeModel(%v, %v) showSearch = %v, want %v", tree, test.opts, got.showSearch, test.want.showSearch)
			}

			if got.navigationTimeout != test.want.navigationTimeout {
				t.Errorf("NewTuiTreeModel(%v, %v) navigationTimeout = %v, want %v", tree, test.opts, got.navigationTimeout, test.want.navigationTimeout)
			}

			if got.searchTimeout != test.want.searchTimeout {
				t.Errorf("NewTuiTreeModel(%v, %v) searchTimeout = %v, want %v", tree, test.opts, got.searchTimeout, test.want.searchTimeout)
			}

			if got.disableNavBar != test.want.disableNavBar {
				t.Errorf("NewTuiTreeModel(%v, %v) disableNavBar = %v, want %v", tree, test.opts, got.disableNavBar, test.want.disableNavBar)
			}

			if got.viewport == nil {
				t.Errorf("NewTuiTreeModel(%v, %v) viewport = nil, want non-nil", tree, test.opts)
			}
		})
	}
}

func TestTuiTreeModelOptions(t *testing.T) {
	model := &TuiTreeModel[string]{}

	tests := []struct {
		name string
		opt  TuiTreeModelOption[string]
		want func(*TuiTreeModel[string]) bool
	}{
		{
			name: "WithTuiWidth",
			opt:  WithTuiWidth[string](100),
			want: func(m *TuiTreeModel[string]) bool { return m.width == 100 },
		},
		{
			name: "WithTuiHeight",
			opt:  WithTuiHeight[string](50),
			want: func(m *TuiTreeModel[string]) bool { return m.height == 50 },
		},
		{
			name: "WithTuiAllowResize_true",
			opt:  WithTuiAllowResize[string](true),
			want: func(m *TuiTreeModel[string]) bool { return m.allowResize == true },
		},
		{
			name: "WithTuiAllowResize_false",
			opt:  WithTuiAllowResize[string](false),
			want: func(m *TuiTreeModel[string]) bool { return m.allowResize == false },
		},
		{
			name: "WithTuiNavigationTimeout",
			opt:  WithTuiNavigationTimeout[string](250 * time.Millisecond),
			want: func(m *TuiTreeModel[string]) bool { return m.navigationTimeout == 250*time.Millisecond },
		},
		{
			name: "WithTuiSearchTimeout",
			opt:  WithTuiSearchTimeout[string](750 * time.Millisecond),
			want: func(m *TuiTreeModel[string]) bool { return m.searchTimeout == 750*time.Millisecond },
		},
		{
			name: "WithTuiDisableNavBar_true",
			opt:  WithTuiDisableNavBar[string](true),
			want: func(m *TuiTreeModel[string]) bool { return m.disableNavBar == true },
		},
		{
			name: "WithTuiDisableNavBar_false",
			opt:  WithTuiDisableNavBar[string](false),
			want: func(m *TuiTreeModel[string]) bool { return m.disableNavBar == false },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model = &TuiTreeModel[string]{}
			test.opt(model)

			if !test.want(model) {
				t.Errorf("%s Option did not set expected value", test.name)
			}
		})
	}
}

func TestWithTuiKeyMap(t *testing.T) {
	customKeyMap := KeyMap{
		Quit: []string{"q"},
		Up:   []string{"k"},
		Down: []string{"j"},
	}

	nodes := []*Node[string]{NewNode("root", "root", "root")}
	tree := NewTree(nodes)
	model := NewTuiTreeModel(tree, WithTuiKeyMap[string](customKeyMap))

	if diff := cmp.Diff(customKeyMap, model.keyMap); diff != "" {
		t.Errorf("WithTuiKeyMap keyMap mismatch (-want +got):\n%s", diff)
	}
}

func TestDefaultKeyMap(t *testing.T) {
	want := KeyMap{
		Quit:         []string{"esc"},
		Up:           []string{"up"},
		Down:         []string{"down"},
		Toggle:       []string{"right", "left"},
		Reset:        []string{"ctrl+r"},
		ExtendUp:     []string{"shift+up"},
		ExtendDown:   []string{"shift+down"},
		SearchStart:  []string{"enter"},
		SearchAccept: []string{"enter"},
		SearchCancel: []string{"esc"},
		SearchDelete: []string{"backspace", "delete"},
	}

	got := DefaultKeyMap()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("DefaultKeyMap() mismatch (-want +got):\n%s", diff)
	}
}

func TestTuiTreeModelInit(t *testing.T) {
	nodes := []*Node[string]{NewNode("root", "root", "root")}
	tree := NewTree(nodes)
	model := NewTuiTreeModel(tree)

	cmd := model.Init()
	if cmd != nil {
		t.Errorf("Init() = %v, want nil", cmd)
	}
}

func TestTuiTreeModelUpdate_WindowSizeMsg(t *testing.T) {
	nodes := []*Node[string]{NewNode("root", "root", "root")}
	tree := NewTree(nodes)

	tests := []struct {
		name        string
		allowResize bool
		msg         tea.WindowSizeMsg
		wantWidth   int
		wantHeight  int
	}{
		{
			name:        "resize_allowed",
			allowResize: true,
			msg:         tea.WindowSizeMsg{Width: 120, Height: 40},
			wantWidth:   120,
			wantHeight:  40,
		},
		{
			name:        "resize_disabled",
			allowResize: false,
			msg:         tea.WindowSizeMsg{Width: 120, Height: 40},
			wantWidth:   80,
			wantHeight:  24,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model := NewTuiTreeModel(tree, WithTuiAllowResize[string](test.allowResize))

			updatedModel, cmd := model.Update(test.msg)
			got, ok := updatedModel.(*TuiTreeModel[string])
			if !ok {
				t.Fatalf("Update(%v) returned wrong type, want *TuiTreeModel[string]", test.msg)
			}

			if got.width != test.wantWidth {
				t.Errorf("Update(%v) width = %v, want %v", test.msg, got.width, test.wantWidth)
			}

			if got.height != test.wantHeight {
				t.Errorf("Update(%v) height = %v, want %v", test.msg, got.height, test.wantHeight)
			}

			if cmd != nil {
				t.Errorf("Update(%v) cmd = %v, want nil", test.msg, cmd)
			}
		})
	}
}

func TestBeginSearch(t *testing.T) {
	nodes := []*Node[string]{NewNode("root", "root", "root")}
	tree := NewTree(nodes)
	model := NewTuiTreeModel(tree)

	model.BeginSearch()

	if !model.showSearch {
		t.Errorf("BeginSearch() showSearch = %v, want true", model.showSearch)
	}

	if model.searchTerm != "" {
		t.Errorf("BeginSearch() searchTerm = %q, want \"\"", model.searchTerm)
	}
}

func TestEndSearch(t *testing.T) {
	nodes := []*Node[string]{NewNode("root", "root", "root")}
	tree := NewTree(nodes)
	model := NewTuiTreeModel(tree)

	model.showSearch = true
	model.searchTerm = "test"

	model.EndSearch()

	if model.showSearch {
		t.Errorf("EndSearch() showSearch = %v, want false", model.showSearch)
	}

	if model.searchTerm != "" {
		t.Errorf("EndSearch() searchTerm = %q, want \"\"", model.searchTerm)
	}
}

func TestUpdateViewportDimensions(t *testing.T) {
	nodes := []*Node[string]{NewNode("root", "root", "root")}
	tree := NewTree(nodes)

	tests := []struct {
		name          string
		width         int
		height        int
		showSearch    bool
		disableNavBar bool
		wantWidth     int
		wantHeight    int
	}{
		{
			name:          "normal_mode",
			width:         80,
			height:        24,
			showSearch:    false,
			disableNavBar: false,
			wantWidth:     80,
			wantHeight:    21,
		},
		{
			name:          "search_mode",
			width:         80,
			height:        24,
			showSearch:    true,
			disableNavBar: false,
			wantWidth:     80,
			wantHeight:    19,
		},
		{
			name:          "no_navbar",
			width:         80,
			height:        24,
			showSearch:    false,
			disableNavBar: true,
			wantWidth:     80,
			wantHeight:    24,
		},
		{
			name:          "search_no_navbar",
			width:         100,
			height:        30,
			showSearch:    true,
			disableNavBar: true,
			wantWidth:     100,
			wantHeight:    28,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model := NewTuiTreeModel(tree,
				WithTuiWidth[string](test.width),
				WithTuiHeight[string](test.height),
				WithTuiDisableNavBar[string](test.disableNavBar),
			)

			model.showSearch = test.showSearch
			model.updateViewportDimensions()

			if model.viewport.Width != test.wantWidth {
				t.Errorf("updateViewportDimensions() viewport width = %v, want %v", model.viewport.Width, test.wantWidth)
			}

			if model.viewport.Height != test.wantHeight {
				t.Errorf("updateViewportDimensions() viewport height = %v, want %v", model.viewport.Height, test.wantHeight)
			}
		})
	}
}

func TestAddNavItem(t *testing.T) {
	nodes := []*Node[string]{NewNode("root", "root", "root")}
	tree := NewTree(nodes)
	model := NewTuiTreeModel(tree)

	tests := []struct {
		name  string
		keys  []string
		label string
		want  string
	}{
		{
			name:  "single_key",
			keys:  []string{"up"},
			label: "Up",
			want:  "up: Up",
		},
		{
			name:  "multiple_keys",
			keys:  []string{"right", "left"},
			label: "Toggle",
			want:  "right/left: Toggle",
		},
		{
			name:  "empty_keys",
			keys:  []string{},
			label: "Nothing",
			want:  "",
		},
		{
			name:  "nil_keys",
			keys:  nil,
			label: "Nothing",
			want:  "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := model.addNavItem(test.keys, test.label)
			if got != test.want {
				t.Errorf("addNavItem(%v, %q) = %q, want %q", test.keys, test.label, got, test.want)
			}
		})
	}
}

func TestNavBar(t *testing.T) {
	nodes := []*Node[string]{NewNode("root", "root", "root")}
	tree := NewTree(nodes)
	model := NewTuiTreeModel(tree)

	tests := []struct {
		name            string
		showSearch      bool
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:            "normal_mode",
			showSearch:      false,
			wantContains:    []string{"up: Up", "down: Down", "enter: Search", "esc: Quit", "ctrl+r: Reset"},
			wantNotContains: []string{"Accept", "Cancel"},
		},
		{
			name:            "search_mode",
			showSearch:      true,
			wantContains:    []string{"up: Up", "down: Down", "enter: Accept", "esc: Cancel", "ctrl+r: Reset"},
			wantNotContains: []string{"Search", "Quit"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model.showSearch = test.showSearch
			got := model.NavBar()

			for _, want := range test.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("NavBar() showSearch=%v, result %q does not contain %q", test.showSearch, got, want)
				}
			}

			for _, notWant := range test.wantNotContains {
				if strings.Contains(got, notWant) {
					t.Errorf("NavBar() showSearch=%v, result %q should not contain %q", test.showSearch, got, notWant)
				}
			}
		})
	}
}

func TestHandleKeypress_Navigation(t *testing.T) {
	nodes := []*Node[string]{
		NewNode("root", "root", "root"),
		NewNode("child1", "child1", "child1"),
		NewNode("child2", "child2", "child2"),
	}
	tree := NewTree(nodes)
	model := NewTuiTreeModel(tree)

	tests := []struct {
		name    string
		key     string
		wantCmd tea.Cmd
	}{
		{
			name:    "quit_key",
			key:     "esc",
			wantCmd: tea.Quit,
		},
		{
			name:    "up_key",
			key:     "up",
			wantCmd: nil,
		},
		{
			name:    "down_key",
			key:     "down",
			wantCmd: nil,
		},
		{
			name:    "toggle_key_right",
			key:     "right",
			wantCmd: nil,
		},
		{
			name:    "toggle_key_left",
			key:     "left",
			wantCmd: nil,
		},
		{
			name:    "search_start_key",
			key:     "enter",
			wantCmd: nil,
		},
		{
			name:    "reset_key",
			key:     "ctrl+r",
			wantCmd: nil,
		},
		{
			name:    "unknown_key",
			key:     "x",
			wantCmd: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(test.key)}

			switch test.key {
			case "esc":
				msg = tea.KeyMsg{Type: tea.KeyEsc}
			case "up":
				msg = tea.KeyMsg{Type: tea.KeyUp}
			case "down":
				msg = tea.KeyMsg{Type: tea.KeyDown}
			case "right":
				msg = tea.KeyMsg{Type: tea.KeyRight}
			case "left":
				msg = tea.KeyMsg{Type: tea.KeyLeft}
			case "enter":
				msg = tea.KeyMsg{Type: tea.KeyEnter}
			case "ctrl+r":
				msg = tea.KeyMsg{Type: tea.KeyCtrlR}
			}

			gotModel, gotCmd := model.handleKeypress(msg)

			if gotModel != model {
				t.Errorf("handleKeypress(%q) model = %v, want %v", test.key, gotModel, model)
			}

			if test.wantCmd == nil && gotCmd != nil {
				t.Errorf("handleKeypress(%q) cmd = %v, want nil", test.key, gotCmd)
			} else if test.wantCmd != nil && gotCmd == nil {
				t.Errorf("handleKeypress(%q) cmd = nil, want non-nil", test.key)
			}
		})
	}
}

func TestHandleKeypress_SearchMode(t *testing.T) {
	nodes := []*Node[string]{NewNode("root", "root", "root")}
	tree := NewTree(nodes)
	model := NewTuiTreeModel(tree)
	model.showSearch = true
	model.searchTerm = "test"

	tests := []struct {
		name           string
		key            string
		initialTerm    string
		wantShowSearch bool
		wantSearchTerm string
	}{
		{
			name:           "search_accept",
			key:            "enter",
			initialTerm:    "hello",
			wantShowSearch: false,
			wantSearchTerm: "hello",
		},
		{
			name:           "search_cancel",
			key:            "esc",
			initialTerm:    "hello",
			wantShowSearch: false,
			wantSearchTerm: "",
		},
		{
			name:           "search_delete_backspace",
			key:            "backspace",
			initialTerm:    "hello",
			wantShowSearch: true,
			wantSearchTerm: "hell",
		},
		{
			name:           "search_delete_delete",
			key:            "delete",
			initialTerm:    "hello",
			wantShowSearch: true,
			wantSearchTerm: "hell",
		},
		{
			name:           "search_delete_empty",
			key:            "backspace",
			initialTerm:    "",
			wantShowSearch: true,
			wantSearchTerm: "",
		},
		{
			name:           "reset_in_search",
			key:            "ctrl+r",
			initialTerm:    "hello",
			wantShowSearch: false,
			wantSearchTerm: "",
		},
		{
			name:           "printable_char",
			key:            "a",
			initialTerm:    "test",
			wantShowSearch: true,
			wantSearchTerm: "testa",
		},
		{
			name:           "space_char",
			key:            " ",
			initialTerm:    "test",
			wantShowSearch: true,
			wantSearchTerm: "test ",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model.showSearch = true
			model.searchTerm = test.initialTerm

			var msg tea.KeyMsg
			switch test.key {
			case "enter":
				msg = tea.KeyMsg{Type: tea.KeyEnter}
			case "esc":
				msg = tea.KeyMsg{Type: tea.KeyEsc}
			case "backspace":
				msg = tea.KeyMsg{Type: tea.KeyBackspace}
			case "delete":
				msg = tea.KeyMsg{Type: tea.KeyDelete}
			case "ctrl+r":
				msg = tea.KeyMsg{Type: tea.KeyCtrlR}
			default:
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(test.key)}
			}

			gotModel, _ := model.handleKeypress(msg)
			got, ok := gotModel.(*TuiTreeModel[string])
			if !ok {
				t.Fatalf("handleKeypress(%q) returned wrong type", test.key)
			}

			if got.showSearch != test.wantShowSearch {
				t.Errorf("handleKeypress(%q) showSearch = %v, want %v", test.key, got.showSearch, test.wantShowSearch)
			}

			if got.searchTerm != test.wantSearchTerm {
				t.Errorf("handleKeypress(%q) searchTerm = %q, want %q", test.key, got.searchTerm, test.wantSearchTerm)
			}
		})
	}
}

func TestNavigationMethods(t *testing.T) {
	nodes := []*Node[string]{
		NewNode("root", "root", "root"),
		NewNode("child1", "child1", "child1"),
		NewNode("child2", "child2", "child2"),
	}
	tree := NewTree(nodes)
	model := NewTuiTreeModel(tree)

	t.Run("NavigateUp", func(t *testing.T) {
		model.NavigateUp()
	})

	t.Run("NavigateDown", func(t *testing.T) {
		model.NavigateDown()
	})

	t.Run("Toggle", func(t *testing.T) {
		model.Toggle()
	})

	t.Run("Expand", func(t *testing.T) {
		model.Expand()
	})

	t.Run("Collapse", func(t *testing.T) {
		model.Collapse()
	})
}

func TestSearch(t *testing.T) {
	nodes := []*Node[string]{
		NewNode("root", "root", "root"),
		NewNode("apple", "apple", "apple"),
		NewNode("banana", "banana", "banana"),
		NewNode("cherry", "cherry", "cherry"),
	}
	tree := NewTree(nodes)
	model := NewTuiTreeModel(tree)

	tests := []struct {
		name       string
		searchTerm string
		wantErr    bool
	}{
		{
			name:       "empty_search",
			searchTerm: "",
			wantErr:    false,
		},
		{
			name:       "simple_search",
			searchTerm: "app",
			wantErr:    false,
		},
		{
			name:       "no_match",
			searchTerm: "xyz",
			wantErr:    false,
		},
		{
			name:       "partial_match",
			searchTerm: "an",
			wantErr:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := model.Search(test.searchTerm)

			gotErr := err != nil
			if gotErr != test.wantErr {
				t.Errorf("Search(%q) error = %v, want error %v", test.searchTerm, gotErr, test.wantErr)
			}

			if model.searchTerm != test.searchTerm {
				t.Errorf("Search(%q) searchTerm = %q, want %q", test.searchTerm, model.searchTerm, test.searchTerm)
			}
		})
	}
}

func TestView(t *testing.T) {
	nodes := []*Node[string]{
		NewNode("root", "root", "root"),
		NewNode("child1", "child1", "child1"),
		NewNode("child2", "child2", "child2"),
	}
	tree := NewTree(nodes)

	tests := []struct {
		name            string
		showSearch      bool
		searchTerm      string
		disableNavBar   bool
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:            "normal_mode",
			showSearch:      false,
			searchTerm:      "",
			disableNavBar:   false,
			wantContains:    []string{"root", "───────────────────────────────────────────────────────────────"},
			wantNotContains: []string{"Search:"},
		},
		{
			name:            "search_mode",
			showSearch:      true,
			searchTerm:      "test",
			disableNavBar:   false,
			wantContains:    []string{"Search: test", "root", "───────────────────────────────────────────────────────────────"},
			wantNotContains: []string{},
		},
		{
			name:            "no_navbar",
			showSearch:      false,
			searchTerm:      "",
			disableNavBar:   true,
			wantContains:    []string{"root"},
			wantNotContains: []string{"───────────────────────────────────────────────────────────────"},
		},
		{
			name:            "search_no_navbar",
			showSearch:      true,
			searchTerm:      "hello",
			disableNavBar:   true,
			wantContains:    []string{"Search: hello", "root"},
			wantNotContains: []string{"───────────────────────────────────────────────────────────────"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model := NewTuiTreeModel(tree, WithTuiDisableNavBar[string](test.disableNavBar))
			model.showSearch = test.showSearch
			model.searchTerm = test.searchTerm

			got := model.View()

			for _, want := range test.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("View() showSearch=%v, disableNavBar=%v, result does not contain %q",
						test.showSearch, test.disableNavBar, want)
				}
			}

			for _, notWant := range test.wantNotContains {
				if strings.Contains(got, notWant) {
					t.Errorf("View() showSearch=%v, disableNavBar=%v, result should not contain %q",
						test.showSearch, test.disableNavBar, notWant)
				}
			}

			if strings.Contains(got, "Error rendering tree") {
				t.Errorf("View() returned error: %q", got)
			}
		})
	}
}

func TestExecWithNavigationTimeout(t *testing.T) {
	nodes := []*Node[string]{NewNode("root", "root", "root")}
	tree := NewTree(nodes)
	model := NewTuiTreeModel(tree, WithTuiNavigationTimeout[string](10*time.Millisecond))

	called := false
	model.execWithNavigationTimeout(func(ctx context.Context) error {
		called = true
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	})

	if !called {
		t.Errorf("execWithNavigationTimeout() function was not called")
	}
}

func TestUpdate_UnknownMessage(t *testing.T) {
	nodes := []*Node[string]{NewNode("root", "root", "root")}
	tree := NewTree(nodes)
	model := NewTuiTreeModel(tree)

	type unknownMsg struct{}
	msg := unknownMsg{}

	gotModel, gotCmd := model.Update(msg)

	if gotModel != model {
		t.Errorf("Update(unknownMsg) model = %v, want %v", gotModel, model)
	}

	if gotCmd != nil {
		t.Errorf("Update(unknownMsg) cmd = %v, want nil", gotCmd)
	}
}
