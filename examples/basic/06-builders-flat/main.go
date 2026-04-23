package main

import (
	"context"
	"fmt"

	"github.com/Digital-Shane/treeview/v2"
	"github.com/Digital-Shane/treeview/v2/examples/shared"
)

////////////////////////////////////////////////////////////////////
//   Employee Organizational Chart
////////////////////////////////////////////////////////////////////

// Employee represents an employee record in a flat data structure
type Employee struct {
	ID         string // Unique employee identifier
	Name       string // Employee's full name
	Title      string // Job title
	Department string // Department name
	ManagerID  string // ID of the employee's manager (empty for CEO)
	Level      int    // Organizational level (0=CEO, 1=C-level, 2=Director, etc.)
}

// Domain-specific predicate helper for Employee
func employeeInDepartment(department string) func(*treeview.Node[Employee]) bool {
	return func(n *treeview.Node[Employee]) bool {
		return n.Data().Department == department
	}
}

// createEmployeeProvider creates a DefaultNodeProvider configured for employee hierarchy
func createEmployeeProvider() *treeview.DefaultNodeProvider[Employee] {
	// Icon rules based on department
	executiveIconRule := treeview.WithIconRule(employeeInDepartment("Executive"), "👑")
	engineeringIconRule := treeview.WithIconRule(employeeInDepartment("Engineering"), "⚙️")
	productIconRule := treeview.WithIconRule(employeeInDepartment("Product"), "📱")
	designIconRule := treeview.WithIconRule(employeeInDepartment("Design"), "🎨")
	salesIconRule := treeview.WithIconRule(employeeInDepartment("Sales"), "💼")
	marketingIconRule := treeview.WithIconRule(employeeInDepartment("Marketing"), "📢")
	defaultIconRule := treeview.WithDefaultIcon[Employee]("👤")

	return treeview.NewDefaultNodeProvider(
		executiveIconRule,
		engineeringIconRule,
		productIconRule,
		designIconRule,
		salesIconRule,
		marketingIconRule,
		defaultIconRule,
	)
}

type EmployeeTreeBuilderProvider struct{}

func (p *EmployeeTreeBuilderProvider) ID(d Employee) string {
	return d.ID
}
func (p *EmployeeTreeBuilderProvider) Name(d Employee) string {
	return fmt.Sprintf("%s (%s)", d.Title, d.Department)
}
func (p *EmployeeTreeBuilderProvider) ParentID(d Employee) string {
	return d.ManagerID
}

func showEmployeeHierarchy() (string, error) {
	// Sample flat employee data with parent-child relationships
	employees := []Employee{
		{ID: "ceo", Name: "Alice Johnson", Title: "CEO", Department: "Executive", ManagerID: "", Level: 0},
		{ID: "cto", Name: "Bob Smith", Title: "CTO", Department: "Engineering", ManagerID: "ceo", Level: 1},
		{ID: "cpo", Name: "Carol Davis", Title: "CPO", Department: "Product", ManagerID: "ceo", Level: 1},
		{ID: "cdo", Name: "Diana Wilson", Title: "CDO", Department: "Design", ManagerID: "ceo", Level: 1},
		{ID: "eng_dir", Name: "David Brown", Title: "Engineering Director", Department: "Engineering", ManagerID: "cto", Level: 2},
		{ID: "pm_dir", Name: "Eva Miller", Title: "Product Director", Department: "Product", ManagerID: "cpo", Level: 2},
		{ID: "design_lead", Name: "Frank Taylor", Title: "Design Lead", Department: "Design", ManagerID: "cdo", Level: 2},
		{ID: "senior_eng1", Name: "Grace Adams", Title: "Senior Engineer", Department: "Engineering", ManagerID: "eng_dir", Level: 3},
		{ID: "senior_eng2", Name: "Henry Chen", Title: "Senior Engineer", Department: "Engineering", ManagerID: "eng_dir", Level: 3},
		{ID: "pm1", Name: "Ivy White", Title: "Product Manager", Department: "Product", ManagerID: "pm_dir", Level: 3},
		{ID: "pm2", Name: "Jack Green", Title: "Product Manager", Department: "Product", ManagerID: "pm_dir", Level: 3},
		{ID: "designer1", Name: "Karen Blue", Title: "Senior Designer", Department: "Design", ManagerID: "design_lead", Level: 3},
		{ID: "junior_eng", Name: "Liam Gray", Title: "Junior Engineer", Department: "Engineering", ManagerID: "senior_eng1", Level: 4},
	}

	provider := createEmployeeProvider()

	tree, err := treeview.NewTreeFromFlatData(
		context.Background(),
		employees,
		&EmployeeTreeBuilderProvider{},
		treeview.WithExpandFunc(func(n *treeview.Node[Employee]) bool {
			return n.Data().Level <= 2
		}),
		treeview.WithProvider(provider),
	)
	if err != nil {
		return "", err
	}

	return tree.Render(context.Background())
}

