package workspace

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
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
	"datapack/internal/exporters"
	"datapack/internal/foreignkey"
	"datapack/internal/formula"
	"datapack/internal/i18n"
	"datapack/internal/logging"
	"datapack/internal/navigation"
	"datapack/internal/projecttree"
	"datapack/internal/schemamodel"
	"datapack/internal/schemaregistry"
)

type View struct {
	Project     Project              `json:"project"`
	Settings    EditorSettings       `json:"settings"`
	Tree        projecttree.Tree     `json:"tree"`
	Schemas     schemaregistry.Index `json:"schemas"`
	ForeignKeys foreignkey.Summary   `json:"foreignKeys"`
	Compile     CompileState         `json:"compile"`
	CanUndo     bool                 `json:"canUndo"`
	CanRedo     bool                 `json:"canRedo"`
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
	mu               sync.RWMutex
	configRoot       string
	current          *View
	undoStack        []map[string][]byte
	redoStack        []map[string][]byte
	logger           *logging.Logger
	revision         int
	compiledRevision int
	compileState     CompileState
	settings         EditorSettings
	settingsLoaded   bool
	exporters        exporters.Registry
}

type CellUpdate struct {
	Path       string `json:"path"`
	RowID      string `json:"rowID"`
	Column     string `json:"column"`
	Expression string `json:"expression"`
}

type CellCommentSet struct {
	Path    string `json:"path"`
	RowID   string `json:"rowID"`
	Column  string `json:"column"`
	Comment string `json:"comment"`
}

type ColumnDisplayNameSet struct {
	Path        string `json:"path"`
	Column      string `json:"column"`
	DisplayName string `json:"displayName"`
}

type ColumnWidthSet struct {
	Path   string `json:"path"`
	Column string `json:"column"`
	Width  int    `json:"width"`
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
	Directory string   `json:"directory"`
	Files     []string `json:"files"`
}

type ExportFormat struct {
	ID              string            `json:"id"`
	Labels          map[string]string `json:"labels"`
	SupportsColumns bool              `json:"supportsColumns"`
	Extensions      []string          `json:"extensions,omitempty"`
}

type TableExportSettings struct {
	Path    string                             `json:"path"`
	Formats map[string]document.ExportSettings `json:"formats"`
}

type TableExportSettingsResult struct {
	Table     document.Table `json:"table"`
	Workspace View           `json:"workspace"`
}

type ProjectSettings struct {
	Project   ProjectInfo    `json:"project"`
	Export    ExportConfig   `json:"export"`
	Formats   []ExportFormat `json:"formats"`
	Tables    []ExportTable  `json:"tables"`
	Workspace View           `json:"workspace"`
}

type ProjectInfo struct {
	Path        string   `json:"path"`
	Root        string   `json:"root"`
	TableRoots  []string `json:"tableRoots"`
	SchemaRoots []string `json:"schemaRoots"`
	Exclude     []string `json:"exclude,omitempty"`
}

type ExportTable struct {
	ID         string   `json:"id"`
	Namespace  string   `json:"namespace"`
	Name       string   `json:"name"`
	Path       string   `json:"path"`
	Selected   bool     `json:"selected"`
	Configured bool     `json:"configured"`
	Columns    []string `json:"columns,omitempty"`
}

type ProjectExportSettings struct {
	Directory string          `json:"directory"`
	Targets   []string        `json:"targets"`
	Tables    map[string]bool `json:"tables"`
}

type CompileResult struct {
	Workspace View `json:"workspace"`
}

// Compile runs one full compile and publishes the outcome into the workspace.
func (s *Service) Compile() (CompileResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return CompileResult{}, s.userError("workspace.noProject")
	}
	s.compile()
	return CompileResult{Workspace: *s.current}, nil
}

// SetCompileMode switches between auto and manual compilation and persists the
// user-level editor settings.
func (s *Service) SetCompileMode(mode string) (View, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return View{}, s.userError("workspace.noProject")
	}
	if mode != "auto" && mode != "manual" {
		return View{}, s.userError("workspace.unsupportedBuildMode", mode)
	}
	settings, err := s.loadSettings()
	if err != nil {
		return View{}, err
	}
	settings.Build.Mode = mode
	if err := SaveEditorSettings(s.configRoot, settings); err != nil {
		s.log(logging.Error, "save editor settings failed", map[string]any{"error": err.Error()})
		return View{}, err
	}
	s.settings = settings
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
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.current == nil {
		return navigation.Target{}, s.userError("workspace.noProject")
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
		return navigation.Target{}, s.userError("workspace.unsupportedNavigationKind", request.Kind)
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
	return navigation.Target{}, s.userError("schema.referenceNotFound", reference)
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
		return navigation.Target{}, s.userError("navigation.notForeignKey", request.Column)
	}
	var tables []document.Table
	collectTables(s.current.Tree.Root, &tables)
	for _, table := range tables {
		if logicalTableName(table) == targetTable {
			return navigation.Target{Kind: navigation.KindForeignKey, Path: table.Path, RowID: request.RowID, Column: "id"}, nil
		}
	}
	return navigation.Target{}, s.userError("navigation.fkTargetNotFound", targetTable)
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
			return navigation.Target{}, s.userError("navigation.formulaNoReferences")
		}
		first := references[0]
		return navigation.Target{Kind: navigation.KindFormula, Path: table.Path, RowID: first.RowID, Column: first.Column}, nil
	}
	return navigation.Target{}, s.userError("navigation.noFormula", request.RowID, request.Column)
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
		return navigation.Target{}, s.userError("navigation.noReferences", name, request.RowID)
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

