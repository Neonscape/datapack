package document

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"datapack/internal/schemamodel"
)

func TestLoadTableBuildsEditableDocument(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
_local: {#Position: {x: number, y: number}}
_schema: {localDefinitions: {Position: {kind: "type", type: "struct"}}}
#Row: {
    name: string & != ""
    tags: [...string]
    position: _local.#Position
}
rows: {sword: {name: "Sword", tags: ["weapon"], position: {x: 0, y: 0}}}
_tool: {table: {namespace: "rpg", name: "item"}}
`)

	document, err := LoadTable(path)
	if err != nil {
		t.Fatal(err)
	}
	if document.Namespace != "rpg" || document.Name != "item" || len(document.Columns) != 3 || len(document.Rows) != 1 {
		t.Fatalf("document = %#v", document)
	}
	if document.Columns[1].Schema.Kind != schemamodel.KindArray {
		t.Fatalf("tags schema = %#v", document.Columns[1].Schema)
	}
	if len(document.LocalSchema) != 1 || document.LocalSchema[0].Name != "Position" || document.LocalSchema[0].Kind != "type" {
		t.Fatalf("local schema = %#v", document.LocalSchema)
	}
	if got := document.Rows[0].Values["name"]; got.Expression != `"Sword"` || got.Display != "Sword" {
		t.Fatalf("name cell = %#v", got)
	}
	if got := document.Rows[0].Values["tags"]; got.Expression != `["weapon"]` || got.Display != `["weapon"]` {
		t.Fatalf("tags cell = %#v", got)
	}
}

func TestLoadPublicSchemaBuildsDefinitions(t *testing.T) {
	path := writeSource(t, "common.cue", `package common
#GameID: string & != ""
#NonNegative: >=0
_schema: {definitions: {
    GameID: {kind: "type", type: "string"}
    NonNegative: {kind: "constraint", appliesTo: ["integer", "number"]}
}}`)

	document, err := LoadPublicSchema(path)
	if err != nil {
		t.Fatal(err)
	}
	if document.Package != "common" || len(document.Definitions) != 2 || document.Definitions[0].Name != "GameID" {
		t.Fatalf("document = %#v", document)
	}
}

func TestLoadTableReadsEnabledCSVExport(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}}
_export: {csv: {enabled: true, columns: ["id", "name"]}}`)

	table, err := LoadTable(path)
	if err != nil || table.CSVExport == nil || len(table.CSVExport.Columns) != 2 {
		t.Fatalf("CSV export = %#v, error = %v", table.CSVExport, err)
	}
}

