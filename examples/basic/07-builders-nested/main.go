package main

import (
	"context"
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/Digital-Shane/treeview/v2"
	"github.com/Digital-Shane/treeview/v2/examples/shared"
)

// ==========================================================================
// Nested Company Organization Structure
// ==========================================================================

// CompanyData represents a node in a company's organizational hierarchy
type CompanyData struct {
	ID       string        // Unique identifier for the organizational unit
	Name     string        // Display name of the unit
	Type     string        // Type of unit (company, department, team, person)
	Level    int           // Organizational level (0=company, 1=department, 2=team, 3=person)
	Children []CompanyData // Child organizational units
}

// Domain-specific predicate helpers for CompanyData
func companyHasLevel(level int) func(*treeview.Node[CompanyData]) bool {
	return func(n *treeview.Node[CompanyData]) bool {
		return n.Data().Level == level
	}
}

func companyHasType(nodeType string) func(*treeview.Node[CompanyData]) bool {
	return func(n *treeview.Node[CompanyData]) bool {
		return n.Data().Type == nodeType
	}
}

// companyFormatter formats company nodes with manager annotations
func companyFormatter(node *treeview.Node[CompanyData]) (string, bool) {
	data := node.Data()
	display := data.Name
	if data.Type == "person" && data.Level == 1 {
		display += " (Manager)"
	}
	return display, true
}

// createCompanyProvider creates a DefaultNodeProvider configured for company hierarchy
func createCompanyProvider() *treeview.DefaultNodeProvider[CompanyData] {
	// Icon rules based on company data type
	companyIconRule := treeview.WithIconRule(companyHasType("company"), "🏢")
	departmentIconRule := treeview.WithIconRule(companyHasType("department"), "🏬")
	teamIconRule := treeview.WithIconRule(companyHasType("team"), "👥")
	personIconRule := treeview.WithIconRule(companyHasType("person"), "👤")
	defaultIconRule := treeview.WithDefaultIcon[CompanyData]("❓")

	// Style rules based on organizational level
	companyStyleRule := treeview.WithStyleRule(
		companyHasLevel(0), // Company level
		lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			Foreground(lipgloss.Color("201")).
			Bold(true).
			Underline(true),
		lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			Foreground(lipgloss.Color("201")).
			Bold(true).
			Underline(true).
			Background(lipgloss.Color("238")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("201")),
	)
	departmentStyleRule := treeview.WithStyleRule(
		companyHasLevel(1), // Department level
		lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			Foreground(lipgloss.Color("39")).
			Bold(true),
		lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Background(lipgloss.Color("238")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("201")),
	)
	teamStyleRule := treeview.WithStyleRule(
		companyHasLevel(2), // Team level
		lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			Foreground(lipgloss.Color("120")).
			Italic(true),
		lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			Foreground(lipgloss.Color("120")).
			Italic(true).
			Background(lipgloss.Color("238")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("201")),
	)
	personStyleRule := treeview.WithStyleRule(
		companyHasLevel(3), // Person level
		lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			Foreground(lipgloss.Color("247")),
		lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			Foreground(lipgloss.Color("247")).
			Background(lipgloss.Color("238")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("201")),
	)
	defaultStyleRule := treeview.WithStyleRule(
		func(n *treeview.Node[CompanyData]) bool { return true }, // Default for any other level
		lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			Foreground(lipgloss.Color("250")),
		lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			Foreground(lipgloss.Color("250")).
			Background(lipgloss.Color("238")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("201")),
	)

	// Formatter rule
	formatterRule := treeview.WithFormatter[CompanyData](companyFormatter)

	return treeview.NewDefaultNodeProvider(
		// Icon rules
		companyIconRule, departmentIconRule, teamIconRule, personIconRule, defaultIconRule,
		// Style rules (specific levels first)
		companyStyleRule, departmentStyleRule, teamStyleRule, personStyleRule, defaultStyleRule,
		// Formatter
		formatterRule,
	)
}

type CompanyTreeBuilderProvider struct{}

