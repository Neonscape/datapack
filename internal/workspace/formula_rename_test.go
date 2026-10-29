package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"datapack/internal/document"
	"datapack/internal/formula"
)

func formulaRenameFixture(t *testing.T, ns, schema, rows, formulas string) (*Service, string) {
	t.Helper()
	root := t.TempDir()
	project := filepath.Join(root, "game.dtproj")
	writeFile(t, project, `{"version":1,"tableRoots":["tables"],"schemaRoots":["schema"],"export":{"directory":"exports","targets":["csv"]}}`)
	writeFile(t, filepath.Join(root, "cue.mod", "module.cue"), `module:"example.com/game"
language:{version:"v0.9.0"}`)
	if err := os.MkdirAll(filepath.Join(root, "schema"), 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "tables", "item", "item.cue")
	writeFile(t, path, fmt.Sprintf("package item\n#PrimaryKey:string\n#Row:%s\nrows:%s\n_tool:{table:{namespace:%q,name:\"item\"},formulas:%s}\n", schema, rows, ns, formulas))
	s := NewService(t.TempDir())
	if _, err := s.OpenProject(project); err != nil {
		t.Fatal(err)
	}
	return s, path
}

func assertRenamedFormulaValue(t *testing.T, s *Service, path, row, column, want string) {
	t.Helper()
	table, err := document.LoadTable(path)
	if err != nil {
		t.Fatal(err)
	}
	result := formula.Evaluate(table)
	if len(result.Diagnostics) > 0 || fmt.Sprint(result.Rows[row][column]) != want {
		t.Fatalf("evaluated=%#v, want %s.%s=%s", result, row, column, want)
	}
	compiled, err := s.Compile()
	if err != nil || len(compiled.Workspace.Compile.Diagnostics) > 0 {
		t.Fatalf("compile after rename: %#v %v", compiled.Workspace.Compile.Diagnostics, err)
	}
}

func TestFormulaColumnRenamePreservesScopeAndQualifiedAddresses(t *testing.T) {
	for _, tc := range []struct{ name, old, next, expression, want string }{
		{"dotted and bare", "price", "cost", `rpg.balance.item.one.price + price`, "20"},
		{"brackets", "price", "cost", `rpg['balance']['item']['one']['price'] + 1`, "11"},
		{"local shadows old", "price", "cost", `let price=3; price+1`, "4"},
		{"new name is local", "price", "cost", `let cost=3; price+cost`, "13"},
		{"binding value before scope", "price", "cost", `let price=price+1; price+1`, "12"},
		{"map labels", "price", "cost", `let m={price:2}; price+m.price`, "12"},
		{"function callee", "abs", "absolute", `abs(-2) + abs`, "9"},
		{"registered function callee", "pow", "power", `pow(2,3) + pow`, "15"},
		{"namespace root is column", "rpg", "rootValue", `rpg + rpg.balance.item.one.price`, "13"},
		{"local namespace object", "price", "cost", `let rpg={balance:{item:{one:{price:4}}}}; price+rpg.balance.item.one.price`, "14"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, path := formulaRenameFixture(t, "rpg.balance", `{total:*0|int,price:int,abs:*7|int,pow:*7|int,rpg:*3|int,text:*""|string}`, `{one:{price:10}}`, `{one:{total:`+fmt.Sprintf("%q", tc.expression)+`,text:"\"price rpg.balance.item.one.price abs(-2)\""}}`)
			assertRenamedFormulaValue(t, s, path, "one", "total", tc.want)
			before := fileTree(t, s.current.Project.Root)
			if _, err := s.RenameColumn(ColumnRename{Path: path, OldName: tc.old, NewName: tc.next}); err != nil {
				t.Fatal(err)
			}
			assertRenamedFormulaValue(t, s, path, "one", "total", tc.want)
			assertRenamedFormulaValue(t, s, path, "one", "text", "price rpg.balance.item.one.price abs(-2)")
			after := fileTree(t, s.current.Project.Root)
			if _, err := s.Undo(); err != nil {
				t.Fatal(err)
			}
			assertFileTree(t, s.current.Project.Root, before)
			if _, err := s.Redo(); err != nil {
				t.Fatal(err)
			}
			assertFileTree(t, s.current.Project.Root, after)
		})
	}
}

func TestFormulaColumnRenameDoesNotRenameMatchingRowID(t *testing.T) {
	s, path := formulaRenameFixture(t, "rpg.balance", `{price:int,total:*0|int}`, `{price:{price:10}}`, `{price:{total:"price+1"}}`)
	if _, err := s.RenameColumn(ColumnRename{Path: path, OldName: "price", NewName: "cost"}); err != nil {
		t.Fatal(err)
	}
	assertRenamedFormulaValue(t, s, path, "price", "total", "11")
}

