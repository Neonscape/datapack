package workspace

import (
	"datapack/internal/document"
	"datapack/internal/navigation"
	"datapack/internal/schemamodel"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestSchemaImportAliasValidCUEExport(t *testing.T) {
	s, table, _ := sourceEditFixture(t)
	data, err := os.ReadFile(table)
	if err != nil {
		t.Fatal(err)
	}
	source := strings.Replace(string(data), "package item", "package item\nimport shared \"example.com/game/schema/rules\"", 1)
	source = regexp.MustCompile(`price:\s*int`).ReplaceAllString(source, "price:shared.#Existing")
	writeFile(t, table, source)
	if _, err := s.OpenProject(s.current.Project.Path); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetTableExportSettings(TableExportSettings{Path: table, Formats: map[string]document.ExportSettings{"csv": {Enabled: true, Columns: []string{"id", "name", "price"}}}}); err != nil {
		t.Fatal(err)
	}
	if result, err := s.Export(); err != nil || len(result.Files) != 1 {
		t.Fatalf("export: %#v %v", result, err)
	}
}

func TestSchemaImportAliasEditorMetadataNavigationAndDeleteProtection(t *testing.T) {
	s, table, rules := sourceEditFixture(t)
	data, _ := os.ReadFile(table)
	source := strings.Replace(string(data), "package item", "package item\nimport shared \"example.com/game/schema/rules\"", 1)
	source = regexp.MustCompile(`price:\s*int`).ReplaceAllString(source, "price:shared.#Existing")
	writeFile(t, table, source)
	if _, err := s.OpenProject(s.current.Project.Path); err != nil {
		t.Fatal(err)
	}
	doc, err := s.GetTable(table)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, column := range doc.Columns {
		if column.Name == "price" {
			found = true
			if column.Schema.Reference != "shared.#Existing" || column.EffectiveSchema.Kind != schemamodel.KindInteger {
				t.Fatalf("column: %#v", column)
			}
		}
	}
	if !found {
		t.Fatal("missing price column")
	}
	target, err := s.Navigate(navigation.Request{Kind: navigation.KindSchemaReference, Path: table, Reference: "shared.#Existing", Column: "price"})
	if err != nil || target.Path != rules || target.Definition != "Existing" {
		t.Fatalf("navigation: %#v %v", target, err)
	}
	if _, err := s.SetColumnSchema(ColumnSchemaSet{Path: table, Name: "price", SchemaExpression: "shared.#Existing"}); err != nil {
		t.Fatal(err)
	}
	afterSave, _ := os.ReadFile(table)
	if !strings.Contains(string(afterSave), `import shared "example.com/game/schema/rules"`) {
		t.Fatal("source import alias was rewritten")
	}
	before, state := fileTree(t, s.current.Project.Root), sourceEditState(s)
	if _, err := s.DeleteSchemaDefinition(SchemaDefinitionDelete{Path: rules, Name: "Existing"}); err == nil {
		t.Fatal("aliased reference was ignored by deletion protection")
	}
	assertFileTree(t, s.current.Project.Root, before)
	if sourceEditState(s) != state {
		t.Fatal("rejected deletion changed history")
	}
}