func TestLoadTableReadsFormulas(t *testing.T) {
	path := writeSource(t, "growth.cue", `package growth
#PrimaryKey: string
#Row: {baseAttack: int, attack: int}
rows: {level_01: {baseAttack: 10}}
_tool: {
    table: {namespace: "combat", name: "growth"}
    formulas: {level_01: {attack: "baseAttack * 2"}}
}`)

	table, err := LoadTable(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(table.Formulas) != 1 || table.Formulas[0] != (Formula{RowID: "level_01", Column: "attack", Expression: "baseAttack * 2"}) {
		t.Fatalf("formulas = %#v", table.Formulas)
	}
}

func TestSetCellValueReplacesExistingCellWithCueExpression(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
#Row: {name: string, price: int}
rows: {sword: {name: "Sword", price: 100}}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	table, err := SetCellValue(path, "sword", "price", "125")
	if err != nil {
		t.Fatal(err)
	}
	if got := table.Rows[0].Values["price"]; got.Expression != "125" || got.Display != "125" {
		t.Fatalf("price cell = %#v", got)
	}
	reloaded, err := LoadTable(path)
	if err != nil || reloaded.Rows[0].Values["price"].Expression != "125" {
		t.Fatalf("reloaded table = %#v, error = %v", reloaded, err)
	}
}

func TestSetCellFormulaReplacesFixedValueAndFixedValueRemovesFormula(t *testing.T) {
	path := writeSource(t, "growth.cue", `package growth
#PrimaryKey: string
#Row: {baseAttack: int, attack: int}
rows: {level_01: {baseAttack: 10, attack: 20}}
_tool: {table: {namespace: "combat", name: "growth"}}`)

	table, err := SetCellFormula(path, "level_01", "attack", "baseAttack * 2")
	if err != nil || len(table.Formulas) != 1 || table.Rows[0].Values["attack"].Expression != "" {
		t.Fatalf("SetCellFormula() = %#v, %v", table, err)
	}
	table, err = SetCellValue(path, "level_01", "attack", "25")
	if err != nil || len(table.Formulas) != 0 || table.Rows[0].Values["attack"].Expression != "25" {
		t.Fatalf("SetCellValue() = %#v, %v", table, err)
	}
}

func TestSetCellValueAllowsEditingFormulaDependencies(t *testing.T) {
	path := writeSource(t, "growth.cue", `package growth
#PrimaryKey: string
#Row: {baseAttack: int, attack: int}
rows: {level_01: {baseAttack: 10}}
_tool: {table: {namespace: "combat", name: "growth"}, formulas: {level_01: {attack: "baseAttack * 2"}}}`)

	table, err := SetCellValue(path, "level_01", "baseAttack", "15")
	if err != nil || table.Rows[0].Values["baseAttack"].Expression != "15" {
		t.Fatalf("SetCellValue() = %#v, %v", table, err)
	}
}

func TestSetCellValueRejectsInvalidTargetsWithoutWriting(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := SetCellValue(path, "sword", "missing", "1"); err == nil {
		t.Fatal("SetCellValue() accepted a missing column")
	}
	if _, err := SetCellValue(path, "sword", "name", "["); err == nil {
		t.Fatal("SetCellValue() accepted invalid CUE")
	}
	current, err := os.ReadFile(path)
	if err != nil || string(current) != string(original) {
		t.Fatalf("source changed after rejected update: %q, error = %v", current, err)
	}
}

func TestSetCellValuePersistsValuesOutsideColumnSchema(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
#Row: {price: int & >= 0, rarity: "N" | "R"}
rows: {sword: {price: 100, rarity: "N"}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	if _, err := SetCellValue(path, "sword", "price", "-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := SetCellValue(path, "sword", "rarity", `"X"`); err != nil {
		t.Fatal(err)
	}
	table, err := LoadTable(path)
	if err != nil || table.Rows[0].Values["price"].Expression != "-1" || table.Rows[0].Values["rarity"].Expression != `"X"` {
		t.Fatalf("table = %#v, error = %v", table, err)
	}
}

func TestSetCellValuePersistsValuesOutsideImportedSchema(t *testing.T) {
	root := t.TempDir()
	writeModuleFile(t, filepath.Join(root, "cue.mod", "module.cue"), `module: "example.com/game"
language: {version: "v0.9.0"}`)
	writeModuleFile(t, filepath.Join(root, "schema", "common.cue"), `package common
#NonNegative: >=0`)
	path := filepath.Join(root, "tables", "item.cue")
	writeModuleFile(t, path, `package item
import "example.com/game/schema:common"
#PrimaryKey: string
#Row: {price: int & common.#NonNegative}
rows: {sword: {price: 100}}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	if _, err := SetCellValue(path, "sword", "price", "-1"); err != nil {
		t.Fatal(err)
	}
}

func TestCreateRowPersistsPartialValuesAndClearCellRemovesEmptyRow(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
#Row: {name: string, price: int}
rows: {}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	table, err := CreateRow(path, "sword", map[string]string{"name": `"Sword"`}, nil)
	if err != nil || len(table.Rows) != 1 || table.Rows[0].Values["name"].Display != "Sword" {
		t.Fatalf("CreateRow() = %#v, %v", table, err)
	}
	table, err = ClearCell(path, "sword", "name")
	if err != nil || len(table.Rows) != 0 {
		t.Fatalf("ClearCell() = %#v, %v", table, err)
	}
}

func TestDeleteRowRemovesItsFormulas(t *testing.T) {
	path := writeSource(t, "growth.cue", `package growth
#PrimaryKey: string
#Row: {attack: int}
rows: {level_01: {}}
_tool: {table: {namespace: "combat", name: "growth"}, formulas: {level_01: {attack: "1"}}}`)

	table, err := DeleteRow(path, "level_01")
	if err != nil || len(table.Rows) != 0 || len(table.Formulas) != 0 {
		t.Fatalf("DeleteRow() = %#v, %v", table, err)
	}
}

func TestPasteCellsIsAtomicAndSupportsFormulas(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
#Row: {name: string, price: int}
rows: {sword: {name: "Sword", price: 10}, axe: {name: "Axe", price: 20}}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	table, err := PasteCells(path, []string{"sword", "axe"}, []string{"name", "price"}, [][]string{{`"Steel Sword"`, "=20 + 5"}, {`"Steel Axe"`, "30"}})
	if err != nil || table.Rows[0].Values["name"].Display != "Steel Sword" || formulaExpressionFor(table, "sword", "price") != "20 + 5" || table.Rows[1].Values["price"].Expression != "30" {
		t.Fatalf("PasteCells() = %#v, %v", table, err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := PasteCells(path, []string{"sword"}, []string{"name", "price"}, [][]string{{`"Valid"`, "["}}); err == nil {
		t.Fatal("PasteCells() accepted invalid CUE")
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(before) {
		t.Fatalf("source changed after invalid paste: %q, error = %v", after, err)
	}
}

func TestRenameRowUpdatesFormulaReferencesWithoutChangingStrings(t *testing.T) {
	path := writeSource(t, "growth.cue", `package growth
#PrimaryKey: string
#Row: {label: string, attack: int}
rows: {level_01: {}, level_02: {}}
_tool: {table: {namespace: "combat", name: "growth"}, formulas: {
    level_01: {attack: "1"}
    level_02: {label: `+strconv.Quote(`"level_01"`)+`, attack: "combat.growth.level_01.attack + 1"}
}}`)

	table, err := RenameRow(path, "level_01", "level_03", nil)
	if err != nil {
		t.Fatal(err)
	}
	if formulaExpressionFor(table, "level_02", "label") != `"level_01"` || formulaExpressionFor(table, "level_02", "attack") != "combat.growth.level_03.attack + 1" || formulaExpressionFor(table, "level_03", "attack") != "1" {
		t.Fatalf("formulas = %#v", table.Formulas)
	}
}

func TestRenameColumnUpdatesFormulaReferences(t *testing.T) {
	path := writeSource(t, "growth.cue", `package growth
#PrimaryKey: string
#Row: {baseAttack: int, attack: int}
rows: {level_01: {}, level_02: {}}
_tool: {table: {namespace: "combat", name: "growth"}, formulas: {
    level_01: {attack: "baseAttack * 2"}
    level_02: {attack: "combat.growth.level_01.attack + baseAttack"}
}}`)

	table, err := RenameColumn(path, "baseAttack", "power")
	if err != nil {
		t.Fatal(err)
	}
	if formulaExpressionFor(table, "level_01", "attack") != "power * 2" || formulaExpressionFor(table, "level_02", "attack") != "combat.growth.level_01.attack + power" {
		t.Fatalf("formulas = %#v", table.Formulas)
	}
}

func formulaExpressionFor(table Table, rowID, column string) string {
	for _, formula := range table.Formulas {
		if formula.RowID == rowID && formula.Column == column {
			return formula.Expression
		}
	}
	return ""
}

func TestAddRowInsertsAndValidatesCueStruct(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
#Row: {name: string, price: int & >=0}
rows: {sword: {name: "Sword", price: 100}}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	table, err := AddRow(path, "potion", `{name: "Potion", price: 10}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(table.Rows) != 2 || table.Rows[1].ID != "potion" || table.Rows[1].Values["name"].Display != "Potion" {
		t.Fatalf("table rows = %#v", table.Rows)
	}
	if _, err := AddRow(path, "potion", `{name: "Potion", price: 10}`); err == nil {
		t.Fatal("AddRow() accepted a duplicate row ID")
	}
	if _, err := AddRow(path, "broken", `{name: "Broken", price: -1}`); err == nil {
		t.Fatal("AddRow() accepted a row outside #Row")
	}
	if _, err := AddRow(path, "bad-id", `{name: "Broken", price: 1}`); err == nil {
		t.Fatal("AddRow() accepted an invalid row ID")
	}
}

func TestDeleteRowRemovesExistingRow(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
#Row: {name: string}
rows: {
    sword: {name: "Sword"}
    potion: {name: "Potion"}
}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	table, err := DeleteRow(path, "sword")
	if err != nil {
		t.Fatal(err)
	}
	if len(table.Rows) != 1 || table.Rows[0].ID != "potion" {
		t.Fatalf("table rows = %#v", table.Rows)
	}
	if _, err := DeleteRow(path, "sword"); err == nil {
		t.Fatal("DeleteRow() accepted a missing row")
	}
}

func TestRenameRowUpdatesScalarAndArrayReferences(t *testing.T) {
	root := t.TempDir()
	targetPath := filepath.Join(root, "effect.cue")
	referencePath := filepath.Join(root, "item.cue")
	writeModuleFile(t, targetPath, `package effect
#PrimaryKey: string
#Row: {name: string}
rows: {burn: {name: "Burn"}}
_tool: {table: {namespace: "rpg", name: "effect"}}`)
	writeModuleFile(t, referencePath, `package item
#PrimaryKey: string
#Row: {effectID: string, effectIDs: [...string]}
rows: {sword: {effectID: "burn", effectIDs: ["burn"]}}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	table, err := RenameRow(targetPath, "burn", "burn_large", []RowReference{
		{Path: referencePath, RowID: "sword", Column: "effectID"},
		{Path: referencePath, RowID: "sword", Column: "effectIDs", Index: 0, Array: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(table.Rows) != 1 || table.Rows[0].ID != "burn_large" {
		t.Fatalf("renamed rows = %#v", table.Rows)
	}
	references, err := LoadTable(referencePath)
	if err != nil {
		t.Fatal(err)
	}
	if references.Rows[0].Values["effectID"].Display != "burn_large" || references.Rows[0].Values["effectIDs"].Expression != `["burn_large"]` {
		t.Fatalf("references = %#v", references.Rows[0].Values)
	}
}

func TestAddColumnInitializesAndValidatesEveryRow(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
#Row: {name: string}
rows: {sword: {name: "Sword"}, potion: {name: "Potion"}}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	table, err := AddColumn(path, "price", "int & >=0", "10", -1)
	if err != nil {
		t.Fatal(err)
	}
	if len(table.Columns) != 2 || table.Rows[0].Values["price"].Expression != "10" || table.Rows[1].Values["price"].Expression != "10" {
		t.Fatalf("table = %#v", table)
	}
	if _, err := AddColumn(path, "invalid", "int & >=0", "-1", -1); err == nil {
		t.Fatal("AddColumn() accepted an invalid initial value")
	}
}

func TestAddColumnInsertsAtRequestedPosition(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
#Row: {name: string, rarity: string}
rows: {sword: {name: "Sword", rarity: "N"}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	table, err := AddColumn(path, "price", "int", "0", 1)
	if err != nil || table.Columns[0].Name != "name" || table.Columns[1].Name != "price" || table.Columns[2].Name != "rarity" {
		t.Fatalf("AddColumn() = %#v, %v", table.Columns, err)
	}
}

func TestSetColumnSchemaPersistsBusinessDiagnostics(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
#Row: {price: int}
rows: {sword: {price: 10}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	table, err := SetColumnSchema(path, "price", "int & >=20")
	if err != nil || table.Columns[0].Schema.Requirements[0].Expression != ">=20" {
		t.Fatalf("SetColumnSchema() = %#v, %v", table, err)
	}
	if _, err := SetColumnSchema(path, "price", "["); err == nil {
		t.Fatal("SetColumnSchema() accepted invalid CUE")
	}
}

func TestLoadEmptyTableUsesEmptyCollections(t *testing.T) {
	path := writeSource(t, "empty.cue", `package empty
#PrimaryKey: string
#Row: {}
rows: {}
_tool: {table: {namespace: "game", name: "empty"}, foreignKeys: {}, formulas: {}}
_editor: {columns: {}}`)
	table, err := LoadTable(path)
	if err != nil || table.Columns == nil || table.Rows == nil || table.ForeignKeys == nil || table.Formulas == nil || table.LocalSchema == nil {
		t.Fatalf("LoadTable() = %#v, %v", table, err)
	}
}

func TestDeleteColumnRemovesValuesAndMetadata(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
#Row: {name: string, effectID: string}
rows: {sword: {name: "Sword", effectID: "burn"}}
_tool: {table: {namespace: "rpg", name: "item"}, foreignKeys: {effectID: {table: "rpg.effect"}}, formulas: {effectID: "ignored"}}
_export: {csv: {enabled: true, columns: ["id", "name", "effectID"]}}`)

	table, err := DeleteColumn(path, "effectID")
	if err != nil {
		t.Fatal(err)
	}
	if len(table.Columns) != 1 || table.Columns[0].Name != "name" || table.Rows[0].Values["effectID"].Expression != "" {
		t.Fatalf("table = %#v", table)
	}
	if len(table.ForeignKeys) != 0 || len(table.CSVExport.Columns) != 2 || table.CSVExport.Columns[1] != "name" {
		t.Fatalf("metadata = %#v", table)
	}
}

func TestRenameColumnUpdatesValuesAndMetadata(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
#Row: {effectID: string}
rows: {sword: {effectID: "burn"}}
_tool: {table: {namespace: "rpg", name: "item"}, foreignKeys: {effectID: {table: "rpg.effect"}}, formulas: {effectID: "ignored"}}
_export: {csv: {enabled: true, columns: ["id", "effectID"]}}`)

	table, err := RenameColumn(path, "effectID", "effectKey")
	if err != nil {
		t.Fatal(err)
	}
	if table.Columns[0].Name != "effectKey" || table.Rows[0].Values["effectKey"].Display != "burn" {
		t.Fatalf("table = %#v", table)
	}
	if table.ForeignKeys[0].Column != "effectKey" || table.CSVExport.Columns[1] != "effectKey" {
		t.Fatalf("metadata = %#v", table)
	}
}

func TestMoveColumnReordersSchemaAndCSVExport(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
#Row: {name: string, price: int, rarity: string}
rows: {sword: {name: "Sword", price: 10, rarity: "N"}}
_tool: {table: {namespace: "rpg", name: "item"}}
_export: {csv: {enabled: true, columns: ["id", "name", "price", "rarity"]}}`)

	table, err := MoveColumn(path, "rarity", 0)
	if err != nil {
		t.Fatal(err)
	}
	if table.Columns[0].Name != "rarity" || table.CSVExport.Columns[0] != "id" || table.CSVExport.Columns[1] != "rarity" {
		t.Fatalf("table = %#v", table)
	}
}

func writeSource(t *testing.T, name, source string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeModuleFile(t *testing.T, path, source string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
}
