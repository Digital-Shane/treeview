package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Digital-Shane/treeview/v2"
	"github.com/Digital-Shane/treeview/v2/examples/shared"
)

var tempDir string

// setupTempFiles creates a temporary directory structure for demonstration
func setupTempFiles() error {
	var err error
	tempDir, err = os.MkdirTemp("", "treeview-demo-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}

	// Create directory structure
	dirs := []string{
		"src",
		"src/models",
		"src/controllers",
		"src/views",
		"tests",
		"tests/unit",
		"tests/integration",
		"docs",
		"config",
		".git",
		".git/hooks",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(tempDir, dir), 0755); err != nil {
			return fmt.Errorf("creating dir %s: %w", dir, err)
		}
	}

	// Create files with content
	files := map[string]string{
		"README.md":                          "# Example Project\n\nThis is a demo project for treeview filesystem builder.",
		"go.mod":                             "module example.com/demo\n\ngo 1.21",
		"go.sum":                             "// checksums would go here",
		"main.go":                            "package main\n\nfunc main() {\n\t// Entry point\n}",
		".gitignore":                         "vendor/\n*.log\n.DS_Store",
		".env":                               "APP_ENV=development\nDEBUG=true",
		"src/models/user.go":                 "package models\n\ntype User struct {\n\tID   int\n\tName string\n}",
		"src/models/product.go":              "package models\n\ntype Product struct {\n\tID    int\n\tName  string\n\tPrice float64\n}",
		"src/controllers/user_controller.go": "package controllers\n\n// UserController handles user requests",
		"src/views/index.html":               "<html><body>Welcome</body></html>",
		"src/views/style.css":                "body { margin: 0; padding: 0; }",
		"tests/unit/user_test.go":            "package unit\n\n// User tests",
		"tests/integration/api_test.go":      "package integration\n\n// API tests",
		"docs/api.md":                        "# API Documentation",
		"docs/setup.md":                      "# Setup Guide",
		"config/app.yaml":                    "name: demo\nversion: 1.0.0",
		"config/database.yaml":               "host: localhost\nport: 5432",
		".git/config":                        "[core]\n\trepositoryformatversion = 0",
		".git/HEAD":                          "ref: refs/heads/main",
		"Makefile":                           "build:\n\tgo build -o bin/app",
		"docker-compose.yml":                 "version: '3.8'\nservices:\n  app:\n    image: app:latest",
		"Dockerfile":                         "FROM golang:1.21\nWORKDIR /app",
	}

	for path, content := range files {
		fullPath := filepath.Join(tempDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing file %s: %w", path, err)
		}
	}

	// Create some empty files
	emptyFiles := []string{
		"src/models/order.go",
		"src/controllers/product_controller.go",
		"tests/unit/product_test.go",
		".dockerignore",
		"LICENSE",
	}

	for _, path := range emptyFiles {
		fullPath := filepath.Join(tempDir, path)
		if err := os.WriteFile(fullPath, []byte{}, 0644); err != nil {
			return fmt.Errorf("creating empty file %s: %w", path, err)
		}
	}

	return nil
}

// cleanup removes the temporary directory
func cleanup() {
	if tempDir != "" {
		os.RemoveAll(tempDir)
	}
}

// setupSignalHandler sets up sig term and sig int handling for cleanup
func setupSignalHandler() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nCleaning up temporary files...")
		cleanup()
		os.Exit(0)
	}()
}

// //////////////////////////////////////////////////////////////////
//	Current Directory (.) Resolution
// //////////////////////////////////////////////////////////////////

func currentDirExample() (string, error) {
	tree, err := treeview.NewTreeFromFileSystem(
		context.Background(),
		".", false,
		treeview.WithMaxDepth[treeview.FileInfo](2),
		treeview.WithExpandAll[treeview.FileInfo](),
		treeview.WithProvider[treeview.FileInfo](treeview.NewFileNodeProvider[treeview.FileInfo]()),
	)
	if err != nil {
		return "", fmt.Errorf("building tree: %w", err)
	}

	rendered, err := tree.Render(context.Background())
	if err != nil {
		return "", err
	}
	return rendered, nil
}