func (p *CompanyTreeBuilderProvider) ID(d CompanyData) string {
	return d.ID
}
func (p *CompanyTreeBuilderProvider) Name(d CompanyData) string {
	return fmt.Sprintf("%s ▸ %s", d.Type, d.Name)
}
func (p *CompanyTreeBuilderProvider) Children(d CompanyData) []CompanyData {
	return d.Children
}

// ==========================================================================
// Software Project Structure
// ==========================================================================

// ProjectData represents a node in a software project hierarchy
type ProjectData struct {
	Path        string        // File or directory path
	Title       string        // Display title
	Kind        string        // Type (project, module, package, file)
	Lang        string        // Programming language (for files)
	IsImportant bool          // Whether this item is marked as important
	Items       []ProjectData // Child items
}

// Domain-specific predicate helpers for ProjectData
func projectHasKind(kind string) func(*treeview.Node[ProjectData]) bool {
	return func(n *treeview.Node[ProjectData]) bool {
		return n.Data().Kind == kind
	}
}

func projectHasLang(lang string) func(*treeview.Node[ProjectData]) bool {
	return func(n *treeview.Node[ProjectData]) bool {
		return n.Data().Lang == lang
	}
}

func projectIsImportant() func(*treeview.Node[ProjectData]) bool {
	return func(n *treeview.Node[ProjectData]) bool {
		return n.Data().IsImportant
	}
}

func projectIsFileWithLang(lang string) func(*treeview.Node[ProjectData]) bool {
	return treeview.PredAll(
		projectHasKind("file"),
		projectHasLang(lang),
	)
}

// projectFormatter formats project nodes with language annotations
func projectFormatter(node *treeview.Node[ProjectData]) (string, bool) {
	data := node.Data()
	display := data.Title
	if data.Lang != "" {
		display = fmt.Sprintf("%-20s <%s>", data.Title, data.Lang)
	}
	return display, true
}

// createProjectProvider creates a DefaultNodeProvider configured for project hierarchy
func createProjectProvider() *treeview.DefaultNodeProvider[ProjectData] {
	// Icon rules based on project data kind and language
	projectIconRule := treeview.WithIconRule(projectHasKind("project"), "🏗")
	moduleIconRule := treeview.WithIconRule(projectHasKind("module"), "📂")
	packageIconRule := treeview.WithIconRule(projectHasKind("package"), "📁")
	goFileIconRule := treeview.WithIconRule(projectIsFileWithLang("go"), "🐹")
	jsFileIconRule := treeview.WithIconRule(projectIsFileWithLang("javascript"), "🟨")
	tsFileIconRule := treeview.WithIconRule(projectIsFileWithLang("typescript"), "🔷")
	pyFileIconRule := treeview.WithIconRule(projectIsFileWithLang("python"), "🐍")
	fileIconRule := treeview.WithIconRule(projectHasKind("file"), "📄")
	defaultIconRule := treeview.WithDefaultIcon[ProjectData]("❓")

	// Style rules based on project kind and importance
	projectStyleRule := treeview.WithStyleRule(
		projectHasKind("project"),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("57")).
			Bold(true).
			PaddingLeft(2).
			PaddingRight(2),
		lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("255")).
			Bold(true).
			PaddingLeft(2).
			PaddingRight(2),
	)
	moduleStyleRule := treeview.WithStyleRule(
		projectHasKind("module"),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("228")).
			Background(lipgloss.Color("234")).
			PaddingLeft(1),
		lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("255")).
			Bold(true).
			PaddingLeft(1),
	)
	packageStyleRule := treeview.WithStyleRule(
		projectHasKind("package"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("156")),
		lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("255")).
			Bold(true),
	)
	fileStyleRule := treeview.WithStyleRule(
		projectHasKind("file"),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("251")).
			Faint(true),
		lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("255")).
			Bold(true),
	)
	importantStyleRule := treeview.WithStyleRule(
		projectIsImportant(),
		lipgloss.NewStyle().Underline(true), // Add underline to existing style for important items
		lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("255")).
			Bold(true).
			Underline(true),
	)
	defaultProjectStyleRule := treeview.WithStyleRule(
		func(n *treeview.Node[ProjectData]) bool { return true },
		lipgloss.NewStyle().Foreground(lipgloss.Color("250")),
		lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("255")).
			Bold(true),
	)

	// Formatter rule
	formatterRule := treeview.WithFormatter[ProjectData](projectFormatter)

	return treeview.NewDefaultNodeProvider(
		// Icon rules (most specific first)
		goFileIconRule, jsFileIconRule, tsFileIconRule, pyFileIconRule, fileIconRule,
		projectIconRule, moduleIconRule, packageIconRule, defaultIconRule,
		// Style rules (important rule before others to ensure it applies)
		importantStyleRule, projectStyleRule, moduleStyleRule, packageStyleRule, fileStyleRule, defaultProjectStyleRule,
		// Formatter
		formatterRule,
	)
}

