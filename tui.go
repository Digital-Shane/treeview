package treeview

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// TuiTreeModelOption represents a functional Option that configures a TuiTreeModel instance directly.
type TuiTreeModelOption[T any] func(*TuiTreeModel[T])

// WithTuiWidth sets the viewport width.
func WithTuiWidth[T any](w int) TuiTreeModelOption[T] {
	return func(m *TuiTreeModel[T]) { m.width = w }
}

// WithTuiHeight sets the viewport height.
func WithTuiHeight[T any](h int) TuiTreeModelOption[T] {
	return func(m *TuiTreeModel[T]) { m.height = h }
}

// WithTuiAllowResize sets whether the viewport can be resized. Defaults to true.
func WithTuiAllowResize[T any](allowResize bool) TuiTreeModelOption[T] {
	return func(m *TuiTreeModel[T]) { m.allowResize = allowResize }
}

// WithTuiNavigationTimeout sets the timeout for navigation operations.
func WithTuiNavigationTimeout[T any](d time.Duration) TuiTreeModelOption[T] {
	return func(m *TuiTreeModel[T]) { m.navigationTimeout = d }
}

// WithTuiSearchTimeout sets the timeout for search operations.
func WithTuiSearchTimeout[T any](d time.Duration) TuiTreeModelOption[T] {
	return func(m *TuiTreeModel[T]) { m.searchTimeout = d }
}

// WithTuiKeyMap allows callers to override the key bindings used by the model.
func WithTuiKeyMap[T any](k KeyMap) TuiTreeModelOption[T] {
	return func(m *TuiTreeModel[T]) { m.keyMap = k }
}

// WithTuiDisableNavBar disables the built-in navigation bar at the bottom of the view.
func WithTuiDisableNavBar[T any](disable bool) TuiTreeModelOption[T] {
	return func(m *TuiTreeModel[T]) { m.disableNavBar = disable }
}

// KeyMap groups key bindings for the interactive TUI. Provide your own via
// WithTuiKeyMap if you need to accommodate non-US layouts or match existing shortcuts.
type KeyMap struct {
	// Navigation keys
	Quit     []string
	Up       []string
	Down     []string
	Expand   []string
	Collapse []string
	Toggle   []string
	Reset    []string

	// Multi-focus keys
	ExtendUp   []string
	ExtendDown []string

	// Search keys
	SearchStart  []string
	SearchAccept []string
	SearchCancel []string
	SearchDelete []string
}

// DefaultKeyMap returns a map of basic key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Navigation
		Quit:   []string{"esc"},
		Up:     []string{"up"},
		Down:   []string{"down"},
		Toggle: []string{"right", "left"},
		Reset:  []string{"ctrl+r"},

		// Multi-focus
		ExtendUp:   []string{"shift+up"},
		ExtendDown: []string{"shift+down"},

		// Search
		SearchStart:  []string{"enter"},
		SearchAccept: []string{"enter"},
		SearchCancel: []string{"esc"},
		SearchDelete: []string{"backspace", "delete"},
	}
}

// TuiTreeModel wraps a Tree and exposes it through a Bubble Tea model. It
// handles keyboard navigation, search, and viewport resizing.
//
// Concurrency: All mutating operations are executed within the Tea event loop
// which is single-threaded, so internal state does not need extra locking.
type TuiTreeModel[T any] struct {
	*Tree[T]
	keyMap KeyMap

	width       int
	height      int
	allowResize bool
	viewport    *viewport.Model

	searchTerm string
	showSearch bool

	navigationTimeout time.Duration
	searchTimeout     time.Duration

	disableNavBar bool
}

// NewTuiTreeModel creates an interactive Bubble Tea TUI model using functional options.
// The zero-value configuration applies sensible defaults:
//   - 100ms navigation timeout
//   - 300ms search timeout
//   - DefaultKeyMap for key bindings
//   - DefaultNodeProvider if none specified
//
// Example:
//
//	tree := treeview.NewTree(nodes, treeview.WithSearcher(customSearcher))
//	model := treeview.NewTuiTreeModel(
//	    tree,
//	    treeview.WithTuiWidth(80),
//	    treeview.WithTuiHeight(25),
//	)
func NewTuiTreeModel[T any](tree *Tree[T], opts ...TuiTreeModelOption[T]) *TuiTreeModel[T] {
	// Initialize the TUI model with default components
	vp := viewport.New(80, 24)
	m := &TuiTreeModel[T]{
		Tree:   tree,
		keyMap: DefaultKeyMap(),

		width:       80,
		height:      24,
		allowResize: true,
		viewport:    &vp,

		searchTerm: "",
		showSearch: false,

		searchTimeout:     300 * time.Millisecond,
		navigationTimeout: 100 * time.Millisecond,

		disableNavBar: false,
	}

	// Apply any provided options
	for _, opt := range opts {
		if opt != nil {
			opt(m)
		}
	}

	// Initialize viewport dimensions
	m.updateViewportDimensions()

	return m
}

