package workspace

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"datapack/internal/document"
	"datapack/internal/projecttree"
)

func TestDisplayNameHistoryReopenTreeRegistryAndExports(t *testing.T) {
	s, tablePath, schemaPath := sourceEditFixture(t)
	if _, err := s.SetTableExportSettings(TableExportSettings{Path: tablePath, Formats: map[string]document.ExportSettings{"csv": {Enabled: true, Columns: []string{"id", "name", "price"}}}}); err != nil {
		t.Fatal(err)
	}
	export := func() string {
		t.Helper()
		result, err := s.Export()
		if err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(result.Files[0])
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	beforeExport := export()
	for _, target := range []struct{ path, definition string }{{tablePath, ""}, {tablePath, "Existing"}, {schemaPath, ""}, {schemaPath, "Existing"}} {
		before, _ := os.ReadFile(target.path)
		result, err := s.SetDisplayName(DisplayNameSet{Path: target.path, Definition: target.definition, DisplayName: "中文标题"})
		if err != nil {
			t.Fatal(err)
		}
		after, _ := os.ReadFile(target.path)
		if reflect.DeepEqual(before, after) {
			t.Fatal("display name not persisted")
		}
		if _, err := s.Undo(); err != nil {
			t.Fatal(err)
		}
		undone, _ := os.ReadFile(target.path)
		if !reflect.DeepEqual(before, undone) {
			t.Fatal("undo differs")
		}
		if _, err := s.Redo(); err != nil {
			t.Fatal(err)
		}
		redone, _ := os.ReadFile(target.path)
		if !reflect.DeepEqual(after, redone) {
			t.Fatal("redo differs")
		}
		if target.definition != "" {
			found := false
			for _, def := range result.Workspace.Schemas.Definitions {
				if def.Path == target.path && def.DisplayName == "中文标题" {
					found = true
				}
			}
			if !found {
				t.Fatal("title missing from registry")
			}
		}
	}
	loaded, err := s.GetTable(tablePath)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.DisplayName != "中文标题" || loaded.LocalSchema[0].DisplayName != "中文标题" {
		t.Fatalf("table: %#v", loaded)
	}
	schema, err := s.GetPublicSchema(schemaPath)
	if err != nil {
		t.Fatal(err)
	}
	if schema.DisplayName != "中文标题" || schema.Definitions[0].DisplayName != "中文标题" {
		t.Fatalf("schema: %#v", schema)
	}
	if export() != beforeExport {
		t.Fatal("titles changed CSV contents or headers")
	}
	reopened := NewService(t.TempDir())
	view, err := reopened.OpenProject(s.current.Project.Path)
	if err != nil {
		t.Fatal(err)
	}
	if findNodeByKind(view.Tree.Root, "item", projecttree.NodeTable).DisplayName != "中文标题" || findNodeByKind(view.Tree.Root, filepath.Base(schemaPath), projecttree.NodeSchemaFile).DisplayName != "中文标题" {
		t.Fatal("tree titles missing after reopen")
	}
	if _, err := s.RenameSchemaDefinition(SchemaDefinitionRename{Path: schemaPath, OldName: "Existing", NewName: "Renamed"}); err != nil {
		t.Fatal(err)
	}
	schema, _ = s.GetPublicSchema(schemaPath)
	if schema.Definitions[0].Name != "Renamed" || schema.Definitions[0].DisplayName != "中文标题" {
		t.Fatal("rename lost title")
	}
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	schema, _ = s.GetPublicSchema(schemaPath)
	if schema.Definitions[0].Name != "Existing" || schema.Definitions[0].DisplayName != "中文标题" {
		t.Fatal("undo rename lost title")
	}
	state := sourceEditState(s)
	source, _ := os.ReadFile(tablePath)
	if _, err := s.SetDisplayName(DisplayNameSet{Path: tablePath, Definition: "Missing", DisplayName: "bad"}); err == nil {
		t.Fatal("accepted nonexistent definition")
	}
	unchanged, _ := os.ReadFile(tablePath)
	if !reflect.DeepEqual(source, unchanged) || sourceEditState(s) != state {
		t.Fatal("failed edit changed source/history/state")
	}
}

func TestTableDisplayNameSurvivesRenameDeleteAndTheirHistory(t *testing.T) {
	s, path, _ := sourceEditFixture(t)
	if _, err := s.SetDisplayName(DisplayNameSet{Path: path, DisplayName: "道具表"}); err != nil {
		t.Fatal(err)
	}
	renamed, err := s.RenameTable(TableRename{Path: path, Namespace: "rpg", Name: "inventory"})
	if err != nil {
		t.Fatal(err)
	}
	if renamed.Table.DisplayName != "道具表" || renamed.Table.Name != "inventory" {
		t.Fatal("rename changed display metadata")
	}
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	table, err := s.GetTable(path)
	if err != nil || table.DisplayName != "道具表" {
		t.Fatalf("undo rename: %#v %v", table, err)
	}
	if _, err := s.Redo(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DeleteTable(TableDelete{Path: renamed.Table.Path}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	table, err = s.GetTable(renamed.Table.Path)
	if err != nil || table.DisplayName != "道具表" {
		t.Fatalf("undo delete: %#v %v", table, err)
	}
	if _, err := s.Redo(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(renamed.Table.Path); !os.IsNotExist(err) {
		t.Fatalf("redo delete: %v", err)
	}
}

func TestDisplayNamesKeepExportsWithMultiFileImportedPackage(t *testing.T) {
	s, table, rules := sourceEditFixture(t)
	second := filepath.Join(filepath.Dir(rules), "extra.cue")
	writeFile(t, second, `package rules
#Extra: string
_schema: {definitions: {Extra: {kind: "type", type: "string"}}}`)
	if _, err := s.OpenProject(s.current.Project.Path); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetColumnSchema(ColumnSchemaSet{Path: table, Name: "price", SchemaExpression: "rules.#Existing"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetTableExportSettings(TableExportSettings{Path: table, Formats: map[string]document.ExportSettings{"csv": {Enabled: true, Columns: []string{"id", "name", "price"}}}}); err != nil {
		t.Fatal(err)
	}
	export := func() string {
		t.Helper()
		result, err := s.Export()
		if err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(result.Files[0])
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	validatePackage := func() {
		t.Helper()
		instances := load.Instances([]string{"."}, &load.Config{Dir: filepath.Dir(rules), Package: "rules"})
		if len(instances) != 1 {
			t.Fatal("unexpected schema package count")
		}
		if err := cuecontext.New().BuildInstance(instances[0]).Err(); err != nil {
			t.Fatal(err)
		}
	}
	validatePackage()
	before := export()
	for path, title := range map[string]string{rules: "规则文件甲", second: "规则文件乙"} {
		if _, err := s.SetDisplayName(DisplayNameSet{Path: path, DisplayName: title}); err != nil {
			t.Fatal(err)
		}
	}
	validatePackage()
	if after := export(); after != before {
		t.Fatal("file titles changed imported-package compilation/export")
	}
}

func TestDisplayNamesKeepExportsWithMultiFileTablePackage(t *testing.T) {
	s, first, _ := sourceEditFixture(t)
	second := filepath.Join(filepath.Dir(filepath.Dir(first)), "extra", "extra.cue")
	writeFile(t, second, `package item
#PrimaryKey:string
#Row:{name:string,price:int}
rows:{two:{name:"Other",price:20}}
_tool:{table:{namespace:"rpg",name:"extra"}}`)
	if _, err := s.OpenProject(s.current.Project.Path); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{first, second} {
		if _, err := s.SetTableExportSettings(TableExportSettings{Path: path, Formats: map[string]document.ExportSettings{"csv": {Enabled: true, Columns: []string{"id", "name", "price"}}}}); err != nil {
			t.Fatal(err)
		}
	}
	export := func() map[string]string {
		t.Helper()
		result, err := s.Export()
		if err != nil {
			t.Fatal(err)
		}
		files := map[string]string{}
		for _, path := range result.Files {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			files[filepath.Base(path)] = string(data)
		}
		if len(files) != 2 {
			t.Fatalf("tables: %v", files)
		}
		return files
	}
	before := export()
	for path, title := range map[string]string{first: "道具主表", second: "同包附表"} {
		if _, err := s.SetDisplayName(DisplayNameSet{Path: path, DisplayName: title}); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(before, export()) {
		t.Fatal("table titles changed shared-package compilation/export")
	}
}

func TestDifferentTableIdentitiesInOnePackageAreAlreadyInvalid(t *testing.T) {
	s, table, _ := sourceEditFixture(t)
	second := filepath.Join(filepath.Dir(table), "extra.cue")
	writeFile(t, second, `package item
#PrimaryKey:string
#Row:{name:string,price:int}
rows:{two:{name:"Other",price:20}}
_tool:{table:{namespace:"rpg",name:"extra"}}`)
	if _, err := s.OpenProject(s.current.Project.Path); err != nil {
		t.Fatal(err)
	}
	check := func() {
		t.Helper()
		if _, err := s.Compile(); err != nil {
			t.Fatal(err)
		}
		found := false
		for _, diagnostic := range s.current.Compile.Diagnostics {
			if strings.Contains(diagnostic.Message, "_tool.table.name") {
				found = true
			}
		}
		if !found {
			t.Fatal("expected existing package identity conflict")
		}
	}
	check()
	for path, title := range map[string]string{table: "主表", second: "附表"} {
		if _, err := s.SetDisplayName(DisplayNameSet{Path: path, DisplayName: title}); err != nil {
			t.Fatal(err)
		}
	}
	check()
}
