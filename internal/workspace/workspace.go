package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/cue/parser"
	"cuelang.org/go/cue/token"
	"datapack/internal/compiler"
	"datapack/internal/document"
	"datapack/internal/export/csv"
	"datapack/internal/foreignkey"
	"datapack/internal/formula"
	"datapack/internal/logging"
	"datapack/internal/navigation"
	"datapack/internal/projecttree"
	"datapack/internal/schemaregistry"
)

type View struct {
	Project     Project              `json:"project"`
	Settings    EditorSettings       `json:"settings"`
	Tree        projecttree.Tree     `json:"tree"`
	Schemas     schemaregistry.Index `json:"schemas"`
	ForeignKeys foreignkey.Summary   `json:"foreignKeys"`
	Compile     CompileState         `json:"compile"`
}

// CompileState carries the current build-mode state and the outcome of the
// most recent full compile.
type CompileState struct {
	Mode        string                `json:"mode"`
	Pending     bool                  `json:"pending"`
	CompiledAt  string                `json:"compiledAt,omitempty"`
	Diagnostics []compiler.Diagnostic `json:"diagnostics"`
}

type Service struct {
	configRoot       string
	current          *View
	history          map[string]*history
	logger           *logging.Logger
	revision         int
	compiledRevision int
	compileState     CompileState
}

type history struct {
	undo [][]byte
	redo [][]byte
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

type Paste struct {
	Path    string     `json:"path"`
	RowIDs  []string   `json:"rowIDs"`
	Columns []string   `json:"columns"`
	Values  [][]string `json:"values"`
}

type HistoryResult struct {
	Table     document.Table `json:"table"`
	Workspace View           `json:"workspace"`
	CanUndo   bool           `json:"canUndo"`
	CanRedo   bool           `json:"canRedo"`
}

type FormulaSet struct {
	Path       string `json:"path"`
	RowID      string `json:"rowID"`
	Column     string `json:"column"`
	Expression string `json:"expression"`
}

type FormulaSetResult struct {
	Table     document.Table `json:"table"`
	Workspace View           `json:"workspace"`
}

type RowCreate struct {
	Path     string            `json:"path"`
	RowID    string            `json:"rowID"`
	Values   map[string]string `json:"values"`
	Formulas map[string]string `json:"formulas"`
}

type RowCreateResult struct {
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

type RowDelete struct {
	Path  string `json:"path"`
	RowID string `json:"rowID"`
}

type RowDeleteResult struct {
	Table     document.Table `json:"table"`
	Workspace View           `json:"workspace"`
}

type RowRename struct {
	Path  string `json:"path"`
	OldID string `json:"oldID"`
	NewID string `json:"newID"`
}

type RowRenameResult struct {
	Table     document.Table `json:"table"`
	Workspace View           `json:"workspace"`
}

type ExportResult struct {
	Files []string `json:"files"`
}

type CompileResult struct {
	Workspace View `json:"workspace"`
}

// Compile runs one full compile and publishes the outcome into the workspace.
func (s *Service) Compile() (CompileResult, error) {
	if s.current == nil {
		return CompileResult{}, fmt.Errorf("no project is open")
	}
	s.compile()
	return CompileResult{Workspace: *s.current}, nil
}

// SetCompileMode switches between auto and manual compilation and persists the
// user-level editor settings.
func (s *Service) SetCompileMode(mode string) (View, error) {
	if s.current == nil {
		return View{}, fmt.Errorf("no project is open")
	}
	if mode != "auto" && mode != "manual" {
		return View{}, fmt.Errorf("unsupported build mode %q", mode)
	}
	settings := s.current.Settings
	settings.Build.Mode = mode
	if err := SaveEditorSettings(s.configRoot, settings); err != nil {
		s.log(logging.Error, "save editor settings failed", map[string]any{"error": err.Error()})
		return View{}, err
	}
	s.current.Settings = settings
	tree, err := projecttree.Scan(s.current.Project.Path, s.current.Project.Config.TableRoots, s.current.Project.Config.SchemaRoots, s.current.Project.Config.Exclude)
	if err != nil {
		s.log(logging.Error, "workspace refresh failed", map[string]any{"error": err.Error()})
		return View{}, fmt.Errorf("refresh project tree: %w", err)
	}
	view := buildView(s.current.Project, settings, tree)
	s.publishView(&view)
	s.log(logging.Info, "compile mode changed", map[string]any{"mode": mode})
	return view, nil
}

// Navigate resolves one semantic jump to a document and location.
func (s *Service) Navigate(request navigation.Request) (navigation.Target, error) {
	if s.current == nil {
		return navigation.Target{}, fmt.Errorf("no project is open")
	}
	switch request.Kind {
	case navigation.KindSchemaReference:
		return s.navigateSchemaReference(request)
	case navigation.KindForeignKey:
		return s.navigateForeignKey(request)
	case navigation.KindFormula:
		return s.navigateFormula(request)
	case navigation.KindPrimaryKey:
		return s.navigatePrimaryKey(request)
	default:
		return navigation.Target{}, fmt.Errorf("unsupported navigation kind %q", request.Kind)
	}
}

func (s *Service) navigateSchemaReference(request navigation.Request) (navigation.Target, error) {
	reference := request.Reference
	name := reference
	if marker := strings.LastIndex(reference, ".#"); marker >= 0 {
		name = reference[marker+2:]
	} else {
		name = strings.TrimPrefix(reference, "#")
	}
	id := ""
	switch {
	case strings.HasPrefix(reference, "_local.#"):
		table, err := document.LoadTable(request.Path)
		if err != nil {
			return navigation.Target{}, err
		}
		id = logicalTableName(table) + "._local.#" + name
	case strings.HasPrefix(reference, "#"):
		schema, err := document.LoadPublicSchema(request.Path)
		if err != nil {
			return navigation.Target{}, err
		}
		id = schema.Package + ".#" + name
	default:
		id = reference
	}
	for _, definition := range s.current.Schemas.Definitions {
		if string(definition.ID) == id {
			return navigation.Target{
				Kind:       navigation.KindSchemaReference,
				Path:       definition.Path,
				Local:      definition.Table != "",
				Definition: name,
			}, nil
		}
	}
	return navigation.Target{}, fmt.Errorf("Schema reference %s not found", reference)
}

func (s *Service) navigateForeignKey(request navigation.Request) (navigation.Target, error) {
	source, err := document.LoadTable(request.Path)
	if err != nil {
		return navigation.Target{}, err
	}
	targetTable := ""
	for _, key := range source.ForeignKeys {
		if key.Column == request.Column {
			targetTable = key.Table
			break
		}
	}
	if targetTable == "" {
		return navigation.Target{}, fmt.Errorf("column %q is not a foreign key", request.Column)
	}
	var tables []document.Table
	collectTables(s.current.Tree.Root, &tables)
	for _, table := range tables {
		if logicalTableName(table) == targetTable {
			return navigation.Target{Kind: navigation.KindForeignKey, Path: table.Path, RowID: request.RowID, Column: "id"}, nil
		}
	}
	return navigation.Target{}, fmt.Errorf("foreign key target table %q not found", targetTable)
}

func (s *Service) navigateFormula(request navigation.Request) (navigation.Target, error) {
	table, err := document.LoadTable(request.Path)
	if err != nil {
		return navigation.Target{}, err
	}
	for _, entry := range table.Formulas {
		if entry.RowID != request.RowID || entry.Column != request.Column {
			continue
		}
		references := formula.References(table, entry)
		if len(references) == 0 {
			return navigation.Target{}, fmt.Errorf("formula has no resolvable references")
		}
		first := references[0]
		return navigation.Target{Kind: navigation.KindFormula, Path: table.Path, RowID: first.RowID, Column: first.Column}, nil
	}
	return navigation.Target{}, fmt.Errorf("no formula at %s.%s", request.RowID, request.Column)
}

func (s *Service) navigatePrimaryKey(request navigation.Request) (navigation.Target, error) {
	table, err := document.LoadTable(request.Path)
	if err != nil {
		return navigation.Target{}, err
	}
	name := logicalTableName(table)
	var tables []document.Table
	collectTables(s.current.Tree.Root, &tables)
	usages := foreignkey.Build(tables).Usages[foreignkey.Target{Table: name, RowID: request.RowID}]
	if len(usages) == 0 {
		return navigation.Target{}, fmt.Errorf("no references to %s.%s", name, request.RowID)
	}
	if len(usages) == 1 {
		usage := usages[0]
		return navigation.Target{Kind: navigation.KindPrimaryKey, Path: usage.Path, RowID: usage.RowID, Column: usage.Column}, nil
	}
	target := navigation.Target{Kind: navigation.KindPrimaryKey}
	for _, usage := range usages {
		target.References = append(target.References, navigation.Target{Kind: navigation.KindPrimaryKey, Path: usage.Path, RowID: usage.RowID, Column: usage.Column})
	}
	return target, nil
}

func logicalTableName(table document.Table) string {
	if table.Namespace == "" {
		return table.Name
	}
	return table.Namespace + "." + table.Name
}

type TableCreate struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type TableCreateResult struct {
	Table     document.Table `json:"table"`
	Workspace View           `json:"workspace"`
}

type ColumnAdd struct {
	Path              string `json:"path"`
	Name              string `json:"name"`
	SchemaExpression  string `json:"schemaExpression"`
	InitialExpression string `json:"initialExpression"`
	Position          *int   `json:"position,omitempty"`
}

type ColumnAddResult struct {
	Table     document.Table `json:"table"`
	Workspace View           `json:"workspace"`
}

type ColumnSchemaSet struct {
	Path             string `json:"path"`
	Name             string `json:"name"`
	SchemaExpression string `json:"schemaExpression"`
}

type ColumnSchemaSetResult struct {
	Table     document.Table `json:"table"`
	Workspace View           `json:"workspace"`
}

type ColumnDelete struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

type ColumnDeleteResult struct {
	Table     document.Table `json:"table"`
	Workspace View           `json:"workspace"`
}

type ColumnRename struct {
	Path    string `json:"path"`
	OldName string `json:"oldName"`
	NewName string `json:"newName"`
}

type ColumnRenameResult struct {
	Table     document.Table `json:"table"`
	Workspace View           `json:"workspace"`
}

type ColumnMove struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Position int    `json:"position"`
}

type ColumnMoveResult struct {
	Table     document.Table `json:"table"`
	Workspace View           `json:"workspace"`
}

type SchemaDefinitionCreate struct {
	Path             string   `json:"path"`
	Name             string   `json:"name"`
	Kind             string   `json:"kind"`
	Type             string   `json:"type,omitempty"`
	AppliesTo        []string `json:"appliesTo,omitempty"`
	SchemaExpression string   `json:"schemaExpression"`
}

type SchemaDefinitionSet struct {
	Path             string   `json:"path"`
	Name             string   `json:"name"`
	Kind             string   `json:"kind"`
	Type             string   `json:"type,omitempty"`
	AppliesTo        []string `json:"appliesTo,omitempty"`
	SchemaExpression string   `json:"schemaExpression"`
}

type SchemaDefinitionRename struct {
	Path    string `json:"path"`
	OldName string `json:"oldName"`
	NewName string `json:"newName"`
}

type SchemaDefinitionDelete struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

type SchemaEditResult struct {
	Schema    document.PublicSchema `json:"schema"`
	Table     document.Table        `json:"table"`
	Workspace View                  `json:"workspace"`
}

func NewService(configRoot string) *Service {
	return &Service{configRoot: configRoot, history: map[string]*history{}}
}

func (s *Service) SetLogger(logger *logging.Logger) {
	s.logger = logger
}

func (s *Service) log(level logging.Level, message string, fields map[string]any) {
	if s.logger != nil {
		s.logger.Log(level, message, fields)
	}
}

func NewUserService() (*Service, error) {
	configRoot, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user config directory: %w", err)
	}
	return NewService(configRoot), nil
}