////////////////////////////////////////////////////////////////////
//   Home Directory (~) Resolution
////////////////////////////////////////////////////////////////////

func homeDirExample() (string, error) {
	tree, err := treeview.NewTreeFromFileSystem(
		context.Background(),
		"~", // Home directory
		false,
		treeview.WithMaxDepth[treeview.FileInfo](1),
		treeview.WithExpandAll[treeview.FileInfo](),
		treeview.WithTraversalCap[treeview.FileInfo](1000),
		treeview.WithFilterFunc[treeview.FileInfo](func(item treeview.FileInfo) bool {
			// Filter to show only directories to limit output
			return item.IsDir() && !strings.HasPrefix(item.Name(), ".")
		}),
		treeview.WithProvider[treeview.FileInfo](treeview.NewFileNodeProvider[treeview.FileInfo]()),
	)

	rendered, err := tree.Render(context.Background())
	if err != nil {
		return "", err
	}
	return rendered, nil
}

////////////////////////////////////////////////////////////////////
//   Parent Directory (..) Resolution
////////////////////////////////////////////////////////////////////

func parentDirExample() (string, error) {
	tree, err := treeview.NewTreeFromFileSystem(
		context.Background(),
		"..", // Parent directory
		false,
		treeview.WithMaxDepth[treeview.FileInfo](1),
		treeview.WithExpandAll[treeview.FileInfo](),
		treeview.WithProvider[treeview.FileInfo](treeview.NewFileNodeProvider[treeview.FileInfo]()),
	)
	if err != nil {
		return "", fmt.Errorf("building parent tree: %w", err)
	}

	rendered, err := tree.Render(context.Background())
	if err != nil {
		return "", err
	}

	return rendered, nil
}

////////////////////////////////////////////////////////////////////
//   Rich Project Structure Demo
////////////////////////////////////////////////////////////////////

func richProjectStructureExample() (string, error) {
	tree, err := treeview.NewTreeFromFileSystem(
		context.Background(),
		tempDir,
		false,
		treeview.WithMaxDepth[treeview.FileInfo](4),
		treeview.WithExpandAll[treeview.FileInfo](),
		treeview.WithFilterFunc[treeview.FileInfo](func(item treeview.FileInfo) bool {
			if item.IsDir() {
				return true
			}
			name := strings.ToLower(item.Name())
			// Include various config file types
			return strings.HasSuffix(name, ".yaml") ||
				strings.HasSuffix(name, ".yml") ||
				strings.HasSuffix(name, ".json") ||
				strings.HasSuffix(name, ".toml") ||
				strings.HasSuffix(name, ".env") ||
				strings.HasSuffix(name, ".go") ||
				strings.HasSuffix(name, ".md") ||
				name == "makefile" ||
				name == "dockerfile" ||
				strings.HasPrefix(name, ".")
		}),
		treeview.WithProvider[treeview.FileInfo](treeview.NewFileNodeProvider[treeview.FileInfo]()),
	)
	if err != nil {
		return "", fmt.Errorf("building project tree: %w", err)
	}

	rendered, err := tree.Render(context.Background())
	if err != nil {
		return "", err
	}

	return rendered, nil
}

func main() {
	// Set up signal handler for cleanup
	setupSignalHandler()

	// Create temporary files in a subdirectory for the rich demo
	fmt.Println("Creating temporary demo files...")
	if err := setupTempFiles(); err != nil {
		fmt.Printf("Error setting up temp files: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	steps := []shared.ExampleStep{
		{
			Name: "Current Directory (.) Resolution",
			Func: currentDirExample,
		},
		{
			Name: "Home Directory (~) Resolution (Dirs Only)",
			Func: homeDirExample,
		},
		{
			Name: "Parent Directory (..) Resolution",
			Func: parentDirExample,
		},
		{
			Name: "Rich Project Structure Demo",
			Func: richProjectStructureExample,
		},
	}

	shared.RunExampleStepsWithDelay("FileSystem Builder", steps, 4)
	shared.WaitDelay(4)
}
