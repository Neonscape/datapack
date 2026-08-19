package document

import "testing"

func TestSetCellCommentRoundTripsMultilineAndClears(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}}
`)
	comment := "第一行\n带 \"引号\" 和 \\ 符号"
	table, err := SetCellComment(path, "sword", "name", comment)
	if err != nil || table.CellComments["sword"]["name"] != comment {
		t.Fatalf("set comment = %#v, %v", table.CellComments, err)
	}
	table, err = SetCellComment(path, "sword", "name", "")
	if err != nil || len(table.CellComments) != 0 {
		t.Fatalf("clear comment = %#v, %v", table.CellComments, err)
	}
}