func (s *Service) OpenProject(path string) (View, error) {
	started := time.Now()
	s.log(logging.Info, "open project started", map[string]any{"path": path})
	project, err := LoadProject(path)
	if err != nil {
		s.log(logging.Error, "open project failed", map[string]any{"path": path, "error": err.Error()})
		return View{}, err
	}
	settings, err := LoadEditorSettings(s.configRoot)
	if err != nil {
		s.log(logging.Error, "open project failed", map[string]any{"path": path, "error": err.Error()})
		return View{}, err
	}
	tree, err := projecttree.Scan(project.Path, project.Config.TableRoots, project.Config.SchemaRoots, project.Config.Exclude)
	if err != nil {
		s.log(logging.Error, "open project failed", map[string]any{"path": path, "error": err.Error()})
		return View{}, err
	}
	s.revision = 0
	s.compiledRevision = 0
	view := buildView(project, settings, tree)
	s.current = &view
	s.compile()
	view = *s.current
	s.log(logging.Info, "open project completed", map[string]any{"path": path, "schemas": len(view.Schemas.Definitions), "compile_diagnostics": len(view.Compile.Diagnostics), "duration_ms": time.Since(started).Milliseconds()})
	return view, nil
}

// compile runs one full compile against the current tree and publishes the
// result into the current view.
func (s *Service) compile() compiler.Result {
	started := time.Now()
	var publicSchemas []document.PublicSchema
	var tables []document.Table
	collectDocuments(s.current.Tree.Root, &publicSchemas, &tables)
	result := compiler.CompileAll(publicSchemas, tables)
	appendTreeDiagnostics(s.current.Tree.Root, &result)
	diagnostics := result.Diagnostics
	if diagnostics == nil {
		diagnostics = []compiler.Diagnostic{}
	}
	s.compileState = CompileState{
		Mode:        s.current.Settings.Build.Mode,
		CompiledAt:  time.Now().UTC().Format(time.RFC3339),
		Diagnostics: diagnostics,
	}
	s.compiledRevision = s.revision
	if s.current != nil {
		s.current.Compile = s.compileState
	}
	s.log(logging.Info, "compile completed", map[string]any{"diagnostics": len(result.Diagnostics), "mode": s.compileState.Mode, "duration_ms": time.Since(started).Milliseconds()})
	return result
}

func appendTreeDiagnostics(node projecttree.Node, result *compiler.Result) {
	if node.Kind == projecttree.NodeDamaged {
		for _, diagnostic := range node.Diagnostics {
			result.Diagnostics = append(result.Diagnostics, compiler.Diagnostic{Path: node.Path, Message: diagnostic.Message})
		}
	}
	for _, child := range node.Children {
		appendTreeDiagnostics(child, result)
	}
}