type TableRename struct {
	Path      string `json:"path"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type TableRenameResult struct {
	Table     document.Table `json:"table"`
	Workspace View           `json:"workspace"`
}

type TableDelete struct {
	Path string `json:"path"`
}

type TableDeleteResult struct {
	Workspace View `json:"workspace"`
}

type SchemaFileCreateResult struct {
	Schema    document.PublicSchema `json:"schema"`
	Workspace View                  `json:"workspace"`
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
	registry, err := exporters.Discover()
	if err != nil {
		panic(fmt.Sprintf("discover built-in exporters: %v", err))
	}
	return &Service{configRoot: configRoot, exporters: registry}
}

func (s *Service) SetLogger(logger *logging.Logger) {
	s.mu.Lock()
	defer s.mu.Unlock()
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

// loadSettings returns the cached user-level editor settings, loading them from
// disk on first use.
func (s *Service) loadSettings() (EditorSettings, error) {
	if s.settingsLoaded {
		return s.settings, nil
	}
	settings, err := LoadEditorSettings(s.configRoot)
	if err != nil {
		return EditorSettings{}, err
	}
	s.settings = settings
	s.settingsLoaded = true
	return s.settings, nil
}

// lang returns the current interface language, loading settings on demand.
func (s *Service) lang() string {
	if s.settingsLoaded {
		return s.settings.Language
	}
	if settings, err := LoadEditorSettings(s.configRoot); err == nil {
		return settings.Language
	}
	return DefaultLanguage
}

// userError builds a localized error from a message key.
func (s *Service) userError(key string, args ...any) error {
	return errors.New(i18n.Message(s.lang(), key, args))
}

// GetSettings returns the current user-level editor settings.
func (s *Service) GetSettings() (EditorSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadSettings()
}

// SetLanguage persists a new interface language and re-publishes the open view.
func (s *Service) SetLanguage(language string) (EditorSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !supportedLanguages[language] {
		return EditorSettings{}, s.userError("workspace.unsupportedLanguage", language)
	}
	settings, err := s.loadSettings()
	if err != nil {
		return EditorSettings{}, err
	}
	settings.Language = language
	if err := SaveEditorSettings(s.configRoot, settings); err != nil {
		s.log(logging.Error, "save editor settings failed", map[string]any{"error": err.Error()})
		return EditorSettings{}, err
	}
	s.settings = settings
	if s.current != nil {
		s.current.Settings.Language = language
	}
	s.log(logging.Info, "language changed", map[string]any{"language": language})
	return settings, nil
}

// SetTheme persists a new color theme preference.
func (s *Service) SetTheme(theme string) (EditorSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !supportedThemes[theme] {
		return EditorSettings{}, s.userError("workspace.unsupportedTheme", theme)
	}
	settings, err := s.loadSettings()
	if err != nil {
		return EditorSettings{}, err
	}
	settings.Theme = theme
	if err := SaveEditorSettings(s.configRoot, settings); err != nil {
		s.log(logging.Error, "save editor settings failed", map[string]any{"error": err.Error()})
		return EditorSettings{}, err
	}
	s.settings = settings
	if s.current != nil {
		s.current.Settings.Theme = theme
	}
	s.log(logging.Info, "theme changed", map[string]any{"theme": theme})
	return settings, nil
}

func (s *Service) OpenProject(path string) (View, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	started := time.Now()
	s.log(logging.Info, "open project started", map[string]any{"path": path})
	project, err := LoadProject(path)
	if err != nil {
		s.log(logging.Error, "open project failed", map[string]any{"path": path, "error": err.Error()})
		return View{}, err
	}
	settings, err := s.loadSettings()
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
	s.undoStack = nil
	s.redoStack = nil
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
	collectionDiagnostics := collectDocuments(s.current.Tree.Root, &publicSchemas, &tables)
	result := compiler.CompileAll(publicSchemas, tables)
	result.Diagnostics = append(collectionDiagnostics, result.Diagnostics...)
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
			result.Diagnostics = append(result.Diagnostics, compiler.Diagnostic{Path: node.Path, Key: "workspace.documentLoad", Message: diagnostic.Message})
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
	view.CanUndo = len(s.undoStack) > 0
	view.CanRedo = len(s.redoStack) > 0
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

func (s *Service) Current() (View, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.current == nil {
		return View{}, false
	}
	return *s.current, true
}

func (s *Service) GetProjectSettings() (ProjectSettings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.current == nil {
		return ProjectSettings{}, s.userError("workspace.noProject")
	}
	return s.projectSettings(), nil
}

func (s *Service) GetExportFormats() []ExportFormat {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.exportFormats()
}

func (s *Service) exportFormats() []ExportFormat {
	formats := s.exporters.Formats()
	result := make([]ExportFormat, 0, len(formats))
	for _, format := range formats {
		result = append(result, ExportFormat{ID: format.ID, Labels: format.Labels, SupportsColumns: format.SupportsColumns, Extensions: format.Extensions})
	}
	return result
}

func (s *Service) NormalizeProjectExportDirectory(path string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.current == nil {
		return "", s.userError("workspace.noProject")
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve export directory: %w", err)
	}
	relative, err := filepath.Rel(s.current.Project.Root, absPath)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("selected export directory is outside the project")
	}
	return filepath.ToSlash(relative), nil
}

func (s *Service) SetTableExportSettings(settings TableExportSettings) (TableExportSettingsResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return TableExportSettingsResult{}, s.userError("workspace.noProject")
	}
	if !containsTable(s.current.Tree.Root, settings.Path) {
		return TableExportSettingsResult{}, s.userError("workspace.tableNotInProject", settings.Path)
	}
	table, err := document.LoadTable(settings.Path)
	if err != nil {
		return TableExportSettingsResult{}, err
	}
	if err := s.validateTableExportSettings(table, settings.Formats); err != nil {
		return TableExportSettingsResult{}, err
	}
	updated, view, err := s.edit(settings.Path, func() (document.Table, error) {
		return document.SetTableExports(settings.Path, settings.Formats)
	})
	if err != nil {
		return TableExportSettingsResult{}, err
	}
	return TableExportSettingsResult{Table: withTableMetadata(updated, view), Workspace: view}, nil
}

func (s *Service) SetProjectExportSettings(settings ProjectExportSettings) (ProjectSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return ProjectSettings{}, s.userError("workspace.noProject")
	}
	config := s.current.Project.Config
	if strings.TrimSpace(settings.Directory) == "" {
		return ProjectSettings{}, fmt.Errorf("project export directory must not be empty")
	}
	config.Export.Directory = settings.Directory
	config.Export.Targets = append([]string(nil), settings.Targets...)
	for _, target := range config.Export.Targets {
		if !s.exporters.Has(target) {
			return ProjectSettings{}, fmt.Errorf("unknown exporter format %q", target)
		}
	}
	config.Export.Tables = copyBoolMap(settings.Tables)
	project := s.current.Project
	project.Config = config
	if err := validateProjectConfig(config); err != nil {
		return ProjectSettings{}, err
	}
	before, err := os.ReadFile(project.Path)
	if err != nil {
		return ProjectSettings{}, fmt.Errorf("record project settings source: %w", err)
	}
	if err := SaveProject(project); err != nil {
		return ProjectSettings{}, err
	}
	s.current.Project = project
	s.recordEdit(map[string][]byte{project.Path: before})
	s.current.CanUndo = true
	s.current.CanRedo = false
	return s.projectSettings(), nil
}

func (s *Service) projectSettings() ProjectSettings {
	project := s.current.Project
	tables := []ExportTable{}
	collectExportTables(s.current.Tree.Root, &tables, project.Config.Export.Tables)
	return ProjectSettings{
		Project:   ProjectInfo{Path: project.Path, Root: project.Root, TableRoots: append([]string(nil), project.Config.TableRoots...), SchemaRoots: append([]string(nil), project.Config.SchemaRoots...), Exclude: append([]string(nil), project.Config.Exclude...)},
		Export:    project.Config.Export,
		Formats:   s.exportFormats(),
		Tables:    tables,
		Workspace: *s.current,
	}
}

func (s *Service) validateTableExportSettings(table document.Table, settings map[string]document.ExportSettings) error {
	knownColumns := map[string]bool{"id": true}
	for _, column := range table.Columns {
		knownColumns[column.Name] = true
	}
	for id, format := range settings {
		definition, exists := s.exporters.Format(id)
		if !exists {
			return fmt.Errorf("unknown exporter format %q", id)
		}
		if format.Enabled && definition.SupportsColumns && len(format.Columns) == 0 {
			return fmt.Errorf("export format %q requires at least one column", id)
		}
		seen := map[string]bool{}
		for _, column := range format.Columns {
			if !knownColumns[column] {
				return fmt.Errorf("export format %q references unknown column %q", id, column)
			}
			if seen[column] {
				return fmt.Errorf("export format %q contains duplicate column %q", id, column)
			}
			seen[column] = true
		}
	}
	return nil
}

func collectExportTables(node projecttree.Node, tables *[]ExportTable, selected map[string]bool) {
	if node.Kind == projecttree.NodeTable {
		table, err := document.LoadTable(node.Path)
		if err == nil {
			id := logicalTableName(table)
			chosen := true
			if selected != nil {
				chosen = selected[id]
			}
			entry := ExportTable{ID: id, Namespace: table.Namespace, Name: table.Name, Path: table.Path, Selected: chosen}
			for formatID, export := range table.Exports {
				if export.Enabled {
					entry.Configured = true
					if formatID == "csv" {
						entry.Columns = append([]string(nil), export.Columns...)
					}
				}
			}
			if table.CSVExport != nil && !entry.Configured {
				entry.Configured = true
				entry.Columns = append([]string(nil), table.CSVExport.Columns...)
			}
			*tables = append(*tables, entry)
		}
	}
	for _, child := range node.Children {
		collectExportTables(child, tables, selected)
	}
	sort.Slice(*tables, func(i, j int) bool { return (*tables)[i].ID < (*tables)[j].ID })
}

func copyBoolMap(values map[string]bool) map[string]bool {
	if values == nil {
		return nil
	}
	copy := make(map[string]bool, len(values))
	for key, value := range values {
		copy[key] = value
	}
	return copy
}

// GetTable loads a table document that belongs to the current project. The
// path is resolved through the project tree instead of being read directly from
// an untrusted frontend request.
func (s *Service) GetTable(path string) (document.Table, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.current == nil {
		err := s.userError("workspace.noProject")
		s.log(logging.Error, "get table failed", map[string]any{"path": path, "error": err.Error()})
		return document.Table{}, err
	}
	if !containsTable(s.current.Tree.Root, path) {
		err := s.userError("workspace.tableNotInProject", path)
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
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.current == nil {
		err := s.userError("workspace.noProject")
		s.log(logging.Error, "get public schema failed", map[string]any{"path": path, "error": err.Error()})
		return document.PublicSchema{}, err
	}
	if !s.containsSchemaFile(path) {
		err := s.userError("workspace.schemaFileNotInProject", path)
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
	if s.containsSchemaFile(path) {
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
	if s.containsSchemaFile(path) {
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
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return SchemaEditResult{}, s.userError("workspace.noProject")
	}
	update := document.SchemaDefinitionUpdate{Kind: create.Kind, Type: create.Type, AppliesTo: create.AppliesTo, SchemaExpression: create.SchemaExpression}
	scope, err := s.schemaDefinitionScope(create.Path)
	if err != nil {
		return SchemaEditResult{}, err
	}
	snapshot := s.snapshotFiles(create.Path)
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
	s.recordEdit(snapshot)
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
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return SchemaEditResult{}, s.userError("workspace.noProject")
	}
	update := document.SchemaDefinitionUpdate{Kind: set.Kind, Type: set.Type, AppliesTo: set.AppliesTo, SchemaExpression: set.SchemaExpression}
	scope, err := s.schemaDefinitionScope(set.Path)
	if err != nil {
		return SchemaEditResult{}, err
	}
	snapshot := s.snapshotFiles(set.Path)
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
	s.recordEdit(snapshot)
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
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return SchemaEditResult{}, s.userError("workspace.noProject")
	}
	scope, err := s.schemaDefinitionScope(rename.Path)
	if err != nil {
		return SchemaEditResult{}, err
	}
	var paths []string
	collectSourcePaths(s.current.Tree.Root, &paths)
	snapshot := s.snapshotFiles(paths...)
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
	s.recordEdit(snapshot)
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
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return SchemaEditResult{}, s.userError("workspace.noProject")
	}
	scope, err := s.schemaDefinitionScope(delete.Path)
	if err != nil {
		return SchemaEditResult{}, err
	}
	id, err := s.schemaDefinitionID(delete.Path, delete.Name)
	if err != nil {
		return SchemaEditResult{}, err
	}
	snapshot := s.snapshotFiles(delete.Path)
	for _, definition := range s.current.Schemas.Definitions {
		if definition.ID != id || len(definition.Usages) == 0 {
			continue
		}
		usage := definition.Usages[0]
		err := s.userError("schema.referencedBy", id, usage.Path, usage.Field)
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
	s.recordEdit(snapshot)
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
	expression = schemamodel.NormalizeNegativeComparisons(expression)
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
			return nil, s.userError("schema.referenceAmbiguous", name)
		}
		return nil, s.userError("schema.referenceNotFoundHint", name, name, name)
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
		return "", s.userError("workspace.schemaPackageNotInProject", packageName)
	}
	relative, err := filepath.Rel(s.current.Project.Root, directory)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", s.userError("workspace.schemaPackageOutside", packageName)
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
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return TableCreateResult{}, s.userError("workspace.noProject")
	}
	if !validTableName(create.Name) || !validNamespace(create.Namespace) {
		return TableCreateResult{}, s.userError("workspace.tableNameInvalid")
	}
	var tables []document.Table
	collectTables(s.current.Tree.Root, &tables)
	for _, table := range tables {
		if table.Namespace == create.Namespace && table.Name == create.Name {
			return TableCreateResult{}, s.userError("workspace.tableExists", create.Namespace+"."+create.Name)
		}
	}
	directory := filepath.Join(s.current.Project.Root, s.current.Project.Config.TableRoots[0], create.Namespace, create.Name)
	path := filepath.Join(directory, create.Name+".cue")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return TableCreateResult{}, fmt.Errorf("create table directory: %w", err)
	}
	if _, err := os.Stat(path); err == nil {
		return TableCreateResult{}, s.userError("workspace.tableExists", create.Namespace+"."+create.Name)
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

// RenameTable changes a table's namespace and name, moves its source directory,
// and updates every foreign-key reference to it.
func (s *Service) RenameTable(rename TableRename) (TableRenameResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return TableRenameResult{}, s.userError("workspace.noProject")
	}
	if !containsTable(s.current.Tree.Root, rename.Path) {
		return TableRenameResult{}, s.userError("workspace.tableNotInProject", rename.Path)
	}
	if !validTableName(rename.Name) || !validNamespace(rename.Namespace) {
		return TableRenameResult{}, s.userError("workspace.tableNameInvalid")
	}
	source, err := document.LoadTable(rename.Path)
	if err != nil {
		return TableRenameResult{}, err
	}
	oldName := logicalTableName(source)
	var tables []document.Table
	collectTables(s.current.Tree.Root, &tables)
	for _, table := range tables {
		if table.Path == rename.Path {
			continue
		}
		if table.Namespace == rename.Namespace && table.Name == rename.Name {
			return TableRenameResult{}, s.userError("workspace.tableExists", rename.Namespace+"."+rename.Name)
		}
	}
	newName := rename.Namespace + "." + rename.Name
	references := []document.TableReference{}
	for _, table := range tables {
		if table.Path == rename.Path {
			continue
		}
		for _, key := range table.ForeignKeys {
			if key.Table == oldName {
				references = append(references, document.TableReference{Path: table.Path, Column: key.Column, OldTable: oldName, NewTable: newName})
			}
		}
	}
	newDirectory := filepath.Join(s.current.Project.Root, s.current.Project.Config.TableRoots[0], rename.Namespace, rename.Name)
	newPath := filepath.Join(newDirectory, rename.Name+".cue")
	files := []string{rename.Path, newPath}
	if s.current.Project.Config.Export.Tables != nil {
		files = append(files, s.current.Project.Path)
	}
	for _, reference := range references {
		files = append(files, reference.Path)
	}
	snapshot := s.snapshotFiles(files...)
	if _, err := document.RenameTable(rename.Path, rename.Namespace, rename.Name, references); err != nil {
		return TableRenameResult{}, err
	}
	oldDirectory := filepath.Dir(rename.Path)
	if err := os.MkdirAll(filepath.Dir(newDirectory), 0o755); err != nil {
		return TableRenameResult{}, fmt.Errorf("create renamed table directory: %w", err)
	}
	if err := os.Rename(oldDirectory, newDirectory); err != nil {
		return TableRenameResult{}, fmt.Errorf("move renamed table: %w", err)
	}
	movedSource := filepath.Join(newDirectory, filepath.Base(rename.Path))
	if movedSource != newPath {
		if err := os.Rename(movedSource, newPath); err != nil {
			_ = os.RemoveAll(newDirectory)
			_ = s.restoreSnapshot(snapshot)
			return TableRenameResult{}, fmt.Errorf("rename table source file: %w", err)
		}
	}
	project := s.current.Project
	if migrateExportTableSelection(&project.Config, oldName, newName) {
		if err := SaveProject(project); err != nil {
			_ = os.RemoveAll(newDirectory)
			_ = s.restoreSnapshot(snapshot)
			return TableRenameResult{}, err
		}
		s.current.Project = project
	}
	s.recordEdit(snapshot)
	view, err := s.refreshAfterEdit()
	if err != nil {
		return TableRenameResult{}, err
	}
	table, err := document.LoadTable(filepath.Join(newDirectory, rename.Name+".cue"))
	if err != nil {
		return TableRenameResult{}, fmt.Errorf("reload renamed table: %w", err)
	}
	s.log(logging.Info, "table renamed", map[string]any{"path": rename.Path, "old": oldName, "new": newName})
	return TableRenameResult{Table: withTableMetadata(table, view), Workspace: view}, nil
}

// DeleteTable removes a table source directory unless another table references
// it through a foreign key.
func (s *Service) DeleteTable(delete TableDelete) (TableDeleteResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return TableDeleteResult{}, s.userError("workspace.noProject")
	}
	if !containsTable(s.current.Tree.Root, delete.Path) {
		return TableDeleteResult{}, s.userError("workspace.tableNotInProject", delete.Path)
	}
	source, err := document.LoadTable(delete.Path)
	if err != nil {
		return TableDeleteResult{}, err
	}
	name := logicalTableName(source)
	var tables []document.Table
	collectTables(s.current.Tree.Root, &tables)
	referencedBy := []string{}
	for _, table := range tables {
		if table.Path == delete.Path {
			continue
		}
		for _, key := range table.ForeignKeys {
			if key.Table == name {
				referencedBy = append(referencedBy, logicalTableName(table)+"."+key.Column)
			}
		}
	}
	if len(referencedBy) > 0 {
		err := s.userError("workspace.tableReferenced", name, strings.Join(referencedBy, ", "))
		s.log(logging.Warn, "table delete blocked", map[string]any{"path": delete.Path, "error": err.Error()})
		return TableDeleteResult{}, err
	}
	paths := []string{delete.Path}
	if s.current.Project.Config.Export.Tables != nil {
		paths = append(paths, s.current.Project.Path)
	}
	snapshot := s.snapshotFiles(paths...)
	if err := os.RemoveAll(filepath.Dir(delete.Path)); err != nil {
		return TableDeleteResult{}, fmt.Errorf("delete table source: %w", err)
	}
	project := s.current.Project
	if removeExportTableSelection(&project.Config, name) {
		if err := SaveProject(project); err != nil {
			_ = s.restoreSnapshot(snapshot)
			return TableDeleteResult{}, err
		}
		s.current.Project = project
	}
	s.recordEdit(snapshot)
	view, err := s.refreshAfterEdit()
	if err != nil {
		return TableDeleteResult{}, err
	}
	s.log(logging.Info, "table deleted", map[string]any{"path": delete.Path, "name": name})
	return TableDeleteResult{Workspace: view}, nil
}

// CreateSchemaFile creates an empty public Schema package for a namespace.
func (s *Service) CreateSchemaFile(namespace string) (SchemaFileCreateResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return SchemaFileCreateResult{}, s.userError("workspace.noProject")
	}
	if !validNamespace(namespace) {
		return SchemaFileCreateResult{}, s.userError("workspace.schemaNamespaceInvalid")
	}
	directory := filepath.Join(s.current.Project.Root, s.current.Project.Config.SchemaRoots[0], namespace)
	path := filepath.Join(directory, namespace+".cue")
	if _, err := os.Stat(path); err == nil {
		return SchemaFileCreateResult{}, s.userError("workspace.schemaNamespaceExists", namespace)
	}
	snapshot := s.snapshotFiles(path)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return SchemaFileCreateResult{}, fmt.Errorf("create schema directory: %w", err)
	}
	source := fmt.Sprintf("package %s\n\n_schema: {\n\tversion: 1\n\tdefinitions: {}\n}\n", namespace)
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		return SchemaFileCreateResult{}, fmt.Errorf("write schema source: %w", err)
	}
	schema, err := document.LoadPublicSchema(path)
	if err != nil {
		_ = os.Remove(path)
		return SchemaFileCreateResult{}, fmt.Errorf("validate created schema: %w", err)
	}
	s.recordEdit(snapshot)
	view, err := s.refreshAfterEdit()
	if err != nil {
		return SchemaFileCreateResult{}, err
	}
	s.log(logging.Info, "schema file created", map[string]any{"namespace": namespace})
	return SchemaFileCreateResult{Schema: schema, Workspace: view}, nil
}

// UpdateCell applies one CUE expression to an existing table cell, then
// refreshes the project tree and Schema registry from the saved source.
func (s *Service) UpdateCell(update CellUpdate) (CellUpdateResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return CellUpdateResult{}, s.userError("workspace.noProject")
	}
	if !containsTable(s.current.Tree.Root, update.Path) {
		return CellUpdateResult{}, s.userError("workspace.tableNotInProject", update.Path)
	}
	table, view, err := s.edit(update.Path, func() (document.Table, error) {
		return document.SetCellValue(update.Path, update.RowID, update.Column, update.Expression)
	})
	if err != nil {
		return CellUpdateResult{}, err
	}
	return CellUpdateResult{Table: withTableMetadata(table, view), Workspace: view}, nil
}

func (s *Service) SetCellText(update CellUpdate) (CellUpdateResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return CellUpdateResult{}, s.userError("workspace.noProject")
	}
	if !containsTable(s.current.Tree.Root, update.Path) {
		return CellUpdateResult{}, s.userError("workspace.tableNotInProject", update.Path)
	}
	table, view, err := s.edit(update.Path, func() (document.Table, error) {
		return document.SetCellText(update.Path, update.RowID, update.Column, update.Expression)
	})
	if err != nil {
		return CellUpdateResult{}, err
	}
	return CellUpdateResult{Table: withTableMetadata(table, view), Workspace: view}, nil
}

func (s *Service) SetCellComment(update CellCommentSet) (CellUpdateResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return CellUpdateResult{}, s.userError("workspace.noProject")
	}
	if !containsTable(s.current.Tree.Root, update.Path) {
		return CellUpdateResult{}, s.userError("workspace.tableNotInProject", update.Path)
	}
	table, view, err := s.edit(update.Path, func() (document.Table, error) {
		return document.SetCellComment(update.Path, update.RowID, update.Column, update.Comment)
	})
	if err != nil {
		return CellUpdateResult{}, err
	}
	return CellUpdateResult{Table: withTableMetadata(table, view), Workspace: view}, nil
}

func (s *Service) SetColumnDisplayName(update ColumnDisplayNameSet) (CellUpdateResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return CellUpdateResult{}, s.userError("workspace.noProject")
	}
	if !containsTable(s.current.Tree.Root, update.Path) {
		return CellUpdateResult{}, s.userError("workspace.tableNotInProject", update.Path)
	}
	table, view, err := s.edit(update.Path, func() (document.Table, error) {
		return document.SetColumnDisplayName(update.Path, update.Column, update.DisplayName)
	})
	if err != nil {
		return CellUpdateResult{}, err
	}
	return CellUpdateResult{Table: withTableMetadata(table, view), Workspace: view}, nil
}

func (s *Service) SetColumnWidth(update ColumnWidthSet) (CellUpdateResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return CellUpdateResult{}, s.userError("workspace.noProject")
	}
	if !containsTable(s.current.Tree.Root, update.Path) {
		return CellUpdateResult{}, s.userError("workspace.tableNotInProject", update.Path)
	}
	table, view, err := s.edit(update.Path, func() (document.Table, error) {
		return document.SetColumnWidth(update.Path, update.Column, update.Width)
	})
	if err != nil {
		return CellUpdateResult{}, err
	}
	return CellUpdateResult{Table: withTableMetadata(table, view), Workspace: view}, nil
}

func (s *Service) Paste(paste Paste) (HistoryResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return HistoryResult{}, s.userError("workspace.noProject")
	}
	if !containsTable(s.current.Tree.Root, paste.Path) {
		return HistoryResult{}, s.userError("workspace.tableNotInProject", paste.Path)
	}
	table, view, err := s.edit(paste.Path, func() (document.Table, error) {
		return document.PasteCells(paste.Path, paste.RowIDs, paste.Columns, paste.Values)
	})
	if err != nil {
		return HistoryResult{}, err
	}
	return s.historyResult(table, view), nil
}

func (s *Service) Undo() (HistoryResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.restoreHistory(true)
}

func (s *Service) Redo() (HistoryResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.restoreHistory(false)
}

func (s *Service) SetCellFormula(set FormulaSet) (FormulaSetResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return FormulaSetResult{}, s.userError("workspace.noProject")
	}
	if !containsTable(s.current.Tree.Root, set.Path) {
		return FormulaSetResult{}, s.userError("workspace.tableNotInProject", set.Path)
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
	s.recordEdit(map[string][]byte{path: before})
	view, err := s.refreshAfterEdit()
	if err != nil {
		s.log(logging.Error, "workspace refresh failed after edit", map[string]any{"path": path, "error": err.Error()})
		return document.Table{}, View{}, err
	}
	s.log(logging.Info, "table edit completed", map[string]any{"path": path, "duration_ms": time.Since(started).Milliseconds()})
	return table, view, nil
}

func (s *Service) restoreHistory(undo bool) (HistoryResult, error) {
	if s.current == nil {
		return HistoryResult{}, s.userError("workspace.noProject")
	}
	stack := &s.undoStack
	opposite := &s.redoStack
	if !undo {
		stack, opposite = &s.redoStack, &s.undoStack
	}
	if len(*stack) == 0 {
		if undo {
			return HistoryResult{}, s.userError("workspace.nothingToUndo")
		}
		return HistoryResult{}, s.userError("workspace.nothingToRedo")
	}
	snapshot := (*stack)[len(*stack)-1]
	*stack = (*stack)[:len(*stack)-1]
	current := s.snapshotFiles(snapshotPaths(snapshot)...)
	if err := s.restoreSnapshot(snapshot); err != nil {
		return HistoryResult{}, err
	}
	view, err := s.refreshAfterEdit()
	if err != nil {
		return HistoryResult{}, err
	}
	*opposite = append(*opposite, current)
	view.CanUndo = len(s.undoStack) > 0
	view.CanRedo = len(s.redoStack) > 0
	s.current = &view
	return HistoryResult{Workspace: view, CanUndo: view.CanUndo, CanRedo: view.CanRedo}, nil
}

func (s *Service) historyResult(table document.Table, view View) HistoryResult {
	return HistoryResult{Table: withTableMetadata(table, view), Workspace: view, CanUndo: len(s.undoStack) > 0, CanRedo: len(s.redoStack) > 0}
}

func (s *Service) snapshotFiles(paths ...string) map[string][]byte {
	snapshot := map[string][]byte{}
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			snapshot[path] = nil
			continue
		}
		snapshot[path] = content
	}
	return snapshot
}

func (s *Service) recordEdit(snapshot map[string][]byte) {
	s.undoStack = append(s.undoStack, snapshot)
	s.redoStack = nil
}

func snapshotPaths(snapshot map[string][]byte) []string {
	paths := make([]string, 0, len(snapshot))
	for path := range snapshot {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func (s *Service) restoreSnapshot(snapshot map[string][]byte) error {
	for path, content := range snapshot {
		if content == nil {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove restored file: %w", err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("create restored directory: %w", err)
		}
		if err := writeFileAtomic(path, content); err != nil {
			return fmt.Errorf("restore source: %w", err)
		}
	}
	return nil
}

func writeFileAtomic(path string, content []byte) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".datapack-restore-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.Write(content); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}

func (s *Service) CreateRow(create RowCreate) (RowCreateResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return RowCreateResult{}, s.userError("workspace.noProject")
	}
	if !containsTable(s.current.Tree.Root, create.Path) {
		return RowCreateResult{}, s.userError("workspace.tableNotInProject", create.Path)
	}
	values := maps.Clone(create.Values)
	if values == nil {
		values = map[string]string{}
	}
	formulas := maps.Clone(create.Formulas)
	if formulas == nil {
		formulas = map[string]string{}
	}
	if source, err := document.LoadTable(create.Path); err == nil {
		for _, column := range source.Columns {
			if _, exists := values[column.Name]; exists {
				continue
			}
			if _, formulaExists := formulas[column.Name]; formulaExists {
				continue
			}
			if expression, ok := schemamodel.DefaultExpression(column.Schema); ok {
				values[column.Name] = expression
			}
		}
	}
	table, view, err := s.edit(create.Path, func() (document.Table, error) {
		return document.CreateRow(create.Path, create.RowID, values, formulas)
	})
	if err != nil {
		return RowCreateResult{}, err
	}
	return RowCreateResult{Table: withTableMetadata(table, view), Workspace: view}, nil
}

func (s *Service) ClearCell(update CellUpdate) (CellUpdateResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return CellUpdateResult{}, s.userError("workspace.noProject")
	}
	if !containsTable(s.current.Tree.Root, update.Path) {
		return CellUpdateResult{}, s.userError("workspace.tableNotInProject", update.Path)
	}
	table, view, err := s.edit(update.Path, func() (document.Table, error) {
		source, err := document.LoadTable(update.Path)
		if err == nil {
			for _, column := range source.Columns {
				if column.Name == update.Column {
					if expression, ok := schemamodel.DefaultExpression(column.Schema); ok {
						return document.SetCellValue(update.Path, update.RowID, update.Column, expression)
					}
					return document.Table{}, fmt.Errorf("column %q has no safe default value", update.Column)
				}
			}
		}
		return document.ClearCell(update.Path, update.RowID, update.Column)
	})
	if err != nil {
		return CellUpdateResult{}, err
	}
	return CellUpdateResult{Table: withTableMetadata(table, view), Workspace: view}, nil
}

func (s *Service) AddRow(add RowAdd) (RowAddResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return RowAddResult{}, s.userError("workspace.noProject")
	}
	if !containsTable(s.current.Tree.Root, add.Path) {
		return RowAddResult{}, s.userError("workspace.tableNotInProject", add.Path)
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
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return RowDeleteResult{}, s.userError("workspace.noProject")
	}
	if !containsTable(s.current.Tree.Root, delete.Path) {
		return RowDeleteResult{}, s.userError("workspace.tableNotInProject", delete.Path)
	}
	table, err := document.LoadTable(delete.Path)
	if err != nil {
		return RowDeleteResult{}, fmt.Errorf("load table: %w", err)
	}
	if usage := foreignKeyUsage(s.current.Tree, table, delete.RowID); usage != nil {
		return RowDeleteResult{}, s.userError("workspace.rowReferenced", delete.RowID, usage.Table, usage.RowID, usage.Column)
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
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return RowRenameResult{}, s.userError("workspace.noProject")
	}
	if !containsTable(s.current.Tree.Root, rename.Path) {
		return RowRenameResult{}, s.userError("workspace.tableNotInProject", rename.Path)
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
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		err := s.userError("workspace.noProject")
		s.log(logging.Error, "export failed", map[string]any{"error": err.Error()})
		return ExportResult{}, err
	}
	compiled := s.compile()
	if len(compiled.Diagnostics) > 0 {
		err := s.userError("workspace.exportBlocked", len(compiled.Diagnostics), compiled.Diagnostics[0].Message)
		s.log(logging.Warn, "export blocked", map[string]any{"error": err.Error()})
		return ExportResult{}, err
	}
	destination, err := exportDirectory(s.current.Project)
	if err != nil {
		s.log(logging.Error, "export failed", map[string]any{"error": err.Error()})
		return ExportResult{}, err
	}
	compiled.Package = filterExportPackage(compiled.Package, s.current.Project.Config.Export.Tables)
	if !hasExportableTables(compiled.Package) {
		return ExportResult{}, s.userError("workspace.noSelectedTables")
	}
	if !containsString(s.current.Project.Config.Export.Targets, "csv") {
		return ExportResult{}, s.userError("workspace.noExportTargets")
	}
	result, err := s.exporters.Export("csv", destination, compiled.Package)
	if err != nil {
		s.log(logging.Error, "export failed", map[string]any{"destination": destination, "error": err.Error()})
		return ExportResult{}, err
	}
	s.log(logging.Info, "export completed", map[string]any{"files": len(result.Files), "destination": destination})
	return ExportResult{Directory: destination, Files: result.Files}, nil
}

func filterExportPackage(pkg *compiler.Package, selected map[string]bool) *compiler.Package {
	if pkg == nil || selected == nil {
		return pkg
	}
	filtered := &compiler.Package{Tables: make([]compiler.Table, 0, len(pkg.Tables))}
	for _, table := range pkg.Tables {
		id := table.Name
		if table.Namespace != "" {
			id = table.Namespace + "." + table.Name
		}
		if selected[id] {
			filtered.Tables = append(filtered.Tables, table)
		}
	}
	return filtered
}

func hasExportableTables(pkg *compiler.Package) bool {
	if pkg == nil {
		return false
	}
	for _, table := range pkg.Tables {
		if table.CSV != nil {
			return true
		}
	}
	return false
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func migrateExportTableSelection(config *ProjectConfig, oldID, newID string) bool {
	if config.Export.Tables == nil {
		return false
	}
	selected, exists := config.Export.Tables[oldID]
	if !exists {
		return false
	}
	delete(config.Export.Tables, oldID)
	config.Export.Tables[newID] = selected
	return true
}

func removeExportTableSelection(config *ProjectConfig, id string) bool {
	if config.Export.Tables == nil {
		return false
	}
	if _, exists := config.Export.Tables[id]; !exists {
		return false
	}
	delete(config.Export.Tables, id)
	return true
}

func (s *Service) AddColumn(add ColumnAdd) (ColumnAddResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return ColumnAddResult{}, s.userError("workspace.noProject")
	}
	if !containsTable(s.current.Tree.Root, add.Path) {
		return ColumnAddResult{}, s.userError("workspace.tableNotInProject", add.Path)
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
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return ColumnSchemaSetResult{}, s.userError("workspace.noProject")
	}
	if !containsTable(s.current.Tree.Root, set.Path) {
		return ColumnSchemaSetResult{}, s.userError("workspace.tableNotInProject", set.Path)
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
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return ColumnDeleteResult{}, s.userError("workspace.noProject")
	}
	if !containsTable(s.current.Tree.Root, delete.Path) {
		return ColumnDeleteResult{}, s.userError("workspace.tableNotInProject", delete.Path)
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
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return ColumnRenameResult{}, s.userError("workspace.noProject")
	}
	if !containsTable(s.current.Tree.Root, rename.Path) {
		return ColumnRenameResult{}, s.userError("workspace.tableNotInProject", rename.Path)
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
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil {
		return ColumnMoveResult{}, s.userError("workspace.noProject")
	}
	if !containsTable(s.current.Tree.Root, move.Path) {
		return ColumnMoveResult{}, s.userError("workspace.tableNotInProject", move.Path)
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
	project, err := LoadProject(s.current.Project.Path)
	if err != nil {
		s.log(logging.Error, "workspace refresh failed", map[string]any{"error": err.Error()})
		return View{}, fmt.Errorf("reload project configuration: %w", err)
	}
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
	if (node.Kind == projecttree.NodeSchemaFile || node.Kind == projecttree.NodeSchemaDefinition) && node.Path == path {
		return true
	}
	for _, child := range node.Children {
		if containsSchemaFile(child, path) {
			return true
		}
	}
	return false
}

func (s *Service) containsSchemaFile(path string) bool {
	if containsSchemaFile(s.current.Tree.Root, path) {
		return true
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	for _, root := range s.current.Project.Config.SchemaRoots {
		schemaRoot := filepath.Join(s.current.Project.Root, root)
		relative, err := filepath.Rel(schemaRoot, absPath)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			continue
		}
		_, err = document.LoadPublicSchema(absPath)
		return err == nil
	}
	return false
}

func collectSourcePaths(node projecttree.Node, paths *[]string) {
	if (node.Kind == projecttree.NodeTable || node.Kind == projecttree.NodeSchemaFile || node.Kind == projecttree.NodeSchemaDefinition) && node.Path != "" {
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
	table = withForeignKeyOptions(withDiagnostics(table, view), view.Tree)
	definitions := map[string]schemamodel.Node{}
	for _, definition := range view.Schemas.Definitions {
		definitions[string(definition.ID)] = definition.Schema
	}
	logical := table.Name
	if table.Namespace != "" {
		logical = table.Namespace + "." + table.Name
	}
	for index := range table.Columns {
		table.Columns[index].EffectiveSchema = resolveEffectiveSchema(table.Columns[index].Schema, logical, definitions, map[string]bool{})
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
