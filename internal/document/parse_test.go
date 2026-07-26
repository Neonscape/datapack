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

func writeSource(t *testing.T, name, source string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
