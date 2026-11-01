package workspace

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

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
