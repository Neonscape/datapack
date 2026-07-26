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