func (m *TuiTreeModel[T]) execWithNavigationTimeout(fn func(context.Context) error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.navigationTimeout)
	defer cancel()
	_ = fn(ctx)
}

// Init initializes the TUI model. Required by the Bubble Tea model interface.
func (m *TuiTreeModel[T]) Init() tea.Cmd {
	return nil
}

// Update processes Bubble Tea messages and returns updated model and commands.
func (m *TuiTreeModel[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle different message types from Bubble Tea
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Process keyboard input
		return m.handleKeypress(msg)

	case tea.WindowSizeMsg:
		// If resize is not allowed, do nothing
		if !m.allowResize {
			return m, nil
		}
		// Terminal was resized, update our dimensions
		m.width = msg.Width
		m.height = msg.Height

		// Recalculate viewport to fit new window
		m.updateViewportDimensions()

		// No command to return
		return m, nil

	default:
		// Ignore other message types
		return m, nil
	}
}

func (m *TuiTreeModel[T]) handleKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// In search mode: prioritize search keys
	if m.showSearch {
		switch {
		case slices.Contains(m.keyMap.SearchAccept, key):
			m.showSearch = false
			m.updateViewportDimensions()
			return m, nil

		case slices.Contains(m.keyMap.SearchCancel, key):
			m.EndSearch()
			return m, nil

		case slices.Contains(m.keyMap.SearchDelete, key):
			if len(m.searchTerm) > 0 {
				m.searchTerm = m.searchTerm[:len(m.searchTerm)-1]
				m.Search(m.searchTerm)
			}
			return m, nil

		case slices.Contains(m.keyMap.Reset, key):
			m.EndSearch()
			return m, nil
		}

		// Add printable characters to search
		if len(key) == 1 && key >= " " && key <= "~" {
			m.searchTerm += key
			m.Search(m.searchTerm)
			return m, nil
		}
	}

	// Normal mode: navigation logic
	switch {
	case slices.Contains(m.keyMap.Quit, key):
		return m, tea.Quit
	case slices.Contains(m.keyMap.Up, key):
		m.NavigateUp()
		return m, nil
	case slices.Contains(m.keyMap.Down, key):
		m.NavigateDown()
		return m, nil
	case slices.Contains(m.keyMap.ExtendUp, key):
		m.ExtendFocusUp()
		return m, nil
	case slices.Contains(m.keyMap.ExtendDown, key):
		m.ExtendFocusDown()
		return m, nil
	case slices.Contains(m.keyMap.Expand, key):
		m.Expand()
		return m, nil
	case slices.Contains(m.keyMap.Collapse, key):
		m.Collapse()
		return m, nil
	case slices.Contains(m.keyMap.Toggle, key):
		m.Toggle()
		return m, nil
	case slices.Contains(m.keyMap.SearchStart, key):
		m.BeginSearch()
		return m, nil
	case slices.Contains(m.keyMap.Reset, key):
		m.ShowAll(context.Background())
		return m, nil
	}

	return m, nil
}

// NavigateUp moves the focus one visible node up.
func (m *TuiTreeModel[T]) NavigateUp() {
	m.execWithNavigationTimeout(func(ctx context.Context) error {
		_, err := m.Move(ctx, -1)
		return err
	})
}

// NavigateDown moves the focus one visible node down.
func (m *TuiTreeModel[T]) NavigateDown() {
	m.execWithNavigationTimeout(func(ctx context.Context) error {
		_, err := m.Move(ctx, 1)
		return err
	})
}

// ExtendFocusUp extends the multi-focus selection upward by one node.
func (m *TuiTreeModel[T]) ExtendFocusUp() {
	m.execWithNavigationTimeout(func(ctx context.Context) error {
		_, err := m.MoveExtend(ctx, -1)
		return err
	})
}

// ExtendFocusDown extends the multi-focus selection downward by one node.
func (m *TuiTreeModel[T]) ExtendFocusDown() {
	m.execWithNavigationTimeout(func(ctx context.Context) error {
		_, err := m.MoveExtend(ctx, 1)
		return err
	})
}

// Toggle expands or collapses all currently focused nodes.
func (m *TuiTreeModel[T]) Toggle() {
	m.execWithNavigationTimeout(func(ctx context.Context) error {
		for info, err := range m.AllFocused(ctx) {
			if err != nil {
				return err
			}
			info.Node.Toggle()
		}
		return nil
	})
}

