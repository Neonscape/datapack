package cuesource

import "testing"

func TestParseTableExtractsRestrictedTableShape(t *testing.T) {
	table := ParseTable("item.cue", []byte(`package item

	#PrimaryKey: string
#Row: {
    name: string
    price: int
}
rows: {
    sword: {name: "Sword", price: 100}
    potion: {name: "Potion", price: 10}
}
_local: {
    #Position: {x: number, y: number}
    #Discount: >=0
}
_tool: {
    table: {
        namespace: "rpg"
        name: "item"
    }
}`))

	if len(table.Diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", table.Diagnostics)
	}
	if table.Package != "item" || table.LogicalName() != "rpg.item" {
		t.Fatalf("identity = package %q, logical %q", table.Package, table.LogicalName())
	}
	assertStrings(t, table.Columns, []string{"name", "price"})
	assertStrings(t, table.RowIDs, []string{"potion", "sword"})
	assertStrings(t, table.LocalDefinitions, []string{"Discount", "Position"})
}

func TestParseTableRetainsDiagnosticsForIncompleteTable(t *testing.T) {
	table := ParseTable("broken.cue", []byte(`package broken
_tool: {table: {namespace: "rpg"}}
`))

	if len(table.Diagnostics) != 4 {
		t.Fatalf("diagnostic count = %d, want 4: %#v", len(table.Diagnostics), table.Diagnostics)
	}
	if table.Namespace != "rpg" {
		t.Fatalf("namespace = %q, want rpg", table.Namespace)
	}
}

func assertStrings(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("values = %#v, want %#v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("values = %#v, want %#v", got, want)
		}
	}
}
