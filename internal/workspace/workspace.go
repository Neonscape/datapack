package workspace

import (
	"fmt"
	"os"

	"datapack/internal/document"
	"datapack/internal/projecttree"
	"datapack/internal/schemaregistry"
)

type View struct {
	Project  Project              `json:"project"`
	Settings EditorSettings       `json:"settings"`
	Tree     projecttree.Tree     `json:"tree"`
	Schemas  schemaregistry.Index `json:"schemas"`
}

type Service struct {
	configRoot string
	current    *View
}

type CellUpdate struct {
	Path       string `json:"path"`
	RowID      string `json:"rowID"`
	Column     string `json:"column"`
	Expression string `json:"expression"`
}

type CellUpdateResult struct {
	Table     document.Table `json:"table"`
	Workspace View           `json:"workspace"`
}

type RowAdd struct {
	Path       string `json:"path"`
	RowID      string `json:"rowID"`
	Expression string `json:"expression"`
}

type RowAddResult struct {
	Table     document.Table `json:"table"`
	Workspace View           `json:"workspace"`
}

func NewService(configRoot string) *Service {
	return &Service{configRoot: configRoot}
}

func NewUserService() (*Service, error) {
	configRoot, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user config directory: %w", err)
	}
	return NewService(configRoot), nil
}

func (s *Service) OpenProject(path string) (View, error) {
	project, err := LoadProject(path)
	if err != nil {
		return View{}, err
	}
	settings, err := LoadEditorSettings(s.configRoot)
	if err != nil {
		return View{}, err
	}
	tree, err := projecttree.Scan(project.Path, project.Config.TableRoots, project.Config.SchemaRoots, project.Config.Exclude)
	if err != nil {
		return View{}, err
	}
	view := View{Project: project, Settings: settings, Tree: tree, Schemas: buildSchemaIndex(tree)}
	s.current = &view
	return view, nil
}

func buildSchemaIndex(tree projecttree.Tree) schemaregistry.Index {
	var publicSchemas []document.PublicSchema
	var tables []document.Table
	collectDocuments(tree.Root, &publicSchemas, &tables)
	return schemaregistry.Build(publicSchemas, tables).Index()
}

func collectDocuments(node projecttree.Node, publicSchemas *[]document.PublicSchema, tables *[]document.Table) {
	switch node.Kind {
	case projecttree.NodeSchemaFile:
		if schema, err := document.LoadPublicSchema(node.Path); err == nil {
			*publicSchemas = append(*publicSchemas, schema)
		}
	case projecttree.NodeTable:
		if table, err := document.LoadTable(node.Path); err == nil {
			*tables = append(*tables, table)
		}
	}
	for _, child := range node.Children {
		collectDocuments(child, publicSchemas, tables)
	}
}

func (s *Service) Current() (View, bool) {
	if s.current == nil {
		return View{}, false
	}
	return *s.current, true
}

// GetTable loads a table document that belongs to the current project. The
// path is resolved through the project tree instead of being read directly from
// an untrusted frontend request.
func (s *Service) GetTable(path string) (document.Table, error) {
	if s.current == nil {
		return document.Table{}, fmt.Errorf("no project is open")
	}
	if !containsTable(s.current.Tree.Root, path) {
		return document.Table{}, fmt.Errorf("table %q is not part of the open project", path)
	}
	table, err := document.LoadTable(path)
	if err != nil {
		return document.Table{}, fmt.Errorf("load table: %w", err)
	}
	return table, nil
}

// UpdateCell applies one CUE expression to an existing table cell, then
// refreshes the project tree and Schema registry from the saved source.
func (s *Service) UpdateCell(update CellUpdate) (CellUpdateResult, error) {
	if s.current == nil {
		return CellUpdateResult{}, fmt.Errorf("no project is open")
	}
	if !containsTable(s.current.Tree.Root, update.Path) {
		return CellUpdateResult{}, fmt.Errorf("table %q is not part of the open project", update.Path)
	}
	table, err := document.SetCellValue(update.Path, update.RowID, update.Column, update.Expression)
	if err != nil {
		return CellUpdateResult{}, err
	}
	view, err := s.refreshCurrentView()
	if err != nil {
		return CellUpdateResult{}, err
	}
	return CellUpdateResult{Table: table, Workspace: view}, nil
}

func (s *Service) AddRow(add RowAdd) (RowAddResult, error) {
	if s.current == nil {
		return RowAddResult{}, fmt.Errorf("no project is open")
	}
	if !containsTable(s.current.Tree.Root, add.Path) {
		return RowAddResult{}, fmt.Errorf("table %q is not part of the open project", add.Path)
	}
	table, err := document.AddRow(add.Path, add.RowID, add.Expression)
	if err != nil {
		return RowAddResult{}, err
	}
	view, err := s.refreshCurrentView()
	if err != nil {
		return RowAddResult{}, err
	}
	return RowAddResult{Table: table, Workspace: view}, nil
}

func (s *Service) refreshCurrentView() (View, error) {
	project := s.current.Project
	tree, err := projecttree.Scan(project.Path, project.Config.TableRoots, project.Config.SchemaRoots, project.Config.Exclude)
	if err != nil {
		return View{}, fmt.Errorf("refresh project tree: %w", err)
	}
	view := View{Project: project, Settings: s.current.Settings, Tree: tree, Schemas: buildSchemaIndex(tree)}
	s.current = &view
	return view, nil
}

func containsTable(node projecttree.Node, path string) bool {
	if node.Kind == projecttree.NodeTable && node.Path == path {
		return true
	}
	for _, child := range node.Children {
		if containsTable(child, path) {
			return true
		}
	}
	return false
}
