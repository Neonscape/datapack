package foreignkey

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"datapack/internal/document"
)

func defaultTables(t *testing.T, schema, rows string) []document.Table {
	t.Helper()
	root := t.TempDir()
	var tables []document.Table
	for _, source := range []struct{ name, body string }{
		{"effect", `#Row: {name: string}
rows: {burn: {name: "Burn"}}`},
		{"item", fmt.Sprintf("#Row: %s\nrows: %s\n", schema, rows)},
	} {
		path := filepath.Join(root, source.name, source.name+".cue")
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		keys := "{}"
		if source.name == "item" {
			keys = `{target: {table: "rpg.effect"}}`
		}
		text := fmt.Sprintf("package %s\n#PrimaryKey: string\n%s\n_tool: {table: {namespace: \"rpg\", name: %q}, foreignKeys: %s}\n", source.name, source.body, source.name, keys)
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
		table, err := document.LoadTable(path)
		if err != nil {
			t.Fatal(err)
		}
		tables = append(tables, table)
	}
	return tables
}

func TestDefaultReferencesAreIndexedAndValidated(t *testing.T) {
	for _, tc := range []struct {
		name, schema, rows string
		count              int
		invalid            bool
	}{
		{"scalar omitted", `{target: *"burn" | string}`, `{one:{}}`, 1, false},
		{"array omitted", `{target: *["burn","burn"] | [...string]}`, `{one:{}}`, 2, false},
		{"nullable null", `{target: null | string}`, `{one:{target:null}}`, 0, false},
		{"null uses default", `{target: *"burn" | string | null}`, `{one:{target:null}}`, 1, false},
		{"invalid default", `{target: *"missing" | string}`, `{one:{}}`, 1, true},
		{"null forbidden", `{target: string}`, `{one:{target:null}}`, 0, true},
		{"missing required", `{target: string}`, `{one:{}}`, 0, true},
		{"empty string is value", `{target: string}`, `{one:{target:""}}`, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tables := defaultTables(t, tc.schema, tc.rows)
			before, err := os.ReadFile(tables[1].Path)
			if err != nil {
				t.Fatal(err)
			}
			index := Build(tables)
			target := "burn"
			if tc.name == "invalid default" {
				target = "missing"
			}
			if got := len(index.Usages[Target{Table: "rpg.effect", RowID: target}]); got != tc.count {
				t.Fatalf("usages=%#v, want %d", index.Usages, tc.count)
			}
			if (len(index.Diagnostics) > 0) != tc.invalid {
				t.Fatalf("diagnostics=%#v, invalid=%v", index.Diagnostics, tc.invalid)
			}
			after, err := os.ReadFile(tables[1].Path)
			if err != nil || string(before) != string(after) {
				t.Fatal("index changed source")
			}
		})
	}
}

func TestMissingForeignKeyDeclarationsAreRejectedWithoutRows(t *testing.T) {
	for _, change := range []func(*document.Table){
		func(t *document.Table) { t.ForeignKeys[0].Column = "absent" },
		func(t *document.Table) { t.ForeignKeys[0].Table = "rpg.absent" },
	} {
		tables := defaultTables(t, `{target: null | string}`, `{}`)
		change(&tables[1])
		if result := Build(tables); len(result.Diagnostics) == 0 {
			t.Fatal("invalid empty-table declaration accepted")
		}
	}
}

func TestInvalidForeignKeySchemasAreRejectedWithoutRows(t *testing.T) {
	for _, schema := range []string{`{target:int}`, `{target:[...int]}`, `{target:[string,int]}`, `{target:[...(null|string)]}`} {
		t.Run(schema, func(t *testing.T) {
			if index := Build(defaultTables(t, schema, `{}`)); len(index.Diagnostics) == 0 || !index.Incomplete {
				t.Fatalf("invalid declaration accepted: %#v", index)
			}
		})
	}
}

func TestArrayDefaultRetainsUsagePositionsAndDiagnosticLocations(t *testing.T) {
	index := Build(defaultTables(t, `{target: *["burn","missing","missing"] | [...string]}`, `{one:{}}`))
	usages := index.Usages[Target{Table: "rpg.effect", RowID: "missing"}]
	if len(usages) != 2 || usages[0].Index != 1 || usages[1].Index != 2 || usages[0].Source != SourceDefault || usages[1].Source != SourceDefault {
		t.Fatalf("usages=%#v", usages)
	}
	if len(index.Diagnostics) != 2 || !strings.Contains(index.Diagnostics[0].Message, "target[1]") || !strings.Contains(index.Diagnostics[1].Message, "target[2]") {
		t.Fatalf("diagnostics=%#v", index.Diagnostics)
	}
}
