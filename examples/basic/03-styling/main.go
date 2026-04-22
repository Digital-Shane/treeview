package main

import (
	"context"
	"fmt"

	"charm.land/lipgloss/v2"

	"github.com/Digital-Shane/treeview"
	"github.com/Digital-Shane/treeview/examples/shared"
)

////////////////////////////////////////////////////////////////////
//   Theme Provider Options
////////////////////////////////////////////////////////////////////

// createDarkThemeProvider creates a DefaultNodeProvider with dark theme styling
func createDarkThemeProvider() *treeview.DefaultNodeProvider[string] {
	return treeview.NewDefaultNodeProvider(
		// Icon rules using predicate helpers
		treeview.WithIconRule(treeview.PredHasExtension[string](".go"), "🐹"),
		treeview.WithIconRule(treeview.PredHasExtension[string](".js"), "⚡"),
		treeview.WithIconRule(treeview.PredHasExtension[string](".ts"), "🔷"),
		treeview.WithIconRule(treeview.PredHasExtension[string](".py"), "🐍"),
		treeview.WithIconRule(treeview.PredHasExtension[string](".md"), "📖"),
		treeview.WithIconRule(treeview.PredHasExtension[string](".json", ".yaml", ".yml"), "📋"),
		treeview.WithIconRule(treeview.PredHasExtension[string](".txt", ".log"), "📝"),
		treeview.WithIconRule(treeview.PredNot(treeview.PredContainsText[string](".")), "📁"),
		treeview.WithDefaultIcon[string]("📄"),

		// Dark theme style rules
		treeview.WithStyleRule(
			func(n *treeview.Node[string]) bool { return true }, // Apply to all nodes
			lipgloss.NewStyle().Foreground(lipgloss.Color("#B0BEC5")),
			lipgloss.NewStyle().
				Background(lipgloss.Color("#263238")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true),
		),

		// Dark theme formatter - simple name-based formatting
		treeview.WithFormatter[string](func(node *treeview.Node[string]) (string, bool) {
			if *node.Data() != "" {
				return *node.Data(), true
			}
			return node.ID(), true
		}),
	)
}

// createRetroThemeProvider creates a DefaultNodeProvider with retro/80s theme styling
func createRetroThemeProvider() *treeview.DefaultNodeProvider[string] {
	return treeview.NewDefaultNodeProvider(
		// Retro theme icon rules with geometric symbols
		treeview.WithIconRule(treeview.PredHasExtension[string](".go"), "▶"),
		treeview.WithIconRule(treeview.PredHasExtension[string](".js"), "◉"),
		treeview.WithIconRule(treeview.PredHasExtension[string](".ts"), "◆"),
		treeview.WithIconRule(treeview.PredHasExtension[string](".py"), "●"),
		treeview.WithIconRule(treeview.PredHasExtension[string](".md"), "▪"),
		treeview.WithIconRule(treeview.PredHasExtension[string](".json", ".yaml", ".yml"), "▫"),
		treeview.WithIconRule(treeview.PredHasExtension[string](".txt", ".log"), "■"),
		treeview.WithIconRule(treeview.PredNot(treeview.PredContainsText[string](".")), "▼"),
		treeview.WithDefaultIcon[string]("◯"),

		// Retro theme style rules
		treeview.WithStyleRule(
			func(n *treeview.Node[string]) bool { return true }, // Apply to all nodes
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("#D2691E")).
				Bold(true),
			lipgloss.NewStyle().
				Background(lipgloss.Color("#8B4513")).
				Foreground(lipgloss.Color("#FFFF00")).
				Underline(true).
				Border(lipgloss.RoundedBorder()),
		),

		// Retro theme formatter
		treeview.WithFormatter[string](func(node *treeview.Node[string]) (string, bool) {
			if *node.Data() != "" {
				return *node.Data(), true
			}
			return node.ID(), true
		}),
	)
}

