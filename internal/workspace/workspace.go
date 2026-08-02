package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"datapack/internal/document"
	"datapack/internal/export/csv"
	"datapack/internal/foreignkey"
	"datapack/internal/formula"
	"datapack/internal/projecttree"
	"datapack/internal/schemaregistry"
)

type View struct {
	Project     Project              `json:"project"`
	Settings    EditorSettings       `json:"settings"`
	Tree        projecttree.Tree     `json:"tree"`
	Schemas     schemaregistry.Index `json:"schemas"`
	ForeignKeys foreignkey.Summary   `json:"foreignKeys"`
}

type Service struct {
	configRoot string
	current    *View
	history    map[string]*history
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

type ColumnAdd struct {
	Path              string `json:"path"`
	Name              string `json:"name"`
	SchemaExpression  string `json:"schemaExpression"`
	InitialExpression string `json:"initialExpression"`
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

func NewService(configRoot string) *Service {
	return &Service{configRoot: configRoot, history: map[string]*history{}}
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
	view := buildView(project, settings, tree)
	s.current = &view
	return view, nil
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
		return document.Table{}, fmt.Errorf("no project is open")
	}
	if !containsTable(s.current.Tree.Root, path) {
		return document.Table{}, fmt.Errorf("table %q is not part of the open project", path)
	}
	table, err := document.LoadTable(path)
	if err != nil {
		return document.Table{}, fmt.Errorf("load table: %w", err)
	}
	return withTableMetadata(table, *s.current), nil
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
	before, err := os.ReadFile(path)
	if err != nil {
		return document.Table{}, View{}, fmt.Errorf("record edit source: %w", err)
	}
	table, err := operation()
	if err != nil {
		return document.Table{}, View{}, err
	}
	view, err := s.refreshCurrentView()
	if err != nil {
		return document.Table{}, View{}, err
	}
	record := s.history[path]
	if record == nil {
		record = &history{}
		s.history[path] = record
	}
	record.undo = append(record.undo, before)
	record.redo = nil
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
	view, err := s.refreshCurrentView()
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
	return RowCreateResult{Table: withTableMetadata(table, view)}, nil
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
	return CellUpdateResult{Table: withTableMetadata(table, view)}, nil
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
	view, err := s.refreshCurrentView()
	if err != nil {
		return RowRenameResult{}, err
	}
	return RowRenameResult{Table: table, Workspace: view}, nil
}

// Export writes configured CSV files only when the current workspace has no
// Schema or foreign-key diagnostics.
func (s *Service) Export() (ExportResult, error) {
	if s.current == nil {
		return ExportResult{}, fmt.Errorf("no project is open")
	}
	if len(s.current.Schemas.Diagnostics) > 0 {
		return ExportResult{}, fmt.Errorf("export blocked by %d Schema diagnostics", len(s.current.Schemas.Diagnostics))
	}
	if len(s.current.ForeignKeys.Diagnostics) > 0 {
		return ExportResult{}, fmt.Errorf("export blocked by %d foreign key diagnostics", len(s.current.ForeignKeys.Diagnostics))
	}
	var tables []document.Table
	collectTables(s.current.Tree.Root, &tables)
	destination, err := exportDirectory(s.current.Project)
	if err != nil {
		return ExportResult{}, err
	}
	result, err := csv.Export(destination, tables)
	if err != nil {
		return ExportResult{}, err
	}
	return ExportResult{Files: result.Files}, nil
}

func (s *Service) AddColumn(add ColumnAdd) (ColumnAddResult, error) {
	if s.current == nil {
		return ColumnAddResult{}, fmt.Errorf("no project is open")
	}
	if !containsTable(s.current.Tree.Root, add.Path) {
		return ColumnAddResult{}, fmt.Errorf("table %q is not part of the open project", add.Path)
	}
	table, view, err := s.edit(add.Path, func() (document.Table, error) {
		return document.AddColumn(add.Path, add.Name, add.SchemaExpression, add.InitialExpression)
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
	table, view, err := s.edit(set.Path, func() (document.Table, error) {
		return document.SetColumnSchema(set.Path, set.Name, set.SchemaExpression)
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
	project := s.current.Project
	tree, err := projecttree.Scan(project.Path, project.Config.TableRoots, project.Config.SchemaRoots, project.Config.Exclude)
	if err != nil {
		return View{}, fmt.Errorf("refresh project tree: %w", err)
	}
	view := buildView(project, s.current.Settings, tree)
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

func withDiagnostics(table document.Table, view View) document.Table {
	evaluated := formula.Evaluate(table)
	for _, diagnostic := range evaluated.Diagnostics {
		table.Diagnostics = append(table.Diagnostics, document.Diagnostic{RowID: diagnostic.RowID, Column: diagnostic.Column, Message: diagnostic.Message})
	}
	appendCueDiagnostics(&table, evaluated.Rows)
	for _, diagnostic := range view.ForeignKeys.Diagnostics {
		if diagnostic.Path == table.Path {
			table.Diagnostics = append(table.Diagnostics, document.Diagnostic{RowID: diagnostic.RowID, Column: diagnostic.Column, Message: diagnostic.Message})
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
