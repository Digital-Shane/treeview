package treeview

import (
	"github.com/charmbracelet/lipgloss"
)

// NodeProvider lets you plug custom rendering logic into treeview. The generic
// parameter T represents your domain object that is stored inside each Node.
//
// A provider is responsible for supplying icon, text, and style attributes so
// the renderer can paint a line for a single node. You can return different
// styles depending on whether the node is currently focused in the TUI.
//
// All methods must be safe for concurrent use because renderers may call them
// from multiple goroutines.
type NodeProvider[T any] interface {
	// Icon returns the leading glyph (e.g. folder / file symbol) for the node.
	Icon(node *Node[T]) string

	// Format converts the node's data into a human-readable label that follows the icon.
	Format(node *Node[T]) string

	// Style supplies the lipgloss style for the node based on its focus state.
	Style(node *Node[T], isFocused bool) lipgloss.Style
}

// DefaultNodeProvider is a batteries-included implementation of
// NodeProvider that delivers a pleasant out-of-the-box look & feel.
//
// It is possible to tweak colours, icons, and type-specific styles via the
// exposed setter methods or replace the provider entirely with your own.
//
// Icon theme
// ------------
// A small internal map associates node types ("expanded", "collapsed", "file", …) with emoji
// glyphs. Call SetIcon/SetCollapsedIcon/SetExpandedIcon to override single
// entries or set DisableIcons to true to render a blank two-space placeholder
// instead.
//
// Styles
// ------
// Focused nodes get their own style variants, so they pop out during keyboard
// navigation. You can further specialise style choices per node type via
// SetTypeStyle.
type DefaultNodeProvider[T any] struct {
	defaultStyle lipgloss.Style
	focusedStyle lipgloss.Style
	formatters   []func(node *Node[T]) (string, bool)
	iconRules    []iconRule[T]
	styleRules   []styleRule[T]
	DisableIcons bool
}

type iconRule[T any] struct {
	predicate func(*Node[T]) bool
	icon      string
}

type styleRule[T any] struct {
	predicate    func(*Node[T]) bool
	style        lipgloss.Style
	focusedStyle lipgloss.Style
}

// ProviderOption is a function that configures a DefaultNodeProvider.
type ProviderOption[T any] func(*DefaultNodeProvider[T])

