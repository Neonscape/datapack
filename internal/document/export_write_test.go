package document

import "testing"

func TestSetTableExportsCreatesAndUpdatesCSVSettings(t *testing.T) {
	path := writeSource(t, "item.cue", `package item
#PrimaryKey: string
#Row: {name: string, price: int}
rows: {sword: {name: "Sword", price: 10}}
_tool: {table: {namespace: "rpg", name: "item"}}`)

	table, err := SetTableExports(path, map[string]ExportSettings{"csv": {Enabled: true, Columns: []string{"id", "name"}}})
	if err != nil || table.CSVExport == nil || len(table.CSVExport.Columns) != 2 || !table.Exports["csv"].Enabled {
		t.Fatalf("SetTableExports() = %#v, %v", table, err)
	}
	table, err = SetTableExports(path, map[string]ExportSettings{"csv": {Enabled: false}})
	if err != nil || table.CSVExport != nil || table.Exports["csv"].Enabled {
		t.Fatalf("disabled SetTableExports() = %#v, %v", table, err)
	}
}
