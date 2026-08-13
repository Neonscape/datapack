package document

import (
	"os"
	"strings"
	"testing"
)

func TestCreatePublicSchemaDefinition(t *testing.T) {
	path := writeSource(t, "common.cue", `package common
#GameID: string
_schema: {version: 1, definitions: {GameID: {kind: "type", type: "string"}}}`)

	schema, err := CreatePublicSchemaDefinition(path, "Rarity", SchemaDefinitionUpdate{Kind: "type", Type: "enum", SchemaExpression: `"N" | "R"`})
	if err != nil {
		t.Fatal(err)
	}
	if len(schema.Definitions) != 2 {
		t.Fatalf("definitions = %#v", schema.Definitions)
	}
	rarity := schema.Definitions[1]
	if rarity.Name != "Rarity" || rarity.Kind != "type" || rarity.Type != "enum" || rarity.SchemaExpression != `"N" | "R"` {
		t.Fatalf("Rarity = %#v", rarity)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(source), `#Rarity: "N" | "R"`) || !strings.Contains(string(source), `type: "enum"`) {
		t.Fatalf("source = %s", source)
	}
	if _, err := CreatePublicSchemaDefinition(path, "Rarity", SchemaDefinitionUpdate{Kind: "type", Type: "enum", SchemaExpression: `"N"`}); err == nil {
		t.Fatal("CreatePublicSchemaDefinition() accepted a duplicate definition")
	}
}

func TestSetPublicSchemaDefinition(t *testing.T) {
	path := writeSource(t, "common.cue", `package common
#GameID: string
_schema: {version: 1, definitions: {GameID: {kind: "type", type: "string"}}}`)

	schema, err := SetPublicSchemaDefinition(path, "GameID", SchemaDefinitionUpdate{Kind: "type", Type: "string", SchemaExpression: `string & =~"^[a-z]+$"`})
	if err != nil {
		t.Fatal(err)
	}
	if schema.Definitions[0].SchemaExpression != `string & =~"^[a-z]+$"` {
		t.Fatalf("definition = %#v", schema.Definitions[0])
	}
	if _, err := SetPublicSchemaDefinition(path, "GameID", SchemaDefinitionUpdate{Kind: "type", Type: "string", SchemaExpression: `[`}); err == nil {
		t.Fatal("SetPublicSchemaDefinition() accepted invalid CUE")
	}
	if _, err := SetPublicSchemaDefinition(path, "Missing", SchemaDefinitionUpdate{Kind: "type", Type: "string", SchemaExpression: `string`}); err == nil {
		t.Fatal("SetPublicSchemaDefinition() accepted a missing definition")
	}
	if _, err := SetPublicSchemaDefinition(path, "GameID", SchemaDefinitionUpdate{Kind: "type", Type: "", SchemaExpression: `string`}); err == nil {
		t.Fatal("SetPublicSchemaDefinition() accepted an empty type")
	}
}

func TestDeletePublicSchemaDefinition(t *testing.T) {
	path := writeSource(t, "common.cue", `package common
#GameID: string
#Rarity: "N"
_schema: {version: 1, definitions: {GameID: {kind: "type", type: "string"}, Rarity: {kind: "type", type: "enum"}}}`)

	schema, err := DeletePublicSchemaDefinition(path, "Rarity")
	if err != nil {
		t.Fatal(err)
	}
	if len(schema.Definitions) != 1 || schema.Definitions[0].Name != "GameID" {
		t.Fatalf("definitions = %#v", schema.Definitions)
	}
	if _, err := DeletePublicSchemaDefinition(path, "Rarity"); err == nil {
		t.Fatal("DeletePublicSchemaDefinition() accepted a missing definition")
	}
}