// NewDefaultNodeProvider returns a provider initialised with a
// reasonably neutral colour palette that should look okay on both dark and light terminals.
func NewDefaultNodeProvider[T any](opts ...ProviderOption[T]) *DefaultNodeProvider[T] {
	p := &DefaultNodeProvider[T]{
		defaultStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		focusedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("39")).
			Bold(true),
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// WithDefaultFolderRules is a provider Option that adds icon and style rules for folder nodes.
// A node is considered a folder if its data object implements the `IsDir() bool` method.
func WithDefaultFolderRules[T any]() ProviderOption[T] {
	return func(p *DefaultNodeProvider[T]) {
		// Icons
		p.iconRules = append(p.iconRules,
			iconRule[T]{
				predicate: PredAll(PredIsDir[T](), PredIsExpanded[T]()),
				icon:      "🔽",
			},
			iconRule[T]{
				predicate: PredAll(PredIsDir[T](), PredIsCollapsed[T]()),
				icon:      "▶️",
			},
		)

		// Formatter
		p.formatters = append(p.formatters, func(n *Node[T]) (string, bool) {
			if !PredIsDir[T]()(n) {
				return "", false
			}

			name := n.Name()
			if name == "" {
				name = n.ID()
			}
			return name + "/", true
		})
	}
}

// WithDefaultFileRules is a provider Option that adds a default icon for file nodes.
// A node is considered a file if it is not a folder.
func WithDefaultFileRules[T any]() ProviderOption[T] {
	return func(p *DefaultNodeProvider[T]) {
		p.iconRules = append(p.iconRules, iconRule[T]{
			predicate: PredIsFile[T](),
			icon:      "📄",
		})
	}
}

// WithDefaultIcon is a provider Option that sets the default icon for nodes that
// do not match any other icon rule.
func WithDefaultIcon[T any](icon string) ProviderOption[T] {
	return func(p *DefaultNodeProvider[T]) {
		p.iconRules = append(p.iconRules, iconRule[T]{
			predicate: func(n *Node[T]) bool { return true },
			icon:      icon,
		})
	}
}

// WithIconRule is a provider Option that adds a custom icon rule.
// Rules are evaluated in the order they are added
func WithIconRule[T any](predicate func(*Node[T]) bool, icon string) ProviderOption[T] {
	return func(p *DefaultNodeProvider[T]) {
		p.iconRules = append(p.iconRules, iconRule[T]{
			predicate: predicate,
			icon:      icon,
		})
	}
}

// WithStyleRule is a provider Option that adds a custom style rule.
func WithStyleRule[T any](predicate func(*Node[T]) bool, style, focused lipgloss.Style) ProviderOption[T] {
	return func(p *DefaultNodeProvider[T]) {
		p.styleRules = append(p.styleRules, styleRule[T]{
			predicate:    predicate,
			style:        style,
			focusedStyle: focused,
		})
	}
}

// WithFormatter adds a custom formatter for the node's label. The first
// formatter that returns true will be used.
func WithFormatter[T any](formatter func(node *Node[T]) (string, bool)) ProviderOption[T] {
	return func(p *DefaultNodeProvider[T]) {
		p.formatters = append(p.formatters, formatter)
	}
}

// NewFileNodeProvider returns a pre-configured node provider with sensible defaults for
// rendering file system trees.
func NewFileNodeProvider[T any](opts ...ProviderOption[T]) *DefaultNodeProvider[T] {
	// Default file system rules
	allOpts := []ProviderOption[T]{
		// Order matters, more specific rules go first.
		WithFileExtensionRules[T](),
		WithDefaultFolderRules[T](),
		WithDefaultFileRules[T](),
		WithDefaultIcon[T]("•"),
	}
	// User-provided options will be prepended, so they are evaluated first.
	allOpts = append(opts, allOpts...)
	return NewDefaultNodeProvider(allOpts...)
}

// Icon picks an icon for the node based on its type and expanded state.
func (p *DefaultNodeProvider[T]) Icon(node *Node[T]) string {
	if p.DisableIcons {
		return "  "
	}

	for _, rule := range p.iconRules {
		if rule.predicate(node) {
			return rule.icon
		}
	}

	return "" // Should not be reached if WithDefaultIcon is used
}

// Style returns the lipgloss style that should be applied to the main label of
// a node. Focus takes precedence over type-specific styles.
func (p *DefaultNodeProvider[T]) Style(node *Node[T], isFocused bool) lipgloss.Style {
	for _, rule := range p.styleRules {
		if rule.predicate(node) {
			if isFocused {
				return rule.focusedStyle
			}
			return rule.style
		}
	}

	// If no type-specific style is provided, use the default style.
	if isFocused {
		return p.focusedStyle
	}
	return p.defaultStyle
}

// SetDefaultStyle updates the style used when no type-specific override exists.
func (p *DefaultNodeProvider[T]) SetDefaultStyle(style lipgloss.Style) {
	p.defaultStyle = style
}

// SetFocusedStyle changes the style applied to the focused node when no type-specific override exists.
func (p *DefaultNodeProvider[T]) SetFocusedStyle(style lipgloss.Style) {
	p.focusedStyle = style
}

// Format returns the human-readable label for a node.
func (p *DefaultNodeProvider[T]) Format(node *Node[T]) string {
	for _, formatter := range p.formatters {
		if name, ok := formatter(node); ok {
			return name
		}
	}
	return node.Name()
}

// WithFileExtensionRules is a provider Option that adds icon and style rules based on file extensions.
func WithFileExtensionRules[T any]() ProviderOption[T] {
	return func(p *DefaultNodeProvider[T]) {

		// Icons
		p.iconRules = append(p.iconRules,
			// Hidden files and folders
			iconRule[T]{predicate: PredAll(PredIsHidden[T](), PredIsDir[T](), PredIsExpanded[T]()), icon: "🔽"},
			iconRule[T]{predicate: PredAll(PredIsHidden[T](), PredIsDir[T](), PredIsCollapsed[T]()), icon: "▶️"},
			iconRule[T]{predicate: PredAll(PredIsHidden[T](), PredNot(PredIsDir[T]())), icon: "•"},

			// Programming languages
			iconRule[T]{predicate: PredHasExtension[T](".go"), icon: "🐹"},
			iconRule[T]{predicate: PredHasExtension[T](".java"), icon: "☕"},
			iconRule[T]{predicate: PredHasExtension[T](".md", ".mdx"), icon: "📝"},
			iconRule[T]{predicate: PredHasExtension[T](".sh"), icon: "🐚"},
			iconRule[T]{predicate: PredHasExtension[T](".py"), icon: "🐍"},
			iconRule[T]{predicate: PredHasExtension[T](".cpp", ".c++"), icon: "⚙️"},
			iconRule[T]{predicate: PredHasExtension[T](".c"), icon: "🔬"},
			iconRule[T]{predicate: PredHasExtension[T](".h"), icon: "📋"},
			iconRule[T]{predicate: PredHasExtension[T](".hpp"), icon: "📋"},
			iconRule[T]{predicate: PredHasExtension[T](".js"), icon: "⚡"},
			iconRule[T]{predicate: PredHasExtension[T](".ts"), icon: "🔷"},
			iconRule[T]{predicate: PredHasExtension[T](".html", ".htm"), icon: "🌐"},
			iconRule[T]{predicate: PredHasExtension[T](".css"), icon: "🎨"},
			iconRule[T]{predicate: PredHasExtension[T](".sassy", ".sass", ".scss"), icon: "🎨"},
			iconRule[T]{predicate: PredHasExtension[T](".less"), icon: "🎨"},
			iconRule[T]{predicate: PredHasExtension[T](".json"), icon: "📋"},
			iconRule[T]{predicate: PredHasExtension[T](".yml", ".yaml"), icon: "⚙️"},
			iconRule[T]{predicate: PredHasExtension[T](".xml"), icon: "📄"},
			iconRule[T]{predicate: PredHasExtension[T](".toml"), icon: "📝"},
			iconRule[T]{predicate: PredHasExtension[T](".php"), icon: "🐘"},
			iconRule[T]{predicate: PredHasExtension[T](".rb"), icon: "💎"},
			iconRule[T]{predicate: PredHasExtension[T](".rs"), icon: "🦀"},
			iconRule[T]{predicate: PredHasExtension[T](".swift"), icon: "🐦"},
			iconRule[T]{predicate: PredHasExtension[T](".kt", ".kts"), icon: "🔥"},
			iconRule[T]{predicate: PredHasExtension[T](".scala"), icon: "⚖️"},
			iconRule[T]{predicate: PredHasExtension[T](".pl"), icon: "🔮"},
			iconRule[T]{predicate: PredHasExtension[T](".pm"), icon: "🔮"},
			iconRule[T]{predicate: PredHasExtension[T](".lua"), icon: "🌙"},
			iconRule[T]{predicate: PredHasExtension[T](".r"), icon: "📊"},
			iconRule[T]{predicate: PredHasExtension[T](".hs"), icon: "🔥"},
			iconRule[T]{predicate: PredHasExtension[T](".ex", ".exs"), icon: "💜"},
			iconRule[T]{predicate: PredHasExtension[T](".clj", ".cljs", ".cljc", ".edn"), icon: "⚗️"},
			iconRule[T]{predicate: PredHasExtension[T](".erl", ".hrl"), icon: "🔥"},

			// Document formats
			iconRule[T]{predicate: PredHasExtension[T](".pdf"), icon: "📕"},
			iconRule[T]{predicate: PredHasExtension[T](".doc", ".docx"), icon: "📄"},
			iconRule[T]{predicate: PredHasExtension[T](".xls", ".xlsx"), icon: "📊"},
			iconRule[T]{predicate: PredHasExtension[T](".ppt", ".pptx"), icon: "📈"},
			iconRule[T]{predicate: PredHasExtension[T](".txt"), icon: "📝"},

			// Image formats
			iconRule[T]{predicate: PredHasExtension[T](".png"), icon: "🖼"},
			iconRule[T]{predicate: PredHasExtension[T](".jpg", ".jpeg"), icon: "🖼"},
			iconRule[T]{predicate: PredHasExtension[T](".gif"), icon: "🖼"},
			iconRule[T]{predicate: PredHasExtension[T](".svg"), icon: "🖼"},
			iconRule[T]{predicate: PredHasExtension[T](".bmp"), icon: "🖼"},
			iconRule[T]{predicate: PredHasExtension[T](".ico"), icon: "🖼"},
			iconRule[T]{predicate: PredHasExtension[T](".webp"), icon: "🖼"},

			// Audio formats
			iconRule[T]{predicate: PredHasExtension[T](".mp3"), icon: "🎵"},
			iconRule[T]{predicate: PredHasExtension[T](".wav"), icon: "🎵"},
			iconRule[T]{predicate: PredHasExtension[T](".ogg"), icon: "🎵"},
			iconRule[T]{predicate: PredHasExtension[T](".flac"), icon: "🎵"},
			iconRule[T]{predicate: PredHasExtension[T](".aac"), icon: "🎵"},

			// Video formats
			iconRule[T]{predicate: PredHasExtension[T](".mp4"), icon: "📹"},
			iconRule[T]{predicate: PredHasExtension[T](".mov"), icon: "📹"},
			iconRule[T]{predicate: PredHasExtension[T](".avi"), icon: "📹"},
			iconRule[T]{predicate: PredHasExtension[T](".mkv"), icon: "📹"},
			iconRule[T]{predicate: PredHasExtension[T](".webm"), icon: "📹"},

			// Compressed formats
			iconRule[T]{predicate: PredHasExtension[T](".zip"), icon: "📦"},
			iconRule[T]{predicate: PredHasExtension[T](".rar"), icon: "📦"},
			iconRule[T]{predicate: PredHasExtension[T](".tar"), icon: "📦"},
			iconRule[T]{predicate: PredHasExtension[T](".gz"), icon: "📦"},
			iconRule[T]{predicate: PredHasExtension[T](".bz2"), icon: "📦"},
			iconRule[T]{predicate: PredHasExtension[T](".7z"), icon: "📦"},
			iconRule[T]{predicate: PredHasExtension[T](".xz"), icon: "📦"},

			// Other
			iconRule[T]{predicate: PredHasExtension[T](".git"), icon: "🌱"},
			iconRule[T]{predicate: PredHasExtension[T](".gitignore"), icon: "🚫"},
			iconRule[T]{predicate: PredHasExtension[T](".gitmodules"), icon: "🔗"},
			iconRule[T]{predicate: PredHasExtension[T](".gitattributes"), icon: "⚙️"},
			iconRule[T]{predicate: PredHasExtension[T](".dockerfile", "docker-compose.yml"), icon: "🐳"},
			iconRule[T]{predicate: PredHasExtension[T](".env"), icon: "⚙️"},
			iconRule[T]{predicate: PredHasExtension[T](".log"), icon: "📋"},
			iconRule[T]{predicate: PredHasExtension[T](".sql"), icon: "🗄"},
			iconRule[T]{predicate: PredHasExtension[T](".db"), icon: "🗄"},
			iconRule[T]{predicate: PredHasExtension[T](".sqlite", ".sqlite3"), icon: "🗄"},
			iconRule[T]{predicate: PredHasExtension[T](".bak"), icon: "🗃"},
			iconRule[T]{predicate: PredHasExtension[T](".tmp"), icon: "⏳"},
			iconRule[T]{predicate: PredHasExtension[T](".swp"), icon: "🔄"},
		)

		// Styles
		p.styleRules = append(p.styleRules,
			styleRule[T]{
				predicate: PredIsHidden[T](),
				style:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
				focusedStyle: lipgloss.NewStyle().
					Foreground(lipgloss.Color("240")).
					Background(lipgloss.Color("39")).
					Bold(true),
			},
			styleRule[T]{
				predicate: PredHasExtension[T](".go", ".java", ".md", ".sh", ".py", ".cpp", ".c", ".h", ".hpp", ".js", ".ts", ".html", ".css", ".scss", ".less", ".json", ".yml", ".yaml", ".xml", ".toml", ".php", ".rb", ".rs", ".swift", ".kt", ".kts", ".scala", ".pl", ".pm", ".lua", ".r", ".hs", ".ex", ".exs", ".clj", ".cljs", ".cljc", ".edn", ".erl", ".hrl"),
				style:     lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
				focusedStyle: lipgloss.NewStyle().
					Foreground(lipgloss.Color("205")).
					Background(lipgloss.Color("39")).
					Bold(true),
			},
		)
	}
}
