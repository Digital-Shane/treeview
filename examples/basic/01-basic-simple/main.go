package main

import (
	"context"
	"fmt"

	"github.com/Digital-Shane/treeview/v2"
	"github.com/Digital-Shane/treeview/v2/examples/shared"
)

////////////////////////////////////////////////////////////////////
//   Simple Tree Example
////////////////////////////////////////////////////////////////////

func main() {
	shared.ClearTerminal()
	fmt.Println("Basic Tree With Default Formatting")
	// Create nodes representing a basic tree structure
	root := shared.CreateBasicTreeNodes()
	// Create tree with default options
	tree := treeview.NewTree([]*treeview.Node[string]{root}, treeview.WithExpandAll[string]())

	// Render the tree to a string & print it
	output, _ := tree.Render(context.Background())
	fmt.Println(output)

	shared.WaitDelayThenClearTerminal(3.5)

	// Move focus to the second child and re-render
	fmt.Println("Focus moved to second child")
	tree.SetFocusedID(context.Background(), root.Children()[0].Children()[0].ID())
	output, _ = tree.Render(context.Background())
	fmt.Println(output)

	shared.WaitDelayThenClearTerminal(3.5)

	// Delete focus and re-render
	fmt.Println("Focus deleted")
	tree.SetFocusedID(context.Background(), "")
	output, _ = tree.Render(context.Background())
	fmt.Println(output)

	shared.WaitDelay(3.5)
}
