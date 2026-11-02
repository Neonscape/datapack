package workspace

import (
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"datapack/internal/document"
	"datapack/internal/navigation"
	"datapack/internal/schemamodel"
	"os"
	"path/filepath"
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

func TestSchemaImportImplicitAndExplicitMajorV1(t *testing.T) {
	for _, test := range []struct{ version, alias string }{{"", ""}, {"@v1", ""}, {"", "shared"}, {"@v1", "shared"}} {
		version, alias := test.version, test.alias
		t.Run("major"+version+"/"+alias, func(t *testing.T) {
			s, table, rules := sourceEditFixture(t)
			writeFile(t, filepath.Join(s.current.Project.Root, "cue.mod", "module.cue"), `module:"example.com/game@v1"
language:{version:"v0.9.0"}`)
			data, _ := os.ReadFile(table)
			qualifier, importAlias := "rules", ""
			if alias != "" {
				qualifier, importAlias = alias, alias+" "
			}
			source := strings.Replace(string(data), "package item", "package item\nimport "+importAlias+"\"example.com/game/schema/rules"+version+"\"", 1)
			source = regexp.MustCompile(`price:\s*int`).ReplaceAllString(source, "price:"+qualifier+".#Existing")
			writeFile(t, table, source)
			instances := load.Instances([]string{"."}, &load.Config{Dir: filepath.Dir(table), Package: "item"})
			if len(instances) != 1 {
				t.Fatal("expected one CUE package")
			}
			if err := cuecontext.New().BuildInstance(instances[0]).Err(); err != nil {
				t.Fatalf("CUE must accept import before app indexing: %v", err)
			}
			if _, err := s.OpenProject(s.current.Project.Path); err != nil {
				t.Fatal(err)
			}
			if _, err := s.SetTableExportSettings(TableExportSettings{Path: table, Formats: map[string]document.ExportSettings{"csv": {Enabled: true, Columns: []string{"id", "name", "price"}}}}); err != nil {
				t.Fatal(err)
			}
			if _, err := s.Export(); err != nil {
				t.Fatal(err)
			}
			doc, err := s.GetTable(table)
			if err != nil {
				t.Fatal(err)
			}
			for _, column := range doc.Columns {
				if column.Name == "price" && column.EffectiveSchema.Kind != schemamodel.KindInteger {
					t.Fatal("major version lost effective type")
				}
			}
			target, err := s.Navigate(navigation.Request{Kind: navigation.KindSchemaReference, Path: table, Reference: qualifier + ".#Existing"})
			if err != nil || target.Path != rules {
				t.Fatalf("navigation: %#v %v", target, err)
			}
		})
	}
}

func TestBrokenModuleDoesNotPreventSimpleSchemaRepair(t *testing.T) {
	s, table, _ := sourceEditFixture(t)
	if _, err := s.SetCompileMode("auto"); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(s.current.Project.Root, "cue.mod/module.cue"), `module: {broken:`)
	if _, err := s.OpenProject(s.current.Project.Path); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetColumnSchema(ColumnSchemaSet{Path: table, Name: "price", SchemaExpression: "number"}); err != nil {
		t.Fatalf("legal source repair blocked by module context: %v", err)
	}
	if len(s.current.Compile.Diagnostics) == 0 {
		t.Fatal("broken module must remain a compile diagnostic")
	}
	data, _ := os.ReadFile(table)
	if !regexp.MustCompile(`price:\s*number`).Match(data) {
		t.Fatal("repair was not persisted")
	}
}
