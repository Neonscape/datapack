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

func TestSetCellValueEncodesPlainAndMultilineStrings(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
#Row: {name: string}
rows: {sword: {name: "Sword"}}
_tool: {table: {namespace: "rpg", name: "item"}}
`)
	table, err := SetCellText(path, "sword", "name", "长剑。")
	if err != nil || table.Rows[0].Values["name"].Display != "长剑。" {
		t.Fatalf("plain string = %#v, %v", table, err)
	}
	comment := "第一行\n第二行"
	table, err = SetCellText(path, "sword", "name", comment)
	if err != nil || table.Rows[0].Values["name"].Display != comment {
		t.Fatalf("multiline string = %#v, %v", table, err)
	}
}