// publishView stamps the compile state onto a freshly built view.
func (s *Service) publishView(view *View) {
	view.Compile = s.compileState
	view.Compile.Mode = s.current.Settings.Build.Mode
	view.Compile.Pending = s.compiledRevision < s.revision
	if view.Compile.Diagnostics == nil {
		view.Compile.Diagnostics = []compiler.Diagnostic{}
	}
	s.current = view
}

func buildSchemaIndex(tree projecttree.Tree) schemaregistry.Index {
	var publicSchemas []document.PublicSchema
	var tables []document.Table
	collectDocuments(tree.Root, &publicSchemas, &tables)
	return schemaregistry.Build(publicSchemas, tables).Index()
}

func buildForeignKeyIndex(tree projecttree.Tree) foreignkey.Index {
	var tables []document.Table
	collectTables(tree.Root, &tables)
	return foreignkey.Build(tables)
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
		err := fmt.Errorf("no project is open")
		s.log(logging.Error, "get table failed", map[string]any{"path": path, "error": err.Error()})
		return document.Table{}, err
	}
	if !containsTable(s.current.Tree.Root, path) {
		err := fmt.Errorf("table %q is not part of the open project", path)
		s.log(logging.Error, "get table failed", map[string]any{"path": path, "error": err.Error()})
		return document.Table{}, err
	}
	table, err := document.LoadTable(path)
	if err != nil {
		err = fmt.Errorf("load table: %w", err)
		s.log(logging.Error, "get table failed", map[string]any{"path": path, "error": err.Error()})
		return document.Table{}, err
	}
	s.log(logging.Debug, "table loaded", map[string]any{"path": path, "rows": len(table.Rows), "columns": len(table.Columns)})
	return withTableMetadata(table, *s.current), nil
}

// GetPublicSchema loads a public Schema document that belongs to the current
// project.
func (s *Service) GetPublicSchema(path string) (document.PublicSchema, error) {
	if s.current == nil {
		err := fmt.Errorf("no project is open")
		s.log(logging.Error, "get public schema failed", map[string]any{"path": path, "error": err.Error()})
		return document.PublicSchema{}, err
	}
	if !containsSchemaFile(s.current.Tree.Root, path) {
		err := fmt.Errorf("schema file %q is not part of the open project", path)
		s.log(logging.Error, "get public schema failed", map[string]any{"path": path, "error": err.Error()})
		return document.PublicSchema{}, err
	}
	schema, err := document.LoadPublicSchema(path)
	if err != nil {
		s.log(logging.Error, "get public schema failed", map[string]any{"path": path, "error": err.Error()})
		return document.PublicSchema{}, err
	}
	s.log(logging.Debug, "public schema loaded", map[string]any{"path": path, "definitions": len(schema.Definitions)})
	return schema, nil
}

func (s *Service) schemaDefinitionScope(path string) (string, error) {
	if containsSchemaFile(s.current.Tree.Root, path) {
		return "public", nil
	}
	if containsTable(s.current.Tree.Root, path) {
		return "local", nil
	}
	return "", fmt.Errorf("path %q is not part of the open project", path)
}

func (s *Service) currentSchemaPackage(path string) string {
	if schema, err := document.LoadPublicSchema(path); err == nil {
		return schema.Package
	}
	return ""
}

func (s *Service) schemaDefinitionID(path, name string) (schemaregistry.ID, error) {
	if containsSchemaFile(s.current.Tree.Root, path) {
		schema, err := document.LoadPublicSchema(path)
		if err != nil {
			return "", err
		}
		return schemaregistry.ID(schema.Package + ".#" + name), nil
	}
	table, err := document.LoadTable(path)
	if err != nil {
		return "", err
	}
	logical := table.Name
	if table.Namespace != "" {
		logical = table.Namespace + "." + table.Name
	}
	return schemaregistry.ID(logical + "._local.#" + name), nil
}

// CreateSchemaDefinition adds one definition to a public Schema file or to a
// table's local Schema, then refreshes the workspace snapshot.
func (s *Service) CreateSchemaDefinition(create SchemaDefinitionCreate) (SchemaEditResult, error) {
	if s.current == nil {
		return SchemaEditResult{}, fmt.Errorf("no project is open")
	}
	update := document.SchemaDefinitionUpdate{Kind: create.Kind, Type: create.Type, AppliesTo: create.AppliesTo, SchemaExpression: create.SchemaExpression}
	scope, err := s.schemaDefinitionScope(create.Path)
	if err != nil {
		return SchemaEditResult{}, err
	}
	if scope == "local" {
		normalized, err := s.normalizeSchemaReferences(create.Path, create.SchemaExpression)
		if err != nil {
			return SchemaEditResult{}, err
		}
		update.SchemaExpression = normalized
		if err := s.ensureSchemaImports(create.Path, normalized, ""); err != nil {
			return SchemaEditResult{}, err
		}
	} else if err := s.ensureSchemaImports(create.Path, create.SchemaExpression, s.currentSchemaPackage(create.Path)); err != nil {
		return SchemaEditResult{}, err
	}
	var result SchemaEditResult
	if scope == "public" {
		schema, err := document.CreatePublicSchemaDefinition(create.Path, create.Name, update)
		if err != nil {
			s.log(logging.Error, "create schema definition failed", map[string]any{"path": create.Path, "name": create.Name, "error": err.Error()})
			return SchemaEditResult{}, err
		}
		result.Schema = schema
	} else {
		table, err := document.CreateLocalSchemaDefinition(create.Path, create.Name, update)
		if err != nil {
			s.log(logging.Error, "create local schema definition failed", map[string]any{"path": create.Path, "name": create.Name, "error": err.Error()})
			return SchemaEditResult{}, err
		}
		result.Table = table
	}
	view, err := s.refreshAfterEdit()
	if err != nil {
		return SchemaEditResult{}, err
	}
	if result.Table.Path != "" {
		result.Table = withTableMetadata(result.Table, view)
	}
	result.Workspace = view
	s.log(logging.Info, "schema definition created", map[string]any{"path": create.Path, "name": create.Name, "scope": scope})
	return result, nil
}

// SetSchemaDefinition replaces one definition expression and metadata.
func (s *Service) SetSchemaDefinition(set SchemaDefinitionSet) (SchemaEditResult, error) {
	if s.current == nil {
		return SchemaEditResult{}, fmt.Errorf("no project is open")
	}
	update := document.SchemaDefinitionUpdate{Kind: set.Kind, Type: set.Type, AppliesTo: set.AppliesTo, SchemaExpression: set.SchemaExpression}
	scope, err := s.schemaDefinitionScope(set.Path)
	if err != nil {
		return SchemaEditResult{}, err
	}
	if scope == "local" {
		normalized, err := s.normalizeSchemaReferences(set.Path, set.SchemaExpression)
		if err != nil {
			return SchemaEditResult{}, err
		}
		update.SchemaExpression = normalized
		if err := s.ensureSchemaImports(set.Path, normalized, ""); err != nil {
			return SchemaEditResult{}, err
		}
	} else if err := s.ensureSchemaImports(set.Path, set.SchemaExpression, s.currentSchemaPackage(set.Path)); err != nil {
		return SchemaEditResult{}, err
	}
	var result SchemaEditResult
	if scope == "public" {
		schema, err := document.SetPublicSchemaDefinition(set.Path, set.Name, update)
		if err != nil {
			s.log(logging.Error, "set schema definition failed", map[string]any{"path": set.Path, "name": set.Name, "error": err.Error()})
			return SchemaEditResult{}, err
		}
		result.Schema = schema
	} else {
		table, err := document.SetLocalSchemaDefinition(set.Path, set.Name, update)
		if err != nil {
			s.log(logging.Error, "set local schema definition failed", map[string]any{"path": set.Path, "name": set.Name, "error": err.Error()})
			return SchemaEditResult{}, err
		}
		result.Table = table
	}
	view, err := s.refreshAfterEdit()
	if err != nil {
		return SchemaEditResult{}, err
	}
	if result.Table.Path != "" {
		result.Table = withTableMetadata(result.Table, view)
	}
	result.Workspace = view
	s.log(logging.Info, "schema definition updated", map[string]any{"path": set.Path, "name": set.Name, "scope": scope})
	return result, nil
}

