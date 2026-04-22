package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Digital-Shane/treeview"
	"github.com/Digital-Shane/treeview/examples/shared"
	"charm.land/lipgloss/v2"
)

////////////////////////////////////////////////////////////////////
//   Workflow Status Types
////////////////////////////////////////////////////////////////////

type WorkflowStatus string

const (
	StatusPending   WorkflowStatus = "pending"
	StatusRunning   WorkflowStatus = "running"
	StatusCompleted WorkflowStatus = "completed"
	StatusFailed    WorkflowStatus = "failed"
	StatusSkipped   WorkflowStatus = "skipped"
)

type WorkflowData struct {
	Status    WorkflowStatus
	Duration  time.Duration
	StartTime time.Time
	ErrorMsg  string
}

// Domain-specific predicate helper for WorkflowData
func hasStatus(status WorkflowStatus) func(*treeview.Node[WorkflowData]) bool {
	return func(n *treeview.Node[WorkflowData]) bool {
		return n.Data().Status == status
	}
}

////////////////////////////////////////////////////////////////////
//   Workflow Provider Options
////////////////////////////////////////////////////////////////////

// workflowFormatter formats workflow nodes based on their status and timing
func workflowFormatter(node *treeview.Node[WorkflowData]) (string, bool) {
	data := node.Data()
	display := node.Name()

	if data.Status == StatusRunning {
		if data.Duration > 0 {
			// Show cumulative duration for running workflows
			display += fmt.Sprintf(" (running %v)", data.Duration.Round(time.Second))
		} else if !data.StartTime.IsZero() {
			// Show elapsed time for newly started steps
			elapsed := time.Since(data.StartTime)
			display += fmt.Sprintf(" (running %v)", elapsed.Round(time.Second))
		}
	} else if data.Status == StatusCompleted && data.Duration > 0 {
		display += fmt.Sprintf(" (%v)", data.Duration.Round(time.Millisecond))
	} else if data.Status == StatusFailed && data.ErrorMsg != "" {
		display += fmt.Sprintf(" (error: %s)", data.ErrorMsg)
	}

	return display, true
}

// createWorkflowProvider creates a DefaultNodeProvider configured for workflow visualization
func createWorkflowProvider() *treeview.DefaultNodeProvider[WorkflowData] {
	// Icon rules based on workflow status
	pendingIconRule := treeview.WithIconRule(hasStatus(StatusPending), "⏳")
	runningIconRule := treeview.WithIconRule(hasStatus(StatusRunning), "🔄")
	completedIconRule := treeview.WithIconRule(hasStatus(StatusCompleted), "✅")
	failedIconRule := treeview.WithIconRule(hasStatus(StatusFailed), "❌")
	skippedIconRule := treeview.WithIconRule(hasStatus(StatusSkipped), "⏭️")
	defaultIconRule := treeview.WithDefaultIcon[WorkflowData]("❓")

	// Style rules based on workflow status
	pendingStyleRule := treeview.WithStyleRule(
		hasStatus(StatusPending),
		lipgloss.NewStyle().Foreground(lipgloss.Color("243")), // Gray
		lipgloss.NewStyle().Background(lipgloss.Color("57")).Underline(true).Foreground(lipgloss.Color("243")),
	)
	runningStyleRule := treeview.WithStyleRule(
		hasStatus(StatusRunning),
		lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true), // Blue, bold
		lipgloss.NewStyle().Background(lipgloss.Color("57")).Underline(true).Foreground(lipgloss.Color("33")).Bold(true),
	)
	completedStyleRule := treeview.WithStyleRule(
		hasStatus(StatusCompleted),
		lipgloss.NewStyle().Foreground(lipgloss.Color("76")), // Green
		lipgloss.NewStyle().Background(lipgloss.Color("57")).Underline(true).Foreground(lipgloss.Color("76")),
	)
	failedStyleRule := treeview.WithStyleRule(
		hasStatus(StatusFailed),
		lipgloss.NewStyle().Foreground(lipgloss.Color("196")), // Red
		lipgloss.NewStyle().Background(lipgloss.Color("57")).Underline(true).Foreground(lipgloss.Color("196")),
	)
	skippedStyleRule := treeview.WithStyleRule(
		hasStatus(StatusSkipped),
		lipgloss.NewStyle().Foreground(lipgloss.Color("214")), // Orange
		lipgloss.NewStyle().Background(lipgloss.Color("57")).Underline(true).Foreground(lipgloss.Color("214")),
	)

	// Formatter rule
	formatterRule := treeview.WithFormatter[WorkflowData](workflowFormatter)

	return treeview.NewDefaultNodeProvider(
		// Order matters: most specific rules first
		pendingIconRule,
		runningIconRule,
		completedIconRule,
		failedIconRule,
		skippedIconRule,
		defaultIconRule,
		pendingStyleRule,
		runningStyleRule,
		completedStyleRule,
		failedStyleRule,
		skippedStyleRule,
		formatterRule,
	)
}

////////////////////////////////////////////////////////////////////
//   Workflow Visualization Example
////////////////////////////////////////////////////////////////////

