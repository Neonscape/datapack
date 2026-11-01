package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"datapack/internal/compiler"
	"datapack/internal/document"
	"datapack/internal/schemamodel"
)

func structuralResponseFixture(t *testing.T, mode string) (*Service, string) {
	t.Helper()
	root := t.TempDir()
	project := filepath.Join(root, "game.dtproj")
	writeFile(t, project, `{"version":1,"tableRoots":["tables"],"schemaRoots":["schema"],"export":{"directory":"exports","targets":["csv"]}}`)
	writeFile(t, filepath.Join(root, "cue.mod/module.cue"), `module:"example.com/response@v0"
language:{version:"v0.9.0"}`)
	writeFile(t, filepath.Join(root, "schema/common/common.cue"), `package common
#Rarity:"N"|"R"
_schema:{definitions:{Rarity:{kind:"type",type:"enum"}}}`)
	writeFile(t, filepath.Join(root, "tables/targets/targets.cue"), `package targets
#PrimaryKey:string
#Row:{name:string}
rows:{ref:{name:"Ref"}}
_tool:{table:{namespace:"game",name:"targets"}}`)
	path := filepath.Join(root, "tables/items/items.cue")
	writeFile(t, path, `package items
import "example.com/response/schema/common"
#PrimaryKey:string
#Row:{amount:int,total:*0|int,rarity:common.#Rarity,target:string,note:string}
rows:{a:{amount:10,rarity:"N",target:"ref",note:"keep"},b:{amount:20,rarity:"R",target:"ref",note:"keep"},c:{amount:30,rarity:"N",target:"ref",note:"keep"}}
_tool:{table:{namespace:"game",name:"items"},foreignKeys:{target:{table:"game.targets"}},formulas:{a:{total:"amount * 2"},b:{total:"amount * 2"},c:{total:"amount * 2"}}}`)
	s := NewService(t.TempDir())
	view, err := s.OpenProject(project)
	if err != nil || len(view.Compile.Diagnostics) != 0 {
		t.Fatalf("fixture compile: %#v %v", view.Compile.Diagnostics, err)
	}
	if _, err := s.SetCompileMode(mode); err != nil {
		t.Fatal(err)
	}
	if mode == "manual" { // A previous explicit compile can remain visible after edits.
		diagnostics := []compiler.Diagnostic{{Path: path, RowID: "a", Column: "amount", Message: "previous explicit compile diagnostic"}}
		s.compileState.Diagnostics = diagnostics
		s.current.Compile.Diagnostics = diagnostics
	}
	return s, path
}
func assertStructuralResponse(t *testing.T, s *Service, path string, table document.Table) {
	t.Helper()
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fresh, err := s.GetTable(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(table, fresh) {
		t.Error("structural response differs from fresh GetTable")
	}
	if len(table.Formulas) == 0 {
		t.Fatal("fixture lost surviving formulas")
	}
	for _, formula := range table.Formulas {
		var found bool
		for _, row := range table.Rows {
			if row.ID == formula.RowID {
				cell, ok := row.Values[formula.Column]
				if !ok || !cell.Formula || cell.Expression != formula.Expression || cell.Display == "" {
					t.Errorf("formula preview missing: %s.%s: %#v", row.ID, formula.Column, cell)
				}
				found = true
			}
		}
		if !found {
			t.Fatal("dangling formula row")
		}
	}
	for _, column := range table.Columns {
		if column.Name == "rarity" && column.EffectiveSchema.Kind != schemamodel.KindEnum {
			t.Error("effective enum missing")
		}
		if column.Name == "target" && !reflect.DeepEqual(column.ForeignKeyOptions, []string{"ref"}) {
			t.Errorf("foreign key options: %#v", column.ForeignKeyOptions)
		}
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Fatal("preview read wrote source")
	}
	raw, err := document.LoadTable(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, formula := range raw.Formulas {
		for _, row := range raw.Rows {
			if row.ID == formula.RowID {
				if _, ok := row.Values[formula.Column]; ok {
					t.Error("derived formula value was written as a fixed source cell")
				}
			}
		}
	}
}
func TestStructuralTableResultsMatchGetTable(t *testing.T) {
	for _, mode := range []string{"manual", "auto"} {
		for _, operation := range []string{"add row", "delete a", "delete b", "delete c", "rename row", "add column", "delete column", "rename column", "move column"} {
			t.Run(mode+"/"+operation, func(t *testing.T) {
				s, path := structuralResponseFixture(t, mode)
				position := 0
				var table document.Table
				var err error
				switch operation {
				case "add row":
					var r RowAddResult
					r, err = s.AddRow(RowAdd{Path: path, RowID: "d", Expression: `{amount:40,total:80,rarity:"N",target:"ref",note:"keep"}`})
					table = r.Table
				case "delete a", "delete b", "delete c":
					var r RowDeleteResult
					r, err = s.DeleteRow(RowDelete{Path: path, RowID: operation[len("delete "):]})
					table = r.Table
				case "rename row":
					var r RowRenameResult
					r, err = s.RenameRow(RowRename{Path: path, OldID: "b", NewID: "bee"})
					table = r.Table
				case "add column":
					var r ColumnAddResult
					r, err = s.AddColumn(ColumnAdd{Path: path, Name: "extra", SchemaExpression: "string", InitialExpression: `"new"`, Position: &position})
					table = r.Table
				case "delete column":
					var r ColumnDeleteResult
					r, err = s.DeleteColumn(ColumnDelete{Path: path, Name: "note"})
					table = r.Table
				case "rename column":
					var r ColumnRenameResult
					r, err = s.RenameColumn(ColumnRename{Path: path, OldName: "note", NewName: "memo"})
					table = r.Table
				case "move column":
					var r ColumnMoveResult
					r, err = s.MoveColumn(ColumnMove{Path: path, Name: "note", Position: 0})
					table = r.Table
				}
				if err != nil {
					t.Fatal(err)
				}
				assertStructuralResponse(t, s, path, table)
				if _, err = s.Undo(); err != nil {
					t.Fatal(err)
				}
				restored, err := s.GetTable(path)
				if err != nil {
					t.Fatal(err)
				}
				if len(restored.Rows) != 3 {
					t.Fatal("Undo did not restore rows")
				}
				assertStructuralResponse(t, s, path, restored)
				if _, err = s.Redo(); err != nil {
					t.Fatal(err)
				}
				redone, err := s.GetTable(path)
				if err != nil {
					t.Fatal(err)
				}
				assertStructuralResponse(t, s, path, redone)
			})
		}
	}
}

// The browser regression consumes actual Service JSON. Normal Go suites skip
// exporting; Playwright supplies one private temporary output file.
func TestExportStructuralUIFixtures(t *testing.T) {
	output := os.Getenv("DATAPACK_STRUCTURAL_UI_PATH")
	if output == "" {
		t.Skip("only exports a private fixture for the browser contract")
	}
	type fixture struct {
		Mode        string
		Deleted     string
		Before      document.Table
		RawBefore   document.Table
		View        View
		Schema      document.PublicSchema
		Result      RowDeleteResult
		Fresh       document.Table
		Undo        HistoryResult
		UndoTable   document.Table
		Redo        HistoryResult
		RedoTable   document.Table
		Width       CellUpdateResult
		SourceAfter string
	}
	var fixtures []fixture
	for _, mode := range []string{"manual", "auto"} {
		for _, deleted := range []string{"a", "b", "c"} {
			s, path := structuralResponseFixture(t, mode)
			width, err := s.SetColumnWidth(ColumnWidthSet{Path: path, Column: "amount", Width: 190})
			if err != nil {
				t.Fatal(err)
			}
			if _, err = s.Undo(); err != nil {
				t.Fatal(err)
			}
			before, err := s.GetTable(path)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := document.LoadTable(path)
			if err != nil {
				t.Fatal(err)
			}
			schema, err := s.GetPublicSchema(filepath.Join(s.current.Project.Root, "schema/common/common.cue"))
			if err != nil {
				t.Fatal(err)
			}
			view := *s.current
			result, err := s.DeleteRow(RowDelete{Path: path, RowID: deleted})
			if err != nil {
				t.Fatal(err)
			}
			fresh, err := s.GetTable(path)
			if err != nil {
				t.Fatal(err)
			}
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			undo, err := s.Undo()
			if err != nil {
				t.Fatal(err)
			}
			undoTable, err := s.GetTable(path)
			if err != nil {
				t.Fatal(err)
			}
			redo, err := s.Redo()
			if err != nil {
				t.Fatal(err)
			}
			redoTable, err := s.GetTable(path)
			if err != nil {
				t.Fatal(err)
			}
			fixtures = append(fixtures, fixture{mode, deleted, before, raw, view, schema, result, fresh, undo, undoTable, redo, redoTable, width, string(source)})
		}
	}
	data, err := json.Marshal(fixtures)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(output, data, 0600); err != nil {
		t.Fatal(err)
	}
}
