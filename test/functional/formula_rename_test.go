package functional

import (
	"encoding/csv"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"datapack/internal/workspace"
)

func TestFormulaRenameSequenceKeepsExportedValues(t *testing.T) {
	root, project := emptyProject(t)
	path := filepath.Join(root, "tables", "item", "item.cue")
	writeFile(t, path, `package item
#PrimaryKey:string
#Row:{price:int,total:int,label:string}
rows:{one:{price:10,label:"rpg.balance.item.two.price"},two:{price:20,label:"price and pow(2,3)"}}
_tool:{table:{namespace:"rpg.balance",name:"item"},formulas:{one:{total:"rpg.balance.item.two.price+price"},two:{total:"price+1"}}}
_export:{csv:{enabled:true,columns:["id","price","total","label"]}}
`)
	s := openProject(t, project)
	before, err := s.Export()
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(strings.NewReader(readFile(t, before.Files[0]))).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"id", "price", "total", "label"}, {"one", "10", "30", "rpg.balance.item.two.price"}, {"two", "20", "21", "price and pow(2,3)"}}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("initial rows=%#v", rows)
	}
	if _, err := s.RenameColumn(workspace.ColumnRename{Path: path, OldName: "price", NewName: "cost"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RenameRow(workspace.RowRename{Path: path, OldID: "two", NewID: "second"}); err != nil {
		t.Fatal(err)
	}
	renamed, err := s.RenameTable(workspace.TableRename{Path: path, Namespace: "economy.item.balance", Name: "goods"})
	if err != nil {
		t.Fatal(err)
	}
	if renamed.Table.Namespace != "economy.item.balance" {
		t.Fatal(renamed.Table)
	}
	after, err := s.Export()
	if err != nil {
		t.Fatal(err)
	}
	rows, err = csv.NewReader(strings.NewReader(readFile(t, after.Files[0]))).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	want[0][1] = "cost"
	want[2][0] = "second"
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("renames changed computed or literal values: %#v want %#v", rows, want)
	}
	if filepath.Base(after.Files[0]) != "economy.item.balance.goods.csv" {
		t.Fatal(after.Files)
	}
}