type ProjectTreeBuilderProvider struct{}

func (p *ProjectTreeBuilderProvider) ID(d ProjectData) string {
	return fmt.Sprintf("%s:%s", d.Kind, d.Path)
}
func (p *ProjectTreeBuilderProvider) Name(d ProjectData) string {
	return d.Title
}
func (p *ProjectTreeBuilderProvider) Children(d ProjectData) []ProjectData {
	return d.Items
}

// ==========================================================================
// Multi-Level Menu System
// ==========================================================================

// MenuItem represents a menu item in a hierarchical menu system
type MenuItem struct {
	ID       string     // Unique identifier
	Label    string     // Display label
	Icon     string     // Menu icon
	Type     string     // Menu type (menu, submenu, action, separator)
	Enabled  bool       // Whether the menu item is enabled
	Shortcut string     // Keyboard shortcut
	Children []MenuItem // Sub-menu items
}

// Domain-specific predicate helpers for MenuItem
func menuHasType(menuType string) func(*treeview.Node[MenuItem]) bool {
	return func(n *treeview.Node[MenuItem]) bool {
		return n.Data().Type == menuType
	}
}

func menuIsEnabled() func(*treeview.Node[MenuItem]) bool {
	return func(n *treeview.Node[MenuItem]) bool {
		return n.Data().Enabled
	}
}

func menuIsEnabledOfType(menuType string) func(*treeview.Node[MenuItem]) bool {
	return treeview.PredAll(
		menuHasType(menuType),
		menuIsEnabled(),
	)
}

// menuFormatter formats menu items with shortcuts
func menuFormatter(node *treeview.Node[MenuItem]) (string, bool) {
	item := node.Data()
	display := item.Label
	// Right-align shortcuts
	if item.Shortcut != "" {
		spaces := 25 - len(display)
		if spaces < 2 {
			spaces = 2
		}
		display += fmt.Sprintf("%*s%s", spaces, "", item.Shortcut)
	}
	return display, true
}

// createMenuProvider creates a DefaultNodeProvider configured for menu system
func createMenuProvider() *treeview.DefaultNodeProvider[MenuItem] {
	// Icon rules based on menu item type
	menuIconRule := treeview.WithIconRule(menuHasType("menu"), "📋")
	submenuIconRule := treeview.WithIconRule(menuHasType("submenu"), "📂")
	actionIconRule := treeview.WithIconRule(menuHasType("action"), "⚡")
	defaultMenuIconRule := treeview.WithDefaultIcon[MenuItem]("❓")

	// Style rules based on menu type and enabled state
	disabledStyleRule := treeview.WithStyleRule(
		treeview.PredNot(menuIsEnabled()),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Strikethrough(true),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Strikethrough(true), // Same for focused
	)
	menuStyleRule := treeview.WithStyleRule(
		menuIsEnabledOfType("menu"),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("213")).
			Bold(true),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("213")).
			Bold(true).
			Background(lipgloss.Color("236")).
			PaddingLeft(1).
			PaddingRight(1),
	)
	submenuStyleRule := treeview.WithStyleRule(
		menuIsEnabledOfType("submenu"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("111")),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("111")).
			Bold(true),
	)
	actionStyleRule := treeview.WithStyleRule(
		menuIsEnabledOfType("action"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("114")),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Background(lipgloss.Color("236")),
	)
	defaultMenuStyleRule := treeview.WithStyleRule(
		func(n *treeview.Node[MenuItem]) bool { return n.Data().Enabled },
		lipgloss.NewStyle().Foreground(lipgloss.Color("250")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("250")),
	)

	// Formatter rule
	formatterRule := treeview.WithFormatter[MenuItem](menuFormatter)

	return treeview.NewDefaultNodeProvider(
		// Icon rules
		menuIconRule, submenuIconRule, actionIconRule, defaultMenuIconRule,
		// Style rules (disabled first to override others)
		disabledStyleRule, menuStyleRule, submenuStyleRule, actionStyleRule, defaultMenuStyleRule,
		// Formatter
		formatterRule,
	)
}