// Expand expands all currently focused nodes to show their children.
func (m *TuiTreeModel[T]) Expand() {
	m.execWithNavigationTimeout(func(ctx context.Context) error {
		for info, err := range m.AllFocused(ctx) {
			if err != nil {
				return err
			}
			info.Node.Expand()
		}
		return nil
	})
}

// Collapse collapses all currently focused nodes to hide their children.
func (m *TuiTreeModel[T]) Collapse() {
	m.execWithNavigationTimeout(func(ctx context.Context) error {
		for info, err := range m.AllFocused(ctx) {
			if err != nil {
				return err
			}
			info.Node.Collapse()
		}
		return nil
	})
}

// BeginSearch switches the model into search mode and clears previous term.
func (m *TuiTreeModel[T]) BeginSearch() {
	m.showSearch = true
	m.searchTerm = ""

	m.updateViewportDimensions()
}

// EndSearch exits search mode and clears highlights.
func (m *TuiTreeModel[T]) EndSearch() {
	m.showSearch = false
	m.searchTerm = ""
	m.updateViewportDimensions()

	m.execWithNavigationTimeout(func(ctx context.Context) error {
		return m.ShowAll(ctx)
	})
}

// Search updates the term live as the user types and expands the tree so that
// matches are visible. It returns the result slice so external code can, for
// instance, display the hit count.
func (m *TuiTreeModel[T]) Search(term string) ([]*Node[T], error) {
	// Update the current search term
	m.searchTerm = term

	ctx, cancel := context.WithTimeout(context.Background(), m.searchTimeout)
	defer cancel()

	matches, err := m.SearchAndExpand(ctx, term)
	return matches, err
}

// View renders the tree plus an optional search bar and navigation legend.
func (m *TuiTreeModel[T]) View() string {
	// Render the tree
	result, err := renderTreeWithViewport(context.Background(), m.Tree, m.viewport)
	if err != nil {
		return "Error rendering tree: " + err.Error()
	}

	// Add search UI at the top if in search mode
	if m.showSearch {
		searchUI := "Search: " + m.searchTerm
		result = searchUI + "\n\n" + result
	}

	// Add navigation bar if not disabled
	if !m.disableNavBar {
		result += "\n───────────────────────────────────────────────────────────────\n"
		result += m.NavBar()
	}

	return result
}

// NavBar returns the navigation bar string that shows available keyboard commands.
// This method is exposed so users can create custom navigation bars or extend the default one.
func (m *TuiTreeModel[T]) NavBar() string {
	// Collect navigation items based on current mode
	var navItems []string

	// Show tree navigation
	for _, item := range []string{
		m.addNavItem(m.keyMap.Up, "Up"),
		m.addNavItem(m.keyMap.Down, "Down"),
		m.addNavItem(m.keyMap.ExtendUp, "ExtendUp"),
		m.addNavItem(m.keyMap.ExtendDown, "ExtendDown"),
		m.addNavItem(m.keyMap.Expand, "Expand"),
		m.addNavItem(m.keyMap.Collapse, "Collapse"),
		m.addNavItem(m.keyMap.Toggle, "Toggle"),
	} {
		if item != "" {
			navItems = append(navItems, item)
		}
	}

	if m.showSearch {
		// In search mode: show search-specific actions
		navItems = append(navItems, m.addNavItem(m.keyMap.SearchAccept, "Accept"))
		navItems = append(navItems, m.addNavItem(m.keyMap.SearchCancel, "Cancel"))
	} else {
		// In normal mode, add Search Option
		navItems = append(navItems, m.addNavItem(m.keyMap.SearchStart, "Search"))
		// Add quit Option
		navItems = append(navItems, m.addNavItem(m.keyMap.Quit, "Quit"))
	}

	// Add reset Option
	navItems = append(navItems, m.addNavItem(m.keyMap.Reset, "Reset"))

	// Join all navigation items with consistent spacing
	return strings.Join(navItems, "  ")
}

func (m *TuiTreeModel[T]) addNavItem(keys []string, label string) string {
	if len(keys) > 0 {
		return fmt.Sprintf("%s: %s", strings.Join(keys, "/"), label)
	}
	return ""
}

func (m *TuiTreeModel[T]) updateViewportDimensions() {
	viewHeight := m.height
	if !m.disableNavBar {
		viewHeight -= 3
	}
	if m.showSearch {
		viewHeight -= 2
	}

	m.viewport.Width = m.width
	m.viewport.Height = viewHeight

	// Update tree truncation width to match viewport width
	m.Tree.mu.Lock()
	m.Tree.truncateWidth = m.width
	m.Tree.mu.Unlock()
}