func main() {
	shared.ClearTerminal()

	// Create the workflow tree nodes
	workflowRoot := treeview.NewNode("ci-cd-pipeline", "CI/CD Pipeline", WorkflowData{Status: StatusPending})

	// Prepare Environment
	prepare := treeview.NewNode("prepare", "Prepare Environment", WorkflowData{Status: StatusPending})
	workflowRoot.AddChild(prepare)

	// Build Application
	build := treeview.NewNode("build", "Build Application", WorkflowData{Status: StatusPending})
	compile := treeview.NewNode("compile", "Compile Source Code", WorkflowData{Status: StatusPending})
	packageNode := treeview.NewNode("package", "Package Artifacts", WorkflowData{Status: StatusPending})
	build.AddChild(compile)
	build.AddChild(packageNode)
	workflowRoot.AddChild(build)

	// Test Suite
	test := treeview.NewNode("test", "Test Suite", WorkflowData{Status: StatusPending})
	unitTests := treeview.NewNode("unit-tests", "Unit Tests", WorkflowData{Status: StatusPending})
	integrationTests := treeview.NewNode("integration-tests", "Integration Tests", WorkflowData{Status: StatusPending})
	test.AddChild(unitTests)
	test.AddChild(integrationTests)
	workflowRoot.AddChild(test)

	// Deploy to Production
	deploy := treeview.NewNode("deploy", "Deploy to Production", WorkflowData{Status: StatusPending})
	workflowRoot.AddChild(deploy)

	// Create tree renderer with workflow provider
	workflowTree := treeview.NewTree(
		[]*treeview.Node[WorkflowData]{workflowRoot},
		treeview.WithProvider(createWorkflowProvider()),
		treeview.WithExpandAll[WorkflowData](),
	)

	////////////////////////////////////////////////////////////////////
	//   Initial Workflow State
	////////////////////////////////////////////////////////////////////

	fmt.Println("Workflow visualization:")
	output, _ := workflowTree.Render(context.Background())
	fmt.Println(output)

	shared.WaitDelayThenClearTerminal(3.5)

	////////////////////////////////////////////////////////////////////
	//   Build in Progress
	////////////////////////////////////////////////////////////////////

	fmt.Println("Workflow visualization:")
	// Prep complete, build starting
	prepare.SetData(WorkflowData{
		Status:   StatusCompleted,
		Duration: 30 * time.Second,
	})
	build.SetData(WorkflowData{Status: StatusRunning})
	compile.SetData(WorkflowData{Status: StatusRunning})
	workflowRoot.SetData(WorkflowData{
		Status:   StatusPending,
		Duration: 30 * time.Second,
	})

	output, _ = workflowTree.Render(context.Background())
	fmt.Println(output)

	shared.WaitDelayThenClearTerminal(3.5)

	////////////////////////////////////////////////////////////////////
	//   Build Complete, Tests Running
	////////////////////////////////////////////////////////////////////

	fmt.Println("Workflow visualization:")
	// Build complete, tests starting
	build.SetData(WorkflowData{
		Status:   StatusCompleted,
		Duration: 75 * time.Second,
	})
	compile.SetData(WorkflowData{
		Status:   StatusCompleted,
		Duration: 45 * time.Second,
	})
	packageNode.SetData(WorkflowData{
		Status:   StatusCompleted,
		Duration: 30 * time.Second,
	})
	test.SetData(WorkflowData{Status: StatusRunning})
	unitTests.SetData(WorkflowData{Status: StatusRunning})
	workflowRoot.SetData(WorkflowData{
		Status:   StatusPending,
		Duration: 105 * time.Second,
	})

	output, _ = workflowTree.Render(context.Background())
	fmt.Println(output)

	shared.WaitDelayThenClearTerminal(3.5)

	////////////////////////////////////////////////////////////////////
	//   Unit Tests Complete
	////////////////////////////////////////////////////////////////////

	fmt.Println("Workflow visualization:")
	// Unit tests complete, integration tests running
	unitTests.SetData(WorkflowData{
		Status:   StatusCompleted,
		Duration: 22 * time.Second,
	})
	integrationTests.SetData(WorkflowData{Status: StatusRunning})
	workflowRoot.SetData(WorkflowData{
		Status:   StatusPending,
		Duration: 127 * time.Second,
	})

	output, _ = workflowTree.Render(context.Background())
	fmt.Println(output)

	shared.WaitDelayThenClearTerminal(3.5)

	////////////////////////////////////////////////////////////////////
	//   Tests Complete, Deploying
	////////////////////////////////////////////////////////////////////

	fmt.Println("Workflow visualization:")
	// Tests complete, deployment running
	test.SetData(WorkflowData{
		Status:   StatusCompleted,
		Duration: 58 * time.Second,
	})
	integrationTests.SetData(WorkflowData{
		Status:   StatusCompleted,
		Duration: 36 * time.Second,
	})
	deploy.SetData(WorkflowData{Status: StatusRunning})
	workflowRoot.SetData(WorkflowData{
		Status:   StatusPending,
		Duration: 163 * time.Second,
	})

	output, _ = workflowTree.Render(context.Background())
	fmt.Println(output)

	shared.WaitDelayThenClearTerminal(3.5)

	////////////////////////////////////////////////////////////////////
	//   Workflow Complete
	////////////////////////////////////////////////////////////////////

	fmt.Println("Workflow visualization:")
	// All complete
	deploy.SetData(WorkflowData{
		Status:   StatusCompleted,
		Duration: 85 * time.Second,
	})
	workflowRoot.SetData(WorkflowData{
		Status:   StatusCompleted,
		Duration: 248 * time.Second,
	})

	output, _ = workflowTree.Render(context.Background())
	fmt.Println(output)

	shared.WaitDelay(3.5)
}