// We need a custom wrapper for menu icons since DefaultNodeProvider doesn't easily support dynamic icons
type MenuProvider struct {
	*treeview.DefaultNodeProvider[MenuItem]
}

func NewMenuProvider() *MenuProvider {
	return &MenuProvider{
		DefaultNodeProvider: createMenuProvider(),
	}
}

func (p *MenuProvider) Icon(node *treeview.Node[MenuItem]) string {
	data := node.Data()
	if data.Icon != "" {
		return data.Icon
	}
	return p.DefaultNodeProvider.Icon(node)
}

type MenuTreeBuilderProvider struct{}

func (p *MenuTreeBuilderProvider) ID(d MenuItem) string {
	return d.ID
}
func (p *MenuTreeBuilderProvider) Name(d MenuItem) string {
	return d.Label
}
func (p *MenuTreeBuilderProvider) Children(d MenuItem) []MenuItem {
	return d.Children
}

func showCompanyHierarchy() (string, error) {
	companyData := CompanyData{
		ID:    "techcorp",
		Name:  "TechCorp Industries",
		Type:  "company",
		Level: 0,
		Children: []CompanyData{
			{
				ID:    "engineering",
				Name:  "Engineering",
				Type:  "department",
				Level: 1,
				Children: []CompanyData{
					{
						ID:    "platform",
						Name:  "Platform Team",
						Type:  "team",
						Level: 2,
						Children: []CompanyData{
							{ID: "alex", Name: "Alex Kim", Type: "person", Level: 3},
							{ID: "sarah", Name: "Sarah Chen", Type: "person", Level: 3},
							{ID: "mike", Name: "Mike Johnson", Type: "person", Level: 3},
						},
					},
					{
						ID:    "mobile",
						Name:  "Mobile Team",
						Type:  "team",
						Level: 2,
						Children: []CompanyData{
							{ID: "lisa", Name: "Lisa Park", Type: "person", Level: 3},
							{ID: "tom", Name: "Tom Wilson", Type: "person", Level: 3},
						},
					},
				},
			},
			{
				ID:    "design",
				Name:  "Design",
				Type:  "department",
				Level: 1,
				Children: []CompanyData{
					{
						ID:    "ux",
						Name:  "UX Team",
						Type:  "team",
						Level: 2,
						Children: []CompanyData{
							{ID: "emma", Name: "Emma Davis", Type: "person", Level: 3},
							{ID: "david", Name: "David Brown", Type: "person", Level: 3},
						},
					},
				},
			},
		},
	}

	tree, err := treeview.NewTreeFromNestedData(
		context.Background(),
		[]CompanyData{companyData},
		&CompanyTreeBuilderProvider{},
		treeview.WithExpandAll[CompanyData](),
		treeview.WithProvider(createCompanyProvider()),
	)
	if err != nil {
		return "", err
	}

	return tree.Render(context.Background())
}