// RenameSchemaDefinition renames a definition and rewrites every Schema
// reference across all project source files.
func (s *Service) RenameSchemaDefinition(rename SchemaDefinitionRename) (SchemaEditResult, error) {
	if s.current == nil {
		return SchemaEditResult{}, fmt.Errorf("no project is open")
	}
	scope, err := s.schemaDefinitionScope(rename.Path)
	if err != nil {
		return SchemaEditResult{}, err
	}
	var paths []string
	collectSourcePaths(s.current.Tree.Root, &paths)
	var result SchemaEditResult
	if scope == "public" {
		schema, err := document.RenamePublicSchemaDefinition(rename.Path, rename.OldName, rename.NewName, paths)
		if err != nil {
			s.log(logging.Error, "rename schema definition failed", map[string]any{"path": rename.Path, "name": rename.OldName, "error": err.Error()})
			return SchemaEditResult{}, err
		}
		result.Schema = schema
	} else {
		table, err := document.RenameLocalSchemaDefinition(rename.Path, rename.OldName, rename.NewName, paths)
		if err != nil {
			s.log(logging.Error, "rename local schema definition failed", map[string]any{"path": rename.Path, "name": rename.OldName, "error": err.Error()})
			return SchemaEditResult{}, err
		}
		result.Table = table
	}
	view, err := s.refreshAfterEdit()
	if err != nil {
		return SchemaEditResult{}, err
	}
	if result.Table.Path != "" {
		result.Table = withTableMetadata(result.Table, view)
	}
	result.Workspace = view
	s.log(logging.Info, "schema definition renamed", map[string]any{"path": rename.Path, "old": rename.OldName, "new": rename.NewName})
	return result, nil
}

// DeleteSchemaDefinition removes a definition unless it is still referenced.
func (s *Service) DeleteSchemaDefinition(delete SchemaDefinitionDelete) (SchemaEditResult, error) {
	if s.current == nil {
		return SchemaEditResult{}, fmt.Errorf("no project is open")
	}
	scope, err := s.schemaDefinitionScope(delete.Path)
	if err != nil {
		return SchemaEditResult{}, err
	}
	id, err := s.schemaDefinitionID(delete.Path, delete.Name)
	if err != nil {
		return SchemaEditResult{}, err
	}
	for _, definition := range s.current.Schemas.Definitions {
		if definition.ID != id || len(definition.Usages) == 0 {
			continue
		}
		usage := definition.Usages[0]
		err := fmt.Errorf("Schema %s is referenced by %s field %s", id, usage.Path, usage.Field)
		s.log(logging.Warn, "schema definition delete blocked", map[string]any{"path": delete.Path, "name": delete.Name, "error": err.Error()})
		return SchemaEditResult{}, err
	}
	var result SchemaEditResult
	if scope == "public" {
		schema, err := document.DeletePublicSchemaDefinition(delete.Path, delete.Name)
		if err != nil {
			s.log(logging.Error, "delete schema definition failed", map[string]any{"path": delete.Path, "name": delete.Name, "error": err.Error()})
			return SchemaEditResult{}, err
		}
		result.Schema = schema
	} else {
		table, err := document.DeleteLocalSchemaDefinition(delete.Path, delete.Name)
		if err != nil {
			s.log(logging.Error, "delete local schema definition failed", map[string]any{"path": delete.Path, "name": delete.Name, "error": err.Error()})
			return SchemaEditResult{}, err
		}
		result.Table = table
	}
	view, err := s.refreshAfterEdit()
	if err != nil {
		return SchemaEditResult{}, err
	}
	if result.Table.Path != "" {
		result.Table = withTableMetadata(result.Table, view)
	}
	result.Workspace = view
	s.log(logging.Info, "schema definition deleted", map[string]any{"path": delete.Path, "name": delete.Name, "scope": scope})
	return result, nil
}

// normalizeSchemaReferences resolves bare #Name references in a table Schema
// expression into _local.#Name or package.#Name forms that CUE can resolve
// from the table file scope. Local definitions win over public definitions.
func (s *Service) normalizeSchemaReferences(path, expression string) (string, error) {
	localNames := map[string]bool{}
	if table, err := document.LoadTable(path); err == nil {
		for _, definition := range table.LocalSchema {
			localNames[definition.Name] = true
		}
	}
	publicPackage := map[string]string{}
	ambiguous := map[string]bool{}
	for _, definition := range s.current.Schemas.Definitions {
		if definition.Table != "" {
			continue
		}
		pkg, name, ok := splitDefinitionID(string(definition.ID))
		if !ok {
			continue
		}
		if existing, exists := publicPackage[name]; exists && existing != pkg {
			delete(publicPackage, name)
			ambiguous[name] = true
			continue
		}
		if !ambiguous[name] {
			publicPackage[name] = pkg
		}
	}
	parsed, err := parser.ParseExpr(path, expression)
	if err != nil {
		return "", fmt.Errorf("parse Schema expression: %w", err)
	}
	rewritten, err := rewriteSchemaBareReferences(parsed, func(name string) (ast.Expr, error) {
		short := strings.TrimPrefix(name, "#")
		if localNames[short] {
			return &ast.SelectorExpr{X: ast.NewIdent("_local"), Sel: ast.NewIdent(name)}, nil
		}
		if pkg, exists := publicPackage[short]; exists {
			return &ast.SelectorExpr{X: ast.NewIdent(pkg), Sel: ast.NewIdent(name)}, nil
		}
		if ambiguous[short] {
			return nil, fmt.Errorf("Schema reference %s is ambiguous; qualify it with a package name", name)
		}
		return nil, fmt.Errorf("Schema reference %s not found; local references use _local.%s and public references use package.%s", name, name, name)
	})
	if err != nil {
		return "", err
	}
	formatted, err := format.Node(rewritten)
	if err != nil {
		return "", fmt.Errorf("format Schema expression: %w", err)
	}
	return string(formatted), nil
}

func splitDefinitionID(id string) (string, string, bool) {
	marker := strings.Index(id, ".#")
	if marker < 0 || marker == len(id)-2 {
		return "", "", false
	}
	return id[:marker], id[marker+2:], true
}

