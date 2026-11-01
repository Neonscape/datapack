package workspace

import (
	"datapack/internal/compiler"
	"datapack/internal/document"
	"datapack/internal/foreignkey"
	"datapack/internal/logging"
	"datapack/internal/projecttree"
	"datapack/internal/schemaregistry"
	"time"
)

// compile runs one full compile against the current tree and publishes the
// result into the current view.
func (s *Service) compile() compiler.Result {
	started := time.Now()
	var publicSchemas []document.PublicSchema
	var tables []document.Table
	collectionDiagnostics := collectDocuments(s.current.Tree.Root, &publicSchemas, &tables)
	result := compiler.CompileAll(publicSchemas, tables)
	result.Diagnostics = append(collectionDiagnostics, result.Diagnostics...)
	appendTreeDiagnostics(s.current.Tree.Root, &result)
	state := s.recordCompile(s.current.Settings.Build.Mode, time.Now().UTC().Format(time.RFC3339), result.Diagnostics)
	if s.current != nil {
		s.current.Compile = state
	}
	s.log(logging.Info, "compile completed", map[string]any{"diagnostics": len(result.Diagnostics), "mode": state.Mode, "duration_ms": time.Since(started).Milliseconds()})
	return result
}

func appendTreeDiagnostics(node projecttree.Node, result *compiler.Result) {
	if node.Kind == projecttree.NodeDamaged {
		for _, diagnostic := range node.Diagnostics {
			result.Diagnostics = append(result.Diagnostics, compiler.Diagnostic{Path: node.Path, Key: "workspace.documentLoad", Message: diagnostic.Message})
		}
	}
	for _, child := range node.Children {
		appendTreeDiagnostics(child, result)
	}
}

// publishView stamps the compile state onto a freshly built view.
func (s *Service) publishView(view *View) {
	*view = s.viewLifecycle.publish(*view, s.current.Settings.Build.Mode, len(s.undoStack) > 0, len(s.redoStack) > 0)

	s.current = view
}

func buildSchemaIndex(tree projecttree.Tree) schemaregistry.Index {
	var publicSchemas []document.PublicSchema
	var tables []document.Table
	_ = collectDocuments(tree.Root, &publicSchemas, &tables)
	return schemaregistry.Build(publicSchemas, tables).Index()
}

func buildForeignKeyIndex(tree projecttree.Tree) foreignkey.Index {
	var tables []document.Table
	var diagnostics []foreignkey.Diagnostic
	var collect func(projecttree.Node)
	collect = func(node projecttree.Node) {
		switch node.Kind {
		case projecttree.NodeDamaged:
			diagnostics = append(diagnostics, foreignkey.Diagnostic{Path: node.Path, Message: "cannot resolve references from damaged source"})
		case projecttree.NodeTable:
			table, err := document.LoadTable(node.Path)
			if err != nil {
				diagnostics = append(diagnostics, foreignkey.Diagnostic{Path: node.Path, Message: "cannot load reference source: " + err.Error()})
			} else {
				tables = append(tables, table)
			}
		}
		for _, child := range node.Children {
			collect(child)
		}
	}
	collect(tree.Root)
	index := foreignkey.Build(tables)
	index.Diagnostics = append(diagnostics, index.Diagnostics...)
	index.Incomplete = index.Incomplete || len(diagnostics) > 0
	return index
}

func buildView(project Project, settings EditorSettings, tree projecttree.Tree) View {
	return View{
		Project:     project,
		Settings:    settings,
		Tree:        tree,
		Schemas:     buildSchemaIndex(tree),
		ForeignKeys: buildForeignKeyIndex(tree).Summary(),
	}
}

func collectDocuments(node projecttree.Node, publicSchemas *[]document.PublicSchema, tables *[]document.Table) []compiler.Diagnostic {
	seen := map[string]bool{}
	var diagnostics []compiler.Diagnostic
	collectDocumentsInto(node, publicSchemas, tables, seen, &diagnostics)
	return diagnostics
}

func collectDocumentsInto(node projecttree.Node, publicSchemas *[]document.PublicSchema, tables *[]document.Table, seenSchemas map[string]bool, diagnostics *[]compiler.Diagnostic) {
	switch node.Kind {
	case projecttree.NodeSchemaDefinition:
		if node.Path != "" && !seenSchemas[node.Path] {
			seenSchemas[node.Path] = true
			if schema, err := document.LoadPublicSchema(node.Path); err == nil {
				*publicSchemas = append(*publicSchemas, schema)
			} else {
				*diagnostics = append(*diagnostics, compiler.Diagnostic{Path: node.Path, Key: "workspace.documentLoad", Message: "load public Schema: " + err.Error()})
			}
		}
	case projecttree.NodeTable:
		if table, err := document.LoadTable(node.Path); err == nil {
			*tables = append(*tables, table)
		} else {
			*diagnostics = append(*diagnostics, compiler.Diagnostic{Path: node.Path, Table: node.Name, Key: "workspace.documentLoad", Message: "load table: " + err.Error()})
		}
	}
	for _, child := range node.Children {
		collectDocumentsInto(child, publicSchemas, tables, seenSchemas, diagnostics)
	}
}