func showProjectHierarchy() (string, error) {
	projectData := ProjectData{
		Path:        "webapp",
		Title:       "Modern Web Application",
		Kind:        "project",
		Lang:        "",
		IsImportant: true,
		Items: []ProjectData{
			{
				Path:        "frontend",
				Title:       "Frontend Module",
				Kind:        "module",
				Lang:        "",
				IsImportant: true,
				Items: []ProjectData{
					{
						Path:        "components",
						Title:       "Components",
						Kind:        "package",
						Lang:        "",
						IsImportant: false,
						Items: []ProjectData{
							{Path: "button", Title: "Button.tsx", Kind: "file", Lang: "typescript"},
							{Path: "modal", Title: "Modal.tsx", Kind: "file", Lang: "typescript"},
						},
					},
					{
						Path:  "utils",
						Title: "Utilities",
						Kind:  "package",
						Lang:  "",
						Items: []ProjectData{
							{Path: "api", Title: "api.ts", Kind: "file", Lang: "typescript"},
							{Path: "helpers", Title: "helpers.ts", Kind: "file", Lang: "typescript"},
						},
					},
				},
			},
			{
				Path:        "backend",
				Title:       "Backend Module",
				Kind:        "module",
				Lang:        "",
				IsImportant: true,
				Items: []ProjectData{
					{
						Path:  "handlers",
						Title: "HTTP Handlers",
						Kind:  "package",
						Lang:  "",
						Items: []ProjectData{
							{Path: "user", Title: "user.go", Kind: "file", Lang: "go"},
							{Path: "auth", Title: "auth.go", Kind: "file", Lang: "go"},
						},
					},
					{
						Path:  "models",
						Title: "Data Models",
						Kind:  "package",
						Lang:  "",
						Items: []ProjectData{
							{Path: "user_model", Title: "user.go", Kind: "file", Lang: "go"},
							{Path: "session", Title: "session.go", Kind: "file", Lang: "go"},
						},
					},
				},
			},
		},
	}

	tree, err := treeview.NewTreeFromNestedData(
		context.Background(),
		[]ProjectData{projectData},
		&ProjectTreeBuilderProvider{},
		treeview.WithFilterFunc(func(d ProjectData) bool {
			return d.IsImportant || d.Kind == "file"
		}),
		treeview.WithExpandFunc(func(n *treeview.Node[ProjectData]) bool {
			return n.Data().IsImportant
		}),
		treeview.WithProvider(createProjectProvider()),
	)
	if err != nil {
		return "", err
	}

	return tree.Render(context.Background())
}

func showMenuHierarchy() (string, error) {
	menuData := []MenuItem{
		{
			ID:      "file",
			Label:   "File",
			Icon:    "📁",
			Type:    "menu",
			Enabled: true,
			Children: []MenuItem{
				{ID: "new", Label: "New", Type: "action", Enabled: true, Shortcut: "Ctrl+N"},
				{ID: "open", Label: "Open", Type: "action", Enabled: true, Shortcut: "Ctrl+O"},
				{ID: "save", Label: "Save", Type: "action", Enabled: true, Shortcut: "Ctrl+S"},
				{ID: "save_as", Label: "Save As", Type: "action", Enabled: false, Shortcut: "Ctrl+Shift+S"},
			},
		},
		{
			ID:      "edit",
			Label:   "Edit",
			Icon:    "✏️",
			Type:    "menu",
			Enabled: true,
			Children: []MenuItem{
				{ID: "undo", Label: "Undo", Type: "action", Enabled: true, Shortcut: "Ctrl+Z"},
				{ID: "redo", Label: "Redo", Type: "action", Enabled: false, Shortcut: "Ctrl+Y"},
				{ID: "cut", Label: "Cut", Type: "action", Enabled: true, Shortcut: "Ctrl+X"},
				{ID: "copy", Label: "Copy", Type: "action", Enabled: true, Shortcut: "Ctrl+C"},
				{ID: "paste", Label: "Paste", Type: "action", Enabled: true, Shortcut: "Ctrl+V"},
			},
		},
		{
			ID:      "view",
			Label:   "View",
			Icon:    "👁️",
			Type:    "menu",
			Enabled: true,
			Children: []MenuItem{
				{
					ID:      "zoom",
					Label:   "Zoom",
					Type:    "submenu",
					Enabled: true,
					Children: []MenuItem{
						{ID: "zoom_in", Label: "Zoom In", Type: "action", Enabled: true, Shortcut: "Ctrl++"},
						{ID: "zoom_out", Label: "Zoom Out", Type: "action", Enabled: true, Shortcut: "Ctrl+-"},
						{ID: "zoom_reset", Label: "Reset Zoom", Type: "action", Enabled: true, Shortcut: "Ctrl+0"},
					},
				},
				{ID: "fullscreen", Label: "Full Screen", Type: "action", Enabled: true, Shortcut: "F11"},
			},
		},
	}

	tree, err := treeview.NewTreeFromNestedData(
		context.Background(),
		menuData,
		&MenuTreeBuilderProvider{},
		treeview.WithExpandFunc(func(n *treeview.Node[MenuItem]) bool {
			return n.Data().Type == "menu"
		}),
		treeview.WithProvider(NewMenuProvider()),
	)
	if err != nil {
		return "", err
	}

	return tree.Render(context.Background())
}

