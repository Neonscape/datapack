package workspace

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"datapack/internal/document"
)

func TestSchemaRenameScopesPreserveCompileExportsAndHistory(t *testing.T) {
	for _, local := range []bool{false, true} {
		t.Run(map[bool]string{false: "public", true: "local"}[local], func(t *testing.T) {
			s, table, rules := sourceEditFixture(t)
			data, err := os.ReadFile(table)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.Replace(string(data), "package item", "package item\nimport \"example.com/game/schema/rules\"", 1)
			source = regexp.MustCompile(`price:\s*int`).ReplaceAllString(source, "price:rules.#Existing, bonus:_local.#Existing")
			source = regexp.MustCompile(`price:\s*10`).ReplaceAllString(source, "price:10, bonus:1")
			writeFile(t, table, source)
			if _, err := s.OpenProject(s.current.Project.Path); err != nil {
				t.Fatal(err)
			}
			for _, path := range []string{table, rules} {
				if _, err := s.SetDisplayName(DisplayNameSet{Path: path, Definition: "Existing", DisplayName: "保留显示名"}); err != nil {
					t.Fatal(err)
				}
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
				if len(result.Files) != 1 {
					t.Fatalf("export: %#v", result)
				}
				data, err := os.ReadFile(result.Files[0])
				if err != nil {
					t.Fatal(err)
				}
				return string(data)
			}
			beforeCSV := export()
			before := fileTree(t, s.current.Project.Root)
			target := rules
			if local {
				target = table
			}
			if _, err := s.RenameSchemaDefinition(SchemaDefinitionRename{Path: target, OldName: "Existing", NewName: "Renamed"}); err != nil {
				t.Fatal(err)
			}
			if export() != beforeCSV {
				t.Fatal("schema rename changed compiled CSV")
			}
			after := fileTree(t, s.current.Project.Root)
			if _, err := s.Undo(); err != nil {
				t.Fatal(err)
			}
			assertFileTree(t, s.current.Project.Root, before)
			if export() != beforeCSV {
				t.Fatal("Undo changed CSV")
			}
			if _, err := s.Redo(); err != nil {
				t.Fatal(err)
			}
			assertFileTree(t, s.current.Project.Root, after)
			if export() != beforeCSV {
				t.Fatal("Redo changed CSV")
			}
			if local {
				doc, err := s.GetTable(table)
				if err != nil {
					t.Fatal(err)
				}
				if doc.LocalSchema[0].DisplayName != "保留显示名" {
					t.Fatal("lost local title")
				}
			} else {
				doc, err := s.GetPublicSchema(rules)
				if err != nil {
					t.Fatal(err)
				}
				if doc.Definitions[0].DisplayName != "保留显示名" {
					t.Fatal("lost public title")
				}
			}
		})
	}
}

func TestSchemaRenameCaptureRejectedWithoutStateChanges(t *testing.T) {
	s, table, rules := sourceEditFixture(t)
	writeFile(t, rules, `package rules
#Existing:int
#Wrapper:{#Renamed:string, value:#Existing}
_schema:{definitions:{Existing:{kind:"type",type:"integer"}, Wrapper:{kind:"type",type:"struct"}}}`)
	if _, err := s.OpenProject(s.current.Project.Path); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdateCell(CellUpdate{Path: table, RowID: "one", Column: "name", Expression: `"History"`}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	before, state := fileTree(t, s.current.Project.Root), sourceEditState(s)
	if _, err := s.RenameSchemaDefinition(SchemaDefinitionRename{Path: rules, OldName: "Existing", NewName: "Renamed"}); err == nil || !strings.Contains(err.Error(), "capture") {
		t.Fatalf("capture: %v", err)
	}
	assertFileTree(t, s.current.Project.Root, before)
	if sourceEditState(s) != state {
		t.Fatal("rejected rename changed history or revision")
	}
}

func TestSchemaRenameSameNamedScopesRefreshFailureRollsBack(t *testing.T) {
	for _, local := range []bool{false, true} {
		t.Run(map[bool]string{false: "public", true: "local"}[local], func(t *testing.T) {
			s, table, rules := sourceEditFixture(t)
			writeFile(t, filepath.Join(filepath.Dir(table), "attachment.bin"), "keep\x00bytes")
			writeFile(t, s.current.Project.Path, "invalid config")
			before, state := fileTree(t, s.current.Project.Root), sourceEditState(s)
			target := rules
			if local {
				target = table
			}
			if _, err := s.RenameSchemaDefinition(SchemaDefinitionRename{Path: target, OldName: "Existing", NewName: "Renamed"}); err == nil {
				t.Fatal("refresh failure accepted")
			}
			assertFileTree(t, s.current.Project.Root, before)
			if sourceEditState(s) != state {
				t.Fatal("failed refresh changed state")
			}
		})
	}
}
