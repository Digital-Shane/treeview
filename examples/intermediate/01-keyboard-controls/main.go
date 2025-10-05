package main

import (
	"fmt"
	"os"

	"github.com/Digital-Shane/treeview"
	"github.com/charmbracelet/bubbletea"
)

func main() {
	// Create mock file/folder structure
	nodes := createMockFileSystem()

	// Create tree with comprehensive file system styling using the new provider
	provider := treeview.NewFileNodeProvider[string]()

	// Create tree first
	tree := treeview.NewTree(nodes, treeview.WithProvider[string](provider))

	// Set up TUI model using functional options
	model := treeview.NewTuiTreeModel(
		tree,
		treeview.WithTuiWidth[string](80),
		treeview.WithTuiHeight[string](25),
	)

	// Create the program with navigation help bar
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run the program directly
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

// createMockFileSystem creates a fake file/folder structure for demonstration
func createMockFileSystem() []*treeview.Node[string] {
	// Create root project directory
	root := treeview.NewNodeSimple("project", "My Project")

	// Source code directory
	src := treeview.NewNodeSimple("src", "src")
	src.AddChild(treeview.NewNodeSimple("main.go", "main.go"))
	src.AddChild(treeview.NewNodeSimple("utils.go", "utils.go"))
	src.AddChild(treeview.NewNodeSimple("handlers.go", "handlers.go"))

	// Documentation directory
	docs := treeview.NewNodeSimple("docs", "docs")
	docs.AddChild(treeview.NewNodeSimple("README.md", "README.md"))
	docs.AddChild(treeview.NewNodeSimple("API.md", "API.md"))
	docs.AddChild(treeview.NewNodeSimple("CONTRIBUTING.md", "CONTRIBUTING.md"))

	// Web assets directory
	web := treeview.NewNodeSimple("web", "web")
	web.AddChild(treeview.NewNodeSimple("index.html", "index.html"))
	web.AddChild(treeview.NewNodeSimple("styles.css", "styles.css"))
	web.AddChild(treeview.NewNodeSimple("script.js", "script.js"))

	// Config files
	configs := treeview.NewNodeSimple("config", "config")
	configs.AddChild(treeview.NewNodeSimple("app.yaml", "app.yaml"))
	configs.AddChild(treeview.NewNodeSimple("database.json", "database.json"))

	// Tests directory
	tests := treeview.NewNodeSimple("tests", "tests")
	tests.AddChild(treeview.NewNodeSimple("main_test.go", "main_test.go"))
	tests.AddChild(treeview.NewNodeSimple("utils_test.go", "utils_test.go"))

	// Add subdirectories to root
	root.AddChild(src)
	root.AddChild(docs)
	root.AddChild(web)
	root.AddChild(configs)
	root.AddChild(tests)

	// Add some root files
	root.AddChild(treeview.NewNodeSimple("go.mod", "go.mod"))
	root.AddChild(treeview.NewNodeSimple("go.sum", "go.sum"))
	root.AddChild(treeview.NewNodeSimple(".gitignore", ".gitignore"))
	root.AddChild(treeview.NewNodeSimple("Makefile", "Makefile"))

	// Expand some directories by default for better demo
	root.Expand()
	src.Expand()
	docs.Expand()

	return []*treeview.Node[string]{root}
}