func TestFormulaColumnRenameMigratesFormulaTargetKeys(t *testing.T) {
	s, path := formulaRenameFixture(t, "rpg.balance", `{price:*0|int,total:*0|int}`, `{one:{}}`, `{one:{price:"10",total:"price+1"}}`)
	if _, err := s.RenameColumn(ColumnRename{Path: path, OldName: "price", NewName: "cost"}); err != nil {
		t.Fatal(err)
	}
	assertRenamedFormulaValue(t, s, path, "one", "cost", "10")
	assertRenamedFormulaValue(t, s, path, "one", "total", "11")
}

func TestFormulaRowRenameMigratesDottedAddressesAndTargets(t *testing.T) {
	s, path := formulaRenameFixture(t, "rpg.balance", `{a:*0|int,b:int,text:*""|string}`, `{one:{b:10},two:{b:20}}`, `{one:{a:"rpg.balance.item['two'].b+1",text:"\"rpg.balance.item.two.b\""},two:{a:"b+2"}}`)
	assertRenamedFormulaValue(t, s, path, "one", "a", "21")
	before := fileTree(t, s.current.Project.Root)
	if _, err := s.RenameRow(RowRename{Path: path, OldID: "two", NewID: "second"}); err != nil {
		t.Fatal(err)
	}
	assertRenamedFormulaValue(t, s, path, "one", "a", "21")
	assertRenamedFormulaValue(t, s, path, "second", "a", "22")
	assertRenamedFormulaValue(t, s, path, "one", "text", "rpg.balance.item.two.b")
	after := fileTree(t, s.current.Project.Root)
	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	assertFileTree(t, s.current.Project.Root, before)
	if _, err := s.Redo(); err != nil {
		t.Fatal(err)
	}
	assertFileTree(t, s.current.Project.Root, after)
}

func TestFormulaRenameSupportsQuotedSourceNames(t *testing.T) {
	for _, id := range []string{"", "quoted row"} {
		t.Run("row="+id, func(t *testing.T) {
			expression := fmt.Sprintf("rpg.balance.item[%q].b+1", id)
			s, path := formulaRenameFixture(t, "rpg.balance", `{a:*0|int,b:*10|int}`, fmt.Sprintf("{%q:{b:10},target:{}}", id), fmt.Sprintf("{target:{a:%q}}", expression))
			assertRenamedFormulaValue(t, s, path, "target", "a", "11")
			if _, err := s.RenameRow(RowRename{Path: path, OldID: id, NewID: "renamed"}); err != nil {
				t.Fatal(err)
			}
			assertRenamedFormulaValue(t, s, path, "target", "a", "11")
		})
	}
	t.Run("empty column", func(t *testing.T) {
		s, path := formulaRenameFixture(t, "rpg.balance", `{a:*0|int,"":int}`, `{one:{"":10}}`, `{one:{a:"rpg.balance.item.one['']+1"}}`)
		assertRenamedFormulaValue(t, s, path, "one", "a", "11")
		if _, err := s.RenameColumn(ColumnRename{Path: path, OldName: "", NewName: "renamed"}); err != nil {
			t.Fatal(err)
		}
		assertRenamedFormulaValue(t, s, path, "one", "a", "11")
	})
}

func TestFormulaTableRenameMigratesVariableLengthNamespace(t *testing.T) {
	for _, expression := range []string{`rpg.balance.item.two.b + b`, `rpg['balance']['item']['two']['b'] + b`, `let rpg={balance:{item:{two:{b:20}}}}; rpg.balance.item.two.b + b`} {
		t.Run(expression, func(t *testing.T) {
			s, path := formulaRenameFixture(t, "rpg.balance", `{a:*0|int,b:int,text:*""|string}`, `{one:{b:10},two:{b:20}}`, `{one:{a:`+fmt.Sprintf("%q", expression)+`,text:"\"rpg.balance.item.two.b\""}}`)
			assertRenamedFormulaValue(t, s, path, "one", "a", "30")
			before := fileTree(t, s.current.Project.Root)
			renamed, err := s.RenameTable(TableRename{Path: path, Namespace: "combat.config.scaling", Name: "goods"})
			if err != nil {
				t.Fatal(err)
			}
			assertRenamedFormulaValue(t, s, renamed.Table.Path, "one", "a", "30")
			assertRenamedFormulaValue(t, s, renamed.Table.Path, "one", "text", "rpg.balance.item.two.b")
			after := fileTree(t, s.current.Project.Root)
			if _, err := s.Undo(); err != nil {
				t.Fatal(err)
			}
			assertFileTree(t, s.current.Project.Root, before)
			if _, err := s.Redo(); err != nil {
				t.Fatal(err)
			}
			assertFileTree(t, s.current.Project.Root, after)
		})
	}
}

