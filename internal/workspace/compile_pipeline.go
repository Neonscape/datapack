package workspace

import (
	"datapack/internal/compiler"
	"datapack/internal/document"
	"datapack/internal/foreignkey"
	"datapack/internal/logging"
	"datapack/internal/perftrace"
	"datapack/internal/projecttree"
	"datapack/internal/schemaregistry"
	"strings"
	"time"
)

// compile runs one full compile against the current tree and publishes the
// result into the current view.
func (s *Service) compile() compiler.Result {
	defer perftrace.Start("workspace.compile")()
	started := time.Now()
	if err := s.refreshExternalCompileInputs(); err != nil {
		result := compiler.Result{Diagnostics: []compiler.Diagnostic{{Message: err.Error()}}}
		s.current.Compile = s.recordCompile(s.current.Settings.Build.Mode, time.Now().UTC().Format(time.RFC3339), result.Diagnostics)
		return result
	}
	before, inputErr := s.captureInputs()
	fingerprint, cacheable := before.fingerprint, inputErr == nil && before.cacheable
	if inputErr == nil && cacheable && s.successfulCompile != nil && s.successfulCompile.fingerprint == fingerprint {
		defer perftrace.Start("compiler.successCacheHit")()
		result := s.successfulCompile.result
		result.InputFingerprint = fingerprint
		if s.inputHooks != nil && s.inputHooks.afterCacheHit != nil {
			s.inputHooks.afterCacheHit()
		}
		latest, err := s.captureInputs()
		if err != nil || latest.fingerprint != fingerprint {
			result.Package = nil
			result.Diagnostics = []compiler.Diagnostic{{Message: "project inputs changed after cache lookup; retry"}}
		}
		s.current.Compile = s.recordCompile(s.current.Settings.Build.Mode, time.Now().UTC().Format(time.RFC3339), result.Diagnostics)
		return result
	}
	var publicSchemas []document.PublicSchema
	var tables []document.Table
	collectionDiagnostics := collectDocuments(s.current.Tree.Root, &publicSchemas, &tables)
	result := compiler.CompileAll(publicSchemas, tables)
	result.Diagnostics = append(collectionDiagnostics, result.Diagnostics...)
	appendTreeDiagnostics(s.current.Tree.Root, &result)
	after, afterErr := s.captureInputs()
	if s.inputHooks != nil && s.inputHooks.afterCompile != nil {
		s.inputHooks.afterCompile()
		after, afterErr = s.captureInputs()
	}
	if inputErr != nil || afterErr != nil || after.fingerprint != fingerprint {
		result.Package = nil
		if len(result.Diagnostics) == 0 {
			result.Diagnostics = append(result.Diagnostics, compiler.Diagnostic{Message: "project input snapshot unavailable or changed during compilation; retry"})
		}
	} else if cacheable && len(result.Diagnostics) == 0 && result.Package != nil {
		result.InputFingerprint = fingerprint
		s.successfulCompile = &successfulCompile{fingerprint, result}
	}

	if result.Package != nil && len(result.Diagnostics) == 0 {
		result.InputFingerprint = fingerprint
	}
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
	defer perftrace.Start("workspace.buildSchemaIndex")()
	var publicSchemas []document.PublicSchema
	var tables []document.Table
	_ = collectDocuments(tree.Root, &publicSchemas, &tables)
	return schemaregistry.Build(publicSchemas, tables).Index()
}

func buildForeignKeyIndex(tree projecttree.Tree) foreignkey.Index {
	defer perftrace.Start("workspace.buildForeignKeyIndex")()
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
	defer perftrace.Start("workspace.buildView")()
	var schemas []document.PublicSchema
	var tables []document.Table
	collectionDiagnostics := collectDocuments(tree.Root, &schemas, &tables)
	// The same immutable documents serve both indexes in this operation.
	fk := foreignkey.Build(tables)
	for _, d := range collectionDiagnostics {
		if strings.HasPrefix(d.Message, "load table: ") {
			fk.Diagnostics = append(fk.Diagnostics, foreignkey.Diagnostic{Path: d.Path, Message: "cannot load reference source: " + strings.TrimPrefix(d.Message, "load table: ")})
			fk.Incomplete = true
		}
	}
	var damaged func(projecttree.Node)
	damaged = func(n projecttree.Node) {
		if n.Kind == projecttree.NodeDamaged {
			fk.Diagnostics = append(fk.Diagnostics, foreignkey.Diagnostic{Path: n.Path, Message: "cannot resolve references from damaged source"})
			fk.Incomplete = true
		}
		for _, c := range n.Children {
			damaged(c)
		}
	}
	damaged(tree.Root)
	return View{Project: project, Settings: settings, Tree: tree, Schemas: schemaregistry.Build(schemas, tables).Index(), ForeignKeys: fk.Summary()}
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