////////////////////////////////////////////////////////////////////
//   Database Categories with Relationships
////////////////////////////////////////////////////////////////////

// Category represents a category with parent relationships
type Category struct {
	Code       string // Unique category code
	Label      string // Display label
	Kind       string // Kind (root, branch, leaf)
	ParentCode string // Parent category code
	ItemCount  int    // Number of items in this category
}

// Domain-specific predicate helper for Category
func categoryHasKind(kind string) func(*treeview.Node[Category]) bool {
	return func(n *treeview.Node[Category]) bool {
		return n.Data().Kind == kind
	}
}

// createCategoryProvider creates a DefaultNodeProvider configured for category hierarchy
func createCategoryProvider() *treeview.DefaultNodeProvider[Category] {
	// Icon rules based on category kind
	rootIconRule := treeview.WithIconRule(categoryHasKind("root"), "🗂️")
	branchIconRule := treeview.WithIconRule(categoryHasKind("branch"), "📁")
	leafIconRule := treeview.WithIconRule(categoryHasKind("leaf"), "📄")
	defaultIconRule := treeview.WithDefaultIcon[Category]("❓")

	return treeview.NewDefaultNodeProvider(
		rootIconRule,
		branchIconRule,
		leafIconRule,
		defaultIconRule,
	)
}

type CategoryTreeBuilderProvider struct{}

func (p *CategoryTreeBuilderProvider) ID(d Category) string {
	return d.Code
}
func (p *CategoryTreeBuilderProvider) Name(d Category) string {
	if d.ItemCount > 0 {
		return fmt.Sprintf("%s [%d]", d.Label, d.ItemCount)
	}
	return d.Label
}
func (p *CategoryTreeBuilderProvider) ParentID(d Category) string {
	return d.ParentCode
}

func showCategoryHierarchy() (string, error) {
	// Sample category data with parent relationships
	categories := []Category{
		{Code: "root", Label: "Product Catalog", Kind: "root", ParentCode: "", ItemCount: 0},
		{Code: "electronics", Label: "Electronics", Kind: "branch", ParentCode: "root", ItemCount: 150},
		{Code: "clothing", Label: "Clothing", Kind: "branch", ParentCode: "root", ItemCount: 300},
		{Code: "books", Label: "Books", Kind: "branch", ParentCode: "root", ItemCount: 500},
		{Code: "computers", Label: "Computers", Kind: "branch", ParentCode: "electronics", ItemCount: 45},
		{Code: "phones", Label: "Mobile Phones", Kind: "branch", ParentCode: "electronics", ItemCount: 32},
		{Code: "accessories", Label: "Accessories", Kind: "branch", ParentCode: "electronics", ItemCount: 73},
		{Code: "laptops", Label: "Laptops", Kind: "leaf", ParentCode: "computers", ItemCount: 25},
		{Code: "desktops", Label: "Desktop PCs", Kind: "leaf", ParentCode: "computers", ItemCount: 20},
		{Code: "smartphones", Label: "Smartphones", Kind: "leaf", ParentCode: "phones", ItemCount: 28},
		{Code: "tablets", Label: "Tablets", Kind: "leaf", ParentCode: "phones", ItemCount: 4},
		{Code: "mens", Label: "Men's Clothing", Kind: "branch", ParentCode: "clothing", ItemCount: 150},
		{Code: "womens", Label: "Women's Clothing", Kind: "branch", ParentCode: "clothing", ItemCount: 150},
		{Code: "shirts", Label: "Shirts", Kind: "leaf", ParentCode: "mens", ItemCount: 50},
		{Code: "pants", Label: "Pants", Kind: "leaf", ParentCode: "mens", ItemCount: 40},
		{Code: "dresses", Label: "Dresses", Kind: "leaf", ParentCode: "womens", ItemCount: 60},
		{Code: "fiction", Label: "Fiction", Kind: "leaf", ParentCode: "books", ItemCount: 200},
		{Code: "nonfiction", Label: "Non-Fiction", Kind: "leaf", ParentCode: "books", ItemCount: 300},
	}

	// Create unified category provider
	provider := createCategoryProvider()

	// Build tree from flat data
	tree, err := treeview.NewTreeFromFlatData(
		context.Background(),
		categories,
		&CategoryTreeBuilderProvider{},
		treeview.WithExpandFunc(func(n *treeview.Node[Category]) bool {
			k := n.Data().Kind
			// Expand root and branch nodes for clarity.
			return k == "root" || k == "branch"
		}),
		treeview.WithProvider(provider),
	)
	if err != nil {
		return "", err
	}

	return tree.Render(context.Background())
}

