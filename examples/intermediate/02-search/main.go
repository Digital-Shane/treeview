package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Digital-Shane/treeview"
	"github.com/charmbracelet/bubbletea"
)

// Product represents a product with additional searchable data
type Product struct {
	Name        string
	Category    string
	Description string
	Tags        []string
}

// String implements the Stringer interface for fallback search
func (p Product) String() string {
	return fmt.Sprintf("%s (%s)", p.Name, p.Category)
}

func main() {
	// Create product catalog
	nodes := createProductCatalog()

	// Create tree with custom search policy that extends the default behavior
	tree := treeview.NewTree(nodes,
		// Custom searcher that searches both default fields AND product data
		treeview.WithSearcher(productSearcher),
	)

	// Set up TUI model
	model := treeview.NewTuiTreeModel(
		tree,
		treeview.WithTuiWidth[Product](80),
		treeview.WithTuiHeight[Product](25),
	)

	// Run the program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

// productSearcher extends the default search to include Product-specific fields
// It searches: node name, product name, category, description, and tags
func productSearcher(ctx context.Context, node *treeview.Node[Product], term string) bool {
	if term == "" {
		return false
	}

	product := node.Data()
	lowerTerm := strings.ToLower(term)

	// First, do the default search behavior node name
	if strings.Contains(strings.ToLower(node.Name()), lowerTerm) {
		return true
	}

	// Then extend with Product-specific fields
	if strings.Contains(strings.ToLower(product.Name), lowerTerm) ||
		strings.Contains(strings.ToLower(product.Category), lowerTerm) ||
		strings.Contains(strings.ToLower(product.Description), lowerTerm) {
		return true
	}

	// Search through tags array
	for _, tag := range product.Tags {
		if strings.Contains(strings.ToLower(tag), lowerTerm) {
			return true
		}
	}

	return false
}

// createProductCatalog creates a sample product hierarchy for demonstration
func createProductCatalog() []*treeview.Node[Product] {
	// Root catalog
	catalog := treeview.NewNode("catalog", "📦 Product Catalog", Product{
		Name:        "Product Catalog",
		Category:    "Root",
		Description: "Main product catalog",
		Tags:        []string{"catalog", "products"},
	})

	// Electronics category
	electronics := treeview.NewNode("electronics", "💻 Electronics", Product{
		Name:        "Electronics",
		Category:    "Category",
		Description: "Electronic devices and accessories",
		Tags:        []string{"electronics", "tech", "gadgets"},
	})

	// Laptops
	laptops := []*treeview.Node[Product]{
		treeview.NewNode("laptop-1", "MacBook Pro", Product{
			Name:        "MacBook Pro 16-inch",
			Category:    "Laptop",
			Description: "Professional laptop with M3 chip and Retina display",
			Tags:        []string{"apple", "professional", "m3", "retina", "creative"},
		}),
		treeview.NewNode("laptop-2", "Dell XPS 13", Product{
			Name:        "Dell XPS 13",
			Category:    "Laptop",
			Description: "Ultrabook with Intel Core i7 and premium build quality",
			Tags:        []string{"dell", "ultrabook", "intel", "portable", "business"},
		}),
		treeview.NewNode("laptop-3", "ThinkPad X1 Carbon", Product{
			Name:        "ThinkPad X1 Carbon",
			Category:    "Laptop",
			Description: "Lightweight business laptop with excellent keyboard",
			Tags:        []string{"lenovo", "thinkpad", "business", "lightweight", "keyboard"},
		}),
	}

	// Smartphones
	smartphones := []*treeview.Node[Product]{
		treeview.NewNode("phone-1", "iPhone 15 Pro", Product{
			Name:        "iPhone 15 Pro",
			Category:    "Smartphone",
			Description: "Latest iPhone with titanium design and advanced camera",
			Tags:        []string{"apple", "iphone", "titanium", "camera", "5g"},
		}),
		treeview.NewNode("phone-2", "Samsung Galaxy S24", Product{
			Name:        "Samsung Galaxy S24",
			Category:    "Smartphone",
			Description: "Android flagship with AI features and excellent display",
			Tags:        []string{"samsung", "android", "ai", "display", "flagship"},
		}),
	}

	// Home category
	home := treeview.NewNode("home", "🏠 Home & Garden", Product{
		Name:        "Home & Garden",
		Category:    "Category",
		Description: "Products for home improvement and gardening",
		Tags:        []string{"home", "garden", "furniture", "decor"},
	})

	// Furniture
	furniture := []*treeview.Node[Product]{
		treeview.NewNode("chair-1", "Ergonomic Office Chair", Product{
			Name:        "Herman Miller Aeron",
			Category:    "Furniture",
			Description: "Premium ergonomic office chair for long work sessions",
			Tags:        []string{"herman-miller", "ergonomic", "office", "premium", "mesh"},
		}),
		treeview.NewNode("desk-1", "Standing Desk", Product{
			Name:        "Jarvis Bamboo Standing Desk",
			Category:    "Furniture",
			Description: "Height-adjustable desk made from sustainable bamboo",
			Tags:        []string{"jarvis", "standing", "bamboo", "adjustable", "sustainable"},
		}),
	}

	// Build the hierarchy
	electronics.SetChildren(append(laptops, smartphones...))
	home.SetChildren(furniture)
	catalog.SetChildren([]*treeview.Node[Product]{electronics, home})

	// Expand for better demonstration
	catalog.Expand()
	electronics.Expand()

	return []*treeview.Node[Product]{catalog}
}