func rewriteSchemaBareReferences(expression ast.Expr, resolve func(string) (ast.Expr, error)) (ast.Expr, error) {
	switch value := expression.(type) {
	case *ast.ParenExpr:
		rewritten, err := rewriteSchemaBareReferences(value.X, resolve)
		if err != nil {
			return nil, err
		}
		value.X = rewritten
		return value, nil
	case *ast.BinaryExpr:
		left, err := rewriteSchemaBareReferences(value.X, resolve)
		if err != nil {
			return nil, err
		}
		right, err := rewriteSchemaBareReferences(value.Y, resolve)
		if err != nil {
			return nil, err
		}
		value.X, value.Y = left, right
		return value, nil
	case *ast.UnaryExpr:
		rewritten, err := rewriteSchemaBareReferences(value.X, resolve)
		if err != nil {
			return nil, err
		}
		value.X = rewritten
		return value, nil
	case *ast.SelectorExpr:
		rewritten, err := rewriteSchemaBareReferences(value.X, resolve)
		if err != nil {
			return nil, err
		}
		value.X = rewritten
		return value, nil
	case *ast.Ident:
		if strings.HasPrefix(value.Name, "#") {
			return resolve(value.Name)
		}
		return value, nil
	case *ast.StructLit:
		for index, declaration := range value.Elts {
			field, ok := declaration.(*ast.Field)
			if !ok {
				continue
			}
			rewritten, err := rewriteSchemaBareReferences(field.Value, resolve)
			if err != nil {
				return nil, err
			}
			field.Value = rewritten
			value.Elts[index] = field
		}
		return value, nil
	case *ast.ListLit:
		for index, element := range value.Elts {
			rewritten, err := rewriteSchemaBareReferences(element, resolve)
			if err != nil {
				return nil, err
			}
			value.Elts[index] = rewritten
		}
		return value, nil
	default:
		return value, nil
	}
}

// referencedSchemaPackages collects the package names referenced by qualified
// references in a Schema expression.
func referencedSchemaPackages(expression string) ([]string, error) {
	parsed, err := parser.ParseExpr("schema-reference.cue", expression)
	if err != nil {
		return nil, fmt.Errorf("parse Schema expression: %w", err)
	}
	packages := map[string]bool{}
	ast.Walk(parsed, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if identifier, ok := selector.X.(*ast.Ident); ok && identifier.Name != "_local" {
			packages[identifier.Name] = true
		}
		return true
	}, nil)
	result := make([]string, 0, len(packages))
	for pkg := range packages {
		result = append(result, pkg)
	}
	sort.Strings(result)
	return result, nil
}

// ensureSchemaImports makes sure the source file imports every public Schema
// package referenced by the expression.
func (s *Service) ensureSchemaImports(path, expression, excludePackage string) error {
	packages, err := referencedSchemaPackages(expression)
	if err != nil {
		return err
	}
	for _, pkg := range packages {
		if pkg == excludePackage || !s.knownSchemaPackage(pkg) {
			continue
		}
		importPath, err := s.schemaImportPath(pkg)
		if err != nil {
			return err
		}
		if err := document.EnsureImport(path, importPath); err != nil {
			return fmt.Errorf("import schema package %q: %w", pkg, err)
		}
	}
	return nil
}

func (s *Service) knownSchemaPackage(pkg string) bool {
	for _, definition := range s.current.Schemas.Definitions {
		if definition.Table != "" {
			continue
		}
		if packageName, _, ok := splitDefinitionID(string(definition.ID)); ok && packageName == pkg {
			return true
		}
	}
	return false
}

// schemaImportPath derives the CUE import path for one public Schema package.
func (s *Service) schemaImportPath(packageName string) (string, error) {
	directory := ""
	for _, definition := range s.current.Schemas.Definitions {
		if definition.Table != "" {
			continue
		}
		if pkg, _, ok := splitDefinitionID(string(definition.ID)); ok && pkg == packageName {
			directory = filepath.Dir(definition.Path)
			break
		}
	}
	if directory == "" {
		return "", fmt.Errorf("schema package %q is not part of the open project", packageName)
	}
	relative, err := filepath.Rel(s.current.Project.Root, directory)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("schema package %q is outside the project directory", packageName)
	}
	module, err := s.moduleName()
	if err != nil {
		return "", err
	}
	return module + "/" + filepath.ToSlash(relative) + ":" + packageName, nil
}

func (s *Service) moduleName() (string, error) {
	path := filepath.Join(s.current.Project.Root, "cue.mod", "module.cue")
	source, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot resolve CUE module import: read %s: %w", path, err)
	}
	file, err := parser.ParseFile(path, source, parser.ParseComments, parser.AllErrors)
	if err != nil {
		return "", fmt.Errorf("cannot resolve CUE module import: %w", err)
	}
	for _, declaration := range file.Decls {
		field, ok := declaration.(*ast.Field)
		if !ok {
			continue
		}
		name, _, err := ast.LabelName(field.Label)
		if err != nil || name != "module" {
			continue
		}
		literal, ok := field.Value.(*ast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			continue
		}
		if value, err := strconv.Unquote(literal.Value); err == nil && value != "" {
			return value, nil
		}
	}
	return "", fmt.Errorf("cannot resolve CUE module import: cue.mod/module.cue has no module declaration")
}

func (s *Service) CreateTable(create TableCreate) (TableCreateResult, error) {
	if s.current == nil {
		return TableCreateResult{}, fmt.Errorf("no project is open")
	}
	if !validTableName(create.Name) || !validNamespace(create.Namespace) {
		return TableCreateResult{}, fmt.Errorf("table name and namespace must use lowercase letters, digits, and underscores")
	}
	var tables []document.Table
	collectTables(s.current.Tree.Root, &tables)
	for _, table := range tables {
		if table.Namespace == create.Namespace && table.Name == create.Name {
			return TableCreateResult{}, fmt.Errorf("table %q already exists", create.Namespace+"."+create.Name)
		}
	}
	directory := filepath.Join(s.current.Project.Root, s.current.Project.Config.TableRoots[0], create.Name)
	path := filepath.Join(directory, create.Name+".cue")
	if err := os.Mkdir(directory, 0o755); err != nil {
		if os.IsExist(err) {
			return TableCreateResult{}, fmt.Errorf("table directory %q already exists", create.Name)
		}
		return TableCreateResult{}, fmt.Errorf("create table directory: %w", err)
	}
	source := fmt.Sprintf(`package %s

#PrimaryKey: string

#Row: {}

rows: {}

_tool: {
	version: 1
	table: {
		namespace: %q
		name: %q
	}
	foreignKeys: {}
	formulas: {}
}

_editor: {columns: {}}
`, create.Name, create.Namespace, create.Name)
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		_ = os.Remove(directory)
		return TableCreateResult{}, fmt.Errorf("write table source: %w", err)
	}
	table, err := document.LoadTable(path)
	if err != nil {
		_ = os.Remove(path)
		_ = os.Remove(directory)
		return TableCreateResult{}, fmt.Errorf("validate created table: %w", err)
	}
	view, err := s.refreshAfterEdit()
	if err != nil {
		return TableCreateResult{}, err
	}
	return TableCreateResult{Table: withTableMetadata(table, view), Workspace: view}, nil
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
	table, view, err := s.edit(update.Path, func() (document.Table, error) {
		return document.SetCellValue(update.Path, update.RowID, update.Column, update.Expression)
	})
	if err != nil {
		return CellUpdateResult{}, err
	}
	return CellUpdateResult{Table: withTableMetadata(table, view), Workspace: view}, nil
}