////////////////////////////////////////////////////////////////////
//   Task Dependencies Structure
////////////////////////////////////////////////////////////////////

// Task represents a task with dependencies
type Task struct {
	ID           string // Unique task identifier
	Name         string // Task name
	Status       string // Task status (pending, active, completed, blocked)
	Priority     string // Task priority (low, medium, high, critical)
	DependsOnID  string // ID of task this depends on
	EstimatedHrs int    // Estimated hours
}

// Domain-specific predicate helper for Task
func taskHasStatus(status string) func(*treeview.Node[Task]) bool {
	return func(n *treeview.Node[Task]) bool {
		return n.Data().Status == status
	}
}

// createTaskProvider creates a DefaultNodeProvider configured for task hierarchy
func createTaskProvider() *treeview.DefaultNodeProvider[Task] {
	// Icon rules based on task status
	pendingIconRule := treeview.WithIconRule(taskHasStatus("pending"), "⏳")
	activeIconRule := treeview.WithIconRule(taskHasStatus("active"), "🔄")
	completedIconRule := treeview.WithIconRule(taskHasStatus("completed"), "✅")
	blockedIconRule := treeview.WithIconRule(taskHasStatus("blocked"), "🚫")
	defaultIconRule := treeview.WithDefaultIcon[Task]("❓")

	return treeview.NewDefaultNodeProvider(
		pendingIconRule,
		activeIconRule,
		completedIconRule,
		blockedIconRule,
		defaultIconRule,
	)
}

type TaskTreeBuilderProvider struct{}

func (p *TaskTreeBuilderProvider) ID(d Task) string {
	return d.ID
}
func (p *TaskTreeBuilderProvider) Name(d Task) string {
	return fmt.Sprintf("%s – %s", d.Status, d.Name)
}
func (p *TaskTreeBuilderProvider) ParentID(d Task) string {
	return d.DependsOnID
}

func showTaskDependencies() (string, error) {
	// Sample task data with dependencies
	tasks := []Task{
		{ID: "plan", Name: "Project Planning", Status: "completed", Priority: "high", DependsOnID: "", EstimatedHrs: 8},
		{ID: "design", Name: "System Design", Status: "completed", Priority: "high", DependsOnID: "plan", EstimatedHrs: 16},
		{ID: "db_setup", Name: "Database Setup", Status: "completed", Priority: "medium", DependsOnID: "design", EstimatedHrs: 4},
		{ID: "api_dev", Name: "API Development", Status: "active", Priority: "high", DependsOnID: "db_setup", EstimatedHrs: 24},
		{ID: "ui_design", Name: "UI Design", Status: "completed", Priority: "medium", DependsOnID: "design", EstimatedHrs: 12},
		{ID: "frontend", Name: "Frontend Development", Status: "active", Priority: "high", DependsOnID: "ui_design", EstimatedHrs: 32},
		{ID: "auth", Name: "Authentication System", Status: "pending", Priority: "critical", DependsOnID: "api_dev", EstimatedHrs: 8},
		{ID: "testing", Name: "Integration Testing", Status: "pending", Priority: "high", DependsOnID: "auth", EstimatedHrs: 16},
		{ID: "user_mgmt", Name: "User Management", Status: "pending", Priority: "medium", DependsOnID: "auth", EstimatedHrs: 12},
		{ID: "deployment", Name: "Production Deployment", Status: "blocked", Priority: "critical", DependsOnID: "testing", EstimatedHrs: 4},
	}

	// Create unified task provider
	provider := createTaskProvider()

	// Build tree from flat data
	tree, err := treeview.NewTreeFromFlatData(
		context.Background(),
		tasks,
		&TaskTreeBuilderProvider{},
		treeview.WithExpandFunc(func(n *treeview.Node[Task]) bool {
			st := n.Data().Status
			// Expand completed or active tasks automatically.
			return st == "completed" || st == "active"
		}),
		treeview.WithProvider(provider),
	)
	if err != nil {
		return "", err
	}

	return tree.Render(context.Background())
}

// Main function and step runner
func main() {
	steps := []shared.ExampleStep{
		{
			Name: "Employee Organizational Chart",
			Func: showEmployeeHierarchy,
		},
		{
			Name: "Database Categories with Relationships",
			Func: showCategoryHierarchy,
		},
		{
			Name: "Task Dependencies Structure",
			Func: showTaskDependencies,
		},
	}

	shared.RunExampleStepsWithDelay("Parent-ID Relationship Builders", steps, 4)
	shared.WaitDelay(4)
}
