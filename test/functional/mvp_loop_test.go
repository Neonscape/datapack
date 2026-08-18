package functional

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"datapack/internal/workspace"
)

func TestEmptyProjectCreateTableFormulaAndExport(t *testing.T) {
	root, projectPath := emptyProject(t)
	// CreateSchemaFile writes an empty package. The project tree only indexes
	// public Schema files that already have definitions, so seed one type first.
	schemaPath := filepath.Join(root, "schema", "common.cue")
	writeFile(t, schemaPath, `package common
#GameID: string
_schema: {version: 1, definitions: {GameID: {kind: "type", type: "string"}}}
`)
	service := openProject(t, projectPath)
	if _, err := service.CreateSchemaDefinition(workspace.SchemaDefinitionCreate{
		Path:             schemaPath,
		Name:             "NonNegative",
		Kind:             "constraint",
		AppliesTo:        []string{"integer", "number"},
		SchemaExpression: ">=0",
	}); err != nil {
		t.Fatalf("CreateSchemaDefinition(): %v", err)
	}

	created, err := service.CreateTable(workspace.TableCreate{Namespace: "combat", Name: "growth"})
	if err != nil {
		t.Fatalf("CreateTable(): %v", err)
	}
	path := created.Table.Path

	if _, err := service.AddColumn(workspace.ColumnAdd{Path: path, Name: "baseAttack", SchemaExpression: "int & common.#NonNegative", InitialExpression: "0"}); err != nil {
		t.Fatalf("AddColumn(baseAttack): %v", err)
	}
	if _, err := service.AddColumn(workspace.ColumnAdd{Path: path, Name: "attack", SchemaExpression: "int & common.#NonNegative", InitialExpression: "0"}); err != nil {
		t.Fatalf("AddColumn(attack): %v", err)
	}
	if _, err := service.CreateRow(workspace.RowCreate{
		Path:     path,
		RowID:    "level_01",
		Values:   map[string]string{"baseAttack": "10"},
		Formulas: map[string]string{"attack": "baseAttack * 2"},
	}); err != nil {
		t.Fatalf("CreateRow(): %v", err)
	}

	enableCSVExport(t, path, []string{"id", "baseAttack", "attack"})
	if _, err := service.OpenProject(projectPath); err != nil {
		t.Fatal(err)
	}

	result, err := service.Export()
	if err != nil {
		view, _ := service.Current()
		t.Fatalf("Export(): %v\n%s", err, formatDiagnostics(view.Compile.Diagnostics))
	}
	got := readFile(t, result.Files[0])
	want := "id,baseAttack,attack\nlevel_01,10,20\n"
	if got != want {
		t.Fatalf("CSV = %q, want %q", got, want)
	}

	if _, err := service.UpdateCell(workspace.CellUpdate{Path: path, RowID: "level_01", Column: "baseAttack", Expression: "-1"}); err != nil {
		t.Fatalf("UpdateCell(-1): %v", err)
	}
	if _, err := service.Export(); err == nil {
		t.Fatal("Export() accepted a negative baseAttack")
	}
	if readFile(t, result.Files[0]) != want {
		t.Fatal("blocked export replaced the previous CSV")
	}
	if source := readFile(t, path); !strings.Contains(source, "baseAttack: -1") {
		t.Fatalf("business error was not saved:\n%s", source)
	}
}

func TestForeignKeyProjectExportsUntilReferenceBreaks(t *testing.T) {
	root := t.TempDir()
	projectPath := filepath.Join(root, "game.dtproj")
	writeFile(t, projectPath, `{"version": 1, "tableRoots": ["tables"], "schemaRoots": ["schema"], "export": {"directory": "exports", "targets": ["csv"]}}`)
	writeFile(t, filepath.Join(root, "tables", "effect", "effect.cue"), `package effect
#PrimaryKey: string
#Row: {name: string}
rows: {burn_small: {name: "Burn"}}
_tool: {table: {namespace: "rpg", name: "effect"}, foreignKeys: {}, formulas: {}}
_export: {csv: {enabled: true, columns: ["id", "name"]}}
`)
	itemPath := filepath.Join(root, "tables", "item", "item.cue")
	writeFile(t, itemPath, `package item
#PrimaryKey: string
#Row: {effectIds: [...string]}
rows: {sword: {effectIds: ["burn_small"]}}
_tool: {
	table: {namespace: "rpg", name: "item"}
	foreignKeys: {effectIds: {table: "rpg.effect"}}
	formulas: {}
}
_export: {csv: {enabled: true, columns: ["id", "effectIds"]}}
`)
	if err := os.MkdirAll(filepath.Join(root, "schema"), 0o755); err != nil {
		t.Fatal(err)
	}

	service := openProject(t, projectPath)
	result, err := service.Export()
	if err != nil {
		view, _ := service.Current()
		t.Fatalf("Export(): %v\n%s", err, formatDiagnostics(view.Compile.Diagnostics))
	}
	itemCSV := ""
	for _, file := range result.Files {
		if filepath.Base(file) == "rpg.item.csv" {
			itemCSV = readFile(t, file)
		}
	}
	want := "id,effectIds\nsword,\"[\"\"burn_small\"\"]\"\n"
	if itemCSV != want {
		t.Fatalf("item CSV = %q, want %q", itemCSV, want)
	}

	if _, err := service.UpdateCell(workspace.CellUpdate{Path: itemPath, RowID: "sword", Column: "effectIds", Expression: `["missing"]`}); err != nil {
		t.Fatalf("UpdateCell(): %v", err)
	}
	if _, err := service.Export(); err == nil {
		t.Fatal("Export() accepted a missing array foreign key")
	}
	if readFile(t, filepath.Join(result.Directory, "rpg.item.csv")) != want {
		t.Fatal("blocked export replaced the previous CSV")
	}
}
