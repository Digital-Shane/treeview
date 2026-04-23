// Package shared provides utilities and common functionality for TreeView v2 examples.
package shared

import (
	"flag"
	"fmt"
	"time"

	"github.com/Digital-Shane/treeview/v2"
)

var noDelay = flag.Bool("no-delay", false, "Disable delay between examples")

// ExampleStep represents a single step in an example.
type ExampleStep struct {
	Name string
	Func func() (string, error)
}

// RunExampleStepsWithDelay runs a series of example steps with a title and custom delay.
// The delay parameter specifies seconds between steps (0 for no delay).
func RunExampleStepsWithDelay(title string, steps []ExampleStep, delaySeconds int) {
	for i, step := range steps {
		// Clear terminal before each step to maintain consistent position
		if delaySeconds > 0 {
			ClearTerminal()
		}

		fmt.Printf("\n📊 Example: %s\n\n", title)
		fmt.Printf("➡️  Step %d: %s\n", i+1, step.Name)

		output, err := step.Func()
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}
		// Print the output from the step
		if output != "" {
			fmt.Println(output)
		}

		// Apply delay if not the last step
		if delaySeconds > 0 && i < len(steps)-1 {
			time.Sleep(time.Duration(delaySeconds) * time.Second)
		}
	}
}

// ClearTerminal clears the terminal screen
func ClearTerminal() {
	// ANSI escape code to clear screen and move cursor to top
	fmt.Print("\033[H\033[2J")
}

// WaitDelay waits for the specified number of seconds
func WaitDelay(defaultDelay float64) {
	// Setup command to determine if delay is disabled
	if !flag.Parsed() {
		flag.Parse()
	}

	if *noDelay {
		return
	}
	time.Sleep(time.Duration(defaultDelay) * time.Second)
}

// ClearTerminal clears the terminal screen
func WaitDelayThenClearTerminal(seconds float64) {
	WaitDelay(seconds)
	if seconds > 0 {
		ClearTerminal()
	}
}

// CreateBasicTreeNodes creates a simple tree structure for basic examples.
func CreateBasicTreeNodes() *treeview.Node[string] {
	root := treeview.NewNodeSimple("root", "Project")

	// Add source files
	src := treeview.NewNodeSimple("src", "src/")
	src.AddChild(treeview.NewNodeSimple("main.go", "main.go"))
	src.AddChild(treeview.NewNodeSimple("utils.go", "utils.go"))

	// Add config files
	config := treeview.NewNodeSimple("config", "config/")
	config.AddChild(treeview.NewNodeSimple("app.yaml", "app.yaml"))
	config.AddChild(treeview.NewNodeSimple("database.json", "database.json"))

	// Add docs
	docs := treeview.NewNodeSimple("docs", "docs/")
	docs.AddChild(treeview.NewNodeSimple("readme.md", "README.md"))
	docs.AddChild(treeview.NewNodeSimple("api.md", "API.md"))

	root.AddChild(src)
	root.AddChild(config)
	root.AddChild(docs)

	return root
}