func TestFormulaRenameRejectsUnsafeCaptureAndOrphanTargets(t *testing.T) {
	for _, kind := range []string{"column capture", "table capture", "orphan row formula", "orphan column formula"} {
		t.Run(kind, func(t *testing.T) {
			formulas := `{one:{a:"price+1"}}`
			switch kind {
			case "column capture":
				formulas = `{one:{a:"let cost=3; let rpg=4; price+cost+rpg"}}`
			case "table capture":
				formulas = `{one:{a:"let combat=3; rpg.balance.item.one.price+combat"}}`
			case "orphan row formula":
				formulas = `{one:{a:"price+1"},second:{a:"123"}}`
			case "orphan column formula":
				formulas = `{one:{price:"10",cost:"20"}}`
			}
			rows := `{one:{price:10}}`
			if kind == "orphan column formula" {
				rows = `{one:{}}`
			}
			s, path := formulaRenameFixture(t, "rpg.balance", `{price:*0|int,a:*0|int}`, rows, formulas)
			before := fileTree(t, s.current.Project.Root)
			state := fmt.Sprintf("%#v|%#v|%#v", s.undoStack, s.redoStack, *s.current)
			var err error
			if kind == "table capture" {
				_, err = s.RenameTable(TableRename{Path: path, Namespace: "combat", Name: "goods"})
			} else if kind == "orphan row formula" {
				_, err = s.RenameRow(RowRename{Path: path, OldID: "one", NewID: "second"})
			} else {
				_, err = s.RenameColumn(ColumnRename{Path: path, OldName: "price", NewName: "cost"})
			}
			if err == nil {
				t.Fatal("unsafe rename was accepted")
			}
			assertFileTree(t, s.current.Project.Root, before)
			if fmt.Sprintf("%#v|%#v|%#v", s.undoStack, s.redoStack, *s.current) != state {
				t.Fatal("rejected rename changed workspace or history")
			}
		})
	}
}

func TestFormulaColumnRenameToNamespaceRootKeepsStructRead(t *testing.T) {
	s, path := formulaRenameFixture(t, "rpg", `{value:*0|int,price:{item:int}}`, `{one:{price:{item:7}}}`, `{one:{value:"price.item"}}`)
	assertRenamedFormulaValue(t, s, path, "one", "value", "7")
	if _, err := s.RenameColumn(ColumnRename{Path: path, OldName: "price", NewName: "rpg"}); err != nil {
		t.Fatal(err)
	}
	assertRenamedFormulaValue(t, s, path, "one", "value", "7")
}

func TestFormulaColumnRenameToExprReservedNameKeepsValue(t *testing.T) {
	for _, name := range []string{"nil", "and"} {
		t.Run(name, func(t *testing.T) {
			s, path := formulaRenameFixture(t, "rpg.balance", `{price:int,a:*0|int}`, `{one:{price:10}}`, `{one:{a:"price == nil ? 99 : price+1"}}`)
			if _, err := s.RenameColumn(ColumnRename{Path: path, OldName: "price", NewName: name}); err != nil {
				t.Fatal(err)
			}
			assertRenamedFormulaValue(t, s, path, "one", "a", "11")
		})
	}
}

func TestFormulaTableRenameToUnrepresentableRootFailsWithoutWrites(t *testing.T) {
	s, path := formulaRenameFixture(t, "rpg.balance", `{price:int,a:*0|int}`, `{one:{price:10}}`, `{one:{a:"rpg.balance.item.one.price+1"}}`)
	before := fileTree(t, s.current.Project.Root)
	if _, err := s.RenameTable(TableRename{Path: path, Namespace: "nil", Name: "goods"}); err == nil {
		t.Fatal("namespace root became an Expr literal")
	}
	assertFileTree(t, s.current.Project.Root, before)
}

func TestFormulaTableRenameDoesNotCaptureBareObjectColumn(t *testing.T) {
	for _, root := range []string{"combat", "rpg"} {
		t.Run(root, func(t *testing.T) {
			s, path := formulaRenameFixture(t, "rpg", fmt.Sprintf("{a:*0|int,%s:{goods:int}}", root), fmt.Sprintf("{one:{%s:{goods:7}}}", root), fmt.Sprintf("{one:{a:%q}}", root+".goods"))
			assertRenamedFormulaValue(t, s, path, "one", "a", "7")
			renamed, err := s.RenameTable(TableRename{Path: path, Namespace: root, Name: "goods"})
			if err != nil {
				t.Fatal(err)
			}
			assertRenamedFormulaValue(t, s, renamed.Table.Path, "one", "a", "7")
		})
	}
}

