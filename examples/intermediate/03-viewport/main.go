package main

import (
	"fmt"
	"log"

	"github.com/Digital-Shane/treeview"
	"github.com/charmbracelet/bubbletea"
)

func main() {
	// Create a large, realistic file structure
	nodes := createLargeFileSystem()

	// Create provider
	provider := createViewportProvider()

	// Create tree first
	tree := treeview.NewTree(nodes, treeview.WithProvider[*FileData](provider))

	// Set up TUI model with functional options
	model := treeview.NewTuiTreeModel(
		tree,
		treeview.WithTuiWidth[*FileData](80),
		treeview.WithTuiHeight[*FileData](20),
	)

	// Expand root to show content immediately
	if len(nodes) > 0 {
		root := nodes[0]
		root.Expand()
	}

	// Create the program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

// FileData represents file information compatible with the file system provider
type FileData struct {
	name  string
	isDir bool
	size  int64 // Added for file size
}

// Name returns the file name
func (f *FileData) Name() string {
	return f.name
}

// IsDir returns whether this is a directory
func (f *FileData) IsDir() bool {
	return f.isDir
}

// viewportFormatter formats files with size information
func viewportFormatter(node *treeview.Node[*FileData]) (string, bool) {
	data := *node.Data()
	if data.size > 0 {
		return fmt.Sprintf("%s (%s)", data.Name(), formatFileSize(data.size)), true
	}
	return data.Name(), true
}

// createViewportProvider creates a DefaultNodeProvider with file size formatting
func createViewportProvider() *treeview.DefaultNodeProvider[*FileData] {
	// Start with default file node provider options
	baseOptions := []treeview.ProviderOption[*FileData]{
		treeview.WithDefaultFolderRules[*FileData](),
		treeview.WithDefaultFileRules[*FileData](),
		treeview.WithFileExtensionRules[*FileData](),
		treeview.WithDefaultIcon[*FileData]("•"),
		treeview.WithFormatter[*FileData](viewportFormatter),
	}

	return treeview.NewDefaultNodeProvider(baseOptions...)
}

// formatFileSize formats a file size in bytes to a human-readable string
func formatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * KB
		GB = MB * KB
	)
	if bytes < KB {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < MB {
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	}
	if bytes < GB {
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	}
	return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
}

// FileNode wraps treeview.Node to provide file/directory functionality
type FileNode struct {
	*treeview.Node[*FileData]
}

// NewFileNode creates a new FileNode
func NewFileNode(id string, data *FileData) *FileNode {
	return &FileNode{
		Node: treeview.NewNodeSimple(id, data),
	}
}

// AddFileChild adds a FileNode as a child by converting to the base type
func (fn *FileNode) AddFileChild(child *FileNode) {
	fn.AddChild(child.Node)
}

