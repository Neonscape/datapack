package workspace

import (
	"datapack/internal/document"
	"path/filepath"
	"strconv"
	"testing"
)

func enumRepairFixture(t *testing.T, mode, schema, initial, formula string) (*Service, string) {
	t.Helper()
	root := t.TempDir()
	project := filepath.Join(root, "game.dtproj")
	writeFile(t, project, `{"version":1,"tableRoots":["tables"],"schemaRoots":["schema"],"export":{"directory":"exports","targets":["csv"]}}`)
	writeFile(t, filepath.Join(root, "cue.mod/module.cue"), `module:"example.com/enum@v0"
language:{version:"v0.9.0"}`)
	writeFile(t, filepath.Join(root, "schema/common/common.cue"), `package common
#Marker:string
_schema:{definitions:{Marker:{kind:"type",type:"string"}}}`)
	path := filepath.Join(root, "tables/items/items.cue")
	rank := "rank:" + initial + ","
	formulaMeta := ""
	if formula != "" {
		rank = ""
		formulaMeta = ",formulas:{one:{rank:" + strconv.Quote(formula) + "}}"
	}
	writeFile(t, path, "package items\n#PrimaryKey:string\n#Row:{amount:int,rank:"+schema+"}\nrows:{one:{amount:\"bad\","+rank+"}}\n_tool:{table:{namespace:\"game\",name:\"items\"}"+formulaMeta+"}")
	s := NewService(t.TempDir())
	if _, err := s.OpenProject(project); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetCompileMode(mode); err != nil {
		t.Fatal(err)
	}
	return s, path
}
func TestEnumRepairPreservesLiteralTypeAndOtherInvalidInput(t *testing.T) {
	cases := []struct{ name, schema, initial, target string }{{"number", "1|2", `"BAD"`, "2"}, {"boolean", "false|true", "false", "true"}, {"nullable", "null|1|2", "1", "null"}, {"escaped string", strconv.Quote("plain") + "|" + strconv.Quote("a\"b\\c"), strconv.Quote("plain"), strconv.Quote("a\"b\\c")}}
	for _, mode := range []string{"manual", "auto"} {
		for _, c := range cases {
			t.Run(mode+"/"+c.name, func(t *testing.T) {
				s, path := enumRepairFixture(t, mode, c.schema, c.initial, "")
				result, err := s.UpdateCell(CellUpdate{Path: path, RowID: "one", Column: "rank", Expression: c.target})
				if err != nil {
					t.Fatal(err)
				}
				if cell := result.Table.Rows[0].Values["rank"]; cell.Expression != c.target || cell.Formula {
					t.Fatalf("rank type/expression: %#v", cell)
				}
				raw, err := document.LoadTable(path)
				if err != nil {
					t.Fatal(err)
				}
				if raw.Rows[0].Values["amount"].Expression != `"bad"` || raw.Rows[0].Values["rank"].Expression != c.target {
					t.Fatal("enum repair changed other source or literal type")
				}
				if _, err = s.Compile(); err != nil {
					t.Fatal(err)
				}
				fresh, err := s.GetTable(path)
				if err != nil {
					t.Fatal(err)
				}
				var otherError bool
				for _, d := range fresh.Diagnostics {
					if d.Column == "rank" {
						t.Fatalf("repaired enum still invalid: %#v", d)
					}
					if d.Column == "amount" {
						otherError = true
					}
				}
				if !otherError {
					t.Fatal("unrelated invalid amount was not preserved/diagnosed")
				}
			})
		}
	}
}
func TestFormulaEnumRepairPreservesFormulaSource(t *testing.T) {
	for _, mode := range []string{"manual", "auto"} {
		t.Run(mode, func(t *testing.T) {
			s, path := enumRepairFixture(t, mode, "1|2", "", "missing + 1")
			before, err := s.GetTable(path)
			if err != nil {
				t.Fatal(err)
			}
			if !before.Rows[0].Values["rank"].Formula {
				t.Fatal("missing invalid formula preview")
			}
			result, err := s.SetCellFormula(FormulaSet{Path: path, RowID: "one", Column: "rank", Expression: "1 + 1"})
			if err != nil {
				t.Fatal(err)
			}
			cell := result.Table.Rows[0].Values["rank"]
			if !cell.Formula || cell.Expression != "1 + 1" || cell.Display != "2" {
				t.Fatalf("repaired formula: %#v", cell)
			}
			raw, err := document.LoadTable(path)
			if err != nil {
				t.Fatal(err)
			}
			if _, ok := raw.Rows[0].Values["rank"]; ok {
				t.Fatal("formula was materialized as fixed source")
			}
			if raw.Formulas[0].Expression != "1 + 1" {
				t.Fatal("source formula not repaired")
			}
		})
	}
}