func TestFormulaRenamesPreserveOptionalFieldAccess(t *testing.T) {
	s, path := formulaRenameFixture(t, "rpg.balance", `{value:*0|int,stats:null|{power:int}}`, `{one:{stats:null}}`, `{one:{value:"rpg.balance.item.one.stats?.power ?? 9"}}`)
	assertRenamedFormulaValue(t, s, path, "one", "value", "9")
	before := fileTree(t, s.current.Project.Root)
	if _, err := s.RenameColumn(ColumnRename{Path: path, OldName: "stats", NewName: "details"}); err != nil {
		t.Fatal(err)
	}
	assertRenamedFormulaValue(t, s, path, "one", "value", "9")
	if _, err := s.RenameRow(RowRename{Path: path, OldID: "one", NewID: "first"}); err != nil {
		t.Fatal(err)
	}
	assertRenamedFormulaValue(t, s, path, "first", "value", "9")
	renamed, err := s.RenameTable(TableRename{Path: path, Namespace: "combat", Name: "goods"})
	if err != nil {
		t.Fatal(err)
	}
	assertRenamedFormulaValue(t, s, renamed.Table.Path, "first", "value", "9")
	after := fileTree(t, s.current.Project.Root)
	for i := 0; i < 3; i++ {
		if _, err := s.Undo(); err != nil {
			t.Fatal(err)
		}
	}
	assertFileTree(t, s.current.Project.Root, before)
	for i := 0; i < 3; i++ {
		if _, err := s.Redo(); err != nil {
			t.Fatal(err)
		}
	}
	assertFileTree(t, s.current.Project.Root, after)
}

func TestFormulaRenameRejectsMalformedMetadataWithoutWrites(t *testing.T) {
	for _, formulas := range []string{`1`, `{one:"ignored"}`, `{one:{price:123}}`} {
		t.Run(formulas, func(t *testing.T) {
			s, path := formulaRenameFixture(t, "rpg.balance", `{price:int,total:*0|int}`, `{one:{price:10}}`, formulas)
			before := fileTree(t, s.current.Project.Root)
			if _, err := s.RenameColumn(ColumnRename{Path: path, OldName: "price", NewName: "cost"}); err == nil {
				t.Fatal("malformed formulas accepted")
			}
			assertFileTree(t, s.current.Project.Root, before)
		})
	}
}

func TestColumnRenameMigratesCommentsAndRejectsOrphanMetadata(t *testing.T) {
	for _, orphan := range []string{"", "comment", "column"} {
		t.Run(orphan, func(t *testing.T) {
			s, path := formulaRenameFixture(t, "rpg.balance", `{price:int,total:*0|int}`, `{one:{price:10}}`, `{one:{total:"price+1"}}`)
			columnMetadata, comments := `{price:{title:"Price",width:200}}`, `{one:{price:"price note",total:"keep note"}}`
			if orphan == "comment" {
				comments = `{one:{price:"price note",cost:"orphan note"}}`
			}
			if orphan == "column" {
				columnMetadata = `{price:{title:"Price"},cost:{title:"orphan metadata"}}`
			}
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			writeFile(t, path, string(source)+"\n_editor:{columns:"+columnMetadata+",cellComments:"+comments+"}\n")
			if _, err := s.OpenProject(s.current.Project.Path); err != nil {
				t.Fatal(err)
			}
			before := fileTree(t, s.current.Project.Root)
			result, err := s.RenameColumn(ColumnRename{Path: path, OldName: "price", NewName: "cost"})
			if orphan != "" {
				if err == nil {
					t.Fatal("orphan metadata was overwritten or attached")
				}
				assertFileTree(t, s.current.Project.Root, before)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if result.Table.CellComments["one"]["cost"] != "price note" || result.Table.CellComments["one"]["total"] != "keep note" || result.Table.CellComments["one"]["price"] != "" || result.Table.Columns[0].DisplayName != "Price" {
				t.Fatalf("column metadata=%#v", result.Table)
			}
			assertRenamedFormulaValue(t, s, path, "one", "total", "11")
			if _, err := s.Undo(); err != nil {
				t.Fatal(err)
			}
			assertFileTree(t, s.current.Project.Root, before)
		})
	}
}
