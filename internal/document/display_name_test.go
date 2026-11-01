package document

import (
	"os"
	"strings"
	"testing"
)

func TestDisplayNamesRoundTripPreserveIdentityAndOtherMetadata(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
#Row: {name: string}
rows: {one: {name: "Original"}}
_local: {#Label: string}
_schema: {localDefinitions: {Label: {kind: "type", type: "string"}}}
_tool: {table: {namespace: "rpg", name: "item"}}
_editor: {table: {title: "旧标题", note: "keep"}, columns: {name: {width: 200}}}`)
	if err := os.Chmod(path, 0600); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"", "Label"} {
		if err := SetDisplayName(path, name, "  道具 \"名\"\n说明  "); err != nil {
			t.Fatal(err)
		}
	}
	table, err := LoadTable(path)
	if err != nil {
		t.Fatal(err)
	}
	if table.DisplayName != "道具 \"名\"\n说明" || table.LocalSchema[0].DisplayName != table.DisplayName || table.Name != "item" || table.Namespace != "rpg" || table.Rows[0].ID != "one" || table.Columns[0].Width != 200 {
		t.Fatalf("table = %#v", table)
	}
	if info, _ := os.Stat(path); info.Mode().Perm() != 0600 {
		t.Fatalf("mode = %v", info.Mode())
	}
	if err := SetDisplayName(path, "", " \t "); err != nil {
		t.Fatal(err)
	}
	table, _ = LoadTable(path)
	source, _ := os.ReadFile(path)
	if table.DisplayName != "" || !strings.Contains(string(source), `note: "keep"`) || table.LocalSchema[0].DisplayName == "" {
		t.Fatalf("clear removed unrelated metadata: %s", source)
	}
}

func TestDefinitionTitlesFollowRenameDeleteAndAllowDuplicates(t *testing.T) {
	for _, local := range []bool{false, true} {
		t.Run(map[bool]string{true: "local", false: "public"}[local], func(t *testing.T) {
			source := `package common
#Old: string
#Other: string
_schema: {definitions: {Old: {kind: "type", type: "string"}, Other: {kind: "type", type: "string"}}}`
			if local {
				source = `package item
#PrimaryKey: string
#Row: {name: string}
rows: {one: {name: "Original"}}
_local: {#Old: string, #Other: string}
_schema: {localDefinitions: {Old: {kind: "type", type: "string"}, Other: {kind: "type", type: "string"}}}
_tool: {table: {namespace: "rpg", name: "item"}}`
			}
			path := writeSource(t, "names.cue", source)
			for _, name := range []string{"Old", "Other"} {
				if err := SetDisplayName(path, name, "相同显示名"); err != nil {
					t.Fatal(err)
				}
			}
			if local {
				_, err := RenameLocalSchemaDefinition(path, "Old", "New", nil)
				if err != nil {
					t.Fatal(err)
				}
			} else {
				_, err := RenamePublicSchemaDefinition(path, "Old", "New", nil)
				if err != nil {
					t.Fatal(err)
				}
			}
			var defs []SchemaDefinition
			if local {
				doc, err := LoadTable(path)
				if err != nil {
					t.Fatal(err)
				}
				defs = doc.LocalSchema
			} else {
				doc, err := LoadPublicSchema(path)
				if err != nil {
					t.Fatal(err)
				}
				defs = doc.Definitions
			}
			for _, def := range defs {
				if def.DisplayName != "相同显示名" || def.Name == "Old" {
					t.Fatalf("definitions = %#v", defs)
				}
			}
			if local {
				_, err := DeleteLocalSchemaDefinition(path, "New")
				if err != nil {
					t.Fatal(err)
				}
			} else {
				_, err := DeletePublicSchemaDefinition(path, "New")
				if err != nil {
					t.Fatal(err)
				}
			}
			updated, _ := os.ReadFile(path)
			if strings.Contains(string(updated), "New:") {
				t.Fatalf("orphan metadata after delete: %s", updated)
			}
		})
	}
}

func TestDisplayNameRejectsMalformedOrUnknownTargetWithoutWrites(t *testing.T) {
	for _, metadata := range []string{`_editor: "bad"`, `_editor: {schema: "bad"}`, `_editor: {schemas: "bad"}`} {
		path := writeSource(t, "common.cue", `package common
#Old: string
_schema: {definitions: {Old: {kind: "type",type: "string"}}}
`+metadata)
		for _, name := range []string{"", "Old", "Missing"} {
			if name == "" && strings.Contains(metadata, "schemas:") {
				continue
			}
			if name == "Old" && strings.Contains(metadata, "schema:") {
				continue
			}
			before, _ := os.ReadFile(path)
			if err := SetDisplayName(path, name, "新标题"); err == nil {
				t.Fatalf("accepted %s / %s", metadata, name)
			}
			after, _ := os.ReadFile(path)
			if string(after) != string(before) {
				t.Fatal("failed edit changed source")
			}
		}
	}
	path := writeSource(t, "common.cue", `package common
#Old: string
_schema: {definitions: {Old: {kind: "type",type: "string"}}}
_editor: {schemas: {New: {title: "orphan"}}}`)
	before, _ := os.ReadFile(path)
	if _, err := RenamePublicSchemaDefinition(path, "Old", "New", nil); err == nil {
		t.Fatal("accepted orphan target metadata")
	}
	after, _ := os.ReadFile(path)
	if string(after) != string(before) {
		t.Fatal("failed rename changed source")
	}
}