func (s *Service) Paste(paste Paste) (HistoryResult, error) {
	if s.current == nil {
		return HistoryResult{}, fmt.Errorf("no project is open")
	}
	if !containsTable(s.current.Tree.Root, paste.Path) {
		return HistoryResult{}, fmt.Errorf("table %q is not part of the open project", paste.Path)
	}
	table, view, err := s.edit(paste.Path, func() (document.Table, error) {
		return document.PasteCells(paste.Path, paste.RowIDs, paste.Columns, paste.Values)
	})
	if err != nil {
		return HistoryResult{}, err
	}
	return s.historyResult(paste.Path, table, view), nil
}

func (s *Service) Undo(path string) (HistoryResult, error) {
	return s.restoreHistory(path, true)
}

func (s *Service) Redo(path string) (HistoryResult, error) {
	return s.restoreHistory(path, false)
}

func (s *Service) SetCellFormula(set FormulaSet) (FormulaSetResult, error) {
	if s.current == nil {
		return FormulaSetResult{}, fmt.Errorf("no project is open")
	}
	if !containsTable(s.current.Tree.Root, set.Path) {
		return FormulaSetResult{}, fmt.Errorf("table %q is not part of the open project", set.Path)
	}
	table, view, err := s.edit(set.Path, func() (document.Table, error) {
		return document.SetCellFormula(set.Path, set.RowID, set.Column, set.Expression)
	})
	if err != nil {
		return FormulaSetResult{}, err
	}
	return FormulaSetResult{Table: withTableMetadata(table, view), Workspace: view}, nil
}

func (s *Service) edit(path string, operation func() (document.Table, error)) (document.Table, View, error) {
	started := time.Now()
	before, err := os.ReadFile(path)
	if err != nil {
		s.log(logging.Error, "table edit failed", map[string]any{"path": path, "error": err.Error()})
		return document.Table{}, View{}, fmt.Errorf("record edit source: %w", err)
	}
	table, err := operation()
	if err != nil {
		s.log(logging.Error, "table edit failed", map[string]any{"path": path, "error": err.Error()})
		return document.Table{}, View{}, err
	}
	view, err := s.refreshAfterEdit()
	if err != nil {
		s.log(logging.Error, "workspace refresh failed after edit", map[string]any{"path": path, "error": err.Error()})
		return document.Table{}, View{}, err
	}
	record := s.history[path]
	if record == nil {
		record = &history{}
		s.history[path] = record
	}
	record.undo = append(record.undo, before)
	record.redo = nil
	s.log(logging.Info, "table edit completed", map[string]any{"path": path, "duration_ms": time.Since(started).Milliseconds()})
	return table, view, nil
}

func (s *Service) restoreHistory(path string, undo bool) (HistoryResult, error) {
	if s.current == nil {
		return HistoryResult{}, fmt.Errorf("no project is open")
	}
	if !containsTable(s.current.Tree.Root, path) {
		return HistoryResult{}, fmt.Errorf("table %q is not part of the open project", path)
	}
	record := s.history[path]
	if record == nil || (undo && len(record.undo) == 0) || (!undo && len(record.redo) == 0) {
		return HistoryResult{}, fmt.Errorf("nothing to %s", map[bool]string{true: "undo", false: "redo"}[undo])
	}
	current, err := os.ReadFile(path)
	if err != nil {
		return HistoryResult{}, fmt.Errorf("read history source: %w", err)
	}
	var target []byte
	if undo {
		target = record.undo[len(record.undo)-1]
	} else {
		target = record.redo[len(record.redo)-1]
	}
	table, err := document.RestoreSource(path, target)
	if err != nil {
		return HistoryResult{}, err
	}
	view, err := s.refreshAfterEdit()
	if err != nil {
		return HistoryResult{}, err
	}
	if undo {
		record.undo = record.undo[:len(record.undo)-1]
		record.redo = append(record.redo, current)
	} else {
		record.redo = record.redo[:len(record.redo)-1]
		record.undo = append(record.undo, current)
	}
	return s.historyResult(path, table, view), nil
}

func (s *Service) historyResult(path string, table document.Table, view View) HistoryResult {
	record := s.history[path]
	return HistoryResult{Table: withTableMetadata(table, view), Workspace: view, CanUndo: record != nil && len(record.undo) > 0, CanRedo: record != nil && len(record.redo) > 0}
}

func (s *Service) CreateRow(create RowCreate) (RowCreateResult, error) {
	if s.current == nil {
		return RowCreateResult{}, fmt.Errorf("no project is open")
	}
	if !containsTable(s.current.Tree.Root, create.Path) {
		return RowCreateResult{}, fmt.Errorf("table %q is not part of the open project", create.Path)
	}
	table, view, err := s.edit(create.Path, func() (document.Table, error) {
		return document.CreateRow(create.Path, create.RowID, create.Values, create.Formulas)
	})
	if err != nil {
		return RowCreateResult{}, err
	}
	return RowCreateResult{Table: withTableMetadata(table, view), Workspace: view}, nil
}

func (s *Service) ClearCell(update CellUpdate) (CellUpdateResult, error) {
	if s.current == nil {
		return CellUpdateResult{}, fmt.Errorf("no project is open")
	}
	if !containsTable(s.current.Tree.Root, update.Path) {
		return CellUpdateResult{}, fmt.Errorf("table %q is not part of the open project", update.Path)
	}
	table, view, err := s.edit(update.Path, func() (document.Table, error) {
		return document.ClearCell(update.Path, update.RowID, update.Column)
	})
	if err != nil {
		return CellUpdateResult{}, err
	}
	return CellUpdateResult{Table: withTableMetadata(table, view), Workspace: view}, nil
}

func (s *Service) AddRow(add RowAdd) (RowAddResult, error) {
	if s.current == nil {
		return RowAddResult{}, fmt.Errorf("no project is open")
	}
	if !containsTable(s.current.Tree.Root, add.Path) {
		return RowAddResult{}, fmt.Errorf("table %q is not part of the open project", add.Path)
	}
	table, view, err := s.edit(add.Path, func() (document.Table, error) {
		return document.AddRow(add.Path, add.RowID, add.Expression)
	})
	if err != nil {
		return RowAddResult{}, err
	}
	return RowAddResult{Table: table, Workspace: view}, nil
}

func (s *Service) DeleteRow(delete RowDelete) (RowDeleteResult, error) {
	if s.current == nil {
		return RowDeleteResult{}, fmt.Errorf("no project is open")
	}
	if !containsTable(s.current.Tree.Root, delete.Path) {
		return RowDeleteResult{}, fmt.Errorf("table %q is not part of the open project", delete.Path)
	}
	table, err := document.LoadTable(delete.Path)
	if err != nil {
		return RowDeleteResult{}, fmt.Errorf("load table: %w", err)
	}
	if usage := foreignKeyUsage(s.current.Tree, table, delete.RowID); usage != nil {
		return RowDeleteResult{}, fmt.Errorf("row %q is referenced by %s.%s.%s", delete.RowID, usage.Table, usage.RowID, usage.Column)
	}
	table, view, err := s.edit(delete.Path, func() (document.Table, error) {
		return document.DeleteRow(delete.Path, delete.RowID)
	})
	if err != nil {
		return RowDeleteResult{}, err
	}
	return RowDeleteResult{Table: table, Workspace: view}, nil
}

