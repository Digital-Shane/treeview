package main

import (
	"context"
	"fmt"

	"github.com/Digital-Shane/treeview"
	"github.com/Digital-Shane/treeview/examples/shared"
	"charm.land/lipgloss/v2"
)

////////////////////////////////////////////////////////////////////
//   Comparison Data Types
////////////////////////////////////////////////////////////////////

// ChangeType represents different types of file changes
type ChangeType string

const (
	ChangeNone    ChangeType = "none"
	ChangeRenamed ChangeType = "renamed"
	ChangeDeleted ChangeType = "deleted"
	ChangeAdded   ChangeType = "added"
	ChangeMoved   ChangeType = "moved"
)

// ComparisonData holds the before/after information for a file change
type ComparisonData struct {
	Type    ChangeType
	OldName string
	NewName string
	Symbol  string
}

// FileChange wraps comparison data as node data
type FileChange struct {
	Name   string
	Change ComparisonData
}

////////////////////////////////////////////////////////////////////
//   Predicate Helpers for File Changes
////////////////////////////////////////////////////////////////////

// hasChangeType returns a predicate that checks if a node has the given change type
func hasChangeType(changeType ChangeType) func(*treeview.Node[FileChange]) bool {
	return func(n *treeview.Node[FileChange]) bool {
		return n.Data().Change.Type == changeType
	}
}

////////////////////////////////////////////////////////////////////
//   Provider Options for File Changes
////////////////////////////////////////////////////////////////////

// comparisonFormatter formats nodes based on their change type
func comparisonFormatter(node *treeview.Node[FileChange]) (string, bool) {
	data := node.Data()
	change := data.Change

	switch change.Type {
	case ChangeRenamed:
		return fmt.Sprintf("%s %s %s", change.OldName, change.Symbol, change.NewName), true
	case ChangeDeleted:
		return fmt.Sprintf("%s %s", change.OldName, change.Symbol), true
	case ChangeAdded:
		return fmt.Sprintf("%s %s", change.NewName, change.Symbol), true
	case ChangeMoved:
		return fmt.Sprintf("%s %s %s", change.OldName, change.Symbol, change.NewName), true
	default:
		return data.Name, true
	}
}

// createComparisonProvider creates a DefaultNodeProvider configured for file changes
func createComparisonProvider() *treeview.DefaultNodeProvider[FileChange] {
	return treeview.NewDefaultNodeProvider(
		// Icon rules based on change type (most specific first)
		treeview.WithIconRule(hasChangeType(ChangeRenamed), "📝"),
		treeview.WithIconRule(hasChangeType(ChangeDeleted), "🗑"),
		treeview.WithIconRule(hasChangeType(ChangeAdded), "✨"),
		treeview.WithIconRule(hasChangeType(ChangeMoved), "📁"),
		treeview.WithDefaultIcon[FileChange]("📄"),

		// Style rules based on change type
		treeview.WithStyleRule(
			hasChangeType(ChangeRenamed),
			lipgloss.NewStyle().Foreground(lipgloss.Color("214")), // Orange for renames
			lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("214")).Bold(true),
		),
		treeview.WithStyleRule(
			hasChangeType(ChangeDeleted),
			lipgloss.NewStyle().Foreground(lipgloss.Color("160")).Strikethrough(true), // Red for deletions with strikethrough
			lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("160")).Bold(true).Strikethrough(true),
		),
		treeview.WithStyleRule(
			hasChangeType(ChangeAdded),
			lipgloss.NewStyle().Foreground(lipgloss.Color("34")), // Green for additions
			lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("34")).Bold(true),
		),
		treeview.WithStyleRule(
			hasChangeType(ChangeMoved),
			lipgloss.NewStyle().Foreground(lipgloss.Color("39")), // Blue for moves
			lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("39")).Bold(true),
		),

		// Custom formatter for change operations
		treeview.WithFormatter[FileChange](comparisonFormatter),
	)
}

////////////////////////////////////////////////////////////////////
//   Creates a tree showing common file operation changes.
////////////////////////////////////////////////////////////////////

func createFileChangesTree() *treeview.Node[FileChange] {
	root := treeview.NewNodeSimple("changes", FileChange{Name: "File Changes", Change: ComparisonData{Type: ChangeNone}})

	// Renamed file: cache.db -> cache.db.bak
	renamed := treeview.NewNodeSimple("cache", FileChange{
		Name: "",
		Change: ComparisonData{
			Type:    ChangeRenamed,
			OldName: "cache.db",
			NewName: "cache.db.bak",
			Symbol:  "→",
		},
	})

	// Deleted file
	deleted := treeview.NewNodeSimple("temp", FileChange{
		Name: "",
		Change: ComparisonData{
			Type:    ChangeDeleted,
			OldName: "temp.log",
			Symbol:  "✗",
		},
	})

	// Added file
	added := treeview.NewNodeSimple("new", FileChange{
		Name: "",
		Change: ComparisonData{
			Type:    ChangeAdded,
			NewName: "config.yaml",
			Symbol:  "✨",
		},
	})

	// Moved file: src/utils.go -> lib/utils.go
	moved := treeview.NewNodeSimple("utils", FileChange{
		Name: "",
		Change: ComparisonData{
			Type:    ChangeMoved,
			OldName: "src/utils.go",
			NewName: "lib/utils.go",
			Symbol:  "↗",
		},
	})

	// Add all changes to root
	root.AddChild(renamed)
	root.AddChild(deleted)
	root.AddChild(added)
	root.AddChild(moved)

	// Expand to show changes
	root.Expand()

	return root
}

////////////////////////////////////////////////////////////////////
//   Main Function
//
//   Simple demonstration of file change comparison display.
////////////////////////////////////////////////////////////////////

func main() {
	root := createFileChangesTree()

	// Create tree with custom provider using functional options
	tree := treeview.NewTree(
		[]*treeview.Node[FileChange]{root},
		treeview.WithProvider(createComparisonProvider()),
		treeview.WithExpandAll[FileChange](),
	)

	// Render the tree
	tree.SetFocusedID(context.Background(), "")
	output, err := tree.Render(context.Background())
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(output)
	fmt.Println("\nLegend:")
	fmt.Println("→  Orange = Renamed")
	fmt.Println("✗  Red    = Deleted")
	fmt.Println("✨ Green  = Added")
	fmt.Println("↗  Blue   = Moved")

	shared.WaitDelay(3.5)
}