// createNeonThemeProvider creates a DefaultNodeProvider with neon/cyberpunk theme styling
func createNeonThemeProvider() *treeview.DefaultNodeProvider[string] {
	return treeview.NewDefaultNodeProvider(
		// Neon theme icon rules with cyber symbols
		treeview.WithIconRule(treeview.PredHasExtension[string](".go"), "★"),
		treeview.WithIconRule(treeview.PredHasExtension[string](".js"), "⚡"),
		treeview.WithIconRule(treeview.PredHasExtension[string](".ts"), "◆"),
		treeview.WithIconRule(treeview.PredHasExtension[string](".py"), "●"),
		treeview.WithIconRule(treeview.PredHasExtension[string](".md"), "▪"),
		treeview.WithIconRule(treeview.PredHasExtension[string](".json", ".yaml", ".yml"), "▫"),
		treeview.WithIconRule(treeview.PredHasExtension[string](".txt", ".log"), "■"),
		treeview.WithIconRule(treeview.PredNot(treeview.PredContainsText[string](".")), "♦"),
		treeview.WithDefaultIcon[string]("◯"),

		// Neon theme style rules
		treeview.WithStyleRule(
			func(n *treeview.Node[string]) bool { return true }, // Apply to all nodes
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FFFF")).
				Bold(true).
				Blink(true),
			lipgloss.NewStyle().
				Background(lipgloss.Color("#FF00FF")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Border(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color("#00FF00")),
		),

		// Neon theme formatter
		treeview.WithFormatter[string](func(node *treeview.Node[string]) (string, bool) {
			if *node.Data() != "" {
				return *node.Data(), true
			}
			return node.ID(), true
		}),
	)
}

////////////////////////////////////////////////////////////////////
//   Styling Example
////////////////////////////////////////////////////////////////////

func main() {
	shared.ClearTerminal()
	// Create a sample file tree to display in various styles
	root := CreateSampleFileTree()
	// Expand some nodes for better display
	root[0].Expand()
	children := root[0].Children()
	if len(children) > 0 {
		children[0].Expand()
	}

	////////////////////////////////////////////////////////////////////
	//   Default Styling (Built-in)
	////////////////////////////////////////////////////////////////////

	fmt.Println("Default styling theme (built-in):")
	tree := treeview.NewTree(root)
	output, _ := tree.Render(context.Background())
	fmt.Println(output)

	shared.WaitDelayThenClearTerminal(3.5)

	////////////////////////////////////////////////////////////////////
	//   File System Styling (Built-in)
	////////////////////////////////////////////////////////////////////

	fmt.Println("File system styling theme (built-in):")
	provider := treeview.NewFileNodeProvider[string]()
	tree = treeview.NewTree(root,
		treeview.WithProvider(provider),
	)
	output, _ = tree.Render(context.Background())
	fmt.Println(output)

	shared.WaitDelayThenClearTerminal(3.5)

	////////////////////////////////////////////////////////////////////
	//   Dark Theme
	////////////////////////////////////////////////////////////////////

	fmt.Println("Dark theme styling:")
	tree = treeview.NewTree(root,
		treeview.WithProvider(createDarkThemeProvider()),
	)
	output, _ = tree.Render(context.Background())
	fmt.Println(output)

	shared.WaitDelayThenClearTerminal(3.5)

	////////////////////////////////////////////////////////////////////
	//   Retro Theme
	////////////////////////////////////////////////////////////////////

	fmt.Println("Retro theme styling:")
	tree = treeview.NewTree(root,
		treeview.WithProvider(createRetroThemeProvider()),
	)
	output, _ = tree.Render(context.Background())
	fmt.Println(output)

	shared.WaitDelayThenClearTerminal(3.5)

	////////////////////////////////////////////////////////////////////
	//   Neon Theme
	////////////////////////////////////////////////////////////////////

	fmt.Println("Neon theme styling:")
	tree = treeview.NewTree(root,
		treeview.WithProvider(createNeonThemeProvider()),
	)
	output, _ = tree.Render(context.Background())
	fmt.Println(output)

	shared.WaitDelay(3.5)
}

// CreateSampleFileTree creates a sample file tree for demonstration purposes.
func CreateSampleFileTree() []*treeview.Node[string] {
	root := treeview.NewNodeSimple("src", "src/")
	root.SetName("src/")

	// Go files
	mainGo := treeview.NewNodeSimple("main.go", "main.go")
	mainGo.SetName("main.go")
	utilGo := treeview.NewNodeSimple("util.go", "util.go")
	utilGo.SetName("util.go")

	// JS/TS files
	appJs := treeview.NewNodeSimple("app.js", "app.js")
	appJs.SetName("app.js")
	stylesCss := treeview.NewNodeSimple("styles.css", "styles.css")
	stylesCss.SetName("styles.css")

	// Config files
	configJson := treeview.NewNodeSimple("config.json", "config.json")
	configJson.SetName("config.json")

	// Docs
	readmeMd := treeview.NewNodeSimple("README.md", "README.md")
	readmeMd.SetName("README.md")

	// Sub directory
	apiDir := treeview.NewNodeSimple("api", "api/")
	apiDir.SetName("api/")
	apiGo := treeview.NewNodeSimple("api.go", "api.go")
	apiGo.SetName("api.go")
	apiDir.AddChild(apiGo)

	// Add children to root
	root.AddChild(mainGo)
	root.AddChild(utilGo)
	root.AddChild(appJs)
	root.AddChild(stylesCss)
	root.AddChild(configJson)
	root.AddChild(readmeMd)
	root.AddChild(apiDir)

	return []*treeview.Node[string]{root}
}