// ==========================================================================
// Filtered Organizational View
// ==========================================================================

func showFilteredHierarchy() (string, error) {
	companyData := CompanyData{
		ID:    "techcorp",
		Name:  "TechCorp Industries",
		Type:  "company",
		Level: 0,
		Children: []CompanyData{
			{
				ID:    "engineering",
				Name:  "Engineering",
				Type:  "department",
				Level: 1,
				Children: []CompanyData{
					{
						ID:    "platform",
						Name:  "Platform Team",
						Type:  "team",
						Level: 2,
						Children: []CompanyData{
							{ID: "alex", Name: "Alex Kim (Lead)", Type: "person", Level: 3},
							{ID: "sarah", Name: "Sarah Chen", Type: "person", Level: 3},
							{ID: "mike", Name: "Mike Johnson", Type: "person", Level: 3},
						},
					},
					{
						ID:    "mobile",
						Name:  "Mobile Team",
						Type:  "team",
						Level: 2,
						Children: []CompanyData{
							{ID: "lisa", Name: "Lisa Park (Lead)", Type: "person", Level: 3},
							{ID: "tom", Name: "Tom Wilson", Type: "person", Level: 3},
						},
					},
				},
			},
			{
				ID:    "design",
				Name:  "Design",
				Type:  "department",
				Level: 1,
				Children: []CompanyData{
					{
						ID:    "ux",
						Name:  "UX Team",
						Type:  "team",
						Level: 2,
						Children: []CompanyData{
							{ID: "emma", Name: "Emma Davis (Lead)", Type: "person", Level: 3},
							{ID: "david", Name: "David Brown", Type: "person", Level: 3},
						},
					},
				},
			},
		},
	}

	// Build organisational view while filtering out deeper hierarchy levels.
	tree, err := treeview.NewTreeFromNestedData(
		context.Background(),
		[]CompanyData{companyData},
		&CompanyTreeBuilderProvider{},
		treeview.WithExpandAll[CompanyData](),
		treeview.WithFilterFunc(func(d CompanyData) bool { return d.Level <= 2 }),
		treeview.WithProvider(createCompanyProvider()),
	)
	if err != nil {
		return "", err
	}

	return tree.Render(context.Background())
}

func main() {
	examples := []shared.ExampleStep{
		{Name: "Company Organization Structure", Func: showCompanyHierarchy},
		{Name: "Project Structure with Conditional Expansion", Func: showProjectHierarchy},
		{Name: "Multi-Level Menu System", Func: showMenuHierarchy},
		{Name: "Filtered Organizational View", Func: showFilteredHierarchy},
	}

	shared.RunExampleStepsWithDelay("Nested Data Builders", examples, 4)
	shared.WaitDelay(4)
}