func (s *Service) RenameRow(rename RowRename) (RowRenameResult, error) {
	if s.current == nil {
		return RowRenameResult{}, fmt.Errorf("no project is open")
	}
	if !containsTable(s.current.Tree.Root, rename.Path) {
		return RowRenameResult{}, fmt.Errorf("table %q is not part of the open project", rename.Path)
	}
	target, err := document.LoadTable(rename.Path)
	if err != nil {
		return RowRenameResult{}, fmt.Errorf("load table: %w", err)
	}
	var tables []document.Table
	collectTables(s.current.Tree.Root, &tables)
	targetName := target.Name
	if target.Namespace != "" {
		targetName = target.Namespace + "." + target.Name
	}
	usages := foreignkey.Build(tables).Usages[foreignkey.Target{Table: targetName, RowID: rename.OldID}]
	references := make([]document.RowReference, 0, len(usages))
	for _, usage := range usages {
		references = append(references, document.RowReference{Path: usage.Path, RowID: usage.RowID, Column: usage.Column, Index: usage.Index, Array: usage.Array})
	}
	table, err := document.RenameRow(rename.Path, rename.OldID, rename.NewID, references)
	if err != nil {
		return RowRenameResult{}, err
	}
	view, err := s.refreshAfterEdit()
	if err != nil {
		return RowRenameResult{}, err
	}
	return RowRenameResult{Table: table, Workspace: view}, nil
}

// Export writes configured CSV files only when the current workspace has no
// Schema or foreign-key diagnostics.
// Export writes configured CSV files only when a full compile is clean. The
// exporter receives the compiled package and never re-validates data.
func (s *Service) Export() (ExportResult, error) {
	if s.current == nil {
		err := fmt.Errorf("no project is open")
		s.log(logging.Error, "export failed", map[string]any{"error": err.Error()})
		return ExportResult{}, err
	}
	compiled := s.compile()
	if len(compiled.Diagnostics) > 0 {
		err := fmt.Errorf("export blocked by %d diagnostics: %s", len(compiled.Diagnostics), compiled.Diagnostics[0].Message)
		s.log(logging.Warn, "export blocked", map[string]any{"error": err.Error()})
		return ExportResult{}, err
	}
	destination, err := exportDirectory(s.current.Project)
	if err != nil {
		s.log(logging.Error, "export failed", map[string]any{"error": err.Error()})
		return ExportResult{}, err
	}
	result, err := csv.Export(destination, compiled.Package)
	if err != nil {
		s.log(logging.Error, "export failed", map[string]any{"destination": destination, "error": err.Error()})
		return ExportResult{}, err
	}
	s.log(logging.Info, "export completed", map[string]any{"files": len(result.Files), "destination": destination})
	return ExportResult{Files: result.Files}, nil
}

func (s *Service) AddColumn(add ColumnAdd) (ColumnAddResult, error) {
	if s.current == nil {
		return ColumnAddResult{}, fmt.Errorf("no project is open")
	}
	if !containsTable(s.current.Tree.Root, add.Path) {
		return ColumnAddResult{}, fmt.Errorf("table %q is not part of the open project", add.Path)
	}
	position := -1
	if add.Position != nil {
		position = *add.Position
	}
	expression, err := s.normalizeSchemaReferences(add.Path, add.SchemaExpression)
	if err != nil {
		return ColumnAddResult{}, err
	}
	if err := s.ensureSchemaImports(add.Path, expression, ""); err != nil {
		return ColumnAddResult{}, err
	}
	table, view, err := s.edit(add.Path, func() (document.Table, error) {
		return document.AddColumn(add.Path, add.Name, expression, add.InitialExpression, position)
	})
	if err != nil {
		return ColumnAddResult{}, err
	}
	return ColumnAddResult{Table: table, Workspace: view}, nil
}

func (s *Service) SetColumnSchema(set ColumnSchemaSet) (ColumnSchemaSetResult, error) {
	if s.current == nil {
		return ColumnSchemaSetResult{}, fmt.Errorf("no project is open")
	}
	if !containsTable(s.current.Tree.Root, set.Path) {
		return ColumnSchemaSetResult{}, fmt.Errorf("table %q is not part of the open project", set.Path)
	}
	expression, err := s.normalizeSchemaReferences(set.Path, set.SchemaExpression)
	if err != nil {
		return ColumnSchemaSetResult{}, err
	}
	if err := s.ensureSchemaImports(set.Path, expression, ""); err != nil {
		return ColumnSchemaSetResult{}, err
	}
	table, view, err := s.edit(set.Path, func() (document.Table, error) {
		return document.SetColumnSchema(set.Path, set.Name, expression)
	})
	if err != nil {
		return ColumnSchemaSetResult{}, err
	}
	return ColumnSchemaSetResult{Table: withTableMetadata(table, view), Workspace: view}, nil
}

func (s *Service) DeleteColumn(delete ColumnDelete) (ColumnDeleteResult, error) {
	if s.current == nil {
		return ColumnDeleteResult{}, fmt.Errorf("no project is open")
	}
	if !containsTable(s.current.Tree.Root, delete.Path) {
		return ColumnDeleteResult{}, fmt.Errorf("table %q is not part of the open project", delete.Path)
	}
	table, view, err := s.edit(delete.Path, func() (document.Table, error) {
		return document.DeleteColumn(delete.Path, delete.Name)
	})
	if err != nil {
		return ColumnDeleteResult{}, err
	}
	return ColumnDeleteResult{Table: table, Workspace: view}, nil
}

func (s *Service) RenameColumn(rename ColumnRename) (ColumnRenameResult, error) {
	if s.current == nil {
		return ColumnRenameResult{}, fmt.Errorf("no project is open")
	}
	if !containsTable(s.current.Tree.Root, rename.Path) {
		return ColumnRenameResult{}, fmt.Errorf("table %q is not part of the open project", rename.Path)
	}
	table, view, err := s.edit(rename.Path, func() (document.Table, error) {
		return document.RenameColumn(rename.Path, rename.OldName, rename.NewName)
	})
	if err != nil {
		return ColumnRenameResult{}, err
	}
	return ColumnRenameResult{Table: table, Workspace: view}, nil
}

func (s *Service) MoveColumn(move ColumnMove) (ColumnMoveResult, error) {
	if s.current == nil {
		return ColumnMoveResult{}, fmt.Errorf("no project is open")
	}
	if !containsTable(s.current.Tree.Root, move.Path) {
		return ColumnMoveResult{}, fmt.Errorf("table %q is not part of the open project", move.Path)
	}
	table, view, err := s.edit(move.Path, func() (document.Table, error) {
		return document.MoveColumn(move.Path, move.Name, move.Position)
	})
	if err != nil {
		return ColumnMoveResult{}, err
	}
	return ColumnMoveResult{Table: table, Workspace: view}, nil
}

func exportDirectory(project Project) (string, error) {
	destination := filepath.Clean(filepath.Join(project.Root, project.Config.Export.Directory))
	relative, err := filepath.Rel(project.Root, destination)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("export directory %q escapes the project directory", project.Config.Export.Directory)
	}
	return destination, nil
}

