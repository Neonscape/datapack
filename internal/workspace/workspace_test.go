package workspace

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"datapack/internal/document"
	"datapack/internal/projecttree"
)

func TestOpenProjectLoadsSettingsAndSourceTree(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{
  "version": 1,
  "tableRoots": ["tables"],
  "schemaRoots": ["schema"],
  "exclude": ["tables/generated"],
  "export": {"directory": "exports", "targets": ["csv"]}
}`)
	writeFile(t, filepath.Join(projectRoot, "tables", "items", "item.cue"), `package item
	#PrimaryKey: string
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}}
`)
	writeFile(t, filepath.Join(projectRoot, "tables", "generated", "ignored.cue"), "package generated\n")
	writeFile(t, filepath.Join(projectRoot, "schema", "common", "common.cue"), `package common
#GameID: string
_schema: {definitions: {GameID: {kind: "type", type: "string"}}}
`)

	configRoot := t.TempDir()
	writeFile(t, SettingsPath(configRoot), `{
  "version": 1,
  "build": {"mode": "manual"}
}`)

	view, err := NewService(configRoot).OpenProject(filepath.Join(projectRoot, "game.dtproj"))
	if err != nil {
		t.Fatalf("OpenProject() error = %v", err)
	}
	if view.Settings.Build.Mode != "manual" {
		t.Fatalf("build mode = %q, want manual", view.Settings.Build.Mode)
	}
	if got := findNode(view.Tree.Root, "rpg.item"); got == nil || got.Kind != projecttree.NodeTable {
		t.Fatalf("table source was not discovered: %#v", view.Tree)
	}
	if got := findNodeByKind(view.Tree.Root, "common", projecttree.NodeSchemaFile); got == nil {
		t.Fatalf("schema source was not discovered: %#v", view.Tree)
	}
	if got := findNode(view.Tree.Root, "ignored.cue"); got != nil {
		t.Fatalf("excluded source was discovered: %#v", got)
	}
	if len(view.Schemas.Definitions) != 1 || view.Schemas.Definitions[0].ID != "common.#GameID" {
		t.Fatalf("schema index = %#v", view.Schemas)
	}
}

func TestOpenProjectIndexesSchemaDiagnosticsAndUsages(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{
  "version": 1,
  "tableRoots": ["tables"],
  "schemaRoots": ["schema"],
  "export": {"directory": "exports", "targets": ["csv"]}
}`)
	writeFile(t, filepath.Join(projectRoot, "schema", "common.cue"), `package common
#GameID: string
#Positive: >0
_schema: {definitions: {
  GameID: {kind: "type", type: "string"}
  Positive: {kind: "constraint", appliesTo: ["integer", "number"]}
}}`)
	writeFile(t, filepath.Join(projectRoot, "tables", "item.cue"), `package item
#PrimaryKey: common.#GameID
#Row: {price: string & common.#Positive}
rows: {sword: {price: "invalid"}}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	view, err := NewService(t.TempDir()).OpenProject(filepath.Join(projectRoot, "game.dtproj"))
	if err != nil {
		t.Fatalf("OpenProject() error = %v", err)
	}
	if len(view.Schemas.Definitions) != 2 {
		t.Fatalf("definitions = %#v", view.Schemas.Definitions)
	}
	if len(view.Schemas.Diagnostics) != 1 || view.Schemas.Diagnostics[0].Field != "price" {
		t.Fatalf("diagnostics = %#v", view.Schemas.Diagnostics)
	}
	if got := view.Schemas.Definitions[0].Usages; len(got) != 1 || got[0].Field != "#PrimaryKey" {
		t.Fatalf("GameID usages = %#v", got)
	}
}

func TestLoadEditorSettingsUsesDefaultsWhenFileIsMissing(t *testing.T) {
	settings, err := LoadEditorSettings(t.TempDir())
	if err != nil {
		t.Fatalf("LoadEditorSettings() error = %v", err)
	}
	if settings != DefaultEditorSettings() {
		t.Fatalf("settings = %#v, want %#v", settings, DefaultEditorSettings())
	}
}

func TestLoadProjectRejectsInvalidConfig(t *testing.T) {
	projectRoot := t.TempDir()
	path := filepath.Join(projectRoot, "invalid.dtproj")
	writeFile(t, path, `{
  "version": 1,
  "tableRoots": ["../outside"],
  "schemaRoots": ["schema"],
  "export": {"directory": "exports", "targets": ["csv"]}
}`)

	_, err := LoadProject(path)
	if err == nil || !strings.Contains(err.Error(), "escapes the project directory") {
		t.Fatalf("LoadProject() error = %v, want path escape error", err)
	}
}

func TestGetTableLoadsOnlyDiscoveredProjectTables(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	path := filepath.Join(projectRoot, "tables", "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	table, err := service.GetTable(path)
	if err != nil || table.Rows[0].Values["name"].Display != "Sword" {
		t.Fatalf("GetTable() = %#v, %v", table, err)
	}
	if _, err := service.GetTable(filepath.Join(projectRoot, "outside.cue")); err == nil {
		t.Fatal("GetTable() accepted a path outside the project tree")
	}
}

func TestGetTableIncludesFormulaPreview(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	path := filepath.Join(projectRoot, "tables", "growth.cue")
	writeFile(t, path, `package growth
#PrimaryKey: string
#Row: {baseAttack: int, attack: int}
rows: {level_01: {baseAttack: 10}}
_tool: {table: {namespace: "combat", name: "growth"}, formulas: {level_01: {attack: "baseAttack * 2"}}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	table, err := service.GetTable(path)
	if err != nil || table.Rows[0].Values["attack"].Display != "20" || !table.Rows[0].Values["attack"].Formula {
		t.Fatalf("GetTable() = %#v, %v", table, err)
	}
}

func TestGetTableReportsSchemaErrorsAtTheirCells(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	path := filepath.Join(projectRoot, "tables", "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {price: int & >=0, rarity: "N" | "R"}
rows: {sword: {price: -1, rarity: "X"}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	table, err := service.GetTable(path)
	if err != nil {
		t.Fatal(err)
	}
	if !hasCellDiagnostic(table, "sword", "price") || !hasCellDiagnostic(table, "sword", "rarity") {
		t.Fatalf("diagnostics = %#v", table.Diagnostics)
	}
}

func hasCellDiagnostic(table document.Table, rowID, column string) bool {
	for _, diagnostic := range table.Diagnostics {
		if diagnostic.RowID == rowID && diagnostic.Column == column {
			return true
		}
	}
	return false
}

func TestUpdateCellRefreshesCurrentWorkspace(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	path := filepath.Join(projectRoot, "tables", "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	result, err := service.UpdateCell(CellUpdate{Path: path, RowID: "sword", Column: "name", Expression: `"Steel Sword"`})
	if err != nil {
		t.Fatal(err)
	}
	if got := result.Table.Rows[0].Values["name"].Display; got != "Steel Sword" {
		t.Fatalf("updated display = %q", got)
	}
	current, exists := service.Current()
	if !exists || current.Tree.Root.Path != result.Workspace.Tree.Root.Path {
		t.Fatalf("current workspace = %#v, exists = %v", current, exists)
	}
}

func TestCreateRowWithDiagnosticsPreservesNavigationSnapshot(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	path := filepath.Join(projectRoot, "tables", "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {price: int & >=0}
rows: {}
_tool: {table: {namespace: "rpg", name: "item"}, foreignKeys: {}, formulas: {}}
_editor: {columns: {}}`)
	writeFile(t, filepath.Join(projectRoot, "schema", "common.cue"), `package common
#GameID: string
_schema: {version: 1, definitions: {GameID: {kind: "type", type: "string"}}}`)
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	result, err := service.CreateRow(RowCreate{Path: path, RowID: "sword", Values: map[string]string{"price": "-1"}, Formulas: map[string]string{}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Table.Diagnostics) == 0 || result.Workspace.Tree.Root.Name == "" || findNodeByKind(result.Workspace.Tree.Root, "rpg.item", projecttree.NodeTable) == nil || len(result.Workspace.Schemas.Definitions) != 1 {
		t.Fatalf("CreateRow() = %#v", result)
	}
}

func TestAddRowRefreshesCurrentWorkspace(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	path := filepath.Join(projectRoot, "tables", "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	result, err := service.AddRow(RowAdd{Path: path, RowID: "potion", Expression: `{name: "Potion"}`})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Table.Rows) != 2 || result.Table.Rows[1].ID != "potion" {
		t.Fatalf("rows = %#v", result.Table.Rows)
	}
}

func TestDeleteRowRefreshesCurrentWorkspace(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	path := filepath.Join(projectRoot, "tables", "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {name: string}
rows: {sword: {name: "Sword"}, potion: {name: "Potion"}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	result, err := service.DeleteRow(RowDelete{Path: path, RowID: "sword"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Table.Rows) != 1 || result.Table.Rows[0].ID != "potion" {
		t.Fatalf("rows = %#v", result.Table.Rows)
	}
}

func TestDeleteRowRejectsForeignKeyReference(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	writeFile(t, filepath.Join(projectRoot, "tables", "effect.cue"), `package effect
#PrimaryKey: string
#Row: {name: string}
rows: {burn: {name: "Burn"}}
_tool: {table: {namespace: "rpg", name: "effect"}, foreignKeys: {}}`)
	path := filepath.Join(projectRoot, "tables", "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {effectIDs: [...string]}
rows: {sword: {effectIDs: ["burn"]}}
_tool: {table: {namespace: "rpg", name: "item"}, foreignKeys: {effectIDs: {table: "rpg.effect"}}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	if _, err := service.DeleteRow(RowDelete{Path: filepath.Join(projectRoot, "tables", "effect.cue"), RowID: "burn"}); err == nil {
		t.Fatal("DeleteRow() accepted a referenced target row")
	}
}

func TestOpenProjectReturnsForeignKeyDiagnostics(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	writeFile(t, filepath.Join(projectRoot, "tables", "item.cue"), `package item
#PrimaryKey: string
#Row: {effectID: string}
rows: {sword: {effectID: "missing"}}
_tool: {table: {namespace: "rpg", name: "item"}, foreignKeys: {effectID: {table: "rpg.effect"}}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	view, err := NewService(t.TempDir()).OpenProject(filepath.Join(projectRoot, "game.dtproj"))
	if err != nil {
		t.Fatal(err)
	}
	if len(view.ForeignKeys.Diagnostics) != 1 || view.ForeignKeys.Diagnostics[0].Column != "effectID" {
		t.Fatalf("foreign key diagnostics = %#v", view.ForeignKeys.Diagnostics)
	}
}

func TestExportWritesConfiguredTablesAndRejectsForeignKeyErrors(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	writeFile(t, filepath.Join(projectRoot, "tables", "item.cue"), `package item
#PrimaryKey: string
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}, foreignKeys: {}}
_export: {csv: {enabled: true, columns: ["id", "name"]}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	result, err := service.Export()
	if err != nil || len(result.Files) != 1 {
		t.Fatalf("Export() = %#v, %v", result, err)
	}

	writeFile(t, filepath.Join(projectRoot, "tables", "item.cue"), `package item
#PrimaryKey: string
#Row: {effectID: string}
rows: {sword: {effectID: "missing"}}
_tool: {table: {namespace: "rpg", name: "item"}, foreignKeys: {effectID: {table: "rpg.effect"}}}
_export: {csv: {enabled: true, columns: ["id", "effectID"]}}`)
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Export(); err == nil {
		t.Fatal("Export() accepted foreign key diagnostics")
	}
}

func TestAddColumnRefreshesCurrentWorkspace(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	path := filepath.Join(projectRoot, "tables", "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	result, err := service.AddColumn(ColumnAdd{Path: path, Name: "price", SchemaExpression: "int", InitialExpression: "10"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Table.Columns) != 2 || result.Table.Rows[0].Values["price"].Expression != "10" {
		t.Fatalf("table = %#v", result.Table)
	}
}

func TestDeleteColumnRefreshesCurrentWorkspace(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	path := filepath.Join(projectRoot, "tables", "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {name: string, price: int}
rows: {sword: {name: "Sword", price: 10}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	result, err := service.DeleteColumn(ColumnDelete{Path: path, Name: "price"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Table.Columns) != 1 || result.Table.Columns[0].Name != "name" {
		t.Fatalf("table = %#v", result.Table)
	}
}

func TestRenameColumnRefreshesCurrentWorkspace(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	path := filepath.Join(projectRoot, "tables", "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {price: int}
rows: {sword: {price: 10}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	result, err := service.RenameColumn(ColumnRename{Path: path, OldName: "price", NewName: "cost"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Table.Columns[0].Name != "cost" || result.Table.Rows[0].Values["cost"].Expression != "10" {
		t.Fatalf("table = %#v", result.Table)
	}
}

func TestMoveColumnRefreshesCurrentWorkspace(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	path := filepath.Join(projectRoot, "tables", "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {name: string, price: int}
rows: {sword: {name: "Sword", price: 10}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	result, err := service.MoveColumn(ColumnMove{Path: path, Name: "price", Position: 0})
	if err != nil {
		t.Fatal(err)
	}
	if result.Table.Columns[0].Name != "price" {
		t.Fatalf("table = %#v", result.Table)
	}
}

func TestPasteUndoAndRedo(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	path := filepath.Join(projectRoot, "tables", "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {name: string, price: int}
rows: {sword: {name: "Sword", price: 10}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	pasted, err := service.Paste(Paste{Path: path, RowIDs: []string{"sword"}, Columns: []string{"name", "price"}, Values: [][]string{{`"Steel Sword"`, "=20 + 5"}}})
	if err != nil || !pasted.CanUndo || pasted.CanRedo || pasted.Table.Rows[0].Values["name"].Expression != `"Steel Sword"` {
		t.Fatalf("Paste() = %#v, %v", pasted, err)
	}
	undone, err := service.Undo(path)
	if err != nil || undone.CanUndo || !undone.CanRedo || undone.Table.Rows[0].Values["name"].Expression != `"Sword"` {
		t.Fatalf("Undo() = %#v, %v", undone, err)
	}
	redone, err := service.Redo(path)
	if err != nil || !redone.CanUndo || redone.CanRedo || redone.Table.Rows[0].Values["name"].Expression != `"Steel Sword"` {
		t.Fatalf("Redo() = %#v, %v", redone, err)
	}
}

func TestFormulaAndClearCellAreUndoable(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	path := filepath.Join(projectRoot, "tables", "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {price: int}
rows: {sword: {price: 10}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetCellFormula(FormulaSet{Path: path, RowID: "sword", Column: "price", Expression: "20 + 5"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ClearCell(CellUpdate{Path: path, RowID: "sword", Column: "price"}); err != nil {
		t.Fatal(err)
	}
	undone, err := service.Undo(path)
	if err != nil || formulaExpressionForWorkspace(undone.Table, "sword", "price") != "20 + 5" {
		t.Fatalf("Undo() = %#v, %v", undone, err)
	}
}

func formulaExpressionForWorkspace(table document.Table, rowID, column string) string {
	for _, entry := range table.Formulas {
		if entry.RowID == rowID && entry.Column == column {
			return entry.Expression
		}
	}
	return ""
}

func TestRenameRowUpdatesForeignKeyReferences(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	path := filepath.Join(projectRoot, "tables", "effect.cue")
	writeFile(t, path, `package effect
#PrimaryKey: string
#Row: {name: string}
rows: {burn: {name: "Burn"}}
_tool: {table: {namespace: "rpg", name: "effect"}, foreignKeys: {}}`)
	writeFile(t, filepath.Join(projectRoot, "tables", "item.cue"), `package item
#PrimaryKey: string
#Row: {effectIDs: [...string]}
rows: {sword: {effectIDs: ["burn"]}}
_tool: {table: {namespace: "rpg", name: "item"}, foreignKeys: {effectIDs: {table: "rpg.effect"}}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	result, err := service.RenameRow(RowRename{Path: path, OldID: "burn", NewID: "burn_large"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Table.Rows[0].ID != "burn_large" {
		t.Fatalf("rows = %#v", result.Table.Rows)
	}
	item, err := service.GetTable(filepath.Join(projectRoot, "tables", "item.cue"))
	if err != nil || item.Rows[0].Values["effectIDs"].Expression != `["burn_large"]` {
		t.Fatalf("item = %#v, error = %v", item, err)
	}
}

func TestGetTableIncludesScalarForeignKeyOptions(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	writeFile(t, filepath.Join(projectRoot, "tables", "effect.cue"), `package effect
#PrimaryKey: string
#Row: {name: string}
rows: {burn: {name: "Burn"}, freeze: {name: "Freeze"}}
_tool: {table: {namespace: "rpg", name: "effect"}}`)
	path := filepath.Join(projectRoot, "tables", "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {effectID: string, effectIDs: [...string]}
rows: {sword: {effectID: "burn", effectIDs: ["burn"]}}
_tool: {table: {namespace: "rpg", name: "item"}, foreignKeys: {effectID: {table: "rpg.effect"}, effectIDs: {table: "rpg.effect"}}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	table, err := service.GetTable(path)
	if err != nil {
		t.Fatal(err)
	}
	if table.Columns[0].ForeignKeyTable != "rpg.effect" || !reflect.DeepEqual(table.Columns[0].ForeignKeyOptions, []string{"burn", "freeze"}) || table.Columns[1].ForeignKeyTable != "" {
		t.Fatalf("columns = %#v", table.Columns)
	}
}

func TestSetColumnSchemaRefreshesDiagnostics(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	path := filepath.Join(projectRoot, "tables", "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {price: int}
rows: {sword: {price: 10}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	result, err := service.SetColumnSchema(ColumnSchemaSet{Path: path, Name: "price", SchemaExpression: "int & >=20"})
	if err != nil || len(result.Table.Diagnostics) == 0 || result.Table.Diagnostics[0].Column != "price" {
		t.Fatalf("SetColumnSchema() = %#v, %v", result, err)
	}
}

func TestCreateTableAddsItToTheCurrentProject(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "tables"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	result, err := service.CreateTable(TableCreate{Namespace: "combat", Name: "skill"})
	if err != nil || result.Table.Namespace != "combat" || result.Table.Name != "skill" || !containsTable(result.Workspace.Tree.Root, result.Table.Path) {
		t.Fatalf("CreateTable() = %#v, %v", result, err)
	}
	if _, err := service.CreateTable(TableCreate{Namespace: "combat", Name: "skill"}); err == nil {
		t.Fatal("CreateTable() accepted duplicate table")
	}
}

func TestCreatePublicSchemaDefinitionRefreshesSchemaIndex(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	schemaPath := filepath.Join(projectRoot, "schema", "common.cue")
	writeFile(t, schemaPath, `package common
#GameID: string
_schema: {version: 1, definitions: {GameID: {kind: "type", type: "string"}}}`)
	writeFile(t, filepath.Join(projectRoot, "tables", "item.cue"), `package item
#PrimaryKey: common.#GameID
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	result, err := service.CreateSchemaDefinition(SchemaDefinitionCreate{Path: schemaPath, Name: "Rarity", Kind: "type", Type: "enum", SchemaExpression: `"N" | "R"`})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Schema.Definitions) != 2 || result.Schema.Definitions[1].Name != "Rarity" {
		t.Fatalf("schema = %#v", result.Schema)
	}
	found := false
	for _, definition := range result.Workspace.Schemas.Definitions {
		if definition.ID == "common.#Rarity" {
			found = true
		}
	}
	if !found {
		t.Fatalf("schema index = %#v", result.Workspace.Schemas.Definitions)
	}
}

func TestCreateLocalSchemaDefinitionRefreshesSchemaIndex(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	tablePath := filepath.Join(projectRoot, "tables", "item.cue")
	writeFile(t, tablePath, `package item
#PrimaryKey: string
#Row: {position: _local.#Position}
rows: {sword: {position: {x: 0}}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}

	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	result, err := service.CreateSchemaDefinition(SchemaDefinitionCreate{Path: tablePath, Name: "Position", Kind: "type", Type: "struct", SchemaExpression: `{x: number}`})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, definition := range result.Workspace.Schemas.Definitions {
		if definition.ID == "rpg.item._local.#Position" {
			found = true
		}
	}
	if !found {
		t.Fatalf("schema index = %#v", result.Workspace.Schemas.Definitions)
	}
}

func TestDeleteSchemaDefinitionBlockedWhenReferenced(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	schemaPath := filepath.Join(projectRoot, "schema", "common.cue")
	writeFile(t, schemaPath, `package common
#GameID: string
_schema: {version: 1, definitions: {GameID: {kind: "type", type: "string"}}}`)
	writeFile(t, filepath.Join(projectRoot, "tables", "item.cue"), `package item
#PrimaryKey: common.#GameID
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	_, err := service.DeleteSchemaDefinition(SchemaDefinitionDelete{Path: schemaPath, Name: "GameID"})
	if err == nil || !strings.Contains(err.Error(), "referenced") {
		t.Fatalf("DeleteSchemaDefinition() = %v, want referenced error", err)
	}
}

func TestRenameSchemaDefinitionUpdatesReferencingTable(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	schemaPath := filepath.Join(projectRoot, "schema", "common.cue")
	writeFile(t, schemaPath, `package common
#GameID: string
_schema: {version: 1, definitions: {GameID: {kind: "type", type: "string"}}}`)
	tablePath := filepath.Join(projectRoot, "tables", "item.cue")
	writeFile(t, tablePath, `package item
#PrimaryKey: common.#GameID
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	result, err := service.RenameSchemaDefinition(SchemaDefinitionRename{Path: schemaPath, OldName: "GameID", NewName: "GameIdentifier"})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, definition := range result.Workspace.Schemas.Definitions {
		if definition.ID == "common.#GameIdentifier" {
			found = true
		}
	}
	if !found {
		t.Fatalf("schema index = %#v", result.Workspace.Schemas.Definitions)
	}
	table, err := service.GetTable(tablePath)
	if err != nil {
		t.Fatal(err)
	}
	if table.PrimaryKey.Reference != "common.#GameIdentifier" {
		t.Fatalf("primary key = %#v", table.PrimaryKey)
	}
}

func TestGetPublicSchemaLoadsOnlyProjectSchemaFiles(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	schemaPath := filepath.Join(projectRoot, "schema", "common.cue")
	writeFile(t, schemaPath, `package common
#GameID: string
_schema: {version: 1, definitions: {GameID: {kind: "type", type: "string"}}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "tables"), 0o755); err != nil {
		t.Fatal(err)
	}

	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	schema, err := service.GetPublicSchema(schemaPath)
	if err != nil || len(schema.Definitions) != 1 || schema.Definitions[0].SchemaExpression != "string" {
		t.Fatalf("GetPublicSchema() = %#v, %v", schema, err)
	}
	if _, err := service.GetPublicSchema(filepath.Join(projectRoot, "outside.cue")); err == nil {
		t.Fatal("GetPublicSchema() accepted a path outside the project tree")
	}
}

func TestSetColumnSchemaNormalizesLocalSchemaReference(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	tablePath := filepath.Join(projectRoot, "tables", "item.cue")
	writeFile(t, tablePath, `package item
#PrimaryKey: string
_local: {#tmpDamageBase: int}
_schema: {localDefinitions: {tmpDamageBase: {kind: "type", type: "integer"}}}
#Row: {damage: int}
rows: {sword: {damage: 10}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}

	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	result, err := service.SetColumnSchema(ColumnSchemaSet{Path: tablePath, Name: "damage", SchemaExpression: "#tmpDamageBase"})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, column := range result.Table.Columns {
		if column.Name == "damage" && column.SchemaExpression == "_local.#tmpDamageBase" {
			found = true
		}
	}
	if !found {
		t.Fatalf("columns = %#v", result.Table.Columns)
	}
	for _, diagnostic := range result.Workspace.Schemas.Diagnostics {
		if diagnostic.Field == "damage" {
			t.Fatalf("diagnostics = %#v", result.Workspace.Schemas.Diagnostics)
		}
	}
}

func TestSetColumnSchemaRejectsUnknownBareReference(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	tablePath := filepath.Join(projectRoot, "tables", "item.cue")
	writeFile(t, tablePath, `package item
#PrimaryKey: string
#Row: {damage: int}
rows: {sword: {damage: 10}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}

	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	_, err := service.SetColumnSchema(ColumnSchemaSet{Path: tablePath, Name: "damage", SchemaExpression: "#Missing"})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("SetColumnSchema() = %v, want not found error", err)
	}
}

func TestSetColumnSchemaAddsSchemaImport(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	writeFile(t, filepath.Join(projectRoot, "cue.mod", "module.cue"), `module: "example.com/game"`)
	writeFile(t, filepath.Join(projectRoot, "schema", "common.cue"), `package common
#DamageBaseValue: int & >=0
_schema: {version: 1, definitions: {DamageBaseValue: {kind: "type", type: "integer"}}}`)
	tablePath := filepath.Join(projectRoot, "tables", "item.cue")
	writeFile(t, tablePath, `package item
#PrimaryKey: string
#Row: {damage: int}
rows: {sword: {damage: 10}}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	result, err := service.SetColumnSchema(ColumnSchemaSet{Path: tablePath, Name: "damage", SchemaExpression: "common.#DamageBaseValue"})
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(tablePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(source), `import "example.com/game/schema:common"`) {
		t.Fatalf("source = %s", source)
	}
	for _, diagnostic := range result.Table.Diagnostics {
		if strings.Contains(diagnostic.Message, "reference") {
			t.Fatalf("table diagnostics = %#v", result.Table.Diagnostics)
		}
	}
}

func TestSetSchemaDefinitionAddsCrossPackageImport(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	writeFile(t, filepath.Join(projectRoot, "cue.mod", "module.cue"), `module: "example.com/game"`)
	writeFile(t, filepath.Join(projectRoot, "schema", "common.cue"), `package common
#DamageBaseValue: int & >=0
_schema: {version: 1, definitions: {DamageBaseValue: {kind: "type", type: "integer"}}}`)
	rulesPath := filepath.Join(projectRoot, "schema", "rules.cue")
	writeFile(t, rulesPath, `package rules
#Damage: int
_schema: {version: 1, definitions: {Damage: {kind: "type", type: "integer"}}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "tables"), 0o755); err != nil {
		t.Fatal(err)
	}

	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetSchemaDefinition(SchemaDefinitionSet{Path: rulesPath, Name: "Damage", Kind: "type", Type: "integer", SchemaExpression: "common.#DamageBaseValue"}); err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(rulesPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(source), `import "example.com/game/schema:common"`) {
		t.Fatalf("source = %s", source)
	}
}

func TestSetColumnSchemaRemovesUnusedImport(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	writeFile(t, filepath.Join(projectRoot, "cue.mod", "module.cue"), `module: "example.com/game"`)
	writeFile(t, filepath.Join(projectRoot, "schema", "common.cue"), `package common
#DamageBaseValue: int & >=0
_schema: {version: 1, definitions: {DamageBaseValue: {kind: "type", type: "integer"}}}`)
	tablePath := filepath.Join(projectRoot, "tables", "item.cue")
	writeFile(t, tablePath, `package item
import "example.com/game/schema:common"
#PrimaryKey: string
#Row: {damage: common.#DamageBaseValue}
rows: {sword: {damage: 10}}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	service := NewService(t.TempDir())
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	result, err := service.SetColumnSchema(ColumnSchemaSet{Path: tablePath, Name: "damage", SchemaExpression: "int"})
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(tablePath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(source), "import") {
		t.Fatalf("unused import was not pruned: %s", source)
	}
	for _, diagnostic := range result.Table.Diagnostics {
		if strings.Contains(diagnostic.Message, "imported and not used") {
			t.Fatalf("table diagnostics = %#v", result.Table.Diagnostics)
		}
	}
}

func TestOpenProjectCompilesWorkspace(t *testing.T) {
	projectRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	writeFile(t, filepath.Join(projectRoot, "tables", "item.cue"), `package item
#PrimaryKey: string
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(t.TempDir())
	view, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj"))
	if err != nil {
		t.Fatal(err)
	}
	if view.Compile.Mode != "auto" || view.Compile.Pending || len(view.Compile.Diagnostics) != 0 || view.Compile.CompiledAt == "" {
		t.Fatalf("compile state = %#v", view.Compile)
	}
}

func TestManualModeDefersCompileUntilRequested(t *testing.T) {
	projectRoot := t.TempDir()
	configRoot := t.TempDir()
	writeFile(t, SettingsPath(configRoot), `{
  "version": 1,
  "build": {"mode": "manual"}
}`)
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	path := filepath.Join(projectRoot, "tables", "item.cue")
	writeFile(t, path, `package item
#PrimaryKey: string
#Row: {price: int & >=0}
rows: {sword: {price: 10}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(configRoot)
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	result, err := service.UpdateCell(CellUpdate{Path: path, RowID: "sword", Column: "price", Expression: "-1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Table.Diagnostics) != 0 {
		t.Fatalf("manual mode evaluated diagnostics on edit: %#v", result.Table.Diagnostics)
	}
	if !result.Workspace.Compile.Pending {
		t.Fatalf("compile state = %#v", result.Workspace.Compile)
	}
	compiled, err := service.Compile()
	if err != nil {
		t.Fatal(err)
	}
	if compiled.Workspace.Compile.Pending || len(compiled.Workspace.Compile.Diagnostics) == 0 {
		t.Fatalf("compile state = %#v", compiled.Workspace.Compile)
	}
	table, err := service.GetTable(path)
	if err != nil {
		t.Fatal(err)
	}
	if !hasCellDiagnostic(table, "sword", "price") {
		t.Fatalf("compiled diagnostics = %#v", table.Diagnostics)
	}
	if _, err := service.Export(); err == nil {
		t.Fatal("Export() accepted compile diagnostics in manual mode")
	}
}

func TestSetCompileModePersistsSettings(t *testing.T) {
	projectRoot := t.TempDir()
	configRoot := t.TempDir()
	writeFile(t, filepath.Join(projectRoot, "game.dtproj"), `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	writeFile(t, filepath.Join(projectRoot, "tables", "item.cue"), `package item
#PrimaryKey: string
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	if err := os.MkdirAll(filepath.Join(projectRoot, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}
	service := NewService(configRoot)
	if _, err := service.OpenProject(filepath.Join(projectRoot, "game.dtproj")); err != nil {
		t.Fatal(err)
	}
	view, err := service.SetCompileMode("manual")
	if err != nil {
		t.Fatal(err)
	}
	if view.Compile.Mode != "manual" || view.Settings.Build.Mode != "manual" {
		t.Fatalf("view = %#v", view)
	}
	settings, err := LoadEditorSettings(configRoot)
	if err != nil {
		t.Fatal(err)
	}
	if settings.Build.Mode != "manual" {
		t.Fatalf("settings = %#v", settings)
	}
	if _, err := service.SetCompileMode("invalid"); err == nil {
		t.Fatal("SetCompileMode() accepted an invalid mode")
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%q): %v", path, err)
	}
}

func findNode(node projecttree.Node, name string) *projecttree.Node {
	if node.Name == name {
		return &node
	}
	for _, child := range node.Children {
		if found := findNode(child, name); found != nil {
			return found
		}
	}
	return nil
}

func findNodeByKind(node projecttree.Node, name string, kind projecttree.NodeKind) *projecttree.Node {
	if node.Name == name && node.Kind == kind {
		return &node
	}
	for _, child := range node.Children {
		if found := findNodeByKind(child, name, kind); found != nil {
			return found
		}
	}
	return nil
}
