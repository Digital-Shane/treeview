package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Digital-Shane/treeview/v2"
	"github.com/Digital-Shane/treeview/v2/examples/shared"
)

////////////////////////////////////////////////////////////////////
//   Tree Iterators Example
////////////////////////////////////////////////////////////////////

func main() {
	shared.ClearTerminal()
	fmt.Println("Tree Structure:")

	// Create a sample tree
	root := shared.CreateBasicTreeNodes()
	tree := treeview.NewTree([]*treeview.Node[string]{root}, treeview.WithExpandAll[string]())

	output, _ := tree.Render(context.Background())
	fmt.Println(output)

	shared.WaitDelayThenClearTerminal(2)

	////////////////////////////////////////////////////////////////////
	//   Depth-First Iteration (Default)
	////////////////////////////////////////////////////////////////////

	fmt.Println("Depth-First Iteration (All):")

	order := 1
	for info, err := range tree.All(context.Background()) {
		if err != nil {
			fmt.Printf("Error during iteration: %v\n", err)
			break
		}
		indent := strings.Repeat("  ", info.Depth)
		fmt.Printf("%d. %s%s\n", order, indent, info.Node.Name())
		order++
	}

	shared.WaitDelayThenClearTerminal(3.5)

	////////////////////////////////////////////////////////////////////
	//   Breadth-First Iteration
	////////////////////////////////////////////////////////////////////

	fmt.Println("Breadth-First Iteration:")

	order = 1
	for info, err := range tree.BreadthFirst(context.Background()) {
		if err != nil {
			fmt.Printf("Error during iteration: %v\n", err)
			break
		}
		indent := strings.Repeat("  ", info.Depth)
		fmt.Printf("%d. %s%s\n", order, indent, info.Node.Name())
		order++
	}

	shared.WaitDelayThenClearTerminal(3.5)

	////////////////////////////////////////////////////////////////////
	//   Visible Nodes Only
	////////////////////////////////////////////////////////////////////

	fmt.Println("Visible Nodes Only\n(config/ and database.json hidden):")

	// Hide some nodes
	for info, err := range tree.All(context.Background()) {
		if err != nil {
			fmt.Printf("Error during iteration: %v\n", err)
			break
		}
		if info.Node.ID() == "config" || info.Node.ID() == "database.json" {
			info.Node.SetVisible(false)
		}
	}

	order = 1
	for info, err := range tree.AllVisible(context.Background()) {
		if err != nil {
			fmt.Printf("Error during iteration: %v\n", err)
			break
		}
		indent := strings.Repeat("  ", info.Depth)
		fmt.Printf("%d. %s%s\n", order, indent, info.Node.Name())
		order++
	}

	shared.WaitDelayThenClearTerminal(3.5)

	////////////////////////////////////////////////////////////////////
	//   Context with Timeout
	////////////////////////////////////////////////////////////////////

	fmt.Println("Iteration with 50ms timeout\n(simulating 20ms work per node):")

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	count := 0
	for info, err := range tree.All(ctx) {
		if err != nil {
			fmt.Printf("Iteration stopped due to: %v\n", err)
			break
		}
		count++
		indent := strings.Repeat("  ", info.Depth)
		fmt.Printf("%d. %s%s\n", count, indent, info.Node.Name())
		// Simulate work that takes time
		time.Sleep(20 * time.Millisecond)
	}

	fmt.Printf("\nProcessed %d nodes before timeout\n", count)

	shared.WaitDelayThenClearTerminal(3.5)

	////////////////////////////////////////////////////////////////////
	//   Iteration from Specific Node
	////////////////////////////////////////////////////////////////////

	fmt.Println("Iteration starting from 'src/' node:")

	// Find the src node
	var srcNode *treeview.Node[string]
	for _, child := range root.Children() {
		if child.ID() == "src" {
			srcNode = child
			break
		}
	}

	order = 1
	for info, err := range srcNode.All(context.Background()) {
		if err != nil {
			fmt.Printf("Error during iteration: %v\n", err)
			break
		}
		indent := strings.Repeat("  ", info.Depth)
		fmt.Printf("%d. %s%s\n", order, indent, info.Node.Name())
		order++
	}

	shared.WaitDelay(3.5)
}