func foreignKeyUsage(tree projecttree.Tree, target document.Table, rowID string) *foreignkey.Usage {
	var tables []document.Table
	collectTables(tree.Root, &tables)
	targetName := target.Name
	if target.Namespace != "" {
		targetName = target.Namespace + "." + target.Name
	}
	usages := foreignkey.Build(tables).Usages[foreignkey.Target{Table: targetName, RowID: rowID}]
	if len(usages) == 0 {
		return nil
	}
	return &usages[0]
}

func collectTables(node projecttree.Node, tables *[]document.Table) {
	if node.Kind == projecttree.NodeTable {
		if table, err := document.LoadTable(node.Path); err == nil {
			*tables = append(*tables, table)
		}
	}
	for _, child := range node.Children {
		collectTables(child, tables)
	}
}

func (s *Service) refreshCurrentView() (View, error) {
	started := time.Now()
	project := s.current.Project
	tree, err := projecttree.Scan(project.Path, project.Config.TableRoots, project.Config.SchemaRoots, project.Config.Exclude)
	if err != nil {
		s.log(logging.Error, "workspace refresh failed", map[string]any{"error": err.Error()})
		return View{}, fmt.Errorf("refresh project tree: %w", err)
	}
	s.revision++
	view := buildView(project, s.current.Settings, tree)
	s.publishView(&view)
	s.log(logging.Debug, "workspace refresh completed", map[string]any{"schemas": len(view.Schemas.Definitions), "revision": s.revision, "duration_ms": time.Since(started).Milliseconds()})
	return view, nil
}

// refreshAfterEdit refreshes the workspace snapshot and, in auto mode,
// immediately compiles the edited sources.
func (s *Service) refreshAfterEdit() (View, error) {
	view, err := s.refreshCurrentView()
	if err != nil {
		return View{}, err
	}
	if s.current.Settings.Build.Mode == "auto" {
		s.compile()
		view = *s.current
	}
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

func containsSchemaFile(node projecttree.Node, path string) bool {
	if node.Kind == projecttree.NodeSchemaFile && node.Path == path {
		return true
	}
	for _, child := range node.Children {
		if containsSchemaFile(child, path) {
			return true
		}
	}
	return false
}

func collectSourcePaths(node projecttree.Node, paths *[]string) {
	if (node.Kind == projecttree.NodeTable || node.Kind == projecttree.NodeSchemaFile) && node.Path != "" {
		*paths = append(*paths, node.Path)
	}
	for _, child := range node.Children {
		collectSourcePaths(child, paths)
	}
}

func validTableName(value string) bool {
	if value == "" {
		return false
	}
	for index, character := range value {
		if !(character >= 'a' && character <= 'z' || character >= '0' && character <= '9' && index > 0 || character == '_' && index > 0) {
			return false
		}
	}
	return true
}

func validNamespace(value string) bool {
	if value == "" {
		return false
	}
	for _, part := range strings.Split(value, ".") {
		if !validTableName(part) {
			return false
		}
	}
	return true
}

func withDiagnostics(table document.Table, view View) document.Table {
	evaluated := formula.Evaluate(table)
	if view.Compile.Mode == "manual" {
		for _, diagnostic := range view.Compile.Diagnostics {
			if diagnostic.Path == table.Path {
				table.Diagnostics = append(table.Diagnostics, document.Diagnostic{RowID: diagnostic.RowID, Column: diagnostic.Column, Message: diagnostic.Message})
			}
		}
	} else {
		for _, diagnostic := range evaluated.Diagnostics {
			table.Diagnostics = append(table.Diagnostics, document.Diagnostic{RowID: diagnostic.RowID, Column: diagnostic.Column, Message: diagnostic.Message})
		}
		appendCueDiagnostics(&table, evaluated.Rows)
		for _, diagnostic := range view.ForeignKeys.Diagnostics {
			if diagnostic.Path == table.Path {
				table.Diagnostics = append(table.Diagnostics, document.Diagnostic{RowID: diagnostic.RowID, Column: diagnostic.Column, Message: diagnostic.Message})
			}
		}
	}
	for _, formula := range table.Formulas {
		for index := range table.Rows {
			if table.Rows[index].ID != formula.RowID {
				continue
			}
			value, ok := evaluated.Rows[formula.RowID][formula.Column]
			display := "<formula error>"
			if ok {
				display = fmt.Sprint(value)
			}
			table.Rows[index].Values[formula.Column] = document.Cell{Expression: formula.Expression, Display: display, Formula: true}
			break
		}
	}
	return table
}

func withTableMetadata(table document.Table, view View) document.Table {
	return withForeignKeyOptions(withDiagnostics(table, view), view.Tree)
}

// withForeignKeyOptions exposes target IDs only for scalar foreign keys. Array
// keys keep CUE expression editing until their dedicated list editor exists.
func withForeignKeyOptions(table document.Table, tree projecttree.Tree) document.Table {
	targets := map[string][]string{}
	var tables []document.Table
	collectTables(tree.Root, &tables)
	for _, candidate := range tables {
		name := candidate.Name
		if candidate.Namespace != "" {
			name = candidate.Namespace + "." + candidate.Name
		}
		for _, row := range candidate.Rows {
			targets[name] = append(targets[name], row.ID)
		}
	}
	for index := range table.Columns {
		for _, key := range table.ForeignKeys {
			if key.Column == table.Columns[index].Name && table.Columns[index].Schema.Kind != "array" {
				table.Columns[index].ForeignKeyTable = key.Table
				table.Columns[index].ForeignKeyOptions = targets[key.Table]
				break
			}
		}
	}
	return table
}

func appendCueDiagnostics(table *document.Table, rows map[string]map[string]any) {
	instances := load.Instances([]string{"."}, &load.Config{Dir: filepath.Dir(table.Path)})
	if len(instances) != 1 {
		table.Diagnostics = append(table.Diagnostics, document.Diagnostic{Message: "load table package: expected one package"})
		return
	}
	context := cuecontext.New()
	packageValue := context.BuildInstance(instances[0])
	if err := packageValue.Err(); err != nil {
		table.Diagnostics = append(table.Diagnostics, document.Diagnostic{Message: err.Error()})
		return
	}
	rowSchema := packageValue.LookupPath(cue.MakePath(cue.Def("Row")))
	primaryKey := packageValue.LookupPath(cue.MakePath(cue.Def("PrimaryKey")))
	for _, row := range table.Rows {
		if err := primaryKey.Unify(context.Encode(row.ID)).Validate(cue.Concrete(true)); err != nil {
			table.Diagnostics = append(table.Diagnostics, document.Diagnostic{RowID: row.ID, Message: "invalid ID: " + err.Error()})
		}
		for column, value := range rows[row.ID] {
			fieldSchema := rowSchema.LookupPath(cue.MakePath(cue.Str(column)))
			if err := fieldSchema.Unify(context.Encode(value)).Validate(cue.Concrete(true)); err != nil {
				table.Diagnostics = append(table.Diagnostics, document.Diagnostic{RowID: row.ID, Column: column, Message: err.Error()})
			}
		}
		candidate := rowSchema.Unify(context.Encode(rows[row.ID]))
		if err := candidate.Validate(cue.Concrete(true)); err != nil {
			table.Diagnostics = append(table.Diagnostics, document.Diagnostic{RowID: row.ID, Message: err.Error()})
		}
	}
}