func TestRenamePublicSchemaDefinitionPropagatesReferences(t *testing.T) {
	root := t.TempDir()
	schemaPath := root + "/common.cue"
	tablePath := root + "/item.cue"
	writeModuleFile(t, schemaPath, `package common
#NonNegative: >=0
#Price: int & #NonNegative
_schema: {version: 1, definitions: {NonNegative: {kind: "constraint", appliesTo: ["integer"]}, Price: {kind: "type", type: "integer"}}}`)
	writeModuleFile(t, tablePath, `package item
#PrimaryKey: string
#Row: {price: int & common.#NonNegative}
rows: {sword: {price: 10}}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	schema, err := RenamePublicSchemaDefinition(schemaPath, "NonNegative", "AtLeastZero", []string{tablePath})
	if err != nil {
		t.Fatal(err)
	}
	for _, definition := range schema.Definitions {
		if definition.Name == "NonNegative" {
			t.Fatal("old definition still registered")
		}
		if definition.Name == "AtLeastZero" && definition.SchemaExpression != ">=0" {
			t.Fatalf("AtLeastZero = %#v", definition)
		}
		if definition.Name == "Price" && definition.SchemaExpression != "int & #AtLeastZero" {
			t.Fatalf("Price = %#v", definition)
		}
	}
	table, err := LoadTable(tablePath)
	if err != nil {
		t.Fatal(err)
	}
	if table.Columns[0].SchemaExpression != "int & common.#AtLeastZero" {
		t.Fatalf("referencing column = %#v", table.Columns[0].SchemaExpression)
	}
}

func TestCreateLocalSchemaDefinition(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
#Row: {}
rows: {}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	table, err := CreateLocalSchemaDefinition(path, "Position", SchemaDefinitionUpdate{Kind: "type", Type: "struct", SchemaExpression: `{x: number, y: number}`})
	if err != nil {
		t.Fatal(err)
	}
	if len(table.LocalSchema) != 1 || table.LocalSchema[0].Name != "Position" || table.LocalSchema[0].Type != "struct" || table.LocalSchema[0].SchemaExpression != `{x: number, y: number}` {
		t.Fatalf("local schema = %#v", table.LocalSchema)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(source), "_local:") || !strings.Contains(string(source), "localDefinitions") {
		t.Fatalf("source = %s", source)
	}
}

func TestSetAndDeleteLocalSchemaDefinition(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
_local: {#Position: {x: number, y: number}}
_schema: {localDefinitions: {Position: {kind: "type", type: "struct"}}}
#Row: {position: _local.#Position}
rows: {sword: {position: {x: 0, y: 0}}}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	table, err := SetLocalSchemaDefinition(path, "Position", SchemaDefinitionUpdate{Kind: "type", Type: "struct", SchemaExpression: `{x: number, y: number, z: number}`})
	if err != nil {
		t.Fatal(err)
	}
	if table.LocalSchema[0].SchemaExpression != `{x: number, y: number, z: number}` {
		t.Fatalf("local schema = %#v", table.LocalSchema)
	}
	table, err = DeleteLocalSchemaDefinition(path, "Position")
	if err != nil {
		t.Fatal(err)
	}
	if len(table.LocalSchema) != 0 {
		t.Fatalf("local schema = %#v", table.LocalSchema)
	}
}

func TestEnsureImportAddsImportWhenMissing(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	if err := EnsureImport(path, "example.com/game/schema:common"); err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(source), `import "example.com/game/schema:common"`) {
		t.Fatalf("source = %s", source)
	}
	if err := EnsureImport(path, "example.com/game/schema:common"); err != nil {
		t.Fatalf("EnsureImport() is not idempotent: %v", err)
	}
	if _, err := LoadTable(path); err != nil {
		t.Fatalf("LoadTable() after import = %v", err)
	}
}

func TestSetColumnSchemaPrunesUnusedImport(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
import "example.com/game/schema:common"
#PrimaryKey: string
#Row: {price: int & common.#NonNegative}
rows: {sword: {price: 10}}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	table, err := SetColumnSchema(path, "price", "int")
	if err != nil {
		t.Fatal(err)
	}
	if table.Columns[0].SchemaExpression != "int" {
		t.Fatalf("column = %#v", table.Columns[0])
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(source), "import") {
		t.Fatalf("unused import was not pruned: %s", source)
	}
}

func TestSetColumnSchemaKeepsUsedImport(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
import "example.com/game/schema:common"
#PrimaryKey: common.#GameID
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	if _, err := SetColumnSchema(path, "name", "string & != \"\""); err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(source), `import "example.com/game/schema:common"`) {
		t.Fatalf("used import was pruned: %s", source)
	}
}

func TestDeleteColumnPrunesUnusedImport(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
import "example.com/game/schema:common"
#PrimaryKey: string
#Row: {name: string, price: int & common.#NonNegative}
rows: {sword: {name: "Sword", price: 10}}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	if _, err := DeleteColumn(path, "price"); err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(source), "import") {
		t.Fatalf("unused import was not pruned: %s", source)
	}
}

func TestRenameLocalSchemaDefinitionPropagatesReferences(t *testing.T) {
	root := t.TempDir()
	tablePath := root + "/item.cue"
	otherPath := root + "/equipment.cue"
	writeModuleFile(t, tablePath, `package item
#PrimaryKey: string
_local: {#Position: {x: number, y: number}}
_schema: {localDefinitions: {Position: {kind: "type", type: "struct"}}}
#Row: {position: _local.#Position}
rows: {sword: {position: {x: 0, y: 0}}}
_tool: {table: {namespace: "rpg", name: "item"}}`)
	writeModuleFile(t, otherPath, `package equipment
#PrimaryKey: string
#Row: {slot: int}
rows: {helmet: {slot: 1}}
_tool: {table: {namespace: "rpg", name: "equipment"}}`)

	table, err := RenameLocalSchemaDefinition(tablePath, "Position", "Pos3D", []string{otherPath})
	if err != nil {
		t.Fatal(err)
	}
	if len(table.LocalSchema) != 1 || table.LocalSchema[0].Name != "Pos3D" {
		t.Fatalf("local schema = %#v", table.LocalSchema)
	}
	if table.Columns[0].SchemaExpression != "_local.#Pos3D" {
		t.Fatalf("column = %#v", table.Columns[0].SchemaExpression)
	}
}