// createLargeFileSystem creates a large, realistic project structure for demonstration
func createLargeFileSystem() []*treeview.Node[*FileData] {
	// Create root project directory
	root := NewFileNode("large-project", &FileData{name: "Large Project", isDir: true})

	// Source code directory with multiple subdirectories
	src := NewFileNode("src", &FileData{name: "src", isDir: true})

	// API handlers
	apiHandlers := NewFileNode("handlers", &FileData{name: "handlers", isDir: true})
	apiHandlers.AddFileChild(NewFileNode("auth.go", &FileData{name: "auth.go", isDir: false}))
	apiHandlers.AddFileChild(NewFileNode("users.go", &FileData{name: "users.go", isDir: false}))
	apiHandlers.AddFileChild(NewFileNode("orders.go", &FileData{name: "orders.go", isDir: false}))
	apiHandlers.AddFileChild(NewFileNode("payments.go", &FileData{name: "payments.go", isDir: false}))
	apiHandlers.AddFileChild(NewFileNode("products.go", &FileData{name: "products.go", isDir: false}))
	apiHandlers.AddFileChild(NewFileNode("middleware.go", &FileData{name: "middleware.go", isDir: false}))
	src.AddFileChild(apiHandlers)

	// Models
	models := NewFileNode("models", &FileData{name: "models", isDir: true})
	models.AddFileChild(NewFileNode("user.go", &FileData{name: "user.go", isDir: false}))
	models.AddFileChild(NewFileNode("order.go", &FileData{name: "order.go", isDir: false}))
	models.AddFileChild(NewFileNode("product.go", &FileData{name: "product.go", isDir: false}))
	models.AddFileChild(NewFileNode("payment.go", &FileData{name: "payment.go", isDir: false}))
	models.AddFileChild(NewFileNode("base.go", &FileData{name: "base.go", isDir: false}))
	src.AddFileChild(models)

	// Services
	services := NewFileNode("services", &FileData{name: "services", isDir: true})
	services.AddFileChild(NewFileNode("auth_service.go", &FileData{name: "auth_service.go", isDir: false}))
	services.AddFileChild(NewFileNode("user_service.go", &FileData{name: "user_service.go", isDir: false}))
	services.AddFileChild(NewFileNode("email_service.go", &FileData{name: "email_service.go", isDir: false}))
	services.AddFileChild(NewFileNode("payment_service.go", &FileData{name: "payment_service.go", isDir: false}))
	services.AddFileChild(NewFileNode("notification_service.go", &FileData{name: "notification_service.go", isDir: false}))
	src.AddFileChild(services)

	// Utils
	utils := NewFileNode("utils", &FileData{name: "utils", isDir: true})
	utils.AddFileChild(NewFileNode("crypto.go", &FileData{name: "crypto.go", isDir: false}))
	utils.AddFileChild(NewFileNode("validation.go", &FileData{name: "validation.go", isDir: false}))
	utils.AddFileChild(NewFileNode("logging.go", &FileData{name: "logging.go", isDir: false}))
	utils.AddFileChild(NewFileNode("config.go", &FileData{name: "config.go", isDir: false}))
	utils.AddFileChild(NewFileNode("database.go", &FileData{name: "database.go", isDir: false}))
	src.AddFileChild(utils)

	// Database migrations
	migrations := NewFileNode("migrations", &FileData{name: "migrations", isDir: true})
	migrations.AddFileChild(NewFileNode("001_create_users.sql", &FileData{name: "001_create_users.sql", isDir: false}))
	migrations.AddFileChild(NewFileNode("002_create_orders.sql", &FileData{name: "002_create_orders.sql", isDir: false}))
	migrations.AddFileChild(NewFileNode("003_create_products.sql", &FileData{name: "003_create_products.sql", isDir: false}))
	migrations.AddFileChild(NewFileNode("004_add_indexes.sql", &FileData{name: "004_add_indexes.sql", isDir: false}))
	migrations.AddFileChild(NewFileNode("005_add_payments.sql", &FileData{name: "005_add_payments.sql", isDir: false}))
	src.AddFileChild(migrations)

	src.AddFileChild(NewFileNode("main.go", &FileData{name: "main.go", isDir: false}))
	src.AddFileChild(NewFileNode("server.go", &FileData{name: "server.go", isDir: false}))
	src.AddFileChild(NewFileNode("routes.go", &FileData{name: "routes.go", isDir: false}))

	// Frontend web directory
	web := NewFileNode("web", &FileData{name: "web", isDir: true})

	// Static assets
	assets := NewFileNode("assets", &FileData{name: "assets", isDir: true})

	// CSS
	css := NewFileNode("css", &FileData{name: "css", isDir: true})
	css.AddFileChild(NewFileNode("main.css", &FileData{name: "main.css", isDir: false}))
	css.AddFileChild(NewFileNode("components.css", &FileData{name: "components.css", isDir: false}))
	css.AddFileChild(NewFileNode("themes.css", &FileData{name: "themes.css", isDir: false}))
	css.AddFileChild(NewFileNode("responsive.css", &FileData{name: "responsive.css", isDir: false}))
	assets.AddFileChild(css)

	// JavaScript
	js := NewFileNode("js", &FileData{name: "js", isDir: true})
	js.AddFileChild(NewFileNode("app.js", &FileData{name: "app.js", isDir: false}))
	js.AddFileChild(NewFileNode("components.js", &FileData{name: "components.js", isDir: false}))
	js.AddFileChild(NewFileNode("utils.js", &FileData{name: "utils.js", isDir: false}))
	js.AddFileChild(NewFileNode("api.js", &FileData{name: "api.js", isDir: false}))
	js.AddFileChild(NewFileNode("validation.js", &FileData{name: "validation.js", isDir: false}))
	assets.AddFileChild(js)

	// Images
	images := NewFileNode("images", &FileData{name: "images", isDir: true})
	images.AddFileChild(NewFileNode("logo.png", &FileData{name: "logo.png", isDir: false}))
	images.AddFileChild(NewFileNode("hero.jpg", &FileData{name: "hero.jpg", isDir: false}))
	images.AddFileChild(NewFileNode("profile-default.svg", &FileData{name: "profile-default.svg", isDir: false}))
	images.AddFileChild(NewFileNode("icons.sprite.svg", &FileData{name: "icons.sprite.svg", isDir: false}))
	assets.AddFileChild(images)

	web.AddFileChild(assets)
	web.AddFileChild(NewFileNode("index.html", &FileData{name: "index.html", isDir: false}))
	web.AddFileChild(NewFileNode("login.html", &FileData{name: "login.html", isDir: false}))
	web.AddFileChild(NewFileNode("dashboard.html", &FileData{name: "dashboard.html", isDir: false}))

	// Documentation
	docs := NewFileNode("docs", &FileData{name: "docs", isDir: true})
	docs.AddFileChild(NewFileNode("README.md", &FileData{name: "README.md", isDir: false}))
	docs.AddFileChild(NewFileNode("API.md", &FileData{name: "API.md", isDir: false}))
	docs.AddFileChild(NewFileNode("CONTRIBUTING.md", &FileData{name: "CONTRIBUTING.md", isDir: false}))
	docs.AddFileChild(NewFileNode("DEPLOYMENT.md", &FileData{name: "DEPLOYMENT.md", isDir: false}))
	docs.AddFileChild(NewFileNode("ARCHITECTURE.md", &FileData{name: "ARCHITECTURE.md", isDir: false}))

	// API documentation
	apiDocs := NewFileNode("api", &FileData{name: "api", isDir: true})
	apiDocs.AddFileChild(NewFileNode("openapi.yaml", &FileData{name: "openapi.yaml", isDir: false}))
	apiDocs.AddFileChild(NewFileNode("auth.md", &FileData{name: "auth.md", isDir: false}))
	apiDocs.AddFileChild(NewFileNode("users.md", &FileData{name: "users.md", isDir: false}))
	apiDocs.AddFileChild(NewFileNode("orders.md", &FileData{name: "orders.md", isDir: false}))
	docs.AddFileChild(apiDocs)

	// Configuration files
	config := NewFileNode("config", &FileData{name: "config", isDir: true})
	config.AddFileChild(NewFileNode("app.yaml", &FileData{name: "app.yaml", isDir: false}))
	config.AddFileChild(NewFileNode("database.yaml", &FileData{name: "database.yaml", isDir: false}))
	config.AddFileChild(NewFileNode("redis.yaml", &FileData{name: "redis.yaml", isDir: false}))
	config.AddFileChild(NewFileNode("nginx.conf", &FileData{name: "nginx.conf", isDir: false}))
	config.AddFileChild(NewFileNode("docker-compose.yml", &FileData{name: "docker-compose.yml", isDir: false}))

	// Environment configs
	envs := NewFileNode("environments", &FileData{name: "environments", isDir: true})
	envs.AddFileChild(NewFileNode("development.env", &FileData{name: "development.env", isDir: false}))
	envs.AddFileChild(NewFileNode("staging.env", &FileData{name: "staging.env", isDir: false}))
	envs.AddFileChild(NewFileNode("production.env", &FileData{name: "production.env", isDir: false}))
	config.AddFileChild(envs)

	// Tests directory
	tests := NewFileNode("tests", &FileData{name: "tests", isDir: true})

	// Unit tests
	unit := NewFileNode("unit", &FileData{name: "unit", isDir: true})
	unit.AddFileChild(NewFileNode("auth_test.go", &FileData{name: "auth_test.go", isDir: false}))
	unit.AddFileChild(NewFileNode("users_test.go", &FileData{name: "users_test.go", isDir: false}))
	unit.AddFileChild(NewFileNode("orders_test.go", &FileData{name: "orders_test.go", isDir: false}))
	unit.AddFileChild(NewFileNode("payments_test.go", &FileData{name: "payments_test.go", isDir: false}))
	unit.AddFileChild(NewFileNode("models_test.go", &FileData{name: "models_test.go", isDir: false}))
	tests.AddFileChild(unit)

	// Integration tests
	integration := NewFileNode("integration", &FileData{name: "integration", isDir: true})
	integration.AddFileChild(NewFileNode("api_test.go", &FileData{name: "api_test.go", isDir: false}))
	integration.AddFileChild(NewFileNode("database_test.go", &FileData{name: "database_test.go", isDir: false}))
	integration.AddFileChild(NewFileNode("auth_flow_test.go", &FileData{name: "auth_flow_test.go", isDir: false}))
	tests.AddFileChild(integration)

	// E2E tests
	e2e := NewFileNode("e2e", &FileData{name: "e2e", isDir: true})
	e2e.AddFileChild(NewFileNode("user_journey_test.go", &FileData{name: "user_journey_test.go", isDir: false}))
	e2e.AddFileChild(NewFileNode("order_flow_test.go", &FileData{name: "order_flow_test.go", isDir: false}))
	e2e.AddFileChild(NewFileNode("payment_flow_test.go", &FileData{name: "payment_flow_test.go", isDir: false}))
	tests.AddFileChild(e2e)

	// Scripts directory
	scripts := NewFileNode("scripts", &FileData{name: "scripts", isDir: true})
	scripts.AddFileChild(NewFileNode("build.sh", &FileData{name: "build.sh", isDir: false}))
	scripts.AddFileChild(NewFileNode("deploy.sh", &FileData{name: "deploy.sh", isDir: false}))
	scripts.AddFileChild(NewFileNode("test.sh", &FileData{name: "test.sh", isDir: false}))
	scripts.AddFileChild(NewFileNode("migrate.sh", &FileData{name: "migrate.sh", isDir: false}))
	scripts.AddFileChild(NewFileNode("seed.py", &FileData{name: "seed.py", isDir: false}))
	scripts.AddFileChild(NewFileNode("backup.py", &FileData{name: "backup.py", isDir: false}))

	// Dependencies
	nodeModules := NewFileNode("node_modules", &FileData{name: "node_modules", isDir: true})

	// Some popular packages
	packages := []string{
		"react", "vue", "angular", "express", "lodash", "axios", "moment",
		"webpack", "babel", "eslint", "prettier", "jest", "cypress", "typescript",
	}
	for _, pkg := range packages {
		pkgNode := NewFileNode(pkg, &FileData{name: pkg, isDir: true})
		pkgNode.AddFileChild(NewFileNode(pkg+"/package.json", &FileData{name: "package.json", isDir: false}))
		pkgNode.AddFileChild(NewFileNode(pkg+"/index.js", &FileData{name: "index.js", isDir: false}))
		pkgNode.AddFileChild(NewFileNode(pkg+"/README.md", &FileData{name: "README.md", isDir: false}))
		nodeModules.AddFileChild(pkgNode)
	}

	// Add all major directories to root
	root.AddFileChild(src)
	root.AddFileChild(web)
	root.AddFileChild(docs)
	root.AddFileChild(config)
	root.AddFileChild(tests)
	root.AddFileChild(scripts)
	root.AddFileChild(nodeModules)

	// Root level files
	root.AddFileChild(NewFileNode("go.mod", &FileData{name: "go.mod", isDir: false}))
	root.AddFileChild(NewFileNode("go.sum", &FileData{name: "go.sum", isDir: false}))
	root.AddFileChild(NewFileNode("package.json", &FileData{name: "package.json", isDir: false}))
	root.AddFileChild(NewFileNode("package-lock.json", &FileData{name: "package-lock.json", isDir: false}))
	root.AddFileChild(NewFileNode(".gitignore", &FileData{name: ".gitignore", isDir: false}))
	root.AddFileChild(NewFileNode(".env.example", &FileData{name: ".env.example", isDir: false}))
	root.AddFileChild(NewFileNode("Dockerfile", &FileData{name: "Dockerfile", isDir: false}))
	root.AddFileChild(NewFileNode("Makefile", &FileData{name: "Makefile", isDir: false}))
	root.AddFileChild(NewFileNode("LICENSE", &FileData{name: "LICENSE", isDir: false}))
	root.AddFileChild(NewFileNode("CHANGELOG.md", &FileData{name: "CHANGELOG.md", isDir: false}))

	// Auto-expand a few directories to show content immediately
	src.Expand()
	apiHandlers.Expand()
	docs.Expand()
	tests.Expand()
	unit.Expand()
	nodeModules.Expand()

	return []*treeview.Node[*FileData]{root.Node}
}
