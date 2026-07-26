package document

import (
	"os"
	"path/filepath"
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

func TestSetCellValueRejectsValuesOutsideColumnSchema(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
#Row: {price: int & >= 0, rarity: "N" | "R"}
rows: {sword: {price: 100, rarity: "N"}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := SetCellValue(path, "sword", "price", "-1"); err == nil {
		t.Fatal("SetCellValue() accepted a value outside an integer constraint")
	}
	if _, err := SetCellValue(path, "sword", "rarity", `"X"`); err == nil {
		t.Fatal("SetCellValue() accepted a value outside an enum")
	}
	current, err := os.ReadFile(path)
	if err != nil || string(current) != string(original) {
		t.Fatalf("source changed after invalid schema update: %q, error = %v", current, err)
	}
}

func TestSetCellValueValidatesImportedSchema(t *testing.T) {
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

	if _, err := SetCellValue(path, "sword", "price", "-1"); err == nil {
		t.Fatal("SetCellValue() accepted a value outside an imported constraint")
	}
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
